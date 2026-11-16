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
	"io"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func routeConfig(t *testing.T) NodeConfig {
	c := twoDuplicatePeers(testConfig(t))
	c.Resources.NodeBytes = 256 << 20
	c.Resources.TransportBytes = 64 << 20
	for i := range c.Peers {
		c.Peers[i].Commands = append([]CommandPolicy(nil), c.Peers[i].Commands...)
		c.Peers[i].Commands[0].DestinationHost = serverOrigin.Host
	}
	return c
}
func newRouteTest(t *testing.T, n *Node, mode ExchangePolicy) *Route {
	t.Helper()
	cfg := RouteConfig{Command: requestBody().CommandKey(), Peers: []PeerID{"dra", "dra2"}, ExchangePolicy: mode, MaxAttempts: 2, MaxReplayAge: 2 * time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 5 * time.Second}}
	if mode == SingleAttempt {
		cfg.MaxAttempts = 1
		cfg.MaxReplayAge = 0
		cfg.ReplayDomain = ReplayDomain{}
	}
	r, e := n.NewRoute(cfg)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func routeReply(t *testing.T, co *codec.Codec, remote transport.Conn, req pdu.Message) {
	t.Helper()
	a, e := answerFor(co, req)
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), remote, co, a); e != nil {
		t.Fatal(e)
	}
}
func routeRead(t *testing.T, co *codec.Codec, remote transport.Conn) pdu.Message {
	t.Helper()
	m, e := readWire(ctxTest(t), remote, co)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func routeSuspect(t *testing.T, n *Node, p *Peer, remote transport.Conn, co *codec.Codec) {
	t.Helper()
	n.mu.Lock()
	p.conn.tw = time.Now()
	n.mu.Unlock()
	n.tick(time.Now())
	m := routeRead(t, co, remote)
	if m.Header.CommandCode != 280 {
		t.Fatal(m)
	}
	n.mu.Lock()
	p.conn.tw = time.Now()
	n.mu.Unlock()
	n.tick(time.Now())
	if p.Info().Watchdog != Suspect {
		t.Fatal(p.Info())
	}
}
func routeSameRequest(t *testing.T, co *codec.Codec, a, b pdu.Message) {
	t.Helper()
	x := reqBytes(t, co, a)
	y := reqBytes(t, co, b)
	if a.Header.EndToEnd != b.Header.EndToEnd || !bytes.Equal(x[20:], y[20:]) || a.Header.Flags&pdu.RetransmittedFlag != 0 || b.Header.Flags&pdu.RetransmittedFlag == 0 {
		t.Fatal("request identity or content changed", a.Header, b.Header)
	}
}
func routeCompletion(t *testing.T, n *Node, port *CompletionPort, id RequestID) Result {
	t.Helper()
	c, e := port.Receive(ctxTest(t))
	if e != nil || c.RequestID != id {
		t.Fatal(c, e)
	}
	if _, e = n.RequestInfo(id); !errors.Is(e, ErrRequestNotActive) {
		t.Fatal(e)
	}
	return c.Result
}
func TestRouteConfiguration(t *testing.T) {
	n, _ := newInboundNode(t, routeConfig(t))
	cfg := newRouteTest(t, n, TransportFailover).cfg
	for _, mutate := range []func(*RouteConfig){
		func(c *RouteConfig) { c.ExchangePolicy = 0 }, func(c *RouteConfig) { c.Peers = []PeerID{"dra"} }, func(c *RouteConfig) { c.Peers = []PeerID{"dra", "dra"} }, func(c *RouteConfig) { c.MaxAttempts = 17 }, func(c *RouteConfig) { c.MaxReplayAge = 10 * time.Second }, func(c *RouteConfig) { c.ReplayDomain.Target.Host = "different.example" }, func(c *RouteConfig) { c.Command = pdu.CommandKey{Code: 280, Request: true} },
	} {
		c := cfg
		mutate(&c)
		if _, e := n.NewRoute(c); e == nil {
			t.Fatal(c)
		}
	}
	for _, different := range []bool{false, true} {
		c := routeConfig(t)
		if different {
			c.Peers[1].Commands[0].DestinationHost = "other.example"
		} else {
			for i := range c.Peers {
				c.Peers[i].Commands[0].DestinationHost = ""
			}
		}
		other, _ := newInboundNode(t, c)
		if _, e := other.NewRoute(cfg); e == nil {
			t.Fatal("unbound target allowed")
		}
	}
}
func TestRouteTransportFailoverIdentityAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		_, a := duplicateConnect(t, n, co, tls, "dra")
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		body := requestBody()
		id, e := r.Submit(ctxTest(t), body, RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		body.MSISDN.Value[0] = 0x99
		before := n.Resources()
		n.mu.Lock()
		x := n.requests[id]
		if x.charge.Capacity < 2*x.answerBound {
			t.Error("missing second answer reserve")
		}
		n.mu.Unlock()
		_ = a.Close()
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		if got := n.Resources(); got.OutboundBytes != before.OutboundBytes || got.Outstanding != 1 {
			t.Fatal(before, got)
		}
		routeReply(t, co, b, second)
		result := routeCompletion(t, n, port, id)
		if result.Outcome != AnswerReceived || len(result.Attempts) != 2 || result.SelectedAttempt.Number != 2 || result.Attempts[0].StopReason != StopTransport {
			t.Fatal(result)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRouteOldLateAnswerWinsAndCompletesOnce(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		routeSuspect(t, n, p, a, co)
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		routeReply(t, co, a, first)
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 1 || len(got.Attempts) != 2 {
			t.Fatal(got)
		}
		routeReply(t, co, b, second)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
	})
}
func TestRouteExplicitRetransmitIdempotentAndNoAutomaticSend(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		_, a := duplicateConnect(t, n, co, tls, "dra")
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		if _, e := r.Request(ctxTest(t), requestBody(), RequestOptions{}); !errors.Is(e, ErrModeRequiresSubmit) {
			t.Fatal(e)
		}
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		info, e := n.RequestInfo(id)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
			t.Fatal(e)
		}
		_ = a.Close()
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
		current, _ := n.RequestInfo(id)
		if current.AttemptsUsed != 1 || !current.Deadline.Equal(info.Deadline) {
			t.Fatal(current)
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
			if next := <-ids; next.Number != 2 {
				t.Fatal(next)
			}
		}
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		routeReply(t, co, b, second)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 {
			t.Fatal(got)
		}
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrRequestNotActive) {
			t.Fatal(e)
		}
	})
}
func TestRouteValidErrorsNeverTriggerRecovery(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			n, co := newInboundNode(t, routeConfig(t))
			_, a := duplicateConnect(t, n, co, tls, "dra")
			_, _ = duplicateConnect(t, n, co, tls, "dra2")
			r := newRouteTest(t, n, mode)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			req := routeRead(t, co, a)
			ans, e := co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), a, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || len(got.Attempts) != 1 || got.Answer.Body.(*pdu.ErrorAnswer).ResultCode != 3004 {
				t.Fatal(got)
			}
		}
	})
}
func TestRouteLocalStopsNeverFailover(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, how := range []string{"cancel", "deadline", "peer-close", "node-close", "invalid-answer", "oversize-answer"} {
			t.Run(how, func(t *testing.T) {
				n, co := newInboundNode(t, routeConfig(t))
				p, a := duplicateConnect(t, n, co, tls, "dra")
				_, _ = duplicateConnect(t, n, co, tls, "dra2")
				r := newRouteTest(t, n, TransportFailover)
				port := portTest(t, n, 1)
				id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, a)
				switch how {
				case "cancel":
					_ = n.Cancel(id)
				case "deadline":
					n.mu.Lock()
					d := n.requests[id].deadline
					n.mu.Unlock()
					n.tick(d)
				case "peer-close":
					_ = p.Close()
				case "node-close":
					_ = n.Close()
				case "invalid-answer", "oversize-answer":
					ans, e := answerFor(co, req)
					if e != nil {
						t.Fatal(e)
					}
					raw := reqBytes(t, co, ans)
					if how == "invalid-answer" {
						raw[4] &^= pdu.ProxiableFlag
					} else {
						raw = append(raw, make([]byte, 1028-len(raw))...)
						wire.PutHeader(raw, ans.Header, len(raw))
					}
					sendRejectRaw(t, a, raw)
				}
				got := routeCompletion(t, n, port, id)
				if len(got.Attempts) != 1 || got.Answer.Body != nil || got.Err == nil {
					t.Fatal(got)
				}
				if (how == "invalid-answer" || how == "oversize-answer") && (!errors.Is(got.Err, validation.ErrInvalidAnswer) || got.StopReason != StopProtocol) {
					t.Fatal(got)
				}
			})
		}
	})
}
func TestRouteExplicitNoAlternateKeepsOriginalDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		_, a := duplicateConnect(t, n, co, tls, "dra")
		r := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		_ = a.Close()
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
		info, _ := n.RequestInfo(id)
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrNoEligibleAlternate) {
			t.Fatal(e)
		}
		next, _ := n.RequestInfo(id)
		if !next.Deadline.Equal(info.Deadline) || next.AttemptsUsed != 1 {
			t.Fatal(next)
		}
		n.tick(info.Deadline)
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, context.DeadlineExceeded) || got.Outcome != WrittenNoAnswer {
			t.Fatal(got)
		}
	})
}
func TestRouteAutomaticNoAlternateAndAttemptsExhausted(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, second := range []bool{false, true} {
			n, co := newInboundNode(t, routeConfig(t))
			_, a := duplicateConnect(t, n, co, tls, "dra")
			var b transport.Conn
			if second {
				_, b = duplicateConnect(t, n, co, tls, "dra2")
			}
			r := newRouteTest(t, n, TransportFailover)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			_ = routeRead(t, co, a)
			_ = a.Close()
			want := ErrNoEligibleAlternate
			if second {
				_ = routeRead(t, co, b)
				_ = b.Close()
				want = ErrAttemptsExhausted
			}
			got := routeCompletion(t, n, port, id)
			if !errors.Is(got.Err, want) || got.Outcome != WrittenNoAnswer || !got.Delivery.MayHaveReachedPeer {
				t.Fatal(got)
			}
		}
	})
}

