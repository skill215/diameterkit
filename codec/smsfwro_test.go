// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"math"
	"testing"
)

func legacyCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := smsfwro.Profile()
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
func TestSMSFWIECPairsAndTemplateBytes(t *testing.T) {
	c := legacyCodec(t)
	r := decoded(t, c, "b41-iec-request")
	for _, n := range []string{"iec-request", "iec-enhanced", "signed-min", "signed-max", "string-nonutf8", "string-empty", "cca-success", "cca-denied", "cca-failed", "cca-error", "cca-grant-ddfh", "cca-ddfh-zero"} {
		t.Run(n, func(t *testing.T) {
			raw := fixture(t, "b41-"+n)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if n != "cca-error" && !bytes.Equal(out, raw) {
				t.Fatal("independent source bytes changed")
			}
			size, e := c.Measure(m)
			if e != nil || size != len(raw) {
				t.Fatal(size, e)
			}
			bound, e := c.DecodeBound(m.Header, size)
			if e != nil || bound < uint64(size) {
				t.Fatal(bound, e)
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("decode aliases input")
			}
			if m.Header.Flags&pdu.RequestFlag == 0 {
				if e = c.ValidateAnswer(r, m, roSources()); e != nil {
					t.Fatal(e)
				}
				report := c.ValidateAnswerContext(r, m, roSources(), validation.Context{})
				if !report.Complete {
					t.Fatal(report)
				}
			}
		})
	}
	q := r.Body.(*smsfwro.CCR)
	if len(q.SubscriptionID) != 2 || q.SubscriptionID[0].SubscriptionIDType != 0 || q.SubscriptionID[1].SubscriptionIDType != 1 || q.ServiceInformation.SMSInformation.SMSNode != 0 || q.MultipleServicesCreditControl.RequestedServiceUnit.Value.CCServiceSpecificUnits != 1 {
		t.Fatal(q)
	}
	if q.EventTimestamp != 3908988800 || string(q.UserEquipmentInfo.UserEquipmentInfoValue) != "001010123456789" || !bytes.Equal(q.ServiceInformation.SMSInformation.SMUserDataHeader, []byte{5, 0, 3, 127, 2, 1, 0, 255}) {
		t.Fatal("template mapping")
	}
	if decoded(t, c, "b41-signed-min").Body.(*smsfwro.CCR).ServiceInformation.SMSInformation.SMDataCodingScheme != math.MinInt32 {
		t.Fatal("signed truncation")
	}
	a := decoded(t, c, "b41-cca-grant-ddfh").Body.(*smsfwro.CCA)
	if a.GrantedServiceUnit.Value.CCServiceSpecificUnits != math.MaxUint64 || a.DirectDebitingFailureHandling.Value != 1 {
		t.Fatal(a)
	}
}
func TestSMSFWIECRejectsUnselectedParentsAndRelations(t *testing.T) {
	c := legacyCodec(t)
	for _, n := range []string{"signed-width", "string-nul", "wrong-ue", "wrong-time", "wrong-smsc", "wrong-node", "wrong-action", "wrong-units", "missing-udh", "duplicate-interface", "wrong-vendor", "wrong-parent", "420-scalar", "420-grouped", "token1292", "cca-psp", "cca-ddfh-unknown", "subscriptions-reversed"} {
		t.Run(n, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b41-"+n)); e == nil {
				t.Fatal("accepted unselected template")
			}
		})
	}
	raw := fixture(t, "b41-iec-request")
	raw[4] &^= pdu.ProxiableFlag
	if _, e := c.Decode(raw); e == nil {
		t.Fatal("unselected legacy P0 deployment")
	}

	r := decoded(t, c, "b41-iec-request")
	a := decoded(t, c, "b41-cca-number")
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	a = decoded(t, c, "b41-cca-success")
	if e := c.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "dra.example", Realm: "example.net"}}}); e == nil {
		t.Fatal("adjacent DRA grants terminal authority")
	}
	a.Header.EndToEnd++
	if e := c.ValidateAnswer(r, a, roSources()); e == nil {
		t.Fatal("wrong exchange")
	}
}
func TestSMSFWIECIsolationOwnershipAndCanonical(t *testing.T) {
	c := legacyCodec(t)
	rp, e := ro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	lp, e := smsfwro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	if _, e := schema.Compile(rp, lp); e == nil {
		t.Fatal("two App4 dialects share bundle")
	}
	if _, e := roCodec(t).Decode(fixture(t, "b41-iec-request")); e == nil {
		t.Fatal("legacy interpreted as standard")
	}
	if _, e := c.Decode(fixture(t, "v-iec-ccr-01")); e == nil {
		t.Fatal("standard interpreted as legacy")
	}
	r := decoded(t, c, "b41-iec-request")
	out := encoded(t, c, r)
	snap, e := c.Snapshot(r)
	if e != nil {
		t.Fatal(e)
	}
	q := r.Body.(*smsfwro.CCR)
	q.SubscriptionID[1].SubscriptionIDData[0] = '9'
	q.ServiceInformation.SMSInformation.SMUserDataHeader[0] = 99
	if !bytes.Equal(out, encoded(t, c, snap)) {
		t.Fatal("snapshot alias")
	}
	old, e := c.Canonical(out, 4096)
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"signed-min", "signed-max", "string-empty", "string-nonutf8"} {
		v, e := c.Canonical(fixture(t, "b41-"+n), 4096)
		if e != nil || bytes.Equal(old, v) {
			t.Fatal(n, e)
		}
	}
	// Fixed dialect rejects unknown optional data, rather than guessing a private
	// Vendor/parent/type. Generic E answers remain independent from CCA fields.
	q = snap.Body.(*smsfwro.CCR)
	q.Extra = []avp.Raw{{Key: avp.Key{VendorID: 4242, Code: 1292}, Flags: 128, Data: []byte{1}}}
	if _, e = c.Append(nil, snap); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
func TestSMSFWIECBuilderMatchesIndependentTemplate(t *testing.T) {
	c := legacyCodec(t)
	pid := []byte{0, 255}
	udh := []byte{5, 0, 3, 127, 2, 1, 0, 255}
	q := smsfwro.NewIEC(smsfwro.IECInput{SessionID: "ctf.example;legacy-iec;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", OriginatorSCCPAddress: "15550000002", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, DataCodingScheme: 8, DestinationInterface: 2, UserDataLength: 140, StatusReportRequest: 1, ProtocolID: pid, UserDataHeader: udh})
	clear(pid)
	clear(udh)
	m, e := c.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 0x31323334, EndToEnd: 0x41424344})
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(encoded(t, c, m), fixture(t, "b41-iec-request")) {
		t.Fatal("builder differs from full legacy template")
	}
	// No caller-supplied field silently defaults to an identity, timestamp or fee.
	empty := smsfwro.NewIEC(smsfwro.IECInput{SessionID: "s", ServiceContextID: "explicit"})
	m, e = c.CompleteRequest(empty, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 1, EndToEnd: 2})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.Decode(encoded(t, c, m)); e != nil {
		t.Fatal(e)
	}
	if empty.EventTimestamp != 0 || len(empty.SubscriptionID[1].SubscriptionIDData) != 0 || empty.ServiceInformation.MMSInformation.SubmissionTime != 0 {
		t.Fatal("inferred facts")
	}
}
