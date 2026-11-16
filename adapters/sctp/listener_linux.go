//go:build linux && !386

package sctp

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

type Listener struct {
	raw       *upstream.SCTPListener
	cfg       SetupConfig
	info      transport.ListenerInfo
	accepting atomic.Bool
	closed    chan struct{}
	once      sync.Once
	closeErr  error
}

// Listen validates bounded input, resolves and binds one listener. Query
// SetupConfig.Requirements before Node construction. Binding concurrency remains
// caller-owned; Node.Serve owns the listener from admission and reserves each
// passive candidate BEFORE calling Accept. No background accept queue is made.
func Listen(ctx context.Context, network, address string, cfg SetupConfig) (*Listener, error) {
	info, e := cfg.Requirements()
	if e != nil {
		return nil, e
	}
	if address == "" {
		return nil, ErrConfig
	}
	raw, e := upstream.ListenContext(ctx, network, address, cfg.upstreamConfig())
	if e != nil {
		return nil, configError(e)
	}
	if e = setupContext(ctx); e != nil {
		raw.Close()
		return nil, e
	}
	return &Listener{raw: raw, cfg: cfg, info: info, closed: make(chan struct{})}, nil
}

func (l *Listener) Info() transport.ListenerInfo { return l.info }

// Addr is a bounded frozen bind address (not a dynamic interface enumeration).
func (l *Listener) Addr() net.Addr {
	a, e := l.raw.BoundAddr()
	if e != nil {
		return nil
	}
	return a
}
func (l *Listener) Close() error {
	l.once.Do(func() { close(l.closed); l.closeErr = l.raw.Close() })
	return l.closeErr
}
func (l *Listener) check(ctx context.Context) error {
	select {
	case <-l.closed:
		return net.ErrClosed
	default:
		return setupContext(ctx)
	}
}
func (l *Listener) Accept(ctx context.Context) (transport.Incoming, error) {
	if e := l.check(ctx); e != nil {
		return nil, e
	}
	if !l.accepting.CompareAndSwap(false, true) {
		return nil, ErrConcurrentAccept
	}
	defer l.accepting.Store(false)
	raw, e := l.raw.AcceptSCTPContext(ctx)
	if stopped := l.check(ctx); stopped != nil {
		if raw != nil {
			raw.Close()
		}
		return nil, stopped
	}
	if e != nil {
		return nil, configError(e)
	}
	return &incoming{raw: raw, cfg: l.cfg.Transport, info: transport.Info{BufferBytes: l.info.ConnectionBufferBytes}}, nil
}

type incoming struct {
	raw          *upstream.SCTPConn
	cfg          Config
	info         transport.Info
	mu           sync.Mutex
	used, closed bool
	conn         transport.Conn
	once         sync.Once
	closeErr     error
}

func (i *incoming) Info() transport.Info { return i.info }
func (i *incoming) Close() error {
	i.once.Do(func() {
		i.mu.Lock()
		i.closed = true
		c := i.conn
		i.mu.Unlock()
		if c != nil {
			i.closeErr = c.Close()
		} else {
			i.closeErr = i.raw.Close()
		}
	})
	return i.closeErr
}
func (i *incoming) Establish(ctx context.Context) (transport.Conn, error) {
	i.mu.Lock()
	if i.closed {
		i.mu.Unlock()
		return nil, net.ErrClosed
	}
	if i.used {
		i.mu.Unlock()
		return nil, ErrConfig
	}
	i.used = true
	i.mu.Unlock()
	if e := setupContext(ctx); e != nil {
		i.Close()
		return nil, e
	}
	c, e := New(i.raw, i.cfg)
	i.mu.Lock()
	if i.closed {
		e = net.ErrClosed
	} else if e == nil {
		e = setupContext(ctx)
	}
	if e == nil {
		i.conn = c
	}
	i.mu.Unlock()
	if e != nil {
		if c != nil {
			c.Close()
		}
		i.Close()
		return nil, e
	}
	return c, nil
}

var _ transport.Listener = (*Listener)(nil)
var _ transport.Incoming = (*incoming)(nil)
