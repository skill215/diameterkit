package diameter

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/wire"
)

var ErrDuplicateConflict = errors.New("duplicate identity has conflicting content or policy")

// DuplicatePolicy enables Node-local volatile duplicate suppression. KeepUntil is
// first admission + MaxProcessing + ReplayRetention and never refreshes on hits.
// MaxResultBytes counts encoded answer AVPs except Proxy-Info; MaxEnvelopeBytes
// counts encoded Proxy-Info only. Their sum plus 20 must fit MaxAnswerBytes.
type DuplicatePolicy struct {
	Class                                                            ClassID
	MaxProcessing, ReplayRetention                                   time.Duration
	MaxArrivals, MaxCanonicalBytes, MaxResultBytes, MaxEnvelopeBytes int
}
type duplicateKey struct {
	origin string
	end    uint32
}
type duplicateSemantics struct {
	policy                DuplicatePolicy
	class                 ClassID
	requestMax, answerMax int
	failure               FailurePolicy
}
type duplicateRecord struct {
	key                            duplicateKey
	semantics                      duplicateSemantics
	canonical                      []byte
	charge                         *budget.Reservation
	owner                          *Inbound
	arrivals                       map[*Inbound]bool
	deadline, keepUntil            time.Time
	committed, suppressed, expired bool
	result                         []byte // Validated immutable answer AVPs, no Proxy-Info or header.
	flags                          byte
}

func (n *Node) configureDuplicate(cfg InboundPolicy, ab uint64) (uint64, error) {
	d := cfg.Duplicate
	if cfg.NoCache {
		if d != (DuplicatePolicy{}) {
			return 0, ErrConfig
		}
		return 0, nil
	}
	op, _ := n.cfg.Bundle.Lookup(cfg.Command)
	if op.Canonical == nil || cfg.FailurePolicy != BaseErrorFailures || n.duplicateClasses[d.Class] == nil || d.MaxProcessing != cfg.ExecutionTimeout || d.MaxProcessing <= 0 || d.ReplayRetention <= 0 || d.ReplayRetention > time.Duration(math.MaxInt64)-d.MaxProcessing || d.MaxArrivals < 1 || d.MaxArrivals > 65536 || d.MaxCanonicalBytes < 12 || d.MaxCanonicalBytes > 2*cfg.MaxRequestBytes || d.MaxResultBytes < 1 || d.MaxEnvelopeBytes < 0 || d.MaxResultBytes > cfg.MaxAnswerBytes-20 || d.MaxEnvelopeBytes > cfg.MaxAnswerBytes-20-d.MaxResultBytes {
		return 0, ErrConfig
	}
	extra := uint64(d.MaxCanonicalBytes)*2 + uint64(d.MaxArrivals)*128 + 131072
	if ab > (math.MaxUint64-extra)/8 {
		return 0, ErrConfig
	}
	return ab*8 + extra, nil
}
func semantics(p inboundPolicy) duplicateSemantics {
	return duplicateSemantics{p.Duplicate, p.Class, p.MaxRequestBytes, p.MaxAnswerBytes, p.FailurePolicy}
}

// splitResult operates on an already validated generated answer, stripping all
// Proxy-Info into a separate owned arrival envelope. No first-hop chain is cached.
func splitResult(frame []byte, l wire.Limits) (result, proxy []byte, err error) {
	d := wire.NewDecoder(frame[20:], l)
	for {
		v, ok, e := d.Next()
		if e != nil {
			return nil, nil, e
		}
		if !ok {
			break
		}
		start := 20 + v.Offset
		size := (int(wire.Uint24(frame[start+5:start+8])) + 3) &^ 3
		if v.Key == (avp.Key{Code: 284}) {
			proxy = append(proxy, frame[start:start+size]...)
		} else {
			result = append(result, frame[start:start+size]...)
		}
	}
	return
}

