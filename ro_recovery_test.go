// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func roRecoveryConfig(t *testing.T, family string) NodeConfig {
	var c NodeConfig
	switch family {
	case "standard":
		c = roConfig(t)
	case "sms":
		c = legacyConfig(t)
	case "money":
		c = refundConfig(t)
	default:
		c = combinedConfig(t)
	}
	c = twoDuplicatePeers(c)
	for i := range c.Peers {
		c.Peers[i].Commands = append([]CommandPolicy(nil), c.Peers[i].Commands...)
	}
	return c
}
func roRecoveryBody(t *testing.T, family, kind string) pdu.Body {
	switch family {
	case "standard":
		if kind == "iec" {
			return roBody()
		}
		return roECURBody(kind == "term")
	case "sms":
		if kind == "iec" {
			return legacyBody()
		}
		return legacyECURBody(kind == "term")
	case "money":
		if kind == "refund" {
			return refundBody(t)
		}
		r := generalBody(t)
		if kind == "balance" {
			r.RequestedAction = 2
		}
		return r
	default:
		return combinedBody(t, kind)
	}
}
func roRecoveryAnswer(q pdu.Body, code uint32) pdu.Body {
	switch q := q.(type) {
	case *ro.CCR:
		a := roAnswer(code)
		if q.CCRequestType != 4 {
			a = roECURAnswer(q.CCRequestType == 3, code)
		}
		return a
	case *smsfwro.CCR:
		a := legacyAnswer(code)
		a.CCRequestType = q.CCRequestType
		a.CCRequestNumber = q.CCRequestNumber
		return a
	case *legacy420.CCR:
		a := &legacy420.CCA{AuthApplicationID: 4, ResultCode: code, CCRequestType: q.CCRequestType, CCRequestNumber: q.CCRequestNumber}
		if q.RequestedAction == 2 && code == 2001 {
			a.CheckBalanceResult = avp.Some(uint32(0))
		}
		return a
	case *combined.CCR:
		return combinedAnswer(q, code)
	}
	panic("request")
}
func roRecoveryRoute(t *testing.T, n *Node, mode ExchangePolicy) *Route {
	t.Helper()
	r, e := n.NewRoute(RouteConfig{Command: pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}, Peers: []PeerID{"dra", "dra2"}, ExchangePolicy: mode, MaxAttempts: 2, MaxReplayAge: 2 * time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 5 * time.Second}})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func roRecoveryConnect(t *testing.T, n *Node, co *codec.Codec, id PeerID, a, b transport.Conn) *Peer {
	t.Helper()
	ch := make(chan error, 1)
	go func() {
		req, e := readWire(ctxTest(t), b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "ro-route", AuthApplicationID: []uint32{4}}, n.peers[id].origin)
			if e == nil {
				e = writeWire(ctxTest(t), b, co, ans)
			}
		}
		ch <- e
	}()
	p, e := n.Connect(ctxTest(t), id, a)
	if e != nil {
		t.Fatal(e)
	}
	if e = <-ch; e != nil {
		t.Fatal(e)
	}
	return p
}
func roRecoveryNetwork(t *testing.T, n *Node, co *codec.Codec, id PeerID, encrypted bool) (*Peer, transport.Conn) {
	t.Helper()
	a, b := networkAdapters(t, encrypted)
	return roRecoveryConnect(t, n, co, id, a, b), b
}
func roRecoveryReply(t *testing.T, co *codec.Codec, remote transport.Conn, r pdu.Message, code uint32) {
	t.Helper()
	a, e := co.CompleteAnswer(r, roRecoveryAnswer(r.Body, code), serverOrigin)
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), remote, co, a); e != nil {
		t.Fatal(e)
	}
}
func roAwait(t *testing.T, n *Node, id RequestID) RequestInfo {
	t.Helper()
	waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
	i, e := n.RequestInfo(id)
	if e != nil {
		t.Fatal(e)
	}
	return i
}
func roQuiet(t *testing.T, remote transport.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("unexpected automatic send", e)
	}
}

