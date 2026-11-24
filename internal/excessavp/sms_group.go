// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Share bounded structural checks and exact SMS diagnostic quoting.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"encoding/binary"
	"github.com/skill215/diameterkit/avp"

	"github.com/skill215/diameterkit/wire"
)

// scanSMSGroup consumes the caller's child decoder so ancestor and leaf reads
// share one budget. Nested interface payloads remain opaque. Each consumer
// checks its own rules as fields arrive, after the common structural checks.
func scanSMSGroup(d *wire.Decoder, extraRequired []uint32, visit func(wire.View) bool) bool {
	const group = 2000
	var seen [19]bool
	for {
		v, more, err := d.Next()
		if err != nil {
			return false
		}
		if !more {
			break
		}
		code := v.Key.Code
		switch code {
		case 2001, 2002, 2007, 2008, 2009, 2010, 2011, 2012, 2013, 2014, 2015, 2016, 2017, 2018:
		default:
			return false
		}
		if v.Key.VendorID != 10415 || v.Flags != 192 || seen[code-group] {
			return false
		}
		seen[code-group] = true
		switch code {
		case 2016, 2001, 2007, 2011, 2012:
			if len(v.Data) != 4 {
				return false
			}
		}
		if !visit(v) {
			return false
		}
	}
	for _, code := range [...]uint32{2016, 2001, 2002, 2007, 2009, 2013} {
		if !seen[code-group] {
			return false
		}
	}
	for _, code := range extraRequired {
		if !seen[code-group] {
			return false
		}
	}
	return true
}

func smsAddress(code uint32) bool { return code == 2008 || code == 2010 || code == 2017 }

func quoteServiceLeaf(area []byte, leaf wire.View, group, result uint32, l wire.Limits, budget int) ([]byte, uint32, bool) {
	leafSize := (12 + len(leaf.Data) + 3) &^ 3
	size := 24 + leafSize
	if size > budget || size > l.MaxFieldBytes || size+28 > l.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	for i, c := range [2]uint32{873, group} {
		off := 12 * i
		binary.BigEndian.PutUint32(out[off:], c)
		out[off+4] = 192
		wire.Put24(out[off+5:off+8], uint32(size-off))
		binary.BigEndian.PutUint32(out[off+8:], 10415)
	}
	copy(out[24:], area[leaf.Offset:leaf.Offset+leafSize])
	return out, result, true
}

// Selectors used by the bounded SMS request diagnostics.
type smsRequestRule uint32

const (
	smsClientRule  smsRequestRule = 2018
	ecurOriginator smsRequestRule = 2008
	initialStatus  smsRequestRule = 2014
)

// Select the SMS group using the original decoder budget. ECUR also excludes
// Requested-Action and requires the corresponding INITIAL or TERMINATION sequence.
func smsRequestGroup(src []byte, l wire.Limits, rule smsRequestRule) (*wire.Decoder, []byte, uint32) {
	var forbidden avp.Key
	if rule != smsClientRule {
		forbidden = avp.Key{Code: 436}
	}
	d := wire.NewDecoder(src, l)
	keys := []avp.Key{{Code: 416}, {Code: 415}, {Code: 455}, {Code: 458}, {VendorID: 10415, Code: 873}, {Code: 456}}
	if rule != initialStatus {
		keys = keys[:5]
	}
	var top [6]wire.View
	if !readInterfaceFields(d, forbidden, top[:], keys) {
		return nil, nil, 0
	}
	for _, v := range top[:3] {
		if v.Flags != 64 || len(v.Data) != 4 {
			return nil, nil, 0
		}
	}
	kind, number := binary.BigEndian.Uint32(top[0].Data), binary.BigEndian.Uint32(top[1].Data)
	if (rule != smsClientRule && kind == 4) || !((kind == 1 || kind == 4) && number == 0 || kind == 3 && number == 1) || binary.BigEndian.Uint32(top[2].Data) != 1 || top[3].Flags != 64 || top[4].Flags != 192 {
		return nil, nil, 0
	}
	if rule == initialStatus {
		if kind != 1 || top[5].Flags != 64 {
			return nil, nil, 0
		}
		mscc, e := d.Child(top[5].Data)
		if e != nil {
			return nil, nil, 0
		}
		_, value, ok := requestedUnitLeaf(mscc)
		if !ok || binary.BigEndian.Uint64(value.Data) != 1 {
			return nil, nil, 0
		}
	}
	service, e := d.Child(top[4].Data)
	if e != nil {
		return nil, nil, 0
	}
	middle, ok := interfaceFields(service, avp.Key{VendorID: 10415, Code: 2000})
	if !ok || middle[0].Flags != 192 {
		return nil, nil, 0
	}
	area := middle[0].Data
	d, e = service.Child(area)
	if e != nil {
		return nil, nil, 0
	}

	return d, area, kind
}