// Completed frame read held at the adapter return boundary: each attempt keeps
// its own A and the final completion waits for both receiver borrowers to retire.
type routeReadBarrier struct {
	transport.Conn
	entered, release chan struct{}
	once             sync.Once
}

func (b *routeReadBarrier) ReadFrameInto(ctx context.Context, dst []byte) error {
	e := b.Conn.ReadFrameInto(ctx, dst)
	if e == nil && len(dst) >= 20 && wire.Uint24(dst[5:8]) == 8388647 && dst[4]&pdu.RequestFlag == 0 {
		b.once.Do(func() { close(b.entered); <-b.release })
	}
	return e
}
func connectRouteWrapped(t *testing.T, n *Node, co *codec.Codec, id PeerID, a, b transport.Conn) *Peer {
	t.Helper()
	ch := make(chan error, 1)
	go func() {
		req, e := readWire(ctxTest(t), b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "route-peer", AuthApplicationID: []uint32{appID}}, n.peers[id].origin)
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
func TestRouteConcurrentAnswersKeepSeparateBudgets(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &routeReadBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := connectRouteWrapped(t, n, co, "dra", barrier, b)
		_, other := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, b)
		initial := n.Resources().OutboundBytes
		routeSuspect(t, n, p, b, co)
		second := routeRead(t, co, other)
		routeReply(t, co, b, first)
		<-barrier.entered
		routeReply(t, co, other, second)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		if n.Resources().OutboundBytes != initial || port.count != 0 {
			t.Fatal("early release/publication", n.Resources())
		}
		release.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 2 || len(got.Attempts) != 2 {
			t.Fatal(got)
		}
	})
}

