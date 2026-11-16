package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	diameter "github.com/skill215/diameterkit"
	"testing"
	"time"
)

func TestPublicRecoveryFlow(t *testing.T) {
	for _, scenario := range []string{"error", "restart", "no-answer"} {
		for _, result := range []string{"2001", "2001/4012", "4012", "5012", "3004"} {
			t.Run(scenario+"/"+result, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()
				var out bytes.Buffer
				if _, e := runRecovery(ctx, scenario, result, &out); e != nil {
					t.Fatal(e, out.String())
				}
				if out.Len() == 0 {
					t.Fatal("missing result")
				}
			})
		}
	}
}

func setupRecovery(t *testing.T) (context.Context, *pair) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	t.Cleanup(cancel)
	cfg, e := recoveryConfigs()
	if e != nil {
		t.Fatal(e)
	}
	p, e := connectPair(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := p.close(); e != nil {
			t.Error(e)
		}
		if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
			t.Error("retained resources")
		}
	})
	return ctx, p
}
func prepareIntent(t *testing.T, ctx context.Context, p *pair) recoveryIntent {
	t.Helper()
	body, e := newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	record, e := p.left.node.PrepareRoRecoveryRequest(ctx, body, recoveryOptions())
	if e != nil {
		t.Fatal(e)
	}
	return recoveryIntent{record, recoveryPlan(), diameter.RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: 1, ReservedThrough: 1}}
}
func TestRecoveryConfirmationAndImportRejection(t *testing.T) {
	ctx, old := setupRecovery(t)
	i := prepareIntent(t, ctx, old)
	digest, e := intentDigest(i)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = submitConfirmed(ctx, old.left.node, i, false, digest, old.left.port); e == nil {
		t.Fatal("unconfirmed sent")
	}
	mutated := i
	mutated.Progress.ReservedThrough++
	if _, e = submitConfirmed(ctx, old.left.node, mutated, true, digest, old.left.port); e == nil {
		t.Fatal("changed intent sent")
	}
	if old.left.node.Resources().Outstanding != 0 {
		t.Fatal("rejected intent reserved completion")
	}
	if e = old.close(); e != nil {
		t.Fatal(e)
	}
	ctx, p := setupRecovery(t)
	cp := diameter.RoRecoveryCheckpoint{RecordID: i.Record.RecordID, NextTrigger: diameter.RoRecoveryInitial}
	base := &diameter.RoRecoveryImport{Checkpoint: cp, CheckedAt: time.Now().UTC(), SenderFenced: true, ProgressConfirmed: true, ClockConfirmed: true}
	for _, bad := range []string{"fencing", "progress", "clock", "terminal", "rollback", "digest", "budget"} {
		t.Run(bad, func(t *testing.T) {
			plan := i.Plan
			claim := *base
			plan.Import = &claim
			record := i.Record
			progress := i.Progress
			switch bad {
			case "fencing":
				claim.SenderFenced = false
			case "progress":
				claim.ProgressConfirmed = false
			case "clock":
				claim.ClockConfirmed = false
			case "terminal":
				claim.Checkpoint.NextTrigger = 0
			case "rollback":
				claim.Checkpoint.SegmentNo = 1
				claim.Checkpoint.ReservedThrough = 1
			case "budget":
				plan.MaxAttempts = 2
				progress.SegmentAllowance = 2
				progress.ReservedThrough = 2
			case "digest":
				record.Frame = append([]byte(nil), record.Frame...)
				record.Frame[len(record.Frame)-1] ^= 1
			}
			if _, e := p.left.node.SubmitRoRecovery(ctx, record, plan, progress, p.left.port); e == nil {
				t.Fatal("invalid import admitted")
			}
		})
	}
	i.Plan.Import = base
	digest, e = intentDigest(i)
	if e != nil {
		t.Fatal(e)
	}
	id, e := submitConfirmed(ctx, p.left.node, i, true, digest, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	a, e := newAnswer("2001")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = in.Respond(ctx, a); e != nil {
		t.Fatal(e)
	}
	result, e := receiveRecovery(ctx, p, id)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = nextRecovery(i.Record, result); !errors.Is(e, diameter.ErrRoRecoveryFinal) {
		t.Fatal(e)
	}
	if _, e = p.left.node.SubmitRoRecovery(ctx, i.Record, i.Plan, i.Progress, p.left.port); e == nil {
		t.Fatal("terminal replay")
	}
}
func TestRecoveryTxLateAnswerAndCancel(t *testing.T) {
	for _, cancelAtTx := range []bool{false, true} {
		t.Run(fmt.Sprint(cancelAtTx), func(t *testing.T) {
			ctx, p := setupRecovery(t)
			i := prepareIntent(t, ctx, p)
			i.Plan.ResponseWait = time.Second
			digest, e := intentDigest(i)
			if e != nil {
				t.Fatal(e)
			}
			id, e := submitConfirmed(ctx, p.left.node, i, true, digest, p.left.port)
			if e != nil {
				t.Fatal(e)
			}
			in, e := p.right.peer.Receive(ctx)
			if e != nil {
				t.Fatal(e)
			}
			info, e := p.left.node.RequestInfo(id)
			if e != nil {
				t.Fatal(e)
			}
			timer := time.NewTimer(time.Until(info.AcceptedAt.Add(20 * time.Millisecond)))
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			var owner businessConsumer
			d, e := consumeLocal(businessCharge, txTimeout, fallbackConfig{}, diameter.Result{RequestID: id})
			if e != nil || !owner.notify(d) {
				t.Fatal(e)
			}
			if cancelAtTx {
				if e = p.left.node.Cancel(id); e != nil {
					t.Fatal(e)
				}
			}
			a, e := newAnswer("2001")
			if e != nil {
				t.Fatal(e)
			}
			if _, e = in.Respond(ctx, a); e != nil {
				t.Fatal(e)
			}
			result, e := receiveRecovery(ctx, p, id)
			if e != nil {
				t.Fatal(e)
			}
			if cancelAtTx {
				if result.Err == nil {
					t.Fatal("cancel lost")
				}
			} else if result.Err != nil || result.Outcome != diameter.AnswerReceived {
				t.Fatal(result)
			}
			if owner.notify(businessDecision{Exchange: result}) {
				t.Fatal("late result notified twice")
			}
			wait, stop := context.WithTimeout(ctx, 20*time.Millisecond)
			defer stop()
			if _, e = p.left.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("extra completion", e)
			}
		})
	}
}

func TestRecoveryOptions(t *testing.T) {
	for _, s := range []string{"error", "restart", "no-answer"} {
		if !validRecoveryOptions("recovery", s) || !validFlowOptions("recovery", "2001", "2001", knownAbortSession) {
			t.Fatal(s)
		}
	}
	if validRecoveryOptions("iec", "restart") || validRecoveryOptions("recovery", "unknown") {
		t.Fatal("invalid scenario")
	}
}
