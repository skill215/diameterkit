// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

func importConfig(t *testing.T, family string) NodeConfig {
	c := recordConfig(t, family)
	c.AllowRoRecoveryImport = true
	return c
}
func importDeclaration(c RoRecoveryCheckpoint) *RoRecoveryImport {
	return &RoRecoveryImport{Checkpoint: c, CheckedAt: time.Now().UTC(), SenderFenced: true, ProgressConfirmed: true, ClockConfirmed: true}
}
func preparedCheckpoint(r RoRecoveryRequest) RoRecoveryCheckpoint {
	return RoRecoveryCheckpoint{RecordID: r.RecordID, NextTrigger: RoRecoveryInitial}
}
func importProgress(c RoRecoveryCheckpoint) RoRecoveryProgress {
	return RoRecoveryProgress{SegmentNo: c.SegmentNo + 1, ReservedBefore: c.ReservedThrough, SegmentAllowance: 1, ReservedThrough: c.ReservedThrough + 1, Delivery: c.Delivery}
}
func importPlan(c RoRecoveryCheckpoint) RoRecoveryPlan {
	p := segmentPlan(1)
	p.Trigger = c.NextTrigger
	p.Import = importDeclaration(c)
	return p
}
func oldImportRecord(t *testing.T) RoRecoveryRequest {
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	r := prepareRecord(t, n, roBody(), recordOptions())
	n.Close()
	if e := n.Wait(ctxTest(t)); e != nil {
		t.Fatal(e)
	}
	return copyRecord(t, r)
}
func changeImportID(r RoRecoveryRequest, id uint32) RoRecoveryRequest {
	r.Frame = append([]byte(nil), r.Frame...)
	binary.BigEndian.PutUint32(r.Frame[16:20], id)
	r.Digest = roRecordDigest(r)
	return r
}

func TestRoImportErrorCheckpointAndCrossTarget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		old, co := newInboundNode(t, crossConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, old, co, "dra", tls)
		port := portTest(t, old, 1)
		r := prepareRecord(t, old, roBody(), crossOptions())
		plan := errorPlan(RoErrorSMSFWBeforeTx)
		plan.MaxAttempts = 1
		id := submitSegment(t, old, r, plan, segmentProgress(1), port)
		req := routeRead(t, co, remote)
		crossReply(t, co, remote, req, serverOrigin, 5012)
		got := routeCompletion(t, old, port, id)
		cp := got.RoRecovery.Checkpoint()
		if cp.ErrorResultCode != 5012 || cp.NextTrigger != RoRecoveryErrorAnswer || cp.ReplayPending {
			t.Fatal(cp)
		}
		old.Close()
		if e := old.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		cfg := crossConfig(t, "standard")
		cfg.AllowRoRecoveryImport = true
		n, co := newInboundNode(t, cfg)
		_, remote = roRecoveryNetwork(t, n, co, "dra", tls)
		port = portTest(t, n, 1)
		plan = importPlan(cp)
		plan.TargetID = "ocs-b"
		id = submitSegment(t, n, copyRecord(t, r), plan, importProgress(cp), port)
		req = routeRead(t, co, remote)
		crossIdentity(t, n, co, r, req)
		if req.Header.Flags&pdu.RetransmittedFlag != 0 {
			t.Fatal("error checkpoint lost T=0 lineage")
		}
		crossReply(t, co, remote, req, crossOrigin, 2001)
		got = routeCompletion(t, n, port, id)
		if got.Err != nil || got.Attempts[0].TargetID != "ocs-b" || got.RoRecovery.CanContinue {
			t.Fatal(got)
		}
		cfg.RoTargets[1].EvidenceID = "different-deployment"
		drift, _ := newInboundNode(t, cfg)
		if _, e := drift.SubmitRoRecovery(ctxTest(t), r, plan, importProgress(cp), portTest(t, drift, 1)); e == nil {
			t.Fatal("import accepted changed authority")
		}
	})
}

