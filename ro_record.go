package diameter

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"math"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// RoRecoveryLimits bounds preparation and local recovery segments. Zero disables
// both. Deployment retention is a declaration, never proof of OCS deduplication.
type RoRecoveryLimits struct {
	MaxRecordBytes, MaxTargets, MaxTotalAttempts, MaxSegments, MaxAttemptsPerSegment int
	MaxReplayAge, MaxResponseWait                                                    time.Duration
}

// RoRecoveryTarget selects a frozen outbound binding without selecting a connection.
// Distinct TargetIDs must identify distinct terminals; alternate paths belong to
// the later segment plan, not to this immutable target set.
type RoRecoveryTarget struct {
	ID        string
	Peer      PeerID
	Retention time.Duration
}
type RoRecoveryOptions struct {
	SessionEvidenceID string // Application evidence for this frozen TERM session; no library session store.
	PolicyID          string
	Targets           []RoRecoveryTarget
	Limits            RoRecoveryLimits
	Validation        validation.Context
}

// RoRecoveryBinding contains only protocol/deployment data. It has no Peer pointer,
// connection, HBH, request handle, callback or external owner/storage authority.
type RoRecoveryBinding struct {
	AllowedPeers                                []PeerID // Frozen for configured recovery targets; empty for the original Peer-policy fallback.
	ReplayDomain, EvidenceID, SessionEvidenceID string
	TargetID                                    string
	Retention                                   time.Duration
	Policy                                      CommandPolicy
}

// RoRecoveryRequest is owned, versioned protocol data that applications may save.
// Frame has HBH=0 and T=0 and must never be sent directly. Its original E2E is
// reserved locally through ReplayUntil+ReuseGuard and any later borrower retirement.
// The digest detects corruption;
// it authenticates neither its producer nor application progress. Mutating a copy
// does not mutate the Node guard or authorize a different request under this E2E.
// Only empty validation context is supported by the four currently audited CCR
// profiles. ContextName and WireFingerprint freeze that interpretation.
// SubmitRoRecovery accepts local records or explicitly authorized bounded imports.
// The application owns storage, sender fencing and confirmed historical progress.
type RoRecoveryRequest struct {
	SessionEvidenceID                      string
	Version                                uint32
	RecordID                               [16]byte
	WireFingerprint, ContextName, PolicyID string
	CreatedAt, ReplayUntil                 time.Time
	Limits                                 RoRecoveryLimits
	Targets                                []RoRecoveryBinding
	MaxFrameBytes                          int
	Frame                                  []byte
	Digest                                 [32]byte
}

var ErrRoRecoveryRecord = errors.New("invalid or unsupported Ro recovery record")

func validRoLimits(l RoRecoveryLimits) bool {
	return l.MaxRecordBytes >= 1024 && l.MaxRecordBytes <= 4<<20 && l.MaxTargets > 0 && l.MaxTargets <= 16 &&
		l.MaxTotalAttempts > 0 && l.MaxTotalAttempts <= 256 && l.MaxSegments > 0 && l.MaxSegments <= l.MaxTotalAttempts &&
		l.MaxAttemptsPerSegment > 0 && l.MaxAttemptsPerSegment <= 16 && l.MaxAttemptsPerSegment <= l.MaxTotalAttempts &&
		l.MaxReplayAge > 0 && l.MaxReplayAge <= 365*24*time.Hour && l.MaxResponseWait > 0 && l.MaxResponseWait <= l.MaxReplayAge
}
func withinRoLimits(l, max RoRecoveryLimits) bool {
	return validRoLimits(l) && l.MaxRecordBytes <= max.MaxRecordBytes && l.MaxTargets <= max.MaxTargets &&
		l.MaxTotalAttempts <= max.MaxTotalAttempts && l.MaxSegments <= max.MaxSegments && l.MaxAttemptsPerSegment <= max.MaxAttemptsPerSegment &&
		l.MaxReplayAge <= max.MaxReplayAge && l.MaxResponseWait <= max.MaxResponseWait
}

