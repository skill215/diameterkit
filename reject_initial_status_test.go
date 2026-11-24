// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded SM-Status exclusion on combined ECUR INITIAL requests.
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

func appendStatusField(t *testing.T, raw []byte, code uint32, data []byte) []byte {
	t.Helper()
	return unitAppendRaw(t, raw, smsPath(code)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, data))
}
func TestInitialStatusFrozen(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	src, err := os.ReadFile("testdata/codec/b43-init-status.hex")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := hex.DecodeString(string(bytes.Join(bytes.Fields(src), nil)))
	if err != nil {
		t.Fatal(err)
	}
	_, cause := co.Decode(raw)
	want := smsDataQuote(t, 2014, []byte{0})
	clear(want[len(want)-3:]) // Archived fixture uses zero padding.
	q, c, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, len(want))
	clear(raw)
	if !ok || c != 5008 || !bytes.Equal(q, want) {
		t.Fatal(cause, c, q)
	}
}
func TestInitialStatusWire(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	original := reqBytes(t, co, combinedDuplicateRequest(t, n, co, "initial", 14201, 24201))
	for _, payload := range [][]byte{nil, {0}, {0xff, 0, 1}, bytes.Repeat([]byte{0x80}, 257)} {
		t.Run(fmt.Sprint(len(payload)), func(t *testing.T) {
			raw := appendStatusField(t, original, 2014, payload)
			_, cause := co.Decode(raw)
			want := smsDataQuote(t, 2014, payload)
			q, c, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, len(want))
			if !ok || c != 5008 || !bytes.Equal(q, want) {
				t.Fatal(cause, c, q)
			}
			if _, _, ok = excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, len(want)-1); ok {
				t.Fatal("short quote budget")
			}
			reply, e := n.rejectFrame(raw, 4096, rejectDecode, cause)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply), rejectDecode, cause); e != nil {
				t.Fatal(e)
			}
			if _, e = n.rejectFrame(raw, len(reply)-1, rejectDecode, cause); e == nil {
				t.Fatal("short answer budget")
			}
		})
	}
	status := appendStatusField(t, original, 2014, []byte{0})
	_, proof := co.Decode(status)
	discharge := appendStatusField(t, original, 2012, make([]byte, 4))
	both := appendStatusField(t, status, 2012, make([]byte, 4))
	for _, tc := range []struct {
		name  string
		raw   []byte
		leaf  avp.Key
		quote []byte
	}{
		{"discharge", discharge, avp.Key{VendorID: 10415, Code: 2000}, nil},
		{"both", both, avp.Key{VendorID: 10415, Code: 2000}, nil},
		{"reversed", appendStatusField(t, discharge, 2014, nil), avp.Key{VendorID: 10415, Code: 2000}, nil},
		{"used-unit", unitAppendRaw(t, status, lengthPath(avp.Key{Code: 456}), invalidPart(t, avp.Key{Code: 446}, 64, invalidPart(t, avp.Key{Code: 417}, 64, make([]byte, 8)))), avp.Key{Code: 456}, nil},
		{"invalid-unit", equalityEdit(t, status, unitValuePath("initial"), unitValueData("initial", 9)), avp.Key{Code: 417}, unitValueQuote(t, "initial", 9)},
		{"invalid-unit-both", equalityEdit(t, both, unitValuePath("initial"), unitValueData("initial", 0)), avp.Key{Code: 417}, unitValueQuote(t, "initial", 0)},
		{"originator", smsDataRaw(t, status, "initial", 2008, []byte("ok")), avp.Key{VendorID: 10415, Code: 2008}, nil},
		{"action", appendECURAction(status, invalidPart(t, avp.Key{Code: 436}, 64, make([]byte, 4))), avp.Key{Code: 436}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := co.Decode(tc.raw)
			var f *validation.FieldError
			if !errors.As(err, &f) || len(f.Path) == 0 || f.Path[len(f.Path)-1] != tc.leaf {
				t.Fatal(err)
			}
			if _, _, ok := excessavp.QuoteInvalid(tc.raw[20:], proof, n.cfg.Limits, 4096); ok {
				t.Fatal("matched status proof on mixed wire image")
			}
			q, c, ok := excessavp.QuoteInvalid(tc.raw[20:], err, n.cfg.Limits, 4096)
			if tc.quote != nil {
				if !ok || c != 5004 || !bytes.Equal(q, tc.quote) {
					t.Fatal(c, q, err)
				}
				return
			}
			ans, e := n.rejectFrame(tc.raw, 4096, rejectDecode, err)
			if ok || len(ans) != 0 || !errors.Is(e, ErrUnsupported) {
				t.Fatal(err, c, e)
			}
		})
	}
}
func TestInitialStatusNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, mode := range []string{"answer", "default", "mixed", "unit"} {
			t.Run(mode, func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				if mode == "default" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				n, co := newInboundNode(t, cfg)
				defer closeDuplicateNode(t, n)
				p, remote := networkConnectNode(t, n, co, tls)
				req := combinedDuplicateRequest(t, n, co, "initial", 14202, 24202)
				data := []byte{0, 255, 2}
				raw := appendStatusField(t, reqBytes(t, co, req), 2014, data)
				if mode == "mixed" {
					raw = appendStatusField(t, raw, 2012, make([]byte, 4))
				}
				if mode == "unit" {
					raw = equalityEdit(t, raw, unitValuePath("initial"), unitValueData("initial", 7))
				}
				sendRejectRaw(t, remote, raw)
				if mode == "answer" || mode == "unit" {
					code := uint32(5008)
					want := smsDataQuote(t, 2014, data)
					if mode == "unit" {
						code = 5004
						want = unitValueQuote(t, "initial", 7)
					}
					ans := readReject(t, remote, req.Header, code)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) {
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
	})
}
