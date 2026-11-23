// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded money ClientAddressWire value diagnostics without content reinterpretation.
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

type moneyClientEvidence interface{ DiameterMoneyClientValueV1() string }

func quoteMoneyClient(src []byte, cause error, proof moneyClientEvidence, l wire.Limits, budget int) ([]byte, uint32, bool) {
	var f *validation.FieldError
	if l.Validate() != nil || len(src) > l.MaxFrameBytes-20 || budget < 36 || l.MaxFieldBytes < 36 || l.MaxFrameBytes < 64 || l.MaxAVPs < 4 || l.MaxDepth < 3 || proof.DiameterMoneyClientValueV1() != "combined-money-client-wire-v1" || !errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) || !errors.As(cause, &f) || len(f.Path) != 3 {
		return nil, 0, false
	}
	for i, c := range [3]uint32{873, 2000, 2018} {
		if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
			return nil, 0, false
		}
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
	if id != 1 || zone != 0 {
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

	leaf := fields[1]
	if bytes.Equal(leaf.Data, []byte{0, 1, 127, 0, 0, 1}) {
		return nil, 0, false
	}
	// Parent scans validate the complete padded source span. Check the whole
	// variable-size quote before allocation; never truncate client bytes.
	leafSize := (12 + len(leaf.Data) + 3) &^ 3
	size := 24 + leafSize
	if size > budget || size > l.MaxFieldBytes || size+28 > l.MaxFrameBytes {
		return nil, 0, false
	}
	out := make([]byte, size)
	for i, c := range [2]uint32{873, 2000} {
		off := i * 12
		binary.BigEndian.PutUint32(out[off:], c)
		out[off+4] = 192
		wire.Put24(out[off+5:off+8], uint32(size-off))
		binary.BigEndian.PutUint32(out[off+8:], 10415)
	}
	copy(out[24:], info[0].Data[leaf.Offset:leaf.Offset+leafSize])
	return out, 5004, true
}
