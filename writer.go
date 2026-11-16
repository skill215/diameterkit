// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"io"
	"time"

	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
)

func (n *Node) writer(c *connection) {
	defer n.workerDone(c)
	for {
		n.mu.Lock()
		if c.closed {
			n.mu.Unlock()
			return
		}
		// Prepare only the control head outside the state gate. The control
		// reserve and writer borrower stay attached even if Closing removes it.
		if len(c.controls) > 0 && c.controls[0].prepared == nil {
			job := c.controls[0]
			prepared := &preparedControl{}
			c.controls[0].prepared = prepared
			if job.identifier != nil {
				job.identifier.writing = true
			}
			c.writerBusy = true
			n.mu.Unlock()
			frame, snapshot, err := n.prepareControl(c, job)
			n.mu.Lock()
			if job.identifier != nil {
				job.identifier.writing = false
				if job.identifier.done || err != nil {
					n.retireIdentifier(job.identifier)
				}
			}
			c.writerBusy = false
			current := !c.closed && len(c.controls) > 0 && c.controls[0].prepared == prepared
			if current && err == nil {
				prepared.frame = frame
				prepared.snapshot = snapshot
			}
			n.mu.Unlock()
			n.notify()
			if current && err != nil {
				n.fail(c, err)
				return
			}
			continue
		}
		kind, class, cost := c.pickWrite()
		if kind < 0 {
			n.mu.Unlock()
			select {
			case <-c.wake:
			case <-c.ctx.Done():
			}
			continue
		}
		if kind == scheduleControl {
			c.debitWrite(kind, class, cost)
			job := c.popControl()
			if job.identifier != nil {
				job.identifier.writing = true
			}
			c.writerBusy = true
			n.mu.Unlock()
			err := n.runControl(c, job)
			n.mu.Lock()
			if job.identifier != nil {
				job.identifier.writing = false
				if job.identifier.done || err != nil {
					n.retireIdentifier(job.identifier)
				}
			}
			c.writerBusy = false
			n.mu.Unlock()
			n.notify()
			if err != nil {
				n.fail(c, err)
				return
			}
			continue
		}
		if kind == scheduleAnswer {
			var in *Inbound
			for _, v := range c.answers {
				if v.scheduleClass == class {
					in = v
					break
				}
			}
			if in.result.PrimaryWrite.Err != nil {
				c.removeAnswer(in)
				n.releaseInboundLocked(in)
				n.mu.Unlock()
				continue
			}
			if !n.mayWriteAnswerLocked(in, time.Now()) || !time.Now().Before(in.writeDeadline) {
				c.removeAnswer(in)
				in.result.PrimaryWrite.Err = ErrNotReady
				if !time.Now().Before(in.writeDeadline) {
					in.result.PrimaryWrite.Err = context.DeadlineExceeded
				}
				n.releaseInboundLocked(in)
				n.mu.Unlock()
				continue
			}
			c.debitWrite(kind, class, cost)
			c.removeAnswer(in)
			in.writing = true
			c.writerBusy = true
			n.mu.Unlock()
			n.writeInbound(c, in)
			continue
		}
		var x *exchange
		for _, v := range c.queue {
			if v.scheduleClass == class {
				x = v
				break
			}
		}
		if x.selected || x.stopped {
			c.removeRequest(x)
			n.publishLocked(x)
			n.mu.Unlock()
			continue
		}
		if !c.peer.mayStartAdmittedLocked(c, time.Now()) {
			n.chooseLocked(x, Result{Err: ErrNotReady})
			n.mu.Unlock()
			continue
		}
		if !time.Now().Before(x.deadline) {
			n.chooseLocked(x, Result{Err: context.DeadlineExceeded})
			n.mu.Unlock()
			continue
		}
		// Check the queue-based response deadline even if the Node timer has not run.
		if n.expireRoResponseWaitLocked(x, time.Now()) {
			n.mu.Unlock()
			continue
		}
		// BeginWrite: cancellation after this point cannot claim the frame was never
		// offered. The exchange reservation remains attached until this call returns.
		c.debitWrite(kind, class, cost)
		c.removeRequest(x)
		x.writing = true
		c.writerBusy = true
		n.mu.Unlock()
		ctx, cancel := context.WithTimeout(c.ctx, n.cfg.Timeouts.FrameWrite)
		r := n.writeApplication(ctx, c, x.frame)
		cancel()
		if r.Bytes < 0 || r.Bytes > len(x.frame) {
			r.Uncertain = true
			r.Err = ErrProtocol
			r.Bytes = 0
		}
		if r.Err == nil && (r.Bytes != len(x.frame) || r.Uncertain) {
			r.Err = io.ErrShortWrite
		}
		n.mu.Lock()
		x.delivery = mergeWrite(transportWrite{r.Bytes, r.Uncertain}, len(x.frame))
		x.writing = false
		c.writerBusy = false
		if r.Err != nil {
			reason := ioStopReason(r.Err)
			cause := r.Err
			if c.closeCause != nil {
				reason = StopLocal
				cause = c.closeCause
			}
			n.stopAttemptLocked(x, cause, reason)
		} else {
			n.expireRoResponseWaitLocked(x, time.Now())
		}
		n.publishLocked(x)
		n.mu.Unlock()
		n.notify()
		if r.Err != nil {
			n.failTransport(c, r.Err)
			return
		}
	}
}

