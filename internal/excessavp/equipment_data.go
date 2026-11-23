// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated equipment value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type equipmentDataEvidence interface{ DiameterEquipmentDataValueV1() string }

func quoteEquipmentData(src []byte, cause error, proof equipmentDataEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 20 || limits.MaxFieldBytes < 20 || limits.MaxFrameBytes < 48 || limits.MaxAVPs < 3 || limits.MaxDepth < 2 || proof.DiameterEquipmentDataValueV1() != "combined-equipment-data-nul-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 2 || field.Path[0] != (avp.Key{Code: 458}) || field.Path[1] != (avp.Key{Code: 460}) {
		return nil, 0, false
	}
	d := wire.NewDecoder(src, limits)
	top, ok := interfaceFields(d, avp.Key{Code: 458})
	if !ok || top[0].Flags != 64 {
		return nil, 0, false
	}
	child, err := d.Child(top[0].Data)
	if err != nil {
		return nil, 0, false
	}
	var leaf wire.View
	var seen [2]bool
	// A complete group scan shares the root source budget. Both known children
	// must occur exactly once; any extra child makes the value error non-isolated.
	for {
		v, more, err := child.Next()
		if err != nil {
			return nil, 0, false
		}
		if !more {
			break
		}
		i := 0
		switch v.Key {
		case avp.Key{Code: 459}:
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 0 {
				return nil, 0, false
			}
		case avp.Key{Code: 460}:
			i = 1
			leaf = v
			if bytes.IndexByte(v.Data, 0) < 0 {
				return nil, 0, false
			}
		default:
			return nil, 0, false
		}
		if seen[i] || v.Flags != 64 {
			return nil, 0, false
		}
		seen[i] = true
	}
	if !seen[0] || !seen[1] {
		return nil, 0, false
	}
	// Only the original padded value survives; the type and other fields
	// remain outside the quote. All source/output checks precede allocation.
	leafSize := (8 + len(leaf.Data) + 3) &^ 3
	size := 8 + leafSize
	if size > maxBytes || size > limits.MaxFieldBytes || size+28 > limits.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	binary.BigEndian.PutUint32(out, 458)
	out[4] = 64
	wire.Put24(out[5:8], uint32(size))
	copy(out[8:], top[0].Data[leaf.Offset:leaf.Offset+leafSize])
	return out, 5004, true
}
