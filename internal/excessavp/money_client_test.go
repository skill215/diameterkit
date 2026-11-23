// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Exact variable client instances, padding, prefix controls and independent fuzz oracle.
// SPDX-AI-Date: 2026-11-23

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

type moneyClientMetadata struct {
	name  string
	cause error
}

func (m moneyClientMetadata) Error() string                      { return "client" }
func (m moneyClientMetadata) Unwrap() error                      { return m.cause }
func (m moneyClientMetadata) DiameterMoneyClientValueV1() string { return m.name }
func moneyClientCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2018}}, Cause: moneyClientMetadata{"combined-money-client-wire-v1", errors.Join(validation.ErrUnsupported, validation.ErrField)}}
}
func moneyClientFixture(action uint32, data []byte) []byte {
	return interfaceEdit(moneyParameterFixture(action, false, 1), []uint32{873, 2000, 2018}, func([]byte) []byte { return interfacePart(2018, data) })
}
func moneyClientExpected(data []byte) []byte {
	return interfacePart(873, interfacePart(2000, interfacePart(2018, data)))
}
func TestMoneyClientConsumer(t *testing.T) {
	samples := [][]byte{{0, 1, 127, 0, 0, 1}}
	for _, n := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 255, 1024} {
		samples = append(samples, bytes.Repeat([]byte{1}, n))
	}
	for i := 0; i < 6; i++ {
		b := []byte{0, 1, 127, 0, 0, 1}
		b[i] ^= 1
		samples = append(samples, b)
	}
	for _, action := range []uint32{0, 1, 2, 3, ^uint32(0)} {
		for i, data := range samples {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%d/%t", action, i, reverse), func(t *testing.T) {
					raw := moneyClientFixture(action, data)
					if reverse {
						d := wire.NewDecoder(raw, wire.DefaultLimits())
						var rev []byte
						for {
							v, yes, e := d.Next()
							if e != nil {
								t.Fatal(e)
							}
							if !yes {
								break
							}
							n := 8 + len(v.Data)
							if v.Flags&128 != 0 {
								n += 4
							}
							n = (n + 3) &^ 3
							rev = append(append([]byte{}, raw[v.Offset:v.Offset+n]...), rev...)
						}
						raw = rev
					}
					want := (action == 0 || action == 2) && !bytes.Equal(data, []byte{0, 1, 127, 0, 0, 1})
					expected := moneyClientExpected(data)
					out, code, ok := QuoteInvalid(raw, moneyClientCause(), wire.DefaultLimits(), len(expected))
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
func TestMoneyClientControls(t *testing.T) {
	for _, zone := range []bool{false, true} {
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "count", "depth", "repeat", "bare", "name", "sentinel", "path", "both", "kind", "number", "action", "trailing", "root-count", "opaque", "client", "mms", "standard"} {
			t.Run(fmt.Sprintf("%t/%s", zone, fault), func(t *testing.T) {
				raw := moneyClientFixture(0, nil)
				var cause error = moneyClientCause()
				l := wire.DefaultLimits()
				budget := 36
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
					l.MaxDepth = 2
				case "repeat":
					l.MaxRepeated = 0
				case "bare":
					cause = validation.ErrField
				case "name":
					cause.(*validation.FieldError).Cause = moneyClientMetadata{"unknown", validation.ErrField}
				case "sentinel":
					m := cause.(*validation.FieldError).Cause.(moneyClientMetadata)
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
					raw = moneyClientFixture(1, nil)
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
					raw = interfaceEdit(raw, []uint32{873, 2000, 2018}, func([]byte) []byte { return interfacePart(2018, []byte{0, 1, 127, 0, 0, 1}) })
				case "mms":
					raw = interfaceEdit(raw, []uint32{873, 877}, func([]byte) []byte { return interfacePart(877, []byte{1}) })
				case "standard":
					raw = interfaceEdit(raw, []uint32{456, 437, 413}, func([]byte) []byte { return unitPart(413, []byte{1}) })
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "exact" || fault == "opaque" || fault == "mms" || fault == "standard"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, moneyClientExpected(nil))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, p := range [][]uint32{{416}, {415}, {436}, {439}, {23}, {456}, {456, 437}, {456, 437, 413}, {873}, {873, 2000}, {873, 877}, {873, 2000, 2016}, {873, 2000, 2018}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%t/%v/%s", zone, p, fault), func(t *testing.T) {
					raw := interfaceEdit(moneyClientFixture(2, nil), p, func(b []byte) []byte {
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
					if out, code, ok := QuoteInvalid(raw, moneyClientCause(), wire.DefaultLimits(), 36); ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
		for _, c := range []uint32{455, 458, 1292, 446, 872, 420, 2008, 2010, 2017, 2011, 2015, 2014, 2012} {
			t.Run(fmt.Sprintf("%t/forbidden/%d", zone, c), func(t *testing.T) {
				raw := moneyClientFixture(0, nil)
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
				if _, _, ok := QuoteInvalid(raw, moneyClientCause(), wire.DefaultLimits(), 36); ok {
					t.Fatal("forbidden")
				}
			})
		}
		for _, p := range [][]uint32{{416}, {415}, {436}, {439}, {23}, {873, 2000, 2016}} {
			for _, width := range []int{0, 1, 3, 5, 8} {
				t.Run(fmt.Sprintf("%t/%v/width%d", zone, p, width), func(t *testing.T) {
					raw := interfaceEdit(moneyClientFixture(2, nil), p, func(b []byte) []byte {
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
					if _, _, ok := QuoteInvalid(raw, moneyClientCause(), wire.DefaultLimits(), 36); ok {
						t.Fatal("width")
					}
				})
			}
		}
	}
}
func TestMoneyClientExactLimits(t *testing.T) {
	for _, n := range []int{0, 1, 3, 4, 5, 6, 7, 8, 17, 255, 1024} {
		for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth"} {
			t.Run(fmt.Sprintf("%d/%s", n, fault), func(t *testing.T) {
				data := bytes.Repeat([]byte{1}, n)
				aligned := (n + 3) &^ 3
				raw := moneyClientFixture(2, data)
				l := wire.DefaultLimits()
				l.MaxAVPs = 13
				l.MaxFrameBytes = 172 + aligned
				l.MaxFieldBytes = 52 + aligned
				l.MaxDepth = 3
				l.MaxRepeated = 1
				budget := 36 + aligned
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
				out, code, ok := QuoteInvalid(raw, moneyClientCause(), l, budget)
				if len(raw)+20 != 172+aligned || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, moneyClientExpected(data))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestMoneyClientPadding(t *testing.T) {
	for _, data := range [][]byte{{1}, {1, 2}, {1, 2, 3}, {0, 1, 127, 0, 0, 1}} {
		t.Run(fmt.Sprint(len(data)), func(t *testing.T) {
			leaf := interfacePart(2018, data)
			for i := 12 + len(data); i < len(leaf); i++ {
				leaf[i] = byte(0xa0 + i)
			}
			raw := interfaceEdit(moneyClientFixture(0, data), []uint32{873, 2000, 2018}, func([]byte) []byte { return leaf })
			out, code, ok := QuoteInvalid(raw, moneyClientCause(), wire.DefaultLimits(), 4096)
			want := !bytes.Equal(data, []byte{0, 1, 127, 0, 0, 1})
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, interfacePart(873, interfacePart(2000, leaf)))) {
				t.Fatal(code, out)
			}
		})
	}
}
func fuzzMoneyClient(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, moneyClientCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) < 36 || len(out) > int(budget) || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes {
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
	if word(id) != 1 || word(tz) != 0 {
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
	leaf := get(sm, avp.Key{VendorID: 10415, Code: 2018}, 192)
	if bytes.Equal(leaf.Data, []byte{0, 1, 127, 0, 0, 1}) {
		t.Fatal("valid value")
	}
	size := (12 + len(leaf.Data) + 3) &^ 3
	expected := interfacePart(873, interfacePart(2000, sms.Data[leaf.Offset:leaf.Offset+size]))
	if !bytes.Equal(out, expected) {
		t.Fatal("exact leaf")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, moneyClientCause(), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
