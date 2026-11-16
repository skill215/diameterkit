package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func combinedConfig(t *testing.T) NodeConfig {
	c := legacyConfig(t)
	b, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := combined.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(b, p)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func combinedBody(t *testing.T, kind string) *combined.CCR {
	t.Helper()
	ec := combined.ECURInput{SessionID: "ctf.example;combined;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{0, 255}}
	switch kind {
	case "initial":
		return combined.NewInitial(ec)
	case "term":
		return combined.NewTermination(combined.TerminationInput{ECURInput: ec, UsedServiceSpecificUnits: 1, Status: []byte{0, 255}, DischargeTime: 3908988830})
	case "iec":
		return combined.NewIEC(combined.IECInput{SessionID: ec.SessionID, ServiceContextID: ec.ServiceContextID, SubscriptionMSISDN: ec.SubscriptionMSISDN, SubscriptionIMSI: ec.SubscriptionIMSI, SMSCAddress: ec.SMSCAddress, OriginatorSCCPAddress: "vmsc", RecipientSCCPAddress: ec.RecipientSCCPAddress, OriginatorMSISDN: ec.OriginatorMSISDN, RecipientMSISDN: ec.RecipientMSISDN, SMSKey: ec.SMSKey, Timestamp: ec.Timestamp, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{0, 255}})
	case "refund":
		q, e := combined.NewRefund(combined.RefundInput{SessionID: ec.SessionID, ServiceContextID: ec.ServiceContextID, SubscriptionMSISDN: ec.SubscriptionMSISDN, SubscriptionIMSI: ec.SubscriptionIMSI, Timestamp: ec.Timestamp, ValueDigits: -5, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 1, RefundInformation: avp.Some([]byte{0, 255})})
		if e != nil {
			t.Fatal(e)
		}
		return q
	default:
		v := combined.GeneralInput{SessionID: ec.SessionID, ServiceContextID: ec.ServiceContextID, SubscriptionMSISDN: ec.SubscriptionMSISDN, SubscriptionIMSI: ec.SubscriptionIMSI, Timestamp: ec.Timestamp, ValueDigits: -5, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 1, OriginatorMSISDN: ec.OriginatorMSISDN}
		if kind == "balance" {
			return combined.NewBalanceCheck(v)
		}
		return combined.NewDirectDebit(v)
	}
}
func combinedAnswer(q *combined.CCR, code uint32) *combined.CCA {
	a := &combined.CCA{AuthApplicationID: 4, CCRequestType: q.CCRequestType, CCRequestNumber: q.CCRequestNumber, ResultCode: code}
	if code == 2001 && q.RequestedAction.Present && q.RequestedAction.Value == 2 {
		a.CheckBalanceResult = avp.Some(uint32(1))
	}
	return a
}
func TestCombinedNetworkAllTemplates(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
			for _, code := range []uint32{2001, 4012, 5012, 3004} {
				q := combinedBody(t, kind)
				id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
				if e != nil {
					t.Fatal(kind, e)
				}
				q.SubscriptionID[0].SubscriptionIDData[0] = '9'
				q.RequestedAction = avp.Some(uint32(99))
				r := routeRead(t, co, remote)
				b := r.Body.(*combined.CCR)
				if b.SubscriptionID[0].SubscriptionIDData[0] != '1' {
					t.Fatal("submit alias")
				}
				var body pdu.Body = combinedAnswer(b, code)
				if code == 3004 {
					body = &pdu.ErrorAnswer{ResultCode: code}
				}
				a, e := co.CompleteAnswer(r, body, serverOrigin)
				if e != nil {
					t.Fatal(kind, code, e)
				}
				if e = writeWire(ctxTest(t), remote, co, a); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || !got.Validation.Complete {
					t.Fatal(kind, code, got)
				}
			}
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(quiet, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic business request", e)
		}
	})
}
func TestCombinedNetworkInboundOriginalBranch(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		for i, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
			r, e := co.CompleteRequest(combinedBody(t, kind), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(200 + i), EndToEnd: uint32(300 + i)})
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
			// Mutating the delivered type discriminator cannot change answer authority.
			in.Message().Body.(*combined.CCR).MultipleServicesIndicator = avp.Optional[int32]{}
			in.Message().Body.(*combined.CCR).RequestedAction = avp.Some(uint32(0))
			bad := combinedAnswer(r.Body.(*combined.CCR), 2001)
			bad.CheckBalanceResult = avp.Some(uint32(0))
			if kind == "balance" {
				bad.CheckBalanceResult = avp.Optional[uint32]{}
			}
			if got, e := in.Respond(ctxTest(t), bad); e == nil || got.Committed {
				t.Fatal(kind, "wrong branch accepted", e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if got, e := in.Respond(ctx, combinedAnswer(r.Body.(*combined.CCR), 2001)); got.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(got, e)
			}
			if got, e := in.Respond(ctxTest(t), combinedAnswer(r.Body.(*combined.CCR), 2001)); e != nil || !got.Committed {
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
func TestCombinedNetworkCancelInvalidAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), combinedBody(t, "refund"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		a, e := co.CompleteAnswer(r, combinedAnswer(r.Body.(*combined.CCR), 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		id, e = p.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = co.CompleteAnswer(r, combinedAnswer(r.Body.(*combined.CCR), 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := co.Append(nil, a)
		if e != nil {
			t.Fatal(e)
		}
		raw = replaceAVP32(t, raw, 415, 1)
		if v := remote.WriteFrame(ctxTest(t), raw); v.Err != nil {
			t.Fatal(v.Err)
		}
		if got := routeCompletion(t, n, port, id); !errors.Is(got.Err, validation.ErrInvalidAnswer) {
			t.Fatal(got)
		}
		id, e = p.Submit(ctxTest(t), combinedBody(t, "initial"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, remote)
		n.Close()
		if got := routeCompletion(t, n, port, id); got.Err == nil {
			t.Fatal(got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("duplicate completion", e)
		}
	})
}
