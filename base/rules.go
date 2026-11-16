// Package base contains the explicitly named Base control codec subset.
// It does not implement a Peer state machine or capability negotiation.
package base

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// BASE-NORMAL-RESULT-1: RFC 6733 section 7.1; 3xxx uses the separate E grammar.
func normalResult(code uint32) error {
	if code < 1000 || code >= 6000 || code >= 3000 && code < 4000 {
		return validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	return nil
}
func validateCEAResult(m *CEA, mode validation.Mode) error {
	if err := normalResult(m.ResultCode); err != nil {
		return err
	}
	return ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode)
}
func validateDWAResult(m *DWA, mode validation.Mode) error {
	if err := normalResult(m.ResultCode); err != nil {
		return err
	}
	return ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode)
}
func validateDPAResult(m *DPA, mode validation.Mode) error {
	if err := normalResult(m.ResultCode); err != nil {
		return err
	}
	return ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode)
}
