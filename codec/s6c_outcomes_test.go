package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func TestS6cMultiOutcomeFrozenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	pairs := [][2]string{{"v24-s6c-rd-r", "v24-s6c-rd-a"}, {"c24-rdr-nodiag-r", "v24-s6c-rd-a"}, {"v24-s6c-rd-r", "f24-s6c-rd-5001"}, {"v24-s6c-rd-r", "f24-s6c-rd-5558"}, {"b22-rdr-sgsn-absent", "b22-rda-sgsn-ok"}, {"b22-rdr-msc-sgsn", "b22-rda-pair-5558"}, {"b22-rdr-mme-sgsn", "b22-rda-pair-5001"}}
	for _, pair := range pairs {
		t.Run(pair[0]+"/"+pair[1], func(t *testing.T) {
			req, ans := decoded(t, c, pair[0]), decoded(t, c, pair[1])
			if r := c.ValidateAnswerContext(req, ans, routingSources(), validation.Context{}); !r.Complete {
				t.Fatal(r)
			}
			for _, m := range []pdu.Message{req, ans} {
				raw := encoded(t, c, m)
				out, e := c.Decode(raw)
				if e != nil || !bytes.Equal(raw, encoded(t, c, out)) {
					t.Fatal(e)
				}
				if size, e := c.Measure(m); e != nil || size != len(raw) {
					t.Fatal(size, e)
				}
				if bound, e := c.DecodeBound(m.Header, len(raw)); e != nil || bound < uint64(len(raw)) {
					t.Fatal(bound, e)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), encoded(t, c, out)) {
					t.Fatal("borrowed frame")
				}
			}
			a, e := c.Canonical(fixture(t, pair[0]), 8192)
			if e != nil {
				t.Fatal(e)
			}
			b, e := c.Canonical(encoded(t, c, req), 8192)
			if e != nil || !bytes.Equal(a, b) {
				t.Fatal(e)
			}
			ans.Header.EndToEnd++
			if e = c.ValidateAnswer(req, ans, routingSources()); !errors.Is(e, validation.ErrUnmatchedAnswer) {
				t.Fatal(e)
			}
		})
	}
	for _, name := range []string{"n24-rdr-diag256-r", "n24-rdr-diag3-r", "n24-rdr-dupdiag-r", "n24-rdr-successdiag-r", "n24-rdr-mscmme-r", "n24-rda-emptyid"} {
		if _, e := c.Decode(fixture(t, name)); e == nil {
			t.Fatal(name)
		}
	}
}
func multiOutcome(mask int, cause uint32, diag avp.Optional[uint32]) s6c.SMDeliveryOutcome {
	var o s6c.SMDeliveryOutcome
	if mask&1 != 0 {
		o.MMESMDeliveryOutcome = avp.Some(s6c.MMESMDeliveryOutcome{SMDeliveryCause: avp.Some(cause), AbsentUserDiagnosticSM: diag})
	}
	if mask&2 != 0 {
		o.MSCSMDeliveryOutcome = avp.Some(s6c.MSCSMDeliveryOutcome{SMDeliveryCause: avp.Some(cause), AbsentUserDiagnosticSM: diag})
	}
	if mask&4 != 0 {
		o.SGSNSMDeliveryOutcome = avp.Some(s6c.SGSNSMDeliveryOutcome{SMDeliveryCause: avp.Some(cause), AbsentUserDiagnosticSM: diag})
	}
	return o
}
func TestS6cMultiOutcomeNodeMatrix(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for mask := 0; mask < 8; mask++ {
		for cause := uint32(0); cause <= 3; cause++ {
			m := decoded(t, c, "v24-s6c-rd-r")
			m.Body.(*s6c.RDR).SMDeliveryOutcome = multiOutcome(mask, cause, avp.Optional[uint32]{})
			e := c.Validate(m)
			valid := mask != 0 && mask&3 != 3 && cause <= 2
			if (e == nil) != valid {
				t.Fatal(mask, cause, e)
			}
			if valid {
				if _, e = c.Decode(encoded(t, c, m)); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	// Every assigned/reserved/octet-overflow value at each parent, preserving zero.
	for _, mask := range []int{1, 2, 4} {
		for d := uint32(0); d <= 256; d++ {
			m := decoded(t, c, "v24-s6c-rd-r")
			m.Body.(*s6c.RDR).SMDeliveryOutcome = multiOutcome(mask, 1, avp.Some(d))
			allowed := map[int]map[uint32]bool{1: {13: true}, 2: {0: true, 1: true, 2: true, 3: true, 4: true, 9: true, 13: true}, 4: {2: true, 5: true, 6: true, 7: true, 8: true, 10: true, 13: true}}
			e := c.Validate(m)
			if (e == nil) != allowed[mask][d] {
				t.Fatal(mask, d, e)
			}
			if e != nil {
				want := validation.ErrUnsupported
				if d == 256 {
					want = validation.ErrField
				}
				if !errors.Is(e, want) {
					t.Fatal(mask, d, e)
				}
			}
		}
	}
	for _, mask := range []int{1, 2, 4} {
		for _, cause := range []uint32{0, 2} {
			m := decoded(t, c, "v24-s6c-rd-r")
			d := uint32(13)
			if mask == 2 {
				d = 0
			}
			if mask == 4 {
				d = 5
			}
			m.Body.(*s6c.RDR).SMDeliveryOutcome = multiOutcome(mask, cause, avp.Some(d))
			if e := c.Validate(m); !errors.Is(e, validation.ErrField) {
				t.Fatal(mask, cause, e)
			}
		}
	}
}
func TestS6cMultiOutcomeWireParentsAndCanonical(t *testing.T) {
	c := newCodec(t, codec.Config{})
	dup := func(v []byte) []byte { return append(v, v...) }
	remove := func([]byte) []byte { return nil }
	for _, parent := range []uint32{3318, 3319} {
		for _, path := range [][]uint32{{3316, parent}, {3316, parent, 3321}, {3316, parent, 3322}} {
			if _, e := c.Decode(rewriteAVP(t, fixture(t, "b22-rdr-msc-sgsn"), path, dup)); e == nil {
				t.Fatal(path)
			}
		}
		raw := rewriteAVP(t, fixture(t, "b22-rdr-msc-sgsn"), []uint32{3316, parent, 3321}, remove)
		if _, e := c.Decode(raw); !errors.Is(e, validation.ErrMissing) {
			t.Fatal(e)
		}
		raw = rewriteAVP(t, fixture(t, "b22-rdr-msc-sgsn"), []uint32{3316, parent, 3322}, func(v []byte) []byte { v[7] = 15; return v })
		if _, e := c.Decode(raw); e == nil {
			t.Fatal("3-octet diagnostic", parent)
		}
	}
	req := decoded(t, c, "b22-rdr-msc-sgsn")
	baseline, e := c.Canonical(encoded(t, c, req), 8192)
	if e != nil {
		t.Fatal(e)
	}
	req.Body.(*s6c.RDR).SMDeliveryOutcome.MSCSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Present = false
	absent, e := c.Canonical(encoded(t, c, req), 8192)
	if e != nil || bytes.Equal(baseline, absent) {
		t.Fatal("zero collapsed", e)
	}
	// Changing only one parent's diagnostic must not be merged into the other.
	req = decoded(t, c, "b22-rdr-msc-sgsn")
	req.Body.(*s6c.RDR).SMDeliveryOutcome.SGSNSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value = 6
	changed, e := c.Canonical(encoded(t, c, req), 8192)
	if e != nil || bytes.Equal(baseline, changed) {
		t.Fatal("parent content collapsed", e)
	}
	for _, code := range []uint32{3336, 3337} {
		req = decoded(t, c, "b22-rdr-msc-sgsn")
		req.Body.(*s6c.RDR).SMDeliveryOutcome.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: code}, Flags: 128}}
		if e = c.Validate(req); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(code, e)
		}
	}
}
func TestS6cMultiOutcomeResultBranches(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v24-s6c-rd-r", "b22-rdr-sgsn-absent", "b22-rdr-msc-sgsn", "b22-rdr-mme-sgsn"} {
		req := decoded(t, c, name)
		for _, body := range []pdu.Body{
			&s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))},
			&s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserIdentifier: avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf9}})},
			&s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})},
			&s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5558})},
			&s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 3004},
		} {
			ans, e := c.CompleteAnswer(req, body, routingSources().Terminals[0])
			if e != nil {
				t.Fatal(e)
			}
			if e = c.ValidateAnswer(req, ans, routingSources()); e != nil {
				t.Fatal(e)
			}
			if e = c.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "other.example", Realm: "example.net"}}}); e == nil {
				t.Fatal("source bypass")
			}
		}
	}
	for _, mutate := range []func(*s6c.RDA){
		func(b *s6c.RDA) { b.UserIdentifier = avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21}}) },
		func(b *s6c.RDA) { b.ResultCode = avp.Some(uint32(2001)) },
		func(b *s6c.RDA) { b.ExperimentalResult.Value.ExperimentalResultCode = 5550 },
		func(b *s6c.RDA) { b.ExperimentalResult.Value.VendorID = 0 },
	} {
		m := decoded(t, c, "f24-s6c-rd-5558")
		mutate(m.Body.(*s6c.RDA))
		if e := c.Validate(m); e == nil {
			t.Fatal("invalid result")
		}
	}
}
