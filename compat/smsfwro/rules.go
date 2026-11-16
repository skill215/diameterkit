// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package smsfwro implements the explicitly selected sms-fw IEC and no-UPDATE ECUR template dialect.
// Compile its Profile in a dedicated Bundle, separately from app/ro. The caller
// binds that Bundle to an approved OCS deployment; application ID 4 alone does
// not identify this dialect. No template files are read at runtime.
package smsfwro

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// This named template family is closed: no unproved private fields, alternate
// parents or type inference through Extra, including M=0. E answers retain the
// independent Base error envelope. Subsequent dialects require explicit schema.
func extras(xs []avp.Raw) error {
	if len(xs) > 0 {
		return validation.At(xs[0].Key, validation.ErrUnsupported)
	}
	return nil
}
func at(c uint32, e error) error       { return validation.At(avp.Key{Code: c}, e) }
func vendorAt(c uint32, e error) error { return validation.At(avp.Key{VendorID: 10415, Code: c}, e) }
func cstring(v []byte) error {
	// The original writer used strlen, not a UTF-8 or Address encoder. Reject
	// embedded NUL instead of silently discarding the suffix; empty is explicit.
	if bytes.IndexByte(v, 0) >= 0 {
		return validation.ErrField
	}
	return nil
}
func validateServiceUnitSubset(m *ServiceUnit, _ validation.Mode) error { return extras(m.Extra) }
func validateSubscriptionIDSubset(m *SubscriptionID, _ validation.Mode) error {
	if e := cstring(m.SubscriptionIDData); e != nil {
		return at(444, e)
	}
	return extras(m.Extra)
}
func validateUserEquipmentInfoSubset(m *UserEquipmentInfo, _ validation.Mode) error {
	if m.UserEquipmentInfoType != 0 {
		return at(459, validation.ErrUnsupported)
	}
	if e := cstring(m.UserEquipmentInfoValue); e != nil {
		return at(460, e)
	}
	return extras(m.Extra)
}
func validateRequestMSCCSubset(m *RequestMSCC, _ validation.Mode) error {
	if m.RequestedServiceUnit.Present && m.RequestedServiceUnit.Value.CCServiceSpecificUnits != 1 {
		return at(437, at(417, validation.ErrUnsupported))
	}
	return extras(m.Extra)
}
func validateInterfaceSubset(m *Interface, _ validation.Mode) error { return extras(m.Extra) }
func validateSMSInformationSubset(m *SMSInformation, _ validation.Mode) error {
	if m.SMSNode != 0 {
		return vendorAt(2016, validation.ErrUnsupported)
	}
	if m.SMMessageType != 0 {
		return vendorAt(2007, validation.ErrUnsupported)
	}
	if m.OriginatorSCCPAddress.Present {
		if e := cstring(m.OriginatorSCCPAddress.Value); e != nil {
			return vendorAt(2008, e)
		}
	}
	for _, x := range []struct {
		c uint32
		b []byte
	}{{2018, m.ClientAddress}, {2010, m.RecipientSCCPAddress}, {2017, m.SMSCAddress}} {
		if e := cstring(x.b); e != nil {
			return vendorAt(x.c, e)
		}
	}
	if !bytes.Equal(m.ClientAddress, m.SMSCAddress) {
		return vendorAt(2018, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateAddressSubset(m *Address, _ validation.Mode) error {
	if m.AddressType != 1 {
		return vendorAt(899, validation.ErrUnsupported)
	}
	if e := cstring(m.AddressData); e != nil {
		return vendorAt(897, e)
	}
	return extras(m.Extra)
}
func validateMMSInformationSubset(m *MMSInformation, _ validation.Mode) error {
	if e := cstring(m.SMSKey); e != nil {
		return vendorAt(1210, e)
	}
	return extras(m.Extra)
}
func validateServiceInformationSubset(m *ServiceInformation, _ validation.Mode) error {
	return extras(m.Extra)
}
func validateCCRSubset(m *CCR, _ validation.Mode) error {
	if !selectedSequence(m.CCRequestType, m.CCRequestNumber) {
		return at(416, validation.ErrUnsupported)
	}
	sms := m.ServiceInformation.SMSInformation
	mscc := m.MultipleServicesCreditControl
	if m.CCRequestType == 4 {
		if !m.RequestedAction.Present || m.RequestedAction.Value != 0 {
			return at(436, validation.ErrUnsupported)
		}
		if !sms.OriginatorSCCPAddress.Present || sms.OriginatorInterface.InterfaceType != 1 {
			return vendorAt(2008, validation.ErrUnsupported)
		}
	} else {
		if m.RequestedAction.Present {
			return at(436, validation.ErrUnsupported)
		}
		if sms.OriginatorSCCPAddress.Present {
			return vendorAt(2008, validation.ErrUnsupported)
		}
		if sms.DestinationInterface.InterfaceType != 2 || sms.OriginatorInterface.InterfaceType != 3 {
			return vendorAt(2006, validation.ErrUnsupported)
		}
	}
	if m.CCRequestType == 3 {
		if mscc.RequestedServiceUnit.Present || !mscc.UsedServiceUnit.Present || !mscc.ReportingReason.Present || mscc.ReportingReason.Value != 2 {
			return at(456, validation.ErrUnsupported)
		}
		if !sms.SMStatus.Present {
			return vendorAt(2014, validation.ErrMissing)
		}
		if !sms.DischargeTime.Present {
			return vendorAt(2012, validation.ErrMissing)
		}
	} else {
		if !mscc.RequestedServiceUnit.Present || mscc.UsedServiceUnit.Present || mscc.ReportingReason.Present {
			return at(456, validation.ErrUnsupported)
		}
		if sms.SMStatus.Present || sms.DischargeTime.Present {
			return vendorAt(2000, validation.ErrUnsupported)
		}
	}

	if m.MultipleServicesIndicator != 1 {
		return at(455, validation.ErrUnsupported)
	}
	if len(m.SubscriptionID) != 2 || m.SubscriptionID[0].SubscriptionIDType != 0 || m.SubscriptionID[1].SubscriptionIDType != 1 {
		return at(443, validation.ErrUnsupported)
	}
	if !bytes.Equal(m.SubscriptionID[1].SubscriptionIDData, m.UserEquipmentInfo.UserEquipmentInfoValue) {
		return at(458, at(460, validation.ErrField))
	}
	if m.EventTimestamp != m.ServiceInformation.MMSInformation.SubmissionTime {
		return vendorAt(873, vendorAt(877, vendorAt(1202, validation.ErrField)))
	}
	return extras(m.Extra)
}
func validateCCASubset(m *CCA, mode validation.Mode) error {
	if mode != validation.Receive && (m.ExperimentalResult.Present || len(m.SupportedFeatures) != 0) {
		return validation.ErrUnsupported
	}
	if m.MultipleServicesCreditControl.Present && m.MultipleServicesCreditControl.Value.ResultCode == 5002 && m.CCRequestType != 3 {
		return at(456, at(268, validation.ErrUnsupported))
	}

	// Receive retains CC identity for explicit original-request correlation.
	if mode != validation.Receive && !selectedSequence(m.CCRequestType, m.CCRequestNumber) {
		return validation.ErrUnsupported
	}
	switch m.ResultCode {
	case 4002:
		// Received legacy wrapper evidence only; not a standard CCA send license.
		if mode != validation.Receive {
			return at(268, validation.ErrUnsupported)
		}

	case 2001, 4010, 4011, 4012, 5030:
	case 5002:
		if m.CCRequestType != 3 {
			return at(268, validation.ErrUnsupported)
		}
	case 5011:
	case 5031:
		if !m.FailedAVP.Present {
			return at(279, validation.ErrMissing)
		}
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		if e := base.ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode); e != nil {
			return e
		}
	default:
		return at(268, validation.ErrUnsupported)
	}
	if m.CCRequestType == 3 && m.GrantedServiceUnit.Present {
		return at(431, validation.ErrUnsupported)
	}
	if m.CCRequestType != 4 && m.DirectDebitingFailureHandling.Present {
		return at(428, validation.ErrUnsupported)
	}

	if m.DirectDebitingFailureHandling.Present && m.DirectDebitingFailureHandling.Value > 1 {
		return at(428, validation.ErrUnsupported)
	}
	if m.CostInformation.Present && !answerDetailsAllowed(m) {
		return at(423, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateCCARelation(r *CCR, a *CCA) error {
	if r.CCRequestType != a.CCRequestType {
		return at(416, validation.ErrInvalidAnswer)
	}
	if r.CCRequestNumber != a.CCRequestNumber {
		return at(415, validation.ErrInvalidAnswer)
	}
	return nil
}
func checkLegacySMS(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	if !ctx.Empty() {
		return validation.ErrContext
	}
	if _, ok := request.(*CCR); !ok {
		return validation.ErrContextRequired
	}
	if a, ok := answer.(*CCA); ok {
		return checkReceivedResult(a)
	}
	return nil
}

func selectedSequence(kind, number uint32) bool {
	return (kind == 1 || kind == 4) && number == 0 || kind == 3 && number == 1
}

func validateCostValueSubset(m *CostValue, _ validation.Mode) error { return extras(m.Extra) }
func validateCostInformationSubset(m *CostInformation, _ validation.Mode) error {
	return extras(m.Extra)
}

// One explicitly selected result-only MSCC. No service key, grant, token or
// default child success is inferred from its position or from the command code.
func validateServiceResultSubset(m *ServiceResult, _ validation.Mode) error {
	switch m.ResultCode {
	case 2001, 4010, 4011, 4012, 5002, 5012, 5030:
	default:
		return at(268, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}

// This predicate only restricts the selected financial field subset. It is NOT
// a verdict of charging success or an instruction to grant/deliver/retry.
func answerDetailsAllowed(a *CCA) bool {
	return a.ResultCode == 2001 && (!a.MultipleServicesCreditControl.Present || a.MultipleServicesCreditControl.Value.ResultCode == 2001)
}

// These groups are received evidence, not feature negotiation or a send license.
func validateReceivedExperimentalResult(m *ExperimentalResult, _ validation.Mode) error {
	return extras(m.Extra)
}
func validateReceivedSupportedFeatures(m *SupportedFeatures, _ validation.Mode) error {
	return extras(m.Extra)
}

func checkReceivedResult(a *CCA) error {
	if a.ResultCode == 4002 {
		return at(268, validation.ErrUnsupported)
	}
	if a.ExperimentalResult.Present {
		e := a.ExperimentalResult.Value
		if a.ResultCode == 2001 && e.VendorID == 10415 && e.ExperimentalResultCode == 5011 {
			return at(297, validation.ErrResultConflict)
		}
		return at(297, validation.ErrUnsupported)
	}
	if len(a.SupportedFeatures) != 0 {
		return validation.At(avp.Key{VendorID: 10415, Code: 628}, validation.ErrUnsupported)
	}
	if a.ResultCode == 5031 && a.FailedAVP.Value.Incomplete {
		return at(279, validation.ErrUnsupported)
	}
	return nil
}

// E cannot be used to manufacture the unresolved normal CCA experimental pair.
// Rating diagnostics remain mandatory even when the E grammar is selected.
func validateCreditError(a *pdu.ErrorAnswer, _ validation.Mode) error {
	if a.ExperimentalResult.Present {
		return at(297, validation.ErrUnsupported)
	}
	if a.ResultCode == 5031 && !a.FailedAVP.Present {
		return at(279, validation.ErrMissing)
	}
	return nil
}
