//go:build linux && !386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
)

func managedConfigs(t *testing.T) (diameter.NodeConfig, diameter.NodeConfig) {
	t.Helper()
	a, b := nodeConfig(t), nodeConfig(t)
	a.Peers[0].Origin = serverOrigin
	b.Origin = serverOrigin
	b.Peers[0].Origin = a.Origin
	b.Peers[0].Commands = nil
	b.Resources.NodeBytes = 256 << 20
	b.Resources.InboundClasses = []diameter.ClassConfig{{ID: "sms", Bytes: 64 << 20, Slots: 4}}
	b.Peers[0].MaxInbound = 4
	b.Peers[0].MaxInboundBytes = 64 << 20
	b.Peers[0].InboundCommands = []diameter.InboundPolicy{{Command: pdu.CommandKey{ApplicationID: appID, Code: 8388647, Request: true}, Class: "sms", Sources: []pdu.Origin{a.Origin}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, ExecutionTimeout: 2 * time.Second, NoCache: true}}
	r, e := DefaultSetupConfig().Requirements()
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []*diameter.NodeConfig{&a, &b} {
		c.Resources.AdapterBytesPerConnection = r.ConnectionBufferBytes
		c.Resources.ListenerBytesPerPeer = r.BufferBytes
		c.Timeouts.Watchdog = 6 * time.Second
		c.Peers[0].Maintain = diameter.MaintainPolicy{InitialBackoff: 20 * time.Millisecond, MaxBackoff: 80 * time.Millisecond, StablePeriod: time.Minute}
	}
	return a, b
}

type managedRun struct {
	cancel context.CancelFunc
	done   chan error
}

