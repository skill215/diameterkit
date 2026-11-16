package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// RoutingQuery uses international decimal digits without '+' or TON/NPI.
// Fields supplies optional protocol fields and endpoints, including the typed
// T4/IMS UserName/SMSMICorrelationID. SessionID, AuthSessionState, MSISDN and
// SCAddress in Fields must be empty. Named context remains explicit at use time.
type RoutingQuery struct {
	SessionID string
	MSISDN    avp.Optional[string]
	SCAddress string
	Fields    SRR
}

// NewSRR constructs an owned request for the current fixed S6c Profile, with
// no I/O or context inference. Runtime-owned envelope fields may be omitted.
func NewSRR(q RoutingQuery) (*SRR, error) {
	r := q.Fields
	if r.SessionID != "" || r.AuthSessionState != 0 || r.MSISDN.Present || len(r.SCAddress) != 0 {
		return nil, validation.ErrEnvelope
	}
	r.SessionID, r.AuthSessionState = q.SessionID, 1
	var err error
	r.SCAddress, err = avp.EncodeTBCD(q.SCAddress)
	if err != nil {
		return nil, validation.At(smsKey(3300), err)
	}
	if q.MSISDN.Present {
		b, e := avp.EncodeTBCD(q.MSISDN.Value)
		if e != nil {
			return nil, validation.At(smsKey(701), e)
		}
		r.MSISDN = avp.Some(b)
	}
	if err = validateSRR(&r, validation.Build); err != nil {
		return nil, err
	}
	return cloneSRR(&r), nil
}
