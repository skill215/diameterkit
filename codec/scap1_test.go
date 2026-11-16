package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func scapCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := scap1.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(bp, sp)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestSCAP1PairsAndOwnedInstances(t *testing.T) {
	c := scapCodec(t)
	r := decoded(t, c, "b45-acr")
	for _, name := range []string{"acr", "acr-envelope", "aca-success", "aca-failed", "aca-diagnostic", "aca-error", "aca-application", "time-zero", "time-max", "subscription-empty", "subscription-raw", "param-empty", "param-text", "param-raw"} {
		t.Run(name, func(t *testing.T) {
			raw := fixture(t, "b45-"+name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if name != "aca-error" && !bytes.Equal(raw, out) {
				t.Fatal("source projection differs")
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
			if q, ok := m.Body.(*scap1.ACR); ok {
				if len(q.ServiceParameterInfo) != 11 {
					t.Fatal("repeated groups lost")
				}
				if !bytes.Equal(q.ServiceParameterInfo[2].ParameterValue, q.ServiceParameterInfo[5].ParameterValue) {
					t.Fatal("origin pairing")
				}
				q.ServiceParameterInfo[2].ParameterValue[0] = '9'
				if bytes.Equal(q.ServiceParameterInfo[2].ParameterValue, q.ServiceParameterInfo[5].ParameterValue) {
					t.Fatal("distinct parents alias")
				}
				q.ServiceParameterInfo[0].ParameterType = 99
			} else {
				if e = c.ValidateAnswer(r, m, roSources()); e != nil {
					t.Fatal(e)
				}
				if report := c.ValidateAnswerContext(r, m, roSources(), validation.Context{}); !report.Complete {
					t.Fatal(report)
				}
				if a, ok := m.Body.(*scap1.ACA); ok {
					a.AccountingRecordNumber = 77
				}
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot alias")
			}
		})
	}
}
func TestSCAP1RejectsTypesParentsAndBranches(t *testing.T) {
	c := scapCodec(t)
	for _, name := range []string{"subscription-nul", "missing-vsai", "missing-acct", "missing-subscription", "missing-subscriber", "missing-unit-type", "missing-unit-value", "missing-digits", "missing-number", "missing-record-type", "missing-action", "missing-timestamp", "bad-digits-width", "bad-number-width", "units-zero", "units-max", "wrong-unit-type", "wrong-record-type", "wrong-number", "wrong-action", "wrong-vendor-value", "wrong-acct", "auth-in-vsai", "wrong-unit-vendor", "standard-action", "standard-time", "unknown-group-child", "unexpected-cc-type", "param-nul", "missing-param", "extra-param", "swapped-params", "duplicate-param-type", "origin-copy-mismatch", "wrong-literal", "wrong-timezone", "missing-param-value", "wrong-param-parent", "aca-missing-number", "aca-ro-field", "aca-unselected-result", "header-ro", "header-acct", "header-private"} {
		t.Run(name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b45-"+name)); e == nil {
				t.Fatal("unsupported source shape")
			}
		})
	}
	r := decoded(t, c, "b45-acr")
	for _, name := range []string{"aca-number", "aca-record-type", "aca-number-negative"} {
		a := decoded(t, c, "b45-"+name)
		if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(name, e)
		}
	}
	a := decoded(t, c, "b45-aca-success")
	a.Header.EndToEnd++
	if e := c.ValidateAnswer(r, a, roSources()); e == nil {
		t.Fatal("wrong exchange")
	}
	a = decoded(t, c, "b45-aca-success")
	if e := c.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "dra.example", Realm: "example.net"}}}); e == nil {
		t.Fatal("adjacent DRA as terminal")
	}
	raw := fixture(t, "b45-acr")
	raw[4] &^= pdu.ProxiableFlag
	if _, e := c.Decode(raw); e == nil {
		t.Fatal("unselected P0")
	}
}
func eventInput() scap1.EventInput {
	return scap1.EventInput{SessionID: "ctf.example;legacy-scap1;1", SubscriptionMSISDN: avp.Some("15551234567"), Timestamp: avp.Some(avp.Time(3908988800)), DataCodingScheme: avp.Some("8"), OriginatorMSISDN: avp.Some("15551234567"), OriginatorVMSC: avp.Some("15550000001"), TrafficCase: avp.Some("MO"), DestinationMSISDN: avp.Some("15557654321"), ConcatenatedMessages: avp.Some("1"), UserDataLength: avp.Some("140")}
}
func TestSCAP1BuilderPresenceAndCanonical(t *testing.T) {
	c := scapCodec(t)
	q, e := scap1.NewEvent(eventInput())
	if e != nil {
		t.Fatal(e)
	}
	m, e := c.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: "ctf.example", OriginRealm: "example.net", DestinationRealm: "example.net", HopByHop: 0x51525354, EndToEnd: 0x61626364})
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(encoded(t, c, m), fixture(t, "b45-acr")) {
		t.Fatal("complete47-node source request differs")
	}
	mutations := []func(*scap1.EventInput){func(x *scap1.EventInput) { x.SubscriptionMSISDN.Present = false }, func(x *scap1.EventInput) { x.Timestamp.Present = false }, func(x *scap1.EventInput) { x.DataCodingScheme.Present = false }, func(x *scap1.EventInput) { x.OriginatorMSISDN.Present = false }, func(x *scap1.EventInput) { x.OriginatorVMSC.Present = false }, func(x *scap1.EventInput) { x.TrafficCase.Present = false }, func(x *scap1.EventInput) { x.DestinationMSISDN.Present = false }, func(x *scap1.EventInput) { x.ConcatenatedMessages.Present = false }, func(x *scap1.EventInput) { x.UserDataLength.Present = false }}
	for _, mut := range mutations {
		v := eventInput()
		mut(&v)
		if _, e := scap1.NewEvent(v); !errors.Is(e, validation.ErrMissing) {
			t.Fatal("missing input became zero/empty", e)
		}
	}
	v := eventInput()
	v.Timestamp = avp.Some(avp.Time(0))
	v.DataCodingScheme = avp.Some("")
	v.OriginatorMSISDN = avp.Some("")
	v.SubscriptionMSISDN = avp.Some("")
	q, e = scap1.NewEvent(v)
	if e != nil {
		t.Fatal(e)
	}
	q.OriginHost = "ctf.example"
	q.OriginRealm = "example.net"
	q.DestinationRealm = "example.net"
	if e = c.ValidateBody(q); e != nil {
		t.Fatal("explicit empty legacy text", e)
	}
	if q.EventTimestamp != 0 || q.RequestedServiceUnit.UnitValue.ValueDigits != 1 {
		t.Fatal("time/one-event literal")
	}
	var seen [][]byte
	for _, name := range []string{"acr", "time-zero", "time-max", "subscription-empty", "subscription-raw", "param-empty", "param-text", "param-raw"} {
		b, e := c.Canonical(fixture(t, "b45-"+name), 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, old := range seen {
			if bytes.Equal(old, b) {
				t.Fatal("canonical collapsed content", name)
			}
		}
		seen = append(seen, b)
	}
}
func TestSCAP1NamespaceAndBoundedOperations(t *testing.T) {
	c := scapCodec(t)
	for _, other := range []*codec.Codec{legacyCodec(t), refundCodec(t), roCodec(t)} {
		if _, e := other.Decode(fixture(t, "b45-acr")); e == nil {
			t.Fatal("implicitSCAP registration")
		}
	}
	for _, name := range []string{"b41-iec-request", "b42-refund-request", "v-iec-ccr-01"} {
		if _, e := c.Decode(fixture(t, name)); e == nil {
			t.Fatal("CCR accepted bySCAP")
		}
	}
	sp, e := scap1.Profile()
	if e != nil {
		t.Fatal(e)
	}
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(bp, sp)
	if e != nil {
		t.Fatal(e)
	}
	for _, code := range []uint32{257, 280, 282, 271} {
		op, ok := b.Lookup(pdu.CommandKey{ApplicationID: 0, Code: code, Request: true})
		if !ok || op.TransportReplay {
			t.Fatal("control/ACR namespace", code)
		}
	}
	if _, e = schema.Compile(bp, sp, sp); e == nil {
		t.Fatal("duplicate fixed binding")
	}
	limits := c.Limits()
	limits.MaxRepeated = 10
	limited, e := codec.New(b, codec.Config{Limits: limits})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = limited.Decode(fixture(t, "b45-acr")); e == nil {
		t.Fatal("eleven-parent budget bypass")
	}
	limits = c.Limits()
	limits.MaxAVPs = 50
	limits.MaxRepeated = 16
	limited, e = codec.New(b, codec.Config{Limits: limits})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = limited.Decode(fixture(t, "b45-acr")); e == nil {
		t.Fatal("51AVP budget bypass")
	}
	// Both completion directions associate the accounting fields, not CC fields.
	r := decoded(t, c, "b45-acr")
	a, e := c.CompleteAnswer(r, &scap1.ACA{ResultCode: 2001, AccountingRecordType: 1}, pdu.Origin{Host: "ocs.example", Realm: "example.net"})
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(encoded(t, c, a), fixture(t, "b45-aca-success")) {
		t.Fatal("ACA pair")
	}
	if _, e = c.CompleteAnswer(r, &scap1.ACA{ResultCode: 2001, AccountingRecordType: 1, AccountingRecordNumber: 1}, pdu.Origin{Host: "ocs.example", Realm: "example.net"}); e == nil {
		t.Fatal("wrong local number")
	}
}
