// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Address-Type value acceptance.
// SPDX-AI-Date: 2026-11-19

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

type addressMetadata struct {
	name  string
	cause error
}

func (m addressMetadata) Error() string                      { return "address" }
func (m addressMetadata) Unwrap() error                      { return m.cause }
func (m addressMetadata) DiameterAddressTypeValueV1() string { return m.name }
func addressCause(group uint32) *validation.FieldError {
	p := []avp.Key{}
	for _, c := range []uint32{873, 877, group, 899} {
		p = append(p, avp.Key{VendorID: 10415, Code: c})
	}
	return &validation.FieldError{Path: p, Cause: addressMetadata{"combined-address-type-v1", errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func addressWord(v int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return interfacePart(899, b)
}
func addressFixture(group uint32, v int32, data []byte, reverse bool) []byte {
	a, b := addressWord(v), interfacePart(897, data)
	if reverse {
		a, b = b, a
	}
	return interfacePart(873, interfacePart(877, interfacePart(group, append(a, b...))))
}
func addressExpected(group uint32, v int32) []byte {
	return interfacePart(873, interfacePart(877, interfacePart(group, addressWord(v))))
}
func TestAddressValueConsumer(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, v := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for _, data := range [][]byte{nil, []byte("123"), bytes.Repeat([]byte{'A'}, 1024), {0}, {'A', 0, 'B'}} {
				for _, reverse := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/%d/%x/%t", group, v, data, reverse), func(t *testing.T) {
						raw := addressFixture(group, v, data, reverse)
						out, code, ok := QuoteInvalid(raw, addressCause(group), wire.DefaultLimits(), 52)
						want := v != 1 && bytes.IndexByte(data, 0) < 0
						if ok != want || ok && (code != 5004 || !bytes.Equal(out, addressExpected(group, v))) || !ok && (out != nil || code != 0) {
							t.Fatal(code, out)
						}
						clear(raw)
						if ok && !bytes.Equal(out, addressExpected(group, v)) {
							t.Fatal("borrowed")
						}
					})
				}
			}
		}
	}
}
func TestAddressValueControls(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "path", "vendor", "sentinel", "valid", "nul", "extra", "mandatory-extra", "trailing", "opaque-root", "opaque-sibling", "root-count"} {
			t.Run(fmt.Sprintf("%d/%s", group, fault), func(t *testing.T) {
				raw := addressFixture(group, -1, []byte("id"), false)
				var cause error = addressCause(group)
				l := wire.DefaultLimits()
				budget := 52
				switch fault {
				case "budget":
					budget = 51
				case "negative":
					budget = -1
				case "field":
					l.MaxFieldBytes = 51
				case "frame":
					l.MaxFrameBytes = len(raw) + 19
				case "count":
					l.MaxAVPs = 4
				case "depth":
					l.MaxDepth = 3
				case "repeat":
					l.MaxRepeated = 0
				case "bare":
					cause = validation.At(avp.Key{VendorID: 10415, Code: 899}, validation.ErrField)
				case "name":
					cause.(*validation.FieldError).Cause = addressMetadata{"unknown", errors.Join(validation.ErrUnsupported, validation.ErrField)}
				case "path":
					cause.(*validation.FieldError).Path[2].Code = 999
				case "vendor":
					cause.(*validation.FieldError).Path[2].VendorID = 0
				case "sentinel":
					cause.(*validation.FieldError).Cause = addressMetadata{"combined-address-type-v1", validation.ErrUnsupported}
				case "valid":
					raw = addressFixture(group, 1, nil, false)
				case "nul":
					raw = addressFixture(group, -1, []byte{0}, false)
				case "extra", "mandatory-extra":
					extra := unitPart(9999, nil)
					if fault == "extra" {
						extra[4] = 0
					}
					raw = interfaceEdit(raw, []uint32{873, 877, group}, func(b []byte) []byte { return interfacePart(group, append(b[12:], extra...)) })
				case "trailing":
					raw = append(raw, 1)
				case "opaque-root":
					raw = append(raw, unitPart(9999, []byte{1})...)
				case "opaque-sibling":
					other := uint32(886)
					if group == 886 {
						other = 1201
					}
					raw = interfaceEdit(raw, []uint32{873, 877}, func(b []byte) []byte { return interfacePart(877, append(b[12:], interfacePart(other, []byte{1})...)) })
				case "root-count":
					raw = append(raw, unitPart(9999, nil)...)
					l.MaxAVPs = 5
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "exact" || fault == "opaque-root" || fault == "opaque-sibling"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, addressExpected(group, -1))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, p := range [][]uint32{{873}, {873, 877}, {873, 877, group}, {873, 877, group, 899}, {873, 877, group, 897}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%d/%v/%s", group, p, fault), func(t *testing.T) {
					raw := interfaceEdit(addressFixture(group, -1, []byte("id"), false), p, func(b []byte) []byte {
						switch fault {
						case "missing":
							return nil
						case "duplicate":
							return append(b, b...)
						case "flags":
							b[4] |= 32
						case "vendor":
							binary.BigEndian.PutUint32(b[8:12], 10416)
						case "trailing":
							if len(p) < 4 {
								return interfacePart(p[len(p)-1], append(b[12:], 1))
							}
							return append(b, 1)
						}
						return b
					})
					out, code, ok := QuoteInvalid(raw, addressCause(group), wire.DefaultLimits(), 52)
					if ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
		for _, width := range []int{0, 1, 3, 5, 8} {
			raw := interfaceEdit(addressFixture(group, -1, nil, false), []uint32{873, 877, group, 899}, func([]byte) []byte { return interfacePart(899, make([]byte, width)) })
			if _, _, ok := QuoteInvalid(raw, addressCause(group), wire.DefaultLimits(), 52); ok {
				t.Fatal("width", width)
			}
		}
	}
}
func TestAddressValueExactLimits(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		l := wire.DefaultLimits()
		l.MaxAVPs = 5
		l.MaxFieldBytes = 52
		l.MaxFrameBytes = 84
		l.MaxDepth = 4
		l.MaxRepeated = 1
		raw := addressFixture(group, -1, nil, false)
		out, code, ok := QuoteInvalid(raw, addressCause(group), l, 52)
		if len(raw)+20 != 84 || !ok || code != 5004 || !bytes.Equal(out, addressExpected(group, -1)) {
			t.Fatal(code, out)
		}
	}
}
func fuzzAddressValue(t *testing.T, raw []byte, budget uint16, group uint32) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, addressCause(group), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) != 52 || len(out) > int(budget) {
		t.Fatal("bound")
	}
	source, quote := raw, out
	d := wire.NewDecoder(source, l)
	for _, c := range []uint32{873, 877, group} {
		var found wire.View
		n := 0
		for {
			v, yes, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !yes {
				break
			}
			if v.Key == (avp.Key{VendorID: 10415, Code: c}) {
				found = v
				n++
			}
		}
		if n != 1 || found.Flags != 192 {
			t.Fatal("ancestor")
		}
		q := wire.NewDecoder(quote, l)
		v, yes, e := q.Next()
		if e != nil || !yes || q.Remaining() != 0 || v.Key != found.Key || v.Flags != 192 {
			t.Fatal("quote ancestor")
		}
		source, quote = found.Data, v.Data
		d, e = d.Child(source)
		if e != nil {
			t.Fatal(e)
		}
	}
	counts := [2]int{}
	var leaf wire.View
	for {
		v, yes, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !yes {
			break
		}
		if v.Flags != 192 {
			t.Fatal("flags")
		}
		switch v.Key {
		case avp.Key{VendorID: 10415, Code: 899}:
			counts[0]++
			leaf = v
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) == 1 {
				t.Fatal("type")
			}
		case avp.Key{VendorID: 10415, Code: 897}:
			counts[1]++
			if bytes.IndexByte(v.Data, 0) >= 0 {
				t.Fatal("data")
			}
		default:
			t.Fatal("extra")
		}
	}
	if counts != ([2]int{1, 1}) || !bytes.Equal(quote, source[leaf.Offset:leaf.Offset+16]) {
		t.Fatal("exact leaf")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, addressCause(group), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
