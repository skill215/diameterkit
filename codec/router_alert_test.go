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
	"github.com/skill215/diameterkit/compat/sgdalert"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func routerEvidence(m pdu.Message) s6c.RouterAlertEvidence {
	var b *s6c.ALR
	switch v := m.Body.(type) {
	case *s6c.ALR:
		b = v
	case *sgdalert.ALR:
		b = (*s6c.ALR)(v)
	}
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return s6c.RouterAlertEvidence{Router: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, GMSC: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}, PriorAlertToken: b.UserIdentifier.UserName.Value, PriorGMSCAddress: sc}
}
func routerContext(t *testing.T, e s6c.RouterAlertEvidence) validation.Context {
	t.Helper()
	c, err := s6c.RouterAlert(e)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestRouterAlertPairsAndEvidence(t *testing.T) {
	c := legacyAlertCodec(t)
	for _, prefix := range []string{"b74-router-", "b74-router-legacy-"} {
		t.Run(prefix, func(t *testing.T) {
			r := decoded(t, c, prefix+"r")
			e := routerEvidence(r)
			ctx := routerContext(t, e)
			if v := c.ValidateContext(r, validation.Context{}); v.State != validation.NeedsApplication || v.Complete {
				t.Fatal(v)
			}
			if v := c.ValidateContext(r, ctx); !v.Complete {
				t.Fatal(v)
			}
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.GMSC}}
			for _, name := range []string{"a", "failure", "e"} {
				a := decoded(t, c, prefix+name)
				for _, v := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
					if !v.Complete {
						t.Fatal(name, v)
					}
				}
				a.Header.EndToEnd++
				if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
					t.Fatal(err)
				}
			}
			for _, change := range []func(*s6c.RouterAlertEvidence){func(e *s6c.RouterAlertEvidence) { e.Router.Host = "dra.example" }, func(e *s6c.RouterAlertEvidence) { e.GMSC.Realm = "other.example" }, func(e *s6c.RouterAlertEvidence) { e.PriorAlertToken = "Router-alert-21" }, func(e *s6c.RouterAlertEvidence) { e.PriorGMSCAddress = "12345" }} {
				wrong := e
				change(&wrong)
				if c.ValidateContext(r, routerContext(t, wrong)).Complete {
					t.Fatal("wrong provenance accepted")
				}
			}
			for _, role := range []s6c.AlertOriginRole{s6c.AlertFromHSS, s6c.AlertFromMME, s6c.AlertFromSGSN} {
				wrong, err := s6c.AlertRole(s6c.AlertRoleEvidence{Role: role, Requester: e.Router, Responder: e.GMSC})
				if err != nil {
					t.Fatal(err)
				}
				if c.ValidateContext(r, wrong).Complete {
					t.Fatal("token accepted as IMSI")
				}
			}
			if !bytes.Equal(encoded(t, c, r), fixture(t, prefix+"r")) {
				t.Fatal("wire changed")
			}
			snap, err := c.Snapshot(r)
			if err != nil {
				t.Fatal(err)
			}
			if b, ok := r.Body.(*s6c.ALR); ok {
				b.SCAddress[0] ^= 255
			} else {
				r.Body.(*sgdalert.ALR).SCAddress[0] ^= 255
			}
			if !bytes.Equal(encoded(t, c, snap), fixture(t, prefix+"r")) {
				t.Fatal("borrowed snapshot")
			}
		})
	}
}

func TestRouterAlertTokenSyntaxAndContextBudget(t *testing.T) {
	c := legacyAlertCodec(t)
	for _, name := range []string{"r", "reachable", "both", "move-no-node", "numeric", "short", "unicode", "decomposed", "max"} {
		t.Run(name, func(t *testing.T) {
			m := decoded(t, c, "b74-router-"+name)
			e := routerEvidence(m)
			ctx := routerContext(t, e)
			if !c.ValidateContext(m, ctx).Complete {
				t.Fatal(c.ValidateContext(m, ctx))
			}
			if !bytes.Equal(encoded(t, c, m), fixture(t, "b74-router-"+name)) {
				t.Fatal("token normalized")
			}
			if name == "decomposed" {
				e.PriorAlertToken = "令牌-é"
				if c.ValidateContext(m, routerContext(t, e)).Complete {
					t.Fatal("Unicode normalized")
				}
			}
		})
	}
	for _, name := range []string{"long", "empty", "nul", "invalid-utf8", "zero-event", "reachable-node"} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b74-router-"+name)); e == nil {
				t.Fatal("bad static shape")
			}
		})
	}
	m := decoded(t, c, "b74-router-missing-event")
	if v := c.ValidateContext(m, routerContext(t, routerEvidence(m))); v.Complete {
		t.Fatal("Router event missing")
	}
	e := routerEvidence(decoded(t, c, "b74-router-r"))
	for _, token := range []string{"", strings.Repeat("x", 257), "bad\x00token", "\xff"} {
		wrong := e
		wrong.PriorAlertToken = token
		if _, err := s6c.RouterAlert(wrong); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"", "+123", "1234567890123456"} {
		wrong := e
		wrong.PriorGMSCAddress = bad
		if _, err := s6c.RouterAlert(wrong); err == nil {
			t.Fatal("invalid address")
		}
	}
	wrong := e
	wrong.Router.Host = "https://router.example"
	if _, err := s6c.RouterAlert(wrong); err == nil {
		t.Fatal("invalid endpoint")
	}
	if reflect.TypeFor[validation.Context]().Size()+4*253+s6c.MaxRouterAlertTokenBytes+15 > 2048 {
		t.Fatal("exceeds existing context reserve")
	}
	ctx := routerContext(t, e)
	view, ok := ctx.RouterAlert()
	if !ok {
		t.Fatal("kind")
	}
	view.UserName = "mutated"
	if !c.ValidateContext(decoded(t, c, "b74-router-r"), ctx).Complete {
		t.Fatal("context view mutates retained evidence")
	}
	if _, ok := ctx.S6cT4(); ok {
		t.Fatal("shared purpose")
	}
	if _, ok := ctx.S6cIMSAlert(); ok {
		t.Fatal("shared purpose")
	}
}