func TestRoRecoveryFixedTargetConfiguration(t *testing.T) {
	for _, family := range []string{"standard", "sms", "money", "combined"} {
		t.Run(family, func(t *testing.T) {
			n, _ := newInboundNode(t, roRecoveryConfig(t, family))
			cfg := roRecoveryRoute(t, n, ExplicitRetransmit).cfg
			for _, change := range []func(*RouteConfig){func(c *RouteConfig) { c.ReplayDomain = ReplayDomain{} }, func(c *RouteConfig) { c.ReplayDomain.Target.Host = "other-ocs.example" }, func(c *RouteConfig) { c.ReplayDomain.Retention = time.Second }, func(c *RouteConfig) { c.MaxAttempts = 17 }, func(c *RouteConfig) { c.MaxReplayAge = 0 }} {
				bad := cfg
				change(&bad)
				if _, e := n.NewRoute(bad); e == nil {
					t.Fatal(bad)
				}
			}
			for _, why := range []string{"realm-only", "different-target", "multiple-terminal-sources"} {
				c := roRecoveryConfig(t, family)
				switch why {
				case "realm-only":
					c.Peers[0].Commands[0].DestinationHost = ""
				case "different-target":
					c.Peers[1].Commands[0].DestinationHost = "other-ocs.example"
				case "multiple-terminal-sources":
					c.Peers[1].Commands[0].Sources.Terminals = []pdu.Origin{serverOrigin, {Host: "other-ocs.example", Realm: serverOrigin.Realm}}
				}
				other, _ := newInboundNode(t, c)
				if _, e := other.NewRoute(cfg); !errors.Is(e, ErrConfig) {
					t.Fatal(why, e)
				}
			}
		})
	}
}
func TestRoRecoveryAllProfilesAndUses(t *testing.T) {
	for _, family := range []string{"standard", "sms", "money", "combined"} {
		t.Run(family, func(t *testing.T) {
			eachNetwork(t, func(t *testing.T, encrypted bool) {
				kinds := []string{"iec", "initial", "term"}
				if family == "money" {
					kinds = []string{"general", "balance", "refund"}
				}
				if family == "combined" {
					kinds = append(kinds, "general", "balance", "refund")
				}
				for _, mode := range []ExchangePolicy{ExplicitRetransmit, TransportFailover} {
					for _, kind := range kinds {
						t.Run(kind+map[ExchangePolicy]string{ExplicitRetransmit: "-explicit", TransportFailover: "-auto"}[mode], func(t *testing.T) {
							n, co := newInboundNode(t, roRecoveryConfig(t, family))
							_, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
							_, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
							r := roRecoveryRoute(t, n, mode)
							port := portTest(t, n, 1)
							body := roRecoveryBody(t, family, kind)
							id, e := r.Submit(ctxTest(t), body, RequestOptions{}, port)
							if e != nil {
								t.Fatal(e)
							}
							first := routeRead(t, co, a)
							switch q := body.(type) {
							case *ro.CCR:
								q.CCRequestNumber = 99
								q.SessionID = "mutated"
							case *smsfwro.CCR:
								q.CCRequestNumber = 99
								q.SessionID = "mutated"
							case *legacy420.CCR:
								q.RequestedAction = 3
								q.RefundInformation.Value = []byte("mutated")
							case *combined.CCR:
								q.CCRequestNumber = 99
								q.SessionID = "mutated"
							}
							before, _ := n.RequestInfo(id)
							reserved := n.Resources().OutboundBytes
							n.mu.Lock()
							x := n.requests[id]
							enough := x.charge.Capacity >= 2*x.answerBound
							n.mu.Unlock()
							if !enough {
								t.Fatal("missing answer reserves")
							}
							if mode == ExplicitRetransmit {
								if _, e = n.Retransmit(ctxTest(t), id, before.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
									t.Fatal("business timer authorized retransmission", e)
								}
							}
							_ = a.Close()
							if mode == ExplicitRetransmit {
								info := roAwait(t, n, id)
								n.mu.Lock()
								queued := len(n.peers["dra2"].conn.pending)
								n.mu.Unlock()
								if queued != 0 {
									t.Fatal("explicit mode sent automatically")
								}
								if !info.Deadline.Equal(before.Deadline) || !info.ReplayEligible {
									t.Fatal(info)
								}
								if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
									t.Fatal(e)
								}
							}
							second := routeRead(t, co, b)
							routeSameRequest(t, co, first, second)
							if got := n.Resources(); got.OutboundBytes != reserved || got.Outstanding != 1 {
								t.Fatal("logical reserve changed", got)
							}
							roRecoveryReply(t, co, b, second, 2001)
							got := routeCompletion(t, n, port, id)
							if got.Outcome != AnswerReceived || !got.Validation.Complete || got.SelectedAttempt.Number != 2 || len(got.Attempts) != 2 || got.Attempts[0].ConnectionID == got.Attempts[1].ConnectionID {
								t.Fatal(got)
							}
							if _, e = n.Retransmit(ctxTest(t), id, before.CurrentAttempt); !errors.Is(e, ErrRequestNotActive) {
								t.Fatal(e)
							}
							if n.Resources().OutboundBytes != 0 {
								t.Fatal(n.Resources())
							}
						})
					}
				}
			})
		})
	}
}
func TestRoRecoveryAnswersNeverAuthorizeRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, mode := range []ExchangePolicy{ExplicitRetransmit, TransportFailover} {
			for _, kind := range []string{"denied", "rating", "version", "e-busy", "conflict", "service-denied", "ddfh"} {
				n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
				_, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
				_, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
				r := roRecoveryRoute(t, n, mode)
				port := portTest(t, n, 1)
				id, e := r.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, a)
				body := combinedAnswer(req.Body.(*combined.CCR), 4012)
				switch kind {
				case "rating":
					body.ResultCode = 5031
					body.FailedAVP = ratingDiagnostic()
				case "version":
					body.ResultCode = 5011
				case "conflict", "service-denied":
					body.ResultCode = 2001
				case "ddfh":
					body.DirectDebitingFailureHandling = avp.Some(uint32(1))
				}
				if kind == "service-denied" {
					body.MultipleServicesCreditControl = avp.Some(combined.ServiceResult{ResultCode: 4012})
				}
				var reply pdu.Body = body
				if kind == "e-busy" {
					reply = &pdu.ErrorAnswer{ResultCode: 3004}
				}
				ans, e := co.CompleteAnswer(req, reply, serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if kind == "conflict" {
					sendRejectRaw(t, a, resultWire(t, co, ans, true, 10415, true))
				} else if e = writeWire(ctxTest(t), a, co, ans); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || len(got.Attempts) != 1 {
					t.Fatal(kind, got)
				}
				if kind == "conflict" && (got.Validation.Complete || !errors.Is(got.Validation.Err(), validation.ErrResultConflict)) {
					t.Fatal(got)
				}
				_ = a.Close()
				roQuiet(t, b)
			}
		}
	})
}
func TestRoRecoveryOldAnswerWhileAwaitingOrAfterSuccessor(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, after := range []bool{false, true} {
			n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
			p, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
			_, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), combinedBody(t, "refund"), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			first := routeRead(t, co, a)
			routeSuspect(t, n, p, a, co)
			info := roAwait(t, n, id)
			var second pdu.Message
			if after {
				if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
					t.Fatal(e)
				}
				second = routeRead(t, co, b)
				routeSameRequest(t, co, first, second)
			}
			ans, e := co.CompleteAnswer(first, combinedAnswer(first.Body.(*combined.CCR), 2001), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			sendRejectRaw(t, a, resultWire(t, co, ans, true, 10415, true))
			got := routeCompletion(t, n, port, id)
			if got.SelectedAttempt.Number != 1 || got.Validation.Complete || !errors.Is(got.Validation.Err(), validation.ErrResultConflict) {
				t.Fatal(got)
			}
			if after {
				roRecoveryReply(t, co, b, second, 2001)
			}
			quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			if _, e = port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("second completion", e)
			}
			cancel()
		}
	})
}

