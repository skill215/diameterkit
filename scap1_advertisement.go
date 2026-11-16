// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/wire"
)

// SCAP1Advertisement is a peer-scoped handshake policy, not a codec profile or
// application authorization. LocalCapabilities always use standard strict VSAI.
type SCAP1Advertisement uint8

const (
	SCAP1StandardAdvertisement SCAP1Advertisement = iota
	// SCAP1LegacyDualAdvertisement sends the reviewed sms-fw pair: Vendor193 /
	// Acct19302 followed by Vendor193 / Auth19302 / Acct19302. Receive accepts
	// standard capabilities or this exact adjacent pair. The redundant dual group
	// is removed from the owned handshake buffer before standard Base validation;
	// it never grants Auth19302. Other malformed groups still fail standard rules.
	// No ACR/ACA grammar, replay permission or persistent state changes.
	SCAP1LegacyDualAdvertisement
)

// Fixed source layout from DccaClientWrapper.cc and dia-base-commands.cc.
// These byte strings are immutable, including exact Vendor/M flags and order.
const scap1StrictGroup = "\x00\x00\x01\x04\x40\x00\x00\x20" +
	"\x00\x00\x01\x0a\x40\x00\x00\x0c\x00\x00\x00\xc1" +
	"\x00\x00\x01\x03\x40\x00\x00\x0c\x00\x00\x4b\x66"
const scap1DualGroup = "\x00\x00\x01\x04\x40\x00\x00\x2c" +
	"\x00\x00\x01\x0a\x40\x00\x00\x0c\x00\x00\x00\xc1" +
	"\x00\x00\x01\x02\x40\x00\x00\x0c\x00\x00\x4b\x66" +
	"\x00\x00\x01\x03\x40\x00\x00\x0c\x00\x00\x4b\x66"

func validSCAP1LegacyCapabilities(c LocalCapabilities) bool {
	count := 0
	for _, v := range c.VendorSpecificApplicationID {
		if v.VendorID == 193 {
			count++
			if v.AuthApplicationID.Present || !v.AcctApplicationID.Present || v.AcctApplicationID.Value != 19302 || len(v.Extra) != 0 {
				return false
			}
		}
	}
	return count == 1
}

// The scan uses existing bounded wire primitives, not an alternative decoder.
// Count the original VSAIs and their children BEFORE removing anything, so the
// compatibility pair cannot bypass frame, AVP, repeated, field or depth limits.
func scanSCAP1Advertisement(b []byte, l wire.Limits) (strict, dual int, err error) {
	strict, dual = -1, -1
	d := wire.NewDecoder(b[20:], l)
	groups := 0
	for {
		v, ok, e := d.Next()
		if e != nil || !ok {
			return strict, dual, e
		}
		if v.Key != (avp.Key{Code: 260}) {
			continue
		}
		groups++
		if groups > l.MaxRepeated {
			return -1, -1, wire.ErrLimit
		}
		child, e := d.Child(v.Data)
		if e != nil {
			return -1, -1, e
		}
		for {
			_, ok, e := child.Next()
			if e != nil {
				return -1, -1, e
			}
			if !ok {
				break
			}
		}
		at := 20 + v.Offset
		raw := b[at : at+8+len(v.Data)] // vendor zero has an eight-byte header
		switch {
		case bytes.Equal(raw, []byte(scap1StrictGroup)):
			if strict != -1 {
				return -1, -1, ErrProtocol
			}
			strict = at
		case bytes.Equal(raw, []byte(scap1DualGroup)):
			if dual != -1 {
				return -1, -1, ErrProtocol
			}
			dual = at
		}
	}
}

// b is exclusively owned. Receive shrinks in place; send requires pre-reserved
// capacity. Only the exact redundant group changes; ordinary Base processing
// still validates every remaining field and the original header/source relation.
func projectSCAP1Advertisement(b []byte, l wire.Limits, send bool) ([]byte, error) {
	h, size, err := wire.Header(b, l)
	if err != nil {
		return nil, err
	}
	if size != len(b) || h.ApplicationID != 0 || h.CommandCode != 257 {
		return nil, ErrProtocol
	}
	if h.Flags&pdu.ErrorFlag != 0 {
		return b, nil
	}
	strict, dual, err := scanSCAP1Advertisement(b, l)
	if err != nil {
		return nil, err
	}
	if send {
		if strict < 0 || dual >= 0 {
			return nil, ErrProtocol
		}
		n := len(b)
		if n > l.MaxFrameBytes-len(scap1DualGroup) || cap(b)-n < len(scap1DualGroup) {
			return nil, wire.ErrLimit
		}
		at := strict + len(scap1StrictGroup)
		b = b[:n+len(scap1DualGroup)]
		copy(b[at+len(scap1DualGroup):], b[at:n])
		copy(b[at:], scap1DualGroup)
		if _, _, err = scanSCAP1Advertisement(b, l); err != nil {
			return nil, err
		}
	} else if dual >= 0 {
		if strict < 0 || dual != strict+len(scap1StrictGroup) {
			return nil, ErrProtocol
		}
		copy(b[dual:], b[dual+len(scap1DualGroup):])
		b = b[:len(b)-len(scap1DualGroup)]
	}
	wire.Put24(b[1:4], uint32(len(b)))
	return b, nil
}
