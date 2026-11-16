// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
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

var accessOrigin = pdu.Origin{Host: "old-access.example", Realm: "example"}

func accessAlertConfig(t *testing.T, cache bool) NodeConfig {
	c := reportAlertConfig(t, cache)
	c.Peers[0].Commands = c.Peers[0].Commands[1:]
	c.Peers[0].InboundCommands = c.Peers[0].InboundCommands[1:]
	c.Peers[0].Commands[0].DestinationHost = gmscOrigin.Host
	c.Peers[0].Commands[0].Sources.Terminals = []pdu.Origin{gmscOrigin}
	c.Peers[0].InboundCommands[0].Sources = []pdu.Origin{accessOrigin}
	return c
}
func accessAlertBody(event uint32, sgsn bool) *s6c.ALR {
	b := &s6c.ALR{SessionID: "access.example;alert;1", AuthSessionState: 1, UserIdentifier: s6c.AlertUserIdentifier{UserName: avp.Some("001010123456789")}, SCAddress: []byte{0x21, 0x43, 0xf5}, SMSGMSCAlertEvent: avp.Some(event)}
	if event&2 != 0 {
		n := s6c.ServingNode{MMEName: avp.Some("new-mme.example"), MMERealm: avp.Some("example"), MMENumber: avp.Some([]byte{0x21, 0xf3})}
		if sgsn {
			n = s6c.ServingNode{SGSNNumber: avp.Some([]byte{0x21, 0xf4})}
		}
		b.ServingNode = avp.Some(n)
	}
	return b
}
func alertEvidence(t *testing.T, role s6c.AlertOriginRole, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, e := s6c.AlertRole(s6c.AlertRoleEvidence{Role: role, Requester: requester, Responder: responder})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func receiveAccessAlert(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req, e := co.CompleteRequest(accessAlertBody(3, true), pdu.RequestEnvelope{OriginHost: accessOrigin.Host, OriginRealm: accessOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
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
func TestS6cAccessAlertInboundRoleFreeze(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, accessAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveAccessAlert(t, n, p, co, remote, 3001)
		if r := in.Validation(); r.State != validation.NeedsApplication || r.Complete {
			t.Fatal(r)
		}
		if _, e := in.Respond(ctxTest(t), reportAnswer(true, 2001)); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		for _, ctx := range []validation.Context{alertEvidence(t, s6c.AlertFromHSS, accessOrigin, n.cfg.Origin), alertEvidence(t, s6c.AlertFromMME, draOrigin, n.cfg.Origin), alertEvidence(t, s6c.AlertFromMME, accessOrigin, gmscOrigin)} {
			if r := in.Check(ctx); r.Complete {
				t.Fatal(r)
			}
		}
		view := in.Message().Body.(*s6c.ALR)
		view.UserIdentifier.UserName = avp.Some("modified")
		view.SMSGMSCAlertEvent.Value = 0
		view.ServingNode.Value = s6c.ServingNode{}
		ctx := alertEvidence(t, s6c.AlertFromMME, accessOrigin, n.cfg.Origin) // Old MME -> new SGSN is valid.
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(alertEvidence(t, s6c.AlertFromSGSN, accessOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("changed frozen role")
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, reportAnswer(true, 2001)); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if r, e := in.Respond(ctxTest(t), reportAnswer(true, 2001)); e != nil || !r.Committed {
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
func TestS6cAccessAlertOutboundBindingAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, accessAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		body := accessAlertBody(2, false)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		wrong := alertEvidence(t, s6c.AlertFromSGSN, n.cfg.Origin, draOrigin)
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: wrong}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		for i, answer := range []pdu.Body{reportAnswer(true, 2001), reportAnswer(true, 5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			b := accessAlertBody(uint32(i+1), i%2 == 0)
			ctx := alertEvidence(t, s6c.AlertFromSGSN, n.cfg.Origin, gmscOrigin)
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			b.SMSGMSCAlertEvent.Value = 0
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, answer, gmscOrigin)
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
		id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: alertEvidence(t, s6c.AlertFromMME, n.cfg.Origin, gmscOrigin)}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, e := co.CompleteAnswer(req, reportAnswer(true, 2001), gmscOrigin)
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
func TestS6cAccessAlertDuplicateCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, accessAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveAccessAlert(t, n, p, co, remote, 3101)
		if r := in.Check(alertEvidence(t, s6c.AlertFromSGSN, accessOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*s6c.ALR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 2}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), reportAnswer(true, 2001)); e != nil || !r.Committed {
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
			b := c.Body.(*s6c.ALR)
			if i == 0 {
				b.SMSGMSCAlertEvent.Value = 2
			} else {
				b.ServingNode.Value.SGSNNumber.Value[0] = 0x43
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
func TestS6cAccessAlertRoleCheckCloseAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := accessAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveAccessAlert(t, n, p, co, remote, 3201)
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
		ctx := alertEvidence(t, s6c.AlertFromMME, accessOrigin, n.cfg.Origin)
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), reportAnswer(true, 2001)) }()
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
