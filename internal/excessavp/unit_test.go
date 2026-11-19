// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Unit value original-instance, condition, budget and fuzz acceptance.
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

type unitMetadata struct {
	rule  string
	cause error
}

func (m *unitMetadata) Error() string               { return "unit value" }
func (m *unitMetadata) Unwrap() error               { return m.cause }
func (m *unitMetadata) DiameterUnitValueV1() string { return m.rule }
func unitCause(reason bool) *validation.FieldError {
	path := []avp.Key{{Code: 456}, {Code: 437}, {Code: 417}}
	rule := "combined-sms-requested-units-v1"
	if reason {
		path = []avp.Key{{Code: 456}, {VendorID: 10415, Code: 872}}
		rule = "combined-termination-reporting-reason-v1"
	}
	return &validation.FieldError{Path: path, Cause: &unitMetadata{rule, errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func unitPart(code uint32, data []byte) []byte {
	b := make([]byte, (8+len(data)+3)&^3)
	binary.BigEndian.PutUint32(b, code)
	b[4] = 64
	wire.Put24(b[5:8], uint32(8+len(data)))
	copy(b[8:], data)
	return b
}
func unitLeaf(reason bool, value uint64) []byte {
	if reason {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(value))
		return interfacePart(872, b)
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, value)
	return unitPart(417, b)
}
func unitFixture(kind uint32, value uint64) []byte {
	raw := append(actionWord(416, kind), actionWord(455, 1)...)
	raw = append(raw, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
	var mscc []byte
	if kind == 3 {
		mscc = append(unitPart(446, unitLeaf(false, 7)), unitLeaf(true, value)...)
	} else {
		mscc = unitPart(437, unitLeaf(false, value))
	}
	return append(raw, unitPart(456, mscc)...)
}
func unitExpected(reason bool, value uint64) []byte {
	leaf := unitLeaf(reason, value)
	if !reason {
		leaf = unitPart(437, leaf)
	}
	return unitPart(456, leaf)
}
func TestUnitValueConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3, 4} {
		for _, value := range []uint64{0, 1, 2, 3, 0x7fffffff, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)} {
			t.Run(fmt.Sprintf("%d/%d", kind, value), func(t *testing.T) {
				reason := kind == 3
				raw := unitFixture(kind, value)
				expected := unitExpected(reason, value)
				want := value != 1
				if reason {
					want = uint32(value) != 2
				}
				out, code, ok := QuoteInvalid(raw, unitCause(reason), wire.DefaultLimits(), len(expected))
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
func TestUnitValueControls(t *testing.T) {
	for _, reason := range []bool{false, true} {
		kind := uint32(1)
		if reason {
			kind = 3
		}
		selected := []uint32{456, 437, 417}
		if reason {
			selected = []uint32{456, 872}
		}
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "source-count", "output-count", "depth", "repeat", "bare", "sentinel", "name", "path", "kind", "indicator", "equipment", "valid", "width", "forbidden-requested-used", "forbidden-reason-money", "forbidden-standard", "counterpart-width", "trailing", "opaque-sibling"} {
			t.Run(fmt.Sprintf("%t/%s", reason, fault), func(t *testing.T) {
				raw := unitFixture(kind, 0)
				cause := unitCause(reason)
				l := wire.DefaultLimits()
				budget := len(unitExpected(reason, 0))
				edit := func(path []uint32, f func([]byte) []byte) { raw = interfaceEdit(raw, path, f) }
				addMSCC := func(b []byte) {
					edit([]uint32{456}, func(old []byte) []byte { return unitPart(456, append(old[8:], b...)) })
				}
				switch fault {
				case "budget":
					budget--
				case "negative":
					budget = -1
				case "field":
					l.MaxFieldBytes = budget - 1
				case "frame":
					l.MaxFrameBytes = len(raw) + 19
				case "source-count":
					l.MaxAVPs = 5
					if reason {
						l.MaxAVPs = 6
					}
				case "output-count":
					l.MaxAVPs = len(selected)
				case "depth":
					l.MaxDepth = len(selected) - 1
				case "repeat":
					l.MaxRepeated = 0
				case "bare":
					cause.Cause = validation.ErrField
				case "sentinel":
					cause.Cause.(*unitMetadata).cause = validation.ErrUnsupported
				case "name":
					cause.Cause.(*unitMetadata).rule = "unreviewed"
				case "path":
					cause.Path[0].Code = 437
				case "kind":
					edit([]uint32{416}, func([]byte) []byte { return actionWord(416, 2) })
				case "indicator":
					edit([]uint32{455}, func([]byte) []byte { return actionWord(455, 0) })
				case "equipment":
					edit([]uint32{458}, func([]byte) []byte { return nil })
				case "valid":
					v := uint64(1)
					if reason {
						v = 2
					}
					edit(selected, func([]byte) []byte { return unitLeaf(reason, v) })
				case "width":
					edit(selected, func([]byte) []byte {
						if reason {
							return interfacePart(872, []byte{1})
						}
						return unitPart(417, []byte{1})
					})
				case "forbidden-requested-used":
					code := uint32(446)
					if reason {
						code = 437
					}
					addMSCC(unitPart(code, unitLeaf(false, 1)))
				case "forbidden-reason-money":
					if reason {
						addMSCC(unitPart(437, nil))
					} else {
						addMSCC(unitLeaf(true, 2))
					}
				case "forbidden-standard":
					if reason {
						addMSCC(unitPart(437, nil))
					} else {
						edit([]uint32{456, 437}, func(b []byte) []byte { return unitPart(437, append(b[8:], unitPart(413, nil)...)) })
					}
				case "counterpart-width":
					if reason {
						edit([]uint32{456, 446, 417}, func([]byte) []byte { return unitPart(417, []byte{1}) })
					} else {
						edit([]uint32{456, 437}, func(b []byte) []byte { return unitPart(437, append(b[8:], unitPart(420, nil)...)) })
					}
				case "trailing":
					raw = append(raw, 1)
				case "opaque-sibling":
					addMSCC(unitPart(9999, []byte{1, 2, 3}))
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "exact" || fault == "opaque-sibling"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, unitExpected(reason, 0))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		paths := [][]uint32{{416}, {455}, {458}, {456}, selected}
		if reason {
			paths = append(paths, []uint32{456, 446}, []uint32{456, 446, 417})
		} else {
			paths = append(paths, []uint32{456, 437})
		}
		for _, path := range paths {
			for _, fault := range []string{"missing", "duplicate", "flags", "key", "trailing"} {
				t.Run(fmt.Sprintf("%t/%v/%s", reason, path, fault), func(t *testing.T) {
					raw := interfaceEdit(unitFixture(kind, 0), path, func(b []byte) []byte {
						switch fault {
						case "missing":
							return nil
						case "duplicate":
							return append(b, b...)
						case "flags":
							b[4] |= 32
						case "key":
							binary.BigEndian.PutUint32(b, 9999)
						case "trailing":
							if path[len(path)-1] == 456 || path[len(path)-1] == 437 || path[len(path)-1] == 446 {
								return unitPart(path[len(path)-1], append(b[8:], 1))
							}
							return append(b, 1)
						}
						return b
					})
					if out, code, ok := QuoteInvalid(raw, unitCause(reason), wire.DefaultLimits(), 32); ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
	}
}
func TestUnitValueExactLimits(t *testing.T) {
	for _, reason := range []bool{false, true} {
		kind := uint32(1)
		l := wire.DefaultLimits()
		l.MaxAVPs = 6
		l.MaxFieldBytes = 32
		l.MaxFrameBytes = 84
		l.MaxDepth = 3
		if reason {
			kind = 3
			l.MaxAVPs = 7
			l.MaxFieldBytes = 40
			l.MaxFrameBytes = 100
			l.MaxDepth = 2
		}
		l.MaxRepeated = 1
		raw := unitFixture(kind, 0)
		out, code, ok := QuoteInvalid(raw, unitCause(reason), l, len(unitExpected(reason, 0)))
		if len(raw)+20 != l.MaxFrameBytes || !ok || code != 5004 || !bytes.Equal(out, unitExpected(reason, 0)) {
			t.Fatal(len(raw), code, out)
		}
	}
}
func fuzzUnitValue(t *testing.T, raw []byte, budget uint16, reason bool) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, unitCause(reason), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial quote")
		}
		return
	}
	expectedLen := 32
	if reason {
		expectedLen = 24
	}
	if code != 5004 || len(out) != expectedLen || len(out) > int(budget) {
		t.Fatal("bounds")
	}
	source, quote := raw, out
	path := unitCause(reason).Path
	for i, key := range path {
		d := wire.NewDecoder(source, l)
		var found wire.View
		count := 0
		for {
			v, yes, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !yes {
				break
			}
			if v.Key == key {
				found = v
				count++
			}
		}
		q := wire.NewDecoder(quote, l)
		v, yes, e := q.Next()
		if e != nil || !yes || count != 1 || q.Remaining() != 0 || v.Key != key || v.Flags != found.Flags {
			t.Fatal("instance grammar")
		}
		if i == len(path)-1 {
			if !bytes.Equal(quote, source[found.Offset:found.Offset+16]) {
				t.Fatal("changed leaf")
			}
			if reason {
				if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) == 2 {
					t.Fatal("reason")
				}
			} else if len(v.Data) != 8 || binary.BigEndian.Uint64(v.Data) == 1 {
				t.Fatal("units")
			}
		}
		source, quote = found.Data, v.Data
	}
	copied := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copied, unitCause(reason), l, int(budget))
	clear(copied)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
