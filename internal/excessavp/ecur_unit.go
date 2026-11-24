// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group exclusions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// The reviewed group has an eight-byte header and one complete UInt64 AVP.
// Its enclosing MSCC contributes only its own header to Failed-AVP.
func quoteExcludedECURUnit(src []byte, cause error, l wire.Limits, budget int) ([]byte, uint32, bool) {
	const size = 32
	var f *validation.FieldError
	if l.Validate() != nil || budget < size || l.MaxFieldBytes < size || l.MaxFrameBytes < 28+size || len(src) > l.MaxFrameBytes-20 || l.MaxDepth < 3 || !errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &f) || len(f.Path) != 2 || f.Path[0] != (avp.Key{Code: 456}) {
		return nil, 0, false
	}
	d := wire.NewDecoder(src, l)
	var top [6]wire.View
	if !readInterfaceFields(d, avp.Key{Code: 436}, top[:], []avp.Key{{Code: 416}, {Code: 415}, {Code: 455}, {Code: 458}, {Code: 456}, {VendorID: 10415, Code: 873}}) {
		return nil, 0, false
	}
	for _, v := range top[:3] {
		if v.Flags != 64 || len(v.Data) != 4 {
			return nil, 0, false
		}
	}
	kind := binary.BigEndian.Uint32(top[0].Data)
	number := binary.BigEndian.Uint32(top[1].Data)
	if !(kind == 1 && number == 0 && f.Path[1] == (avp.Key{Code: 446}) || kind == 3 && number == 1 && f.Path[1] == (avp.Key{Code: 437})) || binary.BigEndian.Uint32(top[2].Data) != 1 || top[3].Flags != 64 || top[4].Flags != 64 || top[5].Flags != 192 {
		return nil, 0, false
	}
	service, e := d.Child(top[5].Data)
	if e != nil {
		return nil, 0, false
	}
	middle, ok := interfaceFields(service, avp.Key{VendorID: 10415, Code: 2000})
	if !ok || middle[0].Flags != 192 {
		return nil, 0, false
	}
	sms, e := service.Child(middle[0].Data)
	if e != nil {
		return nil, 0, false
	}
	if _, ok := scanECURExclusionSMS(sms, kind, 0); !ok {
		return nil, 0, false
	}
	mscc, e := d.Child(top[4].Data)
	if e != nil {
		return nil, 0, false
	}
	group, ok := excludedUnitGroup(mscc, kind)
	if !ok {
		return nil, 0, false
	}
	out := make([]byte, size)
	binary.BigEndian.PutUint32(out, 456)
	out[4] = 64
	wire.Put24(out[5:8], size)
	copy(out[8:], top[4].Data[group.Offset:group.Offset+24])
	return out, 5008, true
}

// Check every MSCC child, including the required sibling, on one decoder budget.
func excludedUnitGroup(mscc *wire.Decoder, kind uint32) (wire.View, bool) {
	var requested, used, reason bool
	var selected wire.View
	for {
		v, more, e := mscc.Next()
		if e != nil {
			return wire.View{}, false
		}
		if !more {
			break
		}
		switch v.Key {
		case avp.Key{Code: 437}, avp.Key{Code: 446}:
			if v.Flags != 64 || len(v.Data) != 16 {
				return wire.View{}, false
			}
			child, e := mscc.Child(v.Data)
			if e != nil {
				return wire.View{}, false
			}
			leaf, more, e := child.Next()
			if e != nil || !more || leaf.Key != (avp.Key{Code: 417}) || leaf.Flags != 64 || len(leaf.Data) != 8 {
				return wire.View{}, false
			}
			if v.Key.Code == 437 {
				if requested || kind == 1 && binary.BigEndian.Uint64(leaf.Data) != 1 {
					return wire.View{}, false
				}
				requested = true
				if kind == 3 {
					selected = v
				}
			} else {
				if used {
					return wire.View{}, false
				}
				used = true
				if kind == 1 {
					selected = v
				}
			}
		case avp.Key{VendorID: 10415, Code: 872}:
			if kind != 3 || reason || v.Flags != 192 || len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 2 {
				return wire.View{}, false
			}
			reason = true
		default:
			return wire.View{}, false
		}
	}
	return selected, requested && used && (kind != 3 || reason)
}
