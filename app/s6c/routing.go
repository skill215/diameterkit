package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// nodeShape selects five explicit ordinary-HSS representations. A combined
// MSC number + MME name/realm is one logical node, not a two-node result.
func nodeShape(sname, srealm, snumber, mname, mrealm, mnumber, msc bool) (sgsn bool, err error) {
	if snumber && !mname && !mrealm && !mnumber && !msc && sname == srealm {
		return true, nil
	}
	if !sname && !srealm && !snumber {
		if !mname && !mrealm && !mnumber && msc {
			return false, nil
		}
		if mname && mrealm && mnumber != msc {
			return false, nil
		}
	}
	return false, validation.ErrField
}
func nodeExtra(extra []avp.Raw) error {
	// SMSF groups remain outside this slice, even when a known key arrives M0.
	return unsupportedFields(extra, smsKey(3338), smsKey(3339), smsKey(3340), smsKey(3341), smsKey(3342), smsKey(3343))
}
func validateServingNodeSubset(m *ServingNode, mode validation.Mode) error {
	if err := nodeExtra(m.Extra); err != nil {
		return err
	}
	if hasIPSMGW(m.IPSMGWNumber, m.IPSMGWName, m.IPSMGWRealm, m.IPSMGWSBISupportIndicator) {
		if err := checkIPSMGWShape(m.IPSMGWNumber, m.SGSNName.Present || m.SGSNRealm.Present || m.SGSNNumber.Present || m.MMEName.Present || m.MMERealm.Present || m.MMENumber.Present || m.MSCNumber.Present); err != nil {
			return err
		}
		if mode != validation.Receive {
			return checkIPSMGWSelection(m.IPSMGWRealm, m.IPSMGWSBISupportIndicator)
		}
		return nil
	}
	_, err := nodeShape(m.SGSNName.Present, m.SGSNRealm.Present, m.SGSNNumber.Present, m.MMEName.Present, m.MMERealm.Present, m.MMENumber.Present, m.MSCNumber.Present)
	return err
}
func validateAdditionalNodeSubset(m *AdditionalServingNode, mode validation.Mode) error {
	if err := nodeExtra(m.Extra); err != nil {
		return err
	}
	if hasIPSMGW(m.IPSMGWNumber, m.IPSMGWName, m.IPSMGWRealm, m.IPSMGWSBISupportIndicator) {
		if err := checkIPSMGWShape(m.IPSMGWNumber, m.SGSNName.Present || m.SGSNRealm.Present || m.SGSNNumber.Present || m.MMEName.Present || m.MMERealm.Present || m.MMENumber.Present || m.MSCNumber.Present); err != nil {
			return err
		}
		if mode != validation.Receive {
			return validation.At(smsKey(3100), validation.ErrUnsupported)
		}
		return nil
	}
	_, err := nodeShape(m.SGSNName.Present, m.SGSNRealm.Present, m.SGSNNumber.Present, m.MMEName.Present, m.MMERealm.Present, m.MMENumber.Present, m.MSCNumber.Present)
	return err
}
func validateSRAAdjuncts(m *SRA, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3344), smsKey(3345), smsKey(3334), smsKey(3335), smsKey(3100), smsKey(3101), smsKey(3112), smsKey(3348)); err != nil {
		return err
	}
	if m.LMSI.Present && len(m.LMSI.Value) != 4 {
		return validation.At(smsKey(2400), validation.ErrField)
	}
	if mode != validation.Receive && m.MWDStatus.Present && m.MWDStatus.Value&^uint32(31) != 0 {
		return validation.At(smsKey(3312), validation.ErrUnsupported)
	}
	if m.AdditionalServingNode.Present {
		if !m.ServingNode.Present {
			return validation.At(smsKey(2401), validation.ErrMissing)
		}
		if m.ServingNode.Value.IPSMGWNumber.Present || m.AdditionalServingNode.Value.IPSMGWNumber.Present {
			// Only the single IP-SM-GW serving-node result is selected. Keep
			// well-typed received combinations for an explicit scope report.
			if mode != validation.Receive {
				return validation.At(smsKey(2406), validation.ErrUnsupported)
			}
			return nil
		}
		// In this ordinary-HSS profile, precisely one of the two nodes must be SGSN.
		// This excludes duplicate SGSNs, duplicate MSC/MME, and split MSC+MME.
		if m.ServingNode.Value.SGSNNumber.Present == m.AdditionalServingNode.Value.SGSNNumber.Present {
			return validation.At(smsKey(2406), validation.ErrField)
		}
	}
	return nil
}
func checkSRANodeRequest(request *SRR, answer *SRA) error {
	if answer.ResultCode.Present && answer.ResultCode.Value == 2001 && !answer.ServingNode.Present && !request.SMDeliveryNotIntended.Present {
		return validation.At(smsKey(2401), validation.ErrMissing)
	}
	// A T4 request IMSI cannot change. IMS requests carry an HSS-ID instead;
	// equality or inequality of its bytes cannot establish the answer identity.
	if request.UserName.Present && !request.SMSMICorrelationID.Present && answer.UserName.Present && request.UserName.Value != answer.UserName.Value {
		return validation.At(avp.Key{Code: 1}, validation.ErrInvalidAnswer)
	}
	if answer.AdditionalServingNode.Present && (!request.SRRFlags.Present || request.SRRFlags.Value&1 == 0) {
		return validation.At(smsKey(2406), validation.ErrInvalidAnswer)
	}
	return nil
}

// EncodeSMEA constructs only the international decimal address subset of
// SM-RP-SMEA: digit count, TOA=0x91, then TBCD digits. It never derives an SME
// from SC-Address or MSISDN, and does not add MAP IEI/length wrappers.
func EncodeSMEA(digits string) ([]byte, error) {
	b, err := avp.EncodeTBCD(digits)
	if err != nil {
		return nil, validation.ErrField
	}
	out := make([]byte, 2+len(b))
	out[0] = byte(len(digits))
	out[1] = 0x91
	copy(out[2:], b)
	return out, nil
}

// DecodeSMEA distinguishes unsupported TON/NPI from malformed selected content.
// It does not validate number assignment, normalize digits or convert other TONs.
func DecodeSMEA(b []byte) (string, error) {
	if len(b) < 3 || len(b) > 12 || b[1]&0x80 == 0 {
		return "", validation.ErrField
	}
	if b[1] != 0x91 {
		return "", validation.ErrUnsupported
	}
	n := int(b[0])
	if n < 1 || n > 15 || len(b) != 2+(n+1)/2 {
		return "", validation.ErrField
	}
	s, err := avp.DecodeTBCD(b[2:])
	if err != nil || len(s) != n {
		return "", validation.ErrField
	}
	return s, nil
}
