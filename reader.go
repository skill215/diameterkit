// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func (n *Node) reader(c *connection) {
	defer n.workerDone(c)
	for {
		var prefix [20]byte
		var err error
		if c.prefixResult != nil {
			r := <-c.prefixResult
			prefix, err = r.prefix, r.err
			c.prefixResult = nil
		} else {
			prefix, err = n.readPrefix(c.ctx, c)
		}
		if err != nil {
			n.failTransport(c, err)
			return
		}
		h, size, err := wire.Header(prefix[:], n.cfg.Limits)
		if err != nil {
			n.mu.Lock()
			if x := c.pending[h.HopByHop]; x != nil && !x.selected && x.snapshot.MatchesHeader(h) {
				n.chooseLocked(x, Result{Err: errors.Join(validation.ErrInvalidAnswer, err)})
			}
			n.mu.Unlock()
			n.fail(c, err)
			return
		}
		if h.ApplicationID == 0 && !(c.peer.scap1Accounting && h.CommandCode == 271) {
			if h.Flags&pdu.RequestFlag != 0 && h.CommandCode != 257 && h.CommandCode != 280 && h.CommandCode != 282 && c.peer.unknownRequests == BaseErrorFailures {
				err = n.readEmergency(c, h, size, rejectCommand, nil)
			} else {
				err = n.readControl(c, h, size)
			}
			if err != nil {
				n.fail(c, err)
				return
			}
			continue
		}

		if h.Flags&pdu.RequestFlag != 0 {
			n.mu.Lock()
			drop := !c.identifierDrain.IsZero() || c.peer.lifecycle == Draining || c.peer.protocol == Closing || c.peer.watchdog == Reopen
			n.mu.Unlock()
			if drop {
				if size > n.cfg.Resources.MaxDiscardBytes {
					n.fail(c, wire.ErrLimit)
					return
				}
				err = n.discardFrame(c.ctx, c, size)
				if err != nil {
					err = transportFailure{err}
				}
			} else {
				err = n.readInbound(c, h, size)
			}
			if err != nil {
				n.fail(c, err)
				return
			}
			continue
		}
		n.mu.Lock()
		x := c.pending[h.HopByHop]
		match := x != nil && !x.selected && x.snapshot.MatchesHeader(h)
		if c.closed {
			n.mu.Unlock()
			return
		}
		if !match {
			dropOnly := c.peer.protocol == Closing || c.peer.watchdog == Reopen || c.peer.lifecycle == Draining
			n.mu.Unlock()
			if h.Flags&pdu.RequestFlag != 0 && !dropOnly {
				n.fail(c, ErrUnsupported)
				return
			}
			if size > n.cfg.Resources.MaxDiscardBytes {
				n.fail(c, wire.ErrLimit)
				return
			}
			if err = n.discardFrame(c.ctx, c, size); err != nil {
				n.failTransport(c, err)
				return
			}
			continue
		}
		if size > x.answerMax {
			n.chooseLocked(x, Result{Err: errors.Join(validation.ErrInvalidAnswer, wire.ErrLimit), StopReason: StopProtocol})
			n.mu.Unlock()
			n.fail(c, wire.ErrLimit)
			return
		}
		// One reader owns the attempt's existing A until decode/candidate retirement.
		x.receiving = true
		n.mu.Unlock()
		b := make([]byte, size)
		err = n.readFrameInto(c.ctx, c, b)
		complete := err == nil
		var answer pdu.Message
		var report validation.Report
		var roCode uint32
		if err == nil {
			answer, report, err = x.snapshot.DecodeAnswerOwnedReport(b)
			if err == nil && x.group != nil && x.group.recovery != nil {
				roCode = roValidatedBaseResult(b, n.cfg.Limits)
				if x.group.recovery.errorContinuation == RoErrorSMSFWBeforeTx && !time.Now().Before(x.group.recovery.errorUntil) && !permitsRoError(RoErrorTransient, roCode) {
					roCode = 0
				}
			}
		}
		n.mu.Lock()
		x.receiving = false
		if err == nil {
			n.aliveLocked(c, false)
			x.roErrorCode = roCode
			n.chooseLocked(x, Result{Answer: answer, Validation: report})
		} else if !complete {
			reason := ioStopReason(err)
			cause := err
			if c.closeCause != nil {
				reason = StopLocal
				cause = c.closeCause
			}
			n.stopAttemptLocked(x, cause, reason)
		} else {
			n.chooseLocked(x, Result{Err: err, StopReason: StopProtocol, Validation: report})
		}
		n.publishLocked(x)
		n.mu.Unlock()
		n.notify()
		// Fully consumed but invalid answer completes only its exchange. A frame
		// read failure loses the stream boundary and must terminate the connection.
		if !complete {
			n.failTransport(c, err)
			return
		}
		var de *wire.DecodeError
		if err != nil && !errors.Is(err, validation.ErrInvalidAnswer) && !errors.As(err, &de) {
			n.fail(c, err)
			return
		}
	}
}
func (n *Node) readControl(c *connection, h pdu.Header, size int) error {
	if h.CommandCode != 280 && h.CommandCode != 282 {
		return ErrProtocol
	}
	n.mu.Lock()
	candidate := c.probe
	if h.CommandCode == 282 {
		candidate = c.disconnect
	}
	discard := h.Flags&pdu.RequestFlag == 0 && (candidate == nil || !candidate.MatchesHeader(h)) || c.peer.protocol == Closing && h.Flags&pdu.RequestFlag != 0 && h.CommandCode == 280
	n.mu.Unlock()
	if discard {
		if size > n.cfg.Resources.MaxDiscardBytes {
			return wire.ErrLimit
		}
		if err := n.discardFrame(c.ctx, c, size); err != nil {
			return transportFailure{err}
		}
		return nil
	}
	b := make([]byte, size)
	if err := n.readFrameInto(c.ctx, c, b); err != nil {
		return transportFailure{err}
	}
	m, err := n.control.DecodeOwned(b)
	if err != nil {
		return err
	}
	switch v := m.Body.(type) {
	case *base.DWR:
		if !sameOrigin(pdu.Origin{Host: v.OriginHost, Realm: v.OriginRealm}, c.peer.origin) {
			return ErrProtocol
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		if c.closed {
			return nil
		}
		if c.peer.protocol == Closing {
			return nil
		}
		n.aliveLocked(c, false)
		if c.closed {
			return nil
		}
		if len(c.controls) >= 2 {
			return ErrBackpressure
		}
		c.enqueueControl(controlJob{request: m, answer: &base.DWA{ResultCode: 2001}, deadline: time.Now().Add(n.cfg.Timeouts.FrameWrite)})
		c.signal()
		n.notify()
		return nil
	case *base.DPR:
		if !sameOrigin(pdu.Origin{Host: v.OriginHost, Realm: v.OriginRealm}, c.peer.origin) {
			return ErrProtocol
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		if c.closed {
			return nil
		}
		if c.remoteDPR {
			return nil
		}
		c.remoteDPR = true
		c.disconnectCause = v.DisconnectCause
		if v.DisconnectCause != 0 && c.peer.manager != nil {
			c.peer.manager.suppressed = true
			c.peer.manager.cause = ErrRemoteDPR
			c.peer.manager.next = time.Time{}
		}
		n.enterClosingLocked(c, time.Now().Add(n.cfg.Timeouts.CloseReserve), ErrRemoteDPR)
		c.enqueueControl(controlJob{request: m, answer: &base.DPA{ResultCode: 2001}, closeAfter: true, deadline: earlier(c.closingDeadline, time.Now().Add(n.cfg.Timeouts.FrameWrite))})
		c.signal()
		n.notify()
		return nil
	default:
		n.mu.Lock()
		pending := c.probe
		if h.CommandCode == 282 {
			pending = c.disconnect
		}
		n.mu.Unlock()
		if pending == nil || !pending.MatchesHeader(h) {
			return nil
		}
		if err = pending.CheckAnswer(m); err != nil {
			return err
		}
		n.mu.Lock()
		closeNow := false
		if !c.closed {
			if h.CommandCode == 282 && c.disconnect == pending {
				switch v := m.Body.(type) {
				case *base.DPA:
					c.peer.lastDPAResult = v.ResultCode
				case *pdu.ErrorAnswer:
					c.peer.lastDPAResult = v.ResultCode
				}
				n.stopConnectionLocked(c, c.closeCause)
				closeNow = true
			} else if h.CommandCode == 280 && c.probe == pending {
				n.retireIdentifier(c.probeIdentifier)
				c.probeIdentifier = nil
				c.probe = nil
				n.aliveLocked(c, true)
			}
		}
		n.mu.Unlock()
		if closeNow {
			_ = c.io.Close()
		}
		n.notify()
		return nil
	}
}
func (n *Node) aliveLocked(c *connection, matchedDWA bool) {
	if c.closed || c.peer.protocol == Closing {
		return
	}
	// The absolute barrier wins even if the Node timer has not run yet.
	if !c.sendBarrierUntil.IsZero() && !time.Now().Before(c.sendBarrierUntil) {
		n.stopConnectionLocked(c, context.DeadlineExceeded)
		return
	}
	c.sendBarrierUntil = time.Time{}
	p := c.peer
	if p.watchdog == Reopen {
		if matchedDWA {
			c.numDWA++
			if c.numDWA >= 3 {
				p.watchdog = Okay
				c.okaySince = time.Now()
				p.signalLocked()
				c.tw = time.Now().Add(n.watchdogDelay(c))
			}
		}
		return
	}
	if p.watchdog == Suspect {
		p.watchdog = Okay
		c.okaySince = time.Now()
		p.signalLocked()
	}
	if !c.probeQueued {
		c.tw = time.Now().Add(n.watchdogDelay(c))
	}
}
