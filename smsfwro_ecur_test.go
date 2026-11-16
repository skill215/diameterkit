package diameter

import (
	"context"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func legacyECURBody(term bool) *smsfwro.CCR {
	in := smsfwro.ECURInput{SessionID: "ctf.example;legacy-ecur;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{0, 255}}
	if term {
		return smsfwro.NewTermination(smsfwro.TerminationInput{ECURInput: in, UsedServiceSpecificUnits: 1, Status: []byte{0, 255}, DischargeTime: 3908988830})
	}
	return smsfwro.NewInitial(in)
}
func legacyECURAnswer(term bool, code uint32) *smsfwro.CCA {
	a := &smsfwro.CCA{AuthApplicationID: 4, CCRequestType: 1, ResultCode: code}
	if term {
		a.CCRequestType = 3
		a.CCRequestNumber = 1
	} else if code == 2001 {
		a.GrantedServiceUnit = avp.Some(smsfwro.ServiceUnit{CCServiceSpecificUnits: 1})
	}
	return a
}
func TestSMSFWECURNetworkCycleAndNoAutomaticActions(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, termCode := range []uint32{2001, 5002} {
			body := legacyECURBody(false)
			body.SessionID += fmt.Sprint(termCode)
			id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			initial := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(initial, legacyECURAnswer(false, 2001), serverOrigin)
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
			// Caller explicitly constructs and checks the second operation after the
			// first has completed. No Node-owned business session or automatic TERM.
			q := legacyECURBody(true)
			ib := initial.Body.(*smsfwro.CCR)
			q.SessionID = ib.SessionID
			q.OriginHost = ib.OriginHost
			q.OriginRealm = ib.OriginRealm
			q.DestinationHost = ib.DestinationHost
			q.DestinationRealm = ib.DestinationRealm
			if e = smsfwro.CheckECURContinuation(ib, q); e != nil {
				t.Fatal(e)
			}
			id, e = p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.MultipleServicesCreditControl.UsedServiceUnit.Value.CCServiceSpecificUnits = 77
			term := routeRead(t, co, remote)
			if term.Header.HopByHop == initial.Header.HopByHop || term.Header.EndToEnd == initial.Header.EndToEnd {
				t.Fatal("exchange identity reused")
			}
			if term.Body.(*smsfwro.CCR).MultipleServicesCreditControl.UsedServiceUnit.Value.CCServiceSpecificUnits != 1 {
				t.Fatal("caller alias")
			}
			if e = smsfwro.CheckECURContinuation(ib, term.Body.(*smsfwro.CCR)); e != nil {
				t.Fatal(e)
			}
			ans, e = co.CompleteAnswer(term, legacyECURAnswer(true, termCode), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got = routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete || got.Answer.Body.(*smsfwro.CCA).ResultCode != termCode {
				t.Fatal(got)
			}

		}
		// Same static Bundle/connection still handles the earlier IEC dialect;
		// no per-request dictionary selection and no additional App4 registration.
		id, e := p.Submit(ctxTest(t), legacyBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		iec := routeRead(t, co, remote)
		ans, e := co.CompleteAnswer(iec, legacyAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); got.Outcome != AnswerReceived {
			t.Fatal(got)
		}

		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = remote.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic INITIAL/refund/retry", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestSMSFWECURInboundPairsAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, tc := range []struct {
			term bool
			body pdu.Body
		}{{false, legacyECURAnswer(false, 2001)}, {false, legacyECURAnswer(false, 4012)}, {false, &pdu.ErrorAnswer{ResultCode: 3004}}, {true, legacyECURAnswer(true, 2001)}, {true, legacyECURAnswer(true, 5002)}, {true, &pdu.ErrorAnswer{ResultCode: 3004}}} {
			req, e := co.CompleteRequest(legacyECURBody(tc.term), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(1100 + i), EndToEnd: uint32(2100 + i)})
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
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			in.Message().Body.(*smsfwro.CCR).CCRequestNumber = 22
			if tc.term {
				wrong := legacyECURAnswer(false, 2001)
				if got, e := in.Respond(ctxTest(t), wrong); got.Committed || !errors.Is(e, validation.ErrInvalidAnswer) {
					t.Fatal(got, e)
				}
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			if got, e := in.Respond(canceled, tc.body); got.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(got, e)
			}
			if got, e := in.Respond(ctxTest(t), tc.body); e != nil || !got.Committed {
				t.Fatal(got, e)
			}
			ans := routeRead(t, co, remote)
			if e = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestSMSFWECURCancelLateAndInvalidAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), legacyECURBody(true), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		ans, e := co.CompleteAnswer(req, legacyECURAnswer(true, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late second completion", e)
		}
		id, e = p.Submit(ctxTest(t), legacyECURBody(false), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req = routeRead(t, co, remote)
		altered, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		r := altered.Body.(*smsfwro.CCR)
		replacement := legacyECURBody(true)
		replacement.OriginHost = r.OriginHost
		replacement.OriginRealm = r.OriginRealm
		replacement.DestinationHost = r.DestinationHost
		replacement.DestinationRealm = r.DestinationRealm
		altered.Body = replacement
		ans, e = co.CompleteAnswer(altered, legacyECURAnswer(true, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		got = routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
