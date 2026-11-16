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
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"reflect"
	"testing"
)

func TestRoECURFrozenCycle(t *testing.T) {
	c := roCodec(t)
	for _, pair := range [][2]string{{"v-sms-ccr-01", "v-sms-cca-01"}, {"v-sms-ccr-01", "v-sms-cca-02"}, {"v-sms-ccr-01", "b40-init-a-error"}, {"v-term-ccr-01", "v-term-cca-01"}, {"v-term-ccr-01", "v-term-cca-02"}, {"v-term-ccr-01", "b40-term-a-mscc"}, {"v-term-ccr-01", "b40-term-a-error"}} {
		t.Run(pair[1], func(t *testing.T) {
			r := decoded(t, c, pair[0])
			a := decoded(t, c, pair[1])
			if e := c.ValidateAnswer(r, a, roSources()); e != nil {
				t.Fatal(e)
			}
			if report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{}); !report.Complete {
				t.Fatal(report)
			}
			for _, name := range pair {
				raw := fixture(t, name)
				m, e := c.Decode(raw)
				if e != nil {
					t.Fatal(e)
				}
				out := encoded(t, c, m)
				if m.Header.Flags&pdu.ErrorFlag == 0 && !bytes.Equal(raw, out) {
					t.Fatal("frozen bytes changed", name)
				}
				back, e := c.Decode(out)
				if e != nil || !reflect.DeepEqual(m, back) {
					t.Fatal("roundtrip", e)
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
			}
		})
	}
	initial := decoded(t, c, "v-sms-ccr-01")
	term := decoded(t, c, "v-term-ccr-01")
	if e := ro.CheckECURContinuation(initial.Body.(*ro.CCR), term.Body.(*ro.CCR)); e != nil {
		t.Fatal(e)
	}
	if initial.Header.HopByHop == term.Header.HopByHop || initial.Header.EndToEnd == term.Header.EndToEnd {
		t.Fatal("same exchange")
	}
	if e := c.ValidateAnswer(initial, decoded(t, c, "v-term-cca-01"), roSources()); e == nil {
		t.Fatal("termination completed initial")
	}
	// Pure request relation does not manufacture a grant or persist an INITIAL.
	if e := ro.CheckECURContinuation(nil, term.Body.(*ro.CCR)); !errors.Is(e, validation.ErrType) {
		t.Fatal(e)
	}
	before := encoded(t, c, term)
	if e := ro.CheckECURContinuation(initial.Body.(*ro.CCR), term.Body.(*ro.CCR)); e != nil || !bytes.Equal(before, encoded(t, c, term)) {
		t.Fatal("mutated term", e)
	}
}
func TestRoECURContinuationMismatches(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name   string
		change func(*ro.CCR)
		want   error
	}{
		{"session", func(r *ro.CCR) { r.SessionID = "other;1" }, validation.ErrContext},
		{"requester", func(r *ro.CCR) { r.OriginHost = "other.example" }, validation.ErrContext},
		{"requester-realm", func(r *ro.CCR) { r.OriginRealm = "other.net" }, validation.ErrContext},
		{"ocs", func(r *ro.CCR) { r.DestinationHost = avp.Some("other.example") }, validation.ErrContext},
		{"realm", func(r *ro.CCR) { r.DestinationRealm = "other.net" }, validation.ErrContext},
		{"realm-only", func(r *ro.CCR) { r.DestinationHost = avp.Optional[string]{} }, validation.ErrContextRequired},
		{"service", func(r *ro.CCR) { r.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2 }, validation.ErrContext},
		{"sequence", func(r *ro.CCR) { r.CCRequestNumber = 2 }, validation.ErrUnsupported},
		{"service-context", func(r *ro.CCR) { r.ServiceContextID = "other" }, validation.ErrUnsupported},
		{"casefold", func(r *ro.CCR) { r.DestinationHost = avp.Some("OCS.EXAMPLE"); r.OriginHost = "CTF.EXAMPLE" }, nil},
		{"units-not-inferred", func(r *ro.CCR) {
			r.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits = avp.Some(^uint64(0))
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initial := decoded(t, c, "v-sms-ccr-01").Body.(*ro.CCR)
			term := decoded(t, c, "v-term-ccr-01").Body.(*ro.CCR)
			tc.change(term)
			if e := ro.CheckECURContinuation(initial, term); !errors.Is(e, tc.want) {
				t.Fatal(e)
			}
		})
	}
	r := decoded(t, c, "v-term-ccr-01")
	a := decoded(t, c, "b40-term-a-wrong-number")
	if e := c.ValidateAnswer(r, a, roSources()); !errors.Is(e, validation.ErrInvalidAnswer) {
		t.Fatal(e)
	}
	a = decoded(t, c, "b40-term-a-refund")
	if report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{}); report.Complete {
		t.Fatal("refund on ECUR", report)
	}
	if _, e := c.Append(nil, a); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
