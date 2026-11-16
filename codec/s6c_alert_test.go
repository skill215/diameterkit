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
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestS6cReportAlertGoldenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v-rdr-01", "v-rda-01", "v-rda-02", "v-rda-03", "v-alr-01", "v-ala-01", "v-ala-02"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, name)
			m, err := c.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded(t, c, m), raw) {
				t.Fatal("pre-implementation fixture changed")
			}
			clear(raw)
			if !bytes.Equal(encoded(t, c, m), fixture(t, name)) {
				t.Fatal("decode borrowed input")
			}
			size, err := c.Measure(m)
			if err != nil || size != len(fixture(t, name)) {
				t.Fatal(size, err)
			}
			if bound, err := c.DecodeBound(m.Header, size); err != nil || bound < uint64(size) {
				t.Fatal(bound, err)
			}
			if m.Header.Flags&pdu.RequestFlag != 0 {
				return
			}
			request := "v-rdr-01"
			host := "hss.example"
			if m.Header.CommandCode == 8388648 {
				request = "v-alr-01"
				host = "iwmsc.example"
			}
			src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: host, Realm: "example.net"}}}
			req := decoded(t, c, request)
			if err = c.ValidateAnswer(req, m, src); err != nil {
				t.Fatal(err)
			}
			if r := c.ValidateAnswerContext(req, m, src, validation.Context{}); !r.Complete {
				t.Fatal(r)
			}
		})
	}
}

func TestS6cOutcomeRulesAndPaths(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, tc := range []struct {
		name   string
		mutate func(*s6c.RDR)
		want   error
		path   []uint32
	}{
		{"empty outcome", func(r *s6c.RDR) { r.SMDeliveryOutcome = s6c.SMDeliveryOutcome{} }, validation.ErrMissing, []uint32{3316}},
		{"no cause", func(r *s6c.RDR) {
			r.SMDeliveryOutcome.MMESMDeliveryOutcome.Value.SMDeliveryCause = avp.Optional[uint32]{}
		}, validation.ErrMissing, []uint32{3316, 3317, 3321}},
		{"cause reserved", func(r *s6c.RDR) { r.SMDeliveryOutcome.MMESMDeliveryOutcome.Value.SMDeliveryCause.Value = 3 }, validation.ErrUnsupported, []uint32{3316, 3317, 3321}},
		{"diagnostic without absence", func(r *s6c.RDR) {
			r.SMDeliveryOutcome.MMESMDeliveryOutcome.Value.AbsentUserDiagnosticSM = avp.Some(uint32(13))
		}, validation.ErrField, []uint32{3316, 3317, 3322}},
		{"diagnostic over 255", func(r *s6c.RDR) {
			m := &r.SMDeliveryOutcome.MMESMDeliveryOutcome.Value
			m.SMDeliveryCause.Value = 1
			m.AbsentUserDiagnosticSM = avp.Some(uint32(256))
		}, validation.ErrField, []uint32{3316, 3317, 3322}},
		{"MME diagnostic zero is not absent", func(r *s6c.RDR) {
			m := &r.SMDeliveryOutcome.MMESMDeliveryOutcome.Value
			m.SMDeliveryCause.Value = 1
			m.AbsentUserDiagnosticSM = avp.Some(uint32(0))
		}, validation.ErrUnsupported, []uint32{3316, 3317, 3322}},
		{"typed IP-SM-GW cannot hide in Extra", func(r *s6c.RDR) {
			r.SMDeliveryOutcome.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3320}, Flags: 128}}
		}, validation.ErrDuplicate, []uint32{3316, 3320}},
		{"typed SMSMI cannot hide in Extra", func(r *s6c.RDR) { r.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3324}, Flags: 128}} }, validation.ErrDuplicate, []uint32{3324}},
		{"typed identity cannot hide in Extra", func(r *s6c.RDR) { r.UserIdentifier.Extra = []avp.Raw{{Key: avp.Key{Code: 1}, Data: []byte("00101")}} }, validation.ErrDuplicate, []uint32{3102, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decoded(t, c, "v-rdr-01")
			tc.mutate(m.Body.(*s6c.RDR))
			err := c.Validate(m)
			if !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
			var f *validation.FieldError
			if !errors.As(err, &f) || len(f.Path) != len(tc.path) {
				t.Fatal(err)
			}
			for i, k := range f.Path {
				if k.Code != tc.path[i] {
					t.Fatal(f.Path)
				}
			}
		})
	}
	for _, cause := range []uint32{0, 1, 2} {
		m := decoded(t, c, "v-rdr-01")
		n := &m.Body.(*s6c.RDR).SMDeliveryOutcome.MMESMDeliveryOutcome.Value
		n.SMDeliveryCause.Value = cause
		if cause == 1 {
			n.AbsentUserDiagnosticSM = avp.Some(uint32(13))
		}
		if _, err := c.Decode(encoded(t, c, m)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestS6cResultIsolation(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, tc := range []struct {
		name string
		body pdu.Body
		want error
	}{
		{"RDA missing result", &s6c.RDA{AuthSessionState: 1}, validation.ErrField},
		{"RDA dual result", &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, validation.ErrField},
		{"RDA bad vendor", &s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 999, ExperimentalResultCode: 5558})}, validation.ErrUnsupported},
		{"RDA failed identity", &s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5558}), UserIdentifier: avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21}})}, validation.ErrField},
		{"RDA base5001 needs diagnostic", &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5001))}, validation.ErrMissing},
		{"ALA experimental5001", &s6c.ALA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, validation.ErrUnsupported},
		{"ALA missing result", &s6c.ALA{AuthSessionState: 1}, validation.ErrField},
		{"ALA normal3004", &s6c.ALA{AuthSessionState: 1, ResultCode: avp.Some(uint32(3004))}, validation.ErrUnsupported},
		{"SRA cannot inherit RDA5558", &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5558})}, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.ValidateBody(tc.body); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	for _, kind := range []string{"rda", "ala"} {
		req := decoded(t, c, "v-rdr-01")
		var body pdu.Body = &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}
		if kind == "ala" {
			req = decoded(t, c, "v-alr-01")
			body = &s6c.ALA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}
		}
		if _, err := c.CompleteAnswer(req, body, pdu.Origin{Host: "hss.example", Realm: "example.net"}); err != nil {
			t.Fatal(err)
		}
	}
}

