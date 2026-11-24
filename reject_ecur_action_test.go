// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Requested-Action exclusion for combined ECUR requests.
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
	"github.com/skill215/diameterkit/wire"
)

func appendECURAction(raw, action []byte) []byte {
	out := append(raw[:len(raw):len(raw)], action...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}

func TestECURActionFrozen(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, op := range []string{"init", "term"} {
		t.Run(op, func(t *testing.T) {
			src, err := os.ReadFile("testdata/codec/b43-" + op + "-action.hex")
			if err != nil {
				t.Fatal(err)
			}
			raw, err := hex.DecodeString(string(bytes.Join(bytes.Fields(src), nil)))
			if err != nil {
				t.Fatal(err)
			}
			_, cause := co.Decode(raw)
			want := invalidPart(t, avp.Key{Code: 436}, 64, make([]byte, 4))
			q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 12)
			if !ok || code != 5008 || !bytes.Equal(q, want) {
				t.Fatal(cause, code, q)
			}
			clear(raw)
			if !bytes.Equal(q, want) {
				t.Fatal("borrowed action")
			}
		})
	}
}

func TestECURActionOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term"} {
		raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 14001, 24001))
		action := invalidPart(t, avp.Key{Code: 436}, 64, []byte{1, 2, 3, 4})
		for _, tc := range []struct {
			name string
			raw  []byte
			code uint32
			path avp.Key
		}{
			{"isolated", appendECURAction(raw, action), 5008, avp.Key{Code: 436}},
			{"later-unit", appendECURAction(equalityEdit(t, raw, unitValuePath(kind), unitValueData(kind, 9)), action), 0, avp.Key{Code: 436}},
			{"earlier-time", appendECURAction(equalityEdit(t, raw, recipientPath(1202), make([]byte, 4)), action), 0, avp.Key{VendorID: 10415, Code: 1202}},
			{"earlier-sequence", appendECURAction(equalityEdit(t, raw, lengthPath(avp.Key{Code: 415}), []byte{0, 0, 0, 8}), action), 0, avp.Key{Code: 416}},
			{"earlier-equipment-nul", appendECURAction(equipmentDataRaw(t, raw, kind, []byte{0}), action), 5004, avp.Key{Code: 460}},
			{"action-length", appendECURAction(raw, invalidPart(t, avp.Key{Code: 436}, 64, []byte{1})), 5014, avp.Key{Code: 436}},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				_, cause := co.Decode(tc.raw)
				var f *validation.FieldError
				if !errors.As(cause, &f) || len(f.Path) == 0 || f.Path[len(f.Path)-1] != tc.path {
					t.Fatal(cause)
				}
				q, c, ok := excessavp.QuoteInvalid(tc.raw[20:], cause, n.cfg.Limits, 4096)
				if c != tc.code || ok != (c != 0) {
					t.Fatal(cause, c, ok)
				}
				reply, e := n.rejectFrame(tc.raw, 4096, rejectDecode, cause)
				if c == 0 {
					if !errors.Is(e, ErrUnsupported) || len(reply) != 0 {
						t.Fatal(e, reply)
					}
					return
				}
				if e != nil {
					t.Fatal(e)
				}
				if c == 5008 {
					if !bytes.Equal(q, action) {
						t.Fatal(q)
					}
					if _, _, ok := excessavp.QuoteInvalid(tc.raw[20:], cause, n.cfg.Limits, 11); ok {
						t.Fatal("short quote")
					}
					if _, e = n.rejectFrame(tc.raw, len(reply), rejectDecode, cause); e != nil {
						t.Fatal(e)
					}
					if _, e = n.rejectFrame(tc.raw, len(reply)-1, rejectDecode, cause); e == nil {
						t.Fatal("short answer")
					}
				}
			})
		}
	}
}

func TestECURActionNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"initial", "term"} {
			for _, mode := range []string{"answer", "default", "mixed"} {
				t.Run(kind+"/"+mode, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					if mode != "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, tls)
					req := combinedDuplicateRequest(t, n, co, kind, 14002, 24002)
					raw := reqBytes(t, co, req)
					if mode == "mixed" {
						raw = equalityEdit(t, raw, unitValuePath(kind), unitValueData(kind, 9))
					}
					want := invalidPart(t, avp.Key{Code: 436}, 64, []byte{255, 255, 255, 255})
					sendRejectRaw(t, remote, appendECURAction(raw, want))
					if mode == "answer" {
						ans := readReject(t, remote, req.Header, 5008)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) {
							t.Fatal(ans.FailedAVP)
						}
						noDuplicateDelivery(t, p)
						return
					}
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected answer %x", b)
					}
				})
			}
		}
	})
}
