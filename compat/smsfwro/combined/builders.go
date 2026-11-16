// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package combined implements the explicitly bound five-template sms-fw wire dialect.
// Its single decoder keeps ClientAddressWire opaque; builders preserve each source
// representation without guessing types. Compile separately from other App4 profiles.
package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
)

type IECInput = smsfwro.IECInput
type ECURInput = smsfwro.ECURInput
type TerminationInput = smsfwro.TerminationInput
type GeneralInput = legacy420.GeneralInput
type RefundInput = legacy420.RefundInput

// NewIEC builds the original one-unit IEC with unchanged legacy STRING bytes.
func NewIEC(v IECInput) *CCR { return fromSMSCCR(smsfwro.NewIEC(v)) }

// NewInitial builds the explicit no-UPDATE INITIAL; it grants no credit.
func NewInitial(v ECURInput) *CCR { return fromSMSCCR(smsfwro.NewInitial(v)) }

// NewTermination builds a new exchange; no state, clock or automatic send is used.
func NewTermination(v TerminationInput) *CCR { return fromSMSCCR(smsfwro.NewTermination(v)) }

// NewDirectDebit preserves the general template's fixed RFC Address wire bytes.
func NewDirectDebit(v GeneralInput) *CCR { return fromMoneyCCR(legacy420.NewDirectDebit(v)) }

// NewBalanceCheck creates an inquiry, not a credit reservation.
func NewBalanceCheck(v GeneralInput) *CCR { return fromMoneyCCR(legacy420.NewBalanceCheck(v)) }

// NewRefund requires an explicit token, including an explicitly empty one.
// Financial eligibility and approved OCS receipt mapping belong to the caller.
func NewRefund(v RefundInput) (*CCR, error) {
	r, e := legacy420.NewRefund(v)
	if e != nil {
		return nil, e
	}
	return fromMoneyCCR(r), nil
}

