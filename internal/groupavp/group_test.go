// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Constraint provenance, omission minima, budgets and bounded fuzz.
// SPDX-AI-Date: 2026-11-17

package groupavp_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/groupavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func part(code uint32, flags byte, data []byte) []byte {
	out := make([]byte, (8+len(data)+3)&^3)
	w := wire.NewEncoder(out, wire.DefaultLimits())
	w.Bytes(avp.Key{Code: code}, flags, data)
	return out
}
func body(kind string) []byte {
	vendor := part(266, 64, []byte{0, 0, 40, 175})
	switch kind {
	case "both":
		return append(append(part(259, 0, []byte{0, 0, 0, 9}), vendor...), part(258, 64, []byte{1, 0, 0, 1})...)
	case "neither":
		return vendor
	case "host":
		return part(33, 64, []byte{0, 255, 1})
	case "empty":
		return nil
	case "state":
		return part(280, 64, []byte("p"))
	}
	panic(kind)
}
func decode(src []byte, kind string, l wire.Limits) error {
	d := wire.NewDecoder(src, l)
	if kind == "both" || kind == "neither" {
		_, e := base.DecodeVSAI(d)
		return e
	}
	_, e := base.DecodeProxy(d)
	return e
}
func TestGroupConstraints(t *testing.T) {
	for _, kind := range []string{"both", "neither", "host", "empty", "state"} {
		t.Run(kind, func(t *testing.T) {
			src := body(kind)
			l := wire.DefaultLimits()
			e := decode(src, kind, l)
			q, code, ok := groupavp.Quote(src, e, l, 24)
			if !ok {
				t.Fatal(e)
			}
			var want []byte
			switch kind {
			case "both":
				want = append(part(259, 0, []byte{0, 0, 0, 9}), part(258, 64, []byte{1, 0, 0, 1})...)
				if code != 5009 || !errors.Is(e, validation.ErrField) || !errors.Is(e, validation.ErrDuplicate) {
					t.Fatal(code, e)
				}
			case "neither":
				want = append(part(258, 64, make([]byte, 4)), part(259, 64, make([]byte, 4))...)
			case "host", "empty":
				want = part(280, 64, []byte{0})
			case "state":
				want = part(33, 64, nil)
			}
			if kind != "both" && (code != 5005 || !errors.Is(e, validation.ErrMissing)) {
				t.Fatal(code, e)
			}
			if !bytes.Equal(q, want) {
				t.Fatalf("quote %x want %x", q, want)
			}
			if x, c, ok := groupavp.Quote(src, e, l, len(q)-1); ok || x != nil || c != 0 {
				t.Fatal("reply bound")
			}
			clear(src)
			if !bytes.Equal(q, want) {
				t.Fatal("borrowed diagnostic")
			}
		})
	}
	valid := append(part(280, 64, []byte("p")), part(33, 64, nil)...)
	if e := decode(valid, "state", wire.DefaultLimits()); e != nil {
		t.Fatal("empty State is present", e)
	}
}

type proof struct {
	key          avp.Key
	flags        byte
	offset       int
	rule         string
	child, cause error
}

