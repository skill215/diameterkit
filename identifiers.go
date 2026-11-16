package diameter

import (
	"encoding/hex"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"strconv"
	"time"
)

var (
	ErrIdentifierCollision    = errors.New("end-to-end candidate already active or quarantined")
	ErrIdentifierSourceFailed = errors.New("end-to-end source failed")
	ErrIdentifierRange        = errors.New("end-to-end candidate outside application space")
)

// EndToEndSource supplies application E2E candidates only. It must be bounded,
// nonblocking and perform no I/O. A Node serializes calls outside its state gate;
// concurrent admission while the source is busy fails with ErrIdentifiers.
// Shared instances must be concurrency-safe across Nodes. No global uniqueness
// or persistence is supplied by this interface.
type EndToEndSource interface{ NextEndToEnd() (uint32, error) }
type SequenceSpace struct {
	Prefix      uint32
	CounterBits uint8
}
type IdentifierConfig struct {
	RecoverySpaces    []SequenceSpace // Up to16 full ranges disjoint from ordinary application/control space.
	Space             SequenceSpace
	ReuseGuard        time.Duration
	EndToEnd          EndToEndSource
	MaxSessionIDBytes int
}
type IdentifierClass struct {
	ID    ClassID
	Slots int
}
type IdentifierInfo struct {
	RecoverySpaces    []SequenceSpace
	Stopped           bool
	Space             SequenceSpace
	ReuseGuard        time.Duration
	MaxSessionIDBytes int
	CustomSource      bool
	ApplicationBits   uint8
	// Allocated counts are the built-in current-round counters; custom application sources report zero.
	ApplicationAllocated, ControlAllocated, ApplicationActive, ControlActive uint64
	GuardSlots, Reserved, Active, Quarantine, Free                           int
	Classes                                                                  []IdentifierClassUsage
}
type IdentifierClassUsage struct {
	ID          ClassID
	Slots, Used int
}
type idSlot struct {
	value       uint32
	class, next int
	state       uint8
	reuseAfter  time.Time
}

const (
	idFree uint8 = iota
	idReserved
	idActive
	idQuarantine
)

type identifierState struct {
	custom                     bool
	reserved, active           int
	config                     IdentifierConfig
	half                       uint64
	prefix                     uint32
	appActive, controlActive   uint64
	appRetired, controlRetired time.Time
	session                    uint64
	slots                      []idSlot
	buckets, expiry            []int
	free                       int
	classes                    []IdentifierClassUsage
	classIndex                 map[ClassID]int
	sourceBusy                 bool
}

// identifierLease is owned by one logical request or control exchange, never
// one Route attempt. A control writer may outlive its answer/close decision.
type identifierLease struct {
	slot                            int
	control, done, writing, retired bool
}

