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
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func t4RoutingFacts(r pdu.Message) s6c.T4NoMSISDNEvidence {
	b := r.Body.(*s6c.SRR)
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return s6c.T4NoMSISDNEvidence{Procedure: s6c.T4Routing, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, IMSI: b.UserName.Value, SCAddress: sc}
}
func TestS6cT4RoutingPairsAndOwnership(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, pair := range [][2]string{{"d21-t4-srr-r", "d21-t4-srr-a"}, {"d21-t4-srr-r", "d21-t4-srr-f"}, {"d21-t4-srr-r", "b27-t4-sra-e"}, {"b27-t4-srr-gprs", "b27-t4-sra-extra-node"}} {
		t.Run(pair[1], func(t *testing.T) {
			r, a := decoded(t, c, pair[0]), decoded(t, c, pair[1])
			e := t4RoutingFacts(r)
			ctx := s6cT4Context(t, e)
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
			if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
				t.Fatal(p)
			}
			if p := c.ValidateContext(r, ctx); !p.Complete {
				t.Fatal(p)
			}
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
				if !p.Complete {
					t.Fatal(p)
				}
			}
			if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); a.Header.Flags&pdu.ErrorFlag == 0 && p.State != validation.NeedsApplication {
				t.Fatal(p)
			}
			for _, m := range []pdu.Message{r, a} {
				raw := encoded(t, c, m)
				owned, err := c.Decode(raw)
				if err != nil {
					t.Fatal(err)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), encoded(t, c, owned)) {
					t.Fatal("borrowed bytes")
				}
				n, err := c.Measure(m)
				if err != nil || n != len(encoded(t, c, m)) {
					t.Fatal(n, err)
				}
				if bound, err := c.DecodeBound(m.Header, n); err != nil || bound < uint64(n) {
					t.Fatal(bound, err)
				}
			}
			canon, err := c.Canonical(fixture(t, pair[0]), 4096)
			if err != nil {
				t.Fatal(err)
			}
			other, err := c.Canonical(encoded(t, c, r), 4096)
			if err != nil || !bytes.Equal(canon, other) {
				t.Fatal(err)
			}
			a.Header.EndToEnd++
			if err = c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
				t.Fatal(err)
			}
		})
	}
}
func TestS6cT4RoutingIdentityPurposeAndNodes(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "d21-t4-srr-r")
	e := t4RoutingFacts(r)
	ctx := s6cT4Context(t, e)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	for _, mutate := range []func(*s6c.T4NoMSISDNEvidence){func(e *s6c.T4NoMSISDNEvidence) { e.Procedure = s6c.T4DeliveryReport }, func(e *s6c.T4NoMSISDNEvidence) { e.Procedure = s6c.T4HSSAlert }, func(e *s6c.T4NoMSISDNEvidence) { e.IMSI = "001010123456780" }, func(e *s6c.T4NoMSISDNEvidence) { e.SCAddress = "123" }, func(e *s6c.T4NoMSISDNEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.T4NoMSISDNEvidence) { e.Responder.Realm = "other.example" }} {
		wrong := e
		mutate(&wrong)
		if p := c.ValidateContext(r, s6cT4Context(t, wrong)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"b27-t4-sra-otherimsi", "b27-t4-sra-extra-node"} {
		a := decoded(t, c, name)
		if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrInvalidAnswer) {
			t.Fatal(name, err)
		}
		if p := c.ValidateAnswerContext(r, a, src, ctx); p.Complete {
			t.Fatal(p)
		}
	}
	// The selected five node shapes remain available under the known IMSI context.
	for i := range 5 {
		a := decoded(t, c, "d21-t4-srr-a")
		a.Body.(*s6c.SRA).ServingNode = avp.Some(routingNode(i))
		if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
			t.Fatal(i, p)
		}
	}
	b := r.Body.(*s6c.SRR)
	b.DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	// Ordinary MSISDN queries retain their zero-context behavior, without guessing IMSI.
	ordinary := decoded(t, c, "v-srr-01")
	if p := c.ValidateContext(ordinary, validation.Context{}); !p.Complete {
		t.Fatal(p)
	}
	if p := c.ValidateContext(ordinary, ctx); p.Complete {
		t.Fatal(p)
	}
	for _, name := range []string{"b26-t4-rdr-mme", "d21-t4-alr-r"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("cross-command purpose", p)
		}
	}
}
func TestS6cT4RoutingWireControls(t *testing.T) {
	c := newCodec(t, codec.Config{})
	if e := c.ValidateAnswer(decoded(t, c, "d21-t4-srr-r"), decoded(t, c, "b27-t4-sra-nonode"), routingSources()); e == nil {
		t.Fatal("delivery answer without node")
	}

	for _, name := range []string{"b27-t4-srr-bothid", "b27-t4-srr-noid", "b27-t4-srr-dupimsi", "b27-t4-srr-badimsi", "b27-t4-sra-noimsi"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	for _, key := range []avp.Key{{VendorID: 10415, Code: 3324}, {VendorID: 10415, Code: 3311}} {
		r := decoded(t, c, "d21-t4-srr-r")
		r.Body.(*s6c.SRR).Extra = []avp.Raw{{Key: key, Flags: 128}}
		want := validation.ErrUnsupported
		if key.Code == 3324 || key.Code == 3311 {
			want = validation.ErrDuplicate
		}
		if err := c.Validate(r); !errors.Is(err, want) {
			t.Fatal(err)
		}
	}
	for _, edit := range []func([]byte) []byte{
		func(v []byte) []byte { v[8] = 0xc3; v[9] = 0x28; return v },
		func(v []byte) []byte { v[8] = 0; return v },
	} {
		if _, err := c.Decode(rewriteAVP(t, fixture(t, "d21-t4-srr-r"), []uint32{1}, edit)); err == nil {
			t.Fatal("invalid UTF8/NUL")
		}
	}
	r := decoded(t, c, "b27-t4-srr-gprs")
	r.Body.(*s6c.SRR).SMRPSMEA.Present = false
	if err := c.Validate(r); !errors.Is(err, validation.ErrMissing) {
		t.Fatal(err)
	}
	// Presence, not residual optional Value, controls whether MSISDN is sent.
	r = decoded(t, c, "d21-t4-srr-r")
	r.Body.(*s6c.SRR).MSISDN.Value = []byte{255}
	if _, err := c.Snapshot(r); err != nil {
		t.Fatal(err)
	}
	a := decoded(t, c, "d21-t4-srr-f")
	a.Body.(*s6c.SRA).UserName = avp.Some("001010123456789")
	if err := c.Validate(a); !errors.Is(err, validation.ErrField) {
		t.Fatal("failure identity", err)
	}
}