func startManaged(t *testing.T, f func(context.Context) error) *managedRun {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	r := &managedRun{cancel, make(chan error, 1)}
	go func() { r.done <- f(ctx) }()
	t.Cleanup(cancel)
	return r
}
func (r *managedRun) wait(t *testing.T) error {
	t.Helper()
	select {
	case e := <-r.done:
		return e
	case <-time.After(4 * time.Second):
		t.Fatal("managed operation did not return")
		return nil
	}
}
func managedRequest(t *testing.T, p, q *diameter.Peer) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		in, e := q.Receive(testContext(t))
		if e == nil {
			_, e = in.Respond(testContext(t), success())
		}
		done <- e
	}()
	r, e := p.Request(testContext(t), requestBody(), diameter.RequestOptions{})
	if e != nil || r.Outcome != diameter.AnswerReceived {
		t.Fatal(r, e)
	}
	if e = <-done; e != nil {
		t.Fatal(e)
	}
}
func TestManagedSCTPDialServeS6cAndHandoff(t *testing.T) {
	a, b := managedConfigs(t)
	n, m := ownNode(t, a), ownNode(t, b)
	l := listenTest(t)
	q, _ := m.Peer("dra")
	run := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", l) })
	waitFor(t, func() bool { return q.Info().AcceptWaiting })
	if q.Info().Candidates != 1 {
		t.Fatal(q.Info())
	}
	p, e := n.Dial(testContext(t), "dra", dialerTest(t, l))
	if e != nil {
		t.Fatal(e)
	}
	waitFor(t, func() bool { return q.Info().Ready })
	managedRequest(t, p, q)
	run.cancel()
	if e = run.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if !q.Info().Ready || q.Info().Serving {
		t.Fatal(q.Info())
	}
	managedRequest(t, p, q)
	if e = n.Shutdown(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if e = n.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	m.Close()
	if e = m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if p.Info().LastDPAResult != 2001 || p.Info().Candidates != 0 || q.Info().Candidates != 0 || n.Resources().Outstanding != 0 || m.Resources().InboundOutstanding != 0 || n.Resources().OutboundBytes != 0 || m.Resources().InboundBytes != 0 {
		t.Fatal(p.Info(), q.Info(), n.Resources(), m.Resources())
	}
}

type countedDialer struct {
	transport.Dialer
	calls atomic.Int32
	mu    sync.Mutex
	last  transport.Conn
}

func (d *countedDialer) Dial(ctx context.Context) (transport.Conn, error) {
	d.calls.Add(1)
	c, e := d.Dialer.Dial(ctx)
	if e == nil {
		d.mu.Lock()
		d.last = c
		d.mu.Unlock()
	}
	return c, e
}
func (d *countedDialer) faultClose() {
	d.mu.Lock()
	c := d.last
	d.mu.Unlock()
	if c != nil {
		c.Close()
	}
}
func TestManagedSCTPAdmissionAndIdleStops(t *testing.T) {
	for _, action := range []string{"cancel", "node-close", "peer-close", "shutdown"} {
		t.Run(action, func(t *testing.T) {
			_, b := managedConfigs(t)
			m := ownNode(t, b)
			q, _ := m.Peer("dra")
			l := listenTest(t)
			r := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", l) })
			waitFor(t, func() bool { return q.Info().AcceptWaiting })
			switch action {
			case "cancel":
				r.cancel()
			case "node-close":
				m.Close()
			case "peer-close":
				q.Close()
			case "shutdown":
				if e := m.Shutdown(testContext(t)); e != nil {
					t.Fatal(e)
				}
			}
			if e := r.wait(t); e == nil {
				t.Fatal("missing stop cause")
			}
			if e := q.Wait(testContext(t)); e != nil {
				t.Fatal(e)
			}
			if q.Info().Candidates != 0 || q.Info().Serving || l.accepting.Load() {
				t.Fatal(q.Info())
			}
		})
	}
	t.Run("dial-budget", func(t *testing.T) {
		a, _ := managedConfigs(t)
		a.Resources.AdapterBytesPerConnection--
		n := ownNode(t, a)
		d := &countedDialer{Dialer: dialerTest(t, listenTest(t))}
		if _, e := n.Dial(testContext(t), "dra", d); !errors.Is(e, diameter.ErrBackpressure) || d.calls.Load() != 0 {
			t.Fatal(e, d.calls.Load())
		}
		if e := n.Maintain(testContext(t), "dra", d); !errors.Is(e, diameter.ErrBackpressure) || d.calls.Load() != 0 {
			t.Fatal(e, d.calls.Load())
		}
	})
	for _, candidate := range []bool{false, true} {
		name := "listener-budget"
		if candidate {
			name = "passive-candidate-budget"
		}
		t.Run(name, func(t *testing.T) {
			_, b := managedConfigs(t)
			if candidate {
				b.Resources.AdapterBytesPerConnection--
			} else {
				b.Resources.ListenerBytesPerPeer--
			}
			m := ownNode(t, b)
			l := listenTest(t)
			if e := m.Serve(testContext(t), "dra", l); !errors.Is(e, diameter.ErrBackpressure) {
				t.Fatal(e)
			}
			if _, e := l.Accept(testContext(t)); !errors.Is(e, net.ErrClosed) {
				t.Fatal(e)
			}
			q, _ := m.Peer("dra")
			if q.Info().Candidates != 0 || q.Info().Serving {
				t.Fatal(q.Info())
			}
		})
	}
}
func TestManagedSCTPFailureBackoffAndCancel(t *testing.T) {
	a, _ := managedConfigs(t)
	n := ownNode(t, a)
	l := listenTest(t)
	d := &countedDialer{Dialer: dialerTest(t, l)}
	l.Close()
	p, _ := n.Peer("dra")
	r := startManaged(t, func(ctx context.Context) error { return n.Maintain(ctx, "dra", d) })
	waitFor(t, func() bool { return p.Info().ReconnectFailures >= 2 })
	if d.calls.Load() < 2 || p.Info().Ready || p.Info().NextDial.IsZero() {
		t.Fatal(p.Info(), d.calls.Load())
	}
	r.cancel()
	if e := r.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if p.Info().Candidates != 0 || p.Info().Managed {
		t.Fatal(p.Info())
	}
}
func TestManagedSCTPReconnectDoesNotReplay(t *testing.T) {
	a, b := managedConfigs(t)
	n, m := ownNode(t, a), ownNode(t, b)
	l := listenTest(t)
	d := &countedDialer{Dialer: dialerTest(t, l)}
	s := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", l) })
	r := startManaged(t, func(ctx context.Context) error { return n.Maintain(ctx, "dra", d) })
	p, _ := n.Peer("dra")
	q, _ := m.Peer("dra")
	waitFor(t, func() bool { return p.Info().Ready && q.Info().Ready })
	first := p.Info().ConnectionID
	port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 64 << 20})
	if e != nil {
		t.Fatal(e)
	}
	defer port.Close(testContext(t))
	id, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := q.Receive(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	d.faultClose() // Test fault injection only; production callers relinquish socket ownership.
	completed, e := port.Receive(testContext(t))
	if e != nil || completed.RequestID != id || completed.Result.Outcome != diameter.WrittenNoAnswer {
		t.Fatal(completed, e)
	}
	waitFor(t, func() bool { return p.Info().ConnectionID != first && p.Info().Watchdog == diameter.Reopen })
	if _, e = in.Respond(testContext(t), success()); e == nil {
		t.Fatal("old generation accepted response")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	for !p.Info().Ready || !q.Info().Ready {
		select {
		case <-ctx.Done():
			t.Fatal("REOPEN did not complete", p.Info(), q.Info())
		case <-time.After(10 * time.Millisecond):
		}
	}
	noReplay, c := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if _, e = q.Receive(noReplay); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("automatic replay", e)
	}
	c()
	noDuplicate, c := context.WithTimeout(context.Background(), 10*time.Millisecond)
	if _, e = port.Receive(noDuplicate); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("duplicate completion", e)
	}
	c()
	managedRequest(t, p, q)
	r.cancel()
	if e = r.wait(t); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	s.cancel()
	s.wait(t)
	m.Close()
	if e = m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if p.Info().Candidates != 0 || p.Info().Managed || n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 || m.Resources().InboundOutstanding != 0 || m.Resources().InboundBytes != 0 {
		t.Fatal(p.Info(), n.Resources(), m.Resources())
	}
}

