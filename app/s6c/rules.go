// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-18

// Package s6c contains bounded SMS routing, reporting and alert codec slices.
// Each command has its own selected result and field branches.
// Common E envelopes are separate; this is not the complete S6c grammar.
package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/validation"
)

// Shared success syntax; request/context selects IMSI, Router token or MCCMNC.
func validateSRAResult(m *SRA, mode validation.Mode) error {
	if err := validateSRAAdjuncts(m, mode); err != nil {
		return err
	}
	if m.ResultCode.Present == m.ExperimentalResult.Present {
		return validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	absent := m.ExperimentalResult.Present && m.ExperimentalResult.Value.VendorID == 10415 && m.ExperimentalResult.Value.ExperimentalResultCode == 5550
	if err := validateSRADiagnostics(m, mode); err != nil {
		return err
	}
	if hasSRADiagnostics(m) && !absent {
		return validation.ErrUnsupported
	} // Selected failure slice, not a universal prohibition.
	if m.ResultCode.Present {
		if m.ResultCode.Value != 2001 {
			if err := validateSMSBaseResult(m.ResultCode.Value, m.FailedAVP.Present, mode); err != nil {
				return err
			}
			if hasSRASuccessData(m) {
				return validation.ErrField
			}
			return nil
		}
		if !m.UserName.Present {
			return validation.ErrMissing
		}
	} else {
		r := m.ExperimentalResult.Value
		if r.VendorID != 10415 {
			return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
		switch r.ExperimentalResultCode {
		case 5001, 5550, 5552, 5556, 5557:
		default:
			return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
		if absent {
			if m.ServingNode.Present || m.AdditionalServingNode.Present || m.LMSI.Present {
				return validation.ErrField
			}
		} else if hasSRASuccessData(m) {
			return validation.ErrField
		}
	}
	if m.UserName.Present {
		if m.ResultCode.Present && m.ResultCode.Value == 2001 {
			return validation.At(avp.Key{Code: 1}, smsidentity.AlertToken(m.UserName.Value))
		}
		return validation.At(avp.Key{Code: 1}, imsi(m.UserName.Value))
	}
	return nil
}
func hasSRASuccessData(m *SRA) bool {
	return m.UserName.Present || m.ServingNode.Present || m.AdditionalServingNode.Present || m.LMSI.Present || m.UserIdentifier.Present || m.MWDStatus.Present
}

func validateSRRSubset(m *SRR, mode validation.Mode) error {
	if !m.UserName.Present && !m.MSISDN.Present {
		return validation.ErrMissing
	}
	if err := validateSMSRequestIdentity(m.UserName, m.MSISDN, nil); err != nil {
		return err
	}
	if m.SMRPMTI.Present && m.SMRPMTI.Value > 1 {
		return validation.At(smsKey(3308), validation.ErrUnsupported)
	}
	if mode != validation.Receive && m.SRRFlags.Present && m.SRRFlags.Value&^uint32(7) != 0 {
		return validation.At(smsKey(3310), validation.ErrUnsupported)
	}
	if m.SRRFlags.Present && m.SRRFlags.Value&1 != 0 && !m.SMRPSMEA.Present {
		if mode == validation.Receive {
			return validation.At(smsKey(3309), &srrSMEAMissing{flags: m.SRRFlags.Value})
		}
		return validation.At(smsKey(3309), validation.ErrMissing)
	}
	if m.SMRPSMEA.Present {
		if _, err := DecodeSMEA(m.SMRPSMEA.Value); err != nil {
			return validation.At(smsKey(3309), err)
		}
	}
	if m.SMDeliveryNotIntended.Present && m.SMDeliveryNotIntended.Value != OnlyMCCMNCRequested {
		return validation.At(smsKey(3311), validation.ErrUnsupported)
	}
	// Known alternative identity/delivery modes cannot opt in by clearing M.
	if m.SMSMICorrelationID.Present {
		if !m.UserName.Present || m.MSISDN.Present {
			return validation.ErrField
		}
		if err := hssID(m.UserName.Value); err != nil {
			return validation.At(avp.Key{Code: 1}, err)
		}
	}
	return unsupportedFields(m.Extra, smsKey(3325), smsKey(3326), smsKey(3327))
}

// All field restrictions below are the declared slice, not full S6c support.
// Known out-of-slice fields are rejected even with M cleared; unknown M=0
// extensions remain owned opaque data and cannot satisfy any required field.
func unsupportedFields(extra []avp.Raw, keys ...avp.Key) error {
	for _, raw := range extra {
		for _, k := range keys {
			if raw.Key == k {
				return validation.At(k, validation.ErrUnsupported)
			}
		}
	}
	return nil
}
func smsKey(code uint32) avp.Key { return avp.Key{VendorID: 10415, Code: code} }
func validateUserIdentifierSubset(m *UserIdentifier, _ validation.Mode) error {
	return unsupportedFields(m.Extra, avp.Key{Code: 1}, smsKey(3111), smsKey(2400))
}

// A reported node is productive in this slice: it includes its own cause.
// Other node fields and reserved diagnostics need a separately audited scope.
func validateNodeOutcome(cause, diagnostic avp.Optional[uint32], allowed uint16) error {
	if !cause.Present {
		return validation.At(smsKey(3321), validation.ErrMissing)
	}
	if cause.Value > 2 {
		return validation.At(smsKey(3321), validation.ErrUnsupported)
	}
	if diagnostic.Present {
		if cause.Value != 1 || diagnostic.Value > 255 {
			return validation.At(smsKey(3322), validation.ErrField)
		}
		if diagnostic.Value > 13 || allowed&(1<<diagnostic.Value) == 0 {
			return validation.At(smsKey(3322), validation.ErrUnsupported)
		}
	}
	return nil
}
func validateMMEOutcomeSubset(m *MMESMDeliveryOutcome, _ validation.Mode) error {
	return validateNodeOutcome(m.SMDeliveryCause, m.AbsentUserDiagnosticSM, 1<<13)
}
func validateMSCOutcomeSubset(m *MSCSMDeliveryOutcome, _ validation.Mode) error {
	return validateNodeOutcome(m.SMDeliveryCause, m.AbsentUserDiagnosticSM, 1<<0|1<<1|1<<2|1<<3|1<<4|1<<9|1<<13)
}
func validateSGSNOutcomeSubset(m *SGSNSMDeliveryOutcome, _ validation.Mode) error {
	return validateNodeOutcome(m.SMDeliveryCause, m.AbsentUserDiagnosticSM, 1<<2|1<<5|1<<6|1<<7|1<<8|1<<10|1<<13)
}
func validateOutcomeSubset(m *SMDeliveryOutcome, _ validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3336), smsKey(3337), smsKey(3321), smsKey(3322)); err != nil {
		return err
	}
	if !m.MMESMDeliveryOutcome.Present && !m.MSCSMDeliveryOutcome.Present && !m.SGSNSMDeliveryOutcome.Present && !m.IPSMGWSMDeliveryOutcome.Present {
		return validation.ErrMissing
	}
	if m.MMESMDeliveryOutcome.Present && m.MSCSMDeliveryOutcome.Present {
		return validation.At(smsKey(3318), validation.ErrField)
	}
	return nil
}
func validateRDRSubset(m *RDR, mode validation.Mode) error {
	if mode != validation.Receive && m.RDRFlags.Present && m.RDRFlags.Value & ^uint32(1) != 0 {
		return validation.At(smsKey(3323), validation.ErrUnsupported)
	}
	if m.SMSMICorrelationID.Present {
		if !m.UserIdentifier.UserName.Present || m.UserIdentifier.MSISDN.Present {
			return validation.At(smsKey(3102), validation.ErrField)
		}
		if err := hssID(m.UserIdentifier.UserName.Value); err != nil {
			return validation.At(smsKey(3102), validation.At(avp.Key{Code: 1}, err))
		}
	}
	return unsupportedFields(m.Extra, smsKey(3325), smsKey(3326), smsKey(3327))
}
func validateSMSBaseResult(code uint32, diagnostic bool, mode validation.Mode) error {
	switch code {
	case 2001:
		return nil
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		return base.ValidateFailureDiagnostic(code, diagnostic, mode)
	default:
		return validation.At(avp.Key{Code: 268}, validation.ErrUnsupported)
	}
}
func validateRDAResult(m *RDA, mode validation.Mode) error {
	if m.ResultCode.Present == m.ExperimentalResult.Present {
		return validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	if m.ResultCode.Present {
		if err := validateSMSBaseResult(m.ResultCode.Value, m.FailedAVP.Present, mode); err != nil {
			return err
		}
	} else {
		r := m.ExperimentalResult.Value
		if r.VendorID != 10415 || r.ExperimentalResultCode != 5001 && r.ExperimentalResultCode != 5558 {
			return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
	}
	if m.UserIdentifier.Present && (!m.ResultCode.Present || m.ResultCode.Value != 2001) {
		return validation.At(smsKey(3102), validation.ErrField)
	}
	return nil
}
func validateALAResult(m *ALA, mode validation.Mode) error {
	if m.ResultCode.Present == m.ExperimentalResult.Present {
		return validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	if m.ExperimentalResult.Present {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	return validateSMSBaseResult(m.ResultCode.Value, m.FailedAVP.Present, mode)
}