// A successful write may already be on the wire while its adapter still borrows
// the frame. A late old answer cannot free the shared Q or publish early.
func TestRouteLateAnswerWaitsForAlternateWriter(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, firstRemote := duplicateConnect(t, n, co, tls, "dra")
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, request: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		connectRouteWrapped(t, n, co, "dra2", barrier, b)
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, firstRemote)
		routeSuspect(t, n, p, firstRemote, co)
		_ = routeRead(t, co, b)
		<-barrier.entered
		routeReply(t, co, firstRemote, first)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		_ = n.Cancel(id)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		release.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 1 || got.Attempts[1].Delivery.Bytes == 0 || got.Err != nil {
			t.Fatal(got)
		}
	})
}

func TestRouteExplicitBackpressureAndHistoricalPeerClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := routeConfig(t)
		cfg.Peers[1].MaxPending = 1
		n, co := newInboundNode(t, cfg)
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		busyPort := portTest(t, n, 1)
		busy, e := p2.Submit(ctxTest(t), requestBody(), RequestOptions{}, busyPort)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, b)
		r := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		_ = a.Close()
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
		info, _ := n.RequestInfo(id)
		_ = p.Close()
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		after, _ := n.RequestInfo(id)
		if after.AttemptsUsed != 1 {
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
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, b)
		routeReply(t, co, b, second)
		_ = routeCompletion(t, n, port, id)
	})
}
func TestRouteAwaitingStopsWithoutConnections(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, how := range []string{"cancel", "node-close", "port-close", "shutdown"} {
			t.Run(how, func(t *testing.T) {
				n, co := newInboundNode(t, routeConfig(t))
				_, a := duplicateConnect(t, n, co, tls, "dra")
				r := newRouteTest(t, n, ExplicitRetransmit)
				port := portTest(t, n, 1)
				id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				_ = routeRead(t, co, a)
				_ = a.Close()
				waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == AwaitingRetransmit })
				switch how {
				case "cancel":
					_ = n.Cancel(id)
				case "node-close":
					_ = n.Close()
				case "port-close":
					if e = port.Close(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
				case "shutdown":
					if e = n.Shutdown(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
				}
				if how != "port-close" {
					got := routeCompletion(t, n, port, id)
					if got.Err == nil || got.Outcome != WrittenNoAnswer {
						t.Fatal(got)
					}
				}
				if n.Resources().Outstanding != 0 {
					t.Fatal(n.Resources())
				}
			})
		}
	})
}
func TestRoutePeerSubmissionRemainsSingle(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, _ := duplicateConnect(t, n, co, tls, "dra2")
		_ = newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		_ = a.Close()
		got, e := port.Receive(ctxTest(t))
		if e != nil || got.RequestID != id || len(got.Result.Attempts) != 0 || got.Result.Outcome != WrittenNoAnswer {
			t.Fatal(got, e)
		}
		n.mu.Lock()
		pending := len(p2.conn.pending)
		n.mu.Unlock()
		if pending != 0 {
			t.Fatal("Peer implicitly failed over")
		}
	})
}
func TestRouteOldAnswerWhileAwaitingRetransmit(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		_, _ = duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, a)
		routeSuspect(t, n, p, a, co)
		info, _ := n.RequestInfo(id)
		if info.Phase != AwaitingRetransmit {
			t.Fatal(info)
		}
		routeReply(t, co, a, req)
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 1 || len(got.Attempts) != 1 {
			t.Fatal(got)
		}
	})
}
func TestRoutePerAttemptConnectionAndErrorSources(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := routeConfig(t)
		cfg.Peers[1].Commands[0].Sources.ErrorAgents = []pdu.Origin{cfg.Peers[1].Origin}
		n, co := newInboundNode(t, cfg)
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, a)
		n.mu.Lock()
		p2.conn.hbhSeed = first.Header.HopByHop + 100
		n.mu.Unlock()
		routeSuspect(t, n, p, a, co)
		second := routeRead(t, co, b)
		routeReply(t, co, b, first) // Old HBH on another connection is unmatched, not a candidate.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		answer, e := co.CompleteAnswer(second, &pdu.ErrorAnswer{ResultCode: 3002}, cfg.Peers[1].Origin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), b, co, answer); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if got.Err != nil || got.SelectedAttempt.Number != 2 {
			t.Fatal(got)
		}
	})
}
func TestRouteWaitsForFailedWriterBeforeReusingFrame(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, request: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := connectRouteWrapped(t, n, co, "dra", barrier, b)
		p2, other := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, b)
		<-barrier.entered
		n.mu.Lock()
		old := p.conn
		n.mu.Unlock()
		n.failTransport(old, io.EOF)
		n.mu.Lock()
		attempts := len(n.requests[id].group.attempts)
		queued := len(p2.conn.queue)
		n.mu.Unlock()
		if attempts != 1 || queued != 0 {
			t.Fatal("new write overlapped borrowed old frame")
		}
		release.Do(func() { close(barrier.release) })
		second := routeRead(t, co, other)
		routeSameRequest(t, co, first, second)
		routeReply(t, co, other, second)
		_ = routeCompletion(t, n, port, id)
	})
}
func TestRouteQueuedAlternateCanceledKeepsPriorEvidence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		c, d := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: c, request: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p2 := connectRouteWrapped(t, n, co, "dra2", barrier, d)
		busyPort := portTest(t, n, 1)
		busy, e := p2.Submit(ctxTest(t), requestBody(), RequestOptions{}, busyPort)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, d)
		<-barrier.entered
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		routeSuspect(t, n, p, a, co)
		_ = n.Cancel(id)
		got := routeCompletion(t, n, port, id)
		if got.Outcome != WrittenNoAnswer || !got.Delivery.MayHaveReachedPeer || len(got.Attempts) != 2 || got.Attempts[1].Delivery.Bytes != 0 {
			t.Fatal(got)
		}
		release.Do(func() { close(barrier.release) })
		_ = n.Cancel(busy)
		if _, e = busyPort.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		n.mu.Lock()
		queued := len(p2.conn.queue)
		n.mu.Unlock()
		if queued != 0 {
			t.Fatal("canceled alternate remained queued")
		}
	})
}

