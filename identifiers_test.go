// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type idSource func() (uint32, error)

func (f idSource) NextEndToEnd() (uint32, error) { return f() }
func customIdentifiers(c NodeConfig, src EndToEndSource, slots int) NodeConfig {
	c.Identifiers.EndToEnd = src
	c.Resources.IdentifierGuardSlots = slots
	c.Resources.IdentifierBytes = 4096 + uint64(slots)*128 + 256*uint64(len(c.Resources.Classes))
	c.Resources.IdentifierClasses = []IdentifierClass{{"sms", slots}}
	return c
}
func TestIdentifierConfigurationAndOwnedInfo(t *testing.T) {
	n, _ := newInboundNode(t, testConfig(t))
	v := n.IdentifierInfo()
	if v.Space.CounterBits != 32 || v.ApplicationBits != 31 || v.ReuseGuard != 240*time.Second || v.MaxSessionIDBytes != 512 || v.GuardSlots != 0 {
		t.Fatal(v)
	}
	for _, mutate := range []func(*NodeConfig){
		func(c *NodeConfig) { c.Identifiers.Space = SequenceSpace{1, 32} }, func(c *NodeConfig) { c.Identifiers.Space.CounterBits = 1 }, func(c *NodeConfig) { c.Identifiers.ReuseGuard = time.Second }, func(c *NodeConfig) { c.Identifiers.MaxSessionIDBytes = 10 }, func(c *NodeConfig) { c.Identifiers.MaxSessionIDBytes = 4097 },
		func(c *NodeConfig) { c.Identifiers.EndToEnd = idSource(func() (uint32, error) { return 1, nil }) }, func(c *NodeConfig) { c.Resources.IdentifierGuardSlots = 1 },
	} {
		cfg := testConfig(t)
		mutate(&cfg)
		if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
			if bad != nil {
				_ = bad.Close()
			}
			t.Fatal(e)
		}
	}
	cfg := customIdentifiers(testConfig(t), idSource(func() (uint32, error) { return 1, nil }), 4)
	exact, _ := newInboundNode(t, cfg)
	info := exact.IdentifierInfo()
	info.Classes[0].Slots = 100
	cfg.Resources.IdentifierClasses[0].Slots = 100
	if exact.IdentifierInfo().Classes[0].Slots != 4 {
		t.Fatal("mutable config")
	}
	cfg = customIdentifiers(testConfig(t), idSource(func() (uint32, error) { return 1, nil }), 4)
	cfg.Resources.IdentifierBytes--
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			_ = bad.Close()
		}
		t.Fatal(e)
	}
}
func TestIdentifierSequencePartitionRolloverAndGuard(t *testing.T) {
	cfg := testConfig(t)
	cfg.Identifiers.Space = SequenceSpace{Prefix: 9, CounterBits: 2}
	n, _ := newInboundNode(t, cfg)
	n.mu.Lock()
	defer n.mu.Unlock()
	var app, ctrl []*identifierLease
	seen := map[uint32]bool{}
	for _, control := range []bool{false, true} {
		for range 2 {
			id, e := n.nextE2E(control)
			if e != nil || seen[id] || id < 36 || id > 39 {
				t.Fatal(id, e)
			}
			if (id&2 != 0) != control {
				t.Fatal(id)
			}
			seen[id] = true
			l := &identifierLease{control: control, slot: -1}
			if control {
				ctrl = append(ctrl, l)
			} else {
				app = append(app, l)
			}
		}
	}
	for _, control := range []bool{false, true} {
		if _, e := n.nextE2E(control); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
	}
	for _, l := range app {
		n.retireIdentifier(l)
	}
	if _, e := n.nextE2E(false); !errors.Is(e, ErrIdentifiers) {
		t.Fatal("reuse before guard", e)
	}
	n.identifiers.appRetired = time.Now().Add(-n.identifiers.config.ReuseGuard - time.Second)
	id, e := n.nextE2E(false)
	if e != nil || !seen[id] {
		t.Fatal(id, e)
	}
	if _, e = n.nextE2E(true); !errors.Is(e, ErrIdentifiers) {
		t.Fatal("borrowed application half", e)
	}
	n.retireIdentifier(&identifierLease{slot: -1})
	for _, l := range ctrl {
		n.retireIdentifier(l)
	}
}
func TestIdentifierSessionOwnershipAndExhaustion(t *testing.T) {
	n, _ := newInboundNode(t, testConfig(t))
	var wg sync.WaitGroup
	ids := make(chan string, 128)
	for range 128 {
		wg.Go(func() {
			s, e := n.NewSessionID()
			if e != nil {
				t.Error(e)
			}
			ids <- s
		})
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{}
	for s := range ids {
		parts := strings.Split(s, ";")
		if len(parts) != 4 || parts[0] != n.cfg.Origin.Host || len(parts[3]) != 32 || seen[s] {
			t.Fatal(s)
		}
		if _, e := strconv.ParseUint(parts[2], 10, 32); e != nil {
			t.Fatal(e)
		}
		seen[s] = true
	}
	n.mu.Lock()
	n.identifiers.session = ^uint64(0)
	n.mu.Unlock()
	if _, e := n.NewSessionID(); !errors.Is(e, ErrIdentifiers) {
		t.Fatal(e)
	}
	_ = n.Close()
	if _, e := n.NewSessionID(); !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
}
func TestIdentifierGuardIndexAndExpiryOracle(t *testing.T) {
	cfg := customIdentifiers(testConfig(t), idSource(func() (uint32, error) { return 0, nil }), 31)
	s, e := compileIdentifiers(cfg.Identifiers, cfg.Resources, cfg.Origin.Host)
	if e != nil {
		t.Fatal(e)
	}
	n := &Node{identifiers: s}
	live := map[uint32]*identifierLease{}
	r := rand.New(rand.NewSource(215))
	now := time.Now()
	for range 6000 {
		if len(live) > 0 && r.Intn(3) == 0 {
			for id, l := range live {
				n.retireIdentifier(l)
				delete(live, id)
				break
			}
			n.identifiers.expire(now.Add(500 * time.Second))
		} else {
			l, e := n.identifiers.reserve("sms", now)
			if e != nil {
				continue
			}
			n.identifiers.sourceBusy = false
			id := uint32(r.Intn(64))
			if r.Intn(2) == 0 {
				id *= 64
			} // mixed homes plus deliberately colliding chains
			e = n.identifiers.install(l, id)
			if _, ok := live[id]; ok {
				if !errors.Is(e, ErrIdentifierCollision) {
					t.Fatal(e)
				}
				n.retireIdentifier(l)
			} else {
				if e != nil {
					t.Fatal(e)
				}
				live[id] = l
			}
		}
		for id, l := range live {
			i, _ := n.identifiers.find(id)
			if i != l.slot {
				t.Fatal(id, i, l.slot)
			}
		}
	}
	for _, l := range live {
		n.retireIdentifier(l)
	}
	n.identifiers.expire(now.Add(500 * time.Second))
	if n.identifiers.classes[0].Used != 0 || len(n.identifiers.expiry) != 0 {
		t.Fatal(n.identifiers.classes)
	}
}
func TestIdentifierCustomCollisionRangeAndFailure(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		var candidate atomic.Uint32
		candidate.Store(7)
		var calls atomic.Int32
		var fail atomic.Bool
		src := idSource(func() (uint32, error) {
			calls.Add(1)
			if fail.Load() {
				return 0, io.EOF
			}
			return candidate.Load(), nil
		})
		n, co := newInboundNode(t, customIdentifiers(inboundConfig(t), src, 4))
		a, b := networkAdapters(t, tls)
		p := connectInboundTransport(t, n, co, a, b)
		port := portTest(t, n, 8)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, b)
		if req.Header.EndToEnd != 7 {
			t.Fatal(req.Header)
		}
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifierCollision) {
			t.Fatal(e)
		}
		candidate.Store(0x80000000)
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifierRange) {
			t.Fatal(e)
		}
		fail.Store(true)
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifierSourceFailed) || !errors.Is(e, io.EOF) {
			t.Fatal(e)
		}
		fail.Store(false)
		candidate.Store(7)
		routeReply(t, co, b, req)
		waitFor(t, func() bool { return n.IdentifierInfo().Quarantine == 1 })
		// Completed but unconsumed results do not extend the identifier window.
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifierCollision) {
			t.Fatal(e)
		}
		if c, e := port.Receive(ctxTest(t)); e != nil || c.RequestID != id {
			t.Fatal(c, e)
		}
		if n.IdentifierInfo().Quarantine != 1 || n.IdentifierInfo().Reserved != 0 {
			t.Fatal(n.IdentifierInfo())
		}
		n.mu.Lock()
		n.identifiers.expire(time.Now().Add(n.identifiers.config.ReuseGuard))
		n.mu.Unlock()
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, b, routeRead(t, co, b))
		if _, e = port.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if calls.Load() != 6 {
			t.Fatal(calls.Load())
		}
	})
}
func TestIdentifierSourceBusyCloseAndWait(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		var calls atomic.Int32
		n, co := newInboundNode(t, customIdentifiers(inboundConfig(t), idSource(func() (uint32, error) { calls.Add(1); close(entered); <-release; return 3, nil }), 2))
		a, b := networkAdapters(t, tls)
		p := connectInboundTransport(t, n, co, a, b)
		port := portTest(t, n, 4)
		done := make(chan error, 1)
		go func() { _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); done <- e }()
		<-entered
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
		if n.IdentifierInfo().Reserved != 1 || calls.Load() != 1 {
			t.Fatal(n.IdentifierInfo())
		}
		_ = n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if e := n.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		unblock()
		if e := <-done; !errors.Is(e, ErrClosed) {
			t.Fatal(e)
		}
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 || n.IdentifierInfo().GuardSlots != 0 {
			t.Fatal(n.Resources(), n.IdentifierInfo())
		}
	})
}

