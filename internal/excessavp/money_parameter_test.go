// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Exact parameter instances, conditional counterparts, budgets and fuzz oracle.
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

type moneyParameterMetadata struct {
	name  string
	cause error
}

func (m moneyParameterMetadata) Error() string                         { return "parameter" }
func (m moneyParameterMetadata) Unwrap() error                         { return m.cause }
func (m moneyParameterMetadata) DiameterMoneyParameterValueV1() string { return m.name }
func moneyParameterCause(zone bool) *validation.FieldError {
	k := avp.Key{Code: 439}
	name := "combined-money-service-identifier-v1"
	if zone {
		k = avp.Key{VendorID: 10425, Code: 23}
		name = "combined-money-timezone-v1"
	}
	return &validation.FieldError{Path: []avp.Key{k}, Cause: moneyParameterMetadata{name, validation.ErrField}}
}
func moneyParameterWord(zone bool, v int32) []byte {
	if !zone {
		return actionWord(439, uint32(v))
	}
	b := smsWord(23, v)
	binary.BigEndian.PutUint32(b[8:], 10425)
	return b
}
func moneyParameterFixture(action uint32, zone bool, v int32) []byte {
	id, tz := int32(1), int32(0)
	if zone {
		tz = v
	} else {
		id = v
	}
	raw := append(actionWord(416, 4), actionWord(415, 0)...)
	raw = append(raw, actionWord(436, action)...)
	raw = append(raw, moneyParameterWord(false, id)...)
	raw = append(raw, moneyParameterWord(true, tz)...)
	raw = append(raw, unitPart(456, unitPart(437, unitPart(413, nil)))...)
	sms := interfacePart(2000, append(smsWord(2016, 2), interfacePart(2018, nil)...))
	return append(raw, interfacePart(873, append(sms, interfacePart(877, nil)...))...)
}
func TestMoneyParameterConsumer(t *testing.T) {
	for _, action := range []uint32{0, 1, 2, 3, ^uint32(0)} {
		for _, zone := range []bool{false, true} {
			for _, v := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
				for _, reverse := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/%t/%d/%t", action, zone, v, reverse), func(t *testing.T) {
						raw := moneyParameterFixture(action, zone, v)
						if reverse {
							d := wire.NewDecoder(raw, wire.DefaultLimits())
							var rev []byte
							for {
								view, yes, e := d.Next()
								if e != nil {
									t.Fatal(e)
								}
								if !yes {
									break
								}
								n := len(view.Data) + 8
								if view.Flags&128 != 0 {
									n += 4
								}
								n = (n + 3) &^ 3
								rev = append(append([]byte{}, raw[view.Offset:view.Offset+n]...), rev...)
							}
							raw = rev
						}
						out, code, ok := QuoteInvalid(raw, moneyParameterCause(zone), wire.DefaultLimits(), 16)
						want := (action == 0 || action == 2) && ((zone && v != 0) || (!zone && v != 1))
						if ok != want || ok && (code != 5004 || !bytes.Equal(out, moneyParameterWord(zone, v))) || !ok && (out != nil || code != 0) {
							t.Fatal(code, out)
						}
						clear(raw)
						if ok && !bytes.Equal(out, moneyParameterWord(zone, v)) {
							t.Fatal("borrowed")
						}
					})
				}
			}
		}
	}
}
func TestMoneyParameterControls(t *testing.T) {
	for _, zone := range []bool{false, true} {
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "sentinel", "path", "both", "kind", "number", "action", "trailing", "root-count", "opaque", "client", "mms", "standard"} {
			t.Run(fmt.Sprintf("%t/%s", zone, fault), func(t *testing.T) {
				raw := moneyParameterFixture(0, zone, -1)
				var cause error = moneyParameterCause(zone)
				l := wire.DefaultLimits()
				budget := len(moneyParameterWord(zone, -1))
				switch fault {
				case "budget":
					budget--
				case "negative":
					budget = -1
				case "field":
					l.MaxFieldBytes = 51
				case "frame":
					l.MaxFrameBytes = len(raw) + 19
				case "count":
					l.MaxAVPs = 12
				case "depth":
					l.MaxDepth = 1
				case "repeat":
					l.MaxRepeated = 0
				case "bare":
					cause = validation.ErrField
				case "name":
					cause.(*validation.FieldError).Cause = moneyParameterMetadata{"unknown", validation.ErrField}
				case "sentinel":
					m := cause.(*validation.FieldError).Cause.(moneyParameterMetadata)
					m.cause = validation.ErrUnsupported
					cause.(*validation.FieldError).Cause = m
				case "path":
					cause.(*validation.FieldError).Path[0].VendorID = 999
				case "both":
					c := uint32(439)
					if !zone {
						c = 23
					}
					raw = interfaceEdit(raw, []uint32{c}, func([]byte) []byte { return moneyParameterWord(!zone, -1) })
				case "action":
					raw = moneyParameterFixture(1, zone, -1)
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
						l.MaxAVPs = 13
					}
				case "client":
					raw = interfaceEdit(raw, []uint32{873, 2000, 2018}, func([]byte) []byte { return interfacePart(2018, bytes.Repeat([]byte{0}, 1000)) })
				case "mms":
					raw = interfaceEdit(raw, []uint32{873, 877}, func([]byte) []byte { return interfacePart(877, []byte{1}) })
				case "standard":
					raw = interfaceEdit(raw, []uint32{456, 437, 413}, func([]byte) []byte { return unitPart(413, []byte{1}) })
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "exact" || fault == "opaque" || fault == "client" || fault == "mms" || fault == "standard"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, moneyParameterWord(zone, -1))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, p := range [][]uint32{{416}, {415}, {436}, {439}, {23}, {456}, {456, 437}, {456, 437, 413}, {873}, {873, 2000}, {873, 877}, {873, 2000, 2016}, {873, 2000, 2018}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%t/%v/%s", zone, p, fault), func(t *testing.T) {
					raw := interfaceEdit(moneyParameterFixture(2, zone, -1), p, func(b []byte) []byte {
						switch fault {
						case "missing":
							return nil
						case "duplicate":
							return append(b, b...)
						case "flags":
							b[4] |= 32
						case "vendor":
							if b[4]&128 != 0 {
								binary.BigEndian.PutUint32(b[8:], 999)
							} else {
								data := b[8:]
								b = interfacePart(p[len(p)-1], data)
							}
						case "trailing":
							return append(b, 1)
						}
						return b
					})
					if out, code, ok := QuoteInvalid(raw, moneyParameterCause(zone), wire.DefaultLimits(), 16); ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
		for _, c := range []uint32{455, 458, 1292, 446, 872, 420, 2008, 2010, 2017, 2011, 2015, 2014, 2012} {
			t.Run(fmt.Sprintf("%t/forbidden/%d", zone, c), func(t *testing.T) {
				raw := moneyParameterFixture(0, zone, -1)
				switch c {
				case 455, 458, 1292:
					raw = append(raw, unitPart(c, nil)...)
				case 446, 872:
					raw = interfaceEdit(raw, []uint32{456}, func(b []byte) []byte {
						x := unitPart(c, nil)
						if c == 872 {
							x = interfacePart(c, nil)
						}
						return unitPart(456, append(b[8:], x...))
					})
				case 420:
					raw = interfaceEdit(raw, []uint32{456, 437}, func(b []byte) []byte { return unitPart(437, append(b[8:], unitPart(420, nil)...)) })
				default:
					raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], interfacePart(c, nil)...)) })
				}
				if _, _, ok := QuoteInvalid(raw, moneyParameterCause(zone), wire.DefaultLimits(), 16); ok {
					t.Fatal("forbidden")
				}
			})
		}
		for _, p := range [][]uint32{{416}, {415}, {436}, {439}, {23}, {873, 2000, 2016}} {
			for _, width := range []int{0, 1, 3, 5, 8} {
				t.Run(fmt.Sprintf("%t/%v/width%d", zone, p, width), func(t *testing.T) {
					raw := interfaceEdit(moneyParameterFixture(2, zone, -1), p, func(b []byte) []byte {
						head := 8
						if b[4]&128 != 0 {
							head = 12
						}
						b = append(b[:head:head], make([]byte, width)...)
						wire.Put24(b[5:8], uint32(len(b)))
						for len(b)%4 != 0 {
							b = append(b, 0)
						}
						return b
					})
					if _, _, ok := QuoteInvalid(raw, moneyParameterCause(zone), wire.DefaultLimits(), 16); ok {
						t.Fatal("width")
					}
				})
			}
		}
	}
}
func TestMoneyParameterExactLimits(t *testing.T) {
	for _, zone := range []bool{false, true} {
		l := wire.DefaultLimits()
		l.MaxAVPs = 13
		l.MaxFrameBytes = 172
		l.MaxFieldBytes = 52
		l.MaxDepth = 2
		l.MaxRepeated = 1
		raw := moneyParameterFixture(2, zone, -1)
		out, code, ok := QuoteInvalid(raw, moneyParameterCause(zone), l, len(moneyParameterWord(zone, -1)))
		if len(raw)+20 != 172 || !ok || code != 5004 || !bytes.Equal(out, moneyParameterWord(zone, -1)) {
			t.Fatal(code, out)
		}
	}
}
func fuzzMoneyParameter(t *testing.T, raw []byte, budget uint16, zone bool) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, moneyParameterCause(zone), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	size := 12
	if zone {
		size = 16
	}
	if code != 5004 || len(out) != size || len(out) > int(budget) {
		t.Fatal("bound")
	}
	// Enumerate complete source areas independently; never reuse production quote
	// walkers. The shared decoder counter covers root and both nested branches.
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
	word := func(v wire.View) uint32 {
		if len(v.Data) != 4 {
			t.Fatal("width")
		}
		return binary.BigEndian.Uint32(v.Data)
	}
	d := wire.NewDecoder(raw, l)
	root := scan(d)
	for _, c := range []uint32{455, 458, 1292} {
		absent(root, avp.Key{Code: c})
	}
	if word(get(root, avp.Key{Code: 416}, 64)) != 4 || word(get(root, avp.Key{Code: 415}, 64)) != 0 {
		t.Fatal("sequence")
	}
	action := word(get(root, avp.Key{Code: 436}, 64))
	if action != 0 && action != 2 {
		t.Fatal("action")
	}
	id := get(root, avp.Key{Code: 439}, 64)
	tz := get(root, avp.Key{VendorID: 10425, Code: 23}, 192)
	if zone && (word(id) != 1 || word(tz) == 0) || !zone && (word(id) == 1 || word(tz) != 0) {
		t.Fatal("isolation")
	}
	mscc := get(root, avp.Key{Code: 456}, 64)
	ud := child(d, mscc.Data)
	u := scan(ud)
	absent(u, avp.Key{Code: 446})
	absent(u, avp.Key{VendorID: 10415, Code: 872})
	requested := get(u, avp.Key{Code: 437}, 64)
	r := scan(child(ud, requested.Data))
	absent(r, avp.Key{Code: 420})
	get(r, avp.Key{Code: 413}, 64)
	service := get(root, avp.Key{VendorID: 10415, Code: 873}, 192)
	sd := child(d, service.Data)
	s := scan(sd)
	get(s, avp.Key{VendorID: 10415, Code: 877}, 192)
	sms := get(s, avp.Key{VendorID: 10415, Code: 2000}, 192)
	sm := scan(child(sd, sms.Data))
	for _, c := range []uint32{2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		absent(sm, avp.Key{VendorID: 10415, Code: c})
	}
	if word(get(sm, avp.Key{VendorID: 10415, Code: 2016}, 192)) != 2 {
		t.Fatal("node")
	}
	get(sm, avp.Key{VendorID: 10415, Code: 2018}, 192)
	leaf := id
	if zone {
		leaf = tz
	}
	if !bytes.Equal(out, raw[leaf.Offset:leaf.Offset+size]) {
		t.Fatal("exact leaf")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, moneyParameterCause(zone), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
