package codec_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func abortCodec(t *testing.T) *codec.Codec {
	t.Helper()
	p, e := abort.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}

func TestRoAbortProfileComposition(t *testing.T) {
	for name, makeProfile := range map[string]func() (schema.Profile, error){"standard": ro.Profile, "smsfw": smsfwro.Profile, "legacy420": legacy420.Profile, "combined": combined.Profile} {
		t.Run(name, func(t *testing.T) {
			p, e := makeProfile()
			if e != nil {
				t.Fatal(e)
			}
			a, e := abort.Profile()
			if e != nil {
				t.Fatal(e)
			}
			bp, e := base.Profile()
			if e != nil {
				t.Fatal(e)
			}
			b, e := schema.Compile(bp, p, a)
			if e != nil {
				t.Fatal(e)
			}
			c, e := codec.New(b, codec.Config{})
			if e != nil {
				t.Fatal(e)
			}
			for _, n := range []string{"asr", "asa-2001", "asa-5002", "asa-5012", "asa-e"} {
				if _, e = c.Decode(fixture(t, "b69-"+n)); e != nil {
					t.Fatal(n, e)
				}
			}
		})
	}
}

func TestRoAbortIndependentPairsAndOwnership(t *testing.T) {
	c := abortCodec(t)
	r := decoded(t, c, "b69-asr")
	sources := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "ctf.example", Realm: "example"}}}
	for _, name := range []string{"asr", "asr-min", "asr-unknown-optional", "asa-2001", "asa-5002", "asa-5012", "asa-options", "asa-e", "asa-e-diagnostic", "asa-unknown-optional"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, "b69-"+name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if !bytes.Equal(out, raw) {
				t.Fatal("independent wire changed")
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("input alias")
			}
			snap, e := c.Snapshot(m)
			if e != nil {
				t.Fatal(e)
			}
			if strings.HasPrefix(name, "asa") {
				if e = c.ValidateAnswer(r, m, sources); e != nil {
					t.Fatal(e)
				}
				if report := c.ValidateAnswerContext(r, m, sources, validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
			}
			if v, ok := m.Body.(*abort.ASR); ok && len(v.ProxyInfo) > 0 {
				v.ProxyInfo[0].ProxyState[0] = 99
			}
			if v, ok := m.Body.(*abort.ASA); ok && len(v.ProxyInfo) > 0 {
				v.ProxyInfo[0].ProxyState[0] = 99
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot alias")
			}
			n, e := c.Measure(snap)
			if e != nil || n != len(out) {
				t.Fatal(n, e)
			}
			bound, e := c.DecodeBound(snap.Header, n)
			if e != nil || bound < uint64(n) {
				t.Fatal(bound, e)
			}
		})
	}
	for _, code := range []uint32{abort.Success, abort.UnknownSessionID, abort.UnableToComply} {
		a, e := c.CompleteAnswer(r, &abort.ASA{ResultCode: code}, sources.Terminals[0])
		if e != nil {
			t.Fatal(e)
		}
		v := a.Body.(*abort.ASA)
		if v.OriginStateID.Present || v.UserName.Present {
			t.Fatal("remote state/user implicitly echoed")
		}
		if v.SessionID != r.Body.(*abort.ASR).SessionID || v.ResultCode != code {
			t.Fatal(v)
		}
	}
	if _, e := c.CompleteAnswer(r, &abort.ASA{}, sources.Terminals[0]); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal("implicit success", e)
	}
}

func TestRoAbortIndependentCounterexamples(t *testing.T) {
	c := abortCodec(t)
	for _, dir := range []string{"asr", "asa"} {
		fields := []string{"session", "origin-host", "origin-realm"}
		if dir == "asr" {
			fields = append(fields, "destination-realm", "destination-host", "auth-app")
		} else {
			fields = append(fields, "result")
		}
		for _, f := range fields {
			for kind, want := range map[string]error{"missing": validation.ErrMissing, "duplicate": validation.ErrDuplicate} {
				name := dir + "-" + kind + "-" + f
				t.Run(name, func(t *testing.T) {
					if _, e := c.Decode(fixture(t, "b69-"+name)); !errors.Is(e, want) {
						t.Fatal(e, want)
					}
				})
			}
		}
	}
	for name, want := range map[string]error{
		"asr-wrong-app": validation.ErrField, "asr-width": avp.ErrValue, "asr-unknown-mandatory": validation.ErrUnsupported, "asa-unknown-mandatory": validation.ErrUnsupported,
		"asr-bad-user": avp.ErrValue, "asr-empty-destination": avp.ErrValue, "asa-zero-result": validation.ErrUnsupported, "asa-unselected-result": validation.ErrUnsupported, "asa-normal-3004": validation.ErrUnsupported, "asa-e-success": validation.ErrField, "asa-e-missing-diagnostic": validation.ErrMissing, "asa-e-experimental": validation.ErrUnsupported, "asr-result": validation.ErrUnsupported,
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b69-"+name)); !errors.Is(e, want) {
				t.Fatal(e, want)
			}
		})
	}
	for _, code := range []string{"258", "415", "416", "436", "456", "292", "297"} {
		t.Run("unselected-"+code, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b69-asa-unselected-"+code)); !errors.Is(e, validation.ErrUnsupported) {
				t.Fatal(e)
			}
		})
	}
	r := decoded(t, c, "b69-asr")
	src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "ctf.example", Realm: "example"}}}
	for _, name := range []string{"session", "source", "proxy"} {
		t.Run("association-"+name, func(t *testing.T) {
			a := decoded(t, c, "b69-asa-wrong-"+name)
			if e := c.ValidateAnswer(r, a, src); !errors.Is(e, validation.ErrInvalidAnswer) {
				t.Fatal(e)
			}
		})
	}
	for _, field := range []string{"hop", "end", "app", "command", "flags"} {
		t.Run("header-"+field, func(t *testing.T) {
			a := decoded(t, c, "b69-asa-2001")
			switch field {
			case "hop":
				a.Header.HopByHop++
			case "end":
				a.Header.EndToEnd++
			case "app":
				a.Header.ApplicationID++
			case "command":
				a.Header.CommandCode++
			case "flags":
				a.Header.Flags |= pdu.RequestFlag
			}
			if e := c.ValidateAnswer(r, a, src); e == nil {
				t.Fatal("mismatched header accepted")
			}
		})
	}
}

func TestRoAbortCanonicalAndNoReplay(t *testing.T) {
	c := abortCodec(t)
	raw := fixture(t, "b69-asr")
	original, e := c.Canonical(raw, 4096)
	if e != nil {
		t.Fatal(e)
	}
	r := decoded(t, c, "b69-asr")
	q := r.Body.(*abort.ASR)
	r.Header.HopByHop++
	r.Header.Flags |= pdu.RetransmittedFlag
	q.ProxyInfo[0].ProxyState = []byte{1}
	q.RouteRecord = []string{"other.example"}
	got, e := c.Canonical(encoded(t, c, r), 4096)
	if e != nil || !bytes.Equal(got, original) {
		t.Fatal("arrival envelope became business content", e)
	}
	q.OriginStateID = avp.Some(uint32(1))
	got, e = c.Canonical(encoded(t, c, r), 4096)
	if e != nil || bytes.Equal(got, original) {
		t.Fatal("changed content hidden", e)
	}
	// Profile's replay permission is intentionally left false, independently of canonical support.
	p, e := abort.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	op, ok := b.Lookup((&abort.ASR{}).CommandKey())
	if !ok || op.TransportReplay || op.Canonical == nil {
		t.Fatal("unexpected operation contract")
	}
}
