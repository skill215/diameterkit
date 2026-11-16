package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"math"
	"testing"
)

func refundCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := legacy420.Profile()
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
func TestLegacy420RefundPairsAndOwnedValues(t *testing.T) {
	c := refundCodec(t)
	r := decoded(t, c, "b42-refund-request")
	for _, name := range []string{"refund-request", "refund-enhanced", "cca-success", "cca-denied", "cca-failed", "cca-error", "money-minimum", "money-maximum", "money-zero", "money-negative", "empty-token", "token-changed"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, "b42-"+name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if name != "cca-error" && !bytes.Equal(out, raw) {
				t.Fatal("source projection bytes changed")
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
			if q, ok := m.Body.(*legacy420.CCR); ok {
				q.SubscriptionID[0].SubscriptionIDData[0] = '9'
				if len(q.RefundInformation.Value) > 0 {
					q.RefundInformation.Value[0] = 55
				}
				q.MultipleServicesCreditControl.RequestedServiceUnit.Money.Value.UnitValue.ValueDigits = 99
			} else {
				if e = c.ValidateAnswer(r, m, roSources()); e != nil {
					t.Fatal(e)
				}
				if v := c.ValidateAnswerContext(r, m, roSources(), validation.Context{}); !v.Complete {
					t.Fatal(v)
				}
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot alias")
			}
		})
	}
	min := decoded(t, c, "b42-money-minimum").Body.(*legacy420.CCR).MultipleServicesCreditControl.RequestedServiceUnit
	max := decoded(t, c, "b42-money-maximum").Body.(*legacy420.CCR).MultipleServicesCreditControl.RequestedServiceUnit
	if min.Money.Value.UnitValue.ValueDigits != math.MinInt64 || min.Money.Value.UnitValue.Exponent != math.MinInt32 || min.Money.Value.CurrencyCode != math.MinInt32 || max.Money.Value.UnitValue.ValueDigits != math.MaxInt64 || max.CCServiceSpecificUnits != math.MaxUint64 {
		t.Fatal(min, max)
	}
}
func TestLegacy420RejectsTypesPathsAndAssociations(t *testing.T) {
	c := refundCodec(t)
	for _, name := range []string{"digits-width4", "exponent-width8", "currency-width8", "unit-width4", "420-scalar", "missing-money", "missing-unit-value", "missing-digits", "missing-exponent", "missing-currency", "missing-units", "missing-token", "duplicate-money", "duplicate-token", "413-alias", "money-vendor", "token-vendor", "2022-alias", "wrong-parent", "wrong-action", "wrong-number", "wrong-node", "cca-money", "cca-token"} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b42-"+name)); e == nil {
				t.Fatal("accepted unsupported form")
			}
		})
	}
	r := decoded(t, c, "b42-refund-request")
	a := decoded(t, c, "b42-cca-number")
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	a = decoded(t, c, "b42-cca-success")
	if e := c.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "dra.example", Realm: "example.net"}}}); e == nil {
		t.Fatal("wrong source")
	}
	a.Header.EndToEnd++
	if e := c.ValidateAnswer(r, a, roSources()); e == nil {
		t.Fatal("wrong exchange")
	}
	raw := fixture(t, "b42-refund-request")
	raw[4] &^= pdu.ProxiableFlag
	if _, e := c.Decode(raw); e == nil {
		t.Fatal("P0 deployment")
	}
	// Action cannot switch the fixed type interpreter even for malformed inputs.
	raw = fixture(t, "b42-420-scalar")
	for i := 20; i+12 <= len(raw); i++ {
		if bytes.Equal(raw[i:i+4], []byte{0, 0, 1, 180}) {
			clear(raw[i+8 : i+12])
			break
		}
	}
	if _, e := c.Decode(raw); e == nil {
		t.Fatal("action-selected scalar decoder")
	}
}
func TestLegacy420ProfileIsolationAndCanonical(t *testing.T) {
	lp, e := legacy420.Profile()
	if e != nil {
		t.Fatal(e)
	}
	ip, e := smsfwro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	rp, e := ro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []schema.Profile{ip, rp} {
		if _, e := schema.Compile(lp, p); e == nil {
			t.Fatal("ambiguous App4 binding")
		}
	}
	for _, c := range []*codec.Codec{legacyCodec(t), roCodec(t)} {
		if _, e := c.Decode(fixture(t, "b42-refund-request")); e == nil {
			t.Fatal("legacy money accepted by other profile")
		}
	}
	c := refundCodec(t)
	for _, n := range []string{"b41-iec-request", "v-iec-ccr-01"} {
		if _, e := c.Decode(fixture(t, n)); e == nil {
			t.Fatal("other profile bypass")
		}
	}
	var seen [][]byte
	for _, n := range []string{"refund-request", "money-negative", "money-minimum", "money-maximum", "money-zero", "empty-token", "token-changed"} {
		b, e := c.Canonical(fixture(t, "b42-"+n), 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, old := range seen {
			if bytes.Equal(b, old) {
				t.Fatal("refund content collapsed", n)
			}
		}
		seen = append(seen, b)
	}
	q := decoded(t, c, "b42-refund-request")
	q.Body.(*legacy420.CCR).MultipleServicesCreditControl.RequestedServiceUnit.Extra = []avp.Raw{{Key: avp.Key{Code: 413}, Data: []byte{1}}}
	if _, e := c.Append(nil, q); e == nil {
		t.Fatal("M0 alias", e)
	}
}
func TestLegacy420ExplicitBuilderAndNoFinancialDefaults(t *testing.T) {
	c := refundCodec(t)
	token := []byte{0, 255, 128, 0, 1}
	pid := []byte{0, 255}
	in := legacy420.RefundInput{SessionID: "ctf.example;legacy-refund;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", Timestamp: 3908988800, ValueDigits: 1, Exponent: 0, CurrencyCode: 840, ServiceSpecificUnits: 1, DataCodingScheme: 8, DestinationInterface: 2, OriginatorInterface: 1, ProtocolID: pid, RefundInformation: avp.Some(token)}
	q, e := legacy420.NewRefund(in)
	if e != nil {
		t.Fatal(e)
	}
	clear(token)
	clear(pid)
	m, e := c.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 0x31323334, EndToEnd: 0x41424344})
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(encoded(t, c, m), fixture(t, "b42-refund-request")) {
		t.Fatal("full source template differs")
	}
	if _, e = legacy420.NewRefund(legacy420.RefundInput{}); !errors.Is(e, validation.ErrMissing) {
		t.Fatal("missing token became empty", e)
	}
	zero, e := legacy420.NewRefund(legacy420.RefundInput{SessionID: "s", ServiceContextID: "explicit", RefundInformation: avp.Some([]byte{})})
	if e != nil {
		t.Fatal(e)
	}
	u := zero.MultipleServicesCreditControl.RequestedServiceUnit
	if u.Money.Value.UnitValue.ValueDigits != 0 || u.Money.Value.CurrencyCode != 0 || u.CCServiceSpecificUnits != 0 {
		t.Fatal("sample financial defaults")
	}
	m, e = c.CompleteRequest(zero, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 1, EndToEnd: 2})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.Decode(encoded(t, c, m)); e != nil {
		t.Fatal(e)
	}
}
