// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func moEvidence(r pdu.Message, iface sgd.MOInterface) sgd.MOForwardingEvidence {
	e := t4Evidence(r)
	return sgd.MOForwardingEvidence{Interface: iface, Requester: e.Requester, Responder: e.Responder, SenderIMSI: e.SenderIMSI, SenderMSISDN: e.SenderMSISDN, SCAddress: e.SCAddress}
}
func moContext(t *testing.T, e sgd.MOForwardingEvidence) validation.Context {
	t.Helper()
	v, err := sgd.MOForwarding(e)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSGDMOInterfaceFrozenPairs(t *testing.T) {
	c := sgdCodec(t)
	for _, tc := range []struct {
		r, a  string
		iface sgd.MOInterface
		want  error
	}{
		{"v24-sgd-mo-r", "f24-sgd-mo-5001", sgd.MOInterfaceSGd, nil},
		{"v24-gdd-mo-r", "c24-gdd-mo-unknown", sgd.MOInterfaceGdd, validation.ErrUnsupported},
	} {
		r, a := decoded(t, c, tc.r), decoded(t, c, tc.a)
		e := moEvidence(r, tc.iface)
		ctx := moContext(t, e)
		src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
		if p := c.ValidateContext(r, ctx); !p.Complete {
			t.Fatal(p)
		}
		if p := c.ValidateAnswerContext(r, a, src, validation.Context{}); p.State != validation.NeedsApplication {
			t.Fatal(p)
		}
		for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
			if !errors.Is(p.Err(), tc.want) || p.Complete != (tc.want == nil) {
				t.Fatal(p)
			}
		}
		raw := encoded(t, c, a)
		if !bytes.Equal(raw, fixture(t, tc.a)) {
			t.Fatal("wire changed")
		}
		owned, err := c.Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		clear(raw)
		if !bytes.Equal(encoded(t, c, owned), fixture(t, tc.a)) {
			t.Fatal("borrowed")
		}
		n, err := c.Measure(a)
		if err != nil || n != len(fixture(t, tc.a)) {
			t.Fatal(n, err)
		}
		if b, err := c.DecodeBound(a.Header, n); err != nil || b < uint64(n) {
			t.Fatal(b, err)
		}
		// Labels such as mme/sgsn are not protocol evidence; only the declared interface changes semantics.
		e.Interface = sgd.MOInterfaceSGd
		if p := c.ValidateAnswerContext(r, a, src, moContext(t, e)); !p.Complete {
			t.Fatal(p)
		}
		for _, edit := range []func(*pdu.Message){func(a *pdu.Message) { a.Header.HopByHop++ }, func(a *pdu.Message) { a.Header.EndToEnd++ }, func(a *pdu.Message) { a.Body.(*sgd.OFA).SessionID += "wrong" }} {
			x, err := c.Snapshot(a)
			if err != nil {
				t.Fatal(err)
			}
			edit(&x)
			if p := c.ValidateAnswerContext(r, x, src, ctx); p.Complete {
				t.Fatal(p)
			}
		}
	}
	for _, name := range []string{"b29-ofa-vendor", "b29-ofa-dupresult", "b29-ofa-dualresult"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
}
func TestSGDMOOriginalIdentityAndSource(t *testing.T) {
	c := sgdCodec(t)
	r, a := decoded(t, c, "v24-sgd-mo-r"), decoded(t, c, "f24-sgd-mo-5001")
	e := moEvidence(r, sgd.MOInterfaceSGd)
	ctx := moContext(t, e)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	for _, edit := range []func(*sgd.MOForwardingEvidence){
		func(e *sgd.MOForwardingEvidence) { e.SenderIMSI = "001010123456780" },
		func(e *sgd.MOForwardingEvidence) { e.SenderMSISDN = avp.Some("123") },
		func(e *sgd.MOForwardingEvidence) { e.SenderMSISDN.Present = false },
		func(e *sgd.MOForwardingEvidence) { e.SCAddress = "124" },
		func(e *sgd.MOForwardingEvidence) { e.Requester.Host = "dra.example" },
		func(e *sgd.MOForwardingEvidence) { e.Responder.Host = "dra.example" },
	} {
		x := e
		edit(&x)
		if p := c.ValidateContext(r, moContext(t, x)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, name := range []string{"b29-mo-nomsisdn", "b29-mo-othermsisdn"} {
		x := decoded(t, c, name)
		if p := c.ValidateContext(x, ctx); p.Complete {
			t.Fatal(p)
		}
		xe := moEvidence(x, sgd.MOInterfaceSGd)
		p := c.ValidateAnswerContext(x, a, src, moContext(t, xe))
		if name == "b29-mo-nomsisdn" {
			if !errors.Is(p.Err(), validation.ErrMissing) {
				t.Fatal(p)
			}
		} else if !p.Complete {
			t.Fatal(p)
		}
		if name == "b29-mo-nomsisdn" {
			if err := c.ValidateAnswer(x, a, src); !errors.Is(err, validation.ErrMissing) {
				t.Fatal(err)
			}
			if p := c.ValidateAnswerContext(x, a, src, validation.Context{}); !errors.Is(p.Err(), validation.ErrMissing) {
				t.Fatal(p)
			}
		}
	}
	other := decoded(t, c, "b29-ofa-origin")
	if p := c.ValidateAnswerContext(r, other, src, ctx); !errors.Is(p.Err(), validation.ErrSource) {
		t.Fatal(p)
	}
	// Even an overly broad source list cannot override the frozen terminal evidence.
	src.Terminals = append(src.Terminals, pdu.Origin{Host: "dra.example", Realm: e.Responder.Realm})
	if p := c.ValidateAnswerContext(r, other, src, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	for _, edit := range []func(*sgd.MOForwardingEvidence){func(e *sgd.MOForwardingEvidence) { e.Interface = 0 }, func(e *sgd.MOForwardingEvidence) { e.Interface = 3 }, func(e *sgd.MOForwardingEvidence) { e.SCAddress = "+123" }, func(e *sgd.MOForwardingEvidence) { e.SenderMSISDN = avp.Some("") }, func(e *sgd.MOForwardingEvidence) { e.Requester.Host = strings.Repeat("x", 254) }} {
		x := e
		edit(&x)
		if _, err := sgd.MOForwarding(x); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	b := r.Body.(*sgd.OFR)
	b.DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	if reflect.TypeFor[validation.Context]().Size()+4*253+3*15+253 > 2048 {
		t.Fatal("context budget exceeded")
	}
}
func TestSGDMONamespaceAndT4Isolation(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "v24-sgd-mo-r")
	e := moEvidence(r, sgd.MOInterfaceSGd)
	ctx := moContext(t, e)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
	a := decoded(t, c, "f24-sgd-mo-5001")
	t4 := t4Context(t, t4Evidence(r))
	if p := c.ValidateAnswerContext(r, a, src, t4); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
	a.Body.(*sgd.OFA).ExternalIdentifier = avp.Some("device-001@example.net")
	if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrUnsupported) {
		t.Fatal(p)
	}
	a.Body.(*sgd.OFA).ExternalIdentifier.Present = false
	a.Body.(*sgd.OFA).SMDeliveryFailureCause = avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 3})
	if err := c.Validate(a); !errors.Is(err, validation.ErrField) {
		t.Fatal(err)
	}
	for _, iface := range []sgd.MOInterface{sgd.MOInterfaceSGd, sgd.MOInterfaceGdd} {
		e.Interface = iface
		ctx = moContext(t, e)
		for _, body := range []pdu.Body{&sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}, &sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5552})}, &pdu.ErrorAnswer{ResultCode: 3004}, &pdu.ErrorAnswer{ResultCode: 5001, FailedAVP: avp.Some(pdu.FailedAVP{Data: rawAVP(999999, nil)})}, &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5001)), FailedAVP: avp.Some(pdu.FailedAVP{Data: rawAVP(999999, nil)})}} {
			ans, err := c.CompleteAnswer(r, body, e.Responder)
			if err != nil {
				t.Fatal(err)
			}
			if p := c.ValidateAnswerContext(r, ans, src, ctx); !p.Complete {
				t.Fatal(p)
			}
		}
	}
	mt := decoded(t, c, "v24-sgd-mt-r")
	if p := c.ValidateContext(mt, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
}
