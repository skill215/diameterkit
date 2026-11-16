package legacy420

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// RefundInput contains explicit application values, not inferred financial defaults.
// Digits/exponent and legacy signed currency are encoded exactly, without float
// conversion, rounding, currency lookup or multiplication. Zero is a real value.
// The old sample used 1/0/840 and one unit; this API does not silently insert them.
// Timestamp is already Diameter/NTP wire time. No clock or Unix conversion occurs.
// RefundInformation must be Present, including an explicitly empty token. Its
// origin, mapping from an approved OCS answer, ownership and eligibility remain
// application responsibilities; the library cannot verify a financial receipt.
type RefundInput struct {
	SessionID, ServiceContextID                                              string
	SubscriptionMSISDN, SubscriptionIMSI                                     string
	Timestamp                                                                avp.Time
	ValueDigits                                                              int64
	Exponent, CurrencyCode                                                   int32
	ServiceSpecificUnits                                                     uint64
	DataCodingScheme, DestinationInterface, MessageType, OriginatorInterface int32
	ProtocolID                                                               []byte
	RefundInformation                                                        avp.Optional[[]byte]
}

// NewRefund creates an owned EVENT/REFUND body, copying explicit values and only
// the template's structural constants. The caller supplies a new Session-Id and
// submits a new exchange, not a retransmission of a completed debit. This helper
// neither allocates nor verifies cross-process identity, and sends nothing.
// Subsequent codec/Submit validation and a fixed OCS binding are required.
func NewRefund(v RefundInput) (*CCR, error) {
	if !v.RefundInformation.Present {
		return nil, at(1292, validation.ErrMissing)
	}
	return &CCR{
		SessionID: v.SessionID, ServiceContextID: v.ServiceContextID, AuthApplicationID: 4, CCRequestType: 4, RequestedAction: 1,
		SubscriptionID:                []SubscriptionID{{SubscriptionIDType: 0, SubscriptionIDData: []byte(v.SubscriptionMSISDN)}, {SubscriptionIDType: 1, SubscriptionIDData: []byte(v.SubscriptionIMSI)}},
		EventTimestamp:                v.Timestamp,
		MultipleServicesCreditControl: RequestMSCC{RequestedServiceUnit: RequestedUnit{Money: avp.Some(Money{UnitValue: UnitValue{ValueDigits: v.ValueDigits, Exponent: v.Exponent}, CurrencyCode: v.CurrencyCode}), CCServiceSpecificUnits: v.ServiceSpecificUnits}},
		ServiceInformation:            ServiceInformation{SMSInformation: SMSInformation{SMSNode: 2, SMDataCodingScheme: v.DataCodingScheme, DestinationInterface: Interface{InterfaceType: v.DestinationInterface}, SMMessageType: v.MessageType, OriginatorInterface: Interface{InterfaceType: v.OriginatorInterface}, SMProtocolID: append([]byte(nil), v.ProtocolID...)}},
		RefundInformation:             avp.Some(append([]byte(nil), v.RefundInformation.Value...)),
	}, nil
}
