// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func trPlan(attempts int, wait time.Duration) RoRecoveryPlan {
	p := segmentPlan(attempts)
	p.ResponseWait = wait
	return p
}
func trInfo(t *testing.T, n *Node, id RequestID) RequestInfo {
	t.Helper()
	i, e := n.RequestInfo(id)
	if e != nil {
		t.Fatal(e)
	}
	return i
}
func trExpire(t *testing.T, n *Node, id RequestID) RequestInfo {
	t.Helper()
	i := trInfo(t, n, id)
	n.tick(i.ResponseWaitUntil)
	return trInfo(t, n, id)
}

func TestRoResponseWaitConfigurationAndIdempotence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 2)
		r := prepareRecord(t, n, roBody(), recordOptions())
		for _, wait := range []time.Duration{-1, r.Limits.MaxResponseWait + 1} {
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, trPlan(2, wait), segmentProgress(2), port); !errors.Is(e, ErrConfig) {
				t.Fatal(e)
			}
		}
		if n.Resources().Pending != 0 || n.seq != 0 {
			t.Fatal("invalid Tr admitted")
		}
		plan := trPlan(2, 100*time.Millisecond)
		id := submitSegment(t, n, r, plan, segmentProgress(2), port)
		req := routeRead(t, co, remote)
		i := trInfo(t, n, id)
		if i.AcceptedAt.IsZero() || !i.AttemptAcceptedAt.Equal(i.AcceptedAt) || i.ResponseWaitUntil.Sub(i.AcceptedAt) != plan.ResponseWait || i.ReplayReason != ReplayNone {
			t.Fatal(i)
		}
		changed := plan
		changed.ResponseWait++
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, changed, segmentProgress(2), port); !errors.Is(e, ErrRoRecoveryConflict) {
			t.Fatal("changed Tr duplicate", e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		again, e := n.SubmitRoRecovery(canceled, r, plan, segmentProgress(2), port)
		if e != nil || again != id {
			t.Fatal(e)
		}
		after := trInfo(t, n, id)
		if !after.AcceptedAt.Equal(i.AcceptedAt) || !after.ResponseWaitUntil.Equal(i.ResponseWaitUntil) {
			t.Fatal("duplicate extended Tr")
		}
		roRecoveryReply(t, co, remote, req, 2001)
		got := routeCompletion(t, n, port, id)
		if !got.Attempts[0].AcceptedAt.Equal(i.AcceptedAt) || !got.Attempts[0].ResponseWaitUntil.Equal(i.ResponseWaitUntil) {
			t.Fatal(got)
		}
	})
}

func TestRoResponseWaitProfilesHealthyConnectionAndLateAnswers(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, family := range []string{"standard", "sms", "money", "combined"} {
			for _, when := range []string{"awaiting", "successor"} {
				n, co := newInboundNode(t, recordConfig(t, family))
				aPeer, a := roRecoveryNetwork(t, n, co, "dra", tls)
				_, b := roRecoveryNetwork(t, n, co, "dra2", tls)
				port := portTest(t, n, 1)
				kind := "initial"
				if family == "money" {
					kind = "refund"
				}
				r := prepareRecord(t, n, roRecoveryBody(t, family, kind), recordOptions())
				id := submitSegment(t, n, r, trPlan(2, 30*time.Millisecond), segmentProgress(2), port)
				first := routeRead(t, co, a)
				before := trInfo(t, n, id)
				if _, e := n.Retransmit(ctxTest(t), id, before.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
					t.Fatal(e)
				}
				info := roAwait(t, n, id)
				if info.ReplayReason != ResponseWaitExpired || info.TransportStopReason != StopNone || !info.Deadline.Equal(before.Deadline) {
					t.Fatal(info)
				}
				n.mu.Lock()
				healthy := aPeer.readyLocked() && !aPeer.conn.closed && len(aPeer.conn.pending) == 1
				queued := len(n.peers["dra2"].conn.pending)
				held := n.requests[id].charge.Capacity
				n.mu.Unlock()
				if !healthy || queued != 0 || held == 0 {
					t.Fatal("Tr changed connection, freed budget or sent automatically")
				}
				if when == "successor" {
					next, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt)
					if e != nil {
						t.Fatal(e)
					}
					second := routeRead(t, co, b)
					routeSameRequest(t, co, first, second)
					current := trInfo(t, n, id)
					if !current.AcceptedAt.Equal(before.AcceptedAt) || !current.AttemptAcceptedAt.After(before.AcceptedAt) || current.ResponseWaitUntil.Sub(current.AttemptAcceptedAt) != 30*time.Millisecond {
						t.Fatal(current)
					}
					same, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt)
					if e != nil || same != next {
						t.Fatal(e)
					}
				}
				roRecoveryReply(t, co, a, first, 2001)
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || got.SelectedAttempt.Number != 1 || got.RoRecovery.CanContinue {
					t.Fatal(got)
				}
				segmentIdle(t, n)
			}
		}
	})
}

