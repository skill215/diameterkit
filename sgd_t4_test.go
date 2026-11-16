package diameter

// TCP/TLS here exercise the transport-neutral runtime only. TS29.337 T4 SCTP
// deployment acceptance is separate and remains pending; these are not T4 peers.
import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func t4RuntimeEvidence(requester, responder pdu.Origin) sgd.T4ChargingEvidence {
	return sgd.T4ChargingEvidence{Requester: requester, Responder: responder, SenderIMSI: "001010123456789", SenderMSISDN: avp.Some("124"), SCAddress: "123", ExternalIdentifier: avp.Some("device-001@example.net")}
}
func t4RuntimeContext(t *testing.T, e sgd.T4ChargingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.T4Charging(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func t4Answer(failed bool) *sgd.OFA {
	a := &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), ExternalIdentifier: avp.Some("device-001@example.net")}
	if failed {
		a.ResultCode.Present = false
		a.ExperimentalResult = avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555})
		a.SMDeliveryFailureCause = avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 4})
	}
	return a
}
func TestSGDT4InboundChargingFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, false, false, 5001)
		// Wire does not identify T4; ordinary zero-context readiness is not T4 proof.
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		if r, e := in.Respond(ctxTest(t), t4Answer(false)); r.Committed || !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(r, e)
		}
		facts := t4RuntimeEvidence(gmscOrigin, n.cfg.Origin)
		wrong := facts
		wrong.SenderIMSI = "001010123456780"
		if r := in.Check(t4RuntimeContext(t, wrong)); r.Complete {
			t.Fatal(r)
		}
		view := in.Message().Body.(*sgd.OFR)
		view.UserIdentifier.UserName = avp.Some("changed")
		view.SCAddress[0] = 0xff
		ctx := t4RuntimeContext(t, facts)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		wrong = facts
		wrong.ExternalIdentifier.Value = "other@example.net"
		if r := in.Check(t4RuntimeContext(t, wrong)); r.Complete {
			t.Fatal("changed frozen mapping")
		}
		if r, e := in.Respond(ctxTest(t), smsAnswer(false, 2001)); r.Committed || !errors.Is(e, validation.ErrMissing) {
			t.Fatal(r, e)
		}
		a := t4Answer(false)
		a.ExternalIdentifier.Value = "other@example.net"
		if r, e := in.Respond(ctxTest(t), a); r.Committed || !errors.Is(e, validation.ErrField) {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(canceled, t4Answer(true)); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		if r, e := in.Respond(ctxTest(t), t4Answer(true)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		ans := duplicateAnswer(t, n, co, remote, req)
		if r := co.ValidateAnswerContext(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
			t.Fatal(r)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("unexpected retry/Ro request", e)
		}
	})
}
func TestSGDT4OutboundMappingAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		facts := t4RuntimeEvidence(n.cfg.Origin, serverOrigin)
		ctx := t4RuntimeContext(t, facts)
		wrong := facts
		wrong.Responder = draOrigin
		if _, e := p.Submit(ctxTest(t), smsBody(false), RequestOptions{Validation: t4RuntimeContext(t, wrong)}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		bad := t4Answer(false)
		bad.ExternalIdentifier.Value = "other@example.net"
		for _, tc := range []struct {
			answer  pdu.Body
			unknown bool
			state   validation.State
		}{
			{t4Answer(false), false, validation.Ready}, {t4Answer(true), false, validation.Ready}, {smsAnswer(false, 2001), false, validation.Rejected}, {bad, false, validation.Rejected}, {t4Answer(false), true, validation.NeedsApplication}, {&pdu.ErrorAnswer{ResultCode: 3004}, true, validation.Ready},
		} {
			e := facts
			if tc.unknown {
				e.ExternalIdentifier.Present = false
			}
			b := smsBody(false)
			id, err := p.Submit(ctxTest(t), b, RequestOptions{Validation: t4RuntimeContext(t, e)}, port)
			if err != nil {
				t.Fatal(err)
			}
			b.(*sgd.OFR).UserIdentifier.UserName = avp.Some("mutated")
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
			if tc.unknown && tc.state == validation.NeedsApplication {
				if r := co.ValidateAnswerContext(req, got.Answer, pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}}, ctx); !r.Complete {
					t.Fatal(r)
				}
			}
		}
		id, e := p.Submit(ctxTest(t), smsBody(false), RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		ans, e := co.CompleteAnswer(req, t4Answer(false), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestSGDT4DuplicateCacheAndIdentityConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, false, false, 5101)
		ctx := t4RuntimeContext(t, t4RuntimeEvidence(gmscOrigin, n.cfg.Origin))
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*sgd.OFR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 5}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), t4Answer(false)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if r := co.ValidateAnswerContext(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
				t.Fatal(r)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*sgd.OFA); a.ExternalIdentifier.Value != "device-001@example.net" {
			t.Fatal(a)
		}
		for i := range 3 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(5110 + i)
			b := conflict.Body.(*sgd.OFR)
			switch i {
			case 0:
				b.UserIdentifier.UserName = avp.Some("001010123456780")
			case 1:
				b.UserIdentifier.MSISDN.Value[0] = 0x43
			case 2:
				b.SCAddress[0] = 0x43
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
func TestSGDT4BudgetAndCheckClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := smsConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, false, false, 5201)
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
		ctx := t4RuntimeContext(t, t4RuntimeEvidence(gmscOrigin, n.cfg.Origin))
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), t4Answer(false)) }()
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
