package codec_test

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func TestSMSFWGeneralPairsAndAmounts(t *testing.T) {
	c := refundCodec(t)
	debit := decoded(t, c, "b44-general-request")
	balance := decoded(t, c, "b44-balance-request")
	for _, name := range []string{"general-request", "general-enhanced", "balance-request", "cca-success", "cca-denied", "cca-failed", "cca-error", "cca-cost-units", "cca-cost-absent-exponent", "cca-cost-zero", "cca-cost-limits", "cca-cost-unit", "balance-enough", "balance-no-credit", "money-negative", "currency-negative", "raw-originator"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, "b44-"+name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if name != "cca-error" && !bytes.Equal(out, raw) {
				t.Fatal("source bytes differ")
			}
			n, e := c.Measure(m)
			if e != nil || n != len(out) {
				t.Fatal(n, e)
			}
			bound, e := c.DecodeBound(m.Header, n)
			if e != nil || bound < uint64(n) {
				t.Fatal(bound, e)
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("input borrowed")
			}
			snap, e := c.Snapshot(m)
			if e != nil {
				t.Fatal(e)
			}
			switch b := m.Body.(type) {
			case *legacy420.CCR:
				b.ServiceInformation.SMSInformation.ClientAddress.Value.Data[0] = 1
				b.ServiceInformation.MMSInformation.Value.OriginatorAddress.AddressData[0] = 0
				b.MultipleServicesCreditControl.RequestedServiceUnit.StandardMoney.Value.UnitValue.ValueDigits = 88
			default:
				r := debit
				if name == "balance-enough" || name == "balance-no-credit" {
					r = balance
				}
				if e = c.ValidateAnswer(r, m, roSources()); e != nil {
					t.Fatal(e)
				}
				if v := c.ValidateAnswerContext(r, m, roSources(), validation.Context{}); !v.Complete {
					t.Fatal(v)
				}
				if a, ok := m.Body.(*legacy420.CCA); ok {
					a.CostInformation.Value.UnitValue.ValueDigits = 999
					a.CheckBalanceResult = avp.Some(uint32(99))
				}
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot aliases")
			}
		})
	}
	a := decoded(t, c, "b44-cca-cost-limits").Body.(*legacy420.CCA)
	if a.CostInformation.Value.UnitValue.ValueDigits != math.MinInt64 || a.CostInformation.Value.UnitValue.Exponent.Value != math.MinInt32 || a.CostInformation.Value.CurrencyCode != math.MaxUint32 || a.GrantedServiceUnit.Value.CCServiceSpecificUnits != math.MaxUint64 {
		t.Fatal(a)
	}
	z := decoded(t, c, "b44-cca-cost-zero").Body.(*legacy420.CCA)
	absent := decoded(t, c, "b44-cca-cost-absent-exponent").Body.(*legacy420.CCA)
	if !z.GrantedServiceUnit.Present || !z.CostInformation.Value.UnitValue.Exponent.Present || absent.CostInformation.Value.UnitValue.Exponent.Present {
		t.Fatal("zero/absence collapsed")
	}
	// Refunds can now carry the standard top-level refunded-unit result too.
	r := decoded(t, c, "b42-refund-request")
	a2 := decoded(t, c, "b42-cca-gsu")
	if e := c.ValidateAnswer(r, a2, roSources()); e != nil {
		t.Fatal(e)
	}
}

func TestSMSFWGeneralRejectsShapesAndOriginalAction(t *testing.T) {
	c := refundCodec(t)
	for _, name := range []string{"balance-unknown", "failed-cost", "failed-balance", "failed-grant", "nul-originator", "digits-width4", "units-width4", "draft-address", "string-address", "ipv6-address", "timezone-nonzero", "service-nonliteral", "refund-action", "price-action", "missing-money", "missing-exponent", "missing-currency", "missing-timezone", "missing-service", "missing-client", "missing-mms", "wrong-timezone-vendor", "money420-alias", "both-money", "duplicate-money", "unexpected-token", "cost-missing-unit-value", "cost-missing-digits", "cost-missing-currency", "cost-missing-units", "cost-bad-digits", "cost-bad-exponent", "cost-bad-currency", "duplicate-cost", "cost-wrong-parent", "unknown-cost-child"} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b44-"+name)); e == nil {
				t.Fatal("unsupported shape accepted")
			}
		})
	}
	for _, tc := range [][2]string{{"balance-request", "cca-success"}, {"balance-request", "balance-grant"}, {"general-request", "balance-enough"}, {"general-request", "cca-number"}} {
		r := decoded(t, c, "b44-"+tc[0])
		a := decoded(t, c, "b44-"+tc[1])
		if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(tc, e)
		}
	}
	for _, name := range []string{"cca-denied", "cca-error"} {
		if e := c.ValidateAnswer(decoded(t, c, "b44-balance-request"), decoded(t, c, "b44-"+name), roSources()); e != nil {
			t.Fatal(e)
		}
	}
	// Other bindings cannot interpret this template by inspecting Action or bytes.
	for _, co := range []interface {
		Decode([]byte) (pdu.Message, error)
	}{legacyCodec(t), roCodec(t)} {
		if _, e := co.Decode(fixture(t, "b44-general-request")); e == nil {
			t.Fatal("profile isolation")
		}
	}
}

func TestSMSFWGeneralBuilderAndCanonical(t *testing.T) {
	c := refundCodec(t)
	pid := []byte{0, 255}
	in := legacy420.GeneralInput{SessionID: "ctf.example;legacy-general;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", Timestamp: 3908988800, ValueDigits: 1, CurrencyCode: 840, ServiceSpecificUnits: 1, DataCodingScheme: 8, DestinationInterface: 2, OriginatorInterface: 1, ProtocolID: pid, OriginatorMSISDN: "15551234567"}
	for _, tc := range []struct {
		name string
		r    *legacy420.CCR
	}{{"general-request", legacy420.NewDirectDebit(in)}, {"balance-request", legacy420.NewBalanceCheck(in)}} {
		clear(pid)
		m, e := c.CompleteRequest(tc.r, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 0x31323334, EndToEnd: 0x41424344})
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(encoded(t, c, m), fixture(t, "b44-"+tc.name)) {
			t.Fatal("builder not source projection", tc.name)
		}
	}
	zero := legacy420.NewDirectDebit(legacy420.GeneralInput{})
	u := zero.MultipleServicesCreditControl.RequestedServiceUnit
	if u.StandardMoney.Value.UnitValue.ValueDigits != 0 || u.StandardMoney.Value.CurrencyCode != 0 || u.CCServiceSpecificUnits != 0 {
		t.Fatal("sample financial defaults")
	}
	var seen [][]byte
	for _, name := range []string{"general-request", "balance-request", "money-negative", "currency-negative", "raw-originator"} {
		b, e := c.Canonical(fixture(t, "b44-"+name), 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, old := range seen {
			if bytes.Equal(old, b) {
				t.Fatal("content collapsed", name)
			}
		}
		seen = append(seen, b)
	}
}
