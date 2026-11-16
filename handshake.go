package diameter

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"
	"strings"
	"time"

	"github.com/skill215/diameterkit/transport"
)

type prefixResult struct {
	prefix [20]byte
	err    error
}

func (p *Peer) signalLocked() { close(p.changed); p.changed = make(chan struct{}) }
func (p *Peer) refreshHandshakeLocked() {
	if p.conn != nil {
		return
	}
	count := 0
	var candidate *connection
	for _, c := range p.candidates {
		if c != nil && !c.closed && !c.accepting {
			count++
			candidate = c
		}
	}
	switch count {
	case 0:
		p.protocol = ProtocolClosed
	case 1:
		if candidate.initiator && candidate.io == nil {
			p.protocol = WaitConnAck
		} else if candidate.initiator {
			p.protocol = WaitCEA
		} else {
			p.protocol = WaitCER
		}
	case 2:
		p.protocol = WaitReturns
		for _, c := range p.candidates {
			if c != nil && c.initiator && c.io == nil {
				p.protocol = WaitConnAckElect
			}
		}
	}
}
func (n *Node) retireConnectionLocked(c *connection) {
	p := c.peer
	if p.conn == c {
		p.conn = nil
	}
	if p.chosen == c {
		p.chosen = nil
	}
	for i, v := range p.candidates {
		if v == c {
			p.candidates[i] = nil
		}
	}
	p.refreshHandshakeLocked()
	close(c.done)
	p.signalLocked()
}
func (n *Node) Connect(ctx context.Context, id PeerID, conn transport.Conn) (*Peer, error) {
	return n.attach(ctx, id, conn, true)
}

// Accept binds an already established transport to a configured peer. An opposite
// direction candidate may compete during handshake; open connections are never replaced.
func (n *Node) Accept(ctx context.Context, id PeerID, conn transport.Conn) (*Peer, error) {
	return n.attach(ctx, id, conn, false)
}
func (n *Node) attach(ctx context.Context, id PeerID, conn transport.Conn, initiator bool) (*Peer, error) {
	if conn == nil {
		return nil, ErrConfig
	}
	c, err := n.admitConnection(ctx, id, conn, conn.Info(), initiator, nil, nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return n.handshakeConnection(ctx, c)
}

// A nil transport reserves the outbound candidate before a Dialer is called.
// The slot and Node task remain held until the callback and all I/O return.
func (n *Node) admitConnection(ctx context.Context, id PeerID, conn transport.Conn, info transport.Info, initiator bool, owner *maintenance, server *serving) (*connection, error) {
	var seed [4]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return nil, err
	}
	n.mu.Lock()
	p := n.peers[id]
	err := ctx.Err()
	slot := -1
	if err == nil {
		switch {
		case n.closed || n.shutting:
			err = ErrClosed
		case p == nil:
			err = ErrConfig
		case p.lifecycle != Active:
			err = ErrPeerClosed
		case !initiator && p.server != server:
			err = ErrServing
		case initiator && p.manager != owner:
			err = ErrManaged
		case p.manager != nil && (p.manager.stopping || p.manager.ctx.Err() != nil):
			err = ErrNotReady
		case p.conn != nil || p.chosen != nil:
			err = ErrNotReady
		case info.BufferBytes > n.cfg.Resources.AdapterBytesPerConnection:
			err = ErrBackpressure
		case n.connSeq == math.MaxUint64:
			err = ErrIdentifiers
		}
	}
	if err == nil {
		for i, c := range p.candidates {
			if c == nil {
				slot = i
			} else if c.initiator == initiator || c.closed {
				err = ErrNotReady
			}
		}
		if slot < 0 {
			err = ErrBackpressure
		}
	}
	if err != nil {
		n.mu.Unlock()
		return nil, err
	}
	deadline := time.Now().Add(n.cfg.Timeouts.Handshake)
	if d, ok := ctx.Deadline(); ok {
		deadline = earlier(deadline, d)
	}
	if server != nil {
		deadline = time.Time{}
	}
	empty := p.candidates[0] == nil && p.candidates[1] == nil
	if empty {
		p.handshakeDeadline = deadline
		p.cause = nil
		p.lastDPAResult = 0
		p.transport = TransportFacts{}
	} else if server == nil {
		deadline = earlier(deadline, p.handshakeDeadline)
	}
	life, cancel := context.WithCancel(context.Background())
	n.connSeq++
	c := &connection{accepting: server != nil, schedule: n.scheduler.newWriter(), peer: p, id: ConnectionID{n.boot, n.connSeq}, io: conn, ctx: life, cancel: cancel, initiator: initiator, handshakeDeadline: deadline, pending: make(map[uint32]*exchange), queue: make([]*exchange, 0, p.maxPending), controls: make([]controlJob, 0, 2), wake: make(chan struct{}, 1), done: make(chan struct{}), hbhSeed: binary.BigEndian.Uint32(seed[:]), jitter: binary.BigEndian.Uint32(seed[:]) | 1}
	c.answers = make([]*Inbound, 0, p.maxInbound+n.cfg.Resources.RejectSlots)
	p.candidates[slot] = c
	p.refreshHandshakeLocked()
	p.signalLocked()
	n.wg.Add(1)
	n.mu.Unlock()
	return c, nil
}
func (n *Node) handshakeConnection(ctx context.Context, c *connection) (*Peer, error) {
	hctx, hcancel := context.WithDeadline(c.ctx, c.handshakeDeadline)
	// The caller deadline is already included in hctx. Forward only explicit
	// cancellation, so racing deadline timers cannot relabel expiry as Canceled.
	stop := context.AfterFunc(ctx, func() {
		if ctx.Err() == context.Canceled {
			hcancel()
		}
	})
	apps, err := n.handshake(hctx, c, c.initiator)
	if hctx.Err() != nil {
		err = hctx.Err()
	}
	stop()
	hcancel()
	return n.finishConnection(ctx, c, apps, err)
}
func (n *Node) finishConnection(ctx context.Context, c *connection, apps *negotiatedApplications, err error) (*Peer, error) {
	p := c.peer
	n.mu.Lock()
	n.retireIdentifier(c.cerIdentifier)
	c.cerIdentifier = nil
	if c.closed {
		err = c.cause
	} else if ctx.Err() != nil {
		err = ctx.Err()
	} else if !c.handshakeDeadline.IsZero() && !time.Now().Before(c.handshakeDeadline) {
		// Socket deadline wakeup can precede context timer delivery. The fixed
		// admission deadline, not timer scheduling order, determines the result.
		err = context.DeadlineExceeded
	}
	if err == nil && (n.closed || n.shutting || p.lifecycle != Active) {
		err = ErrClosed
	}
	if err == nil && p.chosen != nil && p.chosen != c {
		err = ErrElectionLost
	}
	var losers []transport.Conn
	if err == nil {
		p.chosen = c
		p.conn = c
		p.transport = c.transport
		c.established = true
		c.apps = apps
		for _, other := range p.candidates {
			if other != nil && other != c && !other.closed {
				n.stopConnectionLocked(other, ErrElectionLost)
				if other.io != nil {
					losers = append(losers, other.io)
				}
			}
		}
		if c.initiator {
			p.protocol = IOpen
		} else {
			p.protocol = ROpen
		}
		if p.everOpen {
			p.watchdog = Reopen
			c.tw = time.Now()
		} else {
			p.watchdog = Okay
			c.okaySince = time.Now()
			c.tw = time.Now().Add(n.watchdogDelay(c))
		}
		if _, messages := c.io.(transport.MessageWriter); messages && !c.initiator {
			c.sendBarrierUntil = time.Now().Add(n.cfg.Timeouts.Handshake)
			// No I/O under the state gate; control encoding/writing stays in the writer.
			if e := n.queueProbeLocked(c, time.Now()); e != nil {
				n.stopConnectionLocked(c, e)
				n.mu.Unlock()
				_ = c.io.Close()
				for _, io := range losers {
					_ = io.Close()
				}
				if c.prefixDone != nil {
					<-c.prefixDone
				}
				n.mu.Lock()
				n.retireConnectionLocked(c)
				n.mu.Unlock()
				n.wg.Done()
				n.notify()
				return nil, e
			}
		}
		p.everOpen = true
		c.workers = 2
		n.wg.Add(2)
		p.signalLocked()
		n.mu.Unlock()
		for _, io := range losers {
			_ = io.Close()
		}
		go n.reader(c)
		go n.writer(c)
		n.wg.Done()
		n.notify()
		return p, nil
	}
	n.stopConnectionLocked(c, err)
	n.mu.Unlock()
	if c.io != nil {
		_ = c.io.Close()
	}
	if c.prefixDone != nil {
		<-c.prefixDone
	}
	n.mu.Lock()
	n.retireConnectionLocked(c)
	n.mu.Unlock()
	n.wg.Done()
	n.notify()
	return nil, err
}

