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
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func shBody(code uint32) pdu.Body {
	identity := sh.UserIdentity{PublicIdentity: avp.Some("sip:alice@example.net")}
	switch code {
	case 306:
		return &sh.UDR{SessionID: "as.example;sh;pull", VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, UserIdentity: identity, DataReference: 24}
	case 307:
		return &sh.PUR{SessionID: "as.example;sh;update", VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, UserIdentity: identity, DataReference: 24, UserData: []byte("<Sh-Data/>")}
	case 308:
		return &sh.SNR{SessionID: "as.example;sh;subscribe", VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, UserIdentity: identity, DataReference: []uint32{25}, OneTimeNotification: avp.Some(uint32(0)), ExpiryTime: avp.Some(avp.Time(0))}
	default:
		return &sh.PNR{SessionID: "hss.example;sh;notification", VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, UserIdentity: identity, UserData: []byte("<Sh-Data/>")}
	}
}
func shAnswer(command, code uint32, experiment bool) pdu.Body {
	rc := avp.Some(code)
	ex := avp.Optional[sh.ExperimentalResult]{}
	if experiment {
		rc = avp.Optional[uint32]{}
		ex = avp.Some(sh.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
	}
	switch command {
	case 306:
		a := &sh.UDA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
		if !experiment && code == 2001 {
			a.UserData = avp.Some([]byte("<Sh-Data/>"))
		}
		return a
	case 307:
		return &sh.PUA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	case 308:
		return &sh.SNA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	default:
		return &sh.PNA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	}
}
func shConfig(t *testing.T, cache bool) NodeConfig {
	t.Helper()
	cfg := rejectConfig(t)
	if cache {
		cfg = duplicateConfig(t)
	}
	profiles := []schema.Profile{}
	for _, makeProfile := range []func() (schema.Profile, error){base.Profile, s6c.Profile, sgd.Profile, sh.Profile} {
		p, e := makeProfile()
		if e != nil {
			t.Fatal(e)
		}
		profiles = append(profiles, p)
	}
	var e error
	cfg.Bundle, e = schema.Compile(profiles...)
	if e != nil {
		t.Fatal(e)
	}
	// Announce Sh through the Base VSAI capability, distinct from normal-message VSAI.
	cfg.Capabilities.VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{sh.Application()}
	out, in := cfg.Peers[0].Commands[0], cfg.Peers[0].InboundCommands[0]
	cfg.Peers[0].Commands = nil
	cfg.Peers[0].InboundCommands = nil
	for code := uint32(306); code <= 309; code++ {
		p, q := out, in
		p.Command = shBody(code).CommandKey()
		q.Command = p.Command
		p.DestinationHost = serverOrigin.Host
		q.Sources = []pdu.Origin{gmscOrigin}
		if code == 309 {
			p.DestinationHost = gmscOrigin.Host
			p.Sources.Terminals = []pdu.Origin{gmscOrigin}
			q.Sources = []pdu.Origin{serverOrigin}
		}
		cfg.Peers[0].Commands = append(cfg.Peers[0].Commands, p)
		cfg.Peers[0].InboundCommands = append(cfg.Peers[0].InboundCommands, q)
	}
	return cfg
}
func shRequest(t *testing.T, n *Node, co *codec.Codec, code, hbh uint32) pdu.Message {
	t.Helper()
	origin := gmscOrigin
	if code == 309 {
		origin = serverOrigin
	}
	req, e := co.CompleteRequest(shBody(code), pdu.RequestEnvelope{OriginHost: origin.Host, OriginRealm: origin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 10000})
	if e != nil {
		t.Fatal(e)
	}
	return req
}
func sendSh(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, code, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req := shRequest(t, n, co, code, hbh)
	if e := writeWire(ctxTest(t), remote, co, req); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, req
}
func TestShFourIndependentExchanges(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 4)
		ids := map[RequestID]uint32{}
		for code := uint32(306); code <= 309; code++ {
			id, e := p.Submit(ctxTest(t), shBody(code), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			ids[id] = code
		}
		reqs := make([]pdu.Message, 4)
		hbhs, e2es := map[uint32]bool{}, map[uint32]bool{}
		for i := range reqs {
			reqs[i] = routeRead(t, co, remote)
			h := reqs[i].Header
			if hbhs[h.HopByHop] || e2es[h.EndToEnd] {
				t.Fatal("shared identifiers")
			}
			hbhs[h.HopByHop], e2es[h.EndToEnd] = true, true
		}
		for i := 3; i >= 0; i-- {
			req := reqs[i]
			origin := serverOrigin
			if req.Header.CommandCode == 309 {
				origin = gmscOrigin
			}
			a, e := co.CompleteAnswer(req, shAnswer(req.Header.CommandCode, 2001, false), origin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
		}
		for range 4 {
			r, e := port.Receive(ctxTest(t))
			if e != nil || r.Result.Outcome != AnswerReceived || !r.Result.Validation.Complete || r.Result.Answer.Header.CommandCode != ids[r.RequestID] {
				t.Fatal(r, e)
			}
			delete(ids, r.RequestID)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestShNotificationBeforeSubscriptionAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		id, e := p.Submit(ctxTest(t), shBody(308), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		snr := routeRead(t, co, remote)
		in, pnr := sendSh(t, n, p, co, remote, 309, 900)
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		if r, e := in.Respond(ctxTest(t), shAnswer(309, 2001, false)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, pnr)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(ctx)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("PNR consumed SNR pending", e)
		}
		sna, e := co.CompleteAnswer(snr, shAnswer(308, 5104, true), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, sna); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id || r.Result.Answer.Header.CommandCode != 308 {
			t.Fatal(r, e)
		}
		// Even after failure, no subscription registry suppresses another actual PNR.
		in, pnr = sendSh(t, n, p, co, remote, 309, 901)
		if r, e := in.Respond(ctxTest(t), shAnswer(309, 5107, true)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, pnr)
	})
}
func TestShInboundNormalExperimentalAndE(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		hbh := uint32(920)
		for command := uint32(306); command <= 309; command++ {
			for _, kind := range []int{0, 1, 2, 3} {
				body := shAnswer(command, 2001, false)
				switch kind {
				case 1:
					code := uint32(5102)
					if command == 307 {
						code = 5103
					}
					if command == 308 {
						code = 5104
					}
					if command == 309 {
						code = 5008
					}
					body = shAnswer(command, code, true)
				case 2:
					body = shAnswer(command, 5012, false)
				case 3:
					body = &pdu.ErrorAnswer{ResultCode: 3004}
				}
				in, req := sendSh(t, n, p, co, remote, command, hbh)
				hbh++
				if r, e := in.Respond(ctxTest(t), body); e != nil || !r.Committed {
					t.Fatal(r, e)
				}
				ans := duplicateAnswer(t, n, co, remote, req)
				if kind == 3 {
					if a, ok := ans.Body.(*pdu.ErrorAnswer); !ok || a.VendorSpecificApplicationID.Present {
						t.Fatal("E inherited normal VSAI", ans)
					}
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
			}
		}
	})
}
func TestShPrivateRequestAndPreparationRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSh(t, n, p, co, remote, 306, 950)
		view := in.Message().Body.(*sh.UDR)
		view.SessionID = "mutated"
		view.VendorSpecificApplicationID.AuthApplicationID.Value = 0
		view.UserIdentity.PublicIdentity.Value = "invalid"
		bad := shAnswer(306, 2001, false).(*sh.UDA)
		bad.VendorSpecificApplicationID.AuthApplicationID.Value = 16777312
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || e == nil {
			t.Fatal(r, e)
		}
		bad = shAnswer(306, 2001, false).(*sh.UDA)
		bad.UserData = avp.Optional[[]byte]{}
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrMissing) {
			t.Fatal(r, e)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(ctx, shAnswer(306, 2001, false)); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		if !in.Check(validation.Context{}).Complete {
			t.Fatal(in.Validation())
		}
		if r, e := in.Respond(ctxTest(t), shAnswer(306, 2001, false)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
	})
}
func TestShOpaqueDuplicateArrivals(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		for _, code := range []uint32{307, 309} {
			in, req := sendSh(t, n, p, co, remote, code, code+700)
			dup, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			dup.Header.HopByHop++
			dup.Header.Flags |= pdu.RetransmittedFlag
			proxies := []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255}}}
			if code == 307 {
				dup.Body.(*sh.PUR).ProxyInfo = proxies
			} else {
				dup.Body.(*sh.PNR).ProxyInfo = proxies
			}
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			noDuplicateDelivery(t, p)
			if r, e := in.Respond(ctxTest(t), shAnswer(code, 2001, false)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			for range 2 {
				ans := routeRead(t, co, remote)
				original := req
				if ans.Header.HopByHop == dup.Header.HopByHop {
					original = dup
				}
				if e = co.ValidateAnswer(original, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
					t.Fatal(e)
				}
			}
			dup.Header.HopByHop++
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			duplicateAnswer(t, n, co, remote, dup)
			noDuplicateDelivery(t, p)
			// Whitespace that an XML parser might discard is still a content conflict.
			if code == 307 {
				b := dup.Body.(*sh.PUR)
				b.UserData = append(b.UserData, ' ')
			} else {
				b := dup.Body.(*sh.PNR)
				b.UserData = append(b.UserData, ' ')
			}
			dup.Header.HopByHop++
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			a := duplicateAnswer(t, n, co, remote, dup)
			if a.Body.(*pdu.ErrorAnswer).ResultCode != 5012 {
				t.Fatal(a)
			}
			noDuplicateDelivery(t, p)
		}
	})
}
func TestShBudgetsAndCloseRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := shConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		large := shBody(307).(*sh.PUR)
		large.UserData = make([]byte, 1500)
		before := n.Resources()
		if _, e := p.Submit(ctxTest(t), large, RequestOptions{}, port); e == nil {
			t.Fatal("unbudgeted request")
		}
		if after := n.Resources(); before.Outstanding != after.Outstanding || before.OutboundBytes != after.OutboundBytes {
			t.Fatal(before, after)
		}
		in, _ := sendSh(t, n, p, co, remote, 307, 970)
		over := shRequest(t, n, co, 309, 971)
		if e := writeWire(ctxTest(t), remote, co, over); e != nil {
			t.Fatal(e)
		}
		a := duplicateAnswer(t, n, co, remote, over)
		if a.Body.(*pdu.ErrorAnswer).ResultCode != 3004 {
			t.Fatal(a)
		}
		if r, e := in.Respond(ctxTest(t), shAnswer(307, 2001, false)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		routeRead(t, co, remote)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		in, _ = sendSh(t, n, p, co, remote, 309, 972)
		start := make(chan struct{})
		results := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r, _ := in.Respond(ctxTest(t), shAnswer(309, 2001, false))
				results <- r
			}()
		}
		close(start)
		_ = n.Close()
		wg.Wait()
		close(results)
		commits := 0
		for r := range results {
			if r.Committed {
				commits++
			}
		}
		if commits > 1 {
			t.Fatal("multiple commits")
		}
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
	})
}
func TestShCancelLateAnswerAndNoReplay(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), shBody(307), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id || !errors.Is(r.Result.Err, context.Canceled) {
			t.Fatal(r, e)
		}
		a, e := co.CompleteAnswer(req, shAnswer(307, 2001, false), serverOrigin)
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
		for code := uint32(306); code <= 309; code++ {
			target := serverOrigin
			if code == 309 {
				target = gmscOrigin
			}
			for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
				if _, e = n.NewRoute(RouteConfig{Command: shBody(code).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: mode, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: target, Retention: 2 * time.Second}}); e == nil {
					t.Fatal("unaudited Sh replay")
				}
			}
		}
	})
}
