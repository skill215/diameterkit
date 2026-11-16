package diameter

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	dtcp "github.com/skill215/diameterkit/transport/tcp"
	dtls "github.com/skill215/diameterkit/transport/tls"
)

func serveConfig(t *testing.T) NodeConfig {
	c := inboundConfig(t)
	c.Resources.ListenerBytesPerPeer = dtcp.ListenerBufferBytes
	return c
}
func newServeListener(t *testing.T, encrypted bool) (transport.Listener, transport.Dialer) {
	t.Helper()
	raw, e := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = raw.Close() })
	var l transport.Listener
	var d transport.Dialer
	if encrypted {
		client, server := tlsConfigs(t)
		l, e = dtls.NewListener(raw, dtls.Config{TLS: server, HandshakeTimeout: time.Second})
		if e != nil {
			t.Fatal(e)
		}
		d, e = dtls.NewDialer("tcp4", raw.Addr().String(), dtls.Config{TLS: client, HandshakeTimeout: time.Second})
	} else {
		l, e = dtcp.NewListener(raw)
		if e != nil {
			t.Fatal(e)
		}
		d, e = dtcp.NewDialer("tcp4", raw.Addr().String())
	}
	if e != nil {
		t.Fatal(e)
	}
	return l, d
}

type serveRun struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

func startServe(t *testing.T, n *Node, l transport.Listener) *serveRun {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	r := &serveRun{cancel: cancel, done: make(chan struct{})}
	go func() { r.err = n.Serve(ctx, "dra", l); close(r.done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.done:
		case <-time.After(2 * time.Second):
			t.Error("Serve cleanup blocked")
		}
	})
	return r
}
func (r *serveRun) wait(t *testing.T) error {
	t.Helper()
	select {
	case <-r.done:
		return r.err
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return")
		return nil
	}
}
func serveRemote(t *testing.T, n *Node, co *codec.Codec, d transport.Dialer) transport.Conn {
	t.Helper()
	c, e := d.Dial(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = c.Close() })
	if e = writeWire(ctxTest(t), c, co, peerCER(n)); e != nil {
		t.Fatal(e)
	}
	ans := routeRead(t, co, c)
	if ans.Header.CommandCode != 257 || ans.Body.(*base.CEA).ResultCode != 2001 {
		t.Fatal(ans)
	}
	p, _ := n.Peer("dra")
	waitFor(t, func() bool { return p.Info().Ready })
	return c
}
func TestServeTCPAndTLSInboundOutboundAndCancel(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		r := startServe(t, n, l)
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		if p.Info().Protocol != ProtocolClosed || p.Info().Candidates != 1 {
			t.Fatal(p.Info())
		}
		remote := serveRemote(t, n, co, d)
		in, req := receiveInbound(t, n, p, remote, co, 100)
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		if e := co.ValidateAnswer(req, routeRead(t, co, remote), pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		r.cancel()
		if e := r.wait(t); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if !p.Info().Ready || p.Info().Serving {
			t.Fatal("Serve cancellation closed established peer", p.Info())
		}
		port := portTest(t, n, 2)
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, remote, routeRead(t, co, remote))
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
	})
}
func TestServeIdleCancellationAndShutdown(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, action := range []string{"cancel", "node-close", "peer-close", "shutdown"} {
			t.Run(action, func(t *testing.T) {
				n, _ := newInboundNode(t, serveConfig(t))
				p, _ := n.Peer("dra")
				l, _ := newServeListener(t, tls)
				r := startServe(t, n, l)
				waitFor(t, func() bool { return p.Info().AcceptWaiting })
				switch action {
				case "cancel":
					r.cancel()
				case "node-close":
					_ = n.Close()
				case "peer-close":
					_ = p.Close()
				case "shutdown":
					if e := n.Shutdown(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
				}
				if e := r.wait(t); e == nil {
					t.Fatal("missing stop cause")
				}
				if e := p.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				if p.Info().Serving || p.Info().Candidates != 0 || n.Resources().OutboundBytes != 0 {
					t.Fatal(p.Info(), n.Resources())
				}
			})
		}
	})
}
func TestServeIdleDoesNotSpendHandshakeDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := serveConfig(t)
		cfg.Timeouts.Handshake = 60 * time.Millisecond
		n, co := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		r := startServe(t, n, l)
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		time.Sleep(80 * time.Millisecond)
		select {
		case <-r.done:
			t.Fatal(r.err)
		default:
		}
		_ = serveRemote(t, n, co, d)
	})
}
func TestListenerCanceledAcceptCanBeReused(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		l, d := newServeListener(t, tls)
		defer l.Close()
		for range 5 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			incoming, e := l.Accept(ctx)
			cancel()
			if incoming != nil || !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal(incoming, e)
			}
		}
		incoming := make(chan transport.Incoming, 1)
		errs := make(chan error, 1)
		go func() { i, e := l.Accept(ctxTest(t)); incoming <- i; errs <- e }()
		conn := make(chan transport.Conn, 1)
		go func() { c, e := d.Dial(ctxTest(t)); conn <- c; errs <- e }()
		i := <-incoming
		if e := <-errs; e != nil {
			t.Fatal(e)
		}
		server, e := i.Establish(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		defer server.Close()
		client := <-conn
		if e := <-errs; e != nil {
			t.Fatal(e)
		}
		defer client.Close()
		f := frame(20, 0)
		if w := client.WriteFrame(ctxTest(t), f); w.Err != nil {
			t.Fatal(w)
		}
		if _, e := server.ReadPrefix(ctxTest(t), time.Second); e != nil {
			t.Fatal(e)
		}
	})
}

