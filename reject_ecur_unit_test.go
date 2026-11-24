// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group exclusions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package diameter

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/validation"
)

func ecurUnitBytes(t *testing.T, code uint32, value uint64) []byte {
	t.Helper()
	return invalidPart(t, avp.Key{Code: code}, 64, invalidPart(t, avp.Key{Code: 417}, 64, unitValueData("initial", value)))
}

func TestECURUnitFrozen(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for name, code := range map[string]uint32{"b43-init-usu": 446, "b43-term-rsu": 437} {
		t.Run(name, func(t *testing.T) {
			src, e := os.ReadFile("testdata/codec/" + name + ".hex")
			if e != nil {
				t.Fatal(e)
			}
			raw, e := hex.DecodeString(string(bytes.Join(bytes.Fields(src), nil)))
			if e != nil {
				t.Fatal(e)
			}
			_, err := co.Decode(raw)
			want := invalidPart(t, avp.Key{Code: 456}, 64, ecurUnitBytes(t, code, 1))
			q, c, ok := excessavp.QuoteInvalid(raw[20:], err, n.cfg.Limits, 32)
			if _, _, short := excessavp.QuoteInvalid(raw[20:], err, n.cfg.Limits, 31); short {
				t.Fatal("short group quote accepted")
			}
			reply, e := n.rejectFrame(raw, 4096, rejectDecode, err)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply), rejectDecode, err); e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply)-1, rejectDecode, err); e == nil {
				t.Fatal("short answer accepted")
			}
			clear(raw)
			if !ok || c != 5008 || !bytes.Equal(q, want) {
				t.Fatal(err, c, q)
			}
		})
	}
}

func TestECURUnitNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for op, code := range map[string]uint32{"initial": 446, "term": 437} {
			for _, policy := range []string{"answer", "default", "mixed", "earlier-action"} {
				t.Run(op+"/"+policy, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if policy == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, tls)
					req := combinedDuplicateRequest(t, n, co, op, 14302, 24302)
					group := ecurUnitBytes(t, code, 0xfedcba9876543210)
					raw := unitAppendRaw(t, reqBytes(t, co, req), lengthPath(avp.Key{Code: 456}), group)
					if policy == "mixed" {
						raw = equalityEdit(t, raw, unitValuePath(op), unitValueData(op, 9))
					}
					if policy == "earlier-action" {
						raw = appendECURAction(raw, invalidPart(t, avp.Key{Code: 436}, 64, make([]byte, 4)))
					}
					if policy == "mixed" || policy == "earlier-action" {
						_, err := co.Decode(raw)
						var f *validation.FieldError
						want := uint32(456)
						if policy == "earlier-action" {
							want = 436
						}
						if !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{Code: want}) {
							t.Fatal(err)
						}
					}
					sendRejectRaw(t, remote, raw)
					if policy == "answer" {
						ans := readReject(t, remote, req.Header, 5008)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, invalidPart(t, avp.Key{Code: 456}, 64, group)) {
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
