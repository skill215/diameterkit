package diameter

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type RoRecoveryTrigger uint8

const (
	RoRecoveryInitial RoRecoveryTrigger = iota + 1
	RoRecoveryNoAnswer
	RoRecoveryErrorAnswer
)

// RoRecoveryPlan is one finite, explicit segment. TargetID selects the first
// target; Alternates contains later authorized terminals. Peers are paths to each target.
// Deadline can shorten the configured timeout and never extends ReplayUntil.
// ResponseWait optionally enables the queue-based Tr gate for every attempt.
// Zero retains transport-stop-only recovery. It is independent of business Tx.
type RoRecoveryPlan struct {
	Import                 *RoRecoveryImport   // First external admission only; keep identical for an active duplicate.
	Alternates             []RoRecoveryPath    // Ordered, distinct frozen targets after the primary target.
	ErrorContinuation      RoErrorContinuation // Explicit policy for answers received by this segment.
	ErrorContinuationUntil time.Time           // Required only for SMSFWBeforeTx; after this instant only transient codes qualify.
	ResponseWait           time.Duration       // Zero disables Tr; positive values must fit the frozen record limit.
	TargetID               string
	Peers                  []PeerID
	MaxAttempts            int
	Deadline               time.Time
	Trigger                RoRecoveryTrigger
}

// RoRecoveryProgress declares the application's already reserved send allowance.
// Unused allowance is not refunded. Delivery is cumulative previous-segment
// evidence, exactly as returned by RoRecoveryResult, not an override of T or HBH.
// No storage/ownership or cross-process anti-rollback guarantee is implied.
type RoRecoveryProgress struct {
	SegmentNo, ReservedBefore, SegmentAllowance, ReservedThrough uint32
	Delivery                                                     DeliveryEvidence
}
type RoRecoveryResult struct {
	ReplayPending              bool   // Possible write since the last authorized error answer.
	ErrorResultCode            uint32 // Eligible top-level base error; zero otherwise.
	RecordID                   [16]byte
	SegmentNo, ReservedThrough uint32
	CumulativeDelivery         DeliveryEvidence
	CanContinue                bool              // Local continuation eligibility, not a business decision.
	NextTrigger                RoRecoveryTrigger // Zero when terminal; derived from the settled local result.
}

var (
	ErrRoRecoveryNotPrepared    = errors.New("Ro recovery record not prepared by this Node")
	ErrRoRecoveryProgress       = errors.New("invalid or regressed Ro recovery progress")
	ErrRoRecoveryBusy           = errors.New("Ro recovery record has a preparing or active segment")
	ErrRoRecoveryConflict       = errors.New("Ro recovery segment differs from accepted segment")
	ErrRoSegmentAlreadyFinished = errors.New("Ro recovery segment already finished")
	ErrRoRecoveryFinal          = errors.New("Ro recovery record cannot continue")
)

type roSegmentState struct {
	targets                  []roPreparedTarget
	visited                  uint16
	errorUntil               time.Time
	errorContinuation        RoErrorContinuation
	replayPending            bool // Possible sends since the last authorized error answer; separate from cumulative history.
	responseWait             time.Duration
	slot                     int
	prior                    DeliveryEvidence
	maxSegments, maxAttempts uint32
}