func (p *proof) Error() string { return "test metadata" }
func (p *proof) Unwrap() error { return p.cause }
func (p *proof) DiameterGroupAVPV1() (avp.Key, byte, int, string, error) {
	return p.key, p.flags, p.offset, p.rule, p.child
}
func TestGroupProofBounds(t *testing.T) {
	for _, kind := range []string{"valid", "budget", "negative-budget", "field", "count", "depth", "frame", "repeat", "bare", "path", "offset", "flags", "rule", "parent", "sentinel", "terminal-key", "terminal-child", "changed-condition", "duplicate", "width", "mandatory", "extras", "truncated"} {
		t.Run(kind, func(t *testing.T) {
			l := wire.DefaultLimits()
			budget := 64
			data := body("both")
			leaf := &proof{rule: "vsai-both", cause: validation.ErrDuplicate}
			first := part(260, 64, append(body("neither"), part(258, 64, make([]byte, 4))...))
			parent := &proof{key: avp.Key{Code: 260}, flags: 64, offset: len(first), child: leaf, cause: leaf}
			field := &validation.FieldError{Path: []avp.Key{{Code: 260}}, Cause: parent}
			var e error = field
			switch kind {
			case "budget":
				budget = 31
			case "negative-budget":
				budget = -1
			case "field":
				l.MaxFieldBytes = 23
			case "count":
				l.MaxAVPs = 3
				l.MaxRepeated = 3
			case "depth":
				l.MaxDepth = 1
			case "frame":
				l.MaxFrameBytes = 40
				l.MaxFieldBytes = 40
			case "repeat":
				l.MaxRepeated = 1
			case "bare":
				e = validation.At(avp.Key{Code: 260}, validation.ErrDuplicate)
			case "path":
				field.Path = []avp.Key{{Code: 284}}
			case "offset":
				parent.offset = 0
			case "flags":
				parent.flags = 0
			case "rule":
				leaf.rule = "unknown"
			case "parent":
				parent.key.Code = 284
				field.Path[0].Code = 284
			case "sentinel":
				leaf.cause = validation.ErrField
			case "terminal-key":
				leaf.key.Code = 258
			case "terminal-child":
				leaf.child = validation.ErrDuplicate
			case "changed-condition":
				data = body("neither")
			case "duplicate":
				data = append(data, part(258, 64, make([]byte, 4))...)
			case "width":
				data = append(body("neither"), part(258, 64, []byte{0})...)
			case "mandatory":
				data = append(data, part(99999, 64, []byte{1})...)
			case "extras":
				l.MaxRepeated = 2
				for range 3 {
					data = append(data, part(99999, 0, nil)...)
				}
			case "truncated":
				data = append(data, 1)
			}
			src := append(first, part(parent.key.Code, 64, data)...)
			q, c, ok := groupavp.Quote(src, e, l, budget)
			if ok != (kind == "valid") || !ok && (q != nil || c != 0) {
				t.Fatal("unsafe proof", kind, c)
			}
			if ok {
				d := wire.NewDecoder(q, l)
				v, _, err := d.Next()
				want := append(part(259, 0, []byte{0, 0, 0, 9}), part(258, 64, []byte{1, 0, 0, 1})...)
				if err != nil || !bytes.Equal(v.Data, want) {
					t.Fatal("wrong repeated instance")
				}
			}
		})
	}
}

func FuzzGroupQuote(f *testing.F) {
	for i, kind := range []string{"both", "neither", "host", "state", "empty"} {
		f.Add(body(kind), uint8(i), uint16(64))
	}
	f.Fuzz(func(t *testing.T, src []byte, selector uint8, budget uint16) {
		if len(src) > 4096 {
			return
		}
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 8192
		l.MaxFieldBytes = 4096
		l.MaxAVPs = 64
		l.MaxRepeated = 16
		l.MaxDepth = 4
		kinds := []string{"both", "neither", "host", "state", "empty"}
		kind := kinds[int(selector)%len(kinds)]
		e := decode(src, kind, l)
		q, code, ok := groupavp.Quote(src, e, l, int(budget))
		if !ok {
			if q != nil || code != 0 {
				t.Fatal("partial result")
			}
			return
		}
		if len(q) > int(budget) || len(q) > 24 || code != 5005 && code != 5009 {
			t.Fatal("unbounded output")
		}
		if _, _, ok := groupavp.Quote(src, validation.ErrMissing, l, int(budget)); ok {
			t.Fatal("bare sentinel")
		}
		saved := append([]byte(nil), q...)
		clear(src)
		if !bytes.Equal(saved, q) {
			t.Fatal("borrowed")
		}
	})
}

func TestGroupProxyRecheck(t *testing.T) {
	l := wire.DefaultLimits()
	for _, kind := range []string{"host", "state"} {
		t.Run(kind, func(t *testing.T) {
			original := body(kind)
			cause := decode(original, kind, l)
			for name, src := range map[string][]byte{
				"present-empty-state": append(part(280, 64, []byte("p")), part(33, 64, nil)...),
				"invalid-host":        part(280, 64, []byte("bad_name")),
				"duplicate-state":     append(part(33, 64, nil), part(33, 64, nil)...),
				"unknown-mandatory":   append(append([]byte{}, original...), part(99999, 64, nil)...),
				"trailing-framing":    append(append([]byte{}, original...), 1),
			} {
				t.Run(name, func(t *testing.T) {
					if _, _, ok := groupavp.Quote(src, cause, l, 64); ok {
						t.Fatal("stale or unsafe evidence")
					}
				})
			}
		})
	}
}
