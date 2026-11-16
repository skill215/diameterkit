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
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestSMSFWECURSourcePairsAndOwnership(t *testing.T) {
	c := legacyCodec(t)
	for _, name := range []string{"init-request", "init-enhanced", "init-a-2001", "init-a-4012", "init-a-5012", "init-a-3004", "term-request", "term-enhanced", "term-a-2001", "term-a-5002", "term-a-5012", "term-a-3004", "term-usage-zero", "term-usage-max", "term-empty-status", "term-zero-discharge", "init-a-gsu", "term-sender-order"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, "b43-"+name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			// The raw TERM template consumes SMS_KEY before UDL; typed sender has one
			// fixed order. No source input was silently rewritten to obtain exactness.
			if !strings.HasPrefix(name, "term-") && !strings.HasSuffix(name, "3004") && !bytes.Equal(out, raw) {
				t.Fatal("initial/answer bytes changed")
			}
			if name == "term-sender-order" && !bytes.Equal(out, raw) {
				t.Fatal("sender order")
			}
			n, e := c.Measure(m)
			if e != nil || n != len(out) {
				t.Fatal(n, e)
			}
			bound, e := c.DecodeBound(m.Header, n)
			if e != nil || bound < uint64(n) {
				t.Fatal(bound, e)
			}
			again, e := c.Decode(out)
			if e != nil || !reflect.DeepEqual(again, m) {
				t.Fatal("semantic roundtrip", e)
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("input alias")
			}
			snap, e := c.Snapshot(m)
			if e != nil {
				t.Fatal(e)
			}
			if q, ok := m.Body.(*smsfwro.CCR); ok {
				q.SubscriptionID[0].SubscriptionIDData[0] = '9'
				if len(q.ServiceInformation.SMSInformation.SMStatus.Value) > 0 {
					q.ServiceInformation.SMSInformation.SMStatus.Value[0] = 99
				}
				q.MultipleServicesCreditControl.UsedServiceUnit.Value.CCServiceSpecificUnits = 77
			} else {
				kind := "init"
				if strings.HasPrefix(name, "term") {
					kind = "term"
				}
				r := decoded(t, c, "b43-"+kind+"-request")
				if e = c.ValidateAnswer(r, m, roSources()); e != nil {
					t.Fatal(e)
				}
				if report := c.ValidateAnswerContext(r, m, roSources(), validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot alias")
			}
		})
	}
	term := decoded(t, c, "b43-term-request")
	if !bytes.Equal(encoded(t, c, term), fixture(t, "b43-term-sender-order")) {
		t.Fatal("difference exceeds declared order swap")
	}
	q := term.Body.(*smsfwro.CCR)
	if q.RequestedAction.Present || q.MultipleServicesCreditControl.RequestedServiceUnit.Present || q.MultipleServicesCreditControl.ReportingReason.Value != 2 || !bytes.Equal(q.ServiceInformation.SMSInformation.SMStatus.Value, []byte{0, 255}) || q.ServiceInformation.SMSInformation.DischargeTime.Value != 3908988830 {
		t.Fatal(q)
	}
	if got := decoded(t, c, "b43-term-usage-max").Body.(*smsfwro.CCR).MultipleServicesCreditControl.UsedServiceUnit.Value.CCServiceSpecificUnits; got != math.MaxUint64 {
		t.Fatal(got)
	}
}
func TestSMSFWECURRejectedBranchesAndOriginalRequest(t *testing.T) {
	c := legacyCodec(t)
	for _, name := range []string{"init-action", "term-action", "init-wrong-sequence", "term-wrong-sequence", "init-wrong-destination-interface", "term-wrong-destination-interface", "init-wrong-origin-interface", "term-wrong-origin-interface", "init-origin-sccp", "term-origin-sccp", "init-wrong-imsi-copy", "term-wrong-imsi-copy", "init-wrong-time-copy", "term-wrong-time-copy", "term-width4", "term-no-usu", "term-empty-usu", "term-no-final", "term-wrong-final", "term-final-parent", "term-rsu", "init-usu", "init-no-rsu", "init-status", "term-no-status", "term-no-discharge", "term-duplicate-status", "term-token", "init-420", "term-a-gsu"} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b43-"+name)); e == nil {
				t.Fatal("unselected branch accepted")
			}
		})
	}
	r := decoded(t, c, "b43-term-request")
	a := decoded(t, c, "b43-term-a-wrong-sequence")
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	for _, p := range []string{"b43-init-request", "b43-term-request"} {
		if _, e := roCodec(t).Decode(fixture(t, p)); e == nil {
			t.Fatal("legacy accepted as standard")
		}
		if _, e := refundCodec(t).Decode(fixture(t, p)); e == nil {
			t.Fatal("legacy accepted as refund")
		}
	}
	// The pre-existing IEC branch must still require action and its own interface.
	iec := decoded(t, c, "b41-iec-request")
	iec.Body.(*smsfwro.CCR).RequestedAction = avp.Optional[uint32]{}
	if _, e := c.Append(nil, iec); e == nil {
		t.Fatal("IEC lost action requirement")
	}
}
func TestSMSFWECURContinuationBoundaries(t *testing.T) {
	c := legacyCodec(t)
	initial := decoded(t, c, "b43-init-enhanced").Body.(*smsfwro.CCR)
	term := decoded(t, c, "b43-term-enhanced").Body.(*smsfwro.CCR)
	if e := smsfwro.CheckECURContinuation(initial, term); e != nil {
		t.Fatal(e)
	}
	if e := smsfwro.CheckECURContinuation(nil, term); !errors.Is(e, validation.ErrType) {
		t.Fatal(e)
	}
	mutations := map[string]func(*smsfwro.CCR){
		"session": func(q *smsfwro.CCR) { q.SessionID += "other" }, "context": func(q *smsfwro.CCR) { q.ServiceContextID += "other" },
		"ctf": func(q *smsfwro.CCR) { q.OriginHost = "other.example" }, "ctf-realm": func(q *smsfwro.CCR) { q.OriginRealm = "other.example" },
		"ocs": func(q *smsfwro.CCR) { q.DestinationHost = avp.Some("other.example") }, "ocs-realm": func(q *smsfwro.CCR) { q.DestinationRealm = "other.example" },
		"realm-only": func(q *smsfwro.CCR) { q.DestinationHost = avp.Optional[string]{} }, "sequence": func(q *smsfwro.CCR) { q.CCRequestNumber = 2 },
		"msisdn": func(q *smsfwro.CCR) { q.SubscriptionID[0].SubscriptionIDData = []byte("other") },
		"imsi": func(q *smsfwro.CCR) {
			q.SubscriptionID[1].SubscriptionIDData = []byte("other")
			q.UserEquipmentInfo.UserEquipmentInfoValue = []byte("other")
		},
		"key": func(q *smsfwro.CCR) { q.ServiceInformation.MMSInformation.SMSKey = []byte("other") },
		"sender": func(q *smsfwro.CCR) {
			q.ServiceInformation.MMSInformation.OriginatorAddress.AddressData = []byte("other")
		},
		"recipient": func(q *smsfwro.CCR) {
			q.ServiceInformation.MMSInformation.RecipientAddress.AddressData = []byte("other")
		},
	}
	for name, fn := range mutations {
		t.Run(name, func(t *testing.T) {
			q := decoded(t, c, "b43-term-enhanced").Body.(*smsfwro.CCR)
			fn(q)
			if e := smsfwro.CheckECURContinuation(initial, q); e == nil {
				t.Fatal("mismatch accepted")
			}
		})
	}
	// A relation check is not grant, timing or delivery validation.
	term.OriginHost = strings.ToUpper(term.OriginHost)
	term.DestinationHost.Value = strings.ToUpper(term.DestinationHost.Value)
	term.MultipleServicesCreditControl.UsedServiceUnit.Value.CCServiceSpecificUnits = math.MaxUint64
	term.ServiceInformation.SMSInformation.DischargeTime = avp.Some(avp.Time(0))
	before := *term
	if e := smsfwro.CheckECURContinuation(initial, term); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, *term) {
		t.Fatal("helper mutation")
	}
}
func TestSMSFWECURBuildersAndCanonical(t *testing.T) {
	c := legacyCodec(t)
	status := []byte{0, 255}
	in := smsfwro.ECURInput{SessionID: "ctf.example;legacy-ecur;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-0001", Timestamp: 3908988800, DataCodingScheme: 8, UserDataLength: 140, StatusReportRequest: 1, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{5, 0, 3, 127, 2, 1, 0, 255}}
	for _, term := range []bool{false, true} {
		q := smsfwro.NewInitial(in)
		name := "init-request"
		h, e := uint32(0x31323334), uint32(0x41424344)
		if term {
			q = smsfwro.NewTermination(smsfwro.TerminationInput{ECURInput: in, UsedServiceSpecificUnits: 1, Status: status, DischargeTime: 3908988830})
			name = "term-sender-order"
			h, e = 0xb1b2b3b4, 0xc1c2c3c4
			clear(status)
		}
		m, err := c.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: h, EndToEnd: e})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(encoded(t, c, m), fixture(t, "b43-"+name)) {
			t.Fatal(name)
		}
	}
	var seen [][]byte
	for _, name := range []string{"term-request", "term-usage-zero", "term-usage-max", "term-empty-status", "term-zero-discharge", "term-sender-order"} {
		b, e := c.Canonical(fixture(t, "b43-"+name), 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, a := range seen {
			if bytes.Equal(a, b) {
				t.Fatal("lost raw usage/status/time/order", name)
			}
		}
		seen = append(seen, b)
	}
}