func TestIdentifierClassQuarantineIsolationAndControl(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := twoDuplicatePeers(testConfig(t))
		cfg.Resources.TransportBytes = 128 << 20
		cfg.Resources.NodeBytes = 512 << 20
		cfg.Resources.Classes = []ClassConfig{{"a", 32 << 20, 8}, {"b", 32 << 20, 8}}
		var calls atomic.Uint32
		cfg = customIdentifiers(cfg, idSource(func() (uint32, error) { return calls.Add(1), nil }), 4)
		cfg.Resources.IdentifierClasses = []IdentifierClass{{"a", 2}, {"b", 2}}
		for i, class := range []ClassID{"a", "b"} {
			cfg.Peers[i].Commands = append([]CommandPolicy(nil), cfg.Peers[i].Commands...)
			cfg.Peers[i].Commands[0].Class = class
		}
		n, co := newInboundNode(t, cfg)
		p1, r1 := duplicateConnect(t, n, co, tls, "dra")
		p2, r2 := duplicateConnect(t, n, co, tls, "dra2")
		port := portTest(t, n, 8)
		for range 2 {
			if _, e := p1.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
				t.Fatal(e)
			}
			routeReply(t, co, r1, routeRead(t, co, r1))
			if _, e := port.Receive(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
		}
		if _, e := p1.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
		if calls.Load() != 2 {
			t.Fatal("source called before guard admission")
		}
		if _, e := p2.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, r2, routeRead(t, co, r2))
		if _, e := port.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		n.mu.Lock()
		p1.conn.tw = time.Now()
		n.mu.Unlock()
		n.tick(time.Now())
		dwr := routeRead(t, co, r1)
		if dwr.Header.EndToEnd&0x80000000 == 0 || calls.Load() != 3 {
			t.Fatal(dwr.Header, calls.Load())
		}
		dwa, e := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), r1, co, dwa); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.IdentifierInfo().ControlActive == 0 })
		v := n.IdentifierInfo()
		if v.Quarantine != 3 || v.Free != 1 || v.Classes[0].Used != 2 || v.Classes[1].Used != 1 {
			t.Fatal(v)
		}
	})
}
func TestIdentifierLateWriterRetainsActiveGuard(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, customIdentifiers(inboundConfig(t), idSource(func() (uint32, error) { return 10, nil }), 1))
		p, remote, barrier, release := drrBlockedWriter(t, n, co, tls)
		port := portTest(t, n, 2)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		waitBarrier(t, barrier)
		routeReply(t, co, remote, req)
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return n.requests[id].selected })
		if v := n.IdentifierInfo(); v.Active != 1 || v.Quarantine != 0 {
			t.Fatal(v)
		}
		_ = n.Cancel(id)
		if _, e = p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
		release()
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
		if v := n.IdentifierInfo(); v.Active != 0 || v.Quarantine != 1 {
			t.Fatal(v)
		}
	})
}
func TestIdentifierOriginalIdentityRouteKeepsOneGuard(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		var calls atomic.Uint32
		n, co := newInboundNode(t, customIdentifiers(routeConfig(t), idSource(func() (uint32, error) { calls.Add(1); return 17, nil }), 1))
		_, r1 := duplicateConnect(t, n, co, tls, "dra")
		_, r2 := duplicateConnect(t, n, co, tls, "dra2")
		route := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 2)
		id, e := route.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, r1)
		_ = r1.Close()
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
		if v := n.IdentifierInfo(); v.Active != 1 || v.Quarantine != 0 {
			t.Fatal(v)
		}
		if _, e = n.Retransmit(ctxTest(t), id, AttemptID{id, 1}); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, r2)
		routeSameRequest(t, co, first, second)
		if calls.Load() != 1 || n.IdentifierInfo().Active != 1 {
			t.Fatal(calls.Load(), n.IdentifierInfo())
		}
		routeReply(t, co, r2, second)
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
		if n.IdentifierInfo().Quarantine != 1 {
			t.Fatal(n.IdentifierInfo())
		}
	})
}

