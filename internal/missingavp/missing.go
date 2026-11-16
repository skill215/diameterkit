// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded missing-AVP decode evidence and diagnostic construction.
// SPDX-AI-Date: 2026-11-16

// Package missingavp carries schema-known omissions through generated decoders.
// Evidence contains keys and flags only; it never retains request bytes or values.
package missingavp

import (
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Generated decoders implement this versioned convention on private error types
// so external generated packages need no import of a core internal package.
// As with all schema callbacks, this is trusted code, not wire authentication.
type evidence interface {
	DiameterMissingAVPV1() (avp.Key, byte, int, error)
}

// Quote returns the minimal ancestor chain and zero-filled fixed-width leaf.
// The extra Failed-AVP wrapper consumes one depth and one AVP from the limits.
// A modified public FieldError path cannot silently redirect private evidence.
func Quote(err error, limits wire.Limits, maxBytes int) ([]byte, bool) {
	var field *validation.FieldError
	var proof evidence
	if limits.Validate() != nil || maxBytes < 0 || !errors.Is(err, validation.ErrMissing) || !errors.As(err, &field) || !errors.As(err, &proof) {
		return nil, false
	}
	// Freeze the callback metadata once, with a fixed maximum of 64 nodes.
	type node struct {
		key   avp.Key
		flags byte
		width int
	}
	var chain [64]node
	count, size := 0, 0
	for p := proof; p != nil; {
		key, flags, width, child := p.DiameterMissingAVPV1()
		if count >= len(field.Path) || count >= limits.MaxDepth || count+2 > limits.MaxAVPs || field.Path[count] != key || flags&0x1f != 0 || (key.VendorID != 0) != (flags&avp.VendorFlag != 0) {
			return nil, false
		}
		chain[count] = node{key, flags, width}
		count++
		size += 8
		if key.VendorID != 0 {
			size += 4
		}
		if child == nil {
			if width != 4 && width != 8 {
				return nil, false
			}
			size += width
			break
		}
		var ok bool
		p, ok = child.(evidence)
		if !ok || width != 0 {
			return nil, false
		}
	}
	if count != len(field.Path) || size > maxBytes || size > limits.MaxFieldBytes || size+28 > limits.MaxFrameBytes {
		return nil, false
	}
	out := make([]byte, size)
	w := wire.NewEncoder(out, limits)
	var marks [64]wire.GroupMark
	for i := 0; i < count-1; i++ {
		marks[i] = w.BeginGroup(chain[i].key, chain[i].flags)
	}
	leaf := chain[count-1]
	var zero [8]byte
	w.Bytes(leaf.key, leaf.flags, zero[:leaf.width])
	for i := count - 2; i >= 0; i-- {
		w.EndGroup(marks[i])
	}
	return out, w.Err() == nil
}
