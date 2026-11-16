package codecengine

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"strings"
)

func (e *Engine) Measure(m pdu.Message) (int, error) { _, n, err := e.measure(m); return n, err }

// DecodeBound is a pre-allocation logical reservation estimate, not an RSS cap.
func (e *Engine) DecodeBound(h pdu.Header, n int) (uint64, error) {
	if n < 20 || n%4 != 0 || n > e.limits.MaxFrameBytes {
		return 0, wire.ErrLimit
	}
	c, err := e.operation(h)
	if err != nil {
		return 0, err
	}
	return c.DecodeBound(n, e.limits)
}

// Snapshot validates/measures before cloning. The result can be mutated by its
// owner; it is never a runtime's private snapshot unless that runtime owns it.
func (e *Engine) Snapshot(m pdu.Message) (pdu.Message, error) {
	c, _, err := e.measure(m)
	if err != nil {
		return pdu.Message{}, err
	}
	b, err := c.Clone(m.Body)
	if err != nil {
		return pdu.Message{}, err
	}
	m.Body = b
	return m, nil
}

func validateSources(s pdu.AnswerSources) error {
	if len(s.Terminals) == 0 || len(s.Terminals)+len(s.ErrorAgents) > 64 {
		return validation.ErrSource
	}
	for _, list := range [][]pdu.Origin{s.Terminals, s.ErrorAgents} {
		for _, v := range list {
			if avp.Identity(v.Host) != nil || avp.Identity(v.Realm) != nil {
				return validation.ErrSource
			}
		}
	}
	return nil
}
func allowed(origin pdu.Origin, list []pdu.Origin) bool {
	for _, v := range list {
		if strings.EqualFold(v.Host, origin.Host) && strings.EqualFold(v.Realm, origin.Realm) {
			return true
		}
	}
	return false
}

// RequestSnapshot is internal, immutable correlation evidence. The owning runtime
// must first select the actual connection/attempt and must arbitrate completion.
// It carries no pending registry, timer, fd, storage hook or completion authority.
type RequestSnapshot struct {
	context  validation.Context
	engine   *Engine
	request  pdu.Message
	envelope pdu.Envelope
	sources  pdu.AnswerSources
	bound    uint64
}

func (e *Engine) RequestBound(m pdu.Message, s pdu.AnswerSources) (uint64, error) {
	return e.requestBound(m, s, validation.Send)
}
func (e *Engine) requestBound(m pdu.Message, s pdu.AnswerSources, mode validation.Mode) (uint64, error) {
	if err := validateSources(s); err != nil {
		return 0, err
	}
	if m.Header.Flags&pdu.RequestFlag == 0 {
		return 0, validation.ErrType
	}
	c, n, err := e.measureMode(m, mode)
	if err != nil {
		return 0, err
	}
	answerKey := c.Key
	answerKey.Request = false
	answer, ok := e.bundle.Lookup(answerKey)
	if !ok || answer.Proxiable != c.Proxiable {
		return 0, validation.ErrUnsupported
	}
	bound, err := c.DecodeBound(n, e.limits)
	if err != nil {
		return 0, err
	}
	extra := uint64(len(s.Terminals)+len(s.ErrorAgents))*1024 + 512
	if bound > ^uint64(0)-extra {
		return 0, wire.ErrLimit
	}
	return bound + extra, nil
}
func (e *Engine) PrepareRequest(m pdu.Message, s pdu.AnswerSources) (*RequestSnapshot, error) {
	return e.prepareRequest(m, s, validation.Send)
}