// Only App4/272 traffic is intercepted; Base handshake/watchdog stays real.
type roRecoveryBarrier struct {
	transport.Conn
	write            bool
	entered, release chan struct{}
	once             sync.Once
}

func (b *roRecoveryBarrier) WriteFrame(ctx context.Context, p []byte) transport.WriteResult {
	r := b.Conn.WriteFrame(ctx, p)
	if b.write && len(p) >= 20 && wire.Uint24(p[5:8]) == 272 && p[4]&pdu.RequestFlag != 0 {
		b.once.Do(func() { close(b.entered); <-b.release })
	}
	return r
}
func (b *roRecoveryBarrier) ReadFrameInto(ctx context.Context, p []byte) error {
	e := b.Conn.ReadFrameInto(ctx, p)
	if !b.write && e == nil && len(p) >= 20 && wire.Uint24(p[5:8]) == 272 && p[4]&pdu.RequestFlag == 0 {
		b.once.Do(func() { close(b.entered); <-b.release })
	}
	return e
}

type roRecoveryFault struct {
	transport.Conn
	bytes     int
	uncertain bool
}

func (f *roRecoveryFault) WriteFrame(ctx context.Context, p []byte) transport.WriteResult {
	if len(p) >= 20 && wire.Uint24(p[5:8]) == 272 && p[4]&pdu.RequestFlag != 0 {
		return transport.WriteResult{Bytes: f.bytes, Uncertain: f.uncertain, Err: io.ErrUnexpectedEOF}
	}
	return f.Conn.WriteFrame(ctx, p)
}
func TestRoRecoveryIndependentAnswerBudgets(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
		a, b := networkAdapters(t, encrypted)
		barrier := &roRecoveryBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := roRecoveryConnect(t, n, co, "dra", barrier, b)
		_, other := roRecoveryNetwork(t, n, co, "dra2", encrypted)
		r := roRecoveryRoute(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), combinedBody(t, "initial"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, b)
		reserved := n.Resources().OutboundBytes
		routeSuspect(t, n, p, b, co)
		info := roAwait(t, n, id)
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, other)
		roRecoveryReply(t, co, b, first, 4012)
		<-barrier.entered
		roRecoveryReply(t, co, other, second, 2001)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		n.mu.Lock()
		queued := port.count
		n.mu.Unlock()
		if queued != 0 || n.Resources().OutboundBytes != reserved {
			t.Fatal("borrowed answer released early", n.Resources())
		}
		release.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 2 || len(got.Attempts) != 2 || got.Err != nil {
			t.Fatal(got)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRoRecoveryWriterSettlementBarrier(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, oldWriter := range []bool{true, false} {
			n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
			a, b := networkAdapters(t, encrypted)
			barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
			var release sync.Once
			defer release.Do(func() { close(barrier.release) })
			var p *Peer
			var firstRemote, secondRemote transport.Conn
			if oldWriter {
				p = roRecoveryConnect(t, n, co, "dra", barrier, b)
				firstRemote = b
				_, secondRemote = roRecoveryNetwork(t, n, co, "dra2", encrypted)
			} else {
				p, firstRemote = roRecoveryNetwork(t, n, co, "dra", encrypted)
				roRecoveryConnect(t, n, co, "dra2", barrier, b)
				secondRemote = b
			}
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), combinedBody(t, "term"), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			first := routeRead(t, co, firstRemote)
			if oldWriter {
				<-barrier.entered
				n.mu.Lock()
				conn := p.conn
				n.mu.Unlock()
				n.failTransport(conn, io.EOF)
				info, e := n.RequestInfo(id)
				if e != nil || info.ReplayEligible {
					t.Fatal(info, e)
				}
				if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
					t.Fatal(e)
				}
				release.Do(func() { close(barrier.release) })
			} else {
				routeSuspect(t, n, p, firstRemote, co)
			}
			info := roAwait(t, n, id)
			if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
				t.Fatal(e)
			}
			second := routeRead(t, co, secondRemote)
			routeSameRequest(t, co, first, second)
			if oldWriter {
				roRecoveryReply(t, co, secondRemote, second, 2001)
			} else {
				<-barrier.entered
				roRecoveryReply(t, co, firstRemote, first, 2001)
				waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
				_ = n.Cancel(id)
				quiet, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
				if _, e = port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("writer still borrows Q", e)
				}
				cancel()
				release.Do(func() { close(barrier.release) })
			}
			got := routeCompletion(t, n, port, id)
			want := uint32(1)
			if oldWriter {
				want = 2
			}
			if got.SelectedAttempt.Number != want || len(got.Attempts) != 2 {
				t.Fatal(got)
			}
		}
	})
}
func TestRoRecoveryDeliveryEvidenceControlsT(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, state := range []struct {
			bytes     int
			uncertain bool
		}{{0, false}, {8, false}, {0, true}} {
			n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
			a, b := networkAdapters(t, encrypted)
			roRecoveryConnect(t, n, co, "dra", &roRecoveryFault{a, state.bytes, state.uncertain}, b)
			_, other := roRecoveryNetwork(t, n, co, "dra2", encrypted)
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			info := roAwait(t, n, id)
			if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
				t.Fatal(e)
			}
			second := routeRead(t, co, other)
			if (second.Header.Flags&pdu.RetransmittedFlag != 0) != (state.bytes > 0 || state.uncertain) {
				t.Fatal(state, second.Header)
			}
			roRecoveryReply(t, co, other, second, 2001)
			got := routeCompletion(t, n, port, id)
			if got.Attempts[0].Delivery.Bytes != state.bytes || got.Attempts[0].Delivery.Uncertain != state.uncertain {
				t.Fatal(got)
			}
		}
	})
}
func TestRoRecoveryBackpressureAndIdempotentAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := roRecoveryConfig(t, "combined")
		cfg.Peers[1].MaxPending = 1
		n, co := newInboundNode(t, cfg)
		_, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
		p2, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
		busyPort := portTest(t, n, 1)
		busy, e := p2.Submit(ctxTest(t), combinedBody(t, "balance"), RequestOptions{}, busyPort)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, b)
		r := roRecoveryRoute(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), combinedBody(t, "refund"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		_ = a.Close()
		info := roAwait(t, n, id)
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		after, _ := n.RequestInfo(id)
		if after.AttemptsUsed != 1 || !after.Deadline.Equal(info.Deadline) {
			t.Fatal(after)
		}
		_ = n.Cancel(busy)
		if _, e = busyPort.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e = n.Retransmit(canceled, id, info.CurrentAttempt); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		var wg sync.WaitGroup
		ids := make(chan AttemptID, 2)
		errs := make(chan error, 2)
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				next, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt)
				ids <- next
				errs <- e
			}()
		}
		wg.Wait()
		for range 2 {
			if e = <-errs; e != nil {
				t.Fatal(e)
			}
			if v := <-ids; v.Number != 2 {
				t.Fatal(v)
			}
		}
		if v, e := n.Retransmit(canceled, id, info.CurrentAttempt); e != nil || v.Number != 2 {
			t.Fatal("committed edge not idempotent", v, e)
		}
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		roRecoveryReply(t, co, b, second, 2001)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 {
			t.Fatal(got)
		}
	})
}
func TestRoRecoveryAwaitingTerminationAndPeerSingleAttempt(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, why := range []string{"cancel", "deadline", "close", "shutdown", "port-close", "peer-single"} {
			n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
			p, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
			_, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			var id RequestID
			var e error
			if why == "peer-single" {
				id, e = p.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
			} else {
				id, e = r.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
			}
			if e != nil {
				t.Fatal(e)
			}
			_ = routeRead(t, co, a)
			_ = a.Close()
			if why != "peer-single" {
				info := roAwait(t, n, id)
				switch why {
				case "cancel":
					_ = n.Cancel(id)
				case "deadline":
					n.tick(info.Deadline)
				case "close":
					_ = n.Close()
				case "shutdown":
					_ = n.Shutdown(ctxTest(t))
				case "port-close":
					if e = port.Close(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
				}
			}
			if why != "port-close" {
				got := routeCompletion(t, n, port, id)
				if got.Err == nil || got.Outcome != WrittenNoAnswer || len(got.Attempts) > 1 {
					t.Fatal(why, got)
				}
			}
			if why == "peer-single" || why == "cancel" || why == "deadline" {
				roQuiet(t, b)
			}
			if n.Resources().Outstanding != 0 {
				t.Fatal(n.Resources())
			}
		}
	})
}
func TestRoRecoveryRejectsUnselectedRequestsBeforeAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, family := range []string{"standard", "sms", "money", "combined"} {
			n, co := newInboundNode(t, roRecoveryConfig(t, family))
			_, remote := roRecoveryNetwork(t, n, co, "dra", encrypted)
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			for _, invalid := range []string{"action", "update"} {
				kind := "iec"
				if family == "money" {
					kind = "general"
				}
				q := roRecoveryBody(t, family, kind)
				switch q := q.(type) {
				case *ro.CCR:
					if invalid == "action" {
						q.RequestedAction = avp.Some(uint32(3))
					} else {
						q.CCRequestType = 2
					}
				case *smsfwro.CCR:
					if invalid == "action" {
						q.RequestedAction = avp.Some(uint32(3))
					} else {
						q.CCRequestType = 2
					}
				case *legacy420.CCR:
					if invalid == "action" {
						q.RequestedAction = 3
					} else {
						q.CCRequestType = 2
					}
				case *combined.CCR:
					if invalid == "action" {
						q.RequestedAction = avp.Some(uint32(3))
					} else {
						q.CCRequestType = 2
					}
				}
				n.mu.Lock()
				before := n.seq
				n.mu.Unlock()
				if _, e := r.Submit(ctxTest(t), q, RequestOptions{}, port); e == nil {
					t.Fatal("unselected request admitted", family, invalid)
				}
				n.mu.Lock()
				same := n.seq == before
				n.mu.Unlock()
				if !same || n.Resources().Outstanding != 0 {
					t.Fatal("rejected request allocated identity")
				}
			}
			roQuiet(t, remote)
		}
	})
}