// CheckIntegrity checks bounded record structure and its checksum without a Node,
// allocating a decoded body, or granting send/import authority. Segment admission
// additionally validates the current profile, all bindings, age,
// identifier conflicts, application progress and original-request rules.
func (r RoRecoveryRequest) CheckIntegrity() error {
	if r.Version != 2 || r.RecordID == ([16]byte{}) || len(r.WireFingerprint) != 64 || len(r.ContextName) > 64 ||
		len(r.SessionEvidenceID) > 128 || len(r.PolicyID) == 0 || len(r.PolicyID) > 128 || !validRoLimits(r.Limits) || len(r.Targets) == 0 || len(r.Targets) > r.Limits.MaxTargets ||
		r.CreatedAt.IsZero() || r.CreatedAt.Year() < 1970 || r.CreatedAt.Year() > 9998 || !r.ReplayUntil.Equal(r.CreatedAt.Add(r.Limits.MaxReplayAge)) ||
		len(r.Frame) < 20 || len(r.Frame) > r.MaxFrameBytes || r.MaxFrameBytes > wire.MaxWireSize {
		return ErrRoRecoveryRecord
	}
	for i, t := range r.Targets {
		p := t.Policy
		if len(t.ReplayDomain) > 128 || len(t.EvidenceID) > 128 || len(t.SessionEvidenceID) > 128 || (t.ReplayDomain == "") != (t.EvidenceID == "") || len(t.TargetID) == 0 || len(t.TargetID) > 128 || t.Retention < r.Limits.MaxReplayAge || t.Retention > 365*24*time.Hour ||
			p.Command != (pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}) || p.ApplicationBinding != HeaderApplication ||
			len(p.Class) == 0 || len(p.Class) > 128 || p.Class != r.Targets[0].Policy.Class ||
			avp.Identity(p.DestinationHost) != nil || avp.Identity(p.DestinationRealm) != nil ||
			p.MaxRequestBytes < 20 || p.MaxRequestBytes > wire.MaxWireSize || p.MaxRequestBytes%4 != 0 || p.MaxAnswerBytes < 20 || p.MaxAnswerBytes > wire.MaxWireSize || p.MaxAnswerBytes%4 != 0 || p.Timeout <= 0 ||
			len(p.Sources.Terminals) != 1 || len(p.Sources.ErrorAgents) > 63 || !sameOrigin(p.Sources.Terminals[0], pdu.Origin{Host: p.DestinationHost, Realm: p.DestinationRealm}) {
			return ErrRoRecoveryRecord
		}
		if len(t.AllowedPeers) > 64 || t.ReplayDomain != "" && len(t.AllowedPeers) == 0 {
			return ErrRoRecoveryRecord
		}
		for j, id := range t.AllowedPeers {
			if len(id) == 0 || len(id) > 128 {
				return ErrRoRecoveryRecord
			}
			for _, old := range t.AllowedPeers[:j] {
				if old == id {
					return ErrRoRecoveryRecord
				}
			}
		}
		for _, o := range p.Sources.ErrorAgents {
			if avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
				return ErrRoRecoveryRecord
			}
		}
		for _, before := range r.Targets[:i] {
			if before.TargetID == t.TargetID || sameOrigin(before.Policy.Sources.Terminals[0], p.Sources.Terminals[0]) {
				return ErrRoRecoveryRecord
			}
		}
	}
	if roRecordMetadata(r)+r.MaxFrameBytes > r.Limits.MaxRecordBytes {
		return wire.ErrLimit
	}
	limits := wire.DefaultLimits()
	limits.MaxFrameBytes = wire.MaxWireSize
	h, size, err := wire.Header(r.Frame, limits)
	if err != nil || size != len(r.Frame) || h.Key() != (pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}) || h.Flags != pdu.RequestFlag|pdu.ProxiableFlag || h.HopByHop != 0 || roRecordDigest(r) != r.Digest {
		return ErrRoRecoveryRecord
	}
	return nil
}

