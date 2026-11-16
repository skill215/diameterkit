package diameter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func routingRequest(t *testing.T) *s6c.SRR {
	t.Helper()
	b := requestBody()
	b.SRRFlags = avp.Some(uint32(5)) // GPRS and SMS-delivery Single-Attempt, not a transport policy.
	sme, err := s6c.EncodeSMEA("12345")
	if err != nil {
		t.Fatal(err)
	}
	b.SMRPSMEA = avp.Some(sme)
	b.SMRPMTI = avp.Some(uint32(0))
	return b
}
func routingAnswer() *s6c.SRA {
	return &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"),
		ServingNode:           avp.Some(s6c.ServingNode{SGSNNumber: avp.Some([]byte{0x21, 0xf3})}),
		AdditionalServingNode: avp.Some(s6c.AdditionalServingNode{MSCNumber: avp.Some([]byte{0x21, 0xf4})}),
		LMSI:                  avp.Some([]byte{1, 2, 3, 4}), MWDStatus: avp.Some(uint32(0)),
		UserIdentifier: avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf6}})}
}
func TestS6cRoutingPrivateRequestAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for _, permitted := range []bool{false, true} {
			req := inboundRequest(t, n, co, 701)
			if permitted {
				req.Header.HopByHop++
				req.Header.EndToEnd++
				b := req.Body.(*s6c.SRR)
				r := routingRequest(t)
				b.SRRFlags = r.SRRFlags
				b.SMRPSMEA = r.SMRPSMEA
			}
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			// Neither adding nor removing GPRS in the public view changes the original evidence.
			in.Message().Body.(*s6c.SRR).SRRFlags = avp.Some(uint32(1))
			if permitted {
				in.Message().Body.(*s6c.SRR).SRRFlags = avp.Optional[uint32]{}
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, e = in.Respond(ctx, routingAnswer()); e == nil {
				t.Fatal("canceled preparation committed")
			}
			r, e := in.Respond(ctxTest(t), routingAnswer())
			if !permitted {
				if !errors.Is(e, validation.ErrInvalidAnswer) || r.Committed {
					t.Fatal(r, e)
				}
				good := routingAnswer()
				good.AdditionalServingNode.Present = false
				good.LMSI.Present = false
				r, e = in.Respond(ctxTest(t), good)
			}
			if e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			a := duplicateAnswer(t, n, co, remote, req).Body.(*s6c.SRA)
			if a.AdditionalServingNode.Present != permitted || !a.MWDStatus.Present || a.MWDStatus.Value != 0 {
				t.Fatal(a)
			}
			waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		}
	})
}
func TestS6cRoutingRouteIdentityBudgetAndLateAnswer(t *testing.T) {
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
func TestS6cRoutingDuplicateCacheAndContentConflict(t *testing.T) {
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
		if result, err := in.Respond(ctxTest(t), routingAnswer()); err != nil || !result.Committed {
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
		if !cached.AdditionalServingNode.Present || !cached.LMSI.Present || !cached.MWDStatus.Present {
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
func TestS6cRoutingCancelThenLateAnswer(t *testing.T) {
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
		ans, e := co.CompleteAnswer(req, routingAnswer(), serverOrigin)
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
