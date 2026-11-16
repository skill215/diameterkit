package diameter

import (
	"context"
	"errors"
	"math"
	"syscall"
	"time"

	"github.com/skill215/diameterkit/transport"
)

var ErrManaged = errors.New("peer already has an active connection manager")

// MaintainPolicy is opt-in. Backoff is min(InitialBackoff*2^failures,
// MaxBackoff) plus [0,Jitter]. Only continuous Ready time resets failures.
// An all-zero policy disables Maintain, not manual Dial or Connect.
type MaintainPolicy struct {
	InitialBackoff, MaxBackoff, Jitter, StablePeriod time.Duration
}

func (p MaintainPolicy) valid() bool {
	return p == (MaintainPolicy{}) || p.InitialBackoff > 0 && p.MaxBackoff >= p.InitialBackoff && p.Jitter >= 0 && p.MaxBackoff <= time.Duration(math.MaxInt64)-p.Jitter && p.StablePeriod > 0
}

type maintenance struct {
	ctx                  context.Context
	policy               MaintainPolicy
	failures             uint32
	jitter               uint32
	next                 time.Time
	cause                error
	suppressed, stopping bool
}

// Dial reserves an outbound candidate before invoking the dialer. Transport,
// TLS, CER/CEA and election share Timeouts.Handshake, clipped by ctx. Return
// means established, not necessarily Ready (a reconnection starts in REOPEN).
func (n *Node) Dial(ctx context.Context, id PeerID, d transport.Dialer) (*Peer, error) {
	return n.dial(ctx, id, d, nil)
}
func (n *Node) dial(ctx context.Context, id PeerID, d transport.Dialer, owner *maintenance) (*Peer, error) {
	if d == nil {
		return nil, ErrConfig
	}
	declared := d.Info()
	c, err := n.admitConnection(ctx, id, nil, declared, true, owner, nil)
	if err != nil {
		return nil, err
	}
	dctx, cancel := context.WithDeadline(c.ctx, c.handshakeDeadline)
	defer cancel()
	stop := context.AfterFunc(ctx, func() {
		if ctx.Err() == context.Canceled {
			cancel()
		}
	})
	defer stop()
	conn, err := d.Dial(dctx)
	if err == nil {
		err = dctx.Err()
	}
	if err == nil && conn == nil {
		err = ErrConfig
	}
	if err == nil && conn.Info().BufferBytes > declared.BufferBytes {
		err = ErrBackpressure
	}
	n.mu.Lock()
	if c.closed {
		err = c.cause
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil && (n.closed || n.shutting || c.peer.lifecycle != Active) {
		err = ErrClosed
	}
	if err == nil && !time.Now().Before(c.handshakeDeadline) {
		err = context.DeadlineExceeded
	}
	if err == nil {
		c.io = conn
		c.peer.refreshHandshakeLocked()
		c.peer.signalLocked()
	}
	n.mu.Unlock()
	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return n.finishConnection(ctx, c, nil, err)
	}
	return n.handshakeConnection(dctx, c)
}

