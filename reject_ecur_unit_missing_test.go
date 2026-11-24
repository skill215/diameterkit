// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group omissions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package diameter

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestECURUnitMissingFrozen(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for fixture, code := range map[string]uint32{"init-no-rsu": 437, "term-no-usu": 446} {
		t.Run(fixture, func(t *testing.T) {
			src, e := os.ReadFile("testdata/codec/b43-" + fixture + ".hex")
			if e != nil {
				t.Fatal(e)
			}
			raw, e := hex.DecodeString(string(bytes.Join(bytes.Fields(src), nil)))
			if e != nil {
				t.Fatal(e)
			}
			_, cause := co.Decode(raw)
			var f *validation.FieldError
			if !errors.Is(cause, validation.ErrMissing) || !errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &f) || len(f.Path) != 2 || f.Path[0] != (avp.Key{Code: 456}) || f.Path[1] != (avp.Key{Code: code}) {
				t.Fatal(cause)
			}
			want := invalidPart(t, avp.Key{Code: 456}, 64, ecurUnitBytes(t, code, 0))
			for _, variant := range []string{"exact", "bytes", "field", "frame", "depth", "count", "bare", "path"} {
				t.Run(variant, func(t *testing.T) {
					budget := 32
					l := wire.Limits{MaxFrameBytes: 60, MaxFieldBytes: 32, MaxDepth: 3, MaxAVPs: 4, MaxRepeated: 1}
					proof := cause
					switch variant {
					case "bytes":
						budget--
					case "field":
						l.MaxFieldBytes--
					case "frame":
						l.MaxFrameBytes--
					case "depth":
						l.MaxDepth--
					case "count":
						l.MaxAVPs--
					case "bare":
						proof = &validation.FieldError{Path: f.Path, Cause: validation.ErrMissing}
					case "path":
						proof = &validation.FieldError{Path: []avp.Key{{Code: 456}, {Code: 417}}, Cause: f.Cause}
					}
					q, ok := missingavp.Quote(proof, l, budget)
					if variant == "exact" {
						if !ok || !bytes.Equal(q, want) {
							t.Fatal(q, cause)
						}
						return
					}
					if ok || q != nil {
						t.Fatal("partial or unproven quote", q)
					}
				})
			}
			reply, e := n.rejectFrame(raw, 4096, rejectDecode, cause)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply), rejectDecode, cause); e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply)-1, rejectDecode, cause); e == nil {
				t.Fatal("short answer accepted")
			}
		})
	}
}

func TestECURUnitMissingNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for op, code := range map[string]uint32{"initial": 437, "term": 446} {
			for _, variant := range []string{"answer", "default", "mixed-status", "mixed-unit", "earlier-action", "child-only"} {
				t.Run(op+"/"+variant, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if variant == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, tls)
					req := combinedDuplicateRequest(t, n, co, op, 14401, 24401)
					path := lengthPath(avp.Key{Code: 456}, avp.Key{Code: code})
					if variant == "child-only" {
						path = append(path, lengthPath(avp.Key{Code: 417})...)
					}
					raw := missingRaw(t, co, req, path, false)
					switch variant {
					case "mixed-status":
						if op == "initial" {
							raw = appendStatusField(t, raw, 2014, nil)
						} else {
							raw = smsOmit(t, raw, 2012)
						}
					case "mixed-unit":
						if op == "initial" {
							raw = unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 456}), ecurUnitBytes(t, 446, 1))
						} else {
							raw = mmsGroupOmit(t, raw, unitValuePath("term"))
						}
					case "earlier-action":
						raw = appendECURAction(raw, invalidPart(t, avp.Key{Code: 436}, 64, make([]byte, 4)))
					}
					_, cause := co.Decode(raw)
					if variant == "child-only" {
						var f *validation.FieldError
						if !errors.As(cause, &f) || len(f.Path) != 3 || f.Path[2] != (avp.Key{Code: 417}) {
							t.Fatal(cause)
						}
					} else if variant == "mixed-status" || variant == "mixed-unit" || variant == "earlier-action" {
						want := uint32(456)
						if variant == "earlier-action" {
							want = 436
						}
						var f *validation.FieldError
						if errors.Is(cause, validation.ErrMissing) || !errors.As(cause, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{Code: want}) {
							t.Fatal(cause)
						}
					}
					sendRejectRaw(t, remote, raw)
					if variant == "answer" || variant == "child-only" {
						ans := readReject(t, remote, req.Header, 5005)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, invalidPart(t, avp.Key{Code: 456}, 64, ecurUnitBytes(t, code, 0))) {
							t.Fatal(ans.FailedAVP)
						}
						noDuplicateDelivery(t, p)
						return
					}
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected reply %x", b)
					}
				})
			}
		}
	})
}
