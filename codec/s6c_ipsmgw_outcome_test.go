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

func ipOutcome(cause uint32, diag avp.Optional[uint32]) avp.Optional[s6c.IPSMGWSMDeliveryOutcome] {
	return avp.Some(s6c.IPSMGWSMDeliveryOutcome{SMDeliveryCause: avp.Some(cause), AbsentUserDiagnosticSM: diag})
}
func TestS6cIPOutcomeFrozenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"capacity", "absent", "diag11", "diag12", "success", "msc-sgsn-ip", "mme-ip"} {
		r := decoded(t, c, "b31-rdr-"+name)
		if p := c.ValidateContext(r, validation.Context{}); !p.Complete {
			t.Fatal(p)
		}
		for _, ans := range []string{"v24-s6c-rd-a", "f24-s6c-rd-5001", "f24-s6c-rd-5558"} {
			a := decoded(t, c, ans)
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, routingSources(), validation.Context{}), c.ValidateResponseContext(r, a, routingSources(), validation.Context{})} {
				if !p.Complete {
					t.Fatal(name, ans, p)
				}
			}
		}
		e, err := c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 5012}, routingSources().Terminals[0])
		if err != nil {
			t.Fatal(err)
		}
		if p := c.ValidateAnswerContext(r, e, routingSources(), validation.Context{}); !p.Complete {
			t.Fatal(p)
		}
		raw := encoded(t, c, r)
		owned, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, r), encoded(t, c, owned)) {
			t.Fatal("borrowed report")
		}
		n, err := c.Measure(r)
		if err != nil || n != len(encoded(t, c, r)) {
			t.Fatal(n, err)
		}
		if bound, err := c.DecodeBound(r.Header, n); err != nil || bound < uint64(n) {
			t.Fatal(bound, err)
		}
		canon, err := c.Canonical(fixture(t, "b31-rdr-"+name), 8192)
		if err != nil {
			t.Fatal(err)
		}
		again, err := c.Canonical(encoded(t, c, r), 8192)
		if err != nil || !bytes.Equal(canon, again) {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"nocause", "diag256", "success-diag", "dupgroup", "vendorcause", "parentdiag"} {
		if _, err := c.Decode(fixture(t, "b31-rdr-"+name)); err == nil {
			t.Fatal(name)
		}
	}
}
func TestS6cIPOutcomeNodeAndDiagnosticMatrix(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for mask := 0; mask < 8; mask++ {
		for cause := uint32(0); cause < 4; cause++ {
			m := decoded(t, c, "b31-rdr-diag11")
			r := m.Body.(*s6c.RDR)
			r.SMDeliveryOutcome = multiOutcome(mask, 1, avp.Optional[uint32]{})
			r.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome = ipOutcome(cause, avp.Optional[uint32]{})
			err := c.Validate(m)
			if (err == nil) != (mask&3 != 3 && cause <= 2) {
				t.Fatal(mask, cause, err)
			}
		}
	}
	// Presence is not a default. Only11/12 are selected; all other octet values
	// can be retained and reported, but never inherit another parent's meaning.
	for d := uint32(0); d <= 256; d++ {
		raw := rewriteAVP(t, fixture(t, "b31-rdr-diag11"), []uint32{3316, 3320, 3322}, func(v []byte) []byte { v[12] = 0; v[13] = 0; v[14] = byte(d >> 8); v[15] = byte(d); return v })
		m, err := c.Decode(raw)
		if d == 256 {
			if err == nil {
				t.Fatal(d)
			}
			continue
		}
		if err != nil {
			t.Fatal(d, err)
		}
		clear(raw)
		if m.Body.(*s6c.RDR).SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value != d {
			t.Fatal("receive lost value")
		}
		report := c.ValidateContext(m, validation.Context{})
		if report.Complete != (d == 11 || d == 12) {
			t.Fatal(d, report)
		}
		if d != 11 && d != 12 && !errors.Is(report.Err(), validation.ErrUnsupported) {
			t.Fatal(d, report)
		}
		if (c.Validate(m) == nil) != (d == 11 || d == 12) {
			t.Fatal("send scope", d)
		}
	}
	for _, name := range []string{"diag0", "diag13", "cause3"} {
		m := decoded(t, c, "b31-rdr-"+name)
		if p := c.ValidateContext(m, validation.Context{}); p.State != validation.Rejected || !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(p)
		}
	}
	for _, cause := range []uint32{0, 2} {
		m := decoded(t, c, "b31-rdr-diag11")
		m.Body.(*s6c.RDR).SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value.SMDeliveryCause.Value = cause
		if err := c.Validate(m); !errors.Is(err, validation.ErrField) {
			t.Fatal(err)
		}
	}
}
func TestS6cIPOutcomeParentsCanonicalAndIsolation(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, path := range [][]uint32{{3316, 3320}, {3316, 3320, 3321}, {3316, 3320, 3322}} {
		raw := rewriteAVP(t, fixture(t, "b31-rdr-diag11"), path, func(v []byte) []byte { return append(v, v...) })
		if _, err := c.Decode(raw); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	for _, code := range []uint32{3321, 3322} {
		raw := rewriteAVP(t, fixture(t, "b31-rdr-diag11"), []uint32{3316, 3320, code}, func(v []byte) []byte { v[7] = 15; return v })
		if _, err := c.Decode(raw); err == nil {
			t.Fatal(code)
		}
	}
	base, err := c.Canonical(fixture(t, "b31-rdr-msc-sgsn-ip"), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*s6c.RDR){
		func(r *s6c.RDR) {
			r.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Present = false
		},
		func(r *s6c.RDR) { r.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value = 12 },
		func(r *s6c.RDR) {
			r.SMDeliveryOutcome.MSCSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Present = false
		},
	} {
		m := decoded(t, c, "b31-rdr-msc-sgsn-ip")
		edit(m.Body.(*s6c.RDR))
		b, err := c.Canonical(encoded(t, c, m), 8192)
		if err != nil || bytes.Equal(base, b) {
			t.Fatal("parent information collapsed", err)
		}
	}
	r := decoded(t, c, "b31-rdr-diag11")
	body := r.Body.(*s6c.RDR)
	body.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3324}, Flags: 128}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal("typed SMSMI cannot hide in Extra", err)
	}
	r = decoded(t, c, "b31-rdr-diag11")
	body = r.Body.(*s6c.RDR)
	body.SMDeliveryOutcome.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3336}, Flags: 128}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
	body.SMDeliveryOutcome.Extra = []avp.Raw{{Key: avp.Key{VendorID: 99999, Code: 3320}, Flags: 128, Data: []byte{1, 2, 3}}}
	if err := c.Validate(r); err != nil {
		t.Fatal(err)
	}
	// A known typed field cannot be smuggled through Extra, even with M cleared.
	body.SMDeliveryOutcome.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3320}, Flags: 128}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal(err)
	}
}
