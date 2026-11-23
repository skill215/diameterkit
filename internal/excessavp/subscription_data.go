// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated subscription value NUL diagnostics with bounded original quotes.
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

type subscriptionDataEvidence interface{ DiameterSubscriptionDataValueV1() string }

func quoteSubscriptionData(src []byte, cause error, proof subscriptionDataEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 20 || limits.MaxFieldBytes < 20 || limits.MaxFrameBytes < 48 || limits.MaxAVPs < 3 || limits.MaxDepth < 2 || proof.DiameterSubscriptionDataValueV1() != "combined-subscription-data-nul-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 2 || field.Path[0] != (avp.Key{Code: 443}) || field.Path[1] != (avp.Key{Code: 444}) {
		return nil, 0, false
	}
	d := wire.NewDecoder(src, limits)
	var groups [2]wire.View
	count := 0
	// Scan the whole root under one source budget, retaining only the two schema
	// slots. The path has no occurrence index: resolve it in decoder wire order.
	for {
		v, more, err := d.Next()
		if err != nil {
			return nil, 0, false
		}
		if !more {
			break
		}
		if v.Key != (avp.Key{Code: 443}) {
			continue
		}
		if v.Flags != 64 || count == len(groups) || count == limits.MaxRepeated {
			return nil, 0, false
		}
		groups[count] = v
		count++
	}
	var area []byte
	var leaf wire.View
	found := false
	for _, group := range groups[:count] {
		child, err := d.Child(group.Data)
		if err != nil {
			return nil, 0, false
		}
		var seen [2]bool
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
			case avp.Key{Code: 450}:
				if len(v.Data) != 4 {
					return nil, 0, false
				}
			case avp.Key{Code: 444}:
				i = 1
				leaf = v
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
		if bytes.IndexByte(leaf.Data, 0) >= 0 {
			area = group.Data
			found = true
			break
		}
		// Earlier groups must pass this callback. Later groups stay opaque because
		// the trusted decoder stopped at the selected failure. Type pairing is a
		// later request rule, not a prerequisite for this group's data constraint.
	}
	if !found {
		return nil, 0, false
	}
	leafSize := (8 + len(leaf.Data) + 3) &^ 3
	size := 8 + leafSize
	if size > maxBytes || size > limits.MaxFieldBytes || size+28 > limits.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	binary.BigEndian.PutUint32(out, 443)
	out[4] = 64
	wire.Put24(out[5:8], uint32(size))
	copy(out[8:], area[leaf.Offset:leaf.Offset+leafSize])
	return out, 5004, true
}
