// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS client wire NUL and equality diagnostics with original bounded quotes.
// SPDX-AI-Date: 2026-11-24

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

type smsClientMetadata struct {
	name  string
	cause error
}

func (m smsClientMetadata) Error() string                    { return "SMS client" }
func (m smsClientMetadata) Unwrap() error                    { return m.cause }
func (m smsClientMetadata) DiameterSMSClientValueV1() string { return m.name }
func smsClientCause(nul bool) *validation.FieldError {
	name := "combined-sms-client-mismatch-v1"
	if nul {
		name = "combined-sms-client-nul-v1"
	}
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2018}}, Cause: smsClientMetadata{name, validation.ErrField}}
}
func smsClientFixture(kind uint32, data []byte, mask int) []byte {
	var root, b []byte
	number := uint32(0)
	if kind == 3 {
		number = 1
	}
	for i, c := range []uint32{416, 415, 455} {
		v := make([]byte, 4)
		binary.BigEndian.PutUint32(v, []uint32{kind, number, 1}[i])
		root = append(root, unitPart(c, v)...)
	}
	root = append(root, unitPart(458, nil)...)
	for _, c := range []uint32{2016, 2001, 2002, 2007, 2009, 2013, 2018, 2010, 2017, 2011, 2015} {
		var v []byte
		switch c {
		case 2016, 2001, 2007, 2011:
			v = make([]byte, 4)
		case 2010, 2017:
			v = []byte("ok")
		case 2018:
			v = data
		case 2013, 2015:
			v = []byte{0}
		}
		b = append(b, interfacePart(c, v)...)
	}
	for i, c := range []uint32{2008, 2014, 2012} {
		if mask&(1<<i) != 0 {
			v := []byte("ok")
			if c == 2014 {
				v = []byte{0}
			}
			if c == 2012 {
				v = make([]byte, 4)
			}
			b = append(b, interfacePart(c, v)...)
		}
	}
	return append(root, interfacePart(873, interfacePart(2000, b))...)
}
func smsClientExpected(data []byte) []byte { return smsDataExpected(2018, data) }
func TestSMSClientConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3, 4} {
		for sample, data := range [][]byte{nil, []byte("ok"), []byte("different"), {0}, {1, 0, 2}, bytes.Repeat([]byte{1}, 1024), bytes.Repeat([]byte{0}, 1024)} {
			for mask := 0; mask < 8; mask++ {
				for _, nul := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/%d/%d/%t", kind, sample, mask, nul), func(t *testing.T) {
						raw := smsClientFixture(kind, data, mask)
						expected := smsClientExpected(data)
						out, code, ok := QuoteInvalid(raw, smsClientCause(nul), wire.DefaultLimits(), len(expected))
						hasNUL := bytes.IndexByte(data, 0) >= 0
						want := hasNUL == nul && (nul || !bytes.Equal(data, []byte("ok")))
						if ok != want || ok && (code != 5004 || !bytes.Equal(out, expected)) || !ok && (out != nil || code != 0) {
							t.Fatal(out, code)
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
}
func TestSMSClientExactLimits(t *testing.T) {
	for _, nul := range []bool{false, true} {
		for _, n := range []int{0, 1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
			if nul && n == 0 {
				continue
			}
			for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth", "repeat"} {
				t.Run(fmt.Sprintf("%t/%d/%s", nul, n, fault), func(t *testing.T) {
					data := bytes.Repeat([]byte{1}, n)
					if nul {
						data[n-1] = 0
					}
					raw := smsClientFixture(4, data, 0)
					a := (n + 3) &^ 3
					l := wire.DefaultLimits()
					l.MaxFrameBytes = len(raw) + 20
					l.MaxFieldBytes = len(raw) - 56
					l.MaxAVPs = 17
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
					out, code, ok := QuoteInvalid(raw, smsClientCause(nul), l, budget)
					if ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, smsClientExpected(data))) || !ok && (out != nil || code != 0) {
						t.Fatal(len(raw), l, code, out)
					}
				})
			}
		}
	}
}
func TestSMSClientControls(t *testing.T) {
	for _, nul := range []bool{false, true} {
		for _, fault := range []string{"bare", "name", "sentinel", "path", "vendor", "negative", "trailing", "valid", "wrong-proof", "counterpart-nul", "node", "type", "sequence", "kind", "indicator", "extra", "opaque-root", "root-count", "opaque-interface", "later-mms", "padding"} {
			t.Run(fmt.Sprintf("%t/%s", nul, fault), func(t *testing.T) {
				data := []byte{1, 2, 3}
				if nul {
					data[1] = 0
				}
				raw := smsClientFixture(4, data, 7)
				var cause error = smsClientCause(nul)
				l := wire.DefaultLimits()
				budget := 4096
				want := false
				edit := func(path []uint32, v []byte) {
					raw = interfaceEdit(raw, path, func(b []byte) []byte {
						c := path[len(path)-1]
						if len(path) == 1 && c != 873 {
							return unitPart(c, v)
						}
						return interfacePart(c, v)
					})
				}
				switch fault {
				case "bare":
					cause = validation.At(avp.Key{VendorID: 10415, Code: 2018}, validation.ErrField)
				case "name":
					cause.(*validation.FieldError).Cause = smsClientMetadata{"unknown", validation.ErrField}
				case "sentinel":
					cause.(*validation.FieldError).Cause = smsClientMetadata{cause.(*validation.FieldError).Cause.(smsClientMetadata).name, validation.ErrUnsupported}
				case "path":
					cause.(*validation.FieldError).Path[1].Code = 877
				case "vendor":
					cause.(*validation.FieldError).Path[2].VendorID = 0
				case "negative":
					budget = -1
				case "trailing":
					raw = append(raw, 1)
				case "valid":
					edit([]uint32{873, 2000, 2018}, []byte("ok"))
				case "wrong-proof":
					cause = smsClientCause(!nul)
				case "counterpart-nul":
					edit([]uint32{873, 2000, 2017}, []byte{0})
				case "node":
					edit([]uint32{873, 2000, 2016}, []byte{0, 0, 0, 9})
				case "type":
					edit([]uint32{873, 2000, 2007}, []byte{0, 0, 0, 9})
				case "sequence":
					edit([]uint32{415}, []byte{0, 0, 0, 9})
				case "kind":
					edit([]uint32{416}, []byte{0, 0, 0, 2})
				case "indicator":
					edit([]uint32{455}, []byte{0, 0, 0, 0})
				case "extra":
					raw = interfaceEdit(raw, []uint32{873, 2000}, func(b []byte) []byte { return interfacePart(2000, append(b[12:], interfacePart(9999, nil)...)) })
				case "opaque-root", "root-count":
					raw = append(raw, unitPart(9999, []byte{1})...)
					want = fault == "opaque-root"
					if !want {
						l.MaxAVPs = 20
					}
				case "opaque-interface":
					edit([]uint32{873, 2000, 2009}, []byte{1})
					want = true
				case "later-mms":
					raw = interfaceEdit(raw, []uint32{873}, func(b []byte) []byte { return interfacePart(873, append(b[12:], interfacePart(877, nil)...)) })
					want = true
				case "padding":
					raw = interfaceEdit(raw, []uint32{873, 2000, 2018}, func(b []byte) []byte { b[len(b)-1] = 0xff; return b })
					want = true
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				if ok != want || ok && (code != 5004 || len(out) != 40) || !ok && (out != nil || code != 0) {
					t.Fatal(fault, code, out)
				}
				if fault == "padding" && out[len(out)-1] != 0xff {
					t.Fatal("padding lost")
				}
			})
		}
	}
}
func TestSMSClientStructure(t *testing.T) {
	paths := [][]uint32{{416}, {415}, {455}, {458}, {873}, {873, 2000}}
	for _, c := range []uint32{2016, 2001, 2002, 2007, 2009, 2013, 2018, 2010, 2017, 2011, 2015, 2008, 2014, 2012} {
		paths = append(paths, []uint32{873, 2000, c})
	}
	for index, path := range paths {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%d/%s", index, fault), func(t *testing.T) {
				raw := smsClientFixture(4, []byte{0}, 7)
				raw = interfaceEdit(raw, path, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					case "flags":
						b[4] ^= 64
					case "vendor":
						if b[4]&128 != 0 {
							binary.BigEndian.PutUint32(b[8:], 1)
						} else {
							b[3] ^= 1
						}
					case "trailing":
						return append(b, 1)
					}
					return b
				})
				out, code, ok := QuoteInvalid(raw, smsClientCause(true), wire.DefaultLimits(), 4096)
				optional := len(path) == 3 && (path[2] == 2008 || path[2] == 2014 || path[2] == 2012)
				want := fault == "missing" && optional
				if ok != want || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
	for _, c := range []uint32{416, 415, 455, 2016, 2001, 2007, 2011, 2012} {
		for _, n := range []int{0, 1, 3, 5, 8} {
			t.Run(fmt.Sprintf("width/%d/%d", c, n), func(t *testing.T) {
				path := []uint32{c}
				if c > 2000 {
					path = []uint32{873, 2000, c}
				}
				raw := smsClientFixture(4, []byte{0}, 7)
				raw = interfaceEdit(raw, path, func(b []byte) []byte {
					if c > 2000 {
						return interfacePart(c, make([]byte, n))
					}
					return unitPart(c, make([]byte, n))
				})
				if _, _, ok := QuoteInvalid(raw, smsClientCause(true), wire.DefaultLimits(), 4096); ok {
					t.Fatal("width accepted")
				}
			})
		}
	}
}

