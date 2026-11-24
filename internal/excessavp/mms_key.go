// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated MMS SMSKey NUL-value diagnostics.
// SPDX-AI-Date: 2026-11-23

package excessavp

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type mmsKeyEvidence interface{ DiameterMMSKeyValueV1() string }

func quoteMMSKey(src []byte, cause error, proof mmsKeyEvidence, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 40 || l.MaxFieldBytes < 40 || l.MaxFrameBytes < 68 || l.MaxAVPs < 4 || l.MaxDepth < 3 || proof.DiameterMMSKeyValueV1() != "combined-mms-key-nul-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	for i, c := range [3]uint32{873, 877, 1210} {
		if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, l)
	area := src
	for _, k := range f.Path[:2] {
		fields, ok := interfaceFields(d, k)
		if !ok || fields[0].Flags != 192 {
			return nil, 0, false
		}
		area = fields[0].Data
		child, e := d.Child(area)
		if e != nil {
			return nil, 0, false
		}
		d = child
	}
	var seen [7]bool
	var leaf wire.View
	// Close the selected MMS group, while leaving address payload validation to
	// the decoder that issued the evidence. All containing scans share a counter.
	for {
		v, more, e := d.Next()
		if e != nil {
			return nil, 0, false
		}
		if !more {
			break
		}
		if v.Key.VendorID != 10415 || v.Flags != 192 {
			return nil, 0, false
		}
		i := -1
		switch v.Key.Code {
		case 886:
			i = 0
		case 1201:
			i = 1
		case 1202:
			i = 2
		case 1209:
			i = 3
		case 1212:
			i = 4
		case 1210:
			i = 5
			leaf = v
		case 1216:
			i = 6
		}
		if i < 0 || seen[i] {
			return nil, 0, false
		}
		seen[i] = true
		if (i == 2 || i == 3 || i == 4 || i == 6) && len(v.Data) != 4 {
			return nil, 0, false
		}
	}
	if !seen[0] || !seen[5] || bytes.IndexByte(leaf.Data, 0) < 0 {
		return nil, 0, false
	}
	return quoteServiceLeaf(area, leaf, 877, 5004, l, budget)
}
