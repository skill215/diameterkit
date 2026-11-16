// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
)

// Test harness simulates the application's explicit acknowledgement. It does not
// claim that memory, JSON or this helper supplies durable storage or fencing.
func handoffConfirm(t *testing.T, i handoffIntent) handoffReceipt {
	t.Helper()
	h, e := handoffDigest(i)
	if e != nil {
		t.Fatal(e)
	}
	return handoffReceipt{State: handoffCommitted, IntentDigest: h}
}
func handoffSend(t *testing.T, n *Node, i handoffIntent, port *CompletionPort) RequestID {
	t.Helper()
	id, e := handoffSubmit(ctxTest(t), n, i, handoffConfirm(t, i), port)
	if e != nil {
		t.Fatal(e)
	}
	return id
}

func TestRoHandoffCommitUnknownAndExactIntent(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		i := handoffIntent{Record: prepareRecord(t, n, roBody(), recordOptions()), Plan: segmentPlan(1), Progress: segmentProgress(1)}
		ack := handoffConfirm(t, i)
		for _, state := range []handoffCommit{handoffUnknownCommit, handoffRejectedCommit} {
			r := ack
			r.State = state
			if _, e := handoffSubmit(ctxTest(t), n, i, r, port); !errors.Is(e, errHandoffUnconfirmed) {
				t.Fatal(e)
			}
		}
		for _, change := range []func(*handoffIntent){
			func(x *handoffIntent) { x.Plan.TargetID = "ocs-b" },
			func(x *handoffIntent) { x.Progress.ReservedThrough++ },
			func(x *handoffIntent) { x.Plan.Deadline = time.Now().Add(time.Second) },
			func(x *handoffIntent) { x.Record.RecordID[0] ^= 1 },
		} {
			changed := i
			change(&changed)
			if _, e := handoffSubmit(ctxTest(t), n, changed, ack, port); !errors.Is(e, errHandoffUnconfirmed) {
				t.Fatal("different intent accepted", e)
			}
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		// Confirmation arrives for the exact original intent; only now submit once.
		id, e := handoffSubmit(ctxTest(t), n, i, ack, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		roRecoveryReply(t, co, remote, req, 2001)
		got := routeCompletion(t, n, port, id)
		if got.Err != nil || got.RoRecovery.ReservedThrough != 1 {
			t.Fatal(got)
		}
	})
}

func TestRoHandoffCrashConsumesReservedIntent(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, sent := range []bool{false, true} {
			old, co := newInboundNode(t, recordConfig(t, "standard"))
			i := handoffIntent{Record: prepareRecord(t, old, roBody(), recordOptions()), Plan: segmentPlan(2), Progress: segmentProgress(2)}
			// A confirmed intent may have reached Submit; absence of saved Completion is
			// not evidence that no bytes were sent. Exercise both indistinguishable cases.
			if sent {
				_, remote := roRecoveryNetwork(t, old, co, "dra", tls)
				handoffSend(t, old, i, portTest(t, old, 1))
				routeRead(t, co, remote)
			}
			old.Close()
			if e := old.Wait(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
			cp := handoffCrashCheckpoint(i)
			if cp.ReservedThrough != 2 || cp.Delivery.Bytes != 0 || !cp.Unknown {
				t.Fatal(cp)
			}
			n, co := newInboundNode(t, importConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			next := handoffIntent{Record: copyRecord(t, i.Record), Plan: importPlan(cp), Progress: importProgress(cp)}
			id := handoffSend(t, n, next, port)
			req := routeRead(t, co, remote)
			if req.Header.Flags&pdu.RetransmittedFlag == 0 {
				t.Fatal("uncertain intent lost T")
			}
			roRecoveryReply(t, co, remote, req, 2001)
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.ReservedThrough != 3 || !got.RoRecovery.CumulativeDelivery.Uncertain {
				t.Fatal(got)
			}
		}
	})
}

