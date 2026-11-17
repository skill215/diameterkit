// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: First-excess diagnostic provenance, bounds and fuzz tests.
// SPDX-AI-Date: 2026-11-17

package excessavp

import (
	"bytes"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type metadata struct {
	key             avp.Key
	flags           byte
	offset, maximum int
	child           error
}

func (*metadata) Error() string { return validation.ErrDuplicate.Error() }
func (*metadata) Unwrap() error { return validation.ErrDuplicate }
func (e *metadata) DiameterExcessAVPV1() (avp.Key, byte, int, int, error) {
	return e.key, e.flags, e.offset, e.maximum, e.child
}

func excessFixture() ([]byte, *validation.FieldError) {
	b := make([]byte, 36)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	k := avp.Key{Code: 450}
	for _, n := range []uint32{1, 2, 3} {
		w.Uint32(k, 0x40, n)
	}
	return b, validation.At(k, &metadata{key: k, flags: 0x40, offset: 12, maximum: 1}).(*validation.FieldError)
}

func TestExcessQuoteFirstInstanceAndLimits(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*wire.Limits, *int, *[]byte, *validation.FieldError)
		ok     bool
		offset int
	}{
		{"exact", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { *n = 12 }, true, 12},
		{"max-two", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) {
			m := e.Cause.(*metadata)
			m.maximum = 2
			m.offset = 24
		}, true, 24},
		{"third-is-not-first-excess", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).offset = 24 }, false, 0},
		{"first-is-not-excess", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).offset = 0 }, false, 0},
		{"negative-offset", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).offset = -4 }, false, 0},
		{"unaligned-offset", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).offset = 13 }, false, 0},
		{"past-end", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) {
			e.Cause.(*metadata).offset = 1000000
		}, false, 0},
		{"path-changed", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Path[0].Code++ }, false, 0},
		{"flags-changed", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).flags = 0 }, false, 0},
		{"unproven", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause = validation.ErrDuplicate }, false, 0},
		{"invalid-maximum", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { e.Cause.(*metadata).maximum = 257 }, false, 0},
		{"reply-bytes", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { *n = 11 }, false, 0},
		{"field-bytes", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { l.MaxFieldBytes = 11 }, false, 0},
		{"local-repeats", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { l.MaxRepeated = 1 }, false, 0},
		{"count", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { l.MaxAVPs = 1; l.MaxRepeated = 1 }, false, 0},
		{"truncated", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { *b = (*b)[:23] }, false, 0},
		{"bad-length", func(l *wire.Limits, n *int, b *[]byte, e *validation.FieldError) { wire.Put24((*b)[17:20], 7) }, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, e := excessFixture()
			l := wire.DefaultLimits()
			n := 1024
			tc.change(&l, &n, &b, e)
			out, ok := Quote(b, e, l, n)
			if ok != tc.ok {
				t.Fatalf("ok=%v bytes=%x", ok, out)
			}
			if !ok {
				return
			}
			if !bytes.Equal(out, b[tc.offset:tc.offset+12]) || !errors.Is(e, validation.ErrDuplicate) {
				t.Fatal("wrong first excess")
			}
			frozen := append([]byte(nil), out...)
			clear(b)
			if !bytes.Equal(out, frozen) {
				t.Fatal("borrowed payload")
			}
		})
	}
}

func TestExcessQuoteGroupedPath(t *testing.T) {
	k := avp.Key{Code: 450}
	group := avp.Key{VendorID: 10415, Code: 628}
	b, e := excessFixture()
	raw := make([]byte, 60)
	w := wire.NewEncoder(raw, wire.DefaultLimits())
	w.Bytes(group, 0x80, b[:12])
	w.Bytes(group, 0xc0, b[:24])
	inner := e.Cause.(*metadata)
	outer := &metadata{key: group, flags: 0xc0, offset: 24, child: inner}
	err := validation.At(group, e).(*validation.FieldError)
	err.Cause = outer
	l := wire.DefaultLimits()
	q, ok := Quote(raw, err, l, 1024)
	if !ok {
		t.Fatal("missing quote")
	}
	d := wire.NewDecoder(q, l)
	v, found, x := d.Next()
	if x != nil || !found || v.Key != group || v.Flags != 0xc0 || !bytes.Equal(v.Data, b[12:24]) || d.Remaining() != 0 {
		t.Fatal("wrong repeated parent", v, x)
	}
	l.MaxDepth = 1
	if _, ok = Quote(raw, err, l, 1024); ok {
		t.Fatal("wrapper depth omitted")
	}
	outer.flags |= 1
	raw[28] |= 1
	if _, ok = Quote(raw, err, wire.DefaultLimits(), 1024); ok {
		t.Fatal("repaired parent flags")
	}
	if k != inner.key {
		t.Fatal("fixture changed")
	}
}

func FuzzExcessQuote(f *testing.F) {
	b, _ := excessFixture()
	f.Add(b, uint16(12), uint16(1), uint16(1024))
	f.Add(b, uint16(24), uint16(2), uint16(12))
	f.Add([]byte{1}, uint16(0), uint16(1), uint16(20))
	f.Fuzz(func(t *testing.T, raw []byte, offset, maximum, budget uint16) {
		k := avp.Key{Code: 450}
		err := validation.At(k, &metadata{key: k, flags: 0x40, offset: int(offset), maximum: int(maximum)})
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 4096
		l.MaxFieldBytes = 2048
		l.MaxAVPs = 64
		l.MaxRepeated = 64
		out, ok := Quote(raw, err, l, int(budget))
		if !ok {
			return
		}
		if len(out) > int(budget) || len(out) > l.MaxFieldBytes {
			t.Fatal("unbounded quote")
		}
		d := wire.NewDecoder(raw, l)
		seen := 0
		for {
			v, found, e := d.Next()
			if e != nil || !found {
				t.Fatal("invented quote", e)
			}
			if v.Key == k {
				seen++
			}
			if v.Offset == int(offset) {
				if seen != int(maximum)+1 || !bytes.Equal(out, raw[v.Offset:v.Offset+(8+len(v.Data)+3)&^3]) {
					t.Fatal("wrong instance")
				}
				break
			}
		}
	})
}
