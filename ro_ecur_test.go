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
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func roECURBody(term bool) *ro.CCR {
	r := roBody()
	r.SessionID = "ctf.example;ecur;1"
	r.CCRequestType = ro.InitialRequest
	r.RequestedAction = avp.Optional[uint32]{}
	if term {
		r.CCRequestType = ro.TerminationRequest
		r.CCRequestNumber = 1
		s := &r.MultipleServicesCreditControl[0]
		s.RequestedServiceUnit = avp.Optional[ro.ServiceUnit]{}
		s.UsedServiceUnit = []ro.ServiceUnit{{CCServiceSpecificUnits: avp.Some(uint64(1))}}
		s.ReportingReason = []uint32{ro.ReportingFinal}
	}
	return r
}
func roECURAnswer(term bool, code uint32) *ro.CCA {
	a := roAnswer(code)
	a.CCRequestType = ro.InitialRequest
	if term {
		a.CCRequestType = ro.TerminationRequest
		a.CCRequestNumber = 1
		a.MultipleServicesCreditControl = nil
	} else if len(a.MultipleServicesCreditControl) > 0 {
		a.MultipleServicesCreditControl[0].RefundInformation = avp.Optional[[]byte]{}
	}
	return a
}
func TestRoECURNetworkCycleAndNoAutomaticActions(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, termCode := range []uint32{2001, 5002} {
			body := roECURBody(false)
			body.SessionID += fmt.Sprint(termCode)
			id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			initial := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(initial, roECURAnswer(false, 2001), serverOrigin)
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
			q := roECURBody(true)
			ib := initial.Body.(*ro.CCR)
			q.SessionID = ib.SessionID
			q.OriginHost = ib.OriginHost
			q.OriginRealm = ib.OriginRealm
			q.DestinationHost = ib.DestinationHost
			q.DestinationRealm = ib.DestinationRealm
			if e = ro.CheckECURContinuation(ib, q); e != nil {
				t.Fatal(e)
			}
			id, e = p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value = 77
			term := routeRead(t, co, remote)
			if term.Header.HopByHop == initial.Header.HopByHop || term.Header.EndToEnd == initial.Header.EndToEnd {
				t.Fatal("exchange identity reused")
			}
			if term.Body.(*ro.CCR).MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value != 1 {
				t.Fatal("caller alias")
			}
			if e = ro.CheckECURContinuation(ib, term.Body.(*ro.CCR)); e != nil {
				t.Fatal(e)
			}
			ans, e = co.CompleteAnswer(term, roECURAnswer(true, termCode), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got = routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete || got.Answer.Body.(*ro.CCA).ResultCode != termCode {
				t.Fatal(got)
			}

		}
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e := remote.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic INITIAL/refund/retry", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestRoECURInboundPairsAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, tc := range []struct {
			term bool
			body pdu.Body
		}{{false, roECURAnswer(false, 2001)}, {false, roECURAnswer(false, 4012)}, {false, &pdu.ErrorAnswer{ResultCode: 3004}}, {true, roECURAnswer(true, 2001)}, {true, roECURAnswer(true, 5002)}, {true, &pdu.ErrorAnswer{ResultCode: 3004}}} {
			req, e := co.CompleteRequest(roECURBody(tc.term), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(1100 + i), EndToEnd: uint32(2100 + i)})
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
			in.Message().Body.(*ro.CCR).CCRequestNumber = 22
			if tc.term {
				wrong := roECURAnswer(true, 2001)
				wrong.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{2}, ResultCode: avp.Some(uint32(2001))}}
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
func TestRoECURCancelLateAndInvalidAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), roECURBody(true), RequestOptions{}, port)
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
		ans, e := co.CompleteAnswer(req, roECURAnswer(true, 2001), serverOrigin)
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
		id, e = p.Submit(ctxTest(t), roECURBody(false), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req = routeRead(t, co, remote)
		altered, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		r := altered.Body.(*ro.CCR)
		replacement := roECURBody(true)
		replacement.OriginHost = r.OriginHost
		replacement.OriginRealm = r.OriginRealm
		replacement.DestinationHost = r.DestinationHost
		replacement.DestinationRealm = r.DestinationRealm
		altered.Body = replacement
		ans, e = co.CompleteAnswer(altered, roECURAnswer(true, 2001), serverOrigin)
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