// A passive loser waits for the outbound CEA or transport failure. One bounded
// prefix read detects EOF without consuming/allocating an application body. If
// promoted, the very same read is handed to the runtime reader, not canceled.
func (n *Node) selectPassive(ctx context.Context, c *connection) error {
	p := c.peer
	for {
		n.mu.Lock()
		if c.closed {
			err := c.cause
			n.mu.Unlock()
			return err
		}
		if n.closed || n.shutting || p.lifecycle != Active {
			n.mu.Unlock()
			return ErrClosed
		}
		if p.conn != nil || p.chosen != nil && p.chosen != c {
			n.mu.Unlock()
			return ErrElectionLost
		}
		var active *connection
		for _, other := range p.candidates {
			if other != nil && other != c && !other.closed {
				active = other
			}
		}
		if active == nil || strings.Compare(strings.ToLower(n.cfg.Origin.Host), strings.ToLower(p.origin.Host)) > 0 {
			p.chosen = c
			var loser transport.Conn
			if active != nil {
				n.stopConnectionLocked(active, ErrElectionLost)
				loser = active.io
			}
			p.signalLocked()
			n.mu.Unlock()
			if loser != nil {
				_ = loser.Close()
			}
			return nil
		}
		if c.prefixResult == nil {
			c.prefixResult = make(chan prefixResult, 1)
			c.prefixDone = make(chan struct{})
			n.wg.Add(1)
			go func() {
				defer n.wg.Done()
				defer close(c.prefixDone)
				prefix, err := n.readPrefix(c.ctx, c)
				c.prefixResult <- prefixResult{prefix, err}
			}()
		}
		changed := p.changed
		ch := c.prefixResult
		n.mu.Unlock()
		select {
		case <-changed:
		case r := <-ch:
			if r.err != nil {
				return r.err
			}
			return ErrProtocol
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