type observedListener struct {
	transport.Listener
	accepts atomic.Int32
	closes  atomic.Int32
}

func (l *observedListener) Accept(ctx context.Context) (transport.Incoming, error) {
	l.accepts.Add(1)
	return l.Listener.Accept(ctx)
}
func (l *observedListener) Close() error { l.closes.Add(1); return l.Listener.Close() }
func TestServeDoesNotAcceptPastCandidateCapacity(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		raw, d := newServeListener(t, tls)
		l := &observedListener{Listener: raw}
		r := startServe(t, n, l)
		remote := serveRemote(t, n, co, d)
		// Queue another TCP socket in the kernel backlog; no Go-side accept/handshake.
		addr := raw.(interface{ Addr() net.Addr }).Addr().String()
		extra, e := net.DialTimeout("tcp4", addr, time.Second)
		if e != nil {
			t.Fatal(e)
		}
		defer extra.Close()
		if l.accepts.Load() != 1 || p.Info().Candidates != 1 {
			t.Fatal(l.accepts.Load(), p.Info())
		}
		r.cancel()
		_ = r.wait(t)
		if !p.Info().Ready || l.accepts.Load() != 1 {
			t.Fatal(p.Info())
		}
		_ = remote
	})
}
func TestServeTLSAndCERDeadlineAndRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := serveConfig(t)
		cfg.Timeouts.Handshake = 45 * time.Millisecond
		n, co := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		r := startServe(t, n, l)
		addr := l.(interface{ Addr() net.Addr }).Addr().String()
		// TLS gets no ClientHello; TCP gets no CER. Both consume a finite candidate.
		raw, e := net.DialTimeout("tcp4", addr, time.Second)
		if e != nil {
			t.Fatal(e)
		}
		defer raw.Close()
		waitFor(t, func() bool { return p.Info().ServeLastError != nil })
		if !errors.Is(p.Info().ServeLastError, context.DeadlineExceeded) {
			t.Fatal(p.Info())
		}
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		_ = serveRemote(t, n, co, d)
		r.cancel()
		_ = r.wait(t)
	})
}
func TestServeInvalidCERDoesNotInstallPeer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		r := startServe(t, n, l)
		remote, e := d.Dial(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		defer remote.Close()
		req := peerCER(n)
		req.Body.(*base.CER).OriginHost = "wrong.example"
		if e = writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return p.Info().ServeLastError != nil })
		if p.Info().Ready {
			t.Fatal(p.Info())
		}
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		_ = serveRemote(t, n, co, d)
		r.cancel()
		_ = r.wait(t)
	})
}

