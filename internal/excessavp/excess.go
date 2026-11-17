// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded first-excess AVP diagnostic construction.
// SPDX-AI-Date: 2026-11-17

// Package excessavp resolves generated cardinality evidence against the request.
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

// Quote copies the first instance exceeding a compiled maximum, together with
// minimal ancestor headers. src is the AVP area of the original request. The
// caller must also validate the full response, including known diagnostic groups.
func Quote(src []byte, cause error, limits wire.Limits, maxBytes int) ([]byte, bool) {
	var field *validation.FieldError
	var proof evidence
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 0 || !errors.Is(cause, validation.ErrDuplicate) || !errors.As(cause, &field) || !errors.As(cause, &proof) {
		return nil, false
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
			if depth+1 != len(field.Path) || maximum < 1 || maximum > 256 || seen != maximum+1 || seen > limits.MaxRepeated {
				return nil, false
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
