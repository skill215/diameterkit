// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit SCAP1 ordered omission diagnostics and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package cardinalityavp

import (
	"bytes"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestCardinalitySCAPLimits(t *testing.T) {
	r, _ := selected("scap1-parameters-v1")
	cause := validation.At(r.group, &proof{"scap1-parameters-v1", 0})
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 468
	l.MaxFieldBytes = 440
	l.MaxAVPs = 34
	l.MaxRepeated = 11
	l.MaxDepth = 2
	want := encoded(t, r, make([]uint32, 11), nil)
	q, ok := Quote(nil, cause, l, 440)
	if !ok || len(q) != 440 || !bytes.Equal(q, want) {
		t.Fatal("exact limits", q)
	}
	for _, kind := range []string{"bytes", "frame", "field", "avps", "repeat", "depth"} {
		t.Run(kind, func(t *testing.T) {
			limit := l
			budget := 440
			switch kind {
			case "bytes":
				budget--
			case "frame":
				limit.MaxFrameBytes--
			case "field":
				limit.MaxFieldBytes--
			case "avps":
				limit.MaxAVPs--
			case "repeat":
				limit.MaxRepeated--
			case "depth":
				limit.MaxDepth--
			}
			if q, ok := Quote(nil, cause, limit, budget); ok || q != nil {
				t.Fatal("limit bypass", q)
			}
		})
	}
	for _, types := range [][]uint32{{0, 1}, {3, 13}, {0, 5, 17, 2}, {13}} {
		raw := encoded(t, r, types, []byte("private-value"))
		q, ok := Quote(raw, validation.At(r.group, &proof{"scap1-parameters-v1", len(types)}), wire.DefaultLimits(), 440)
		if !ok || !bytes.Equal(q, encoded(t, r, make([]uint32, 11-len(types)), nil)) {
			t.Fatal(types, q)
		}
		clear(raw)
		if bytes.Contains(q, []byte("private-value")) {
			t.Fatal("copied value")
		}
	}
	for _, types := range [][]uint32{{1, 0}, {0, 0}, {13, 3}, {0, 42}, {0xffffffff}, r.types} {
		raw := encoded(t, r, types, nil)
		if q, ok := Quote(raw, validation.At(r.group, &proof{"scap1-parameters-v1", len(types)}), wire.DefaultLimits(), 440); ok || q != nil {
			t.Fatal("unsafe type sequence", types)
		}
	}
}
