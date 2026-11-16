// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
)

func earlier(a, b time.Time) time.Time {
	if a.IsZero() || !b.IsZero() && b.Before(a) {
		return b
	}
	return a
}
func (p *Peer) quiescedLocked() bool {
	return p.candidates[0] == nil && p.candidates[1] == nil && p.preparing == 0 && p.manager == nil && p.server == nil
}
func (n *Node) shutdownLimit(ctx context.Context) time.Time {
	d := time.Now().Add(n.cfg.Timeouts.ShutdownTimeout)
	if t, ok := ctx.Deadline(); ok {
		d = earlier(d, t)
	}
	if ctx.Err() != nil {
		d = time.Now()
	}
	return d
}
func (n *Node) shutdownPeerLocked(p *Peer, d time.Time) {
	if p.lifecycle == Closed {
		return
	}
	p.lifecycle = Draining
	n.stopServingLocked(p, ErrShutdown)
	p.shutdownDeadline = earlier(p.shutdownDeadline, d)
	drain := p.shutdownDeadline.Add(-n.cfg.Timeouts.CloseReserve)
	if drain.Before(time.Now()) {
		drain = time.Now()
	}
	p.drainDeadline = earlier(p.drainDeadline, drain)
	if p.conn != nil && !p.conn.closingDeadline.IsZero() {
		p.conn.closingDeadline = earlier(p.conn.closingDeadline, p.shutdownDeadline)
	}
	p.signalLocked()
}

// Shutdown permanently stops new work in this peer's management domain. Its ctx
// bounds the whole drain/close operation; cancellation upgrades it to immediate close.
func (p *Peer) Shutdown(ctx context.Context) error {
	n := p.node
	fallback := n.shutdownLimit(ctx)
	n.mu.Lock()
	n.shutdownPeerLocked(p, fallback)
	n.mu.Unlock()
	n.tick(time.Now())
	n.notify()
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		n.mu.Lock()
		if p.lifecycle == Closed && p.quiescedLocked() {
			err := p.shutdownErr
			n.mu.Unlock()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		deadline := earlier(fallback, p.shutdownDeadline)
		ch := p.changed
		n.mu.Unlock()
		timer.Reset(max(time.Until(deadline), 0))
		select {
		case <-ch:
		case <-ctx.Done():
			n.mu.Lock()
			n.shutdownPeerLocked(p, time.Now())
			n.mu.Unlock()
			n.tick(time.Now())
			return ctx.Err()
		case <-timer.C:
			n.mu.Lock()
			n.shutdownPeerLocked(p, time.Now())
			n.mu.Unlock()
			n.tick(time.Now())
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return context.DeadlineExceeded
		}
	}
}

