// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Exact equipment data instances, padding, counterpart controls and fuzz oracle.
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

type equipmentDataMetadata struct {
	name  string
	cause error
}

func (m equipmentDataMetadata) Error() string                        { return "data" }
func (m equipmentDataMetadata) Unwrap() error                        { return m.cause }
func (m equipmentDataMetadata) DiameterEquipmentDataValueV1() string { return m.name }
func equipmentDataCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{Code: 458}, {Code: 460}}, Cause: equipmentDataMetadata{"combined-equipment-data-nul-v1", validation.ErrField}}
}
func equipmentDataExpected(data []byte) []byte { return unitPart(458, unitPart(460, data)) }
func TestEquipmentDataConsumer(t *testing.T) {
	for _, typ := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
		for i, data := range [][]byte{nil, {}, []byte("id"), {0}, {0, 1}, {1, 0}, {1, 0, 2}, {1, 2, 3, 0}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%d/%t", typ, i, reverse), func(t *testing.T) {
					raw := equipmentFixture(typ, data, reverse)
					expected := equipmentDataExpected(data)
					out, code, ok := QuoteInvalid(raw, equipmentDataCause(), wire.DefaultLimits(), len(expected))
					want := typ == 0 && bytes.IndexByte(data, 0) >= 0
					if ok != want || ok && (code != 5004 || !bytes.Equal(out, expected)) || !ok && (out != nil || code != 0) {
						t.Fatal(code, out)
					}
					clear(raw)
					if ok && !bytes.Equal(out, expected) {
						t.Fatal("borrowed")
					}
				})
			}
		}
	}
}
func TestEquipmentDataControls(t *testing.T) {
	for _, fault := range []string{"exact", "budget", "negative", "field", "source-frame", "output-frame", "count", "depth", "repeat", "bare", "name", "path", "sentinel", "valid", "nul", "extra", "mandatory-extra", "trailing", "opaque-root", "root-count"} {
		t.Run(fault, func(t *testing.T) {
			raw := equipmentFixture(0, []byte{0}, false)
			var cause error = equipmentDataCause()
			l := wire.DefaultLimits()
			budget := 20
			switch fault {
			case "budget":
				budget = 19
			case "negative":
				budget = -1
			case "field":
				l.MaxFieldBytes = 19
			case "source-frame":
				l.MaxFrameBytes = len(raw) + 19
			case "output-frame":
				l.MaxFrameBytes = 47
			case "count":
				l.MaxAVPs = 2
			case "depth":
				l.MaxDepth = 1
			case "repeat":
				l.MaxRepeated = 0
			case "bare":
				cause = validation.At(avp.Key{Code: 458}, validation.At(avp.Key{Code: 459}, validation.ErrField))
			case "name":
				cause.(*validation.FieldError).Cause = equipmentDataMetadata{"unknown", validation.ErrField}
			case "path":
				cause.(*validation.FieldError).Path[1].Code = 459
			case "sentinel":
				cause.(*validation.FieldError).Cause = equipmentDataMetadata{"combined-equipment-data-nul-v1", validation.ErrUnsupported}
			case "valid":
				raw = equipmentFixture(0, nil, false)
			case "nul":
				raw = equipmentFixture(-1, []byte{0}, false)
			case "extra", "mandatory-extra":
				extra := unitPart(9999, nil)
				if fault == "extra" {
					extra[4] = 0
				}
				raw = unitPart(458, append(raw[8:], extra...))
			case "trailing":
				raw = append(raw, 1)
			case "opaque-root":
				raw = append(raw, unitPart(9999, []byte{1, 2, 3})...)
			case "root-count":
				raw = append(raw, unitPart(9999, nil)...)
				l.MaxAVPs = 3
			}
			out, code, ok := QuoteInvalid(raw, cause, l, budget)
			want := fault == "exact" || fault == "opaque-root"
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, equipmentDataExpected([]byte{0}))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, path := range [][]uint32{{458}, {458, 459}, {458, 460}} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%v/%s", path, fault), func(t *testing.T) {
				raw := interfaceEdit(equipmentFixture(0, []byte{0}, false), path, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					case "flags":
						b[4] |= 32
					case "vendor":
						return interfacePart(path[len(path)-1], b[8:])
					case "trailing":
						if len(path) == 1 {
							return unitPart(458, append(b[8:], 1))
						}
						return append(b, 1)
					}
					return b
				})
				out, code, ok := QuoteInvalid(raw, equipmentDataCause(), wire.DefaultLimits(), 20)
				if ok || out != nil || code != 0 {
					t.Fatal(code, out)
				}
			})
		}
	}
	for _, width := range []int{0, 1, 3, 5, 8} {
		raw := unitPart(458, append(unitPart(459, make([]byte, width)), unitPart(460, nil)...))
		if _, _, ok := QuoteInvalid(raw, equipmentDataCause(), wire.DefaultLimits(), 20); ok {
			t.Fatal("width", width)
		}
	}
}
func TestEquipmentDataExactLimits(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
		for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth"} {
			t.Run(fmt.Sprintf("%d/%s", n, fault), func(t *testing.T) {
				data := make([]byte, n)
				raw := equipmentFixture(0, data, false)
				aligned := (n + 3) &^ 3
				l := wire.DefaultLimits()
				l.MaxAVPs = 3
				l.MaxFrameBytes = 48 + aligned
				l.MaxFieldBytes = 20 + aligned
				l.MaxDepth = 2
				l.MaxRepeated = 1
				budget := 16 + aligned
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
				out, code, ok := QuoteInvalid(raw, equipmentDataCause(), l, budget)
				if len(raw)+20 != 48+aligned || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, equipmentDataExpected(data))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestEquipmentDataPadding(t *testing.T) {
	for _, data := range [][]byte{{0}, {1, 0}, {1, 0, 2}, []byte("id")} {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/%t", data, reverse), func(t *testing.T) {
				leaf := unitPart(460, data)
				for i := 8 + len(data); i < len(leaf); i++ {
					leaf[i] = 0xee
				}
				typ := actionWord(459, 0)
				b := append(append([]byte{}, typ...), leaf...)
				if reverse {
					b = append(append([]byte{}, leaf...), typ...)
				}
				raw := unitPart(458, b)
				out, code, ok := QuoteInvalid(raw, equipmentDataCause(), wire.DefaultLimits(), 4096)
				want := bytes.IndexByte(data, 0) >= 0
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, unitPart(458, leaf))) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func fuzzEquipmentData(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, equipmentDataCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) < 20 || len(out) > int(budget) || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes {
		t.Fatal("bounds")
	}
	d := wire.NewDecoder(raw, l)
	var group wire.View
	n := 0
	for {
		v, yes, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !yes {
			break
		}
		if v.Key == (avp.Key{Code: 458}) {
			group = v
			n++
		}
	}
	if n != 1 || group.Flags != 64 {
		t.Fatal("group")
	}
	child, e := d.Child(group.Data)
	if e != nil {
		t.Fatal(e)
	}
	counts := [2]int{}
	var selected wire.View
	for {
		v, yes, e := child.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !yes {
			break
		}
		if v.Flags != 64 {
			t.Fatal("flags")
		}
		switch v.Key {
		case avp.Key{Code: 459}:
			counts[0]++
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 0 {
				t.Fatal("type")
			}
		case avp.Key{Code: 460}:
			counts[1]++
			selected = v
			if bytes.IndexByte(v.Data, 0) < 0 {
				t.Fatal("counterpart")
			}
		default:
			t.Fatal("extra")
		}
	}
	size := (8 + len(selected.Data) + 3) &^ 3
	if counts != ([2]int{1, 1}) || !bytes.Equal(out, unitPart(458, group.Data[selected.Offset:selected.Offset+size])) {
		t.Fatal("exact quote")
	}
	copied := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copied, equipmentDataCause(), l, int(budget))
	clear(copied)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
