// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"testing"
	"time"
)

func legacyDetailedAnswer(q *combined.CCR, top, child uint32) *combined.CCA {
	a := combinedAnswer(q, top)
	a.MultipleServicesCreditControl = avp.Some(combined.ServiceResult{ResultCode: child})
	if top == 2001 && child == 2001 {
		a.CostInformation = avp.Some(combined.CostInformation{UnitValue: combined.CostValue{ValueDigits: -123}, CurrencyCode: 840})
	} else {
		a.CheckBalanceResult = avp.Optional[uint32]{}
	}
	if q.CCRequestType == 4 && q.RequestedAction.Present && q.RequestedAction.Value == 0 {
		a.DirectDebitingFailureHandling = avp.Some(uint32(0))
	}
	return a
}
func TestLegacyAnswerNetworkLayersAndDetails(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
			for _, codes := range [][2]uint32{{2001, 2001}, {2001, 4012}, {5012, 2001}, {4010, 4010}, {4011, 4011}, {5030, 5030}} {
				id, e := p.Submit(ctxTest(t), combinedBody(t, kind), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				r := routeRead(t, co, remote)
				a, e := co.CompleteAnswer(r, legacyDetailedAnswer(r.Body.(*combined.CCR), codes[0], codes[1]), serverOrigin)
				if e != nil {
					t.Fatal(kind, codes, e)
				}
				if e = writeWire(ctxTest(t), remote, co, a); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || !got.Validation.Complete {
					t.Fatal(got)
				}
				body := got.Answer.Body.(*combined.CCA)
				if body.ResultCode != codes[0] || !body.MultipleServicesCreditControl.Present || body.MultipleServicesCreditControl.Value.ResultCode != codes[1] {
					t.Fatal("result layer lost", body)
				}
				if body.CostInformation.Present && (body.CostInformation.Value.UnitValue.ValueDigits != -123 || body.CostInformation.Value.UnitValue.Exponent.Present) {
					t.Fatal("cost default/truncation")
				}
			}
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(quiet, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("result caused automatic charging/retry", e)
		}
	})
}
func TestLegacyAnswerNetworkPrepareAndLate(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		r, e := co.CompleteRequest(combinedBody(t, "balance"), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 800, EndToEnd: 900})
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, r); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		in.Message().Body.(*combined.CCR).RequestedAction = avp.Some(uint32(0))
		bad := legacyDetailedAnswer(r.Body.(*combined.CCR), 2001, 2001)
		bad.DirectDebitingFailureHandling = avp.Some(uint32(1))
		if got, e := in.Respond(ctxTest(t), bad); e == nil || got.Committed {
			t.Fatal("mutated request enabledDDFH", got, e)
		}
		bad.DirectDebitingFailureHandling = avp.Optional[uint32]{}
		bad.MultipleServicesCreditControl.Value.ResultCode = 4999
		if got, e := in.Respond(ctxTest(t), bad); e == nil || got.Committed {
			t.Fatal("unknown child accepted", got, e)
		}
		good := legacyDetailedAnswer(r.Body.(*combined.CCR), 2001, 4012)
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if got, e := in.Respond(canceled, good); got.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(got, e)
		}
		if got, e := in.Respond(ctxTest(t), good); e != nil || !got.Committed {
			t.Fatal(got, e)
		}
		a := routeRead(t, co, remote)
		if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		a, e = co.CompleteAnswer(r, legacyDetailedAnswer(r.Body.(*combined.CCR), 2001, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		quiet, done := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer done()
		if _, e = port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late detailed answer completed twice", e)
		}
		n.Close()
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
