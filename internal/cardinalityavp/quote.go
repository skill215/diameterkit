// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Independently checked bounded repeated-group omission diagnostics.
// SPDX-AI-Date: 2026-11-18

// Package cardinalityavp checks named repeated-group omissions against the wire.
package cardinalityavp

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type evidence interface{ DiameterMissingCardinalityV1() (string, int) }

func HasEvidence(err error) bool { var p evidence; return errors.As(err, &p) }

type rule struct {
	group, discriminator, value avp.Key
	flags                       byte
	types                       []uint32
}

func selected(name string) (rule, bool) {
	switch name {
	case "combined-subscriptions-v1":
		return rule{avp.Key{Code: 443}, avp.Key{Code: 450}, avp.Key{Code: 444}, 64, []uint32{0, 1}}, true
	}
	return rule{}, false
}

// Quote rechecks the count and ordered type subsequence from the original wire.
// Schema callbacks are trusted code. Missing instances get zero-filled minimum
// groups, never business values or inferred discriminator values.
func Quote(src []byte, cause error, limits wire.Limits, budget int) ([]byte, bool) {
	var proof evidence
	var field *validation.FieldError
	if limits.Validate() != nil || budget < 0 || len(src) > limits.MaxFrameBytes-20 || !errors.Is(cause, validation.ErrMissing) || !errors.As(cause, &proof) || !errors.As(cause, &field) {
		return nil, false
	}
	name, have := proof.DiameterMissingCardinalityV1()
	r, ok := selected(name)
	if !ok || have < 0 || have >= len(r.types) || len(field.Path) != 1 || field.Path[0] != r.group {
		return nil, false
	}
	d := wire.NewDecoder(src, limits)
	count, previous := 0, -1
	for {
		v, found, err := d.Next()
		if err != nil {
			return nil, false
		}
		if !found {
			break
		}
		if v.Key != r.group {
			continue
		}
		count++
		if count > have || count > limits.MaxRepeated || v.Flags&0x1f != 0 {
			return nil, false
		}
		child, err := d.Child(v.Data)
		if err != nil {
			return nil, false
		}
		seenType, seenValue := false, false
		var typ uint32
		for {
			leaf, found, err := child.Next()
			if err != nil {
				return nil, false
			}
			if !found {
				break
			}
			if leaf.Flags&0x1f != 0 {
				return nil, false
			}
			switch leaf.Key {
			case r.discriminator:
				if seenType {
					return nil, false
				}
				seenType = true
				typ, err = wire.U32(leaf.Data)
				if err != nil {
					return nil, false
				}
			case r.value:
				if seenValue || bytes.IndexByte(leaf.Data, 0) >= 0 {
					return nil, false
				}
				seenValue = true
			default:
				return nil, false
			}
		}
		if !seenType || !seenValue {
			return nil, false
		}
		position := -1
		for i, t := range r.types {
			if t == typ {
				position = i
				break
			}
		}
		if position <= previous {
			return nil, false
		}
		previous = position
	}
	if count != have {
		return nil, false
	}
	missing := len(r.types) - have
	head := 8
	if r.group.VendorID != 0 {
		head = 12
	}
	size := missing * (3*head + 4)
	if missing > limits.MaxRepeated || 3*missing+1 > limits.MaxAVPs || limits.MaxDepth < 2 || size > budget || size > limits.MaxFieldBytes || size > limits.MaxFrameBytes-28 {
		return nil, false
	}
	out := make([]byte, size)
	w := wire.NewEncoder(out, limits)
	for i := 0; i < missing; i++ {
		mark := w.BeginGroup(r.group, r.flags)
		w.Uint32(r.discriminator, r.flags, 0)
		w.Bytes(r.value, r.flags, nil)
		w.EndGroup(mark)
	}
	if w.Err() != nil {
		return nil, false
	}
	return out, true
}
