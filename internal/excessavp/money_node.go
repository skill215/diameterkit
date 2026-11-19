// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated money-template SMS-Node value diagnostics.
// SPDX-AI-Date: 2026-11-19

package excessavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type moneyNodeEvidence interface{ DiameterMoneyNodeValueV1() string }

func quoteMoneyNode(src []byte, cause error, proof moneyNodeEvidence, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 40 || limits.MaxFieldBytes < 40 || limits.MaxFrameBytes < 68 || limits.MaxAVPs < 4 || limits.MaxDepth < 3 || proof.DiameterMoneyNodeValueV1() != "combined-money-sms-node-v1" || !errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &field) || len(field.Path) != 3 {
		return nil, 0, false
	}
	path := [3]avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2016}}
	for i, k := range path {
		if field.Path[i] != k {
			return nil, 0, false
		}
	}
	// Family, sequence, action and unit presence are raw counterparts. All scans
	// use one source counter; subscription and money payload validation stays in
	// the trusted decoder that produced the evidence.
	d := wire.NewDecoder(src, limits)
	top, ok := unitFields(d, []avp.Key{{Code: 455}, {Code: 458}}, avp.Key{Code: 416}, avp.Key{Code: 415}, avp.Key{Code: 436}, avp.Key{Code: 456}, path[0])
	if !ok {
		return nil, 0, false
	}
	for i := 0; i < 3; i++ {
		if top[i].Flags != 64 || len(top[i].Data) != 4 {
			return nil, 0, false
		}
	}
	if binary.BigEndian.Uint32(top[0].Data) != 4 || binary.BigEndian.Uint32(top[1].Data) != 0 || binary.BigEndian.Uint32(top[2].Data) > 2 || top[3].Flags != 64 || top[4].Flags != 192 {
		return nil, 0, false
	}
	units, e := d.Child(top[3].Data)
	if e != nil {
		return nil, 0, false
	}
	u, ok := unitFields(units, []avp.Key{{Code: 446}, {VendorID: 10415, Code: 872}}, avp.Key{Code: 437})
	if !ok || u[0].Flags != 64 {
		return nil, 0, false
	}
	service, e := d.Child(top[4].Data)
	if e != nil {
		return nil, 0, false
	}
	middle, ok := interfaceFields(service, path[1])
	if !ok || middle[0].Flags != 192 {
		return nil, 0, false
	}
	sms, e := service.Child(middle[0].Data)
	if e != nil {
		return nil, 0, false
	}
	forbidden := [7]avp.Key{}
	for i, c := range [7]uint32{2008, 2010, 2017, 2011, 2015, 2014, 2012} {
		forbidden[i] = avp.Key{VendorID: 10415, Code: c}
	}
	leaf, ok := unitFields(sms, forbidden[:], path[2])
	if !ok || leaf[0].Flags != 192 || len(leaf[0].Data) != 4 || binary.BigEndian.Uint32(leaf[0].Data) == 2 {
		return nil, 0, false
	}
	// Allocate only after complete source scans and output limits pass. Copy no
	// SMS addresses, unit data, or unrelated children.
	out := make([]byte, 40)
	for i, k := range path[:2] {
		off := i * 12
		binary.BigEndian.PutUint32(out[off:], k.Code)
		out[off+4] = 192
		wire.Put24(out[off+5:off+8], uint32(40-off))
		binary.BigEndian.PutUint32(out[off+8:], 10415)
	}
	copy(out[24:], middle[0].Data[leaf[0].Offset:leaf[0].Offset+16])
	return out, 5004, true
}
