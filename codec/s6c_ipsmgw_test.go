// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/wire"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/validation"
)

func TestS6cIPSMGWSelectedPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v25-srr-two-r")
	for _, name := range []string{"n", "nh", "nb0", "nhb1", "vendorrealm", "unknown"} {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, "b30-ipsm-"+name)
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, routingSources(), validation.Context{}), c.ValidateResponseContext(r, a, routingSources(), validation.Context{})} {
				if !p.Complete {
					t.Fatal(p)
				}
			}
			b := a.Body.(*s6c.SRA)
			node := b.ServingNode.Value
			if !node.IPSMGWNumber.Present || node.IPSMGWRealm.Present {
				t.Fatal(node)
			}
			if name == "nb0" && (!node.IPSMGWSBISupportIndicator.Present || node.IPSMGWSBISupportIndicator.Value != 0) {
				t.Fatal("zero lost")
			}
			if name == "n" && node.IPSMGWSBISupportIndicator.Present {
				t.Fatal("invented SBI")
			}
			raw := encoded(t, c, a)
			owned, err := c.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			clear(raw)
			if !bytes.Equal(encoded(t, c, owned), encoded(t, c, a)) {
				t.Fatal("borrowed")
			}
			n, err := c.Measure(a)
			if err != nil || n != len(encoded(t, c, a)) {
				t.Fatal(n, err)
			}
			if bound, err := c.DecodeBound(a.Header, n); err != nil || bound < uint64(n) {
				t.Fatal(bound, err)
			}
			// A single IP-SM-GW node does not require GPRS or SMEA.
			for _, flags := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(2)), avp.Some(uint32(4))} {
				x, err := c.Snapshot(r)
				if err != nil {
					t.Fatal(err)
				}
				x.Body.(*s6c.SRR).SRRFlags = flags
				x.Body.(*s6c.SRR).SMRPSMEA.Present = false
				if p := c.ValidateAnswerContext(x, a, routingSources(), validation.Context{}); !p.Complete {
					t.Fatal(p)
				}
			}
			// Failure results do not gain productive nodes; missing success IMSI remains invalid.
			b.UserName.Present = false
			if err := c.Validate(a); !errors.Is(err, validation.ErrMissing) {
				t.Fatal(err)
			}
			b.UserName = avp.Some("001010123456789")
			b.ResultCode.Value = 5012
			if err := c.Validate(a); err == nil {
				t.Fatal("failure with node")
			}
		})
	}
}
func TestS6cIPSMGWReceiveAndReportBoundary(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v25-srr-two-r")
	for _, name := range []string{"b30-ipsm-nhr", "b30-ipsm-sbi2", "b30-ipsm-addn", "c24-add-ipsmgw", "c27-realm-nr", "c27-realm-nhrb"} {
		a := decoded(t, c, name)
		// Old candidates use different requests; only relabel an owned ordinary
		// request envelope for this explicitly synthetic relation/control check.
		x, err := c.Snapshot(r)
		if err != nil {
			t.Fatal(err)
		}
		x.Header.HopByHop = a.Header.HopByHop
		x.Header.EndToEnd = a.Header.EndToEnd
		x.Body.(*s6c.SRR).SessionID = a.Body.(*s6c.SRA).SessionID
		if p := c.ValidateAnswerContext(x, a, routingSources(), validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) || p.State == validation.NeedsApplication {
			t.Fatal(name, p)
		}
		if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(name, err)
		}
		// An explicit ordinary requester-role context cannot remove a profile boundary.
		ctx := routingRoleContext(t, routingRoleFacts(x, s6c.RoutingFromGMSC, avp.Optional[bool]{}))
		if p := c.ValidateAnswerContext(x, a, routingSources(), ctx); !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(name, p)
		}
		node := a.Body.(*s6c.SRA).ServingNode.Value
		if name == "b30-ipsm-nhr" && node.IPSMGWRealm.Value != "example.net" {
			t.Fatal(node)
		}
		if name == "b30-ipsm-sbi2" && node.IPSMGWSBISupportIndicator.Value != 2 {
			t.Fatal(node)
		}
		if name == "b30-ipsm-addn" {
			x.Body.(*s6c.SRR).SRRFlags.Present = false
			if err := c.ValidateAnswer(x, a, routingSources()); !errors.Is(err, validation.ErrInvalidAnswer) {
				t.Fatal(err)
			}
		}
	}
	for _, name := range []string{"b30-ipsm-nonumber", "b30-ipsm-badnumber", "b30-ipsm-duprealm", "b30-ipsm-mixed", "n27-realm-text", "n27-realm-nouser"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
}
func TestS6cIPSMGWParentAndT4Isolation(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v25-srr-two-r")
	a := decoded(t, c, "b30-ipsm-nh")
	// Pairing a productive IP-SM-GW with an Additional ordinary node is not selected.
	a.Body.(*s6c.SRA).AdditionalServingNode = avp.Some(s6c.AdditionalServingNode{SGSNNumber: avp.Some([]byte{0x21})})
	if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
	// ALR reuses ServingNode but must not inherit SRA's new node family.
	alert := decoded(t, c, "v-alr-move-01")
	alert.Body.(*s6c.ALR).ServingNode = avp.Some(decoded(t, c, "b30-ipsm-nh").Body.(*s6c.SRA).ServingNode.Value)
	if err := c.Validate(alert); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
	// T4 keeps its original IMSI and purpose checks; node content does not authorize a different user.
	tr := decoded(t, c, "d21-t4-srr-r")
	facts := t4RoutingFacts(tr)
	ctx := s6cT4Context(t, facts)
	body := decoded(t, c, "b30-ipsm-nh").Body.(*s6c.SRA)
	body.SessionID = ""
	body.OriginHost = ""
	body.OriginRealm = ""
	ans, err := c.CompleteAnswer(tr, body, routingSources().Terminals[0])
	if err != nil {
		t.Fatal(err)
	}
	if p := c.ValidateAnswerContext(tr, ans, routingSources(), ctx); !p.Complete {
		t.Fatal(p)
	}
	ans.Body.(*s6c.SRA).UserName.Value = "001010123456780"
	if err := c.ValidateAnswer(tr, ans, routingSources()); !errors.Is(err, validation.ErrInvalidAnswer) {
		t.Fatal(err)
	}
	// Nested leaf cannot be supplied at top-level, or by a vendor-zero shadow.
	wrong := decoded(t, c, "b30-ipsm-n")
	wrong.Body.(*s6c.SRA).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3112}, Flags: 128, Data: []byte("example.net")}}
	if p := c.ValidateAnswerContext(r, wrong, routingSources(), validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
}

