// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func ipsmAnswer() *s6c.SRA {
	return &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"), ServingNode: avp.Some(s6c.ServingNode{IPSMGWNumber: avp.Some([]byte{0x21, 0xf3}), IPSMGWName: avp.Some("isg.example"), IPSMGWSBISupportIndicator: avp.Some(uint32(0))})}
}
func TestS6cIPSMGWInboundPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 6301)
		for _, edit := range []func(*s6c.SRA){func(a *s6c.SRA) { a.ServingNode.Value.IPSMGWRealm = avp.Some("example.net") }, func(a *s6c.SRA) { a.ServingNode.Value.IPSMGWSBISupportIndicator.Value = 2 }, func(a *s6c.SRA) {
			a.AdditionalServingNode = avp.Some(s6c.AdditionalServingNode{IPSMGWNumber: avp.Some([]byte{0x21})})
		}} {
			bad := ipsmAnswer()
			edit(bad)
			if r, err := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(r, err)
			}
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, err := in.Respond(canceled, ipsmAnswer()); r.Committed || !errors.Is(err, context.Canceled) {
			t.Fatal(r, err)
		}
		answer := ipsmAnswer()
		if r, err := in.Respond(ctxTest(t), answer); !r.Committed || err != nil {
			t.Fatal(r, err)
		}
		answer.ServingNode.Value.IPSMGWNumber.Value[0] = 0xff
		got := duplicateAnswer(t, n, co, remote, req)
		if p := co.ValidateAnswerContext(req, got, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, validation.Context{}); !p.Complete {
			t.Fatal(p)
		}
		if b := got.Body.(*s6c.SRA).ServingNode.Value; b.IPSMGWNumber.Value[0] != 0x21 || b.IPSMGWRealm.Present || !b.IPSMGWSBISupportIndicator.Present {
			t.Fatal(b)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
	})
}

// Add a well-framed leaf without passing the library's send selection; models a
// remote answer outside our selected semantics. Existing field counts stay valid.
func ipsmRemoteField(t *testing.T, raw []byte, code uint32, flags byte, data []byte) []byte {
	t.Helper()
	leaf := make([]byte, (12+len(data)+3)&^3)
	binary.BigEndian.PutUint32(leaf, code)
	leaf[4] = flags
	wire.Put24(leaf[5:8], uint32(12+len(data)))
	binary.BigEndian.PutUint32(leaf[8:12], 10415)
	copy(leaf[12:], data)
	for off := 20; off < len(raw); {
		n := int(binary.BigEndian.Uint32(raw[off+4:off+8]) & 0xffffff)
		if binary.BigEndian.Uint32(raw[off:]) == 2401 {
			end := off + (n+3)&^3
			out := append([]byte{}, raw[:end]...)
			out = append(out, leaf...)
			out = append(out, raw[end:]...)
			wire.Put24(out[off+5:off+8], uint32(n+len(leaf)))
			wire.Put24(out[1:4], uint32(len(out)))
			return out
		}
		off += (n + 3) &^ 3
	}
	t.Fatal("missing node")
	return nil
}
func TestS6cIPSMGWReceivedReportAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, realm := range []bool{true, false} {
			id, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
			if err != nil {
				t.Fatal(err)
			}
			req := routeRead(t, co, remote)
			body := ipsmAnswer()
			body.ServingNode.Value.IPSMGWSBISupportIndicator.Present = false
			ans, err := co.CompleteAnswer(req, body, serverOrigin)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := co.Append(nil, ans)
			if err != nil {
				t.Fatal(err)
			}
			if realm {
				raw = ipsmRemoteField(t, raw, 3112, 192, []byte("example.net"))
			} else {
				raw = ipsmRemoteField(t, raw, 3348, 128, []byte{0, 0, 0, 2})
			}
			sendRejectRaw(t, remote, raw)
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Outcome != AnswerReceived || !errors.Is(got.Validation.Err(), validation.ErrUnsupported) {
				t.Fatal(got)
			}
			node := got.Answer.Body.(*s6c.SRA).ServingNode.Value
			if realm && node.IPSMGWRealm.Value != "example.net" || !realm && node.IPSMGWSBISupportIndicator.Value != 2 {
				t.Fatal(node)
			}
			sendRejectRaw(t, remote, raw)
			short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			_, err = port.Receive(short)
			cancel()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			// Address data never modifies configured peers or starts an automatic request.
			if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
				t.Fatal(n.Resources())
			}
		}
	})
}
func TestS6cIPSMGWRouteIdentityBudgetAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		body := routingRequest(t)
		id, e := r.Submit(ctxTest(t), body, RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		before := n.Resources()
		body.SRRFlags.Value = 0
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
			ans, err := co.CompleteAnswer(req, ipsmAnswer(), serverOrigin)
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
func TestS6cIPSMGWDuplicateCacheAndContentConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		req := inboundRequest(t, n, co, 801)
		r := routingRequest(t)
		body := req.Body.(*s6c.SRR)
		body.SRRFlags = r.SRRFlags
		body.SMRPSMEA = r.SMRPSMEA
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		dup := otherArrival(t, co, req, 802)
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if result, err := in.Respond(ctxTest(t), ipsmAnswer()); err != nil || !result.Committed {
			t.Fatal(result, err)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if e = co.ValidateAnswer(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		late := otherArrival(t, co, req, 803)
		if e = writeWire(ctxTest(t), remote, co, late); e != nil {
			t.Fatal(e)
		}
		cached := duplicateAnswer(t, n, co, remote, late).Body.(*s6c.SRA)
		if !cached.ServingNode.Value.IPSMGWName.Present || !cached.ServingNode.Value.IPSMGWSBISupportIndicator.Present || cached.ServingNode.Value.IPSMGWSBISupportIndicator.Value != 0 {
			t.Fatal(cached)
		}
		for i := range 2 {
			conflict := otherArrival(t, co, req, uint32(804+i))
			c := conflict.Body.(*s6c.SRR)
			if i == 0 {
				c.SRRFlags.Value = 1
			} else {
				c.SMRPSMEA.Value[2] = 0x43
			}
			if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
				t.Fatal(e)
			}
			ans := routeRead(t, co, remote)
			if ans.Header.Flags&pdu.ErrorFlag == 0 {
				t.Fatal("content conflict reused success", ans)
			}
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}
func TestS6cIPSMGWCancelThenLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), routingRequest(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if got.Outcome == AnswerReceived {
			t.Fatal(got)
		}
		ans, e := co.CompleteAnswer(req, ipsmAnswer(), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestS6cIPSMGWBudgetAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, _ := receiveInbound(t, n, p, remote, co, 3971)
		blocked := inboundRequest(t, n, co, 3972)
		if e := writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		ctx := sraRole(t, s6c.RoutingFromGMSC, avp.Optional[bool]{}, serverOrigin, n.cfg.Origin)
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), ipsmAnswer()) }()
		close(start)
		_ = n.Close()
		wg.Wait()
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
