package diameter

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type RequestOptions struct {
	// TargetID selects an explicitly configured Peer.RequestTargets entry. Empty
	// uses the default command policy, or the Route's frozen target when submitted
	// through a Route. It never authorizes an arbitrary destination or retry.
	TargetID   string
	Deadline   time.Time
	Validation validation.Context
}
type exchange struct {
	roTarget                                int
	roErrorCode                             uint32 // Validated top-level base result of this answer candidate only.
	acceptedAt, responseWaitUntil           time.Time
	responseWaitExpired                     bool
	identifier                              *identifierLease
	scheduleClass                           int
	group                                   *routeExchange
	attemptID                               AttemptID
	attemptPeer                             *Peer
	attemptConnection                       ConnectionID
	peerCharge                              *budget.Reservation
	stopped                                 bool
	stopReason                              StopReason
	stopErr                                 error
	id                                      RequestID
	conn                                    *connection
	port                                    *CompletionPort
	charge                                  *budget.Reservation
	snapshot                                *codecengine.RequestSnapshot
	frame                                   []byte
	deadline                                time.Time
	answerMax                               int
	answerBound                             uint64
	hbh                                     uint32
	writing, receiving, selected, published bool
	result                                  Result
	delivery                                DeliveryEvidence
}

// Submit uses ctx only for admission. A successful request lives until its
// configured deadline, Cancel, a terminal answer or connection termination.
func (p *Peer) Submit(ctx context.Context, body pdu.Body, opts RequestOptions, port *CompletionPort) (RequestID, error) {
	return p.submit(ctx, body, opts, port, nil)
}
func (p *Peer) submit(ctx context.Context, body pdu.Body, opts RequestOptions, port *CompletionPort, route *Route) (RequestID, error) {
	n := p.node
	if body == nil || port == nil || port.node != n {
		return RequestID{}, ErrConfig
	}
	policy, ok := p.requestPolicy(body.CommandKey(), opts.TargetID)
	if !ok {
		if opts.TargetID != "" {
			return RequestID{}, ErrConfig
		}
		return RequestID{}, ErrUnsupported
	}
	opContext, _ := n.cfg.Bundle.Lookup(policy.Command)
	if !opContext.NoContext && opContext.CheckContext == nil {
		return RequestID{}, validation.ErrUnsupported
	}
	size, err := n.engine.MeasureBody(body)
	if err != nil {
		return RequestID{}, err
	}
	if size > policy.MaxRequestBytes {
		return RequestID{}, wire.ErrLimit
	}
	// Reserve worst-case completion/clone/decode before copying caller input.
	op, _ := n.cfg.Bundle.Lookup(policy.Command)
	qb, err := op.DecodeBound(policy.MaxRequestBytes, n.cfg.Limits)
	if err != nil {
		return RequestID{}, err
	}
	ah := pdu.Header{ApplicationID: policy.Command.ApplicationID, CommandCode: policy.Command.Code, Flags: pdu.ProxiableFlag}
	ab, err := n.engine.DecodeBound(ah, policy.MaxAnswerBytes)
	if err != nil {
		return RequestID{}, err
	}
	ah.Flags |= pdu.ErrorFlag
	eb, err := n.engine.DecodeBound(ah, policy.MaxAnswerBytes)
	if err != nil {
		return RequestID{}, err
	}
	ab = max(ab, eb)
	if qb > (math.MaxUint64-131072)/4 {
		return RequestID{}, ErrBackpressure
	}
	cap, err := add(4*qb+131072, ab)
	if err != nil {
		return RequestID{}, ErrBackpressure
	}
	peerBytes := cap
	if route != nil {
		if ab > math.MaxUint64/uint64(route.cfg.MaxAttempts)-2048 {
			return RequestID{}, ErrBackpressure
		}
		cap, err = add(4*qb+131072, uint64(route.cfg.MaxAttempts)*(ab+2048))
		if err != nil {
			return RequestID{}, ErrBackpressure
		}
	}
	now := time.Now()
	deadline := now.Add(policy.Timeout)
	if route != nil && route.cfg.ExchangePolicy != SingleAttempt {
		deadline = earlier(deadline, now.Add(route.cfg.MaxReplayAge))
	}
	if !opts.Deadline.IsZero() && opts.Deadline.Before(deadline) {
		deadline = opts.Deadline
	}
	n.mu.Lock()
	if n.closed || n.shutting || port.closed || p.lifecycle != Active {
		n.mu.Unlock()
		return RequestID{}, ErrClosed
	}
	if err = ctx.Err(); err != nil {
		n.mu.Unlock()
		return RequestID{}, err
	}
	if !deadline.After(now) {
		n.mu.Unlock()
		return RequestID{}, context.DeadlineExceeded
	}
	if !p.readyLocked() || !p.conn.apps.allows(policy.Command, policy.ApplicationBinding) {
		n.mu.Unlock()
		return RequestID{}, ErrNotReady
	}
	c := p.conn
	var charge, peerCharge *budget.Reservation
	if route == nil {
		charge, err = budget.Reserve(cap, 1, "preparing", &n.total, n.classes[policy.Class], &p.account, &port.account)
	} else {
		charge, err = budget.Reserve(cap, 1, "preparing", &n.total, n.classes[policy.Class], &port.account)
		if err == nil {
			peerCharge, err = budget.Reserve(peerBytes, 1, "route-peer", &p.account)
			if err != nil {
				charge.Release()
			}
		}
	}
	if err != nil {
		n.mu.Unlock()
		return RequestID{}, ErrBackpressure
	}
	hbh, err := c.nextHBH(false)
	var e2e uint32
	var identifier *identifierLease
	if err == nil {
		if n.identifiers.config.EndToEnd != nil {
			identifier, err = n.identifiers.reserve(policy.Class, time.Now())
		} else {
			e2e, err = n.nextE2E(false)
			if err == nil {
				identifier = &identifierLease{slot: -1}
			}
		}
	}
	if err != nil {
		n.retireIdentifier(identifier)
		charge.Release()
		if peerCharge != nil {
			peerCharge.Release()
		}
		port.signal()
		n.mu.Unlock()
		return RequestID{}, err
	}
	n.wg.Add(1)
	p.preparing++
	n.mu.Unlock()
	defer func() {
		n.mu.Lock()
		p.preparing--
		p.signalLocked()
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
	}()
	if n.identifiers.config.EndToEnd != nil {
		candidate, sourceErr := n.identifiers.config.EndToEnd.NextEndToEnd()
		n.mu.Lock()
		n.identifiers.sourceBusy = false
		if sourceErr != nil {
			err = errors.Join(ErrIdentifiers, ErrIdentifierSourceFailed, sourceErr)
		} else if ctx.Err() != nil {
			err = ctx.Err()
		} else if n.closed || n.shutting || p.lifecycle != Active || port.closed {
			err = ErrClosed
		} else {
			err = n.installApplicationIdentifier(identifier, candidate)
			e2e = candidate
		}
		if err != nil {
			n.retireIdentifier(identifier)
			charge.Release()
			if peerCharge != nil {
				peerCharge.Release()
			}
			port.signal()
			n.mu.Unlock()
			return RequestID{}, err
		}
		n.mu.Unlock()
	}
	m, err := n.engine.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, DestinationHost: policy.DestinationHost, DestinationRealm: policy.DestinationRealm, HopByHop: hbh, EndToEnd: e2e})
	var snapshot *codecengine.RequestSnapshot
	var frame []byte
	if err == nil {
		var env pdu.Envelope
		env, err = op.Envelope(m.Body)
		if err == nil && (!env.DestinationRealm.Present || env.DestinationRealm.Value != policy.DestinationRealm || (policy.DestinationHost == "" && env.DestinationHost.Present) || (policy.DestinationHost != "" && (!env.DestinationHost.Present || env.DestinationHost.Value != policy.DestinationHost))) {
			err = validation.ErrEnvelope
		}
	}
	if err == nil {
		var size int
		size, err = n.engine.Measure(m)
		if size > policy.MaxRequestBytes {
			err = wire.ErrLimit
		}
	}
	if err == nil {
		snapshot, err = n.engine.PrepareRequest(m, policy.Sources)
		if err == nil {
			err = snapshot.BindContext(opts.Validation)
		}
	}
	if err == nil {
		frame, err = n.engine.Append(nil, m)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err == nil {
		err = ctx.Err()
	}
	if err == nil && (n.closed || n.shutting || port.closed || p.lifecycle != Active) {
		err = ErrClosed
	}
	if err == nil && (!p.readyLocked() || p.conn != c) {
		err = ErrNotReady
	}
	if err == nil && !time.Now().Before(deadline) {
		err = context.DeadlineExceeded
	}
	if err == nil && n.seq == math.MaxUint64 {
		err = ErrIdentifiers
	}
	if err != nil {
		n.abandonIdentifier(identifier)
		charge.Release()
		if peerCharge != nil {
			peerCharge.Release()
		}
		port.signal()
		return RequestID{}, err
	}
	n.seq++
	id := RequestID{n.boot, n.seq}
	x := &exchange{identifier: identifier, scheduleClass: n.scheduler.requestIndex[policy.Class], id: id, conn: c, port: port, charge: charge, snapshot: snapshot, frame: frame, deadline: deadline, answerMax: policy.MaxAnswerBytes, answerBound: ab, hbh: hbh, delivery: DeliveryEvidence{FrameBytes: len(frame)}}
	if route != nil {
		g := &routeExchange{route: route, root: x, peerBytes: peerBytes, frame: frame, phase: AttemptActive, attempts: make([]*exchange, 0, route.cfg.MaxAttempts)}
		x.group = g
		x.attemptID = AttemptID{id, 1}
		x.attemptPeer = p
		x.attemptConnection = c.id
		x.peerCharge = peerCharge
		g.attempts = append(g.attempts, x)
	}
	charge.Transfer("exchange")
	n.requests[id] = x
	c.pending[hbh] = x
	c.enqueueRequest(x)
	c.signal()
	n.notify()
	return id, nil
}