func TestRoECURRawCounterexamples(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name string
		want error
	}{
		{"term-no-usu", validation.ErrUnsupported}, {"term-empty-usu", validation.ErrMissing}, {"term-width4", avp.ErrValue}, {"term-dup-usu", validation.ErrUnsupported}, {"term-dup-unit", validation.ErrDuplicate},
		{"term-no-reason", validation.ErrUnsupported}, {"term-threshold", validation.ErrUnsupported}, {"term-dup-reason", validation.ErrUnsupported}, {"term-usu-reason", validation.ErrUnsupported}, {"term-top-reason", validation.ErrUnsupported}, {"term-rsu", validation.ErrUnsupported},
		{"init-usu", validation.ErrUnsupported}, {"init-action", validation.ErrUnsupported}, {"term-action", validation.ErrUnsupported}, {"init-wrong-number", validation.ErrUnsupported}, {"term-wrong-number", validation.ErrUnsupported}, {"init-update", validation.ErrUnsupported},
		{"init-a-no-gsu", validation.ErrMissing}, {"term-a-gsu", validation.ErrUnsupported},
		{"init-imei15", validation.ErrUnsupported}, {"init-equipment-type", validation.ErrUnsupported}, {"init-rat-width", validation.ErrField}, {"init-rat-nr", validation.ErrUnsupported}, {"init-uli-short", validation.ErrField}, {"init-uli-plmn", validation.ErrField}, {"init-timezone-spare", validation.ErrUnsupported}, {"init-nested-equipment", validation.ErrUnsupported}, {"term-sms-result", validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b40-"+tc.name)); !errors.Is(e, tc.want) {
				t.Fatal(e)
			}
		})
	}
}
func TestRoECURUnitsOwnershipAndCanonical(t *testing.T) {
	c := roCodec(t)
	var canons [][]byte
	for _, name := range []string{"v-term-ccr-01", "b40-term-zero", "b40-term-max"} {
		m := decoded(t, c, name)
		out := encoded(t, c, m)
		if !bytes.Equal(out, fixture(t, name)) {
			t.Fatal(name)
		}
		snap, e := c.Snapshot(m)
		if e != nil {
			t.Fatal(e)
		}
		b := m.Body.(*ro.CCR)
		b.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value = 22
		b.MultipleServicesCreditControl[0].ReportingReason[0] = 1
		if !bytes.Equal(out, encoded(t, c, snap)) {
			t.Fatal("USU/reason alias")
		}
		canon, e := c.Canonical(out, 4096)
		if e != nil {
			t.Fatal(e)
		}
		for _, old := range canons {
			if bytes.Equal(old, canon) {
				t.Fatal("used units collapsed")
			}
		}
		canons = append(canons, canon)
	}
	for _, name := range []string{"b40-init-no-rsu", "b40-init-empty-rsu", "b40-init-uli-ecgi"} {
		m := decoded(t, c, name)
		if !bytes.Equal(encoded(t, c, m), fixture(t, name)) {
			t.Fatal(name)
		}
	}
	m := decoded(t, c, "v-sms-ccr-01")
	snap, e := c.Snapshot(m)
	if e != nil {
		t.Fatal(e)
	}
	b := m.Body.(*ro.CCR)
	b.UserEquipmentInfo.Value.UserEquipmentInfoValue[0] = '0'
	b.ServiceInformation.Value.PSInformation.Value.UserLocationInfo.Value[1] = 9
	if !bytes.Equal(encoded(t, c, snap), fixture(t, "v-sms-ccr-01")) {
		t.Fatal("equipment/location alias")
	}
	m = decoded(t, c, "b40-init-timezone-reserved")
	if m.Body.(*ro.CCR).ServiceInformation.Value.PSInformation.Value.MSTimeZone.Value[1] != 128 {
		t.Fatal("reserved bits lost")
	}
	if _, e := c.Append(nil, m); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal("reserved bits send", e)
	}
	a, e := c.Canonical(fixture(t, "b40-init-timezone-reserved"), 4096)
	if e != nil {
		t.Fatal(e)
	}
	normal, e := c.Canonical(fixture(t, "v-sms-ccr-01"), 4096)
	if e != nil || bytes.Equal(a, normal) {
		t.Fatal("timezone canonical rewritten", e)
	}
}

