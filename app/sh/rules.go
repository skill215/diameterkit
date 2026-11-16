// Package sh implements a finite Sh AVP slice: Ref24 pull/update, bounded
// Ref11/12/25 subscriptions and unsubscriptions, and independent notifications. UserData remains opaque;
// successful wire validation does not validate XML or accept a business operation.
package sh

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

const ApplicationID uint32 = 16777217

// Application returns an explicit owned VSAI for normal Sh messages. E answers
// have their own grammar and do not require this group. No field is defaulted by
// the codec from an empty struct or from an adjacent DRA's capabilities.
func Application() pdu.VendorSpecificApplicationID {
	return pdu.VendorSpecificApplicationID{VendorID: 10415, AuthApplicationID: avp.Some(ApplicationID)}
}
func key(code uint32) avp.Key { return avp.Key{VendorID: 10415, Code: code} }
func checkApplication(v pdu.VendorSpecificApplicationID) error {
	if v.VendorID != 10415 {
		return validation.At(avp.Key{Code: 260}, validation.At(avp.Key{Code: 266}, validation.ErrField))
	}
	if !v.AuthApplicationID.Present || v.AuthApplicationID.Value != ApplicationID || v.AcctApplicationID.Present {
		return validation.At(avp.Key{Code: 260}, validation.At(avp.Key{Code: 258}, validation.ErrField))
	}
	return nil
}
func unsupported(extra []avp.Raw, keys ...avp.Key) error {
	for _, raw := range extra {
		for _, k := range keys {
			if raw.Key == k {
				return validation.At(k, validation.ErrUnsupported)
			}
		}
	}
	return nil
}

