package diameter

import (
	"context"
	"encoding/binary"
	"errors"
	"time"
)

// RoRecoveryCheckpoint is a finite application-owned observation, not authenticated
// history. For an uncertain crash, reserve the entire possibly admitted segment,
// set Unknown, and provide conservative Uncertain/MayHaveReachedPeer evidence.
type RoRecoveryCheckpoint struct {
	RecordID                   [16]byte
	SegmentNo, ReservedThrough uint32
	Delivery                   DeliveryEvidence
	NextTrigger                RoRecoveryTrigger
	ReplayPending              bool
	ErrorResultCode            uint32 // Base namespace only, from a previously validated allowed answer.
	Unknown                    bool
}

// RoRecoveryImport declares application checks completed outside this library.
// No flag is proof of durable commit, fencing, trustworthy time or remote deduplication.
type RoRecoveryImport struct {
	Checkpoint                                      RoRecoveryCheckpoint
	CheckedAt                                       time.Time
	SenderFenced, ProgressConfirmed, ClockConfirmed bool
}

var ErrRoRecoveryImport = errors.New("invalid or unauthorized Ro recovery import")

// Checkpoint returns protocol progress only. The application must decide whether
// this result was durably observed and fence the previous sender before import.
func (r RoRecoveryResult) Checkpoint() RoRecoveryCheckpoint {
	return RoRecoveryCheckpoint{RecordID: r.RecordID, SegmentNo: r.SegmentNo, ReservedThrough: r.ReservedThrough, Delivery: r.CumulativeDelivery, NextTrigger: r.NextTrigger, ReplayPending: r.ReplayPending, ErrorResultCode: r.ErrorResultCode}
}
func (w *roDigestWriter) imported(v *RoRecoveryImport) {
	if v == nil {
		w.number(0)
		return
	}
	w.number(1)
	c := v.Checkpoint
	w.data(c.RecordID[:])
	w.number(uint64(c.SegmentNo))
	w.number(uint64(c.ReservedThrough))
	w.number(uint64(c.NextTrigger))
	w.number(uint64(c.ErrorResultCode))
	w.number(uint64(c.Delivery.Bytes))
	w.number(uint64(c.Delivery.FrameBytes))
	for _, b := range []bool{c.Delivery.Uncertain, c.Delivery.MayHaveReachedPeer, c.ReplayPending, c.Unknown, v.SenderFenced, v.ProgressConfirmed, v.ClockConfirmed} {
		if b {
			w.number(1)
		} else {
			w.number(0)
		}
	}
	w.number(uint64(v.CheckedAt.Unix()))
	w.number(uint64(v.CheckedAt.Nanosecond()))
}
func sequenceBounds(s SequenceSpace) (uint64, uint64) {
	lo := uint64(s.Prefix) << s.CounterBits
	return lo, lo + (uint64(1) << s.CounterBits)
}
func (n *Node) roImportIdentifier(id uint32, now time.Time) error {
	s := &n.identifiers
	v := uint64(id)
	lo := uint64(s.prefix)
	if v >= lo && v < lo+s.half {
		if s.custom {
			if s.sourceBusy {
				return ErrIdentifiers
			}
			s.expire(now)
			if slot, _ := s.find(id); slot >= 0 {
				return errors.Join(ErrIdentifiers, ErrIdentifierCollision)
			}
		} else {
			// The sequence is circular within its half, with a fixed round seed.
			offset := ((v - lo) - uint64(n.appSeed)) & (s.half - 1)
			if offset < n.appCount {
				return errors.Join(ErrIdentifiers, ErrIdentifierCollision)
			}
		}
		return nil
	}
	for _, space := range s.config.RecoverySpaces {
		a, b := sequenceBounds(space)
		if v >= a && v < b {
			return nil
		}
	}
	return errors.Join(ErrIdentifiers, ErrIdentifierRange)
}
func validateRoImport(r RoRecoveryRequest, v RoRecoveryImport, now time.Time) error {
	c := v.Checkpoint
	d := c.Delivery
	if !v.SenderFenced || !v.ProgressConfirmed || !v.ClockConfirmed || v.CheckedAt.IsZero() || v.CheckedAt.Before(r.CreatedAt) || now.Before(v.CheckedAt) || now.Before(r.CreatedAt) || !now.Before(r.ReplayUntil) {
		return ErrRoRecoveryImport
	}
	if c.RecordID != r.RecordID || c.SegmentNo >= uint32(r.Limits.MaxSegments) || c.ReservedThrough >= uint32(r.Limits.MaxTotalAttempts) || c.ReservedThrough < c.SegmentNo || d.Bytes < 0 || d.FrameBytes < 0 || d.FrameBytes > r.MaxFrameBytes || d.Bytes > d.FrameBytes || d.MayHaveReachedPeer != (d.Bytes > 0 || d.Uncertain) || c.ReplayPending && !d.MayHaveReachedPeer {
		return ErrRoRecoveryProgress
	}
	if c.SegmentNo == 0 {
		if c.ReservedThrough != 0 || d != (DeliveryEvidence{}) || c.NextTrigger != RoRecoveryInitial || c.ReplayPending || c.Unknown || c.ErrorResultCode != 0 {
			return ErrRoRecoveryProgress
		}
		return nil
	}
	if d.FrameBytes < len(r.Frame) {
		return ErrRoRecoveryProgress
	}
	switch c.NextTrigger {
	case RoRecoveryNoAnswer:
		if c.ErrorResultCode != 0 || c.Unknown && (!d.Uncertain || !d.MayHaveReachedPeer || !c.ReplayPending) {
			return ErrRoRecoveryProgress
		}
	case RoRecoveryErrorAnswer:
		mode := RoErrorTransient
		if r.PolicyID == RoSMSFWRecoveryPolicy {
			mode = RoErrorSMSFWBeforeTx
		} else if r.PolicyID != RoTransientRecoveryPolicy {
			return ErrRoRecoveryImport
		}
		if c.Unknown || c.ReplayPending || !d.MayHaveReachedPeer || !permitsRoError(mode, c.ErrorResultCode) {
			return ErrRoRecoveryProgress
		}
	default:
		return ErrRoRecoveryFinal
	}
	return nil
}

