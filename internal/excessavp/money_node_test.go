// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Money SMS-Node raw counterparts, bounded quotes and independent fuzz oracle.
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

type moneyNodeMetadata struct {
	name  string
	cause error
}

func (m moneyNodeMetadata) Error() string                    { return "money node" }
func (m moneyNodeMetadata) Unwrap() error                    { return m.cause }
func (m moneyNodeMetadata) DiameterMoneyNodeValueV1() string { return m.name }
func moneyNodeCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2016}}, Cause: moneyNodeMetadata{"combined-money-sms-node-v1", errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func moneyNodeFixture(action uint32, node int32) []byte {
	raw := append(actionWord(416, 4), actionWord(415, 0)...)
	raw = append(raw, actionWord(436, action)...)
	raw = append(raw, unitPart(456, unitPart(437, nil))...)
	return append(raw, smsExpected(true, node)...)
}
func TestMoneyNodeConsumer(t *testing.T) {
	for _, a := range []uint32{0, 1, 2, 3, ^uint32(0)} {
		for _, v := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			t.Run(fmt.Sprintf("%d/%d", a, v), func(t *testing.T) {
				raw := moneyNodeFixture(a, v)
				out, code, ok := QuoteInvalid(raw, moneyNodeCause(), wire.DefaultLimits(), 40)
				want := a <= 2 && v != 2
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, smsExpected(true, v))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
				clear(raw)
				if ok && !bytes.Equal(out, smsExpected(true, v)) {
					t.Fatal("borrowed")
				}
			})
		}
	}
}
func TestMoneyNodeControls(t *testing.T) {
	for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "sentinel", "path", "valid", "kind", "number", "action", "trailing", "root-count", "opaque", "message"} {
		t.Run(fault, func(t *testing.T) {
			raw := moneyNodeFixture(0, -1)
			var cause error = moneyNodeCause()
			l := wire.DefaultLimits()
			budget := 40
			switch fault {
			case "budget":
				budget = 39
			case "negative":
				budget = -1
			case "field":
				l.MaxFieldBytes = 39
			case "frame":
				l.MaxFrameBytes = len(raw) + 19
			case "count":
				l.MaxAVPs = 7
			case "depth":
				l.MaxDepth = 2
			case "repeat":
				l.MaxRepeated = 0
			case "bare":
				cause = validation.ErrField
			case "name":
				cause.(*validation.FieldError).Cause = moneyNodeMetadata{"unknown", errors.Join(validation.ErrField, validation.ErrUnsupported)}
			case "sentinel":
				cause.(*validation.FieldError).Cause = moneyNodeMetadata{"combined-money-sms-node-v1", validation.ErrUnsupported}
			case "path":
				cause.(*validation.FieldError).Path[0].VendorID = 0
			case "valid":
				raw = moneyNodeFixture(0, 2)
			case "action":
				raw = moneyNodeFixture(3, -1)
			case "kind", "number":
				c := uint32(416)
				if fault == "number" {
					c = 415
				}
				raw = interfaceEdit(raw, []uint32{c}, func([]byte) []byte { return actionWord(c, 1) })
			case "trailing":
				raw = append(raw, 1)
			case "opaque", "root-count":
				raw = append(raw, unitPart(9999, []byte{1})...)
				if fault == "root-count" {
					l.MaxAVPs = 8
				}
			case "message":
				raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], smsWord(2007, -99)...)) })
			}
			out, code, ok := QuoteInvalid(raw, cause, l, budget)
			want := fault == "exact" || fault == "opaque" || fault == "message"
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, smsExpected(true, -1))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, p := range [][]uint32{{416}, {415}, {436}, {456}, {456, 437}, {873}, {873, 2000}, {873, 2000, 2016}} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%v/%s", p, fault), func(t *testing.T) {
				raw := interfaceEdit(moneyNodeFixture(2, -1), p, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					case "flags":
						b[4] |= 32
					case "vendor":
						if b[4]&128 != 0 {
							binary.BigEndian.PutUint32(b[8:], 10416)
						} else {
							binary.BigEndian.PutUint32(b, 9999)
						}
					case "trailing":
						if p[len(p)-1] == 456 || p[len(p)-1] == 437 {
							return unitPart(p[len(p)-1], append(b[8:], 1))
						}
						if len(p) < 3 && p[0] == 873 {
							return interfacePart(p[len(p)-1], append(b[12:], 1))
						}
						return append(b, 1)
					}
					return b
				})
				_, _, ok := QuoteInvalid(raw, moneyNodeCause(), wire.DefaultLimits(), 40)
				if ok && !(len(p) == 2 && p[0] == 456 && fault == "trailing") {
					t.Fatal("accepted")
				}
				if !ok && len(p) == 2 && p[0] == 456 && fault == "trailing" {
					t.Fatal("requested-unit content is opaque")
				}
			})
		}
	}
	for _, c := range []uint32{455, 458, 446, 872, 2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		t.Run(fmt.Sprintf("forbidden/%d", c), func(t *testing.T) {
			raw := moneyNodeFixture(1, -1)
			switch c {
			case 455, 458:
				raw = append(raw, unitPart(c, nil)...)
			case 446, 872:
				raw = interfaceEdit(raw, []uint32{456}, func(b []byte) []byte {
					x := unitPart(c, nil)
					if c == 872 {
						x = interfacePart(c, nil)
					}
					return unitPart(456, append(b[8:], x...))
				})
			default:
				raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], interfacePart(c, nil)...)) })
			}
			if out, code, ok := QuoteInvalid(raw, moneyNodeCause(), wire.DefaultLimits(), 40); ok || out != nil || code != 0 {
				t.Fatal(code, out)
			}
		})
	}
	for _, p := range [][]uint32{{416}, {415}, {436}, {873, 2000, 2016}} {
		for _, width := range []int{0, 1, 3, 5, 8} {
			raw := interfaceEdit(moneyNodeFixture(0, -1), p, func([]byte) []byte {
				if len(p) == 1 {
					return unitPart(p[0], make([]byte, width))
				}
				return interfacePart(2016, make([]byte, width))
			})
			if _, _, ok := QuoteInvalid(raw, moneyNodeCause(), wire.DefaultLimits(), 40); ok {
				t.Fatal("width", p, width)
			}
		}
	}
}
func TestMoneyNodeExactLimits(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxAVPs = 8
	l.MaxFrameBytes = 112
	l.MaxFieldBytes = 40
	l.MaxDepth = 3
	l.MaxRepeated = 1
	raw := moneyNodeFixture(2, -1)
	out, code, ok := QuoteInvalid(raw, moneyNodeCause(), l, 40)
	if len(raw)+20 != 112 || !ok || code != 5004 || !bytes.Equal(out, smsExpected(true, -1)) {
		t.Fatal(code, out)
	}
}
func fuzzMoneyNode(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, moneyNodeCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) != 40 || len(out) > int(budget) {
		t.Fatal("bound")
	}
	// Independent oracle: enumerate full source areas with one shared decoder
	// counter, then check branch counterparts and reconstruct the exact leaf.
	scan := func(d *wire.Decoder) map[avp.Key][]wire.View {
		m := map[avp.Key][]wire.View{}
		for {
			v, yes, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !yes {
				break
			}
			m[v.Key] = append(m[v.Key], v)
		}
		return m
	}
	get := func(m map[avp.Key][]wire.View, k avp.Key, flags byte) wire.View {
		v := m[k]
		if len(v) != 1 || v[0].Flags != flags {
			t.Fatal("instance", k)
		}
		return v[0]
	}
	absent := func(m map[avp.Key][]wire.View, k avp.Key) {
		if len(m[k]) != 0 {
			t.Fatal("forbidden", k)
		}
	}
	child := func(d *wire.Decoder, b []byte) *wire.Decoder {
		c, e := d.Child(b)
		if e != nil {
			t.Fatal(e)
		}
		return c
	}
	d := wire.NewDecoder(raw, l)
	top := scan(d)
	for _, c := range []uint32{455, 458} {
		absent(top, avp.Key{Code: c})
	}
	for _, c := range []uint32{416, 415, 436} {
		v := get(top, avp.Key{Code: c}, 64)
		if len(v.Data) != 4 {
			t.Fatal("width")
		}
		n := binary.BigEndian.Uint32(v.Data)
		if c == 416 && n != 4 || c == 415 && n != 0 || c == 436 && n > 2 {
			t.Fatal("branch")
		}
	}
	units := get(top, avp.Key{Code: 456}, 64)
	u := scan(child(d, units.Data))
	get(u, avp.Key{Code: 437}, 64)
	absent(u, avp.Key{Code: 446})
	absent(u, avp.Key{VendorID: 10415, Code: 872})
	service := get(top, avp.Key{VendorID: 10415, Code: 873}, 192)
	sd := child(d, service.Data)
	s := scan(sd)
	sms := get(s, avp.Key{VendorID: 10415, Code: 2000}, 192)
	sm := scan(child(sd, sms.Data))
	for _, c := range []uint32{2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		absent(sm, avp.Key{VendorID: 10415, Code: c})
	}
	leaf := get(sm, avp.Key{VendorID: 10415, Code: 2016}, 192)
	if len(leaf.Data) != 4 || binary.BigEndian.Uint32(leaf.Data) == 2 {
		t.Fatal("value")
	}
	expected := interfacePart(873, interfacePart(2000, sms.Data[leaf.Offset:leaf.Offset+16]))
	if !bytes.Equal(out, expected) {
		t.Fatal("exact quote")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, moneyNodeCause(), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
