// Package sgd implements selected SGd/Gdd MO/MT, T4 and IMS forwarding slices.
// TPDU bytes are opaque. Delivery, retry, subscriber lookup and persistence belong
// to the application; the selected grammar is not full TS 29.338 support.
package sgd

import (
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/validation"
)

func smsKey(code uint32) avp.Key { return avp.Key{VendorID: 10415, Code: code} }

// MaxRouterAlertTokenBytes is the selected UTF8 token bound shared with S6c ALR.
const MaxRouterAlertTokenBytes = smsidentity.MaxAlertTokenBytes

func validateAlertIdentity(m *AlertUserIdentifier, _ validation.Mode) error {
	if err := smsidentity.AlertToken(m.UserName); err != nil {
		return validation.At(avp.Key{Code: 1}, err)
	}
	return unsupportedFields(m.Extra, 701, 3111, 2400)
}
func unsupportedFields(extra []avp.Raw, codes ...uint32) error {
	for _, raw := range extra {
		for _, code := range codes {
			if raw.Key == smsKey(code) {
				return validation.At(raw.Key, validation.ErrUnsupported)
			}
		}
	}
	return nil
}
func imsi(s string) error {
	if len(s) < 5 || len(s) > 15 {
		return validation.ErrField
	}
	for _, b := range s {
		if b < '0' || b > '9' {
			return validation.ErrField
		}
	}
	return nil
}
func tpdu(b []byte) error {
	// TS 29.338 bounds SM-RP-UI to 200 octets. No TPDU interpretation or
	// additional minimum length is inferred from the opaque OctetString type.
	if len(b) > 200 {
		return validation.At(smsKey(3301), validation.ErrField)
	}
	return nil
}
func validateIdentitySubset(m *UserIdentifier, _ validation.Mode) error {
	if m.UserName.Present {
		if err := imsi(m.UserName.Value); err != nil {
			return validation.At(avp.Key{Code: 1}, err)
		}
	}
	return unsupportedFields(m.Extra, 3111, 2400)
}
func flags(v avp.Optional[uint32], code uint32, mode validation.Mode) error {
	if mode != validation.Receive && v.Present && v.Value & ^uint32(1) != 0 {
		return validation.At(smsKey(code), validation.ErrUnsupported)
	}
	return nil
}
func validateOFRSubset(m *OFR, mode validation.Mode) error {
	if err := tpdu(m.SMRPUI); err != nil {
		return err
	}
	if err := flags(m.OFRFlags, 3328, mode); err != nil {
		return err
	}
	if m.OFRFlags.Present && m.OFRFlags.Value&1 != 0 && m.EPSLocationInformation.Present {
		return validation.At(smsKey(1496), validation.ErrUnsupported)
	}
	if m.SMSMICorrelationID.Present {
		if m.UserIdentifier.UserName.Present {
			if m.UserIdentifier.MSISDN.Present {
				digits, err := avp.DecodeTBCD(m.UserIdentifier.MSISDN.Value)
				if err != nil || digits == "000000000000000" {
					return validation.At(smsKey(3102), validation.At(smsKey(701), validation.ErrField))
				}
			}
		} else {
			if !m.UserIdentifier.MSISDN.Present {
				return validation.At(smsKey(3102), validation.At(smsKey(701), validation.ErrMissing))
			}
			digits, err := avp.DecodeTBCD(m.UserIdentifier.MSISDN.Value)
			if err != nil || digits != "000000000000000" {
				return validation.At(smsKey(3102), validation.At(smsKey(701), validation.ErrField))
			}
		}
		if !m.SMDeliveryOutcome.Present {
			return validation.At(smsKey(3316), validation.ErrMissing)
		}
		// This bounded IP-SM-GW slice does not select access-node flags/location.
		if m.OFRFlags.Present || m.EPSLocationInformation.Present {
			return validation.ErrUnsupported
		}
	} else {
		if !m.UserIdentifier.UserName.Present {
			return validation.At(smsKey(3102), validation.At(avp.Key{Code: 1}, validation.ErrMissing))
		}
		if m.SMDeliveryOutcome.Present {
			return validation.At(smsKey(3316), validation.ErrUnsupported)
		}
	}
	return unsupportedFields(m.Extra, 3325, 3326, 3327, 3320, 3321, 3322)
}
func validateTFRSubset(m *TFR, mode validation.Mode) error {
	checkIdentity := imsi
	if m.SMSMICorrelationID.Present {
		checkIdentity = smsidentity.HSSID
	}
	if err := checkIdentity(m.UserName); err != nil {
		return validation.At(avp.Key{Code: 1}, err)
	}
	if err := tpdu(m.SMRPUI); err != nil {
		return err
	}
	if m.SMSMICorrelationID.Present {
		if m.MMENumberForMTSMS.Present || m.SGSNNumber.Present {
			return validation.At(smsKey(3324), validation.ErrField)
		}
	} else if m.MMENumberForMTSMS.Present && m.SGSNNumber.Present {
		return validation.At(smsKey(1645), validation.ErrField)
	}
	if err := flags(m.TFRFlags, 3302, mode); err != nil {
		return err
	}
	if m.MaximumRetransmissionTime.Present && !m.SMSGMSCAddress.Present {
		return validation.At(smsKey(3332), validation.ErrMissing)
	}
	if mode != validation.Receive && m.SMSMICorrelationID.Present {
		if err := validateIMSMTRequest(m); err != nil {
			return err
		}
	}
	return unsupportedFields(m.Extra, 3325, 3326, 3327)
}
func plmn(b []byte) bool {
	// MCC digits: octet1 low/high, octet2 low; MNC digits: octet3 low/high,
	// octet2 high (F only for a two-digit MNC). Leading zeroes remain significant.
	return b[0]&15 <= 9 && b[0]>>4 <= 9 && b[1]&15 <= 9 && (b[1]>>4 <= 9 || b[1]>>4 == 15) && b[2]&15 <= 9 && b[2]>>4 <= 9
}
func validateMMELocationSubset(m *MMELocationInformation, _ validation.Mode) error {
	if m.EUTRANCellGlobalIdentity.Present {
		b := m.EUTRANCellGlobalIdentity.Value
		if len(b) != 7 || !plmn(b) || b[3]&0xf0 != 0 {
			return validation.At(smsKey(1602), validation.ErrField)
		}
	}
	if m.TrackingAreaIdentity.Present {
		b := m.TrackingAreaIdentity.Value
		if len(b) != 5 || !plmn(b) {
			return validation.At(smsKey(1603), validation.ErrField)
		}
		if b[3] == 0 && b[4] == 0 || b[3] == 255 && b[4] == 254 {
			return validation.At(smsKey(1603), validation.ErrUnsupported)
		}
	}
	// Other location leaves carry freshness/current-location semantics that this
	// ordinary ECGI/TAI slice has not audited, even if their M bit is cleared.
	return unsupportedFields(m.Extra, 1608, 1609, 1610, 1611, 2319, 4013)
}
func validateEPSSubset(m *EPSLocationInformation, _ validation.Mode) error {
	return unsupportedFields(m.Extra, 1601)
}
func result(code avp.Optional[uint32], experimental avp.Optional[ExperimentalResult], failed bool, mode validation.Mode) (uint32, error) {
	if code.Present == experimental.Present {
		return 0, validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	if experimental.Present {
		if experimental.Value.VendorID != 10415 || experimental.Value.ExperimentalResultCode == 0 {
			return 0, validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
		return experimental.Value.ExperimentalResultCode, nil
	}
	switch code.Value {
	case 2001:
		return 0, nil
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		return 0, base.ValidateFailureDiagnostic(code.Value, failed, mode)
	default:
		return 0, validation.At(avp.Key{Code: 268}, validation.ErrUnsupported)
	}
}
func failure(cause avp.Optional[SMDeliveryFailureCause], code, lo, hi uint32) error {
	if cause.Present != (code == 5555) {
		return validation.At(smsKey(3303), validation.ErrField)
	}
	if cause.Present && (cause.Value.SMEnumeratedDeliveryFailureCause < lo || cause.Value.SMEnumeratedDeliveryFailureCause > hi) {
		return validation.At(smsKey(3303), validation.At(smsKey(3304), validation.ErrUnsupported))
	}
	return nil
}
func validateOFAResult(m *OFA, mode validation.Mode) error {
	code, err := result(m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, mode)
	if err != nil {
		return err
	}
	if code != 0 && code != 5001 && code != 5552 && code != 5555 {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	if err = failure(m.SMDeliveryFailureCause, code, 3, 6); err != nil {
		return err
	}
	if m.SMRPUI.Present {
		if err = tpdu(m.SMRPUI.Value); err != nil {
			return err
		}
	}
	if m.ExternalIdentifier.Present {
		err := externalIdentifierASCII(m.ExternalIdentifier.Value)
		// Owned receive preserves syntactically UTF8 text outside the selected
		// semantic subset. Complete contextual checks still report unsupported.
		if err != nil && !(mode == validation.Receive && errors.Is(err, validation.ErrUnsupported)) {
			return validation.At(smsKey(3111), err)
		}
	}
	return unsupportedFields(m.Extra, 3102)
}
func validateTFAResult(m *TFA, mode validation.Mode) error {
	code, err := result(m.ResultCode, m.ExperimentalResult, m.FailedAVP.Present, mode)
	if err != nil {
		return err
	}
	switch code {
	case 0, 5001, 5550, 5551, 5552, 5553, 5554, 5555, 5557:
	default:
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	if err = failure(m.SMDeliveryFailureCause, code, 0, 2); err != nil {
		return err
	}
	if m.AbsentUserDiagnosticSM.Present {
		if code != 5550 || m.AbsentUserDiagnosticSM.Value > 255 {
			return validation.At(smsKey(3322), validation.ErrField)
		}
		if m.AbsentUserDiagnosticSM.Value != 13 {
			return validation.At(smsKey(3322), validation.ErrUnsupported)
		}
	}
	if m.EPSLocationInformation.Present && (!m.ResultCode.Present || m.ResultCode.Value != 2001) {
		return validation.At(smsKey(1496), validation.ErrUnsupported)
	}
	if m.SMRPUI.Present {
		if err = tpdu(m.SMRPUI.Value); err != nil {
			return err
		}
	}
	if m.RequestedRetransmissionTime.Present && code != 5550 {
		return validation.At(smsKey(3331), validation.ErrField)
	}
	// The selected Router token output is the time-bearing 5550 branch.
	// Tokens on other results are not certified by this bounded slice.
	if m.UserIdentifier.Present && !m.RequestedRetransmissionTime.Present {
		return validation.At(smsKey(3102), validation.ErrUnsupported)
	}
	return nil
}
func validateTFARelation(request *TFR, answer *TFA) error {
	if answer.RequestedRetransmissionTime.Present && !request.MaximumRetransmissionTime.Present {
		return validation.At(smsKey(3331), validation.ErrField)
	}
	if !request.MMENumberForMTSMS.Present {
		if answer.ExperimentalResult.Present && answer.ExperimentalResult.Value.ExperimentalResultCode == 5557 {
			return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
		if answer.EPSLocationInformation.Present {
			return validation.At(smsKey(1496), validation.ErrUnsupported)
		}
	}
	return nil
}

// The SGd unknown-user result refers to the original MSISDN, not its IMSI.
func validateOFARelation(request *OFR, answer *OFA) error {
	if moUnknown(answer) && !request.UserIdentifier.MSISDN.Present {
		return validation.At(smsKey(3102), validation.At(smsKey(701), validation.ErrMissing))
	}
	return nil
}