// Maintain runs one cancellable manager for an initially offline configured
// peer. The caller owns this goroutine. Cancellation closes its connection
// domain and waits for callbacks/I/O to return, but leaves the Peer reusable.
// Shutdown instead transfers existing connections to the drain/close lifecycle.
func (n *Node) Maintain(ctx context.Context, id PeerID, d transport.Dialer) error {
	if d == nil {
		return ErrConfig
	}
	info := d.Info()
	n.mu.Lock()
	p := n.peers[id]
	var err error
	switch {
	case ctx.Err() != nil:
		err = ctx.Err()
	case n.closed || n.shutting:
		err = ErrClosed
	case p == nil:
		err = ErrConfig
	case p.lifecycle != Active:
		err = ErrPeerClosed
	case p.manager != nil:
		err = ErrManaged
	case p.maintainPolicy == (MaintainPolicy{}):
		err = ErrConfig
	case !p.canStartMaintenanceLocked():
		err = ErrNotReady
	case info.BufferBytes > n.cfg.Resources.AdapterBytesPerConnection:
		err = ErrBackpressure
	}
	if err != nil {
		n.mu.Unlock()
		return err
	}
	m := &maintenance{ctx: ctx, policy: p.maintainPolicy, jitter: uint32(time.Now().UnixNano()) | 1}
	p.manager = m
	p.signalLocked()
	n.wg.Add(1)
	n.mu.Unlock()
	// Cancellation can race an uncooperative Dial. Revoke installation/Close
	// immediately while keeping the task and candidate slot until it returns.
	canceled := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { n.cancelMaintenance(p, m, ctx.Err()); close(canceled) })
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	defer func() {
		if !stop() {
			<-canceled
		}
		n.mu.Lock()
		p.manager = nil
		p.signalLocked()
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
	}()
	for {
		n.mu.Lock()
		if n.closed || n.shutting || p.lifecycle != Active {
			m.stopping = true
			cause := ErrPeerClosed
			if n.closed {
				cause = ErrClosed
			} else if n.shutting || p.lifecycle == Draining {
				cause = ErrShutdown
			}
			n.mu.Unlock()
			return cause
		}
		if ctx.Err() != nil {
			n.mu.Unlock()
			n.cancelMaintenance(p, m, ctx.Err())
			// No replacement manager may start until all old connection borrowers exit.
			n.mu.Lock()
			var wait <-chan struct{}
			for _, c := range p.candidates {
				if c != nil {
					wait = c.done
					break
				}
			}
			n.mu.Unlock()
			if wait != nil {
				<-wait
				continue
			}
			return ctx.Err()
		}
		now := time.Now()
		var wakeAt time.Time
		active := false
		for _, c := range p.candidates {
			if c != nil && (!c.accepting || c.closed) {
				active = true
			}
		}
		if c := p.conn; c != nil && p.readyLocked() && !c.okaySince.IsZero() && m.failures != 0 {
			stable := c.okaySince.Add(m.policy.StablePeriod)
			if !now.Before(stable) {
				m.failures = 0
				p.signalLocked()
			} else {
				wakeAt = stable
			}
		}
		if !active && !m.suppressed {
			if m.next.IsZero() || !now.Before(m.next) {
				m.next = time.Time{}
				n.mu.Unlock()
				_, e := n.dial(ctx, id, d, m)
				n.mu.Lock()
				// A competing passive connection or election outcome is not a failed
				// network attempt. All other dial/handshake failures have one backoff edge.
				competing := (errors.Is(e, ErrElectionLost) || errors.Is(e, ErrNotReady)) && (p.candidates[0] != nil || p.candidates[1] != nil || !m.next.IsZero() || m.suppressed)
				if e != nil && ctx.Err() == nil && p.lifecycle == Active && !n.closed && !n.shutting && !competing {
					n.scheduleReconnectLocked(p, m, e, false)
				}
				n.mu.Unlock()
				continue
			}
			wakeAt = m.next
		}
		changed := p.changed
		n.mu.Unlock()
		if wakeAt.IsZero() {
			select {
			case <-changed:
			case <-ctx.Done():
			}
		} else {
			timer.Reset(max(time.Until(wakeAt), 0))
			select {
			case <-changed:
			case <-ctx.Done():
			case <-timer.C:
			}
			timer.Stop()
		}
	}
}
func (n *Node) cancelMaintenance(p *Peer, m *maintenance, cause error) {
	n.mu.Lock()
	if p.manager != m || n.shutting || p.lifecycle != Active {
		n.mu.Unlock()
		return
	}
	m.stopping = true
	m.next = time.Time{}
	var conns []transport.Conn
	for _, c := range p.candidates {
		if c != nil {
			// Local intent takes precedence over subsequent EOF/reset from Close.
			if c.closeCause == nil {
				c.closeCause = cause
			}
			n.stopConnectionLocked(c, cause)
			if c.io != nil {
				conns = append(conns, c.io)
			}
		}
	}
	p.signalLocked()
	n.mu.Unlock()
	for _, c := range conns {
		_ = c.Close()
	}
	n.notify()
}
func (n *Node) maintainLossLocked(c *connection, cause error) {
	p := c.peer
	m := p.manager
	if m == nil || m.stopping || m.ctx.Err() != nil || p.lifecycle != Active || n.closed || n.shutting {
		return
	}
	if !c.okaySince.IsZero() && time.Since(c.okaySince) >= m.policy.StablePeriod {
		m.failures = 0
	}
	reboot := c.remoteDPR && c.disconnectCause == 0 || !c.identifierDrain.IsZero()
	n.scheduleReconnectLocked(p, m, cause, reboot)
}
func (n *Node) scheduleReconnectLocked(p *Peer, m *maintenance, cause error, reboot bool) {
	m.cause = cause
	if m.suppressed || !reboot && !reconnectError(cause) {
		m.suppressed = true
		m.next = time.Time{}
		p.signalLocked()
		return
	}
	delay := m.policy.InitialBackoff
	for i := uint32(0); i < m.failures && delay < m.policy.MaxBackoff; i++ {
		if delay > m.policy.MaxBackoff/2 {
			delay = m.policy.MaxBackoff
		} else {
			delay *= 2
		}
	}
	m.jitter ^= m.jitter << 13
	m.jitter ^= m.jitter >> 17
	m.jitter ^= m.jitter << 5
	if m.policy.Jitter > 0 {
		delay += time.Duration(uint64(m.jitter) % (uint64(m.policy.Jitter) + 1))
	}
	m.next = time.Now().Add(delay)
	if m.failures < math.MaxUint32 {
		m.failures++
	}
	p.signalLocked()
}
func reconnectError(err error) bool {
	return ioStopReason(err) == StopTransport || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, ErrWatchdog)
}

// Wait waits for this peer's connection/manager/preparation borrowers, not
// application Inbounds or completion-port consumption. It does not close a Peer.
func (p *Peer) Wait(ctx context.Context) error {
	for {
		p.node.mu.Lock()
		if p.quiescedLocked() {
			p.node.mu.Unlock()
			return nil
		}
		ch := p.changed
		p.node.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Peer) canStartMaintenanceLocked() bool {
	if p.preparing != 0 {
		return false
	}
	for _, c := range p.candidates {
		if c != nil && (!c.accepting || c.closed) {
			return false
		}
	}
	return true
}