func roPlanDigest(plan RoRecoveryPlan, p RoRecoveryProgress) [32]byte {
	w := roDigestWriter{h: sha256.New()}
	w.str("diameterkit/ro-local-segment/v5")
	w.imported(plan.Import)
	w.number(uint64(plan.ResponseWait))
	w.number(uint64(plan.ErrorContinuation))
	w.str(plan.TargetID)
	w.number(uint64(len(plan.Alternates)))
	for _, a := range plan.Alternates {
		w.str(a.TargetID)
		w.number(uint64(len(a.Peers)))
		for _, id := range a.Peers {
			w.str(string(id))
		}
	}
	w.number(uint64(plan.Trigger))
	w.number(uint64(plan.MaxAttempts))
	for _, deadline := range []time.Time{plan.Deadline, plan.ErrorContinuationUntil} {
		if deadline.IsZero() {
			w.number(0)
		} else {
			w.number(1)
			w.number(uint64(deadline.Unix()))
			w.number(uint64(deadline.Nanosecond()))
		}
	}
	w.number(uint64(len(plan.Peers)))
	for _, id := range plan.Peers {
		w.str(string(id))
	}
	for _, v := range []uint64{uint64(p.SegmentNo), uint64(p.ReservedBefore), uint64(p.SegmentAllowance), uint64(p.ReservedThrough), uint64(p.Delivery.Bytes), uint64(p.Delivery.FrameBytes)} {
		w.number(v)
	}
	if p.Delivery.Uncertain {
		w.number(1)
	} else {
		w.number(0)
	}
	if p.Delivery.MayHaveReachedPeer {
		w.number(1)
	} else {
		w.number(0)
	}
	var out [32]byte
	copy(out[:], w.h.Sum(nil))
	return out
}
func sameRoPolicy(a, b CommandPolicy) bool {
	if a.ApplicationBinding != b.ApplicationBinding || a.Command != b.Command || a.Class != b.Class || a.DestinationHost != b.DestinationHost || a.DestinationRealm != b.DestinationRealm || a.MaxRequestBytes != b.MaxRequestBytes || a.MaxAnswerBytes != b.MaxAnswerBytes || a.Timeout != b.Timeout || len(a.Sources.Terminals) != len(b.Sources.Terminals) || len(a.Sources.ErrorAgents) != len(b.Sources.ErrorAgents) {
		return false
	}
	for i, o := range a.Sources.Terminals {
		if o != b.Sources.Terminals[i] {
			return false
		}
	}
	for i, o := range a.Sources.ErrorAgents {
		if o != b.Sources.ErrorAgents[i] {
			return false
		}
	}
	return true
}

// roSegmentGate runs under the Node gate. An accepted duplicate is resolved
// before context/readiness/budget checks and never re-enqueues or extends time.
func (n *Node) roSegmentGate(r RoRecoveryRequest, plan RoRecoveryPlan, p RoRecoveryProgress, digest [32]byte, port *CompletionPort) (int, RequestID, error) {
	if n.closed || n.shutting || port.closed {
		return -1, RequestID{}, ErrClosed
	}
	e2e := binary.BigEndian.Uint32(r.Frame[16:20])
	if !n.roRecovery.contains(e2e, time.Now()) {
		return -1, RequestID{}, ErrRoRecoveryNotPrepared
	}
	slot := n.roRecovery.byE2E[e2e]
	g := &n.roRecovery.slots[slot]
	if g.recordID != r.RecordID || g.digest != r.Digest {
		return -1, RequestID{}, ErrRoRecoveryConflict
	}
	if p.SegmentNo == g.segmentNo && g.segmentNo != 0 {
		if g.active == (RequestID{}) {
			return -1, RequestID{}, ErrRoSegmentAlreadyFinished
		}
		x := n.requests[g.active]
		if digest != g.segmentDigest || x == nil || x.port != port {
			return -1, RequestID{}, ErrRoRecoveryConflict
		}
		return slot, g.active, nil
	}
	if p.SegmentNo < g.segmentNo {
		return -1, RequestID{}, ErrRoSegmentAlreadyFinished
	}
	if g.admitting || g.active != (RequestID{}) {
		return -1, RequestID{}, ErrRoRecoveryBusy
	}
	if !time.Now().Before(g.replayUntil) || time.Now().Before(r.CreatedAt) || !time.Now().Before(r.ReplayUntil) {
		return -1, RequestID{}, context.DeadlineExceeded
	}
	if g.segmentNo != 0 && !g.continuable {
		return -1, RequestID{}, ErrRoRecoveryFinal
	}
	if p.SegmentNo != g.segmentNo+1 || p.ReservedBefore < g.reservedThrough || p.Delivery != g.cumulative {
		return -1, RequestID{}, ErrRoRecoveryProgress
	}
	if g.segmentNo == 0 {
		if plan.Trigger != RoRecoveryInitial || p.ReservedBefore != 0 {
			return -1, RequestID{}, ErrRoRecoveryProgress
		}
	} else if plan.Trigger != g.nextTrigger {
		return -1, RequestID{}, ErrRoRecoveryProgress
	}
	return slot, RequestID{}, nil
}