// Metadata charge is a versioned conservative owned representation bound, not
// the size of an application's JSON/database encoding. Include the longest target
// frame even though the record stores only the initial target frame.
func roBindingMetadata(id string, p CommandPolicy) int {
	n := 512 + len(id) + len(p.Class) + len(p.DestinationHost) + len(p.DestinationRealm)
	for _, list := range [][]pdu.Origin{p.Sources.Terminals, p.Sources.ErrorAgents} {
		for _, v := range list {
			n += 64 + len(v.Host) + len(v.Realm)
		}
	}
	return n
}
func roRecordMetadata(r RoRecoveryRequest) int {
	n := 1024 + len(r.WireFingerprint) + len(r.ContextName) + len(r.PolicyID) + len(r.SessionEvidenceID)
	for _, t := range r.Targets {
		n += roBindingMetadata(t.TargetID, t.Policy) + roBindingAuthorityBytes(t)
	}
	return n
}

type roDigestWriter struct {
	h       hash.Hash
	scratch [8]byte
}

func (w *roDigestWriter) number(n uint64) {
	binary.BigEndian.PutUint64(w.scratch[:], n)
	_, _ = w.h.Write(w.scratch[:])
}
func (w *roDigestWriter) data(b []byte) { w.number(uint64(len(b))); _, _ = w.h.Write(b) }
func (w *roDigestWriter) str(s string)  { w.data([]byte(s)) }
func roRecordDigest(r RoRecoveryRequest) [32]byte {
	w := roDigestWriter{h: sha256.New()}
	w.str("diameterkit/ro-recovery-record/v2")
	w.number(uint64(r.Version))
	w.data(r.RecordID[:])
	w.str(r.WireFingerprint)
	w.str(r.ContextName)
	w.str(r.PolicyID)
	w.str(r.SessionEvidenceID)
	for _, t := range []time.Time{r.CreatedAt, r.ReplayUntil} {
		w.number(uint64(t.Unix()))
		w.number(uint64(t.Nanosecond()))
	}
	l := r.Limits
	for _, n := range []uint64{uint64(l.MaxRecordBytes), uint64(l.MaxTargets), uint64(l.MaxTotalAttempts), uint64(l.MaxSegments), uint64(l.MaxAttemptsPerSegment), uint64(l.MaxReplayAge), uint64(l.MaxResponseWait), uint64(r.MaxFrameBytes), uint64(len(r.Targets))} {
		w.number(n)
	}
	for _, t := range r.Targets {
		p := t.Policy
		w.str(t.TargetID)
		w.str(t.ReplayDomain)
		w.str(t.EvidenceID)
		w.str(t.SessionEvidenceID)
		w.number(uint64(len(t.AllowedPeers)))
		for _, id := range t.AllowedPeers {
			w.str(string(id))
		}
		w.number(uint64(t.Retention))
		w.number(uint64(p.ApplicationBinding))
		w.number(uint64(p.Command.ApplicationID))
		w.number(uint64(p.Command.Code))
		if p.Command.Request {
			w.number(1)
		} else {
			w.number(0)
		}
		w.str(string(p.Class))
		w.str(p.DestinationHost)
		w.str(p.DestinationRealm)
		w.number(uint64(p.MaxRequestBytes))
		w.number(uint64(p.MaxAnswerBytes))
		w.number(uint64(p.Timeout))
		for _, list := range [][]pdu.Origin{p.Sources.Terminals, p.Sources.ErrorAgents} {
			w.number(uint64(len(list)))
			for _, o := range list {
				w.str(o.Host)
				w.str(o.Realm)
			}
		}
	}
	w.data(r.Frame)
	var sum [32]byte
	copy(sum[:], w.h.Sum(nil))
	return sum
}

