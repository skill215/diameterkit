package diameter

import (
	"context"
	"errors"
	"io"
	"math"
	"strings"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
)

type Lifecycle uint8

const (
	Active Lifecycle = iota
	Closed
	Draining
)

type ProtocolState uint8

const (
	ProtocolClosed ProtocolState = iota
	WaitCEA
	WaitCER
	IOpen
	ROpen
	Closing
	WaitReturns
	WaitConnAck
	WaitConnAckElect
)

type WatchdogState uint8

const (
	Initial WatchdogState = iota
	Okay
	Suspect
	Down
	Reopen
)

type PeerInfo struct {
	SendBarrier                     bool
	Transport                       TransportFacts
	IdentifierDraining              bool
	ServeFailures                   uint64
	ServeLastError                  error
	Serving, AcceptWaiting          bool
	Managed, ReconnectSuppressed    bool
	ReconnectFailures               uint32
	NextDial, ConnectDeadline       time.Time
	ReconnectCause                  error
	ID                              PeerID
	Lifecycle                       Lifecycle
	Protocol                        ProtocolState
	Watchdog                        WatchdogState
	Ready                           bool
	ConnectionID                    ConnectionID
	Pending                         int
	Cause                           error
	Candidates                      int
	DrainDeadline, ShutdownDeadline time.Time
	LastDPAResult                   uint32
}
type Peer struct {
	scap1Advertisement              SCAP1Advertisement
	targets                         map[requestTargetKey]CommandPolicy
	transport                       TransportFacts
	scap1Accounting                 bool // At least one explicitly bound inbound/outbound command.
	server                          *serving
	manager                         *maintenance
	maintainPolicy                  MaintainPolicy
	node                            *Node
	id                              PeerID
	origin                          pdu.Origin
	policies                        map[pdu.CommandKey]CommandPolicy
	maxPending                      int
	maxInbound                      int
	inboundAccount                  budget.Account
	inboundPolicies                 map[pdu.CommandKey]inboundPolicy
	inboundQueue                    []*Inbound
	unknownRequests                 FailurePolicy
	account                         budget.Account
	lifecycle                       Lifecycle
	protocol                        ProtocolState
	watchdog                        WatchdogState
	everOpen                        bool
	conn                            *connection
	cause                           error
	candidates                      [2]*connection
	chosen                          *connection
	handshakeDeadline               time.Time
	changed                         chan struct{}
	preparing                       int
	drainDeadline, shutdownDeadline time.Time
	shutdownErr                     error
	lastDPAResult                   uint32
}
type controlJob struct {
	identifier *identifierLease
	prepared   *preparedControl
	request    pdu.Message
	answer     pdu.Body
	probe      bool
	disconnect bool
	closeAfter bool
	deadline   time.Time
}
type connection struct {
	sendBarrierUntil                                     time.Time
	transport                                            TransportFacts
	identifierDrain                                      time.Time
	cerIdentifier, probeIdentifier, disconnectIdentifier *identifierLease
	accepting                                            bool
	okaySince                                            time.Time
	disconnectCause                                      uint32
	peer                                                 *Peer
	id                                                   ConnectionID
	io                                                   transport.Conn
	ctx                                                  context.Context
	cancel                                               context.CancelFunc
	closed                                               bool
	cause                                                error
	pending                                              map[uint32]*exchange
	queue                                                []*exchange
	answers                                              []*Inbound
	schedule                                             writerSchedule
	controls                                             []controlJob
	wake                                                 chan struct{}
	done                                                 chan struct{}
	workers                                              int
	jitter                                               uint32
	hbhSeed                                              uint32
	hbhCount                                             uint64
	apps                                                 *negotiatedApplications
	tw, closingDeadline                                  time.Time
	probeQueued                                          bool
	probe                                                *codecengine.RequestSnapshot
	probeDeadline                                        time.Time
	numDWA                                               int
	initiator, established, writerBusy                   bool
	handshakeDeadline                                    time.Time
	prefixResult                                         chan prefixResult
	prefixDone                                           chan struct{}
	localDPR, remoteDPR                                  bool
	disconnect                                           *codecengine.RequestSnapshot
	closeCause                                           error
}

