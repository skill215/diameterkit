package codec_test

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/app/ro/dcca"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func dccaCodec(t *testing.T) *codec.Codec {
	t.Helper()
	p, e := dcca.Profile()
	if e != nil {
		t.Fatal(e)
	}
	a, e := abort.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(p, a)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func dccaSources() pdu.AnswerSources {
	return pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "ocs.example", Realm: "example"}}}
}
func TestDCCAIndependentPairsAndOwnedUnits(t *testing.T) {
	c := dccaCodec(t)
	for _, action := range []string{"balance", "debit", "refund"} {
		for _, suffix := range []string{"", "-zero", "-max", "-money", "-money-min", "-money-max"} {
			name := "r-" + action + suffix
			t.Run(name, func(t *testing.T) {
				raw := fixture(t, "b70-"+name)
				m, e := c.Decode(raw)
				if e != nil {
					t.Fatal(e)
				}
				out := encoded(t, c, m)
				if !bytes.Equal(out, raw) {
					t.Fatal("independent bytes changed")
				}
				clear(raw)
				if !bytes.Equal(out, encoded(t, c, m)) {
					t.Fatal("aliased input")
				}
				snap, e := c.Snapshot(m)
				if e != nil {
					t.Fatal(e)
				}
				r := m.Body.(*dcca.CCR)
				r.SubscriptionID[0].SubscriptionIDData = "9"
				if !bytes.Equal(out, encoded(t, c, snap)) {
					t.Fatal("snapshot alias")
				}
				v := snap.Body.(*dcca.CCR).RequestedServiceUnit
				if suffix == "-zero" && (!v.Present || !v.Value.CCServiceSpecificUnits.Present || v.Value.CCServiceSpecificUnits.Value != 0) {
					t.Fatal("lost zero")
				}
				if suffix == "-max" && v.Value.CCServiceSpecificUnits.Value != math.MaxUint64 {
					t.Fatal("truncated uint64")
				}
				if suffix == "-money-min" && (v.Value.CCMoney.Value.UnitValue.ValueDigits != math.MinInt64 || v.Value.CCMoney.Value.UnitValue.Exponent.Value != math.MinInt32 || v.Value.CCMoney.Value.CurrencyCode != math.MaxUint32) {
					t.Fatal("money changed")
				}
				n, e := c.Measure(snap)
				if e != nil || n != len(out) {
					t.Fatal(n, e)
				}
				bound, e := c.DecodeBound(snap.Header, n)
				if e != nil || bound < uint64(n) {
					t.Fatal(bound, e)
				}
				if report := c.ValidateContext(snap, validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
			})
		}
	}
	for _, action := range []string{"balance", "debit", "refund"} {
		r := decoded(t, c, "b70-r-"+action)
		names := []string{"a-4010", "a-4011", "a-4012", "a-5011", "a-5012", "a-5030", "a-5031", "a-e"}
		switch action {
		case "balance":
			names = append(names, "a-balance-enough", "a-balance-no")
		case "debit":
			names = append(names, "a-debit", "a-debit-zero", "a-debit-money")
		case "refund":
			names = append(names, "a-refund", "a-refund-cost")
		}
		for _, name := range names {
			t.Run(action+"/"+name, func(t *testing.T) {
				a := decoded(t, c, "b70-"+name)
				if e := c.ValidateAnswer(r, a, dccaSources()); e != nil {
					t.Fatal(e)
				}
				if report := c.ValidateAnswerContext(r, a, dccaSources(), validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
				if !bytes.Equal(encoded(t, c, a), fixture(t, "b70-"+name)) {
					t.Fatal("answer changed")
				}
			})
		}
	}
}
func TestDCCABalanceMissingUnknownAndOriginalAction(t *testing.T) {
	c := dccaCodec(t)
	r := decoded(t, c, "b70-r-balance")
	for _, name := range []string{"a-balance-missing", "a-balance-unknown"} {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, "b70-"+name)
			if e := c.ValidateAnswer(r, a, dccaSources()); e != nil {
				t.Fatal("wire association should retain answer", e)
			}
			report := c.ValidateAnswerContext(r, a, dccaSources(), validation.Context{})
			if report.Complete {
				t.Fatal("invented balance", report)
			}
			v := a.Body.(*dcca.CCA).CheckBalanceResult
			if name == "a-balance-missing" && (v.Present || !reflect.DeepEqual(report.Violations, []error{validation.ErrMissing})) {
				t.Fatal(report, v)
			}
			if name == "a-balance-unknown" && (!v.Present || v.Value != math.MaxUint32 || !reflect.DeepEqual(report.LocalErrors, []error{validation.ErrUnsupported})) {
				t.Fatal(report, v)
			}
		})
	}
	for _, action := range []string{"debit", "refund"} {
		r := decoded(t, c, "b70-r-"+action)
		if e := c.ValidateAnswer(r, decoded(t, c, "b70-a-balance-enough"), dccaSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(e)
		}
	}
	for _, name := range []string{"a-number", "a-type", "a-debit"} {
		if e := c.ValidateAnswer(r, decoded(t, c, "b70-"+name), dccaSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(name, e)
		}
	}
	// The same 2001 without GSU is incomplete for debit but complete for refund.
	a := decoded(t, c, "b70-a-refund")
	if c.ValidateAnswerContext(decoded(t, c, "b70-r-debit"), a, dccaSources(), validation.Context{}).Complete {
		t.Fatal("debit grant inferred")
	}
	if !c.ValidateAnswerContext(decoded(t, c, "b70-r-refund"), a, dccaSources(), validation.Context{}).Complete {
		t.Fatal("refund incorrectly requires GSU")
	}
	if e := c.ValidateBody(&dcca.CCA{ResultCode: 2001, AuthApplicationID: 4, CCRequestType: 4, CheckBalanceResult: avp.Some(uint32(99))}); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
