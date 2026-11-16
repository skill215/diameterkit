// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
)

type LocalCapabilities struct {
	HostIPAddress                        []avp.Address
	VendorID                             uint32
	ProductName                          string
	AuthApplicationID, AcctApplicationID []uint32
	VendorSpecificApplicationID          []pdu.VendorSpecificApplicationID
}
type Node struct {
	roTargets              map[string]roTargetConfig
	roRecovery             roRecoveryState
	identifiers            identifierState
	scheduler              schedulerPlan
	stopped                bool
	mu                     sync.Mutex // admission, BeginWrite, completion and all ledger mutations
	cfg                    NodeConfig
	engine                 *codecengine.Engine
	control                *codecengine.Engine
	peers                  map[PeerID]*Peer
	classes                map[ClassID]*budget.Account
	inboundClasses         map[ClassID]*budget.Account
	duplicateClasses       map[ClassID]*budget.Account
	duplicates             map[duplicateKey]*duplicateRecord
	duplicateConflicts     uint64
	inbounds               map[RequestID]*Inbound
	rejectAccount          budget.Account
	rejectBound            uint64
	total                  budget.Account
	ports                  map[*CompletionPort]bool
	requests               map[RequestID]*exchange
	boot                   [16]byte
	seq, connSeq           uint64
	appSeed, controlSeed   uint32
	appCount, controlCount uint64
	closed                 bool
	shutting               bool
	shutdownDeadline       time.Time
	shutdownErr            error
	shutdownChanged        chan struct{}
	wake                   chan struct{}
	done                   chan struct{}
	wg                     sync.WaitGroup
	fixed                  uint64
	controlBound           uint64
}