// SubmitRoRecovery accepts a locally prepared original identity without invoking
// the identifier source. Frozen authorized targets and locally settled no-answer
// or explicitly enabled named error continuation are admitted. External records
// additionally require opt-in import checks. ctx controls admission only; accepted work uses its fixed deadline.
func (n *Node) SubmitRoRecovery(ctx context.Context, r RoRecoveryRequest, plan RoRecoveryPlan, p RoRecoveryProgress, port *CompletionPort) (RequestID, error) {
	var zero RequestID
	if port == nil || port.node != n {
		return zero, ErrConfig
	}
	if err := r.CheckIntegrity(); err != nil {
		return zero, err
	}
	if !withinRoLimits(r.Limits, n.cfg.RoRecovery) || r.WireFingerprint != n.cfg.Bundle.Fingerprint() {
		return zero, ErrRoRecoveryRecord
	}
	selectedIndex := roTargetIndex(r, plan.TargetID)
	if len(plan.Peers) < 1 || len(plan.Peers) > 64 || len(plan.TargetID) > 128 || selectedIndex < 0 || len(plan.Alternates) > 15 {
		return zero, ErrUnsupported
	}
	if plan.Trigger != RoRecoveryInitial && plan.Trigger != RoRecoveryNoAnswer && plan.Trigger != RoRecoveryErrorAnswer {
		return zero, ErrUnsupported
	}
	if plan.MaxAttempts < 1 || plan.MaxAttempts > r.Limits.MaxAttemptsPerSegment || p.SegmentNo == 0 || p.SegmentNo > uint32(r.Limits.MaxSegments) || p.SegmentAllowance == 0 || p.SegmentAllowance > uint32(r.Limits.MaxTotalAttempts) || p.ReservedBefore > uint32(r.Limits.MaxTotalAttempts) || p.ReservedThrough != p.ReservedBefore+p.SegmentAllowance || p.ReservedThrough > uint32(r.Limits.MaxTotalAttempts) || uint32(plan.MaxAttempts) > p.SegmentAllowance {
		return zero, ErrRoRecoveryProgress
	}
	if plan.ResponseWait < 0 || plan.ResponseWait > r.Limits.MaxResponseWait {
		return zero, ErrConfig
	}
	if !validRoErrorContinuation(r.PolicyID, plan.ErrorContinuation) || (plan.ErrorContinuation == RoErrorSMSFWBeforeTx) != !plan.ErrorContinuationUntil.IsZero() || plan.ErrorContinuationUntil.After(r.ReplayUntil) {
		return zero, ErrConfig
	}
	policy := r.Targets[selectedIndex].Policy
	op, ok := n.cfg.Bundle.Lookup(policy.Command)
	if !ok || !op.TransportReplay || op.ContextName != r.ContextName {
		return zero, ErrUnsupported
	}
	paths := append([]RoRecoveryPath{{TargetID: plan.TargetID, Peers: plan.Peers}}, plan.Alternates...)
	totalPaths := 0
	for i, path := range paths {
		index := roTargetIndex(r, path.TargetID)
		if index < 0 || len(path.Peers) == 0 || len(path.Peers) > 64 || !roCrossAllowed(r, index) {
			return zero, ErrUnsupported
		}
		totalPaths += len(path.Peers)
		if totalPaths > 64 {
			return zero, ErrConfig
		}
		for _, old := range paths[:i] {
			if old.TargetID == path.TargetID {
				return zero, ErrConfig
			}
		}
		for j, id := range path.Peers {
			if len(id) > 128 || !n.roPathAllowed(r.Targets[index], id) {
				return zero, ErrConfig
			}
			for _, old := range path.Peers[:j] {
				if old == id {
					return zero, ErrConfig
				}
			}
		}
	}
	maxRequest := 0
	maxAnswer := 0
	minTimeout := policy.Timeout
	for _, target := range r.Targets {
		found := false
		for _, peer := range n.peers {
			if n.roPathAllowed(target, peer.id) {
				found = true
				break
			}
		}
		if !found || target.Policy.MaxRequestBytes > n.cfg.Limits.MaxFrameBytes || target.Policy.MaxAnswerBytes > n.cfg.Limits.MaxFrameBytes {
			return zero, ErrRoRecoveryRecord
		}
		maxRequest = max(maxRequest, target.Policy.MaxRequestBytes)
		maxAnswer = max(maxAnswer, target.Policy.MaxAnswerBytes)
		minTimeout = min(minTimeout, target.Policy.Timeout)
	}
	qb, err := op.DecodeBound(maxRequest, n.cfg.Limits)
	if err != nil {
		return zero, err
	}
	ah := pdu.Header{ApplicationID: 4, CommandCode: 272, Flags: pdu.ProxiableFlag}
	ab, err := n.engine.DecodeBound(ah, maxAnswer)
	if err != nil {
		return zero, err
	}
	ah.Flags |= pdu.ErrorFlag
	eb, err := n.engine.DecodeBound(ah, maxAnswer)
	if err != nil {
		return zero, err
	}
	ab = max(ab, eb)
	extra := uint64(r.Limits.MaxRecordBytes)*2 + 131072
	factor := uint64(4 + 2*(len(paths)-1))
	if qb > (math.MaxUint64-extra)/factor {
		return zero, ErrBackpressure
	}
	q := factor*qb + extra
	unit, err := add(ab, 2048)
	if err != nil || unit > math.MaxUint64/uint64(plan.MaxAttempts) {
		return zero, ErrBackpressure
	}
	capacity, err := add(q, unit*uint64(plan.MaxAttempts))
	if err != nil {
		return zero, ErrBackpressure
	}
	peerBytes, err := add(q, unit)
	if err != nil {
		return zero, ErrBackpressure
	}
	digest := roPlanDigest(plan, p)
	committed := false
	n.mu.Lock()
	if plan.Import != nil {
		e2e := binary.BigEndian.Uint32(r.Frame[16:20])
		if n.roRecovery.contains(e2e, time.Now()) {
			old := n.roRecovery.slots[n.roRecovery.byE2E[e2e]]
			if old.active == (RequestID{}) {
				n.mu.Unlock()
				if old.segmentNo > 0 && !old.continuable {
					return zero, ErrRoRecoveryFinal
				}
				return zero, ErrRoRecoveryConflict
			}
		} else {
			imported, e := n.beginRoImport(r, *plan.Import, ctx)
			if e != nil {
				n.mu.Unlock()
				return zero, e
			}
			n.wg.Add(1)
			defer func() {
				n.mu.Lock()
				if !committed {
					n.roRecovery.release(imported)
				}
				n.mu.Unlock()
				n.wg.Done()
				n.notify()
			}()
		}
	}

	slot, existing, err := n.roSegmentGate(r, plan, p, digest, port)
	if err != nil || existing != zero {
		n.mu.Unlock()
		return existing, err
	}
	if err = ctx.Err(); err != nil {
		n.mu.Unlock()
		return zero, err
	}
	now := time.Now()
	deadline := earlier(now.Add(minTimeout), n.roRecovery.slots[slot].replayUntil)
	if !plan.Deadline.IsZero() {
		deadline = earlier(deadline, plan.Deadline)
	}
	if !now.Before(deadline) {
		n.mu.Unlock()
		return zero, context.DeadlineExceeded
	}
	charge, err := budget.Reserve(capacity, 1, "ro-segment-preparation", &n.total, n.classes[policy.Class], &port.account)
	if err != nil {
		n.mu.Unlock()
		return zero, ErrBackpressure
	}
	var peer *Peer
	var pc *budget.Reservation
	last := error(ErrNoEligibleAlternate)
	for _, id := range plan.Peers {
		candidate := n.peers[id]
		if !candidate.readyLocked() || !candidate.conn.apps.allows(policy.Command, policy.ApplicationBinding) {
			continue
		}
		pc, err = budget.Reserve(peerBytes, 1, "route-peer", &candidate.account)
		if err != nil {
			last = ErrBackpressure
			continue
		}
		peer = candidate
		break
	}
	if peer == nil {
		charge.Release()
		n.mu.Unlock()
		return zero, last
	}
	c := peer.conn
	n.roRecovery.slots[slot].admitting = true
	peer.preparing++
	n.wg.Add(1)
	n.mu.Unlock()
	defer func() {
		n.mu.Lock()
		n.roRecovery.slots[slot].admitting = false
		if !committed {
			charge.Release()
			pc.Release()
			port.signal()
		}
		peer.preparing--
		peer.signalLocked()
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
	}()
	// Owned selected frame/snapshot plus bounded all-target validation workspace.
	// Do not encode a decoded application object: preserve original private bytes.
	original, err := n.engine.Decode(r.Frame)
	if err != nil {
		return zero, err
	}
	env, err := op.Envelope(original.Body)
	if err != nil || env.Origin != n.cfg.Origin || !env.DestinationHost.Present || env.DestinationHost.Value != r.Targets[0].Policy.DestinationHost || !env.DestinationRealm.Present || env.DestinationRealm.Value != r.Targets[0].Policy.DestinationRealm {
		return zero, validation.ErrEnvelope
	}
	var frame []byte
	var snapshot *codecengine.RequestSnapshot
	largest := 0
	prepared := make([]roPreparedTarget, len(paths))
	for i, target := range r.Targets {
		f, e := roTargetFrame(r.Frame, pdu.Origin{Host: target.Policy.DestinationHost, Realm: target.Policy.DestinationRealm}, n.cfg.Limits, min(target.Policy.MaxRequestBytes, r.Limits.MaxRecordBytes-roRecordMetadata(r)))
		if e != nil {
			return zero, e
		}
		msg, e := n.engine.Decode(f)
		if e != nil {
			return zero, e
		}
		snap, e := n.engine.PrepareRequest(msg, target.Policy.Sources)
		if e != nil {
			return zero, e
		}
		if e = snap.BindContext(validation.Context{}); e != nil {
			return zero, e
		}
		for j, path := range paths {
			if path.TargetID != target.TargetID {
				continue
			}
			ownedPolicy := n.peers[path.Peers[0]].policies[policy.Command]
			if config, ok := n.roTargets[path.TargetID]; ok {
				ownedPolicy = config.binding.Policy
			}
			t := roPreparedTarget{index: i, id: avp.CloneString(path.TargetID), policy: ownedPolicy, frame: f, snapshot: snap}
			for _, id := range path.Peers {
				t.peers = append(t.peers, n.peers[id])
			}
			prepared[j] = t
		}
		largest = max(largest, len(f))
		if i == selectedIndex {
			frame = f
			snapshot = snap
		}
	}
	if largest != r.MaxFrameBytes {
		return zero, ErrRoRecoveryRecord
	}
	route := &Route{node: n, policy: prepared[0].policy, cfg: RouteConfig{Command: policy.Command, ExchangePolicy: ExplicitRetransmit, MaxAttempts: plan.MaxAttempts, MaxReplayAge: r.Limits.MaxReplayAge, ReplayDomain: ReplayDomain{Target: policy.Sources.Terminals[0], Retention: r.Targets[selectedIndex].Retention}}}
	for _, id := range plan.Peers {
		route.peers = append(route.peers, n.peers[id])
		route.cfg.Peers = append(route.cfg.Peers, PeerID(avp.CloneString(string(id))))
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err = ctx.Err(); err != nil {
		return zero, err
	}
	if n.closed || n.shutting || port.closed || peer.lifecycle != Active {
		return zero, ErrClosed
	}
	if !peer.readyLocked() || peer.conn != c {
		return zero, ErrNotReady
	}
	if !time.Now().Before(deadline) || time.Now().Before(r.CreatedAt) || !time.Now().Before(r.ReplayUntil) {
		return zero, context.DeadlineExceeded
	}
	if n.seq == math.MaxUint64 {
		return zero, ErrIdentifiers
	}
	hbh, err := c.nextHBH(false)
	if err != nil {
		return zero, err
	}
	guard := &n.roRecovery.slots[slot]
	snapshot = snapshot.ForAttempt(hbh, guard.replayPending, route.policy.Sources)
	wire.PutHeader(frame, snapshot.Header(), len(frame))
	n.seq++
	id := RequestID{n.boot, n.seq}
	x := &exchange{roTarget: selectedIndex, scheduleClass: n.scheduler.requestIndex[policy.Class], id: id, conn: c, port: port, charge: charge, snapshot: snapshot, frame: frame, deadline: deadline, answerMax: policy.MaxAnswerBytes, answerBound: ab, hbh: hbh, delivery: DeliveryEvidence{FrameBytes: len(frame)}, attemptID: AttemptID{id, 1}, attemptPeer: peer, attemptConnection: c.id, peerCharge: pc}
	g := &routeExchange{route: route, root: x, peerBytes: peerBytes, frame: frame, phase: AttemptActive, attempts: make([]*exchange, 0, plan.MaxAttempts), recovery: &roSegmentState{targets: prepared, visited: uint16(1) << selectedIndex, errorContinuation: plan.ErrorContinuation, errorUntil: plan.ErrorContinuationUntil, replayPending: guard.replayPending, responseWait: plan.ResponseWait, slot: slot, prior: guard.cumulative, maxSegments: uint32(r.Limits.MaxSegments), maxAttempts: uint32(r.Limits.MaxTotalAttempts)}}
	x.group = g
	startRoResponseWait(x, time.Now())
	g.attempts = append(g.attempts, x)
	// Only this commit consumes a segment and its entire reserved allowance.
	guard.segmentNo = p.SegmentNo
	guard.reservedThrough = p.ReservedThrough
	guard.segmentDigest = digest
	guard.active = id
	guard.continuable = false
	guard.nextTrigger = 0
	charge.Transfer("exchange")
	n.requests[id] = x
	c.pending[hbh] = x
	c.enqueueRequest(x)
	committed = true
	c.signal()
	n.notify()
	return id, nil
}

func (n *Node) finishRoSegmentLocked(g *routeExchange, result *Result) {
	if g.recovery == nil {
		return
	}
	guard := &n.roRecovery.slots[g.recovery.slot]
	d := result.Delivery
	prior := g.recovery.prior
	d.Bytes = max(d.Bytes, prior.Bytes)
	d.FrameBytes = max(d.FrameBytes, prior.FrameBytes)
	d.Uncertain = d.Uncertain || prior.Uncertain
	d.MayHaveReachedPeer = d.MayHaveReachedPeer || prior.MayHaveReachedPeer
	guard.cumulative = d
	guard.active = RequestID{}
	guard.replayPending = g.recovery.replayPending || result.Delivery.MayHaveReachedPeer
	guard.nextTrigger = 0
	if result.Answer.Body != nil && result.Err == nil {
		// The winning candidate's code came from the validated, associated frame,
		// never from mutable caller Result data or a child/experimental result.
		if g.winner != nil && g.winner.attemptPeer.lifecycle == Active && permitsRoError(g.recovery.errorContinuation, g.winner.roErrorCode) {
			guard.nextTrigger = RoRecoveryErrorAnswer
			guard.replayPending = false
		}
	} else if result.Answer.Body == nil && (errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, context.DeadlineExceeded) || result.StopReason == StopTransport || result.StopReason == StopWatchdog || result.StopReason == StopAttemptsExhausted || result.StopReason == StopResponseWait) {
		guard.nextTrigger = RoRecoveryNoAnswer
	}
	guard.continuable = !n.closed && !n.shutting && !g.root.port.closed && time.Now().Before(guard.replayUntil) && guard.segmentNo < g.recovery.maxSegments && guard.reservedThrough < g.recovery.maxAttempts && guard.nextTrigger != 0
	if !guard.continuable {
		guard.nextTrigger = 0
	}
	guard.keepUntil = maxTime(guard.keepUntil, time.Now().Add(n.identifiers.config.ReuseGuard))
	var code uint32
	if guard.nextTrigger == RoRecoveryErrorAnswer && g.winner != nil {
		code = g.winner.roErrorCode
	}
	result.RoRecovery = &RoRecoveryResult{ReplayPending: guard.replayPending, ErrorResultCode: code, RecordID: guard.recordID, SegmentNo: guard.segmentNo, ReservedThrough: guard.reservedThrough, CumulativeDelivery: d, CanContinue: guard.continuable, NextTrigger: guard.nextTrigger}
}
func maxTime(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}
