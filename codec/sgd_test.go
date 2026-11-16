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
	"reflect"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func sgdCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	gp, e := sgd.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(bp, sp, gp)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestSGDGoldenPairs(t *testing.T) {
	c := sgdCodec(t)
	for _, pair := range [][2]string{{"v-ofr-01", "v-ofa-01"}, {"v-ofr-01", "v-ofa-02"}, {"v-tfr-01", "v-tfa-01"}, {"v-tfr-01", "v-tfa-02"}, {"v-gdd-tfr-01", "v-gdd-tfa-01"}, {"v-gdd-tfr-01", "v-gdd-tfa-02"}, {"v24-sgd-mo-r", "v24-sgd-mo-a"}, {"v24-gdd-mo-r", "v24-gdd-mo-a"}} {
		t.Run(pair[1], func(t *testing.T) {
			req := decoded(t, c, pair[0])
			ans := decoded(t, c, pair[1])
			var host, realm string
			switch a := ans.Body.(type) {
			case *sgd.OFA:
				host, realm = a.OriginHost, a.OriginRealm
			case *sgd.TFA:
				host, realm = a.OriginHost, a.OriginRealm
			}
			src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: host, Realm: realm}}}
			if e := c.ValidateAnswer(req, ans, src); e != nil {
				t.Fatal(e)
			}
			if r := c.ValidateAnswerContext(req, ans, src, validation.Context{}); !r.Complete {
				t.Fatal(r)
			}
			for _, name := range pair {
				raw := fixture(t, name)
				m, e := c.Decode(raw)
				if e != nil {
					t.Fatal(e)
				}
				out := encoded(t, c, m)
				// V-OFR-01 predates the leaf vectors and has a different legal root AVP
				// order. Generated output is deterministic, not a lossless wire proxy.
				if name != "v-ofr-01" && !bytes.Equal(raw, out) {
					t.Fatalf("changed fixture %s", name)
				}
				again, e := c.Decode(out)
				if e != nil || !reflect.DeepEqual(m, again) {
					t.Fatal("semantic re-encode", e)
				}
				if name == "v-ofr-01" {
					a, e := c.Canonical(raw, 4096)
					if e != nil {
						t.Fatal(e)
					}
					b, e := c.Canonical(out, 4096)
					if e != nil || !bytes.Equal(a, b) {
						t.Fatal("canonical reorder", e)
					}
				}
				clear(raw)
				if !bytes.Equal(out, encoded(t, c, m)) {
					t.Fatal("borrowed input")
				}
				n, e := c.Measure(m)
				if e != nil || n != len(out) {
					t.Fatal(n, e)
				}
				if b, e := c.DecodeBound(m.Header, n); e != nil || b < uint64(n) {
					t.Fatal(b, e)
				}
			}
		})
	}
}
func TestSGDRequestAndLocationRules(t *testing.T) {
	c := sgdCodec(t)
	for _, tc := range []struct {
		name, file string
		mutate     func(pdu.Body)
		want       error
	}{
		{"IMSI nondigit", "v-ofr-01", func(b pdu.Body) { b.(*sgd.OFR).UserIdentifier.UserName = avp.Some("0010x") }, validation.ErrField},
		{"IMSI too short", "v-tfr-01", func(b pdu.Body) { b.(*sgd.TFR).UserName = "1234" }, validation.ErrField},
		{"two MT targets", "v-tfr-01", func(b pdu.Body) { b.(*sgd.TFR).SGSNNumber = avp.Some([]byte{0x21}) }, validation.ErrField},
		{"missing host", "v-tfr-01", func(b pdu.Body) { b.(*sgd.TFR).DestinationHost = "" }, avp.ErrValue},
		{"TPDU too long", "v-ofr-01", func(b pdu.Body) { b.(*sgd.OFR).SMRPUI = make([]byte, 201) }, validation.ErrField},
		{"empty TPDU opaque", "v-ofr-01", func(b pdu.Body) { b.(*sgd.OFR).SMRPUI = []byte{} }, nil},
		{"max opaque TPDU", "v-tfr-01", func(b pdu.Body) { b.(*sgd.TFR).SMRPUI = bytes.Repeat([]byte{255}, 200) }, nil},
		{"Gdd MME location", "v-ofr-01", func(b pdu.Body) { b.(*sgd.OFR).OFRFlags = avp.Some(uint32(1)) }, validation.ErrUnsupported},
		{"reserved flags send", "v-tfr-01", func(b pdu.Body) { b.(*sgd.TFR).TFRFlags = avp.Some(uint32(2)) }, validation.ErrUnsupported},
		{"typed retransmission cannot hide in Extra", "v-tfr-01", func(b pdu.Body) {
			b.(*sgd.TFR).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3330}, Flags: 128, Data: []byte{0, 0, 0, 1}}}
		}, validation.ErrDuplicate},
		{"T4 identity M0", "v-ofr-01", func(b pdu.Body) {
			b.(*sgd.OFR).UserIdentifier.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3111}, Flags: 128, Data: []byte("x")}}
		}, validation.ErrUnsupported},
		{"ECGI spare bits", "v-ofr-01", func(b pdu.Body) {
			b.(*sgd.OFR).EPSLocationInformation.Value.MMELocationInformation.EUTRANCellGlobalIdentity.Value[3] = 0x10
		}, validation.ErrField},
		{"PLMN filler wrong digit", "v-ofr-01", func(b pdu.Body) {
			b.(*sgd.OFR).EPSLocationInformation.Value.MMELocationInformation.TrackingAreaIdentity.Value[0] = 0xf0
		}, validation.ErrField},
		{"deleted TAC unsupported", "v-ofr-01", func(b pdu.Body) {
			x := b.(*sgd.OFR).EPSLocationInformation.Value.MMELocationInformation.TrackingAreaIdentity.Value
			x[3] = 255
			x[4] = 254
		}, validation.ErrUnsupported},
		{"SGSN location M0", "v-ofr-01", func(b pdu.Body) {
			b.(*sgd.OFR).EPSLocationInformation.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 1601}, Flags: 128}}
		}, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decoded(t, c, tc.file)
			tc.mutate(m.Body)
			e := c.Validate(m)
			if !errors.Is(e, tc.want) {
				t.Fatal(e)
			}
		})
	}
}
func TestSGDResultsAndCauseDirections(t *testing.T) {
	c := sgdCodec(t)
	for _, mt := range []bool{false, true} {
		name := "v-ofr-01"
		if mt {
			name = "v-tfr-01"
		}
		req := decoded(t, c, name)
		for _, code := range []uint32{0, 5001, 5550, 5551, 5552, 5553, 5554, 5555, 5556, 5557, 5558, 5011} {
			for cause := uint32(0); cause <= 7; cause++ {
				ex := avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
				fc := avp.Optional[sgd.SMDeliveryFailureCause]{}
				if code == 5555 {
					fc = avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: cause, SMDiagnosticInfo: avp.Some([]byte{0, 255, 0})})
				}
				var body pdu.Body = &sgd.OFA{AuthSessionState: 1, ExperimentalResult: ex, SMDeliveryFailureCause: fc}
				supported := code == 5001 || code == 5552 || code == 5555 && cause >= 3 && cause <= 6
				if mt {
					body = &sgd.TFA{AuthSessionState: 1, ExperimentalResult: ex, SMDeliveryFailureCause: fc}
					supported = code == 5001 || code >= 5550 && code <= 5554 || code == 5555 && cause <= 2 || code == 5557
				}
				_, e := c.CompleteAnswer(req, body, pdu.Origin{Host: "node.example", Realm: "example.net"})
				if (e == nil) != supported {
					t.Fatalf("MT=%t code=%d cause=%d: %v", mt, code, cause, e)
				}
			}
		}
	}
	req := decoded(t, c, "v-tfr-01")
	for _, n := range []uint32{0, 13, 255, 256} {
		a := &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5550}), AbsentUserDiagnosticSM: avp.Some(n)}
		_, e := c.CompleteAnswer(req, a, pdu.Origin{Host: "mme.example", Realm: "example.net"})
		if (e == nil) != (n == 13) {
			t.Fatal(n, e)
		}
	}
	for _, body := range []pdu.Body{&sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555})}, &sgd.TFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), SMDeliveryFailureCause: avp.Some(sgd.SMDeliveryFailureCause{})}} {
		key := body.CommandKey()
		r := req
		if key.Code == 8388645 {
			r = decoded(t, c, "v-ofr-01")
		}
		if _, e := c.CompleteAnswer(r, body, pdu.Origin{Host: "node.example", Realm: "example.net"}); e == nil {
			t.Fatal("missing/contradictory cause")
		}
	}
}
func TestSGDOriginalTargetRelationAndE(t *testing.T) {
	c := sgdCodec(t)
	req := decoded(t, c, "v-gdd-tfr-01")
	origin := pdu.Origin{Host: "sgsn.example", Realm: "example.net"}
	src := pdu.AnswerSources{Terminals: []pdu.Origin{origin}}
	for _, body := range []*sgd.TFA{{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5557})}, {AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), EPSLocationInformation: decoded(t, c, "v-ofr-01").Body.(*sgd.OFR).EPSLocationInformation}} {
		if _, e := c.CompleteAnswer(req, body, origin); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
		mme := decoded(t, c, "v-gdd-tfr-01")
		b := mme.Body.(*sgd.TFR)
		b.MMENumberForMTSMS = b.SGSNNumber
		b.SGSNNumber = avp.Optional[[]byte]{}
		ans, e := c.CompleteAnswer(mme, body, origin)
		if e != nil {
			t.Fatal(e)
		}
		if e = c.ValidateAnswer(req, ans, src); !errors.Is(e, validation.ErrInvalidAnswer) || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
		out, e := c.DecodeAnswer(req, encoded(t, c, ans), src)
		if out.Body != nil || !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(out, e)
		}
		if r := c.ValidateAnswerContext(req, ans, src, validation.Context{}); r.Complete || r.State != validation.Rejected {
			t.Fatal(r)
		}
	}
	ans, e := c.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, origin)
	if e != nil {
		t.Fatal(e)
	}
	if e = c.ValidateAnswer(req, ans, src); e != nil {
		t.Fatal(e)
	}
}
func TestSGDIndependentWireNegatives(t *testing.T) {
	c := sgdCodec(t)
	for _, tc := range []struct {
		name, file string
		path       []uint32
		edit       func([]byte) []byte
	}{
		{"missing TPDU", "v-ofr-01", []uint32{3301}, func([]byte) []byte { return nil }},
		{"duplicate TPDU", "v-tfr-01", []uint32{3301}, func(b []byte) []byte { return append(b, b...) }},
		{"missing IMSI in identity", "v-ofr-01", []uint32{3102, 1}, func([]byte) []byte { return nil }},
		{"bad TBCD", "v-ofr-01", []uint32{3300}, func(b []byte) []byte { b[12] = 0xfa; return b }},
		{"ECGI wrong width", "v-ofr-01", []uint32{1496, 1600, 1602}, func(b []byte) []byte { wire.Put24(b[5:8], 18); return b }},
		{"bad Time width", "v-tfr-01", []uint32{3307}, func(b []byte) []byte { wire.Put24(b[5:8], 15); return b }},
		{"wrong MO cause", "v-ofa-02", []uint32{3303, 3304}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[12:], 0); return b }},
		{"wrong MT cause", "v-tfa-02", []uint32{3303, 3304}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[12:], 3); return b }},
		{"missing failure cause", "v-tfa-02", []uint32{3303, 3304}, func([]byte) []byte { return nil }},
		{"wrong namespace", "v-tfa-02", []uint32{297, 266}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[8:], 0); return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, e := c.Decode(rewriteAVP(t, fixture(t, tc.file), tc.path, tc.edit)); e == nil {
				t.Fatal("accepted invalid frame")
			}
		})
	}
}
func TestSGDFlagsTimeCanonicalAndExtensionOwnership(t *testing.T) {
	c := sgdCodec(t)
	req := decoded(t, c, "v-tfr-01")
	b := req.Body.(*sgd.TFR)
	for _, value := range []avp.Time{0, 0xffffffff, 1} {
		b.SMDeliveryStartTime = avp.Some(value)
		m, e := c.Decode(encoded(t, c, req))
		if e != nil || m.Body.(*sgd.TFR).SMDeliveryStartTime != b.SMDeliveryStartTime {
			t.Fatal(e)
		}
	}
	b.TFRFlags = avp.Some(uint32(1))
	raw := encoded(t, c, req)
	unknown := rewriteAVP(t, raw, []uint32{3302}, func(v []byte) []byte { binary.BigEndian.PutUint32(v[12:], 0x80000001); return v })
	received, e := c.Decode(unknown)
	if e != nil {
		t.Fatal(e)
	}
	if e = c.Validate(received); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	a, e := c.CompleteAnswer(received, &sgd.TFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}, pdu.Origin{Host: "mme.example", Realm: "example.net"})
	if e != nil {
		t.Fatal(e)
	}
	if e = c.ValidateAnswer(received, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "mme.example", Realm: "example.net"}}}); e != nil {
		t.Fatal(e)
	}
	before, e := c.Canonical(raw, 4096)
	if e != nil {
		t.Fatal(e)
	}
	after, e := c.Canonical(unknown, 4096)
	if e != nil || bytes.Equal(before, after) {
		t.Fatal("flags collapsed", e)
	}
	mutated := rewriteAVP(t, raw, []uint32{3301}, func(v []byte) []byte { v[12] ^= 1; return v })
	after, e = c.Canonical(mutated, 4096)
	if e != nil || bytes.Equal(before, after) {
		t.Fatal("TPDU collapsed", e)
	}
	b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 999, Code: 777}, Flags: 128, Data: []byte{0, 255, 1}}}
	out := encoded(t, c, req)
	m, e := c.Decode(out)
	if e != nil {
		t.Fatal(e)
	}
	clear(out)
	if !bytes.Equal(m.Body.(*sgd.TFR).Extra[0].Data, []byte{0, 255, 1}) {
		t.Fatal("borrowed extension")
	}
	gp, _ := sgd.Profile()
	bundle, _ := schema.Compile(gp)
	for _, body := range []pdu.Body{&sgd.OFR{}, &sgd.TFR{}} {
		op, _ := bundle.Lookup(body.CommandKey())
		if op.TransportReplay || (!op.NoContext && op.CheckContext == nil) || op.Canonical == nil {
			t.Fatal(op.Key)
		}
	}
}
