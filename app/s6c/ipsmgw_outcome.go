package s6c

import "github.com/skill215/diameterkit/validation"

// The selected ordinary-MSISDN, T4 and IMS report slices keep each parent's cause. Unknown
// values on receive remain owned for a scope report; they never imply cause1,
// a default diagnostic, or permission to update MWD or schedule delivery.
func validateIPSMGWOutcomeSubset(m *IPSMGWSMDeliveryOutcome, mode validation.Mode) error {
	if !m.SMDeliveryCause.Present {
		return validation.At(smsKey(3321), validation.ErrMissing)
	}
	cause := m.SMDeliveryCause.Value
	if m.AbsentUserDiagnosticSM.Present {
		d := m.AbsentUserDiagnosticSM.Value
		if d > 255 || cause == 0 || cause == 2 {
			return validation.At(smsKey(3322), validation.ErrField)
		}
	}
	if mode != validation.Receive {
		if cause > 2 {
			return validation.At(smsKey(3321), validation.ErrUnsupported)
		}
		if m.AbsentUserDiagnosticSM.Present && m.AbsentUserDiagnosticSM.Value != 11 && m.AbsentUserDiagnosticSM.Value != 12 {
			return validation.At(smsKey(3322), validation.ErrUnsupported)
		}
	}
	return nil
}
