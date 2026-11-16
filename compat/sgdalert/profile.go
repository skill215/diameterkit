// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package sgdalert supplies the explicit legacy-sgd-alert-v1 deployment binding.
// It binds only ALR/ALA to application 16777313, using the implemented S6c field
// and context rules. It is not a standard SGd command or a general app-ID alias.
package sgdalert

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

const ApplicationID uint32 = 16777313
const Version = "legacy-sgd-alert-v1.1"

// ALR has the S6c ALR fields, with a distinct, fixed command identity. Use
// s6c.AlertRole, s6c.RouterAlert or the selected S6c IMS/T4 context for evidence.
// Conversion from *s6c.ALR borrows its fields; Clone/completion own their output.
type ALR s6c.ALR
type ALA s6c.ALA

func (*ALR) CommandKey() pdu.CommandKey {
	return pdu.CommandKey{ApplicationID: ApplicationID, Code: 8388648, Request: true}
}
func (*ALA) CommandKey() pdu.CommandKey {
	return pdu.CommandKey{ApplicationID: ApplicationID, Code: 8388648}
}

// Profile must be explicitly supplied to schema.Compile and authorized in the
// Node's command/source/destination bindings. It adds no replay permission.
// The digest includes the exact S6c definition and version, so changes to the
// delegated field/context rules also change the resulting Bundle fingerprint.
func Profile() (schema.Profile, error) {
	p, err := s6c.Profile()
	if err != nil {
		return schema.Profile{}, err
	}
	b, err := schema.Compile(p)
	if err != nil {
		return schema.Profile{}, err
	}
	r, ok := b.Lookup((*s6c.ALR)(nil).CommandKey())
	if !ok {
		return schema.Profile{}, schema.ErrDefinition
	}
	a, ok := b.Lookup((*s6c.ALA)(nil).CommandKey())
	if !ok {
		return schema.Profile{}, schema.ErrDefinition
	}
	sum := sha256.Sum256([]byte(Version + "\n" + p.Name() + "\n" + p.Version() + "\n" + p.Digest()))
	return schema.NewProfile("legacy-sgd-alert-v1", Version, hex.EncodeToString(sum[:]), bind(r), bind(a))
}

// These conversions are private and fixed to two named types. They do not
// inspect or rewrite wire headers, clone borrowed bodies, or retain state.
func standard(b pdu.Body, request bool) (pdu.Body, error) {
	if request {
		if v, ok := b.(*ALR); ok && v != nil {
			return (*s6c.ALR)(v), nil
		}
	} else {
		if v, ok := b.(*ALA); ok && v != nil {
			return (*s6c.ALA)(v), nil
		}
	}
	return nil, validation.ErrType
}
func legacy(b pdu.Body, err error) (pdu.Body, error) {
	if err != nil {
		return nil, err
	}
	switch v := b.(type) {
	case *s6c.ALR:
		return (*ALR)(v), nil
	case *s6c.ALA:
		return (*ALA)(v), nil
	default:
		return nil, validation.ErrType
	}
}

func bind(src schema.Command) schema.Command {
	c := src
	c.Key.ApplicationID = ApplicationID
	c.Decode = func(d *wire.Decoder) (pdu.Body, error) { return legacy(src.Decode(d)) }
	c.Encode = func(e *wire.Encoder, b pdu.Body) error {
		v, err := standard(b, src.Key.Request)
		if err != nil {
			return err
		}
		return src.Encode(e, v)
	}
	c.Validate = func(b pdu.Body, mode validation.Mode) error {
		v, err := standard(b, src.Key.Request)
		if err != nil {
			return err
		}
		return src.Validate(v, mode)
	}
	c.Clone = func(b pdu.Body) (pdu.Body, error) {
		v, err := standard(b, src.Key.Request)
		if err != nil {
			return nil, err
		}
		return legacy(src.Clone(v))
	}
	c.Envelope = func(b pdu.Body) (pdu.Envelope, error) {
		v, err := standard(b, src.Key.Request)
		if err != nil {
			return pdu.Envelope{}, err
		}
		return src.Envelope(v)
	}
	if src.CompleteRequest != nil {
		c.CompleteRequest = func(b pdu.Body, env pdu.RequestEnvelope) (pdu.Body, error) {
			v, err := standard(b, true)
			if err != nil {
				return nil, err
			}
			return legacy(src.CompleteRequest(v, env))
		}
	}
	if src.CompleteAnswer != nil {
		c.CompleteAnswer = func(b pdu.Body, env pdu.AnswerEnvelope) (pdu.Body, error) {
			v, err := standard(b, false)
			if err != nil {
				return nil, err
			}
			return legacy(src.CompleteAnswer(v, env))
		}
	}
	if src.CheckContext != nil {
		c.CheckContext = func(r, a pdu.Body, ctx validation.Context, mode validation.Mode) error {
			r, err := standard(r, true)
			if err != nil {
				return err
			}
			if a != nil {
				a, err = standard(a, false)
				if err != nil {
					return err
				}
			}
			return src.CheckContext(r, a, ctx, mode)
		}
	}
	if src.CheckAnswer != nil {
		c.CheckAnswer = func(r, a pdu.Body) error {
			r, err := standard(r, true)
			if err != nil {
				return err
			}
			a, err = standard(a, false)
			if err != nil {
				return err
			}
			return src.CheckAnswer(r, a)
		}
	}
	if src.Error != nil {
		e := *src.Error
		// The shared E body carries its own command key. Validate a shallow copy
		// under the source grammar; decoded output keeps the actual legacy key.
		e.Validate = func(b pdu.Body, mode validation.Mode) error {
			v, ok := b.(*pdu.ErrorAnswer)
			if !ok || v == nil || v.Command != c.Key && !(mode == validation.Build && v.Command == (pdu.CommandKey{})) {
				return validation.ErrType
			}
			copy := *v
			if copy.Command != (pdu.CommandKey{}) {
				copy.Command = src.Key
			}
			return src.Error.Validate(&copy, mode)
		}
		e.Decode = func(d *wire.Decoder) (pdu.Body, error) {
			b, err := src.Error.Decode(d)
			if err != nil {
				return nil, err
			}
			v, ok := b.(*pdu.ErrorAnswer)
			if !ok || v == nil {
				return nil, validation.ErrType
			}
			v.Command = c.Key
			return v, nil
		}
		c.Error = &e
	}
	// Canonical sees the untouched header (including Application-Id).
	// DecodeBound is identical: pointer conversions add no owned allocation.
	return c
}
