package diameter

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// ExchangePolicy selects the sole scheduler for one logical request.
// Its zero value is invalid. Peer.Submit always remains SingleAttempt.
type ExchangePolicy uint8

const (
	SingleAttempt ExchangePolicy = iota + 1
	TransportFailover
	ExplicitRetransmit
)

type StopReason uint8

const (
	StopNone StopReason = iota
	StopTransport
	StopWatchdog
	StopProtocol
	StopLocal
	StopNoAlternate
	StopAttemptsExhausted
	StopBackpressure
	StopResponseWait
)

// ReplayReason describes the current attempt; it grants no admission by itself.
type ReplayReason uint8

const (
	ReplayNone ReplayReason = iota
	ReplayTransportStopped
	ReplayWatchdogStopped
	ResponseWaitExpired
)

type RequestPhase uint8

const (
	AttemptActive RequestPhase = iota + 1
	AwaitingRetransmit
	Completing
)

type AttemptID struct {
	RequestID RequestID
	Number    uint32
}

// AttemptResult is an owned, bounded summary; it does not retain a connection.
type AttemptResult struct {
	TargetID                      string // Ro recovery target; empty for ordinary routes.
	AcceptedAt, ResponseWaitUntil time.Time
	ID                            AttemptID
	PeerID                        PeerID
	ConnectionID                  ConnectionID
	HopByHop                      uint32
	Retransmitted                 bool
	Delivery                      DeliveryEvidence
	StopReason                    StopReason
	Err                           error
}
type RequestInfo struct {
	TargetID string // Current Ro recovery target.
	// Ro recovery only: segment queue acceptance, current attempt acceptance and
	// its Tr deadline (zero when disabled). Ordinary Route requests leave these zero.
	AcceptedAt, AttemptAcceptedAt, ResponseWaitUntil time.Time
	// Expiry can be observed while a writer is still retiring. Only ReplayEligible
	// permits Retransmit; a reason alone never authorizes another attempt.
	ReplayReason        ReplayReason
	RequestID           RequestID
	Phase               RequestPhase
	CurrentAttempt      AttemptID
	ReplayEligible      bool
	AttemptsUsed        int
	Deadline            time.Time
	Delivery            DeliveryEvidence
	TransportStopReason StopReason
}

var (
	ErrNoEligibleAlternate  = errors.New("no ready unused route connection")
	ErrAttemptsExhausted    = errors.New("route attempts exhausted")
	ErrRequestNotActive     = errors.New("request not active")
	ErrForeignRequest       = errors.New("request belongs to another node")
	ErrAttemptNotReplayable = errors.New("attempt not eligible for retransmission")
	ErrStaleAttempt         = errors.New("stale retransmission predecessor")
	ErrModeRequiresSubmit   = errors.New("explicit retransmission requires Submit")
)

// ReplayDomain is a deployment declaration, not proof of remote deduplication.
// This slice accepts exactly the configured Destination-Host/Realm and bounds
// the exchange age by the declared terminal retention. No target migration.
type ReplayDomain struct {
	Target    pdu.Origin
	Retention time.Duration
}
type RouteConfig struct {
	// TargetID selects the same authorized terminal on every candidate Peer.
	// It is immutable for the Route and all attempts; empty keeps default binding.
	TargetID       string
	Command        pdu.CommandKey
	Peers          []PeerID // Ordered finite candidates; all selections stay within this set.
	ExchangePolicy ExchangePolicy
	MaxAttempts    int
	MaxReplayAge   time.Duration
	ReplayDomain   ReplayDomain
}
type Route struct {
	node   *Node
	cfg    RouteConfig
	peers  []*Peer
	policy CommandPolicy
}

