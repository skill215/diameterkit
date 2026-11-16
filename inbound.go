// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"io"
	"math"
	"strings"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

var (
	ErrInboundDone  = errors.New("inbound already decided")
	ErrResponseBusy = errors.New("inbound response preparation in progress")
)

type inboundPolicy struct {
	InboundPolicy
	reserve, prepareBound, cacheBound uint64
}
type InboundInfo struct {
	RequestID         RequestID
	PeerID            PeerID
	ConnectionID      ConnectionID
	ExecutionDeadline time.Time
}
type ResponseWriteState uint8

const (
	ResponseNotWritten ResponseWriteState = iota
	ResponsePartlyWritten
	ResponseWritten
	ResponseWriteUncertain
)

type ResponseWrite struct {
	State    ResponseWriteState
	Delivery DeliveryEvidence
	Err      error
}
type ResponseResult struct {
	Committed    bool
	PrimaryWrite ResponseWrite
}

// Inbound owns one logical submission right and one original arrival. Message is
// the caller's independent mutable view; the private response envelope is never
// derived from that view. With DuplicatePolicy, one logical submission serves
// bounded arrivals and a volatile result cache; no business session is maintained.
type Inbound struct {
	featureContext                              validation.Context
	validation                                  validation.Report
	scheduleClass                               int
	emergency                                   bool
	node                                        *Node
	peer                                        *Peer
	conn                                        *connection
	info                                        InboundInfo
	policy                                      inboundPolicy
	request                                     pdu.Message
	view                                        pdu.Message
	frame                                       []byte
	timeoutFrame                                []byte
	timeoutResult                               []byte
	proxy                                       []byte
	arrivalHeader                               pdu.Header
	duplicate, replay                           *duplicateRecord
	charge                                      *budget.Reservation
	delivered, busy, writing, decided, released bool
	terminal                                    error
	result                                      ResponseResult
	writeDeadline                               time.Time
	done, finished                              chan struct{}
}

func (in *Inbound) Info() InboundInfo     { return in.info }
func (in *Inbound) Message() pdu.Message  { return in.view }
func (in *Inbound) Done() <-chan struct{} { return in.done }

func (n *Node) configureInbound(p *Peer, pc PeerConfig) error {
	p.maxInbound = pc.MaxInbound
	p.inboundPolicies = make(map[pdu.CommandKey]inboundPolicy)
	if len(pc.InboundCommands) == 0 {
		if pc.MaxInbound != 0 || pc.MaxInboundBytes != 0 {
			return ErrConfig
		}
		return nil
	}
	if len(pc.InboundCommands) > 256 || pc.MaxInbound <= 0 || pc.MaxInboundBytes == 0 {
		return ErrConfig
	}
	p.inboundAccount = budget.Account{LimitBytes: pc.MaxInboundBytes, LimitSlots: uint64(pc.MaxInbound)}
	p.inboundQueue = make([]*Inbound, 0, pc.MaxInbound)
	for _, cfg := range pc.InboundCommands {
		if cfg.FailurePolicy > BaseErrorFailures || cfg.FailurePolicy == BaseErrorFailures && n.cfg.Resources.RejectSlots == 0 {
			return ErrConfig
		}
		if !cfg.Command.Request || !validApplicationBinding(cfg.Command, cfg.ApplicationBinding, n.cfg.Capabilities) || n.inboundClasses[cfg.Class] == nil || cfg.ExecutionTimeout <= 0 || cfg.ExecutionTimeout > n.cfg.Timeouts.Request || len(cfg.Sources) == 0 || len(cfg.Sources) > 64 {
			return ErrConfig
		}
		if cfg.ApplicationBinding == SCAP1Accounting {
			p.scap1Accounting = true
		}
		if _, found := p.inboundPolicies[cfg.Command]; found {
			return ErrConfig
		}
		op, ok := n.cfg.Bundle.Lookup(cfg.Command)
		if !ok || !op.Proxiable || (!op.NoContext && op.CheckContext == nil) {
			return ErrConfig
		}
		ah := pdu.Header{ApplicationID: cfg.Command.ApplicationID, CommandCode: cfg.Command.Code, Flags: pdu.ProxiableFlag}
		rb, err := op.DecodeBound(cfg.MaxRequestBytes, n.cfg.Limits)
		if err != nil {
			return ErrConfig
		}
		ab, err := n.engine.DecodeBound(ah, cfg.MaxAnswerBytes)
		if err != nil {
			return ErrConfig
		}
		ah.Flags |= pdu.ErrorFlag
		eb, err := n.engine.DecodeBound(ah, cfg.MaxAnswerBytes)
		if err != nil {
			return ErrConfig
		}
		ab = max(ab, eb)
		if cfg.MaxRequestBytes < 20 || cfg.MaxRequestBytes%4 != 0 || cfg.MaxRequestBytes > n.cfg.Limits.MaxFrameBytes || rb > (math.MaxUint64-131072)/8 || ab > (math.MaxUint64-131072-8*rb)/8 {
			return ErrConfig
		}
		// Covers wire+owned request, private evidence, mutable delivery view, answer
		// completion clones, encode frame and bounded bookkeeping; no second charge.
		reserve := 8*(rb+ab) + 131072
		cfg.Sources = append([]pdu.Origin(nil), cfg.Sources...)
		for i, v := range cfg.Sources {
			if avp.Identity(v.Host) != nil || avp.Identity(v.Realm) != nil {
				return ErrConfig
			}
			cfg.Sources[i] = pdu.Origin{Host: avp.CloneString(v.Host), Realm: avp.CloneString(v.Realm)}
		}
		cacheBound, err := n.configureDuplicate(cfg, ab)
		if err != nil {
			return err
		}
		if !cfg.NoCache {
			// CanonicalV1 may retain child output while composing each ancestor.
			scratch := uint64(cfg.Duplicate.MaxCanonicalBytes)*uint64(2*n.cfg.Limits.MaxDepth+4) + uint64(min(n.cfg.Limits.MaxAVPs, (cfg.MaxRequestBytes-20)/8))*128
			reserve, err = add(reserve, scratch)
			if err != nil {
				return err
			}
		}
		p.inboundPolicies[cfg.Command] = inboundPolicy{cfg, reserve, 2*(rb+ab) + 1024, cacheBound}
	}
	return nil
}

