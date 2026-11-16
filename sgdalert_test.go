// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/sgdalert"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func legacyAlertConfig(t *testing.T, cache bool) NodeConfig {
	c := accessAlertConfig(t, cache)
	var profiles []schema.Profile
	for _, f := range []func() (schema.Profile, error){base.Profile, s6c.Profile, sgd.Profile, sgdalert.Profile} {
		p, e := f()
		if e != nil {
			t.Fatal(e)
		}
		profiles = append(profiles, p)
	}
	var e error
	c.Bundle, e = schema.Compile(profiles...)
	if e != nil {
		t.Fatal(e)
	}
	c.Capabilities.AuthApplicationID = append(c.Capabilities.AuthApplicationID, sgdalert.ApplicationID)
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	out.Command = (*sgdalert.ALR)(nil).CommandKey()
	in.Command = out.Command
	c.Peers[0].Commands = append(c.Peers[0].Commands, out)
	c.Peers[0].InboundCommands = append(c.Peers[0].InboundCommands, in)
	return c
}
func legacyAlertAnswer(code uint32) *sgdalert.ALA {
	return &sgdalert.ALA{AuthSessionState: 1, ResultCode: avp.Some(code)}
}

func TestLegacySGdAlertRuntimePairsAndCancel(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		body := (*sgdalert.ALR)(accessAlertBody(3, true))
		if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		ctx := alertEvidence(t, s6c.AlertFromSGSN, n.cfg.Origin, gmscOrigin)
		for _, a := range []pdu.Body{legacyAlertAnswer(2001), legacyAlertAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}, &pdu.ErrorAnswer{ResultCode: 5012}} {
			id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			r := routeRead(t, co, remote)
			if r.Header.ApplicationID != sgdalert.ApplicationID {
				t.Fatal(r)
			}
			ans, e := co.CompleteAnswer(r, a, gmscOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			if result := routeCompletion(t, n, port, id); result.Outcome != AnswerReceived || !result.Validation.Complete {
				t.Fatal(result)
			}
		}
		id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		_ = n.Cancel(id)
		if c := routeCompletion(t, n, port, id); !errors.Is(c.Err, context.Canceled) || c.Outcome != WrittenNoAnswer {
			t.Fatal(c)
		}
		a, e := co.CompleteAnswer(r, legacyAlertAnswer(2001), gmscOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late answer completed twice", e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}

func TestLegacySGdAlertRuntimeInbound(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, a := range []pdu.Body{legacyAlertAnswer(2001), legacyAlertAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			r, e := co.CompleteRequest((*sgdalert.ALR)(accessAlertBody(3, true)), pdu.RequestEnvelope{OriginHost: accessOrigin.Host, OriginRealm: accessOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(7300 + i), EndToEnd: uint32(8300 + i)})
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
			if _, e = in.Respond(ctxTest(t), a); !errors.Is(e, validation.ErrContextRequired) {
				t.Fatal(e)
			}
			view := in.Message().Body.(*sgdalert.ALR)
			view.SMSGMSCAlertEvent.Value = 0
			view.UserIdentifier.UserName = avp.Some("mutated")
			if v := in.Check(alertEvidence(t, s6c.AlertFromSGSN, accessOrigin, n.cfg.Origin)); !v.Complete {
				t.Fatal(v)
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			if _, e = in.Respond(canceled, a); !errors.Is(e, context.Canceled) {
				t.Fatal(e)
			}
			if _, e = in.Respond(ctxTest(t), reportAnswer(true, 2001)); e == nil {
				t.Fatal("wrong binding committed")
			}
			if v, e := in.Respond(ctxTest(t), a); e != nil || !v.Committed {
				t.Fatal(v, e)
			}
			ans := routeRead(t, co, remote)
			if e := co.ValidateAnswer(r, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(short, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic business follow-up", e)
		}
	})
}

func TestLegacySGdAlertRuntimeDuplicateIsolation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		r, e := co.CompleteRequest((*sgdalert.ALR)(accessAlertBody(3, true)), pdu.RequestEnvelope{OriginHost: accessOrigin.Host, OriginRealm: accessOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7400, EndToEnd: 8400})
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
		if v := in.Check(alertEvidence(t, s6c.AlertFromSGSN, accessOrigin, n.cfg.Origin)); !v.Complete {
			t.Fatal(v)
		}
		dup, e := co.Snapshot(r)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*sgdalert.ALR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other.example", ProxyState: []byte{7, 3}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if v, e := in.Respond(ctxTest(t), legacyAlertAnswer(2001)); e != nil || !v.Committed {
			t.Fatal(v, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := r
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if e := co.ValidateAnswer(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, dup)
		// Same Origin/E2E with a different Application-Id is a content conflict,
		// never a cache hit for the other binding or a new business delivery.
		standard, e := co.Snapshot(r)
		if e != nil {
			t.Fatal(e)
		}
		standard.Header.ApplicationID = appID
		standard.Header.HopByHop += 10
		standard.Body = (*s6c.ALR)(standard.Body.(*sgdalert.ALR))
		if e = writeWire(ctxTest(t), remote, co, standard); e != nil {
			t.Fatal(e)
		}
		a := routeRead(t, co, remote)
		if a.Header.ApplicationID != appID {
			t.Fatal(a)
		}
		if b, ok := a.Body.(*pdu.ErrorAnswer); !ok || b.ResultCode != 5012 {
			t.Fatal(a)
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}

func TestLegacySGdAlertRuntimeCloseRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := legacyAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		r, e := co.CompleteRequest((*sgdalert.ALR)(accessAlertBody(3, true)), pdu.RequestEnvelope{OriginHost: accessOrigin.Host, OriginRealm: accessOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7500, EndToEnd: 8500})
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
		if v := in.Check(alertEvidence(t, s6c.AlertFromSGSN, accessOrigin, n.cfg.Origin)); !v.Complete {
			t.Fatal(v)
		}
		blocked, e := co.Snapshot(r)
		if e != nil {
			t.Fatal(e)
		}
		blocked.Header.HopByHop++
		blocked.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 || a.Command.ApplicationID != sgdalert.ApplicationID {
			t.Fatal(a)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Go(func() { <-start; _, _ = in.Respond(context.Background(), legacyAlertAnswer(2001)) })
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
