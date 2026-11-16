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
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func segmentPlan(attempts int) RoRecoveryPlan {
	return RoRecoveryPlan{TargetID: "ocs-a", Peers: []PeerID{"dra", "dra2"}, MaxAttempts: attempts, Trigger: RoRecoveryInitial}
}
func segmentProgress(allowance uint32) RoRecoveryProgress {
	return RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: allowance, ReservedThrough: allowance}
}
func nextSegmentProgress(r Result, allowance uint32) RoRecoveryProgress {
	p := r.RoRecovery
	return RoRecoveryProgress{SegmentNo: p.SegmentNo + 1, ReservedBefore: p.ReservedThrough, SegmentAllowance: allowance, ReservedThrough: p.ReservedThrough + allowance, Delivery: p.CumulativeDelivery}
}
func submitSegment(t *testing.T, n *Node, r RoRecoveryRequest, plan RoRecoveryPlan, p RoRecoveryProgress, port *CompletionPort) RequestID {
	t.Helper()
	id, e := n.SubmitRoRecovery(ctxTest(t), r, plan, p, port)
	if e != nil {
		t.Fatal(e)
	}
	return id
}
func segmentIdle(t *testing.T, n *Node) {
	t.Helper()
	waitFor(t, func() bool { return n.Resources().OutboundBytes == 0 && n.Resources().Pending == 0 })
}

func TestRoSegmentProfilesOriginalIdentityAndCompletion(t *testing.T) {
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
				t.Run(family+"/"+kind, func(t *testing.T) {
					n, co := newInboundNode(t, recordConfig(t, family))
					_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
					port := portTest(t, n, 2)
					r := prepareRecord(t, n, roRecoveryBody(t, family, kind), recordOptions())
					// Assert the use label actually selected its CC request type;
					// combinedBody's default is the general EVENT template.
					wantType := uint32(4)
					if kind == "initial" {
						wantType = 1
					}
					if kind == "term" {
						wantType = 3
					}
					d := wire.NewDecoder(r.Frame[20:], n.cfg.Limits)
					foundType := false
					for {
						v, ok, e := d.Next()
						if e != nil {
							t.Fatal(e)
						}
						if !ok {
							break
						}
						if v.Key.VendorID == 0 && v.Key.Code == 416 {
							value, e := wire.U32(v.Data)
							if e != nil || value != wantType {
								t.Fatal("wrong template for label", kind, value, e)
							}
							foundType = true
						}
					}
					if !foundType {
						t.Fatal("missing CC-Request-Type")
					}

					plan := segmentPlan(1)
					progress := segmentProgress(2)
					before := n.IdentifierInfo().ApplicationAllocated
					id := submitSegment(t, n, r, plan, progress, port)
					req := routeRead(t, co, remote)
					encoded, e := co.Append(nil, req)
					if e != nil || !bytes.Equal(r.Frame[16:], encoded[16:]) || req.Header.HopByHop == 0 || req.Header.Flags&pdu.RetransmittedFlag != 0 {
						t.Fatal("original identity changed", e)
					}
					if n.IdentifierInfo().ApplicationAllocated != before {
						t.Fatal("recovery allocated a new E2E")
					}
					info, e := n.RequestInfo(id)
					if e != nil || info.AttemptsUsed != 1 || info.ReplayEligible || info.Deadline.After(r.ReplayUntil) {
						t.Fatal(info, e)
					}
					canceled, cancel := context.WithCancel(context.Background())
					cancel()
					again, e := n.SubmitRoRecovery(canceled, r, plan, progress, port)
					if e != nil || again != id {
						t.Fatal("active duplicate", again, e)
					}
					// Caller records are detached from the live encoded request/snapshot.
					saved := copyRecord(t, r)
					r.Frame[len(r.Frame)-1] ^= 1
					r.Targets[0].Policy.Sources.Terminals[0].Host = "changed.example"
					roRecoveryReply(t, co, remote, req, 2001)
					got := routeCompletion(t, n, port, id)
					if got.Outcome != AnswerReceived || got.RoRecovery == nil || got.RoRecovery.RecordID != saved.RecordID || got.RoRecovery.ReservedThrough != 2 || got.RoRecovery.CanContinue || len(got.Attempts) != 1 {
						t.Fatal(got)
					}
					if _, e = n.SubmitRoRecovery(ctxTest(t), saved, plan, progress, port); !errors.Is(e, ErrRoSegmentAlreadyFinished) {
						t.Fatal(e)
					}
					plan.Trigger = RoRecoveryNoAnswer
					if _, e = n.SubmitRoRecovery(ctxTest(t), saved, plan, nextSegmentProgress(got, 1), port); !errors.Is(e, ErrRoRecoveryFinal) {
						t.Fatal("successful record revived", e)
					}
					segmentIdle(t, n)
				})
			}
		}
	})
}