// NewRoute freezes ordered candidates and validates their command bindings.
// Multi-attempt modes require explicit same-terminal replay authority.
func (n *Node) NewRoute(cfg RouteConfig) (*Route, error) {
	if len(cfg.TargetID) > 128 || len(cfg.Peers) < 1 || len(cfg.Peers) > 64 || !cfg.Command.Request || cfg.Command.ApplicationID == 0 && cfg.Command.Code != 271 || cfg.ExchangePolicy < SingleAttempt || cfg.ExchangePolicy > ExplicitRetransmit {
		return nil, ErrConfig
	}
	if cfg.ExchangePolicy == SingleAttempt {
		if cfg.MaxAttempts != 1 || cfg.MaxReplayAge != 0 || cfg.ReplayDomain != (ReplayDomain{}) {
			return nil, ErrConfig
		}
	} else {
		op, ok := n.cfg.Bundle.Lookup(cfg.Command)
		if !ok || !op.TransportReplay {
			return nil, ErrUnsupported
		}
		if avp.Identity(cfg.ReplayDomain.Target.Host) != nil || avp.Identity(cfg.ReplayDomain.Target.Realm) != nil || cfg.MaxAttempts < 2 || cfg.MaxAttempts > 16 || cfg.MaxReplayAge <= 0 || cfg.ReplayDomain.Retention < cfg.MaxReplayAge || cfg.ExchangePolicy == TransportFailover && len(cfg.Peers) < 2 {
			return nil, ErrConfig
		}
	}
	r := &Route{node: n, cfg: cfg}
	r.cfg.TargetID = avp.CloneString(cfg.TargetID)
	r.cfg.Peers = append([]PeerID(nil), cfg.Peers...)
	seen := map[PeerID]bool{}
	for _, id := range cfg.Peers {
		p, ok := n.peers[id]
		if !ok || seen[id] {
			return nil, ErrConfig
		}
		seen[id] = true
		policy, ok := p.requestPolicy(cfg.Command, cfg.TargetID)
		if !ok {
			return nil, ErrConfig
		}
		if len(r.peers) == 0 {
			r.policy = policy
		} else {
			first := r.policy
			if first.ApplicationBinding != policy.ApplicationBinding || first.Class != policy.Class || !strings.EqualFold(first.DestinationHost, policy.DestinationHost) || !strings.EqualFold(first.DestinationRealm, policy.DestinationRealm) || first.MaxRequestBytes != policy.MaxRequestBytes || first.MaxAnswerBytes != policy.MaxAnswerBytes || first.Timeout != policy.Timeout {
				return nil, ErrConfig
			}
		}
		if cfg.ExchangePolicy != SingleAttempt {
			target := cfg.ReplayDomain.Target
			if policy.DestinationHost == "" || !sameOrigin(target, pdu.Origin{Host: policy.DestinationHost, Realm: policy.DestinationRealm}) || len(policy.Sources.Terminals) != 1 || !sameOrigin(target, policy.Sources.Terminals[0]) {
				return nil, ErrConfig
			}
		}
		r.peers = append(r.peers, p)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed || n.shutting {
		return nil, ErrClosed
	}
	return r, nil
}

// Submit uses ctx for admission only. The accepted logical request retains its
// original deadline, identity and completion port through all attempts.
func (r *Route) Submit(ctx context.Context, body pdu.Body, opts RequestOptions, port *CompletionPort) (RequestID, error) {
	r.node.mu.Lock()
	closed := r.node.closed || r.node.shutting
	r.node.mu.Unlock()
	if closed {
		return RequestID{}, ErrClosed
	}
	if body == nil || body.CommandKey() != r.cfg.Command {
		return RequestID{}, ErrUnsupported
	}
	if opts.TargetID != "" && opts.TargetID != r.cfg.TargetID {
		return RequestID{}, ErrConfig
	}
	opts.TargetID = r.cfg.TargetID
	last := error(ErrNoEligibleAlternate)
	for _, p := range r.peers {
		r.node.mu.Lock()
		ready := p.readyLocked()
		r.node.mu.Unlock()
		if !ready {
			continue
		}
		id, err := p.submit(ctx, body, opts, port, r)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, ErrNotReady) && !errors.Is(err, ErrBackpressure) {
			return RequestID{}, err
		}
		last = err
	}
	if err := ctx.Err(); err != nil {
		return RequestID{}, err
	}
	return RequestID{}, last
}
func (r *Route) Request(ctx context.Context, body pdu.Body, opts RequestOptions) (Result, error) {
	if r.cfg.ExchangePolicy == ExplicitRetransmit {
		return Result{}, ErrModeRequiresSubmit
	}
	n := r.node
	port, err := n.NewCompletionPort(CompletionLimits{1, n.total.LimitBytes})
	if err != nil {
		return Result{}, err
	}
	defer port.Close(context.Background())
	id, err := r.Submit(ctx, body, opts, port)
	if err != nil {
		return Result{}, err
	}
	c, err := port.Receive(ctx)
	if err != nil {
		_ = n.Cancel(id)
		c, err = port.Receive(context.Background())
	}
	if err != nil {
		return Result{}, err
	}
	return c.Result, c.Result.Err
}

type routeExchange struct {
	recovery            *roSegmentState
	route               *Route
	root                *exchange
	attempts            []*exchange
	peerBytes           uint64
	frame               []byte
	phase               RequestPhase
	selected, published bool
	result              Result
	winner              *exchange
}

