// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group exclusions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type ecurUnitProof struct{}

func (ecurUnitProof) Error() string               { return "excluded unit" }
func (ecurUnitProof) Unwrap() error               { return validation.ErrUnsupported }
func (ecurUnitProof) DiameterExcludedECURUnitV1() {}
func ecurUnitCause(code uint32) error {
	return &validation.FieldError{Path: []avp.Key{{Code: 456}, {Code: code}}, Cause: ecurUnitProof{}}
}
func ecurUnitArea(kind uint32) []byte {
	// Start with the normal SMS group by removing the synthetic originator field.
	top := interfaceEdit(ecurOriginatorArea(kind, nil, 3, 2), []uint32{873, 2000, 2008}, func([]byte) []byte { return nil })
	groups := append(unitPart(446, unitLeaf(false, 17)), unitPart(437, unitLeaf(false, 1))...)
	if kind == 3 {
		groups = append(groups, unitLeaf(true, 2)...)
	}
	return append(top, unitPart(456, groups)...)
}
func TestECURUnitConsumer(t *testing.T) {
	for _, kind := range []uint32{1, 3} {
		code, other := uint32(446), uint32(437)
		if kind == 3 {
			code, other = other, code
		}
		for _, variant := range []string{"exact", "reverse", "short", "count", "depth", "field", "frame", "width", "flags", "duplicate", "sibling", "money", "missing-peer", "wrong-proof", "action", "status", "trailing", "sequence", "reason"} {
			t.Run(fmt.Sprintf("%d/%s", kind, variant), func(t *testing.T) {
				raw := ecurUnitArea(kind)
				l := wire.DefaultLimits()
				budget := 32
				cause := ecurUnitCause(code)
				path := []uint32{456, code}
				switch variant {
				case "reverse":
					raw = interfaceEdit(raw, []uint32{456}, func(b []byte) []byte {
						data := append([]byte(nil), b[8:32]...)
						copy(b[8:], b[32:56])
						copy(b[32:], data)
						return b
					})
				case "short":
					budget--
				case "count":
					l.MaxAVPs = 10
				case "depth":
					l.MaxDepth = 2
				case "field":
					l.MaxFieldBytes = 31
				case "frame":
					l.MaxFrameBytes = len(raw) + 19
				case "width":
					raw = interfaceEdit(raw, append(path, 417), func([]byte) []byte { return unitPart(417, make([]byte, 4)) })
				case "flags":
					raw = interfaceEdit(raw, path, func(b []byte) []byte { b[4] = 0; return b })
				case "duplicate":
					raw = interfaceEdit(raw, path, func(b []byte) []byte { return append(b, b...) })
				case "sibling":
					raw = interfaceEdit(raw, path, func(b []byte) []byte { return append(b, unitPart(777, nil)...) })
				case "money":
					raw = interfaceEdit(raw, []uint32{456, 437, 417}, func(b []byte) []byte { return append(b, unitPart(413, nil)...) })
				case "missing-peer":
					raw = interfaceEdit(raw, []uint32{456, other}, func([]byte) []byte { return nil })
				case "wrong-proof":
					cause = ecurUnitCause(other)
				case "action":
					raw = append(raw, actionWord(436, 0)...)
				case "status":
					raw = interfaceEdit(raw, []uint32{873, 2000, 2015}, func(b []byte) []byte { return append(b, interfacePart(2014, nil)...) })
				case "trailing":
					raw = append(raw, 42)
				case "sequence":
					raw = interfaceEdit(raw, []uint32{415}, func([]byte) []byte { return actionWord(415, 8) })
				case "reason":
					raw = interfaceEdit(raw, path, func(b []byte) []byte { return append(b, unitLeaf(true, 7)...) })
				}
				q, c, ok := QuoteInvalid(raw, cause, l, budget)
				if variant != "exact" && variant != "reverse" {
					if ok || c != 0 || q != nil {
						t.Fatal(c, q)
					}
					return
				}
				value := uint64(17)
				if kind == 3 {
					value = 1
				}
				if !ok || c != 5008 || !bytes.Equal(q, unitPart(456, unitPart(code, unitLeaf(false, value)))) {
					t.Fatal(c, q)
				}
			})
		}
	}
}

func fuzzECURUnit(t *testing.T, raw []byte, budget uint16) {
	for _, code := range [...]uint32{437, 446} {
		q, c, ok := QuoteInvalid(raw, ecurUnitCause(code), wire.DefaultLimits(), int(budget))
		if !ok {
			if c != 0 || q != nil {
				t.Fatal(c, q)
			}
			continue
		}
		if c != 5008 || len(q) != 32 || len(q) > int(budget) || !bytes.Equal(q[:8], []byte{0, 0, 1, 200, 64, 0, 0, 32}) || binary.BigEndian.Uint32(q[8:]) != code || !bytes.Contains(raw, q[8:]) {
			t.Fatal(c, q)
		}
	}
}