func TestRoResponseWaitSingleLastAttemptAndOverallDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"single", "last", "overall"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, a := roRecoveryNetwork(t, n, co, "dra", tls)
			_, b := roRecoveryNetwork(t, n, co, "dra2", tls)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), recordOptions())
			attempts := 1
			if kind == "last" {
				attempts = 2
			}
			plan := trPlan(attempts, 25*time.Millisecond)
			if kind == "overall" {
				plan.Deadline = time.Now().Add(20 * time.Millisecond)
				plan.ResponseWait = time.Second
			}
			id := submitSegment(t, n, r, plan, segmentProgress(uint32(attempts)), port)
			routeRead(t, co, a)
			if attempts == 2 {
				info := roAwait(t, n, id)
				if _, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
					t.Fatal(e)
				}
				routeRead(t, co, b)
			}
			got := routeCompletion(t, n, port, id)
			if kind == "overall" {
				if !errors.Is(got.Err, context.DeadlineExceeded) || got.StopReason == StopResponseWait {
					t.Fatal(got)
				}
			} else if !errors.Is(got.Err, ErrResponseWaitExpired) || got.StopReason != StopResponseWait || got.Attempts[attempts-1].StopReason != StopResponseWait {
				t.Fatal(got)
			}
			if !got.RoRecovery.CanContinue || len(got.Attempts) != attempts {
				t.Fatal(got)
			}
			segmentIdle(t, n)
		}
	})
}

func TestRoResponseWaitBusinessTxAndDisabledGate(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, enabled := range []bool{false, true} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), recordOptions())
			plan := trPlan(2, 0)
			if enabled {
				plan.ResponseWait = 500 * time.Millisecond
			}
			id := submitSegment(t, n, r, plan, segmentProgress(2), port)
			req := routeRead(t, co, remote)
			before := trInfo(t, n, id)
			tx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			_, e := port.Receive(tx)
			cancel()
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal(e)
			}
			after := trInfo(t, n, id)
			if after.ReplayEligible || !after.ResponseWaitUntil.Equal(before.ResponseWaitUntil) {
				t.Fatal("business Tx altered Tr")
			}
			if !enabled && !after.ResponseWaitUntil.IsZero() {
				t.Fatal(after)
			}
			if _, e = n.Retransmit(ctxTest(t), id, after.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
				t.Fatal(e)
			}
			roRecoveryReply(t, co, remote, req, 2001)
			routeCompletion(t, n, port, id)
		}
	})
}

func TestRoResponseWaitQueueAgeNotWritten(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, attempts := range []int{1, 2} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			a, b := networkAdapters(t, tls)
			barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
			var release sync.Once
			defer release.Do(func() { close(barrier.release) })
			peer := roRecoveryConnect(t, n, co, "dra", barrier, b)
			_, alternate := roRecoveryNetwork(t, n, co, "dra2", tls)
			blockPort := portTest(t, n, 1)
			blocker, e := peer.Submit(ctxTest(t), roBody(), RequestOptions{}, blockPort)
			if e != nil {
				t.Fatal(e)
			}
			routeRead(t, co, b)
			<-barrier.entered
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), recordOptions())
			id := submitSegment(t, n, r, trPlan(attempts, 25*time.Millisecond), segmentProgress(uint32(attempts)), port)
			if attempts == 1 {
				got := routeCompletion(t, n, port, id)
				if got.Outcome != NotWritten || got.Delivery.MayHaveReachedPeer || !errors.Is(got.Err, ErrResponseWaitExpired) {
					t.Fatal(got)
				}
			} else {
				info := roAwait(t, n, id)
				if info.Delivery.MayHaveReachedPeer {
					t.Fatal(info)
				}
				if _, e = n.Retransmit(ctxTest(t), id, info.CurrentAttempt); e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, alternate)
				if req.Header.Flags&pdu.RetransmittedFlag != 0 {
					t.Fatal("unsent queued request set T")
				}
				roRecoveryReply(t, co, alternate, req, 2001)
				got := routeCompletion(t, n, port, id)
				if got.Attempts[0].Delivery.MayHaveReachedPeer {
					t.Fatal(got)
				}
			}
			n.Cancel(blocker)
			release.Do(func() { close(barrier.release) })
			routeCompletion(t, n, blockPort, blocker)
			segmentIdle(t, n)
		}
	})
}