func fuzzSMSClient(t *testing.T, raw []byte, budget uint16, nul bool) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, smsClientCause(nul), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) < 36 || len(out) > int(budget) || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes {
		t.Fatal("bounds")
	}
	d := wire.NewDecoder(raw, l)
	root := map[uint32]wire.View{}
	counts := map[uint32]int{}
	for {
		v, more, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		if v.Key == (avp.Key{Code: 416}) || v.Key == (avp.Key{Code: 415}) || v.Key == (avp.Key{Code: 455}) || v.Key == (avp.Key{Code: 458}) || v.Key == (avp.Key{VendorID: 10415, Code: 873}) {
			counts[v.Key.Code]++
			root[v.Key.Code] = v
		}
	}
	for _, c := range []uint32{416, 415, 455, 458, 873} {
		if counts[c] != 1 {
			t.Fatal("root count")
		}
		flag := byte(64)
		if c == 873 {
			flag = 192
		}
		if root[c].Flags != flag {
			t.Fatal("root flag")
		}
	}
	for _, c := range []uint32{416, 415, 455} {
		if len(root[c].Data) != 4 {
			t.Fatal("root width")
		}
	}
	kind, number := binary.BigEndian.Uint32(root[416].Data), binary.BigEndian.Uint32(root[415].Data)
	if !((kind == 1 || kind == 4) && number == 0 || kind == 3 && number == 1) || binary.BigEndian.Uint32(root[455].Data) != 1 {
		t.Fatal("root values")
	}
	service, e := d.Child(root[873].Data)
	if e != nil {
		t.Fatal(e)
	}
	var group wire.View
	n := 0
	for {
		v, more, e := service.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		if v.Key == (avp.Key{VendorID: 10415, Code: 2000}) {
			group = v
			n++
		}
	}
	if n != 1 || group.Flags != 192 {
		t.Fatal("group")
	}
	sms, e := service.Child(group.Data)
	if e != nil {
		t.Fatal(e)
	}
	fields := map[uint32]wire.View{}
	for {
		v, more, e := sms.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		if v.Flags != 192 || v.Key.VendorID != 10415 {
			t.Fatal("SMS flags")
		}
		if _, yes := fields[v.Key.Code]; yes {
			t.Fatal("duplicate")
		}
		fields[v.Key.Code] = v
		switch v.Key.Code {
		case 2016, 2001, 2007, 2011, 2012:
			if len(v.Data) != 4 {
				t.Fatal("width")
			}
		case 2008, 2010, 2017:
			if bytes.IndexByte(v.Data, 0) >= 0 {
				t.Fatal("address NUL")
			}
		case 2002, 2009, 2013, 2018, 2015, 2014:
		default:
			t.Fatal("extra")
		}
	}
	for _, c := range []uint32{2016, 2001, 2002, 2007, 2009, 2013, 2018, 2010, 2017, 2011, 2015} {
		if _, yes := fields[c]; !yes {
			t.Fatal("missing")
		}
	}
	if binary.BigEndian.Uint32(fields[2016].Data) != 0 || binary.BigEndian.Uint32(fields[2007].Data) != 0 {
		t.Fatal("node/type")
	}
	leaf := fields[2018]
	has := bytes.IndexByte(leaf.Data, 0) >= 0
	if has != nul || !nul && bytes.Equal(leaf.Data, fields[2017].Data) {
		t.Fatal("condition")
	}
	quote := out
	for _, c := range []uint32{873, 2000} {
		q := wire.NewDecoder(quote, l)
		v, more, e := q.Next()
		if e != nil || !more || q.Remaining() != 0 || v.Key != (avp.Key{VendorID: 10415, Code: c}) || v.Flags != 192 {
			t.Fatal("quote group")
		}
		quote = v.Data
	}
	size := (12 + len(leaf.Data) + 3) &^ 3
	if !bytes.Equal(quote, group.Data[leaf.Offset:leaf.Offset+size]) {
		t.Fatal("exact leaf")
	}
	clone := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(clone, smsClientCause(nul), l, int(budget))
	clear(clone)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}

func TestSMSClientEmptyEqual(t *testing.T) {
	raw := smsClientFixture(4, nil, 0)
	raw = interfaceEdit(raw, []uint32{873, 2000, 2017}, func(b []byte) []byte { return interfacePart(2017, nil) })
	for _, nul := range []bool{false, true} {
		if _, _, ok := QuoteInvalid(raw, smsClientCause(nul), wire.DefaultLimits(), 4096); ok {
			t.Fatal("empty equality")
		}
	}
}