type idControlBarrier struct {
	transport.Conn
	entered, release chan struct{}
	once             sync.Once
}

func (b *idControlBarrier) WriteFrame(ctx context.Context, f []byte) transport.WriteResult {
	r := b.Conn.WriteFrame(ctx, f)
	if len(f) >= 20 && f[5] == 0 && f[6] == 1 && f[7] == 24 && f[4]&pdu.RequestFlag != 0 {
		b.once.Do(func() { close(b.entered); <-b.release })
	}
	return r
}
func TestIdentifierControlRetiresAfterAnswerAndWriter(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &idControlBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		release := func() { once.Do(func() { close(barrier.release) }) }
		defer release()
		p := connectInboundTransport(t, n, co, barrier, b)
		if n.IdentifierInfo().ControlActive != 0 {
			t.Fatal("CER lease survived handshake")
		}
		n.mu.Lock()
		p.conn.tw = time.Now()
		n.mu.Unlock()
		n.tick(time.Now())
		dwr := routeRead(t, co, b)
		<-barrier.entered
		dwa, e := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), b, co, dwa); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn.probe == nil })
		if n.IdentifierInfo().ControlActive != 1 {
			t.Fatal("released borrowed control ID")
		}
		_ = p.Close()
		if n.IdentifierInfo().ControlActive != 1 {
			t.Fatal("Close released borrowed control ID")
		}
		release()
		if e = p.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.IdentifierInfo().ControlActive != 0 {
			t.Fatal(n.IdentifierInfo())
		}
	})
}
func TestIdentifierHBHExhaustionDrainsWithControlReserve(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		a, b := networkAdapters(t, tls)
		p := connectInboundTransport(t, n, co, a, b)
		port := portTest(t, n, 4)
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, b)
		n.mu.Lock()
		p.conn.hbhCount = (1 << 32) - 1024
		n.mu.Unlock()
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
		if !p.Info().IdentifierDraining || p.Info().Ready || p.Info().Lifecycle != Active {
			t.Fatal(p.Info())
		}
		routeReply(t, co, b, req)
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
		dpr := routeRead(t, co, b)
		if dpr.Header.CommandCode != 282 {
			t.Fatal(dpr)
		}
		dpa, e := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), b, co, dpa); e != nil {
			t.Fatal(e)
		}
		if e = p.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if p.Info().Lifecycle != Active || n.IdentifierInfo().ControlActive != 0 {
			t.Fatal(p.Info(), n.IdentifierInfo())
		}
	})
}
func TestIdentifierPreparationFailureReleasesGuard(t *testing.T) {
	n, co := newInboundNode(t, customIdentifiers(inboundConfig(t), idSource(func() (uint32, error) { return 11, nil }), 1))
	p, _, _ := connectTest(t, n, co)
	port := portTest(t, n, 2)
	body := requestBody()
	body.DestinationRealm = "other.example"
	if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); e == nil {
		t.Fatal("conflicting binding accepted")
	}
	if v := n.IdentifierInfo(); v.Free != 1 || v.Quarantine != 0 || v.Active != 0 {
		t.Fatal(v)
	}
}