type stubListener struct {
	info            transport.ListenerInfo
	accept          func(context.Context) (transport.Incoming, error)
	close           func() error
	accepts, closes atomic.Int32
}

func (l *stubListener) Info() transport.ListenerInfo { return l.info }
func (l *stubListener) Accept(ctx context.Context) (transport.Incoming, error) {
	l.accepts.Add(1)
	return l.accept(ctx)
}
func (l *stubListener) Close() error {
	l.closes.Add(1)
	if l.close != nil {
		return l.close()
	}
	return nil
}

type stubIncoming struct {
	bytes         uint64
	establish     func(context.Context) (transport.Conn, error)
	close         func() error
	calls, closes atomic.Int32
}

func (i *stubIncoming) Info() transport.Info { return transport.Info{BufferBytes: i.bytes} }
func (i *stubIncoming) Establish(ctx context.Context) (transport.Conn, error) {
	i.calls.Add(1)
	return i.establish(ctx)
}
func (i *stubIncoming) Close() error {
	i.closes.Add(1)
	if i.close != nil {
		return i.close()
	}
	return nil
}
func TestServeAdmissionAndListenerOwnership(t *testing.T) {
	for _, mode := range []string{"disabled", "listener-limit", "conn-limit", "unknown", "canceled", "closed"} {
		t.Run(mode, func(t *testing.T) {
			cfg := serveConfig(t)
			if mode == "disabled" {
				cfg.Resources.ListenerBytesPerPeer = 0
			}
			n, _ := newInboundNode(t, cfg)
			l := &stubListener{accept: func(context.Context) (transport.Incoming, error) {
				t.Error("Accept before admission")
				return nil, io.EOF
			}}
			if mode == "listener-limit" {
				l.info.BufferBytes = cfg.Resources.ListenerBytesPerPeer + 1
			}
			if mode == "conn-limit" {
				l.info.ConnectionBufferBytes = cfg.Resources.AdapterBytesPerConnection + 1
			}
			id := PeerID("dra")
			if mode == "unknown" {
				id = "missing"
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "canceled" {
				cancel()
			}
			if mode == "closed" {
				_ = n.Close()
			}
			if e := n.Serve(ctx, id, l); e == nil {
				t.Fatal("invalid admission accepted")
			}
			if l.accepts.Load() != 0 || l.closes.Load() != 1 {
				t.Fatal(l.accepts.Load(), l.closes.Load())
			}
		})
	}
}
func TestServeExclusiveOwnerAndManualAccept(t *testing.T) {
	n, co := newInboundNode(t, serveConfig(t))
	p, _ := n.Peer("dra")
	l, _ := newServeListener(t, false)
	r := startServe(t, n, l)
	waitFor(t, func() bool { return p.Info().AcceptWaiting })
	other := &stubListener{}
	if e := n.Serve(ctxTest(t), "dra", other); !errors.Is(e, ErrServing) || other.closes.Load() != 1 {
		t.Fatal(e)
	}
	a, b := networkAdapters(t, false)
	defer b.Close()
	if _, e := n.Accept(ctxTest(t), "dra", a); !errors.Is(e, ErrServing) {
		t.Fatal(e)
	}
	r.cancel()
	_ = r.wait(t)
	replacement, d := newServeListener(t, false)
	again := startServe(t, n, replacement)
	_ = serveRemote(t, n, co, d)
	again.cancel()
	_ = again.wait(t)
}
func TestServeLateAcceptRetainsCandidateAndWait(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, _ := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		a, b := networkAdapters(t, tls)
		defer b.Close()
		incoming := &stubIncoming{close: a.Close, establish: func(context.Context) (transport.Conn, error) { return a, nil }}
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		l := &stubListener{accept: func(context.Context) (transport.Incoming, error) { close(entered); <-release; return incoming, nil }}
		r := startServe(t, n, l)
		<-entered
		r.cancel()
		waitFor(t, func() bool { return l.closes.Load() == 1 })
		if p.Info().Candidates != 1 || !p.Info().Serving {
			t.Fatal("borrower released early", p.Info())
		}
		_ = n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if e := n.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		unblock()
		if e := r.wait(t); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if incoming.calls.Load() != 0 || incoming.closes.Load() != 1 || p.Info().Candidates != 0 {
			t.Fatal(incoming.calls.Load(), incoming.closes.Load(), p.Info())
		}
	})
}
func TestServeLateEstablishRetainsCandidateAndClosesResult(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, _ := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		a, b := networkAdapters(t, tls)
		defer b.Close()
		tracked := &trackedDialConn{Conn: a}
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		incoming := &stubIncoming{close: tracked.Close, establish: func(context.Context) (transport.Conn, error) { close(entered); <-release; return tracked, nil }}
		l := &stubListener{info: transport.ListenerInfo{ConnectionBufferBytes: a.Info().BufferBytes}, accept: func(context.Context) (transport.Incoming, error) { return incoming, nil }}
		r := startServe(t, n, l)
		<-entered
		_ = p.Close()
		waitFor(t, func() bool { return tracked.Closed() })
		if p.Info().Candidates != 1 {
			t.Fatal(p.Info())
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if e := p.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		unblock()
		if e := r.wait(t); !errors.Is(e, ErrPeerClosed) {
			t.Fatal(e)
		}
		if e := p.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if p.Info().Ready || p.Info().Candidates != 0 {
			t.Fatal(p.Info())
		}
	})
}
func TestServeIncomingAndEstablishedDeclarationChecks(t *testing.T) {
	for _, mode := range []string{"accept-error", "nil-incoming", "incoming-limit", "establish-error", "nil-conn", "conn-limit"} {
		t.Run(mode, func(t *testing.T) {
			n, _ := newInboundNode(t, serveConfig(t))
			a, b := networkAdapters(t, false)
			defer b.Close()
			tracked := &trackedDialConn{Conn: a}
			incoming := &stubIncoming{close: tracked.Close, establish: func(context.Context) (transport.Conn, error) { return tracked, nil }}
			l := &stubListener{info: transport.ListenerInfo{ConnectionBufferBytes: a.Info().BufferBytes}, accept: func(context.Context) (transport.Incoming, error) { return incoming, nil }}
			switch mode {
			case "accept-error":
				l.accept = func(context.Context) (transport.Incoming, error) { return incoming, io.EOF }
			case "nil-incoming":
				l.accept = func(context.Context) (transport.Incoming, error) { return nil, nil }
				defer tracked.Close()
			case "incoming-limit":
				incoming.bytes = l.info.ConnectionBufferBytes + 1
			case "establish-error":
				incoming.establish = func(context.Context) (transport.Conn, error) { return tracked, ErrConfig }
			case "nil-conn":
				incoming.establish = func(context.Context) (transport.Conn, error) { return nil, nil }
			case "conn-limit":
				l.info.ConnectionBufferBytes = 0
			}
			if e := n.Serve(ctxTest(t), "dra", l); e == nil {
				t.Fatal("invalid adapter accepted")
			}
			if mode != "nil-incoming" && !tracked.Closed() {
				t.Fatal("owned socket leaked")
			}
			p, _ := n.Peer("dra")
			if p.Info().Candidates != 0 || p.Info().Serving {
				t.Fatal(p.Info())
			}
		})
	}
}
func TestServeShutdownKeepsEstablishedDrain(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		r := startServe(t, n, l)
		remote := serveRemote(t, n, co, d)
		port := portTest(t, n, 2)
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		done := make(chan error, 1)
		go func() { done <- n.Shutdown(ctxTest(t)) }()
		waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
		r.cancel()
		if e := r.wait(t); !errors.Is(e, ErrShutdown) {
			t.Fatal(e)
		}
		routeReply(t, co, remote, req)
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
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
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
	})
}
func TestServeFixedReserve(t *testing.T) {
	cfg := serveConfig(t)
	n, _ := newInboundNode(t, cfg)
	rp := cfg.Resources
	minimum := 2*(n.controlBound+rp.AdapterBytesPerConnection) + 4096 + uint64(cfg.Peers[0].MaxPending)*1024 + uint64(cfg.Peers[0].MaxInbound)*512 + uint64(rp.MaxPorts*rp.MaxPortOutstanding)*512 + uint64(rp.MaxPorts)*512 + schedulerReserve(rp, 1) + listenerReserve(rp, 1)
	cfg.Resources.TransportBytes = minimum - 1
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			_ = bad.Close()
		}
		t.Fatal(e)
	}
	cfg.Resources.TransportBytes = minimum
	exact, _ := newInboundNode(t, cfg)
	p, _ := exact.Peer("dra")
	before := exact.Resources()
	l, _ := newServeListener(t, false)
	r := startServe(t, exact, l)
	waitFor(t, func() bool { return p.Info().AcceptWaiting })
	if exact.Resources() != before {
		t.Fatal("listening borrowed business quota", exact.Resources(), before)
	}
	r.cancel()
	_ = r.wait(t)
	cfg.Resources.ListenerBytesPerPeer = 1<<20 + 1
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			_ = bad.Close()
		}
		t.Fatal(e)
	}
}

