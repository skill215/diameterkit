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

func t4Requester(alert bool) pdu.Origin {
	if alert {
		return serverOrigin
	}
	return gmscOrigin
}
func t4Responder(alert bool) pdu.Origin {
	if alert {
		return iwmscOrigin
	}
	return serverOrigin
}
func t4Body(alert bool) pdu.Body {
	if alert {
		b := alertBody()
		b.UserIdentifier = s6c.AlertUserIdentifier{UserName: avp.Some("001010123456789")}
		return b
	}
	b := reportBody()
	b.UserIdentifier = s6c.ReportUserIdentifier{UserName: avp.Some("001010123456789")}
	return b
}
func mutateT4(body pdu.Body, sc bool) {
	switch b := body.(type) {
	case *s6c.RDR:
		if sc {
			b.SCAddress[0] = 0x43
		} else {
			b.UserIdentifier.UserName = avp.Some("001010123456780")
		}
	case *s6c.ALR:
		if sc {
			b.SCAddress[0] = 0x43
		} else {
			b.UserIdentifier.UserName = avp.Some("001010123456780")
		}
	}
}
func t4Context(t *testing.T, alert bool, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	procedure := s6c.T4DeliveryReport
	if alert {
		procedure = s6c.T4HSSAlert
	}
	ctx, e := s6c.T4NoMSISDN(s6c.T4NoMSISDNEvidence{Procedure: procedure, Requester: requester, Responder: responder, IMSI: "001010123456789", SCAddress: "12345"})
	if e != nil {
		t.Fatal(e)
	}
	return ctx
}
func receiveT4Request(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, alert bool, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	origin := t4Requester(alert)
	req, e := co.CompleteRequest(t4Body(alert), pdu.RequestEnvelope{OriginHost: origin.Host, OriginRealm: origin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
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
func TestS6cT4InboundRoleFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, alert := range []bool{false, true} {
			n, co := newInboundNode(t, reportAlertConfig(t, false))
			p, remote := networkConnectNode(t, n, co, tls)
			in, req := receiveT4Request(t, n, p, co, remote, alert, 3001)
			if r := in.Validation(); r.State != validation.NeedsApplication || r.Complete {
				t.Fatal(r)
			}
			if _, e := in.Respond(ctxTest(t), reportAnswer(alert, 2001)); !errors.Is(e, validation.ErrContextRequired) {
				t.Fatal(e)
			}
			for _, ctx := range []validation.Context{t4Context(t, !alert, t4Requester(alert), n.cfg.Origin), t4Context(t, alert, draOrigin, n.cfg.Origin), t4Context(t, alert, t4Requester(alert), t4Responder(alert))} {
				if r := in.Check(ctx); r.Complete {
					t.Fatal(r)
				}
			}
			mutateT4(in.Message().Body, false)
			mutateT4(in.Message().Body, true)
			ctx := t4Context(t, alert, t4Requester(alert), n.cfg.Origin)
			if r := in.Check(ctx); !r.Complete {
				t.Fatal(r)
			}
			if r := in.Check(t4Context(t, !alert, t4Requester(alert), n.cfg.Origin)); r.Complete {
				t.Fatal("changed frozen role")
			}
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			if _, e := in.Respond(canceled, reportAnswer(alert, 2001)); !errors.Is(e, context.Canceled) {
				t.Fatal(e)
			}
			if r, e := in.Respond(ctxTest(t), reportAnswer(alert, 2001)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			duplicateAnswer(t, n, co, remote, req)
			waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
			short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("automatic SRR or SMS", e)
			}
		}
	})
}
func TestS6cT4OutboundBindingAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, alert := range []bool{false, true} {
			n, co := newInboundNode(t, reportAlertConfig(t, false))
			p, remote := networkConnectNode(t, n, co, tls)
			port := portTest(t, n, 1)
			body := t4Body(alert)
			if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
				t.Fatal(e)
			}
			wrong := t4Context(t, alert, n.cfg.Origin, draOrigin)
			if _, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: wrong}, port); !errors.Is(e, validation.ErrContext) {
				t.Fatal(e)
			}
			if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
				t.Fatal(n.Resources())
			}
			answers := []pdu.Body{reportAnswer(alert, 2001), reportAnswer(alert, 5012), &pdu.ErrorAnswer{ResultCode: 3004}}
			if !alert {
				answers = append(answers,
					&s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})},
					&s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5558})},
					&s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserIdentifier: avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf3}})})
			}
			for _, answer := range answers {
				b := t4Body(alert)
				ctx := t4Context(t, alert, n.cfg.Origin, t4Responder(alert))
				id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
				if e != nil {
					t.Fatal(e)
				}
				mutateT4(b, false)
				req := routeRead(t, co, remote)
				ans, e := co.CompleteAnswer(req, answer, t4Responder(alert))
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
			id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: t4Context(t, alert, n.cfg.Origin, t4Responder(alert))}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, remote)
			_ = n.Cancel(id)
			routeCompletion(t, n, port, id)
			a, e := co.CompleteAnswer(req, reportAnswer(alert, 2001), t4Responder(alert))
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
		}
	})
}
func TestS6cT4DuplicateCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, alert := range []bool{false, true} {
			n, co := newInboundNode(t, reportAlertConfig(t, true))
			p, remote := networkConnectNode(t, n, co, tls)
			in, req := receiveT4Request(t, n, p, co, remote, alert, 3101)
			if r := in.Check(t4Context(t, alert, t4Requester(alert), n.cfg.Origin)); !r.Complete {
				t.Fatal(r)
			}
			dup, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			dup.Header.HopByHop++
			dup.Header.Flags |= pdu.RetransmittedFlag
			proxies := []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 2}}}
			if alert {
				dup.Body.(*s6c.ALR).ProxyInfo = proxies
			} else {
				dup.Body.(*s6c.RDR).ProxyInfo = proxies
			}
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
			noDuplicateDelivery(t, p)
			if r, e := in.Respond(ctxTest(t), reportAnswer(alert, 2001)); e != nil || !r.Committed {
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
			for i := range 2 {
				c, e := co.Snapshot(req)
				if e != nil {
					t.Fatal(e)
				}
				c.Header.HopByHop = uint32(3110 + i)
				mutateT4(c.Body, i == 1)
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
		}
	})
}
func TestS6cT4RoleCheckCloseAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, alert := range []bool{false, true} {
			cfg := reportAlertConfig(t, false)
			cfg.Resources.InboundClasses[0].Slots = 1
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, tls)
			in, req := receiveT4Request(t, n, p, co, remote, alert, 3201)
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
			ctx := t4Context(t, alert, t4Requester(alert), n.cfg.Origin)
			wg.Add(2)
			go func() { defer wg.Done(); <-start; in.Check(ctx) }()
			go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), reportAnswer(alert, 2001)) }()
			close(start)
			_ = n.Close()
			wg.Wait()
			if e = n.Wait(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
			if n.Resources().InboundBytes != 0 {
				t.Fatal(n.Resources())
			}
		}
	})
}