// Node gate held. This provisional guard has no network work or new E2E allocation.
// The caller pins it through all preparation and rolls it back on any refusal.
func (n *Node) beginRoImport(r RoRecoveryRequest, v RoRecoveryImport, ctx context.Context) (int, error) {
	if !n.cfg.AllowRoRecoveryImport {
		return -1, ErrRoRecoveryImport
	}
	if n.closed || n.shutting {
		return -1, ErrClosed
	}
	if e := ctx.Err(); e != nil {
		return -1, e
	}
	now := time.Now()
	if e := validateRoImport(r, v, now); e != nil {
		return -1, e
	}
	id := binary.BigEndian.Uint32(r.Frame[16:20])
	if e := n.roImportIdentifier(id, now); e != nil {
		return -1, e
	}
	n.roRecovery.expire(now)
	for _, g := range n.roRecovery.slots {
		if g.occupied && g.recordID == r.RecordID {
			return -1, ErrRoRecoveryConflict
		}
	}
	slot, e := n.roRecovery.reserve(r.Targets[0].Policy.Class, now)
	if e != nil {
		return -1, e
	}
	until := now.Add(r.ReplayUntil.Sub(now)) // retain a process-local monotonic deadline
	n.roRecovery.commit(slot, id, r.RecordID, r.Digest, until.Add(n.identifiers.config.ReuseGuard))
	g := &n.roRecovery.slots[slot]
	c := v.Checkpoint
	g.replayUntil = until
	g.segmentNo = c.SegmentNo
	g.reservedThrough = c.ReservedThrough
	g.cumulative = c.Delivery
	g.replayPending = c.ReplayPending
	g.nextTrigger = c.NextTrigger
	g.continuable = true
	return slot, nil
}