// PrepareRoRecoveryRequest freezes a validated original request under Node/class
// scratch and identifier quotas. It performs no network I/O, allocates no RequestID,
// pending or completion, and never calls storage. ctx controls preparation only.
func (n *Node) PrepareRoRecoveryRequest(ctx context.Context, body pdu.Body, opts RoRecoveryOptions) (RoRecoveryRequest, error) {
	var zero RoRecoveryRequest
	if len(opts.SessionEvidenceID) > 128 {
		return zero, ErrConfig
	}
	if !opts.Validation.Empty() {
		return zero, validation.ErrContext
	}
	if body == nil || body.CommandKey() != (pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}) {
		return zero, ErrUnsupported
	}
	op, ok := n.cfg.Bundle.Lookup(body.CommandKey())
	if !ok || !op.TransportReplay || !op.Proxiable || (!op.NoContext && op.CheckContext == nil) {
		return zero, ErrUnsupported
	}
	if !withinRoLimits(opts.Limits, n.cfg.RoRecovery) || len(opts.PolicyID) == 0 || len(opts.PolicyID) > 128 || len(opts.Targets) == 0 || len(opts.Targets) > opts.Limits.MaxTargets {
		return zero, ErrConfig
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	size, err := n.engine.MeasureBody(body)
	if err != nil {
		return zero, err
	}
	if size > opts.Limits.MaxRecordBytes {
		return zero, wire.ErrLimit
	}
	// Inspect bounded configured bindings before reserving scratch or cloning input.
	var class ClassID
	maxRequest := 0
	metadata := 1024 + len(n.cfg.Bundle.Fingerprint()) + len(op.ContextName) + len(opts.PolicyID) + len(opts.SessionEvidenceID)
	for i, t := range opts.Targets {
		p := n.peers[t.Peer]
		if p == nil || len(t.ID) == 0 || len(t.ID) > 128 || t.Retention < opts.Limits.MaxReplayAge || t.Retention > 365*24*time.Hour {
			return zero, ErrConfig
		}
		binding, found := n.roBinding(t, body.CommandKey())
		policy := binding.Policy
		if !found || policy.ApplicationBinding != HeaderApplication || policy.DestinationHost == "" || len(policy.Sources.Terminals) != 1 || !sameOrigin(policy.Sources.Terminals[0], pdu.Origin{Host: policy.DestinationHost, Realm: policy.DestinationRealm}) {
			return zero, ErrConfig
		}
		if i == 0 {
			class = policy.Class
		} else if policy.Class != class {
			return zero, ErrConfig
		}
		for _, before := range opts.Targets[:i] {
			old, _ := n.roBinding(before, body.CommandKey())
			previous := old.Policy
			if before.ID == t.ID || sameOrigin(previous.Sources.Terminals[0], policy.Sources.Terminals[0]) {
				return zero, ErrConfig
			}
		}
		metadata += roBindingMetadata(t.ID, policy) + roBindingAuthorityBytes(binding)
		maxRequest = max(maxRequest, policy.MaxRequestBytes)
	}
	if size > maxRequest || metadata+size > opts.Limits.MaxRecordBytes {
		return zero, wire.ErrLimit
	}
	qb, err := op.DecodeBound(maxRequest, n.cfg.Limits)
	if err != nil {
		return zero, err
	}
	extra := uint64(opts.Limits.MaxRecordBytes)*2 + 131072
	if qb > (math.MaxUint64-extra)/4 {
		return zero, ErrBackpressure
	}
	n.mu.Lock()
	if n.closed || n.shutting {
		n.mu.Unlock()
		return zero, ErrClosed
	}
	charge, err := budget.Reserve(4*qb+extra, 1, "ro-record-preparation", &n.total, n.classes[class])
	if err != nil {
		n.mu.Unlock()
		return zero, ErrBackpressure
	}
	slot, err := n.roRecovery.reserve(class, time.Now())
	if err != nil {
		charge.Release()
		n.mu.Unlock()
		return zero, err
	}
	n.wg.Add(1)
	n.mu.Unlock()
	committed := false
	var lease *identifierLease
	defer func() {
		n.mu.Lock()
		if !committed {
			n.roRecovery.release(slot)
		}
		n.abandonIdentifier(lease)
		charge.Release()
		n.mu.Unlock()
		n.wg.Done()
		n.notify()
	}()
	now := time.Now()
	r := RoRecoveryRequest{Version: 2, SessionEvidenceID: avp.CloneString(opts.SessionEvidenceID), WireFingerprint: avp.CloneString(n.cfg.Bundle.Fingerprint()), ContextName: avp.CloneString(op.ContextName), PolicyID: avp.CloneString(opts.PolicyID), CreatedAt: now.UTC(), ReplayUntil: now.Add(opts.Limits.MaxReplayAge).UTC(), Limits: opts.Limits}
	for _, t := range opts.Targets {
		b, _ := n.roBinding(t, body.CommandKey())
		p := b.Policy
		p.Class = ClassID(avp.CloneString(string(p.Class)))
		p.DestinationHost = avp.CloneString(p.DestinationHost)
		p.DestinationRealm = avp.CloneString(p.DestinationRealm)
		p.Sources = cloneRoSources(p.Sources)
		b.AllowedPeers = append([]PeerID(nil), b.AllowedPeers...)
		b.TargetID = avp.CloneString(t.ID)
		b.Policy = p
		r.Targets = append(r.Targets, b)
	}
	first := r.Targets[0].Policy
	m, err := n.engine.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, DestinationHost: first.DestinationHost, DestinationRealm: first.DestinationRealm})
	if err != nil {
		return zero, err
	}
	r.Frame, err = n.engine.Append(nil, m)
	if err != nil {
		return zero, err
	}
	for _, t := range r.Targets {
		frame, e := roTargetFrame(r.Frame, pdu.Origin{Host: t.Policy.DestinationHost, Realm: t.Policy.DestinationRealm}, n.cfg.Limits, min(t.Policy.MaxRequestBytes, r.Limits.MaxRecordBytes-metadata))
		if e != nil {
			return zero, e
		}
		msg, e := n.engine.Decode(frame)
		if e != nil {
			return zero, e
		}
		snapshot, e := n.engine.PrepareRequest(msg, t.Policy.Sources)
		if e != nil {
			return zero, e
		}
		if e = snapshot.BindContext(opts.Validation); e != nil {
			return zero, e
		}
		r.MaxFrameBytes = max(r.MaxFrameBytes, len(frame))
	}
	if _, err = rand.Read(r.RecordID[:]); err != nil {
		return zero, err
	}
	n.mu.Lock()
	err = ctx.Err()
	if err == nil && (n.closed || n.shutting) {
		err = ErrClosed
	}
	if err == nil && !time.Now().Before(now.Add(r.Limits.MaxReplayAge)) {
		err = context.DeadlineExceeded
	}
	var e2e uint32
	if err == nil {
		if n.identifiers.custom {
			lease, err = n.identifiers.reserve(class, time.Now())
		} else {
			e2e, err = n.nextE2E(false)
			if err == nil {
				lease = &identifierLease{slot: -1}
			}
		}
	}
	n.mu.Unlock()
	if err != nil {
		return zero, err
	}
	if n.identifiers.custom {
		candidate, sourceErr := n.identifiers.config.EndToEnd.NextEndToEnd()
		n.mu.Lock()
		n.identifiers.sourceBusy = false
		switch {
		case sourceErr != nil:
			err = errors.Join(ErrIdentifiers, ErrIdentifierSourceFailed, sourceErr)
		case ctx.Err() != nil:
			err = ctx.Err()
		case n.closed || n.shutting:
			err = ErrClosed
		default:
			err = n.installApplicationIdentifier(lease, candidate)
			e2e = candidate
		}
		n.mu.Unlock()
		if err != nil {
			return zero, err
		}
	}
	binary.BigEndian.PutUint32(r.Frame[16:20], e2e)
	r.Digest = roRecordDigest(r)
	if err = r.CheckIntegrity(); err != nil {
		return zero, err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err = ctx.Err(); err != nil {
		return zero, err
	}
	if n.closed || n.shutting {
		return zero, ErrClosed
	}
	if !time.Now().Before(now.Add(r.Limits.MaxReplayAge)) {
		return zero, context.DeadlineExceeded
	}
	n.roRecovery.commit(slot, e2e, r.RecordID, r.Digest, now.Add(r.Limits.MaxReplayAge).Add(n.identifiers.config.ReuseGuard))
	n.roRecovery.slots[slot].replayUntil = now.Add(r.Limits.MaxReplayAge)
	committed = true
	return r, nil
}

