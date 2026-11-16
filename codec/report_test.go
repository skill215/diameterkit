// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func TestContextReportPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, pair := range [][2]string{{"v-srr-01", "v-sra-01"}, {"b2-srr-proxy", "b2-sra-normal-proxy"}, {"b2-srr-proxy", "b2-sra-e-proxy"}, {"b2-srr", "e28-sra-5012"}} {
		t.Run(pair[1], func(t *testing.T) {
			req, ans := decoded(t, c, pair[0]), decoded(t, c, pair[1])
			ctx := validation.Context{}
			r := c.ValidateContext(req, ctx)
			if !r.Complete || r.State != validation.Ready || len(r.CheckedRules) != 2 || !strings.HasPrefix(r.Scope, "context-v2/") {
				t.Fatal(r)
			}
			r = c.ValidateContext(ans, ctx)
			if r.Complete || r.State != validation.NeedsApplication || len(r.MissingContext) != 2 || len(r.LocalErrors) != 0 {
				t.Fatal(r)
			}
			r.MissingContext[0] = "caller mutation"
			r = c.ValidateAnswerContext(req, ans, sources(), ctx)
			if !r.Complete || r.State != validation.Ready || len(r.CheckedRules) != 3 || len(r.MissingContext) != 0 {
				t.Fatal(r)
			}
			r = c.ValidateAnswerContext(req, ans, pdu.AnswerSources{}, ctx)
			if r.Complete || len(r.LocalErrors) != 1 || len(r.Violations) != 0 {
				t.Fatal(r)
			}
		})
	}
}

func TestContextReportStageFailures(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "b2-srr-proxy")
	for _, tc := range []struct {
		name    string
		mutate  func(*pdu.Message)
		checked int
		want    error
		local   bool
	}{
		{"header", func(m *pdu.Message) { m.Header.EndToEnd++ }, 0, validation.ErrUnmatchedAnswer, false},
		{"wire", func(m *pdu.Message) { m.Body.(*s6c.SRA).OriginHost = "" }, 0, validation.ErrField, false},
		{"session", func(m *pdu.Message) { m.Body.(*s6c.SRA).SessionID = "wrong" }, 1, validation.ErrEnvelope, false},
		{"source", func(m *pdu.Message) { m.Body.(*s6c.SRA).OriginHost = "other.example" }, 1, validation.ErrSource, false},
		{"unsupported", func(m *pdu.Message) { m.Body.(*s6c.SRA).ResultCode.Value = 9999 }, 0, validation.ErrUnsupported, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ans := decoded(t, c, "b2-sra-normal-proxy")
			tc.mutate(&ans)
			r := c.ValidateAnswerContext(req, ans, sources(), validation.Context{})
			found := r.Violations
			if tc.local {
				found = r.LocalErrors
			}
			if r.Complete || r.State != validation.Rejected || len(r.CheckedRules) != tc.checked || len(found) != 1 || !errors.Is(found[0], tc.want) || len(r.MissingContext) != 0 {
				t.Fatal(r)
			}
		})
	}
}

func TestContextAuditPermission(t *testing.T) {
	p, _ := s6c.Profile()
	b, _ := schema.Compile(p)
	req, _ := b.Lookup((&s6c.SRR{}).CommandKey())
	ans, _ := b.Lookup((&s6c.SRA{}).CommandKey())
	if req.NoContext || ans.NoContext || req.CheckContext == nil || ans.CheckContext == nil {
		t.Fatal("audit missing")
	}
	req.CheckContext = nil
	req.ContextName = ""
	other, err := schema.NewProfile(p.Name(), p.Version(), p.Digest(), req, ans)
	if err != nil {
		t.Fatal(err)
	}
	noAudit, _ := schema.Compile(other)
	if noAudit.Fingerprint() == b.Fingerprint() {
		t.Fatal("context audit absent from fingerprint")
	}
	c, _ := codec.New(noAudit, codec.Config{})
	m := decoded(t, c, "v-srr-01") // wire functionality remains available
	r := c.ValidateContext(m, validation.Context{})
	if r.Complete || len(r.LocalErrors) != 1 || !errors.Is(r.LocalErrors[0], validation.ErrUnsupported) || len(r.MissingContext) != 0 {
		t.Fatal(r)
	}
	a := decoded(t, c, "v-sra-01")
	r = c.ValidateAnswerContext(m, a, sources(), validation.Context{})
	if r.Complete || len(r.LocalErrors) != 1 {
		t.Fatal(r)
	}
}

func TestBaseContextReportPair(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req, ans := decoded(t, c, "base-dwr"), decoded(t, c, "base-dwa")
	body := ans.Body.(*base.DWA)
	sources := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: body.OriginHost, Realm: body.OriginRealm}}}
	r := c.ValidateAnswerContext(req, ans, sources, validation.Context{})
	if !r.Complete || len(r.CheckedRules) != 3 {
		t.Fatal(r)
	}
	r = c.ValidateContext(ans, validation.Context{})
	if r.Complete || len(r.MissingContext) != 2 {
		t.Fatal(r)
	}
}