func compileIdentifiers(cfg IdentifierConfig, r ResourcePlan, host string) (identifierState, error) {
	s := identifierState{config: cfg, free: -1}
	if cfg.Space == (SequenceSpace{}) {
		cfg.Space.CounterBits = 32
	}
	bits := cfg.Space.CounterBits
	if bits < 2 || bits > 32 || uint64(cfg.Space.Prefix) >= uint64(1)<<(32-bits) {
		return s, ErrConfig
	}
	if cfg.ReuseGuard == 0 {
		cfg.ReuseGuard = 240 * time.Second
	}
	if cfg.ReuseGuard < 240*time.Second || cfg.ReuseGuard > 365*24*time.Hour {
		return s, ErrConfig
	}
	if cfg.MaxSessionIDBytes == 0 {
		cfg.MaxSessionIDBytes = 512
	}
	// Two decimal uint32 words, separators and a 128-bit hex boot nonce.
	if cfg.MaxSessionIDBytes < len(host)+55 || cfg.MaxSessionIDBytes > 4096 {
		return s, ErrConfig
	}
	if len(cfg.RecoverySpaces) > 16 {
		return s, ErrConfig
	}
	a, b := sequenceBounds(cfg.Space)
	for i, space := range cfg.RecoverySpaces {
		if space.CounterBits < 1 || space.CounterBits > 31 || uint64(space.Prefix) >= uint64(1)<<(32-space.CounterBits) {
			return s, ErrConfig
		}
		lo, hi := sequenceBounds(space)
		if lo < b && a < hi {
			return s, ErrConfig
		}
		for _, old := range cfg.RecoverySpaces[:i] {
			x, y := sequenceBounds(old)
			if lo < y && x < hi {
				return s, ErrConfig
			}
		}
	}
	cfg.RecoverySpaces = append([]SequenceSpace(nil), cfg.RecoverySpaces...)
	s.config = cfg
	s.custom = cfg.EndToEnd != nil
	s.half = uint64(1) << (bits - 1)
	s.prefix = uint32(uint64(cfg.Space.Prefix) << bits)
	slots := r.IdentifierGuardSlots
	if cfg.EndToEnd == nil {
		if slots != 0 || len(r.IdentifierClasses) != 0 {
			return s, ErrConfig
		}
		return s, nil
	}
	if slots <= 0 || slots > 1<<20 || len(r.IdentifierClasses) != len(r.Classes) {
		return s, ErrConfig
	}
	// Fixed slots, <=4*slots hash buckets and expiry heap, class metadata/index.
	bound := uint64(4096) + uint64(slots)*128 + uint64(len(r.Classes))*256
	if r.IdentifierBytes < bound {
		return s, ErrConfig
	}
	s.classIndex = make(map[ClassID]int, len(r.Classes))
	s.classes = make([]IdentifierClassUsage, len(r.Classes))
	known := make(map[ClassID]bool, len(r.Classes))
	for _, c := range r.Classes {
		if c.ID == "" || len(c.ID) > 128 || known[c.ID] {
			return s, ErrConfig
		}
		known[c.ID] = true
	}
	sum := 0
	for i, c := range r.IdentifierClasses {
		if !known[c.ID] || c.Slots <= 0 || c.Slots > slots {
			return s, ErrConfig
		}
		if _, ok := s.classIndex[c.ID]; ok {
			return s, ErrConfig
		}
		sum += c.Slots
		id := ClassID(avp.CloneString(string(c.ID)))
		s.classIndex[id] = i
		s.classes[i] = IdentifierClassUsage{ID: id, Slots: c.Slots}
	}
	if sum > slots {
		return s, ErrConfig
	}
	s.slots = make([]idSlot, slots)
	s.free = 0
	for i := range s.slots {
		s.slots[i].next = i + 1
	}
	s.slots[slots-1].next = -1
	bucketCount := 2
	for bucketCount < 2*slots {
		bucketCount *= 2
	}
	s.buckets = make([]int, bucketCount)
	s.expiry = make([]int, 0, slots)
	return s, nil
}
func (s *identifierState) home(id uint32) int { return int(id*2654435761) & (len(s.buckets) - 1) }
func (s *identifierState) find(id uint32) (int, int) {
	mask := len(s.buckets) - 1
	j := s.home(id)
	for range s.buckets {
		v := s.buckets[j]
		if v == 0 {
			return -1, j
		}
		if s.slots[v-1].value == id {
			return v - 1, j
		}
		j = (j + 1) & mask
	}
	panic("identifier index capacity invariant")
}
func (s *identifierState) remove(id uint32) {
	_, hole := s.find(id)
	mask := len(s.buckets) - 1
	for j := (hole + 1) & mask; s.buckets[j] != 0; j = (j + 1) & mask {
		home := s.home(s.slots[s.buckets[j]-1].value)
		if (j-home)&mask >= (j-hole)&mask {
			s.buckets[hole] = s.buckets[j]
			hole = j
		}
	}
	s.buckets[hole] = 0
}
func (s *identifierState) release(slot int) {
	v := &s.slots[slot]
	s.classes[v.class].Used--
	if v.state == idReserved {
		s.reserved--
	} else if v.state == idActive {
		s.active--
	}
	*v = idSlot{next: s.free}
	s.free = slot
}
func (s *identifierState) expire(now time.Time) {
	for len(s.expiry) > 0 && !now.Before(s.slots[s.expiry[0]].reuseAfter) {
		slot := s.expiry[0]
		last := len(s.expiry) - 1
		s.expiry[0] = s.expiry[last]
		s.expiry = s.expiry[:last]
		for i := 0; 2*i+1 < len(s.expiry); {
			j := 2*i + 1
			if j+1 < len(s.expiry) && s.slots[s.expiry[j+1]].reuseAfter.Before(s.slots[s.expiry[j]].reuseAfter) {
				j++
			}
			if !s.slots[s.expiry[j]].reuseAfter.Before(s.slots[s.expiry[i]].reuseAfter) {
				break
			}
			s.expiry[i], s.expiry[j] = s.expiry[j], s.expiry[i]
			i = j
		}
		s.remove(s.slots[slot].value)
		s.release(slot)
	}
}
func (s *identifierState) reserve(class ClassID, now time.Time) (*identifierLease, error) {
	s.expire(now)
	i := s.classIndex[class]
	if s.sourceBusy || s.free < 0 || s.classes[i].Used >= s.classes[i].Slots {
		return nil, ErrIdentifiers
	}
	slot := s.free
	s.free = s.slots[slot].next
	s.slots[slot] = idSlot{class: i, state: idReserved}
	s.classes[i].Used++
	s.sourceBusy = true
	s.reserved++
	return &identifierLease{slot: slot}, nil
}
func (s *identifierState) install(l *identifierLease, id uint32) error {
	if uint64(id) < uint64(s.prefix) || uint64(id)-uint64(s.prefix) >= s.half {
		return errors.Join(ErrIdentifiers, ErrIdentifierRange)
	}
	found, bucket := s.find(id)
	if found >= 0 {
		return errors.Join(ErrIdentifiers, ErrIdentifierCollision)
	}
	s.reserved--
	s.active++
	s.slots[l.slot].state = idActive
	s.slots[l.slot].value = id
	s.buckets[bucket] = l.slot + 1
	return nil
}
func (n *Node) nextE2E(control bool) (uint32, error) {
	s := &n.identifiers
	count, seed, active, retired := &n.appCount, n.appSeed, &s.appActive, &s.appRetired
	if control {
		count, seed, active, retired = &n.controlCount, n.controlSeed, &s.controlActive, &s.controlRetired
	}
	now := time.Now()
	if *count >= s.half {
		if *active != 0 || retired.IsZero() || now.Before(retired.Add(s.config.ReuseGuard)) {
			return 0, ErrIdentifiers
		}
		*count = 0
	}
	v := s.prefix | uint32((uint64(seed)+*count)&(s.half-1))
	*count++
	// One bounded candidate per call: never scan or retry around a recovery guard.
	if !control && n.roRecovery.contains(v, now) {
		*retired = now
		return 0, errors.Join(ErrIdentifiers, ErrIdentifierCollision)
	}
	*active++
	if control {
		v |= uint32(s.half)
	}
	return v, nil
}
func (n *Node) retireIdentifier(l *identifierLease) {
	if l == nil || l.retired {
		return
	}
	l.done = true
	if l.writing {
		return
	}
	l.retired = true
	s := &n.identifiers
	now := time.Now()
	if l.control {
		s.controlActive--
		s.controlRetired = now
		return
	}
	if s.config.EndToEnd == nil {
		s.appActive--
		s.appRetired = now
		return
	}
	v := &s.slots[l.slot]
	if v.state == idReserved {
		s.release(l.slot)
		return
	}
	s.active--
	v.state = idQuarantine
	v.reuseAfter = now.Add(s.config.ReuseGuard)
	s.expiry = append(s.expiry, l.slot)
	for i := len(s.expiry) - 1; i > 0; {
		p := (i - 1) / 2
		if !s.slots[s.expiry[i]].reuseAfter.Before(s.slots[s.expiry[p]].reuseAfter) {
			break
		}
		s.expiry[i], s.expiry[p] = s.expiry[p], s.expiry[i]
		i = p
	}
}