func add(a, b uint64) (uint64, error) {
	if b > math.MaxUint64-a {
		return 0, ErrConfig
	}
	return a + b, nil
}
func NewNode(cfg NodeConfig) (*Node, error) {
	if avp.Identity(cfg.Origin.Host) != nil || avp.Identity(cfg.Origin.Realm) != nil || cfg.Bundle == nil || cfg.Limits.Validate() != nil {
		return nil, ErrConfig
	}
	cfg.Origin = pdu.Origin{Host: avp.CloneString(cfg.Origin.Host), Realm: avp.CloneString(cfg.Origin.Realm)}
	t := cfg.Timeouts
	if t.Handshake <= 0 || t.FrameRead <= 0 || t.FrameWrite <= 0 || t.Request <= 0 || t.Watchdog < 6*time.Second || t.ShutdownTimeout <= 0 || t.DPAWait <= 0 || t.CloseReserve > t.ShutdownTimeout || t.CloseReserve <= 0 || t.ClosingGrace <= 0 || t.ClosingGrace > t.CloseReserve {
		return nil, ErrConfig
	}
	r := cfg.Resources
	if len(cfg.Peers) == 0 || len(cfg.Peers) > 1024 || len(r.Classes) == 0 || len(r.Classes) > 256 || r.MaxPorts <= 0 || r.MaxPorts > 4096 || r.MaxPortOutstanding <= 0 || r.MaxPortOutstanding > 65536 || r.IdentifierBytes < 4096 || r.MaxDiscardBytes < 20 || r.MaxDiscardBytes > cfg.Limits.MaxFrameBytes {
		return nil, ErrConfig
	}
	if r.ListenerBytesPerPeer > 1<<20 {
		return nil, ErrConfig
	}
	identifiers, err := compileIdentifiers(cfg.Identifiers, r, cfg.Origin.Host)
	if err != nil {
		return nil, err
	}
	if (cfg.AllowRoRecoveryImport && cfg.RoRecovery == (RoRecoveryLimits{})) || (len(identifiers.config.RecoverySpaces) > 0 && !cfg.AllowRoRecoveryImport) {
		return nil, ErrConfig
	}
	recovery, err := compileRoRecovery(cfg.RoRecovery, r, identifiers.custom, len(identifiers.config.RecoverySpaces))
	if err != nil {
		return nil, err
	}
	scheduler, err := compileScheduler(cfg.Scheduler, r, cfg.Limits.MaxFrameBytes)
	if err != nil {
		return nil, err
	}
	engine, err := codecengine.New(cfg.Bundle, cfg.Limits)
	if err != nil {
		return nil, err
	}
	// The control tables are fixed Base implementations, not extension overrides.
	bp, err := base.Profile()
	if err != nil {
		return nil, err
	}
	bb, err := schema.Compile(bp)
	if err != nil {
		return nil, err
	}
	controlEngine, err := codecengine.New(bb, cfg.Limits)
	if err != nil {
		return nil, err
	}
	for _, key := range []pdu.CommandKey{{Code: 257, Request: true}, {Code: 257}, {Code: 280, Request: true}, {Code: 280}, {Code: 282, Request: true}, {Code: 282}} {
		if _, ok := cfg.Bundle.Lookup(key); !ok {
			return nil, ErrConfig
		}
	}

	caps := cfg.Capabilities
	cer := &base.CER{OriginHost: cfg.Origin.Host, OriginRealm: cfg.Origin.Realm, HostIPAddress: caps.HostIPAddress, VendorID: caps.VendorID, ProductName: caps.ProductName, AuthApplicationID: caps.AuthApplicationID, AcctApplicationID: caps.AcctApplicationID, VendorSpecificApplicationID: caps.VendorSpecificApplicationID}
	cm := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257}, Body: cer}
	owned, err := controlEngine.Snapshot(cm)
	if err != nil {
		return nil, ErrConfig
	}
	c := owned.Body.(*base.CER)
	cfg.Capabilities = LocalCapabilities{c.HostIPAddress, c.VendorID, c.ProductName, c.AuthApplicationID, c.AcctApplicationID, c.VendorSpecificApplicationID}
	// Control RX, TX, queued request and association scratch are independently reserved.
	cb, err := controlEngine.DecodeBound(cm.Header, cfg.Limits.MaxFrameBytes)
	if err != nil || cb > math.MaxUint64/8 {
		return nil, ErrConfig
	}
	for _, code := range []uint32{257, 280, 282} {
		for _, flags := range []byte{0, pdu.ErrorFlag, pdu.RequestFlag} {
			b, e := controlEngine.DecodeBound(pdu.Header{CommandCode: code, Flags: flags}, cfg.Limits.MaxFrameBytes)
			if e != nil {
				return nil, ErrConfig
			}
			cb = max(cb, b)
		}
	}
	control := cb*8 + uint64(cfg.Limits.MaxFrameBytes)*4 + 8192 + transportEventReserve
	per, err := add(control, r.AdapterBytesPerConnection)
	if err != nil || per > math.MaxUint64/(2*uint64(len(cfg.Peers))) {
		return nil, ErrConfig
	}
	rejectBound, err := rejectionReserve(cfg, engine)
	if err != nil {
		return nil, err
	}
	minTransport, err := add(2*per*uint64(len(cfg.Peers)), rejectBound*uint64(r.RejectSlots))
	if err != nil {
		return nil, err
	}
	for _, pc := range cfg.Peers {
		if !pc.Maintain.valid() {
			return nil, ErrConfig
		}
		if pc.MaxPending <= 0 || pc.MaxPending > 65536 || pc.MaxInbound < 0 || pc.MaxInbound > 65536 {
			return nil, ErrConfig
		}
		minTransport, err = add(minTransport, 4096+uint64(pc.MaxPending)*1024+uint64(pc.MaxInbound)*512+uint64(r.RejectSlots)*128)
		if err != nil {
			return nil, err
		}
	}
	rings := uint64(r.MaxPorts)*uint64(r.MaxPortOutstanding)*512 + uint64(r.MaxPorts)*512
	minTransport, err = add(minTransport, rings+schedulerReserve(r, len(cfg.Peers))+listenerReserve(r, len(cfg.Peers)))
	if err != nil || r.TransportBytes < minTransport {
		return nil, ErrConfig
	}
	targetBytes, err := roTargetConfigReserve(cfg)
	if err != nil {
		return nil, err
	}
	requestTargetBytes, err := requestTargetReserve(cfg)
	if err != nil {
		return nil, err
	}
	fixed, err := add(r.TransportBytes, r.IdentifierBytes)
	if err == nil {
		fixed, err = add(fixed, targetBytes)
	}
	if err == nil {
		fixed, err = add(fixed, requestTargetBytes)
	}
	if err != nil {
		return nil, err
	}
	n := &Node{roRecovery: recovery, identifiers: identifiers, scheduler: scheduler, shutdownChanged: make(chan struct{}), cfg: cfg, engine: engine, control: controlEngine, peers: make(map[PeerID]*Peer), classes: make(map[ClassID]*budget.Account), ports: make(map[*CompletionPort]bool), requests: make(map[RequestID]*exchange), wake: make(chan struct{}, 1), done: make(chan struct{}), fixed: fixed, controlBound: control}
	n.inboundClasses = make(map[ClassID]*budget.Account)
	n.duplicateClasses = make(map[ClassID]*budget.Account)
	n.duplicates = make(map[duplicateKey]*duplicateRecord)
	n.inbounds = make(map[RequestID]*Inbound)
	n.rejectBound = rejectBound
	n.rejectAccount = budget.Account{LimitBytes: rejectBound * uint64(r.RejectSlots), LimitSlots: uint64(r.RejectSlots)}
	sum := fixed
	slots := uint64(0)
	for _, cl := range r.Classes {
		if cl.ID == "" || len(cl.ID) > 128 || cl.Bytes == 0 || cl.Slots <= 0 || cl.Slots > 1<<20 || n.classes[cl.ID] != nil {
			return nil, ErrConfig
		}
		sum, err = add(sum, cl.Bytes)
		if err != nil {
			return nil, err
		}
		slots += uint64(cl.Slots)
		n.classes[cl.ID] = &budget.Account{LimitBytes: cl.Bytes, LimitSlots: uint64(cl.Slots)}
	}
	if len(r.InboundClasses) > 256 {
		return nil, ErrConfig
	}
	for _, cl := range r.InboundClasses {
		if cl.ID == "" || len(cl.ID) > 128 || cl.Bytes == 0 || cl.Slots <= 0 || cl.Slots > 1<<20 || n.inboundClasses[cl.ID] != nil {
			return nil, ErrConfig
		}
		sum, err = add(sum, cl.Bytes)
		if err != nil {
			return nil, err
		}
		slots += uint64(cl.Slots)
		n.inboundClasses[cl.ID] = &budget.Account{LimitBytes: cl.Bytes, LimitSlots: uint64(cl.Slots)}
	}
	if len(r.DuplicateClasses) > 256 {
		return nil, ErrConfig
	}
	for _, cl := range r.DuplicateClasses {
		if cl.ID == "" || len(cl.ID) > 128 || cl.Bytes == 0 || cl.Slots <= 0 || cl.Slots > 1<<20 || n.duplicateClasses[cl.ID] != nil {
			return nil, ErrConfig
		}
		sum, err = add(sum, cl.Bytes)
		if err != nil {
			return nil, err
		}
		slots += uint64(cl.Slots)
		n.duplicateClasses[cl.ID] = &budget.Account{LimitBytes: cl.Bytes, LimitSlots: uint64(cl.Slots)}
	}
	if sum > r.NodeBytes {
		return nil, ErrConfig
	}
	n.total = budget.Account{LimitBytes: r.NodeBytes - fixed, LimitSlots: slots}
	for _, pc := range cfg.Peers {
		if pc.ID == "" || len(pc.ID) > 128 || n.peers[pc.ID] != nil || avp.Identity(pc.Origin.Host) != nil || avp.Identity(pc.Origin.Realm) != nil || strings.EqualFold(pc.Origin.Host, cfg.Origin.Host) || pc.MaxPending <= 0 || pc.MaxPending > 65536 || pc.MaxBytes == 0 || len(pc.Commands)+len(pc.InboundCommands) == 0 || len(pc.Commands) > 256 {
			return nil, ErrConfig
		}
		pc.Origin = pdu.Origin{Host: avp.CloneString(pc.Origin.Host), Realm: avp.CloneString(pc.Origin.Realm)}
		p := &Peer{node: n, id: pc.ID, maintainPolicy: pc.Maintain, origin: pc.Origin, account: budget.Account{LimitBytes: pc.MaxBytes, LimitSlots: uint64(pc.MaxPending)}, policies: make(map[pdu.CommandKey]CommandPolicy), maxPending: pc.MaxPending, lifecycle: Active, protocol: ProtocolClosed, watchdog: Initial, changed: make(chan struct{})}
		for _, policy := range pc.Commands {
			if !policy.Command.Request || !validApplicationBinding(policy.Command, policy.ApplicationBinding, cfg.Capabilities) || n.classes[policy.Class] == nil || policy.Timeout <= 0 || policy.Timeout > t.Request || policy.MaxRequestBytes < 20 || policy.MaxRequestBytes%4 != 0 || policy.MaxRequestBytes > cfg.Limits.MaxFrameBytes || policy.MaxAnswerBytes < 20 || policy.MaxAnswerBytes%4 != 0 || policy.MaxAnswerBytes > cfg.Limits.MaxFrameBytes || avp.Identity(policy.DestinationRealm) != nil || policy.DestinationHost != "" && avp.Identity(policy.DestinationHost) != nil {
				return nil, ErrConfig
			}
			if policy.ApplicationBinding == SCAP1Accounting {
				if policy.DestinationHost == "" || len(policy.Sources.Terminals) != 1 || !sameOrigin(policy.Sources.Terminals[0], pdu.Origin{Host: policy.DestinationHost, Realm: policy.DestinationRealm}) {
					return nil, ErrConfig
				}
				p.scap1Accounting = true
			}
			if _, ok := p.policies[policy.Command]; ok {
				return nil, ErrConfig
			}
			op, ok := cfg.Bundle.Lookup(policy.Command)
			if !ok || !op.Proxiable {
				return nil, ErrConfig
			}
			answer := policy.Command
			answer.Request = false
			if answerOp, found := cfg.Bundle.Lookup(answer); !found || answerOp.Error == nil {
				return nil, ErrConfig
			}
			if len(policy.Sources.Terminals) == 0 || len(policy.Sources.Terminals)+len(policy.Sources.ErrorAgents) > 64 {
				return nil, ErrConfig
			}
			for _, list := range [][]pdu.Origin{policy.Sources.Terminals, policy.Sources.ErrorAgents} {
				for _, v := range list {
					if avp.Identity(v.Host) != nil || avp.Identity(v.Realm) != nil {
						return nil, ErrConfig
					}
				}
			}
			policy.Sources.Terminals = append([]pdu.Origin(nil), policy.Sources.Terminals...)
			policy.Sources.ErrorAgents = append([]pdu.Origin(nil), policy.Sources.ErrorAgents...)
			for _, list := range [][]pdu.Origin{policy.Sources.Terminals, policy.Sources.ErrorAgents} {
				for i := range list {
					list[i] = pdu.Origin{Host: avp.CloneString(list[i].Host), Realm: avp.CloneString(list[i].Realm)}
				}
			}
			policy.DestinationHost = avp.CloneString(policy.DestinationHost)
			policy.DestinationRealm = avp.CloneString(policy.DestinationRealm)
			p.policies[policy.Command] = policy
		}
		if err = p.configureRequestTargets(pc.RequestTargets); err != nil {
			return nil, err
		}
		if pc.UnknownRequests > BaseErrorFailures || pc.UnknownRequests == BaseErrorFailures && r.RejectSlots == 0 {
			return nil, ErrConfig
		}
		p.unknownRequests = pc.UnknownRequests
		if err = n.configureInbound(p, pc); err != nil {
			return nil, err
		}
		if pc.SCAP1Advertisement > SCAP1LegacyDualAdvertisement {
			return nil, ErrConfig
		}
		if pc.SCAP1Advertisement == SCAP1LegacyDualAdvertisement {
			if !p.scap1Accounting || !validSCAP1LegacyCapabilities(cfg.Capabilities) {
				return nil, ErrConfig
			}
			// Fail configuration before opening a socket if the expanded CER
			// exceeds the same frame/AVP/depth/repeat limits used at handshake.
			answer, e := controlEngine.CompleteAnswer(cm, n.cea(2001), cfg.Origin)
			if e != nil {
				return nil, ErrConfig
			}
			for _, message := range []pdu.Message{cm, answer} {
				raw, e := controlEngine.Append(make([]byte, 0, cfg.Limits.MaxFrameBytes), message)
				if e != nil {
					return nil, ErrConfig
				}
				if _, e = projectSCAP1Advertisement(raw, cfg.Limits, true); e != nil {
					return nil, ErrConfig
				}
			}
		}
		p.scap1Advertisement = pc.SCAP1Advertisement
		n.peers[pc.ID] = p
	}
	if _, err = rand.Read(n.boot[:]); err != nil {
		return nil, err
	}
	n.appSeed = binary.BigEndian.Uint32(n.boot[:4]) & 0x7fffffff
	n.controlSeed = binary.BigEndian.Uint32(n.boot[4:8]) & 0x7fffffff
	// Do not retain caller-owned configuration slices.
	n.cfg.Peers = nil
	n.cfg.Scheduler = SchedulerConfig{}
	n.cfg.Identifiers = IdentifierConfig{}
	n.cfg.Resources.IdentifierClasses = nil
	n.cfg.Resources.RoRecoveryClasses = nil
	n.cfg.Resources.Classes = nil
	n.cfg.Resources.InboundClasses = nil
	n.cfg.Resources.DuplicateClasses = nil
	if err = n.configureRoTargets(cfg.RoTargets); err != nil {
		return nil, err
	}
	n.wg.Add(1)
	n.cfg.RoTargets = nil // owned compiled target table only
	go n.timerLoop()
	go func() {
		n.wg.Wait()
		n.mu.Lock()
		n.stopped = true
		n.roRecovery = roRecoveryState{}
		n.identifiers.slots = nil
		n.identifiers.buckets = nil
		n.identifiers.expiry = nil
		n.identifiers.classes = nil
		n.identifiers.classIndex = nil
		n.identifiers.config.EndToEnd = nil
		if len(n.ports) == 0 {
			n.fixed = 0
		}
		n.mu.Unlock()
		close(n.done)
	}()
	return n, nil
}
func (n *Node) Peer(id PeerID) (*Peer, bool) { p, ok := n.peers[id]; return p, ok }
func (n *Node) notify() {
	select {
	case n.wake <- struct{}{}:
	default:
	}
}

