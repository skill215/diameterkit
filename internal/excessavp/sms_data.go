// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated SMS address NUL-value diagnostics.
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
	if selected != 2008 && selected != 2010 && selected != 2017 {
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
	var seen [14]bool
	var leaf wire.View
	bad := 0
	// Close the SMS group and check all address counterparts. Interface payloads
	// remain opaque to this consumer; the trusted decoder validates them first.
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
		case 2016:
			i = 0
		case 2001:
			i = 1
		case 2002:
			i = 2
		case 2007:
			i = 3
		case 2009:
			i = 4
		case 2013:
			i = 5
		case 2018:
			i = 6
		case 2008:
			i = 7
		case 2010:
			i = 8
		case 2017:
			i = 9
		case 2011:
			i = 10
		case 2015:
			i = 11
		case 2014:
			i = 12
		case 2012:
			i = 13
		}
		if i < 0 || seen[i] {
			return nil, 0, false
		}
		seen[i] = true
		if (i == 0 || i == 1 || i == 3 || i == 10 || i == 13) && len(v.Data) != 4 {
			return nil, 0, false
		}
		if i >= 7 && i <= 9 && bytes.IndexByte(v.Data, 0) >= 0 {
			bad++
		}
		if v.Key.Code == selected {
			leaf = v
		}
	}
	for _, yes := range seen[:6] {
		if !yes {
			return nil, 0, false
		}
	}
	if bad != 1 || leaf.Key.Code != selected || bytes.IndexByte(leaf.Data, 0) < 0 {
		return nil, 0, false
	}

	leafSize := (12 + len(leaf.Data) + 3) &^ 3
	size := 24 + leafSize
	if size > budget || size > l.MaxFieldBytes || size+28 > l.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	for i, c := range [2]uint32{873, 2000} {
		off := 12 * i
		binary.BigEndian.PutUint32(out[off:], c)
		out[off+4] = 192
		wire.Put24(out[off+5:off+8], uint32(size-off))
		binary.BigEndian.PutUint32(out[off+8:], 10415)
	}
	copy(out[24:], area[leaf.Offset:leaf.Offset+leafSize])
	return out, 5004, true
}