func TestRoResponseWaitWriterRetirementAndNoTimerSpin(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, b)
		_, alternate := roRecoveryNetwork(t, n, co, "dra2", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 25*time.Millisecond), segmentProgress(2), port)
		first := routeRead(t, co, b)
		<-barrier.entered
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.ReplayReason == ResponseWaitExpired })
		i := trInfo(t, n, id)
		if i.ReplayEligible || i.Phase != AttemptActive {
			t.Fatal("live writer authorized successor", i)
		}
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
			t.Fatal(e)
		}
		n.mu.Lock()
		next := n.nextDeadlineLocked()
		held := n.requests[id].charge.Capacity
		n.mu.Unlock()
		if !next.After(time.Now()) || held == 0 {
			t.Fatal("expired Tr remained on timer", next)
		}
		release.Do(func() { close(barrier.release) })
		i = roAwait(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, alternate)
		routeSameRequest(t, co, first, second)
		roRecoveryReply(t, co, alternate, second, 2001)
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 2 {
			t.Fatal(got)
		}
	})
}

func TestRoResponseWaitReaderBudgetAndOneCompletion(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, b)
		_, alternate := roRecoveryNetwork(t, n, co, "dra2", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 500*time.Millisecond), segmentProgress(2), port)
		first := routeRead(t, co, b)
		roRecoveryReply(t, co, b, first, 4012)
		<-barrier.entered
		i := trExpire(t, n, id)
		waitFor(t, func() bool { return trInfo(t, n, id).ReplayEligible })
		i = trInfo(t, n, id)
		if !i.ReplayEligible {
			t.Fatal(i)
		}
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, alternate)
		roRecoveryReply(t, co, alternate, second, 2001)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		quiet, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, e := port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("completion before A borrower", e)
		}
		if n.Resources().OutboundBytes == 0 {
			t.Fatal("early A/Q release")
		}
		release.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 2 || got.Outcome != AnswerReceived {
			t.Fatal(got)
		}
		segmentIdle(t, n)
	})
}

func TestRoResponseWaitOldAnswerSuppressesQueuedSuccessor(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, firstRemote := roRecoveryNetwork(t, n, co, "dra", tls)
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		peer := roRecoveryConnect(t, n, co, "dra2", barrier, b)
		blockPort := portTest(t, n, 1)
		block, e := peer.Submit(ctxTest(t), roBody(), RequestOptions{}, blockPort)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, b)
		<-barrier.entered
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 500*time.Millisecond), segmentProgress(2), port)
		first := routeRead(t, co, firstRemote)
		i := trExpire(t, n, id)
		waitFor(t, func() bool { return trInfo(t, n, id).ReplayEligible })
		i = trInfo(t, n, id)
		if _, e = n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		roRecoveryReply(t, co, firstRemote, first, 2001)
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 1 || got.Attempts[1].Delivery.MayHaveReachedPeer || got.Attempts[1].Delivery.Bytes != 0 {
			t.Fatal(got)
		}
		n.Cancel(block)
		release.Do(func() { close(barrier.release) })
		routeCompletion(t, n, blockPort, block)
		segmentIdle(t, n)
		roQuiet(t, b)
	})
}

func TestRoResponseWaitOldAnswerWaitsForSuccessorWriter(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, firstRemote := roRecoveryNetwork(t, n, co, "dra", tls)
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra2", barrier, b)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 500*time.Millisecond), segmentProgress(2), port)
		first := routeRead(t, co, firstRemote)
		i := trExpire(t, n, id)
		// Remote receipt can precede local writer retirement. Only the
		// successor writer is intentionally held by this test's barrier.
		waitFor(t, func() bool { return trInfo(t, n, id).ReplayEligible })
		i = trInfo(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, b)
		<-barrier.entered
		roRecoveryReply(t, co, firstRemote, first, 2001)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		quiet, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, e := port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) || n.Resources().OutboundBytes == 0 {
			t.Fatal("completion or release before successor writer retired", e)
		}
		release.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 1 || !got.Attempts[1].Delivery.MayHaveReachedPeer || got.Attempts[1].Delivery.Bytes == 0 {
			t.Fatal(got)
		}
		segmentIdle(t, n)
	})
}

