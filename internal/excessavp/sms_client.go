// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS client wire NUL and equality diagnostics with original bounded quotes.
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

type smsClientEvidence interface{ DiameterSMSClientValueV1() string }

func quoteSMSClient(src []byte, cause error, proof smsClientEvidence, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var f *validation.FieldError
	name := proof.DiameterSMSClientValueV1()
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 36 || l.MaxFieldBytes < 36 || l.MaxFrameBytes < 64 || l.MaxAVPs < 17 || l.MaxDepth < 3 || (name != "combined-sms-client-nul-v1" && name != "combined-sms-client-mismatch-v1") || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	for i, c := range [3]uint32{873, 2000, 2018} {
		if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
			return nil, 0, false
		}
	}
	d := wire.NewDecoder(src, l)
	top, ok := interfaceFields(d, avp.Key{Code: 416}, avp.Key{Code: 415}, avp.Key{Code: 455}, avp.Key{Code: 458}, f.Path[0])
	if !ok {
		return nil, 0, false
	}
	for _, v := range top[:3] {
		if v.Flags != 64 || len(v.Data) != 4 {
			return nil, 0, false
		}
	}
	kind, number := binary.BigEndian.Uint32(top[0].Data), binary.BigEndian.Uint32(top[1].Data)
	if !((kind == 1 || kind == 4) && number == 0 || kind == 3 && number == 1) || binary.BigEndian.Uint32(top[2].Data) != 1 || top[3].Flags != 64 || top[4].Flags != 192 {
		return nil, 0, false
	}
	service, e := d.Child(top[4].Data)
	if e != nil {
		return nil, 0, false
	}
	middle, ok := interfaceFields(service, f.Path[1])
	if !ok || middle[0].Flags != 192 {
		return nil, 0, false
	}
	area := middle[0].Data
	d, e = service.Child(area)
	if e != nil {
		return nil, 0, false
	}

	var seen [14]bool
	var leaf, smsc wire.View
	// The trusted request callback has passed earlier subscription/equipment rules.
	// Close the SMS group; nested interface payloads remain opaque.
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
			return nil, 0, false
		}
		if v.Key.Code == 2018 {
			leaf = v
		}
		if v.Key.Code == 2017 {
			smsc = v
		}
		if (i == 0 || i == 3) && binary.BigEndian.Uint32(v.Data) != 0 {
			return nil, 0, false
		}
	}
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 8, 9, 10, 11} {
		if !seen[i] {
			return nil, 0, false
		}
	}
	nul := bytes.IndexByte(leaf.Data, 0) >= 0
	if name == "combined-sms-client-nul-v1" {
		if !nul {
			return nil, 0, false
		}
	} else if nul || bytes.Equal(leaf.Data, smsc.Data) {
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
