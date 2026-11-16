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
	"testing"
)

func routingSources() pdu.AnswerSources {
	return pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "hss.example", Realm: "example.net"}}}
}

var routingAnswers = []string{"v25-sra-sgsn-msc", "v25-sra-mme", "v25-sra-sgsn-n", "v25-sra-sgsn-3", "v25-sra-msc", "v25-sra-msc-comb"}

func TestS6cRoutingFrozenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "v25-srr-two-r")
	for _, name := range append([]string{"v25-srr-two-r"}, routingAnswers...) {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			again, e := c.Decode(out)
			if e != nil || !bytes.Equal(out, encoded(t, c, again)) {
				t.Fatal(e)
			}
			if n, e := c.Measure(m); e != nil || n != len(raw) {
				t.Fatal(n, e)
			}
			if bound, e := c.DecodeBound(m.Header, len(raw)); e != nil || bound < uint64(len(raw)) {
				t.Fatal(bound, e)
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("borrowed frame")
			}
			if name == "v25-srr-two-r" {
				return
			}
			if e = c.ValidateAnswer(req, m, routingSources()); e != nil {
				t.Fatal(e)
			}
			if r := c.ValidateAnswerContext(req, m, routingSources(), validation.Context{}); !r.Complete {
				t.Fatal(r)
			}
			for _, flag := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(2)), avp.Some(uint32(4))} {
				bad := decoded(t, c, "v25-srr-two-r")
				b := bad.Body.(*s6c.SRR)
				b.SRRFlags = flag
				// Advertising SMSF support must not substitute for GPRS-Indicator.
				b.SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}
				if e = c.ValidateAnswer(bad, m, routingSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
					t.Fatal("missing GPRS", e)
				}
			}
			m.Header.EndToEnd++
			if e = c.ValidateAnswer(req, m, routingSources()); !errors.Is(e, validation.ErrUnmatchedAnswer) {
				t.Fatal(e)
			}
		})
	}
}
func routingNode(n int) s6c.ServingNode {
	switch n {
	case 0:
		return s6c.ServingNode{SGSNNumber: avp.Some([]byte{0x21, 0xf3})}
	case 1:
		return s6c.ServingNode{SGSNNumber: avp.Some([]byte{0x21, 0xf3}), SGSNName: avp.Some("sgsn.example"), SGSNRealm: avp.Some("example.net")}
	case 2:
		return s6c.ServingNode{MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example.net"), MMENumber: avp.Some([]byte{0x21, 0xf4})}
	case 3:
		return s6c.ServingNode{MSCNumber: avp.Some([]byte{0x21, 0xf5})}
	default:
		return s6c.ServingNode{MSCNumber: avp.Some([]byte{0x21, 0xf5}), MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example.net")}
	}
}
func extraNode(n s6c.ServingNode) s6c.AdditionalServingNode {
	return s6c.AdditionalServingNode{SGSNName: n.SGSNName, SGSNRealm: n.SGSNRealm, SGSNNumber: n.SGSNNumber, MMEName: n.MMEName, MMERealm: n.MMERealm, MMENumber: n.MMENumber, MSCNumber: n.MSCNumber, Extra: n.Extra}
}
func TestS6cRoutingShapesCrossGroupsAndSuccessData(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "v25-srr-two-r")
	for i := 0; i < 5; i++ {
		for j := -1; j < 5; j++ {
			a := decoded(t, c, "v-sra-01").Body.(*s6c.SRA)
			a.SessionID = ""
			a.ServingNode = avp.Some(routingNode(i))
			a.AdditionalServingNode = avp.Optional[s6c.AdditionalServingNode]{}
			if j >= 0 {
				a.AdditionalServingNode = avp.Some(extraNode(routingNode(j)))
			}
			// Clear original envelope; completion owns correlation and selected local source.
			a.OriginHost = ""
			a.OriginRealm = ""
			_, e := c.CompleteAnswer(req, a, routingSources().Terminals[0])
			valid := j < 0 || (i < 2) != (j < 2)
			if (e == nil) != valid {
				t.Fatal(i, j, e)
			}
		}
	}
	for _, mutate := range []func(*s6c.SRA){
		func(m *s6c.SRA) { m.ServingNode.Value = s6c.ServingNode{} },
		func(m *s6c.SRA) { m.ServingNode.Value.MMERealm = avp.Optional[string]{} },
		func(m *s6c.SRA) { m.ServingNode.Value.SGSNNumber = avp.Some([]byte{0x21}) },
		func(m *s6c.SRA) {
			m.AdditionalServingNode = avp.Some(s6c.AdditionalServingNode{SGSNName: avp.Some("sgsn.example")})
		},
		func(m *s6c.SRA) {
			m.AdditionalServingNode = avp.Some(extraNode(routingNode(0)))
			m.ServingNode.Present = false
		},
		func(m *s6c.SRA) {
			m.ServingNode.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3100}, Flags: 128, Data: []byte{0x21}}}
		},
		func(m *s6c.SRA) {
			m.AdditionalServingNode = avp.Some(extraNode(routingNode(0)))
			m.AdditionalServingNode.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3100}, Flags: 128, Data: []byte{0x21}}}
		},
		func(m *s6c.SRA) { m.LMSI = avp.Some([]byte{1, 2, 3}) },
		func(m *s6c.SRA) { m.UserName.Present = false },
		func(m *s6c.SRA) { m.ResultCode = avp.Some(uint32(5012)) },
	} {
		m := decoded(t, c, "v-sra-01")
		mutate(m.Body.(*s6c.SRA))
		if e := c.Validate(m); e == nil {
			t.Fatal("invalid selected node/branch")
		}
	}
	// No invented LMSI or Alert-MSISDN when application facts are absent.
	m := decoded(t, c, "v25-sra-msc")
	b := m.Body.(*s6c.SRA)
	if b.LMSI.Present || b.UserIdentifier.Present {
		t.Fatal("synthesized data")
	}
	for _, code := range []uint32{5001, 5012} {
		m := decoded(t, c, "v-sra-02")
		b := m.Body.(*s6c.SRA)
		b.ResultCode = avp.Some(code)
		b.ExperimentalResult.Present = false
		b.MWDStatus = avp.Some(uint32(0))
		if e := c.Validate(m); e == nil {
			t.Fatal("failure success adjunct")
		}
	}
}
func TestS6cSMEAControlsAndRequestFlags(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v25-srr-two-r", "c27-smea-even", "c27-smea-max15", "c27-smea-min1", "c27-smea-bit0-clear"} {
		m := decoded(t, c, name)
		b := m.Body.(*s6c.SRR)
		number, e := s6c.DecodeSMEA(b.SMRPSMEA.Value)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := s6c.EncodeSMEA(number)
		if e != nil || !bytes.Equal(raw, b.SMRPSMEA.Value) {
			t.Fatal(e)
		}
		first, e := c.Canonical(fixture(t, name), 8192)
		if e != nil {
			t.Fatal(e)
		}
		second, e := c.Canonical(encoded(t, c, m), 8192)
		if e != nil || !bytes.Equal(first, second) {
			t.Fatal(e)
		}
	}
	for _, name := range []string{"n25-srr-nosmea", "n25-smea-length", "n25-smea-toa", "n27-smea-pad", "n27-smea-digit", "n27-smea-stripped", "n27-smea-tlv", "c27-smea-16digits"} {
		if _, e := c.Decode(fixture(t, name)); e == nil {
			t.Fatal(name)
		}
	}
	if _, e := c.Decode(fixture(t, "c27-smea-national")); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	for f := uint32(0); f < 8; f++ {
		for _, mti := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(1))} {
			m := decoded(t, c, "v25-srr-two-r")
			b := m.Body.(*s6c.SRR)
			b.SRRFlags = avp.Some(f)
			b.SMRPMTI = mti
			if e := c.Validate(m); e != nil {
				t.Fatal(f, mti, e)
			}
		}
	}
	// Unknown flag bits are received/retained, but cannot be emitted by Validate.
	raw := fixture(t, "v25-srr-two-r")
	for i := 20; i < len(raw); {
		n := int(raw[i+5])<<16 | int(raw[i+6])<<8 | int(raw[i+7])
		if bytes.Equal(raw[i:i+4], []byte{0, 0, 12, 238}) {
			raw[i+12] = 128
		}
		i += (n + 3) &^ 3
	}
	m, e := c.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	if m.Body.(*s6c.SRR).SRRFlags.Value != 0x80000001 {
		t.Fatal("lost flags")
	}
	if e = c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	answer := decoded(t, c, "v25-sra-msc")
	if e = c.ValidateAnswer(m, answer, routingSources()); e != nil {
		t.Fatal("original receive rules", e)
	}
}

