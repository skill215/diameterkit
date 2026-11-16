// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"sync"
	"testing"
	"time"
)

func t4RoutingBody(t *testing.T) *s6c.SRR {
	b := routingRequest(t)
	b.MSISDN = avp.Optional[[]byte]{}
	b.UserName = avp.Some("001010123456789")
	return b
}
func mutateT4Routing(body pdu.Body, sc bool) {
	b := body.(*s6c.SRR)
	if sc {
		b.SCAddress[0] = 0x43
	} else {
		b.UserName = avp.Some("001010123456780")
	}
}
func t4RoutingContext(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, e := s6c.T4NoMSISDN(s6c.T4NoMSISDNEvidence{Procedure: s6c.T4Routing, Requester: requester, Responder: responder, IMSI: "001010123456789", SCAddress: "12345"})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func receiveT4Routing(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req, e := co.CompleteRequest(t4RoutingBody(t), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), remote, co, req); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, req
}
func TestS6cT4RoutingInboundRoleFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveT4Routing(t, n, p, co, remote, 3001)
		if r := in.Validation(); r.State != validation.NeedsApplication || r.Complete {
			t.Fatal(r)
		}
		if _, e := in.Respond(ctxTest(t), routingAnswer()); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		for _, ctx := range []validation.Context{t4Context(t, false, serverOrigin, n.cfg.Origin), t4RoutingContext(t, draOrigin, n.cfg.Origin), t4RoutingContext(t, serverOrigin, serverOrigin)} {
			if r := in.Check(ctx); r.Complete {
				t.Fatal(r)
			}
		}
		mutateT4Routing(in.Message().Body, false)
		mutateT4Routing(in.Message().Body, true)
		ctx := t4RoutingContext(t, serverOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(t4Context(t, false, serverOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("changed frozen role")
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, routingAnswer()); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		bad := routingAnswer()
		bad.UserName = avp.Some("001010123456780")
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), routingAnswer()); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic SRR or SMS", e)
		}
	})
}
func TestS6cT4RoutingOutboundBindingAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		body := t4RoutingBody(t)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		wrong := t4RoutingContext(t, n.cfg.Origin, draOrigin)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: wrong}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		answers := []pdu.Body{routingAnswer(), &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 3004}}
		answers = append(answers, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})})
		for _, answer := range answers {
			b := t4RoutingBody(t)
			ctx := t4RoutingContext(t, n.cfg.Origin, serverOrigin)
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			mutateT4Routing(b, false)
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, answer, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
		}
		id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: t4RoutingContext(t, n.cfg.Origin, serverOrigin)}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, e := co.CompleteAnswer(req, routingAnswer(), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestS6cT4RoutingDuplicateCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveT4Routing(t, n, p, co, remote, 3101)
		if r := in.Check(t4RoutingContext(t, serverOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		proxies := []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 2}}}
		dup.Body.(*s6c.SRR).ProxyInfo = proxies
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), routingAnswer()); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			r := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				r = dup
			}
			if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, dup)
		for i := range 2 {
			c, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			c.Header.HopByHop = uint32(3110 + i)
			mutateT4Routing(c.Body, i == 1)
			if e = writeWire(ctxTest(t), remote, co, c); e != nil {
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
func TestS6cT4RoutingRoleCheckCloseAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveT4Routing(t, n, p, co, remote, 3201)
		blocked, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		blocked.Header.HopByHop++
		blocked.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		ctx := t4RoutingContext(t, serverOrigin, n.cfg.Origin)
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), routingAnswer()) }()
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
func TestS6cT4RoutingRouteIdentityBudgetAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		body := t4RoutingBody(t)
		id, e := r.Submit(ctxTest(t), body, RequestOptions{Validation: t4RoutingContext(t, n.cfg.Origin, serverOrigin)}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		before := n.Resources()
		body.UserName.Value = "001010123456780"
		body.SMRPSMEA.Value[2] = 0x99
		n.mu.Lock()
		x := n.requests[id]
		reserved := x.charge.Capacity >= 2*x.answerBound
		n.mu.Unlock()
		if !reserved {
			t.Fatal("missing per-attempt answer reserve")
		}
		routeSuspect(t, n, p, a, co)
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		if got := n.Resources(); got.Outstanding != 1 || got.OutboundBytes != before.OutboundBytes {
			t.Fatal(got, before)
		}
		// An old attempt may win; both arrivals retain their own correlation and budget.
		for i, req := range []pdu.Message{first, second} {
			ans, err := co.CompleteAnswer(req, routingAnswer(), serverOrigin)
			if err != nil {
				t.Fatal(err)
			}
			remote := a
			if i == 1 {
				remote = b
			}
			if err = writeWire(ctxTest(t), remote, co, ans); err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || got.SelectedAttempt.Number != 1 || len(got.Attempts) != 2 || !got.Validation.Complete {
					t.Fatal(got)
				}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
