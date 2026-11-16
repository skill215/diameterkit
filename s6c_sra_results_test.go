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
)

func sraRole(t *testing.T, role s6c.RoutingOriginRole, unri avp.Optional[bool], requester, responder pdu.Origin) validation.Context {
	t.Helper()
	c, e := s6c.RoutingRole(s6c.RoutingRoleEvidence{Role: role, UNRI: unri, Requester: requester, Responder: responder})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func sraFailure(code uint32) *s6c.SRA {
	a := &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})}
	if code == 5550 {
		a.MSCAbsentUserDiagnosticSM = avp.Some(uint32(0))
		a.MWDStatus = avp.Some(uint32(0))
	}
	return a
}
func TestS6cSRAResultsTransportPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, t4 := range []bool{false, true} {
			for _, code := range []uint32{5550, 5552, 5556, 5557} {
				body := requestBody()
				ctx := sraRole(t, s6c.RoutingFromGMSC, avp.Optional[bool]{}, n.cfg.Origin, serverOrigin)
				if t4 {
					body = t4RoutingBody(t)
					ctx = t4RoutingContext(t, n.cfg.Origin, serverOrigin)
				}
				id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: ctx}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, remote)
				ans, e := co.CompleteAnswer(req, sraFailure(code), serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Err != nil || got.Outcome != AnswerReceived || !got.Validation.Complete || got.Answer.Body.(*s6c.SRA).ExperimentalResult.Value.ExperimentalResultCode != code {
					t.Fatal(got)
				}
				// Matching inbound flow proves normal failure Respond, not just outbound parsing.
				var incoming pdu.Message
				if t4 {
					b := t4RoutingBody(t)
					incoming, e = co.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: code, EndToEnd: code + 100})
				} else {
					incoming = inboundRequest(t, n, co, code)
				}
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, incoming); e != nil {
					t.Fatal(e)
				}
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				ctx = sraRole(t, s6c.RoutingFromGMSC, avp.Optional[bool]{}, serverOrigin, n.cfg.Origin)
				if t4 {
					ctx = t4RoutingContext(t, serverOrigin, n.cfg.Origin)
				}
				if r := in.Check(ctx); !r.Complete {
					t.Fatal(r)
				}
				if r, e := in.Respond(ctxTest(t), sraFailure(code)); e != nil || !r.Committed {
					t.Fatal(r, e)
				}
				duplicateAnswer(t, n, co, remote, incoming)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
			}
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic retry/MWD action", e)
		}
	})
}
func TestS6cSRAAbsentInboundFreezeAndRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 3901)
		if _, e := in.Respond(ctxTest(t), sraFailure(5550)); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		wrong := sraRole(t, s6c.RoutingFromIPSMGW, avp.Some(false), draOrigin, n.cfg.Origin)
		if r := in.Check(wrong); r.Complete {
			t.Fatal(r)
		}
		in.Message().Body.(*s6c.SRR).SRRFlags = avp.Some(uint32(7))
		ctx := sraRole(t, s6c.RoutingFromIPSMGW, avp.Some(false), serverOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(sraRole(t, s6c.RoutingFromIPSMGW, avp.Some(true), serverOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("UNRI replaced after freeze")
		}
		if _, e := in.Respond(ctxTest(t), sraFailure(5550)); !errors.Is(e, validation.ErrMissing) {
			t.Fatal(e)
		}
		a := sraFailure(5550)
		a.UserName = avp.Some("001010123456789")
		a.MMEAbsentUserDiagnosticSM = avp.Some(uint32(14))
		if _, e := in.Respond(ctxTest(t), a); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
		a.MMEAbsentUserDiagnosticSM = avp.Some(uint32(13))
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, a); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if r, e := in.Respond(ctxTest(t), a); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		got := duplicateAnswer(t, n, co, remote, req).Body.(*s6c.SRA)
		if !got.MSCAbsentUserDiagnosticSM.Present || got.MSCAbsentUserDiagnosticSM.Value != 0 || !got.UserName.Present {
			t.Fatal(got)
		}
	})
}
func TestS6cSRAIncompleteAndReservedRuntimeReports(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		unknown := sraRole(t, s6c.RoutingFromIPSMGW, avp.Optional[bool]{}, n.cfg.Origin, serverOrigin)
		for _, reserved := range []bool{false, true} {
			ctx := unknown
			if reserved {
				ctx = sraRole(t, s6c.RoutingFromGMSC, avp.Optional[bool]{}, n.cfg.Origin, serverOrigin)
			}
			id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, remote)
			a := sraFailure(5550)
			a.MMEAbsentUserDiagnosticSM = avp.Some(uint32(13))
			ans, e := co.CompleteAnswer(req, a, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			raw, e := co.Append(nil, ans)
			if e != nil {
				t.Fatal(e)
			}
			if reserved {
				found := false
				for off := 20; off < len(raw); {
					size := int(binary.BigEndian.Uint32(raw[off+4:off+8]) & 0xffffff)
					if binary.BigEndian.Uint32(raw[off:]) == 3313 {
						binary.BigEndian.PutUint32(raw[off+12:], 14)
						found = true
					}
					off += (size + 3) &^ 3
				}
				if !found {
					t.Fatal("missing diagnostic")
				}
			}
			sendRejectRaw(t, remote, raw)
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Outcome != AnswerReceived || got.Answer.Body == nil || got.Validation.Complete {
				t.Fatal(got)
			}
			if reserved {
				if !errors.Is(got.Validation.Err(), validation.ErrUnsupported) || got.Answer.Body.(*s6c.SRA).MMEAbsentUserDiagnosticSM.Value != 14 {
					t.Fatal(got)
				}
			} else {
				if got.Validation.State != validation.NeedsApplication {
					t.Fatal(got)
				}
				known := sraRole(t, s6c.RoutingFromIPSMGW, avp.Some(true), n.cfg.Origin, serverOrigin)
				if r := co.ValidateAnswerContext(req, got.Answer, pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}}, known); !r.Complete {
					t.Fatal(r)
				}
			}
			sendRejectRaw(t, remote, raw) // Late duplicate cannot complete twice, even when semantic report is incomplete.
			short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			_, e = port.Receive(short)
			cancel()
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal(e)
			}
		}
	})
}
func TestS6cSRAFailureDuplicateCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 3951)
		if r := in.Check(sraRole(t, s6c.RoutingFromIPSMGW, avp.Some(false), serverOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		dup := otherArrival(t, co, req, 3952)
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		a := sraFailure(5550)
		a.UserName = avp.Some("001010123456789")
		if r, e := in.Respond(ctxTest(t), a); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			ans := routeRead(t, co, remote)
			r := req
			if ans.Header.HopByHop == dup.Header.HopByHop {
				r = dup
			}
			if e := co.ValidateAnswer(r, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		dup.Header.HopByHop++
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		cached := duplicateAnswer(t, n, co, remote, dup).Body.(*s6c.SRA)
		if !cached.UserName.Present || !cached.MSCAbsentUserDiagnosticSM.Present || cached.MSCAbsentUserDiagnosticSM.Value != 0 {
			t.Fatal(cached)
		}
		dup.Header.HopByHop++
		dup.Body.(*s6c.SRR).SCAddress[0] ^= 1
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		if ans, ok := routeRead(t, co, remote).Body.(*pdu.ErrorAnswer); !ok || ans.ResultCode != 5012 {
			t.Fatal(ans)
		}
		noDuplicateDelivery(t, p)
	})
}
func TestS6cSRAFailureBudgetAndClose(t *testing.T) {
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
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), sraFailure(5550)) }()
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