func TestRoRecoverySinglePeerReconnect(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := roRecoveryConfig(t, "combined")
		cfg.Peers[0].Commands[0].Timeout = time.Minute
		cfg.Timeouts.Request = time.Minute
		n, co := newInboundNode(t, cfg)
		p, a := roRecoveryNetwork(t, n, co, "dra", tls)
		r, e := n.NewRoute(RouteConfig{Command: combinedBody(t, "initial").CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Minute, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Minute}})
		if e != nil {
			t.Fatal(e)
		}
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), combinedBody(t, "initial"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		_ = a.Close()
		waitFor(t, func() bool {
			i, _ := n.RequestInfo(id)
			return i.Phase == AwaitingRetransmit && p.Info().Candidates == 0
		})
		info, _ := n.RequestInfo(id)
		p, b := roRecoveryNetwork(t, n, co, "dra", tls)
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrNoEligibleAlternate) {
			t.Fatal(e)
		}
		for i := 0; i < 3; i++ {
			if i > 0 {
				n.mu.Lock()
				p.conn.tw = time.Now()
				n.mu.Unlock()
				n.tick(time.Now())
			}
			probe := routeRead(t, co, b)
			if probe.Header.CommandCode != 280 {
				t.Fatal(probe)
			}
			answer, e := co.CompleteAnswer(probe, &base.DWA{ResultCode: 2001}, draOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), b, co, answer); e != nil {
				t.Fatal(e)
			}
			want := i + 1
			waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn.numDWA == want })
		}
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		roRecoveryReply(t, co, b, second, 2001)
		got := routeCompletion(t, n, port, id)
		if got.Attempts[0].ConnectionID == got.Attempts[1].ConnectionID {
			t.Fatal(got)
		}
	})
}
func TestRoRecoveryAdmissionReservesAllAnswers(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
		_, a := roRecoveryNetwork(t, n, co, "dra", tls)
		r := roRecoveryRoute(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), combinedBody(t, "initial"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		n.mu.Lock()
		cap := n.requests[id].charge.Capacity
		bound := n.requests[id].answerBound
		n.mu.Unlock()
		_ = n.Cancel(id)
		_ = routeCompletion(t, n, port, id)
		small, e := n.NewCompletionPort(CompletionLimits{1, cap - bound})
		if e != nil {
			t.Fatal(e)
		}
		defer small.Close(context.Background())
		n.mu.Lock()
		seq := n.seq
		n.mu.Unlock()
		if _, e = r.Submit(ctxTest(t), combinedBody(t, "initial"), RequestOptions{}, small); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		n.mu.Lock()
		same := n.seq == seq
		n.mu.Unlock()
		if !same || n.Resources().Outstanding != 0 {
			t.Fatal("failed total reserve allocated a request", n.Resources())
		}
	})
}

