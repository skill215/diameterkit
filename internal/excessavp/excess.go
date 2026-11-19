// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded present-AVP and typed-length diagnostic construction.
// SPDX-AI-Date: 2026-11-17

// Package excessavp resolves compiled present-AVP evidence against the request.
// Errors retain offsets and schema limits, never request bytes or subscriber values.
package excessavp

import (
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Generated external packages implement this convention on private error types.
// Schema callbacks are trusted code; this metadata does not authenticate a peer.
type evidence interface {
	DiameterExcessAVPV1() (avp.Key, byte, int, int, error)
}

type invalidEvidence interface {
	DiameterInvalidAVPV1() (avp.Key, byte, int, uint32, error)
}

// A length mapping additionally names a generated wire type. The walker checks
// its length rule again against the selected instance; ErrValue alone is not proof.
type lengthEvidence interface {
	DiameterLengthAVPV1() string
}

func invalidLength(kind string, data []byte) bool {
	switch kind {
	case "u32", "i32", "time":
		return len(data) != 4
	case "u64", "i64":
		return len(data) != 8
	case "tbcd":
		return len(data) == 0 || len(data) > 8
	case "address":
		if len(data) < 2 {
			return true
		}
		family := binary.BigEndian.Uint16(data)
		return family == 1 && len(data) != 6 || family == 2 && len(data) != 18
	default:
		return false
	}
}

// invalidNode adapts value evidence to the common bounded instance walker.
// Negative detail values are reserved for the selected result codes.
type invalidNode struct{ invalidEvidence }

func (n invalidNode) Error() string { return "invalid AVP evidence" }
func (n invalidNode) DiameterExcessAVPV1() (avp.Key, byte, int, int, error) {
	k, flags, offset, code, child := n.DiameterInvalidAVPV1()
	if p, ok := child.(invalidEvidence); ok {
		child = invalidNode{p}
	}
	// Check before converting to int, including on 32-bit targets.
	if code != 0 && code != 5001 && code != 5004 && code != 5014 {
		return k, flags, offset, -1, child
	}
	return k, flags, offset, -int(code), child
}

// QuoteInvalid quotes a selected isolated value, unknown Mandatory AVP or
// typed-length failure from generated or handwritten decoders. Bare sentinels,
// unsupported domain rules, framing and resource errors supply no proof.
func QuoteInvalid(src []byte, cause error, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var parameter moneyParameterEvidence
	if errors.As(cause, &parameter) {
		return quoteMoneyParameter(src, cause, parameter, limits, maxBytes)
	}
	var moneyNode moneyNodeEvidence
	if errors.As(cause, &moneyNode) {
		return quoteMoneyNode(src, cause, moneyNode, limits, maxBytes)
	}
	var address addressEvidence
	if errors.As(cause, &address) {
		return quoteAddressValue(src, cause, address, limits, maxBytes)
	}
	var equipment equipmentEvidence
	if errors.As(cause, &equipment) {
		return quoteEquipmentValue(src, cause, equipment, limits, maxBytes)
	}
	var indicator indicatorEvidence
	if errors.As(cause, &indicator) {
		return quoteIndicatorValue(src, cause, indicator, limits, maxBytes)
	}
	var unit unitEvidence
	if errors.As(cause, &unit) {
		return quoteUnitValue(src, cause, unit, limits, maxBytes)
	}
	var sms smsEvidence
	if errors.As(cause, &sms) {
		return quoteSMSValue(src, cause, sms, limits, maxBytes)
	}
	var inter interfaceEvidence
	if errors.As(cause, &inter) {
		return quoteInterface(src, cause, inter, limits, maxBytes)
	}
	var action actionEvidence
	if errors.As(cause, &action) {
		return quoteAction(src, cause, action, limits, maxBytes)
	}
	var code uint32
	out, ok := quote(src, cause, limits, maxBytes, &code)
	if !ok {
		return nil, 0, false
	}
	return out, code, true
}

// Quote copies the first instance exceeding a compiled maximum, together with
// minimal ancestor headers. src is the AVP area of the original request. The
// caller must also validate the full response, including known diagnostic groups.
func Quote(src []byte, cause error, limits wire.Limits, maxBytes int) ([]byte, bool) {
	return quote(src, cause, limits, maxBytes, nil)
}

func quote(src []byte, cause error, limits wire.Limits, maxBytes int, result *uint32) ([]byte, bool) {
	var field *validation.FieldError
	var proof evidence
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 0 || !errors.As(cause, &field) {
		return nil, false
	}
	if result == nil {
		if !errors.Is(cause, validation.ErrDuplicate) || !errors.As(cause, &proof) {
			return nil, false
		}
	} else {
		var invalid invalidEvidence
		if !errors.As(cause, &invalid) {
			return nil, false
		}
		proof = invalidNode{invalid}
	}
	var parents [64]wire.View
	depth, headers := 0, 0
	d := wire.NewDecoder(src, limits)
	for {
		if depth >= len(field.Path) || depth >= limits.MaxDepth || depth+2 > limits.MaxAVPs {
			return nil, false
		}
		key, flags, offset, maximum, child := proof.DiameterExcessAVPV1()
		if key != field.Path[depth] || offset < 0 || offset%4 != 0 {
			return nil, false
		}
		var selected wire.View
		seen := 0
		for {
			v, ok, err := d.Next()
			if err != nil || !ok || v.Offset > offset {
				return nil, false
			}
			if v.Key == key {
				seen++
			}
			if v.Offset == offset {
				if v.Key != key || v.Flags != flags {
					return nil, false
				}
				selected = v
				break
			}
		}
		if child == nil {
			if depth+1 != len(field.Path) || seen > limits.MaxRepeated {
				return nil, false
			}
			if result == nil {
				if maximum < 1 || maximum > 256 || seen != maximum+1 {
					return nil, false
				}
			} else {
				switch maximum {
				case -5001:
					if flags&avp.MandatoryFlag == 0 || !errors.Is(cause, validation.ErrUnsupported) {
						return nil, false
					}
				case -5004:
					if !errors.Is(cause, validation.ErrField) && !errors.Is(cause, avp.ErrValue) {
						return nil, false
					}
				case -5014:
					n, ok := proof.(invalidNode)
					if !ok || !errors.Is(cause, avp.ErrValue) {
						return nil, false
					}
					typed, ok := n.invalidEvidence.(lengthEvidence)
					if !ok || !invalidLength(typed.DiameterLengthAVPV1(), selected.Data) {
						return nil, false
					}
				default:
					return nil, false
				}
				*result = uint32(-maximum)
			}
			head := 8
			if key.VendorID != 0 {
				head = 12
			}
			length := (head + len(selected.Data) + 3) &^ 3
			size := headers + length
			if size > maxBytes || size > limits.MaxFieldBytes || size > limits.MaxFrameBytes-28 {
				return nil, false
			}
			// Allocation is bounded before copying any payload. Only the offending
			// instance retains its complete header/data/padding; no sibling is copied.
			out := make([]byte, size)
			pos := 0
			for i := 0; i < depth; i++ {
				p := parents[i]
				h := 8
				if p.Key.VendorID != 0 {
					h = 12
				}
				binary.BigEndian.PutUint32(out[pos:], p.Key.Code)
				out[pos+4] = p.Flags
				wire.Put24(out[pos+5:pos+8], uint32(size-pos))
				if h == 12 {
					binary.BigEndian.PutUint32(out[pos+8:], p.Key.VendorID)
				}
				pos += h
			}
			copy(out[pos:], src[offset:offset+length])
			return out, true
		}
		if maximum != 0 || flags&0x1f != 0 {
			return nil, false
		}
		var ok bool
		proof, ok = child.(evidence)
		if !ok {
			return nil, false
		}
		parents[depth] = wire.View{Key: key, Flags: flags}
		headers += 8
		if key.VendorID != 0 {
			headers += 4
		}
		depth++
		var err error
		d, err = d.Child(selected.Data)
		if err != nil {
			return nil, false
		}
		src = selected.Data
	}
}