type preparedControl struct {
	frame    []byte
	snapshot *codecengine.RequestSnapshot
}

func (n *Node) prepareControl(c *connection, job controlJob) ([]byte, *codecengine.RequestSnapshot, error) {
	m := job.request
	var err error
	var snapshot *codecengine.RequestSnapshot
	if job.answer != nil {
		m, err = n.control.CompleteAnswer(job.request, job.answer, n.cfg.Origin)
	}
	if err != nil {
		return nil, nil, err
	}
	frame, err := n.control.Append(nil, m)
	if err != nil {
		return nil, nil, err
	}
	if job.probe || job.disconnect {
		snapshot, err = n.control.PrepareRequest(m, pdu.AnswerSources{Terminals: []pdu.Origin{c.peer.origin}})
		if err != nil {
			return nil, nil, err
		}
	}
	return frame, snapshot, nil
}
func (n *Node) runControl(c *connection, job controlJob) error {
	frame, snapshot := job.prepared.frame, job.prepared.snapshot

	n.mu.Lock()
	if c.closed {
		n.mu.Unlock()
		return ErrClosed
	}
	job.deadline = earlier(job.deadline, c.closingDeadline)
	if !time.Now().Before(job.deadline) {
		n.mu.Unlock()
		return context.DeadlineExceeded
	}
	if c.peer.protocol == Closing && !job.closeAfter && !job.disconnect {
		n.mu.Unlock()
		return nil
	}
	if job.disconnect {
		c.disconnect = snapshot
	}
	if job.probe {
		c.probeQueued = false
		c.probe = snapshot
		c.tw = time.Now().Add(n.watchdogDelay(c))
	}
	n.mu.Unlock()
	n.notify()
	ctx, cancel := context.WithDeadline(c.ctx, job.deadline)
	r := c.io.WriteFrame(ctx, frame)
	cancel()
	if r.Err != nil {
		return transportFailure{r.Err}
	}
	if r.Bytes != len(frame) || r.Uncertain {
		return transportFailure{io.ErrShortWrite}
	}
	if job.closeAfter || job.disconnect {
		n.mu.Lock()
		wait := n.cfg.Timeouts.ClosingGrace
		if job.disconnect {
			wait = n.cfg.Timeouts.DPAWait
		}
		d := time.Now().Add(wait)
		if !c.closed && d.Before(c.closingDeadline) {
			c.closingDeadline = d
		}
		n.mu.Unlock()
		n.notify()
	}
	return nil
}
