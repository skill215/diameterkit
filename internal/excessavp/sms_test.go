// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded SMS value instance, budget and fuzz acceptance.
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

type smsMetadata struct {
	rule  string
	cause error
}

func (m *smsMetadata) Error() string              { return "sms value" }
func (m *smsMetadata) Unwrap() error              { return m.cause }
func (m *smsMetadata) DiameterSMSValueV1() string { return m.rule }
func smsCause(node bool) *validation.FieldError {
	code := uint32(2007)
	rule := "combined-sm-message-type-v1"
	if node {
		code = 2016
		rule = "combined-sms-node-v1"
	}
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: code}}, Cause: &smsMetadata{rule, errors.Join(validation.ErrField, validation.ErrUnsupported)}}
}
func smsWord(code uint32, value int32) []byte {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(value))
	return interfacePart(code, data[:])
}
func smsFixture(kind uint32, node bool, value int32) []byte {
	a, b := int32(0), int32(0)
	if node {
		a = value
	} else {
		b = value
	}
	raw := append(actionWord(416, kind), actionWord(455, 1)...)
	raw = append(raw, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
	return append(raw, interfacePart(873, interfacePart(2000, append(smsWord(2016, a), smsWord(2007, b)...)))...)
}
func smsExpected(node bool, value int32) []byte {
	code := uint32(2007)
	if node {
		code = 2016
	}
	return interfacePart(873, interfacePart(2000, smsWord(code, value)))
}
func TestSMSValueConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3, 4} {
		for _, node := range []bool{false, true} {
			for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
				t.Run(fmt.Sprintf("%d/%t/%d", kind, node, value), func(t *testing.T) {
					raw := smsFixture(kind, node, value)
					out, code, ok := QuoteInvalid(raw, smsCause(node), wire.DefaultLimits(), 40)
					if ok != (value != 0) || ok && (code != 5004 || !bytes.Equal(out, smsExpected(node, value))) || !ok && (out != nil || code != 0) {
						t.Fatal(code, out)
					}
					clear(raw)
					if ok && !bytes.Equal(out, smsExpected(node, value)) {
						t.Fatal("borrowed")
					}
				})
			}
		}
	}
}
func TestSMSValueControls(t *testing.T) {
	for _, node := range []bool{false, true} {
		selected, other := uint32(2007), uint32(2016)
		if node {
			selected, other = other, selected
		}
		for _, fault := range []string{"exact", "budget", "negative", "field", "frame", "source-count", "output-count", "depth", "repeat", "bare", "sentinel", "name", "path", "kind", "indicator", "equipment", "both", "valid", "counterpart-width", "selected-width", "trailing", "siblings"} {
			t.Run(fmt.Sprintf("%t/%s", node, fault), func(t *testing.T) {
				raw := smsFixture(4, node, -1)
				cause := smsCause(node)
				l := wire.DefaultLimits()
				budget := 40
				edit := func(path []uint32, f func([]byte) []byte) { raw = interfaceEdit(raw, path, f) }
				switch fault {
				case "budget":
					budget = 39
				case "negative":
					budget = -1
				case "field":
					l.MaxFieldBytes = 39
				case "frame":
					l.MaxFrameBytes = len(raw) + 19
				case "source-count":
					l.MaxAVPs = 6
				case "output-count":
					l.MaxAVPs = 3
				case "depth":
					l.MaxDepth = 2
				case "repeat":
					l.MaxRepeated = 0
				case "bare":
					cause.Cause = validation.ErrField
				case "sentinel":
					cause.Cause.(*smsMetadata).cause = validation.ErrUnsupported
				case "name":
					cause.Cause.(*smsMetadata).rule = "unreviewed"
				case "path":
					cause.Path[2].Code = other
				case "kind":
					edit([]uint32{416}, func([]byte) []byte { return actionWord(416, 2) })
				case "indicator":
					edit([]uint32{455}, func([]byte) []byte { return actionWord(455, 0) })
				case "equipment":
					edit([]uint32{458}, func([]byte) []byte { return nil })
				case "both":
					edit([]uint32{873, 2000, other}, func([]byte) []byte { return smsWord(other, -1) })
				case "valid":
					edit([]uint32{873, 2000, selected}, func([]byte) []byte { return smsWord(selected, 0) })
				case "counterpart-width":
					edit([]uint32{873, 2000, other}, func([]byte) []byte { return interfacePart(other, []byte{0}) })
				case "selected-width":
					edit([]uint32{873, 2000, selected}, func([]byte) []byte { return interfacePart(selected, []byte{1}) })
				case "trailing":
					raw = append(raw, 1)
				case "siblings":
					edit([]uint32{873, 2000}, func(b []byte) []byte {
						return interfacePart(2000, append(append(append([]byte{}, b[28:44]...), interfacePart(9999, []byte{7, 8, 9})...), b[12:28]...))
					})
				}
				out, code, ok := QuoteInvalid(raw, cause, l, budget)
				want := fault == "exact" || fault == "siblings"
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, smsExpected(node, -1))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
		for _, path := range [][]uint32{{416}, {455}, {458}, {873}, {873, 2000}, {873, 2000, selected}, {873, 2000, other}} {
			for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
				t.Run(fmt.Sprintf("%t/%v/%s", node, path, fault), func(t *testing.T) {
					raw := interfaceEdit(smsFixture(1, node, -1), path, func(b []byte) []byte {
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
							if path[len(path)-1] == 873 || path[len(path)-1] == 2000 {
								return interfacePart(path[len(path)-1], append(b[12:], 1))
							}
							return append(b, 1)
						}
						return b
					})
					if out, code, ok := QuoteInvalid(raw, smsCause(node), wire.DefaultLimits(), 40); ok || out != nil || code != 0 {
						t.Fatal(code, out)
					}
				})
			}
		}
	}
}
func TestSMSValueExactLimits(t *testing.T) {
	for _, node := range []bool{false, true} {
		raw := smsFixture(3, node, -1)
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 108
		l.MaxFieldBytes = 44
		l.MaxAVPs = 7
		l.MaxDepth = 3
		l.MaxRepeated = 1
		if len(raw) != 88 {
			t.Fatal(len(raw))
		}
		out, code, ok := QuoteInvalid(raw, smsCause(node), l, 40)
		if !ok || code != 5004 || !bytes.Equal(out, smsExpected(node, -1)) {
			t.Fatal(code, out)
		}
	}
}
func fuzzSMSValue(t *testing.T, raw []byte, budget uint16, node bool) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, smsCause(node), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial quote")
		}
		return
	}
	if code != 5004 || len(out) != 40 || len(out) > int(budget) {
		t.Fatal("bounds")
	}
	selected := uint32(2007)
	if node {
		selected = 2016
	}
	source := raw
	quote := out
	for _, key := range []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: selected}} {
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
		if e != nil || !yes || q.Remaining() != 0 || v.Key != key || v.Flags != 192 || count != 1 {
			t.Fatal("instance grammar")
		}
		if key.Code == selected {
			if len(found.Data) != 4 || binary.BigEndian.Uint32(found.Data) == 0 || !bytes.Equal(quote, source[found.Offset:found.Offset+16]) {
				t.Fatal("changed leaf")
			}
		}
		source = found.Data
		quote = v.Data
	}
	frozen := append([]byte{}, out...)
	copyOfSource := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyOfSource, smsCause(node), l, int(budget))
	clear(copyOfSource)
	if !yes || !bytes.Equal(frozen, again) {
		t.Fatal("borrowed")
	}
}