func (p *Peer) readyLocked() bool {
	return p.lifecycle == Active && p.conn != nil && !p.conn.closed && p.conn.identifierDrain.IsZero() && (p.protocol == IOpen || p.protocol == ROpen) && p.watchdog == Okay
}
func (p *Peer) Info() PeerInfo {
	n := p.node
	n.mu.Lock()
	defer n.mu.Unlock()
	v := PeerInfo{Transport: p.transport, Serving: p.server != nil, ID: p.id, Lifecycle: p.lifecycle, Protocol: p.protocol, Watchdog: p.watchdog, Ready: p.readyLocked(), Cause: p.cause}
	if s := p.server; s != nil {
		v.ServeFailures = s.failures
		v.ServeLastError = s.lastError
	}
	v.DrainDeadline = p.drainDeadline
	v.ShutdownDeadline = p.shutdownDeadline
	v.LastDPAResult = p.lastDPAResult
	if m := p.manager; m != nil {
		v.Managed = true
		v.ReconnectSuppressed = m.suppressed
		v.ReconnectFailures = m.failures
		v.NextDial = m.next
		v.ReconnectCause = m.cause
	}
	for _, c := range p.candidates {
		if c != nil {
			v.Candidates++
			if c.accepting {
				v.AcceptWaiting = true
			}
			if !c.established && c.initiator {
				v.ConnectDeadline = c.handshakeDeadline
			}
		}
	}
	if p.conn != nil {
		v.SendBarrier = !p.conn.sendBarrierUntil.IsZero()
		v.IdentifierDraining = !p.conn.identifierDrain.IsZero()
		v.ConnectionID = p.conn.id
		v.Pending = len(p.conn.pending)
	}
	return v
}
func (p *Peer) Close() error {
	n := p.node
	n.mu.Lock()
	p.lifecycle = Closed
	n.stopServingLocked(p, ErrPeerClosed)
	for len(p.inboundQueue) > 0 {
		n.endInboundLocked(p.inboundQueue[0], ErrPeerClosed)
	}
	var conns []transport.Conn
	for _, c := range p.candidates {
		if c != nil {
			n.stopConnectionLocked(c, ErrPeerClosed)
			if c.io != nil {
				conns = append(conns, c.io)
			}
		}
	}
	p.signalLocked()
	n.mu.Unlock()
	for _, c := range conns {
		_ = c.Close()
	}
	n.notify()
	return nil
}
func (c *connection) signal() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}
func (c *connection) nextHBH(control bool) (uint32, error) {
	limit := uint64(1) << 32
	if !control {
		limit -= 1024
	}
	if c.hbhCount >= limit {
		if !control {
			c.peer.node.startIdentifierDrainLocked(c)
		}
		return 0, ErrIdentifiers
	}
	v := c.hbhSeed + uint32(c.hbhCount)
	c.hbhCount++
	return v, nil
}
func (n *Node) stopConnectionLocked(c *connection, cause error) {
	n.stopConnectionReasonLocked(c, cause, StopLocal)
}
func (n *Node) stopConnectionReasonLocked(c *connection, cause error, reason StopReason) {
	if c.closed {
		return
	}
	if c.closeCause != nil {
		cause = c.closeCause
		reason = StopLocal
	}
	if c.peer.conn == c {
		n.maintainLossLocked(c, cause)
	}
	c.closed = true
	c.cause = cause
	c.cancel()
	n.retireIdentifier(c.probeIdentifier)
	n.retireIdentifier(c.disconnectIdentifier)
	c.probeIdentifier = nil
	c.disconnectIdentifier = nil
	c.probe = nil
	c.disconnect = nil
	c.probeQueued = false
	c.controls = nil
	c.schedule.category.empty(scheduleControl)
	c.tw = time.Time{}
	p := c.peer
	if p.conn == c {
		p.protocol = ProtocolClosed
		p.cause = cause
		if p.everOpen {
			p.watchdog = Down
		} else {
			p.watchdog = Initial
		}
	}
	if p.chosen == c {
		p.chosen = nil
	}
	n.detachInboundConnectionLocked(c, cause)
	for _, x := range c.pending {
		n.stopAttemptLocked(x, cause, reason)
	}
	p.refreshHandshakeLocked()
	p.signalLocked()
	c.signal()
}
func (n *Node) fail(c *connection, cause error) {
	n.mu.Lock()
	reason := StopProtocol
	var failure transportFailure
	if errors.As(cause, &failure) {
		cause = failure.error
		reason = ioStopReason(cause)
	}
	n.stopConnectionReasonLocked(c, cause, reason)
	n.mu.Unlock()
	if c.io != nil {
		_ = c.io.Close()
	}
	n.notify()
}
func (n *Node) failTransport(c *connection, cause error) {
	n.mu.Lock()
	n.stopConnectionReasonLocked(c, cause, ioStopReason(cause))
	n.mu.Unlock()
	if c.io != nil {
		_ = c.io.Close()
	}
	n.notify()
}
func (n *Node) workerDone(c *connection) {
	n.mu.Lock()
	c.workers--
	if c.workers == 0 {
		n.retireConnectionLocked(c)
	}
	n.mu.Unlock()
	n.wg.Done()
	n.notify()
}
func (n *Node) cer() *base.CER {
	v := n.cfg.Capabilities
	return &base.CER{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, HostIPAddress: v.HostIPAddress, VendorID: v.VendorID, ProductName: v.ProductName, AuthApplicationID: v.AuthApplicationID, AcctApplicationID: v.AcctApplicationID, VendorSpecificApplicationID: v.VendorSpecificApplicationID}
}
func (n *Node) cea(code uint32) *base.CEA {
	v := n.cfg.Capabilities
	return &base.CEA{ResultCode: code, HostIPAddress: v.HostIPAddress, VendorID: v.VendorID, ProductName: v.ProductName, AuthApplicationID: v.AuthApplicationID, AcctApplicationID: v.AcctApplicationID, VendorSpecificApplicationID: v.VendorSpecificApplicationID}
}
func sameOrigin(a, b pdu.Origin) bool {
	return strings.EqualFold(a.Host, b.Host) && strings.EqualFold(a.Realm, b.Realm)
}
func (n *Node) commonApps(auth, acct []uint32, vs []pdu.VendorSpecificApplicationID) map[uint32]bool {
	remote := make(map[uint32]bool)
	relay := false
	for _, list := range [][]uint32{auth, acct} {
		for _, id := range list {
			remote[id] = true
			if id == math.MaxUint32 {
				relay = true
			}
		}
	}
	for _, v := range vs {
		if v.AuthApplicationID.Present {
			remote[v.AuthApplicationID.Value] = true
		}
		if v.AcctApplicationID.Present {
			remote[v.AcctApplicationID.Value] = true
		}
	}
	local := n.cfg.Capabilities
	common := make(map[uint32]bool)
	for _, list := range [][]uint32{local.AuthApplicationID, local.AcctApplicationID} {
		for _, id := range list {
			if id != 0 && id != math.MaxUint32 && (relay || remote[id]) {
				common[id] = true
			}
		}
	}
	for _, v := range local.VendorSpecificApplicationID {
		for _, id := range []uint32{v.AuthApplicationID.Value, v.AcctApplicationID.Value} {
			if id != 0 && id != math.MaxUint32 && (relay || remote[id]) {
				common[id] = true
			}
		}
	}
	return common
}
func (n *Node) handshake(ctx context.Context, c *connection, initiator bool) (*negotiatedApplications, error) {
	var req pdu.Message
	if initiator {
		n.mu.Lock()
		h, e := c.nextHBH(true)
		var end uint32
		if e == nil {
			end, e = n.nextE2E(true)
			if e == nil {
				c.cerIdentifier = &identifierLease{control: true, slot: -1}
			}
		}
		n.mu.Unlock()
		if e != nil {
			return nil, e
		}
		req = pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: h, EndToEnd: end}, Body: n.cer()}
		if e = n.writeControl(ctx, c, req); e != nil {
			return nil, e
		}
	}
	prefix, err := n.readPrefix(ctx, c)
	if err != nil {
		return nil, err
	}
	h, size, err := wire.Header(prefix[:], n.cfg.Limits)
	if err != nil {
		return nil, err
	}
	if h.ApplicationID != 0 || h.CommandCode != 257 || (h.Flags&pdu.RequestFlag != 0) == initiator {
		return nil, ErrProtocol
	}
	b := make([]byte, size)
	if err = n.readFrameInto(ctx, c, b); err != nil {
		return nil, err
	}
	if c.peer.scap1Advertisement == SCAP1LegacyDualAdvertisement {
		b, err = projectSCAP1Advertisement(b, n.cfg.Limits, false)
		if err != nil {
			return nil, err
		}
	}
	if initiator {
		m, err := n.control.DecodeAnswer(req, b, pdu.AnswerSources{Terminals: []pdu.Origin{c.peer.origin}})
		if err != nil {
			return nil, err
		}
		answer, ok := m.Body.(*base.CEA)
		if !ok || answer.ResultCode != 2001 {
			return nil, ErrProtocol
		}
		apps := n.negotiateApplications(c.peer, answer.AuthApplicationID, answer.AcctApplicationID, answer.VendorSpecificApplicationID)
		if !apps.usable(c.peer) {
			return nil, ErrProtocol
		}
		return apps, nil
	}
	req, err = n.control.DecodeOwned(b)
	if err != nil {
		return nil, err
	}
	cer, ok := req.Body.(*base.CER)
	if !ok || !sameOrigin(pdu.Origin{Host: cer.OriginHost, Realm: cer.OriginRealm}, c.peer.origin) {
		return nil, ErrProtocol
	}
	apps := n.negotiateApplications(c.peer, cer.AuthApplicationID, cer.AcctApplicationID, cer.VendorSpecificApplicationID)
	code := uint32(2001)
	if !apps.usable(c.peer) {
		code = 5010
	}
	if code == 2001 {
		if err = n.selectPassive(ctx, c); err != nil {
			return nil, err
		}
	}
	ans, err := n.control.CompleteAnswer(req, n.cea(code), n.cfg.Origin)
	if err != nil {
		return nil, err
	}
	if err = n.writeControl(ctx, c, ans); err != nil {
		return nil, err
	}
	if code != 2001 {
		return nil, ErrProtocol
	}
	return apps, nil
}
func (n *Node) writeControl(ctx context.Context, c *connection, m pdu.Message) error {
	legacy := c.peer.scap1Advertisement == SCAP1LegacyDualAdvertisement && m.Header.CommandCode == 257 && m.Header.ApplicationID == 0 && m.Header.Flags&pdu.ErrorFlag == 0
	var dst []byte
	if legacy {
		// One bounded TX frame, including the 44-byte legacy group. Appending the
		// group cannot allocate a second frame or escape the control reservation.
		dst = make([]byte, 0, n.cfg.Limits.MaxFrameBytes)
	}
	b, err := n.control.Append(dst, m)
	if err != nil {
		return err
	}
	if legacy {
		b, err = projectSCAP1Advertisement(b, n.cfg.Limits, true)
		if err != nil {
			return err
		}
	}
	r := c.io.WriteFrame(ctx, b)
	if r.Err != nil {
		return r.Err
	}
	if r.Uncertain || r.Bytes != len(b) {
		return io.ErrShortWrite
	}
	return nil
}
