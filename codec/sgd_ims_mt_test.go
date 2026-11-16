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

func imsMTEvidence(r pdu.Message) sgd.IMSMTForwardingEvidence {
	b := r.Body.(*sgd.TFR)
	m := b.SMSMICorrelationID.Value
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return sgd.IMSMTForwardingEvidence{Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost, Realm: b.DestinationRealm}, HSSID: b.UserName, OriginatingSIPURI: m.OriginatingSIPURI, DestinationSIPURI: m.DestinationSIPURI, SCAddress: sc}
}
func imsMTContext(t *testing.T, e sgd.IMSMTForwardingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.IMSMTForwarding(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSGDIMSMTPairsAndOwned(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"d21-ims-tfr-r", "b36-timer0-start0", "b36-time-absent", "b36-uri-escaped"} {
		r := decoded(t, c, name)
		b := r.Body.(*sgd.TFR)
		switch name {
		case "d21-ims-tfr-r":
			if !b.SMDeliveryTimer.Present || b.SMDeliveryTimer.Value != 30 || !b.SMDeliveryStartTime.Present || b.SMDeliveryStartTime.Value != 0xed2723f0 {
				t.Fatal("frozen timer", b)
			}
		case "b36-timer0-start0":
			if !b.SMDeliveryTimer.Present || b.SMDeliveryTimer.Value != 0 || !b.SMDeliveryStartTime.Present || b.SMDeliveryStartTime.Value != 0 {
				t.Fatal("present zero", b)
			}
		case "b36-time-absent":
			if b.SMDeliveryTimer.Present || b.SMDeliveryStartTime.Present {
				t.Fatal("absent time", b)
			}
		}
		e := imsMTEvidence(r)
		ctx := imsMTContext(t, e)
		src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}, ErrorAgents: []pdu.Origin{{Host: "dra.example", Realm: e.Responder.Realm}}}
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		for _, aName := range []string{"d21-ims-tfr-a", "d21-ims-tfr-f", "b36-tfa-5552", "b36-cause1", "b36-cause2"} {
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
				body = &sgd.TFA{AuthSessionState: 1, ResultCode: avp.Some(code)}
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
func TestSGDIMSMTEvidenceAndBudgets(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "d21-ims-tfr-r")
	e := imsMTEvidence(r)
	ctx := imsMTContext(t, e)
	// New evidence retains endpoints/HSS-ID/SC plus an inline URI digest. Both
	// original URI strings remain in the budgeted request, not in Context storage.
	if reflect.TypeFor[validation.Context]().Size()+4*253+30 > 2048 {
		t.Fatal("new context bound")
	}
	if reflect.TypeFor[validation.Context]().Size()+4*253+3*15+256 > 2048 {
		t.Fatal("older context bound")
	}
	for _, edit := range []func(*sgd.IMSMTForwardingEvidence){func(e *sgd.IMSMTForwardingEvidence) { e.HSSID = "001010000000002" }, func(e *sgd.IMSMTForwardingEvidence) { e.OriginatingSIPURI = "sip:%61lice@example.net" }, func(e *sgd.IMSMTForwardingEvidence) { e.DestinationSIPURI = "sip:bob@Example.net" }, func(e *sgd.IMSMTForwardingEvidence) {
		e.OriginatingSIPURI, e.DestinationSIPURI = e.DestinationSIPURI, e.OriginatingSIPURI
	}, func(e *sgd.IMSMTForwardingEvidence) { e.SCAddress = "123" }, func(e *sgd.IMSMTForwardingEvidence) { e.Responder.Host = "dra.example" }, func(e *sgd.IMSMTForwardingEvidence) { e.Requester.Realm = "other.example" }} {
		x := e
		edit(&x)
		if p := c.ValidateContext(r, imsMTContext(t, x)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, edit := range []func(*sgd.IMSMTForwardingEvidence){func(e *sgd.IMSMTForwardingEvidence) { e.HSSID = "00101" }, func(e *sgd.IMSMTForwardingEvidence) { e.HSSID = "0010100000000001" }, func(e *sgd.IMSMTForwardingEvidence) { e.SCAddress = "+123" }, func(e *sgd.IMSMTForwardingEvidence) { e.OriginatingSIPURI = "sip:@example.net" }, func(e *sgd.IMSMTForwardingEvidence) { e.DestinationSIPURI = "sip:bob@example.net;transport=tcp" }, func(e *sgd.IMSMTForwardingEvidence) { e.Requester.Host = "" }} {
		x := e
		edit(&x)
		if _, err := sgd.IMSMTForwarding(x); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	e.Requester.Host = strings.ToUpper(e.Requester.Host)
	if p := c.ValidateContext(r, imsMTContext(t, e)); !p.Complete {
		t.Fatal(p)
	}
	e.HSSID = "001010000000002"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("mutated evidence", p)
	}
	e = imsMTEvidence(r)
	e.OriginatingSIPURI = "sip:" + strings.Repeat("x", 240) + "@example.net"
	e.DestinationSIPURI = e.OriginatingSIPURI
	if _, err := sgd.IMSMTForwarding(e); err != nil {
		t.Fatal(err)
	}
	e.OriginatingSIPURI += "x"
	if _, err := sgd.IMSMTForwarding(e); err == nil {
		t.Fatal("URI bound")
	}
	e = imsMTEvidence(r)
	ordinary := moContext(t, sgd.MOForwardingEvidence{Interface: sgd.MOInterfaceSGd, Requester: e.Requester, Responder: e.Responder, SenderIMSI: "001010123456789", SCAddress: e.SCAddress})
	t4, err := sgd.T4Charging(sgd.T4ChargingEvidence{Requester: e.Requester, Responder: e.Responder, SenderIMSI: "001010123456789", SCAddress: e.SCAddress})
	if err != nil {
		t.Fatal(err)
	}
	report, err := s6c.IMSDeliveryReport(s6c.IMSDeliveryReportEvidence{Requester: e.Requester, Responder: e.Responder, HSSID: e.HSSID, DestinationSIPURI: e.DestinationSIPURI, SCAddress: e.SCAddress})
	if err != nil {
		t.Fatal(err)
	}
	for _, wrong := range []validation.Context{ordinary, t4, report, retransmissionContext(t, retransmissionEvidence(r, 0)), imsMOContext(t, imsMOEvidence(decoded(t, c, "v21-ims-ofr-r")))} {
		if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"v-ofr-01", "v-tfr-01"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("cross purpose", p)
		}
	}
	r.Body.(*sgd.TFR).DestinationHost = "other.example"
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
}

func TestSGDIMSMTWireParentsAndResults(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"n26-tfr-hssid-nested", "n27-tfr-uri3326", "n27-tfr-uri3327", "b36-no-origin", "b36-no-destination", "b36-dup-origin", "b36-dup-group", "b36-hssid5", "b36-hssid16", "b36-wrong-vendor", "b36-node1645", "b36-node1489", "b36-cause3"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"b36-uri-params", "b36-max-ret", "b36-gmsc-only"} {
		r := decoded(t, c, name)
		if p := c.ValidateContext(r, validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(name, p)
		}
		if _, err := c.Append(nil, r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(name, err)
		}
	}
	r6 := decoded(t, c, "b36-hssid6")
	if p := c.ValidateContext(r6, imsMTContext(t, imsMTEvidence(r6))); !p.Complete {
		t.Fatal(p)
	}
	for _, path := range [][]uint32{{1}, {3324, 3326}, {3324, 3327}, {3306}, {3307}} {
		if _, err := c.Decode(rewriteAVP(t, fixture(t, "d21-ims-tfr-r"), path, func(b []byte) []byte { return append(b, b...) })); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	r := decoded(t, c, "d21-ims-tfr-r")
	e := imsMTEvidence(r)
	ctx := imsMTContext(t, e)
	sources := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	for _, name := range []string{"b36-tfa-5001", "b36-tfa-5550", "b36-tfa-5551", "b36-tfa-5553", "b36-tfa-5554"} {
		a := decoded(t, c, name)
		for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, sources, ctx), c.ValidateResponseContext(r, a, sources, ctx)} {
			if !errors.Is(p.Err(), validation.ErrUnsupported) {
				t.Fatal(name, p)
			}
		}
	}
	for _, edit := range []func(*pdu.Message){func(a *pdu.Message) { a.Header.HopByHop++ }, func(a *pdu.Message) { a.Header.CommandCode++ }, func(a *pdu.Message) { a.Body.(*sgd.TFA).SessionID += "wrong" }, func(a *pdu.Message) { a.Body.(*sgd.TFA).OriginHost = "dra.example" }} {
		a := decoded(t, c, "d21-ims-tfr-a")
		edit(&a)
		if p := c.ValidateAnswerContext(r, a, sources, ctx); p.Complete {
			t.Fatal(p)
		}
	}
	b := r.Body.(*sgd.TFR)
	for _, code := range []uint32{3325, 3326, 3327} {
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
	b.SMSMICorrelationID.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3325}, Flags: 128, Data: []byte("001010000000001")}}
	if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
}
func TestSGDIMSMTCanonicalAndOpaqueTimeTPDU(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "d21-ims-tfr-r")
	base, err := c.Canonical(encoded(t, c, r), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*sgd.TFR){func(b *sgd.TFR) { b.UserName = "001010000000002" }, func(b *sgd.TFR) { b.SMSMICorrelationID.Value.OriginatingSIPURI = "sip:%61lice@example.net" }, func(b *sgd.TFR) { b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net" }, func(b *sgd.TFR) { b.SCAddress[0] ^= 0x10 }, func(b *sgd.TFR) { b.SMRPUI[0] ^= 1 }, func(b *sgd.TFR) { b.SMDeliveryTimer.Present = false }, func(b *sgd.TFR) { b.SMDeliveryStartTime = avp.Some(avp.Time(0)) }} {
		m, err := c.Snapshot(r)
		if err != nil {
			t.Fatal(err)
		}
		edit(m.Body.(*sgd.TFR))
		x, err := c.Canonical(encoded(t, c, m), 8192)
		if err != nil || bytes.Equal(base, x) {
			t.Fatal("collapsed content", err)
		}
	}
	b := r.Body.(*sgd.TFR)
	b.SMDeliveryTimer = avp.Some(^uint32(0))
	b.SMDeliveryStartTime = avp.Some(avp.Time(^uint32(0)))
	if p := c.ValidateContext(r, imsMTContext(t, imsMTEvidence(r))); !p.Complete {
		t.Fatal(p)
	}
	b.SMRPUI = make([]byte, 201)
	if err := c.Validate(r); !errors.Is(err, validation.ErrField) {
		t.Fatal(err)
	}
}
