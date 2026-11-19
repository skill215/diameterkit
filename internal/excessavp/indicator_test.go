// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined indicator value acceptance.
// SPDX-AI-Date: 2026-11-19

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

type indicatorMetadata struct {
	name  string
	cause error
}

func (m indicatorMetadata) Error() string                    { return "indicator" }
func (m indicatorMetadata) Unwrap() error                    { return m.cause }
func (m indicatorMetadata) DiameterIndicatorValueV1() string { return m.name }
func indicatorCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{Code: 455}}, Cause: indicatorMetadata{"combined-sms-indicator-v1", validation.ErrField}}
}
func indicatorFixture(kind uint32, value int32) []byte {
	b := append(actionWord(416, kind), actionWord(455, uint32(value))...)
	return append(b, unitPart(458, nil)...)
}
func TestIndicatorValueConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3, 4} {
		for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			t.Run(fmt.Sprintf("%d/%d", kind, value), func(t *testing.T) {
				raw := indicatorFixture(kind, value)
				out, code, ok := QuoteInvalid(raw, indicatorCause(), wire.DefaultLimits(), 12)
				want := value != 1
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, actionWord(455, uint32(value)))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
				clear(raw)
				if ok && !bytes.Equal(out, actionWord(455, uint32(value))) {
					t.Fatal("borrowed")
				}
			})
		}
	}
}
func TestIndicatorValueControls(t *testing.T) {
	for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "path", "sentinel", "kind", "value", "service", "timezone", "refund", "trailing", "opaque"} {
		t.Run(fault, func(t *testing.T) {
			raw := indicatorFixture(1, 0)
			var cause error = indicatorCause()
			l := wire.DefaultLimits()
			budget := 12
			switch fault {
			case "budget":
				budget = 11
			case "negative":
				budget = -1
			case "field":
				l.MaxFieldBytes = 11
			case "frame":
				l.MaxFrameBytes = len(raw) + 19
			case "count":
				l.MaxAVPs = 2
			case "depth":
				l.MaxDepth = 0
			case "repeat":
				l.MaxRepeated = 0
			case "bare":
				cause = validation.At(avp.Key{Code: 455}, validation.ErrField)
			case "name":
				cause.(*validation.FieldError).Cause = indicatorMetadata{"unreviewed", validation.ErrField}
			case "sentinel":
				cause.(*validation.FieldError).Cause = indicatorMetadata{"combined-sms-indicator-v1", validation.ErrUnsupported}
			case "path":
				cause.(*validation.FieldError).Path[0].Code = 416
			case "kind":
				raw = indicatorFixture(2, 0)
			case "value":
				raw = indicatorFixture(1, 1)
			case "service":
				raw = append(raw, actionWord(439, 1)...)
			case "timezone":
				b := make([]byte, 16)
				binary.BigEndian.PutUint32(b, 23)
				b[4] = 192
				wire.Put24(b[5:8], 16)
				binary.BigEndian.PutUint32(b[8:12], 10425)
				raw = append(raw, b...)
			case "refund":
				raw = append(raw, unitPart(1292, nil)...)
			case "trailing":
				raw = append(raw, 1)
			case "opaque":
				raw = append(raw, unitPart(9999, []byte{1, 2, 3})...)
			}
			out, code, ok := QuoteInvalid(raw, cause, l, budget)
			want := fault == "exact" || fault == "opaque"
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, actionWord(455, 0))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, key := range []uint32{416, 455, 458} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "width"} {
			t.Run(fmt.Sprintf("%d/%s", key, fault), func(t *testing.T) {
				raw := interfaceEdit(indicatorFixture(1, 0), []uint32{key}, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					case "flags":
						b[4] |= 32
					case "vendor":
						return interfacePart(key, b[8:])
					case "width":
						if key == 458 {
							return unitPart(key, []byte{1})
						}
						return unitPart(key, []byte{1})
					}
					return b
				})
				out, code, ok := QuoteInvalid(raw, indicatorCause(), wire.DefaultLimits(), 12)
				// The presence marker's child payload remains opaque; its schema validates it before the root rule.
				want := key == 458 && fault == "width"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, actionWord(455, 0))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestIndicatorValueExactLimits(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxAVPs = 3
	l.MaxFieldBytes = 12
	l.MaxFrameBytes = 52
	l.MaxDepth = 1
	l.MaxRepeated = 1
	out, code, ok := QuoteInvalid(indicatorFixture(1, 0), indicatorCause(), l, 12)
	if !ok || code != 5004 || !bytes.Equal(out, actionWord(455, 0)) {
		t.Fatal(code, out)
	}
}
func fuzzIndicatorValue(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, indicatorCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) != 12 || len(out) > int(budget) {
		t.Fatal("bound")
	}
	counts := map[avp.Key]int{}
	d := wire.NewDecoder(raw, l)
	var selected wire.View
	for {
		v, yes, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !yes {
			break
		}
		counts[v.Key]++
		switch v.Key {
		case avp.Key{Code: 416}:
			if v.Flags != 64 || len(v.Data) != 4 {
				t.Fatal("kind shape")
			}
			n := binary.BigEndian.Uint32(v.Data)
			if n != 1 && n != 3 && n != 4 {
				t.Fatal("kind")
			}
		case avp.Key{Code: 455}:
			selected = v
			if v.Flags != 64 || len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) == 1 {
				t.Fatal("value")
			}
		case avp.Key{Code: 458}:
			if v.Flags != 64 {
				t.Fatal("marker")
			}
		}
	}
	for _, k := range []avp.Key{{Code: 416}, {Code: 455}, {Code: 458}} {
		if counts[k] != 1 {
			t.Fatal("ambiguous")
		}
	}
	for _, k := range []avp.Key{{Code: 439}, {VendorID: 10425, Code: 23}, {Code: 1292}} {
		if counts[k] != 0 {
			t.Fatal("mixed")
		}
	}
	if !bytes.Equal(out, raw[selected.Offset:selected.Offset+12]) {
		t.Fatal("instance")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, indicatorCause(), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
