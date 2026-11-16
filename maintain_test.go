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
	"math"
	"net"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/testkit"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	dtcp "github.com/skill215/diameterkit/transport/tcp"
	dtls "github.com/skill215/diameterkit/transport/tls"
)

type trackedDialConn struct {
	transport.Conn
	closed atomic.Bool
}

func (c *trackedDialConn) Close() error { c.closed.Store(true); return c.Conn.Close() }
func (c *trackedDialConn) Closed() bool { return c.closed.Load() }

type testDialer struct {
	bytes uint64
	call  func(context.Context) (transport.Conn, error)
}

func (d testDialer) Info() transport.Info                             { return transport.Info{BufferBytes: d.bytes} }
func (d testDialer) Dial(ctx context.Context) (transport.Conn, error) { return d.call(ctx) }
func maintainConfig(t *testing.T) NodeConfig {
	c := testConfig(t)
	c.Peers[0].Maintain = MaintainPolicy{InitialBackoff: 10 * time.Millisecond, MaxBackoff: 40 * time.Millisecond, StablePeriod: time.Second}
	return c
}

type maintainRun struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

func startMaintain(t *testing.T, n *Node, d transport.Dialer) *maintainRun {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	r := &maintainRun{cancel: cancel, done: make(chan struct{})}
	go func() { r.err = n.Maintain(ctx, "dra", d); close(r.done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.done:
		case <-time.After(2 * time.Second):
			t.Error("manager did not exit")
		}
	})
	return r
}
func (r *maintainRun) wait(t *testing.T) error {
	t.Helper()
	select {
	case <-r.done:
		return r.err
	case <-time.After(2 * time.Second):
		t.Fatal("manager still running")
		return nil
	}
}
func dialListener(t *testing.T, encrypted bool) (*net.TCPListener, transport.Dialer) {
	t.Helper()
	l, e := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = l.Close() })
	if encrypted {
		client, _ := tlsConfigs(t)
		d, e := dtls.NewDialer("tcp4", l.Addr().String(), dtls.Config{TLS: client, HandshakeTimeout: time.Second})
		if e != nil {
			t.Fatal(e)
		}
		return l, d
	}
	d, e := dtcp.NewDialer("tcp4", l.Addr().String())
	if e != nil {
		t.Fatal(e)
	}
	return l, d
}
func acceptDial(t *testing.T, l *net.TCPListener, encrypted bool) transport.Conn {
	t.Helper()
	_ = l.SetDeadline(time.Now().Add(2 * time.Second))
	raw, e := l.AcceptTCP()
	if e != nil {
		t.Fatal(e)
	}
	var c transport.Conn
	if encrypted {
		_, cfg := tlsConfigs(t)
		c, e = dtls.Server(ctxTest(t), raw, dtls.Config{TLS: cfg, HandshakeTimeout: time.Second})
	} else {
		c, e = dtcp.New(raw)
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}
func answerCER(t *testing.T, n *Node, co *codec.Codec, c transport.Conn) {
	t.Helper()
	req := routeRead(t, co, c)
	if req.Header.CommandCode != 257 {
		t.Fatal(req)
	}
	ans, e := co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "dial-test", AuthApplicationID: []uint32{appID}}, draOrigin)
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), c, co, ans); e != nil {
		t.Fatal(e)
	}
}
func finishReopen(t *testing.T, n *Node, p *Peer, co *codec.Codec, c transport.Conn) {
	t.Helper()
	waitFor(t, func() bool { return p.Info().Watchdog == Reopen })
	for i := 0; i < 3; i++ {
		if i > 0 {
			n.mu.Lock()
			p.conn.tw = time.Now()
			n.mu.Unlock()
			n.tick(time.Now())
		}
		req := routeRead(t, co, c)
		if req.Header.CommandCode != 280 {
			t.Fatal("business replay before Ready", req)
		}
		ans, e := co.CompleteAnswer(req, &base.DWA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), c, co, ans); e != nil {
			t.Fatal(e)
		}
		want := i + 1
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn != nil && p.conn.numDWA == want })
	}
	if !p.Info().Ready {
		t.Fatal(p.Info())
	}
}