func removeAttemptQueue(x *exchange) {
	if x.conn == nil {
		return
	}
	x.conn.removeRequest(x)
}

type transportFailure struct{ error }

func ioStopReason(err error) StopReason {
	if errors.Is(err, transport.ErrNotificationLimit) {
		return StopBackpressure
	}
	if errors.Is(err, transport.ErrAssociationClosed) {
		return StopTransport
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrClosed) || errors.Is(err, ErrShutdown) || errors.Is(err, ErrPeerClosed) || errors.Is(err, ErrRemoteDPR) {
		return StopLocal
	}
	if errors.Is(err, ErrProtocol) || errors.Is(err, ErrConfig) || errors.Is(err, ErrBackpressure) || errors.Is(err, transport.ErrFrameState) || errors.Is(err, wire.ErrLimit) || errors.Is(err, wire.ErrFrame) || errors.Is(err, wire.ErrAVP) || errors.Is(err, validation.ErrInvalidAnswer) {
		return StopProtocol
	}
	// Only actual I/O sites call this function. Unknown adapter/TLS errors are
	// terminal: a protocol alert must not masquerade as a failed transport path.
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.ErrShortWrite) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, net.ErrClosed) || errors.Is(err, context.DeadlineExceeded) {
		return StopTransport
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return StopTransport
	}
	for _, e := range []syscall.Errno{syscall.ECONNRESET, syscall.ECONNABORTED, syscall.ESHUTDOWN, syscall.EPIPE, syscall.ETIMEDOUT, syscall.ENETRESET, syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.ENOTCONN} {
		if errors.Is(err, e) {
			return StopTransport
		}
	}
	return StopProtocol
}
func (n *Node) stopAttemptLocked(x *exchange, err error, reason StopReason) {
	if x.group == nil {
		n.chooseLocked(x, Result{Err: err})
		return
	}
	g := x.group
	// Tr stops scheduling without declaring the connection failed. A later real
	// stop of the current attempt must still honor local/protocol termination.
	if x.stopped && x.stopReason == StopResponseWait && reason != StopResponseWait && g.attempts[len(g.attempts)-1] == x {
		x.stopped = false
	}
	if !x.stopped {
		x.stopped = true
		x.stopReason = reason
		x.stopErr = err
		removeAttemptQueue(x)
		if reason != StopTransport && reason != StopWatchdog && reason != StopResponseWait || g.route.cfg.ExchangePolicy == SingleAttempt {
			n.chooseRouteLocked(g, x, Result{Err: err, StopReason: reason})
		}
	}
	n.advanceRouteLocked(g)
}
func (n *Node) retireAttemptPeerLocked(x *exchange) {
	if x.peerCharge != nil && x.conn != nil && x.stopped && x.conn.closed && !x.writing && !x.receiving {
		x.peerCharge.Release()
		x.peerCharge = nil
		delete(x.conn.pending, x.hbh)
		x.conn = nil
		x.frame = nil
	}
}
func (n *Node) advanceRouteLocked(g *routeExchange) {
	if g.published {
		return
	}
	for _, a := range g.attempts {
		n.retireAttemptPeerLocked(a)
	}
	if g.selected {
		n.publishRouteLocked(g)
		return
	}
	cur := g.attempts[len(g.attempts)-1]
	if !cur.stopped || cur.writing {
		return
	}
	if n.closed || n.shutting {
		n.chooseRouteLocked(g, cur, Result{Err: ErrShutdown, StopReason: StopLocal})
		return
	}
	if !time.Now().Before(g.root.deadline) {
		n.chooseRouteLocked(g, cur, Result{Err: context.DeadlineExceeded, StopReason: StopLocal})
		return
	}
	if len(g.attempts) >= g.route.cfg.MaxAttempts {
		if cur.stopReason == StopResponseWait {
			n.chooseRouteLocked(g, cur, Result{Err: ErrResponseWaitExpired, StopReason: StopResponseWait})
			return
		}
		n.chooseRouteLocked(g, cur, Result{Err: ErrAttemptsExhausted, StopReason: StopAttemptsExhausted})
		return
	}
	if g.route.cfg.ExchangePolicy == ExplicitRetransmit {
		g.phase = AwaitingRetransmit
		return
	}
	if _, err := n.addRouteAttemptLocked(g); err != nil {
		reason := StopNoAlternate
		if errors.Is(err, ErrBackpressure) {
			reason = StopBackpressure
		}
		n.chooseRouteLocked(g, cur, Result{Err: err, StopReason: reason})
	}
}
func (n *Node) addRouteAttemptLocked(g *routeExchange) (AttemptID, error) {
	if g.recovery != nil && len(g.recovery.targets) > 0 {
		return n.addRoTargetAttemptLocked(g)
	}
	last := error(ErrNoEligibleAlternate)
	for _, p := range g.route.peers {
		if !p.readyLocked() || !p.conn.apps.allows(g.route.cfg.Command, g.route.policy.ApplicationBinding) {
			continue
		}
		c := p.conn
		used := false
		for _, a := range g.attempts {
			if a.attemptConnection == c.id {
				used = true
				break
			}
		}
		if used {
			continue
		}
		pc, err := budget.Reserve(g.peerBytes, 1, "route-peer", &p.account)
		if err != nil {
			last = ErrBackpressure
			continue
		}
		hbh, err := c.nextHBH(false)
		if err != nil {
			pc.Release()
			last = err
			continue
		}
		retransmit := g.recovery != nil && g.recovery.replayPending
		for _, a := range g.attempts {
			if a.writing {
				pc.Release()
				return AttemptID{}, ErrAttemptNotReplayable
			}
			retransmit = retransmit || a.delivery.MayHaveReachedPeer
		}
		policy, _ := p.requestPolicy(g.route.cfg.Command, g.route.cfg.TargetID)
		snapshot := g.root.snapshot.ForAttempt(hbh, retransmit, policy.Sources)
		wire.PutHeader(g.frame, snapshot.Header(), len(g.frame))
		id := AttemptID{g.root.id, uint32(len(g.attempts) + 1)}
		a := &exchange{scheduleClass: g.root.scheduleClass, id: g.root.id, conn: c, port: g.root.port, charge: g.root.charge, snapshot: snapshot, frame: g.frame, deadline: g.root.deadline, answerMax: g.root.answerMax, answerBound: g.root.answerBound, hbh: hbh, delivery: DeliveryEvidence{FrameBytes: len(g.frame)}, group: g, attemptID: id, peerCharge: pc, attemptPeer: p, attemptConnection: c.id}
		startRoResponseWait(a, time.Now())
		g.attempts = append(g.attempts, a)
		g.phase = AttemptActive
		c.pending[hbh] = a
		c.enqueueRequest(a)
		c.signal()
		n.notify()
		return id, nil
	}
	return AttemptID{}, last
}
func (n *Node) chooseRouteLocked(g *routeExchange, winner *exchange, result Result) {
	if g.selected {
		return
	}
	g.selected = true
	g.phase = Completing
	if result.Err != nil && result.StopReason == StopNone {
		result.StopReason = StopLocal
	}
	g.result = result
	g.winner = winner
	for _, a := range g.attempts {
		a.selected = true
		removeAttemptQueue(a)
	}
	n.publishRouteLocked(g)
}
func aggregateAttempts(g *routeExchange) DeliveryEvidence {
	var d DeliveryEvidence
	for _, a := range g.attempts {
		d.FrameBytes = max(d.FrameBytes, a.delivery.FrameBytes)
		d.Bytes = max(d.Bytes, a.delivery.Bytes)
		d.Uncertain = d.Uncertain || a.delivery.Uncertain
		d.MayHaveReachedPeer = d.MayHaveReachedPeer || a.delivery.MayHaveReachedPeer
	}
	return d
}
func (n *Node) publishRouteLocked(g *routeExchange) {
	if !g.selected || g.published {
		return
	}
	for _, a := range g.attempts {
		if a.writing || a.receiving {
			return
		}
	}
	g.published = true
	n.retireIdentifier(g.root.identifier)
	x := g.root
	delete(n.requests, x.id)
	result := g.result
	result.RequestID = x.id
	result.Delivery = aggregateAttempts(g)
	result.Attempts = make([]AttemptResult, len(g.attempts))
	anyWritten := false
	for i, a := range g.attempts {
		h := a.snapshot.Header()
		result.Attempts[i] = AttemptResult{TargetID: roAttemptTarget(g, a), ID: a.attemptID, PeerID: a.attemptPeer.id, ConnectionID: a.attemptConnection, HopByHop: a.hbh, Retransmitted: h.Flags&pdu.RetransmittedFlag != 0, Delivery: a.delivery, StopReason: a.stopReason, Err: a.stopErr, AcceptedAt: a.acceptedAt, ResponseWaitUntil: a.responseWaitUntil}
		anyWritten = anyWritten || (a.delivery.Bytes == a.delivery.FrameBytes && a.delivery.FrameBytes > 0 && !a.delivery.Uncertain)
		if a.conn != nil {
			delete(a.conn.pending, a.hbh)
		}
		a.conn = nil
		if a.peerCharge != nil {
			a.peerCharge.Release()
			a.peerCharge = nil
		}
		a.published = true
		a.frame = nil
		a.snapshot = nil
	}
	selected := g.winner
	if selected == nil {
		selected = g.attempts[len(g.attempts)-1]
	}
	result.PeerID = selected.attemptPeer.id
	result.ConnectionID = selected.attemptConnection
	if result.Answer.Body != nil {
		result.SelectedAttempt = selected.attemptID
		result.Outcome = AnswerReceived
	} else if anyWritten {
		result.Outcome = WrittenNoAnswer
	} else if result.Delivery.Uncertain {
		result.Outcome = WriteUncertain
	} else if result.Delivery.Bytes > 0 {
		result.Outcome = PartlyWritten
	} else {
		result.Outcome = NotWritten
	}
	n.finishRoSegmentLocked(g, &result)
	if g.recovery != nil {
		g.recovery.targets = nil
	}
	keep := uint64(1024 + len(g.attempts)*512)
	if result.Answer.Body != nil {
		keep += x.answerBound
	}
	_ = x.charge.SetHeld(0)
	_ = x.charge.Shrink(keep)
	x.charge.Transfer("completion")
	if x.port.closed {
		x.charge.Release()
	} else {
		idx := (x.port.head + x.port.count) % len(x.port.queue)
		x.port.queue[idx] = completionEntry{Completion{x.id, result}, x.charge}
		x.port.count++
	}
	g.frame = nil
	g.result = Result{}
	x.port.signal()
	n.notify()
}

