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

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func TestRoRecoveryCompiledPermission(t *testing.T) {
	for _, makeProfile := range []func() (schema.Profile, error){ro.Profile, smsfwro.Profile, legacy420.Profile, combined.Profile} {
		bp, _ := base.Profile()
		rp, e := makeProfile()
		if e != nil {
			t.Fatal(e)
		}
		b, e := schema.Compile(bp, rp)
		if e != nil {
			t.Fatal(e)
		}
		key := pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}
		op, ok := b.Lookup(key)
		if !ok || !op.TransportReplay || op.Canonical == nil {
			t.Fatal("audited CCR lacks permission")
		}
		op.TransportReplay = false
		frozen, _ := b.Lookup(key)
		if !frozen.TransportReplay {
			t.Fatal("mutable permission")
		}
		key.Request = false
		ans, _ := b.Lookup(key)
		if ans.TransportReplay {
			t.Fatal("answer replay allowed")
		}
	}
	bp, _ := base.Profile()
	sp, _ := scap1.Profile()
	b, e := schema.Compile(bp, sp)
	if e != nil {
		t.Fatal(e)
	}
	for _, key := range []pdu.CommandKey{{Code: 271, Request: true}, {Code: 280, Request: true}} {
		op, ok := b.Lookup(key)
		if !ok || op.TransportReplay {
			t.Fatal("unaudited command replay")
		}
	}
}
func TestRoRecoveryFrozenIdentityPairs(t *testing.T) {
	for _, family := range []string{"standard", "sms", "money", "combined"} {
		t.Run(family, func(t *testing.T) {
			var c *codec.Codec
			var kinds []string
			switch family {
			case "standard":
				c = roCodec(t)
				kinds = []string{"std-iec", "std-initial", "std-term"}
			case "sms":
				c = legacyCodec(t)
				kinds = []string{"iec", "initial", "term"}
			case "money":
				c = refundCodec(t)
				kinds = []string{"general", "balance", "refund"}
			default:
				c = combinedCodec(t)
				kinds = []string{"iec", "initial", "term", "general", "balance", "refund"}
			}
			for _, kind := range kinds {
				t.Run(kind, func(t *testing.T) {
					q := "b47-" + kind + "-request"
					a := "b47-" + kind + "-success"
					switch kind {
					case "std-iec":
						q = "v-iec-ccr-01"
						a = "v-iec-cca-01"
					case "std-initial":
						q = "v-sms-ccr-01"
						a = "v-sms-cca-01"
					case "std-term":
						q = "v-term-ccr-01"
						a = "v-term-cca-01"
					}
					original := fixture(t, q)
					canonical, e := c.Canonical(original, 1<<20)
					if e != nil {
						t.Fatal(e)
					}
					oldAnswer := decoded(t, c, a)
					answer := decoded(t, c, "b50-"+kind+"-answer")
					for _, suffix := range []string{"unwritten", "retransmit"} {
						raw := fixture(t, "b50-"+kind+"-"+suffix)
						req, e := c.Decode(raw)
						if e != nil {
							t.Fatal(e)
						}
						if !bytes.Equal(raw[20:], original[20:]) || !bytes.Equal(raw[16:20], original[16:20]) || (req.Header.Flags&pdu.RetransmittedFlag != 0) != (suffix == "retransmit") {
							t.Fatal("identity/content/T changed")
						}
						key, e := c.Canonical(raw, 1<<20)
						if e != nil || !bytes.Equal(key, canonical) {
							t.Fatal("header entered duplicate content", e)
						}
						if e = c.ValidateAnswer(req, answer, roSources()); e != nil {
							t.Fatal(e)
						}
						if report := c.ValidateAnswerContext(req, answer, roSources(), validation.Context{}); !report.Complete {
							t.Fatal(report)
						}
						if e = c.ValidateAnswer(req, oldAnswer, roSources()); !errors.Is(e, validation.ErrUnmatchedAnswer) {
							t.Fatal("wrong HBH correlated", e)
						}
						// The separate money profile has its own fixed field order;
						// replay preserves that profile's encoded frame, not a different sender's order.
						out := encoded(t, c, req)
						initial := encoded(t, c, decoded(t, c, q))
						if !bytes.Equal(out[16:], initial[16:]) {
							t.Fatal("profile encoding changed across attempts")
						}
						back, err := c.Decode(out)
						if err != nil || !bytes.Equal(out, encoded(t, c, back)) || back.Header != req.Header {
							t.Fatal("unstable roundtrip/identity", err)
						}
					}
				})
			}
		})
	}
}
