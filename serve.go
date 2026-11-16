package diameter

import (
	"context"
	"errors"
	"time"

	"github.com/skill215/diameterkit/transport"
)

var ErrServing = errors.New("peer already has a listener owner")

type serving struct {
	cancel    context.CancelFunc
	cause     error
	candidate *connection
	failures  uint64
	lastError error
}

func listenerReserve(r ResourcePlan, peers int) uint64 {
	if r.ListenerBytesPerPeer == 0 {
		return 0
	}
	return uint64(peers) * (4096 + r.ListenerBytesPerPeer)
}
func (n *Node) stopServingLocked(p *Peer, cause error) {
	if s := p.server; s != nil && s.cause == nil {
		s.cause = cause
		s.cancel()
		p.signalLocked()
	}
}

// Serve owns l, including on admission failure. It serially accepts for one
// configured Peer, taking a passive candidate before Accept. Handshake failures
// close that candidate and retry after 100ms; Accept errors terminate Serve.
// Cancellation closes the listener and unfinished candidate, not established
// connections. Peer/Node Close or Shutdown own those connections. Return waits
// for all listener/handshake callbacks; Node/Peer.Wait include this borrower.
func (n *Node) Serve(ctx context.Context, id PeerID, l transport.Listener) (result error) {
	if l == nil {
		return ErrConfig
	}
	info := l.Info()
	n.mu.Lock()
	p := n.peers[id]
	switch {
	case ctx.Err() != nil:
		result = ctx.Err()
	case n.closed || n.shutting:
		result = ErrClosed
	case p == nil:
		result = ErrConfig
	case p.lifecycle != Active:
		result = ErrPeerClosed
	case p.server != nil:
		result = ErrServing
	case n.cfg.Resources.ListenerBytesPerPeer == 0:
		result = ErrConfig
	case info.BufferBytes > n.cfg.Resources.ListenerBytesPerPeer || info.ConnectionBufferBytes > n.cfg.Resources.AdapterBytesPerConnection:
		result = ErrBackpressure
	}
	if result != nil {
		n.mu.Unlock()
		_ = l.Close()
		return result
	}
	life, cancel := context.WithCancel(context.Background())
	s := &serving{cancel: cancel}
	p.server = s
	p.signalLocked()
	n.wg.Add(1)
	n.mu.Unlock()
	closed := make(chan struct{})
	context.AfterFunc(life, func() {
		n.mu.Lock()
		c := s.candidate
		var io transport.Conn
		if c != nil && !c.established {
			n.stopConnectionLocked(c, s.cause)
			io = c.io
		}
		n.mu.Unlock()
		if io != nil {
			_ = io.Close()
		}
		_ = l.Close()
		close(closed)
	})
	callerDone := make(chan struct{})
	stopCaller := context.AfterFunc(ctx, func() { n.mu.Lock(); n.stopServingLocked(p, ctx.Err()); n.mu.Unlock(); close(callerDone) })
	defer func() {
		if !stopCaller() {
			<-callerDone
		}
		n.mu.Lock()
		n.stopServingLocked(p, result)
		if s.cause == nil {
			s.cause = ErrClosed
			s.cancel()
		}
		n.mu.Unlock()
		<-closed
		n.mu.Lock()
		if s.cause != nil {
			result = s.cause
		}
		p.server = nil
		p.signalLocked()
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
	}()
	for {
		n.mu.Lock()
		if s.cause != nil {
			result = s.cause
			n.mu.Unlock()
			return
		}
		ready := p.conn == nil && p.chosen == nil
		for _, c := range p.candidates {
			if c != nil && (!c.initiator || c.closed) {
				ready = false
			}
		}
		if p.manager != nil && (p.manager.stopping || p.manager.ctx.Err() != nil) {
			ready = false
		}
		changed := p.changed
		n.mu.Unlock()
		if !ready {
			select {
			case <-changed:
			case <-life.Done():
			}
			continue
		}
		c, err := n.admitConnection(life, id, nil, transport.Info{BufferBytes: info.ConnectionBufferBytes}, false, nil, s)
		if errors.Is(err, ErrNotReady) {
			continue
		}
		if err != nil {
			return err
		}
		n.mu.Lock()
		s.candidate = c
		if s.cause != nil {
			n.stopConnectionLocked(c, s.cause)
		}
		n.mu.Unlock()
		err, terminal := n.serveCandidate(ctx, c, l, info)
		n.mu.Lock()
		s.candidate = nil
		if err != nil {
			s.lastError = err
			if s.failures < ^uint64(0) {
				s.failures++
			}
			p.signalLocked()
		}
		cause := s.cause
		n.mu.Unlock()
		if cause != nil {
			return cause
		}
		if terminal {
			return err
		}
		if err != nil && !errors.Is(err, ErrElectionLost) {
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-timer.C:
			case <-life.Done():
			}
			timer.Stop()
		}
	}
}
func (n *Node) serveCandidate(caller context.Context, c *connection, l transport.Listener, declared transport.ListenerInfo) (error, bool) {
	incoming, err := l.Accept(c.ctx)
	n.mu.Lock()
	stopped := c.closed
	if stopped {
		err = c.cause
	}
	n.mu.Unlock()
	if err == nil && incoming == nil {
		err = ErrConfig
	}
	if err != nil {
		if incoming != nil {
			_ = incoming.Close()
		}
		_, e := n.finishConnection(caller, c, nil, err)
		return e, !stopped
	}
	if incoming.Info().BufferBytes > declared.ConnectionBufferBytes {
		_ = incoming.Close()
		_, e := n.finishConnection(caller, c, nil, ErrBackpressure)
		return e, true
	}
	n.mu.Lock()
	if c.closed || caller.Err() != nil {
		err = c.cause
		if err == nil {
			err = caller.Err()
		}
		n.mu.Unlock()
		_ = incoming.Close()
		_, e := n.finishConnection(caller, c, nil, err)
		return e, false
	}
	c.accepting = false
	deadline := time.Now().Add(n.cfg.Timeouts.Handshake)
	if d, ok := caller.Deadline(); ok {
		deadline = earlier(deadline, d)
	}
	for _, other := range c.peer.candidates {
		if other != nil && other != c && !other.closed && !other.accepting {
			deadline = earlier(deadline, other.handshakeDeadline)
		}
	}
	c.handshakeDeadline = deadline
	c.peer.handshakeDeadline = deadline
	c.peer.refreshHandshakeLocked()
	c.peer.signalLocked()
	n.mu.Unlock()
	hctx, cancel := context.WithDeadline(c.ctx, deadline)
	closed := make(chan struct{})
	stop := context.AfterFunc(hctx, func() { _ = incoming.Close(); close(closed) })
	conn, err := incoming.Establish(hctx)
	if !stop() {
		<-closed
	}
	if err == nil {
		err = hctx.Err()
	}
	if err == nil && conn == nil {
		err = ErrConfig
	}
	if err == nil && conn.Info().BufferBytes > declared.ConnectionBufferBytes {
		err = ErrBackpressure
	}
	n.mu.Lock()
	if c.closed {
		err = c.cause
	} else if caller.Err() != nil {
		err = caller.Err()
	} else if !time.Now().Before(deadline) {
		err = context.DeadlineExceeded
	}
	if err == nil {
		c.io = conn
	}
	n.mu.Unlock()
	if err != nil {
		cancel()
		if conn != nil {
			_ = conn.Close()
		}
		_ = incoming.Close()
		_, e := n.finishConnection(caller, c, nil, err)
		return e, errors.Is(err, ErrConfig) || errors.Is(err, ErrBackpressure)
	}
	_, err = n.handshakeConnection(caller, c)
	cancel()
	return err, false
}
