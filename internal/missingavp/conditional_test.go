// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional evidence, exact minimum bytes and refusal boundaries.
// SPDX-AI-Date: 2026-11-18

package missingavp

import (
	"bytes"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type conditionalProof struct {
	*omission
	rule      string
	condition uint32
}

func (p *conditionalProof) DiameterMissingConditionalV1() (string, uint32) {
	return p.rule, p.condition
}

type mixedConditionalProof struct{ *conditionalProof }

func (*mixedConditionalProof) DiameterMissingVariableV1() string { return "octets" }

func TestConditionalMinimum(t *testing.T) {
	for _, kind := range []string{"exact", "reserved-receive-bits", "budget", "field", "frame", "count", "depth", "rule", "condition", "width", "key", "flags", "path", "parent", "child", "legacy", "bare", "mixed-type"} {
		t.Run(kind, func(t *testing.T) {
			key := avp.Key{VendorID: 10415, Code: 3309}
			proof := &conditionalProof{omission: &omission{key: key, flags: 192, width: 3}, rule: "s6c-srr-gprs-smea-v1", condition: 1}
			f := &validation.FieldError{Path: []avp.Key{key}, Cause: proof}
			var cause error = f
			l, budget := wire.DefaultLimits(), 16
			switch kind {
			case "reserved-receive-bits":
				proof.condition = 0x80000001
			case "budget":
				budget--
			case "field":
				l.MaxFieldBytes = 15
			case "frame":
				l.MaxFrameBytes = 43
				l.MaxFieldBytes = 43
			case "count":
				l.MaxAVPs = 1
				l.MaxRepeated = 1
			case "depth":
				l.MaxDepth = 0
			case "rule":
				proof.rule = "unknown"
			case "condition":
				proof.condition = 2
			case "width":
				proof.width = 4
			case "key":
				proof.key.Code++
			case "flags":
				proof.flags = 128
			case "path":
				f.Path[0].Code = 263
			case "parent":
				cause = validation.At(avp.Key{Code: 284}, &omission{key: avp.Key{Code: 284}, flags: 64, child: proof})
				cause.(*validation.FieldError).Path = append([]avp.Key{{Code: 284}}, key)
			case "child":
				proof.width = 0
				proof.child = &omission{key: avp.Key{Code: 266}, flags: 64, width: 4}
				f.Path = append(f.Path, avp.Key{Code: 266})
			case "legacy":
				f.Cause = proof.omission
			case "bare":
				f.Cause = validation.ErrMissing
			case "mixed-type":
				f.Cause = &mixedConditionalProof{proof}
			}
			quote, ok := Quote(cause, l, budget)
			wantOK := kind == "exact" || kind == "reserved-receive-bits"
			if ok != wantOK || !ok && quote != nil {
				t.Fatal(kind, quote, ok)
			}
			if !ok {
				return
			}
			want := []byte{0, 0, 12, 237, 192, 0, 0, 15, 0, 0, 40, 175, 0, 0, 0, 0}
			if !bytes.Equal(quote, want) {
				t.Fatalf("quote %x want %x", quote, want)
			}
			quote[12] = 99
			again, ok := Quote(cause, l, budget)
			if !ok || !bytes.Equal(again, want) {
				t.Fatal("shared output", again)
			}
		})
	}
}

func FuzzConditionalMinimum(f *testing.F) {
	f.Add(uint32(1), int16(3), uint8(192), uint16(16), true)
	f.Add(uint32(0), int16(4), uint8(192), uint16(15), false)
	f.Fuzz(func(t *testing.T, condition uint32, width int16, flags uint8, budget uint16, known bool) {
		key := avp.Key{VendorID: 10415, Code: 3309}
		rule := "unknown"
		if known {
			rule = "s6c-srr-gprs-smea-v1"
		}
		p := &conditionalProof{omission: &omission{key: key, flags: flags, width: int(width)}, rule: rule, condition: condition}
		quote, ok := Quote(validation.At(key, p), wire.DefaultLimits(), int(budget))
		wantOK := known && condition&1 != 0 && width == 3 && flags == 192 && budget >= 16
		if ok != wantOK || !ok && quote != nil {
			t.Fatal("conditional boundary", quote, ok)
		}
		if ok && (len(quote) != 16 || wire.Uint24(quote[5:8]) != 15 || !bytes.Equal(quote[12:], []byte{0, 0, 0, 0})) {
			t.Fatal("unsafe quote", quote)
		}
	})
}
