// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated SMS template value quotes.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type smsEvidence interface{ DiameterSMSValueV1() string }

func quoteSMSValue(src []byte, cause error, proof smsEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 40 ||
		limits.MaxFieldBytes < 40 || limits.MaxFrameBytes < 68 || limits.MaxAVPs < 4 || limits.MaxDepth < 3 ||
		!errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) ||
		!errors.As(cause, &field) || len(field.Path) != 3 {
		return nil, 0, false
	}
	selected := 0
	switch proof.DiameterSMSValueV1() {
	case "combined-sms-node-v1":
	case "combined-sm-message-type-v1":
		selected = 1
	default:
		return nil, 0, false
	}
	leaves := [2]avp.Key{{VendorID: 10415, Code: 2016}, {VendorID: 10415, Code: 2007}}
	path := [3]avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, leaves[selected]}
	for i, key := range path {
		if field.Path[i] != key {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, limits)
	top, ok := interfaceFields(d, avp.Key{Code: 416}, avp.Key{Code: 455}, avp.Key{Code: 458}, path[0])
	if !ok || top[0].Flags != 64 || len(top[0].Data) != 4 || top[1].Flags != 64 || len(top[1].Data) != 4 ||
		binary.BigEndian.Uint32(top[1].Data) != 1 || top[2].Flags != 64 || top[3].Flags != 192 {
		return nil, 0, false
	}
	kind := binary.BigEndian.Uint32(top[0].Data)
	if kind != 1 && kind != 3 && kind != 4 {
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
	values, ok := interfaceFields(sms, leaves[:]...)
	if !ok {
		return nil, 0, false
	}
	// Both selected and counterpart instances consume the same source budget.
	for i, v := range values[:2] {
		if v.Flags != 192 || len(v.Data) != 4 || (i == selected) == (binary.BigEndian.Uint32(v.Data) == 0) {
			return nil, 0, false
		}
	}
	// Allocate only after all source and output limits have been checked.
	out := make([]byte, 40)
	for i := 0; i < 2; i++ {
		offset := i * 12
		binary.BigEndian.PutUint32(out[offset:], path[i].Code)
		out[offset+4] = 192
		wire.Put24(out[offset+5:offset+8], uint32(40-offset))
		binary.BigEndian.PutUint32(out[offset+8:], 10415)
	}
	leaf := values[selected]
	copy(out[24:], middle[0].Data[leaf.Offset:leaf.Offset+16])
	return out, 5004, true
}
