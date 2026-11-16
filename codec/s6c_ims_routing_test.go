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

func imsRoutingEvidence(r pdu.Message) s6c.IMSRoutingEvidence {
	b := r.Body.(*s6c.SRR)
	sc, _ := avp.DecodeTBCD(b.SCAddress)
	return s6c.IMSRoutingEvidence{Role: s6c.RoutingFromGMSC, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, HSSID: b.UserName.Value, DestinationSIPURI: b.SMSMICorrelationID.Value.DestinationSIPURI, SCAddress: sc}
}
func imsRoutingContext(t *testing.T, e s6c.IMSRoutingEvidence) validation.Context {
	t.Helper()
	ctx, err := s6c.IMSRouting(e)
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}
func TestS6cIMSRoutingPairsAndOwned(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"v21-ims-srr-r", "b34-hssid6", "b34-uri-escaped"} {
		r := decoded(t, c, name)
		e := imsRoutingEvidence(r)
		ctx := imsRoutingContext(t, e)
		src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}, ErrorAgents: []pdu.Origin{{Host: "dra.example", Realm: "example.net"}}}
		if p := c.ValidateContext(r, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		for _, answer := range []string{"v21-ims-srr-a", "v21-ims-srr-f", "b34-sra-5550", "b34-sra-5552", "b34-sra-5556", "b34-sra-5557"} {
			a := decoded(t, c, answer)
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
				if !p.Complete {
					t.Fatal(name, answer, p)
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
			t.Fatal("borrowed input")
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
func TestS6cIMSRoutingEvidenceAndBounds(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v21-ims-srr-r")
	e := imsRoutingEvidence(r)
	ctx := imsRoutingContext(t, e)
	if reflect.TypeFor[validation.Context]().Size()+4*253+30+256 > 2048 {
		t.Fatal("existing context reserve exceeded")
	}
	for _, edit := range []func(*s6c.IMSRoutingEvidence){func(e *s6c.IMSRoutingEvidence) { e.HSSID = "001010000000002" }, func(e *s6c.IMSRoutingEvidence) { e.DestinationSIPURI = "sip:%62ob@example.net" }, func(e *s6c.IMSRoutingEvidence) { e.SCAddress = "123" }, func(e *s6c.IMSRoutingEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.IMSRoutingEvidence) { e.Responder.Realm = "other.example" }} {
		bad := e
		edit(&bad)
		if p := c.ValidateContext(r, imsRoutingContext(t, bad)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, edit := range []func(*s6c.IMSRoutingEvidence){func(e *s6c.IMSRoutingEvidence) { e.Role = 0 }, func(e *s6c.IMSRoutingEvidence) { e.UNRI = avp.Some(false) }, func(e *s6c.IMSRoutingEvidence) { e.HSSID = "00101" }, func(e *s6c.IMSRoutingEvidence) { e.SCAddress = "+123" }, func(e *s6c.IMSRoutingEvidence) { e.DestinationSIPURI = "sip:@example.net" }, func(e *s6c.IMSRoutingEvidence) { e.Responder.Host = "" }} {
		bad := e
		edit(&bad)
		if _, err := s6c.IMSRouting(bad); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	e.Requester.Host = strings.ToUpper(e.Requester.Host)
	if p := c.ValidateContext(r, imsRoutingContext(t, e)); !p.Complete {
		t.Fatal(p)
	}
	e.DestinationSIPURI = "sip:changed@example.net"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("borrowed context", p)
	}
	e = imsRoutingEvidence(r)
	e.DestinationSIPURI = "sip:" + strings.Repeat("x", 240) + "@example.net"
	if _, err := s6c.IMSRouting(e); err != nil {
		t.Fatal(err)
	}
	e.DestinationSIPURI += "x"
	if _, err := s6c.IMSRouting(e); err == nil {
		t.Fatal("unbounded URI")
	}
	role, err := s6c.RoutingRole(s6c.RoutingRoleEvidence{Role: s6c.RoutingFromGMSC, Requester: e.Requester, Responder: e.Responder})
	if err != nil {
		t.Fatal(err)
	}
	t4 := s6cT4Context(t, s6c.T4NoMSISDNEvidence{Procedure: s6c.T4Routing, Requester: e.Requester, Responder: e.Responder, IMSI: "001010000000001", SCAddress: e.SCAddress})
	report := imsReportContext(t, s6c.IMSDeliveryReportEvidence{Requester: e.Requester, Responder: e.Responder, HSSID: e.HSSID, DestinationSIPURI: "sip:bob@example.net", SCAddress: e.SCAddress})
	for _, wrong := range []validation.Context{role, t4, report} {
		if p := c.ValidateContext(r, wrong); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"d21-t4-srr-r", "v21-ims-rdr-r", "v21-ims-alr-r", "v-srr-01"} {
		if p := c.ValidateContext(decoded(t, c, name), ctx); p.Complete {
			t.Fatal("purpose escaped", name)
		}
	}
	r.Body.(*s6c.SRR).DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
}
func TestS6cIMSRoutingResultIdentityAndUNRI(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "v21-ims-srr-r")
	e := imsRoutingEvidence(r)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	// Equality cannot prove allocated subscriber identity. It must not become a
	// protocol inequality rule; applications must never synthesize IMSI from HSS-ID.
	for _, name := range []string{"v21-ims-srr-a", "b34-sra-samebytes"} {
		if p := c.ValidateAnswerContext(r, decoded(t, c, name), src, imsRoutingContext(t, e)); !p.Complete {
			t.Fatal(p)
		}
	}
	for _, role := range []s6c.RoutingOriginRole{s6c.RoutingFromGMSC, s6c.RoutingFromIPSMGW} {
		for _, unri := range []avp.Optional[bool]{{}, avp.Some(false), avp.Some(true)} {
			if role == s6c.RoutingFromGMSC && unri.Present {
				continue
			}
			e.Role = role
			e.UNRI = unri
			ctx := imsRoutingContext(t, e)
			if p := c.ValidateContext(r, ctx); !p.Complete {
				t.Fatal(p)
			}
			a := decoded(t, c, "b34-sra-5550")
			p := c.ValidateAnswerContext(r, a, src, ctx)
			switch {
			case role == s6c.RoutingFromGMSC || unri.Present && unri.Value:
				if !p.Complete {
					t.Fatal(p)
				}
			case !unri.Present:
				if p.State != validation.NeedsApplication {
					t.Fatal(p)
				}
			default:
				if !errors.Is(p.Err(), validation.ErrMissing) {
					t.Fatal(p)
				}
			}
			a.Body.(*s6c.SRA).UserName = avp.Some("001010123456789")
			if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
				t.Fatal(p)
			}
		}
	}
	// Additional node still depends on this original request's GPRS bit; IMS
	// identity selection does not weaken existing node/Realm selection rules.
	e = imsRoutingEvidence(r)
	ctx := imsRoutingContext(t, e)
	a := decoded(t, c, "v21-ims-srr-a")
	a.Body.(*s6c.SRA).ServingNode = avp.Some(s6c.ServingNode{MSCNumber: avp.Some([]byte{0x21})})
	a.Body.(*s6c.SRA).AdditionalServingNode = avp.Some(s6c.AdditionalServingNode{SGSNNumber: avp.Some([]byte{0x21})})
	if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrInvalidAnswer) {
		t.Fatal(p)
	}
	r.Body.(*s6c.SRR).SRRFlags = avp.Some(uint32(1))
	smea, err := s6c.EncodeSMEA("12345")
	if err != nil {
		t.Fatal(err)
	}
	r.Body.(*s6c.SRR).SMRPSMEA = avp.Some(smea)
	if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
		t.Fatal(p)
	}
	a = decoded(t, c, "v21-ims-srr-a")
	a.Body.(*s6c.SRA).ServingNode.Value.IPSMGWRealm = avp.Some("example.net")
	if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
}
func TestS6cIMSRoutingWireAndCanonical(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"n26-srr-hssid-nested", "n26-srr-originuri", "n27-srr-uri3327", "b34-no-uri", "b34-dup-uri", "b34-dup-group", "b34-hssid5", "b34-hssid16", "b34-both-ids", "b34-wrong-vendor", "b34-sra-noimsi"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	r := decoded(t, c, "b34-uri-params")
	if p := c.ValidateContext(r, validation.Context{}); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
	for _, path := range [][]uint32{{1}, {3324}, {3324, 3327}} {
		if _, err := c.Decode(rewriteAVP(t, fixture(t, "v21-ims-srr-r"), path, func(v []byte) []byte { return append(v, v...) })); !errors.Is(err, validation.ErrDuplicate) {
			t.Fatal(path, err)
		}
	}
	r = decoded(t, c, "v21-ims-srr-r")
	b := r.Body.(*s6c.SRR)
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
	base, err := c.Canonical(fixture(t, "v21-ims-srr-r"), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b34-uri-escaped", "b34-hssid6"} {
		v, err := c.Canonical(fixture(t, name), 8192)
		if err != nil || bytes.Equal(base, v) {
			t.Fatal("identity collapsed", err)
		}
	}
}
