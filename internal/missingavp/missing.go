// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded missing-AVP decode evidence and diagnostic construction.
// SPDX-AI-Date: 2026-11-18

// Package missingavp carries schema-known omissions through generated decoders.
// Evidence contains schema metadata and selected scalar condition facts, never
// borrowed request buffers or subscriber values.
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

// Zero/one-byte leaves require an explicit wire-type proof. Legacy fixed-width
// producers retain their existing convention and behavior.
type variableEvidence interface{ DiameterMissingVariableV1() string }

func variableWidth(kind string, width int) bool {
	switch kind {
	case "identity", "tbcd", "utf8-nonempty", "octets-nonempty":
		return width == 1
	case "utf8", "octets":
		return width == 0
	}
	return false
}

// Quote returns a minimum ancestor chain and a proven zero-filled diagnostic.
// The extra Failed-AVP wrapper consumes one depth and one AVP from the limits.
// A modified public FieldError path cannot silently redirect private evidence.
func Quote(err error, limits wire.Limits, maxBytes int) ([]byte, bool) {
	var typed groupEvidenceV2
	if errors.As(err, &typed) {
		return quoteGroup(err, typed, 2, limits, maxBytes)
	}
	var group groupEvidence
	if errors.As(err, &group) {
		return quoteGroup(err, group, 1, limits, maxBytes)
	}
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
		kind := ""
		if typed, ok := p.(variableEvidence); ok {
			kind = typed.DiameterMissingVariableV1()
		}
		if count >= len(field.Path) || count >= limits.MaxDepth || count+2 > limits.MaxAVPs || field.Path[count] != key || flags&0x1f != 0 || (key.VendorID != 0) != (flags&avp.VendorFlag != 0) {
			return nil, false
		}
		chain[count] = node{key, flags, width}
		count++
		size += 8
		if key.VendorID != 0 {
			size += 4
		}
		conditional, conditioned := p.(conditionalEvidence)
		if conditioned && (child != nil || count != 1 || kind != "" || !conditionalMinimum(conditional, key, flags, width)) {
			return nil, false
		}
		if child == nil {
			if !conditioned && (kind != "" && !variableWidth(kind, width) || kind == "" && width != 4 && width != 8) {
				return nil, false
			}
			size += (width + 3) &^ 3
			break
		}
		var ok bool
		p, ok = child.(evidence)
		if !ok || width != 0 || kind != "" {
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
	if w.Err() != nil {
		return nil, false
	}
	return out, true
}