func TestS6cRoutingWireParentsAndPresence(t *testing.T) {
	c := newCodec(t, codec.Config{})
	dup := func(v []byte) []byte { return append(v, v...) }
	remove := func([]byte) []byte { return nil }
	for _, tc := range []struct {
		file string
		path []uint32
		edit func([]byte) []byte
	}{
		{"v25-sra-sgsn-msc", []uint32{2401}, dup}, {"v25-sra-sgsn-msc", []uint32{2406}, dup},
		{"v25-sra-sgsn-msc", []uint32{2401, 1489}, dup}, {"v25-sra-sgsn-msc", []uint32{2406, 2403}, dup},
		{"v25-sra-sgsn-msc", []uint32{2401, 1489}, remove}, {"v25-sra-sgsn-msc", []uint32{2406, 2403}, remove},
		{"v25-sra-sgsn-3", []uint32{2406, 2410}, remove}, {"v25-sra-sgsn-3", []uint32{2406, 2409}, dup},
		{"v25-sra-sgsn-msc", []uint32{2400}, func(v []byte) []byte { v[7] = 15; return v }},
		{"v25-sra-sgsn-msc", []uint32{3312}, func(v []byte) []byte { v[7] = 15; return v }},
		{"v25-sra-sgsn-msc", []uint32{3102, 701}, remove},
		{"v25-srr-two-r", []uint32{3310}, dup}, {"v25-srr-two-r", []uint32{3309}, dup},
	} {
		if _, e := c.Decode(rewriteAVP(t, fixture(t, tc.file), tc.path, tc.edit)); e == nil {
			t.Fatal(tc.file, tc.path)
		}
	}
	// Future MWD bits survive receiving without becoming legal local send values.
	raw := rewriteAVP(t, fixture(t, "v25-sra-sgsn-msc"), []uint32{3312}, func(v []byte) []byte { v[12] = 128; return v })
	m, e := c.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	if m.Body.(*s6c.SRA).MWDStatus.Value != 0x80000000 {
		t.Fatal("MWD bit lost")
	}
	if e = c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	if e = c.ValidateAnswer(decoded(t, c, "v25-srr-two-r"), m, routingSources()); e != nil {
		t.Fatal(e)
	}
	// Absent and explicit zero flags are distinct duplicate contents.
	req := decoded(t, c, "c27-smea-bit0-clear")
	req.Body.(*s6c.SRR).SRRFlags = avp.Some(uint32(0))
	present, e := c.Canonical(encoded(t, c, req), 8192)
	if e != nil {
		t.Fatal(e)
	}
	req.Body.(*s6c.SRR).SRRFlags.Present = false
	absent, e := c.Canonical(encoded(t, c, req), 8192)
	if e != nil || bytes.Equal(present, absent) {
		t.Fatal("presence collapsed", e)
	}
}
