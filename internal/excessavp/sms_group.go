// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Share bounded structural checks and exact SMS diagnostic quoting.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"encoding/binary"

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

func quoteSMSLeaf(area []byte, leaf wire.View, l wire.Limits, budget int) ([]byte, uint32, bool) {
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
