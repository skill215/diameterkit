// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded combined money MMS exclusion and acceptance.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"bytes"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type moneyMMSProof struct{}

func (moneyMMSProof) Error() string                   { return "excluded money key" }
func (moneyMMSProof) Unwrap() error                   { return validation.ErrUnsupported }
func (moneyMMSProof) DiameterMoneyMMSKeyExclusionV1() {}
func moneyMMSCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 877}, {VendorID: 10415, Code: 1210}}, Cause: moneyMMSProof{}}
}
func moneyMMSFixture(action uint32, key []byte) []byte {
	raw := moneyClientFixture(action, []byte{0, 1, 127, 0, 0, 1})
	address := interfacePart(886, append(smsWord(899, 1), interfacePart(897, []byte("from"))...))
	return interfaceEdit(raw, []uint32{873, 877}, func([]byte) []byte { return interfacePart(877, append(address, interfacePart(1210, key)...)) })
}
func TestMoneyMMSQuoteLimits(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("abc"), bytes.Repeat([]byte{'x'}, 257)} {
		raw := moneyMMSFixture(0, data)
		want := interfacePart(873, interfacePart(877, interfacePart(1210, data)))
		limits := wire.DefaultLimits()
		// The Service-Information payload is the largest source field.
		d := wire.NewDecoder(raw, limits)
		for {
			v, yes, err := d.Next()
			if err != nil {
				t.Fatal(err)
			}
			if !yes {
				break
			}
			if v.Key.Code == 873 {
				limits.MaxFieldBytes = len(v.Data)
			}
		}
		limits.MaxAVPs = 17
		limits.MaxRepeated = 1
		limits.MaxDepth = 3
		limits.MaxFrameBytes = len(raw) + 20
		for _, short := range []string{"none", "count", "depth", "field", "frame", "quote"} {
			l := limits
			budget := len(want)
			switch short {
			case "count":
				l.MaxAVPs--
			case "depth":
				l.MaxDepth--
			case "field":
				l.MaxFieldBytes--
			case "frame":
				l.MaxFrameBytes--
			case "quote":
				budget--
			}
			q, c, ok := QuoteInvalid(raw, moneyMMSCause(), l, budget)
			if short == "none" {
				if !ok || c != 5008 || !bytes.Equal(q, want) {
					t.Fatal(short, c, q)
				}
			} else if ok || q != nil || c != 0 {
				t.Fatal(short, c, q)
			}
		}
	}
}
func TestMoneyMMSEvidenceMismatch(t *testing.T) {
	raw := moneyMMSFixture(2, []byte("abc"))
	for _, tc := range []struct {
		path        []uint32
		replacement []byte
	}{
		{[]uint32{873, 877, 886, 899}, smsWord(899, 0)},
		{[]uint32{873, 877, 886, 897}, interfacePart(897, []byte{0})},
		{[]uint32{873, 877, 1210}, interfacePart(1210, []byte{0})},
		{[]uint32{873, 877, 1210}, append(interfacePart(1210, nil), interfacePart(1210, nil)...)},
		{[]uint32{873, 877, 886}, nil},
		{[]uint32{873, 877, 1210}, append(interfacePart(1210, nil), 0, 1, 2)},
		{[]uint32{873, 2000, 2018}, interfacePart(2018, []byte("wrong"))},
		{[]uint32{436}, actionWord(436, 1)},
	} {
		altered := interfaceEdit(raw, tc.path, func([]byte) []byte { return tc.replacement })
		if q, c, ok := QuoteInvalid(altered, moneyMMSCause(), wire.DefaultLimits(), 4096); ok || q != nil || c != 0 {
			t.Fatal(tc.path, c, q)
		}
	}
	f := moneyMMSCause()
	f.Path[1].Code = 2000
	if _, _, ok := QuoteInvalid(raw, f, wire.DefaultLimits(), 4096); ok {
		t.Fatal("wrong path")
	}
}

func fuzzExcludedMoneyMMSKey(t *testing.T, raw []byte, budget uint16) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	q, c, ok := QuoteInvalid(raw, moneyMMSCause(), l, int(budget))
	if !ok {
		if q != nil || c != 0 {
			t.Fatal("partial quote")
		}
		return
	}
	if c != 5008 || len(q) > int(budget) || len(q) > l.MaxFieldBytes || len(q)+28 > l.MaxFrameBytes {
		t.Fatal("bounds", c, len(q))
	}
	area := raw
	for _, code := range []uint32{873, 877, 1210} {
		d := wire.NewDecoder(area, l)
		var child []byte
		found := false
		for {
			v, yes, err := d.Next()
			if err != nil {
				t.Fatal(err)
			}
			if !yes {
				break
			}
			if v.Key != (avp.Key{VendorID: 10415, Code: code}) {
				continue
			}
			if found || v.Flags != 192 {
				t.Fatal("instance", code)
			}
			found = true
			child = v.Data
			if code == 1210 {
				end := v.Offset + ((12 + len(v.Data) + 3) &^ 3)
				want := interfacePart(873, interfacePart(877, area[v.Offset:end]))
				if bytes.IndexByte(v.Data, 0) >= 0 || !bytes.Equal(want, q) {
					t.Fatal("original key")
				}
			}
		}
		if !found {
			t.Fatal("missing", code)
		}
		area = child
	}
}
