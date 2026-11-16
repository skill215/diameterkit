// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"reflect"
	"strings"
	"testing"
)

func specialRoutingContext(t *testing.T, r pdu.Message, mcc bool, prefix avp.Optional[string]) validation.Context {
	t.Helper()
	b := r.Body.(*s6c.SRR)
	src, dst := pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}
	var ctx validation.Context
	var err error
	if mcc {
		ctx, err = s6c.MCCMNCRouting(s6c.MCCMNCRoutingEvidence{Requester: src, Responder: dst, MCCMNC: prefix})
	} else {
		ctx, err = s6c.RouterRouting(s6c.RouterRoutingEvidence{GMSC: src, Router: dst, HPLMNEnforced: true})
	}
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}
func TestSpecialRoutingPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, tc := range []struct {
		r       string
		answers []string
		mcc     bool
		prefix  string
	}{
		{"router-r", []string{"router-a", "numeric-a", "unicode-a", "max-a", "failure", "e", "absent-a"}, false, ""},
		{"mcc-r", []string{"mcc-a", "mcc-node-a", "mcc-short-a", "numeric-a", "failure", "e", "absent-a"}, true, "00101"},
		{"mcc-r", []string{"mnc3-a"}, true, "001001"},
	} {
		t.Run(tc.r+tc.prefix, func(t *testing.T) {
			r := decoded(t, c, "b76-routing-"+tc.r)
			ctx := specialRoutingContext(t, r, tc.mcc, avp.Some(tc.prefix))
			if v := c.ValidateContext(r, ctx); !v.Complete {
				t.Fatal(v)
			}
			for _, n := range tc.answers {
				t.Run(n, func(t *testing.T) {
					a := decoded(t, c, "b76-routing-"+n)
					for _, v := range []validation.Report{c.ValidateAnswerContext(r, a, routingSources(), ctx), c.ValidateResponseContext(r, a, routingSources(), ctx)} {
						if !v.Complete {
							t.Fatal(v)
						}
					}
					for _, m := range []pdu.Message{r, a} {
						raw := encoded(t, c, m)
						again, err := c.Decode(raw)
						if err != nil {
							t.Fatal(err)
						}
						clear(raw)
						if !reflect.DeepEqual(m, again) {
							t.Fatal("borrowed data")
						}
						sz, err := c.Measure(m)
						if err != nil {
							t.Fatal(err)
						}
						if bound, err := c.DecodeBound(m.Header, sz); err != nil || bound < uint64(sz) {
							t.Fatal(bound, err)
						}
					}
					a.Header.EndToEnd++
					if err := c.ValidateAnswer(r, a, routingSources()); !errors.Is(err, validation.ErrUnmatchedAnswer) {
						t.Fatal(err)
					}
				})
			}
		})
	}
}
func TestSpecialRoutingConditions(t *testing.T) {
	c := newCodec(t, codec.Config{})
	router := decoded(t, c, "b76-routing-router-r")
	query := decoded(t, c, "b76-routing-mcc-r")
	rctx := specialRoutingContext(t, router, false, avp.Optional[string]{})
	qctx := specialRoutingContext(t, query, true, avp.Some("00101"))
	unknown := specialRoutingContext(t, query, true, avp.Optional[string]{})
	if v := c.ValidateContext(query, validation.Context{}); v.State != validation.NeedsApplication {
		t.Fatal(v)
	}
	if v := c.ValidateContext(query, unknown); !v.Complete {
		t.Fatal(v)
	}
	for _, tc := range []struct {
		r, a  pdu.Message
		ctx   validation.Context
		state validation.State
	}{
		{router, decoded(t, c, "b76-routing-router-a"), validation.Context{}, validation.NeedsApplication},
		{query, decoded(t, c, "b76-routing-mcc-a"), unknown, validation.NeedsApplication},
		{query, decoded(t, c, "b76-routing-failure"), unknown, validation.Ready},
		{query, decoded(t, c, "b76-routing-router-a"), qctx, validation.Rejected},
		{query, decoded(t, c, "b76-routing-prefix-only-a"), qctx, validation.Rejected},
		{query, decoded(t, c, "b76-routing-wrong-prefix-a"), qctx, validation.Rejected},
		{router, decoded(t, c, "b76-routing-router-nonode"), rctx, validation.Rejected},
		{router, decoded(t, c, "b76-routing-mcc-a"), rctx, validation.Rejected},
		{query, decoded(t, c, "b76-routing-mcc-a"), rctx, validation.Rejected},
		{router, decoded(t, c, "b76-routing-router-a"), qctx, validation.Rejected},
	} {
		if v := c.ValidateAnswerContext(tc.r, tc.a, routingSources(), tc.ctx); v.State != tc.state {
			t.Fatal(tc.state, v)
		}
	}
	// No-context numeric ordinary responses retain legacy semantics; numeric bytes
	// alone cannot prove a Router token or dummy MSIN. Explicit purposes isolate use.
	if v := c.ValidateAnswerContext(router, decoded(t, c, "b76-routing-numeric-a"), routingSources(), validation.Context{}); !v.Complete {
		t.Fatal(v)
	}
	for _, r := range []pdu.Message{decoded(t, c, "d21-t4-srr-r"), decoded(t, c, "b76-routing-mcc-r")} {
		if v := c.ValidateContext(r, rctx); v.Complete {
			t.Fatal("cross-purpose accepted")
		}
	}
	wrong := query
	wrong.Body.(*s6c.SRR).DestinationHost = avp.Some("dra.example")
	if v := c.ValidateContext(wrong, qctx); v.Complete {
		t.Fatal("DRA substituted")
	}
	if v := c.ValidateContext(decoded(t, c, "b74-router-r"), rctx); v.Complete {
		t.Fatal("SRR context used for ALR")
	}
}
func TestSpecialRoutingWireAndBudget(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, n := range []string{"only-imsi-r", "reserved-r", "negative-r", "duplicate-purpose-r", "short-purpose-r", "long-a", "nul-a", "empty-a", "utf8-a", "missing-name-a", "duplicate-name-a"} {
		t.Run(n, func(t *testing.T) {
			if _, err := c.Decode(fixture(t, "b76-routing-"+n)); err == nil {
				t.Fatal("bad wire accepted")
			}
		})
	}
	r := decoded(t, c, "b76-routing-mcc-r")
	before, err := c.Canonical(encoded(t, c, r), 4096)
	if err != nil {
		t.Fatal(err)
	}
	r.Body.(*s6c.SRR).SMDeliveryNotIntended.Present = false
	after, err := c.Canonical(encoded(t, c, r), 4096)
	if err != nil || bytes.Equal(before, after) {
		t.Fatal("purpose omitted from duplicate comparison", err)
	}
	r.Body.(*s6c.SRR).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3311}, Flags: 128, Data: []byte{0, 0, 0, 1}}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal(err)
	}
	src, dst := pdu.Origin{Host: "gmsc.example", Realm: "example.net"}, pdu.Origin{Host: "router.example", Realm: "example.net"}
	if _, err := s6c.RouterRouting(s6c.RouterRoutingEvidence{GMSC: src, Router: dst}); !errors.Is(err, validation.ErrContext) {
		t.Fatal(err)
	}
	for _, prefix := range []string{"", "0010", "0010100", "001a1"} {
		if _, err := s6c.MCCMNCRouting(s6c.MCCMNCRoutingEvidence{Requester: src, Responder: dst, MCCMNC: avp.Some(prefix)}); !errors.Is(err, validation.ErrContext) {
			t.Fatal(prefix, err)
		}
	}
	if _, err := s6c.RouterRouting(s6c.RouterRoutingEvidence{GMSC: src, Router: pdu.Origin{Host: strings.Repeat("x", 254), Realm: dst.Realm}, HPLMNEnforced: true}); !errors.Is(err, validation.ErrContext) {
		t.Fatal(err)
	}
	a, err := s6c.MCCMNCRouting(s6c.MCCMNCRoutingEvidence{Requester: src, Responder: dst, MCCMNC: avp.Optional[string]{Value: "discarded"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s6c.MCCMNCRouting(s6c.MCCMNCRoutingEvidence{Requester: src, Responder: dst})
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("absent payload retained", err)
	}
	if n := reflect.TypeFor[validation.Context]().Size() + 4*253 + 6; n > 2048 {
		t.Fatal("context budget", n)
	}
}
