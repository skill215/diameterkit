package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/sgdalert"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func legacyAlertCodec(t *testing.T) *codec.Codec {
	t.Helper()
	var ps []schema.Profile
	for _, f := range []func() (schema.Profile, error){base.Profile, s6c.Profile, sgd.Profile, sgdalert.Profile} {
		p, e := f()
		if e != nil {
			t.Fatal(e)
		}
		ps = append(ps, p)
	}
	b, e := schema.Compile(ps...)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}

func TestLegacySGdAlertFrozenVectors(t *testing.T) {
	c := legacyAlertCodec(t)
	for _, name := range []string{"v-alr-01", "v-ala-01", "v-ala-02", "v-alr-move-01", "v-ala-move-01", "v-ala-move-02", "v21-sgsn-alr-r", "v21-sgsn-alr-a", "v21-sgsn-alr-f", "v21-ims-alr-r", "v21-ims-alr-a", "v21-ims-alr-f", "d21-t4-alr-r", "d21-t4-alr-a", "d21-t4-alr-f", "b23-alr-reachable", "b23-alr-move-nonode", "b23-alr-both", "b23-alr-future-bits", "b23-alr-zero", "b23-alr-noevent", "b23-alr-reachable-node", "v21-router-alr-r", "n26-alr-originuri", "n26-alr-hssid-nested", "n27-alr-uri3327"} {
		t.Run(name, func(t *testing.T) {
			original := fixture(t, name)
			raw := fixture(t, "b73-"+name)
			if !bytes.Equal(original[:8], raw[:8]) || !bytes.Equal(original[12:], raw[12:]) {
				t.Fatal("projection changed fields/identifiers")
			}
			s, se := c.Decode(original)
			m, e := c.Decode(raw)
			if (se == nil) != (e == nil) || se != nil && se.Error() != e.Error() {
				t.Fatalf("standard %v legacy %v", se, e)
			}
			if e != nil {
				return
			}
			if m.Body.CommandKey().ApplicationID != sgdalert.ApplicationID {
				t.Fatal(m)
			}
			switch m.Body.(type) {
			case *sgdalert.ALR, *sgdalert.ALA, *pdu.ErrorAnswer:
			default:
				t.Fatal("wrong body type")
			}
			if (c.Validate(m) == nil) != (c.Validate(s) == nil) {
				t.Fatal("send rules drift")
			}
			// Future bits are receive-only, exactly as in the source profile.
			if c.Validate(m) == nil {
				want := encoded(t, c, s) // Source E grammar may complete Auth-Session-State.
				binary.BigEndian.PutUint32(want[8:12], sgdalert.ApplicationID)
				if got := encoded(t, c, m); !bytes.Equal(want, got) {
					t.Fatal("wire differs from source grammar")
				}
				copy, e := c.Snapshot(m)
				if e != nil {
					t.Fatal(e)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, copy), encoded(t, c, m)) {
					t.Fatal("borrowed source")
				}
			}
			x, e := c.DecodeBound(m.Header, len(original))
			y, se := c.DecodeBound(s.Header, len(original))
			if e != nil || se != nil || x != y {
				t.Fatal(x, y, e, se)
			}
			if m.Header.Flags&pdu.RequestFlag != 0 {
				a, e := c.Canonical(fixture(t, "b73-"+name), 8192)
				b, se := c.Canonical(original, 8192)
				if (e == nil) != (se == nil) || e == nil && bytes.Equal(a, b) {
					t.Fatal("application identity lost in duplicate comparison", e, se)
				}
			}
		})
	}
}

func TestLegacySGdAlertPairsAndOwnership(t *testing.T) {
	c := legacyAlertCodec(t)
	for _, pair := range [][2]string{{"v-alr-01", "v-ala-01"}, {"v-alr-01", "v-ala-02"}, {"v-alr-move-01", "v-ala-move-01"}, {"v-alr-move-01", "v-ala-move-02"}, {"v21-sgsn-alr-r", "v21-sgsn-alr-a"}, {"v21-sgsn-alr-r", "v21-sgsn-alr-f"}} {
		t.Run(pair[1], func(t *testing.T) {
			r, a := decoded(t, c, "b73-"+pair[0]), decoded(t, c, "b73-"+pair[1])
			b := r.Body.(*sgdalert.ALR)
			role := s6c.AlertFromHSS
			if b.UserIdentifier.UserName.Present {
				role = s6c.AlertFromSGSN
			}
			ctx, e := s6c.AlertRole(s6c.AlertRoleEvidence{Role: role, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}})
			if e != nil {
				t.Fatal(e)
			}
			sources := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: b.DestinationHost, Realm: b.DestinationRealm}}, ErrorAgents: []pdu.Origin{{Host: "dra.example", Realm: "example"}}}
			if report := c.ValidateAnswerContext(r, a, sources, ctx); !report.Complete {
				t.Fatal(report)
			}
			a.Header.ApplicationID = 16777312
			if e := c.ValidateAnswer(r, a, sources); !errors.Is(e, validation.ErrUnmatchedAnswer) {
				t.Fatal(e)
			}
		})
	}
	r := decoded(t, c, "b73-v-alr-01")
	b := r.Body.(*sgdalert.ALR)
	copy, e := c.Snapshot(r)
	if e != nil {
		t.Fatal(e)
	}
	b.SCAddress[0] ^= 0xff
	b.UserIdentifier.MSISDN.Value[0] ^= 0xff
	if !bytes.Equal(encoded(t, c, copy), fixture(t, "b73-v-alr-01")) {
		t.Fatal("clone borrows nested fields")
	}
	for _, answer := range []pdu.Body{&sgdalert.ALA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}, &sgdalert.ALA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 3004}, &pdu.ErrorAnswer{ResultCode: 5012}} {
		a, e := c.CompleteAnswer(copy, answer, pdu.Origin{Host: "local.example", Realm: "example"})
		if e != nil {
			t.Fatal(e)
		}
		if a.Header.ApplicationID != sgdalert.ApplicationID || a.Body.CommandKey() != (*sgdalert.ALA)(nil).CommandKey() {
			t.Fatal(a)
		}
		raw := encoded(t, c, a)
		decoded, e := c.Decode(raw)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(raw, encoded(t, c, decoded)) {
			t.Fatal("E/normal roundtrip")
		}
		if x, ok := answer.(*pdu.ErrorAnswer); ok && x.Command != (pdu.CommandKey{}) {
			t.Fatal("mutated caller error template")
		}
	}
}

