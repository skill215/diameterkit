// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded independently checked group-constraint diagnostics.
// SPDX-AI-Date: 2026-11-17

// Package groupavp quotes selected handwritten conditional and omission rules.
package groupavp

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// External generated packages implement this without internal imports. Metadata
// retains no request bytes; callbacks are trusted schema code.
type evidence interface {
	DiameterGroupAVPV1() (avp.Key, byte, int, string, error)
}

// Quote rechecks a selected constraint against the original AVP area, returning
// owned bytes and minimal ancestors. Bare sentinels supply no evidence.
func Quote(src []byte, cause error, limits wire.Limits, maxBytes int) ([]byte, uint32, bool) {
	var field *validation.FieldError
	var proof evidence
	if limits.Validate() != nil || len(src) > limits.MaxFrameBytes-20 || maxBytes < 0 || !errors.As(cause, &field) || !errors.As(cause, &proof) {
		return nil, 0, false
	}
	var parents [64]wire.View
	depth, headers := 0, 0
	d := wire.NewDecoder(src, limits)
	for {
		key, flags, offset, rule, child := proof.DiameterGroupAVPV1()
		if child == nil {
			if flags != 0 || offset != 0 {
				return nil, 0, false
			}
			group := uint32(260)
			pathLen := depth
			switch rule {
			case "vsai-both", "vsai-neither":
				if key != (avp.Key{}) {
					return nil, 0, false
				}
			case "proxy-host-missing", "proxy-state-missing":
				group = 284
				pathLen++
				expected := uint32(280)
				if rule == "proxy-state-missing" {
					expected = 33
				}
				if key != (avp.Key{Code: expected}) || depth >= len(field.Path) || field.Path[depth] != key {
					return nil, 0, false
				}
			default:
				return nil, 0, false
			}
			if len(field.Path) != pathLen || depth > 0 && parents[depth-1].Key != (avp.Key{Code: group}) {
				return nil, 0, false
			}
			code := uint32(5005)
			if rule == "vsai-both" {
				code = 5009
				if !errors.Is(cause, validation.ErrDuplicate) {
					return nil, 0, false
				}
			} else if !errors.Is(cause, validation.ErrMissing) {
				return nil, 0, false
			}
			leaves, ok := check(d, rule, limits)
			if !ok {
				return nil, 0, false
			}
			payload, count := 24, 2
			if group == 284 {
				count = 1
				payload = 12
				if key.Code == 33 {
					payload = 8
				}
			}
			size := headers + payload
			if depth+1 > limits.MaxDepth || depth+count+1 > limits.MaxAVPs || size > maxBytes || size > limits.MaxFieldBytes || size > limits.MaxFrameBytes-28 {
				return nil, 0, false
			}
			out := make([]byte, size)
			pos := 0
			for i := 0; i < depth; i++ {
				p := parents[i]
				pos += header(out[pos:], p.Key, p.Flags, size-pos)
			}
			if rule == "vsai-both" {
				for _, v := range leaves {
					copy(out[pos:], src[v.Offset:v.Offset+12])
					pos += 12
				}
			} else if group == 260 {
				header(out[pos:], avp.Key{Code: 258}, avp.MandatoryFlag, 12)
				header(out[pos+12:], avp.Key{Code: 259}, avp.MandatoryFlag, 12)
			} else {
				length := 8
				if key.Code == 280 {
					length = 9
				}
				header(out[pos:], key, avp.MandatoryFlag, length)
			}
			return out, code, true
		}
		if rule != "" || depth >= len(parents) || depth >= len(field.Path) || depth >= limits.MaxDepth || key != field.Path[depth] || flags&0x1f != 0 || offset < 0 || offset%4 != 0 {
			return nil, 0, false
		}
		var selected wire.View
		seen := 0
		for {
			v, ok, err := d.Next()
			if err != nil || !ok || v.Offset > offset {
				return nil, 0, false
			}
			if v.Key == key {
				seen++
			}
			if v.Offset == offset {
				if v.Key != key || v.Flags != flags || seen > limits.MaxRepeated {
					return nil, 0, false
				}
				selected = v
				break
			}
		}
		var ok bool
		proof, ok = child.(evidence)
		if !ok {
			return nil, 0, false
		}
		parents[depth] = wire.View{Key: key, Flags: flags}
		depth++
		headers += 8
		if key.VendorID != 0 {
			headers += 4
		}
		var err error
		d, err = d.Child(selected.Data)
		if err != nil {
			return nil, 0, false
		}
		src = selected.Data
	}
}

// Check the whole selected group, including trailing framing and extra budgets.
func check(d *wire.Decoder, rule string, limits wire.Limits) ([2]wire.View, bool) {
	var leaves [2]wire.View
	vendor, auth, acct, host, state, extras, n := 0, 0, 0, 0, 0, 0, 0
	vsai := rule == "vsai-both" || rule == "vsai-neither"
	for {
		v, ok, err := d.Next()
		if err != nil {
			return leaves, false
		}
		if !ok {
			break
		}
		known := false
		if v.Key.VendorID == 0 {
			if vsai {
				switch v.Key.Code {
				case 266:
					vendor++
					known = true
				case 258:
					auth++
					known = true
				case 259:
					acct++
					known = true
				}
				if known && len(v.Data) != 4 {
					return leaves, false
				}
				if known && v.Key.Code != 266 {
					if n >= len(leaves) {
						return leaves, false
					}
					leaves[n] = v
					n++
				}
			} else {
				switch v.Key.Code {
				case 280:
					host++
					known = true
					if len(v.Data) > 253 || avp.Identity(string(v.Data)) != nil {
						return leaves, false
					}
				case 33:
					state++
					known = true
				}
			}
		}
		if vendor > 1 || auth > 1 || acct > 1 || host > 1 || state > 1 {
			return leaves, false
		}
		if !known {
			extras++
			if v.Flags&avp.MandatoryFlag != 0 || extras > limits.MaxRepeated {
				return leaves, false
			}
		}
	}
	switch rule {
	case "vsai-both":
		return leaves, vendor == 1 && auth == 1 && acct == 1
	case "vsai-neither":
		return leaves, vendor == 1 && auth == 0 && acct == 0
	case "proxy-host-missing":
		return leaves, host == 0
	case "proxy-state-missing":
		return leaves, host == 1 && state == 0
	}
	return leaves, false
}

func header(dst []byte, key avp.Key, flags byte, length int) int {
	binary.BigEndian.PutUint32(dst, key.Code)
	dst[4] = flags
	wire.Put24(dst[5:8], uint32(length))
	if key.VendorID != 0 {
		binary.BigEndian.PutUint32(dst[8:], key.VendorID)
		return 12
	}
	return 8
}