// attachDuplicateLocked is the sole lookup/create gate. The input has already
// passed its own source/destination policy; no peer identity enters the key.
func (n *Node) attachDuplicateLocked(in *Inbound, canonical []byte, origin string) error {
	key := duplicateKey{strings.ToLower(origin), in.request.Header.EndToEnd}
	now := time.Now()
	r := n.duplicates[key]
	if r != nil && !now.Before(r.keepUntil) {
		if r.owner != nil && !r.owner.decided {
			n.expireInboundLocked(r.owner)
		}
		n.retireDuplicateLocked(r)
		r = nil
	}
	if r != nil {
		if r.semantics != semantics(in.policy) || !bytes.Equal(r.canonical, canonical) {
			if n.duplicateConflicts < math.MaxUint64 {
				n.duplicateConflicts++
			}
			return ErrDuplicateConflict
		}
		if r.suppressed {
			return ErrInboundDone
		}
		active := 0
		for a := range r.arrivals {
			if a.conn != nil {
				active++
			}
		}
		if active >= in.policy.Duplicate.MaxArrivals {
			return ErrBackpressure
		}
		r.arrivals[in] = true
		in.duplicate = r
		in.info.ExecutionDeadline = r.deadline
		in.request = pdu.Message{}
		in.view = pdu.Message{}
		in.timeoutFrame = nil
		in.timeoutResult = nil
		if r.committed {
			n.queueReplayLocked(in, r)
		}
		return nil
	}
	charge, err := budget.Reserve(in.policy.cacheBound, 1, "duplicate-processing", &n.total, n.duplicateClasses[in.policy.Duplicate.Class])
	if err != nil {
		return ErrBackpressure
	}
	r = &duplicateRecord{key: key, semantics: semantics(in.policy), canonical: canonical, charge: charge, owner: in, arrivals: map[*Inbound]bool{in: true}, deadline: in.info.ExecutionDeadline, keepUntil: in.info.ExecutionDeadline.Add(in.policy.Duplicate.ReplayRetention)}
	n.duplicates[key] = r
	in.duplicate = r
	return nil
}
func (n *Node) queueReplayLocked(in *Inbound, r *duplicateRecord) {
	if in.decided {
		return
	}
	in.decided = true
	in.terminal = ErrInboundDone
	in.result.Committed = true
	close(in.done)
	in.replay = r
	in.writeDeadline = time.Now().Add(n.cfg.Timeouts.FrameWrite)
	in.result.PrimaryWrite.Delivery.FrameBytes = 20 + len(r.result) + len(in.proxy)
	if !n.mayWriteAnswerLocked(in, time.Now()) {
		if in.result.PrimaryWrite.Err == nil {
			in.result.PrimaryWrite.Err = ErrNotReady
		}
		n.releaseInboundLocked(in)
	} else {
		in.conn.enqueueAnswer(in)
		in.conn.signal()
	}
	n.notify()
}
func (n *Node) commitDuplicateLocked(in *Inbound, result []byte, flags byte, suppressed bool) {
	r := in.duplicate
	if r == nil || r.owner != in || r.committed {
		return
	}
	r.committed = true
	r.suppressed = suppressed
	r.result = result
	r.flags = flags
	r.charge.Transfer("duplicate-result")
	for a := range r.arrivals {
		if a != in {
			if suppressed {
				n.endInboundLocked(a, ErrInboundDone)
			} else {
				n.queueReplayLocked(a, r)
			}
		}
	}
}
func (n *Node) retireDuplicateLocked(r *duplicateRecord) {
	r.expired = true
	if n.duplicates[r.key] == r {
		delete(n.duplicates, r.key)
	}
	n.releaseDuplicateLocked(r)
}
func (n *Node) releaseDuplicateLocked(r *duplicateRecord) {
	if r.expired && len(r.arrivals) == 0 {
		r.canonical = nil
		r.result = nil
		r.owner = nil
		r.charge.Release()
	}
}
func (n *Node) closeDuplicatesLocked() {
	for _, r := range n.duplicates {
		n.retireDuplicateLocked(r)
	}
}
func (n *Node) releaseDuplicateArrivalLocked(in *Inbound) {
	if r := in.duplicate; r != nil {
		delete(r.arrivals, in)
		if r.owner == in {
			r.owner = nil
		}
		n.releaseDuplicateLocked(r)
		in.duplicate = nil
	}
	in.replay = nil
	in.proxy = nil
	in.timeoutResult = nil
}