// Receive waits across transport disconnection. Its ctx only cancels the wait.
// Explicit peer close stops delivery; Node.Close terminates all logical handles.
func (p *Peer) Receive(ctx context.Context) (*Inbound, error) {
	n := p.node
	for {
		n.mu.Lock()
		if err := ctx.Err(); err != nil {
			n.mu.Unlock()
			return nil, err
		}
		if n.closed || p.lifecycle == Closed {
			n.mu.Unlock()
			return nil, ErrClosed
		}
		if len(p.inboundQueue) > 0 {
			in := p.inboundQueue[0]
			if !time.Now().Before(in.info.ExecutionDeadline) {
				n.expireInboundLocked(in)
				n.mu.Unlock()
				continue
			}
			p.inboundQueue = removeInbound(p.inboundQueue, in)
			in.delivered = true
			n.mu.Unlock()
			return in, nil
		}
		ch := p.changed
		n.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
func removeInbound(queue []*Inbound, in *Inbound) []*Inbound {
	for i, v := range queue {
		if v == in {
			copy(queue[i:], queue[i+1:])
			queue[len(queue)-1] = nil
			return queue[:len(queue)-1]
		}
	}
	return queue
}
func (n *Node) readInbound(c *connection, h pdu.Header, size int) error {
	p := c.peer
	policy, ok := p.inboundPolicies[h.Key()]
	n.mu.Lock()
	binding := policy.ApplicationBinding
	if !ok && p.scap1Accounting && h.ApplicationID == 0 && h.CommandCode == 271 {
		binding = SCAP1Accounting
	}
	hasApp := c.apps.allows(h.Key(), binding)
	n.mu.Unlock()
	if !ok || !hasApp {
		if p.unknownRequests == BaseErrorFailures || ok && policy.FailurePolicy == BaseErrorFailures {
			reason := rejectCommand
			if !hasApp {
				reason = rejectApplication
			}
			return n.readEmergency(c, h, size, reason, nil)
		}
		return ErrUnsupported
	}
	if size > policy.MaxRequestBytes {
		return wire.ErrLimit
	}
	n.mu.Lock()
	if n.closed || n.shutting || !p.readyLocked() || p.conn != c || !c.apps.allows(h.Key(), policy.ApplicationBinding) {
		n.mu.Unlock()
		return ErrNotReady
	}
	if n.seq == math.MaxUint64 {
		n.mu.Unlock()
		return ErrIdentifiers
	}
	charge, err := budget.Reserve(policy.reserve, 1, "inbound", &n.total, n.inboundClasses[policy.Class], &p.inboundAccount)
	if err != nil {
		n.mu.Unlock()
		if policy.FailurePolicy == BaseErrorFailures {
			return n.readEmergency(c, h, size, rejectOverload, &policy)
		}
		return ErrBackpressure
	}
	n.seq++
	id := RequestID{n.boot, n.seq}
	in := &Inbound{node: n, peer: p, conn: c, policy: policy, charge: charge, busy: true, info: InboundInfo{id, p.id, c.id, time.Now().Add(policy.ExecutionTimeout)}, done: make(chan struct{}), finished: make(chan struct{})}
	n.inbounds[id] = in
	n.mu.Unlock()
	n.notify()
	b := make([]byte, size)
	ctx, cancel := context.WithDeadline(c.ctx, in.info.ExecutionDeadline)
	err = n.readFrameInto(ctx, c, b)
	cancel()
	var request, view pdu.Message
	var report validation.Report
	var reject, timeout, canonical, timeoutResult, proxy []byte
	var origin string
	complete := err == nil
	if err == nil {
		request, err = n.engine.DecodeOwned(b)
	}
	reason := rejectMissing
	if err == nil {
		err = checkInboundBinding(request, policy, n)
		if errors.Is(err, validation.ErrSource) {
			reason = rejectSource
		} else if err != nil {
			reason = rejectDestination
		}
	}
	if complete && err != nil && policy.FailurePolicy == BaseErrorFailures {
		reject, err = n.rejectFrame(b, policy.MaxAnswerBytes, reason, err)
	} else if err == nil {
		op, _ := n.cfg.Bundle.Lookup(h.Key())
		if policy.FailurePolicy == BaseErrorFailures {
			timeout, err = n.rejectFrame(b, policy.MaxAnswerBytes, rejectDeadline, nil)
		}
		if err == nil && !policy.NoCache {
			canonical, err = op.Canonical(b, n.cfg.Limits, policy.Duplicate.MaxCanonicalBytes)
			if err == nil && (len(canonical) < 12 || len(canonical) > policy.Duplicate.MaxCanonicalBytes) {
				err = wire.ErrLimit
			}
			if err == nil {
				timeoutResult, proxy, err = splitResult(timeout, n.cfg.Limits)
			}
			if err == nil && (len(proxy) > policy.Duplicate.MaxEnvelopeBytes || len(timeoutResult) > policy.Duplicate.MaxResultBytes) {
				err = wire.ErrLimit
			}
			if err == nil {
				var env pdu.Envelope
				env, err = op.Envelope(request.Body)
				origin = env.Origin.Host
			}
		}
		if err == nil {
			report = n.engine.ValidateContext(request, validation.Context{})
			if report.State == validation.Rejected {
				err = report.Err()
			}
		}
		if err == nil {
			report.CheckedRules = append(report.CheckedRules, "INBOUND-BINDING")
			view.Header = request.Header
			view.Body, err = op.Clone(request.Body)
		}
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	in.busy = false
	if err == nil && (in.decided || n.closed || p.lifecycle == Closed || c.closed) {
		err = ErrClosed
	}
	if err == nil && !time.Now().Before(in.info.ExecutionDeadline) {
		err = context.DeadlineExceeded
	}
	if err != nil {
		n.endInboundLocked(in, err)
		n.releaseInboundLocked(in)
		if !complete {
			return transportFailure{err}
		}
		return err
	}
	if len(reject) > 0 {
		n.queueAutomaticLocked(in, reject, ErrInboundDone)
		return nil
	}
	in.timeoutFrame = timeout
	in.request = request
	in.validation = report
	in.view = view
	in.arrivalHeader = request.Header
	in.proxy = proxy
	in.timeoutResult = timeoutResult
	if !policy.NoCache {
		err = n.attachDuplicateLocked(in, canonical, origin)
		if err != nil {
			// Conflict is a local protocol failure, never another business delivery.
			// The prevalidated 5012 frame uses only this arrival's envelope.
			if errors.Is(err, ErrDuplicateConflict) {
				n.queueAutomaticLocked(in, timeout, err)
				return nil
			}
			n.endInboundLocked(in, err)
			return err
		}
		if in.duplicate == nil || in.duplicate.owner != in {
			n.aliveLocked(c, false)
			return nil
		}
	}
	p.inboundQueue = append(p.inboundQueue, in)
	p.signalLocked()
	n.aliveLocked(c, false)
	n.notify()
	return nil
}

// Respond validates and bounds a complete answer before acquiring the one-time
// submission right. Failed validation or pre-commit ctx cancellation permits a
// retry. After Committed, ctx no longer cancels the write: return waits for its
// bounded primary write/close barrier. Committed is not peer/business success.
func (in *Inbound) Respond(ctx context.Context, body pdu.Body) (ResponseResult, error) {
	return in.respond(ctx, body, 0)
}
func (in *Inbound) respond(ctx context.Context, body pdu.Body, failure ProtocolFailure) (ResponseResult, error) {
	n := in.node
	n.mu.Lock()
	if in.decided {
		err := in.terminal
		if err == nil {
			err = ErrInboundDone
		}
		n.mu.Unlock()
		return ResponseResult{}, err
	}
	if err := ctx.Err(); err != nil {
		n.mu.Unlock()
		return ResponseResult{}, err
	}
	if !time.Now().Before(in.info.ExecutionDeadline) {
		n.expireInboundLocked(in)
		n.mu.Unlock()
		return ResponseResult{}, context.DeadlineExceeded
	}
	if failure == 0 && !in.validation.Complete {
		n.mu.Unlock()
		return ResponseResult{}, validation.ErrContextRequired
	}
	if in.busy {
		n.mu.Unlock()
		return ResponseResult{}, ErrResponseBusy
	}
	in.busy = true
	in.peer.preparing++
	n.wg.Add(1)
	request := in.request
	n.mu.Unlock()
	var frame, cached []byte
	var err error
	if failure != 0 {
		body, err = publicFailureBody(request, n, failure)
	}
	var bound uint64
	if err == nil {
		bound, err = n.engine.CompleteAnswerBound(request, body, n.cfg.Origin)
	}
	if err == nil && bound > in.policy.prepareBound {
		err = wire.ErrLimit
	}
	if err == nil {
		var answer pdu.Message
		answer, err = n.engine.CompleteAnswer(request, body, n.cfg.Origin)
		if err == nil && failure == 0 {
			r := n.engine.ValidateResponseContext(request, answer, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, in.featureContext)
			if !r.Complete {
				err = r.Err()
			}
		}
		if err == nil {
			var size int
			size, err = n.engine.Measure(answer)
			if err == nil && size > in.policy.MaxAnswerBytes {
				err = wire.ErrLimit
			}
			if err == nil {
				frame, err = n.engine.Append(nil, answer)
			}
		}
	}
	if err == nil && in.duplicate != nil {
		cached, _, err = splitResult(frame, n.cfg.Limits)
		if err == nil && len(cached) > in.policy.Duplicate.MaxResultBytes {
			err = wire.ErrLimit
		}
	}
	n.mu.Lock()
	in.busy = false
	in.peer.preparing--
	in.peer.signalLocked()
	if err == nil {
		err = ctx.Err()
	}
	if in.decided {
		err = in.terminal
		if err == nil {
			err = ErrInboundDone
		}
	}
	if !in.decided && !time.Now().Before(in.info.ExecutionDeadline) {
		n.expireInboundLocked(in)
		err = context.DeadlineExceeded
	}
	if err != nil {
		n.releaseInboundLocked(in)
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
		return ResponseResult{}, err
	}
	n.commitDuplicateLocked(in, cached, frame[4], false)
	in.decided = true
	in.result.Committed = true
	in.result.PrimaryWrite.Delivery.FrameBytes = len(frame)
	close(in.done)
	in.frame = frame
	in.timeoutFrame = nil
	in.writeDeadline = time.Now().Add(n.cfg.Timeouts.FrameWrite)
	if !n.mayWriteAnswerLocked(in, time.Now()) {
		if in.result.PrimaryWrite.Err == nil {
			in.result.PrimaryWrite.Err = ErrNotReady
		}
		n.releaseInboundLocked(in)
	} else {
		in.conn.enqueueAnswer(in)
		in.conn.signal()
	}
	n.mu.Unlock()
	n.wg.Done()
	n.notify()
	<-in.finished
	return in.result, in.result.PrimaryWrite.Err
}
func (n *Node) mayWriteAnswerLocked(in *Inbound, now time.Time) bool {
	c := in.conn
	p := in.peer
	return c != nil && !c.closed && p.conn == c && p.lifecycle != Closed && (p.protocol == IOpen || p.protocol == ROpen) && (p.watchdog == Okay || p.watchdog == Suspect) && (p.lifecycle != Draining || now.Before(p.drainDeadline))
}
func (n *Node) releaseInboundLocked(in *Inbound) {
	if in.released || !in.decided || in.busy || in.writing {
		return
	}
	// A committed queued response still owns its writer slot until it is written
	// or explicitly suppressed. Callers settle a queued write before releasing it.
	if in.result.Committed && (in.frame != nil || in.replay != nil) && in.result.PrimaryWrite.Err == nil {
		return
	}
	in.released = true
	delete(n.inbounds, in.info.RequestID)
	in.peer.inboundQueue = removeInbound(in.peer.inboundQueue, in)
	if in.conn != nil {
		in.conn.removeAnswer(in)
	}
	in.request = pdu.Message{}
	in.timeoutFrame = nil
	in.frame = nil
	in.conn = nil
	if !in.delivered {
		in.view = pdu.Message{}
	}
	n.releaseDuplicateArrivalLocked(in)
	in.charge.Release()
	close(in.finished)
	in.peer.signalLocked()
	n.notify()
}
func (n *Node) endInboundLocked(in *Inbound, err error) {
	if !in.decided {
		n.commitDuplicateLocked(in, nil, 0, true)
		in.decided = true
		in.terminal = err
		close(in.done)
		in.peer.inboundQueue = removeInbound(in.peer.inboundQueue, in)
	}
	if in.result.PrimaryWrite.Err == nil {
		in.result.PrimaryWrite.Err = err
	}
	n.releaseInboundLocked(in)
}
func (n *Node) detachInboundConnectionLocked(c *connection, err error) {
	for _, in := range n.inbounds {
		if in.conn == c {
			if in.result.PrimaryWrite.Err == nil {
				in.result.PrimaryWrite.Err = err
			}
			if !in.writing {
				in.conn = nil
			}
			if in.duplicate != nil && in.duplicate.owner != in && !in.decided {
				n.endInboundLocked(in, err)
			}
			if in.result.Committed {
				n.releaseInboundLocked(in)
			}
		}
	}
	c.clearAnswers()
}
func (n *Node) activeInboundLocked(c *connection) bool {
	for _, in := range n.inbounds {
		if in.conn == c {
			return true
		}
	}
	return false
}
func (n *Node) suppressAnswersLocked(c *connection, err error) {
	for len(c.answers) > 0 {
		in := c.answers[0]
		c.removeAnswer(in)
		in.result.PrimaryWrite.Err = err
		n.releaseInboundLocked(in)
	}
}
func (n *Node) writeInbound(c *connection, in *Inbound) {
	// The writer has already set writing under the state gate; Close may select
	// failure but cannot release the borrowed frame until this call returns.
	frame := in.frame
	var assemblyErr error
	if in.replay != nil {
		r := in.replay
		frame = make([]byte, 20+len(r.result)+len(in.proxy))
		h := in.arrivalHeader
		h.Flags = r.flags
		wire.PutHeader(frame, h, len(frame))
		copy(frame[20:], r.result)
		copy(frame[20+len(r.result):], in.proxy)
		var assembled pdu.Message
		assembled, assemblyErr = n.engine.DecodeOwned(frame)
		if assemblyErr == nil {
			assemblyErr = n.engine.Validate(assembled)
		}
		// Assembly borrows W/C but is not permission to start I/O after Closing.
		n.mu.Lock()
		if assemblyErr == nil && !n.mayWriteAnswerLocked(in, time.Now()) {
			assemblyErr = ErrNotReady
		}
		if assemblyErr == nil && !time.Now().Before(in.writeDeadline) {
			assemblyErr = context.DeadlineExceeded
		}
		n.mu.Unlock()
	}
	ctx, cancel := context.WithDeadline(c.ctx, in.writeDeadline)
	r := transport.WriteResult{Err: assemblyErr}
	if assemblyErr == nil {
		r = n.writeApplication(ctx, c, frame)
	}
	cancel()
	if r.Bytes < 0 || r.Bytes > len(frame) {
		r.Bytes = 0
		r.Uncertain = true
		r.Err = ErrProtocol
	}
	if r.Err == nil && (r.Bytes != len(frame) || r.Uncertain) {
		r.Err = io.ErrShortWrite
	}
	n.mu.Lock()
	in.writing = false
	c.writerBusy = false
	delivery := mergeWrite(transportWrite{r.Bytes, r.Uncertain}, len(frame))
	state := ResponseNotWritten
	switch {
	case r.Uncertain:
		state = ResponseWriteUncertain
	case r.Bytes == len(frame):
		state = ResponseWritten
	case r.Bytes > 0:
		state = ResponsePartlyWritten
	}
	// Actual completed I/O is authoritative even if Close raced its return.
	in.result.PrimaryWrite = ResponseWrite{state, delivery, r.Err}
	in.frame = nil
	in.replay = nil
	n.releaseInboundLocked(in)
	n.mu.Unlock()
	n.notify()
	if r.Err != nil {
		n.fail(c, r.Err)
	}
}

func checkInboundBinding(request pdu.Message, policy inboundPolicy, n *Node) error {
	op, _ := n.cfg.Bundle.Lookup(request.Header.Key())
	env, err := op.Envelope(request.Body)
	if err != nil {
		return err
	}
	allowed := false
	for _, origin := range policy.Sources {
		if sameOrigin(env.Origin, origin) {
			allowed = true
		}
	}
	if !allowed {
		return validation.ErrSource
	}
	if !env.DestinationRealm.Present || !strings.EqualFold(env.DestinationRealm.Value, n.cfg.Origin.Realm) || env.DestinationHost.Present && !strings.EqualFold(env.DestinationHost.Value, n.cfg.Origin.Host) {
		return validation.ErrEnvelope
	}
	return nil
}

// Validation returns owned evidence for the immutable received request. A
// committed response does not undo validation; expiry/close prevents further use.
func (in *Inbound) Validation() validation.Report {
	return in.Check(validation.Context{})
}

// Check freezes successful named evidence on the private request. It borrows the
// admitted workspace while running compiled pure rules outside the state lock.
// Failed checks leave the previous verdict intact. Frozen evidence cannot change.
func (in *Inbound) Check(ctx validation.Context) validation.Report {
	n := in.node
	n.mu.Lock()
	r := in.validation.Clone()
	expired := func() bool {
		return !in.result.Committed && (in.decided || !time.Now().Before(in.info.ExecutionDeadline) || n.closed)
	}
	if expired() {
		r.Complete, r.State = false, validation.Expired
		r.MissingContext = nil
		r.Violations = nil
		r.LocalErrors = []error{validation.ErrExpired}
		n.mu.Unlock()
		return r
	}
	if ctx.Empty() {
		n.mu.Unlock()
		return r
	}
	if in.result.Committed || !in.featureContext.Empty() {
		if ctx != in.featureContext {
			r.Failure(validation.ErrContext)
		}
		n.mu.Unlock()
		return r
	}
	if in.busy {
		r.Failure(validation.ErrContext)
		n.mu.Unlock()
		return r
	}
	_, _, host, realm, present := ctx.Endpoints()
	if present && (!strings.EqualFold(host, n.cfg.Origin.Host) || !strings.EqualFold(realm, n.cfg.Origin.Realm)) {
		r.Failure(validation.ErrContext)
		n.mu.Unlock()
		return r
	}
	in.busy = true
	in.peer.preparing++
	n.wg.Add(1)
	request := in.request
	n.mu.Unlock()
	r = n.engine.ValidateContext(request, ctx)
	if r.Complete {
		r.CheckedRules = append(r.CheckedRules, "INBOUND-BINDING")
	}
	n.mu.Lock()
	in.busy = false
	in.peer.preparing--
	in.peer.signalLocked()
	if expired() {
		r.Complete, r.State = false, validation.Expired
		r.MissingContext = nil
		r.Violations = nil
		r.LocalErrors = []error{validation.ErrExpired}
	} else if r.Complete {
		in.featureContext = ctx
		in.validation = r.Clone()
	}
	n.releaseInboundLocked(in)
	n.mu.Unlock()
	n.wg.Done()
	n.notify()
	return r
}