func TestRoSegmentNoAnswerHighWaterAndOldAnswerIsolation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 2)
		r := prepareRecord(t, n, roBody(), recordOptions())
		plan := segmentPlan(1)
		first := submitSegment(t, n, r, plan, segmentProgress(3), port)
		old := routeRead(t, co, remote)
		n.Cancel(first)
		done := routeCompletion(t, n, port, first)
		if !done.RoRecovery.CanContinue || !done.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
			t.Fatal(done)
		}
		plan.Trigger = RoRecoveryNoAnswer
		p := nextSegmentProgress(done, 2)
		bad := p
		bad.ReservedBefore = 1
		bad.ReservedThrough = 3
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, bad, port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal("allowance refund", e)
		}
		bad = p
		bad.Delivery = DeliveryEvidence{}
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, bad, port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal("lost prior write", e)
		}
		second := submitSegment(t, n, r, plan, p, port)
		current := routeRead(t, co, remote)
		if current.Header.EndToEnd != old.Header.EndToEnd || current.Header.HopByHop == old.Header.HopByHop || current.Header.Flags&pdu.RetransmittedFlag == 0 {
			t.Fatal(old.Header, current.Header)
		}
		roRecoveryReply(t, co, remote, old, 2001)
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e := port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("old segment completed new one", e)
		}
		roRecoveryReply(t, co, remote, current, 4012)
		got := routeCompletion(t, n, port, second)
		if got.RoRecovery.SegmentNo != 2 || got.RoRecovery.ReservedThrough != 5 || got.RoRecovery.CanContinue || got.Outcome != AnswerReceived {
			t.Fatal(got)
		}
		segmentIdle(t, n)
	})
}

func TestRoSegmentCumulativeTAndExplicitTransportAttempts(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, fault := range []struct {
			bytes     int
			uncertain bool
		}{{0, false}, {8, false}, {0, true}} {
			t.Run(string(rune('a'+fault.bytes))+map[bool]string{false: "known", true: "unknown"}[fault.uncertain], func(t *testing.T) {
				n, co := newInboundNode(t, recordConfig(t, "combined"))
				a, b := networkAdapters(t, tls)
				roRecoveryConnect(t, n, co, "dra", &roRecoveryFault{Conn: a, bytes: fault.bytes, uncertain: fault.uncertain}, b)
				_, secondRemote := roRecoveryNetwork(t, n, co, "dra2", tls)
				port := portTest(t, n, 2)
				r := prepareRecord(t, n, combinedBody(t, "refund"), recordOptions())
				plan := segmentPlan(1)
				id := submitSegment(t, n, r, plan, segmentProgress(1), port)
				first := routeCompletion(t, n, port, id)
				if !first.RoRecovery.CanContinue {
					t.Fatal(first)
				}
				plan.Trigger = RoRecoveryNoAnswer
				plan.Peers = []PeerID{"dra2"}
				plan.MaxAttempts = 2
				id = submitSegment(t, n, r, plan, nextSegmentProgress(first, 2), port)
				req := routeRead(t, co, secondRemote)
				want := fault.bytes > 0 || fault.uncertain
				if (req.Header.Flags&pdu.RetransmittedFlag != 0) != want || req.Header.EndToEnd != binary.BigEndian.Uint32(r.Frame[16:20]) {
					t.Fatal(req.Header, want)
				}
				n.Cancel(id)
				final := routeCompletion(t, n, port, id)
				if !final.RoRecovery.CumulativeDelivery.MayHaveReachedPeer || final.RoRecovery.ReservedThrough != 3 {
					t.Fatal(final)
				}
			})
		}
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, a := roRecoveryNetwork(t, n, co, "dra", tls)
		_, b := roRecoveryNetwork(t, n, co, "dra2", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, segmentPlan(2), segmentProgress(2), port)
		first := routeRead(t, co, a)
		a.Close()
		info := roAwait(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, b)
		routeSameRequest(t, co, first, second)
		roRecoveryReply(t, co, b, second, 2001)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 || got.RoRecovery.ReservedThrough != 2 || got.SelectedAttempt.Number != 2 {
			t.Fatal(got)
		}
	})
}

