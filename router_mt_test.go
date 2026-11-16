// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func routerMTBody() *sgd.TFR {
	b := smsBody(true).(*sgd.TFR)
	b.MMENumberForMTSMS.Present = false
	b.MaximumRetransmissionTime = avp.Some(avp.Time(16))
	b.SMSGMSCAddress = avp.Some([]byte{0x21, 0xf3})
	b.SMDeliveryStartTime = avp.Some(avp.Time(0xfffffff0))
	return b
}
func routerMTAnswer(seconds avp.Time) *sgd.TFA {
	return &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5550}), RequestedRetransmissionTime: avp.Some(seconds), UserIdentifier: avp.Some(sgd.AlertUserIdentifier{UserName: "router-token-75"})}
}
func routerMTFacts(requester, responder pdu.Origin) sgd.RetransmissionEvidence {
	return sgd.RetransmissionEvidence{Target: sgd.RetransmissionToRouter, RouterAlertSupported: avp.Some(true), Requester: requester, Responder: responder, SamePLMN: avp.Some(true), PowerSaving: avp.Some(true), MaximumEra: 1, RequestedEra: avp.Some(uint32(1))}
}
func routerMTContext(t *testing.T, e sgd.RetransmissionEvidence) validation.Context {
	t.Helper()
	v, err := sgd.Retransmission(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func receiveRouterMT(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r, e := co.CompleteRequest(routerMTBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
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
	return in, r
}
func TestRouterMTInboundFreezeAndStrictSend(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveRouterMT(t, n, p, co, remote, 4001)
		if r := in.Validation(); r.State != validation.NeedsApplication {
			t.Fatal(r)
		}
		if _, e := in.Respond(ctxTest(t), routerMTAnswer(0)); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		facts := routerMTFacts(gmscOrigin, n.cfg.Origin)
		wrong := facts
		wrong.Responder = draOrigin
		if r := in.Check(routerMTContext(t, wrong)); r.Complete {
			t.Fatal(r)
		}
		missing := routerMTAnswer(0)
		missing.UserIdentifier.Present = false
		// Application view cannot replace the original upper bound or target.
		view := in.Message().Body.(*sgd.TFR)
		view.MaximumRetransmissionTime.Value = 100
		view.SMSGMSCAddress.Value[0] = 0xff
		ctx := routerMTContext(t, facts)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if rr, err := in.Respond(ctxTest(t), missing); rr.Committed || !errors.Is(err, validation.ErrMissing) {
			t.Fatal(rr, err)
		}
		wrong = facts
		wrong.MaximumEra = 2
		if r := in.Check(routerMTContext(t, wrong)); r.Complete {
			t.Fatal("changed frozen era")
		}
		for _, tc := range []struct {
			v   avp.Time
			err error
		}{{16, validation.ErrUnsupported}, {17, validation.ErrField}} {
			if r, e := in.Respond(ctxTest(t), routerMTAnswer(tc.v)); r.Committed || !errors.Is(e, tc.err) {
				t.Fatal(r, e)
			}
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(canceled, routerMTAnswer(0)); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), routerMTAnswer(0)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		a := duplicateAnswer(t, n, co, remote, req)
		if r := co.ValidateAnswerContext(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
			t.Fatal(r)
		}
		if !a.Body.(*sgd.TFA).RequestedRetransmissionTime.Present {
			t.Fatal("lost zero")
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("unexpected SMS retry", e)
		}
	})
}
func TestRouterMTOutboundReportsAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		facts := routerMTFacts(n.cfg.Origin, serverOrigin)
		ctx := routerMTContext(t, facts)
		if _, e := p.Submit(ctxTest(t), routerMTBody(), RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		bad := facts
		bad.Responder = draOrigin
		if _, e := p.Submit(ctxTest(t), routerMTBody(), RequestOptions{Validation: routerMTContext(t, bad)}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		bad = facts
		bad.RouterAlertSupported.Present = false
		if _, e := p.Submit(ctxTest(t), routerMTBody(), RequestOptions{Validation: routerMTContext(t, bad)}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		bad.RouterAlertSupported = avp.Some(false)
		if _, e := p.Submit(ctxTest(t), routerMTBody(), RequestOptions{Validation: routerMTContext(t, bad)}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		for _, tc := range []struct {
			answer  pdu.Body
			unknown bool
			state   validation.State
		}{
			{routerMTAnswer(0), false, validation.Ready}, {routerMTAnswer(16), false, validation.NeedsApplication},
			{routerMTAnswer(17), false, validation.Rejected}, {routerMTAnswer(0), true, validation.NeedsApplication},
			{smsAnswer(true, 2001), true, validation.Ready}, {&pdu.ErrorAnswer{ResultCode: 3004}, true, validation.Ready},
		} {
			e := facts
			if tc.unknown {
				e.RequestedEra.Present = false
				e.PowerSaving.Present = false
			}
			b := routerMTBody()
			id, err := p.Submit(ctxTest(t), b, RequestOptions{Validation: routerMTContext(t, e)}, port)
			if err != nil {
				t.Fatal(err)
			}
			b.MaximumRetransmissionTime.Value = 999
			req := routeRead(t, co, remote)
			ans, err := co.CompleteAnswer(req, tc.answer, serverOrigin)
			if err != nil {
				t.Fatal(err)
			}
			if err = writeWire(ctxTest(t), remote, co, ans); err != nil {
				t.Fatal(err)
			}
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Outcome != AnswerReceived || got.Validation.State != tc.state {
				t.Fatal(got)
			}
			if a, ok := got.Answer.Body.(*sgd.TFA); ok && a.RequestedRetransmissionTime.Present && (a.ExperimentalResult.Value.ExperimentalResultCode != 5550 || !a.UserIdentifier.Present || a.UserIdentifier.Value.UserName != "router-token-75") {
				t.Fatal(a)
			}
			if tc.unknown && tc.state == validation.NeedsApplication {
				if report := co.ValidateAnswerContext(req, got.Answer, pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}}, ctx); !report.Complete {
					t.Fatal(report)
				}
			}
		}
		id, err := p.Submit(ctxTest(t), routerMTBody(), RequestOptions{Validation: ctx}, port)
		if err != nil {
			t.Fatal(err)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, err := co.CompleteAnswer(req, routerMTAnswer(0), serverOrigin)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctxTest(t), remote, co, a); err != nil {
			t.Fatal(err)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err = port.Receive(short); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRouterMTDuplicateCacheAndConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveRouterMT(t, n, p, co, remote, 4101)
		ctx := routerMTContext(t, routerMTFacts(gmscOrigin, n.cfg.Origin))
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*sgd.TFR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra2.example", ProxyState: []byte{4, 2}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), routerMTAnswer(0)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if report := co.ValidateAnswerContext(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !report.Complete {
				t.Fatal(report)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		cached := duplicateAnswer(t, n, co, remote, dup)
		if cached.Body.(*sgd.TFA).UserIdentifier.Value.UserName != "router-token-75" {
			t.Fatal("cached token lost")
		}
		for i := range 2 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(4110 + i)
			b := conflict.Body.(*sgd.TFR)
			if i == 0 {
				b.MaximumRetransmissionTime.Value++
			} else {
				b.SMSGMSCAddress.Value[0] = 0x43
			}
			if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
				t.Fatal(e)
			}
			if a, ok := routeRead(t, co, remote).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 5012 {
				t.Fatal(a)
			}
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRouterMTBudgetAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := smsConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveRouterMT(t, n, p, co, remote, 4201)
		other, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		other.Header.HopByHop++
		other.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, other); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, other).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		ctx := routerMTContext(t, routerMTFacts(gmscOrigin, n.cfg.Origin))
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), routerMTAnswer(0)) }()
		close(start)
		_ = n.Close()
		wg.Wait()
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
