package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func multiReportBody(mask int) *s6c.RDR {
	b := reportBody()
	b.SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
	if mask&1 != 0 {
		b.SMDeliveryOutcome.MMESMDeliveryOutcome = avp.Some(s6c.MMESMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(13))})
	}
	if mask&2 != 0 {
		b.SMDeliveryOutcome.MSCSMDeliveryOutcome = avp.Some(s6c.MSCSMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(0))})
	}
	if mask&4 != 0 {
		b.SMDeliveryOutcome.SGSNSMDeliveryOutcome = avp.Some(s6c.SGSNSMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(5))})
	}
	return b
}
func multiReportRequest(t *testing.T, n *Node, co *codec.Codec, mask int, hbh uint32) pdu.Message {
	t.Helper()
	b := multiReportBody(mask)
	req, e := co.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 1000})
	if e != nil {
		t.Fatal(e)
	}
	return req
}
func multiReportReceive(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, mask int, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req := multiReportRequest(t, n, co, mask, hbh)
	if e := writeWire(ctxTest(t), remote, co, req); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, req
}
func multiReportAnswer(result uint32) *s6c.RDA {
	b := &s6c.RDA{AuthSessionState: 1}
	if result == 2001 {
		b.ResultCode = avp.Some(result)
		b.UserIdentifier = avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf6}})
	} else {
		b.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: result})
	}
	return b
}
func TestS6cMultiReportExchanges(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, mask := range []int{1, 2, 4, 5, 6} {
			for _, code := range []uint32{2001, 5001, 5558} {
				body := multiReportBody(mask)
				id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
				if e != nil {
					t.Fatal(mask, e)
				}
				// Caller changes cannot rewrite the admitted report.
				body.SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
				req := routeRead(t, co, remote)
				if o := req.Body.(*s6c.RDR).SMDeliveryOutcome; !o.MMESMDeliveryOutcome.Present && !o.MSCSMDeliveryOutcome.Present && !o.SGSNSMDeliveryOutcome.Present {
					t.Fatal("request mutation escaped")
				}
				ans, e := co.CompleteAnswer(req, multiReportAnswer(code), serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || got.Err != nil || !got.Validation.Complete {
					t.Fatal(got)
				}
				b := got.Answer.Body.(*s6c.RDA)
				if code == 2001 {
					if !b.UserIdentifier.Present {
						t.Fatal(b)
					}
				} else if b.ExperimentalResult.Value.ExperimentalResultCode != code || b.ResultCode.Present {
					t.Fatal(b)
				}
				if n.Resources().OutboundBytes != 0 {
					t.Fatal(n.Resources())
				}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic alert or retry", e)
		}
	})
}
func TestS6cMultiReportInboundBudgetAndRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := reportAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := multiReportReceive(t, n, p, co, remote, 6, 900)
		before := n.Resources()
		if before.InboundOutstanding != 1 || before.InboundBytes == 0 {
			t.Fatal(before)
		}
		blocked := multiReportRequest(t, n, co, 4, 901)
		if e := writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); a.ResultCode != 3004 {
			t.Fatal(a)
		}
		in.Message().Body.(*s6c.RDR).SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
		if !in.Check(validation.Context{}).Complete {
			t.Fatal(in.Validation())
		}
		bad := multiReportAnswer(5558)
		bad.UserIdentifier = multiReportAnswer(2001).UserIdentifier
		if r, e := in.Respond(ctxTest(t), bad); e == nil || r.Committed {
			t.Fatal(r, e)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(ctx, multiReportAnswer(2001)); !errors.Is(e, context.Canceled) || r.Committed {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), multiReportAnswer(2001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		in, req = multiReportReceive(t, n, p, co, remote, 4, 902)
		if r, e := in.Respond(ctxTest(t), multiReportAnswer(5001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
	})
}
func TestS6cMultiReportDuplicateParents(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := multiReportReceive(t, n, p, co, remote, 6, 920)
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*s6c.RDR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{4, 5}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), multiReportAnswer(5558)); e != nil || !r.Committed {
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
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*s6c.RDA); a.ExperimentalResult.Value.ExperimentalResultCode != 5558 {
			t.Fatal(a)
		}
		for i := range 2 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(930 + i)
			o := &conflict.Body.(*s6c.RDR).SMDeliveryOutcome
			if i == 0 {
				o.MSCSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Present = false
			} else {
				o.SGSNSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value = 6
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
func TestS6cMultiReportCancelClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), multiReportBody(6), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); got.Outcome == AnswerReceived {
			t.Fatal(got)
		}
		ans, e := co.CompleteAnswer(req, multiReportAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(ctx)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late completion", e)
		}
		in, _ := multiReportReceive(t, n, p, co, remote, 5, 940)
		start := make(chan struct{})
		results := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r, _ := in.Respond(ctxTest(t), multiReportAnswer(2001))
				results <- r
			}()
		}
		close(start)
		_ = n.Close()
		wg.Wait()
		close(results)
		commits := 0
		for r := range results {
			if r.Committed {
				commits++
			}
		}
		if commits > 1 {
			t.Fatal("multiple commits")
		}
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
