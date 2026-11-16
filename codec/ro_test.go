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
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"math"
	"reflect"
	"testing"
)

func roCodec(t *testing.T) *codec.Codec {
	t.Helper()
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	rp, e := ro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(bp, rp)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func roSources() pdu.AnswerSources {
	return pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "ocs.example", Realm: "example.net"}}}
}
func TestRoIECFrozenPairs(t *testing.T) {
	c := roCodec(t)
	r := decoded(t, c, "v-iec-ccr-01")
	for _, name := range []string{"v-iec-cca-01", "v-iec-cca-02", "b39-a-child-denied", "b39-a-top-failure-child-success", "b39-a-empty-token", "b39-a-error"} {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, name)
			if err := c.ValidateAnswer(r, a, roSources()); err != nil {
				t.Fatal(err)
			}
			report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{})
			if !report.Complete || report.State != validation.Ready {
				t.Fatal(report)
			}
			if report := c.ValidateContext(a, validation.Context{}); report.Complete {
				t.Fatal("answer without original request", report)
			}
			// Ready means the selected protocol checks passed, not permission to debit.
			if name == "b39-a-top-failure-child-success" {
				v := a.Body.(*ro.CCA)
				if v.ResultCode != 4012 || v.MultipleServicesCreditControl[0].ResultCode.Value != 2001 {
					t.Fatal("result flattened")
				}
			}
		})
	}
	for _, name := range []string{"v-iec-ccr-01", "v-iec-cca-01", "v-iec-cca-02", "b39-a-error", "b39-r-zero", "b39-r-high", "b39-r-max", "b39-a-zero", "b39-a-high", "b39-a-max", "b39-r-empty-rsu", "b39-r-no-rsu", "b39-a-child-denied", "b39-a-top-failure-child-success", "b39-a-empty-token"} {
		t.Run(name+"/owned", func(t *testing.T) {
			raw := fixture(t, name)
			m, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			out := encoded(t, c, m)
			if name != "b39-a-error" && !bytes.Equal(out, raw) {
				t.Fatal("independent bytes changed")
			}
			again, e := c.Decode(out)
			if e != nil || !reflect.DeepEqual(again, m) {
				t.Fatal("semantic roundtrip", e)
			}
			clear(raw)
			if !bytes.Equal(out, encoded(t, c, m)) {
				t.Fatal("input alias")
			}
			n, e := c.Measure(m)
			if e != nil || n != len(out) {
				t.Fatal(n, e)
			}
			bound, e := c.DecodeBound(m.Header, n)
			if e != nil || bound < uint64(n) {
				t.Fatal(bound, e)
			}
			snap, e := c.Snapshot(m)
			if e != nil {
				t.Fatal(e)
			}
			if a, ok := m.Body.(*ro.CCA); ok && len(a.MultipleServicesCreditControl) > 0 {
				s := &a.MultipleServicesCreditControl[0]
				s.ServiceIdentifier[0] = 99
				if len(s.RefundInformation.Value) > 0 {
					s.RefundInformation.Value[0] = 55
				}
			}
			if q, ok := m.Body.(*ro.CCR); ok {
				q.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 99
				q.ServiceInformation.Value.SMSInformation.Value.RecipientInfo[0].RecipientSCCPAddress.Value.Data[0] = '9'
			}
			if !bytes.Equal(out, encoded(t, c, snap)) {
				t.Fatal("snapshot alias")
			}
		})
	}
	req := r.Body.(*ro.CCR)
	if !req.RequestedAction.Present || req.RequestedAction.Value != 0 || req.CCRequestNumber != 0 {
		t.Fatal("zero/presence lost")
	}
	a := decoded(t, c, "v-iec-cca-01").Body.(*ro.CCA)
	if !bytes.Equal(a.MultipleServicesCreditControl[0].RefundInformation.Value, []byte{0, 255}) {
		t.Fatal("binary token")
	}
	if v := decoded(t, c, "b39-a-max").Body.(*ro.CCA).MultipleServicesCreditControl[0].GrantedServiceUnit.Value.CCServiceSpecificUnits; !v.Present || v.Value != math.MaxUint64 {
		t.Fatal(v)
	}
}
func TestRoIECRawCounterexamples(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name string
		want error
	}{
		{"r-width4", avp.ErrValue}, {"r-width7", avp.ErrValue}, {"r-width9", avp.ErrValue}, {"a-width4", avp.ErrValue}, {"a-width7", avp.ErrValue}, {"a-width9", avp.ErrValue},
		{"r-duplicate-unit", validation.ErrDuplicate}, {"a-duplicate-unit", validation.ErrDuplicate}, {"r-no-action", validation.ErrMissing},
		{"r-refund", validation.ErrUnsupported}, {"r-initial", validation.ErrUnsupported}, {"r-sequence", validation.ErrUnsupported}, {"r-auth-app", validation.ErrField},
		{"r-extra-time", validation.ErrUnsupported}, {"r-top-unit", validation.ErrUnsupported}, {"r-vendor-unit", validation.ErrUnsupported}, {"r-duplicate-mscc", validation.ErrUnsupported}, {"r-sms-result", validation.ErrUnsupported}, {"r-wrong-sms-parent", validation.ErrUnsupported},
		{"a-no-gsu", validation.ErrMissing}, {"a-empty-gsu", validation.ErrMissing}, {"a-error-app-field", validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := c.Decode(fixture(t, "b39-"+tc.name)); !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
	r := decoded(t, c, "v-iec-ccr-01")
	for _, name := range []string{"a-unmatched-service", "a-unmatched-number", "a-unmatched-type"} {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, "b39-"+name)
			if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
				t.Fatal(e)
			}
		})
	}
	a := decoded(t, c, "b39-a-no-child-result")
	if a.Body.(*ro.CCA).MultipleServicesCreditControl[0].ResultCode.Present {
		t.Fatal("defaulted child result")
	}
	if report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{}); report.Complete {
		t.Fatal(report)
	}
	if _, err := c.Append(nil, a); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("unselected send", err)
	}
}
func TestRoIECPresenceCanonicalExtensionsAndLimits(t *testing.T) {
	c := roCodec(t)
	var canon [][]byte
	for _, name := range []string{"v-iec-ccr-01", "b39-r-zero", "b39-r-empty-rsu", "b39-r-no-rsu", "b39-r-max"} {
		b, err := c.Canonical(fixture(t, name), 4096)
		if err != nil {
			t.Fatal(err)
		}
		for _, old := range canon {
			if bytes.Equal(old, b) {
				t.Fatal("units or group presence collapsed", name)
			}
		}
		canon = append(canon, b)
	}
	r := decoded(t, c, "v-iec-ccr-01")
	q := r.Body.(*ro.CCR)
	q.Extra = []avp.Raw{{Key: avp.Key{VendorID: 99999, Code: 420}, Flags: 128, Data: []byte{0, 255, 1}}}
	out := encoded(t, c, r)
	again, e := c.Decode(out)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(again.Body.(*ro.CCR).Extra[0].Data, []byte{0, 255, 1}) {
		t.Fatal("optional extension")
	}
	q.Extra[0].Flags |= 64
	if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal(err)
	}
	q.Extra = nil
	// Common envelope source and correlation checks still run for App4.
	a := decoded(t, c, "v-iec-cca-01")
	a.Body.(*ro.CCA).SessionID = "wrong"
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	a = decoded(t, c, "v-iec-cca-01")
	if e := c.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "other.example", Realm: "example.net"}}}); !errors.Is(e, validation.ErrSource) {
		t.Fatal(e)
	}
	q.MultipleServicesCreditControl = append(q.MultipleServicesCreditControl, q.MultipleServicesCreditControl[0])
	if e := c.Validate(r); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal("multi service subset", e)
	}
	bp, _ := base.Profile()
	rp, _ := ro.Profile()
	b, _ := schema.Compile(bp, rp)
	l := wire.DefaultLimits()
	l.MaxAVPs = 4
	l.MaxRepeated = 4
	small, e := codec.New(b, codec.Config{Limits: l})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = small.Decode(fixture(t, "v-iec-ccr-01")); !errors.Is(e, wire.ErrLimit) {
		t.Fatal(e)
	}
}

