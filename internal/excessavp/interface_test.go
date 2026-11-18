// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated IEC Originator-Interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type interfaceMetadata struct {
	rule  string
	cause error
}

func (m *interfaceMetadata) Error() string                    { return "interface value" }
func (m *interfaceMetadata) Unwrap() error                    { return m.cause }
func (m *interfaceMetadata) DiameterInterfaceValueV1() string { return m.rule }
func interfaceCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2009}, {VendorID: 10415, Code: 2006}},
		Cause: &interfaceMetadata{"combined-iec-originator-interface-v1", errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func interfacePart(code uint32, data []byte) []byte {
	b := make([]byte, (12+len(data)+3)&^3)
	binary.BigEndian.PutUint32(b, code)
	b[4] = 192
	wire.Put24(b[5:8], uint32(12+len(data)))
	binary.BigEndian.PutUint32(b[8:12], 10415)
	copy(b[12:], data)
	return b
}
func interfaceLeaf(value int32) []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(value))
	return interfacePart(2006, b[:])
}
func interfaceExpected(value int32) []byte {
	return interfacePart(873, interfacePart(2000, interfacePart(2009, interfaceLeaf(value))))
}
func interfaceFixture(value int32) []byte {
	raw := append(actionWord(416, 4), actionWord(436, 0)...)
	raw = append(raw, actionWord(455, 1)...)
	raw = append(raw, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
	sms := append(interfacePart(2008, nil), interfacePart(2009, interfaceLeaf(value))...)
	return append(raw, interfacePart(873, interfacePart(2000, sms))...)
}

// Mutate a known fixture using raw lengths, independent of the production
// unique-key consumer. Rebuild only containing headers/padding.
func interfaceEdit(src []byte, path []uint32, change func([]byte) []byte) []byte {
	var out []byte
	for pos := 0; pos < len(src); {
		n := int(wire.Uint24(src[pos+5 : pos+8]))
		size := (n + 3) &^ 3
		head := 8
		if src[pos+4]&128 != 0 {
			head = 12
		}
		part := append([]byte(nil), src[pos:pos+size]...)
		if binary.BigEndian.Uint32(part) == path[0] {
			if len(path) == 1 {
				part = change(part)
			} else {
				data := interfaceEdit(part[head:n], path[1:], change)
				part = append(part[:head:head], data...)
				wire.Put24(part[5:8], uint32(len(part)))
			}
		}
		out = append(out, part...)
		pos += size
	}
	return out
}
func TestInterfaceValueQuoteControls(t *testing.T) {
	cases := []string{"exact", "bytes", "negative-budget", "field", "frame", "source-count", "output-count", "depth", "repeat", "path0", "path1", "path2", "path3", "short-path", "long-path", "bare", "cause", "rule", "kind", "action", "indicator", "equipment", "originator", "originator-duplicate", "originator-nul", "valid", "length", "leaf-flags", "leaf-vendor", "trailing0", "trailing1", "trailing2", "trailing3"}
	for _, bad := range cases {
		t.Run(bad, func(t *testing.T) {
			raw := interfaceFixture(-1)
			cause := interfaceCause()
			l := wire.DefaultLimits()
			budget := 52
			path := []uint32{873, 2000, 2009, 2006}
			changeWord := func(code uint32, value uint32) {
				raw = interfaceEdit(raw, []uint32{code}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[8:], value); return b })
			}
			switch bad {
			case "bytes":
				budget--
			case "negative-budget":
				budget = -1
			case "field":
				l.MaxFieldBytes = 51
			case "frame":
				l.MaxFrameBytes = len(raw) + 19
			case "source-count":
				l.MaxAVPs = 8
			case "output-count":
				l.MaxAVPs = 4
			case "depth":
				l.MaxDepth = 3
			case "repeat":
				l.MaxRepeated = 0
			case "path0", "path1", "path2", "path3":
				cause.Path[int(bad[4]-'0')].Code++
			case "short-path":
				cause.Path = cause.Path[:3]
			case "long-path":
				cause.Path = append(cause.Path, avp.Key{Code: 2006})
			case "bare":
				cause.Cause = errors.Join(validation.ErrField, validation.ErrUnsupported)
			case "cause":
				cause.Cause.(*interfaceMetadata).cause = validation.ErrUnsupported
			case "rule":
				cause.Cause.(*interfaceMetadata).rule = "unreviewed"
			case "kind":
				changeWord(416, 1)
			case "action":
				changeWord(436, 2)
			case "indicator":
				changeWord(455, 2)
			case "equipment":
				raw = interfaceEdit(raw, []uint32{458}, func([]byte) []byte { return nil })
			case "originator":
				raw = interfaceEdit(raw, []uint32{873, 2000, 2008}, func([]byte) []byte { return nil })
			case "originator-duplicate":
				raw = interfaceEdit(raw, []uint32{873, 2000, 2008}, func(b []byte) []byte { return append(b, b...) })
			case "originator-nul":
				raw = interfaceEdit(raw, []uint32{873, 2000, 2008}, func([]byte) []byte { return interfacePart(2008, []byte{0}) })
			case "valid":
				raw = interfaceEdit(raw, path, func([]byte) []byte { return interfaceLeaf(1) })
			case "length":
				raw = interfaceEdit(raw, path, func([]byte) []byte { return interfacePart(2006, []byte{1}) })
			case "leaf-flags":
				raw = interfaceEdit(raw, path, func(b []byte) []byte { b[4] |= 32; return b })
			case "leaf-vendor":
				raw = interfaceEdit(raw, path, func(b []byte) []byte { binary.BigEndian.PutUint32(b[8:], 10416); return b })
			case "trailing0":
				raw = append(raw, 1)
			case "trailing1", "trailing2", "trailing3":
				depth := int(bad[8] - '0')
				raw = interfaceEdit(raw, path[:depth], func(b []byte) []byte {
					b = append(b, 1)
					wire.Put24(b[5:8], uint32(len(b)))
					for len(b)%4 != 0 {
						b = append(b, 0)
					}
					return b
				})
			}
			out, code, ok := QuoteInvalid(raw, cause, l, budget)
			if ok != (bad == "exact") || !ok && (out != nil || code != 0) {
				t.Fatal(code, ok, out)
			}
			if ok {
				if code != 5004 || !bytes.Equal(out, interfaceExpected(-1)) {
					t.Fatal("wrong instance")
				}
				clear(raw)
				if !bytes.Equal(out, interfaceExpected(-1)) {
					t.Fatal("borrowed")
				}
			}
		})
	}
	for _, level := range []int{0, 1, 2, 3} {
		for _, bad := range []string{"missing", "duplicate", "flags"} {
			t.Run(fmt.Sprintf("%s/%d", bad, level), func(t *testing.T) {
				path := []uint32{873, 2000, 2009, 2006}
				raw := interfaceEdit(interfaceFixture(-1), path[:level+1], func(b []byte) []byte {
					switch bad {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					default:
						b[4] |= 32
						return b
					}
				})
				if out, code, ok := QuoteInvalid(raw, interfaceCause(), wire.DefaultLimits(), 52); ok || code != 0 || out != nil {
					t.Fatal("ambiguous instance", out, code)
				}
			})
		}
	}
}
func TestInterfaceValueExactSourceAndQuoteLimits(t *testing.T) {
	raw := interfaceFixture(-1)
	l := wire.DefaultLimits()
	l.MaxFrameBytes = len(raw) + 20
	l.MaxFieldBytes = 52
	l.MaxAVPs = 9
	l.MaxDepth = 4
	l.MaxRepeated = 1
	out, code, ok := QuoteInvalid(raw, interfaceCause(), l, 52)
	if !ok || code != 5004 || !bytes.Equal(out, interfaceExpected(-1)) {
		t.Fatal(code, ok, out)
	}
}
func TestInterfaceValueSiblingsAndOrdering(t *testing.T) {
	raw := interfaceFixture(-1)
	// Put another instance of the same leaf key in an unrelated sibling; only
	// Originator-Interface's selected child may be quoted.
	raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte {
		data := append(interfacePart(2016, interfaceLeaf(99)), b[12:]...)
		return interfacePart(2000, data)
	})
	raw = interfaceEdit(raw, []uint32{873, 2000, 2009}, func(b []byte) []byte { return interfacePart(2009, append(actionWord(9000, 42), b[12:]...)) })
	out, code, ok := QuoteInvalid(raw, interfaceCause(), wire.DefaultLimits(), 52)
	if !ok || code != 5004 || !bytes.Equal(out, interfaceExpected(-1)) {
		t.Fatal("sibling substituted", out, code)
	}
}