func TestServeAndDialElection(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		low, high := serveConfig(t), serveConfig(t)
		low.Origin.Host = "alpha.example"
		high.Origin.Host = "zulu.example"
		low.Peers[0].Origin = high.Origin
		high.Peers[0].Origin = low.Origin
		a, _ := newInboundNode(t, low)
		b, _ := newInboundNode(t, high)
		al, ad := newServeListener(t, tls)
		bl, bd := newServeListener(t, tls)
		ar, br := startServe(t, a, al), startServe(t, b, bl)
		ap, _ := a.Peer("dra")
		bp, _ := b.Peer("dra")
		waitFor(t, func() bool { return ap.Info().AcceptWaiting && bp.Info().AcceptWaiting })
		entered, release := make(chan struct{}, 2), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		gate := func(d transport.Dialer) transport.Dialer {
			return testDialer{bytes: d.Info().BufferBytes, call: func(ctx context.Context) (transport.Conn, error) {
				c, e := d.Dial(ctx)
				entered <- struct{}{}
				<-release
				return c, e
			}}
		}
		done := make(chan error, 2)
		go func() { _, e := a.Dial(ctxTest(t), "dra", gate(bd)); done <- e }()
		go func() { _, e := b.Dial(ctxTest(t), "dra", gate(ad)); done <- e }()
		<-entered
		<-entered
		if ap.Info().Candidates != 2 || bp.Info().Candidates != 2 {
			t.Fatal(ap.Info(), bp.Info())
		}
		unblock()
		e1, e2 := <-done, <-done
		if (e1 == nil) == (e2 == nil) {
			t.Fatal("expected one active winner", e1, e2)
		}
		waitFor(t, func() bool {
			return ap.Info().Ready && bp.Info().Ready && ap.Info().Candidates == 1 && bp.Info().Candidates == 1
		})
		if ap.Info().Protocol != IOpen || bp.Info().Protocol != ROpen {
			t.Fatal(ap.Info(), bp.Info())
		}
		// Crossed DPR/DPA while Serve owns listeners; both Close paths must quiesce.
		closed := make(chan error, 2)
		go func() { closed <- a.Shutdown(ctxTest(t)) }()
		go func() { closed <- b.Shutdown(ctxTest(t)) }()
		for range 2 {
			if e := <-closed; e != nil {
				t.Fatal(e)
			}
		}
		if e := ar.wait(t); !errors.Is(e, ErrShutdown) {
			t.Fatal(e)
		}
		if e := br.wait(t); !errors.Is(e, ErrShutdown) {
			t.Fatal(e)
		}
	})
}
func TestServePromotesPassiveAfterActiveFailure(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		done := make(chan error, 1)
		go func() {
			_, e := n.Dial(ctxTest(t), "dra", testDialer{call: func(context.Context) (transport.Conn, error) { close(entered); <-release; return nil, io.EOF }})
			done <- e
		}()
		<-entered
		r := startServe(t, n, l)
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		n.mu.Lock()
		idleDeadline := p.server.candidate.handshakeDeadline
		n.mu.Unlock()
		if !idleDeadline.IsZero() {
			t.Fatal("idle Accept inherited active handshake expiry", idleDeadline)
		}
		remote, e := d.Dial(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		defer remote.Close()
		if e = writeWire(ctxTest(t), remote, co, peerCER(n)); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			for _, c := range p.candidates {
				if c != nil && !c.initiator && c.prefixResult != nil {
					return true
				}
			}
			return false
		})
		unblock()
		if e = <-done; !errors.Is(e, io.EOF) {
			t.Fatal(e)
		}
		if a := routeRead(t, co, remote); a.Header.CommandCode != 257 {
			t.Fatal(a)
		}
		waitFor(t, func() bool { return p.Info().Ready && p.Info().Candidates == 1 })
		r.cancel()
		_ = r.wait(t)
	})
}
func TestServeCanCoexistWithMaintain(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := maintainConfig(t)
		cfg.Resources.ListenerBytesPerPeer = dtcp.ListenerBufferBytes
		n, co := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		l, _ := newServeListener(t, tls)
		r := startServe(t, n, l)
		waitFor(t, func() bool { return p.Info().AcceptWaiting })
		outbound, d := dialListener(t, tls)
		m := startMaintain(t, n, d)
		remote := acceptDial(t, outbound, tls)
		defer remote.Close()
		answerCER(t, n, co, remote)
		waitFor(t, func() bool { return p.Info().Ready && p.Info().Candidates == 1 })
		if !p.Info().Serving || !p.Info().Managed {
			t.Fatal(p.Info())
		}
		r.cancel()
		_ = r.wait(t)
		if !p.Info().Ready {
			t.Fatal(p.Info())
		}
		m.cancel()
		_ = m.wait(t)
	})
}
func TestServeCancelDuringHandshake(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, action := range []string{"cancel", "close", "shutdown"} {
			t.Run(action, func(t *testing.T) {
				n, _ := newInboundNode(t, serveConfig(t))
				p, _ := n.Peer("dra")
				l, _ := newServeListener(t, tls)
				r := startServe(t, n, l)
				raw, e := net.DialTimeout("tcp4", l.(interface{ Addr() net.Addr }).Addr().String(), time.Second)
				if e != nil {
					t.Fatal(e)
				}
				defer raw.Close()
				waitFor(t, func() bool { return p.Info().Candidates == 1 && !p.Info().AcceptWaiting })
				switch action {
				case "cancel":
					r.cancel()
				case "close":
					_ = n.Close()
				case "shutdown":
					if e = n.Shutdown(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
				}
				if e = r.wait(t); e == nil {
					t.Fatal("missing cause")
				}
				if e = p.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				if p.Info().Ready || p.Info().Candidates != 0 {
					t.Fatal(p.Info())
				}
			})
		}
	})
}