type gatedDialer struct {
	transport.Dialer
	arrived chan struct{}
	release <-chan struct{}
}

func (d gatedDialer) Dial(ctx context.Context) (transport.Conn, error) {
	c, e := d.Dialer.Dial(ctx)
	if e != nil {
		return nil, e
	}
	close(d.arrived)
	select {
	case <-d.release:
		return c, nil
	case <-ctx.Done():
		c.Close()
		return nil, ctx.Err()
	}
}
func TestManagedSCTPCrossedElectionAndShutdown(t *testing.T) {
	a, b := managedConfigs(t)
	a.Timeouts.Handshake = 2 * time.Second
	b.Timeouts.Handshake = 2 * time.Second
	n, m := ownNode(t, a), ownNode(t, b)
	ln, lm := listenTest(t), listenTest(t)
	sn := startManaged(t, func(ctx context.Context) error { return n.Serve(ctx, "dra", ln) })
	sm := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", lm) })
	p, _ := n.Peer("dra")
	q, _ := m.Peer("dra")
	waitFor(t, func() bool { return p.Info().AcceptWaiting && q.Info().AcceptWaiting })
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	da, db := gatedDialer{dialerTest(t, lm), make(chan struct{}), release}, gatedDialer{dialerTest(t, ln), make(chan struct{}), release}
	out := make(chan error, 2)
	go func() { _, e := n.Dial(testContext(t), "dra", da); out <- e }()
	go func() { _, e := m.Dial(testContext(t), "dra", db); out <- e }()
	select {
	case <-da.arrived:
	case <-time.After(time.Second):
		t.Fatal("first dial")
	}
	select {
	case <-db.arrived:
	case <-time.After(time.Second):
		t.Fatal("second dial")
	}
	once.Do(func() { close(release) })
	e1, e2 := <-out, <-out
	if !((e1 == nil && errors.Is(e2, diameter.ErrElectionLost)) || (e2 == nil && errors.Is(e1, diameter.ErrElectionLost))) {
		t.Fatal(e1, e2)
	}
	waitFor(t, func() bool { return p.Info().Ready && q.Info().Ready })
	// client.example < hss.example: the higher identity keeps the received leg.
	if p.Info().Protocol != diameter.IOpen || q.Info().Protocol != diameter.ROpen {
		t.Fatal(p.Info(), q.Info())
	}
	managedRequest(t, p, q)
	go func() { out <- n.Shutdown(testContext(t)) }()
	go func() { out <- m.Shutdown(testContext(t)) }()
	if e := <-out; e != nil {
		t.Fatal(e)
	}
	if e := <-out; e != nil {
		t.Fatal(e)
	}
	sn.wait(t)
	sm.wait(t)
	if e := n.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if e := m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if p.Info().Candidates != 0 || q.Info().Candidates != 0 {
		t.Fatal(p.Info(), q.Info())
	}
}

