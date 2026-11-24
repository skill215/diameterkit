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
	d, area, _ := smsRequestGroup(src, l, false)
	if d == nil {
		return nil, 0, false
	}

	var leaf, smsc wire.View
	ok := scanSMSGroup(d, []uint32{2018, 2010, 2017, 2011, 2015}, func(v wire.View) bool {
		if smsAddress(v.Key.Code) && bytes.IndexByte(v.Data, 0) >= 0 {
			return false
		}
		switch v.Key.Code {
		case 2018:
			leaf = v
		case 2017:
			smsc = v
		case 2016, 2007:
			return binary.BigEndian.Uint32(v.Data) == 0
		}
		return true
	})
	if !ok {
		return nil, 0, false
	}
	nul := bytes.IndexByte(leaf.Data, 0) >= 0
	if name == "combined-sms-client-nul-v1" {
		if !nul {
			return nil, 0, false
		}
	} else if nul || bytes.Equal(leaf.Data, smsc.Data) {
		return nil, 0, false
	}

	return quoteServiceLeaf(area, leaf, 2000, 5004, l, budget)
}