// IdentifierInfo returns owned configuration and counters, never IDs or source.
func (n *Node) IdentifierInfo() IdentifierInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	s := &n.identifiers
	s.expire(time.Now())
	v := IdentifierInfo{RecoverySpaces: append([]SequenceSpace(nil), s.config.RecoverySpaces...), Stopped: n.stopped, Space: s.config.Space, ReuseGuard: s.config.ReuseGuard, MaxSessionIDBytes: s.config.MaxSessionIDBytes, CustomSource: s.custom, ApplicationBits: s.config.Space.CounterBits - 1, ApplicationAllocated: n.appCount, ControlAllocated: n.controlCount, ApplicationActive: s.appActive, ControlActive: s.controlActive, GuardSlots: len(s.slots), Classes: append([]IdentifierClassUsage(nil), s.classes...)}
	v.Reserved = s.reserved
	v.Active = s.active
	v.Quarantine = len(s.expiry)
	v.Free = len(s.slots) - v.Reserved - v.Active - v.Quarantine
	if v.CustomSource {
		v.ApplicationActive = uint64(v.Active)
	}
	return v
}

// NewSessionID creates an owned bounded ID. It does not create or track a session.
func (n *Node) NewSessionID() (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed || n.shutting {
		return "", ErrClosed
	}
	s := &n.identifiers
	if s.session == ^uint64(0) {
		return "", ErrIdentifiers
	}
	s.session++
	var scratch [512]byte
	b := append(scratch[:0], n.cfg.Origin.Host...)
	b = append(b, ';')
	b = strconv.AppendUint(b, s.session>>32, 10)
	b = append(b, ';')
	b = strconv.AppendUint(b, s.session&0xffffffff, 10)
	b = append(b, ';')
	b = hex.AppendEncode(b, n.boot[:])
	return string(b), nil
}

// A preparation failure was never admitted or offered to the transport.
func (n *Node) abandonIdentifier(l *identifierLease) {
	if l == nil || l.retired {
		return
	}
	if n.identifiers.config.EndToEnd == nil || l.control {
		n.retireIdentifier(l)
		return
	}
	s := &n.identifiers
	if s.slots[l.slot].state == idActive {
		s.remove(s.slots[l.slot].value)
	}
	s.release(l.slot)
	l.retired = true
}
func (n *Node) startIdentifierDrainLocked(c *connection) {
	if c.identifierDrain.IsZero() {
		c.identifierDrain = time.Now().Add(n.cfg.Timeouts.ShutdownTimeout - n.cfg.Timeouts.CloseReserve)
		c.peer.signalLocked()
		n.notify()
	}
}
func (n *Node) advanceIdentifierDrainLocked(c *connection, now time.Time) error {
	if c.identifierDrain.IsZero() || c.closed || c.peer.protocol == Closing {
		return nil
	}
	drained := len(c.pending) == 0 && len(c.queue) == 0 && len(c.controls) == 0 && !c.writerBusy && c.peer.preparing == 0 && !n.activeInboundLocked(c)
	if drained || !now.Before(c.identifierDrain) {
		n.enterClosingLocked(c, now.Add(n.cfg.Timeouts.CloseReserve), ErrIdentifiers)
		return n.queueDPRLocked(c, now)
	}
	return nil
}
