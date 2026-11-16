package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func accessContext(t *testing.T, r pdu.Message, role s6c.AlertOriginRole) validation.Context {
	t.Helper()
	b := r.Body.(*s6c.ALR)
	v, e := s6c.AlertRole(s6c.AlertRoleEvidence{Role: role, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestS6cAccessAlertFrozenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, pair := range [][2]string{{"v-alr-move-01", "v-ala-move-01"}, {"v-alr-move-01", "v-ala-move-02"}, {"v21-sgsn-alr-r", "v21-sgsn-alr-a"}, {"v21-sgsn-alr-r", "v21-sgsn-alr-f"}} {
		req, ans := decoded(t, c, pair[0]), decoded(t, c, pair[1])
		b := req.Body.(*s6c.ALR)
		role := s6c.AlertFromMME
		if pair[0] == "v21-sgsn-alr-r" {
			role = s6c.AlertFromSGSN
		}
		ctx := accessContext(t, req, role)
		sources := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: b.DestinationHost, Realm: b.DestinationRealm}}}
		if r := c.ValidateContext(req, validation.Context{}); r.State != validation.NeedsApplication || r.Complete {
			t.Fatal(r)
		}
		if r := c.ValidateContext(req, ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := c.ValidateAnswerContext(req, ans, sources, ctx); !r.Complete {
			t.Fatal(r)
		}
		for _, m := range []pdu.Message{req, ans} {
			raw := encoded(t, c, m)
			a, e := c.Decode(raw)
			if e != nil || !bytes.Equal(raw, encoded(t, c, a)) {
				t.Fatal(e)
			}
			clear(raw)
			if !bytes.Equal(encoded(t, c, m), encoded(t, c, a)) {
				t.Fatal("borrowed data")
			}
			size, e := c.Measure(m)
			if e != nil {
				t.Fatal(e)
			}
			if n, e := c.DecodeBound(m.Header, size); e != nil || n < uint64(size) {
				t.Fatal(n, e)
			}
		}
		x, e := c.Canonical(fixture(t, pair[0]), 8192)
		if e != nil {
			t.Fatal(e)
		}
		y, e := c.Canonical(encoded(t, c, req), 8192)
		if e != nil || !bytes.Equal(x, y) {
			t.Fatal(e)
		}
		ans.Header.EndToEnd++
		if e = c.ValidateAnswer(req, ans, sources); !errors.Is(e, validation.ErrUnmatchedAnswer) {
			t.Fatal(e)
		}
	}
}
func TestS6cAccessAlertEventsAndNodeShapes(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"b23-alr-reachable", "b23-alr-move-nonode", "b23-alr-both", "b23-alr-future-bits"} {
		m := decoded(t, c, name)
		ctx := accessContext(t, m, s6c.AlertFromMME)
		if r := c.ValidateContext(m, ctx); !r.Complete {
			t.Fatal(name, r)
		}
		if name == "b23-alr-future-bits" {
			if m.Body.(*s6c.ALR).SMSGMSCAlertEvent.Value != 0x80000002 {
				t.Fatal("lost bit")
			}
			if e := c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
				t.Fatal(e)
			}
		} else {
			if e := c.Validate(m); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, name := range []string{"b23-alr-zero", "b23-alr-reachable-node"} {
		if _, e := c.Decode(fixture(t, name)); e == nil {
			t.Fatal(name)
		}
	}
	for _, role := range []s6c.AlertOriginRole{s6c.AlertFromMME, s6c.AlertFromSGSN} {
		for n := -1; n < 5; n++ {
			m := decoded(t, c, "v-alr-move-01")
			b := m.Body.(*s6c.ALR)
			b.ServingNode.Present = n >= 0
			if n >= 0 {
				b.ServingNode.Value = routingNode(n)
			}
			r := c.ValidateContext(m, accessContext(t, m, role))
			if r.Complete != (n < 3) {
				t.Fatal(role, n, r)
			}
		}
	}
	for _, f := range []uint32{0, 1, 2, 3, 4, 0x80000001} {
		m := decoded(t, c, "v-alr-move-01")
		b := m.Body.(*s6c.ALR)
		b.ServingNode.Present = false
		b.SMSGMSCAlertEvent = avp.Some(f)
		e := c.Validate(m)
		if (e == nil) != (f >= 1 && f <= 3) {
			t.Fatal(f, e)
		}
	}
}
func TestS6cAccessAlertIdentityAndWireNegatives(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, mutate := range []func(*s6c.ALR){
		func(b *s6c.ALR) { b.UserIdentifier.UserName.Present = false },
		func(b *s6c.ALR) { b.UserIdentifier.MSISDN = avp.Some([]byte{0x21}) },
		func(b *s6c.ALR) { b.UserIdentifier.UserName = avp.Some("router-token") },
		func(b *s6c.ALR) { b.UserIdentifier.UserName = avp.Some("001") },
		func(b *s6c.ALR) { b.MaximumUEAvailabilityTime = avp.Some(avp.Time(0)) },
		func(b *s6c.ALR) { b.ServingNode.Value.MMERealm.Present = false },
		func(b *s6c.ALR) { b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3324}, Flags: 128}} },
	} {
		m := decoded(t, c, "v-alr-move-01")
		mutate(m.Body.(*s6c.ALR))
		if report := c.ValidateContext(m, accessContext(t, m, s6c.AlertFromMME)); report.Complete {
			t.Fatal("bad shape/role")
		}
	}
	for _, path := range [][]uint32{{3102, 1}, {3333}, {2401}, {2401, 2402}} {
		raw := rewriteAVP(t, fixture(t, "v-alr-move-01"), path, func(v []byte) []byte { return append(v, v...) })
		if _, e := c.Decode(raw); e == nil {
			t.Fatal(path)
		}
	}
	for _, event := range []uint32{1, 2} {
		m := decoded(t, c, "v-alr-01")
		m.Body.(*s6c.ALR).SMSGMSCAlertEvent = avp.Some(event)
		if e := c.Validate(m); e == nil {
			t.Fatal("HSS mixed with access event")
		}
	}
	// Typed RDR identity cannot be shadowed through Extra; RDA retains its MSISDN-only identity.
	m := decoded(t, c, "v-rdr-01")
	m.Body.(*s6c.RDR).UserIdentifier.Extra = []avp.Raw{{Key: avp.Key{Code: 1}, Data: []byte("001010123456789")}}
	if e := c.Validate(m); !errors.Is(e, validation.ErrDuplicate) {
		t.Fatal(e)
	}
}
func TestS6cAccessAlertRoleAndContextKinds(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "v-alr-move-01")
	b := req.Body.(*s6c.ALR)
	good := s6c.AlertRoleEvidence{Role: s6c.AlertFromMME, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}}
	for _, mutate := range []func(*s6c.AlertRoleEvidence){func(e *s6c.AlertRoleEvidence) { e.Role = s6c.AlertFromHSS }, func(e *s6c.AlertRoleEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.AlertRoleEvidence) { e.Responder.Host = "other.example" }} {
		v := good
		mutate(&v)
		ctx, e := s6c.AlertRole(v)
		if e != nil {
			t.Fatal(e)
		}
		if r := c.ValidateContext(req, ctx); r.Complete || !errors.Is(r.Err(), validation.ErrContext) {
			t.Fatal(r)
		}
	}
	for _, role := range []s6c.AlertOriginRole{0, 4} {
		v := good
		v.Role = role
		if _, e := s6c.AlertRole(v); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
	}
	v := good
	v.Requester.Host = "bad host"
	if _, e := s6c.AlertRole(v); e == nil {
		t.Fatal("bad endpoint")
	}
	shctx, e := sh.PeerFeatures(sh.PeerFeatureEvidence{CommandCode: 308, Requester: good.Requester, Responder: good.Responder, RequesterKnown: true, ResponderKnown: true})
	if e != nil {
		t.Fatal(e)
	}
	if r := c.ValidateContext(req, shctx); !errors.Is(r.Err(), validation.ErrContext) {
		t.Fatal(r)
	}
	// A typed S6c context must not masquerade as empty context on plain Sh.
	sc := shCodec(t, codec.Config{})
	shreq := decoded(t, sc, "v25-sh-plain-r")
	if r := sc.ValidateContext(shreq, accessContext(t, req, s6c.AlertFromMME)); !errors.Is(r.Err(), validation.ErrContext) {
		t.Fatal(r)
	}
	hss := decoded(t, c, "v-alr-01")
	if r := c.ValidateContext(hss, validation.Context{}); !r.Complete {
		t.Fatal(r)
	}
	if r := c.ValidateContext(hss, accessContext(t, hss, s6c.AlertFromHSS)); !r.Complete {
		t.Fatal(r)
	}
}