func TestRoResponseWaitNoAlternateBackpressureAndIdempotence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 2)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 500*time.Millisecond), segmentProgress(2), port)
		routeRead(t, co, remote)
		i := trExpire(t, n, id)
		waitFor(t, func() bool { return trInfo(t, n, id).ReplayEligible })
		i = trInfo(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrNoEligibleAlternate) {
			t.Fatal("reused visited healthy connection", e)
		}
		_, alternate := roRecoveryNetwork(t, n, co, "dra2", tls)
		n.mu.Lock()
		limit := n.peers["dra2"].account.LimitBytes
		n.peers["dra2"].account.LimitBytes = 1
		n.mu.Unlock()
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		after := trInfo(t, n, id)
		if after.AttemptsUsed != 1 || !after.Deadline.Equal(i.Deadline) || !after.ResponseWaitUntil.Equal(i.ResponseWaitUntil) {
			t.Fatal("refusal renewed Tr", after)
		}
		n.mu.Lock()
		n.peers["dra2"].account.LimitBytes = limit
		n.mu.Unlock()
		var wg sync.WaitGroup
		results := make(chan AttemptID, 8)
		errs := make(chan error, 8)
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt)
				results <- v
				errs <- e
			}()
		}
		wg.Wait()
		close(results)
		close(errs)
		for e := range errs {
			if e != nil {
				t.Fatal(e)
			}
		}
		for a := range results {
			if a.Number != 2 || a.RequestID != id {
				t.Fatal(a)
			}
		}
		second := routeRead(t, co, alternate)
		roRecoveryReply(t, co, alternate, second, 2001)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 {
			t.Fatal(got)
		}
	})
}

func TestRoResponseWaitStopAndCancelRaces(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, end := range []string{"cancel", "peer", "port", "node", "protocol", "invalid-answer"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			peer, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), recordOptions())
			id := submitSegment(t, n, r, trPlan(2, 500*time.Millisecond), segmentProgress(2), port)
			req := routeRead(t, co, remote)
			trExpire(t, n, id)
			switch end {
			case "cancel":
				n.Cancel(id)
			case "peer":
				peer.Close()
			case "port":
				if e := port.Close(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				segmentIdle(t, n)
				continue
			case "node":
				n.Close()
			case "protocol":
				n.mu.Lock()
				c := peer.conn
				n.mu.Unlock()
				n.fail(c, ErrProtocol)
			case "invalid-answer":
				ans, e := co.CompleteAnswer(req, roAnswer(2001), pdu.Origin{Host: "wrong.example", Realm: "example"})
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
			}
			got := routeCompletion(t, n, port, id)
			if errors.Is(got.Err, ErrResponseWaitExpired) || got.Outcome == AnswerReceived {
				t.Fatal("Tr masked terminal event", end, got)
			}
			if end == "invalid-answer" && !errors.Is(got.Err, validation.ErrInvalidAnswer) {
				t.Fatal(got)
			}
			if end != "cancel" && got.RoRecovery.CanContinue {
				t.Fatal(got)
			}
			segmentIdle(t, n)
		}
	})
}

// A writer may return a fatal adapter error only after its Tr has expired.
type trFailAfterBarrier struct {
	transport.Conn
	entered, release chan struct{}
}

func (b *trFailAfterBarrier) WriteFrame(ctx context.Context, frame []byte) transport.WriteResult {
	if len(frame) > 12 && frame[9] == 0 && frame[10] == 0 && frame[11] == 4 {
		close(b.entered)
		<-b.release
		return transport.WriteResult{Err: ErrProtocol}
	}
	return b.Conn.WriteFrame(ctx, frame)
}
func TestRoResponseWaitLateWriterFailureIsTerminal(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		a, b := networkAdapters(t, tls)
		barrier := &trFailAfterBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, b)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), recordOptions())
		id := submitSegment(t, n, r, trPlan(2, 20*time.Millisecond), segmentProgress(2), port)
		<-barrier.entered
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.ReplayReason == ResponseWaitExpired })
		once.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, ErrProtocol) || got.RoRecovery.CanContinue || got.StopReason != StopProtocol {
			t.Fatal(got)
		}
	})
}
