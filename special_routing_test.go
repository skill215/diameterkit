// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"fmt"
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

func specialBody(mcc bool) *s6c.SRR {
	b := requestBody()
	if mcc {
		b.SMDeliveryNotIntended = avp.Some(s6c.OnlyMCCMNCRequested)
	}
	return b
}
func specialAnswer(mcc bool) *s6c.SRA {
	a := &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("route-token-76")}
	if mcc {
		a.UserName = avp.Some("001010000000000")
	} else {
		a.ServingNode = avp.Some(s6c.ServingNode{MSCNumber: avp.Some([]byte{0x21, 0xf3})})
	}
	return a
}
func specialContext(t *testing.T, mcc, known bool, src, dst pdu.Origin) validation.Context {
	t.Helper()
	var v validation.Context
	var e error
	if mcc {
		v, e = s6c.MCCMNCRouting(s6c.MCCMNCRoutingEvidence{Requester: src, Responder: dst, MCCMNC: avp.Optional[string]{Present: known, Value: "00101"}})
	} else {
		v, e = s6c.RouterRouting(s6c.RouterRoutingEvidence{GMSC: src, Router: dst, HPLMNEnforced: true})
	}
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func eachSpecial(t *testing.T, f func(*testing.T, bool, bool)) {
	t.Helper()
	for _, mcc := range []bool{false, true} {
		t.Run(fmt.Sprintf("mcc=%t", mcc), func(t *testing.T) { eachNetwork(t, func(t *testing.T, tls bool) { f(t, tls, mcc) }) })
	}
}
func receiveSpecial(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, mcc bool, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req, e := co.CompleteRequest(specialBody(mcc), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
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
func TestSpecialRoutingNetwork(t *testing.T) {
	eachSpecial(t, func(t *testing.T, tls, mcc bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		ctx := specialContext(t, mcc, true, n.cfg.Origin, serverOrigin)
		bad := specialContext(t, mcc, true, n.cfg.Origin, draOrigin)
		if _, e := p.Submit(ctxTest(t), specialBody(mcc), RequestOptions{Validation: bad}, port); e == nil {
			t.Fatal("wrong endpoints")
		}
		if mcc {
			if _, e := p.Submit(ctxTest(t), specialBody(mcc), RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
				t.Fatal(e)
			}
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		for _, body := range []pdu.Body{specialAnswer(mcc), &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, sraFailure(5550), &pdu.ErrorAnswer{ResultCode: 3004}} {
			b := specialBody(mcc)
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			b.MSISDN.Value[0] ^= 255
			b.SMDeliveryNotIntended.Present = !mcc
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
		}
		if mcc {
			id, e := p.Submit(ctxTest(t), specialBody(true), RequestOptions{Validation: specialContext(t, true, false, n.cfg.Origin, serverOrigin)}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, specialAnswer(true), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Validation.State != validation.NeedsApplication {
				t.Fatal(got)
			}
			if v := co.ValidateAnswerContext(req, got.Answer, pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}}, ctx); !v.Complete {
				t.Fatal(v)
			}
		}
		id, e := p.Submit(ctxTest(t), specialBody(mcc), RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		ans, e := co.CompleteAnswer(req, specialAnswer(mcc), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late completion", e)
		}
		waitFor(t, func() bool { return n.Resources().OutboundBytes == 0 })
	})
}
func TestSpecialRoutingInboundCache(t *testing.T) {
	eachSpecial(t, func(t *testing.T, tls, mcc bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveSpecial(t, n, p, co, remote, mcc, 7601)
		wrong := specialContext(t, mcc, true, draOrigin, n.cfg.Origin)
		if v := in.Check(wrong); v.Complete {
			t.Fatal(v)
		}
		// Owned public view cannot turn the actual original query into delivery or vice versa.
		in.Message().Body.(*s6c.SRR).SMDeliveryNotIntended.Present = !mcc
		ctx := specialContext(t, mcc, true, serverOrigin, n.cfg.Origin)
		if v := in.Check(ctx); !v.Complete {
			t.Fatal(v)
		}
		if v := in.Check(specialContext(t, !mcc, true, serverOrigin, n.cfg.Origin)); v.Complete {
			t.Fatal("changed purpose after freeze")
		}
		bad := specialAnswer(mcc)
		if mcc {
			bad.UserName.Value = "999990000000000"
		} else {
			bad.ServingNode.Present = false
		}
		if r, e := in.Respond(ctxTest(t), bad); e == nil || r.Committed {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(canceled, specialAnswer(mcc)); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		dup := otherArrival(t, co, req, 7602)
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), specialAnswer(mcc)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if v := co.ValidateAnswerContext(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !v.Complete {
				t.Fatal(v)
			}
		}
		dup.Header.HopByHop++
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		a := duplicateAnswer(t, n, co, remote, dup)
		if a.Body.(*s6c.SRA).UserName.Value != specialAnswer(mcc).UserName.Value {
			t.Fatal("cache identity")
		}
		conflict := otherArrival(t, co, req, 7604)
		conflict.Body.(*s6c.SRR).SMDeliveryNotIntended = avp.Optional[int32]{Present: !mcc, Value: s6c.OnlyMCCMNCRequested}
		if e := writeWire(ctxTest(t), remote, co, conflict); e != nil {
			t.Fatal(e)
		}
		if a, ok := routeRead(t, co, remote).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 5012 {
			t.Fatal(a)
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}
func TestSpecialRoutingBudgetClose(t *testing.T) {
	eachSpecial(t, func(t *testing.T, tls, mcc bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveSpecial(t, n, p, co, remote, mcc, 7701)
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
		ctx := specialContext(t, mcc, true, serverOrigin, n.cfg.Origin)
		if v := in.Check(ctx); !v.Complete {
			t.Fatal(v)
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), specialAnswer(mcc)) }()
		go func() { defer wg.Done(); <-start; _ = n.Close() }()
		close(start)
		wg.Wait()
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