func cloneRoSources(s pdu.AnswerSources) pdu.AnswerSources {
	clone := func(in []pdu.Origin) []pdu.Origin {
		out := make([]pdu.Origin, len(in))
		for i, o := range in {
			out[i] = pdu.Origin{Host: avp.CloneString(o.Host), Realm: avp.CloneString(o.Realm)}
		}
		return out
	}
	return pdu.AnswerSources{Terminals: clone(s.Terminals), ErrorAgents: clone(s.ErrorAgents)}
}

// Replace only the two top-level destinations, preserving every other AVP byte,
// ordering, flag and private payload. No decoded/reencoded application projection.
// The input is a validated locally built CCR. Old per-hop state is prohibited.
func roTargetFrame(frame []byte, target pdu.Origin, limits wire.Limits, maxBytes int) ([]byte, error) {
	if len(frame) < 20 {
		return nil, wire.ErrFrame
	}
	size := 20
	host, realm := 0, 0
	d := wire.NewDecoder(frame[20:], limits)
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		rawSize := (8 + len(v.Data) + 3) &^ 3
		if v.Key.VendorID != 0 {
			rawSize += 4
		}
		if v.Key.VendorID == 0 {
			switch v.Key.Code {
			case 282, 284:
				return nil, validation.ErrEnvelope
			case 293:
				host++
				rawSize = (8 + len(target.Host) + 3) &^ 3
			case 283:
				realm++
				rawSize = (8 + len(target.Realm) + 3) &^ 3
			}
		}
		size += rawSize
	}
	if host != 1 || realm != 1 {
		return nil, validation.ErrEnvelope
	}
	if size > maxBytes || size > limits.MaxFrameBytes {
		return nil, wire.ErrLimit
	}
	out := make([]byte, size)
	copy(out, frame[:20])
	wire.Put24(out[1:4], uint32(size))
	offset := 20
	d = wire.NewDecoder(frame[20:], limits)
	for {
		v, ok, _ := d.Next()
		if !ok {
			break
		}
		rawSize := (8 + len(v.Data) + 3) &^ 3
		if v.Key.VendorID != 0 {
			rawSize += 4
		}
		value := ""
		if v.Key.VendorID == 0 && v.Key.Code == 293 {
			value = target.Host
		}
		if v.Key.VendorID == 0 && v.Key.Code == 283 {
			value = target.Realm
		}
		if value != "" {
			copy(out[offset:offset+8], frame[20+v.Offset:20+v.Offset+8])
			wire.Put24(out[offset+5:offset+8], uint32(8+len(value)))
			copy(out[offset+8:], value)
			offset += (8 + len(value) + 3) &^ 3
		} else {
			copy(out[offset:], frame[20+v.Offset:20+v.Offset+rawSize])
			offset += rawSize
		}
	}
	return out, nil
}

func roBindingAuthorityBytes(b RoRecoveryBinding) int {
	n := len(b.ReplayDomain) + len(b.EvidenceID) + len(b.SessionEvidenceID)
	for _, id := range b.AllowedPeers {
		n += 64 + len(id)
	}
	return n
}
