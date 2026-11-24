// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded conditional EVENT Requested-Action value diagnostics.
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

// This private convention is deliberately limited to two reviewed combined
// EVENT branches. Schema callbacks are trusted code, not wire authentication.
// No generic unsupported/conditional error is sufficient to authorize a quote.
type actionEvidence interface{ DiameterActionValueV1() string }

func quoteAction(src []byte, cause error, proof actionEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 ||
		maxBytes < 12 || limits.MaxFieldBytes < 12 || limits.MaxFrameBytes < 40 ||
		limits.MaxAVPs < 2 || limits.MaxDepth < 1 ||
		!errors.Is(cause, validation.ErrField) || !errors.Is(cause, validation.ErrUnsupported) ||
		!errors.As(cause, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 436}) {
		return nil, 0, false
	}
	rule := proof.DiameterActionValueV1()
	if rule != "combined-iec-event-v1" && rule != "combined-money-event-v1" {
		return nil, 0, false
	}
	// Scan the complete top-level area before allocating. A static rule proof has
	// no wire offset: reject duplicate selectors/action and trailing malformed AVPs.
	// Earlier decoding/validation supplies the rest of the request's semantics.
	d := wire.NewDecoder(src, limits)
	var action wire.View
	var kind, indicator uint32
	var actions, kinds, indicators, equipment int
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, 0, false
		}
		if !ok {
			break
		}
		if v.Key.VendorID != 0 {
			continue
		}
		switch v.Key.Code {
		case 436:
			actions++
			if actions > 1 || v.Flags != avp.MandatoryFlag || len(v.Data) != 4 {
				return nil, 0, false
			}
			action = v
		case 416:
			kinds++
			if kinds > 1 || len(v.Data) != 4 {
				return nil, 0, false
			}
			kind = binary.BigEndian.Uint32(v.Data)
		case 455:
			indicators++
			if indicators > 1 || len(v.Data) != 4 {
				return nil, 0, false
			}
			indicator = binary.BigEndian.Uint32(v.Data)
		case 458:
			equipment++
			if equipment > 1 {
				return nil, 0, false
			}
		}
	}
	if actions != 1 || kinds != 1 || kind != 4 {
		return nil, 0, false
	}
	value := binary.BigEndian.Uint32(action.Data)
	if rule == "combined-iec-event-v1" {
		if indicators != 1 || equipment != 1 || indicator != 1 || value == 0 {
			return nil, 0, false
		}
	} else if indicators != 0 || equipment != 0 || value <= 2 {
		return nil, 0, false
	}
	// Own exactly the original fixed-width instance, including its actual value.
	// Failed-AVP and complete-answer budgets are checked independently by callers.
	out := make([]byte, 12)
	copy(out, src[action.Offset:action.Offset+12])
	return out, 5004, true
}

// The callback establishes the isolated violation. Recheck the precise action
// and SMS operation selectors on the complete bounded top-level wire area.
func quoteExcludedECURAction(src []byte, cause error, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || l.MaxFrameBytes < 40 || l.MaxFieldBytes < 12 || budget < 12 ||
		!errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 436}) {
		return nil, 0, false
	}
	top, ok := interfaceFields(wire.NewDecoder(src, l), avp.Key{Code: 416}, avp.Key{Code: 415}, avp.Key{Code: 436}, avp.Key{Code: 455}, avp.Key{Code: 458})
	if !ok {
		return nil, 0, false
	}
	for _, v := range top[:4] {
		if v.Flags != avp.MandatoryFlag || len(v.Data) != 4 {
			return nil, 0, false
		}
	}
	kind := binary.BigEndian.Uint32(top[0].Data)
	number := binary.BigEndian.Uint32(top[1].Data)
	if !((kind == 1 && number == 0) || (kind == 3 && number == 1)) || binary.BigEndian.Uint32(top[3].Data) != 1 || top[4].Flags != avp.MandatoryFlag {
		return nil, 0, false
	}
	offset := top[2].Offset
	return bytes.Clone(src[offset : offset+12]), 5008, true
}