// Close revokes all new work immediately. Wait is the I/O/borrower barrier;
// unconsumed completions retain their charge until Receive or Port.Close.
func (n *Node) Close() error {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return nil
	}
	n.closed = true
	n.closeDuplicatesLocked()
	for _, x := range n.requests {
		n.chooseLocked(x, Result{Err: ErrClosed, StopReason: StopLocal})
	}
	for _, in := range n.inbounds {
		n.endInboundLocked(in, ErrClosed)
	}
	close(n.shutdownChanged)
	n.shutdownChanged = make(chan struct{})
	var conns []transport.Conn
	for _, p := range n.peers {
		p.lifecycle = Closed
		n.stopServingLocked(p, ErrClosed)
		for _, c := range p.candidates {
			if c != nil {
				n.stopConnectionLocked(c, ErrClosed)
				if c.io != nil {
					conns = append(conns, c.io)
				}
			}
		}
		p.signalLocked()
	}
	n.mu.Unlock()
	for _, c := range conns {
		_ = c.Close()
	}
	n.notify()
	return nil
}
func (n *Node) Wait(ctx context.Context) error {
	select {
	case <-n.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ResourceSnapshot reports logical commitments, not Go heap or kernel buffers.
type ResourceSnapshot struct {
	FixedBytes, OutboundBytes   uint64
	Outstanding, Pending, Ports int
	InboundBytes                uint64
	InboundOutstanding          int
	DuplicateBytes              uint64
	DuplicateRecords            int
	DuplicateConflicts          uint64
	RejectBytes                 uint64
	RejectOutstanding           int
}

func (n *Node) Resources() ResourceSnapshot {
	n.mu.Lock()
	defer n.mu.Unlock()
	s := ResourceSnapshot{FixedBytes: n.fixed, OutboundBytes: n.total.UsedBytes, Outstanding: int(n.total.UsedSlots), Ports: len(n.ports)}
	s.DuplicateConflicts = n.duplicateConflicts
	s.RejectBytes = n.rejectAccount.UsedBytes
	s.RejectOutstanding = int(n.rejectAccount.UsedSlots)
	for _, a := range n.inboundClasses {
		s.InboundBytes += a.UsedBytes
		s.InboundOutstanding += int(a.UsedSlots)
	}
	for _, a := range n.duplicateClasses {
		s.DuplicateBytes += a.UsedBytes
		s.DuplicateRecords += int(a.UsedSlots)
	}
	s.OutboundBytes -= s.DuplicateBytes
	s.Outstanding -= s.DuplicateRecords
	s.OutboundBytes -= s.InboundBytes
	s.Outstanding -= s.InboundOutstanding
	for _, p := range n.peers {
		if p.conn != nil {
			s.Pending += len(p.conn.pending)
		}
	}
	return s
}
