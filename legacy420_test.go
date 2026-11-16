package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func refundBody(t *testing.T) *legacy420.CCR {
	t.Helper()
	r, e := legacy420.NewRefund(legacy420.RefundInput{SessionID: "ctf.example;refund;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", Timestamp: 3908988800, ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 1, ProtocolID: []byte{0, 255}, RefundInformation: avp.Some([]byte{0, 255, 128, 0, 1})})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func refundAnswer(code uint32) *legacy420.CCA {
	return &legacy420.CCA{AuthApplicationID: 4, CCRequestType: 4, ResultCode: code}
}
func refundConfig(t *testing.T) NodeConfig {
	c := roConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := legacy420.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(bp, sp, p)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestLegacy420RefundNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, body := range []pdu.Body{refundAnswer(2001), refundAnswer(4012), refundAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			q := refundBody(t)
			id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.RefundInformation.Value[0] = 99
			r := routeRead(t, co, remote)
			if r.Body.(*legacy420.CCR).RefundInformation.Value[0] != 0 {
				t.Fatal("submit alias")
			}
			a, e := co.CompleteAnswer(r, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e := remote.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic charging/retry", e)
		}
	})
}
func TestLegacy420RefundInboundPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, body := range []pdu.Body{refundAnswer(2001), refundAnswer(4012), refundAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			r, e := co.CompleteRequest(refundBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(500 + i), EndToEnd: uint32(900 + i)})
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
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			in.Message().Body.(*legacy420.CCR).CCRequestNumber = 77
			bad := refundAnswer(2001)
			bad.CCRequestNumber = 1
			if got, e := in.Respond(ctxTest(t), bad); e == nil || got.Committed {
				t.Fatal("bad preparation committed", got, e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if got, e := in.Respond(ctx, body); got.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(got, e)
			}
			if got, e := in.Respond(ctxTest(t), body); e != nil || !got.Committed {
				t.Fatal(got, e)
			}
			a := routeRead(t, co, remote)
			if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestLegacy420RefundCancelLateInvalidAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), refundBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		a, e := co.CompleteAnswer(r, refundAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		id, e = p.Submit(ctxTest(t), refundBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = co.CompleteAnswer(r, refundAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		// Use a complete wire projection: static send rejects wrong request numbers,
		// receive must first decode and then reject against the original request.
		raw, e := co.Append(nil, a)
		if e != nil {
			t.Fatal(e)
		}
		raw = replaceAVP32(t, raw, 415, 1)
		if result := remote.WriteFrame(ctxTest(t), raw); result.Err != nil {
			t.Fatal(result.Err)
		}
		got = routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		id, e = p.Submit(ctxTest(t), refundBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, remote)
		n.Close()
		got = routeCompletion(t, n, port, id)
		if got.Err == nil {
			t.Fatal("close succeeded request", got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
