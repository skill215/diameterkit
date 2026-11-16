// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type resultFixture struct{ File, Request, Expect, Profile string }

func resultFixtures(t *testing.T) []resultFixture {
	t.Helper()
	b, e := os.ReadFile("../testdata/codec/manifest.json")
	if e != nil {
		t.Fatal(e)
	}
	var all, out []resultFixture
	if e = json.Unmarshal(b, &all); e != nil {
		t.Fatal(e)
	}
	for _, f := range all {
		if strings.HasPrefix(f.File, "b49-") {
			out = append(out, f)
		}
	}
	if len(out) != 65 {
		t.Fatal(len(out))
	}
	return out
}
func resultValues(b pdu.Body) (uint32, uint32, uint32, int, pdu.FailedAVP) {
	switch a := b.(type) {
	case *ro.CCA:
		return a.ResultCode, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode, len(a.SupportedFeatures), a.FailedAVP.Value
	case *smsfwro.CCA:
		return a.ResultCode, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode, len(a.SupportedFeatures), a.FailedAVP.Value
	case *legacy420.CCA:
		return a.ResultCode, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode, len(a.SupportedFeatures), a.FailedAVP.Value
	case *combined.CCA:
		return a.ResultCode, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode, len(a.SupportedFeatures), a.FailedAVP.Value
	case *pdu.ErrorAnswer:
		return a.ResultCode, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode, 0, a.FailedAVP.Value
	}
	panic("body")
}
func TestRoResultSafetyPairs(t *testing.T) {
	for _, profile := range []string{"standard", "sms", "money", "combined"} {
		t.Run(profile, func(t *testing.T) {
			var c *codec.Codec
			switch profile {
			case "standard":
				c = roCodec(t)
			case "sms":
				c = legacyCodec(t)
			case "money":
				c = refundCodec(t)
			default:
				c = combinedCodec(t)
			}
			for _, f := range resultFixtures(t) {
				if f.Expect == "reject" || (profile == "standard") != (f.Profile == "standard") {
					continue
				}
				money := strings.Contains(f.Request, "general") || strings.Contains(f.Request, "balance") || strings.Contains(f.Request, "refund")
				if profile == "sms" && money || profile == "money" && !money {
					continue
				}
				t.Run(f.File, func(t *testing.T) {
					r := decoded(t, c, strings.TrimSuffix(f.Request, ".hex"))
					raw := fixture(t, strings.TrimSuffix(f.File, ".hex"))
					original := bytes.Clone(raw)
					a, e := c.DecodeAnswer(r, raw, roSources())
					if e != nil {
						t.Fatal(e)
					}
					report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{})
					switch f.Expect {
					case "pair":
						if !report.Complete {
							t.Fatal(report)
						}
					case "conflict":
						if report.Complete || !errors.Is(report.Err(), validation.ErrResultConflict) {
							t.Fatal(report)
						}
					case "unsupported":
						if report.Complete || !errors.Is(report.Err(), validation.ErrUnsupported) {
							t.Fatal(report)
						}
					}
					top, vendor, code, _, diag := resultValues(a.Body)
					if f.Expect == "conflict" && (top != 2001 || vendor != 10415 || code != 5011) {
						t.Fatal("namespace collapsed", top, vendor, code)
					}
					if strings.Contains(f.File, "-rating") && top != 5031 {
						t.Fatal(top)
					}
					if top == 5031 && len(diag.Data) == 0 {
						t.Fatal("lost diagnostic")
					}
					clear(raw)
					again, e := c.Decode(original)
					if e != nil || !reflect.DeepEqual(a, again) {
						t.Fatal("input alias", e)
					}
					bound, e := c.DecodeBound(a.Header, len(original))
					if e != nil || bound < uint64(len(original)) {
						t.Fatal(bound, e)
					}
					if f.Expect == "pair" {
						out := encoded(t, c, a)
						back, e := c.Decode(out)
						if e != nil || !reflect.DeepEqual(a, back) {
							t.Fatal("roundtrip", e)
						}
						snap, e := c.Snapshot(a)
						if e != nil || !bytes.Equal(out, encoded(t, c, snap)) {
							t.Fatal(e)
						}
						if _, e = c.CompleteAnswer(r, a.Body, pdu.Origin{Host: "ocs.example", Realm: "example.net"}); e != nil {
							t.Fatal(e)
						}
					} else {
						if _, e = c.Measure(a); e == nil {
							t.Fatal("unreviewed send accepted")
						}
						if _, e = c.CompleteAnswer(r, a.Body, pdu.Origin{Host: "ocs.example", Realm: "example.net"}); e == nil {
							t.Fatal("automatic pairing accepted")
						}
					}
					wrong := a
					wrong.Header.EndToEnd++
					if e = c.ValidateAnswer(r, wrong, roSources()); !errors.Is(e, validation.ErrUnmatchedAnswer) {
						t.Fatal(e)
					}
					if e = c.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "other.example", Realm: "example.net"}}}); !errors.Is(e, validation.ErrSource) {
						t.Fatal(e)
					}
				})
			}
		})
	}
}
func TestRoResultSafetyMalformedAndSendBoundaries(t *testing.T) {
	for _, c := range []*codec.Codec{legacyCodec(t), refundCodec(t), combinedCodec(t)} {
		for _, f := range resultFixtures(t) {
			if f.Expect != "reject" {
				continue
			}
			if _, e := c.Decode(fixture(t, strings.TrimSuffix(f.File, ".hex"))); e == nil {
				t.Fatal("malformed accepted", f.File)
			}
		}
	}
	// Historical b48 negatives have new receive behavior; their bytes never change.
	for _, c := range []*codec.Codec{legacyCodec(t), refundCodec(t), combinedCodec(t)} {
		a := decoded(t, c, "b48-unselected-5011")
		top, vendor, code, _, _ := resultValues(a.Body)
		if top != 5011 || vendor != 0 || code != 0 {
			t.Fatal("base5011 mapped to feature failure")
		}
		a = decoded(t, c, "b48-experimental-5011")
		if _, e := c.Measure(a); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
	}
	c := combinedCodec(t)
	r := decoded(t, c, "b47-iec-request")
	for _, b := range []pdu.Body{&combined.CCA{ResultCode: 5031, AuthApplicationID: 4, CCRequestType: 4}, &pdu.ErrorAnswer{ResultCode: 5031}} {
		if _, e := c.CompleteAnswer(r, b, pdu.Origin{Host: "ocs.example", Realm: "example.net"}); !errors.Is(e, validation.ErrMissing) {
			t.Fatal("missing5031 diagnostic", e)
		}
	}
	// None of the four compiled profiles permits E to bypass the normal pairing gate.
	for _, tc := range []struct {
		c *codec.Codec
		r string
	}{{roCodec(t), "v-iec-ccr-01"}, {legacyCodec(t), "b47-iec-request"}, {refundCodec(t), "b47-refund-request"}, {combinedCodec(t), "b47-iec-request"}} {
		r := decoded(t, tc.c, tc.r)
		_, e := tc.c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 5012, ExperimentalResult: avp.Some(pdu.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})}, pdu.Origin{Host: "ocs.example", Realm: "example.net"})
		if !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal("E bypass", e)
		}
	}
}
func TestRoResultSafetyBudgetsAndQuotedData(t *testing.T) {
	bp, _ := base.Profile()
	rp, _ := combined.Profile()
	b, e := schema.Compile(bp, rp)
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"b49-iec-rating", "b49-features-multiple"} {
		raw := fixture(t, name)
		c, e := codec.New(b, codec.Config{})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = c.Decode(raw); e != nil {
			t.Fatal(e)
		}
		l := wire.DefaultLimits()
		l.MaxAVPs = 0
		d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
		for {
			_, ok, err := d.Next()
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				break
			}
			l.MaxAVPs++
		}
		l.MaxRepeated = l.MaxAVPs
		small, e := codec.New(b, codec.Config{Limits: l})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = small.Decode(raw); !errors.Is(e, wire.ErrLimit) {
			t.Fatal("inner fields escaped count", e)
		}
	}
	c := combinedCodec(t)
	a := decoded(t, c, "b49-rating-quoted-incomplete").Body.(*combined.CCA)
	if !a.FailedAVP.Value.Incomplete || !bytes.Equal(a.FailedAVP.Value.Data, []byte{1, 2, 3}) {
		t.Fatal(a.FailedAVP)
	}
	// Decoder owns lists, flags, unknown bits and every received result value.
	a = decoded(t, c, "b49-features-m1").Body.(*combined.CCA)
	if len(a.SupportedFeatures) != 1 || !a.SupportedFeatures[0].Mandatory || a.SupportedFeatures[0].FeatureList != 0x80000000 {
		t.Fatal(a.SupportedFeatures)
	}
	a = decoded(t, c, "b49-exp-unknown-vendor").Body.(*combined.CCA)
	if a.ExperimentalResult.Value.VendorID != 999 {
		t.Fatal(a.ExperimentalResult)
	}
	r := decoded(t, c, "b47-iec-request")
	m := decoded(t, c, "b49-iec-conflict")
	m.Body.(*combined.CCA).CCRequestNumber++
	if e := c.ValidateAnswer(r, m, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal("conflict bypassed sequence", e)
	}
}