func TestRoHandoffKnownTerminalCannotRestart(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{2001, 4012} {
			old, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, old, co, "dra", tls)
			port := portTest(t, old, 1)
			i := handoffIntent{Record: prepareRecord(t, old, roBody(), recordOptions()), Plan: segmentPlan(1), Progress: segmentProgress(1)}
			id := handoffSend(t, old, i, port)
			req := routeRead(t, co, remote)
			roRecoveryReply(t, co, remote, req, code)
			got := routeCompletion(t, old, port, id)
			cp := got.RoRecovery.Checkpoint()
			if cp.NextTrigger != 0 {
				t.Fatal(cp)
			}
			old.Close()
			if e := old.Wait(ctxTest(t)); e != nil {
				t.Fatal(e)
			}
			n, _ := newInboundNode(t, importConfig(t, "standard"))
			next := handoffIntent{Record: copyRecord(t, i.Record), Plan: importPlan(cp), Progress: importProgress(cp)}
			// Deliberately attempt to override just the plan; the saved terminal checkpoint remains authoritative.
			next.Plan.Trigger = RoRecoveryNoAnswer
			_, e := handoffSubmit(ctxTest(t), n, next, handoffConfirm(t, next), portTest(t, n, 1))
			if !errors.Is(e, ErrRoRecoveryFinal) {
				t.Fatal(e)
			}
			if n.RoRecoveryInfo().Held != 0 || n.Resources().OutboundBytes != 0 {
				t.Fatal("terminal imported")
			}
		}
	})
}

func TestRoHandoffTxDoesNotCompleteProtocol(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, cancelAtTx := range []bool{false, true} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			plan := segmentPlan(2)
			plan.ResponseWait = time.Second
			i := handoffIntent{Record: prepareRecord(t, n, roBody(), recordOptions()), Plan: plan, Progress: segmentProgress(2)}
			id := handoffSend(t, n, i, port)
			req := routeRead(t, co, remote)
			info, e := n.RequestInfo(id)
			if e != nil {
				t.Fatal(e)
			}
			tx := info.AcceptedAt.Add(10 * time.Millisecond)
			timer := time.NewTimer(max(time.Until(tx), 0))
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctxTest(t).Done():
				t.Fatal("Tx wait")
			}
			business := handoffBusiness{}
			notifications := 0
			if business.finishOnce() {
				notifications++
			} // application timeout, never a fabricated CCA
			if cancelAtTx {
				n.Cancel(id) // explicit old TIMEOUT_RETRY_ENABLED=0 policy
				got := routeCompletion(t, n, port, id)
				if !errors.Is(got.Err, context.Canceled) {
					t.Fatal(got)
				}
				roRecoveryReply(t, co, remote, req, 2001) // deliberately outside the accepted result window
			} else {
				if _, e := n.Retransmit(ctxTest(t), id, info.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
					t.Fatal("Tx granted replay", e)
				}
				roRecoveryReply(t, co, remote, req, 2001)
				got := routeCompletion(t, n, port, id)
				if got.Err != nil || got.Answer.Body == nil {
					t.Fatal(got)
				}
			}
			if business.finishOnce() {
				notifications++
			} // handle background result without waking caller twice
			if notifications != 1 {
				t.Fatal(notifications)
			}
			short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			_, e = port.Receive(short)
			cancel()
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("second Completion", e)
			}
			segmentIdle(t, n)
		}
	})
}