// Independent bounded TLV rewriting keeps the rest of each original document
// frame intact, including all ancestors' lengths. It never calls the generator.
func rewriteAVP(t *testing.T, frame []byte, path []uint32, edit func([]byte) []byte) []byte {
	t.Helper()
	var rewrite func([]byte, []uint32) []byte
	rewrite = func(b []byte, keys []uint32) []byte {
		for off := 0; off < len(b); {
			if len(b)-off < 8 {
				t.Fatal("bad test frame")
			}
			n := int(b[off+5])<<16 | int(b[off+6])<<8 | int(b[off+7])
			padded := (n + 3) &^ 3
			if n < 8 || off+padded > len(b) {
				t.Fatal("bad test AVP")
			}
			if binary.BigEndian.Uint32(b[off:]) == keys[0] {
				v := append([]byte(nil), b[off:off+padded]...)
				if len(keys) == 1 {
					v = edit(v)
				} else {
					header := 8
					if v[4]&128 != 0 {
						header = 12
					}
					payload := rewrite(v[header:n], keys[1:])
					v = append(v[:header], payload...)
					wire.Put24(v[5:8], uint32(len(v)))
				}
				out := append([]byte(nil), b[:off]...)
				out = append(out, v...)
				return append(out, b[off+padded:]...)
			}
			off += padded
		}
		t.Fatal("missing test AVP", keys)
		return nil
	}
	out := append([]byte(nil), frame[:20]...)
	out = append(out, rewrite(frame[20:], path)...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}

func TestS6cNestedWireNegatives(t *testing.T) {
	c := newCodec(t, codec.Config{})
	dup := func(v []byte) []byte { return append(v, v...) }
	remove := func([]byte) []byte { return nil }
	short := func(v []byte) []byte { wire.Put24(v[5:8], uint32(len(v)-1)); return v }
	for _, tc := range []struct {
		name, file string
		path       []uint32
		edit       func([]byte) []byte
	}{
		{"RDR identity missing", "v-rdr-01", []uint32{3102}, remove},
		{"RDR number duplicate", "v-rdr-01", []uint32{3102, 701}, dup},
		{"RDR MME duplicate", "v-rdr-01", []uint32{3316, 3317}, dup},
		{"RDR cause missing", "v-rdr-01", []uint32{3316, 3317, 3321}, remove},
		{"RDR cause duplicate", "v-rdr-01", []uint32{3316, 3317, 3321}, dup},
		{"RDR cause width3", "v-rdr-01", []uint32{3316, 3317, 3321}, short},
		{"RDA result duplicate", "v-rda-01", []uint32{268}, dup},
		{"RDA vendor missing", "v-rda-02", []uint32{297, 266}, remove},
		{"RDA alert number missing", "v-rda-01", []uint32{3102, 701}, remove},
		{"ALR destination missing", "v-alr-01", []uint32{293}, remove},
		{"ALR SC missing", "v-alr-01", []uint32{3300}, remove},
		{"ALR time duplicate", "v-alr-01", []uint32{3329}, dup},
		{"ALR time width3", "v-alr-01", []uint32{3329}, short},
		{"ALA no result", "v-ala-01", []uint32{268}, remove},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := c.Decode(rewriteAVP(t, fixture(t, tc.file), tc.path, tc.edit)); err == nil {
				t.Fatal("accepted invalid frame")
			}
		})
	}
}

