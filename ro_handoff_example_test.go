package diameter

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/skill215/diameterkit/avp"
)

// These test-only application examples deliberately contain no storage provider,
// worker, automatic retry or exported library policy. A single application owner
// serializes decisions and keeps the intent immutable while confirming/submitting it.
type handoffIntent struct {
	Record   RoRecoveryRequest
	Plan     RoRecoveryPlan
	Progress RoRecoveryProgress
}
type handoffCommit uint8

const (
	handoffUnknownCommit handoffCommit = iota
	handoffCommitted
	handoffRejectedCommit
)

type handoffReceipt struct {
	State        handoffCommit
	IntentDigest [32]byte
}

var errHandoffUnconfirmed = errors.New("application intent is not confirmed")

// This sample hash binds an application acknowledgement to the exact intent,
// including its plan and counters. JSON/hash are illustrative, not a library
// storage format, authentication protocol or durable-commit proof.
func handoffDigest(i handoffIntent) ([32]byte, error) {
	b, e := json.Marshal(i)
	if e != nil {
		return [32]byte{}, e
	}
	return sha256.Sum256(b), nil
}
func handoffSubmit(ctx context.Context, n *Node, i handoffIntent, receipt handoffReceipt, port *CompletionPort) (RequestID, error) {
	digest, err := handoffDigest(i)
	if err != nil {
		return RequestID{}, err
	}
	if receipt.State != handoffCommitted || receipt.IntentDigest != digest {
		return RequestID{}, errHandoffUnconfirmed
	}
	// Application confirms this exact intent BEFORE this call, never in a library callback.
	return n.SubmitRoRecovery(ctx, i.Record, i.Plan, i.Progress, port)
}

// Only after the application fences the old sender and confirms that this intent
// may have been admitted. A known terminal result MUST NOT be replaced by this
// construction. MaxFrameBytes covers every frozen target, including longer hosts.
func handoffCrashCheckpoint(i handoffIntent) RoRecoveryCheckpoint {
	d := i.Progress.Delivery
	d.FrameBytes = max(d.FrameBytes, i.Record.MaxFrameBytes)
	d.Uncertain, d.MayHaveReachedPeer = true, true
	return RoRecoveryCheckpoint{
		RecordID: i.Record.RecordID, SegmentNo: i.Progress.SegmentNo,
		ReservedThrough: i.Progress.ReservedThrough, Delivery: d,
		NextTrigger: RoRecoveryNoAnswer, ReplayPending: true, Unknown: true,
	}
}

// One application event-loop owner; this is not a concurrent library completion port.
type handoffBusiness struct{ completed bool }

func (b *handoffBusiness) finishOnce() bool {
	if b.completed {
		return false
	}
	b.completed = true
	return true
}

// Named legacy DIRECT_DEBITING example only. The caller extracts remote from an
// applicable, validated CCA, never a timeout, E answer or another operation.
func handoffDDFH(directDebiting bool, configured uint32, remote avp.Optional[uint32]) (uint32, error) {
	if !directDebiting || configured > 1 || remote.Present && remote.Value > 1 {
		return 0, ErrConfig
	}
	if remote.Present {
		return remote.Value, nil
	}
	return configured, nil
}

// Named sms-fw no-backup policy: qualified pre-Tx error saves for either DDFH;
// exhausted Tr without an answer saves only for configured DDFH=0. The caller
// has already checked eligibility/phase. "save" requests application work, not
// a claim that anything was stored. No protocol result is rewritten here.
func handoffNoBackup(ddfh uint32, qualifiedError bool) (save bool, business string) {
	if qualifiedError {
		if ddfh == 1 {
			return true, "temporary-error"
		}
		return true, "actual-error"
	}
	return ddfh == 0, "no-answer"
}

func ExampleNode_SubmitRoRecovery_commitUnknown() {
	// Real code first calls PrepareRoRecoveryRequest and confirms the entire
	// record/plan/progress intent in its own storage/ownership protocol.
	intent := handoffIntent{}
	receipt := handoffReceipt{State: handoffUnknownCommit}
	// Nil Node/Port are safe only in this deliberately rejected example: no
	// library call is made while commit status is unknown. Use real objects after confirmation.
	_, err := handoffSubmit(context.Background(), nil, intent, receipt, nil)
	fmt.Println(errors.Is(err, errHandoffUnconfirmed))
	// Output: true
}

func ExampleRoRecoveryImport_crashUnknown() {
	// Illustrative already-confirmed intent: third segment reserved two attempts.
	intent := handoffIntent{
		Record:   RoRecoveryRequest{RecordID: [16]byte{1}, MaxFrameBytes: 4096},
		Progress: RoRecoveryProgress{SegmentNo: 3, ReservedBefore: 4, SegmentAllowance: 2, ReservedThrough: 6},
	}
	cp := handoffCrashCheckpoint(intent)
	fmt.Println(cp.SegmentNo, cp.ReservedThrough, cp.Delivery.Uncertain, cp.ReplayPending)
	// Application must confirm fencing/history/time and save the NEXT intent before
	// using cp in Plan.Import on a new Node; the sample record is not a valid wire record.
	// Output: 3 6 true true
}

func ExampleRoRecoveryResult_Checkpoint_terminal() {
	// NextTrigger=0 represents a known terminal result. Preserve it; do not replace
	// it with an Unknown checkpoint merely because the process is restarting.
	result := RoRecoveryResult{RecordID: [16]byte{1}, SegmentNo: 2, ReservedThrough: 4}
	cp := result.Checkpoint()
	fmt.Println(cp.NextTrigger == 0, cp.ReservedThrough)
	// Output: true 4
}
