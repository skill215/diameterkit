package smsfwro

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// ECURInput supplies explicit values for the legacy INITIAL/TERMINATION pair.
// Timestamp is Diameter/NTP time, including a meaningful zero. Strings preserve
// legacy bytes; protocol/user-data headers are opaque. Interfaces are the actual
// templates' fixed destination2/origin3, not inferred from an endpoint role.
type ECURInput struct {
	SessionID, ServiceContextID                                string
	SubscriptionMSISDN, SubscriptionIMSI                       string
	SMSCAddress, RecipientSCCPAddress                          string
	OriginatorMSISDN, RecipientMSISDN, SMSKey                  string
	Timestamp                                                  avp.Time
	DataCodingScheme, ReplyPath, Priority, StatusReportRequest int32
	UserDataLength                                             uint32
	ProtocolID, UserDataHeader                                 []byte
}

// NewInitial constructs the original one-unit INITIAL template with owned data.
// It sends nothing, creates no session state and does not establish a grant.
func NewInitial(v ECURInput) *CCR {
	q := NewIEC(IECInput{SessionID: v.SessionID, ServiceContextID: v.ServiceContextID, SubscriptionMSISDN: v.SubscriptionMSISDN, SubscriptionIMSI: v.SubscriptionIMSI, SMSCAddress: v.SMSCAddress, RecipientSCCPAddress: v.RecipientSCCPAddress, OriginatorMSISDN: v.OriginatorMSISDN, RecipientMSISDN: v.RecipientMSISDN, SMSKey: v.SMSKey, Timestamp: v.Timestamp, DataCodingScheme: v.DataCodingScheme, DestinationInterface: 2, ReplyPath: v.ReplyPath, Priority: v.Priority, StatusReportRequest: v.StatusReportRequest, UserDataLength: v.UserDataLength, ProtocolID: v.ProtocolID, UserDataHeader: v.UserDataHeader})
	q.CCRequestType = 1
	q.RequestedAction = avp.Optional[uint32]{}
	q.ServiceInformation.SMSInformation.OriginatorSCCPAddress = avp.Optional[[]byte]{}
	q.ServiceInformation.SMSInformation.OriginatorInterface.InterfaceType = 3
	return q
}

// TerminationInput carries caller-selected usage and opaque outcome data. Empty
// Status and zero usage/time are explicit values, not absent fields. No grant,
// TPDU delivery status or elapsed time is inferred from these representations.
type TerminationInput struct {
	ECURInput
	UsedServiceSpecificUnits uint64
	Status                   []byte
	DischargeTime            avp.Time
}

// NewTermination constructs a new exchange in the same caller-supplied Session,
// with sequence3/1, FINAL2 and explicit usage; it is not a retransmission of INITIAL.
func NewTermination(v TerminationInput) *CCR {
	q := NewInitial(v.ECURInput)
	q.CCRequestType = 3
	q.CCRequestNumber = 1
	q.MultipleServicesCreditControl = RequestMSCC{UsedServiceUnit: avp.Some(ServiceUnit{CCServiceSpecificUnits: v.UsedServiceSpecificUnits}), ReportingReason: avp.Some(int32(2))}
	q.ServiceInformation.SMSInformation.SMStatus = avp.Some(append([]byte(nil), v.Status...))
	q.ServiceInformation.SMSInformation.DischargeTime = avp.Some(v.DischargeTime)
	return q
}

// CheckECURContinuation validates two caller-supplied complete request bodies:
// the no-UPDATE sequence, same Session/CTF/fixed OCS/ServiceContext, ordered
// subscriptions, SMSKey and originating/recipient MMS addresses. It neither
// retains input nor checks a prior CCA, usage against a grant, delivery success,
// time ordering, routing changes or financial state. Each answer must separately
// be validated against its own request. Do not mutate inputs concurrently.
func CheckECURContinuation(initial, termination *CCR) error {
	if initial == nil || termination == nil {
		return validation.ErrType
	}
	if e := validateCCR(initial, validation.Send); e != nil {
		return e
	}
	if e := validateCCR(termination, validation.Send); e != nil {
		return e
	}
	if initial.CCRequestType != 1 || termination.CCRequestType != 3 {
		return at(416, validation.ErrUnsupported)
	}
	if initial.SessionID != termination.SessionID {
		return at(263, validation.ErrContext)
	}
	if initial.ServiceContextID != termination.ServiceContextID {
		return at(461, validation.ErrContext)
	}
	if !strings.EqualFold(initial.OriginHost, termination.OriginHost) || !strings.EqualFold(initial.OriginRealm, termination.OriginRealm) {
		return at(264, validation.ErrContext)
	}
	if !initial.DestinationHost.Present || !termination.DestinationHost.Present {
		return at(293, validation.ErrContextRequired)
	}
	if !strings.EqualFold(initial.DestinationHost.Value, termination.DestinationHost.Value) || !strings.EqualFold(initial.DestinationRealm, termination.DestinationRealm) {
		return at(293, validation.ErrContext)
	}
	for i := range initial.SubscriptionID {
		if !bytes.Equal(initial.SubscriptionID[i].SubscriptionIDData, termination.SubscriptionID[i].SubscriptionIDData) {
			return at(443, validation.ErrContext)
		}
	}
	a, b := initial.ServiceInformation.MMSInformation, termination.ServiceInformation.MMSInformation
	if !bytes.Equal(a.SMSKey, b.SMSKey) || !bytes.Equal(a.OriginatorAddress.AddressData, b.OriginatorAddress.AddressData) || !bytes.Equal(a.RecipientAddress.AddressData, b.RecipientAddress.AddressData) {
		return vendorAt(877, validation.ErrContext)
	}
	return nil
}
