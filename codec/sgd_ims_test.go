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
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"reflect"
	"strings"
	"testing"
)

func imsMOEvidence(r pdu.Message) sgd.IMSMOForwardingEvidence {
	b := r.Body.(*sgd.OFR)
	m := b.SMSMICorrelationID.Value
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	e := sgd.IMSMOForwardingEvidence{Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, HSSID: m.HSSID, OriginatingSIPURI: m.OriginatingSIPURI, DestinationSIPURI: m.DestinationSIPURI, SCAddress: sc}
	if b.UserIdentifier.UserName.Present {
		e.SenderIMSI = b.UserIdentifier.UserName
		if b.UserIdentifier.MSISDN.Present {
			n, _ := avp.DecodeTBCD(b.UserIdentifier.MSISDN.Value)
			e.SenderMSISDN = avp.Some(n)
		}
	}
	return e
}
func imsMOContext(t *testing.T, e sgd.IMSMOForwardingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.IMSMOForwarding(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSGDIMSMOPairsAndOwned(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"v21-ims-ofr-r", "v27-ims-ofr-absent", "b35-diag12", "b35-uri-escaped"} {
		r := decoded(t, c, name)
		e := imsMOEvidence(r)
		ctx := imsMOContext(t, e)
		src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}, ErrorAgents: []pdu.Origin{{Host: "dra.example", Realm: e.Responder.Realm}}}
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		for _, aName := range []string{"v21-ims-ofr-a", "v21-ims-ofr-f", "b35-ofa-5552"} {
			a := decoded(t, c, aName)
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
				if !p.Complete {
					t.Fatal(name, aName, p)
				}
			}
			if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); p.State != validation.NeedsApplication {
				t.Fatal(p)
			}
			a.Header.EndToEnd++
			if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrUnmatchedAnswer) {
				t.Fatal(p)
			}
		}
		for _, code := range []uint32{3004, 5012} {
			var body pdu.Body = &pdu.ErrorAnswer{ResultCode: code}
			origin := src.ErrorAgents[0]
			if code == 5012 {
				body = &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(code)}
				origin = e.Responder
			}
			a, err := c.CompleteAnswer(r, body, origin)
			if err != nil {
				t.Fatal(err)
			}
			if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
				t.Fatal(p)
			}
		}
		raw := fixture(t, name)
		owned, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, r), encoded(t, c, owned)) {
			t.Fatal("borrowed bytes")
		}
		n, err := c.Measure(r)
		if err != nil || n != len(encoded(t, c, r)) {
			t.Fatal(n, err)
		}
		if bound, err := c.DecodeBound(r.Header, n); err != nil || bound < uint64(n) {
			t.Fatal(bound, err)
		}
		x, err := c.Canonical(fixture(t, name), 8192)
		if err != nil {
			t.Fatal(err)
		}
		y, err := c.Canonical(encoded(t, c, r), 8192)
		if err != nil || !bytes.Equal(x, y) {
			t.Fatal(err)
		}
	}
}
func TestSGDIMSMOEvidenceAndBudgets(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "v21-ims-ofr-r")
	e := imsMOEvidence(r)
	ctx := imsMOContext(t, e)
	// New evidence retains endpoints/HSS-ID/SC plus an inline URI digest. Both
	// original URI strings remain in the budgeted request, not in Context storage.
	if reflect.TypeFor[validation.Context]().Size()+4*253+30 > 2048 {
		t.Fatal("new context bound")
	}
	if reflect.TypeFor[validation.Context]().Size()+4*253+3*15+256 > 2048 {
		t.Fatal("older context bound")
	}
	for _, edit := range []func(*sgd.IMSMOForwardingEvidence){func(e *sgd.IMSMOForwardingEvidence) { e.HSSID = "001010000000002" }, func(e *sgd.IMSMOForwardingEvidence) { e.OriginatingSIPURI = "sip:%61lice@example.net" }, func(e *sgd.IMSMOForwardingEvidence) { e.DestinationSIPURI = "sip:bob@Example.net" }, func(e *sgd.IMSMOForwardingEvidence) {
		e.OriginatingSIPURI, e.DestinationSIPURI = e.DestinationSIPURI, e.OriginatingSIPURI
	}, func(e *sgd.IMSMOForwardingEvidence) { e.SCAddress = "123" }, func(e *sgd.IMSMOForwardingEvidence) { e.Responder.Host = "dra.example" }, func(e *sgd.IMSMOForwardingEvidence) { e.Requester.Realm = "other.example" }} {
		x := e
		edit(&x)
		if p := c.ValidateContext(r, imsMOContext(t, x)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, edit := range []func(*sgd.IMSMOForwardingEvidence){func(e *sgd.IMSMOForwardingEvidence) { e.HSSID = "00101" }, func(e *sgd.IMSMOForwardingEvidence) { e.HSSID = "0010100000000001" }, func(e *sgd.IMSMOForwardingEvidence) { e.SCAddress = "+123" }, func(e *sgd.IMSMOForwardingEvidence) { e.OriginatingSIPURI = "sip:@example.net" }, func(e *sgd.IMSMOForwardingEvidence) { e.DestinationSIPURI = "sip:bob@example.net;transport=tcp" }, func(e *sgd.IMSMOForwardingEvidence) { e.Requester.Host = "" }} {
		x := e
		edit(&x)
		if _, err := sgd.IMSMOForwarding(x); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	e.Requester.Host = strings.ToUpper(e.Requester.Host)
	if p := c.ValidateContext(r, imsMOContext(t, e)); !p.Complete {
		t.Fatal(p)
	}
	e.HSSID = "001010000000002"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("mutated evidence", p)
	}
	e = imsMOEvidence(r)
	e.OriginatingSIPURI = "sip:" + strings.Repeat("x", 240) + "@example.net"
	e.DestinationSIPURI = e.OriginatingSIPURI
	if _, err := sgd.IMSMOForwarding(e); err != nil {
		t.Fatal(err)
	}
	e.OriginatingSIPURI += "x"
	if _, err := sgd.IMSMOForwarding(e); err == nil {
		t.Fatal("URI bound")
	}
	e = imsMOEvidence(r)
	ordinary := moContext(t, sgd.MOForwardingEvidence{Interface: sgd.MOInterfaceSGd, Requester: e.Requester, Responder: e.Responder, SenderIMSI: "001010123456789", SCAddress: e.SCAddress})
	t4, err := sgd.T4Charging(sgd.T4ChargingEvidence{Requester: e.Requester, Responder: e.Responder, SenderIMSI: "001010123456789", SCAddress: e.SCAddress})
	if err != nil {
		t.Fatal(err)
	}
	report, err := s6c.IMSDeliveryReport(s6c.IMSDeliveryReportEvidence{Requester: e.Requester, Responder: e.Responder, HSSID: e.HSSID, DestinationSIPURI: e.DestinationSIPURI, SCAddress: e.SCAddress})
	if err != nil {
		t.Fatal(err)
	}
	for _, wrong := range []validation.Context{ordinary, t4, report} {
		if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"v-ofr-01", "v-tfr-01"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("cross purpose", p)
		}
	}
	r.Body.(*sgd.OFR).DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
}
func TestSGDIMSMOWireParentsAndResults(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"n27-ofr-nocause", "n27-ofr-diag256", "b35-no-hssid", "b35-no-origin", "b35-no-destination", "b35-dup-origin", "b35-dup-group", "b35-hssid5", "b35-hssid-utf8", "b35-wrong-vendor", "b35-dummy14", "b35-dummy16", "b35-real-number", "b35-both-ids", "b35-no-outcome", "b35-empty-outcome", "b35-wrong-outcome-parent", "b35-cause0-diag11"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"b35-cause2", "b35-cause99", "b35-diag0", "b35-uri-params"} {
		r := decoded(t, c, name)
		if p := c.ValidateContext(r, validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(name, p)
		}
		if _, err := c.Append(nil, r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(name, err)
		}
	}
	for _, path := range [][]uint32{{3102, 701}, {3324, 3325}, {3324, 3326}, {3324, 3327}, {3316}, {3316, 3320}, {3316, 3320, 3321}, {3316, 3320, 3322}} {
		if _, err := c.Decode(rewriteAVP(t, fixture(t, "v27-ims-ofr-absent"), path, func(b []byte) []byte { return append(b, b...) })); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	for _, path := range [][]uint32{{3324, 3326}, {3324, 3327}} {
		if _, err := c.Decode(rewriteAVP(t, fixture(t, "v21-ims-ofr-r"), path, func(b []byte) []byte { b[12] = 0xc3; b[13] = 0x28; return b })); err == nil {
			t.Fatal("UTF8", path)
		}
	}
	r := decoded(t, c, "v21-ims-ofr-r")
	e := imsMOEvidence(r)
	ctx := imsMOContext(t, e)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	for _, name := range []string{"b35-ofa-5001", "b35-ofa-external"} {
		a := decoded(t, c, name)
		for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
			if !errors.Is(p.Err(), validation.ErrUnsupported) {
				t.Fatal(p)
			}
		}
	}
	// Dummy presence must not turn the ordinary SGd unknown-MSISDN rule into permission.
	for _, cause := range []uint32{3, 4, 5, 6} {
		a := decoded(t, c, "v21-ims-ofr-f")
		a.Body.(*sgd.OFA).SMDeliveryFailureCause.Value.SMEnumeratedDeliveryFailureCause = cause
		if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
			t.Fatal(p)
		}
	}
	b := r.Body.(*sgd.OFR)
	b.OFRFlags = avp.Some(uint32(0))
	if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
	b.OFRFlags.Present = false
	for _, code := range []uint32{3325, 3326, 3327, 3320, 3321, 3322} {
		b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: code}, Flags: 128}}
		if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(code, err)
		}
	}
	b.Extra = nil
	b.SMSMICorrelationID.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 99999, Code: 3325}, Flags: 128, Data: []byte{255}}}
	if err := c.Validate(r); err != nil {
		t.Fatal(err)
	}
	b.SMSMICorrelationID.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3325}, Flags: 128}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal(err)
	}
}
func TestSGDIMSMOCanonicalAndOpaqueTPDU(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "v21-ims-ofr-r")
	base, err := c.Canonical(encoded(t, c, r), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*sgd.OFR){func(b *sgd.OFR) { b.SMSMICorrelationID.Value.HSSID = "001010000000002" }, func(b *sgd.OFR) { b.SMSMICorrelationID.Value.OriginatingSIPURI = "sip:%61lice@example.net" }, func(b *sgd.OFR) { b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net" }, func(b *sgd.OFR) { b.SMDeliveryOutcome.Value.IPSMGWSMDeliveryOutcome.SMDeliveryCause = 1 }, func(b *sgd.OFR) { b.SMRPUI[0] ^= 1 }} {
		m, err := c.Snapshot(r)
		if err != nil {
			t.Fatal(err)
		}
		edit(m.Body.(*sgd.OFR))
		x, err := c.Canonical(encoded(t, c, m), 8192)
		if err != nil || bytes.Equal(base, x) {
			t.Fatal("content collapsed", err)
		}
	}
	b := r.Body.(*sgd.OFR)
	b.UserIdentifier.UserName.Value = "ignored"
	if _, err := c.Snapshot(r); err != nil {
		t.Fatal("absent Value", err)
	}
	b.SMRPUI = make([]byte, 201)
	if err := c.Validate(r); !errors.Is(err, validation.ErrField) {
		t.Fatal(err)
	}
}
