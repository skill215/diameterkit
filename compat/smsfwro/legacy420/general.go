// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package legacy420

import "github.com/skill215/diameterkit/avp"

// GeneralInput supplies values for the actual general sms-fw template. Amounts
// and units are explicit, including zero; no sample financial defaults are used.
// Timestamp is already Diameter wire time. OriginatorMSISDN follows the legacy
// STRING writer: bytes without NUL, not a number normalization or UTF-8 service.
type GeneralInput struct {
	SessionID, ServiceContextID                                              string
	SubscriptionMSISDN, SubscriptionIMSI                                     string
	Timestamp                                                                avp.Time
	ValueDigits                                                              int64
	Exponent, CurrencyCode                                                   int32
	ServiceSpecificUnits                                                     uint64
	DataCodingScheme, DestinationInterface, MessageType, OriginatorInterface int32
	ProtocolID                                                               []byte
	OriginatorMSISDN                                                         string
}

// NewDirectDebit constructs EVENT4/0/Action0 from the general template, with
// ordinary RFC Address encoding. It sends nothing and allocates no identity.
func NewDirectDebit(v GeneralInput) *CCR { return newGeneral(v, 0) }

// NewBalanceCheck constructs EVENT4/0/Action2. A successful balance answer is
// evidence only for this inquiry, never a reservation or permission to debit.
func NewBalanceCheck(v GeneralInput) *CCR { return newGeneral(v, 2) }

func newGeneral(v GeneralInput, action uint32) *CCR {
	return &CCR{
		SessionID: v.SessionID, ServiceContextID: v.ServiceContextID, AuthApplicationID: 4, CCRequestType: 4, RequestedAction: action,
		SubscriptionID: []SubscriptionID{{SubscriptionIDType: 0, SubscriptionIDData: []byte(v.SubscriptionMSISDN)}, {SubscriptionIDType: 1, SubscriptionIDData: []byte(v.SubscriptionIMSI)}},
		EventTimestamp: v.Timestamp, ServiceIdentifier: avp.Some(int32(1)), Timezone: avp.Some(int32(0)),
		MultipleServicesCreditControl: RequestMSCC{RequestedServiceUnit: RequestedUnit{StandardMoney: avp.Some(Money{UnitValue: UnitValue{ValueDigits: v.ValueDigits, Exponent: v.Exponent}, CurrencyCode: v.CurrencyCode}), CCServiceSpecificUnits: v.ServiceSpecificUnits}},
		ServiceInformation: ServiceInformation{
			SMSInformation: SMSInformation{SMSNode: 2, SMDataCodingScheme: v.DataCodingScheme, DestinationInterface: Interface{InterfaceType: v.DestinationInterface}, SMMessageType: v.MessageType, OriginatorInterface: Interface{InterfaceType: v.OriginatorInterface}, SMProtocolID: append([]byte(nil), v.ProtocolID...), ClientAddress: avp.Some(avp.Address{Family: 1, Data: []byte{127, 0, 0, 1}})},
			MMSInformation: avp.Some(MMSInformation{OriginatorAddress: OriginatorAddress{AddressType: 1, AddressData: []byte(v.OriginatorMSISDN)}}),
		},
	}
}