func fromSMSSubscriptionIDToSubscriptionID(m *smsfwro.SubscriptionID) *SubscriptionID {
	c := &SubscriptionID{}
	c.SubscriptionIDType = m.SubscriptionIDType
	c.SubscriptionIDData = m.SubscriptionIDData
	return c
}
func fromSMSServiceUnitToRequestedUnit(m *smsfwro.ServiceUnit) *RequestedUnit {
	c := &RequestedUnit{}
	c.CCServiceSpecificUnits = m.CCServiceSpecificUnits
	return c
}
func fromSMSServiceUnitToServiceUnit(m *smsfwro.ServiceUnit) *ServiceUnit {
	c := &ServiceUnit{}
	c.CCServiceSpecificUnits = m.CCServiceSpecificUnits
	return c
}
func fromSMSRequestMSCCToRequestMSCC(m *smsfwro.RequestMSCC) *RequestMSCC {
	c := &RequestMSCC{}
	if m.RequestedServiceUnit.Present {
		c.RequestedServiceUnit = avp.Some(*fromSMSServiceUnitToRequestedUnit(&m.RequestedServiceUnit.Value))
	}
	if m.UsedServiceUnit.Present {
		c.UsedServiceUnit = avp.Some(*fromSMSServiceUnitToServiceUnit(&m.UsedServiceUnit.Value))
	}
	if m.ReportingReason.Present {
		c.ReportingReason = avp.Some(m.ReportingReason.Value)
	}
	return c
}
func fromSMSUserEquipmentInfoToUserEquipmentInfo(m *smsfwro.UserEquipmentInfo) *UserEquipmentInfo {
	c := &UserEquipmentInfo{}
	c.UserEquipmentInfoType = m.UserEquipmentInfoType
	c.UserEquipmentInfoValue = m.UserEquipmentInfoValue
	return c
}
func fromSMSInterfaceToInterface(m *smsfwro.Interface) *Interface {
	c := &Interface{}
	c.InterfaceType = m.InterfaceType
	return c
}
func fromSMSSMSInformationToSMSInformation(m *smsfwro.SMSInformation) *SMSInformation {
	c := &SMSInformation{}
	c.SMSNode = m.SMSNode
	c.ClientAddressWire = avp.Some(m.ClientAddress)
	if m.OriginatorSCCPAddress.Present {
		c.OriginatorSCCPAddress = avp.Some(m.OriginatorSCCPAddress.Value)
	}
	c.RecipientSCCPAddress = avp.Some(m.RecipientSCCPAddress)
	c.SMSCAddress = avp.Some(m.SMSCAddress)
	c.SMDataCodingScheme = m.SMDataCodingScheme
	c.DestinationInterface = *fromSMSInterfaceToInterface(&m.DestinationInterface)
	c.SMMessageType = m.SMMessageType
	c.OriginatorInterface = *fromSMSInterfaceToInterface(&m.OriginatorInterface)
	c.SMProtocolID = m.SMProtocolID
	c.ReplyPathRequested = avp.Some(m.ReplyPathRequested)
	c.SMUserDataHeader = avp.Some(m.SMUserDataHeader)
	if m.SMStatus.Present {
		c.SMStatus = avp.Some(m.SMStatus.Value)
	}
	if m.DischargeTime.Present {
		c.DischargeTime = avp.Some(m.DischargeTime.Value)
	}
	return c
}
func fromSMSAddressToAddress(m *smsfwro.Address) *Address {
	c := &Address{}
	c.AddressType = m.AddressType
	c.AddressData = m.AddressData
	return c
}
func fromSMSMMSInformationToMMSInformation(m *smsfwro.MMSInformation) *MMSInformation {
	c := &MMSInformation{}
	c.OriginatorAddress = *fromSMSAddressToAddress(&m.OriginatorAddress)
	c.RecipientAddress = avp.Some(*fromSMSAddressToAddress(&m.RecipientAddress))
	c.SubmissionTime = avp.Some(m.SubmissionTime)
	c.Priority = avp.Some(m.Priority)
	c.UserDataLength = avp.Some(m.UserDataLength)
	c.SMSKey = avp.Some(m.SMSKey)
	c.StatusReportRequest = avp.Some(m.StatusReportRequest)
	return c
}
func fromSMSServiceInformationToServiceInformation(m *smsfwro.ServiceInformation) *ServiceInformation {
	c := &ServiceInformation{}
	c.SMSInformation = *fromSMSSMSInformationToSMSInformation(&m.SMSInformation)
	c.MMSInformation = avp.Some(*fromSMSMMSInformationToMMSInformation(&m.MMSInformation))
	return c
}
func fromSMSCCRToCCR(m *smsfwro.CCR) *CCR {
	c := &CCR{}
	c.SessionID = m.SessionID
	c.OriginHost = m.OriginHost
	c.OriginRealm = m.OriginRealm
	c.DestinationRealm = m.DestinationRealm
	c.AuthApplicationID = m.AuthApplicationID
	c.ServiceContextID = m.ServiceContextID
	c.CCRequestType = m.CCRequestType
	c.CCRequestNumber = m.CCRequestNumber
	if m.RequestedAction.Present {
		c.RequestedAction = avp.Some(m.RequestedAction.Value)
	}
	if m.DestinationHost.Present {
		c.DestinationHost = avp.Some(m.DestinationHost.Value)
	}
	if m.OriginStateID.Present {
		c.OriginStateID = avp.Some(m.OriginStateID.Value)
	}
	for _, v := range m.SubscriptionID {
		c.SubscriptionID = append(c.SubscriptionID, *fromSMSSubscriptionIDToSubscriptionID(&v))
	}
	c.EventTimestamp = m.EventTimestamp
	c.MultipleServicesIndicator = avp.Some(m.MultipleServicesIndicator)
	c.MultipleServicesCreditControl = *fromSMSRequestMSCCToRequestMSCC(&m.MultipleServicesCreditControl)
	c.UserEquipmentInfo = avp.Some(*fromSMSUserEquipmentInfoToUserEquipmentInfo(&m.UserEquipmentInfo))
	c.ServiceInformation = *fromSMSServiceInformationToServiceInformation(&m.ServiceInformation)
	for _, v := range m.ProxyInfo {
		c.ProxyInfo = append(c.ProxyInfo, v)
	}
	for _, v := range m.RouteRecord {
		c.RouteRecord = append(c.RouteRecord, v)
	}
	return c
}
func fromSMSCCR(m *smsfwro.CCR) *CCR { return fromSMSCCRToCCR(m) }
func fromMoneySubscriptionIDToSubscriptionID(m *legacy420.SubscriptionID) *SubscriptionID {
	c := &SubscriptionID{}
	c.SubscriptionIDType = m.SubscriptionIDType
	c.SubscriptionIDData = m.SubscriptionIDData
	return c
}
func fromMoneyUnitValueToUnitValue(m *legacy420.UnitValue) *UnitValue {
	c := &UnitValue{}
	c.ValueDigits = m.ValueDigits
	c.Exponent = m.Exponent
	return c
}
func fromMoneyMoneyToMoney(m *legacy420.Money) *Money {
	c := &Money{}
	c.UnitValue = *fromMoneyUnitValueToUnitValue(&m.UnitValue)
	c.CurrencyCode = m.CurrencyCode
	return c
}
func fromMoneyRequestedUnitToRequestedUnit(m *legacy420.RequestedUnit) *RequestedUnit {
	c := &RequestedUnit{}
	if m.Money.Present {
		c.Money = avp.Some(*fromMoneyMoneyToMoney(&m.Money.Value))
	}
	if m.StandardMoney.Present {
		c.StandardMoney = avp.Some(*fromMoneyMoneyToMoney(&m.StandardMoney.Value))
	}
	c.CCServiceSpecificUnits = m.CCServiceSpecificUnits
	return c
}
func fromMoneyRequestMSCCToRequestMSCC(m *legacy420.RequestMSCC) *RequestMSCC {
	c := &RequestMSCC{}
	c.RequestedServiceUnit = avp.Some(*fromMoneyRequestedUnitToRequestedUnit(&m.RequestedServiceUnit))
	return c
}
func fromMoneyInterfaceToInterface(m *legacy420.Interface) *Interface {
	c := &Interface{}
	c.InterfaceType = m.InterfaceType
	return c
}
func fromMoneySMSInformationToSMSInformation(m *legacy420.SMSInformation) *SMSInformation {
	c := &SMSInformation{}
	c.SMSNode = m.SMSNode
	c.SMDataCodingScheme = m.SMDataCodingScheme
	c.DestinationInterface = *fromMoneyInterfaceToInterface(&m.DestinationInterface)
	c.SMMessageType = m.SMMessageType
	c.OriginatorInterface = *fromMoneyInterfaceToInterface(&m.OriginatorInterface)
	c.SMProtocolID = m.SMProtocolID
	if m.ClientAddress.Present {
		c.ClientAddressWire = avp.Some([]byte{0, 1, 127, 0, 0, 1})
	}
	return c
}
func fromMoneyOriginatorAddressToAddress(m *legacy420.OriginatorAddress) *Address {
	c := &Address{}
	c.AddressType = m.AddressType
	c.AddressData = m.AddressData
	return c
}
func fromMoneyMMSInformationToMMSInformation(m *legacy420.MMSInformation) *MMSInformation {
	c := &MMSInformation{}
	c.OriginatorAddress = *fromMoneyOriginatorAddressToAddress(&m.OriginatorAddress)
	return c
}
func fromMoneyServiceInformationToServiceInformation(m *legacy420.ServiceInformation) *ServiceInformation {
	c := &ServiceInformation{}
	c.SMSInformation = *fromMoneySMSInformationToSMSInformation(&m.SMSInformation)
	if m.MMSInformation.Present {
		c.MMSInformation = avp.Some(*fromMoneyMMSInformationToMMSInformation(&m.MMSInformation.Value))
	}
	return c
}
func fromMoneyCCRToCCR(m *legacy420.CCR) *CCR {
	c := &CCR{}
	c.SessionID = m.SessionID
	c.OriginHost = m.OriginHost
	c.OriginRealm = m.OriginRealm
	c.DestinationRealm = m.DestinationRealm
	c.AuthApplicationID = m.AuthApplicationID
	c.ServiceContextID = m.ServiceContextID
	c.CCRequestType = m.CCRequestType
	c.CCRequestNumber = m.CCRequestNumber
	c.RequestedAction = avp.Some(m.RequestedAction)
	if m.DestinationHost.Present {
		c.DestinationHost = avp.Some(m.DestinationHost.Value)
	}
	if m.OriginStateID.Present {
		c.OriginStateID = avp.Some(m.OriginStateID.Value)
	}
	for _, v := range m.SubscriptionID {
		c.SubscriptionID = append(c.SubscriptionID, *fromMoneySubscriptionIDToSubscriptionID(&v))
	}
	c.EventTimestamp = m.EventTimestamp
	if m.ServiceIdentifier.Present {
		c.ServiceIdentifier = avp.Some(m.ServiceIdentifier.Value)
	}
	if m.Timezone.Present {
		c.Timezone = avp.Some(m.Timezone.Value)
	}
	c.MultipleServicesCreditControl = *fromMoneyRequestMSCCToRequestMSCC(&m.MultipleServicesCreditControl)
	c.ServiceInformation = *fromMoneyServiceInformationToServiceInformation(&m.ServiceInformation)
	if m.RefundInformation.Present {
		c.RefundInformation = avp.Some(m.RefundInformation.Value)
	}
	for _, v := range m.ProxyInfo {
		c.ProxyInfo = append(c.ProxyInfo, v)
	}
	for _, v := range m.RouteRecord {
		c.RouteRecord = append(c.RouteRecord, v)
	}
	return c
}
func fromMoneyCCR(m *legacy420.CCR) *CCR { return fromMoneyCCRToCCR(m) }
