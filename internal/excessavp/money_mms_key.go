// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded combined money MMS exclusion and acceptance.
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

type moneyMMSKeyExclusion interface{ DiameterMoneyMMSKeyExclusionV1() }

// closedVendorPair requires both named children exactly once and exhausts the
// group, including trailing siblings, before returning either view.
func closedVendorPair(d *wire.Decoder, first, second uint32) ([2]wire.View, bool) {
	var pair [2]wire.View
	for {
		v, more, err := d.Next()
		if err != nil {
			return pair, false
		}
		if !more {
			return pair, pair[0].Key.Code == first && pair[1].Key.Code == second
		}
		if v.Key.VendorID != 10415 || v.Flags != 192 {
			return pair, false
		}
		var i int
		switch v.Key.Code {
		case first:
		case second:
			i = 1
		default:
			return pair, false
		}
		if pair[i].Key.Code != 0 {
			return pair, false
		}
		pair[i] = v
	}
}

func quoteExcludedMoneyMMSKey(src []byte, cause error, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 36 || l.MaxFieldBytes < 36 || l.MaxFrameBytes < 64 || l.MaxDepth < 3 || !errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	for i, c := range [...]uint32{873, 877, 1210} {
		if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
			return nil, 0, false
		}
	}
	g, ok := scanMoneyClientGroups(wire.NewDecoder(src, l))
	if !ok || !validMoneyClientBytes(g.client.Data) {
		return nil, 0, false
	}
	mms, err := g.service.Child(g.mms.Data)
	if err != nil {
		return nil, 0, false
	}
	pair, ok := closedVendorPair(mms, 886, 1210)
	if !ok || bytes.IndexByte(pair[1].Data, 0) >= 0 {
		return nil, 0, false
	}
	addr, err := mms.Child(pair[0].Data)
	if err != nil {
		return nil, 0, false
	}
	fields, ok := closedVendorPair(addr, 899, 897)
	if !ok || len(fields[0].Data) != 4 || binary.BigEndian.Uint32(fields[0].Data) != 1 || bytes.IndexByte(fields[1].Data, 0) >= 0 {
		return nil, 0, false
	}
	return quoteServiceLeaf(g.mms.Data, pair[1], 877, 5008, l, budget)
}
