// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded exact-instance Originator/Recipient Address-Data NUL diagnostics.
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

type addressDataEvidence interface{ DiameterAddressDataValueV1() string }

func quoteAddressData(src []byte, cause error, proof addressDataEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 52 || limits.MaxFieldBytes < 52 || limits.MaxFrameBytes < 80 || limits.MaxAVPs < 5 || limits.MaxDepth < 4 || proof.DiameterAddressDataValueV1() != "combined-address-data-nul-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 4 {
		return nil, 0, false
	}
	for i, key := range field.Path {
		if key.VendorID != 10415 || (i == 0 && key.Code != 873) || (i == 1 && key.Code != 877) || (i == 2 && key.Code != 886 && key.Code != 1201) || (i == 3 && key.Code != 897) {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, limits)
	area := src
	// Complete each containing-area scan. All descendants share the source
	// counter. The other address group and unrelated sibling payloads stay opaque.
	for _, key := range field.Path[:3] {
		fields, ok := interfaceFields(d, key)
		if !ok || fields[0].Flags != 192 {
			return nil, 0, false
		}
		area = fields[0].Data
		child, err := d.Child(area)
		if err != nil {
			return nil, 0, false
		}
		d = child
	}
	var leaf wire.View
	var seen [2]bool
	for {
		v, more, err := d.Next()
		if err != nil {
			return nil, 0, false
		}
		if !more {
			break
		}
		i := 0
		switch v.Key {
		case avp.Key{VendorID: 10415, Code: 899}:
			if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 1 {
				return nil, 0, false
			}
		case avp.Key{VendorID: 10415, Code: 897}:
			i = 1
			leaf = v
			if bytes.IndexByte(v.Data, 0) < 0 {
				return nil, 0, false
			}
		default:
			return nil, 0, false
		}
		if seen[i] || v.Flags != 192 {
			return nil, 0, false
		}
		seen[i] = true
	}
	if !seen[0] || !seen[1] {
		return nil, 0, false
	}
	// Preserve only the complete original padded data under minimal ancestors.
	leafSize := (12 + len(leaf.Data) + 3) &^ 3
	size := 36 + leafSize
	if size > maxBytes || size > limits.MaxFieldBytes || size+28 > limits.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	for i, key := range field.Path[:3] {
		off := i * 12
		binary.BigEndian.PutUint32(out[off:], key.Code)
		out[off+4] = 192
		wire.Put24(out[off+5:off+8], uint32(size-off))
		binary.BigEndian.PutUint32(out[off+8:], 10415)
	}
	copy(out[36:], area[leaf.Offset:leaf.Offset+leafSize])
	return out, 5004, true
}