func TestRoECURPSAndDirectionBoundaries(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name string
		edit func(*ro.CCR)
		want error
	}{
		{"equipment-leading-zero", func(r *ro.CCR) { r.UserEquipmentInfo.Value.UserEquipmentInfoValue = []byte("0001542032375101") }, nil},
		{"equipment-TBCD", func(r *ro.CCR) { r.UserEquipmentInfo.Value.UserEquipmentInfoValue = make([]byte, 8) }, validation.ErrUnsupported},
		{"equipment-nondigit", func(r *ro.CCR) { r.UserEquipmentInfo.Value.UserEquipmentInfoValue = []byte("490154203237510F") }, validation.ErrUnsupported},
		{"ULI-empty", func(r *ro.CCR) { r.ServiceInformation.Value.PSInformation.Value.UserLocationInfo = avp.Some([]byte{}) }, validation.ErrField},
		{"ULI-no-tag", func(r *ro.CCR) {
			r.ServiceInformation.Value.PSInformation.Value.UserLocationInfo = avp.Some([]byte{0, 0xf1, 0x10, 0, 1})
		}, validation.ErrUnsupported},
		{"ULI-ECGI-spare", func(r *ro.CCR) {
			r.ServiceInformation.Value.PSInformation.Value.UserLocationInfo = avp.Some([]byte{129, 0, 0xf1, 0x10, 0xf1, 2, 3, 4})
		}, validation.ErrUnsupported},
		{"timezone-width", func(r *ro.CCR) { r.ServiceInformation.Value.PSInformation.Value.MSTimeZone = avp.Some([]byte{0x23}) }, validation.ErrField},
		{"timezone-digit", func(r *ro.CCR) { r.ServiceInformation.Value.PSInformation.Value.MSTimeZone = avp.Some([]byte{0xa3, 0}) }, validation.ErrField},
		{"negative-timezone-DST2", func(r *ro.CCR) { r.ServiceInformation.Value.PSInformation.Value.MSTimeZone = avp.Some([]byte{0x2b, 2}) }, nil},
		{"PS-wrong-parent-M0", func(r *ro.CCR) {
			r.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 21}, Flags: 128, Data: []byte{6}}}
		}, validation.ErrUnsupported},
		{"sequence-M0", func(r *ro.CCR) {
			r.ServiceInformation.Value.SMSInformation.Value.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3408}, Flags: 128, Data: []byte{0, 0, 0, 1}}}
		}, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := decoded(t, c, "v-sms-ccr-01")
			tc.edit(r.Body.(*ro.CCR))
			if err := c.Validate(r); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	for value := 0; value <= 255; value++ {
		r := decoded(t, c, "v-sms-ccr-01")
		r.Body.(*ro.CCR).ServiceInformation.Value.PSInformation.Value.RATType.Value[0] = byte(value)
		err := c.Validate(r)
		if (err == nil) != (value >= 1 && value <= 9) {
			t.Fatalf("RAT %d: %v", value, err)
		}
	}
	// Event requests must not acquire ECUR usage fields by sharing the typed MSCC.
	for _, reason := range []bool{false, true} {
		r := decoded(t, c, "v-iec-ccr-01")
		s := &r.Body.(*ro.CCR).MultipleServicesCreditControl[0]
		if reason {
			s.ReportingReason = []uint32{2}
		} else {
			s.UsedServiceUnit = []ro.ServiceUnit{{CCServiceSpecificUnits: avp.Some(uint64(0))}}
		}
		if err := c.Validate(r); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
	}
}
