// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

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

func TestCombinedPublicCrossTarget(t *testing.T) {
	for _, action := range []string{"debit", "balance", "refund", "initial", "term"} {
		for _, code := range []uint32{2001, 4012, 5012, 3004} {
			t.Run(fmt.Sprintf("%s/%d", action, code), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()
				var out bytes.Buffer
				if e := runCompat(ctx, action, code, &out); e != nil {
					t.Fatal(e)
				}
				if out.Len() == 0 {
					t.Fatal("missing result")
				}
			})
		}
	}
}

func TestCombinedTargetAuthorityRejections(t *testing.T) {
	for _, bad := range []string{"domain", "target-session", "application-session", "unknown-target", "changed-record", "regressed-progress"} {
		t.Run(bad, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cfg, secondary, e := compatConfigs()
			if e != nil {
				t.Fatal(e)
			}
			if bad == "domain" {
				cfg[0].RoTargets[1].ReplayDomain = "unrelated-domain"
			}
			if bad == "target-session" {
				cfg[0].RoTargets[1].SessionEvidenceID = ""
			}
			primary, backup, e := openCompatTargets(ctx, cfg, secondary)
			if e != nil {
				t.Fatal(e)
			}
			defer func() {
				if e := primary.close(); e != nil {
					t.Error(e)
				}
				if e := backup.close(); e != nil {
					t.Error(e)
				}
				for _, n := range []*diameter.Node{primary.left.node, primary.right.node, backup.right.node} {
					if n.Resources() != (diameter.ResourceSnapshot{}) {
						t.Error("resource leak")
					}
				}
			}()
			body, e := compatRequest("term", "ctf.example;existing-session")
			if e != nil {
				t.Fatal(e)
			}
			options := compatOptions()
			if bad == "application-session" {
				options.SessionEvidenceID = ""
			}
			record, e := primary.left.node.PrepareRoRecoveryRequest(ctx, body, options)
			if e != nil {
				t.Fatal(e)
			}
			intent := recoveryIntent{record, recoveryPlan(), diameter.RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: 1, ReservedThrough: 1}}
			digest, e := intentDigest(intent)
			if e != nil {
				t.Fatal(e)
			}
			id, e := submitConfirmed(ctx, primary.left.node, intent, true, digest, primary.left.port)
			if e != nil {
				t.Fatal(e)
			}
			in, e := primary.right.peer.Receive(ctx)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = in.Respond(ctx, compatAnswer(body, 3004)); e != nil {
				t.Fatal(e)
			}
			first, e := receiveRecovery(ctx, primary, id)
			if e != nil {
				t.Fatal(e)
			}
			next, e := nextRecovery(record, first)
			if e != nil {
				t.Fatal(e)
			}
			next.Plan.TargetID = "backup"
			next.Plan.Peers = []diameter.PeerID{backupPeer}
			switch bad {
			case "unknown-target":
				next.Plan.TargetID = "not-frozen"
			case "changed-record":
				next.Record.Frame = append([]byte(nil), record.Frame...)
				next.Record.Frame[len(next.Record.Frame)-1] ^= 1
			case "regressed-progress":
				next.Progress.ReservedBefore = 0
				next.Progress.ReservedThrough = 1
			}
			if _, e = primary.left.node.SubmitRoRecovery(ctx, next.Record, next.Plan, next.Progress, primary.left.port); e == nil {
				t.Fatal("unauthorized continuation")
			}
			if primary.left.node.Resources().Outstanding != 0 {
				t.Fatal("rejection reserved completion")
			}
			wait, stop := context.WithTimeout(ctx, 20*time.Millisecond)
			defer stop()
			if _, e = backup.right.peer.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("rejection sent request", e)
			}
		})
	}
}
func TestCombinedNewOperationAndOptions(t *testing.T) {
	// Refund is a separately authorized operation with its own Session; recovery
	// of that refund retains its identity. This does not infer refund eligibility.
	debit, e := compatRequest("debit", "ctf;debit")
	if e != nil {
		t.Fatal(e)
	}
	refund, e := compatRequest("refund", "ctf;refund")
	if e != nil {
		t.Fatal(e)
	}
	if debit.SessionID == refund.SessionID || refund.RequestedAction.Value != 1 || !refund.RefundInformation.Present || !bytes.Equal(refund.RefundInformation.Value, []byte{0, 255}) {
		t.Fatal("refund identity/token")
	}
	for _, action := range []string{"debit", "balance", "refund", "initial", "term"} {
		if !validCompatOptions("compat", action) {
			t.Fatal(action)
		}
	}
	if validCompatOptions("compat", "unknown") || validCompatOptions("iec", "refund") || validFlowOptions("compat", "2001/4012", "2001", knownAbortSession) {
		t.Fatal("invalid options")
	}
}
