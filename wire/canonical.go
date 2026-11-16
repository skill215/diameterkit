// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package wire

import (
	"cmp"
	"encoding/binary"
	"slices"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
)

// CanonicalField is a compiled field rule, not a runtime ignore list. Callers
// must validate the original request before applying these comparison rules.
// A nil Group denotes an opaque leaf; grouped fields preserve child sequence.
type CanonicalField struct {
	PreserveMandatory bool // Only compiled fields with M-dependent semantics.
	Key               avp.Key
	Identity          bool
	Group             []CanonicalField
}

// CanonicalRequest returns an owned CanonicalV1 comparison sequence. Top-level
// known keys sort stably between unknown barriers; nested groups stay ordered.
// Header identity excludes HBH/T/E2E. E2E belongs to the duplicate lookup key.
func CanonicalRequest(src []byte, l Limits, maxBytes int, fields []CanonicalField, ignoreVSAI bool) ([]byte, error) {
	h, n, err := Header(src, l)
	if err != nil {
		return nil, err
	}
	if n != len(src) || h.Flags&pdu.RequestFlag == 0 || maxBytes < 12 {
		return nil, ErrFrame
	}
	out := make([]byte, 12)
	out[0] = 1
	out[1] = h.Flags & (pdu.RequestFlag | pdu.ProxiableFlag | pdu.ErrorFlag)
	binary.BigEndian.PutUint32(out[4:8], h.ApplicationID)
	binary.BigEndian.PutUint32(out[8:12], h.CommandCode)
	body, err := canonicalAVPs(NewDecoder(src[20:], l), fields, true, ignoreVSAI, maxBytes-12)
	if err != nil {
		return nil, err
	}
	return append(out, body...), nil
}

type canonicalPart struct {
	key   avp.Key
	data  []byte
	known bool
}

func canonicalAVPs(d *Decoder, fields []CanonicalField, top, ignoreVSAI bool, limit int) ([]byte, error) {
	var parts []canonicalPart
	total := 0
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if top && v.Key.VendorID == 0 && (v.Key.Code == 284 || v.Key.Code == 282 || ignoreVSAI && v.Key.Code == 260) {
			continue
		}
		var rule *CanonicalField
		for i := range fields {
			if fields[i].Key == v.Key {
				rule = &fields[i]
				break
			}
		}
		if limit-total < 13 {
			return nil, ErrLimit
		}
		data := v.Data
		if rule != nil && rule.Group != nil {
			child, err := d.Child(v.Data)
			if err != nil {
				return nil, err
			}
			data, err = canonicalAVPs(child, rule.Group, false, false, limit-total-13)
			if err != nil {
				return nil, err
			}
		}
		if len(data) > limit-total-13 {
			return nil, ErrLimit
		}
		b := make([]byte, 13+len(data))
		binary.BigEndian.PutUint32(b[:4], v.Key.VendorID)
		binary.BigEndian.PutUint32(b[4:8], v.Key.Code)
		b[8] = v.Flags & 0xe0
		if rule != nil && !rule.PreserveMandatory {
			b[8] &^= avp.MandatoryFlag
		}
		binary.BigEndian.PutUint32(b[9:13], uint32(len(data)))
		copy(b[13:], data)
		if rule != nil && rule.Identity {
			for i := 13; i < len(b); i++ {
				if b[i] >= 'A' && b[i] <= 'Z' {
					b[i] += 'a' - 'A'
				}
			}
		}
		total += len(b)
		parts = append(parts, canonicalPart{v.Key, b, rule != nil})
	}
	if top {
		for start := 0; start < len(parts); {
			if !parts[start].known {
				start++
				continue
			}
			end := start + 1
			for end < len(parts) && parts[end].known {
				end++
			}
			slices.SortStableFunc(parts[start:end], func(a, b canonicalPart) int {
				if a.key.VendorID != b.key.VendorID {
					return cmp.Compare(a.key.VendorID, b.key.VendorID)
				}
				return cmp.Compare(a.key.Code, b.key.Code)
			})
			start = end
		}
	}
	out := make([]byte, 0, total)
	for _, p := range parts {
		out = append(out, p.data...)
	}
	return out, nil
}
