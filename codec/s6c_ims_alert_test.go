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

func imsAlertEvidence(r pdu.Message) s6c.IMSHSSAlertEvidence {
	b := r.Body.(*s6c.ALR)
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return s6c.IMSHSSAlertEvidence{Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}, DestinationSIPURI: b.SMSMICorrelationID.Value.DestinationSIPURI, SCAddress: sc}
}
func imsAlertContext(t *testing.T, e s6c.IMSHSSAlertEvidence) validation.Context {
	t.Helper()
	ctx, err := s6c.IMSHSSAlert(e)
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}
func TestS6cIMSAlertPairsAndOwned(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v21-ims-alr-r", "b33-time0", "b33-time-max", "b33-uri-escaped"} {
		r := decoded(t, c, name)
		e := imsAlertEvidence(r)
		ctx := imsAlertContext(t, e)
		src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}, ErrorAgents: []pdu.Origin{{Host: "dra.example", Realm: "example.net"}}}
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		for _, answer := range []string{"v21-ims-alr-a", "v21-ims-alr-f", "b33-ala-5012"} {
			a := decoded(t, c, answer)
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
				if !p.Complete {
					t.Fatal(name, answer, p)
				}
			}
			p := c.ValidateAnswerContext(r, a, src, validation.Context{})
			if (a.Header.Flags&pdu.ErrorFlag != 0) != p.Complete {
				t.Fatal(p)
			}
			a.Header.EndToEnd++
			if p := c.ValidateAnswerContext(r, a, src, ctx); p.Complete {
				t.Fatal("unmatched answer")
			}
		}
		// E agent is authorized independently of the declared IWMSC terminal.
		a, err := c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 3004}, src.ErrorAgents[0])
		if err != nil {
			t.Fatal(err)
		}
		if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
			t.Fatal(p)
		}
		raw := fixture(t, name)
		m, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, r), encoded(t, c, m)) {
			t.Fatal("borrowed data")
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
func TestS6cIMSAlertPurposeAndBounds(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v21-ims-alr-r")
	e := imsAlertEvidence(r)
	ctx := imsAlertContext(t, e)
	if reflect.TypeFor[validation.Context]().Size()+4*253+15+256 > 2048 {
		t.Fatal("context exceeds existing reservation")
	}
	for _, edit := range []func(*s6c.IMSHSSAlertEvidence){func(e *s6c.IMSHSSAlertEvidence) { e.SCAddress = "123" }, func(e *s6c.IMSHSSAlertEvidence) { e.DestinationSIPURI = "sip:bob@Example.net" }, func(e *s6c.IMSHSSAlertEvidence) { e.DestinationSIPURI = "sip:%62ob@example.net" }, func(e *s6c.IMSHSSAlertEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.IMSHSSAlertEvidence) { e.Responder.Realm = "other.example" }} {
		bad := e
		edit(&bad)
		if p := c.ValidateContext(r, imsAlertContext(t, bad)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, bad := range []s6c.IMSHSSAlertEvidence{{}, {Requester: e.Requester, Responder: e.Responder, DestinationSIPURI: e.DestinationSIPURI, SCAddress: "+123"}, {Requester: e.Requester, Responder: e.Responder, SCAddress: e.SCAddress, DestinationSIPURI: "sip:@example.net"}} {
		if _, err := s6c.IMSHSSAlert(bad); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	e.Requester.Host = strings.ToUpper(e.Requester.Host)
	if p := c.ValidateContext(r, imsAlertContext(t, e)); !p.Complete {
		t.Fatal(p)
	}
	e.DestinationSIPURI = "sip:edited@example.net"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("borrowed evidence", p)
	}
	e = imsAlertEvidence(r)
	e.DestinationSIPURI = "sip:" + strings.Repeat("x", 240) + "@example.net"
	if _, err := s6c.IMSHSSAlert(e); err != nil {
		t.Fatal(err)
	}
	e.DestinationSIPURI += "x"
	if _, err := s6c.IMSHSSAlert(e); err == nil {
		t.Fatal("unbounded evidence")
	}
	role, err := s6c.AlertRole(s6c.AlertRoleEvidence{Role: s6c.AlertFromHSS, Requester: imsAlertEvidence(r).Requester, Responder: imsAlertEvidence(r).Responder})
	if err != nil {
		t.Fatal(err)
	}
	report := imsReportContext(t, s6c.IMSDeliveryReportEvidence{Requester: imsAlertEvidence(r).Requester, Responder: imsAlertEvidence(r).Responder, HSSID: "001010000000001", DestinationSIPURI: imsAlertEvidence(r).DestinationSIPURI, SCAddress: imsAlertEvidence(r).SCAddress})
	t4 := s6cT4Context(t, s6c.T4NoMSISDNEvidence{Procedure: s6c.T4HSSAlert, Requester: imsAlertEvidence(r).Requester, Responder: imsAlertEvidence(r).Responder, IMSI: "001010123456789", SCAddress: imsAlertEvidence(r).SCAddress})
	for _, wrong := range []validation.Context{role, report, t4} {
		if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"v21-ims-rdr-r", "d21-t4-alr-r", "v-alr-01"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("purpose escaped", name)
		}
	}
	r.Body.(*s6c.ALR).SMSMICorrelationID.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal("dummy did not prove IMS", p)
	}
}
func TestS6cIMSAlertIdentityParentsAndResults(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"n26-alr-hssid-nested", "n26-alr-originuri", "n27-alr-uri3327", "b33-no-uri", "b33-dup-uri", "b33-dup-group", "b33-dummy14", "b33-dummy16", "b33-real-number", "b33-username", "b33-both-ids", "b33-wrong-vendor"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	r := decoded(t, c, "b33-uri-params")
	if p := c.ValidateContext(r, validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
	for _, path := range [][]uint32{{3102}, {3102, 701}, {3324}, {3324, 3327}, {3329}} {
		raw := rewriteAVP(t, fixture(t, "b33-time0"), path, func(v []byte) []byte { return append(v, v...) })
		if _, err := c.Decode(raw); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	raw := rewriteAVP(t, fixture(t, "b33-time0"), []uint32{3329}, func(v []byte) []byte { v[7] = 15; return v })
	if _, err := c.Decode(raw); err == nil {
		t.Fatal("Time width")
	}
	for _, edit := range []func(*s6c.ALR){func(b *s6c.ALR) { b.SMSGMSCAlertEvent = avp.Some(uint32(1)) }, func(b *s6c.ALR) { b.ServingNode = avp.Some(s6c.ServingNode{SGSNNumber: avp.Some([]byte{0x21})}) }} {
		r := decoded(t, c, "v21-ims-alr-r")
		edit(r.Body.(*s6c.ALR))
		if err := c.Validate(r); err == nil {
			t.Fatal("access procedure escaped")
		}
	}
	r = decoded(t, c, "v21-ims-alr-r")
	b := r.Body.(*s6c.ALR)
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
	a := decoded(t, c, "v21-ims-alr-a")
	a.Body.(*s6c.ALA).ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})
	a.Body.(*s6c.ALA).ResultCode.Present = false
	if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("invented ALA experiment", err)
	}
}
func TestS6cIMSAlertCanonicalURIAndTime(t *testing.T) {
	c := newCodec(t, codec.Config{})
	base, err := c.Canonical(fixture(t, "v21-ims-alr-r"), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b33-time0", "b33-time-max", "b33-uri-escaped"} {
		v, err := c.Canonical(fixture(t, name), 8192)
		if err != nil || bytes.Equal(base, v) {
			t.Fatal("presence or identity collapsed", err)
		}
	}
	r := decoded(t, c, "v21-ims-alr-r")
	r.Body.(*s6c.ALR).SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net"
	v, err := c.Canonical(encoded(t, c, r), 8192)
	if err != nil || bytes.Equal(base, v) {
		t.Fatal(err)
	}
}
