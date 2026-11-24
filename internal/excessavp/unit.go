// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Reporting-Reason and requested-unit value quotes.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type unitEvidence interface{ DiameterUnitValueV1() string }

// Complete one containing-area scan. Forbidden and unique required instances
// share the decoder's source counter; no second scan or retained payload.
func unitFields(d *wire.Decoder, forbidden []avp.Key, keys ...avp.Key) ([5]wire.View, bool) {
	var values [5]wire.View
	var seen [5]bool
	if len(keys) == 0 || len(keys) > len(values) {
		return values, false
	}
	for {
		v, ok, err := d.Next()
		if err != nil {
			return values, false
		}
		if !ok {
			break
		}
		for _, key := range forbidden {
			if v.Key == key {
				return values, false
			}
		}
		for i, key := range keys {
			if v.Key == key {
				if seen[i] {
					return values, false
				}
				seen[i] = true
				values[i] = v
			}
		}
	}
	for i := range keys {
		if !seen[i] {
			return values, false
		}
	}
	return values, true
}

func quoteUnitValue(src []byte, cause error, proof unitEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	reason := false
	switch proof.DiameterUnitValueV1() {
	case "combined-termination-reporting-reason-v1":
		reason = true
	case "combined-sms-requested-units-v1":
	default:
		return nil, 0, false
	}
	size, depth := 32, 3
	path := []avp.Key{{Code: 456}, {Code: 437}, {Code: 417}}
	if reason {
		size, depth = 24, 2
		path = []avp.Key{{Code: 456}, {VendorID: 10415, Code: 872}}
	}
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < size || limits.MaxFieldBytes < size || limits.MaxFrameBytes < 28+size || limits.MaxAVPs < depth+1 || limits.MaxDepth < depth || !errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != len(path) {
		return nil, 0, false
	}
	for i, key := range path {
		if field.Path[i] != key {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, limits)
	top, ok := interfaceFields(d, avp.Key{Code: 416}, avp.Key{Code: 455}, avp.Key{Code: 458}, path[0])
	if !ok || top[0].Flags != 64 || len(top[0].Data) != 4 || top[1].Flags != 64 || len(top[1].Data) != 4 || binary.BigEndian.Uint32(top[1].Data) != 1 || top[2].Flags != 64 || top[3].Flags != 64 {
		return nil, 0, false
	}
	kind := binary.BigEndian.Uint32(top[0].Data)
	if reason && kind != 3 || !reason && kind != 1 && kind != 4 {
		return nil, 0, false
	}
	mscc, err := d.Child(top[3].Data)
	if err != nil {
		return nil, 0, false
	}
	var leaf wire.View
	var area []byte
	if reason {
		fields, ok := unitFields(mscc, []avp.Key{{Code: 437}}, avp.Key{Code: 446}, path[1])
		if !ok || fields[0].Flags != 64 || fields[1].Flags != 192 || len(fields[1].Data) != 4 || int32(binary.BigEndian.Uint32(fields[1].Data)) == 2 {
			return nil, 0, false
		}
		used, err := mscc.Child(fields[0].Data)
		if err != nil {
			return nil, 0, false
		}
		values, ok := interfaceFields(used, avp.Key{Code: 417})
		if !ok || values[0].Flags != 64 || len(values[0].Data) != 8 {
			return nil, 0, false
		}
		leaf, area = fields[1], top[3].Data
	} else {
		var ok bool
		area, leaf, ok = requestedUnitLeaf(mscc)
		if !ok || binary.BigEndian.Uint64(leaf.Data) == 1 {
			return nil, 0, false
		}
	}
	// The two leaf wire types both occupy exactly 16 original bytes. No sibling
	// content survives. Source and complete quote limits precede allocation.
	out := make([]byte, size)
	for i := 0; i < depth-1; i++ {
		offset := i * 8
		binary.BigEndian.PutUint32(out[offset:], path[i].Code)
		out[offset+4] = 64
		wire.Put24(out[offset+5:offset+8], uint32(size-offset))
	}
	copy(out[size-16:], area[leaf.Offset:leaf.Offset+16])
	return out, 5004, true
}

// Return the sole structurally eligible requested-unit value and its original containing area.
func requestedUnitLeaf(mscc *wire.Decoder) ([]byte, wire.View, bool) {
	fields, ok := unitFields(mscc, []avp.Key{{Code: 446}, {VendorID: 10415, Code: 872}}, avp.Key{Code: 437})
	if !ok || fields[0].Flags != 64 {
		return nil, wire.View{}, false
	}
	requested, err := mscc.Child(fields[0].Data)
	if err != nil {
		return nil, wire.View{}, false
	}
	values, ok := unitFields(requested, []avp.Key{{Code: 420}, {Code: 413}}, avp.Key{Code: 417})
	if !ok || values[0].Flags != 64 || len(values[0].Data) != 8 {
		return nil, wire.View{}, false
	}
	return fields[0].Data, values[0], true
}
