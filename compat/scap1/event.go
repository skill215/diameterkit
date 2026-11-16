// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package scap1

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// EventInput contains explicit values after the application's Getter-equivalent
// decisions. Optional distinguishes a missing input from an explicitly empty
// legacy STRING or zero wire Time. Text is not parsed, normalized or reformatted;
// DCS/sequence/length text must be supplied by the application. The library does
// not reproduce S1 structure access, charged-party selection or fake-VMSC policy.
type EventInput struct {
	SessionID                                          string
	SubscriptionMSISDN                                 avp.Optional[string]
	Timestamp                                          avp.Optional[avp.Time]
	DataCodingScheme, OriginatorMSISDN, OriginatorVMSC avp.Optional[string]
	TrafficCase, DestinationMSISDN                     avp.Optional[string]
	ConcatenatedMessages, UserDataLength               avp.Optional[string]
}

// NewEvent constructs the source template's EVENT record1/number0/action0 and
// literal one event unit. It creates no identity, sends nothing and stores no
// accounting state. CompleteRequest supplies the explicit routing envelope;
// codec validation is still required. Node use requires explicit SCAP1Accounting
// policy and capability configuration; no advertisement is installed here.
func NewEvent(v EventInput) (*ACR, error) {
	if !v.SubscriptionMSISDN.Present {
		return nil, at(554, 193, validation.ErrMissing)
	}
	if !v.Timestamp.Present {
		return nil, at(610, 193, validation.ErrMissing)
	}
	vals := []avp.Optional[string]{v.DataCodingScheme, v.OriginatorMSISDN, v.OriginatorVMSC, v.TrafficCase, v.DestinationMSISDN, v.ConcatenatedMessages, v.UserDataLength}
	for _, x := range vals {
		if !x.Present {
			return nil, at(607, 193, validation.ErrMissing)
		}
	}
	types := [...]int32{0, 1, 5, 8, 14, 16, 17, 18, 2, 3, 13}
	values := [...]string{"1", v.DataCodingScheme.Value, v.OriginatorMSISDN.Value, v.OriginatorVMSC.Value, v.TrafficCase.Value, v.OriginatorMSISDN.Value, v.DestinationMSISDN.Value, "4", v.ConcatenatedMessages.Value, v.UserDataLength.Value, "+0600"}
	r := &ACR{SessionID: v.SessionID, VendorSpecificApplicationID: VendorApplication{VendorID: 193, AcctApplicationID: 19302}, SubscriptionID: SubscriptionID{SubscriptionIDType: 0, SubscriptionIDData: []byte(v.SubscriptionMSISDN.Value)}, AccountingRecordType: 1, RequestedServiceUnit: RequestedUnit{UnitType: 2, UnitValue: UnitValue{ValueDigits: 1}}, EventTimestamp: v.Timestamp.Value}
	for i, t := range types {
		r.ServiceParameterInfo = append(r.ServiceParameterInfo, ServiceParameter{ParameterType: t, ParameterValue: []byte(values[i])})
	}
	return r, nil
}
