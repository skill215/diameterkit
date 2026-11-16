// Package ro implements bounded standard SMS IEC and no-UPDATE ECUR slices.
// It preserves charging units and opaque refund information; rating, delivery,
// charging history, refund decisions and recovery policy belong to the caller.
package ro

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

const (
	ApplicationID       uint32 = 4
	InitialRequest      uint32 = 1
	TerminationRequest  uint32 = 3
	EventRequest        uint32 = 4
	ReportingFinal      uint32 = 2
	DirectDebiting      uint32 = 0
	SMSServiceContextID        = "18.32274@3gpp.org"
)

// Known charging fields cannot bypass the selected parent/direction/subset by
// clearing M. Unrecognized optional extensions still follow the common policy.
func extras(v []avp.Raw) error {
	for _, x := range v {
		k := x.Key
		if k.VendorID == 0 && (k.Code >= 411 && k.Code <= 461 || k.Code == 258 || k.Code == 268 || k.Code == 277 || k.Code == 297 || k.Code == 417) || k.VendorID == 10415 && (k.Code >= 2000 && k.Code <= 2040 || k.Code == 21 || k.Code == 22 || k.Code == 23 || k.Code == 872 || k.Code == 873 || k.Code == 874 || k.Code == 3408 || k.Code == 3409 || k.Code == 628) {
			return validation.At(k, validation.ErrUnsupported)
		}
	}
	return nil
}
func at(code uint32, err error) error { return validation.At(avp.Key{Code: code}, err) }
func smsAt(code uint32, err error) error {
	return validation.At(avp.Key{VendorID: 10415, Code: code}, err)
}
func decimal(s string) bool {
	if len(s) == 0 || len(s) > 15 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func validateUnitsSubset(m *ServiceUnit, _ validation.Mode) error { return extras(m.Extra) }
func validateSubscriptionSubset(m *SubscriptionID, _ validation.Mode) error {
	if m.SubscriptionIDType != 0 {
		return at(450, validation.ErrUnsupported)
	}
	if !decimal(m.SubscriptionIDData) {
		return at(444, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateInterfaceSubset(m *Interface, _ validation.Mode) error {
	if m.InterfaceType.Present && m.InterfaceType.Value > 4 {
		return smsAt(2006, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateRecipientSubset(m *RecipientInfo, _ validation.Mode) error {
	if m.RecipientSCCPAddress.Present {
		a := m.RecipientSCCPAddress.Value
		if a.Family != 8 {
			return smsAt(2010, validation.ErrUnsupported)
		}
		if !decimal(string(a.Data)) {
			return smsAt(2010, validation.ErrField)
		}
	}
	return extras(m.Extra)
}
func validateSMSSubset(m *SMSInformation, _ validation.Mode) error {
	// These slices select the IP-SM-GW/Submission values of V-IEC/V-SMS/V-TERM.
	// Other nodes/message types are future subsets, not globally illegal values.
	if !m.SMSNode.Present || !m.SMMessageType.Present {
		return validation.ErrUnsupported
	}
	if m.SMSNode.Value != 1 || m.SMMessageType.Value != 0 {
		return validation.ErrUnsupported
	}
	return extras(m.Extra)
}
func validateServiceSubset(m *ServiceInformation, _ validation.Mode) error {
	if !m.SMSInformation.Present {
		return smsAt(2000, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func serviceKey(ids []uint32) error {
	if len(ids) != 1 {
		return at(439, validation.ErrUnsupported)
	}
	return nil
}
func validateRequestMSCCSubset(m *RequestMSCC, _ validation.Mode) error {
	if err := serviceKey(m.ServiceIdentifier); err != nil {
		return err
	}
	return extras(m.Extra)
}
func validateAnswerMSCCSubset(m *AnswerMSCC, _ validation.Mode) error {
	if err := serviceKey(m.ServiceIdentifier); err != nil {
		return err
	}
	if m.ResultCode.Present && m.ResultCode.Value != 2001 && m.ResultCode.Value != 4010 && m.ResultCode.Value != 4011 && m.ResultCode.Value != 4012 && m.ResultCode.Value != 5030 {
		return at(268, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateCCRSubset(m *CCR, _ validation.Mode) error {
	if m.ServiceContextID != SMSServiceContextID {
		return at(461, validation.ErrUnsupported)
	}
	if !selectedSequence(m.CCRequestType, m.CCRequestNumber) {
		return at(416, validation.ErrUnsupported)
	}
	if m.CCRequestType == EventRequest {
		if !m.RequestedAction.Present {
			return at(436, validation.ErrMissing)
		}
		if m.RequestedAction.Value != DirectDebiting {
			return at(436, validation.ErrUnsupported)
		}
	} else if m.RequestedAction.Present {
		return at(436, validation.ErrUnsupported)
	}
	if !m.MultipleServicesIndicator.Present || m.MultipleServicesIndicator.Value != 1 {
		return at(455, validation.ErrUnsupported)
	}
	if len(m.MultipleServicesCreditControl) != 1 {
		return at(456, validation.ErrUnsupported)
	}
	if !m.ServiceInformation.Present {
		return smsAt(873, validation.ErrUnsupported)
	}
	s := m.MultipleServicesCreditControl[0]
	if m.CCRequestType == TerminationRequest {
		if s.RequestedServiceUnit.Present {
			return at(456, at(437, validation.ErrUnsupported))
		}
		if len(s.UsedServiceUnit) != 1 {
			return at(456, at(446, validation.ErrUnsupported))
		}
		if !s.UsedServiceUnit[0].CCServiceSpecificUnits.Present {
			return at(456, at(446, at(417, validation.ErrMissing)))
		}
		if len(s.ReportingReason) != 1 || s.ReportingReason[0] != ReportingFinal {
			return at(456, smsAt(872, validation.ErrUnsupported))
		}
	} else if len(s.UsedServiceUnit) != 0 || len(s.ReportingReason) != 0 {
		return at(456, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateCCASubset(m *CCA, mode validation.Mode) error {
	if mode != validation.Receive && (m.ExperimentalResult.Present || len(m.SupportedFeatures) != 0) {
		return validation.ErrUnsupported
	}
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
		if m.CCRequestType != TerminationRequest {
			return at(268, validation.ErrUnsupported)
		}
	case 5011:
	case 5031:
		if !m.FailedAVP.Present {
			return at(279, validation.ErrMissing)
		}
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		if err := base.ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode); err != nil {
			return err
		}
	default:
		return at(268, validation.ErrUnsupported)
	}
	if len(m.MultipleServicesCreditControl) > 1 {
		return at(456, validation.ErrUnsupported)
	}
	if m.CCRequestType != TerminationRequest && m.ResultCode == 2001 && !m.ExperimentalResult.Present && len(m.MultipleServicesCreditControl) == 0 {
		return at(456, validation.ErrMissing)
	}
	for _, s := range m.MultipleServicesCreditControl {
		if m.CCRequestType == TerminationRequest && s.GrantedServiceUnit.Present {
			return at(456, at(431, validation.ErrUnsupported))
		}
		if mode != validation.Receive && m.CCRequestType != EventRequest && s.RefundInformation.Present {
			return at(456, smsAt(2022, validation.ErrUnsupported))
		}
		// Absence is retained on receive. This first slice does not infer a child
		// result or a grant from the command result when no child result is present.
		if mode != validation.Receive && !s.ResultCode.Present {
			return at(456, at(268, validation.ErrUnsupported))
		}
		if m.CCRequestType != TerminationRequest && m.ResultCode == 2001 && !m.ExperimentalResult.Present && s.ResultCode.Present && s.ResultCode.Value == 2001 && (!s.GrantedServiceUnit.Present || !s.GrantedServiceUnit.Value.CCServiceSpecificUnits.Present) {
			return at(456, at(431, validation.ErrMissing))
		}
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
	for _, s := range a.MultipleServicesCreditControl {
		if len(r.MultipleServicesCreditControl) != 1 || len(s.ServiceIdentifier) != 1 || len(r.MultipleServicesCreditControl[0].ServiceIdentifier) != 1 || s.ServiceIdentifier[0] != r.MultipleServicesCreditControl[0].ServiceIdentifier[0] {
			return at(456, at(439, validation.ErrInvalidAnswer))
		}
	}
	return nil
}
func checkSMSCreditControl(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	if !ctx.Empty() {
		return validation.ErrContext
	}
	if _, ok := request.(*CCR); !ok {
		return validation.ErrContextRequired
	}
	if a, ok := answer.(*CCA); ok {
		if err := checkReceivedResult(a); err != nil {
			return err
		}
		for _, s := range a.MultipleServicesCreditControl {
			if a.CCRequestType != EventRequest && s.RefundInformation.Present {
				return at(456, smsAt(2022, validation.ErrUnsupported))
			}
			if !s.ResultCode.Present {
				return at(456, at(268, validation.ErrUnsupported))
			}
		}
	}
	return nil
}

// This profile selects the no-interim ECUR path only, not a general sequence allocator.
func selectedSequence(kind, number uint32) bool {
	return (kind == InitialRequest || kind == EventRequest) && number == 0 || kind == TerminationRequest && number == 1
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
