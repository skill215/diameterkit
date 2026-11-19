// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package combined

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func at(c uint32, e error) error       { return validation.At(avp.Key{Code: c}, e) }
func vendorAt(c uint32, e error) error { return validation.At(avp.Key{VendorID: 10415, Code: c}, e) }
func extras(v []avp.Raw) error {
	if len(v) > 0 {
		return validation.At(v[0].Key, validation.ErrUnsupported)
	}
	return nil
}
func cstring(v []byte) error {
	if bytes.IndexByte(v, 0) >= 0 {
		return validation.ErrField
	}
	return nil
}
func validateSubscriptionIDSubset(m *SubscriptionID, _ validation.Mode) error {
	if e := cstring(m.SubscriptionIDData); e != nil {
		return at(444, e)
	}
	return extras(m.Extra)
}
func validateUserEquipmentInfoSubset(m *UserEquipmentInfo, mode validation.Mode) error {
	if m.UserEquipmentInfoType != 0 {
		if mode == validation.Receive && cstring(m.UserEquipmentInfoValue) == nil && len(m.Extra) == 0 {
			return at(459, equipmentTypeValue{})
		}
		return at(459, validation.ErrUnsupported)
	}
	if e := cstring(m.UserEquipmentInfoValue); e != nil {
		return at(460, e)
	}
	return extras(m.Extra)
}
func validateServiceUnitSubset(m *ServiceUnit, _ validation.Mode) error { return extras(m.Extra) }
func validateRequestMSCCSubset(m *RequestMSCC, _ validation.Mode) error { return extras(m.Extra) }
func validateInterfaceSubset(m *Interface, _ validation.Mode) error     { return extras(m.Extra) }
func validateSMSInformationSubset(m *SMSInformation, _ validation.Mode) error {
	// 2018 always remains wire octets. The complete request rule checks its exact
	// template relation; neither this decoder nor the builder parses its content.
	for _, v := range []avp.Optional[[]byte]{m.OriginatorSCCPAddress, m.RecipientSCCPAddress, m.SMSCAddress} {
		if v.Present {
			if e := cstring(v.Value); e != nil {
				return e
			}
		}
	}
	return extras(m.Extra)
}
func validateAddressSubset(m *Address, mode validation.Mode) error {
	if m.AddressType != 1 {
		if mode == validation.Receive && cstring(m.AddressData) == nil && len(m.Extra) == 0 {
			return vendorAt(899, addressTypeValue{})
		}
		return vendorAt(899, validation.ErrUnsupported)
	}
	if e := cstring(m.AddressData); e != nil {
		return vendorAt(897, e)
	}
	return extras(m.Extra)
}
func validateMMSInformationSubset(m *MMSInformation, _ validation.Mode) error {
	if m.SMSKey.Present {
		if e := cstring(m.SMSKey.Value); e != nil {
			return vendorAt(1210, e)
		}
	}
	return extras(m.Extra)
}
func validateServiceInformationSubset(m *ServiceInformation, _ validation.Mode) error {
	return extras(m.Extra)
}
func validateUnitValueSubset(m *UnitValue, _ validation.Mode) error { return extras(m.Extra) }
func validateMoneySubset(m *Money, _ validation.Mode) error         { return extras(m.Extra) }
func validateRequestedUnitSubset(m *RequestedUnit, _ validation.Mode) error {
	if m.Money.Present && m.StandardMoney.Present {
		return at(437, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateCostValueSubset(m *CostValue, _ validation.Mode) error { return extras(m.Extra) }
func validateCostInformationSubset(m *CostInformation, _ validation.Mode) error {
	return extras(m.Extra)
}
func selectedSequence(kind, number uint32) bool {
	return (kind == 1 || kind == 4) && number == 0 || kind == 3 && number == 1
}
func validateCCRSubset(m *CCR, mode validation.Mode) error {
	if !selectedSequence(m.CCRequestType, m.CCRequestNumber) {
		return at(416, validation.ErrUnsupported)
	}
	if len(m.SubscriptionID) != 2 || m.SubscriptionID[0].SubscriptionIDType != 0 || m.SubscriptionID[1].SubscriptionIDType != 1 {
		if mode == validation.Receive {
			if err := diagnoseCCRSubscriptions(m); err != nil {
				return err
			}
		}
		return at(443, validation.ErrUnsupported)
	}
	var err error
	// The structurally required template markers select constraints only. Field
	// types and decoder operations are fixed for every branch of this profile.
	if m.MultipleServicesIndicator.Present || m.UserEquipmentInfo.Present || m.CCRequestType != 4 {
		err = validateSMSRequest(m, mode)
	} else {
		err = validateMoneyRequest(m, mode)
	}
	if err != nil {
		return err
	}
	return extras(m.Extra)
}
func validateSMSRequest(m *CCR, mode validation.Mode) error {
	if !m.MultipleServicesIndicator.Present || m.MultipleServicesIndicator.Value != 1 || !m.UserEquipmentInfo.Present || m.ServiceIdentifier.Present || m.Timezone.Present || m.RefundInformation.Present {
		if mode == validation.Receive && m.MultipleServicesIndicator.Present && m.UserEquipmentInfo.Present && !m.ServiceIdentifier.Present && !m.Timezone.Present && !m.RefundInformation.Present {
			return at(455, indicatorValue{})
		}
		return validation.ErrField
	}
	if !bytes.Equal(m.SubscriptionID[1].SubscriptionIDData, m.UserEquipmentInfo.Value.UserEquipmentInfoValue) {
		return at(458, at(460, validation.ErrField))
	}
	s := m.ServiceInformation.SMSInformation
	u := m.MultipleServicesCreditControl
	if s.SMSNode != 0 || s.SMMessageType != 0 {
		if mode == validation.Receive {
			if s.SMSNode != 0 && s.SMMessageType == 0 {
				return invalidSMSValue(true)
			}
			if s.SMMessageType != 0 && s.SMSNode == 0 {
				return invalidSMSValue(false)
			}
		}
		return vendorAt(2000, validation.ErrUnsupported)
	}
	if !s.ClientAddressWire.Present || !s.RecipientSCCPAddress.Present || !s.SMSCAddress.Present || !s.ReplyPathRequested.Present || !s.SMUserDataHeader.Present {
		if mode == validation.Receive {
			fields := [...]smsOmissionField{{s.ClientAddressWire.Present, 2018}, {s.RecipientSCCPAddress.Present, 2010}, {s.SMSCAddress.Present, 2017}, {s.ReplyPathRequested.Present, 2011}, {s.SMUserDataHeader.Present, 2015}}
			if err := missingSMSCondition(fields[:]); err != nil {
				return err
			}
		}

		return vendorAt(2000, validation.ErrMissing)
	}
	if e := cstring(s.ClientAddressWire.Value); e != nil {
		return vendorAt(2018, e)
	}
	if !bytes.Equal(s.ClientAddressWire.Value, s.SMSCAddress.Value) {
		return vendorAt(2018, validation.ErrField)
	}
	if !m.ServiceInformation.MMSInformation.Present {
		if mode == validation.Receive {
			return missingSMSMMSGroup()
		}
		return vendorAt(877, validation.ErrMissing)
	}
	mm := m.ServiceInformation.MMSInformation.Value
	if !mm.RecipientAddress.Present || !mm.SubmissionTime.Present || !mm.Priority.Present || !mm.UserDataLength.Present || !mm.SMSKey.Present || !mm.StatusReportRequest.Present {
		if mode == validation.Receive {
			if err := missingSMSMMSField(&mm); err != nil {
				return err
			}
		}
		return vendorAt(877, validation.ErrMissing)
	}
	if m.EventTimestamp != mm.SubmissionTime.Value {
		return vendorAt(1202, validation.ErrField)
	}
	if m.CCRequestType == 4 {
		if !m.RequestedAction.Present || m.RequestedAction.Value != 0 {
			if mode == validation.Receive && !m.RequestedAction.Present {
				return missingRequestedAction()
			}
			if mode == validation.Receive && m.RequestedAction.Present {
				return invalidRequestedAction("combined-iec-event-v1")
			}
			return at(436, validation.ErrUnsupported)
		}
		if !s.OriginatorSCCPAddress.Present || s.OriginatorInterface.InterfaceType != 1 {
			if mode == validation.Receive && !s.OriginatorSCCPAddress.Present && s.OriginatorInterface.InterfaceType == 1 {
				return missingIECOriginator()
			}
			if mode == validation.Receive && s.OriginatorSCCPAddress.Present && s.OriginatorInterface.InterfaceType != 1 {
				return invalidIECInterface()
			}
			return vendorAt(2008, validation.ErrUnsupported)
		}
	} else {
		if m.RequestedAction.Present {
			return at(436, validation.ErrUnsupported)
		}
		if s.OriginatorSCCPAddress.Present || s.DestinationInterface.InterfaceType != 2 || s.OriginatorInterface.InterfaceType != 3 {
			if mode == validation.Receive && !s.OriginatorSCCPAddress.Present {
				if s.DestinationInterface.InterfaceType == 2 && s.OriginatorInterface.InterfaceType != 3 {
					return invalidECURInterface(true)
				}
				if s.OriginatorInterface.InterfaceType == 3 && s.DestinationInterface.InterfaceType != 2 {
					return invalidECURInterface(false)
				}
			}
			return vendorAt(2008, validation.ErrUnsupported)
		}
	}
	if m.CCRequestType == 3 {
		if u.RequestedServiceUnit.Present || !u.UsedServiceUnit.Present || !u.ReportingReason.Present || u.ReportingReason.Value != 2 {
			if mode == validation.Receive && !u.RequestedServiceUnit.Present && u.UsedServiceUnit.Present && u.ReportingReason.Present {
				return invalidUnitValue(true)
			}
			return at(456, validation.ErrUnsupported)
		}
		if !s.SMStatus.Present || !s.DischargeTime.Present {
			if mode == validation.Receive {
				fields := [...]smsOmissionField{{s.SMStatus.Present, 2014}, {s.DischargeTime.Present, 2012}}
				if err := missingSMSCondition(fields[:]); err != nil {
					return err
				}
			}

			return vendorAt(2000, validation.ErrMissing)
		}
	} else {
		if !u.RequestedServiceUnit.Present || u.UsedServiceUnit.Present || u.ReportingReason.Present {
			return at(456, validation.ErrUnsupported)
		}
		r := u.RequestedServiceUnit.Value
		if r.Money.Present || r.StandardMoney.Present || r.CCServiceSpecificUnits != 1 {
			if mode == validation.Receive && !r.Money.Present && !r.StandardMoney.Present {
				return invalidUnitValue(false)
			}
			return at(437, validation.ErrUnsupported)
		}
		if s.SMStatus.Present || s.DischargeTime.Present {
			return vendorAt(2000, validation.ErrUnsupported)
		}
	}
	return nil
}
func validateMoneyRequest(m *CCR, mode validation.Mode) error {
	if !m.RequestedAction.Present || m.RequestedAction.Value > 2 {
		if mode == validation.Receive && !m.RequestedAction.Present {
			return missingRequestedAction()
		}
		if mode == validation.Receive && m.RequestedAction.Present {
			return invalidRequestedAction("combined-money-event-v1")
		}
		return at(436, validation.ErrUnsupported)
	}
	s := m.ServiceInformation.SMSInformation
	u := m.MultipleServicesCreditControl
	if !u.RequestedServiceUnit.Present || u.UsedServiceUnit.Present || u.ReportingReason.Present {
		return at(456, validation.ErrUnsupported)
	}
	if s.SMSNode != 2 || s.OriginatorSCCPAddress.Present || s.RecipientSCCPAddress.Present || s.SMSCAddress.Present || s.ReplyPathRequested.Present || s.SMUserDataHeader.Present || s.SMStatus.Present || s.DischargeTime.Present {
		if mode == validation.Receive && s.SMSNode != 2 && !s.OriginatorSCCPAddress.Present && !s.RecipientSCCPAddress.Present && !s.SMSCAddress.Present && !s.ReplyPathRequested.Present && !s.SMUserDataHeader.Present && !s.SMStatus.Present && !s.DischargeTime.Present {
			return invalidMoneyNode()
		}
		return vendorAt(2000, validation.ErrUnsupported)
	}
	r := u.RequestedServiceUnit.Value
	if m.RequestedAction.Value == 1 {
		if !r.Money.Present || r.StandardMoney.Present || !m.RefundInformation.Present || m.ServiceIdentifier.Present || m.Timezone.Present || s.ClientAddressWire.Present || m.ServiceInformation.MMSInformation.Present {
			return validation.ErrField
		}
	} else {
		if r.Money.Present || !r.StandardMoney.Present || m.RefundInformation.Present || !m.ServiceIdentifier.Present || m.ServiceIdentifier.Value != 1 || !m.Timezone.Present || m.Timezone.Value != 0 || !s.ClientAddressWire.Present || !m.ServiceInformation.MMSInformation.Present {
			return validation.ErrField
		}
		if !bytes.Equal(s.ClientAddressWire.Value, []byte{0, 1, 127, 0, 0, 1}) {
			return vendorAt(2018, validation.ErrUnsupported)
		}
		mm := m.ServiceInformation.MMSInformation.Value
		if mm.RecipientAddress.Present || mm.SubmissionTime.Present || mm.Priority.Present || mm.UserDataLength.Present || mm.SMSKey.Present || mm.StatusReportRequest.Present {
			return vendorAt(877, validation.ErrUnsupported)
		}
	}
	return nil
}
func validateCCASubset(m *CCA, mode validation.Mode) error {
	if mode != validation.Receive && (m.ExperimentalResult.Present || len(m.SupportedFeatures) != 0) {
		return validation.ErrUnsupported
	}
	if m.MultipleServicesCreditControl.Present && m.MultipleServicesCreditControl.Value.ResultCode == 5002 && m.CCRequestType != 3 {
		return at(456, at(268, validation.ErrUnsupported))
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
	if m.DirectDebitingFailureHandling.Present && (m.CCRequestType != 4 || m.DirectDebitingFailureHandling.Value > 1) {
		return at(428, validation.ErrUnsupported)
	}
	if m.CheckBalanceResult.Present && m.CheckBalanceResult.Value > 1 {
		return at(422, validation.ErrUnsupported)
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
	if r.MultipleServicesIndicator.Present {
		// Cost is now explicitly selected for this family on clean2001 only.
		// Balance remains tied to the original balance request.
		if a.CheckBalanceResult.Present || a.CostInformation.Present && !answerDetailsAllowed(a) {
			return validation.ErrInvalidAnswer
		}
	} else {
		if a.DirectDebitingFailureHandling.Present && r.RequestedAction.Value != 0 {
			return at(428, validation.ErrInvalidAnswer)
		}
		if !answerDetailsAllowed(a) && (a.GrantedServiceUnit.Present || a.CostInformation.Present || a.CheckBalanceResult.Present) {
			return validation.ErrInvalidAnswer
		}
		if r.RequestedAction.Value == 2 {
			if a.GrantedServiceUnit.Present || answerDetailsAllowed(a) && !a.ExperimentalResult.Present && !a.CheckBalanceResult.Present {
				return validation.ErrInvalidAnswer
			}
		} else if a.CheckBalanceResult.Present {
			return at(422, validation.ErrInvalidAnswer)
		}
	}
	return nil
}
func checkCombinedSMS(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
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
