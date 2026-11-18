// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated INITIAL/TERMINATION interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package excessavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func quoteECURInterface(src []byte, cause error, proof interfaceEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 52 ||
		limits.MaxFieldBytes < 52 || limits.MaxFrameBytes < 80 || limits.MaxAVPs < 5 || limits.MaxDepth < 4 ||
		!errors.Is(cause, validation.ErrField) || !errors.Is(cause, validation.ErrUnsupported) ||
		!errors.As(cause, &field) || len(field.Path) != 4 {
		return nil, 0, false
	}
	selected := 0
	switch proof.DiameterInterfaceValueV1() {
	case "combined-ecur-originator-interface-v1":
	case "combined-ecur-destination-interface-v1":
		selected = 1
	default:
		return nil, 0, false
	}
	groups := [2]avp.Key{{VendorID: 10415, Code: 2009}, {VendorID: 10415, Code: 2002}}
	path := [4]avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, groups[selected], {VendorID: 10415, Code: 2006}}
	for i, key := range path {
		if field.Path[i] != key {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, limits)
	top, ok := interfaceFieldsAbsent(d, avp.Key{Code: 436}, avp.Key{Code: 416}, avp.Key{Code: 455}, avp.Key{Code: 458}, path[0])
	if !ok || top[0].Flags != 64 || len(top[0].Data) != 4 || top[1].Flags != 64 || len(top[1].Data) != 4 || binary.BigEndian.Uint32(top[1].Data) != 1 || top[2].Flags != 64 || top[3].Flags != 192 {
		return nil, 0, false
	}
	kind := binary.BigEndian.Uint32(top[0].Data)
	if kind != 1 && kind != 3 {
		return nil, 0, false
	}
	service, err := d.Child(top[3].Data)
	if err != nil {
		return nil, 0, false
	}
	middle, ok := interfaceFields(service, path[1])
	if !ok || middle[0].Flags != 192 {
		return nil, 0, false
	}
	sms, err := service.Child(middle[0].Data)
	if err != nil {
		return nil, 0, false
	}
	parents, ok := interfaceFieldsAbsent(sms, avp.Key{VendorID: 10415, Code: 2008}, groups[:]...)
	if !ok || parents[0].Flags != 192 || parents[1].Flags != 192 {
		return nil, 0, false
	}
	var leaves [2]wire.View
	// Both siblings use the same source counter, including the valid counterpart.
	for i := range parents[:2] {
		child, err := sms.Child(parents[i].Data)
		if err != nil {
			return nil, 0, false
		}
		leaf, ok := interfaceFields(child, path[3])
		if !ok || leaf[0].Flags != 192 || len(leaf[0].Data) != 4 {
			return nil, 0, false
		}
		leaves[i] = leaf[0]
		value := int32(binary.BigEndian.Uint32(leaf[0].Data))
		expected := int32(3 - i)
		if (i == selected) == (value == expected) {
			return nil, 0, false
		}
	}
	// Quote only the selected original leaf; no sibling or subscriber data.
	out := make([]byte, 52)
	for i := 0; i < 3; i++ {
		offset := i * 12
		binary.BigEndian.PutUint32(out[offset:], path[i].Code)
		out[offset+4] = 192
		wire.Put24(out[offset+5:offset+8], uint32(52-offset))
		binary.BigEndian.PutUint32(out[offset+8:], 10415)
	}
	leaf := leaves[selected]
	copy(out[36:], parents[selected].Data[leaf.Offset:leaf.Offset+16])
	return out, 5004, true
}
