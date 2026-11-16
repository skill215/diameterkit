// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package codec provides owned, offline Diameter encoding and decoding.
package codec

import (
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"io"
)

type Limits = wire.Limits
type Config struct{ Limits Limits }
type Codec struct{ engine *codecengine.Engine }

func New(b *schema.Bundle, cfg Config) (*Codec, error) {
	l := cfg.Limits
	if l == (Limits{}) {
		l = wire.DefaultLimits()
	}
	e, err := codecengine.New(b, l)
	if err != nil {
		return nil, err
	}
	return &Codec{e}, nil
}
func (c *Codec) Limits() Limits                                   { return c.engine.Limits() }
func (c *Codec) Decode(src []byte) (pdu.Message, error)           { return c.engine.Decode(src) }
func (c *Codec) Read(r io.Reader) (pdu.Message, error)            { return c.engine.Read(r) }
func (c *Codec) Append(dst []byte, m pdu.Message) ([]byte, error) { return c.engine.Append(dst, m) }
func (c *Codec) ValidateBody(b pdu.Body) error                    { return c.engine.ValidateBody(b) }
func (c *Codec) Validate(m pdu.Message) error                     { return c.engine.Validate(m) }
func (c *Codec) CompleteRequest(b pdu.Body, e pdu.RequestEnvelope) (pdu.Message, error) {
	return c.engine.CompleteRequest(b, e)
}

// Measure returns the validated encoded size without producing an output frame.
func (c *Codec) Measure(m pdu.Message) (int, error)              { return c.engine.Measure(m) }
func (c *Codec) DecodeBound(h pdu.Header, n int) (uint64, error) { return c.engine.DecodeBound(h, n) }
func (c *Codec) Snapshot(m pdu.Message) (pdu.Message, error)     { return c.engine.Snapshot(m) }

// ValidateAnswer performs offline protocol/source association, not connection or completion ownership.
func (c *Codec) ValidateAnswer(request, answer pdu.Message, sources pdu.AnswerSources) error {
	return c.engine.ValidateAnswer(request, answer, sources)
}
func (c *Codec) DecodeAnswer(request pdu.Message, src []byte, sources pdu.AnswerSources) (pdu.Message, error) {
	return c.engine.DecodeAnswer(request, src, sources)
}
func (c *Codec) CompleteAnswer(request pdu.Message, body pdu.Body, origin pdu.Origin) (pdu.Message, error) {
	return c.engine.CompleteAnswer(request, body, origin)
}
func ReadFrame(r io.Reader, maxBytes int) ([]byte, error) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = maxBytes
	if l.MaxFieldBytes > maxBytes {
		l.MaxFieldBytes = maxBytes
	}
	return codecengine.ReadFrame(r, l)
}
func WriteFrame(w io.Writer, b []byte, maxBytes int) (int, error) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = maxBytes
	if l.MaxFieldBytes > maxBytes {
		l.MaxFieldBytes = maxBytes
	}
	return codecengine.WriteFrame(w, b, l)
}

func (c *Codec) CompleteAnswerBound(request pdu.Message, body pdu.Body, origin pdu.Origin) (uint64, error) {
	return c.engine.CompleteAnswerBound(request, body, origin)
}

// Canonical validates the original wire request and returns owned comparison bytes.
// It never normalizes a mutable decoded application view.
func (c *Codec) Canonical(src []byte, maxBytes int) ([]byte, error) {
	return c.engine.Canonical(src, maxBytes)
}

// ValidateContext reports the compiled slice's static/context checks. For answers
// without original request/source evidence, MissingContext names both inputs.
func (c *Codec) ValidateContext(m pdu.Message, ctx validation.Context) validation.Report {
	return c.engine.ValidateContext(m, ctx)
}

// ValidateAnswerContext supplies offline request/source evidence. It does not
// establish live connection ownership or authorize completion of an exchange.
func (c *Codec) ValidateAnswerContext(request, answer pdu.Message, sources pdu.AnswerSources, ctx validation.Context) validation.Report {
	return c.engine.ValidateAnswerContext(request, answer, sources, ctx)
}

// ValidateResponseContext checks an offline response using receive rules for
// the original request and send rules for the answer. CompleteAnswer/Append
// alone validate wire shape and association, not caller-supplied protocol facts.
func (c *Codec) ValidateResponseContext(request, answer pdu.Message, sources pdu.AnswerSources, ctx validation.Context) validation.Report {
	return c.engine.ValidateResponseContext(request, answer, sources, ctx)
}