func TestRoSegmentAdmissionRejectionsDoNotConsumeProgress(t *testing.T) {
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	port := portTest(t, n, 2)
	r := prepareRecord(t, n, roBody(), recordOptions())
	plan := segmentPlan(1)
	p := segmentProgress(1)
	for _, mutate := range []func(*RoRecoveryProgress){
		func(p *RoRecoveryProgress) { p.SegmentNo = 0 }, func(p *RoRecoveryProgress) { p.SegmentNo = 5 }, func(p *RoRecoveryProgress) { p.SegmentNo = 2 }, func(p *RoRecoveryProgress) { p.ReservedBefore = 1; p.ReservedThrough = 2 }, func(p *RoRecoveryProgress) { p.ReservedThrough = 2 }, func(p *RoRecoveryProgress) { p.SegmentAllowance = 0; p.ReservedThrough = 0 }, func(p *RoRecoveryProgress) {
			p.ReservedBefore = ^uint32(0)
			p.SegmentAllowance = 2
			p.ReservedThrough = 1
		}, func(p *RoRecoveryProgress) { p.Delivery.MayHaveReachedPeer = true },
	} {
		bad := p
		mutate(&bad)
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, bad, port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal(e)
		}
	}
	for _, mutate := range []func(*RoRecoveryPlan){
		func(p *RoRecoveryPlan) { p.Peers = nil }, func(p *RoRecoveryPlan) { p.Peers = []PeerID{"missing"} }, func(p *RoRecoveryPlan) { p.Peers = []PeerID{"dra", "dra"} }, func(p *RoRecoveryPlan) { p.TargetID = "other" }, func(p *RoRecoveryPlan) { p.MaxAttempts = 0 }, func(p *RoRecoveryPlan) { p.MaxAttempts = 5 }, func(p *RoRecoveryPlan) { p.Trigger = 0 },
	} {
		bad := plan
		mutate(&bad)
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, bad, p, port); e == nil {
			t.Fatal("bad plan")
		}
	}
	changed := copyRecord(t, r)
	changed.Frame[len(changed.Frame)-1] ^= 1
	changed.Digest = roRecordDigest(changed)
	if _, e := n.SubmitRoRecovery(ctxTest(t), changed, plan, p, port); !errors.Is(e, ErrRoRecoveryConflict) {
		t.Fatal("rehash changed original", e)
	}
	changed = copyRecord(t, r)
	changed.WireFingerprint = strings.Repeat("a", 64)
	changed.Digest = roRecordDigest(changed)
	if _, e := n.SubmitRoRecovery(ctxTest(t), changed, plan, p, port); !errors.Is(e, ErrRoRecoveryRecord) {
		t.Fatal(e)
	}
	foreign, _ := newInboundNode(t, recordConfig(t, "standard"))
	otherPort := portTest(t, foreign, 1)
	if _, e := foreign.SubmitRoRecovery(ctxTest(t), r, plan, p, otherPort); !errors.Is(e, ErrRoRecoveryNotPrepared) {
		t.Fatal("restart imported", e)
	}
	if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, p, otherPort); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := n.SubmitRoRecovery(ctx, r, plan, p, port); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, p, port); !errors.Is(e, ErrNoEligibleAlternate) {
		t.Fatal(e)
	}
	n.mu.Lock()
	guard := n.roRecovery.slots[n.roRecovery.byE2E[binary.BigEndian.Uint32(r.Frame[16:20])]]
	seq := n.seq
	n.mu.Unlock()
	if guard.segmentNo != 0 || guard.reservedThrough != 0 || guard.admitting || seq != 0 || n.Resources().Outstanding != 0 {
		t.Fatal(guard, n.Resources())
	}
}

