// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Exact subscription data instances, padding, counterpart controls and fuzz oracle.
// SPDX-AI-Date: 2026-11-23

package excessavp

import (
	"bytes"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type subscriptionDataMetadata struct {
	name  string
	cause error
}

func (m subscriptionDataMetadata) Error() string                           { return "data" }
func (m subscriptionDataMetadata) Unwrap() error                           { return m.cause }
func (m subscriptionDataMetadata) DiameterSubscriptionDataValueV1() string { return m.name }
func subscriptionDataCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{Code: 443}, {Code: 444}}, Cause: subscriptionDataMetadata{"combined-subscription-data-nul-v1", validation.ErrField}}
}
func subscriptionDataExpected(data []byte) []byte { return unitPart(443, unitPart(444, data)) }
func TestSubscriptionDataConsumer(t *testing.T) {
	for _, typ := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
		for i, data := range [][]byte{nil, {}, []byte("id"), {0}, {0, 1}, {1, 0}, {1, 0, 2}, {1, 2, 3, 0}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%d/%t", typ, i, reverse), func(t *testing.T) {
					raw := subscriptionFixture(typ, data, reverse)
					expected := subscriptionDataExpected(data)
					out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), len(expected))
					want := bytes.IndexByte(data, 0) >= 0
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
func TestSubscriptionDataControls(t *testing.T) {
	for _, fault := range []string{"exact", "budget", "negative", "field", "source-frame", "output-frame", "count", "depth", "repeat", "bare", "name", "path", "sentinel", "valid", "extra", "mandatory-extra", "trailing", "opaque-root", "root-count"} {
		t.Run(fault, func(t *testing.T) {
			raw := subscriptionFixture(0, []byte{0}, false)
			var cause error = subscriptionDataCause()
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
				cause = validation.At(avp.Key{Code: 443}, validation.At(avp.Key{Code: 450}, validation.ErrField))
			case "name":
				cause.(*validation.FieldError).Cause = subscriptionDataMetadata{"unknown", validation.ErrField}
			case "path":
				cause.(*validation.FieldError).Path[1].Code = 450
			case "sentinel":
				cause.(*validation.FieldError).Cause = subscriptionDataMetadata{"combined-subscription-data-nul-v1", validation.ErrUnsupported}
			case "valid":
				raw = subscriptionFixture(0, nil, false)
			case "extra", "mandatory-extra":
				extra := unitPart(9999, nil)
				if fault == "extra" {
					extra[4] = 0
				}
				raw = unitPart(443, append(raw[8:], extra...))
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
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, subscriptionDataExpected([]byte{0}))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, path := range [][]uint32{{443}, {443, 450}, {443, 444}} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%v/%s", path, fault), func(t *testing.T) {
				raw := interfaceEdit(subscriptionFixture(0, []byte{0}, false), path, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						if len(path) == 1 {
							return append(append(append([]byte{}, b...), b...), b...)
						}
						return append(b, b...)
					case "flags":
						b[4] |= 32
					case "vendor":
						return interfacePart(path[len(path)-1], b[8:])
					case "trailing":
						if len(path) == 1 {
							return unitPart(443, append(b[8:], 1))
						}
						return append(b, 1)
					}
					return b
				})
				out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), 20)
				if ok || out != nil || code != 0 {
					t.Fatal(code, out)
				}
			})
		}
	}
	for _, width := range []int{0, 1, 3, 5, 8} {
		raw := unitPart(443, append(unitPart(450, make([]byte, width)), unitPart(444, []byte{0})...))
		if _, _, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), 20); ok {
			t.Fatal("width", width)
		}
	}
}
func TestSubscriptionDataExactLimits(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
		for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth"} {
			t.Run(fmt.Sprintf("%d/%s", n, fault), func(t *testing.T) {
				data := make([]byte, n)
				raw := subscriptionFixture(0, data, false)
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
				out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), l, budget)
				if len(raw)+20 != 48+aligned || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, subscriptionDataExpected(data))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestSubscriptionDataPadding(t *testing.T) {
	for _, data := range [][]byte{{0}, {1, 0}, {1, 0, 2}, []byte("id")} {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/%t", data, reverse), func(t *testing.T) {
				leaf := unitPart(444, data)
				for i := 8 + len(data); i < len(leaf); i++ {
					leaf[i] = 0xee
				}
				typ := actionWord(450, 0)
				b := append(append([]byte{}, typ...), leaf...)
				if reverse {
					b = append(append([]byte{}, leaf...), typ...)
				}
				raw := unitPart(443, b)
				out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), 4096)
				want := bytes.IndexByte(data, 0) >= 0
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, unitPart(443, leaf))) {
					t.Fatal(code, out)
				}
			})
		}
	}
}