func TestRoRecoveryInvalidAnswersAndAttemptIsolation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, why := range []string{"number", "session", "source", "oversize", "old-header-on-new-connection"} {
			n, co := newInboundNode(t, roRecoveryConfig(t, "combined"))
			p, a := roRecoveryNetwork(t, n, co, "dra", encrypted)
			_, b := roRecoveryNetwork(t, n, co, "dra2", encrypted)
			r := roRecoveryRoute(t, n, ExplicitRetransmit)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), combinedBody(t, "term"), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			first := routeRead(t, co, a)
			if why == "old-header-on-new-connection" {
				// Make the second connection's HBH allocation different before admission.
				n.mu.Lock()
				_, e = n.peers["dra2"].conn.nextHBH(false)
				n.mu.Unlock()
				if e != nil {
					t.Fatal(e)
				}
				routeSuspect(t, n, p, a, co)
				info := roAwait(t, n, id)
				if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
					t.Fatal(e)
				}
				second := routeRead(t, co, b)
				if first.Header.HopByHop == second.Header.HopByHop {
					t.Fatal("test needs distinct local ids")
				}
				roRecoveryReply(t, co, b, first, 2001)
				// A valid old header on the wrong physical connection must not consume A or completion.
				roRecoveryReply(t, co, b, second, 4012)
				got := routeCompletion(t, n, port, id)
				if got.SelectedAttempt.Number != 2 || got.Answer.Body.(*combined.CCA).ResultCode != 4012 {
					t.Fatal(got)
				}
				continue
			}
			ans, e := co.CompleteAnswer(first, combinedAnswer(first.Body.(*combined.CCR), 2001), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			switch why {
			case "number":
				// Inject below: the send validator rejects an invalid TERM number.
			case "session":
				ans.Body.(*combined.CCA).SessionID = "wrong"
			case "source":
				ans.Body.(*combined.CCA).OriginHost = "other-ocs.example"
			}
			raw, e := co.Append(nil, ans)
			if e != nil {
				t.Fatal(e)
			}
			if why == "number" {
				raw = replaceAVP32(t, raw, 415, 2)
			}
			if why == "oversize" {
				max := n.peers["dra"].policies[r.cfg.Command].MaxAnswerBytes
				raw = append(raw, make([]byte, max+4-len(raw))...)
				wire.PutHeader(raw, ans.Header, len(raw))
			}
			sendRejectRaw(t, a, raw)
			got := routeCompletion(t, n, port, id)
			if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil || len(got.Attempts) != 1 {
				t.Fatal(why, got)
			}
			roQuiet(t, b)
		}
	})
}
