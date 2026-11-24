// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Originator-SCCP-Address exclusion on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package diameter

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/validation"
)

func TestECUROriginatorFrozen(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, op := range []string{"init", "term"} {
		t.Run(op, func(t *testing.T) {
			src, e := os.ReadFile("testdata/codec/b43-" + op + "-origin-sccp.hex")
			if e != nil {
				t.Fatal(e)
			}
			raw, e := hex.DecodeString(string(bytes.Join(bytes.Fields(src), nil)))
			if e != nil {
				t.Fatal(e)
			}
			_, cause := co.Decode(raw)
			want := smsDataQuote(t, 2008, []byte("15550000002"))
			want[len(want)-1] = 0 // Archived encoder uses zero padding.
			q, c, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, len(want))
			if !ok || c != 5008 || !bytes.Equal(q, want) {
				t.Fatal(cause, c, q)
			}
			clear(raw)
			if !bytes.Equal(q, want) {
				t.Fatal("borrowed input")
			}
		})
	}
}
func TestECUROriginatorWire(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, op := range []string{"initial", "term"} {
		raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, op, 14101, 24101))
		for _, data := range [][]byte{nil, {0xff}, {1, 2, 3, 4}, bytes.Repeat([]byte{0xfe}, 257)} {
			t.Run(fmt.Sprintf("%s/%d", op, len(data)), func(t *testing.T) {
				frame := smsDataRaw(t, raw, op, 2008, data)
				want := smsDataQuote(t, 2008, data)
				_, err := co.Decode(frame)
				if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) {
					t.Fatal(err)
				}
				q, c, ok := excessavp.QuoteInvalid(frame[20:], err, n.cfg.Limits, len(want))
				if !ok || c != 5008 || !bytes.Equal(q, want) {
					t.Fatal(err, c, q)
				}
				if _, _, ok = excessavp.QuoteInvalid(frame[20:], err, n.cfg.Limits, len(want)-1); ok {
					t.Fatal("short quote budget")
				}
				answer, e := n.rejectFrame(frame, 4096, rejectDecode, err)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = n.rejectFrame(frame, len(answer), rejectDecode, err); e != nil {
					t.Fatal(e)
				}
				if _, e = n.rejectFrame(frame, len(answer)-1, rejectDecode, err); e == nil {
					t.Fatal("short answer budget")
				}
			})
		}
		frame := smsDataRaw(t, raw, op, 2008, []byte("ok"))
		for _, tc := range []struct {
			name string
			raw  []byte
			code uint32
			leaf avp.Key
		}{
			{"origin", equalityEdit(t, frame, ecurInterfacePath(true), make([]byte, 4)), 0, avp.Key{VendorID: 10415, Code: 2008}},
			{"destination", equalityEdit(t, frame, ecurInterfacePath(false), make([]byte, 4)), 0, avp.Key{VendorID: 10415, Code: 2008}},
			{"later-units", equalityEdit(t, frame, unitValuePath(op), unitValueData(op, 9)), 0, avp.Key{VendorID: 10415, Code: 2008}},
			{"earlier-time", equalityEdit(t, frame, recipientPath(1202), make([]byte, 4)), 0, avp.Key{VendorID: 10415, Code: 1202}},
			{"earlier-action", appendECURAction(frame, invalidPart(t, avp.Key{Code: 436}, 64, make([]byte, 4))), 0, avp.Key{Code: 436}},
			{"address-nul", smsDataRaw(t, raw, op, 2008, []byte{0}), 5004, avp.Key{VendorID: 10415, Code: 2008}},
		} {
			t.Run(op+"/"+tc.name, func(t *testing.T) {
				_, err := co.Decode(tc.raw)
				var f *validation.FieldError
				if !errors.As(err, &f) || len(f.Path) == 0 || f.Path[len(f.Path)-1] != tc.leaf {
					t.Fatal(err)
				}
				q, c, ok := excessavp.QuoteInvalid(tc.raw[20:], err, n.cfg.Limits, 4096)
				if c != tc.code || ok != (c != 0) {
					t.Fatal(c, err)
				}
				if c == 0 {
					ans, e := n.rejectFrame(tc.raw, 4096, rejectDecode, err)
					if len(ans) != 0 || !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				} else if !bytes.Equal(q, smsDataQuote(t, 2008, []byte{0})) {
					t.Fatal("wrong NUL quote", q)
				}
			})
		}
	}
}
func TestECUROriginatorNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, op := range []string{"initial", "term"} {
			for _, mode := range []string{"answer", "default", "mixed"} {
				t.Run(op+"/"+mode, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					if mode != "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, tls)
					req := combinedDuplicateRequest(t, n, co, op, 14102, 24102)
					data := []byte{255, 2, 3}
					raw := smsDataRaw(t, reqBytes(t, co, req), op, 2008, data)
					if mode == "mixed" {
						raw = equalityEdit(t, raw, ecurInterfacePath(true), make([]byte, 4))
					}
					sendRejectRaw(t, remote, raw)
					if mode == "answer" {
						ans := readReject(t, remote, req.Header, 5008)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, smsDataQuote(t, 2008, data)) {
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
