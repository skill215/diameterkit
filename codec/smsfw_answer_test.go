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
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"math"
	"os"
	"strings"
	"testing"
)

type answerFixture struct{ File, Request, Expect string }

func answerFixtures(t *testing.T) []answerFixture {
	t.Helper()
	b, e := os.ReadFile("../testdata/codec/manifest.json")
	if e != nil {
		t.Fatal(e)
	}
	var all []answerFixture
	if e = json.Unmarshal(b, &all); e != nil {
		t.Fatal(e)
	}
	var out []answerFixture
	for _, x := range all {
		if strings.HasPrefix(x.File, "b48-") {
			out = append(out, x)
		}
	}
	if len(out) != 67 {
		t.Fatal(len(out))
	}
	return out
}
func answerFields(b pdu.Body) (uint32, avp.Optional[uint32]) {
	switch a := b.(type) {
	case *smsfwro.CCA:
		if a.MultipleServicesCreditControl.Present {
			return a.ResultCode, avp.Some(a.MultipleServicesCreditControl.Value.ResultCode)
		}
		return a.ResultCode, avp.Optional[uint32]{}
	case *legacy420.CCA:
		if a.MultipleServicesCreditControl.Present {
			return a.ResultCode, avp.Some(a.MultipleServicesCreditControl.Value.ResultCode)
		}
		return a.ResultCode, avp.Optional[uint32]{}
	case *combined.CCA:
		if a.MultipleServicesCreditControl.Present {
			return a.ResultCode, avp.Some(a.MultipleServicesCreditControl.Value.ResultCode)
		}
		return a.ResultCode, avp.Optional[uint32]{}
	}
	panic("wrong body")
}
func mutateAnswer(b pdu.Body) {
	switch a := b.(type) {
	case *smsfwro.CCA:
		a.ResultCode = 4012
		a.MultipleServicesCreditControl.Value.ResultCode = 4012
	case *legacy420.CCA:
		a.ResultCode = 4012
		a.MultipleServicesCreditControl.Value.ResultCode = 4012
	case *combined.CCA:
		a.ResultCode = 4012
		a.MultipleServicesCreditControl.Value.ResultCode = 4012
	}
}
func TestLegacyAnswerPairsAcrossProfiles(t *testing.T) {
	for _, profile := range []string{"sms", "money", "combined"} {
		t.Run(profile, func(t *testing.T) {
			var c *codec.Codec
			switch profile {
			case "sms":
				c = legacyCodec(t)
			case "money":
				c = refundCodec(t)
			default:
				c = combinedCodec(t)
			}
			for _, f := range answerFixtures(t) {
				if f.Expect != "pair" {
					continue
				}
				money := strings.Contains(f.Request, "general") || strings.Contains(f.Request, "balance") || strings.Contains(f.Request, "refund")
				if profile == "sms" && money || profile == "money" && !money {
					continue
				}
				t.Run(f.File, func(t *testing.T) {
					r := decoded(t, c, strings.TrimSuffix(f.Request, ".hex"))
					raw := fixture(t, strings.TrimSuffix(f.File, ".hex"))
					a, e := c.Decode(raw)
					if e != nil {
						t.Fatal(e)
					}
					top, child := answerFields(a.Body)
					switch {
					case strings.Contains(f.File, "child-denied"):
						if top != 2001 || !child.Present || child.Value != 4012 {
							t.Fatal("levels collapsed", top, child)
						}
					case strings.Contains(f.File, "top-failure"):
						if top != 5012 || !child.Present || child.Value != 2001 {
							t.Fatal("child overrode failure", top, child)
						}
					case strings.Contains(f.File, "child-success"):
						if top != 2001 || !child.Present || child.Value != 2001 {
							t.Fatal(top, child)
						}
					default:
						if child.Present {
							t.Fatal("invented child success")
						}
					}
					if e = c.ValidateAnswer(r, a, roSources()); e != nil {
						t.Fatal(e)
					}
					if v := c.ValidateAnswerContext(r, a, roSources(), validation.Context{}); !v.Complete {
						t.Fatal(v)
					}
					out := encoded(t, c, a)
					n, e := c.Measure(a)
					if e != nil || n != len(out) {
						t.Fatal(n, e)
					}
					bound, e := c.DecodeBound(a.Header, n)
					if e != nil || bound < uint64(n) {
						t.Fatal(bound, e)
					}
					snap, e := c.Snapshot(a)
					if e != nil {
						t.Fatal(e)
					}
					clear(raw)
					mutateAnswer(a.Body)
					if !bytes.Equal(out, encoded(t, c, snap)) {
						t.Fatal("answer alias")
					}
					again, e := c.Decode(out)
					if e != nil {
						t.Fatal(e)
					}
					if e = c.ValidateAnswer(r, again, roSources()); e != nil {
						t.Fatal(e)
					}
				})
			}
		})
	}
}
func TestLegacyAnswerBoundaries(t *testing.T) {
	for _, c := range []*codec.Codec{legacyCodec(t), refundCodec(t), combinedCodec(t)} {
		for _, f := range answerFixtures(t) {
			if f.Expect != "reject" || f.File == "b48-unselected-5011.hex" || f.File == "b48-experimental-5011.hex" {
				continue
			}
			a, e := c.Decode(fixture(t, strings.TrimSuffix(f.File, ".hex")))
			if e == nil { // Combined union defers original-family financial fields to relation.
				r := decoded(t, c, "b47-iec-request")
				e = c.ValidateAnswer(r, a, roSources())
			}
			if e == nil {
				t.Fatal("negative accepted", f.File)
			}
		}
	}
	for _, c := range []*codec.Codec{refundCodec(t), combinedCodec(t)} {
		for _, kind := range []string{"balance", "refund"} {
			r := decoded(t, c, "b47-"+kind+"-request")
			a := decoded(t, c, "b48-"+kind+"-ddfh")
			if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
				t.Fatal(kind, e)
			}
			if _, e := c.CompleteAnswer(r, a.Body, pdu.Origin{Host: "ocs.example", Realm: "example.net"}); e == nil {
				t.Fatal("wrong-action local answer")
			}
		}
	}
	// Child success cannot satisfy a missing mandatory command result. Malformed
	// child length must also be rejected before any successful correlation.
	c := combinedCodec(t)
	r := decoded(t, c, "b47-iec-request")
	for _, n := range []string{"missing-top", "child-width", "duplicate-mscc"} {
		if _, e := c.DecodeAnswer(r, fixture(t, "b48-"+n), roSources()); e == nil {
			t.Fatal(n)
		}
	}
}
func TestLegacyAnswerCostPresenceAndRequestRelation(t *testing.T) {
	c := legacyCodec(t)
	absent := decoded(t, c, "b48-iec-cost").Body.(*smsfwro.CCA)
	zero := decoded(t, c, "b48-cost-zero").Body.(*smsfwro.CCA)
	limit := decoded(t, c, "b48-cost-limits").Body.(*smsfwro.CCA)
	if absent.CostInformation.Value.UnitValue.Exponent.Present || !zero.CostInformation.Value.UnitValue.Exponent.Present || zero.CostInformation.Value.UnitValue.Exponent.Value != 0 {
		t.Fatal("exponent absence collapsed")
	}
	if limit.CostInformation.Value.UnitValue.ValueDigits != math.MinInt64 || limit.CostInformation.Value.UnitValue.Exponent.Value != math.MinInt32 || limit.CostInformation.Value.CurrencyCode != math.MaxUint32 {
		t.Fatal(limit)
	}
	// The original sequence/source remains mandatory after adding service results.
	r := decoded(t, c, "b47-initial-request")
	a := decoded(t, c, "b48-initial-child-denied")
	a.Body.(*smsfwro.CCA).CCRequestNumber = 1
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	// Explicit local construction uses the same layered fields without replacing top.
	cc := combinedCodec(t)
	q := decoded(t, cc, "b47-general-request")
	ans, e := cc.CompleteAnswer(q, &combined.CCA{AuthApplicationID: 4, CCRequestType: 4, ResultCode: 2001, DirectDebitingFailureHandling: avp.Some(uint32(0)), MultipleServicesCreditControl: avp.Some(combined.ServiceResult{ResultCode: 4012})}, pdu.Origin{Host: "ocs.example", Realm: "example.net"})
	if e != nil {
		t.Fatal(e)
	}
	top, child := answerFields(ans.Body)
	if top != 2001 || child.Value != 4012 {
		t.Fatal("result overwritten")
	}
	// A child failure is a valid answer to a balance request without a fabricated422.
	balance := decoded(t, cc, "b47-balance-request")
	a2 := decoded(t, cc, "b48-balance-child-denied")
	if e = cc.ValidateAnswer(balance, a2, roSources()); e != nil {
		t.Fatal(e)
	}
	a2.Body.(*combined.CCA).CheckBalanceResult = avp.Some(uint32(0))
	if e = cc.ValidateAnswer(balance, a2, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal("failure became enough-credit", e)
	}
}
