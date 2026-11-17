// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Typed-length evidence, exact quotes, ownership, limits and fuzz acceptance.
// SPDX-AI-Date: 2026-11-17

package excessavp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type lengthMetadata struct {
	*invalidMetadata
	kind string
}

func (m *lengthMetadata) DiameterLengthAVPV1() string { return m.kind }

func lengthBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	b := make([]byte, (8+len(data)+3)&^3)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	w.Bytes(avp.Key{Code: 450}, 0x40, data)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	for i := 8 + len(data); i < len(b); i++ {
		b[i] = 0xa5
	}
	return b
}

func TestLengthQuoteTypes(t *testing.T) {
	for _, kind := range []string{"u32", "i32", "time", "u64", "i64", "tbcd", "address4", "address6", "address-other", "address-zero", "unknown"} {
		for n := 0; n <= 20; n++ {
			t.Run(fmt.Sprintf("%s/%d", kind, n), func(t *testing.T) {
				data := make([]byte, n)
				typ := kind
				want := false
				switch kind {
				case "u32", "i32", "time":
					want = n != 4
				case "u64", "i64":
					want = n != 8
				case "tbcd":
					want = n == 0 || n > 8
				case "address4", "address6", "address-other", "address-zero":
					typ = "address"
					family := uint16(99)
					expected := 0
					if kind == "address4" {
						family = 1
						expected = 6
					}
					if kind == "address6" {
						family = 2
						expected = 18
					}
					if kind == "address-zero" {
						family = 0
					}
					if n >= 2 {
						binary.BigEndian.PutUint16(data, family)
					}
					want = n < 2 || expected != 0 && n != expected
				}
				leaf := lengthBytes(t, data)
				raw := append(lengthBytes(t, []byte{0, 0, 0, 1}), leaf...)
				m := &lengthMetadata{&invalidMetadata{key: avp.Key{Code: 450}, flags: 0x40, offset: 12, code: 5014, cause: avp.ErrValue}, typ}
				out, code, ok := QuoteInvalid(raw, validation.At(m.key, m), wire.DefaultLimits(), 1024)
				if ok != want {
					t.Fatalf("ok=%v want=%v", ok, want)
				}
				if !ok {
					if out != nil || code != 0 {
						t.Fatal("partial result")
					}
					return
				}
				if code != 5014 || !bytes.Equal(out, leaf) {
					t.Fatal("changed instance")
				}
				clear(raw)
				if !bytes.Equal(out, leaf) {
					t.Fatal("borrowed input")
				}
			})
		}
	}
}

func TestLengthQuoteBoundaries(t *testing.T) {
	for _, kind := range []string{"valid", "bare", "no-type", "wrong-cause", "wrong-type", "valid-length", "offset", "flags", "path", "truncated", "reply", "field", "count", "depth", "cyclic"} {
		t.Run(kind, func(t *testing.T) {
			raw := lengthBytes(t, []byte{1, 2, 3})
			l := wire.DefaultLimits()
			budget := 1024
			m := &lengthMetadata{&invalidMetadata{key: avp.Key{Code: 450}, flags: 0x40, code: 5014, cause: avp.ErrValue}, "u32"}
			e := validation.At(m.key, m).(*validation.FieldError)
			switch kind {
			case "bare":
				e.Cause = avp.ErrValue
			case "no-type":
				e.Cause = m.invalidMetadata
			case "wrong-cause":
				m.cause = wire.ErrLimit
			case "wrong-type":
				m.kind = "grouped"
			case "valid-length":
				raw = lengthBytes(t, []byte{1, 2, 3, 4})
			case "offset":
				m.offset = 4
			case "flags":
				m.flags = 0
			case "path":
				e.Path[0].Code++
			case "truncated":
				raw = raw[:len(raw)-1]
			case "reply":
				budget = len(raw) - 1
			case "field":
				l.MaxFieldBytes = 2
			case "count":
				l.MaxAVPs = 1
			case "depth":
				l.MaxDepth = 0
			case "cyclic":
				m.code = 0
				m.child = m
			}
			out, code, ok := QuoteInvalid(raw, e, l, budget)
			if ok != (kind == "valid") {
				t.Fatal(ok, code)
			}
			if !ok && (out != nil || code != 0) {
				t.Fatal("partial result")
			}
		})
	}
}

func FuzzLengthQuote(f *testing.F) {
	f.Add([]byte{0, 0, 1, 194, 0x40, 0, 0, 11, 1, 2, 3, 0xa5}, uint16(0), uint8(0), uint16(1024))
	f.Add([]byte{0, 0, 1, 194, 0x40, 0, 0, 8}, uint16(0), uint8(5), uint16(8))
	f.Add([]byte{1}, uint16(4), uint8(255), uint16(0))
	f.Fuzz(func(t *testing.T, raw []byte, offset uint16, kind uint8, budget uint16) {
		types := []string{"u32", "i32", "time", "u64", "i64", "tbcd", "address", "unknown"}
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 4096
		l.MaxFieldBytes = 2048
		l.MaxAVPs = 64
		l.MaxRepeated = 64
		m := &lengthMetadata{&invalidMetadata{key: avp.Key{Code: 450}, flags: 0x40, offset: int(offset), code: 5014, cause: avp.ErrValue}, types[int(kind)%len(types)]}
		out, code, ok := QuoteInvalid(raw, validation.At(m.key, m), l, int(budget))
		if !ok {
			return
		}
		if code != 5014 || len(out) > int(budget) || len(out) > l.MaxFieldBytes {
			t.Fatal("unbounded result")
		}
		d := wire.NewDecoder(raw, l)
		for {
			v, found, e := d.Next()
			if e != nil || !found {
				t.Fatal("invented instance", e)
			}
			if v.Offset != int(offset) {
				continue
			}
			size := (8 + len(v.Data) + 3) &^ 3
			if v.Key != m.key || v.Flags != m.flags || !bytes.Equal(out, raw[v.Offset:v.Offset+size]) {
				t.Fatal("changed bytes")
			}
			return
		}
	})
}
