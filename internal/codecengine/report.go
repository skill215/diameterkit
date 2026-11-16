// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codecengine

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

func (e *Engine) report(h pdu.Header) validation.Report {
	return validation.Report{Scope: fmt.Sprintf("context-v2/%s/%d/%d/R=%t/E=%t", e.bundle.Fingerprint(), h.ApplicationID, h.CommandCode, h.Flags&pdu.RequestFlag != 0, h.Flags&pdu.ErrorFlag != 0)}
}
func (e *Engine) contextStage(r *validation.Report, m pdu.Message, request pdu.Body, ctx validation.Context, mode validation.Mode) {
	op, err := e.operation(m.Header)
	if err != nil {
		r.Failure(err)
		return
	}
	if m.Header.Flags&pdu.ErrorFlag != 0 {
		r.CheckedRules = append(r.CheckedRules, "APPLICATION-E-ENVELOPE")
		r.Complete, r.State = true, validation.Ready
		return
	}
	if op.CheckContext != nil {
		var answer pdu.Body
		if m.Header.Flags&pdu.RequestFlag != 0 {
			request = m.Body
		} else {
			answer = m.Body
		}
		err = op.CheckContext(request, answer, ctx, mode)
		if errors.Is(err, validation.ErrContextRequired) {
			r.State = validation.NeedsApplication
			r.MissingContext = []string{op.ContextName}
			return
		}
		if err != nil {
			r.Failure(err)
			return
		}
		r.CheckedRules = append(r.CheckedRules, op.ContextName)
	} else {
		if !ctx.Empty() {
			r.Failure(validation.ErrContext)
			return
		}
		if !op.NoContext {
			r.Failure(validation.ErrUnsupported)
			return
		}
		r.CheckedRules = append(r.CheckedRules, "APPLICATION-NO-CONTEXT")
	}
	r.Complete, r.State = true, validation.Ready
}
func (e *Engine) ValidateContext(m pdu.Message, ctx validation.Context) validation.Report {
	return e.validateContextMode(m, ctx, validation.Receive)
}
func (e *Engine) validateContextMode(m pdu.Message, ctx validation.Context, mode validation.Mode) validation.Report {
	r := e.report(m.Header)
	if _, _, err := e.measureMode(m, mode); err != nil {
		r.Failure(err)
		return r
	}
	r.CheckedRules = append(r.CheckedRules, "WIRE-STATIC")
	e.contextStage(&r, m, nil, ctx, mode)
	if m.Header.Flags&pdu.RequestFlag == 0 && r.State != validation.Rejected {
		r.Complete, r.State = false, validation.NeedsApplication
		// Original-request association is the next required stage; it then checks context.
		r.MissingContext = []string{"OriginalRequest", "AnswerSources"}
	}
	return r
}

// BindContext checks runtime destination evidence before the request becomes visible.
// A feature binding chooses exactly one end terminal, not an adjacent agent.
func (r *RequestSnapshot) BindContext(ctx validation.Context) error {
	_, _, host, realm, present := ctx.Endpoints()
	if present && (len(r.sources.Terminals) != 1 || !strings.EqualFold(r.sources.Terminals[0].Host, host) || !strings.EqualFold(r.sources.Terminals[0].Realm, realm)) {
		return validation.ErrContext
	}
	report := r.engine.validateContextMode(r.request, ctx, validation.Send)
	if err := report.Err(); err != nil {
		return err
	}
	r.context = ctx
	return nil
}
func (e *Engine) ValidateAnswerContext(request, answer pdu.Message, sources pdu.AnswerSources, ctx validation.Context) validation.Report {
	return e.validateAnswerContextMode(request, answer, sources, ctx, validation.Receive)
}

// ValidateResponseContext uses receive rules for the original request and send
// rules for the new answer. In particular receive-only tolerated M bits survive.
func (e *Engine) ValidateResponseContext(request, answer pdu.Message, sources pdu.AnswerSources, ctx validation.Context) validation.Report {
	return e.validateAnswerContextMode(request, answer, sources, ctx, validation.Send)
}
func (e *Engine) validateAnswerContextMode(request, answer pdu.Message, sources pdu.AnswerSources, ctx validation.Context, mode validation.Mode) validation.Report {
	r := e.report(answer.Header)
	snap, err := e.prepareRequest(request, sources, validation.Receive)
	if err != nil {
		r.Failure(validation.ErrContext)
		return r
	}
	op, _ := e.operation(request.Header)
	if !op.NoContext && op.CheckContext == nil {
		r.Failure(validation.ErrUnsupported)
		return r
	}
	return snap.validateAnswerContextMode(answer, ctx, mode)
}
func (r *RequestSnapshot) ValidateAnswerContext(answer pdu.Message, ctx validation.Context) validation.Report {
	return r.validateAnswerContextMode(answer, ctx, validation.Receive)
}
func (r *RequestSnapshot) validateAnswerContextMode(answer pdu.Message, ctx validation.Context, mode validation.Mode) validation.Report {
	report := r.engine.report(answer.Header)
	if !r.MatchesHeader(answer.Header) {
		report.Failure(validation.ErrUnmatchedAnswer)
		return report
	}
	if _, _, err := r.engine.measureMode(answer, mode); err != nil {
		report.Failure(err)
		return report
	}
	report.CheckedRules = append(report.CheckedRules, "WIRE-STATIC")
	if err := r.CheckAnswer(answer); err != nil {
		report.Failure(err)
		return report
	}
	report.CheckedRules = append(report.CheckedRules, "EXCHANGE-ASSOCIATION")
	r.engine.contextStage(&report, answer, r.request.Body, ctx, mode)
	return report
}
