// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func t4Evidence(r pdu.Message) sgd.T4ChargingEvidence {
	b := r.Body.(*sgd.OFR)
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	e := sgd.T4ChargingEvidence{Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, SenderIMSI: b.UserIdentifier.UserName.Value, SCAddress: sc, ExternalIdentifier: avp.Some("device-001@example.net")}
	if b.UserIdentifier.MSISDN.Present {
		n, _ := avp.DecodeTBCD(b.UserIdentifier.MSISDN.Value)
		e.SenderMSISDN = avp.Some(n)
	}
	return e
}
func t4Context(t *testing.T, e sgd.T4ChargingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.T4Charging(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSGDT4FrozenPairsAndFlags(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "v25-t4-ofr-r")
	e := t4Evidence(r)
	ctx := t4Context(t, e)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal(p)
	}
	for _, tc := range []struct{ file, send string }{{"v25-t4-ofa-a", "b25-t4-ofa-m0"}, {"f25-t4-ofa-5555", "b25-t4-fail-m0"}, {"b25-t4-ofa-m0", "b25-t4-ofa-m0"}, {"b25-t4-fail-m0", "b25-t4-fail-m0"}} {
		a := decoded(t, c, tc.file)
		if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
			t.Fatal(p)
		}
		if p := c.ValidateResponseContext(r, a, src, ctx); !p.Complete {
			t.Fatal(p)
		}
		raw := encoded(t, c, a)
		if !bytes.Equal(raw, fixture(t, tc.send)) {
			t.Fatal("send must clear only3111 M", tc.file)
		}
		again, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, again), encoded(t, c, a)) {
			t.Fatal("borrowed answer")
		}
		n, err := c.Measure(a)
		if err != nil {
			t.Fatal(err)
		}
		if bound, err := c.DecodeBound(a.Header, n); err != nil || bound < uint64(n) {
			t.Fatal(bound, err)
		}
		a.Header.ApplicationID = 16777311
		if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"n25-t4-noext", "n26-ofa-vendor-shadow"} {
		a := decoded(t, c, name)
		if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrMissing) {
			t.Fatal(name, p)
		}
		if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); !p.Complete {
			t.Fatal("ordinary MO absence allowed", p)
		}
	}
	for _, name := range []string{"n25-t4-nested", "n25-t4-nai", "n25-t4-wrongapp", "n25-l092-text", "n25-l092-dup"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	a, err := c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 3004}, e.Responder)
	if err != nil {
		t.Fatal(err)
	}
	if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); !p.Complete {
		t.Fatal(p)
	}
	if p := c.ValidateResponseContext(r, a, src, ctx); !p.Complete {
		t.Fatal(p)
	}
}
func TestSGDT4OriginalIdentityAndEvidence(t *testing.T) {
	c := sgdCodec(t)
	r, a := decoded(t, c, "v25-t4-ofr-r"), decoded(t, c, "b25-t4-ofa-m0")
	base := t4Evidence(r)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{base.Responder}}
	for _, tc := range []struct {
		name     string
		edit     func(*sgd.T4ChargingEvidence)
		req, ans validation.State
	}{
		{"unknown mapping", func(e *sgd.T4ChargingEvidence) { e.ExternalIdentifier.Present = false }, validation.Ready, validation.NeedsApplication},
		{"other mapping", func(e *sgd.T4ChargingEvidence) { e.ExternalIdentifier.Value = "device-002@example.net" }, validation.Ready, validation.Rejected},
		{"local case significant", func(e *sgd.T4ChargingEvidence) { e.ExternalIdentifier.Value = "Device-001@example.net" }, validation.Ready, validation.Rejected},
		{"other sender", func(e *sgd.T4ChargingEvidence) { e.SenderIMSI = "001010123456780" }, validation.Rejected, validation.Rejected},
		{"missing MSISDN", func(e *sgd.T4ChargingEvidence) { e.SenderMSISDN.Present = false }, validation.Rejected, validation.Rejected},
		{"different MSISDN", func(e *sgd.T4ChargingEvidence) { e.SenderMSISDN = avp.Some("1234") }, validation.Rejected, validation.Rejected},
		{"different SC", func(e *sgd.T4ChargingEvidence) { e.SCAddress = "1234" }, validation.Rejected, validation.Rejected},
		{"DRA sender", func(e *sgd.T4ChargingEvidence) { e.Requester.Host = "dra.example" }, validation.Rejected, validation.Rejected},
		{"DRA responder", func(e *sgd.T4ChargingEvidence) { e.Responder.Host = "dra.example" }, validation.Rejected, validation.Rejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := base
			tc.edit(&e)
			v := t4Context(t, e)
			if p := c.ValidateContext(r, v); p.State != tc.req {
				t.Fatal(p)
			}
			if p := c.ValidateAnswerContext(r, a, src, v); p.State != tc.ans {
				t.Fatal(p)
			}
		})
	}
	for _, edit := range []func(*sgd.T4ChargingEvidence){func(e *sgd.T4ChargingEvidence) { e.SenderIMSI = "opaque-token" }, func(e *sgd.T4ChargingEvidence) { e.SCAddress = "+1234" }, func(e *sgd.T4ChargingEvidence) { e.SenderMSISDN = avp.Some("") }, func(e *sgd.T4ChargingEvidence) { e.ExternalIdentifier = avp.Some("@example.net") }, func(e *sgd.T4ChargingEvidence) { e.Requester.Host = strings.Repeat("x", 254) }} {
		e := base
		edit(&e)
		if _, err := sgd.T4Charging(e); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	ctx := t4Context(t, base)
	b := r.Body.(*sgd.OFR)
	b.DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	b.DestinationHost = avp.Some(base.Responder.Host)
	b.UserIdentifier.MSISDN.Present = false
	base.SenderMSISDN.Present = false
	if p := c.ValidateContext(r, t4Context(t, base)); !p.Complete {
		t.Fatal("known IMSI without MSISDN", p)
	}
	before, err := c.Canonical(encoded(t, c, r), 4096)
	if err != nil {
		t.Fatal(err)
	}
	b.UserIdentifier.UserName = avp.Some("001010123456780")
	after, err := c.Canonical(encoded(t, c, r), 4096)
	if err != nil || bytes.Equal(before, after) {
		t.Fatal("sender collapsed", err)
	}
	mt := decoded(t, c, "v24-sgd-mt-r")
	if p := c.ValidateContext(mt, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	wrong := retransmissionContext(t, retransmissionEvidence(mt, 0))
	if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
}
func TestSGDT4ExternalIdentifierASCIIAndReceiveBoundary(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "v25-t4-ofr-r")
	e := t4Evidence(r)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	for _, s := range []string{"device-001@example.net", `user\@site@example.net`, `\(user\)@example.net`, `a\.b@example.net`, `a\\b@example.net`, `a\ b@example.net`, "!#$%&'*+-/=?^_`{|}~@example.net", "a.b@EXAMPLE.NET", strings.Repeat("a", 241) + "@example.net"} {
		e.ExternalIdentifier = avp.Some(s)
		ctx := t4Context(t, e)
		a := decoded(t, c, "b25-t4-ofa-m0")
		a.Body.(*sgd.OFA).ExternalIdentifier = avp.Some(s)
		if p := c.ValidateResponseContext(r, a, src, ctx); !p.Complete {
			t.Fatal(s, p)
		}
		again, err := c.Decode(encoded(t, c, a))
		if err != nil || again.Body.(*sgd.OFA).ExternalIdentifier.Value != s {
			t.Fatal(s, err)
		}
	}
	for _, s := range []string{"", "@example.net", "u@", "u@example", "u@@example.net", ".u@example.net", "u.@example.net", "u..v@example.net", `u\@example.net`, "u v@example.net", "u:bad@example.net", "u@bad_.net", "u@-bad.net", "u@example.net.", "u\x00@example.net", "u\\\x01@example.net"} {
		a := decoded(t, c, "b25-t4-ofa-m0")
		a.Body.(*sgd.OFA).ExternalIdentifier = avp.Some(s)
		if err := c.Validate(a); err == nil {
			t.Fatal(s)
		}
	}
	e = t4Evidence(r)
	ctx := t4Context(t, e)
	for _, s := range []string{"设备@example.net", strings.Repeat("a", 242) + "@example.net"} {
		raw := rewriteAVP(t, fixture(t, "b25-t4-ofa-m0"), []uint32{3111}, func(v []byte) []byte {
			v = append(v[:12], []byte(s)...)
			wire.Put24(v[5:8], uint32(len(v)))
			return append(v, make([]byte, (4-len(v)%4)%4)...)
		})
		a, err := c.Decode(raw)
		if err != nil || a.Body.(*sgd.OFA).ExternalIdentifier.Value != s {
			t.Fatal(err)
		}
		if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(p)
		}
		if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
	}
}
