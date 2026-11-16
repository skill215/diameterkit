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
	"strings"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
)

func TestPublicBindingRecoveryMatrix(t *testing.T) {
	for _, binding := range []string{"combined", "smsfwro", "legacy420"} {
		for _, action := range []string{"debit", "balance", "refund", "initial", "term"} {
			if !validBoundCompat(binding, action, "error") {
				continue
			}
			for _, scenario := range []string{"error", "restart", "no-answer", "no-answer-import"} {
				// Batch89 already covers this exact slice; retain its regression tests.
				if binding == "combined" && scenario == "error" {
					continue
				}
				for _, code := range []uint32{2001, 4012, 5012, 3004} {
					t.Run(fmt.Sprintf("%s/%s/%s/%d", binding, action, scenario, code), func(t *testing.T) {
						ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
						defer cancel()
						var out bytes.Buffer
						if e := runBoundCompat(ctx, binding, action, scenario, code, &out); e != nil {
							t.Fatal(e)
						}
						if !strings.Contains(out.String(), "reserved=2") || !strings.Contains(out.String(), "scenario="+scenario) {
							t.Fatal(out.String())
						}
					})
				}
			}
		}
	}
}

func TestPublicBindingSelection(t *testing.T) {
	for _, v := range [][3]string{{"smsfwro", "balance", "error"}, {"smsfwro", "refund", "error"}, {"legacy420", "initial", "error"}, {"legacy420", "term", "error"}, {"unknown", "debit", "error"}, {"combined", "debit", "unknown"}} {
		if validBoundCompat(v[0], v[1], v[2]) {
			t.Fatal(v)
		}
		if e := runBoundCompat(context.Background(), v[0], v[1], v[2], 2001, &bytes.Buffer{}); !errors.Is(e, diameter.ErrConfig) {
			t.Fatal(v, e)
		}
	}
	if validCompatSelection("iec", "smsfwro", "debit", "error") || !validRecoveryOptions("compat", "no-answer-import") || validRecoveryOptions("recovery", "no-answer-import") {
		t.Fatal("CLI selection")
	}
}

// Each fixture produces an actual written/no-answer completion, fences all old
// owners, and submits only through the public import API on fresh Nodes.
func TestPublicCrossTargetImportRejections(t *testing.T) {
	for _, binding := range []string{"combined", "smsfwro", "legacy420"} {
		t.Run(binding, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()
			cfg, secondary, e := boundCompatConfigs(binding)
			if e != nil {
				t.Fatal(e)
			}
			old, oldBackup, e := openCompatTargets(ctx, cfg, secondary)
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() {
				if e := closeCompatTargets(old, oldBackup); e != nil {
					t.Error(e)
				}
			})
			body, e := boundCompatRequest(binding, "debit", "ctf;import-rejection")
			if e != nil {
				t.Fatal(e)
			}
			record, e := old.left.node.PrepareRoRecoveryRequest(ctx, body, compatOptions())
			if e != nil {
				t.Fatal(e)
			}
			initial := recoveryIntent{record, recoveryPlan(), diameter.RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: 1, ReservedThrough: 1}}
			id, e := old.left.node.SubmitRoRecovery(ctx, record, initial.Plan, initial.Progress, old.left.port)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = old.right.peer.Receive(ctx); e != nil {
				t.Fatal(e)
			}
			first, e := receiveRecovery(ctx, old, id)
			if e != nil {
				t.Fatal(e)
			}
			if first.Outcome != diameter.WrittenNoAnswer || !first.RoRecovery.ReplayPending {
				t.Fatal(first)
			}
			next, e := nextRecovery(record, first)
			if e != nil {
				t.Fatal(e)
			}
			if e = closeCompatTargets(old, oldBackup); e != nil {
				t.Fatal(e)
			}
			next.Plan.TargetID, next.Plan.Peers = "backup", []diameter.PeerID{backupPeer}
			for _, bad := range []string{"fencing", "progress", "clock", "checkpoint", "domain", "evidence", "binding", "import-disabled"} {
				t.Run(bad, func(t *testing.T) {
					cfg, secondary, e := boundCompatConfigs(binding)
					if e != nil {
						t.Fatal(e)
					}
					if bad == "domain" {
						cfg[0].RoTargets[1].ReplayDomain = "unrelated"
					}
					if bad == "evidence" {
						cfg[0].RoTargets[1].EvidenceID = "changed-authority"
					}
					if bad == "binding" {
						other := "combined"
						if binding == other {
							other = "smsfwro"
						}
						cfg, secondary, e = boundCompatConfigs(other)
						if e != nil {
							t.Fatal(e)
						}
					}
					if bad == "import-disabled" {
						cfg[0].AllowRoRecoveryImport = false
					}
					p, backup, e := openCompatTargets(ctx, cfg, secondary)
					if e != nil {
						t.Fatal(e)
					}
					defer func() {
						if e := closeCompatTargets(p, backup); e != nil {
							t.Error(e)
						}
					}()
					i := next
					i.Plan.Import = &diameter.RoRecoveryImport{Checkpoint: first.RoRecovery.Checkpoint(), CheckedAt: time.Now().UTC(), SenderFenced: true, ProgressConfirmed: true, ClockConfirmed: true}
					switch bad {
					case "fencing":
						i.Plan.Import.SenderFenced = false
					case "progress":
						i.Plan.Import.ProgressConfirmed = false
					case "clock":
						i.Plan.Import.ClockConfirmed = false
					case "checkpoint":
						i.Plan.Import.Checkpoint.ReservedThrough = 0
					}
					if _, e = p.left.node.SubmitRoRecovery(ctx, i.Record, i.Plan, i.Progress, p.left.port); e == nil {
						t.Fatal("invalid import admitted")
					}
					if p.left.node.Resources().Outstanding != 0 {
						t.Fatal("rejection retained completion")
					}
					wait, stop := context.WithTimeout(ctx, 20*time.Millisecond)
					defer stop()
					if _, e = backup.right.peer.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
						t.Fatal("rejection sent request", e)
					}
				})
			}
		})
	}
}