// RequestInfo reports active Route requests only; it stores no completion history.
func (n *Node) RequestInfo(id RequestID) (RequestInfo, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if id.Boot != n.boot {
		return RequestInfo{}, ErrForeignRequest
	}
	x := n.requests[id]
	if x == nil {
		return RequestInfo{}, ErrRequestNotActive
	}
	if x.group == nil {
		return RequestInfo{}, ErrUnsupported
	}
	g := x.group
	cur := g.attempts[len(g.attempts)-1]
	transportReason := cur.stopReason
	if transportReason == StopResponseWait {
		transportReason = StopNone
	}
	reason := ReplayNone
	switch cur.stopReason {
	case StopTransport:
		reason = ReplayTransportStopped
	case StopWatchdog:
		reason = ReplayWatchdogStopped
	case StopResponseWait:
		reason = ResponseWaitExpired
	}
	if cur.responseWaitExpired && cur.stopReason == StopNone {
		reason = ResponseWaitExpired
	}
	return RequestInfo{TargetID: roAttemptTarget(g, cur), RequestID: id, Phase: g.phase, CurrentAttempt: cur.attemptID, ReplayEligible: g.phase == AwaitingRetransmit, AttemptsUsed: len(g.attempts), Deadline: x.deadline, Delivery: aggregateAttempts(g), TransportStopReason: transportReason, AcceptedAt: x.acceptedAt, AttemptAcceptedAt: cur.acceptedAt, ResponseWaitUntil: cur.responseWaitUntil, ReplayReason: reason}, nil
}