func TestIdentifierSmallSpaceOnWireAndInboundEcho(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := inboundConfig(t)
		cfg.Identifiers.Space = SequenceSpace{Prefix: 5, CounterBits: 2}
		n, co := newInboundNode(t, cfg)
		a, b := networkAdapters(t, tls)
		p := connectInboundTransport(t, n, co, a, b)
		port := portTest(t, n, 4)
		seen := map[uint32]bool{}
		for range 2 {
			body := requestBody()
			body.SessionID, _ = n.NewSessionID()
			if _, e := p.Submit(ctxTest(t), body, RequestOptions{}, port); e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, b)
			if req.Header.EndToEnd < 20 || req.Header.EndToEnd > 21 || seen[req.Header.EndToEnd] {
				t.Fatal(req.Header)
			}
			seen[req.Header.EndToEnd] = true
			routeReply(t, co, b, req)
			if _, e := port.Receive(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
		}
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifiers) {
			t.Fatal(e)
		}
		before := n.IdentifierInfo()
		in, req := receiveInbound(t, n, p, b, co, 999)
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		answer := routeRead(t, co, b)
		if !pdu.MatchesAnswerHeader(req.Header, answer.Header) || n.IdentifierInfo().ApplicationAllocated != before.ApplicationAllocated {
			t.Fatal(answer.Header)
		}
		if n.IdentifierInfo().ApplicationActive != 0 {
			t.Fatal(n.IdentifierInfo())
		}
	})
}
func TestIdentifierCancelAdmissionReleasesReservedSlot(t *testing.T) {
	n, co := newInboundNode(t, customIdentifiers(inboundConfig(t), idSource(func() (uint32, error) { return 3, nil }), 1))
	p, _, _ := connectTest(t, n, co)
	port := portTest(t, n, 2)
	ctx, cancel := context.WithCancel(context.Background())
	// Reentering a read-only API proves the source executes outside the state gate.
	n.mu.Lock()
	n.identifiers.config.EndToEnd = idSource(func() (uint32, error) { _ = n.IdentifierInfo(); cancel(); return 3, nil })
	n.mu.Unlock()
	if _, e := p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if v := n.IdentifierInfo(); v.Free != 1 || v.Reserved != 0 || v.Active != 0 {
		t.Fatal(v)
	}
}
func TestIdentifierExactExpiryDoesNotDependOnCompletionConsumption(t *testing.T) {
	cfg := customIdentifiers(testConfig(t), idSource(func() (uint32, error) { return 4, nil }), 1)
	s, e := compileIdentifiers(cfg.Identifiers, cfg.Resources, cfg.Origin.Host)
	if e != nil {
		t.Fatal(e)
	}
	n := &Node{identifiers: s}
	l, e := n.identifiers.reserve("sms", time.Now())
	if e != nil {
		t.Fatal(e)
	}
	n.identifiers.sourceBusy = false
	if e = n.identifiers.install(l, 4); e != nil {
		t.Fatal(e)
	}
	n.retireIdentifier(l)
	deadline := n.identifiers.slots[l.slot].reuseAfter
	n.identifiers.expire(deadline.Add(-time.Nanosecond))
	if n.identifiers.free != -1 {
		t.Fatal("early guard reuse")
	}
	n.identifiers.expire(deadline)
	if n.identifiers.free != l.slot {
		t.Fatal("expired slot not free")
	}
	if allocs := testing.AllocsPerRun(100, func() { n.identifiers.expire(deadline) }); allocs != 0 {
		t.Fatal(allocs)
	}
}