func TestS6cTimePresenceAndAlertBoundary(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, v := range []avp.Time{0, 1, 0xfffffff0, 0xffffffff} {
		m := decoded(t, c, "v-alr-01")
		m.Body.(*s6c.ALR).MaximumUEAvailabilityTime = avp.Some(v)
		out, err := c.Decode(encoded(t, c, m))
		if err != nil {
			t.Fatal(err)
		}
		got := out.Body.(*s6c.ALR).MaximumUEAvailabilityTime
		if !got.Present || got.Value != v {
			t.Fatal(got)
		}
	}
	m := decoded(t, c, "v-alr-01")
	m.Body.(*s6c.ALR).MaximumUEAvailabilityTime = avp.Optional[avp.Time]{}
	if out, err := c.Decode(encoded(t, c, m)); err != nil || out.Body.(*s6c.ALR).MaximumUEAvailabilityTime.Present {
		t.Fatal(out, err)
	}
	for _, code := range []uint32{3324, 3333, 2401} {
		m := decoded(t, c, "v-alr-01")
		m.Body.(*s6c.ALR).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: code}, Flags: 128}}
		if err := c.Validate(m); err == nil {
			t.Fatal(code, err)
		}
	}
}

func TestS6cFlagsReceiveEvidenceAndCanonical(t *testing.T) {
	c := newCodec(t, codec.Config{})
	m := decoded(t, c, "v-rdr-01")
	m.Body.(*s6c.RDR).RDRFlags = avp.Some(uint32(1))
	raw := encoded(t, c, m)
	unknown := rewriteAVP(t, raw, []uint32{3323}, func(v []byte) []byte { binary.BigEndian.PutUint32(v[12:], 0x80000001); return v })
	received, err := c.Decode(unknown)
	if err != nil {
		t.Fatal(err)
	}
	if received.Body.(*s6c.RDR).RDRFlags.Value != 0x80000001 {
		t.Fatal("unknown flag discarded")
	}
	if err = c.Validate(received); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("send must reject reserved bits", err)
	}
	if r := c.ValidateContext(received, validation.Context{}); !r.Complete {
		t.Fatal(r)
	}
	answer, err := c.CompleteAnswer(received, &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}, sources().Terminals[0])
	if err != nil {
		t.Fatal(err)
	}
	if err = c.ValidateAnswer(received, answer, sources()); err != nil {
		t.Fatal("receive-only evidence rejected", err)
	}
	if r := c.ValidateAnswerContext(received, answer, sources(), validation.Context{}); !r.Complete {
		t.Fatal(r)
	}
	if _, err = c.DecodeAnswer(received, encoded(t, c, answer), sources()); err != nil {
		t.Fatal(err)
	}
	a, err := c.Canonical(unknown, 4096)
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Canonical(raw, 4096)
	if err != nil || bytes.Equal(a, b) {
		t.Fatal("reserved flag collapsed", err)
	}
	for _, name := range []string{"v-rdr-01", "v-alr-01"} {
		original := fixture(t, name)
		before, err := c.Canonical(original, 4096)
		if err != nil {
			t.Fatal(err)
		}
		original[4] ^= pdu.RetransmittedFlag
		original[12] ^= 1
		after, err := c.Canonical(original, 4096)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal(err)
		}
	}
	profile, _ := s6c.Profile()
	bundle, _ := schema.Compile(profile)
	for _, body := range []pdu.Body{&s6c.RDR{}, &s6c.ALR{}} {
		op, _ := bundle.Lookup(body.CommandKey())
		if op.TransportReplay || (!op.NoContext && op.CheckContext == nil) || op.Canonical == nil {
			t.Fatal(op.Key)
		}
	}
}
