// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated MMS SMSKey NUL-value diagnostics.
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

type mmsKeyMetadata struct {
	name  string
	cause error
}

func (m mmsKeyMetadata) Error() string                 { return "key" }
func (m mmsKeyMetadata) Unwrap() error                 { return m.cause }
func (m mmsKeyMetadata) DiameterMMSKeyValueV1() string { return m.name }
func mmsKeyCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 877}, {VendorID: 10415, Code: 1210}}, Cause: mmsKeyMetadata{"combined-mms-key-nul-v1", validation.ErrField}}
}
func mmsKeyFixture(data []byte, mask int) []byte {
	b := interfacePart(886, nil)
	for i, c := range []uint32{1201, 1202, 1209, 1212, 1216} {
		if mask&(1<<i) != 0 {
			v := []byte{1, 2, 3, 4}
			if c == 1201 {
				v = nil
			}
			b = append(b, interfacePart(c, v)...)
		}
	}
	b = append(b, interfacePart(1210, data)...)
	return interfacePart(873, interfacePart(877, b))
}
func mmsKeyExpected(data []byte) []byte {
	return interfacePart(873, interfacePart(877, interfacePart(1210, data)))
}
func TestMMSKeyConsumer(t *testing.T) {
	samples := [][]byte{nil, {}, []byte("key"), {0}, {1, 0}, {0, 1, 2}, {1, 0, 2, 3}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)}
	for i, data := range samples {
		for mask := 0; mask < 32; mask++ {
			t.Run(fmt.Sprintf("%d/%d", i, mask), func(t *testing.T) {
				raw := mmsKeyFixture(data, mask)
				expected := mmsKeyExpected(data)
				out, code, ok := QuoteInvalid(raw, mmsKeyCause(), wire.DefaultLimits(), len(expected))
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
func TestMMSKeyExactLimits(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 255, 1024} {
		for _, fault := range []string{"exact", "budget", "frame", "field", "count", "depth", "negative"} {
			t.Run(fmt.Sprintf("%d/%s", n, fault), func(t *testing.T) {
				data := make([]byte, n)
				raw := mmsKeyFixture(data, 0)
				aligned := (n + 3) &^ 3
				l := wire.DefaultLimits()
				l.MaxFrameBytes = 68 + aligned
				l.MaxFieldBytes = 36 + aligned
				l.MaxAVPs = 4
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
				case "negative":
					budget = -1
				}
				out, code, ok := QuoteInvalid(raw, mmsKeyCause(), l, budget)
				if len(raw)+20 != 68+aligned || ok != (fault == "exact") || ok && (code != 5004 || !bytes.Equal(out, mmsKeyExpected(data))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func TestMMSKeyControls(t *testing.T) {
	for _, fault := range []string{"bare", "name", "sentinel", "path", "short-path", "long-path", "valid", "absent", "extra", "vendor-extra", "trailing", "root-count", "opaque-root", "opaque-address", "opaque-service"} {
		t.Run(fault, func(t *testing.T) {
			raw := mmsKeyFixture([]byte{0}, 0)
			var cause error = mmsKeyCause()
			l := wire.DefaultLimits()
			switch fault {
			case "bare":
				cause = validation.ErrField
			case "name":
				cause.(*validation.FieldError).Cause = mmsKeyMetadata{"unknown", validation.ErrField}
			case "sentinel":
				cause.(*validation.FieldError).Cause = mmsKeyMetadata{"combined-mms-key-nul-v1", validation.ErrUnsupported}
			case "path":
				cause.(*validation.FieldError).Path[2].Code++
			case "short-path":
				cause.(*validation.FieldError).Path = cause.(*validation.FieldError).Path[:2]
			case "long-path":
				cause.(*validation.FieldError).Path = append(cause.(*validation.FieldError).Path, avp.Key{Code: 1})
			case "valid":
				raw = mmsKeyFixture([]byte("key"), 0)
			case "absent":
				raw = interfaceEdit(raw, []uint32{873, 877, 1210}, func([]byte) []byte { return nil })
			case "extra", "vendor-extra":
				raw = interfaceEdit(raw, []uint32{873, 877}, func(b []byte) []byte {
					x := unitPart(9999, []byte{1})
					if fault == "vendor-extra" {
						x = interfacePart(9999, []byte{1})
					}
					return interfacePart(877, append(b[12:], x...))
				})
			case "trailing":
				raw = append(raw, 1)
			case "root-count", "opaque-root":
				raw = append(raw, unitPart(9999, []byte{1})...)
				if fault == "root-count" {
					l.MaxAVPs = 4
				}
			case "opaque-address":
				raw = interfaceEdit(raw, []uint32{873, 877, 886}, func([]byte) []byte { return interfacePart(886, []byte{1}) })
			case "opaque-service":
				raw = interfaceEdit(raw, []uint32{873}, func(b []byte) []byte { return interfacePart(873, append(b[12:], interfacePart(2000, []byte{1})...)) })
			}
			out, code, ok := QuoteInvalid(raw, cause, l, 40)
			want := fault == "opaque-root" || fault == "opaque-address" || fault == "opaque-service"
			if ok != want || ok && (code != 5004 || !bytes.Equal(out, mmsKeyExpected([]byte{0}))) || !ok && (out != nil || code != 0) {
				t.Fatal(code, out)
			}
		})
	}
	for _, path := range [][]uint32{{873}, {873, 877}, {873, 877, 886}, {873, 877, 1210}, {873, 877, 1201}, {873, 877, 1202}, {873, 877, 1209}, {873, 877, 1212}, {873, 877, 1216}} {
		for _, fault := range []string{"missing", "duplicate", "flags", "vendor", "trailing"} {
			t.Run(fmt.Sprintf("%v/%s", path, fault), func(t *testing.T) {
				raw := interfaceEdit(mmsKeyFixture([]byte{0}, 31), path, func(b []byte) []byte {
					switch fault {
					case "missing":
						return nil
					case "duplicate":
						return append(b, b...)
					case "flags":
						b[4] |= 32
					case "vendor":
						binary.BigEndian.PutUint32(b[8:], 999)
					case "trailing":
						return append(b, 1)
					}
					return b
				})
				out, code, ok := QuoteInvalid(raw, mmsKeyCause(), wire.DefaultLimits(), 40)
				c := path[len(path)-1]
				want := fault == "missing" && (c == 1201 || c == 1202 || c == 1209 || c == 1212 || c == 1216)
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, mmsKeyExpected([]byte{0}))) || !ok && (out != nil || code != 0) {
					t.Fatal(code, out)
				}
			})
		}
	}
	for _, c := range []uint32{1202, 1209, 1212, 1216} {
		for _, n := range []int{0, 1, 3, 5, 8} {
			t.Run(fmt.Sprintf("%d/width%d", c, n), func(t *testing.T) {
				raw := interfaceEdit(mmsKeyFixture([]byte{0}, 31), []uint32{873, 877, c}, func([]byte) []byte { return interfacePart(c, make([]byte, n)) })
				if _, _, ok := QuoteInvalid(raw, mmsKeyCause(), wire.DefaultLimits(), 40); ok {
					t.Fatal("width")
				}
			})
		}
	}
}
func TestMMSKeyPaddingAndOrder(t *testing.T) {
	for _, data := range [][]byte{{0}, {1, 0}, {1, 0, 2}, []byte("key")} {
		for _, first := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/%t", data, first), func(t *testing.T) {
				leaf := interfacePart(1210, data)
				for i := 12 + len(data); i < len(leaf); i++ {
					leaf[i] = 0xee
				}
				origin := interfacePart(886, nil)
				b := append(append([]byte{}, origin...), leaf...)
				if first {
					b = append(append([]byte{}, leaf...), origin...)
				}
				raw := interfacePart(873, interfacePart(877, b))
				out, code, ok := QuoteInvalid(raw, mmsKeyCause(), wire.DefaultLimits(), 4096)
				want := bytes.IndexByte(data, 0) >= 0
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, interfacePart(873, interfacePart(877, leaf)))) {
					t.Fatal(code, out)
				}
			})
		}
	}
}
func fuzzMMSKey(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	out, code, ok := QuoteInvalid(raw, mmsKeyCause(), l, int(budget))
	if !ok {
		if out != nil || code != 0 {
			t.Fatal("partial")
		}
		return
	}
	if code != 5004 || len(out) < 40 || len(out) > int(budget) || len(out) > l.MaxFieldBytes || len(out)+28 > l.MaxFrameBytes || len(raw)+20 > l.MaxFrameBytes {
		t.Fatal("bound")
	}
	// Independently enumerate full source areas; never use production walkers.
	d := wire.NewDecoder(raw, l)
	area := raw
	for _, c := range []uint32{873, 877} {
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
			if v.Key == (avp.Key{VendorID: 10415, Code: c}) {
				found = v
				count++
			}
		}
		if count != 1 || found.Flags != 192 {
			t.Fatal("ancestor")
		}
		area = found.Data
		var e error
		d, e = d.Child(area)
		if e != nil {
			t.Fatal(e)
		}
	}
	counts := map[uint32]int{}
	var leaf wire.View
	for {
		v, yes, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !yes {
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
		case 886, 1201:
		case 1202, 1209, 1212, 1216:
			if len(v.Data) != 4 {
				t.Fatal("width")
			}
		case 1210:
			leaf = v
			if bytes.IndexByte(v.Data, 0) < 0 {
				t.Fatal("valid key")
			}
		default:
			t.Fatal("extra")
		}
	}
	if counts[886] != 1 || counts[1210] != 1 {
		t.Fatal("required")
	}
	size := (12 + len(leaf.Data) + 3) &^ 3
	expected := interfacePart(873, interfacePart(877, area[leaf.Offset:leaf.Offset+size]))
	if !bytes.Equal(out, expected) {
		t.Fatal("original")
	}
	copyRaw := append([]byte{}, raw...)
	again, _, yes := QuoteInvalid(copyRaw, mmsKeyCause(), l, int(budget))
	clear(copyRaw)
	if !yes || !bytes.Equal(out, again) {
		t.Fatal("borrowed")
	}
}
