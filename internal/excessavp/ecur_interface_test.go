// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: INITIAL/TERMINATION interface value acceptance.
// SPDX-AI-Date: 2026-11-18

package excessavp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func ecurCause(origin bool) *validation.FieldError {
	c := interfaceCause()
	name := "combined-ecur-destination-interface-v1"
	code := uint32(2002)
	if origin {
		name = "combined-ecur-originator-interface-v1"
		code = 2009
	}
	c.Cause.(*interfaceMetadata).rule = name
	c.Path[2].Code = code
	return c
}
func ecurFixture(kind uint32, origin bool, value int32) []byte {
	a, b := int32(3), int32(2)
	if origin {
		a = value
	} else {
		b = value
	}
	sms := append(interfacePart(2009, interfaceLeaf(a)), interfacePart(2002, interfaceLeaf(b))...)
	raw := append(actionWord(416, kind), actionWord(455, 1)...)
	raw = append(raw, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
	return append(raw, interfacePart(873, interfacePart(2000, sms))...)
}
func ecurExpected(origin bool, value int32) []byte {
	code := uint32(2002)
	if origin {
		code = 2009
	}
	return interfacePart(873, interfacePart(2000, interfacePart(code, interfaceLeaf(value))))
}
func TestECURInterfaceValueConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3} {
		for _, origin := range []bool{false, true} {
			for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
				t.Run(fmt.Sprintf("%d/%t/%d", kind, origin, value), func(t *testing.T) {
					raw := ecurFixture(kind, origin, value)
					expected := int32(2)
					if origin {
						expected = 3
					}
					out, code, ok := QuoteInvalid(raw, ecurCause(origin), wire.DefaultLimits(), 52)
					if ok != (value != expected) || ok && (code != 5004 || !bytes.Equal(out, ecurExpected(origin, value))) {
						t.Fatal(code, out)
					}
					if !ok && (out != nil || code != 0) {
						t.Fatal("partial quote")
					}
					clear(raw)
					if ok && !bytes.Equal(out, ecurExpected(origin, value)) {
						t.Fatal("borrowed")
					}
				})
			}
		}
	}
}
func TestECURInterfaceValueControls(t *testing.T) {
	for _, origin := range []bool{false, true} {
		selected, other := uint32(2002), uint32(2009)
		if origin {
			selected, other = other, selected
		}
		for _, bad := range []string{"exact", "bytes", "negative", "field", "frame", "source-count", "output-count", "depth", "repeat", "bare", "cause", "name", "path", "kind", "indicator", "equipment", "action", "address", "both", "valid", "counterpart-length", "counterpart-flags", "counterpart-duplicate", "counterpart-trailing", "trailing", "selector-duplicate", "selector-flags", "selector-length", "sibling-order"} {
			t.Run(fmt.Sprintf("%t/%s", origin, bad), func(t *testing.T) {
				raw := ecurFixture(1, origin, -1)
				cause := ecurCause(origin)
				l := wire.DefaultLimits()
				budget := 52
				edit := func(path []uint32, f func([]byte) []byte) { raw = interfaceEdit(raw, path, f) }
				switch bad {
				case "bytes":
					budget = 51
				case "negative":
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
				case "bare":
					cause.Cause = validation.ErrField
				case "cause":
					cause.Cause.(*interfaceMetadata).cause = validation.ErrUnsupported
				case "name":
					cause.Cause.(*interfaceMetadata).rule = "unreviewed"
				case "path":
					cause.Path[2].Code = other
				case "kind":
					edit([]uint32{416}, func([]byte) []byte { return actionWord(416, 4) })
				case "indicator":
					edit([]uint32{455}, func([]byte) []byte { return actionWord(455, 0) })
				case "equipment":
					edit([]uint32{458}, func([]byte) []byte { return nil })
				case "action":
					raw = append(raw, actionWord(436, 0)...)
				case "address":
					edit([]uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], interfacePart(2008, nil)...)) })
				case "both":
					edit([]uint32{873, 2000, other, 2006}, func([]byte) []byte { return interfaceLeaf(-1) })
				case "valid":
					value := int32(2)
					if origin {
						value = 3
					}
					edit([]uint32{873, 2000, selected, 2006}, func([]byte) []byte { return interfaceLeaf(value) })
				case "counterpart-length":
					edit([]uint32{873, 2000, other, 2006}, func([]byte) []byte { return interfacePart(2006, []byte{0}) })
				case "counterpart-flags":
					edit([]uint32{873, 2000, other, 2006}, func(b []byte) []byte { b[4] |= 32; return b })
				case "counterpart-duplicate":
					edit([]uint32{873, 2000, other, 2006}, func(b []byte) []byte { return append(b, b...) })
				case "counterpart-trailing":
					edit([]uint32{873, 2000, other}, func(b []byte) []byte { return interfacePart(other, append(b[12:], 1)) })
				case "trailing":
					raw = append(raw, 1)
				case "selector-duplicate":
					edit([]uint32{416}, func(b []byte) []byte { return append(b, b...) })
				case "selector-flags":
					edit([]uint32{416}, func(b []byte) []byte { b[4] |= 32; return b })
				case "selector-length":
					edit([]uint32{416}, func(b []byte) []byte { wire.Put24(b[5:8], 9); return b })
				case "sibling-order":
					edit([]uint32{873, 2000}, func(b []byte) []byte {
						data := append(append([]byte{}, b[40:68]...), b[12:40]...)
						data = append(interfacePart(2016, interfaceLeaf(99)), data...)
						return interfacePart(2000, data)
					})
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := bad == "exact" || bad == "sibling-order"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, ecurExpected(origin, -1))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, parent := range []uint32{selected, other} {
			for _, depth := range []int{1, 2, 3, 4} {
				for _, bad := range []string{"missing", "duplicate", "flags", "vendor"} {
					t.Run(fmt.Sprintf("%t/%d/%d/%s", origin, parent, depth, bad), func(t *testing.T) {
						path := []uint32{873, 2000, parent, 2006}
						raw := interfaceEdit(ecurFixture(3, origin, -1), path[:depth], func(b []byte) []byte {
							switch bad {
							case "missing":
								return nil
							case "duplicate":
								return append(b, b...)
							case "flags":
								b[4] |= 32
							case "vendor":
								binary.BigEndian.PutUint32(b[8:], 10416)
							}
							return b
						})
						if out, code, ok := QuoteInvalid(raw, ecurCause(origin), wire.DefaultLimits(), 52); ok || out != nil || code != 0 {
							t.Fatal(code, out)
						}
					})
				}
			}
		}
	}
}
func TestECURInterfaceValueExactLimits(t *testing.T) {
	for _, origin := range []bool{false, true} {
		raw := ecurFixture(3, origin, -1)
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 132
		l.MaxFieldBytes = 68
		l.MaxAVPs = 9
		l.MaxDepth = 4
		l.MaxRepeated = 1
		if len(raw) != 112 {
			t.Fatal(len(raw))
		}
		q, code, ok := QuoteInvalid(raw, ecurCause(origin), l, 52)
		if !ok || code != 5004 || !bytes.Equal(q, ecurExpected(origin, -1)) {
			t.Fatal(code, q)
		}
	}
}

func fuzzECURInterfaceValue(t *testing.T, raw []byte, budget uint16, origin bool) {
	group := uint32(2002)
	expected := int32(2)
	if origin {
		group = 2009
		expected = 3
	}
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, ecurCause(origin), l, int(budget))
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
	for depth, want := range []uint32{873, 2000, group, 2006} {
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
	if len(leaf) != 4 || int32(binary.BigEndian.Uint32(leaf)) == expected {
		t.Fatal("valid or malformed value")
	}
	// Locate exactly one matching original node at every selected level.
	source := raw
	keyPath := []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: group}, {VendorID: 10415, Code: 2006}}
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
