package diameter

import (
	"context"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"sync"
	"testing"
	"time"
)

func imsRoutingBody(t *testing.T) *s6c.SRR {
	b := routingRequest(t)
	b.MSISDN = avp.Optional[[]byte]{}
	b.UserName = avp.Some("001010000000001")
	b.SMSMICorrelationID = avp.Some(s6c.RoutingSMSMICorrelationID{DestinationSIPURI: "sip:bob@example.net"})
	return b
}
func mutateIMSRouting(body pdu.Body, sc bool) {
	b := body.(*s6c.SRR)
	if sc {
		b.SCAddress[0] = 0x43
	} else {
		b.UserName = avp.Some("001010123456780")
	}
}
func imsRoutingContext(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, e := s6c.IMSRouting(s6c.IMSRoutingEvidence{Role: s6c.RoutingFromGMSC, Requester: requester, Responder: responder, HSSID: "001010000000001", DestinationSIPURI: "sip:bob@example.net", SCAddress: "12345"})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func receiveIMSRouting(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req, e := co.CompleteRequest(imsRoutingBody(t), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
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
	return in, req
}
func TestS6cIMSRoutingInboundRoleFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSRouting(t, n, p, co, remote, 3001)
		if r := in.Validation(); r.State != validation.NeedsApplication || r.Complete {
			t.Fatal(r)
		}
		if _, e := in.Respond(ctxTest(t), routingAnswer()); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		for _, ctx := range []validation.Context{t4Context(t, false, serverOrigin, n.cfg.Origin), imsRoutingContext(t, draOrigin, n.cfg.Origin), imsRoutingContext(t, serverOrigin, serverOrigin)} {
			if r := in.Check(ctx); r.Complete {
				t.Fatal(r)
			}
		}
		mutateIMSRouting(in.Message().Body, false)
		mutateIMSRouting(in.Message().Body, true)
		ctx := imsRoutingContext(t, serverOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(t4Context(t, false, serverOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("changed frozen role")
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, routingAnswer()); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		bad := routingAnswer()
		bad.UserName.Present = false
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrMissing) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), routingAnswer()); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic SRR or SMS", e)
		}
	})
}
func TestS6cIMSRoutingOutboundBindingAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		body := imsRoutingBody(t)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		wrong := imsRoutingContext(t, n.cfg.Origin, draOrigin)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: wrong}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		answers := []pdu.Body{routingAnswer(), &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 3004}}
		answers = append(answers, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})})
		for _, answer := range answers {
			b := imsRoutingBody(t)
			ctx := imsRoutingContext(t, n.cfg.Origin, serverOrigin)
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			mutateIMSRouting(b, false)
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, answer, serverOrigin)
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
		}
		id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: imsRoutingContext(t, n.cfg.Origin, serverOrigin)}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, e := co.CompleteAnswer(req, routingAnswer(), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestS6cIMSRoutingDuplicateCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSRouting(t, n, p, co, remote, 3101)
		if r := in.Check(imsRoutingContext(t, serverOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		proxies := []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 2}}}
		dup.Body.(*s6c.SRR).ProxyInfo = proxies
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), routingAnswer()); e != nil || !r.Committed {
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
		duplicateAnswer(t, n, co, remote, dup)
		for i := range 3 {
			c, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			c.Header.HopByHop = uint32(3110 + i)
			mutateIMSRouting(c.Body, i == 1)
			if i == 2 {
				c.Body.(*s6c.SRR).UserName = req.Body.(*s6c.SRR).UserName
				c.Body.(*s6c.SRR).SMSMICorrelationID.Value.DestinationSIPURI = "sip:%62ob@example.net"
			}
			if e = writeWire(ctxTest(t), remote, co, c); e != nil {
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
func TestS6cIMSRoutingRoleCheckCloseAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSRouting(t, n, p, co, remote, 3201)
		blocked, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		blocked.Header.HopByHop++
		blocked.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		ctx := imsRoutingContext(t, serverOrigin, n.cfg.Origin)
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), routingAnswer()) }()
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
func TestS6cIMSRoutingRouteIdentityBudgetAndLateAnswer(t *testing.T) {
	for _, policy := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
		t.Run(fmt.Sprint(policy), func(t *testing.T) {
			eachNetwork(t, func(t *testing.T, tls bool) {
				n, co := newInboundNode(t, routeConfig(t))
				p, a := duplicateConnect(t, n, co, tls, "dra")
				_, b := duplicateConnect(t, n, co, tls, "dra2")
				r := newRouteTest(t, n, policy)
				port := portTest(t, n, 1)
				body := imsRoutingBody(t)
				id, e := r.Submit(ctxTest(t), body, RequestOptions{Validation: imsRoutingContext(t, n.cfg.Origin, serverOrigin)}, port)
				if e != nil {
					t.Fatal(e)
				}
				first := routeRead(t, co, a)
				before := n.Resources()
				body.UserName.Value = "001010123456780"
				body.SMRPSMEA.Value[2] = 0x99
				body.SMSMICorrelationID.Value.DestinationSIPURI = "sip:changed@example.net"
				n.mu.Lock()
				x := n.requests[id]
				reserved := x.charge.Capacity >= 2*x.answerBound
				n.mu.Unlock()
				if !reserved {
					t.Fatal("missing per-attempt answer reserve")
				}
				routeSuspect(t, n, p, a, co)
				if policy == ExplicitRetransmit {
					waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
					info, _ := n.RequestInfo(id)
					canceled, cancel := context.WithCancel(context.Background())
					cancel()
					if _, e := n.Retransmit(canceled, id, info.CurrentAttempt); !errors.Is(e, context.Canceled) {
						t.Fatal(e)
					}
					if _, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
						t.Fatal(e)
					}
				}
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

		})
	}
}

