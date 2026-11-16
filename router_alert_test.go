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
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/sgdalert"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

var routerOrigin = pdu.Origin{Host: "router.example", Realm: "example"}

func routerAlertRuntimeConfig(t *testing.T, cache bool) NodeConfig {
	c := legacyAlertConfig(t, cache)
	for i := range c.Peers[0].InboundCommands {
		c.Peers[0].InboundCommands[i].Sources = []pdu.Origin{routerOrigin}
	}
	return c
}
func routerAlertBody(legacy bool) pdu.Body {
	b := accessAlertBody(3, true)
	b.UserIdentifier.UserName = avp.Some("router-alert-74")
	if legacy {
		return (*sgdalert.ALR)(b)
	}
	return b
}
func routerAlertAnswer(legacy bool, code uint32) pdu.Body {
	if legacy {
		return legacyAlertAnswer(code)
	}
	return reportAnswer(true, code)
}
func routerAlertView(b pdu.Body) *s6c.ALR {
	if v, ok := b.(*s6c.ALR); ok {
		return v
	}
	return (*s6c.ALR)(b.(*sgdalert.ALR))
}
func routerRuntimeContext(t *testing.T, router, gmsc pdu.Origin, token, sc string) validation.Context {
	t.Helper()
	c, e := s6c.RouterAlert(s6c.RouterAlertEvidence{Router: router, GMSC: gmsc, PriorAlertToken: token, PriorGMSCAddress: sc})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func eachRouterBinding(t *testing.T, f func(*testing.T, bool, bool)) {
	t.Helper()
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%t", legacy), func(t *testing.T) { eachNetwork(t, func(t *testing.T, tls bool) { f(t, tls, legacy) }) })
	}
}

func TestRouterAlertRuntimeOutbound(t *testing.T) {
	eachRouterBinding(t, func(t *testing.T, tls, legacy bool) {
		n, co := newInboundNode(t, routerAlertRuntimeConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		ctx := routerRuntimeContext(t, n.cfg.Origin, gmscOrigin, "router-alert-74", "12345")
		for _, bad := range []validation.Context{{}, routerRuntimeContext(t, n.cfg.Origin, gmscOrigin, "other-token", "12345"), routerRuntimeContext(t, n.cfg.Origin, gmscOrigin, "router-alert-74", "54321"), routerRuntimeContext(t, routerOrigin, gmscOrigin, "router-alert-74", "12345")} {
			if _, e := p.Submit(ctxTest(t), routerAlertBody(legacy), RequestOptions{Validation: bad}, port); e == nil {
				t.Fatal("missing/wrong evidence admitted")
			}
			if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
				t.Fatal(n.Resources())
			}
		}
		for _, answer := range []pdu.Body{routerAlertAnswer(legacy, 2001), routerAlertAnswer(legacy, 5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			body := routerAlertBody(legacy)
			id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			routerAlertView(body).UserIdentifier.UserName = avp.Some("changed")
			r := routeRead(t, co, remote)
			if routerAlertView(r.Body).UserIdentifier.UserName.Value != "router-alert-74" {
				t.Fatal("caller view retained")
			}
			a, e := co.CompleteAnswer(r, answer, gmscOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			if result := routeCompletion(t, n, port, id); result.Outcome != AnswerReceived || !result.Validation.Complete {
				t.Fatal(result)
			}
		}
		id, e := p.Submit(ctxTest(t), routerAlertBody(legacy), RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		_ = n.Cancel(id)
		if result := routeCompletion(t, n, port, id); !errors.Is(result.Err, context.Canceled) || result.Outcome != WrittenNoAnswer {
			t.Fatal(result)
		}
		a, e := co.CompleteAnswer(r, routerAlertAnswer(legacy, 2001), gmscOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late answer second completion", e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}

func TestRouterAlertRuntimeInboundCache(t *testing.T) {
	eachRouterBinding(t, func(t *testing.T, tls, legacy bool) {
		n, co := newInboundNode(t, routerAlertRuntimeConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		r, e := co.CompleteRequest(routerAlertBody(legacy), pdu.RequestEnvelope{OriginHost: routerOrigin.Host, OriginRealm: routerOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7401, EndToEnd: 8401})
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
		if in.Validation().Complete {
			t.Fatal("history guessed from token")
		}
		if _, e = in.Respond(ctxTest(t), routerAlertAnswer(legacy, 2001)); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		if in.Check(routerRuntimeContext(t, routerOrigin, n.cfg.Origin, "WRONG", "12345")).Complete {
			t.Fatal("wrong mapping accepted")
		}
		routerAlertView(in.Message().Body).UserIdentifier.UserName = avp.Some("mutated")
		if v := in.Check(routerRuntimeContext(t, routerOrigin, n.cfg.Origin, "router-alert-74", "12345")); !v.Complete {
			t.Fatal(v)
		}
		if in.Check(routerRuntimeContext(t, routerOrigin, n.cfg.Origin, "other", "12345")).Complete {
			t.Fatal("changed frozen evidence")
		}
		dup, e := co.Snapshot(r)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		routerAlertView(dup.Body).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{7, 4}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e = in.Respond(canceled, routerAlertAnswer(legacy, 2001)); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if v, e := in.Respond(ctxTest(t), routerAlertAnswer(legacy, 2001)); e != nil || !v.Committed {
			t.Fatal(v, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			orig := r
			if a.Header.HopByHop == dup.Header.HopByHop {
				orig = dup
			}
			if e := co.ValidateAnswer(orig, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, dup)
		conflict, e := co.Snapshot(r)
		if e != nil {
			t.Fatal(e)
		}
		conflict.Header.HopByHop += 10
		routerAlertView(conflict.Body).UserIdentifier.UserName = avp.Some("Router-alert-74")
		if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, conflict).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 5012 {
			t.Fatal(a)
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
		short, stop := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer stop()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic SRR/SMS", e)
		}
	})
}

func TestRouterAlertRuntimeBudgetAndClose(t *testing.T) {
	eachRouterBinding(t, func(t *testing.T, tls, legacy bool) {
		cfg := routerAlertRuntimeConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		r, e := co.CompleteRequest(routerAlertBody(legacy), pdu.RequestEnvelope{OriginHost: routerOrigin.Host, OriginRealm: routerOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7501, EndToEnd: 8501})
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
		r.Header.HopByHop++
		r.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, r); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, r).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		if v := in.Check(routerRuntimeContext(t, routerOrigin, n.cfg.Origin, "router-alert-74", "12345")); !v.Complete {
			t.Fatal(v)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Go(func() { <-start; _, _ = in.Respond(context.Background(), routerAlertAnswer(legacy, 5012)) })
		wg.Go(func() { <-start; _ = n.Close() })
		close(start)
		wg.Wait()
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 || n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}