// These known features require another audited slice. Clearing M cannot opt in.
func ordinary(extra []avp.Raw) error {
	return unsupported(extra, avp.Key{Code: 1}, key(634), key(636), key(3111), key(602), key(704), key(706), key(707), key(708), key(710), key(711), key(650), key(713), key(714), key(715), key(717), key(718), key(719), key(720), avp.Key{Code: 621}, avp.Key{Code: 623}, avp.Key{Code: 650})
}
func hexDigit(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}
func publicIdentity(s string) error {
	if strings.HasPrefix(s, "tel:+") {
		digits := s[5:]
		if len(digits) < 1 || len(digits) > 15 || digits[0] == '0' {
			return validation.ErrField
		}
		for i := range digits {
			if digits[i] < '0' || digits[i] > '9' {
				return validation.ErrUnsupported
			}
		}
		return nil
	}
	if !strings.HasPrefix(s, "sip:") {
		return validation.ErrUnsupported
	}
	address := s[4:]
	at := strings.IndexByte(address, '@')
	if at <= 0 || at == len(address)-1 || strings.LastIndexByte(address, '@') != at {
		return validation.ErrField
	}
	user, host := address[:at], address[at+1:]
	// This intentionally selects DNS hosts, not the full SIP URI grammar. URI
	// equality is not DiameterIdentity equality; never normalize the supplied URI.
	if strings.ContainsAny(host, ":;?[]") {
		return validation.ErrUnsupported
	}
	if err := avp.Identity(host); err != nil {
		return validation.ErrField
	}
	last := host[strings.LastIndexByte(host, '.')+1:]
	if !(last[0] >= 'a' && last[0] <= 'z' || last[0] >= 'A' && last[0] <= 'Z') {
		return validation.ErrUnsupported
	}
	for i := 0; i < len(user); i++ {
		b := user[i]
		if b == '%' {
			if i+2 >= len(user) || !hexDigit(user[i+1]) || !hexDigit(user[i+2]) {
				return validation.ErrField
			}
			i += 2
			continue
		}
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("-_.!~*'()&=+$,;?/", rune(b)) {
			continue
		}
		return validation.ErrUnsupported
	}
	return nil
}
func validateIdentitySubset(m *UserIdentity, _ validation.Mode) error {
	if m.PublicIdentity.Present == m.MSISDN.Present {
		return validation.ErrField
	}
	if m.PublicIdentity.Present {
		if err := publicIdentity(m.PublicIdentity.Value); err != nil {
			return validation.At(key(601), err)
		}
	}
	return unsupported(m.Extra, key(3111), avp.Key{Code: 1})
}
func validateUDRSubset(m *UDR, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, false, mode); err != nil {
		return err
	}
	if err := checkApplication(m.VendorSpecificApplicationID); err != nil {
		return err
	}
	if m.DataReference != 24 {
		return validation.At(key(703), validation.ErrUnsupported)
	}
	return ordinary(m.Extra)
}
func validatePURSubset(m *PUR, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, false, mode); err != nil {
		return err
	}
	if err := checkApplication(m.VendorSpecificApplicationID); err != nil {
		return err
	}
	if m.DataReference != 24 {
		return validation.At(key(703), validation.ErrUnsupported)
	}
	return ordinary(m.Extra)
}
func validateSNRSubset(m *SNR, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, false, mode); err != nil {
		return err
	}
	if err := checkApplication(m.VendorSpecificApplicationID); err != nil {
		return err
	}
	if m.SubsReqType > 1 {
		return validation.At(key(705), validation.ErrUnsupported)
	}
	var seen uint8
	for _, ref := range m.DataReference {
		bit := refBit(ref)
		if bit == 0 {
			return validation.At(key(703), validation.ErrUnsupported)
		}
		if seen&bit != 0 {
			return validation.At(key(703), validation.ErrDuplicate)
		}
		seen |= bit
	}
	if seen&3 != 0 && !m.UserIdentity.PublicIdentity.Present {
		return validation.At(key(700), validation.ErrUnsupported)
	}
	if seen&4 != 0 {
		if !m.OneTimeNotification.Present {
			return validation.At(key(712), validation.ErrMissing)
		}
		if m.OneTimeNotification.Value != 0 || m.SendDataIndication.Present {
			return validation.ErrUnsupported
		}
	} else if m.OneTimeNotification.Present {
		return validation.At(key(712), validation.ErrUnsupported)
	}
	if m.SendDataIndication.Present && m.SendDataIndication.Value > 1 {
		return validation.At(key(710), validation.ErrUnsupported)
	}
	if len(m.DataReference) > 1 && !hasNotif(m.SupportedFeatures, mode != validation.Receive) {
		return validation.At(key(628), validation.ErrMissing)
	}

	return ordinary(m.Extra)
}
func validatePNRSubset(m *PNR, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, false, mode); err != nil {
		return err
	}
	if err := checkApplication(m.VendorSpecificApplicationID); err != nil {
		return err
	}
	if err := unsupported(m.Extra, key(703), key(705), key(709), key(712)); err != nil {
		return err
	}
	return ordinary(m.Extra)
}
func result(v pdu.VendorSpecificApplicationID, code avp.Optional[uint32], experimental avp.Optional[ExperimentalResult], failed bool, extra []avp.Raw, mode validation.Mode, allowed ...uint32) (bool, error) {
	if err := checkApplication(v); err != nil {
		return false, err
	}
	if err := ordinary(extra); err != nil {
		return false, err
	}
	if code.Present == experimental.Present {
		return false, validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	if experimental.Present {
		e := experimental.Value
		if e.VendorID == 10415 {
			for _, c := range allowed {
				if c == e.ExperimentalResultCode {
					return false, nil
				}
			}
		}
		return false, validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	switch code.Value {
	case 2001:
		return true, nil
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		return false, base.ValidateFailureDiagnostic(code.Value, failed, mode)
	default:
		return false, validation.At(avp.Key{Code: 268}, validation.ErrUnsupported)
	}
}
func validateUDAResult(m *UDA, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, true, mode); err != nil {
		return err
	}
	success, err := result(m.VendorSpecificApplicationID, m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, m.Extra, mode, 5011, 5001, 5101, 5102, 4100)
	if err != nil {
		return err
	}
	if success && !m.UserData.Present {
		return validation.At(key(702), validation.ErrMissing)
	}
	if !success && m.UserData.Present {
		return validation.At(key(702), validation.ErrField)
	}
	return nil
}
func validatePUAResult(m *PUA, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, true, mode); err != nil {
		return err
	}
	if err := unsupported(m.Extra, key(703), key(702), key(709)); err != nil {
		return err
	}
	_, err := result(m.VendorSpecificApplicationID, m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, m.Extra, mode, 5011, 5001, 5101, 5103, 4101)
	return err
}
func validateSNAResult(m *SNA, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, true, mode); err != nil {
		return err
	}
	success, err := result(m.VendorSpecificApplicationID, m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, m.Extra, mode, 5011, 5001, 5101, 5104)
	if err != nil {
		return err
	}
	if !success {
		if m.UserData.Present {
			return validation.At(key(702), validation.ErrField)
		}
		if m.ExpiryTime.Present {
			return validation.At(key(709), validation.ErrField)
		}
	}
	return nil
}
func validatePNAResult(m *PNA, mode validation.Mode) error {
	if err := featureShape(m.SupportedFeatures, true, mode); err != nil {
		return err
	}
	if err := unsupported(m.Extra, key(703), key(702), key(709)); err != nil {
		return err
	}
	_, err := result(m.VendorSpecificApplicationID, m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, m.Extra, mode, 5011, 5001, 5008, 5100, 5107)
	return err
}