func TestDialConfigurationAndAdmission(t *testing.T) {
	for _, bad := range []MaintainPolicy{{InitialBackoff: -1}, {InitialBackoff: 2, MaxBackoff: 1, StablePeriod: 1}, {InitialBackoff: 1, MaxBackoff: math.MaxInt64, Jitter: 1, StablePeriod: 1}, {InitialBackoff: 1, MaxBackoff: 1, StablePeriod: 0}} {
		cfg := maintainConfig(t)
		cfg.Peers[0].Maintain = bad
		if n, e := NewNode(cfg); e == nil {
			_ = n.Close()
			t.Fatal("bad policy accepted")
		}
	}
	n, _ := newTestNode(t)
	var calls atomic.Int32
	d := testDialer{call: func(context.Context) (transport.Conn, error) { calls.Add(1); return nil, io.EOF }}
	if e := n.Maintain(ctxTest(t), "dra", d); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
	if _, e := n.Dial(ctxTest(t), "missing", d); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
	d.bytes = n.cfg.Resources.AdapterBytesPerConnection + 1
	if _, e := n.Dial(ctxTest(t), "dra", d); !errors.Is(e, ErrBackpressure) {
		t.Fatal(e)
	}
	if calls.Load() != 0 {
		t.Fatal("dial happened before admission")
	}
	for _, address := range []string{"", ":3868", "localhost:0", "localhost:65536", "localhost:diameter"} {
		if _, e := dtcp.NewDialer("tcp", address); e == nil {
			t.Fatal(address)
		}
	}
	if _, e := dtcp.NewDialer("udp", "localhost:3868"); e == nil {
		t.Fatal("UDP accepted")
	}
	if _, e := dtls.NewDialer("tcp", "localhost:3868", dtls.Config{}); e == nil {
		t.Fatal("TLS config omitted")
	}
}
func TestDialOwnsErrorAndLateConnections(t *testing.T) {
	for _, kind := range []string{"error", "underdeclared", "cancel-late", "close-late", "shutdown-late"} {
		t.Run(kind, func(t *testing.T) {
			n, _ := newTestNode(t)
			p, _ := n.Peer("dra")
			raw, b := testkit.Pipe()
			a := &trackedDialConn{Conn: raw}
			defer b.Close()
			entered, release := make(chan struct{}), make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			d := testDialer{bytes: a.Info().BufferBytes, call: func(context.Context) (transport.Conn, error) {
				close(entered)
				<-release
				if kind == "error" {
					return a, io.EOF
				}
				return a, nil
			}}
			if kind == "underdeclared" {
				d.bytes = 0
			}
			out := make(chan error, 1)
			go func() { _, e := n.Dial(ctx, "dra", d); out <- e }()
			<-entered
			if p.Info().Protocol != WaitConnAck || p.Info().Candidates != 1 || p.Info().ConnectDeadline.IsZero() {
				t.Fatal(p.Info())
			}
			switch kind {
			case "cancel-late":
				cancel()
			case "close-late":
				_ = n.Close()
			case "shutdown-late":
				short, c := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer c()
				if e := n.Shutdown(short); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal(e)
				}
			}
			if kind == "close-late" || kind == "shutdown-late" {
				short, c := context.WithTimeout(context.Background(), 10*time.Millisecond)
				e := n.Wait(short)
				c()
				if !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("lost callback barrier", e)
				}
			}
			close(release)
			if e := <-out; e == nil {
				t.Fatal("late/error connection installed")
			}
			if !a.Closed() {
				t.Fatal("returned connection leaked")
			}
			if p.Info().Candidates != 0 || p.Info().Ready {
				t.Fatal(p.Info())
			}
		})
	}
}
func TestDialTotalDeadlineIncludesTransportAndCER(t *testing.T) {
	n, _ := newTestNode(t)
	raw, b := testkit.Pipe()
	a := &trackedDialConn{Conn: raw}
	defer b.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	var deadline time.Time
	d := testDialer{bytes: a.Info().BufferBytes, call: func(ctx context.Context) (transport.Conn, error) {
		deadline, _ = ctx.Deadline()
		<-time.After(30 * time.Millisecond)
		return a, nil
	}}
	start := time.Now()
	_, e := n.Dial(ctx, "dra", d)
	if !errors.Is(e, context.DeadlineExceeded) || time.Since(start) > 300*time.Millisecond {
		t.Fatal(e, time.Since(start))
	}
	want, _ := ctx.Deadline()
	if !deadline.Equal(want) {
		t.Fatal(deadline, want)
	}
	if !a.Closed() {
		t.Fatal("expired transport retained")
	}
}
func TestDialRealTCPAndTLS(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newTestNode(t)
		l, d := dialListener(t, tls)
		out := make(chan error, 1)
		go func() { _, e := n.Dial(ctxTest(t), "dra", d); out <- e }()
		remote := acceptDial(t, l, tls)
		answerCER(t, n, co, remote)
		if e := <-out; e != nil {
			t.Fatal(e)
		}
		p, _ := n.Peer("dra")
		if !p.Info().Ready {
			t.Fatal(p.Info())
		}
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		routeReply(t, co, remote, req)
		result, e := port.Receive(ctxTest(t))
		if e != nil || result.RequestID != id || result.Result.Outcome != AnswerReceived {
			t.Fatal(result, e)
		}
	})
}
func TestMaintainExclusiveOwnerAndCancellation(t *testing.T) {
	n, _ := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	entered := make(chan struct{})
	d := testDialer{call: func(ctx context.Context) (transport.Conn, error) { close(entered); <-ctx.Done(); return nil, ctx.Err() }}
	r := startMaintain(t, n, d)
	<-entered
	if e := n.Maintain(ctxTest(t), "dra", d); !errors.Is(e, ErrManaged) {
		t.Fatal(e)
	}
	if _, e := n.Dial(ctxTest(t), "dra", d); !errors.Is(e, ErrManaged) {
		t.Fatal(e)
	}
	raw, b := testkit.Pipe()
	a := &trackedDialConn{Conn: raw}
	defer b.Close()
	if _, e := n.Connect(ctxTest(t), "dra", a); !errors.Is(e, ErrManaged) || !a.Closed() {
		t.Fatal(e)
	}
	r.cancel()
	if e := r.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if e := p.Wait(ctxTest(t)); e != nil {
		t.Fatal(e)
	}
	if p.Info().Managed || p.Info().Lifecycle != Active || p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
	// Cancellation does not permanently close the management domain.
	r2 := startMaintain(t, n, testDialer{call: func(ctx context.Context) (transport.Conn, error) { <-ctx.Done(); return nil, ctx.Err() }})
	waitFor(t, func() bool { return p.Info().Managed })
	r2.cancel()
	_ = r2.wait(t)
}
func TestMaintainBackoffAndSuppression(t *testing.T) {
	cfg := maintainConfig(t)
	cfg.Peers[0].Maintain = MaintainPolicy{InitialBackoff: 20 * time.Millisecond, MaxBackoff: 40 * time.Millisecond, StablePeriod: time.Second}
	n, _ := newInboundNode(t, cfg)
	p, _ := n.Peer("dra")
	calls := make(chan time.Time, 8)
	var count atomic.Int32
	r := startMaintain(t, n, testDialer{call: func(context.Context) (transport.Conn, error) {
		calls <- time.Now()
		if count.Add(1) == 4 {
			return nil, ErrProtocol
		}
		return nil, syscall.ECONNREFUSED
	}})
	var times []time.Time
	for range 4 {
		select {
		case at := <-calls:
			times = append(times, at)
		case <-time.After(time.Second):
			t.Fatal("missing retry")
		}
	}
	for i, min := range []time.Duration{20 * time.Millisecond, 40 * time.Millisecond, 40 * time.Millisecond} {
		if times[i+1].Sub(times[i]) < min {
			t.Fatal(times)
		}
	}
	waitFor(t, func() bool { return p.Info().ReconnectSuppressed })
	if !p.Info().NextDial.IsZero() || p.Info().ReconnectFailures != 3 {
		t.Fatal(p.Info())
	}
	select {
	case <-calls:
		t.Fatal("suppressed manager redialed")
	case <-time.After(60 * time.Millisecond):
	}
	r.cancel()
	_ = r.wait(t)
}
func TestMaintainBackoffBoundsAndContinuousStability(t *testing.T) {
	n, _ := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	n.mu.Lock()
	m := &maintenance{policy: MaintainPolicy{InitialBackoff: time.Second, MaxBackoff: 4 * time.Second, Jitter: time.Second, StablePeriod: time.Second}, jitter: 1, ctx: context.Background()}
	for _, failures := range []uint32{0, 1, 2, 63, math.MaxUint32} {
		m.failures = failures
		before := time.Now()
		n.scheduleReconnectLocked(p, m, io.EOF, false)
		min := time.Second
		if failures == 1 {
			min = 2 * time.Second
		} else if failures >= 2 {
			min = 4 * time.Second
		}
		delay := m.next.Sub(before)
		if delay < min || delay > min+time.Second+time.Millisecond {
			t.Fatal(delay, failures)
		}
	}
	n.mu.Unlock()
}
func TestMaintainTransientReconnectNeverReplaysBusiness(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := maintainConfig(t)
		cfg.Peers[0].Commands[0].DestinationHost = serverOrigin.Host
		cfg.Timeouts.Request = time.Minute
		cfg.Peers[0].Commands[0].Timeout = time.Minute
		n, co := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		l, d := dialListener(t, tls)
		r := startMaintain(t, n, d)
		first := acceptDial(t, l, tls)
		answerCER(t, n, co, first)
		waitFor(t, func() bool { return p.Info().Ready })
		route, e := n.NewRoute(RouteConfig{Command: requestBody().CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Minute, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Minute}})
		if e != nil {
			t.Fatal(e)
		}
		port := portTest(t, n, 2)
		id, e := route.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		original := routeRead(t, co, first)
		old := p.Info().ConnectionID
		single, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, first)
		_ = first.Close()
		completed, e := port.Receive(ctxTest(t))
		if e != nil || completed.RequestID != single {
			t.Fatal(completed, e)
		}
		second := acceptDial(t, l, tls)
		answerCER(t, n, co, second)
		finishReopen(t, n, p, co, second)
		if old == p.Info().ConnectionID {
			t.Fatal("connection identity reused")
		}
		info, e := n.RequestInfo(id)
		if e != nil || info.Phase != AwaitingRetransmit || info.AttemptsUsed != 1 || p.Info().Pending != 0 {
			t.Fatal(info, e, p.Info())
		}
		// A same-connection DWR flushes the writer behind any illicit queued replay.
		n.mu.Lock()
		p.conn.tw = time.Now()
		n.mu.Unlock()
		n.tick(time.Now())
		probe := routeRead(t, co, second)
		if probe.Header.CommandCode != 280 {
			t.Fatal("automatic business replay", probe)
		}
		ans, e := co.CompleteAnswer(probe, &base.DWA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), second, co, ans); e != nil {
			t.Fatal(e)
		}
		if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		replay := routeRead(t, co, second)
		routeSameRequest(t, co, original, replay)
		routeReply(t, co, second, replay)
		if got := routeCompletion(t, n, port, id); got.Outcome != AnswerReceived || len(got.Attempts) != 2 {
			t.Fatal(got)
		}
		r.cancel()
		_ = r.wait(t)
	})
}
func TestMaintainCancelDoesNotTriggerRouteFailover(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := routeConfig(t)
		cfg.Peers[0].Maintain = maintainConfig(t).Peers[0].Maintain
		n, co := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		l, d := dialListener(t, tls)
		r := startMaintain(t, n, d)
		a := acceptDial(t, l, tls)
		answerCER(t, n, co, a)
		waitFor(t, func() bool { return p.Info().Ready })
		_, _ = duplicateConnect(t, n, co, tls, "dra2")
		route := newRouteTest(t, n, TransportFailover)
		port := portTest(t, n, 1)
		id, e := route.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, a)
		r.cancel()
		_ = r.wait(t)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 1 || got.StopReason != StopLocal || !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
	})
}
func TestMaintainCloseAndShutdownDuringDial(t *testing.T) {
	for _, how := range []string{"peer-close", "node-close", "peer-shutdown", "node-shutdown"} {
		t.Run(how, func(t *testing.T) {
			n, _ := newInboundNode(t, maintainConfig(t))
			p, _ := n.Peer("dra")
			entered := make(chan struct{})
			r := startMaintain(t, n, testDialer{call: func(ctx context.Context) (transport.Conn, error) { close(entered); <-ctx.Done(); return nil, ctx.Err() }})
			<-entered
			switch how {
			case "peer-close":
				_ = p.Close()
			case "node-close":
				_ = n.Close()
			case "peer-shutdown":
				if e := p.Shutdown(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
			case "node-shutdown":
				if e := n.Shutdown(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
			}
			if e := r.wait(t); e == nil {
				t.Fatal("manager did not report closure")
			}
			if e := p.Wait(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
			if p.Info().Managed || p.Info().Candidates != 0 || p.Info().Lifecycle != Closed {
				t.Fatal(p.Info())
			}
			if _, e := n.Dial(ctxTest(t), "dra", testDialer{call: func(context.Context) (transport.Conn, error) { t.Error("closed peer dialed"); return nil, io.EOF }}); e == nil {
				t.Fatal("closed peer revived")
			}
		})
	}
}
func TestMaintainLateCallbackKeepsOwnershipAndWait(t *testing.T) {
	n, _ := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	raw, b := testkit.Pipe()
	a := &trackedDialConn{Conn: raw}
	defer b.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	r := startMaintain(t, n, testDialer{call: func(context.Context) (transport.Conn, error) { close(entered); <-release; return a, nil }})
	<-entered
	r.cancel()
	waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.manager.stopping })
	if !p.Info().Managed || p.Info().Candidates != 1 {
		t.Fatal(p.Info())
	}
	short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if e := p.Wait(short); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	if e := n.Maintain(ctxTest(t), "dra", testDialer{}); !errors.Is(e, ErrManaged) {
		t.Fatal(e)
	}
	close(release)
	if e := r.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if !a.Closed() || p.Info().Managed || p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestMaintainDPRReconnectPolicy(t *testing.T) {
	for _, cause := range []uint32{0, 1, 2} {
		t.Run(string(rune('0'+cause)), func(t *testing.T) {
			cfg := maintainConfig(t)
			cfg.Timeouts.ClosingGrace = 10 * time.Millisecond
			n, co := newInboundNode(t, cfg)
			p, _ := n.Peer("dra")
			l, d := dialListener(t, false)
			r := startMaintain(t, n, d)
			remote := acceptDial(t, l, false)
			answerCER(t, n, co, remote)
			waitFor(t, func() bool { return p.Info().Ready })
			req := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 282, HopByHop: 21, EndToEnd: 22}, Body: &base.DPR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, DisconnectCause: cause}}
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			if ans := routeRead(t, co, remote); ans.Header.CommandCode != 282 {
				t.Fatal(ans)
			}
			_ = remote.Close()
			if cause == 0 {
				next := acceptDial(t, l, false)
				answerCER(t, n, co, next)
				waitFor(t, func() bool { return p.Info().Watchdog == Reopen })
			} else {
				waitFor(t, func() bool { return p.Info().ReconnectSuppressed && p.Info().Candidates == 0 })
				if !p.Info().NextDial.IsZero() {
					t.Fatal(p.Info())
				}
			}
			r.cancel()
			_ = r.wait(t)
		})
	}
}