// Fault injection records the adapter's exact/uncertain evidence without creating
// an unbounded partial-frame simulator. Base handshake traffic is unaffected.
type routeWriteFault struct {
	transport.Conn
	bytes     int
	uncertain bool
}

func (f *routeWriteFault) WriteFrame(ctx context.Context, b []byte) transport.WriteResult {
	if len(b) >= 20 && wire.Uint24(b[5:8]) == 8388647 && b[4]&pdu.RequestFlag != 0 {
		return transport.WriteResult{Bytes: f.bytes, Uncertain: f.uncertain, Err: io.ErrUnexpectedEOF}
	}
	return f.Conn.WriteFrame(ctx, b)
}
func TestRouteRetransmitTUsesActualDeliveryEvidence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, state := range []struct {
			bytes     int
			uncertain bool
		}{{0, false}, {8, false}, {0, true}} {
			n, co := newInboundNode(t, routeConfig(t))
			a, b := networkAdapters(t, tls)
			connectRouteWrapped(t, n, co, "dra", &routeWriteFault{a, state.bytes, state.uncertain}, b)
			_, other := duplicateConnect(t, n, co, tls, "dra2")
			r := newRouteTest(t, n, TransportFailover)
			port := portTest(t, n, 1)
			id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			second := routeRead(t, co, other)
			if (second.Header.Flags&pdu.RetransmittedFlag != 0) != (state.bytes > 0 || state.uncertain) {
				t.Fatal(second.Header, state)
			}
			_ = n.Cancel(id)
			got := routeCompletion(t, n, port, id)
			if got.Attempts[0].Delivery.Bytes != state.bytes || got.Attempts[0].Delivery.Uncertain != state.uncertain || got.Outcome != WrittenNoAnswer {
				t.Fatal(got)
			}
		}
	})
}
func TestRouteExplicitSinglePeerReconnect(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := routeConfig(t)
		cfg.Peers[0].Commands[0].Timeout = time.Minute
		cfg.Timeouts.Request = time.Minute
		n, co := newInboundNode(t, cfg)
		p, a := duplicateConnect(t, n, co, tls, "dra")
		r, e := n.NewRoute(RouteConfig{Command: requestBody().CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Minute, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Minute}})
		if e != nil {
			t.Fatal(e)
		}
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
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
		p, b := duplicateConnect(t, n, co, tls, "dra")
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
		routeReply(t, co, b, second)
		got := routeCompletion(t, n, port, id)
		if got.Attempts[0].ConnectionID == got.Attempts[1].ConnectionID {
			t.Fatal(got)
		}
	})
}
func TestRouteAdmissionReservesAllAnswerSlots(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		_, a := duplicateConnect(t, n, co, tls, "dra")
		r := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
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
		if _, e = r.Submit(ctxTest(t), requestBody(), RequestOptions{}, small); !errors.Is(e, ErrBackpressure) {
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

func TestRouteFailureClassification(t *testing.T) {
	for _, err := range []error{io.EOF, io.ErrUnexpectedEOF, io.ErrShortWrite, context.DeadlineExceeded} {
		if got := ioStopReason(err); got != StopTransport {
			t.Fatal(err, got)
		}
	}
	for _, err := range []error{ErrProtocol, wire.ErrFrame, wire.ErrLimit, transport.ErrFrameState, validation.ErrInvalidAnswer, errors.New("TLS protocol alert")} {
		if got := ioStopReason(err); got != StopProtocol {
			t.Fatal(err, got)
		}
	}
	for _, err := range []error{context.Canceled, ErrClosed, ErrRemoteDPR, ErrShutdown, ErrPeerClosed} {
		if got := ioStopReason(err); got != StopLocal {
			t.Fatal(err, got)
		}
	}
}
func TestRouteSingleRequestSelectsReadyPeer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := routeConfig(t)
		for i := range cfg.Peers {
			cfg.Peers[i].Commands[0].DestinationHost = ""
		}
		n, co := newInboundNode(t, cfg)
		_, b := duplicateConnect(t, n, co, tls, "dra2")
		r := newRouteTest(t, n, SingleAttempt)
		ch := make(chan Result, 1)
		go func() {
			result, e := r.Request(ctxTest(t), requestBody(), RequestOptions{})
			if e != nil {
				result.Err = e
			}
			ch <- result
		}()
		req := routeRead(t, co, b)
		routeReply(t, co, b, req)
		got := <-ch
		if got.Err != nil || got.PeerID != "dra2" || len(got.Attempts) != 1 || got.Outcome != AnswerReceived {
			t.Fatal(got)
		}
	})
}
func TestRouteForeignStaleAndModeChecks(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, routeConfig(t))
		_, a := duplicateConnect(t, n, co, tls, "dra")
		r := newRouteTest(t, n, ExplicitRetransmit)
		port := portTest(t, n, 1)
		id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		other := id
		other.Boot[0] ^= 1
		if _, e = n.RequestInfo(other); !errors.Is(e, ErrForeignRequest) {
			t.Fatal(e)
		}
		if _, e = n.Retransmit(ctxTest(t), other, AttemptID{other, 1}); !errors.Is(e, ErrForeignRequest) {
			t.Fatal(e)
		}
		for _, after := range []AttemptID{{id, 0}, {id, 2}, {other, 1}} {
			if _, e = n.Retransmit(ctxTest(t), id, after); !errors.Is(e, ErrStaleAttempt) {
				t.Fatal(after, e)
			}
		}
		_ = n.Cancel(id)
		_ = routeCompletion(t, n, port, id)
		r = newRouteTest(t, n, TransportFailover)
		id, e = r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		if _, e = n.Retransmit(ctxTest(t), id, AttemptID{id, 1}); !errors.Is(e, ErrUnsupported) {
			t.Fatal(e)
		}
		_ = n.Cancel(id)
		_ = routeCompletion(t, n, port, id)
	})
}
