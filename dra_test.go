// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

// Real TCP/TLS boundary; the opposite endpoint models an adjacent relay, not
// the terminal Origin used in subsequent application answers.
func draHandshake(t *testing.T, n *Node, co *codec.Codec, encrypted, passive bool, auth, acct []uint32, success bool) (*Peer, transport.Conn) {
	t.Helper()
	a, b := networkAdapters(t, encrypted)
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		if !passive {
			req, e := readWire(ctx, b, co)
			if e == nil {
				var ans pdu.Message
				ans, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "relay-test", AuthApplicationID: auth, AcctApplicationID: acct}, draOrigin)
				if e == nil {
					e = writeWire(ctx, b, co, ans)
				}
			}
			ch <- e
			return
		}
		req := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 88, EndToEnd: 89}, Body: &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "relay-test", AuthApplicationID: auth, AcctApplicationID: acct}}
		e := writeWire(ctx, b, co, req)
		if e == nil {
			var ans pdu.Message
			ans, e = readWire(ctx, b, co)
			if e == nil {
				v, ok := ans.Body.(*base.CEA)
				want := uint32(5010)
				if success {
					want = 2001
				}
				if !ok || v.ResultCode != want {
					e = errors.New("wrong CEA application decision")
				}
			}
		}
		ch <- e
	}()
	var p *Peer
	var e error
	if passive {
		p, e = n.Accept(ctx, "dra", a)
	} else {
		p, e = n.Connect(ctx, "dra", a)
	}
	if other := <-ch; other != nil {
		t.Fatal(other)
	}
	if success {
		if e != nil || p == nil || !p.Info().Ready {
			t.Fatalf("ready=%v err=%v", p, e)
		}
	} else if e == nil {
		t.Fatal("unshared applications accepted")
	}
	return p, b
}
func TestDRARelayHandshake(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, passive := range []bool{false, true} {
			for _, tc := range []struct {
				name       string
				auth, acct []uint32
				ok         bool
			}{
				{"relay-auth", []uint32{math.MaxUint32}, nil, true},
				{"relay-acct", nil, []uint32{math.MaxUint32}, true},
				{"concrete", []uint32{appID}, nil, true},
				{"unshared", []uint32{4}, nil, false},
			} {
				name := tc.name
				if passive {
					name += "-passive"
				} else {
					name += "-active"
				}
				t.Run(name, func(t *testing.T) {
					n, co := newTestNode(t)
					p, _ := draHandshake(t, n, co, encrypted, passive, tc.auth, tc.acct, tc.ok)
					if tc.ok {
						n.mu.Lock()
						apps := p.conn.apps
						if !apps.allows(requestBody().CommandKey(), HeaderApplication) || apps.allows(pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}, HeaderApplication) || apps.allows(pdu.CommandKey{Code: 271, Request: true}, SCAP1Accounting) {
							t.Error("relay expanded local authorization")
						}
						n.mu.Unlock()
					}
				})
			}
		}
	})
}
func TestDRARelayVendorAndSCAP1Isolation(t *testing.T) {
	n := &Node{cfg: NodeConfig{Capabilities: LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(appID)}, {VendorID: 193, AcctApplicationID: avp.Some(uint32(19302))}}}}}
	p := &Peer{}
	apps := n.negotiateApplications(p, []uint32{math.MaxUint32}, nil, nil)
	if !apps.byID[appID] || apps.scap1Accounting {
		t.Fatal("local vendor application or SCAP1 boundary")
	}
	apps = n.negotiateApplications(p, nil, nil, []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(uint32(math.MaxUint32))}})
	if apps.byID[appID] {
		t.Fatal("vendor pseudo-relay granted wildcard")
	}
}
func withRequestTargets(c NodeConfig, key pdu.CommandKey) NodeConfig {
	for i := range c.Peers {
		c.Peers[i].RequestTargets = []RequestTarget{
			{ID: "one", Command: key, Destination: pdu.Origin{Host: "mme1.example", Realm: "example"}},
			{ID: "two", Command: key, Destination: pdu.Origin{Host: "mme2.other", Realm: "other"}},
		}
	}
	return c
}
func TestDRATargetConfigurationAndBudget(t *testing.T) {
	key := requestBody().CommandKey()
	basecfg := testConfig(t)
	c := withRequestTargets(testConfig(t), key)
	cost, e := requestTargetReserve(c)
	if e != nil {
		t.Fatal(e)
	}
	if uint64(unsafe.Sizeof(CommandPolicy{})+unsafe.Sizeof(requestTargetKey{})+unsafe.Sizeof(pdu.Origin{})) >= 1024 {
		t.Fatal("target ledger too small")
	}
	n, _ := newInboundNode(t, c)
	plain, _ := newInboundNode(t, basecfg)
	if n.Resources().FixedBytes-plain.Resources().FixedBytes != cost {
		t.Fatal("target reserve not charged")
	}
	c.Peers[0].RequestTargets[0].Destination.Host = "mutated.example"
	c.Peers[0].Commands[0].Sources.ErrorAgents[0].Host = "mutated-dra.example"
	p, ok := n.peers["dra"].requestPolicy(key, "one")
	if !ok || p.DestinationHost != "mme1.example" || p.Sources.ErrorAgents[0] != draOrigin || len(p.Sources.Terminals) != 1 || p.Sources.Terminals[0].Host != p.DestinationHost {
		t.Fatal("unfrozen or overbroad policy")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*NodeConfig)
	}{
		{"duplicate", func(c *NodeConfig) { c.Peers[0].RequestTargets[1].ID = "one" }},
		{"unknown-command", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].Command.Code = 123 }},
		{"answer", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].Command.Request = false }},
		{"base-or-scap", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].Command = pdu.CommandKey{Code: 271, Request: true} }},
		{"empty-id", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].ID = "" }},
		{"long-id", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].ID = strings.Repeat("a", 129) }},
		{"bad-host", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].Destination.Host = "" }},
		{"bad-realm", func(c *NodeConfig) { c.Peers[0].RequestTargets[0].Destination.Realm = "bad realm" }},
		{"count", func(c *NodeConfig) { c.Peers[0].RequestTargets = make([]RequestTarget, 65) }},
		{"one-byte-budget", func(c *NodeConfig) {
			c.Resources.NodeBytes = n.Resources().FixedBytes + c.Resources.Classes[0].Bytes - 1
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := withRequestTargets(testConfig(t), key)
			tc.mutate(&cfg)
			nn, e := NewNode(cfg)
			if e == nil {
				nn.Close()
				nn.Wait(context.Background())
				t.Fatal("invalid target accepted")
			}
		})
	}
}
func TestDRATwoTerminalsAndAnswerAuthority(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		c := withRequestTargets(smsConfig(t, false), smsBody(true).CommandKey())
		n, co := newInboundNode(t, c)
		p, remote := draHandshake(t, n, co, encrypted, false, []uint32{math.MaxUint32}, nil, true)
		port := portTest(t, n, 4)
		before := n.Resources()
		if _, e := p.Submit(ctxTest(t), smsBody(true), RequestOptions{TargetID: "unlisted"}, port); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
		if n.Resources() != before {
			t.Fatal("rejected target consumed quota")
		}
		ids := map[RequestID]string{}
		for _, target := range []string{"one", "two"} {
			id, e := p.Submit(ctxTest(t), smsBody(true), RequestOptions{TargetID: target}, port)
			if e != nil {
				t.Fatal(e)
			}
			ids[id] = target
		}
		reqs := []pdu.Message{routeRead(t, co, remote), routeRead(t, co, remote)}
		seen := map[string]bool{}
		for i := 1; i >= 0; i-- {
			req := reqs[i]
			v := req.Body.(*sgd.TFR)
			host := v.DestinationHost
			seen[host] = true
			origin := pdu.Origin{Host: host, Realm: v.DestinationRealm}
			ans, e := co.CompleteAnswer(req, smsAnswer(true, 2001), origin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
		}
		if !seen["mme1.example"] || !seen["mme2.other"] {
			t.Fatal(seen)
		}
		for range 2 {
			r, e := port.Receive(ctxTest(t))
			if e != nil || r.Result.Err != nil || !r.Result.Validation.Complete || ids[r.RequestID] == "" {
				t.Fatal(r, e)
			}
			delete(ids, r.RequestID)
		}
		// A valid pending key is insufficient: other authorized targets and the DRA
		// cannot answer normal success for this selected terminal.
		for _, origin := range []pdu.Origin{{Host: "mme2.other", Realm: "other"}, draOrigin} {
			id, e := p.Submit(ctxTest(t), smsBody(true), RequestOptions{TargetID: "one"}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, remote)
			ans, e := co.CompleteAnswer(req, smsAnswer(true, 2001), origin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			r, e := port.Receive(ctxTest(t))
			if e != nil || r.RequestID != id || !errors.Is(r.Result.Err, validation.ErrInvalidAnswer) {
				t.Fatal(r, e)
			}
		}
		// The inherited, explicit error-agent permission remains limited to E.
		_, e := p.Submit(ctxTest(t), smsBody(true), RequestOptions{TargetID: "two"}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		ans, e := co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.Result.Err != nil || r.Result.Answer.Header.Flags&pdu.ErrorFlag == 0 {
			t.Fatal(r, e)
		}
		waitFor(t, func() bool {
			return n.Resources().Outstanding == 0 && n.Resources().Pending == 0 && n.Resources().OutboundBytes == 0
		})
	})
}
func TestDRATargetCancelCloseAndLateAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, withRequestTargets(testConfig(t), requestBody().CommandKey()))
		p, remote := draHandshake(t, n, co, encrypted, false, []uint32{math.MaxUint32}, nil, true)
		port := portTest(t, n, 2)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{TargetID: "one"}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		waitFor(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			return !n.requests[id].writing && n.requests[id].delivery.MayHaveReachedPeer
		})
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id || !errors.Is(r.Result.Err, context.Canceled) {
			t.Fatal(r, e)
		}
		a, e := co.CompleteAnswer(req, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, pdu.Origin{Host: "mme1.example", Realm: "example"})
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		id, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{TargetID: "two"}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, remote)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = n.Cancel(id) }()
		go func() { defer wg.Done(); _ = p.Close() }()
		wg.Wait()
		r, e = port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id {
			t.Fatal(r, e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 && n.Resources().Pending == 0 })
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(timeout); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
	})
}
func TestDRATargetRouteRetransmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		c := withRequestTargets(routeConfig(t), requestBody().CommandKey())
		n, co := newInboundNode(t, c)
		p, a := duplicateConnect(t, n, co, encrypted, "dra")
		_, b := duplicateConnect(t, n, co, encrypted, "dra2")
		cfg := RouteConfig{TargetID: "two", Command: requestBody().CommandKey(), Peers: []PeerID{"dra", "dra2"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: 2 * time.Second, ReplayDomain: ReplayDomain{Target: c.Peers[0].RequestTargets[1].Destination, Retention: 5 * time.Second}}
		r, e := n.NewRoute(cfg)
		if e != nil {
			t.Fatal(e)
		}
		bad := cfg
		bad.ReplayDomain.Target = serverOrigin
		if _, e = n.NewRoute(bad); e == nil {
			t.Fatal("target/replay mismatch accepted")
		}
		port := portTest(t, n, 1)
		if _, e = r.Submit(ctxTest(t), requestBody(), RequestOptions{TargetID: "one"}, port); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		routeSuspect(t, n, p, a, co)
		info, e := n.RequestInfo(id)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, b)
		x, _ := co.Append(nil, first)
		y, _ := co.Append(nil, second)
		if first.Header.EndToEnd != second.Header.EndToEnd || second.Header.Flags&pdu.RetransmittedFlag == 0 || !bytes.Equal(x[20:], y[20:]) || second.Body.(*s6c.SRR).DestinationHost.Value != "mme2.other" {
			t.Fatal("replay identity or destination drift")
		}
		ans, e := co.CompleteAnswer(second, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, cfg.ReplayDomain.Target)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), b, co, ans); e != nil {
			t.Fatal(e)
		}
		result := routeCompletion(t, n, port, id)
		if result.Err != nil || len(result.Attempts) != 2 {
			t.Fatal(result)
		}
	})
}
