package codec_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func imsReportEvidence(r pdu.Message) s6c.IMSDeliveryReportEvidence {
	b := r.Body.(*s6c.RDR)
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return s6c.IMSDeliveryReportEvidence{Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, HSSID: b.UserIdentifier.UserName.Value, DestinationSIPURI: b.SMSMICorrelationID.Value.DestinationSIPURI, SCAddress: sc}
}
func imsReportContext(t *testing.T, e s6c.IMSDeliveryReportEvidence) validation.Context {
	t.Helper()
	v, err := s6c.IMSDeliveryReport(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestS6cIMSReportPairsAndOwned(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v21-ims-rdr-r", "b32-hssid6", "b32-uri-escaped"} {
		r := decoded(t, c, name)
		e := imsReportEvidence(r)
		ctx := imsReportContext(t, e)
		sources := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		for _, answer := range []string{"v21-ims-rdr-a", "v21-ims-rdr-f", "b32-rda-5001"} {
			a := decoded(t, c, answer)
			if p := c.ValidateAnswerContext(r, a, sources, ctx); !p.Complete {
				t.Fatal(p)
			}
			if p := c.ValidateResponseContext(r, a, sources, ctx); !p.Complete {
				t.Fatal(p)
			}
			if p := c.ValidateAnswerContext(r, a, sources, validation.Context{}); p.State != validation.NeedsApplication {
				t.Fatal(p)
			}
			a.Header.EndToEnd++
			if p := c.ValidateAnswerContext(r, a, sources, ctx); p.Complete {
				t.Fatal("wrong original accepted")
			}
		}
		a, err := c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 5012}, e.Responder)
		if err != nil {
			t.Fatal(err)
		}
		if p := c.ValidateAnswerContext(r, a, sources, validation.Context{}); !p.Complete {
			t.Fatal(p)
		}
		raw := fixture(t, name)
		m, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, r), encoded(t, c, m)) {
			t.Fatal("borrowed field")
		}
		n, err := c.Measure(r)
		if err != nil || n != len(encoded(t, c, r)) {
			t.Fatal(n, err)
		}
		if bound, err := c.DecodeBound(r.Header, n); err != nil || bound < uint64(n) {
			t.Fatal(bound, err)
		}
		canon, err := c.Canonical(fixture(t, name), 8192)
		if err != nil {
			t.Fatal(err)
		}
		again, err := c.Canonical(encoded(t, c, r), 8192)
		if err != nil || !bytes.Equal(canon, again) {
			t.Fatal(err)
		}
	}
}
func TestS6cIMSReportEvidenceAndBounds(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v21-ims-rdr-r")
	e := imsReportEvidence(r)
	if reflect.TypeFor[validation.Context]().Size()+4*253+2*15+256 > 2048 {
		t.Fatal("IMS evidence exceeds reservation")
	}
	for _, edit := range []func(*s6c.IMSDeliveryReportEvidence){
		func(e *s6c.IMSDeliveryReportEvidence) { e.HSSID = "001010000000002" }, func(e *s6c.IMSDeliveryReportEvidence) { e.SCAddress = "123" }, func(e *s6c.IMSDeliveryReportEvidence) { e.DestinationSIPURI = "sip:Bob@example.net" }, func(e *s6c.IMSDeliveryReportEvidence) { e.DestinationSIPURI = "sip:bob@Example.net" }, func(e *s6c.IMSDeliveryReportEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.IMSDeliveryReportEvidence) { e.Responder.Realm = "other.example" },
	} {
		bad := e
		edit(&bad)
		if p := c.ValidateContext(r, imsReportContext(t, bad)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	ctx := imsReportContext(t, e)
	e.HSSID = "999999"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("factory borrowed input", p)
	}
	e = imsReportEvidence(r)
	e.Requester.Host = strings.ToUpper(e.Requester.Host)
	if p := c.ValidateContext(r, imsReportContext(t, e)); !p.Complete {
		t.Fatal(p)
	}
	for _, bad := range []string{"00101", "0010100000000001", "00101x", ""} {
		e := imsReportEvidence(r)
		e.HSSID = bad
		if _, err := s6c.IMSDeliveryReport(e); !errors.Is(err, validation.ErrContext) {
			t.Fatal(bad, err)
		}
	}
	for _, bad := range []string{"sip:@example.net", "sip:bob@", "sip:%xx@example.net", "sips:bob@example.net", "sip:bob@[::1]", "sip:bob@example.net;transport=tcp", "sip:bob@example.net:5060", "sip:用户@example.net", "sip:bob@例子.net", "sip:bob@127.0.0.1"} {
		e := imsReportEvidence(r)
		e.DestinationSIPURI = bad
		if _, err := s6c.IMSDeliveryReport(e); err == nil {
			t.Fatal(bad)
		}
	}
	e = imsReportEvidence(r)
	e.DestinationSIPURI = "sip:" + strings.Repeat("x", 240) + "@example.net"
	if _, err := s6c.IMSDeliveryReport(e); err != nil {
		t.Fatal(err)
	}
	e.DestinationSIPURI = "sip:" + strings.Repeat("x", 241) + "@example.net"
	if _, err := s6c.IMSDeliveryReport(e); err == nil {
		t.Fatal("unbounded URI")
	}
	// Equal digit strings do not turn a T4 assertion into IMS evidence.
	t4 := s6cT4Evidence(r)
	if p := c.ValidateContext(r, s6cT4Context(t, t4)); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	r.Body.(*s6c.RDR).DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	for _, name := range []string{"b26-t4-rdr-mme", "d21-t4-alr-r", "b31-rdr-diag11"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("context crossed procedure", name)
		}
	}
}
func TestS6cIMSReportFieldNegatives(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"n26-rdr-username-text", "n26-rdr-hssid-nested", "n26-rdr-originuri", "n27-rdr-uri3327", "b32-no-uri", "b32-dup-uri", "b32-dup-group", "b32-hssid5", "b32-hssid16", "b32-both-ids", "b32-msisdn-only", "b32-wrong-vendor"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"b32-uri-params", "b32-uri257"} {
		r := decoded(t, c, name)
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.Rejected || !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(p)
		}
		if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	for _, path := range [][]uint32{{3102, 1}, {3324}, {3324, 3327}} {
		raw := rewriteAVP(t, fixture(t, "v21-ims-rdr-r"), path, func(v []byte) []byte { return append(v, v...) })
		if _, err := c.Decode(raw); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	r := decoded(t, c, "v21-ims-rdr-r")
	b := r.Body.(*s6c.RDR)
	for _, code := range []uint32{3325, 3326, 3327} {
		b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: code}, Flags: 128}}
		if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(code, err)
		}
	}
	b.Extra = nil
	b.SMSMICorrelationID.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 99999, Code: 3325}, Flags: 128, Data: []byte{0xff}}}
	if err := c.Validate(r); err != nil {
		t.Fatal(err)
	}
	b.SMSMICorrelationID.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3327}, Flags: 128}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal(err)
	}
}
func TestS6cIMSReportCanonicalExactIdentities(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v21-ims-rdr-r")
	base, err := c.Canonical(encoded(t, c, r), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*s6c.RDR){func(b *s6c.RDR) { b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net" }, func(b *s6c.RDR) { b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:%62ob@example.net" }, func(b *s6c.RDR) { b.UserIdentifier.UserName.Value = "001010000000002" }} {
		r := decoded(t, c, "v21-ims-rdr-r")
		edit(r.Body.(*s6c.RDR))
		b, err := c.Canonical(encoded(t, c, r), 8192)
		if err != nil || bytes.Equal(base, b) {
			t.Fatal("identity normalized", err)
		}
	}
}