func subscriptionFixture(typ int32, data []byte, reverse bool) []byte {
	a, b := actionWord(450, uint32(typ)), unitPart(444, data)
	if reverse {
		a, b = b, a
	}
	return unitPart(443, append(a, b...))
}
func TestSubscriptionDataRepeated(t *testing.T) {
	samples := [][]byte{nil, []byte("id"), {0}, {1, 0, 2}}
	for i, first := range samples {
		for j, second := range samples {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%d/%t", i, j, reverse), func(t *testing.T) {
					a, b := subscriptionFixture(-1, first, reverse), subscriptionFixture(2147483647, second, !reverse)
					raw := append(a, b...)
					out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), 4096)
					var want []byte
					for _, data := range [][]byte{first, second} {
						if bytes.IndexByte(data, 0) >= 0 {
							want = subscriptionDataExpected(data)
							break
						}
					}
					if ok != (want != nil) || ok && (code != 5004 || !bytes.Equal(out, want)) || !ok && (out != nil || code != 0) {
						t.Fatal(code, out)
					}
					clear(raw)
					if ok && !bytes.Equal(out, want) {
						t.Fatal("borrowed")
					}
				})
			}
		}
	}
	for _, position := range []string{"before", "selected", "after"} {
		for _, fault := range []string{"missing-type", "missing-data", "duplicate-type", "duplicate-data", "extra", "mandatory-extra", "type-width", "trailing"} {
			t.Run(position+"/"+fault, func(t *testing.T) {
				bad := subscriptionFixture(0, []byte{0}, false)
				if position != "selected" {
					bad = subscriptionFixture(0, []byte("ok"), false)
				}
				switch fault {
				case "missing-type":
					bad = unitPart(443, unitPart(444, []byte{0}))
				case "missing-data":
					bad = unitPart(443, actionWord(450, 0))
				case "duplicate-type":
					bad = unitPart(443, append(bad[8:], actionWord(450, 0)...))
				case "duplicate-data":
					bad = unitPart(443, append(bad[8:], unitPart(444, []byte{0})...))
				case "extra", "mandatory-extra":
					part := unitPart(9999, nil)
					if fault == "extra" {
						part[4] = 0
					}
					bad = unitPart(443, append(bad[8:], part...))
				case "type-width":
					bad = unitPart(443, append(unitPart(450, []byte{1}), unitPart(444, []byte{0})...))
				case "trailing":
					bad = unitPart(443, append(bad[8:], 1))
				}
				selected := subscriptionFixture(1, []byte{1, 0, 2}, false)
				raw := bad
				if position == "before" {
					raw = append(bad, selected...)
				}
				if position == "after" {
					raw = append(selected, bad...)
				}
				out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), wire.DefaultLimits(), 4096)
				if ok != (position == "after") || ok && (code != 5004 || !bytes.Equal(out, subscriptionDataExpected([]byte{1, 0, 2}))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestSubscriptionDataRepeatedLimits(t *testing.T) {
	for _, selected := range []int{0, 1} {
		for _, fault := range []string{"exact", "count", "repeat", "frame", "field", "depth", "quote"} {
			t.Run(fmt.Sprintf("%d/%s", selected, fault), func(t *testing.T) {
				first, second := subscriptionFixture(0, []byte{0}, false), subscriptionFixture(1, []byte("ok"), false)
				if selected == 1 {
					first, second = second, first
				}
				raw := append(first, second...)
				l := wire.DefaultLimits()
				l.MaxAVPs = 4 + 2*selected
				l.MaxRepeated = 2
				l.MaxFieldBytes = 24
				l.MaxFrameBytes = 84
				l.MaxDepth = 2
				budget := 20
				switch fault {
				case "count":
					l.MaxAVPs--
				case "repeat":
					l.MaxRepeated--
				case "frame":
					l.MaxFrameBytes--
				case "field":
					l.MaxFieldBytes--
				case "depth":
					l.MaxDepth--
				case "quote":
					budget--
				}
				out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), l, budget)
				if len(raw)+20 != 84 || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, subscriptionDataExpected([]byte{0}))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func fuzzSubscriptionData(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, subscriptionDataCause(), l, int(budget))
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
	var groups []wire.View
	for {
		v, more, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		if v.Key == (avp.Key{Code: 443}) {
			if v.Flags != 64 {
				t.Fatal("group flags")
			}
			groups = append(groups, v)
		}
	}
	if len(groups) < 1 || len(groups) > 2 {
		t.Fatal("root count")
	}
	var want []byte
	for _, g := range groups {
		child, e := d.Child(g.Data)
		if e != nil {
			t.Fatal(e)
		}
		counts := [2]int{}
		var leaf wire.View
		for {
			v, more, e := child.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !more {
				break
			}
			if v.Flags != 64 {
				t.Fatal("child flags")
			}
			switch v.Key {
			case avp.Key{Code: 450}:
				counts[0]++
				if len(v.Data) != 4 {
					t.Fatal("width")
				}
			case avp.Key{Code: 444}:
				counts[1]++
				leaf = v
			default:
				t.Fatal("extra")
			}
		}
		if counts != ([2]int{1, 1}) {
			t.Fatal("child counts")
		}
		if bytes.IndexByte(leaf.Data, 0) >= 0 {
			size := (8 + len(leaf.Data) + 3) &^ 3
			want = unitPart(443, g.Data[leaf.Offset:leaf.Offset+size])
			break
		}
	}
	if want == nil || !bytes.Equal(out, want) {
		t.Fatal("first failing instance")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, subscriptionDataCause(), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
