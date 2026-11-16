package s6c

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// IMSRoutingEvidence declares one MSISDN-less IMS delivery routing query to a
// known HSS. HSSID is the request identity, not an expected subscriber IMSI.
// Successful answers in this slice mean a real IMSI supplied by the HSS; the
// library checks syntax, not subscriber allocation or the existence of an IMSI.
// UNRI has the same external-fact semantics as RoutingRoleEvidence.
type IMSRoutingEvidence struct {
	Role                                RoutingOriginRole
	Requester, Responder                pdu.Origin
	HSSID, DestinationSIPURI, SCAddress string
	UNRI                                avp.Optional[bool]
}

// IMSRouting owns bounded identity/purpose evidence for SRR/SRA only. It does
// not resolve HSSs, synthesize answer identities, change nodes or send an SMS.
func IMSRouting(e IMSRoutingEvidence) (validation.Context, error) {
	if e.Role != RoutingFromGMSC && e.Role != RoutingFromIPSMGW || e.Role == RoutingFromGMSC && e.UNRI.Present {
		return validation.Context{}, validation.ErrContext
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if hssID(e.HSSID) != nil || reportSIPURI(e.DestinationSIPURI) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	s := contextvalue.SMSIdentity{Command: 8388647, Role: uint8(e.Role), RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UserName: strings.Clone(e.HSSID), SCAddress: strings.Clone(e.SCAddress), DestinationSIPURI: strings.Clone(e.DestinationSIPURI), UNRIKnown: e.UNRI.Present}
	if e.UNRI.Present {
		s.UNRI = e.UNRI.Value
	}
	return contextvalue.NewS6cIMSRouting(s), nil
}
func validateRoutingCorrelation(m *RoutingSMSMICorrelationID, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3325), smsKey(3326)); err != nil {
		return err
	}
	if mode != validation.Receive {
		return validation.At(smsKey(3327), reportSIPURI(m.DestinationSIPURI))
	}
	return nil
}
func checkIMSRouting(r *SRR, answer pdu.Body, ctx validation.Context) error {
	if err := validateRoutingCorrelation(&r.SMSMICorrelationID.Value, validation.Send); err != nil {
		return validation.At(smsKey(3324), err)
	}
	if ctx.Empty() {
		return validation.ErrContextRequired
	}
	e, ok := ctx.S6cIMSRouting()
	if !ok {
		return validation.ErrContext
	}
	if !r.DestinationHost.Present || !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) || r.UserName.Value != e.UserName || r.SMSMICorrelationID.Value.DestinationSIPURI != e.DestinationSIPURI {
		return validation.ErrContext
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*SRA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if isAbsentSRA(answer) && e.Role == uint8(RoutingFromIPSMGW) && !a.UserName.Present {
		if !e.UNRIKnown {
			return validation.ErrContextRequired
		}
		if !e.UNRI {
			return validation.At(avp.Key{Code: 1}, validation.ErrMissing)
		}
	}
	return nil
}
