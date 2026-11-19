// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded exact-instance equipment-type value diagnostics.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type equipmentEvidence interface{ DiameterEquipmentTypeValueV1() string }

func quoteEquipmentValue(src []byte, cause error, proof equipmentEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 20 || limits.MaxFieldBytes < 20 || limits.MaxFrameBytes < 48 || limits.MaxAVPs < 3 || limits.MaxDepth < 2 || proof.DiameterEquipmentTypeValueV1() != "combined-equipment-type-v1" || !errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 2 || field.Path[0] != (avp.Key{Code: 458}) || field.Path[1] != (avp.Key{Code: 459}) {
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
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) == 0 {
				return nil, 0, false
			}
			leaf = v
		case avp.Key{Code: 460}:
			i = 1
			if bytes.IndexByte(v.Data, 0) >= 0 {
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
	// Allocation follows all source/output checks. No identity bytes are copied.
	out := make([]byte, 20)
	binary.BigEndian.PutUint32(out, 458)
	out[4] = 64
	wire.Put24(out[5:8], 20)
	copy(out[8:], top[0].Data[leaf.Offset:leaf.Offset+12])
	return out, 5004, true
}