func TestRoSegmentBudgetIdempotenceAndConcurrentAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 2)
		r := prepareRecord(t, n, roBody(), recordOptions())
		plan := segmentPlan(2)
		p := segmentProgress(2)
		small, e := n.NewCompletionPort(CompletionLimits{1, 1})
		if e != nil {
			t.Fatal(e)
		}
		defer small.Close(context.Background())
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, plan, p, small); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		var wg sync.WaitGroup
		ids := make(chan RequestID, 16)
		errs := make(chan error, 16)
		for range 16 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				id, e := n.SubmitRoRecovery(ctxTest(t), r, plan, p, port)
				if e != nil {
					errs <- e
				} else {
					ids <- id
				}
			}()
		}
		wg.Wait()
		close(ids)
		close(errs)
		var id RequestID
		count := 0
		for v := range ids {
			if count > 0 && id != v {
				t.Fatal("duplicate segment IDs")
			}
			id = v
			count++
		}
		if count == 0 {
			t.Fatal("no admission")
		}
		for e := range errs {
			if !errors.Is(e, ErrRoRecoveryBusy) {
				t.Fatal(e)
			}
		}
		req := routeRead(t, co, remote)
		n.mu.Lock()
		x := n.requests[id]
		capacity, bound := x.charge.Capacity, x.answerBound
		deadline := x.deadline
		seq := n.seq
		n.mu.Unlock()
		if capacity < 2*bound {
			t.Fatal("missing per-attempt A")
		}
		changed := plan
		changed.Peers = []PeerID{"dra2", "dra"}
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, changed, p, port); !errors.Is(e, ErrRoRecoveryConflict) {
			t.Fatal(e)
		}
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, plan, p, small); !errors.Is(e, ErrRoRecoveryConflict) {
			t.Fatal("different Port duplicate", e)
		}
		n.mu.Lock()
		same := n.requests[id].deadline.Equal(deadline) && n.seq == seq
		n.mu.Unlock()
		if !same {
			t.Fatal("duplicate changed deadline/ID")
		}
		roRecoveryReply(t, co, remote, req, 2001)
		routeCompletion(t, n, port, id)
		segmentIdle(t, n)
		// Actual total A admission failure, not just a zero-size Port.
		r2 := prepareRecord(t, n, roBody(), recordOptions())
		limited, e := n.NewCompletionPort(CompletionLimits{1, capacity - bound})
		if e != nil {
			t.Fatal(e)
		}
		defer limited.Close(context.Background())
		if _, e = n.SubmitRoRecovery(ctxTest(t), r2, plan, p, limited); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
	})
}

func TestRoSegmentWriterAndReaderRetirement(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, write := range []bool{true, false} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			a, b := networkAdapters(t, tls)
			barrier := &roRecoveryBarrier{Conn: a, write: write, entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			defer once.Do(func() { close(barrier.release) })
			roRecoveryConnect(t, n, co, "dra", barrier, b)
			port := portTest(t, n, 2)
			r := prepareRecord(t, n, roBody(), recordOptions())
			plan := segmentPlan(1)
			p := segmentProgress(1)
			id := submitSegment(t, n, r, plan, p, port)
			req := routeRead(t, co, b)
			if !write {
				roRecoveryReply(t, co, b, req, 2001)
			}
			<-barrier.entered
			n.Cancel(id)
			next := plan
			next.Trigger = RoRecoveryNoAnswer
			progress := RoRecoveryProgress{SegmentNo: 2, ReservedBefore: 1, SegmentAllowance: 1, ReservedThrough: 2}
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, next, progress, port); !errors.Is(e, ErrRoRecoveryBusy) {
				t.Fatal("new segment overlaps borrower", e)
			}
			n.mu.Lock()
			slot := n.roRecovery.byE2E[req.Header.EndToEnd]
			n.roRecovery.expire(n.roRecovery.slots[slot].keepUntil.Add(time.Hour))
			held := n.roRecovery.slots[slot].occupied
			n.mu.Unlock()
			if !held {
				t.Fatal("live guard expired")
			}
			quiet, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			_, e := port.Receive(quiet)
			cancel()
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("completion before borrower", e)
			}
			once.Do(func() { close(barrier.release) })
			got := routeCompletion(t, n, port, id)
			if !errors.Is(got.Err, context.Canceled) || got.RoRecovery == nil {
				t.Fatal(got)
			}
			segmentIdle(t, n)
		}
	})
}

