// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated SMS address NUL-value diagnostics.
// SPDX-AI-Date: 2026-11-23

package excessavp

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type smsDataEvidence interface{ DiameterSMSDataValueV1() string }

func quoteSMSData(src []byte, cause error, proof smsDataEvidence, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 40 || l.MaxFieldBytes < 40 || l.MaxFrameBytes < 68 || l.MaxAVPs < 9 || l.MaxDepth < 3 || proof.DiameterSMSDataValueV1() != "combined-sms-data-nul-v1" || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	if f.Path[0] != (avp.Key{VendorID: 10415, Code: 873}) || f.Path[1] != (avp.Key{VendorID: 10415, Code: 2000}) || f.Path[2].VendorID != 10415 {
		return nil, 0, false
	}
	selected := f.Path[2].Code
	if !smsAddress(selected) {
		return nil, 0, false
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
	var leaf wire.View
	bad := 0
	ok := scanSMSGroup(d, nil, func(v wire.View) bool {
		if smsAddress(v.Key.Code) && bytes.IndexByte(v.Data, 0) >= 0 {
			bad++
			if v.Key.Code == selected {
				leaf = v
			}
		}
		return true
	})
	if !ok || bad != 1 || leaf.Key.Code != selected {
		return nil, 0, false
	}
	return quoteSMSLeaf(area, leaf, l, budget)
}
