// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func s6cT4Evidence(r pdu.Message) s6c.T4NoMSISDNEvidence {
	var e s6c.T4NoMSISDNEvidence
	var sc []byte
	switch b := r.Body.(type) {
	case *s6c.RDR:
		e.Procedure = s6c.T4DeliveryReport
		e.Requester = pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}
		e.Responder = pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}
		e.IMSI = b.UserIdentifier.UserName.Value
		sc = b.SCAddress
	case *s6c.ALR:
		e.Procedure = s6c.T4HSSAlert
		e.Requester = pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}
		e.Responder = pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}
		e.IMSI = b.UserIdentifier.UserName.Value
		sc = b.SCAddress
	}
	e.SCAddress, _ = avp.DecodeTBCD(sc)
	return e
}
func s6cT4Context(t *testing.T, e s6c.T4NoMSISDNEvidence) validation.Context {
	t.Helper()
	v, err := s6c.T4NoMSISDN(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestS6cT4PairsAndOwnership(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, pair := range [][2]string{{"b26-t4-rdr-mme", "d21-t4-rdr-a"}, {"b26-t4-rdr-mme", "d21-t4-rdr-f"}, {"b26-t4-rdr-mme", "b26-t4-rda-5001"}, {"d21-t4-alr-r", "d21-t4-alr-a"}, {"d21-t4-alr-r", "d21-t4-alr-f"}} {
		t.Run(pair[1], func(t *testing.T) {
			r, a := decoded(t, c, pair[0]), decoded(t, c, pair[1])
			e := s6cT4Evidence(r)
			ctx := s6cT4Context(t, e)
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
			if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
				t.Fatal(p)
			}
			if p := c.ValidateContext(r, ctx); !p.Complete {
				t.Fatal(p)
			}
			if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
				t.Fatal(p)
			}
			if p := c.ValidateResponseContext(r, a, src, ctx); !p.Complete {
				t.Fatal(p)
			}
			if a.Header.Flags&pdu.ErrorFlag == 0 {
				if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); p.State != validation.NeedsApplication {
					t.Fatal(p)
				}
			} else {
				if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); !p.Complete {
					t.Fatal(p)
				}
			}
			for _, m := range []pdu.Message{r, a} {
				raw := encoded(t, c, m)
				again, err := c.Decode(raw)
				if err != nil {
					t.Fatal(err)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), encoded(t, c, again)) {
					t.Fatal("borrowed bytes")
				}
				n, err := c.Measure(m)
				if err != nil {
					t.Fatal(err)
				}
				if bound, err := c.DecodeBound(m.Header, n); err != nil || bound < uint64(n) {
					t.Fatal(bound, err)
				}
			}
			canon, err := c.Canonical(fixture(t, pair[0]), 4096)
			if err != nil {
				t.Fatal(err)
			}
			again, err := c.Canonical(encoded(t, c, r), 4096)
			if err != nil || !bytes.Equal(canon, again) {
				t.Fatal(err)
			}
			a.Header.HopByHop++
			if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
				t.Fatal(err)
			}
		})
	}
	// Ordinary RDA Alert-MSISDN remains a separate optional success adjunct.
	r, a := decoded(t, c, "b26-t4-rdr-mme"), decoded(t, c, "d21-t4-rdr-a")
	e := s6cT4Evidence(r)
	a.Body.(*s6c.RDA).UserIdentifier = avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf3}})
	if p := c.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}, s6cT4Context(t, e)); !p.Complete {
		t.Fatal(p)
	}
	for _, name := range []string{"b26-t4-rdr-bothid", "b26-t4-rdr-emptyid"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	// Batch38 enables the existing typed IP parent under explicit T4 report evidence.
	t4IP := decoded(t, c, "d21-t4-rdr-r")
	if p := c.ValidateContext(t4IP, s6cT4Context(t, s6cT4Evidence(t4IP))); !p.Complete {
		t.Fatal(p)
	}
	if p := c.ValidateContext(t4IP, validation.Context{}); p.State != validation.NeedsApplication {
		t.Fatal(p)
	}
	if err := c.Validate(t4IP); err != nil {
		t.Fatal(err)
	}

	// The largest active context (T4 charging from batch25) still fits the stated
	// 2KiB bound even after adding this fixed representation. No RSS assertion.
	if reflect.TypeFor[validation.Context]().Size()+4*253+3*15+253 > 2048 {
		t.Fatal("context bookkeeping bound")
	}
}
func TestS6cT4PurposeRolesAndIdentity(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"b26-t4-rdr-mme", "d21-t4-alr-r"} {
		r := decoded(t, c, name)
		base := s6cT4Evidence(r)
		for _, edit := range []func(*s6c.T4NoMSISDNEvidence){func(e *s6c.T4NoMSISDNEvidence) { e.Procedure = 3 - e.Procedure }, func(e *s6c.T4NoMSISDNEvidence) { e.IMSI = "001010123456780" }, func(e *s6c.T4NoMSISDNEvidence) { e.SCAddress = "123" }, func(e *s6c.T4NoMSISDNEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.T4NoMSISDNEvidence) { e.Responder.Host = "other.example" }} {
			e := base
			edit(&e)
			if p := c.ValidateContext(r, s6cT4Context(t, e)); !errors.Is(p.Err(), validation.ErrContext) {
				t.Fatal(name, p)
			}
		}
		for _, edit := range []func(*s6c.T4NoMSISDNEvidence){func(e *s6c.T4NoMSISDNEvidence) { e.Procedure = 0 }, func(e *s6c.T4NoMSISDNEvidence) { e.IMSI = "123" }, func(e *s6c.T4NoMSISDNEvidence) { e.IMSI = "opaque-token" }, func(e *s6c.T4NoMSISDNEvidence) { e.SCAddress = "+123" }, func(e *s6c.T4NoMSISDNEvidence) { e.Responder.Host = strings.Repeat("x", 254) }} {
			e := base
			edit(&e)
			if _, err := s6c.T4NoMSISDN(e); !errors.Is(err, validation.ErrContext) {
				t.Fatal(err)
			}
		}
		switch b := r.Body.(type) {
		case *s6c.RDR:
			b.UserIdentifier.MSISDN = avp.Some([]byte{0x21})
		case *s6c.ALR:
			b.UserIdentifier.MSISDN = avp.Some([]byte{0x21})
		}
		if err := c.Validate(r); err == nil {
			t.Fatal("identity XOR")
		}
	}
	hss := decoded(t, c, "d21-t4-alr-r")
	facts := s6cT4Evidence(hss)
	ctx := s6cT4Context(t, facts)
	for _, role := range []s6c.AlertOriginRole{s6c.AlertFromHSS, s6c.AlertFromMME, s6c.AlertFromSGSN} {
		if p := c.ValidateContext(hss, accessContext(t, hss, role)); p.Complete {
			t.Fatal("role alone proves T4", p)
		}
	}
	hss.Body.(*s6c.ALR).MaximumUEAvailabilityTime = avp.Some(avp.Time(0))
	if p := c.ValidateContext(hss, ctx); !p.Complete {
		t.Fatal(p)
	}
	hss.Body.(*s6c.ALR).MaximumUEAvailabilityTime.Present = false
	hss.Body.(*s6c.ALR).SMSGMSCAlertEvent = avp.Some(uint32(1))
	if p := c.ValidateContext(hss, ctx); p.Complete {
		t.Fatal("T4 mixed access event", p)
	}
	if p := c.ValidateContext(hss, accessContext(t, hss, s6c.AlertFromMME)); !p.Complete {
		t.Fatal(p)
	}
	noevent := decoded(t, c, "b23-alr-noevent")
	if p := c.ValidateContext(noevent, validation.Context{}); p.State != validation.NeedsApplication {
		t.Fatal(p)
	}
	if p := c.ValidateContext(noevent, accessContext(t, noevent, s6c.AlertFromMME)); !errors.Is(p.Err(), validation.ErrMissing) {
		t.Fatal(p)
	}
	// Explicit MSISDN HSS still works; T4 assertion must not override its identity.
	ordinary := decoded(t, c, "v-alr-01")
	if p := c.ValidateContext(ordinary, validation.Context{}); !p.Complete {
		t.Fatal(p)
	}
	if p := c.ValidateContext(ordinary, ctx); p.Complete {
		t.Fatal(p)
	}
}
func TestS6cT4WireNegativesAndCommandIsolation(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"b26-t4-rdr-mme", "d21-t4-alr-r"} {
		for _, path := range [][]uint32{{3102}, {3102, 1}} {
			raw := rewriteAVP(t, fixture(t, name), path, func(v []byte) []byte { return append(v, v...) })
			if _, err := c.Decode(raw); !errors.Is(err, validation.ErrDuplicate) {
				t.Fatal(name, path, err)
			}
		}
		r := decoded(t, c, name)
		ctx := s6cT4Context(t, s6cT4Evidence(r))
		switch b := r.Body.(type) {
		case *s6c.RDR:
			b.UserIdentifier.UserName = avp.Some("not-an-imsi")
		case *s6c.ALR:
			b.UserIdentifier.UserName = avp.Some("not-an-imsi")
		}
		if _, ok := r.Body.(*s6c.ALR); ok {
			// ALR wire syntax also carries Router tokens. T4 still binds the
			// original numeric IMSI; a token cannot satisfy that purpose.
			if report := c.ValidateContext(r, ctx); report.Complete {
				t.Fatal(report)
			}
		} else if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
		r = decoded(t, c, name)
		switch b := r.Body.(type) {
		case *s6c.RDR:
			b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3324}, Flags: 128}}
		case *s6c.ALR:
			b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3324}, Flags: 128}}
		}
		want := validation.ErrDuplicate
		if err := c.Validate(r); !errors.Is(err, want) {
			t.Fatal(err)
		}
	}
	r := decoded(t, c, "b26-t4-rdr-mme")
	facts := s6cT4Evidence(r)
	ctx := s6cT4Context(t, facts)
	r.Body.(*s6c.RDR).DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	a := decoded(t, c, "d21-t4-rdr-a")
	a.Body.(*s6c.RDA).UserIdentifier = avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21}, Extra: []avp.Raw{{Key: avp.Key{Code: 1}, Data: []byte("001010123456789")}}})
	if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("RDA identity scope broadened", err)
	}
	// HSS no-event shape cannot carry a moved-to node either.
	h := decoded(t, c, "d21-t4-alr-r")
	h.Body.(*s6c.ALR).ServingNode = decoded(t, c, "v-alr-move-01").Body.(*s6c.ALR).ServingNode
	if err := c.Validate(h); !errors.Is(err, validation.ErrField) {
		t.Fatal(err)
	}
}
