// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Address-Data NUL value acceptance.
// SPDX-AI-Date: 2026-11-23

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

type addressDataMetadata struct {
	name  string
	cause error
}

func (m addressDataMetadata) Error() string                      { return "address" }
func (m addressDataMetadata) Unwrap() error                      { return m.cause }
func (m addressDataMetadata) DiameterAddressDataValueV1() string { return m.name }
func addressDataCause(group uint32) *validation.FieldError {
	p := []avp.Key{}
	for _, c := range []uint32{873, 877, group, 897} {
		p = append(p, avp.Key{VendorID: 10415, Code: c})
	}
	return &validation.FieldError{Path: p, Cause: addressDataMetadata{"combined-address-data-nul-v1", validation.ErrField}}
}
func TestAddressDataConsumer(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, v := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for _, data := range [][]byte{nil, []byte("123"), bytes.Repeat([]byte{'A'}, 1024), {0}, {'A', 0, 'B'}} {
				for _, reverse := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/%d/%x/%t", group, v, data, reverse), func(t *testing.T) {
						raw := addressFixture(group, v, data, reverse)
						out, code, ok := QuoteInvalid(raw, addressDataCause(group), wire.DefaultLimits(), 4096)
						want := v == 1 && bytes.IndexByte(data, 0) >= 0
						if ok != want || ok && (code != 5004 || !bytes.Equal(out, addressDataExpected(group, data))) || !ok && (out != nil || code != 0) {
							t.Fatal(code, out)
						}
						clear(raw)
						if ok && !bytes.Equal(out, addressDataExpected(group, data)) {
							t.Fatal("borrowed")
						}
					})
				}
			}
		}
	}
}
func TestAddressDataControls(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "path", "vendor", "sentinel", "valid", "nul", "extra", "mandatory-extra", "trailing", "opaque-root", "opaque-sibling", "root-count"} {
			t.Run(fmt.Sprintf("%d/%s", group, fault), func(t *testing.T) {
				raw := addressFixture(group, 1, []byte{0}, false)
				var cause error = addressDataCause(group)
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
					cause.(*validation.FieldError).Cause = addressDataMetadata{"unknown", validation.ErrField}
				case "path":
					cause.(*validation.FieldError).Path[2].Code = 999
				case "vendor":
					cause.(*validation.FieldError).Path[2].VendorID = 0
				case "sentinel":
					cause.(*validation.FieldError).Cause = addressDataMetadata{"combined-address-data-nul-v1", validation.ErrUnsupported}
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
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, addressDataExpected(group, []byte{0}))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, p := range [][]uint32{{873}, {873, 877}, {873, 877, group}, {873, 877, group, 899}, {873, 877, group, 897}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%d/%v/%s", group, p, fault), func(t *testing.T) {
					raw := interfaceEdit(addressFixture(group, 1, []byte{0}, false), p, func(b []byte) []byte {
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
					out, code, ok := QuoteInvalid(raw, addressDataCause(group), wire.DefaultLimits(), 52)
					if ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
		for _, width := range []int{0, 1, 3, 5, 8} {
			raw := interfaceEdit(addressFixture(group, -1, nil, false), []uint32{873, 877, group, 899}, func([]byte) []byte { return interfacePart(899, make([]byte, width)) })
			if _, _, ok := QuoteInvalid(raw, addressDataCause(group), wire.DefaultLimits(), 52); ok {
				t.Fatal("width", width)
			}
		}
	}
}
func fuzzAddressData(t *testing.T, raw []byte, budget uint16, group uint32) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, addressDataCause(group), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) < 52 || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes || len(out) > int(budget) {
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
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 1 {
				t.Fatal("type")
			}
		case avp.Key{VendorID: 10415, Code: 897}:
			counts[1]++
			leaf = v
			if bytes.IndexByte(v.Data, 0) < 0 {
				t.Fatal("data")
			}
		default:
			t.Fatal("extra")
		}
	}
	if counts != ([2]int{1, 1}) || !bytes.Equal(quote, source[leaf.Offset:leaf.Offset+((12+len(leaf.Data)+3)&^3)]) {
		t.Fatal("exact leaf")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, addressDataCause(group), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}

func addressDataExpected(group uint32, data []byte) []byte {
	return interfacePart(873, interfacePart(877, interfacePart(group, interfacePart(897, data))))
}
func TestAddressDataExactLimits(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
			for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth"} {
				t.Run(fmt.Sprintf("%d/%s", n, fault), func(t *testing.T) {
					data := make([]byte, n)
					raw := addressFixture(group, 1, data, false)
					aligned := (n + 3) &^ 3
					l := wire.DefaultLimits()
					l.MaxAVPs = 5
					l.MaxFrameBytes = 84 + aligned
					l.MaxFieldBytes = 52 + aligned
					l.MaxDepth = 4
					l.MaxRepeated = 1
					budget := 48 + aligned
					switch fault {
					case "budget":
						budget--
					case "frame":
						l.MaxFrameBytes--
					case "field":
						l.MaxFieldBytes--
					case "count":
						l.MaxAVPs--
					case "depth":
						l.MaxDepth--
					}
					out, code, ok := QuoteInvalid(raw, addressDataCause(group), l, budget)
					if len(raw)+20 != 84+aligned || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, addressDataExpected(group, data))) || !ok && (out != nil || code != 0) {
						t.Fatal(code, out)
					}
				})
			}
		}
	}
}

func TestAddressDataPadding(t *testing.T) {
	for _, group := range []uint32{886, 1201} {
		for _, data := range [][]byte{{0}, {1, 0}, {1, 0, 2}, []byte("id")} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%x/%t", data, reverse), func(t *testing.T) {
					leaf := interfacePart(897, data)
					for i := 12 + len(data); i < len(leaf); i++ {
						leaf[i] = 0xee
					}
					typ := addressWord(1)
					b := append(append([]byte{}, typ...), leaf...)
					if reverse {
						b = append(append([]byte{}, leaf...), typ...)
					}
					raw := interfacePart(873, interfacePart(877, interfacePart(group, b)))
					out, code, ok := QuoteInvalid(raw, addressDataCause(group), wire.DefaultLimits(), 4096)
					want := bytes.IndexByte(data, 0) >= 0
					if ok != want || ok && (code != 5004 || !bytes.Equal(out, interfacePart(873, interfacePart(877, interfacePart(group, leaf))))) {
						t.Fatal(code, out)
					}
				})
			}
		}
	}
}