func TestRoImportClassQuotaRetainsCompletedGuards(t *testing.T) {
	r := oldImportRecord(t)
	cfg := importConfig(t, "standard")
	cfg.Resources.RoRecoveryClasses[0].Slots = 1
	n, co := newInboundNode(t, cfg)
	_, remote := roRecoveryNetwork(t, n, co, "dra", false)
	port := portTest(t, n, 1)
	cp := preparedCheckpoint(r)
	id := submitSegment(t, n, r, importPlan(cp), importProgress(cp), port)
	req := routeRead(t, co, remote)
	roRecoveryReply(t, co, remote, req, 2001)
	routeCompletion(t, n, port, id)
	other := changeImportID(r, binary.BigEndian.Uint32(r.Frame[16:20])^1)
	other.RecordID[0] ^= 1
	other.Digest = roRecordDigest(other)
	cp = preparedCheckpoint(other)
	if _, e := n.SubmitRoRecovery(ctxTest(t), other, importPlan(cp), importProgress(cp), port); !errors.Is(e, ErrIdentifiers) {
		t.Fatal(e)
	}
	if n.RoRecoveryInfo().Held != 1 {
		t.Fatal("completed guard evicted")
	}
}
func TestRoImportRestartProfilesAndOriginalIdentity(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, family := range []string{"standard", "sms", "money", "combined"} {
			kinds := []string{"iec", "initial", "term"}
			if family == "money" {
				kinds = []string{"general", "balance", "refund"}
			}
			if family == "combined" {
				kinds = []string{"iec", "initial", "term", "general", "balance", "refund"}
			}
			for _, kind := range kinds {
				old, co := newInboundNode(t, recordConfig(t, family))
				_, remote := roRecoveryNetwork(t, old, co, "dra", tls)
				port := portTest(t, old, 1)
				r := prepareRecord(t, old, roRecoveryBody(t, family, kind), recordOptions())
				id := submitSegment(t, old, r, segmentPlan(1), segmentProgress(1), port)
				first := routeRead(t, co, remote)
				old.Cancel(id)
				result := routeCompletion(t, old, port, id)
				checkpoint := result.RoRecovery.Checkpoint()
				old.Close()
				if e := old.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				r = copyRecord(t, r)
				n, co := newInboundNode(t, importConfig(t, family))
				_, remote = roRecoveryNetwork(t, n, co, "dra", tls)
				port = portTest(t, n, 1)
				p := importPlan(checkpoint)
				next := submitSegment(t, n, r, p, importProgress(checkpoint), port)
				req := routeRead(t, co, remote)
				encoded, e := co.Append(nil, req)
				if e != nil {
					t.Fatal(e)
				}
				if next.Boot == id.Boot || req.Header.EndToEnd != first.Header.EndToEnd || req.Header.Flags&pdu.RetransmittedFlag == 0 || !bytes.Equal(encoded[16:], r.Frame[16:]) {
					t.Fatal("restart identity")
				}
				again, e := n.SubmitRoRecovery(ctxTest(t), r, p, importProgress(checkpoint), port)
				if e != nil || again != next {
					t.Fatal(e)
				}
				changed := p
				v := *p.Import
				v.CheckedAt = v.CheckedAt.Add(-time.Nanosecond)
				changed.Import = &v
				if _, e = n.SubmitRoRecovery(ctxTest(t), r, changed, importProgress(checkpoint), port); !errors.Is(e, ErrRoRecoveryConflict) {
					t.Fatal(e)
				}
				roRecoveryReply(t, co, remote, req, 2001)
				got := routeCompletion(t, n, port, next)
				if got.RoRecovery.SegmentNo != 2 || got.RoRecovery.ReservedThrough != 2 || got.RoRecovery.CanContinue {
					t.Fatal(got)
				}
				p.Import = importDeclaration(checkpoint)
				if _, e = n.SubmitRoRecovery(ctxTest(t), r, p, importProgress(checkpoint), port); !errors.Is(e, ErrRoRecoveryFinal) {
					t.Fatal("terminal resurrected", e)
				}
				if n.IdentifierInfo().ApplicationAllocated != 0 {
					t.Fatal("import allocated new E2E")
				}
				segmentIdle(t, n)
			}
		}
	})
}
func TestRoImportPreparedUnknownAndErrorLineage(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"prepared", "unknown", "error", "after-error-not-written"} {
			old, _ := newInboundNode(t, recordConfig(t, "standard"))
			opts := recordOptions()
			opts.PolicyID = RoSMSFWRecoveryPolicy
			r := prepareRecord(t, old, roBody(), opts)
			old.Close()
			old.Wait(ctxTest(t))
			c := preparedCheckpoint(r)
			if kind != "prepared" {
				c.SegmentNo = 1
				c.ReservedThrough = 2
				c.Delivery = DeliveryEvidence{Bytes: len(r.Frame), FrameBytes: len(r.Frame), MayHaveReachedPeer: true}
				c.NextTrigger = RoRecoveryNoAnswer
			}
			switch kind {
			case "unknown":
				c.Unknown = true
				c.ReplayPending = true
				c.Delivery.Uncertain = true
			case "error":
				c.NextTrigger = RoRecoveryErrorAnswer
				c.ErrorResultCode = 5012
			}
			n, co := newInboundNode(t, importConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			p := importPlan(c)
			id := submitSegment(t, n, r, p, importProgress(c), port)
			req := routeRead(t, co, remote)
			if (req.Header.Flags&pdu.RetransmittedFlag != 0) != (kind == "unknown") {
				t.Fatal("import T lineage", kind)
			}
			n.Cancel(id)
			got := routeCompletion(t, n, port, id)
			if !got.RoRecovery.ReplayPending || !got.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
				t.Fatal(got)
			}
			p.Import = nil
			p.Trigger = got.RoRecovery.NextTrigger
			id = submitSegment(t, n, r, p, nextSegmentProgress(got, 1), port)
			req = routeRead(t, co, remote)
			if req.Header.Flags&pdu.RetransmittedFlag == 0 {
				t.Fatal("later no-answer lost T")
			}
			roRecoveryReply(t, co, remote, req, 2001)
			routeCompletion(t, n, port, id)
		}
	})
}
func TestRoImportAttestationTimeAndProgressRefusals(t *testing.T) {
	r := oldImportRecord(t)
	for _, change := range []func(*RoRecoveryImport){func(v *RoRecoveryImport) { v.SenderFenced = false }, func(v *RoRecoveryImport) { v.ProgressConfirmed = false }, func(v *RoRecoveryImport) { v.ClockConfirmed = false }, func(v *RoRecoveryImport) { v.CheckedAt = time.Now().Add(time.Hour) }, func(v *RoRecoveryImport) { v.CheckedAt = r.CreatedAt.Add(-time.Second) }, func(v *RoRecoveryImport) { v.Checkpoint.RecordID[0] ^= 1 }, func(v *RoRecoveryImport) { v.Checkpoint.ReservedThrough = 8 }, func(v *RoRecoveryImport) { v.Checkpoint.NextTrigger = 0 }, func(v *RoRecoveryImport) { v.Checkpoint.ReplayPending = true }, func(v *RoRecoveryImport) { v.Checkpoint.Unknown = true }} {
		n, _ := newInboundNode(t, importConfig(t, "standard"))
		port := portTest(t, n, 1)
		c := preparedCheckpoint(r)
		p := importPlan(c)
		change(p.Import)
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, importProgress(c), port); e == nil {
			t.Fatal("invalid import accepted")
		}
		if n.RoRecoveryInfo().Held != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal("failed import retained quota")
		}
	}
	for _, kind := range []string{"disabled", "origin", "past", "future", "binding"} {
		cfg := importConfig(t, "standard")
		bad := copyRecord(t, r)
		switch kind {
		case "disabled":
			cfg.AllowRoRecoveryImport = false
		case "origin":
			cfg.Origin.Host = "new-owner.example"
		case "past":
			bad.CreatedAt = time.Now().Add(-2 * time.Minute).UTC()
			bad.ReplayUntil = bad.CreatedAt.Add(bad.Limits.MaxReplayAge)
			bad.Digest = roRecordDigest(bad)
		case "future":
			bad.CreatedAt = time.Now().Add(time.Minute).UTC()
			bad.ReplayUntil = bad.CreatedAt.Add(bad.Limits.MaxReplayAge)
			bad.Digest = roRecordDigest(bad)
		case "binding":
			cfg.Peers[0].Commands[0].Timeout -= time.Millisecond
			cfg.Peers[1].Commands[0].Timeout -= time.Millisecond
		}
		n, co := newInboundNode(t, cfg)
		_, _ = roRecoveryNetwork(t, n, co, "dra", false)
		port := portTest(t, n, 1)
		c := preparedCheckpoint(bad)
		p := importPlan(c)
		if _, e := n.SubmitRoRecovery(ctxTest(t), bad, p, importProgress(c), port); e == nil {
			t.Fatal(kind)
		}
		if n.RoRecoveryInfo().Held != 0 {
			t.Fatal("bad record held", kind)
		}
	}
}
func TestRoImportBuiltInUsedRangeAndWrap(t *testing.T) {
	for _, seed := range []uint32{0, 6, 0x80000006} {
		for _, candidate := range []uint32{0, 1, 2, 6, 7, 8, 15, 16} {
			cfg := importConfig(t, "standard")
			cfg.Identifiers.Space = SequenceSpace{CounterBits: 4}
			n, _ := newInboundNode(t, cfg)
			n.mu.Lock()
			n.appSeed = seed
			n.appCount = 3
			e := n.roImportIdentifier(candidate, time.Now())
			n.mu.Unlock()
			used := false
			for j := uint32(0); j < 3; j++ {
				if (seed+j)&7 == candidate {
					used = true
				}
			}
			want := candidate < 8 && !used
			if (e == nil) != want {
				t.Fatalf("seed%d candidate%d: %v", seed, candidate, e)
			}
		}
	}
	eachNetwork(t, func(t *testing.T, tls bool) {
		r := changeImportID(oldImportRecord(t), 7)
		cfg := importConfig(t, "standard")
		cfg.Identifiers.Space = SequenceSpace{CounterBits: 4}
		n, co := newInboundNode(t, cfg)
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		n.mu.Lock()
		n.appSeed = 6
		n.appCount = 0
		n.mu.Unlock()
		c := preparedCheckpoint(r)
		id := submitSegment(t, n, r, importPlan(c), importProgress(c), port)
		req := routeRead(t, co, remote)
		n.mu.Lock()
		v, e := n.nextE2E(false)
		if e == nil {
			n.identifiers.appActive--
			n.identifiers.appRetired = time.Now()
		}
		_, collision := n.nextE2E(false)
		n.mu.Unlock()
		if v != 6 || e != nil || !errors.Is(collision, ErrIdentifierCollision) {
			t.Fatal(v, e, collision)
		}
		roRecoveryReply(t, co, remote, req, 2001)
		routeCompletion(t, n, port, id)
	})
}
func TestRoImportRecoverySpacesAndOwnedInfo(t *testing.T) {
	for _, spaces := range [][]SequenceSpace{{{Prefix: 0, CounterBits: 4}}, {{Prefix: 1, CounterBits: 4}, {Prefix: 2, CounterBits: 3}}, {{CounterBits: 32}}, {{Prefix: 1, CounterBits: 0}}} {
		cfg := importConfig(t, "standard")
		cfg.Identifiers.Space = SequenceSpace{CounterBits: 4}
		cfg.Identifiers.RecoverySpaces = spaces
		cfg.Resources.IdentifierBytes += 512
		if n, e := NewNode(cfg); e == nil {
			n.Close()
			t.Fatal("overlap or invalid recovery space accepted")
		}
	}
	eachNetwork(t, func(t *testing.T, tls bool) {
		r := changeImportID(oldImportRecord(t), 23)
		cfg := importConfig(t, "standard")
		cfg.Identifiers.Space = SequenceSpace{CounterBits: 4}
		cfg.Identifiers.RecoverySpaces = []SequenceSpace{{Prefix: 1, CounterBits: 4}}
		cfg.Resources.IdentifierBytes += 32
		n, co := newInboundNode(t, cfg)
		cfg.Identifiers.RecoverySpaces[0].Prefix = 2
		info := n.IdentifierInfo()
		info.RecoverySpaces[0].Prefix = 3
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		c := preparedCheckpoint(r)
		id := submitSegment(t, n, r, importPlan(c), importProgress(c), port)
		req := routeRead(t, co, remote)
		if req.Header.EndToEnd != 23 {
			t.Fatal(req.Header)
		}
		roRecoveryReply(t, co, remote, req, 2001)
		routeCompletion(t, n, port, id)
		for _, id := range []uint32{8, 15, 32} {
			n.mu.Lock()
			e := n.roImportIdentifier(id, time.Now())
			n.mu.Unlock()
			if !errors.Is(e, ErrIdentifierRange) {
				t.Fatal(id, e)
			}
		}
	})
}
func TestRoImportCustomSourceCollisionAndNoInvocation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		var calls atomic.Int32
		cfg := customIdentifiers(importConfig(t, "standard"), idSource(func() (uint32, error) { calls.Add(1); return 91, nil }), 4)
		cfg.Resources.IdentifierBytes += 4*512 + 256
		n, co := newInboundNode(t, cfg)
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := changeImportID(oldImportRecord(t), 91)
		c := preparedCheckpoint(r)
		n.mu.Lock()
		lease, e := n.identifiers.reserve("sms", time.Now())
		if e == nil {
			n.identifiers.sourceBusy = false
			e = n.identifiers.install(lease, 91)
		}
		n.mu.Unlock()
		if e != nil {
			t.Fatal(e)
		}
		for _, quarantine := range []bool{false, true} {
			if quarantine {
				n.mu.Lock()
				n.retireIdentifier(lease)
				n.mu.Unlock()
			}
			if _, e = n.SubmitRoRecovery(ctxTest(t), r, importPlan(c), importProgress(c), port); !errors.Is(e, ErrIdentifierCollision) {
				t.Fatal(e)
			}
		}
		n.mu.Lock()
		n.identifiers.slots[lease.slot].reuseAfter = time.Now().Add(-time.Second)
		n.identifiers.expire(time.Now())
		n.mu.Unlock()
		id := submitSegment(t, n, r, importPlan(c), importProgress(c), port)
		req := routeRead(t, co, remote)
		if calls.Load() != 0 {
			t.Fatal("import called E2E source")
		}
		roRecoveryReply(t, co, remote, req, 2001)
		routeCompletion(t, n, port, id)
	})
}
func TestRoImportFailedAdmissionRollbackAndDuplicateRecord(t *testing.T) {
	r := oldImportRecord(t)
	n, co := newInboundNode(t, importConfig(t, "standard"))
	port := portTest(t, n, 2)
	c := preparedCheckpoint(r)
	p := importPlan(c)
	if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, importProgress(c), port); !errors.Is(e, ErrNoEligibleAlternate) {
		t.Fatal(e)
	}
	if n.RoRecoveryInfo().Held != 0 {
		t.Fatal("no-path import leaked")
	}
	_, remote := roRecoveryNetwork(t, n, co, "dra", false)
	n.mu.Lock()
	limit := port.account.LimitBytes
	port.account.LimitBytes = 1
	n.mu.Unlock()
	if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, importProgress(c), port); !errors.Is(e, ErrBackpressure) {
		t.Fatal(e)
	}
	n.mu.Lock()
	port.account.LimitBytes = limit
	n.mu.Unlock()
	if n.RoRecoveryInfo().Held != 0 {
		t.Fatal("budget refusal leaked")
	}
	id := submitSegment(t, n, r, p, importProgress(c), port)
	req := routeRead(t, co, remote)
	other := changeImportID(r, req.Header.EndToEnd^1)
	if _, e := n.SubmitRoRecovery(ctxTest(t), other, importPlan(c), importProgress(c), port); !errors.Is(e, ErrRoRecoveryConflict) {
		t.Fatal("RecordID mapped to another E2E", e)
	}
	roRecoveryReply(t, co, remote, req, 2001)
	routeCompletion(t, n, port, id)
}
func TestRoImportCloseWithBorrowersAndGuardRetention(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		r := oldImportRecord(t)
		n, co := newInboundNode(t, importConfig(t, "standard"))
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, b)
		port := portTest(t, n, 1)
		c := preparedCheckpoint(r)
		id := submitSegment(t, n, r, importPlan(c), importProgress(c), port)
		routeRead(t, co, b)
		<-barrier.entered
		n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		e := n.Wait(ctx)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) || n.Resources().OutboundBytes == 0 {
			t.Fatal("early import retirement", e)
		}
		once.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.RoRecovery.CanContinue {
			t.Fatal(got)
		}
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
	})
}

