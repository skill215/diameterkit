// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sgd

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// RetransmissionTarget identifies the actual end terminal, never an adjacent DRA.
// Router is a separate endpoint leg; the library never forwards/translates it.
type RetransmissionTarget uint8

const (
	RetransmissionToMME RetransmissionTarget = iota + 1
	RetransmissionToSGSN
	RetransmissionToRouter
)

// RetransmissionEvidence is a caller assertion for one original TFR/TFA pair.
// MaximumEra interprets the original maximum. RequestedEra, when known, is the
// application's explicit interpretation of the answer; no era is guessed from
// the high bit, current time or delivery start. Era zero is valid.
// Unknown answer facts may be left absent on Submit: a time-bearing answer then
// remains owned and correlated but NeedsApplication. It can be checked offline
// with new evidence; the active exchange's context is never changed.
// Before Inbound.Check, supply facts needed to send the intended answer, since
// a successful Check freezes the context. PowerSaving includes current UE
// unreachability for this procedure, not just support for a feature.
type RetransmissionEvidence struct {
	Target               RetransmissionTarget
	Requester, Responder pdu.Origin
	SamePLMN             avp.Optional[bool]
	PowerSaving          avp.Optional[bool]
	MaximumEra           uint32
	RequestedEra         avp.Optional[uint32]
	// RouterAlertSupported is required, known and true when a Router-bound
	// request carries Maximum-Retransmission-Time. Only used with ToRouter;
	// adjacent DRA capabilities and the presence of a token do not prove it.
	RouterAlertSupported avp.Optional[bool]
}

// Retransmission owns a bounded context. It does no lookups, scheduling or
// storage. SamePLMN is required for a time-bearing request; PowerSaving and
// RequestedEra are required only when checking a time-bearing normal answer.
func Retransmission(e RetransmissionEvidence) (validation.Context, error) {
	if e.Target != RetransmissionToMME && e.Target != RetransmissionToSGSN && e.Target != RetransmissionToRouter {
		return validation.Context{}, validation.ErrContext
	}
	if e.Target != RetransmissionToRouter && e.RouterAlertSupported.Present {
		return validation.Context{}, validation.ErrContext
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	r := contextvalue.Retransmission{Target: uint8(e.Target), RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), SamePLMNKnown: e.SamePLMN.Present, PowerSavingKnown: e.PowerSaving.Present, MaximumEra: e.MaximumEra, RequestedEraKnown: e.RequestedEra.Present}
	if e.SamePLMN.Present {
		r.SamePLMN = e.SamePLMN.Value
	}
	if e.PowerSaving.Present {
		r.PowerSaving = e.PowerSaving.Value
	}
	if e.RequestedEra.Present {
		r.RequestedEra = e.RequestedEra.Value
	}
	if e.RouterAlertSupported.Present {
		r.RouterAlertKnown, r.RouterAlert = true, e.RouterAlertSupported.Value
	}
	return contextvalue.NewRetransmission(r), nil
}

func checkRetransmission(request, answer pdu.Body, ctx validation.Context, mode validation.Mode) error {
	r, ok := request.(*TFR)
	if !ok {
		return validation.ErrContextRequired
	}
	if r.SMSMICorrelationID.Present {
		return checkIMSMT(r, answer, ctx)
	}
	e, present := ctx.Retransmission()
	if !present {
		if !ctx.Empty() {
			return validation.ErrContext
		}
		if r.MaximumRetransmissionTime.Present || !r.MMENumberForMTSMS.Present && !r.SGSNNumber.Present {
			return validation.ErrContextRequired
		}
		if a, ok := answer.(*TFA); ok && a.UserIdentifier.Present {
			return validation.ErrContextRequired
		}
		return nil
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	router := e.Target == uint8(RetransmissionToRouter)
	if !router && (!r.MMENumberForMTSMS.Present && !r.SGSNNumber.Present || (e.Target == uint8(RetransmissionToMME)) != r.MMENumberForMTSMS.Present) {
		return validation.ErrContext
	}
	if r.MaximumRetransmissionTime.Present {
		if !e.SamePLMNKnown {
			return validation.ErrContextRequired
		}
		if !e.SamePLMN {
			return validation.ErrContext
		}
		if router {
			if !e.RouterAlertKnown {
				return validation.ErrContextRequired
			}
			if !e.RouterAlert {
				return validation.ErrContext
			}
		}
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*TFA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if !router && a.UserIdentifier.Present {
		return validation.At(smsKey(3102), validation.ErrUnsupported)
	}
	if !a.RequestedRetransmissionTime.Present {
		return nil
	}
	if router && !a.UserIdentifier.Present {
		return validation.At(smsKey(3102), validation.ErrMissing)
	}
	if !r.MaximumRetransmissionTime.Present {
		return validation.At(smsKey(3331), validation.ErrField)
	}
	if !e.PowerSavingKnown || !e.RequestedEraKnown {
		return validation.ErrContextRequired
	}
	if !e.PowerSaving {
		return validation.ErrContext
	}
	// Lexicographic era/seconds ordering, no duration arithmetic or overflow.
	after := e.RequestedEra > e.MaximumEra || e.RequestedEra == e.MaximumEra && a.RequestedRetransmissionTime.Value > r.MaximumRetransmissionTime.Value
	equal := e.RequestedEra == e.MaximumEra && a.RequestedRetransmissionTime.Value == r.MaximumRetransmissionTime.Value
	if after {
		return validation.At(smsKey(3331), validation.ErrField)
	}
	if equal {
		// TS29.338 table says <=, procedure says prior to. Neither silent
		// acceptance nor declaring a malformed peer answer is justified.
		if mode == validation.Receive {
			return validation.ErrContextRequired
		}
		return validation.At(smsKey(3331), validation.ErrUnsupported)
	}
	return nil
}