func TestDialElectionDuringTransportEstablishment(t *testing.T) {
	for _, localWins := range []bool{false, true} {
		t.Run(map[bool]string{false: "active-fails", true: "passive-wins"}[localWins], func(t *testing.T) {
			cfg := maintainConfig(t)
			if localWins {
				cfg.Origin.Host = "zulu.example"
			}
			n, co := newInboundNode(t, cfg)
			p, _ := n.Peer("dra")
			entered, release := make(chan struct{}), make(chan struct{})
			raw, outRemote := testkit.Pipe()
			out := &trackedDialConn{Conn: raw}
			defer outRemote.Close()
			r := startMaintain(t, n, testDialer{bytes: 4096, call: func(ctx context.Context) (transport.Conn, error) {
				close(entered)
				<-release
				if localWins {
					return out, nil
				}
				return out, io.EOF
			}})
			<-entered
			in, remote := testkit.Pipe()
			defer remote.Close()
			accepted := make(chan error, 1)
			go func() { _, e := n.Accept(ctxTest(t), "dra", in); accepted <- e }()
			if e := writeMessage(ctxTest(t), remote, co, peerCER(n)); e != nil {
				t.Fatal(e)
			}
			if !localWins {
				waitFor(t, func() bool { return p.Info().Protocol == WaitConnAckElect })
				a, b := testkit.Pipe()
				defer b.Close()
				if _, e := n.Accept(ctxTest(t), "dra", a); e == nil {
					t.Fatal("third candidate accepted")
				}
				close(release)
			}
			answer := routeRead(t, co, remote)
			if answer.Header.CommandCode != 257 {
				t.Fatal(answer)
			}
			if e := <-accepted; e != nil {
				t.Fatal(e)
			}
			if localWins {
				close(release)
			}
			waitFor(t, func() bool { return p.Info().Candidates == 1 })
			if !p.Info().Ready || p.Info().Protocol != ROpen || !out.Closed() {
				t.Fatal(p.Info())
			}
			r.cancel()
			_ = r.wait(t)
		})
	}
}
func TestDialPassiveCandidateCannotExtendDeadline(t *testing.T) {
	cfg := testConfig(t)
	cfg.Timeouts.Handshake = 60 * time.Millisecond
	n, co := newInboundNode(t, cfg)
	p, _ := n.Peer("dra")
	entered := make(chan struct{})
	active := make(chan error, 1)
	go func() {
		_, e := n.Dial(ctxTest(t), "dra", testDialer{call: func(ctx context.Context) (transport.Conn, error) { close(entered); <-ctx.Done(); return nil, ctx.Err() }})
		active <- e
	}()
	<-entered
	original := p.Info().ConnectDeadline
	a, b := testkit.Pipe()
	defer b.Close()
	passive := make(chan error, 1)
	go func() { _, e := n.Accept(ctxTest(t), "dra", a); passive <- e }()
	if e := writeMessage(ctxTest(t), b, co, peerCER(n)); e != nil {
		t.Fatal(e)
	}
	n.mu.Lock()
	for _, c := range p.candidates {
		if c != nil && !c.handshakeDeadline.Equal(original) {
			t.Error("candidate renewed total deadline")
		}
	}
	n.mu.Unlock()
	if e := <-active; !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	if e := <-passive; e == nil {
		t.Fatal("election outlived deadline")
	}
	if p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}
