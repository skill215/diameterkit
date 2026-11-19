// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined equipment-type value acceptance.
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

type equipmentMetadata struct {
	name  string
	cause error
}

func (m equipmentMetadata) Error() string                        { return "equipment" }
func (m equipmentMetadata) Unwrap() error                        { return m.cause }
func (m equipmentMetadata) DiameterEquipmentTypeValueV1() string { return m.name }
func equipmentCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{Code: 458}, {Code: 459}}, Cause: equipmentMetadata{"combined-equipment-type-v1", errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func equipmentFixture(value int32, data []byte, reverse bool) []byte {
	a, b := actionWord(459, uint32(value)), unitPart(460, data)
	if reverse {
		a, b = b, a
	}
	return unitPart(458, append(a, b...))
}
func equipmentExpected(value int32) []byte { return unitPart(458, actionWord(459, uint32(value))) }
func TestEquipmentValueConsumer(t *testing.T) {
	for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
		for _, data := range [][]byte{nil, []byte("identity"), bytes.Repeat([]byte{'A'}, 1024), {0}, {'A', 0, 'B'}} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%x/%t", value, data, reverse), func(t *testing.T) {
					raw := equipmentFixture(value, data, reverse)
					out, code, ok := QuoteInvalid(raw, equipmentCause(), wire.DefaultLimits(), 20)
					want := value != 0 && bytes.IndexByte(data, 0) < 0
					if ok != want || ok && (code != 5004 || !bytes.Equal(out, equipmentExpected(value))) || !ok && (out != nil || code != 0) {
						t.Fatal(code, out)
					}
					clear(raw)
					if ok && !bytes.Equal(out, equipmentExpected(value)) {
						t.Fatal("borrowed")
					}
				})
			}
		}
	}
}
func TestEquipmentValueControls(t *testing.T) {
	for _, fault := range []string{"exact", "budget", "negative", "field", "source-frame", "output-frame", "count", "depth", "repeat", "bare", "name", "path", "sentinel", "valid", "nul", "extra", "mandatory-extra", "trailing", "opaque-root", "root-count"} {
		t.Run(fault, func(t *testing.T) {
			raw := equipmentFixture(-1, []byte("id"), false)
			var cause error = equipmentCause()
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
				cause.(*validation.FieldError).Cause = equipmentMetadata{"unknown", errors.Join(validation.ErrUnsupported, validation.ErrField)}
			case "path":
				cause.(*validation.FieldError).Path[1].Code = 460
			case "sentinel":
				cause.(*validation.FieldError).Cause = equipmentMetadata{"combined-equipment-type-v1", validation.ErrUnsupported}
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
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, equipmentExpected(-1))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, path := range [][]uint32{{458}, {458, 459}, {458, 460}} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%v/%s", path, fault), func(t *testing.T) {
				raw := interfaceEdit(equipmentFixture(-1, []byte("id"), false), path, func(b []byte) []byte {
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
				out, code, ok := QuoteInvalid(raw, equipmentCause(), wire.DefaultLimits(), 20)
				if ok || out != nil || code != 0 {
					t.Fatal(code, out)
				}
			})
		}
	}
	for _, width := range []int{0, 1, 3, 5, 8} {
		raw := unitPart(458, append(unitPart(459, make([]byte, width)), unitPart(460, nil)...))
		if _, _, ok := QuoteInvalid(raw, equipmentCause(), wire.DefaultLimits(), 20); ok {
			t.Fatal("width", width)
		}
	}
}
func TestEquipmentValueExactLimits(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxAVPs = 3
	l.MaxFieldBytes = 20
	l.MaxFrameBytes = 48
	l.MaxDepth = 2
	l.MaxRepeated = 1
	out, code, ok := QuoteInvalid(equipmentFixture(-1, nil, false), equipmentCause(), l, 20)
	if !ok || code != 5004 || !bytes.Equal(out, equipmentExpected(-1)) {
		t.Fatal(code, out)
	}
}
func fuzzEquipmentValue(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, equipmentCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) != 20 || len(out) > int(budget) {
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
			selected = v
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) == 0 {
				t.Fatal("type")
			}
		case avp.Key{Code: 460}:
			counts[1]++
			if bytes.IndexByte(v.Data, 0) >= 0 {
				t.Fatal("counterpart")
			}
		default:
			t.Fatal("extra")
		}
	}
	if counts != ([2]int{1, 1}) || !bytes.Equal(out[8:], group.Data[selected.Offset:selected.Offset+12]) || !bytes.Equal(out[:8], []byte{0, 0, 1, 202, 64, 0, 0, 20}) {
		t.Fatal("exact quote")
	}
	copied := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copied, equipmentCause(), l, int(budget))
	clear(copied)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
