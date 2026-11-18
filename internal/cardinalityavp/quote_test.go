// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Count, order, ownership, budget and malformed-input acceptance.
// SPDX-AI-Date: 2026-11-18

package cardinalityavp

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type proof struct {
	name  string
	count int
}

func (*proof) Error() string                                 { return "cardinality" }
func (*proof) Unwrap() error                                 { return validation.ErrMissing }
func (p *proof) DiameterMissingCardinalityV1() (string, int) { return p.name, p.count }
func encoded(t *testing.T, r rule, types []uint32, value []byte) []byte {
	t.Helper()
	write := func(w *wire.Encoder) {
		for _, typ := range types {
			g := w.BeginGroup(r.group, r.flags)
			w.Uint32(r.discriminator, r.flags, typ)
			w.Bytes(r.value, r.flags, value)
			w.EndGroup(g)
		}
	}
	w := wire.NewEncoder(nil, wire.DefaultLimits())
	write(w)
	out := make([]byte, w.Len())
	w = wire.NewEncoder(out, wire.DefaultLimits())
	write(w)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	return out
}
func TestCardinalityCombined(t *testing.T) {
	r, _ := selected("combined-subscriptions-v1")
	for _, types := range [][]uint32{nil, {0}, {1}} {
		raw := encoded(t, r, types, []byte("secret"))
		cause := validation.At(r.group, &proof{"combined-subscriptions-v1", len(types)})
		q, ok := Quote(raw, cause, wire.DefaultLimits(), 56)
		want := encoded(t, r, make([]uint32, 2-len(types)), nil)
		if !ok || !bytes.Equal(q, want) {
			t.Fatal(types, q)
		}
		clear(raw)
		if !bytes.Equal(q, want) {
			t.Fatal("borrowed")
		}
	}
	for _, types := range [][]uint32{{0, 1}, {1, 0}, {0, 0}, {1, 1}, {2}, {0xffffffff}} {
		raw := encoded(t, r, types, nil)
		if q, ok := Quote(raw, validation.At(r.group, &proof{"combined-subscriptions-v1", len(types)}), wire.DefaultLimits(), 1024); ok || q != nil {
			t.Fatal("unsafe", types, q)
		}
	}
}
func TestCardinalityRefusal(t *testing.T) {
	r, _ := selected("combined-subscriptions-v1")
	for _, kind := range []string{"budget", "field", "frame", "avps", "depth", "repeated", "path", "rule", "count", "negative", "bare", "trailing", "nul", "truncated"} {
		t.Run(kind, func(t *testing.T) {
			p := &proof{"combined-subscriptions-v1", 0}
			l := wire.DefaultLimits()
			budget := 56
			var raw []byte
			var cause error = validation.At(r.group, p)
			switch kind {
			case "budget":
				budget--
			case "field":
				l.MaxFieldBytes = 55
			case "frame":
				l.MaxFrameBytes = 83
				l.MaxFieldBytes = 83
			case "avps":
				l.MaxAVPs = 6
				l.MaxRepeated = 6
			case "depth":
				l.MaxDepth = 1
			case "repeated":
				l.MaxRepeated = 1
			case "path":
				cause = validation.At(avp.Key{Code: 263}, p)
			case "rule":
				p.name = "unreviewed"
			case "count":
				p.count = 1
			case "negative":
				p.count = -1
			case "bare":
				cause = validation.At(r.group, validation.ErrMissing)
			case "trailing":
				raw = []byte{0}
			case "nul":
				raw = encoded(t, r, []uint32{0}, []byte{0})
				p.count = 1
			case "truncated":
				raw = encoded(t, r, []uint32{0}, nil)
				raw = raw[:len(raw)-1]
				p.count = 1
			}
			if q, ok := Quote(raw, cause, l, budget); ok || q != nil {
				t.Fatal(kind, q)
			}
		})
	}
	if !HasEvidence(validation.At(r.group, &proof{})) || HasEvidence(errors.New("bare")) {
		t.Fatal("evidence classification")
	}
}
func FuzzCardinalityQuote(f *testing.F) {
	f.Add([]byte{}, int8(0), uint16(56), false)
	f.Add([]byte{}, int8(0), uint16(440), true)
	f.Add([]byte{0}, int8(1), uint16(55), true)
	f.Fuzz(func(t *testing.T, raw []byte, count int8, budget uint16, other bool) {
		if len(raw) > 8192 {
			return
		}
		name := "combined-subscriptions-v1"
		if other {
			name = "scap1-parameters-v1"
		}
		r, known := selected(name)
		key := r.group
		if !known {
			key = avp.Key{Code: 607, VendorID: 193}
		}
		cause := validation.At(key, &proof{name, int(count)})
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 8192
		l.MaxFieldBytes = 4096
		l.MaxAVPs = 128
		l.MaxRepeated = 16
		q, ok := Quote(raw, cause, l, int(budget))
		if !ok {
			if q != nil {
				t.Fatal("partial output")
			}
			return
		}
		if !known || int(count) < 0 || int(count) >= len(r.types) || len(q) > int(budget) || len(q) > 440 {
			t.Fatal("unbounded quote")
		}
		want := encoded(t, r, make([]uint32, len(r.types)-int(count)), nil)
		if !bytes.Equal(q, want) {
			t.Fatal("nonminimum quote")
		}
	})
}
