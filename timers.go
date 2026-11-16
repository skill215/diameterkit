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
)

// One timer per Node scans bounded active records. No per-request goroutines or
// accumulating timer callbacks survive completion/connection retirement.
func (n *Node) timerLoop() {
	defer n.wg.Done()
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		n.mu.Lock()
		now := time.Now()
		closeList := n.advanceShutdownLocked(now)
		for _, p := range n.peers {
			if c := p.conn; c != nil && !c.closed {
				if e := n.advanceIdentifierDrainLocked(c, now); e != nil {
					n.stopConnectionLocked(c, e)
					closeList = append(closeList, c.io)
				}
			}
		}
		if n.closed {
			n.mu.Unlock()
			for _, c := range closeList {
				_ = c.Close()
			}
			return
		}
		next := n.nextDeadlineLocked()
		n.mu.Unlock()
		for _, c := range closeList {
			_ = c.Close()
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		delay := time.Hour
		if !next.IsZero() {
			delay = max(time.Until(next), 0)
		}
		timer.Reset(delay)
		select {
		case <-n.wake:
		case now := <-timer.C:
			n.tick(now)
		}
	}
}
func (n *Node) nextDeadlineLocked() time.Time {
	var next time.Time
	earlier := func(t time.Time) {
		if !t.IsZero() && (next.IsZero() || t.Before(next)) {
			next = t
		}
	}
	for _, r := range n.duplicates {
		earlier(r.keepUntil)
	}
	for _, in := range n.inbounds {
		if !in.decided {
			earlier(in.info.ExecutionDeadline)
		} else if in.result.Committed && (in.frame != nil || in.replay != nil) && !in.writing && !in.released && in.result.PrimaryWrite.Err == nil {
			earlier(in.writeDeadline)
		}
	}
	for _, x := range n.requests {
		if !x.selected {
			earlier(x.deadline)
			if x.group != nil && x.group.recovery != nil {
				cur := x.group.attempts[len(x.group.attempts)-1]
				if !cur.stopped && !cur.responseWaitExpired {
					earlier(cur.responseWaitUntil)
				}
			}
		}
	}
	for _, p := range n.peers {
		if p.lifecycle == Draining {
			if p.conn != nil && !p.conn.closed && p.protocol != Closing {
				earlier(p.drainDeadline)
			}
			if time.Now().Before(p.shutdownDeadline) {
				earlier(p.shutdownDeadline)
			}
		}
		c := p.conn
		if c == nil || c.closed {
			continue
		}
		if !c.identifierDrain.IsZero() && c.peer.protocol != Closing {
			earlier(c.identifierDrain)
		}
		earlier(c.tw)
		earlier(c.sendBarrierUntil)
		earlier(c.closingDeadline)
		if c.probeQueued {
			earlier(c.probeDeadline)
		}
		for _, job := range c.controls {
			earlier(job.deadline)
		}
	}
	return next
}
func (n *Node) watchdogDelay(c *connection) time.Duration {
	// A bounded per-connection PRNG yields a fresh +/-2s jitter without blocking
	// entropy calls under the state gate. Its seed came from crypto/rand.
	c.jitter ^= c.jitter << 13
	c.jitter ^= c.jitter >> 17
	c.jitter ^= c.jitter << 5
	return n.cfg.Timeouts.Watchdog + time.Duration(int64(c.jitter)%4000000001-2000000000)
}
func (n *Node) tick(now time.Time) {
	n.mu.Lock()
	closeList := n.advanceShutdownLocked(now)
	stop := func(c *connection, err error) { n.stopConnectionLocked(c, err); closeList = append(closeList, c.io) }
	for _, r := range n.duplicates {
		if !now.Before(r.keepUntil) {
			n.retireDuplicateLocked(r)
		}
	}
	for _, in := range n.inbounds {
		if !in.decided && !now.Before(in.info.ExecutionDeadline) {
			n.expireInboundLocked(in)
		} else if in.result.Committed && (in.frame != nil || in.replay != nil) && !in.writing && !now.Before(in.writeDeadline) {
			n.endInboundLocked(in, context.DeadlineExceeded)
		}
	}
	for _, x := range n.requests {
		if !x.selected && !now.Before(x.deadline) {
			n.chooseLocked(x, Result{Err: context.DeadlineExceeded})
		} else if !x.selected && x.group != nil && x.group.recovery != nil {
			n.expireRoResponseWaitLocked(x.group.attempts[len(x.group.attempts)-1], now)
		}
	}
	for _, p := range n.peers {
		c := p.conn
		if c == nil || c.closed {
			continue
		}
		if !c.sendBarrierUntil.IsZero() && !now.Before(c.sendBarrierUntil) {
			stop(c, context.DeadlineExceeded)
			continue
		}
		if err := n.advanceIdentifierDrainLocked(c, now); err != nil {
			stop(c, err)
			continue
		}
		if !c.closingDeadline.IsZero() && !now.Before(c.closingDeadline) {
			if c.localDPR && c.disconnect != nil {
				p.shutdownErr = context.DeadlineExceeded
				if n.shutting {
					n.shutdownErr = context.DeadlineExceeded
				}
			}
			stop(c, c.closeCause)
			continue
		}
		expired := false
		for _, job := range c.controls {
			if !now.Before(job.deadline) {
				expired = true
			}
		}
		if expired || c.probeQueued && !now.Before(c.probeDeadline) {
			stop(c, context.DeadlineExceeded)
			continue
		}
		if c.tw.IsZero() || now.Before(c.tw) {
			continue
		}
		switch p.watchdog {
		case Suspect:
			stop(c, ErrWatchdog)
		case Okay, Reopen:
			if c.probe != nil {
				if p.watchdog == Okay {
					p.watchdog = Suspect
					c.okaySince = time.Time{}
					p.signalLocked()
					for _, x := range c.pending {
						n.stopAttemptLocked(x, ErrWatchdog, StopWatchdog)
					}
				} else {
					if c.numDWA < 0 {
						stop(c, ErrWatchdog)
						continue
					}
					c.numDWA = -1
				}
				c.tw = now.Add(n.watchdogDelay(c))
				continue
			}
			if err := n.queueProbeLocked(c, now); err != nil {
				stop(c, err)
			}

		}
	}
	n.mu.Unlock()
	for _, c := range closeList {
		_ = c.Close()
	}
}

// queueProbeLocked shares the fixed control slot with the CEA send barrier.
func (n *Node) queueProbeLocked(c *connection, now time.Time) error {
	if len(c.controls) >= 2 {
		return ErrBackpressure
	}
	h, err := c.nextHBH(true)
	var end uint32
	if err == nil {
		end, err = n.nextE2E(true)
	}
	if err != nil {
		return err
	}
	c.probeIdentifier = &identifierLease{control: true, slot: -1}
	c.probeQueued = true
	c.probeDeadline = earlier(now.Add(n.cfg.Timeouts.FrameWrite), c.sendBarrierUntil)
	c.tw = time.Time{}
	m := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: h, EndToEnd: end}, Body: &base.DWR{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm}}
	c.enqueueControl(controlJob{identifier: c.probeIdentifier, request: m, probe: true, deadline: c.probeDeadline})
	c.signal()
	return nil
}
