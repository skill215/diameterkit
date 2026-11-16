package diameter

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

func moRuntimeEvidence(requester, responder pdu.Origin, iface sgd.MOInterface) sgd.MOForwardingEvidence {
	return sgd.MOForwardingEvidence{Interface: iface, Requester: requester, Responder: responder, SenderIMSI: "001010123456789", SenderMSISDN: avp.Some("124"), SCAddress: "123"}
}
func moRuntimeContext(t *testing.T, e sgd.MOForwardingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.MOForwarding(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func moUnknownAnswer() *sgd.OFA {
	return &sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}
}
func TestSGDMOInboundFreezeAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, false, false, 5901)
		if r, e := in.Respond(ctxTest(t), moUnknownAnswer()); r.Committed || !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(r, e)
		}
		facts := moRuntimeEvidence(gmscOrigin, n.cfg.Origin, sgd.MOInterfaceSGd)
		wrong := facts
		wrong.SenderMSISDN = avp.Some("125")
		if r := in.Check(moRuntimeContext(t, wrong)); r.Complete {
			t.Fatal(r)
		}
		// The delivered view cannot alter the private original or its identity binding.
		in.Message().Body.(*sgd.OFR).UserIdentifier.MSISDN.Value[0] = 0xff
		ctx := moRuntimeContext(t, facts)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		wrong = facts
		wrong.Interface = sgd.MOInterfaceGdd
		if r := in.Check(moRuntimeContext(t, wrong)); r.Complete {
			t.Fatal("interface changed after freeze")
		}
		bad := moUnknownAnswer()
		bad.ExternalIdentifier = avp.Some("device@example.net")
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(canceled, moUnknownAnswer()); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), moUnknownAnswer()); !r.Committed || e != nil {
			t.Fatal(r, e)
		}
		a := duplicateAnswer(t, n, co, remote, req)
		if r := co.ValidateAnswerContext(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
			t.Fatal(r)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		// Gdd rejects only the unsupported preparation; an ordinary success can follow.
		in, _ = sendSMS(t, n, p, co, remote, false, false, 5902)
		if r := in.Check(moRuntimeContext(t, wrong)); !r.Complete {
			t.Fatal(r)
		}
		if r, e := in.Respond(ctxTest(t), moUnknownAnswer()); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), smsAnswer(false, 2001)); !r.Committed || e != nil {
			t.Fatal(r, e)
		}
		routeRead(t, co, remote)
	})
}
func TestSGDMOOutboundReportsAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		facts := moRuntimeEvidence(n.cfg.Origin, serverOrigin, sgd.MOInterfaceSGd)
		ctx := moRuntimeContext(t, facts)
		wrong := facts
		wrong.Responder = draOrigin
		if _, err := p.Submit(ctxTest(t), smsBody(false), RequestOptions{Validation: moRuntimeContext(t, wrong)}, port); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		for _, tc := range []struct {
			iface sgd.MOInterface
			state validation.State
			want  error
		}{
			{sgd.MOInterfaceSGd, validation.Ready, nil}, {sgd.MOInterfaceGdd, validation.Rejected, validation.ErrUnsupported}, {0, validation.NeedsApplication, validation.ErrContextRequired},
		} {
			v := validation.Context{}
			if tc.iface != 0 {
				e := facts
				e.Interface = tc.iface
				v = moRuntimeContext(t, e)
			}
			b := smsBody(false)
			id, err := p.Submit(ctxTest(t), b, RequestOptions{Validation: v}, port)
			if err != nil {
				t.Fatal(err)
			}
			b.(*sgd.OFR).UserIdentifier.MSISDN.Value[0] = 0xff
			req := routeRead(t, co, remote)
			a, err := co.CompleteAnswer(req, moUnknownAnswer(), serverOrigin)
			if err != nil {
				t.Fatal(err)
			}
			if err = writeWire(ctxTest(t), remote, co, a); err != nil {
				t.Fatal(err)
			}
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Outcome != AnswerReceived || got.Validation.State != tc.state || !errors.Is(got.Validation.Err(), tc.want) {
				t.Fatal(got)
			}
			if got.Answer.Body.(*sgd.OFA).ExperimentalResult.Value.ExperimentalResultCode != 5001 {
				t.Fatal(got.Answer)
			}
			if tc.iface == 0 {
				if r := co.ValidateAnswerContext(req, got.Answer, pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}}, ctx); !r.Complete {
					t.Fatal(r)
				}
			}
			// Duplicate wire answers do not reopen the completed exchange or retry it.
			if err = writeWire(ctxTest(t), remote, co, a); err != nil {
				t.Fatal(err)
			}
			short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			if _, err = port.Receive(short); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			cancel()
		}
		// Missing original MSISDN is a matched-invalid answer, not a request for
		// extra application facts. Build the remote frame from a separate view.
		body := smsBody(false).(*sgd.OFR)
		body.UserIdentifier.MSISDN.Present = false
		absent := facts
		absent.SenderMSISDN.Present = false
		missingID, err := p.Submit(ctxTest(t), body, RequestOptions{Validation: moRuntimeContext(t, absent)}, port)
		if err != nil {
			t.Fatal(err)
		}
		missingReq := routeRead(t, co, remote)
		changed, err := co.Snapshot(missingReq)
		if err != nil {
			t.Fatal(err)
		}
		changed.Body.(*sgd.OFR).UserIdentifier.MSISDN = avp.Some([]byte{0x21, 0xf4})
		invalid, err := co.CompleteAnswer(changed, moUnknownAnswer(), serverOrigin)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctxTest(t), remote, co, invalid); err != nil {
			t.Fatal(err)
		}
		completion, err := port.Receive(ctxTest(t))
		if err != nil || completion.RequestID != missingID || completion.Result.Answer.Body != nil || !errors.Is(completion.Result.Err, validation.ErrInvalidAnswer) {
			t.Fatal(completion, err)
		}

		id, err := p.Submit(ctxTest(t), smsBody(false), RequestOptions{Validation: ctx}, port)
		if err != nil {
			t.Fatal(err)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, err := co.CompleteAnswer(req, moUnknownAnswer(), serverOrigin)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctxTest(t), remote, co, a); err != nil {
			t.Fatal(err)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err = port.Receive(short); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestSGDMODuplicateCacheAndIdentityConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, false, false, 5101)
		ctx := moRuntimeContext(t, moRuntimeEvidence(gmscOrigin, n.cfg.Origin, sgd.MOInterfaceSGd))
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
		if r, e := in.Respond(ctxTest(t), moUnknownAnswer()); e != nil || !r.Committed {
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
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*sgd.OFA); a.ExperimentalResult.Value.ExperimentalResultCode != 5001 {
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
func TestSGDMOBudgetAndCheckClose(t *testing.T) {
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
		ctx := moRuntimeContext(t, moRuntimeEvidence(gmscOrigin, n.cfg.Origin, sgd.MOInterfaceSGd))
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), moUnknownAnswer()) }()
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