// Retransmit admits at most one successor to after. Repeating a successful edge
// returns the same AttemptID while the request is active, even if ctx is canceled.
// Temporary absence of a ready connection or budget leaves the original deadline.
func (n *Node) Retransmit(ctx context.Context, id RequestID, after AttemptID) (AttemptID, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if id.Boot != n.boot {
		return AttemptID{}, ErrForeignRequest
	}
	x := n.requests[id]
	if x == nil || x.selected {
		return AttemptID{}, ErrRequestNotActive
	}
	g := x.group
	if g == nil || g.route.cfg.ExchangePolicy != ExplicitRetransmit {
		return AttemptID{}, ErrUnsupported
	}
	if after.RequestID != id || after.Number == 0 || int(after.Number) > len(g.attempts) {
		return AttemptID{}, ErrStaleAttempt
	}
	if int(after.Number) < len(g.attempts) {
		return g.attempts[after.Number].attemptID, nil
	}
	if err := ctx.Err(); err != nil {
		return AttemptID{}, err
	}
	if n.closed || n.shutting {
		return AttemptID{}, ErrClosed
	}
	if !time.Now().Before(x.deadline) {
		n.chooseRouteLocked(g, x, Result{Err: context.DeadlineExceeded, StopReason: StopLocal})
		return AttemptID{}, context.DeadlineExceeded
	}
	if g.phase != AwaitingRetransmit {
		return AttemptID{}, ErrAttemptNotReplayable
	}
	return n.addRouteAttemptLocked(g)
}