func TestRoSegmentAllAnswersAreTerminal(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{2001, 4012, 5012, 3004} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 2)
			r := prepareRecord(t, n, roBody(), recordOptions())
			plan := segmentPlan(1)
			id := submitSegment(t, n, r, plan, segmentProgress(1), port)
			req := routeRead(t, co, remote)
			if code == 3004 {
				ans, e := co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: code}, serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
			} else {
				roRecoveryReply(t, co, remote, req, code)
			}
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.CanContinue {
				t.Fatal(got)
			}
			plan.Trigger = RoRecoveryNoAnswer
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, nextSegmentProgress(got, uint32(plan.MaxAttempts)), port); !errors.Is(e, ErrRoRecoveryFinal) {
				t.Fatal(e)
			}
		}
	})
}

func TestRoSegmentTotalAllowancesAndDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		plan := segmentPlan(1)
		p := segmentProgress(4)
		id := submitSegment(t, n, r, plan, p, port)
		routeRead(t, co, remote)
		n.Cancel(id)
		first := routeCompletion(t, n, port, id)
		plan.Trigger = RoRecoveryNoAnswer
		plan.Deadline = time.Now().Add(20 * time.Millisecond)
		second := submitSegment(t, n, r, plan, nextSegmentProgress(first, 4), port)
		routeRead(t, co, remote)
		done := routeCompletion(t, n, port, second)
		if !errors.Is(done.Err, context.DeadlineExceeded) || done.RoRecovery.ReservedThrough != 8 {
			t.Fatal(done)
		}
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, nextSegmentProgress(done, 1), port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal("total allowance reset", e)
		}
		// Enforce the internal monotonic recovery deadline as well as the saved UTC time.
		another := prepareRecord(t, n, roBody(), recordOptions())
		n.mu.Lock()
		slot := n.roRecovery.byE2E[binary.BigEndian.Uint32(another.Frame[16:20])]
		n.roRecovery.slots[slot].replayUntil = time.Now().Add(-time.Second)
		n.mu.Unlock()
		if _, e := n.SubmitRoRecovery(ctxTest(t), another, segmentPlan(1), segmentProgress(1), port); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
	})
}

func TestRoSegmentPreparationCloseRollback(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := recordConfig(t, "standard")
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
		n, co := newInboundNode(t, cfg)
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		_ = remote
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		armed.Store(true)
		done := make(chan error, 1)
		go func() { _, e := n.SubmitRoRecovery(ctxTest(t), r, segmentPlan(1), segmentProgress(1), port); done <- e }()
		<-entered
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, segmentPlan(1), segmentProgress(1), port); !errors.Is(e, ErrRoRecoveryBusy) {
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

func TestRoSegmentPriorWriteSurvivesZeroWriteAttempt(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		a, b := networkAdapters(t, tls)
		roRecoveryConnect(t, n, co, "dra2", &roRecoveryFault{Conn: a}, b)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, segmentPlan(1), segmentProgress(1), port)
		first := routeRead(t, co, remote)
		n.Cancel(id)
		done := routeCompletion(t, n, port, id)
		plan := segmentPlan(2)
		plan.Peers = []PeerID{"dra2", "dra"}
		plan.Trigger = RoRecoveryNoAnswer
		id = submitSegment(t, n, r, plan, nextSegmentProgress(done, 2), port)
		info := roAwait(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, remote)
		if second.Header.Flags&pdu.RetransmittedFlag == 0 || second.Header.EndToEnd != first.Header.EndToEnd {
			t.Fatal("zero-byte attempt erased older write", second.Header)
		}
		roRecoveryReply(t, co, remote, second, 2001)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 || got.Attempts[0].Delivery.MayHaveReachedPeer || !got.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
			t.Fatal(got)
		}
	})
}

