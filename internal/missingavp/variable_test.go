// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit variable minimum provenance, padding, budgets and fuzz.
// SPDX-AI-Date: 2026-11-17

package missingavp

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type variableProof struct {
	*omission
	kind string
}

func (p *variableProof) DiameterMissingVariableV1() string { return p.kind }
func TestVariableMinimumQuote(t *testing.T) {
	for _, kind := range []string{"identity", "tbcd", "utf8", "utf8-nonempty", "octets", "octets-nonempty"} {
		t.Run(kind, func(t *testing.T) {
			width := 1
			if kind == "utf8" || kind == "octets" {
				width = 0
			}
			leaf := avp.Key{VendorID: 193, Code: 9002}
			parent := avp.Key{VendorID: 193, Code: 9001}
			for _, negative := range []string{"exact", "bytes", "field", "frame", "count", "depth", "bare", "legacy", "wrong-minimum", "unknown-kind", "path", "flags", "typed-parent"} {
				t.Run(negative, func(t *testing.T) {
					p := &variableProof{omission: &omission{key: leaf, flags: 128, width: width}, kind: kind}
					outer := &variableProof{omission: &omission{key: parent, flags: 192, child: p}}
					f := &validation.FieldError{Path: []avp.Key{parent, leaf}, Cause: outer}
					var cause error = f
					l := wire.DefaultLimits()
					size := 24 + ((width + 3) &^ 3)
					budget := size
					switch negative {
					case "bytes":
						budget--
					case "field":
						l.MaxFieldBytes = size - 1
					case "frame":
						l.MaxFrameBytes = size + 27
						l.MaxFieldBytes = l.MaxFrameBytes
					case "count":
						l.MaxAVPs = 2
						l.MaxRepeated = 2
					case "depth":
						l.MaxDepth = 1
					case "bare":
						cause = validation.At(parent, validation.At(leaf, validation.ErrMissing))
					case "legacy":
						outer.child = p.omission
					case "wrong-minimum":
						p.width = 1 - width
					case "unknown-kind":
						p.kind = "domain"
					case "path":
						f.Path[1].Code++
					case "flags":
						p.flags = 0
					case "typed-parent":
						outer.kind = "octets"
					}
					q, ok := Quote(cause, l, budget)
					if ok != (negative == "exact") || !ok && q != nil {
						t.Fatal("unsafe minimum", negative, q)
					}
					if !ok {
						return
					}
					if len(q) != size {
						t.Fatal(len(q))
					}
					d := wire.NewDecoder(q, l)
					v, _, e := d.Next()
					if e != nil || v.Key != parent || v.Flags != 192 || d.Remaining() != 0 {
						t.Fatal(v, e)
					}
					child, _ := d.Child(v.Data)
					v, found, e := child.Next()
					if e != nil || !found || v.Key != leaf || v.Flags != 128 || !bytes.Equal(v.Data, make([]byte, width)) || child.Remaining() != 0 {
						t.Fatal(v, e)
					}
					if got := wire.Uint24(q[17:20]); got != uint32(12+width) {
						t.Fatal("padded length became payload", got)
					}
					if width == 1 && !bytes.Equal(q[len(q)-3:], []byte{0, 0, 0}) {
						t.Fatal("nonzero padding")
					}
				})
			}
		})
	}
	for _, width := range []int{-1, 2, 4, 8, 1024} {
		p := &variableProof{omission: &omission{key: avp.Key{Code: 1}, width: width}, kind: "identity"}
		if _, ok := Quote(validation.At(p.key, p), wire.DefaultLimits(), 4096); ok {
			t.Fatal("unproven width", width)
		}
	}
}
func FuzzVariableMinimum(f *testing.F) {
	f.Add(uint8(0), int16(1), uint16(28), uint8(2), uint8(128))
	f.Add(uint8(2), int16(0), uint16(24), uint8(2), uint8(128))
	f.Fuzz(func(t *testing.T, selector uint8, width int16, budget uint16, depth, flags uint8) {
		kinds := []string{"identity", "tbcd", "utf8", "utf8-nonempty", "octets", "octets-nonempty", "", "unknown"}
		kind := kinds[int(selector)%len(kinds)]
		key := avp.Key{VendorID: 193, Code: 9002}
		p := &variableProof{omission: &omission{key: key, flags: flags, width: int(width)}, kind: kind}
		cause := validation.At(key, p)
		l := wire.DefaultLimits()
		l.MaxDepth = int(depth % 65)
		l.MaxAVPs = 16
		l.MaxRepeated = 8
		q, ok := Quote(cause, l, int(budget))
		if !ok {
			if q != nil {
				t.Fatal("partial quote")
			}
			return
		}
		if len(q) > int(budget) || len(q) > 20 {
			t.Fatal("unbounded output")
		}
		d := wire.NewDecoder(q, l)
		v, found, e := d.Next()
		if e != nil || !found || v.Key != key || len(v.Data) != int(width) || d.Remaining() != 0 || !bytes.Equal(v.Data, make([]byte, int(width))) {
			t.Fatal("invalid output", e)
		}
	})
}