type ceaWriteBarrier struct {
	transport.Conn
	entered, release chan struct{}
}

func (b *ceaWriteBarrier) WriteFrame(ctx context.Context, frame []byte) transport.WriteResult {
	r := b.Conn.WriteFrame(ctx, frame)
	if len(frame) >= 20 && frame[5] == 0 && frame[6] == 1 && frame[7] == 1 && frame[4]&pdu.RequestFlag == 0 {
		close(b.entered)
		<-b.release
	}
	return r
}
func TestServeCancellationCannotInstallLateCEA(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, serveConfig(t))
		p, _ := n.Peer("dra")
		l, d := newServeListener(t, tls)
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		unblock := func() { once.Do(func() { close(release) }) }
		defer unblock()
		wrapped := &stubListener{info: l.Info(), close: l.Close, accept: func(ctx context.Context) (transport.Incoming, error) {
			in, e := l.Accept(ctx)
			if e != nil {
				return nil, e
			}
			return &stubIncoming{bytes: in.Info().BufferBytes, close: in.Close, establish: func(ctx context.Context) (transport.Conn, error) {
				c, e := in.Establish(ctx)
				if e != nil {
					return nil, e
				}
				return &ceaWriteBarrier{c, entered, release}, nil
			}}, nil
		}}
		r := startServe(t, n, wrapped)
		remote, e := d.Dial(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		defer remote.Close()
		if e = writeWire(ctxTest(t), remote, co, peerCER(n)); e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, remote)
		<-entered
		r.cancel()
		waitFor(t, func() bool { return wrapped.closes.Load() > 0 })
		if p.Info().Ready || p.Info().Candidates != 1 {
			t.Fatal(p.Info())
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if e = p.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		unblock()
		if e = r.wait(t); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if p.Info().Ready || p.Info().Candidates != 0 {
			t.Fatal(p.Info())
		}
	})
}
func TestServeEstablishAndCERShareDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := serveConfig(t)
		cfg.Timeouts.Handshake = 70 * time.Millisecond
		n, _ := newInboundNode(t, cfg)
		p, _ := n.Peer("dra")
		a, b := networkAdapters(t, tls)
		defer b.Close()
		deadlines := make(chan time.Time, 1)
		incoming := &stubIncoming{bytes: a.Info().BufferBytes, close: a.Close, establish: func(ctx context.Context) (transport.Conn, error) {
			d, _ := ctx.Deadline()
			deadlines <- d
			time.Sleep(30 * time.Millisecond)
			return a, nil
		}}
		var first atomic.Bool
		l := &stubListener{info: transport.ListenerInfo{ConnectionBufferBytes: a.Info().BufferBytes}, accept: func(ctx context.Context) (transport.Incoming, error) {
			if first.CompareAndSwap(false, true) {
				return incoming, nil
			}
			<-ctx.Done()
			return nil, ctx.Err()
		}}
		r := startServe(t, n, l)
		deadline := <-deadlines
		waitFor(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			return p.server != nil && p.server.candidate != nil && p.server.candidate.io != nil
		})
		n.mu.Lock()
		actual := p.server.candidate.handshakeDeadline
		n.mu.Unlock()
		if !actual.Equal(deadline) {
			t.Fatal("CER got a fresh deadline", actual, deadline)
		}
		waitFor(t, func() bool { return errors.Is(p.Info().ServeLastError, context.DeadlineExceeded) })
		r.cancel()
		_ = r.wait(t)
	})
}
func TestTLSListenerFreezesConfigAndRejectsUnauthenticated(t *testing.T) {
	raw, e := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	client, server := tlsConfigs(t)
	l, e := dtls.NewListener(raw, dtls.Config{TLS: server, HandshakeTimeout: time.Second})
	if e != nil {
		t.Fatal(e)
	}
	// Changes to the caller's top-level policy must not disable listener mTLS.
	server.ClientAuth = 0
	n, co := newInboundNode(t, serveConfig(t))
	p, _ := n.Peer("dra")
	r := startServe(t, n, l)
	bad := client.Clone()
	bad.Certificates = nil
	d, e := dtls.NewDialer("tcp4", raw.Addr().String(), dtls.Config{TLS: bad, HandshakeTimeout: time.Second})
	if e != nil {
		t.Fatal(e)
	}
	c, e := d.Dial(ctxTest(t))
	if c != nil {
		_, _ = c.ReadPrefix(ctxTest(t), time.Second)
		_ = c.Close()
	}
	waitFor(t, func() bool { return p.Info().ServeFailures == 1 })
	if p.Info().Ready {
		t.Fatal("unauthenticated connection installed")
	}
	good, e := dtls.NewDialer("tcp4", raw.Addr().String(), dtls.Config{TLS: client, HandshakeTimeout: time.Second})
	if e != nil {
		t.Fatal(e)
	}
	_ = serveRemote(t, n, co, good)
	r.cancel()
	_ = r.wait(t)
}