func TestRoSegmentInvalidAnswerAndLifecycleTermination(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, end := range []string{"invalid-answer", "port-close", "node-close"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 2)
			r := prepareRecord(t, n, roBody(), recordOptions())
			plan := segmentPlan(2)
			id := submitSegment(t, n, r, plan, segmentProgress(2), port)
			req := routeRead(t, co, remote)
			switch end {
			case "invalid-answer":
				ans, e := co.CompleteAnswer(req, roAnswer(2001), pdu.Origin{Host: "foreign.example", Realm: "example"})
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.RoRecovery.CanContinue {
					t.Fatal(got)
				}
				plan.Trigger = RoRecoveryNoAnswer
				if _, e = n.SubmitRoRecovery(ctxTest(t), r, plan, nextSegmentProgress(got, uint32(plan.MaxAttempts)), port); !errors.Is(e, ErrRoRecoveryFinal) {
					t.Fatal(e)
				}
			case "port-close":
				if e := port.Close(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				other := portTest(t, n, 1)
				plan.Trigger = RoRecoveryNoAnswer
				n.mu.Lock()
				slot := n.roRecovery.byE2E[req.Header.EndToEnd]
				g := n.roRecovery.slots[slot]
				n.mu.Unlock()
				if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, RoRecoveryProgress{SegmentNo: 2, ReservedBefore: 2, SegmentAllowance: 2, ReservedThrough: 4, Delivery: g.cumulative}, other); !errors.Is(e, ErrRoRecoveryFinal) {
					t.Fatal(e)
				}
			case "node-close":
				n.Close()
				got := routeCompletion(t, n, port, id)
				if got.RoRecovery.CanContinue {
					t.Fatal(got)
				}
				if e := n.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, segmentProgress(2), port); !errors.Is(e, ErrClosed) {
					t.Fatal(e)
				}
			}
			segmentIdle(t, n)
		}
	})
}

func TestRoSegmentCustomSourceIsNotReentered(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		var calls atomic.Int32
		cfg := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) {
			if calls.Add(1) > 1 {
				t.Error("segment called source")
			}
			return 121, nil
		}), 4)
		cfg.Resources.IdentifierBytes += 4*512 + 256
		n, co := newInboundNode(t, cfg)
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, segmentPlan(1), segmentProgress(1), port)
		req := routeRead(t, co, remote)
		if req.Header.EndToEnd != 121 {
			t.Fatal(req.Header)
		}
		roRecoveryReply(t, co, remote, req, 2001)
		routeCompletion(t, n, port, id)
		if calls.Load() != 1 || n.RoRecoveryInfo().Held != 1 || n.IdentifierInfo().Active != 0 {
			t.Fatal(calls.Load(), n.IdentifierInfo())
		}
	})
}

func TestRoSegmentGuardHighWaterLocalAssertions(t *testing.T) {
	// Exercise conservative allowance gaps and binding drift with a test-only
	// previous high-water mark. Failed selection must not advance that mark.
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	port := portTest(t, n, 1)
	r := prepareRecord(t, n, roBody(), recordOptions())
	e2e := binary.BigEndian.Uint32(r.Frame[16:20])
	n.mu.Lock()
	slot := n.roRecovery.byE2E[e2e]
	g := &n.roRecovery.slots[slot]
	g.segmentNo = 1
	g.reservedThrough = 2
	g.continuable = true
	g.nextTrigger = RoRecoveryNoAnswer
	n.mu.Unlock()
	plan := segmentPlan(1)
	plan.Trigger = RoRecoveryNoAnswer
	p := RoRecoveryProgress{SegmentNo: 2, ReservedBefore: 3, SegmentAllowance: 1, ReservedThrough: 4}
	// Conservative skipped allowances are legal but still charged, even with no send.
	if _, e := n.SubmitRoRecovery(ctxTest(t), r, plan, p, port); !errors.Is(e, ErrNoEligibleAlternate) {
		t.Fatal(e)
	}
	n.mu.Lock()
	unchanged := n.roRecovery.slots[slot].reservedThrough == 2
	n.mu.Unlock()
	if !unchanged {
		t.Fatal("failed selection consumed progress")
	}
	changed := copyRecord(t, r)
	changed.Targets[0].Policy.Timeout += time.Nanosecond
	changed.Digest = roRecordDigest(changed)
	if _, e := n.SubmitRoRecovery(ctxTest(t), changed, plan, p, port); !errors.Is(e, ErrConfig) {
		t.Fatal("binding drift", e)
	}
}