func TestLegacySGdAlertExplicitBinding(t *testing.T) {
	c := legacyAlertCodec(t)
	plain := newCodec(t, codec.Config{})
	if _, e := plain.Decode(fixture(t, "b73-v-alr-01")); e == nil {
		t.Fatal("implicit alias")
	}
	p, e := sgdalert.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := b.Lookup((*s6c.ALR)(nil).CommandKey()); ok {
		t.Fatal("standard app rebound")
	}
	for _, code := range []uint32{8388645, 8388646, 8388647, 8388649} {
		if _, ok := b.Lookup(pdu.CommandKey{ApplicationID: sgdalert.ApplicationID, Code: code, Request: true}); ok {
			t.Fatal("non-alert aliased")
		}
	}
	if _, e := schema.Compile(p, p); e == nil {
		t.Fatal("duplicate profile")
	}
	again, _ := sgdalert.Profile()
	bb, _ := schema.Compile(again)
	if b.Fingerprint() != bb.Fingerprint() {
		t.Fatal("unstable fingerprint")
	}
	r := decoded(t, c, "b73-v-alr-01")
	if _, e := c.CompleteAnswer(r, &s6c.ALA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}, pdu.Origin{Host: "a.example", Realm: "example"}); e == nil {
		t.Fatal("cross-bound answer")
	}
	for _, body := range []pdu.Body{(*sgdalert.ALR)(nil), (*sgdalert.ALA)(nil), &pdu.ErrorAnswer{Command: (*s6c.ALA)(nil).CommandKey(), ResultCode: 3004}} {
		key := (*sgdalert.ALR)(nil).CommandKey()
		if _, ok := body.(*sgdalert.ALR); !ok {
			key.Request = false
		}
		op, _ := b.Lookup(key)
		var err error
		if _, ok := body.(*pdu.ErrorAnswer); ok {
			err = op.Error.Validate(body, validation.Build)
		} else {
			err = op.Validate(body, validation.Build)
		}
		if !errors.Is(err, validation.ErrType) {
			t.Fatal(err)
		}
	}
	router := decoded(t, c, "b73-v21-router-alr-r")
	if c.ValidateContext(router, validation.Context{}).Complete {
		t.Fatal("Router token without evidence")
	}
}

func TestLegacySGdAlertIMSAndT4Context(t *testing.T) {
	c := legacyAlertCodec(t)
	for _, prefix := range []string{"v21-ims-alr", "d21-t4-alr"} {
		t.Run(prefix, func(t *testing.T) {
			original := decoded(t, c, prefix+"-r")
			r := decoded(t, c, "b73-"+prefix+"-r")
			b := r.Body.(*sgdalert.ALR)
			var ctx validation.Context
			if prefix == "v21-ims-alr" {
				ctx = imsAlertContext(t, imsAlertEvidence(original))
			} else {
				ctx = s6cT4Context(t, s6cT4Evidence(original))
			}
			if c.ValidateContext(r, validation.Context{}).Complete {
				t.Fatal("missing purpose evidence accepted")
			}
			if v := c.ValidateContext(r, ctx); !v.Complete {
				t.Fatal(v)
			}
			sources := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: b.DestinationHost, Realm: b.DestinationRealm}}}
			for _, suffix := range []string{"-a", "-f"} {
				a := decoded(t, c, "b73-"+prefix+suffix)
				if v := c.ValidateAnswerContext(r, a, sources, ctx); !v.Complete {
					t.Fatal(v)
				}
				if v := c.ValidateResponseContext(r, a, sources, ctx); !v.Complete {
					t.Fatal(v)
				}
			}
		})
	}
}