func TestRoHandoffDDFHPresenceAndActualError(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, remoteDDFH := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(1))} {
			for _, configured := range []uint32{0, 1} {
				n, co := newInboundNode(t, recordConfig(t, "combined"))
				_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
				port := portTest(t, n, 1)
				opts := recordOptions()
				opts.PolicyID = RoSMSFWRecoveryPolicy
				i := handoffIntent{Record: prepareRecord(t, n, combinedBody(t, "general"), opts), Plan: errorPlan(RoErrorSMSFWBeforeTx), Progress: segmentProgress(2)}
				i.Plan.MaxAttempts = 2
				id := handoffSend(t, n, i, port)
				req := routeRead(t, co, remote)
				answer := roRecoveryAnswer(req.Body, 5012).(*combined.CCA)
				answer.DirectDebitingFailureHandling = remoteDDFH
				msg, e := co.CompleteAnswer(req, answer, serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, msg); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				actual := got.Answer.Body.(*combined.CCA)
				effective, e := handoffDDFH(true, configured, actual.DirectDebitingFailureHandling)
				want := configured
				if remoteDDFH.Present {
					want = remoteDDFH.Value
				}
				if e != nil || effective != want || actual.ResultCode != 5012 || got.RoRecovery.NextTrigger != RoRecoveryErrorAnswer {
					t.Fatal(e, effective, got)
				}
				save, businessCode := handoffNoBackup(effective, got.RoRecovery.NextTrigger == RoRecoveryErrorAnswer)
				expectedBusiness := map[uint32]string{0: "actual-error", 1: "temporary-error"}[want]
				if !save || businessCode != expectedBusiness || n.Resources().Outstanding != 0 {
					t.Fatal("wrong no-backup disposition", save, businessCode)
				}
			}
		}
	})
	for _, remote := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(1))} {
		if _, e := handoffDDFH(false, 0, remote); !errors.Is(e, ErrConfig) {
			t.Fatal("DDFH applied to another action", e)
		}
	}
	if _, e := handoffDDFH(true, 0, avp.Some(uint32(2))); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
}

func TestRoHandoffNoAnswerUsesConfiguredDDFH(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, configured := range []uint32{0, 1} {
			n, co := newInboundNode(t, recordConfig(t, "combined"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			plan := segmentPlan(1)
			plan.ResponseWait = 20 * time.Millisecond
			i := handoffIntent{Record: prepareRecord(t, n, combinedBody(t, "general"), recordOptions()), Plan: plan, Progress: segmentProgress(1)}
			id := handoffSend(t, n, i, port)
			routeRead(t, co, remote)
			got := routeCompletion(t, n, port, id)
			if got.Answer.Body != nil || got.StopReason != StopResponseWait || !got.Delivery.MayHaveReachedPeer {
				t.Fatal(got)
			}
			effective, e := handoffDDFH(true, configured, avp.Optional[uint32]{})
			if e != nil {
				t.Fatal(e)
			}
			save, business := handoffNoBackup(effective, false)
			if save != (configured == 0) || business != "no-answer" {
				t.Fatal(save, business)
			}
			// Saving is not automatic: even DDFH=0 has no active request after completion.
			if n.Resources().Outstanding != 0 {
				t.Fatal(n.Resources())
			}
		}
	})
}

func TestRoHandoffLateErrorKeepsOriginalTxDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{3004, 5012} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			opts := recordOptions()
			opts.PolicyID = RoSMSFWRecoveryPolicy
			plan := errorPlan(RoErrorSMSFWBeforeTx)
			// Original business cutoff is already past when this later segment begins.
			// It is carried forward, not reset to now+Tx for each recovery segment.
			plan.ErrorContinuationUntil = time.Now().Add(-time.Millisecond)
			i := handoffIntent{Record: prepareRecord(t, n, roBody(), opts), Plan: plan, Progress: segmentProgress(1)}
			id := handoffSend(t, n, i, port)
			req := routeRead(t, co, remote)
			errorReply(t, co, remote, req, code)
			got := routeCompletion(t, n, port, id)
			if got.Err != nil || got.Answer.Body == nil {
				t.Fatal(got)
			}
			if got.RoRecovery.CanContinue != (code == 3004) {
				t.Fatal("late result used pre-Tx extra error set", got)
			}
			if code == 3004 && got.RoRecovery.ErrorResultCode != code {
				t.Fatal(got)
			}
		}
	})
}
