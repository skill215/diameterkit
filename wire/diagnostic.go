// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package wire

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
)

// Diagnostic only scans provably framed headers, with a fixed set of known
// Grouped keys. Malformed quoted lengths end that scan; they do not invalidate
// the enclosing AVP or authorize resynchronizing past untrusted bytes.
func (d *Decoder) Diagnostic(b []byte, groups []avp.Key) (bool, error) {
	child, err := d.Child(b)
	if err != nil {
		return false, err
	}
	return child.scanDiagnostic(groups)
}
func (d *Decoder) scanDiagnostic(groups []avp.Key) (bool, error) {
	complete := true
	for {
		v, ok, err := d.Next()
		if err != nil {
			if errors.Is(err, ErrLimit) || errors.Is(err, ErrConfig) {
				return false, err
			}
			return false, nil
		}
		if !ok {
			return complete, nil
		}
		for _, k := range groups {
			if k == v.Key {
				valid, err := d.Diagnostic(v.Data, groups)
				if err != nil {
					return false, err
				}
				complete = complete && valid
				break
			}
		}
	}
}

// Diagnostic encodes quoted bytes without applying application value rules,
// while charging recognized inner headers/depth to the enclosing encoder.
func (e *Encoder) Diagnostic(k avp.Key, flags byte, b []byte, groups []avp.Key) {
	e.Bytes(k, flags, b)
	if e.err != nil {
		return
	}
	d := &Decoder{depth: e.depth, limits: e.limits, count: &counter{avps: e.avps}}
	_, e.err = d.Diagnostic(b, groups)
	e.avps = d.count.avps
}