type stalledIncoming struct {
	transport.Incoming
	entered chan struct{}
	release <-chan struct{}
}

func (i stalledIncoming) Establish(ctx context.Context) (transport.Conn, error) {
	close(i.entered)
	<-i.release
	return i.Incoming.Establish(ctx)
}

type stalledListener struct {
	transport.Listener
	entered chan struct{}
	release <-chan struct{}
}

func (l stalledListener) Accept(ctx context.Context) (transport.Incoming, error) {
	i, e := l.Listener.Accept(ctx)
	if e != nil {
		return nil, e
	}
	return stalledIncoming{i, l.entered, l.release}, nil
}
func TestManagedSCTPWaitRetainsEstablishBorrower(t *testing.T) {
	_, b := managedConfigs(t)
	m := ownNode(t, b)
	l := listenTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	s := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", stalledListener{l, entered, release}) })
	c, e := dialerTest(t, l).Dial(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("Establish not entered")
	}
	m.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if e = m.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("borrower retired early", e)
	}
	q, _ := m.Peer("dra")
	if q.Info().Candidates != 1 {
		t.Fatal(q.Info())
	}
	once.Do(func() { close(release) })
	s.wait(t)
	if e = m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if q.Info().Candidates != 0 || q.Info().Ready {
		t.Fatal(q.Info())
	}
	if _, e = read(testContext(t), c); e == nil {
		t.Fatal("late socket remains usable")
	}
}
func TestManagedSCTPHandshakeDeadlineAndIdleListener(t *testing.T) {
	a, b := managedConfigs(t)
	a.Timeouts.Handshake = 40 * time.Millisecond
	b.Timeouts.Handshake = 40 * time.Millisecond
	t.Run("dial-no-CEA", func(t *testing.T) {
		n := ownNode(t, a)
		l := listenTest(t)
		if _, e := n.Dial(testContext(t), "dra", dialerTest(t, l)); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		p, _ := n.Peer("dra")
		if p.Info().Candidates != 0 {
			t.Fatal(p.Info())
		}
	})
	t.Run("idle-then-no-CER", func(t *testing.T) {
		m := ownNode(t, b)
		l := listenTest(t)
		q, _ := m.Peer("dra")
		r := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", l) })
		waitFor(t, func() bool { return q.Info().AcceptWaiting })
		time.Sleep(2 * b.Timeouts.Handshake)
		if !q.Info().AcceptWaiting || q.Info().ServeFailures != 0 {
			t.Fatal(q.Info())
		}
		c, e := dialerTest(t, l).Dial(testContext(t))
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		waitFor(t, func() bool { return q.Info().ServeFailures == 1 })
		if !errors.Is(q.Info().ServeLastError, context.DeadlineExceeded) {
			t.Fatal(q.Info())
		}
		r.cancel()
		r.wait(t)
		if q.Info().Candidates != 0 {
			t.Fatal(q.Info())
		}
	})
}