// Add the selected consumer to the existing invalid-value fuzzer. Successful
// outputs must contain only the exact original selected path and signed value.
func fuzzInterfaceValue(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, interfaceCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial interface quote")
		}
		return
	}
	if code != 5004 || len(out) != 52 || len(out) > int(budget) {
		t.Fatal("interface quote bounds")
	}
	d := wire.NewDecoder(out, l)
	var leaf []byte
	for depth, want := range []uint32{873, 2000, 2009, 2006} {
		v, found, e := d.Next()
		if e != nil || !found || v.Key != (avp.Key{VendorID: 10415, Code: want}) || v.Flags != 192 || d.Remaining() != 0 {
			t.Fatal("quote grammar")
		}
		if depth == 3 {
			leaf = v.Data
		} else {
			d, e = d.Child(v.Data)
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(leaf) != 4 || int32(binary.BigEndian.Uint32(leaf)) == 1 {
		t.Fatal("valid or malformed value")
	}
	// Locate exactly one matching original node at every selected level.
	source := raw
	keyPath := []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2009}, {VendorID: 10415, Code: 2006}}
	for depth, key := range keyPath {
		scan := wire.NewDecoder(source, l)
		var selected wire.View
		count := 0
		for {
			v, found, e := scan.Next()
			if e != nil {
				t.Fatal("malformed source")
			}
			if !found {
				break
			}
			if v.Key == key {
				selected = v
				count++
			}
		}
		if count != 1 {
			t.Fatal("ambiguous source")
		}
		if depth == 3 && !bytes.Equal(out[36:], source[selected.Offset:selected.Offset+16]) {
			t.Fatal("changed original leaf")
		}
		source = selected.Data
	}
}