func TestMaintainCancellationDuringCER(t *testing.T) {
	n, co := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	a, b := testkit.Pipe()
	defer b.Close()
	r := startMaintain(t, n, testDialer{bytes: 4096, call: func(context.Context) (transport.Conn, error) { return a, nil }})
	req := routeRead(t, co, b)
	if req.Header.CommandCode != 257 {
		t.Fatal(req)
	}
	r.cancel()
	if e := r.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if p.Info().Managed || p.Info().Candidates != 0 || p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestMaintainShutdownKeepsDrainOwnership(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, maintainConfig(t))
		p, _ := n.Peer("dra")
		l, d := dialListener(t, tls)
		r := startMaintain(t, n, d)
		remote := acceptDial(t, l, tls)
		answerCER(t, n, co, remote)
		waitFor(t, func() bool { return p.Info().Ready })
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		done := make(chan error, 1)
		go func() { done <- n.Shutdown(ctxTest(t)) }()
		waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
		// Cancellation of the former owner must not defeat already committed drain.
		r.cancel()
		_ = r.wait(t)
		routeReply(t, co, remote, req)
		got, e := port.Receive(ctxTest(t))
		if e != nil || got.RequestID != id || got.Result.Outcome != AnswerReceived {
			t.Fatal(got, e)
		}
		dpr := routeRead(t, co, remote)
		if dpr.Header.CommandCode != 282 {
			t.Fatal(dpr)
		}
		dpa, e := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, dpa); e != nil {
			t.Fatal(e)
		}
		if e = <-done; e != nil {
			t.Fatal(e)
		}
		if p.Info().Managed || p.Info().Lifecycle != Closed {
			t.Fatal(p.Info())
		}
	})
}
func TestMaintainSuppressionAllowsPassiveAndRestart(t *testing.T) {
	n, co := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	var calls atomic.Int32
	d := testDialer{call: func(context.Context) (transport.Conn, error) { calls.Add(1); return nil, ErrProtocol }}
	r := startMaintain(t, n, d)
	waitFor(t, func() bool { return p.Info().ReconnectSuppressed })
	a, b := testkit.Pipe()
	defer b.Close()
	accepted := make(chan error, 1)
	go func() { _, e := n.Accept(ctxTest(t), "dra", a); accepted <- e }()
	if e := writeMessage(ctxTest(t), b, co, peerCER(n)); e != nil {
		t.Fatal(e)
	}
	_ = routeRead(t, co, b)
	if e := <-accepted; e != nil {
		t.Fatal(e)
	}
	if !p.Info().Ready || !p.Info().ReconnectSuppressed {
		t.Fatal(p.Info())
	}
	_ = b.Close()
	waitFor(t, func() bool { return p.Info().Candidates == 0 })
	if calls.Load() != 1 || !p.Info().ReconnectSuppressed {
		t.Fatal(calls.Load(), p.Info())
	}
	r.cancel()
	_ = r.wait(t)
	r2 := startMaintain(t, n, d)
	waitFor(t, func() bool { return calls.Load() == 2 && p.Info().ReconnectSuppressed })
	r2.cancel()
	_ = r2.wait(t)
}
func TestMaintainResetsOnlyAfterContinuousOkay(t *testing.T) {
	cfg := maintainConfig(t)
	cfg.Peers[0].Maintain.StablePeriod = 40 * time.Millisecond
	n, co := newInboundNode(t, cfg)
	p, _ := n.Peer("dra")
	l, real := dialListener(t, false)
	var calls atomic.Int32
	d := testDialer{bytes: 4096, call: func(ctx context.Context) (transport.Conn, error) {
		if calls.Add(1) == 1 {
			return nil, io.EOF
		}
		return real.Dial(ctx)
	}}
	r := startMaintain(t, n, d)
	remote := acceptDial(t, l, false)
	answerCER(t, n, co, remote)
	waitFor(t, func() bool { return p.Info().Ready })
	if p.Info().ReconnectFailures != 1 {
		t.Fatal(p.Info())
	}
	// Force the actual watchdog transition before allowing a stable interval.
	n.mu.Lock()
	p.conn.tw = time.Now()
	n.mu.Unlock()
	n.tick(time.Now())
	probe := routeRead(t, co, remote)
	n.mu.Lock()
	p.conn.tw = time.Now()
	n.mu.Unlock()
	n.tick(time.Now())
	if p.Info().Watchdog != Suspect {
		t.Fatal(p.Info())
	}
	time.Sleep(50 * time.Millisecond)
	if p.Info().ReconnectFailures != 1 {
		t.Fatal("Suspect counted as continuous Okay", p.Info())
	}
	answer, e := co.CompleteAnswer(probe, &base.DWA{ResultCode: 2001}, draOrigin)
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), remote, co, answer); e != nil {
		t.Fatal(e)
	}
	waitFor(t, func() bool { return p.Info().Ready })
	if p.Info().ReconnectFailures != 1 {
		t.Fatal("reset before stable period", p.Info())
	}
	waitFor(t, func() bool { return p.Info().ReconnectFailures == 0 })
	_ = remote.Close()
	waitFor(t, func() bool { return p.Info().ReconnectFailures == 1 })
	r.cancel()
	_ = r.wait(t)
}
func TestMaintainTLSVerificationFailureSuppresses(t *testing.T) {
	n, _ := newInboundNode(t, maintainConfig(t))
	p, _ := n.Peer("dra")
	l, _ := dialListener(t, true)
	client, server := tlsConfigs(t)
	client.ServerName = "wrong.invalid"
	d, e := dtls.NewDialer("tcp4", l.Addr().String(), dtls.Config{TLS: client, HandshakeTimeout: time.Second})
	if e != nil {
		t.Fatal(e)
	}
	client.ServerName = "localhost" // constructor must freeze the top-level config
	r := startMaintain(t, n, d)
	_ = l.SetDeadline(time.Now().Add(time.Second))
	raw, e := l.AcceptTCP()
	if e != nil {
		t.Fatal(e)
	}
	_, e = dtls.Server(ctxTest(t), raw, dtls.Config{TLS: server, HandshakeTimeout: time.Second})
	if e == nil {
		t.Fatal("unverified TLS completed")
	}
	waitFor(t, func() bool { return p.Info().ReconnectSuppressed })
	if p.Info().ReconnectCause == nil || p.Info().Candidates != 0 || p.Info().Ready || p.Info().ReconnectFailures != 0 {
		t.Fatal(p.Info())
	}
	r.cancel()
	_ = r.wait(t)
}
func TestDialTLSHandshakeCancellation(t *testing.T) {
	n, _ := newTestNode(t)
	l, d := dialListener(t, true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := n.Dial(ctx, "dra", d); done <- e }()
	_ = l.SetDeadline(time.Now().Add(time.Second))
	raw, e := l.AcceptTCP()
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	cancel()
	if e = <-done; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	p, _ := n.Peer("dra")
	if p.Info().Candidates != 0 || p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestMaintainCancellationDuringBackoff(t *testing.T) {
	cfg := maintainConfig(t)
	cfg.Peers[0].Maintain.InitialBackoff = time.Hour
	cfg.Peers[0].Maintain.MaxBackoff = time.Hour
	n, _ := newInboundNode(t, cfg)
	p, _ := n.Peer("dra")
	var calls atomic.Int32
	r := startMaintain(t, n, testDialer{call: func(context.Context) (transport.Conn, error) { calls.Add(1); return nil, syscall.ECONNREFUSED }})
	waitFor(t, func() bool { return !p.Info().NextDial.IsZero() })
	r.cancel()
	_ = r.wait(t)
	if calls.Load() != 1 || p.Info().Managed {
		t.Fatal(calls.Load(), p.Info())
	}
}
func TestMaintainFailureClassification(t *testing.T) {
	for _, e := range []error{syscall.ECONNREFUSED, syscall.ECONNRESET, io.EOF, context.DeadlineExceeded, ErrWatchdog} {
		if !reconnectError(e) {
			t.Fatal(e)
		}
	}
	for _, e := range []error{ErrProtocol, ErrConfig, ErrIdentifiers, ErrBackpressure, context.Canceled, ErrPeerClosed, ErrRemoteDPR, ErrShutdown, errors.New("TLS alert")} {
		if reconnectError(e) {
			t.Fatal(e)
		}
	}
}

func TestMaintainDialerErrorsCannotSpin(t *testing.T) {
	for _, cause := range []error{ErrNotReady, ErrElectionLost, ErrBackpressure} {
		t.Run(cause.Error(), func(t *testing.T) {
			n, _ := newInboundNode(t, maintainConfig(t))
			p, _ := n.Peer("dra")
			var calls atomic.Int32
			r := startMaintain(t, n, testDialer{call: func(context.Context) (transport.Conn, error) { calls.Add(1); return nil, cause }})
			waitFor(t, func() bool { return p.Info().ReconnectSuppressed })
			if calls.Load() != 1 {
				t.Fatal("dialer error bypassed backoff/suppression", calls.Load())
			}
			r.cancel()
			_ = r.wait(t)
		})
	}
}
func TestDialFixedReserveAndNoOutboundCharge(t *testing.T) {
	cfg := maintainConfig(t)
	n, _ := newInboundNode(t, cfg)
	rp := cfg.Resources
	baseReserve := 2*(n.controlBound+rp.AdapterBytesPerConnection) + uint64(cfg.Peers[0].MaxPending)*1024 + uint64(rp.MaxPorts)*uint64(rp.MaxPortOutstanding)*512 + uint64(rp.MaxPorts)*512
	cfg.Resources.TransportBytes = baseReserve + schedulerReserve(cfg.Resources, len(cfg.Peers)) + 4095
	if bad, e := NewNode(cfg); e == nil {
		_ = bad.Close()
		t.Fatal("missing manager reserve accepted")
	}
	cfg.Resources.TransportBytes = baseReserve + schedulerReserve(cfg.Resources, len(cfg.Peers)) + 4096
	exact, _ := newInboundNode(t, cfg)
	before := exact.Resources()
	if _, e := exact.Dial(ctxTest(t), "dra", testDialer{call: func(context.Context) (transport.Conn, error) { return nil, io.EOF }}); !errors.Is(e, io.EOF) {
		t.Fatal(e)
	}
	after := exact.Resources()
	if before != after || after.OutboundBytes != 0 {
		t.Fatal(before, after)
	}
}