// Matching 5550 answers finish once even when external UNRI evidence is absent
// or contradicts a missing IMSI. The Result carries that validation outcome.
func TestS6cIMSRoutingUNRIReportsAndFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		makeContext := func(requester, responder pdu.Origin, unri avp.Optional[bool]) validation.Context {
			v, e := s6c.IMSRouting(s6c.IMSRoutingEvidence{Role: s6c.RoutingFromIPSMGW, Requester: requester, Responder: responder, HSSID: "001010000000001", DestinationSIPURI: "sip:bob@example.net", SCAddress: "12345", UNRI: unri})
			if e != nil {
				t.Fatal(e)
			}
			return v
		}
		for _, unri := range []avp.Optional[bool]{{}, avp.Some(false), avp.Some(true)} {
			ctx := makeContext(n.cfg.Origin, serverOrigin, unri)
			id, e := p.Submit(ctxTest(t), imsRoutingBody(t), RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, sraFailure(5550), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived {
				t.Fatal(got)
			}
			switch {
			case !unri.Present:
				if got.Validation.State != validation.NeedsApplication {
					t.Fatal(got)
				}
			case !unri.Value:
				if !errors.Is(got.Validation.Err(), validation.ErrMissing) {
					t.Fatal(got)
				}
			default:
				if !got.Validation.Complete {
					t.Fatal(got)
				}
			}
			if _, e = n.Retransmit(ctxTest(t), id, AttemptID{id, 1}); !errors.Is(e, ErrRequestNotActive) {
				t.Fatal(e)
			}
		}
		in, req := receiveIMSRouting(t, n, p, co, remote, 3501)
		if r := in.Check(makeContext(serverOrigin, n.cfg.Origin, avp.Some(false))); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(makeContext(serverOrigin, n.cfg.Origin, avp.Some(true))); r.Complete {
			t.Fatal("changed frozen UNRI")
		}
		if r, e := in.Respond(ctxTest(t), sraFailure(5550)); r.Committed || !errors.Is(e, validation.ErrMissing) {
			t.Fatal(r, e)
		}
		answer := sraFailure(5550)
		answer.UserName = avp.Some("001010123456789")
		if r, e := in.Respond(ctxTest(t), answer); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 && n.Resources().OutboundBytes == 0 })
	})
}
