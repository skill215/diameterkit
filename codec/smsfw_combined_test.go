// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func combinedCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := combined.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(bp, p)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestCombinedSourcePairsAndOwnership(t *testing.T) {
	c := combinedCodec(t)
	for _, kind := range []string{"iec", "initial", "term", "refund", "general", "balance"} {
		t.Run(kind, func(t *testing.T) {
			raw := fixture(t, "b47-"+kind+"-request")
			r, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, r)
			if !bytes.Equal(raw, out) {
				t.Fatal("fixed sender projection differs")
			}
			n, e := c.Measure(r)
			if e != nil || n != len(out) {
				t.Fatal(n, e)
			}
			bound, e := c.DecodeBound(r.Header, n)
			if e != nil || bound < uint64(n) {
				t.Fatal(bound, e)
			}
			snap, e := c.Snapshot(r)
			if e != nil {
				t.Fatal(e)
			}
			clear(raw)
			q := r.Body.(*combined.CCR)
			q.SubscriptionID[0].SubscriptionIDData[0] = '9'
			q.ServiceInformation.SMSInformation.ClientAddressWire = avp.Some([]byte("mutated"))
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot/input alias")
			}
			for _, result := range []string{"success", "denied", "error"} {
				a := decoded(t, c, "b47-"+kind+"-"+result)
				if e = c.ValidateAnswer(snap, a, roSources()); e != nil {
					t.Fatal(result, e)
				}
				if report := c.ValidateAnswerContext(snap, a, roSources(), validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
				b := encoded(t, c, a)
				again, e := c.Decode(b)
				if e != nil {
					t.Fatal(e)
				}
				if e = c.ValidateAnswer(snap, again, roSources()); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
	// Decode both source orderings but never erase raw order from duplicate evidence.
	orig := fixture(t, "b43-term-request")
	sender := fixture(t, "b47-term-request")
	a := decoded(t, c, "b43-term-request")
	if !bytes.Equal(encoded(t, c, a), sender) {
		t.Fatal("TERM fixed order")
	}
	ca, e := c.Canonical(orig, 4096)
	if e != nil {
		t.Fatal(e)
	}
	cb, e := c.Canonical(sender, 4096)
	if e != nil || bytes.Equal(ca, cb) {
		t.Fatal("lost source order", e)
	}
	for _, name := range []string{"b41-iec-request", "b43-init-request", "b42-refund-request", "b44-general-request", "b44-balance-request"} {
		if _, e := c.Decode(fixture(t, name)); e != nil {
			t.Fatal(name, e)
		}
	}
}
func TestCombinedCrossTemplateRejections(t *testing.T) {
	c := combinedCodec(t)
	for _, name := range []string{"iec-address2018", "general-text2018", "general-draft2018", "general-smsc", "iec-money413", "refund-unit-only", "refund-client", "initial-service", "term-money", "general-ue", "general-mms-key", "iec-unknown-optional"} {
		if _, e := c.Decode(fixture(t, "b47-"+name)); e == nil {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"b41-string-nul", "b41-wrong-ue", "b41-wrong-time", "b41-wrong-smsc", "b41-wrong-node", "b41-wrong-action", "b41-wrong-units", "b41-missing-udh", "b41-420-scalar", "b41-420-grouped", "b41-token1292", "b41-subscriptions-reversed", "b42-missing-token", "b42-missing-money", "b42-413-alias", "b42-wrong-node", "b44-missing-money", "b44-missing-exponent", "b44-missing-timezone", "b44-missing-client", "b44-missing-mms", "b44-money420-alias", "b44-both-money", "b44-wrong-timezone-vendor", "b43-init-420", "b43-init-origin-sccp", "b43-init-no-rsu", "b43-term-no-usu", "b43-term-no-final", "b43-term-no-status", "b43-term-no-discharge", "b43-term-action"} {
		if _, e := c.Decode(fixture(t, name)); e == nil {
			t.Fatal("old constraint lost", name)
		}
	}
	// Static answers are a union; the mandatory original-request relation rejects
	// fields that belong only to a different template's answer subset.
	for _, tc := range [][2]string{{"b47-balance-request", "b44-cca-success"}, {"b47-general-request", "b44-balance-enough"}, {"b47-balance-request", "b44-balance-grant"}, {"b47-general-request", "b44-failed-cost"}} {
		r := decoded(t, c, tc[0])
		a := decoded(t, c, tc[1])
		a.Header = r.Header
		a.Header.Flags &^= pdu.RequestFlag
		a.Body.(*combined.CCA).SessionID = r.Body.(*combined.CCR).SessionID
		if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(tc, e)
		}
	}
	for _, old := range []func() (schema.Profile, error){smsfwro.Profile, legacy420.Profile} {
		a, _ := old()
		b, _ := combined.Profile()
		if _, e := schema.Compile(a, b); !errors.Is(e, schema.ErrDefinition) {
			t.Fatal("duplicate App4 binding", e)
		}
	}
}
func TestCombinedBuildersAndContinuation(t *testing.T) {
	c := combinedCodec(t)
	ec := combined.ECURInput{SessionID: "ctf.example;legacy-ecur;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, DataCodingScheme: 8, UserDataLength: 140, StatusReportRequest: 1, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{5, 0, 3, 127, 2, 1, 0, 255}}
	init := combined.NewInitial(ec)
	term := combined.NewTermination(combined.TerminationInput{ECURInput: ec, UsedServiceSpecificUnits: 1, Status: []byte{0, 255}, DischargeTime: 3908988830})
	for _, tc := range []struct {
		q    *combined.CCR
		name string
	}{{init, "initial"}, {term, "term"}} {
		raw := decoded(t, c, "b47-"+tc.name+"-request")
		r := raw.Body.(*combined.CCR)
		m, e := c.CompleteRequest(tc.q, pdu.RequestEnvelope{OriginHost: r.OriginHost, OriginRealm: r.OriginRealm, DestinationRealm: r.DestinationRealm, HopByHop: raw.Header.HopByHop, EndToEnd: raw.Header.EndToEnd})
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(encoded(t, c, m), fixture(t, "b47-"+tc.name+"-request")) {
			t.Fatal(tc.name, "builder")
		}
	}
	i := decoded(t, c, "b43-init-enhanced").Body.(*combined.CCR)
	q := decoded(t, c, "b43-term-enhanced").Body.(*combined.CCR)
	if e := combined.CheckECURContinuation(i, q); e != nil {
		t.Fatal(e)
	}
	q.ServiceInformation.MMSInformation.Value.SMSKey.Value = []byte("other")
	if e := combined.CheckECURContinuation(i, q); !errors.Is(e, validation.ErrContext) {
		t.Fatal(e)
	}
	// Owned adapters reuse typed source builders, never parse/guess ClientAddressWire.
	iv := combined.IECInput{SessionID: "s", ServiceContextID: "c", SubscriptionMSISDN: "m", SubscriptionIMSI: "i", SMSCAddress: "non-address-text", ProtocolID: []byte{0, 255}}
	ie := combined.NewIEC(iv)
	clear(iv.ProtocolID)
	if string(ie.ServiceInformation.SMSInformation.ClientAddressWire.Value) != "non-address-text" || ie.ServiceInformation.SMSInformation.SMProtocolID[1] != 255 {
		t.Fatal("adapter conversion/alias")
	}
	gv := combined.GeneralInput{SessionID: "s", ServiceContextID: "c", ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 9, ProtocolID: []byte{0, 255}}
	gd := combined.NewDirectDebit(gv)
	gb := combined.NewBalanceCheck(gv)
	clear(gv.ProtocolID)
	if gd.MultipleServicesCreditControl.RequestedServiceUnit.Value.StandardMoney.Value.UnitValue.ValueDigits != -123 || !bytes.Equal(gd.ServiceInformation.SMSInformation.ClientAddressWire.Value, []byte{0, 1, 127, 0, 0, 1}) || gb.RequestedAction.Value != 2 || gd.ServiceInformation.SMSInformation.SMProtocolID[1] != 255 {
		t.Fatal("general adapter")
	}
	if _, e := combined.NewRefund(combined.RefundInput{}); !errors.Is(e, validation.ErrMissing) {
		t.Fatal(e)
	}
	token := []byte{0, 255}
	rr, e := combined.NewRefund(combined.RefundInput{ValueDigits: -5, RefundInformation: avp.Some(token)})
	if e != nil {
		t.Fatal(e)
	}
	clear(token)
	if rr.RefundInformation.Value[1] != 255 || rr.MultipleServicesCreditControl.RequestedServiceUnit.Value.Money.Value.UnitValue.ValueDigits != -5 {
		t.Fatal("refund adapter")
	}
}

func TestCombinedRemainingBuildersExactProjection(t *testing.T) {
	c := combinedCodec(t)
	ie := combined.NewIEC(combined.IECInput{SessionID: "ctf.example;legacy-iec;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", OriginatorSCCPAddress: "15550000002", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, DataCodingScheme: 8, DestinationInterface: 2, UserDataLength: 140, StatusReportRequest: 1, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{5, 0, 3, 127, 2, 1, 0, 255}})
	gv := combined.GeneralInput{SessionID: "ctf.example;legacy-general;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", Timestamp: 3908988800, ValueDigits: 1, CurrencyCode: 840, ServiceSpecificUnits: 1, DataCodingScheme: 8, DestinationInterface: 2, OriginatorInterface: 1, ProtocolID: []byte{0, 255}, OriginatorMSISDN: "15551234567"}
	// Token bytes come from the frozen explicit source input; no CCA-to-token inference.
	token := decoded(t, c, "b42-refund-request").Body.(*combined.CCR).RefundInformation
	rf, e := combined.NewRefund(combined.RefundInput{SessionID: "ctf.example;legacy-refund;1", ServiceContextID: gv.ServiceContextID, SubscriptionMSISDN: gv.SubscriptionMSISDN, SubscriptionIMSI: gv.SubscriptionIMSI, Timestamp: gv.Timestamp, ValueDigits: 1, CurrencyCode: 840, ServiceSpecificUnits: 1, DataCodingScheme: 8, DestinationInterface: 2, OriginatorInterface: 1, ProtocolID: []byte{0, 255}, RefundInformation: token})
	if e != nil {
		t.Fatal(e)
	}
	for name, q := range map[string]*combined.CCR{"iec": ie, "general": combined.NewDirectDebit(gv), "balance": combined.NewBalanceCheck(gv), "refund": rf} {
		want := decoded(t, c, "b47-"+name+"-request")
		r := want.Body.(*combined.CCR)
		q.SessionID = r.SessionID // balance fixture uses its own caller-supplied identity.
		m, e := c.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: r.OriginHost, OriginRealm: r.OriginRealm, DestinationRealm: r.DestinationRealm, HopByHop: want.Header.HopByHop, EndToEnd: want.Header.EndToEnd})
		if e != nil {
			t.Fatal(name, e)
		}
		if !bytes.Equal(encoded(t, c, m), fixture(t, "b47-"+name+"-request")) {
			t.Fatal(name, "source projection differs")
		}
	}
}
