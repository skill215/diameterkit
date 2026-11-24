// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Originator-SCCP-Address exclusion on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func quoteECURExclusion(src []byte, cause error, l wire.Limits, budget int, rule smsRequestRule) ([]byte, uint32, bool) {
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 36 || l.MaxFieldBytes < 36 || l.MaxDepth < 4 || !errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	for i, c := range [...]uint32{873, 2000, uint32(rule)} {
		if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
			return nil, 0, false
		}
	}
	d, area, kind := smsRequestGroup(src, l, rule)
	if d == nil {
		return nil, 0, false
	}
	selected, ok := scanECURExclusionSMS(d, kind, rule)
	if !ok {
		return nil, 0, false
	}
	return quoteServiceLeaf(area, selected, 2000, 5008, l, budget)
}

// A zero rule checks the normal SMS fields accompanying an MSCC exclusion.
func scanECURExclusionSMS(d *wire.Decoder, kind uint32, rule smsRequestRule) (wire.View, bool) {
	required := []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012}
	if kind == 1 {
		required = required[:5]
	}
	if rule != 0 {
		required = append(required, uint32(rule))
	}
	var selected, client, smsc wire.View
	ok := scanSMSGroup(d, required, func(v wire.View) bool {
		if smsAddress(v.Key.Code) && bytes.IndexByte(v.Data, 0) >= 0 {
			return false
		}
		switch v.Key.Code {
		case 2008:
			if rule != ecurOriginator {
				return false
			}
			selected = v
		case 2014:
			if rule == initialStatus {
				selected = v
				return true
			}
			return kind == 3
		case 2018:
			client = v
		case 2017:
			smsc = v
		case 2016, 2007:
			return binary.BigEndian.Uint32(v.Data) == 0
		case 2012:
			return kind == 3
		case 2009, 2002:
			expected := uint32(2)
			if v.Key.Code == 2009 {
				expected = 3
			}
			child, err := d.Child(v.Data)
			if err != nil {
				return false
			}
			types, ok := interfaceFields(child, avp.Key{VendorID: 10415, Code: 2006})
			return ok && types[0].Flags == 192 && len(types[0].Data) == 4 && binary.BigEndian.Uint32(types[0].Data) == expected
		}
		return true
	})
	return selected, ok && bytes.Equal(client.Data, smsc.Data)
}