func TestRoImportPreparationCloseRollback(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := importConfig(t, "standard")
		var armed atomic.Bool
		entered, release := make(chan struct{}), make(chan struct{})
		var once sync.Once
		defer once.Do(func() { close(release) })
		var ops []schema.Command
		for _, key := range []pdu.CommandKey{{Code: 257, Request: true}, {Code: 257}, {Code: 280, Request: true}, {Code: 280}, {Code: 282, Request: true}, {Code: 282}, {ApplicationID: 4, Code: 272, Request: true}, {ApplicationID: 4, Code: 272}} {
			op, _ := cfg.Bundle.Lookup(key)
			if key.ApplicationID == 4 && key.Request {
				clone := op.Clone
				op.Clone = func(b pdu.Body) (pdu.Body, error) {
					if armed.Load() {
						close(entered)
						<-release
					}
					return clone(b)
				}
			}
			ops = append(ops, op)
		}
		profile, e := schema.NewProfile("test-segment-barrier", "1", strings.Repeat("0", 64), ops...)
		if e != nil {
			t.Fatal(e)
		}
		cfg.Bundle, e = schema.Compile(profile)
		if e != nil {
			t.Fatal(e)
		}
		old, _ := newInboundNode(t, cfg)
		r := prepareRecord(t, old, roBody(), recordOptions())
		old.Close()
		if e := old.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		n, co := newInboundNode(t, cfg)
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		_ = remote
		port := portTest(t, n, 1)
		plan := importPlan(preparedCheckpoint(r))
		armed.Store(true)
		done := make(chan error, 1)
		go func() { _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, segmentProgress(1), port); done <- e }()
		<-entered
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, plan, segmentProgress(1), port); !errors.Is(e, ErrRoRecoveryConflict) {
			t.Fatal(e)
		}
		n.Close()
		short, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		e = n.Wait(short)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("Wait skipped preparer", e)
		}
		once.Do(func() { close(release) })
		if e = <-done; !errors.Is(e, ErrClosed) {
			t.Fatal(e)
		}
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.seq != 0 || n.Resources().Outstanding != 0 || n.IdentifierInfo().ApplicationActive != 0 {
			t.Fatal(n.Resources())
		}
	})
}
