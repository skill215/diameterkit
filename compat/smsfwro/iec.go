// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package smsfwro

import "github.com/skill215/diameterkit/avp"

// IECInput contains explicit application values for yppdia_iec_avptemplate.cf.
// Empty strings and zero numbers are real inputs, not lookup/default requests.
// Timestamp is the Diameter/NTP wire value, not Unix seconds. The application
// performs any epoch/era conversion once. No clock or identity lookup is used.
// Strings preserve bytes without UTF-8/number normalization; NUL is rejected by
// profile validation. PID and UDH preserve their length and embedded zeros.
type IECInput struct {
	SessionID, ServiceContextID                                                      string
	SubscriptionMSISDN, SubscriptionIMSI                                             string
	SMSCAddress, OriginatorSCCPAddress, RecipientSCCPAddress                         string
	OriginatorMSISDN, RecipientMSISDN, SMSKey                                        string
	Timestamp                                                                        avp.Time
	DataCodingScheme, DestinationInterface, ReplyPath, Priority, StatusReportRequest int32
	UserDataLength                                                                   uint32
	ProtocolID, UserDataHeader                                                       []byte
}

// NewIEC copies inputs and supplies only constants/copies present in the old
// template. The returned body still requires normal codec/Submit validation and
// an explicit runtime endpoint binding. It sends nothing and does not charge,
// allocate Session-Id, infer missing subscriber data, or select an OCS.
func NewIEC(v IECInput) *CCR {
	return &CCR{
		SessionID: v.SessionID, ServiceContextID: v.ServiceContextID, AuthApplicationID: 4, CCRequestType: 4, RequestedAction: avp.Some(uint32(0)),
		SubscriptionID: []SubscriptionID{{SubscriptionIDType: 0, SubscriptionIDData: []byte(v.SubscriptionMSISDN)}, {SubscriptionIDType: 1, SubscriptionIDData: []byte(v.SubscriptionIMSI)}},
		EventTimestamp: v.Timestamp, MultipleServicesIndicator: 1,
		MultipleServicesCreditControl: RequestMSCC{RequestedServiceUnit: avp.Some(ServiceUnit{CCServiceSpecificUnits: 1})},
		UserEquipmentInfo:             UserEquipmentInfo{UserEquipmentInfoValue: []byte(v.SubscriptionIMSI)},
		ServiceInformation: ServiceInformation{
			SMSInformation: SMSInformation{
				ClientAddress: []byte(v.SMSCAddress), OriginatorSCCPAddress: avp.Some([]byte(v.OriginatorSCCPAddress)), RecipientSCCPAddress: []byte(v.RecipientSCCPAddress), SMSCAddress: []byte(v.SMSCAddress),
				SMDataCodingScheme: v.DataCodingScheme, DestinationInterface: Interface{InterfaceType: v.DestinationInterface}, OriginatorInterface: Interface{InterfaceType: 1},
				SMProtocolID: append([]byte(nil), v.ProtocolID...), ReplyPathRequested: v.ReplyPath, SMUserDataHeader: append([]byte(nil), v.UserDataHeader...),
			},
			MMSInformation: MMSInformation{
				OriginatorAddress: Address{AddressType: 1, AddressData: []byte(v.OriginatorMSISDN)}, RecipientAddress: Address{AddressType: 1, AddressData: []byte(v.RecipientMSISDN)},
				SubmissionTime: v.Timestamp, Priority: v.Priority, UserDataLength: v.UserDataLength, SMSKey: []byte(v.SMSKey), StatusReportRequest: v.StatusReportRequest,
			},
		},
	}
}
