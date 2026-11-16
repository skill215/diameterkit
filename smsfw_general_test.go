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
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func generalBody(t *testing.T) *legacy420.CCR {
	t.Helper()
	return legacy420.NewDirectDebit(legacy420.GeneralInput{SessionID: "ctf.example;general;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", Timestamp: 3908988800, ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 1, ProtocolID: []byte{0, 255}, OriginatorMSISDN: "15551234567"})
}
func generalAnswer(code uint32) *legacy420.CCA {
	a := refundAnswer(code)
	if code == 2001 {
		a.CostInformation = avp.Some(legacy420.CostInformation{UnitValue: legacy420.CostValue{ValueDigits: -123, Exponent: avp.Some(int32(-2))}, CurrencyCode: 840})
		a.GrantedServiceUnit = avp.Some(legacy420.GrantedUnit{CCServiceSpecificUnits: 0})
	}
	return a
}
func TestSMSFWGeneralNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, body := range []pdu.Body{generalAnswer(2001), generalAnswer(4012), generalAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			q := generalBody(t)
			id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.ServiceInformation.MMSInformation.Value.OriginatorAddress.AddressData[0] = 99
			r := routeRead(t, co, remote)
			if r.Body.(*legacy420.CCR).ServiceInformation.MMSInformation.Value.OriginatorAddress.AddressData[0] != '1' {
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
		for _, value := range []uint32{0, 1} {
			q := generalBody(t)
			q.RequestedAction = 2
			id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.RequestedAction = 0 // frozen request must still require a balance result
			r := routeRead(t, co, remote)
			a, e := co.CompleteAnswer(r, &legacy420.CCA{AuthApplicationID: 4, CCRequestType: 4, ResultCode: 2001, CheckBalanceResult: avp.Some(value)}, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete || got.Answer.Body.(*legacy420.CCA).CheckBalanceResult.Value != value {
				t.Fatal(got)
			}
		}
		id, e := p.Submit(ctxTest(t), refundBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		a, e := co.CompleteAnswer(r, generalAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); got.Outcome != AnswerReceived {
			t.Fatal(got)
		}

		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = remote.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic charging/retry", e)
		}
	})
}
func TestSMSFWGeneralInboundPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, body := range []pdu.Body{generalAnswer(2001), generalAnswer(4012), generalAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			r, e := co.CompleteRequest(generalBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(500 + i), EndToEnd: uint32(900 + i)})
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
			bad := generalAnswer(2001)
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
func TestSMSFWGeneralCancelLateInvalidAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), generalBody(t), RequestOptions{}, port)
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
		a, e := co.CompleteAnswer(r, generalAnswer(2001), serverOrigin)
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
		id, e = p.Submit(ctxTest(t), generalBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = co.CompleteAnswer(r, generalAnswer(2001), serverOrigin)
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
		id, e = p.Submit(ctxTest(t), generalBody(t), RequestOptions{}, port)
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

func TestSMSFWGeneralBalanceOriginalAction(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, refundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		q := generalBody(t)
		q.RequestedAction = 2
		id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		// A well-formed direct-debit CCA cannot answer an original balance request.
		// Server-side mutation here deliberately avoids the correct CompleteAnswer guard.
		r.Body.(*legacy420.CCR).RequestedAction = 0
		a, e := co.CompleteAnswer(r, generalAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })

		q = generalBody(t)
		q.RequestedAction = 2
		r, e = co.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 800, EndToEnd: 801})
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
		in.Message().Body.(*legacy420.CCR).RequestedAction = 0
		if got, e := in.Respond(ctxTest(t), generalAnswer(2001)); e == nil || got.Committed {
			t.Fatal("missing balance result committed", got, e)
		}
		reply := refundAnswer(2001)
		reply.CheckBalanceResult = avp.Some(uint32(1))
		if got, e := in.Respond(ctxTest(t), reply); e != nil || !got.Committed {
			t.Fatal(got, e)
		}
		a = routeRead(t, co, remote)
		if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
