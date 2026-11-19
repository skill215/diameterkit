// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated Service-Identifier and Timezone value diagnostics.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type moneyParameterEvidence interface{ DiameterMoneyParameterValueV1() string }

// Seven fixed root counterparts and three forbidden markers share one complete
// scan. Keep this rule-specific bound separate from other diagnostics.
func moneyParameterRoot(d *wire.Decoder) ([7]wire.View, bool) {
	keys := [7]avp.Key{{Code: 416}, {Code: 415}, {Code: 436}, {Code: 456}, {VendorID: 10415, Code: 873}, {Code: 439}, {VendorID: 10425, Code: 23}}
	var values [7]wire.View
	var seen [7]bool
	for {
		v, more, e := d.Next()
		if e != nil {
			return values, false
		}
		if !more {
			break
		}
		if v.Key == (avp.Key{Code: 455}) || v.Key == (avp.Key{Code: 458}) || v.Key == (avp.Key{Code: 1292}) {
			return values, false
		}
		for i, k := range keys {
			if v.Key == k {
				if seen[i] {
					return values, false
				}
				seen[i] = true
				values[i] = v
			}
		}
	}
	for _, yes := range seen {
		if !yes {
			return values, false
		}
	}
	return values, true
}
func quoteMoneyParameter(src []byte, cause error, proof moneyParameterEvidence, l wire.Limits, budget int) ([]byte, uint32, bool) {
	selected, size := 5, 12
	switch proof.DiameterMoneyParameterValueV1() {
	case "combined-money-service-identifier-v1":
	case "combined-money-timezone-v1":
		selected, size = 6, 16
	default:
		return nil, 0, false
	}
	key := avp.Key{Code: 439}
	if selected == 6 {
		key = avp.Key{VendorID: 10425, Code: 23}
	}
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < size || l.MaxFieldBytes < size || l.MaxFrameBytes < size+28 || l.MaxAVPs < 2 || l.MaxDepth < 1 || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &f) || len(f.Path) != 1 || f.Path[0] != key {
		return nil, 0, false
	}
	d := wire.NewDecoder(src, l)
	root, ok := moneyParameterRoot(d)
	if !ok {
		return nil, 0, false
	}
	for _, i := range []int{0, 1, 2, 5, 6} {
		flags := byte(64)
		if i == 6 {
			flags = 192
		}
		if root[i].Flags != flags || len(root[i].Data) != 4 {
			return nil, 0, false
		}
	}
	action := binary.BigEndian.Uint32(root[2].Data)
	if binary.BigEndian.Uint32(root[0].Data) != 4 || binary.BigEndian.Uint32(root[1].Data) != 0 || (action != 0 && action != 2) || root[3].Flags != 64 || root[4].Flags != 192 {
		return nil, 0, false
	}
	id, zone := binary.BigEndian.Uint32(root[5].Data), binary.BigEndian.Uint32(root[6].Data)
	if selected == 5 && (id == 1 || zone != 0) || selected == 6 && (id != 1 || zone == 0) {
		return nil, 0, false
	}
	units, e := d.Child(root[3].Data)
	if e != nil {
		return nil, 0, false
	}
	u, ok := unitFields(units, []avp.Key{{Code: 446}, {VendorID: 10415, Code: 872}}, avp.Key{Code: 437})
	if !ok || u[0].Flags != 64 {
		return nil, 0, false
	}
	requested, e := units.Child(u[0].Data)
	if e != nil {
		return nil, 0, false
	}
	money, ok := unitFields(requested, []avp.Key{{Code: 420}}, avp.Key{Code: 413})
	if !ok || money[0].Flags != 64 {
		return nil, 0, false
	}
	service, e := d.Child(root[4].Data)
	if e != nil {
		return nil, 0, false
	}
	info, ok := interfaceFields(service, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: 877})
	if !ok || info[0].Flags != 192 || info[1].Flags != 192 {
		return nil, 0, false
	}
	sms, e := service.Child(info[0].Data)
	if e != nil {
		return nil, 0, false
	}
	forbidden := [7]avp.Key{}
	for i, c := range [7]uint32{2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		forbidden[i] = avp.Key{VendorID: 10415, Code: c}
	}
	fields, ok := unitFields(sms, forbidden[:], avp.Key{VendorID: 10415, Code: 2016}, avp.Key{VendorID: 10415, Code: 2018})
	if !ok || fields[0].Flags != 192 || len(fields[0].Data) != 4 || binary.BigEndian.Uint32(fields[0].Data) != 2 || fields[1].Flags != 192 {
		return nil, 0, false
	}
	// Client bytes and MMS/money contents belong to other rules. Do not pull later
	// constraints forward. Only the original numeric leaf survives this copy.
	out := make([]byte, size)
	copy(out, src[root[selected].Offset:root[selected].Offset+size])
	return out, 5004, true
}
