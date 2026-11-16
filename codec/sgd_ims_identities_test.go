package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"reflect"
	"strings"
	"testing"
)

func TestSGDIMSMOKnownSourcePairsAndOwnership(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"b37-imsi-only", "b37-imsi-real", "b37-imsi5", "b37-msisdn15", "b37-imsi-equals-hssid"} {
		t.Run(name, func(t *testing.T) {
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
			for _, aName := range []string{"v21-ims-ofr-a", "v21-ims-ofr-f", "b37-ofa-5552"} {
				a := decoded(t, c, aName)
				for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
					if !p.Complete {
						t.Fatal(aName, p)
					}
				}
				a.Header.EndToEnd++
				if p := c.ValidateAnswerContext(r, a, src, ctx); !errors.Is(p.Err(), validation.ErrUnmatchedAnswer) {
					t.Fatal(p)
				}
			}
			for _, code := range []uint32{3004, 5012} {
				var body pdu.Body = &pdu.ErrorAnswer{ResultCode: code}
				o := src.ErrorAgents[0]
				if code == 5012 {
					body = &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(code)}
					o = e.Responder
				}
				a, err := c.CompleteAnswer(r, body, o)
				if err != nil {
					t.Fatal(err)
				}
				if p := c.ValidateAnswerContext(r, a, src, ctx); !p.Complete {
					t.Fatal(p)
				}
			}
			for _, aName := range []string{"b37-ofa-5001", "b37-ofa-external"} {
				a := decoded(t, c, aName)
				for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
					if p.Complete {
						t.Fatal("real identity bypass", p)
					}
					want := validation.ErrUnsupported
					if aName == "b37-ofa-5001" && !e.SenderMSISDN.Present {
						want = validation.ErrMissing
					}
					if !errors.Is(p.Err(), want) {
						t.Fatal(p)
					}
				}
			}
			raw := fixture(t, name)
			owned, err := c.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			clear(raw)
			if !bytes.Equal(encoded(t, c, r), encoded(t, c, owned)) {
				t.Fatal("borrowed identity")
			}
			n, err := c.Measure(r)
			if err != nil || n != len(encoded(t, c, r)) {
				t.Fatal(n, err)
			}
			if bound, err := c.DecodeBound(r.Header, n); err != nil || bound < uint64(n) {
				t.Fatal(bound, err)
			}
			canonical, err := c.Canonical(fixture(t, name), 8192)
			if err != nil {
				t.Fatal(err)
			}
			round, err := c.Canonical(encoded(t, c, r), 8192)
			if err != nil || !bytes.Equal(canonical, round) {
				t.Fatal(err)
			}
		})
	}
}
func TestSGDIMSMOKnownSourceEvidence(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "b37-imsi-real")
	e := imsMOEvidence(r)
	ctx := imsMOContext(t, e)
	// Reusing inactive source storage keeps every older Context budget intact.
	if reflect.TypeFor[validation.Context]().Size() > 704 {
		t.Fatal("Context grew")
	}
	if reflect.TypeFor[validation.Context]().Size()+4*253+4*15 > 2048 || reflect.TypeFor[validation.Context]().Size()+4*253+3*15+256 > 2048 {
		t.Fatal("context budget")
	}
	for _, edit := range []func(*sgd.IMSMOForwardingEvidence){
		func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI = avp.Some("001010123456780") },
		func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Optional[string]{} },
		func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Some("12025550123") },
		func(e *sgd.IMSMOForwardingEvidence) {
			e.SenderIMSI = avp.Optional[string]{}
			e.SenderMSISDN = avp.Optional[string]{}
		},
	} {
		x := e
		edit(&x)
		if p := c.ValidateContext(r, imsMOContext(t, x)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	for _, edit := range []func(*sgd.IMSMOForwardingEvidence){
		func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI = avp.Some("") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI = avp.Some("0010") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI = avp.Some(strings.Repeat("1", 16)) }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI = avp.Some("00101x") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderIMSI.Present = false }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Some("") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Some("+123") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Some("000000000000000") }, func(e *sgd.IMSMOForwardingEvidence) { e.SenderMSISDN = avp.Some(strings.Repeat("1", 16)) },
	} {
		x := e
		edit(&x)
		if _, err := sgd.IMSMOForwarding(x); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	e.SenderIMSI.Value = "001010123456780"
	e.SenderMSISDN.Value = "123"
	if p := c.ValidateContext(r, ctx); !p.Complete {
		t.Fatal("evidence not owned", p)
	}
	dummy := decoded(t, c, "v21-ims-ofr-r")
	x := imsMOEvidence(dummy)
	x.SenderIMSI.Value = strings.Repeat("x", 4096)
	x.SenderMSISDN.Value = strings.Repeat("y", 4096)
	v := imsMOContext(t, x)
	if p := c.ValidateContext(dummy, v); !p.Complete {
		t.Fatal(p)
	}
	si, sn, mp, ok := v.IMSMOSender()
	if !ok || si != "" || sn != "" || mp {
		t.Fatal("retained absent values")
	}
	only := decoded(t, c, "b37-imsi-only")
	x = imsMOEvidence(only)
	x.SenderMSISDN.Value = strings.Repeat("x", 4096)
	if p := c.ValidateContext(only, imsMOContext(t, x)); !p.Complete {
		t.Fatal(p)
	}
	if p := c.ValidateContext(only, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal("presence mismatch", p)
	}
	if p := c.ValidateContext(dummy, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal("known identities authorized dummy", p)
	}
	x = imsMOEvidence(r)
	ordinary := moContext(t, sgd.MOForwardingEvidence{Interface: sgd.MOInterfaceSGd, Requester: x.Requester, Responder: x.Responder, SenderIMSI: x.SenderIMSI.Value, SenderMSISDN: x.SenderMSISDN, SCAddress: x.SCAddress})
	if p := c.ValidateContext(r, ordinary); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	if _, ok := ctx.T4Charging(); ok {
		t.Fatal("source storage exposed T4 purpose")
	}
	if _, ok := ctx.MOForwarding(); ok {
		t.Fatal("source storage exposed ordinary MO purpose")
	}
}
func TestSGDIMSMOKnownSourceWireAndCanonical(t *testing.T) {
	c := sgdCodec(t)
	for _, name := range []string{"b37-imsi4", "b37-imsi16", "b37-imsi-utf8", "b37-imsi-empty", "b37-msisdn16", "b37-msisdn-bad", "b37-msisdn-empty", "b37-imsi-dummy", "b37-empty-group", "b37-real-noimsi", "b37-dup-imsi", "b37-dup-msisdn", "b37-wrong-vendor-imsi"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	r := decoded(t, c, "b37-imsi-real")
	base, err := c.Canonical(encoded(t, c, r), 8192)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*sgd.OFR){func(b *sgd.OFR) { b.UserIdentifier.UserName.Value = "001010123456780" }, func(b *sgd.OFR) { b.UserIdentifier.MSISDN.Present = false }, func(b *sgd.OFR) { b.UserIdentifier.MSISDN.Value[0] = 0x10 }} {
		m, err := c.Snapshot(r)
		if err != nil {
			t.Fatal(err)
		}
		edit(m.Body.(*sgd.OFR))
		x, err := c.Canonical(encoded(t, c, m), 8192)
		if err != nil || bytes.Equal(x, base) {
			t.Fatal("source collapsed", err)
		}
	}
	// HSS-ID and source IMSI may happen to share digits; semantics follow the parent path.
	same := decoded(t, c, "b37-imsi-equals-hssid")
	if p := c.ValidateContext(same, imsMOContext(t, imsMOEvidence(same))); !p.Complete {
		t.Fatal(p)
	}
}