// Received request evidence is validated under Receive rules, never re-sent.
// This allows preserved receive-only flag bits without weakening outbound checks.
func (e *Engine) prepareRequest(m pdu.Message, s pdu.AnswerSources, mode validation.Mode) (*RequestSnapshot, error) {
	bound, err := e.requestBound(m, s, mode)
	if err != nil {
		return nil, err
	}
	c, _ := e.bundle.Lookup(m.Header.Key())
	owned := m
	owned.Body, err = c.Clone(m.Body)
	if err != nil {
		return nil, err
	}
	env, err := c.Envelope(owned.Body)
	if err != nil {
		return nil, err
	}
	return &RequestSnapshot{validation.Context{}, e, owned, env, pdu.AnswerSources{Terminals: cloneOrigins(s.Terminals), ErrorAgents: cloneOrigins(s.ErrorAgents)}, bound}, nil
}
func (r *RequestSnapshot) Bound() uint64      { return r.bound }
func (r *RequestSnapshot) Header() pdu.Header { return r.request.Header }
func (r *RequestSnapshot) MatchesHeader(h pdu.Header) bool {
	return pdu.MatchesAnswerHeader(r.request.Header, h)
}
func invalid(err error) error { return errors.Join(validation.ErrInvalidAnswer, err) }
func (r *RequestSnapshot) CheckAnswer(m pdu.Message) error {
	if !r.MatchesHeader(m.Header) {
		return validation.ErrUnmatchedAnswer
	}
	c, _, err := r.engine.measureMode(m, validation.Receive)
	if err != nil {
		return invalid(err)
	}
	env, err := c.Envelope(m.Body)
	if err != nil {
		return invalid(err)
	}
	if m.Header.Flags&pdu.ProxiableFlag != r.request.Header.Flags&pdu.ProxiableFlag {
		return invalid(wire.ErrFrame)
	}
	if env.SessionID != r.envelope.SessionID {
		return invalid(validation.At(avp.Key{Code: 263}, validation.ErrEnvelope))
	}
	if !pdu.EqualProxies(env.ProxyInfo, r.envelope.ProxyInfo) {
		return invalid(validation.At(avp.Key{Code: 284}, validation.ErrEnvelope))
	}
	isError := m.Header.Flags&pdu.ErrorFlag != 0
	if !isError && r.envelope.AuthSessionState.Present && env.AuthSessionState != r.envelope.AuthSessionState {
		return invalid(validation.At(avp.Key{Code: 277}, validation.ErrEnvelope))
	}
	if !allowed(env.Origin, r.sources.Terminals) && !(isError && allowed(env.Origin, r.sources.ErrorAgents)) {
		return invalid(validation.ErrSource)
	}
	if c.CheckAnswer != nil {
		if err := c.CheckAnswer(r.request.Body, m.Body); err != nil {
			return invalid(err)
		}
	}
	return nil
}

func (e *Engine) ValidateAnswer(request, answer pdu.Message, s pdu.AnswerSources) error {
	r, err := e.prepareRequest(request, s, validation.Receive)
	if err != nil {
		return err
	}
	return r.CheckAnswer(answer)
}

// DecodeAnswerOwned is the runtime's single owned-frame/association entry. It
// returns no partially valid body. Caller-owned buffers use DecodeAnswer instead.
func (r *RequestSnapshot) DecodeAnswerOwned(src []byte) (pdu.Message, error) {
	return r.decodeAnswer(src, true)
}
func (r *RequestSnapshot) DecodeAnswer(src []byte) (pdu.Message, error) {
	return r.decodeAnswer(src, false)
}
func (r *RequestSnapshot) decodeAnswer(src []byte, owned bool) (pdu.Message, error) {
	m, _, err := r.decodeAnswerReport(src, owned)
	return m, err
}

// DecodeAnswerOwnedReport separates wire/exchange failure (no typed answer) from
// a correlated answer with unsupported application semantics (answer retained).
func (r *RequestSnapshot) DecodeAnswerOwnedReport(src []byte) (pdu.Message, validation.Report, error) {
	return r.decodeAnswerReport(src, true)
}
func (r *RequestSnapshot) decodeAnswerReport(src []byte, owned bool) (pdu.Message, validation.Report, error) {
	h, _, err := wire.Header(src, r.engine.limits)
	report := r.engine.report(h)
	if err != nil {
		report.Failure(err)
		if r.MatchesHeader(h) {
			err = invalid(err)
		}
		return pdu.Message{}, report, err
	}
	if !r.MatchesHeader(h) {
		report.Failure(validation.ErrUnmatchedAnswer)
		return pdu.Message{}, report, validation.ErrUnmatchedAnswer
	}
	var m pdu.Message
	if owned {
		m, err = r.engine.DecodeOwned(src)
	} else {
		m, err = r.engine.Decode(src)
	}
	if err != nil {
		report.Failure(err)
		return pdu.Message{}, report, invalid(err)
	}
	report.CheckedRules = append(report.CheckedRules, "WIRE-STATIC")
	if err = r.CheckAnswer(m); err != nil {
		report.Failure(err)
		return pdu.Message{}, report, err
	}
	report.CheckedRules = append(report.CheckedRules, "EXCHANGE-ASSOCIATION")
	r.engine.contextStage(&report, m, r.request.Body, r.context, validation.Receive)
	return m, report, nil
}
func (e *Engine) DecodeAnswer(request pdu.Message, src []byte, s pdu.AnswerSources) (pdu.Message, error) {
	r, err := e.prepareRequest(request, s, validation.Receive)
	if err != nil {
		return pdu.Message{}, err
	}
	return r.DecodeAnswer(src)
}