func TestDCCACounterexamplesAndBindingIsolation(t *testing.T) {
	c := dccaCodec(t)
	for _, f := range []string{"app", "context", "type", "number", "subscription", "action", "service"} {
		for _, kind := range []string{"missing", "duplicate"} {
			name := "r-" + kind + "-" + f
			t.Run(name, func(t *testing.T) {
				want := validation.ErrMissing
				if kind == "duplicate" {
					want = validation.ErrDuplicate
				}
				if _, e := c.Decode(fixture(t, "b70-"+name)); !errors.Is(e, want) {
					t.Fatal(e, want)
				}
			})
		}
	}
	for name, want := range map[string]error{"r-wrong-app": validation.ErrField, "r-wrong-context-format": validation.ErrField, "r-wrong-sms": validation.ErrUnsupported, "r-wrong-initial": validation.ErrUnsupported, "r-wrong-number": validation.ErrUnsupported, "r-wrong-price": validation.ErrUnsupported, "r-wrong-bad-sub": validation.ErrField, "r-duplicate-sub-type": validation.ErrDuplicate, "r-unit-empty": validation.ErrUnsupported, "r-unit-mixed": validation.ErrUnsupported, "r-unit-width": avp.ErrValue, "r-unit-legacy420": validation.ErrUnsupported, "r-mscc": validation.ErrUnsupported, "r-token2022": validation.ErrUnsupported, "r-token1292": validation.ErrUnsupported, "r-unknown-mandatory": validation.ErrUnsupported, "a-duplicate-balance": validation.ErrDuplicate, "a-5031-missing": validation.ErrMissing, "a-experimental": validation.ErrUnsupported} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b70-"+name)); !errors.Is(e, want) {
				t.Fatal(e, want)
			}
		})
	}
	raw := fixture(t, "b70-r-unknown-optional")
	m, e := c.Decode(raw)
	if e != nil || !bytes.Equal(raw, encoded(t, c, m)) {
		t.Fatal(e)
	}
	p, e := dcca.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := ro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = schema.Compile(p, sp); e == nil {
		t.Fatal("ambiguous CCR binding")
	}
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	op, ok := b.Lookup((&dcca.CCR{}).CommandKey())
	if !ok || op.TransportReplay {
		t.Fatal("unreviewed replay")
	}
}
func TestDCCARefundRelation(t *testing.T) {
	c := dccaCodec(t)
	original := decoded(t, c, "b70-r-debit").Body.(*dcca.CCR)
	refund := decoded(t, c, "b70-r-refund").Body.(*dcca.CCR)
	if e := dcca.CheckRefundRelation(original, refund); !errors.Is(e, validation.ErrContext) {
		t.Fatal("reused Session", e)
	}
	refund.SessionID += ";refund"
	if e := dcca.CheckRefundRelation(original, refund); e != nil {
		t.Fatal(e)
	}
	// Changing amount is not decided here: eligibility and amount belong to the application.
	refund.RequestedServiceUnit = avp.Some(dcca.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(0))})
	if e := dcca.CheckRefundRelation(original, refund); e != nil {
		t.Fatal(e)
	}
	for name, change := range map[string]func(*dcca.CCR){"service": func(q *dcca.CCR) { q.ServiceIdentifier++ }, "subscriber": func(q *dcca.CCR) { q.SubscriptionID[0].SubscriptionIDData = "15550000002" }, "context": func(q *dcca.CCR) { q.ServiceContextID = "other@example.net" }, "origin": func(q *dcca.CCR) { q.OriginHost = "other.example" }, "target": func(q *dcca.CCR) { q.DestinationHost = avp.Some("other.example") }, "realm-only": func(q *dcca.CCR) { q.DestinationHost = avp.Optional[string]{} }, "action": func(q *dcca.CCR) { q.RequestedAction = dcca.CheckBalance }, "extension": func(q *dcca.CCR) { q.Extra = []avp.Raw{{Key: avp.Key{Code: 999001}, Data: []byte{1}}} }} {
		t.Run(name, func(t *testing.T) {
			q := *refund
			q.SubscriptionID = append([]dcca.SubscriptionID(nil), refund.SubscriptionID...)
			change(&q)
			if e := dcca.CheckRefundRelation(original, &q); e == nil {
				t.Fatal("unrelated refund passed")
			}
		})
	}
	if e := dcca.CheckRefundRelation(nil, refund); !errors.Is(e, validation.ErrType) {
		t.Fatal(e)
	}
	before := *original
	if e := dcca.CheckRefundRelation(original, refund); e != nil || !reflect.DeepEqual(before, *original) {
		t.Fatal("modified original", e)
	}
}

func TestDCCACanonicalKeepsOperationAndUnitPresence(t *testing.T) {
	c := dccaCodec(t)
	var all [][]byte
	for _, name := range []string{"r-balance", "r-debit", "r-refund", "r-refund-zero", "r-refund-money", "r-refund-max", "r-unknown-optional"} {
		b, e := c.Canonical(fixture(t, "b70-"+name), 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, old := range all {
			if bytes.Equal(b, old) {
				t.Fatal("operation/unit/presence collapsed", name)
			}
		}
		all = append(all, b)
	}
	r := decoded(t, c, "b70-r-balance")
	r.Header.HopByHop++
	r.Header.Flags |= pdu.RetransmittedFlag
	r.Body.(*dcca.CCR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255}}}
	r.Body.(*dcca.CCR).RouteRecord = []string{"dra.example"}
	b, e := c.Canonical(encoded(t, c, r), 4096)
	if e != nil || !bytes.Equal(b, all[0]) {
		t.Fatal("arrival changes became business content", e)
	}
}
