// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package s6c

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// OnlyMCCMNCRequested is SM-Delivery-Not-Intended=1. Value zero (only IMSI)
// is a distinct, currently unsupported query, never the absent-field default.
const OnlyMCCMNCRequested int32 = 1

// RouterRoutingEvidence asserts the GMSC/Router endpoint leg and deployed
// enforcement of routing via the receiving subscriber's HPLMN. It does not
// declare the adjacent DRA a Router or authorize an IMS/T4 request rewrite.
type RouterRoutingEvidence struct {
	GMSC, Router  pdu.Origin
	HPLMNEnforced bool
}

// RouterRouting freezes a bounded assertion for a MSISDN-based SRR and its
// answer. The successful User-Name is an opaque MT correlation ID, not an IMSI.
// The application supplies the token and its mapping; the library stores neither.
func RouterRouting(e RouterRoutingEvidence) (validation.Context, error) {
	if !e.HPLMNEnforced {
		return validation.Context{}, validation.ErrContext
	}
	s, err := routingEndpoints(e.GMSC, e.Router)
	if err != nil {
		return validation.Context{}, err
	}
	return contextvalue.NewRouterRouting(s), nil
}

// MCCMNCRoutingEvidence declares a non-delivery MSISDN query to a fixed
// responder. MCCMNC is an optional, exact 5/6-digit prefix supplied by the
// application. Unknown answer facts may remain absent at submission; successful
// answers then need offline checking with new evidence. No MNC length is guessed.
type MCCMNCRoutingEvidence struct {
	Requester, Responder pdu.Origin
	MCCMNC               avp.Optional[string]
}

// MCCMNCRouting freezes the original query purpose and optional expected PLMN.
// It never assigns a dummy MSIN, resolves number portability or starts delivery.
func MCCMNCRouting(e MCCMNCRoutingEvidence) (validation.Context, error) {
	s, err := routingEndpoints(e.Requester, e.Responder)
	if err != nil {
		return validation.Context{}, err
	}
	if e.MCCMNC.Present {
		if (len(e.MCCMNC.Value) != 5 && len(e.MCCMNC.Value) != 6) || imsi(e.MCCMNC.Value) != nil {
			return validation.Context{}, validation.ErrContext
		}
		s.UNRIKnown = true
		s.UserName = strings.Clone(e.MCCMNC.Value)
	}
	return contextvalue.NewMCCMNCRouting(s), nil
}
func routingEndpoints(requester, responder pdu.Origin) (contextvalue.SMSIdentity, error) {
	for _, o := range []pdu.Origin{requester, responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return contextvalue.SMSIdentity{}, validation.ErrContext
		}
	}
	return contextvalue.SMSIdentity{RequesterHost: strings.Clone(requester.Host), RequesterRealm: strings.Clone(requester.Realm), ResponderHost: strings.Clone(responder.Host), ResponderRealm: strings.Clone(responder.Realm)}, nil
}
func imsi(s string) error {
	if len(s) < 5 || len(s) > 15 {
		return validation.ErrField
	}
	for _, v := range s {
		if v < '0' || v > '9' {
			return validation.ErrField
		}
	}
	return nil
}
func checkSpecialRouting(r *SRR, answer pdu.Body, ctx validation.Context) error {
	e, router := ctx.RouterRouting()
	if !router {
		var ok bool
		e, ok = ctx.MCCMNCRouting()
		if !ok {
			return validation.ErrContext
		}
	}
	if r.UserName.Present || !r.MSISDN.Present || r.SMSMICorrelationID.Present || !r.DestinationHost.Present {
		return validation.ErrContext
	}
	if router && r.SMDeliveryNotIntended.Present || !router && (!r.SMDeliveryNotIntended.Present || r.SMDeliveryNotIntended.Value != OnlyMCCMNCRequested) {
		return validation.ErrContext
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
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
	if !a.ResultCode.Present || a.ResultCode.Value != 2001 {
		return nil
	}
	if router {
		return validation.At(avp.Key{Code: 1}, smsidentity.AlertToken(a.UserName.Value))
	}
	if err := imsi(a.UserName.Value); err != nil {
		return validation.At(avp.Key{Code: 1}, err)
	}
	if !e.UNRIKnown {
		return validation.ErrContextRequired
	}
	// A supplied real IMSI is also permitted: no claim about subscriber allocation
	// or which MSIN bytes are dummy can be established from a string alone.
	if !strings.HasPrefix(a.UserName.Value, e.UserName) || len(a.UserName.Value) <= len(e.UserName) {
		return validation.At(avp.Key{Code: 1}, validation.ErrField)
	}
	return nil
}
