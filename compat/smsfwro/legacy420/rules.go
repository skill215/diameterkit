// Package legacy420 implements the named sms-fw general/refund dialect. Profile fixes
// CCR/456/437/420 as Grouped money; standard CC-Time at that path is unsupported.
// Use a dedicated Bundle and an explicitly approved OCS deployment. This package
// does not read templates, store credentials, determine eligibility or retry.
package legacy420

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func extras(xs []avp.Raw) error {
	if len(xs) > 0 {
		return validation.At(xs[0].Key, validation.ErrUnsupported)
	}
	return nil
}
func at(c uint32, e error) error { return validation.At(avp.Key{Code: c}, e) }
func validateSubscriptionIDSubset(m *SubscriptionID, _ validation.Mode) error {
	if bytes.IndexByte(m.SubscriptionIDData, 0) >= 0 {
		return at(444, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateInterfaceSubset(m *Interface, _ validation.Mode) error { return extras(m.Extra) }
func validateSMSInformationSubset(m *SMSInformation, _ validation.Mode) error {
	if m.ClientAddress.Present && (m.ClientAddress.Value.Family != 1 || !bytes.Equal(m.ClientAddress.Value.Data, []byte{127, 0, 0, 1})) {
		return validation.At(avp.Key{VendorID: 10415, Code: 2018}, validation.ErrUnsupported)
	}
	if m.SMSNode != 2 {
		return validation.At(avp.Key{VendorID: 10415, Code: 2016}, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateServiceInformationSubset(m *ServiceInformation, _ validation.Mode) error {
	return extras(m.Extra)
}
func validateUnitValueSubset(m *UnitValue, _ validation.Mode) error { return extras(m.Extra) }
func validateMoneySubset(m *Money, _ validation.Mode) error         { return extras(m.Extra) }
func validateRequestedUnitSubset(m *RequestedUnit, _ validation.Mode) error {
	if m.Money.Present == m.StandardMoney.Present {
		return at(437, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateRequestMSCCSubset(m *RequestMSCC, _ validation.Mode) error { return extras(m.Extra) }
func validateCCRSubset(m *CCR, _ validation.Mode) error {
	if m.CCRequestType != 4 {
		return at(416, validation.ErrUnsupported)
	}
	if m.CCRequestNumber != 0 {
		return at(415, validation.ErrUnsupported)
	}
	if m.RequestedAction != 0 && m.RequestedAction != 1 && m.RequestedAction != 2 {
		return at(436, validation.ErrUnsupported)
	}
	if len(m.SubscriptionID) != 2 || m.SubscriptionID[0].SubscriptionIDType != 0 || m.SubscriptionID[1].SubscriptionIDType != 1 {
		return at(443, validation.ErrUnsupported)
	}
	u := m.MultipleServicesCreditControl.RequestedServiceUnit
	if m.RequestedAction == 1 {
		if !u.Money.Present || !m.RefundInformation.Present || m.ServiceIdentifier.Present || m.Timezone.Present || m.ServiceInformation.SMSInformation.ClientAddress.Present || m.ServiceInformation.MMSInformation.Present {
			return validation.ErrField
		}
	} else {
		if !u.StandardMoney.Present || m.RefundInformation.Present || !m.ServiceIdentifier.Present || m.ServiceIdentifier.Value != 1 || !m.Timezone.Present || m.Timezone.Value != 0 || !m.ServiceInformation.SMSInformation.ClientAddress.Present || !m.ServiceInformation.MMSInformation.Present {
			return validation.ErrField
		}
	}
	return extras(m.Extra)
}
func validateCCASubset(m *CCA, mode validation.Mode) error {
	if mode != validation.Receive && (m.ExperimentalResult.Present || len(m.SupportedFeatures) != 0) {
		return validation.ErrUnsupported
	}
	if m.DirectDebitingFailureHandling.Present && m.DirectDebitingFailureHandling.Value > 1 {
		return at(428, validation.ErrUnsupported)
	}
	if m.MultipleServicesCreditControl.Present && m.MultipleServicesCreditControl.Value.ResultCode == 5002 && m.CCRequestType != 3 {
		return at(456, at(268, validation.ErrUnsupported))
	}

	if mode != validation.Receive && (m.CCRequestType != 4 || m.CCRequestNumber != 0) {
		return validation.ErrUnsupported
	}
	if m.CheckBalanceResult.Present && m.CheckBalanceResult.Value > 1 {
		return at(422, validation.ErrUnsupported)
	}
	if !answerDetailsAllowed(m) && (m.GrantedServiceUnit.Present || m.CostInformation.Present || m.CheckBalanceResult.Present) {
		return validation.ErrUnsupported
	}
	switch m.ResultCode {
	case 4002:
		// Received legacy wrapper evidence only; not a standard CCA send license.
		if mode != validation.Receive {
			return at(268, validation.ErrUnsupported)
		}

	case 2001, 4010, 4011, 4012, 5030:
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
	return extras(m.Extra)
}
func validateCCARelation(r *CCR, a *CCA) error {
	if r.CCRequestType != a.CCRequestType {
		return at(416, validation.ErrInvalidAnswer)
	}
	if r.CCRequestNumber != a.CCRequestNumber {
		return at(415, validation.ErrInvalidAnswer)
	}
	if a.DirectDebitingFailureHandling.Present && r.RequestedAction != 0 {
		return at(428, validation.ErrInvalidAnswer)
	}
	if r.RequestedAction == 2 {
		if a.GrantedServiceUnit.Present || (answerDetailsAllowed(a) && !a.ExperimentalResult.Present && !a.CheckBalanceResult.Present) {
			return validation.ErrInvalidAnswer
		}
	} else if a.CheckBalanceResult.Present {
		return at(422, validation.ErrInvalidAnswer)
	}
	return nil
}
func checkLegacyEvent(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
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

func validateOriginatorAddressSubset(m *OriginatorAddress, _ validation.Mode) error {
	if m.AddressType != 1 || bytes.IndexByte(m.AddressData, 0) >= 0 {
		return validation.ErrField
	}
	return extras(m.Extra)
}
func validateMMSInformationSubset(m *MMSInformation, _ validation.Mode) error { return extras(m.Extra) }
func validateCostValueSubset(m *CostValue, _ validation.Mode) error           { return extras(m.Extra) }
func validateCostInformationSubset(m *CostInformation, _ validation.Mode) error {
	return extras(m.Extra)
}
func validateGrantedUnitSubset(m *GrantedUnit, _ validation.Mode) error { return extras(m.Extra) }

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
