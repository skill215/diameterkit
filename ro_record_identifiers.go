package diameter

import (
	"errors"
	"time"

	"github.com/skill215/diameterkit/avp"
)

type roRecoveryGuard struct {
	admitting                  bool
	replayUntil                time.Time
	active                     RequestID
	segmentNo, reservedThrough uint32
	segmentDigest              [32]byte
	cumulative                 DeliveryEvidence
	continuable                bool
	nextTrigger                RoRecoveryTrigger
	replayPending              bool
	reserved, occupied         bool
	class                      int
	e2e                        uint32
	recordID                   [16]byte
	digest                     [32]byte
	keepUntil                  time.Time // process-local monotonic deadline; exported record uses UTC
}
type roRecoveryState struct {
	slots      []roRecoveryGuard
	byE2E      map[uint32]int
	classes    []IdentifierClassUsage
	classIndex map[ClassID]int
}

func compileRoRecovery(l RoRecoveryLimits, r ResourcePlan, custom bool, spaces int) (roRecoveryState, error) {
	var s roRecoveryState
	if l == (RoRecoveryLimits{}) {
		if r.RoRecoverySlots != 0 || len(r.RoRecoveryClasses) != 0 {
			return s, ErrConfig
		}
		return s, nil
	}
	if !validRoLimits(l) || r.RoRecoverySlots < 1 || r.RoRecoverySlots > 4096 || len(r.RoRecoveryClasses) == 0 || len(r.RoRecoveryClasses) > len(r.Classes) {
		return s, ErrConfig
	}
	// Fixed guard + index capacity, no complete message or completion history.
	bound := uint64(4096+spaces*32) + uint64(r.RoRecoverySlots)*512 + uint64(len(r.RoRecoveryClasses))*256
	if custom {
		bound += uint64(r.IdentifierGuardSlots)*128 + uint64(len(r.Classes))*256
	}
	if r.IdentifierBytes < bound {
		return s, ErrConfig
	}
	s.slots = make([]roRecoveryGuard, r.RoRecoverySlots)
	s.byE2E = make(map[uint32]int, r.RoRecoverySlots)
	s.classIndex = make(map[ClassID]int, len(r.RoRecoveryClasses))
	known := make(map[ClassID]bool, len(r.Classes))
	for _, cl := range r.Classes {
		known[cl.ID] = true
	}
	sum := 0
	for _, cl := range r.RoRecoveryClasses {
		if !known[cl.ID] || cl.Slots <= 0 || cl.Slots > r.RoRecoverySlots {
			return roRecoveryState{}, ErrConfig
		}
		if _, ok := s.classIndex[cl.ID]; ok {
			return roRecoveryState{}, ErrConfig
		}
		sum += cl.Slots
		id := ClassID(avp.CloneString(string(cl.ID)))
		s.classIndex[id] = len(s.classes)
		s.classes = append(s.classes, IdentifierClassUsage{ID: id, Slots: cl.Slots})
	}
	if sum > r.RoRecoverySlots {
		return roRecoveryState{}, ErrConfig
	}
	return s, nil
}
func (s *roRecoveryState) release(slot int) {
	g := s.slots[slot]
	if g.occupied {
		delete(s.byE2E, g.e2e)
	}
	if g.occupied || g.reserved {
		s.classes[g.class].Used--
		s.slots[slot] = roRecoveryGuard{}
	}
}
func (s *roRecoveryState) expire(now time.Time) {
	for i, g := range s.slots {
		if g.occupied && !g.admitting && g.active == (RequestID{}) && !now.Before(g.keepUntil) {
			s.release(i)
		}
	}
}
func (s *roRecoveryState) contains(e2e uint32, now time.Time) bool {
	slot, ok := s.byE2E[e2e]
	if ok && !s.slots[slot].admitting && s.slots[slot].active == (RequestID{}) && !now.Before(s.slots[slot].keepUntil) {
		s.release(slot)
		return false
	}
	return ok
}
func (s *roRecoveryState) reserve(class ClassID, now time.Time) (int, error) {
	s.expire(now)
	i, ok := s.classIndex[class]
	if !ok || s.classes[i].Used >= s.classes[i].Slots {
		return -1, ErrIdentifiers
	}
	for slot, g := range s.slots {
		if !g.reserved && !g.occupied {
			s.slots[slot] = roRecoveryGuard{class: i, reserved: true}
			s.classes[i].Used++
			return slot, nil
		}
	}
	return -1, ErrIdentifiers
}
func (s *roRecoveryState) commit(slot int, e2e uint32, id [16]byte, digest [32]byte, until time.Time) {
	g := &s.slots[slot]
	g.reserved = false
	g.occupied = true
	g.e2e = e2e
	g.recordID = id
	g.digest = digest
	g.keepUntil = until
	s.byE2E[e2e] = slot
}
func (n *Node) installApplicationIdentifier(l *identifierLease, id uint32) error {
	if n.roRecovery.contains(id, time.Now()) {
		return errors.Join(ErrIdentifiers, ErrIdentifierCollision)
	}
	return n.identifiers.install(l, id)
}

// RoRecoveryInfo reports only owned bounded quota usage. No original request,
// record identifier, E2E, digest, historic answer or storage handle is exposed.
type RoRecoveryInfo struct {
	Stopped                     bool
	Limits                      RoRecoveryLimits
	Slots, Reserved, Held, Free int
	Classes                     []IdentifierClassUsage
}

func (n *Node) RoRecoveryInfo() RoRecoveryInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	s := &n.roRecovery
	s.expire(time.Now())
	out := RoRecoveryInfo{Stopped: n.stopped, Limits: n.cfg.RoRecovery, Slots: len(s.slots), Classes: append([]IdentifierClassUsage(nil), s.classes...)}
	for _, g := range s.slots {
		if g.reserved {
			out.Reserved++
		}
		if g.occupied {
			out.Held++
		}
	}
	out.Free = out.Slots - out.Reserved - out.Held
	return out
}
