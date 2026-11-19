// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded original-instance combined indicator value diagnostics.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type indicatorEvidence interface{ DiameterIndicatorValueV1() string }

func quoteIndicatorValue(src []byte, cause error, proof indicatorEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 12 || limits.MaxFieldBytes < 12 || limits.MaxFrameBytes < 40 || limits.MaxAVPs < 2 || limits.MaxDepth < 1 || proof.DiameterIndicatorValueV1() != "combined-sms-indicator-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 455}) {
		return nil, 0, false
	}
	// Scan the entire containing area once, including forbidden peers. Required
	// markers, the selected instance and unrelated siblings share one AVP budget.
	d := wire.NewDecoder(src, limits)
	fields, ok := unitFields(d, []avp.Key{{Code: 439}, {VendorID: 10425, Code: 23}, {Code: 1292}}, avp.Key{Code: 416}, avp.Key{Code: 455}, avp.Key{Code: 458})
	if !ok || fields[0].Flags != 64 || len(fields[0].Data) != 4 || fields[1].Flags != 64 || len(fields[1].Data) != 4 || fields[2].Flags != 64 {
		return nil, 0, false
	}
	kind := binary.BigEndian.Uint32(fields[0].Data)
	if kind != 1 && kind != 3 && kind != 4 || binary.BigEndian.Uint32(fields[1].Data) == 1 {
		return nil, 0, false
	}
	// No ancestor is needed. Source and output limits precede this owned copy.
	out := make([]byte, 12)
	copy(out, src[fields[1].Offset:fields[1].Offset+12])
	return out, 5004, true
}
