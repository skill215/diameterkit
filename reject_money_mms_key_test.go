// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded combined money MMS exclusion and acceptance.
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

func moneyMMSChildren(t *testing.T, mask int, key []byte) []byte {
	part := func(c uint32, b []byte) []byte { return invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, b) }
	var children []byte
	for i, c := range [...]uint32{1201, 1202, 1209, 1212, 1210, 1216} {
		if mask&(1<<i) == 0 {
			continue
		}
		data := []byte{0, 0, 0, 1}
		switch c {
		case 1201:
			data = append(part(899, data), part(897, []byte("dest"))...)
		case 1210:
			data = key
		}
		children = append(children, part(c, data)...)
	}
	return children
}

func TestMoneyMMSWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance"} {
		original := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13801, 23801))
		key := []byte("key")
		valid := mmsKeyRaw(t, original, kind, key)
		_, proof := co.Decode(valid)
		for mask := 0; mask < 64; mask++ {
			t.Run(fmt.Sprintf("%s/%02x", kind, mask), func(t *testing.T) {
				raw := unitAppendRaw(t, original, mmsGroupPath(), moneyMMSChildren(t, mask, key))
				_, err := co.Decode(raw)
				// The consumer independently rejects a mismatched wire image even with genuine evidence.
				q, c, ok := excessavp.QuoteInvalid(raw[20:], proof, n.cfg.Limits, 4096)
				if mask == 16 {
					if !ok || c != 5008 || !bytes.Equal(q, mmsKeyQuote(t, key)) {
						t.Fatal(c, q)
					}
				} else if ok {
					t.Fatal("accepted mismatched evidence", c)
				}
				if mask == 0 {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if !errors.Is(err, validation.ErrUnsupported) {
					t.Fatal(err)
				}
				_, e := n.rejectFrame(raw, 4096, rejectDecode, err)
				if mask == 16 {
					if e != nil {
						t.Fatal(e)
					}
				} else if !errors.Is(e, ErrUnsupported) {
					t.Fatal("mixed/unsupported mapping", e)
				}
			})
		}
	}
}

func TestMoneyMMSPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance"} {
		req := combinedDuplicateRequest(t, n, co, kind, 13802, 23802)
		badClient, _ := invalidRaw(t, co, req, smsPath(2018), []byte{1}, avp.Key{})
		for _, tc := range []struct {
			name      string
			raw, want []byte
		}{
			{"client-before-priority", unitAppendRaw(t, badClient, mmsGroupPath(), moneyMMSChildren(t, 4, nil)), moneyClientQuote(t, []byte{1})},
			{"client-before-key", mmsKeyRaw(t, badClient, kind, []byte("ok")), moneyClientQuote(t, []byte{1})},
			{"group-before-client", mmsKeyRaw(t, badClient, kind, []byte{0}), mmsKeyQuote(t, []byte{0})},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				_, err := co.Decode(tc.raw)
				q, c, ok := excessavp.QuoteInvalid(tc.raw[20:], err, n.cfg.Limits, 4096)
				if !ok || c != 5004 || !bytes.Equal(q, tc.want) {
					t.Fatal(err, c, q)
				}
			})
		}
	}
}

func TestRejectMoneyMMSKey(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance"} {
			for _, key := range [][]byte{nil, []byte("abc"), bytes.Repeat([]byte{'k'}, 257)} {
				t.Run(fmt.Sprintf("%s/%d", kind, len(key)), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					cfg.Peers[0].InboundCommands[0].MaxRequestBytes = 4096
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 4096
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := combinedDuplicateRequest(t, n, co, kind, 13803, 23803)
					raw := mmsKeyRaw(t, reqBytes(t, co, req), kind, key)
					_, err := co.Decode(raw)
					want := mmsKeyQuote(t, key)
					q, c, ok := excessavp.QuoteInvalid(raw[20:], err, cfg.Limits, len(want))
					if !ok || c != 5008 || !bytes.Equal(q, want) {
						t.Fatal(err, c, q)
					}
					if _, _, ok := excessavp.QuoteInvalid(raw[20:], err, cfg.Limits, len(want)-1); ok {
						t.Fatal("short quote budget")
					}
					reply, e := n.rejectFrame(raw, 4096, rejectDecode, err)
					if e != nil {
						t.Fatal(e)
					}
					if _, e = n.rejectFrame(raw, len(reply), rejectDecode, err); e != nil {
						t.Fatal(e)
					}
					if _, e = n.rejectFrame(raw, len(reply)-1, rejectDecode, err); e == nil {
						t.Fatal("short answer budget")
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5008)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) {
						t.Fatal(ans.FailedAVP)
					}
					clear(raw)
					if !bytes.Equal(q, want) {
						t.Fatal("borrowed bytes")
					}
					noDuplicateDelivery(t, p)
				})
			}
			for _, mask := range []int{16, 4, 20} {
				t.Run(fmt.Sprintf("%s/close-%d", kind, mask), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if mask == 16 {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13804, 23804))
					sendRejectRaw(t, remote, unitAppendRaw(t, raw, mmsGroupPath(), moneyMMSChildren(t, mask, []byte("ok"))))
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected reply %x", b)
					}
				})
			}
		}
	})
}

func TestMoneyMMSFrozenNegative(t *testing.T) {
	text, err := os.ReadFile("testdata/codec/b47-general-mms-key.hex")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := hex.DecodeString(string(bytes.TrimSpace(text)))
	if err != nil {
		t.Fatal(err)
	}
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	_, cause := co.Decode(raw)
	q, c, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
	if !ok || c != 5008 || !bytes.Equal(q, mmsKeyQuote(t, []byte("sms-0001"))) {
		t.Fatal(cause, c, q)
	}
}