// Request uses the same engine and a private single-completion port. Cancellation
// stops waiting only after the selected result and real writer evidence settle.
func (p *Peer) Request(ctx context.Context, body pdu.Body, opts RequestOptions) (Result, error) {
	n := p.node
	port, err := n.NewCompletionPort(CompletionLimits{1, n.total.LimitBytes})
	if err != nil {
		return Result{}, err
	}
	defer port.Close(context.Background())
	id, err := p.Submit(ctx, body, opts, port)
	if err != nil {
		return Result{}, err
	}
	completion, err := port.Receive(ctx)
	if err != nil {
		_ = n.Cancel(id)
		completion, err = port.Receive(context.Background())
	}
	if err != nil {
		return Result{}, err
	}
	return completion.Result, completion.Result.Err
}
func (n *Node) Cancel(id RequestID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	x := n.requests[id]
	if x == nil {
		return nil
	}
	n.chooseLocked(x, Result{Err: context.Canceled})
	n.notify()
	return nil
}
func (n *Node) chooseLocked(x *exchange, result Result) {
	if x.group != nil {
		n.chooseRouteLocked(x.group, x, result)
		return
	}
	if x.selected {
		return
	}
	x.selected = true
	x.result = result
	// Removing queued writes makes queue space bounded by outstanding slots even
	// when cancellation repeatedly wins before the writer is scheduled.
	x.conn.removeRequest(x)
	n.publishLocked(x)
}
func (n *Node) publishLocked(x *exchange) {
	if x.group != nil {
		n.advanceRouteLocked(x.group)
		return
	}
	if !x.selected || x.published || x.writing || x.receiving {
		return
	}
	x.published = true
	n.retireIdentifier(x.identifier)
	delete(n.requests, x.id)
	delete(x.conn.pending, x.hbh)
	r := x.result
	r.RequestID = x.id
	r.PeerID = x.conn.peer.id
	r.ConnectionID = x.conn.id
	r.Delivery = x.delivery
	switch {
	case r.Answer.Body != nil:
		r.Outcome = AnswerReceived
	case r.Delivery.Bytes == r.Delivery.FrameBytes && r.Delivery.FrameBytes > 0 && !r.Delivery.Uncertain:
		r.Outcome = WrittenNoAnswer
	case r.Delivery.Uncertain:
		r.Outcome = WriteUncertain
	case r.Delivery.Bytes > 0:
		r.Outcome = PartlyWritten
	default:
		r.Outcome = NotWritten
	}
	// No borrower remains. Keep a conservative selected-answer bound and result
	// metadata in the same reservation; detach only the retired peer dimension.
	x.frame = nil
	x.snapshot = nil
	x.charge.Detach(&x.conn.peer.account)
	keep := uint64(1024)
	if r.Answer.Body != nil {
		keep += x.answerBound
	}
	_ = x.charge.SetHeld(0)
	_ = x.charge.Shrink(keep)
	x.charge.Transfer("completion")
	port := x.port
	if port.closed {
		x.charge.Release()
	} else {
		idx := (port.head + port.count) % len(port.queue)
		port.queue[idx] = completionEntry{Completion{x.id, r}, x.charge}
		port.count++
	}
	port.signal()
	n.notify()
}
func mergeWrite(result transportWrite, frameBytes int) DeliveryEvidence {
	return DeliveryEvidence{Bytes: result.bytes, FrameBytes: frameBytes, Uncertain: result.uncertain, MayHaveReachedPeer: result.bytes > 0 || result.uncertain}
}

type transportWrite struct {
	bytes     int
	uncertain bool
}