// Shutdown uses one absolute deadline for all peers. Expiry may return before
// a delayed borrower exits; Wait remains the cleanup barrier, independent of Ports.
func (n *Node) Shutdown(ctx context.Context) error {
	fallback := n.shutdownLimit(ctx)
	n.beginShutdown(fallback)
	n.tick(time.Now())
	n.notify()
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		n.mu.Lock()
		deadline := earlier(fallback, n.shutdownDeadline)
		ch := n.shutdownChanged
		err := n.shutdownErr
		stopped := n.stopped
		n.mu.Unlock()
		if stopped {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		timer.Reset(max(time.Until(deadline), 0))
		select {
		case <-n.done:
		case <-ch:
		case <-ctx.Done():
			n.beginShutdown(time.Now())
			n.tick(time.Now())
			n.notify()
			return ctx.Err()
		case <-timer.C:
			n.beginShutdown(time.Now())
			n.tick(time.Now())
			n.notify()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return context.DeadlineExceeded
		}
	}
}
func (n *Node) beginShutdown(d time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return
	}
	n.shutting = true
	for _, x := range n.requests {
		if x.group != nil && x.group.phase == AwaitingRetransmit {
			n.chooseLocked(x, Result{Err: ErrShutdown, StopReason: StopLocal})
		}
	}
	n.shutdownDeadline = earlier(n.shutdownDeadline, d)
	close(n.shutdownChanged)
	n.shutdownChanged = make(chan struct{})
	for _, p := range n.peers {
		n.shutdownPeerLocked(p, n.shutdownDeadline)
	}
}
func (p *Peer) mayStartAdmittedLocked(c *connection, now time.Time) bool {
	return p.conn == c && !c.closed && (p.protocol == IOpen || p.protocol == ROpen) && p.watchdog == Okay && (p.lifecycle == Active || p.lifecycle == Draining && now.Before(p.drainDeadline))
}
func (n *Node) enterClosingLocked(c *connection, d time.Time, cause error) {
	c.peer.protocol = Closing
	c.closingDeadline = earlier(c.closingDeadline, d)
	if c.closeCause == nil {
		c.closeCause = cause
	}
	c.tw = time.Time{}
	n.retireIdentifier(c.probeIdentifier)
	c.probeIdentifier = nil
	c.probe = nil
	c.probeQueued = false
	jobs := c.controls[:0]
	for _, job := range c.controls {
		if job.disconnect || job.closeAfter {
			job.deadline = earlier(job.deadline, c.closingDeadline)
			jobs = append(jobs, job)
		}
	}
	for i := len(jobs); i < len(c.controls); i++ {
		c.controls[i] = controlJob{}
	}
	c.controls = jobs
	if len(jobs) == 0 {
		c.schedule.category.empty(scheduleControl)
	}
	for _, x := range c.pending {
		if !x.writing && x.delivery.Bytes == 0 && !x.delivery.Uncertain {
			n.stopAttemptLocked(x, c.closeCause, StopLocal)
		}
	}
	n.suppressAnswersLocked(c, cause)
	c.signal()
}
func (n *Node) queueDPRLocked(c *connection, now time.Time) error {
	if c.localDPR || c.remoteDPR {
		return nil
	}
	h, err := c.nextHBH(true)
	var end uint32
	if err == nil {
		end, err = n.nextE2E(true)
	}
	if err != nil {
		return err
	}
	c.disconnectIdentifier = &identifierLease{control: true, slot: -1}
	c.localDPR = true
	m := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 282, HopByHop: h, EndToEnd: end}, Body: &base.DPR{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, DisconnectCause: 2}}
	c.enqueueControl(controlJob{identifier: c.disconnectIdentifier, request: m, disconnect: true, deadline: earlier(c.closingDeadline, now.Add(n.cfg.Timeouts.FrameWrite))})
	c.signal()
	return nil
}

// advanceShutdownLocked only produces close actions; callers execute I/O outside the gate.
func (n *Node) advanceShutdownLocked(now time.Time) []transport.Conn {
	var closeList []transport.Conn
	stop := func(c *connection, err error) {
		if !c.closed {
			n.stopConnectionLocked(c, err)
			if c.io != nil {
				closeList = append(closeList, c.io)
			}
		}
	}
	for _, p := range n.peers {
		if p.lifecycle != Draining {
			continue
		}
		if !now.Before(p.shutdownDeadline) {
			// A completed DPA/EOF must not become a timeout merely because cleanup was slow.
			live := false
			for _, c := range p.candidates {
				if c != nil && !c.closed {
					live = true
					stop(c, ErrShutdown)
				}
			}
			if live || p.preparing > 0 {
				p.shutdownErr = context.DeadlineExceeded
				if n.shutting {
					n.shutdownErr = context.DeadlineExceeded
				}
			}
		}
		c := p.conn
		if c == nil || c.closed {
			for _, candidate := range p.candidates {
				if candidate != nil {
					stop(candidate, ErrShutdown)
				}
			}
		} else if p.protocol != Closing {
			drained := len(c.pending) == 0 && len(c.queue) == 0 && len(c.controls) == 0 && !c.writerBusy && p.preparing == 0 && !n.activeInboundLocked(c)
			if drained || !now.Before(p.drainDeadline) {
				n.enterClosingLocked(c, p.shutdownDeadline, ErrShutdown)
				if err := n.queueDPRLocked(c, now); err != nil {
					stop(c, err)
				}
			}
		}
		if p.quiescedLocked() {
			p.lifecycle = Closed
			for len(p.inboundQueue) > 0 {
				n.endInboundLocked(p.inboundQueue[0], ErrShutdown)
			}
			p.signalLocked()
		}
	}
	if n.shutting && !n.closed {
		all := true
		for _, p := range n.peers {
			if p.lifecycle != Closed || !p.quiescedLocked() {
				all = false
				break
			}
		}
		if all {
			n.closeDuplicatesLocked()
			for _, in := range n.inbounds {
				n.endInboundLocked(in, ErrShutdown)
			}
			n.closed = true
		}
	}
	return closeList
}
