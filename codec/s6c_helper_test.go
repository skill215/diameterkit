// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"testing"
)

func TestPureRoutingSpecialIdentity(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, mcc := range []bool{false, true} {
		rn, an, kind := "router-r", "router-a", s6c.IdentityRouterToken
		if mcc {
			rn, an, kind = "mcc-r", "mcc-a", s6c.IdentityMCCMNC
		}
		r := decoded(t, c, "b76-routing-"+rn)
		a := decoded(t, c, "b76-routing-"+an)
		ctx := specialRoutingContext(t, r, mcc, avp.Some("00101"))
		out, e := s6c.InterpretSRA(s6c.RoutingContext{Request: r.Body.(*s6c.SRR), Validation: ctx}, a.Body)
		if e != nil || out.Identity.Kind != kind || out.Outcome != s6c.RoutingSuccess {
			t.Fatal(out, e)
		}
	}
	r := decoded(t, c, "v21-ims-srr-r")
	a := decoded(t, c, "v21-ims-srr-a")
	ctx := imsRoutingContext(t, imsRoutingEvidence(r))
	out, e := s6c.InterpretSRA(s6c.RoutingContext{Request: r.Body.(*s6c.SRR), Validation: ctx}, a.Body)
	if e != nil || out.Identity.Kind != s6c.IdentityIMSI || out.Identity.Value != a.Body.(*s6c.SRA).UserName {
		t.Fatal(out, e)
	}
}

func TestNewSRRSelectedIdentityShapes(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v21-ims-srr-r", "b76-routing-router-r", "b76-routing-mcc-r", "d21-t4-srr-r"} {
		t.Run(name, func(t *testing.T) {
			original := decoded(t, c, name).Body.(*s6c.SRR)
			fields := *original
			sc, e := avp.DecodeTBCD(fields.SCAddress)
			if e != nil {
				t.Fatal(e)
			}
			q := s6c.RoutingQuery{SessionID: fields.SessionID, SCAddress: sc}
			if fields.MSISDN.Present {
				n, e := avp.DecodeTBCD(fields.MSISDN.Value)
				if e != nil {
					t.Fatal(e)
				}
				q.MSISDN = avp.Some(n)
			}
			fields.SessionID = ""
			fields.AuthSessionState = 0
			fields.SCAddress = nil
			fields.MSISDN = avp.Optional[[]byte]{}
			q.Fields = fields
			r, e := s6c.NewSRR(q)
			if e != nil {
				t.Fatal(e)
			}
			if r.UserName != original.UserName || r.SMSMICorrelationID.Present != original.SMSMICorrelationID.Present {
				t.Fatal("identity changed")
			}
		})
	}
}
