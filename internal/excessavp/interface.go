// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated IEC Originator-Interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Trusted schema convention for one reviewed isolated condition. It is not a
// generic unsupported-field mapping or authentication of the original request.
type interfaceEvidence interface{ DiameterInterfaceValueV1() string }

// Scan an entire containing AVP area, requiring one instance of each selected
// key. Child decoders share the source counter; unrelated payload stays opaque.
func interfaceFields(d *wire.Decoder, keys ...avp.Key) ([5]wire.View, bool) {
	return interfaceFieldsAbsent(d, avp.Key{}, keys...)
}

// An optional forbidden key is checked in the same bounded scan.
func interfaceFieldsAbsent(d *wire.Decoder, forbidden avp.Key, keys ...avp.Key) ([5]wire.View, bool) {
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
		if forbidden.Code != 0 && v.Key == forbidden {
			return values, false
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
func quoteInterface(src []byte, cause error, proof interfaceEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	switch proof.DiameterInterfaceValueV1() {
	case "combined-ecur-originator-interface-v1", "combined-ecur-destination-interface-v1":
		return quoteECURInterface(src, cause, proof, limits, maxBytes)
	}
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 52 ||
		limits.MaxFieldBytes < 52 || limits.MaxFrameBytes < 80 || limits.MaxAVPs < 5 || limits.MaxDepth < 4 ||
		proof.DiameterInterfaceValueV1() != "combined-iec-originator-interface-v1" ||
		!errors.Is(cause, validation.ErrField) || !errors.Is(cause, validation.ErrUnsupported) ||
		!errors.As(cause, &field) || len(field.Path) != 4 {
		return nil, 0, false
	}
	path := [4]avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2009}, {VendorID: 10415, Code: 2006}}
	for i, key := range path {
		if field.Path[i] != key {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, limits)
	top, ok := interfaceFields(d, avp.Key{Code: 416}, avp.Key{Code: 436}, avp.Key{Code: 455}, avp.Key{Code: 458}, path[0])
	if !ok {
		return nil, 0, false
	}
	for i, value := range []uint32{4, 0, 1} {
		if top[i].Flags != 64 || len(top[i].Data) != 4 || binary.BigEndian.Uint32(top[i].Data) != value {
			return nil, 0, false
		}
	}
	if top[3].Flags != 64 || top[4].Flags != 192 {
		return nil, 0, false
	}
	service, err := d.Child(top[4].Data)
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
	fields, ok := interfaceFields(sms, avp.Key{VendorID: 10415, Code: 2008}, path[2])
	if !ok || fields[0].Flags != 192 || fields[1].Flags != 192 || bytes.IndexByte(fields[0].Data, 0) >= 0 {
		return nil, 0, false
	}
	inter, err := sms.Child(fields[1].Data)
	if err != nil {
		return nil, 0, false
	}
	leaf, ok := interfaceFields(inter, path[3])
	if !ok || leaf[0].Flags != 192 || len(leaf[0].Data) != 4 || int32(binary.BigEndian.Uint32(leaf[0].Data)) == 1 {
		return nil, 0, false
	}
	// Only three minimal ancestor headers and the original 16-byte leaf survive.
	// All source/quote checks precede this bounded allocation; no sibling copied.
	out := make([]byte, 52)
	for i := 0; i < 3; i++ {
		offset := i * 12
		binary.BigEndian.PutUint32(out[offset:], path[i].Code)
		out[offset+4] = 192
		wire.Put24(out[offset+5:offset+8], uint32(len(out)-offset))
		binary.BigEndian.PutUint32(out[offset+8:], 10415)
	}
	copy(out[36:], fields[1].Data[leaf[0].Offset:leaf[0].Offset+16])
	return out, 5004, true
}