// CompleteAnswer accepts a validated request, not arbitrary invalid bytes. The
// later runtime rejection path must extract trustworthy envelope evidence within
// its separate rejection budget before replying to a malformed request.
func (e *Engine) CompleteAnswer(request pdu.Message, body pdu.Body, origin pdu.Origin) (pdu.Message, error) {
	if _, err := e.CompleteAnswerBound(request, body, origin); err != nil {
		return pdu.Message{}, err
	}
	if request.Header.Flags&pdu.RequestFlag == 0 {
		return pdu.Message{}, validation.ErrType
	}
	if avp.Identity(origin.Host) != nil || avp.Identity(origin.Realm) != nil {
		return pdu.Message{}, validation.ErrEnvelope
	}
	rc, _, err := e.measureMode(request, validation.Receive)
	if err != nil {
		return pdu.Message{}, err
	}
	env, err := rc.Envelope(request.Body)
	if err != nil {
		return pdu.Message{}, err
	}
	header := request.Header
	header.Flags &= pdu.ProxiableFlag
	if _, ok := body.(*pdu.ErrorAnswer); ok {
		header.Flags |= pdu.ErrorFlag
	}
	c, err := e.operation(header)
	if err != nil {
		return pdu.Message{}, err
	}
	if c.CompleteAnswer == nil {
		return pdu.Message{}, validation.ErrUnsupported
	}
	if err = c.Validate(body, validation.Build); err != nil {
		return pdu.Message{}, err
	}
	// Bound caller-owned fields before the first clone, then bound the final envelope.
	w := wire.NewEncoder(nil, e.limits)
	if err = c.Encode(w, body); err != nil {
		return pdu.Message{}, err
	}
	if w.Err() != nil {
		return pdu.Message{}, w.Err()
	}
	if w.Len() > e.limits.MaxFrameBytes-20 {
		return pdu.Message{}, wire.ErrLimit
	}
	b, err := c.CompleteAnswer(body, pdu.AnswerEnvelope{Command: header.Key(), SessionID: env.SessionID, Origin: origin, ProxyInfo: env.ProxyInfo})
	if err != nil {
		return pdu.Message{}, err
	}
	out := pdu.Message{Header: header, Body: b}
	if err = e.Validate(out); err != nil {
		return pdu.Message{}, err
	}
	// Also guard the normal Auth-Session-State relation; it is not filled from defaults.
	answerEnv, err := c.Envelope(b)
	if err != nil {
		return pdu.Message{}, err
	}
	if header.Flags&pdu.ErrorFlag == 0 && env.AuthSessionState.Present && answerEnv.AuthSessionState != env.AuthSessionState {
		return pdu.Message{}, validation.ErrEnvelope
	}
	if c.CheckAnswer != nil {
		if err := c.CheckAnswer(request.Body, b); err != nil {
			return pdu.Message{}, err
		}
	}
	return out, nil
}

// CompleteAnswerBound covers input clones, copied request envelope and output
// workspace, including failure after merging a too-large envelope. A runtime
// reserves this before completion; the offline codec owns no ledger.
func (e *Engine) CompleteAnswerBound(request pdu.Message, body pdu.Body, origin pdu.Origin) (uint64, error) {
	if request.Header.Flags&pdu.RequestFlag == 0 {
		return 0, validation.ErrType
	}
	if avp.Identity(origin.Host) != nil || avp.Identity(origin.Realm) != nil {
		return 0, validation.ErrEnvelope
	}
	rc, n, err := e.measureMode(request, validation.Receive)
	if err != nil {
		return 0, err
	}
	h := request.Header
	h.Flags &= pdu.ProxiableFlag
	if _, ok := body.(*pdu.ErrorAnswer); ok {
		h.Flags |= pdu.ErrorFlag
	}
	c, err := e.operation(h)
	if err != nil {
		return 0, err
	}
	if c.CompleteAnswer == nil {
		return 0, validation.ErrUnsupported
	}
	if err = c.Validate(body, validation.Build); err != nil {
		return 0, err
	}
	w := wire.NewEncoder(nil, e.limits)
	if err = c.Encode(w, body); err != nil {
		return 0, err
	}
	if w.Err() != nil {
		return 0, w.Err()
	}
	if w.Len() > e.limits.MaxFrameBytes-20 {
		return 0, wire.ErrLimit
	}
	rb, err := rc.DecodeBound(n, e.limits)
	if err != nil {
		return 0, err
	}
	ab, err := c.DecodeBound(w.Len()+20, e.limits)
	if err != nil {
		return 0, err
	}
	if rb > (^uint64(0)-1024)/2 || ab > (^uint64(0)-1024)/2-rb {
		return 0, wire.ErrLimit
	}
	return 2*(rb+ab) + 1024, nil
}

func cloneOrigins(src []pdu.Origin) []pdu.Origin {
	out := make([]pdu.Origin, len(src))
	for i, v := range src {
		out[i] = pdu.Origin{Host: avp.CloneString(v.Host), Realm: avp.CloneString(v.Realm)}
	}
	return out
}

// ForAttempt shares immutable request evidence while replacing only hop identity
// and the frozen per-path source policy. The runtime must serialize frame writes.
func (r *RequestSnapshot) ForAttempt(hbh uint32, retransmitted bool, sources pdu.AnswerSources) *RequestSnapshot {
	c := *r
	c.request.Header.HopByHop = hbh
	c.request.Header.Flags &^= pdu.RetransmittedFlag
	if retransmitted {
		c.request.Header.Flags |= pdu.RetransmittedFlag
	}
	c.sources = sources
	return &c
}
