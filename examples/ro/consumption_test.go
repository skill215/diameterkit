// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"fmt"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"reflect"
	"testing"
)

func TestFiveOperationConsumption(t *testing.T) {
	rows := []struct {
		reason localReason
		value  int
		want   [5]string
	}{
		{txTimeout, 0, [5]string{"CHECK_FAILED", "COMMIT_FAILED", "OTHER_ERROR", "COMMIT_FAILED", "CANCEL_FAILED"}},
		{txTimeout, 1, [5]string{"CHECK_SUCCESSFUL", "COMMITTED", "RESERVATION_SUCCESSFUL", "COMMITTED", "CANCELLED"}},
		{txTimeout, 2, [5]string{"NOT_CHARGED", "NOT_CHARGED", "NOT_CHARGED", "COMMITTED", "CANCELLED"}},
		{connectionDown, 0, [5]string{"CHECK_FAILED", "OTHER_ERROR", "OTHER_ERROR", "COMMIT_FAILED", "CANCEL_FAILED"}},
		{connectionDown, 1, [5]string{"NOT_CHARGED", "NOT_CHARGED", "NOT_CHARGED", "COMMITTED", "CANCELLED"}},
		{temporaryError, 1, [5]string{"NOT_CHARGED", "NOT_CHARGED", "NOT_CHARGED", "COMMITTED", "CANCELLED"}},
		{temporaryError, 0, [5]string{"CHECK_SUCCESSFUL", "COMMITTED", "RESERVATION_SUCCESSFUL", "COMMITTED", "CANCELLED"}},
		{systemFailure, 0, [5]string{"SYSTEM_FAILURE", "SYSTEM_FAILURE", "SYSTEM_FAILURE", "SYSTEM_FAILURE", "SYSTEM_FAILURE"}},
	}
	for _, row := range rows {
		for op := businessCheck; op <= businessCancel; op++ {
			t.Run(fmt.Sprintf("reason%d/config%d/op%d", row.reason, row.value, op), func(t *testing.T) {
				original := diameter.Result{Outcome: diameter.WrittenNoAnswer, Delivery: diameter.DeliveryEvidence{Uncertain: true, MayHaveReachedPeer: true}, RequestID: diameter.RequestID{Sequence: 3}}
				d, e := consumeLocal(op, row.reason, fallbackConfig{TxTimeout: row.value, ConnectionDown: row.value, TemporaryError: row.value, ErrorCause: 42, LogLSI: true}, original)
				if e != nil || d.Status != row.want[op] || d.ErrorCause != 42 || !d.LogLSI || !reflect.DeepEqual(d.Exchange, original) {
					t.Fatal(d, e)
				}
				var owner businessConsumer
				if !owner.notify(d) || owner.notify(d) {
					t.Fatal("duplicate business notification")
				}
			})
		}
	}
	if _, e := consumeLocal(99, txTimeout, fallbackConfig{}, diameter.Result{}); e == nil {
		t.Fatal("unknown operation")
	}
	if _, e := consumeLocal(businessCheck, 99, fallbackConfig{}, diameter.Result{}); e == nil {
		t.Fatal("unknown reason")
	}
}
func TestDDFHConsumptionBoundary(t *testing.T) {
	request := &combined.CCR{CCRequestType: 4, RequestedAction: avp.Some(uint32(0))}
	for _, remote := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(1)), avp.Some(uint32(2))} {
		for _, config := range []uint32{0, 1} {
			result := diameter.Result{Outcome: diameter.AnswerReceived, Validation: validation.Report{Complete: true}, Answer: pdu.Message{Body: &combined.CCA{ResultCode: 5012, DirectDebitingFailureHandling: remote}}}
			value, e := debitDDFH(request, result, config)
			if remote.Present && remote.Value == 2 {
				if e == nil {
					t.Fatal("invalid DDFH")
				}
				continue
			}
			want := config
			if remote.Present {
				want = remote.Value
			}
			if e != nil || value != want || result.Answer.Body.(*combined.CCA).ResultCode != 5012 {
				t.Fatal(value, e)
			}
			result.Validation.Complete = false
			if _, e = debitDDFH(request, result, config); e == nil {
				t.Fatal("unvalidated answer")
			}
		}
	}
	for _, kind := range []uint32{1, 3, 4} {
		for _, action := range []uint32{1, 2} {
			bad := *request
			bad.CCRequestType = kind
			bad.RequestedAction = avp.Some(action)
			if _, e := debitDDFH(&bad, diameter.Result{}, 0); e == nil {
				t.Fatal("DDFH leaked to other operation")
			}
		}
	}
	for _, r := range []diameter.Result{{Outcome: diameter.WrittenNoAnswer}, {Outcome: diameter.AnswerReceived, Validation: validation.Report{Complete: true}, Answer: pdu.Message{Body: &pdu.ErrorAnswer{ResultCode: 3004}}}} {
		if v, e := debitDDFH(request, r, 1); e != nil || v != 1 {
			t.Fatal(v, e)
		}
	}
}

func TestDebitNoBackupDecision(t *testing.T) {
	for _, tc := range []struct {
		ddfh            uint32
		qualified, save bool
		status          string
	}{{0, true, true, "actual-error"}, {1, true, true, "temporary-error"}, {0, false, true, "no-answer"}, {1, false, false, "no-answer"}} {
		save, status, e := debitNoBackup(tc.ddfh, tc.qualified)
		if e != nil || save != tc.save || status != tc.status {
			t.Fatal(save, status, e)
		}
	}
	if _, _, e := debitNoBackup(2, false); e == nil {
		t.Fatal("reserved DDFH")
	}
}