func TestS6cIPSMGWLeafCardinalityAndWidth(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, parent := range []uint32{2401, 2406} {
		for _, code := range []uint32{3100, 3101, 3112, 3348} {
			name := "b30-ipsm-nhr"
			if code == 3348 {
				name = "b30-ipsm-nb0"
			}
			raw := fixture(t, name)
			if parent == 2406 {
				raw = rewriteAVP(t, raw, []uint32{2401}, func(v []byte) []byte { binary.BigEndian.PutUint32(v, 2406); return v })
			}
			dup := rewriteAVP(t, raw, []uint32{parent, code}, func(v []byte) []byte { return append(v, v...) })
			if _, err := c.Decode(dup); !errors.Is(err, validation.ErrDuplicate) {
				t.Fatal(parent, code, err)
			}
		}
	}
	for _, size := range []int{0, 3, 5} {
		raw := rewriteAVP(t, fixture(t, "b30-ipsm-nb0"), []uint32{2401, 3348}, func(v []byte) []byte {
			v = append(v[:12], make([]byte, size)...)
			wire.Put24(v[5:8], uint32(len(v)))
			return append(v, make([]byte, (-len(v))&3)...)
		})
		if _, err := c.Decode(raw); err == nil {
			t.Fatal(size)
		}
	}
	raw := fixture(t, "b30-ipsm-nhr")
	a, err := c.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	clear(raw)
	n := a.Body.(*s6c.SRA).ServingNode.Value
	if n.IPSMGWRealm.Value != "example.net" || n.IPSMGWName.Value != "isg.example" || !bytes.Equal(n.IPSMGWNumber.Value, []byte{0x21, 0xf3}) {
		t.Fatal("receive boundary aliases input")
	}
}
