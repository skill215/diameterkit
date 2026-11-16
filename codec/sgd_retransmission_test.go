// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func retransmissionEvidence(r pdu.Message, era uint32) sgd.RetransmissionEvidence {
	b := r.Body.(*sgd.TFR)
	target := sgd.RetransmissionToMME
	if b.SGSNNumber.Present {
		target = sgd.RetransmissionToSGSN
	}
	return sgd.RetransmissionEvidence{Target: target, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}, SamePLMN: avp.Some(true), PowerSaving: avp.Some(true), MaximumEra: era, RequestedEra: avp.Some(era)}
}
func retransmissionContext(t *testing.T, e sgd.RetransmissionEvidence) validation.Context {
	t.Helper()
	v, err := sgd.Retransmission(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSGDRetransmissionFrozenPairs(t *testing.T) {
	c := sgdCodec(t)
	for _, tc := range []struct {
		r, a  string
		era   uint32
		state validation.State
	}{
		{"v24-sgd-mt-r", "v24-sgd-mt-a", 0, validation.Ready},
		{"v24-sgd-mt-r", "v24-sgd-mt-wait", 0, validation.Ready},
		{"v24-gdd-mt-r", "v24-gdd-mt-a", 0, validation.Ready},
		{"v24-gdd-mt-r", "v24-gdd-mt-wait", 0, validation.Ready},
		{"v24-era-mt-r", "v24-era-mt-wait", 1, validation.Ready},
		{"c27-time-zero-r", "c27-time-zero-a", 1, validation.NeedsApplication},
		{"v24-sgd-mt-r", "c24-tfa-equal", 0, validation.NeedsApplication},
		{"v24-sgd-mt-r", "n24-tfa-late", 0, validation.Rejected},
		{"v24-era-mt-r", "n24-era-late", 1, validation.Rejected},
	} {
		t.Run(tc.a, func(t *testing.T) {
			r, a := decoded(t, c, tc.r), decoded(t, c, tc.a)
			e := retransmissionEvidence(r, tc.era)
			ctx := retransmissionContext(t, e)
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
			if report := c.ValidateContext(r, validation.Context{}); report.State != validation.NeedsApplication {
				t.Fatal(report)
			}
			if report := c.ValidateContext(r, ctx); !report.Complete {
				t.Fatal(report)
			}
			if err := c.ValidateAnswer(r, a, src); err != nil {
				t.Fatal(err)
			}
			report := c.ValidateAnswerContext(r, a, src, ctx)
			sendReport := c.ValidateResponseContext(r, a, src, ctx)
			if sendReport.Complete != (tc.state == validation.Ready) {
				t.Fatal(sendReport)
			}
			if tc.state == validation.NeedsApplication && !errors.Is(sendReport.Err(), validation.ErrUnsupported) {
				t.Fatal(sendReport)
			}
			if report.State != tc.state || report.Complete != (tc.state == validation.Ready) {
				t.Fatal(report)
			}
			for _, m := range []pdu.Message{r, a} {
				raw := encoded(t, c, m)
				again, err := c.Decode(raw)
				if err != nil {
					t.Fatal(err)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), encoded(t, c, again)) {
					t.Fatal("ownership")
				}
				n, err := c.Measure(m)
				if err != nil {
					t.Fatal(err)
				}
				if bound, err := c.DecodeBound(m.Header, n); err != nil || bound < uint64(n) {
					t.Fatal(bound, err)
				}
			}
			x, err := c.Canonical(fixture(t, tc.r), 8192)
			if err != nil {
				t.Fatal(err)
			}
			y, err := c.Canonical(encoded(t, c, r), 8192)
			if err != nil || !bytes.Equal(x, y) {
				t.Fatal(err)
			}
			if tc.era == 1 && !a.Body.(*sgd.TFA).RequestedRetransmissionTime.Present {
				t.Fatal("zero lost")
			}
			a.Header.EndToEnd++
			if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
				t.Fatal(err)
			}
		})
	}
	// Identical seconds in distinct eras do not mean equality; raw U32 is never enough.
	for _, tc := range []struct {
		maxEra, reqEra uint32
		max, req       avp.Time
		state          validation.State
	}{
		{1, 0, 0, 0xffffffff, validation.Ready}, {0, 1, 0xffffffff, 0, validation.Rejected},
		{0xffffffff, 0xfffffffe, 0, 0xffffffff, validation.Ready}, {0, 0xffffffff, 1, 0, validation.Rejected},
		{1, 0, 0, 0, validation.Ready},
	} {
		r, a := decoded(t, c, "v24-era-mt-r"), decoded(t, c, "v24-era-mt-wait")
		r.Body.(*sgd.TFR).MaximumRetransmissionTime.Value = tc.max
		a.Body.(*sgd.TFA).RequestedRetransmissionTime.Value = tc.req
		e := retransmissionEvidence(r, tc.maxEra)
		e.RequestedEra = avp.Some(tc.reqEra)
		report := c.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}, retransmissionContext(t, e))
		if report.State != tc.state {
			t.Fatal(tc, report)
		}
	}
}
func TestSGDRetransmissionFieldsAndRelation(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"n24-tfr-nogmsc-r", "n24-tfa-successtime", "n24-tfa-time3"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	for _, tc := range []struct {
		name string
		code uint32
	}{{"v24-sgd-mt-r", 3330}, {"v24-sgd-mt-r", 3332}, {"v24-sgd-mt-wait", 3331}} {
		for _, edit := range []func([]byte) []byte{
			func(v []byte) []byte { return append(v, v...) },
			func(v []byte) []byte { v[7] = 15; v[12] = 0xff; return v[:16] }, // 3-byte payload for Time; invalid trailing nibble/length for this number
		} {
			raw := rewriteAVP(t, fixture(t, tc.name), []uint32{tc.code}, edit)
			if _, err := c.Decode(raw); err == nil {
				t.Fatal(tc)
			}
		}
	}
	r, a := decoded(t, c, "c24-tfr-nomax-r"), decoded(t, c, "v24-sgd-mt-wait")
	e := retransmissionEvidence(r, 0)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrInvalidAnswer) {
		t.Fatal(err)
	}
	if report := c.ValidateContext(r, validation.Context{}); !report.Complete {
		t.Fatal(report)
	}
	r = decoded(t, c, "v24-sgd-mt-r")
	for _, b := range [][]byte{nil, {0xff}, {0x1a}, bytes.Repeat([]byte{0x11}, 9)} {
		r.Body.(*sgd.TFR).SMSGMSCAddress = avp.Some(b)
		if err := c.Validate(r); err == nil {
			t.Fatal(b)
		}
	}
	r = decoded(t, c, "v24-sgd-mt-r")
	canonical, _ := c.Canonical(encoded(t, c, r), 8192)
	r.Body.(*sgd.TFR).MaximumRetransmissionTime.Value++
	other, _ := c.Canonical(encoded(t, c, r), 8192)
	if bytes.Equal(canonical, other) {
		t.Fatal("time omitted from canonical")
	}
}
func TestSGDRetransmissionFactsAndIsolation(t *testing.T) {
	c := sgdCodec(t)
	r, a := decoded(t, c, "v24-sgd-mt-r"), decoded(t, c, "v24-sgd-mt-wait")
	base := retransmissionEvidence(r, 0)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{base.Responder}}
	for _, tc := range []struct {
		name     string
		edit     func(*sgd.RetransmissionEvidence)
		req, ans validation.State
	}{
		{"PLMN unknown", func(e *sgd.RetransmissionEvidence) { e.SamePLMN.Present = false }, validation.NeedsApplication, validation.NeedsApplication},
		{"other PLMN", func(e *sgd.RetransmissionEvidence) { e.SamePLMN.Value = false }, validation.Rejected, validation.Rejected},
		{"power unknown", func(e *sgd.RetransmissionEvidence) { e.PowerSaving.Present = false }, validation.Ready, validation.NeedsApplication},
		{"not power saving", func(e *sgd.RetransmissionEvidence) { e.PowerSaving.Value = false }, validation.Ready, validation.Rejected},
		{"era unknown", func(e *sgd.RetransmissionEvidence) { e.RequestedEra.Present = false }, validation.Ready, validation.NeedsApplication},
		{"wrong target", func(e *sgd.RetransmissionEvidence) { e.Target = sgd.RetransmissionToSGSN }, validation.Rejected, validation.Rejected},
		{"DRA origin", func(e *sgd.RetransmissionEvidence) { e.Requester.Host = "dra.example" }, validation.Rejected, validation.Rejected},
		{"DRA responder", func(e *sgd.RetransmissionEvidence) { e.Responder.Host = "dra.example" }, validation.Rejected, validation.Rejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := base
			tc.edit(&e)
			v := retransmissionContext(t, e)
			if p := c.ValidateContext(r, v); p.State != tc.req {
				t.Fatal(p)
			}
			if p := c.ValidateAnswerContext(r, a, src, v); p.State != tc.ans {
				t.Fatal(p)
			}
		})
	}
	for _, edit := range []func(*sgd.RetransmissionEvidence){func(e *sgd.RetransmissionEvidence) { e.Target = 0 }, func(e *sgd.RetransmissionEvidence) { e.Requester.Host = "" }, func(e *sgd.RetransmissionEvidence) { e.Responder.Realm = string(bytes.Repeat([]byte{'a'}, 254)) }} {
		e := base
		edit(&e)
		if _, err := sgd.Retransmission(e); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	v := retransmissionContext(t, base)
	wrong, err := s6c.AlertRole(s6c.AlertRoleEvidence{Role: s6c.AlertFromMME, Requester: base.Requester, Responder: base.Responder})
	if err != nil {
		t.Fatal(err)
	}
	if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	plain := decoded(t, c, "v-alr-01")
	if p := c.ValidateContext(plain, v); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	// E uses its own grammar/association and never needs normal TFA time evidence.
	ea, err := c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 3004}, base.Responder)
	if err != nil {
		t.Fatal(err)
	}
	if p := c.ValidateAnswerContext(r, ea, src, validation.Context{}); !p.Complete {
		t.Fatal(p)
	}
}
