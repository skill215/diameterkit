// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded SMS address value NUL diagnostics with bounded original quotes.
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

type smsDataMetadata struct {
	name  string
	cause error
}

func (m smsDataMetadata) Error() string                  { return "SMS address" }
func (m smsDataMetadata) Unwrap() error                  { return m.cause }
func (m smsDataMetadata) DiameterSMSDataValueV1() string { return m.name }
func smsDataCause(code uint32) *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: code}}, Cause: smsDataMetadata{"combined-sms-data-nul-v1", validation.ErrField}}
}
func smsDataFixture(code uint32, data []byte, mask int) []byte {
	var b []byte
	for _, c := range []uint32{2016, 2001, 2002, 2007, 2009, 2013} {
		var v []byte
		if c == 2016 || c == 2001 || c == 2007 {
			v = []byte{0, 0, 0, 9}
		}
		b = append(b, interfacePart(c, v)...)
	}
	index := 0
	for _, c := range []uint32{2018, 2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		if c == code {
			continue
		}
		if mask&(1<<index) != 0 {
			v := []byte("ok")
			if c == 2011 || c == 2012 {
				v = []byte{1, 2, 3, 4}
			}
			if c == 2018 || c == 2015 || c == 2014 {
				v = []byte{0}
			}
			b = append(b, interfacePart(c, v)...)
		}
		index++
	}
	b = append(b, interfacePart(code, data)...)
	return interfacePart(873, interfacePart(2000, b))
}
func smsDataExpected(code uint32, data []byte) []byte {
	return interfacePart(873, interfacePart(2000, interfacePart(code, data)))
}
func TestSMSDataConsumer(t *testing.T) {
	for _, code := range []uint32{2008, 2010, 2017} {
		for sample, data := range [][]byte{nil, []byte("ok"), {0}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)} {
			for mask := 0; mask < 128; mask++ {
				t.Run(fmt.Sprintf("%d/%d/%d", code, sample, mask), func(t *testing.T) {
					raw := smsDataFixture(code, data, mask)
					expected := smsDataExpected(code, data)
					out, result, ok := QuoteInvalid(raw, smsDataCause(code), wire.DefaultLimits(), len(expected))
					want := bytes.IndexByte(data, 0) >= 0
					if ok != want || ok && (result != 5004 || !bytes.Equal(out, expected)) || !ok && (out != nil || result != 0) {
						t.Fatal(result, out)
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
func TestSMSDataExactLimits(t *testing.T) {
	for _, code := range []uint32{2008, 2010, 2017} {
		for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
			for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth", "repeat"} {
				t.Run(fmt.Sprintf("%d/%d/%s", code, n, fault), func(t *testing.T) {
					data := make([]byte, n)
					raw := smsDataFixture(code, data, 0)
					a := (n + 3) &^ 3
					l := wire.DefaultLimits()
					l.MaxFrameBytes = 140 + a
					l.MaxFieldBytes = 108 + a
					l.MaxAVPs = 9
					l.MaxDepth = 3
					l.MaxRepeated = 1
					budget := 36 + a
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
					case "repeat":
						l.MaxRepeated--
					}
					out, result, ok := QuoteInvalid(raw, smsDataCause(code), l, budget)
					if len(raw)+20 != 140+a || ok != (fault == "exact") || ok && (result != 5004 || !bytes.Equal(out, smsDataExpected(code, data))) || !ok && (out != nil || result != 0) {
						t.Fatal(result, out)
					}
				})
			}
		}
	}
}
func TestSMSDataControls(t *testing.T) {
	for _, code := range []uint32{2008, 2010, 2017} {
		for _, fault := range []string{"bare", "name", "path", "vendor", "sentinel", "negative", "trailing", "extra", "mandatory-extra", "other-nul", "all-nul", "valid", "opaque-root", "root-count"} {
			t.Run(fmt.Sprintf("%d/%s", code, fault), func(t *testing.T) {
				raw := smsDataFixture(code, []byte{0}, 127)
				var cause error = smsDataCause(code)
				l := wire.DefaultLimits()
				budget := 4096
				switch fault {
				case "bare":
					cause = validation.At(avp.Key{VendorID: 10415, Code: code}, validation.ErrField)
				case "name":
					cause.(*validation.FieldError).Cause = smsDataMetadata{"unknown", validation.ErrField}
				case "path":
					cause.(*validation.FieldError).Path[2].Code = 2018
				case "vendor":
					cause.(*validation.FieldError).Path[2].VendorID = 0
				case "sentinel":
					cause.(*validation.FieldError).Cause = smsDataMetadata{"combined-sms-data-nul-v1", validation.ErrUnsupported}
				case "negative":
					budget = -1
				case "trailing":
					raw = append(raw, 1)
				case "extra", "mandatory-extra":
					part := unitPart(9999, nil)
					if fault == "extra" {
						part[4] = 0
					}
					raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], part...)) })
				case "other-nul", "all-nul":
					for _, c := range []uint32{2008, 2010, 2017} {
						if c != code {
							raw = interfaceEdit(raw, []uint32{873, 2000, c}, func([]byte) []byte { return interfacePart(c, []byte{0}) })
							if fault == "other-nul" {
								break
							}
						}
					}
				case "valid":
					raw = smsDataFixture(code, []byte("ok"), 127)
				case "opaque-root":
					raw = append(raw, unitPart(9999, []byte{1})...)
				case "root-count":
					raw = smsDataFixture(code, []byte{0}, 0)
					raw = append(raw, unitPart(9999, nil)...)
					l.MaxAVPs = 9
					l.MaxRepeated = 9
				}
				out, result, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "opaque-root"
				if ok != want || ok && (result != 5004 || !bytes.Equal(out, smsDataExpected(code, []byte{0}))) || !ok && (out != nil || result != 0) {
					t.Fatal(result, out)
				}
			})
		}
		for _, path := range [][]uint32{{873}, {873, 2000}, {873, 2000, code}, {873, 2000, 2016}, {873, 2000, 2001}, {873, 2000, 2002}, {873, 2000, 2007}, {873, 2000, 2009}, {873, 2000, 2013}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%d/%v/%s", code, path, fault), func(t *testing.T) {
					raw := interfaceEdit(smsDataFixture(code, []byte{0}, 0), path, func(b []byte) []byte {
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
							if len(path) < 3 {
								return interfacePart(path[len(path)-1], append(b[12:], 1))
							}
							return append(b, 1)
						}
						return b
					})
					out, result, ok := QuoteInvalid(raw, smsDataCause(code), wire.DefaultLimits(), 4096)
					if ok || out != nil || result != 0 {
						t.Fatal(result, out)
					}
				})
			}
		}
		for _, c := range []uint32{2016, 2001, 2007, 2011, 2012} {
			for _, n := range []int{0, 1, 3, 5, 8} {
				raw := interfaceEdit(smsDataFixture(code, []byte{0}, 127), []uint32{873, 2000, c}, func([]byte) []byte { return interfacePart(c, make([]byte, n)) })
				if _, _, ok := QuoteInvalid(raw, smsDataCause(code), wire.DefaultLimits(), 4096); ok {
					t.Fatal("width", c, n)
				}
			}
		}
	}
}
func TestSMSDataPaddingAndOrder(t *testing.T) {
	for _, code := range []uint32{2008, 2010, 2017} {
		for _, data := range [][]byte{{0}, {1, 0}, {1, 0, 2}, []byte("ok")} {
			for _, reverse := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/%x/%t", code, data, reverse), func(t *testing.T) {
					leaf := interfacePart(code, data)
					for i := 12 + len(data); i < len(leaf); i++ {
						leaf[i] = 0xee
					}
					raw := interfaceEdit(smsDataFixture(code, data, 127), []uint32{873, 2000, code}, func([]byte) []byte { return leaf })
					if reverse {
						raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte {
							d := wire.NewDecoder(b[12:], wire.DefaultLimits())
							var parts [][]byte
							for {
								v, more, e := d.Next()
								if e != nil {
									t.Fatal(e)
								}
								if !more {
									break
								}
								size := (12 + len(v.Data) + 3) &^ 3
								parts = append(parts, b[12+v.Offset:12+v.Offset+size])
							}
							var body []byte
							for i := len(parts) - 1; i >= 0; i-- {
								body = append(body, parts[i]...)
							}
							return interfacePart(2000, body)
						})
					}
					out, result, ok := QuoteInvalid(raw, smsDataCause(code), wire.DefaultLimits(), 4096)
					want := bytes.IndexByte(data, 0) >= 0
					if ok != want || ok && (result != 5004 || !bytes.Equal(out, interfacePart(873, interfacePart(2000, leaf)))) {
						t.Fatal(result, out)
					}
				})
			}
		}
	}
}
func fuzzSMSData(t *testing.T, raw []byte, budget uint16, code uint32) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, result, ok := QuoteInvalid(raw, smsDataCause(code), l, int(budget))
	if !ok {
		if out != nil || result != 0 {
			t.Fatal("partial")
		}
		return
	}
	if result != 5004 || len(out) < 40 || len(out) > int(budget) || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes {
		t.Fatal("bounds")
	}
	source, quote := raw, out
	d := wire.NewDecoder(raw, l)
	for _, c := range []uint32{873, 2000} {
		var found wire.View
		n := 0
		for {
			v, more, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !more {
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
		v, more, e := q.Next()
		if e != nil || !more || q.Remaining() != 0 || v.Key != found.Key || v.Flags != 192 {
			t.Fatal("quote ancestor")
		}
		source, quote = found.Data, v.Data
		d, e = d.Child(source)
		if e != nil {
			t.Fatal(e)
		}
	}
	counts := map[uint32]int{}
	bad := 0
	var leaf wire.View
	for {
		v, more, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		if v.Key.VendorID != 10415 || v.Flags != 192 {
			t.Fatal("flags/vendor")
		}
		counts[v.Key.Code]++
		if counts[v.Key.Code] != 1 {
			t.Fatal("duplicate")
		}
		switch v.Key.Code {
		case 2016, 2001, 2007, 2011, 2012:
			if len(v.Data) != 4 {
				t.Fatal("width")
			}
		case 2008, 2010, 2017:
			if bytes.IndexByte(v.Data, 0) >= 0 {
				bad++
			}
		case 2002, 2009, 2013, 2018, 2015, 2014:
		default:
			t.Fatal("extra")
		}
		if v.Key.Code == code {
			leaf = v
		}
	}
	for _, c := range []uint32{2016, 2001, 2002, 2007, 2009, 2013} {
		if counts[c] != 1 {
			t.Fatal("required")
		}
	}
	size := (12 + len(leaf.Data) + 3) &^ 3
	if bad != 1 || leaf.Key.Code != code || bytes.IndexByte(leaf.Data, 0) < 0 || !bytes.Equal(quote, source[leaf.Offset:leaf.Offset+size]) {
		t.Fatal("exact isolated leaf")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, smsDataCause(code), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