func TestRoIECResultFamiliesAndSelectedSMSFields(t *testing.T) {
	c := roCodec(t)
	for _, code := range []uint32{4012, 5012, 5001, 5004, 5005, 5007, 5008, 5009, 5014, 3004, 4011, 5011, 5030, 5031} {
		a := decoded(t, c, "v-iec-cca-02")
		a.Body.(*ro.CCA).ResultCode = code
		err := c.Validate(a)
		switch code {
		case 4012, 5012, 5011, 4011, 5030:
			if err != nil {
				t.Fatal(code, err)
			}
		case 5001, 5004, 5005, 5007, 5008, 5009, 5014, 5031:
			if err == nil {
				t.Fatal("missing diagnostic accepted", code)
			}
		default:
			if !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(code, err)
			}
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func(*ro.CCR)
		want   error
	}{
		{"result-in-request-M0", func(r *ro.CCR) { r.Extra = []avp.Raw{{Key: avp.Key{Code: 268}, Data: []byte{0, 0, 7, 209}}} }, validation.ErrUnsupported},
		{"context", func(r *ro.CCR) { r.ServiceContextID = "legacy" }, validation.ErrUnsupported},
		{"subscription", func(r *ro.CCR) { r.SubscriptionID[0].SubscriptionIDType = 1 }, validation.ErrUnsupported},
		{"number-format", func(r *ro.CCR) { r.SubscriptionID[0].SubscriptionIDData = "+12025550123" }, validation.ErrField},
		{"node", func(r *ro.CCR) { r.ServiceInformation.Value.SMSInformation.Value.SMSNode = avp.Some(uint32(2)) }, validation.ErrUnsupported},
		{"type", func(r *ro.CCR) { r.ServiceInformation.Value.SMSInformation.Value.SMMessageType = avp.Some(uint32(1)) }, validation.ErrUnsupported},
		{"interface", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.OriginatorInterface.Value.InterfaceType = avp.Some(uint32(5))
		}, validation.ErrUnsupported},
		{"family", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.RecipientInfo[0].RecipientSCCPAddress.Value = avp.Address{Family: 1, Data: []byte{127, 0, 0, 1}}
		}, validation.ErrUnsupported},
		{"address-data", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.RecipientInfo[0].RecipientSCCPAddress.Value.Data = []byte("tel:123")
		}, validation.ErrField},
		{"optional-interface", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.OriginatorInterface = avp.Optional[ro.Interface]{}
		}, nil},
		{"unknown-M0", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 90001, Code: 9}, Flags: 128, Data: []byte{255}}}
		}, nil},
		{"known-wrong-parent-M0", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 2006}, Flags: 128, Data: []byte{0, 0, 0, 1}}}
		}, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := decoded(t, c, "v-iec-ccr-01")
			tc.mutate(r.Body.(*ro.CCR))
			if err := c.Validate(r); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
}
