// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Invalid-value and unsupported AVP quote provenance and fuzz checks.
// SPDX-AI-Date: 2026-11-17

package excessavp

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type invalidMetadata struct {
	key    avp.Key
	flags  byte
	offset int
	code   uint32
	cause  error
	child  error
}

func (m *invalidMetadata) Error() string { return m.cause.Error() }
func (m *invalidMetadata) Unwrap() error { return m.cause }
func (m *invalidMetadata) DiameterInvalidAVPV1() (avp.Key, byte, int, uint32, error) {
	return m.key, m.flags, m.offset, m.code, m.child
}

func TestInvalidQuoteEvidence(t *testing.T) {
	for _, kind := range []string{"value", "unsupported", "bare", "wrong-code", "overflow-code", "wrong-cause", "optional", "offset", "flags", "path", "truncated", "reply", "field", "depth", "cycle"} {
		t.Run(kind, func(t *testing.T) {
			raw, _ := excessFixture()
			l := wire.DefaultLimits()
			budget := 1024
			m := &invalidMetadata{key: avp.Key{Code: 450}, flags: 0x40, offset: 12, code: 5004, cause: avp.ErrValue}
			e := validation.At(m.key, m).(*validation.FieldError)
			switch kind {
			case "unsupported":
				m.code = 5001
				m.cause = validation.ErrUnsupported
			case "bare":
				e.Cause = avp.ErrValue
			case "wrong-code":
				m.code = 5014
			case "overflow-code":
				m.code = ^uint32(0)
			case "wrong-cause":
				m.cause = wire.ErrLimit
			case "optional":
				m.code = 5001
				m.cause = validation.ErrUnsupported
				m.flags = 0
				raw[16] = 0
			case "offset":
				m.offset = 13
			case "flags":
				m.flags = 0
			case "path":
				e.Path[0].Code++
			case "truncated":
				raw = raw[:23]
			case "reply":
				budget = 11
			case "field":
				l.MaxFieldBytes = 11
			case "depth":
				l.MaxDepth = 0
			case "cycle":
				m.code = 0
				m.child = m
			}
			out, code, ok := QuoteInvalid(raw, e, l, budget)
			want := kind == "value" || kind == "unsupported"
			if ok != want {
				t.Fatalf("ok=%v code=%d", ok, code)
			}
			if !ok {
				if out != nil || code != 0 {
					t.Fatal("partial result")
				}
				return
			}
			if code != m.code || !bytes.Equal(out, raw[12:24]) {
				t.Fatal("wrong instance/code")
			}
			frozen := append([]byte(nil), out...)
			clear(raw)
			if !bytes.Equal(out, frozen) {
				t.Fatal("borrowed input")
			}
		})
	}
}

func FuzzInvalidQuote(f *testing.F) {
	raw, _ := excessFixture()
	f.Add(raw, uint16(12), uint32(5004), uint16(1024))
	f.Add(raw, uint16(24), uint32(5001), uint16(12))
	f.Add([]byte{1}, uint16(0), ^uint32(0), uint16(1))
	f.Add(initialStatusArea([]byte{0}, 1), uint16(0), uint32(5008), uint16(40))
	for _, kind := range []uint32{1, 3} {
		f.Add(ecurUnitArea(kind), uint16(0), uint32(5008), uint16(32))
	}
	f.Add(initialStatusArea(nil, 1), uint16(0), uint32(5008), uint16(36))
	f.Add(ecurOriginatorArea(1, nil, 3, 2), uint16(0), uint32(5008), uint16(36))
	f.Add(ecurOriginatorArea(3, []byte{255, 2, 3}, 3, 2), uint16(0), uint32(5008), uint16(40))
	f.Add(ecurActionFixture(1, 0), uint16(0), uint32(5008), uint16(12))
	f.Add(ecurActionFixture(3, 1), uint16(0), uint32(5008), uint16(12))
	f.Add(actionFixture(false), uint16(0), uint32(5004), uint16(12))
	f.Add(actionFixture(true), uint16(0), uint32(5004), uint16(12))
	f.Add(interfaceFixture(-1), uint16(0), uint32(5004), uint16(52))
	f.Add(interfaceFixture(1), uint16(0), uint32(5004), uint16(51))
	for _, kind := range []uint32{1, 3} {
		for _, origin := range []bool{false, true} {
			f.Add(ecurFixture(kind, origin, -1), uint16(0), uint32(5004), uint16(52))
		}
	}
	for _, kind := range []uint32{1, 3, 4} {
		for _, node := range []bool{false, true} {
			f.Add(smsFixture(kind, node, -1), uint16(0), uint32(5004), uint16(40))
		}
	}
	for _, kind := range []uint32{1, 3, 4} {
		f.Add(unitFixture(kind, 0), uint16(0), uint32(5004), uint16(32))
	}
	for _, kind := range []uint32{1, 3, 4} {
		f.Add(indicatorFixture(kind, 0), uint16(0), uint32(5004), uint16(12))
	}
	f.Add(equipmentFixture(-1, nil, false), uint16(0), uint32(5004), uint16(20))
	f.Add(equipmentFixture(1, []byte("id"), true), uint16(0), uint32(5004), uint16(20))
	for _, group := range []uint32{886, 1201} {
		f.Add(addressFixture(group, 0, nil, false), uint16(0), uint32(5004), uint16(52))
		f.Add(addressFixture(group, -1, []byte("addr"), true), uint16(0), uint32(5004), uint16(52))
	}
	for _, action := range []uint32{0, 1, 2} {
		f.Add(moneyNodeFixture(action, -1), uint16(0), uint32(5004), uint16(40))
	}
	for _, action := range []uint32{0, 2} {
		for _, zone := range []bool{false, true} {
			f.Add(moneyParameterFixture(action, zone, -1), uint16(0), uint32(5004), uint16(16))
		}
	}
	for _, action := range []uint32{0, 2} {
		for _, data := range [][]byte{nil, {1}, {0, 1, 127, 0, 0, 2}, bytes.Repeat([]byte{1}, 17)} {
			f.Add(moneyClientFixture(action, data), uint16(0), uint32(2018), uint16(128))
		}
	}
	for _, data := range [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17)} {
		f.Add(mmsKeyFixture(data, 31), uint16(0), uint32(1210), uint16(128))
	}
	for _, data := range [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17)} {
		f.Add(equipmentFixture(0, data, false), uint16(0), uint32(460), uint16(128))
	}
	for _, group := range []uint32{886, 1201} {
		for _, data := range [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17)} {
			f.Add(addressFixture(group, 1, data, false), uint16(0), uint32(897), uint16(128))
		}
	}
	for _, data := range [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17)} {
		selected := subscriptionFixture(1, data, false)
		f.Add(selected, uint16(0), uint32(444), uint16(128))
		f.Add(append(subscriptionFixture(0, []byte("ok"), true), selected...), uint16(0), uint32(444), uint16(128))
	}
	for _, code := range []uint32{2008, 2010, 2017} {
		for _, data := range [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17)} {
			f.Add(smsDataFixture(code, data, 127), uint16(0), code, uint16(128))
		}
	}
	for _, kind := range []uint32{1, 3, 4} {
		for _, data := range [][]byte{nil, {0}, []byte("different")} {
			f.Add(smsClientFixture(kind, data, 7), uint16(0), uint32(2018), uint16(128))
		}
	}
	for _, action := range []uint32{0, 2} {
		f.Add(moneyMMSFixture(action, []byte("abc")), uint16(0), uint32(5008), uint16(40))
	}
	f.Fuzz(func(t *testing.T, raw []byte, offset uint16, code uint32, budget uint16) {
		checkECURFuzzQuote(t, raw, budget, 2014, initialStatusCause())
		fuzzECURUnit(t, raw, budget)
		fuzzECUROriginator(t, raw, budget)
		fuzzECURAction(t, raw, budget)
		fuzzExcludedMoneyMMSKey(t, raw, budget)
		fuzzSMSClient(t, raw, budget, false)
		fuzzSMSClient(t, raw, budget, true)
		for _, c := range []uint32{2008, 2010, 2017} {
			fuzzSMSData(t, raw, budget, c)
		}
		fuzzSubscriptionData(t, raw, budget)
		fuzzAddressData(t, raw, budget, 886)
		fuzzAddressData(t, raw, budget, 1201)
		fuzzEquipmentData(t, raw, budget)
		fuzzMMSKey(t, raw, budget)
		fuzzMoneyClient(t, raw, budget)
		fuzzMoneyParameter(t, raw, budget, false)
		fuzzMoneyParameter(t, raw, budget, true)
		fuzzMoneyNode(t, raw, budget)
		fuzzAddressValue(t, raw, budget, 886)
		fuzzAddressValue(t, raw, budget, 1201)
		fuzzEquipmentValue(t, raw, budget)
		fuzzIndicatorValue(t, raw, budget)
		fuzzUnitValue(t, raw, budget, false)
		fuzzUnitValue(t, raw, budget, true)
		fuzzSMSValue(t, raw, budget, false)
		fuzzSMSValue(t, raw, budget, true)
		fuzzInterfaceValue(t, raw, budget)
		fuzzECURInterfaceValue(t, raw, budget, false)
		fuzzECURInterfaceValue(t, raw, budget, true)
		fuzzActionValue(t, raw, budget)
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 4096
		l.MaxFieldBytes = 2048
		l.MaxAVPs = 64
		l.MaxRepeated = 64
		cause := avp.ErrValue
		if code == 5001 {
			cause = validation.ErrUnsupported
		}
		m := &invalidMetadata{key: avp.Key{Code: 450}, flags: 0x40, offset: int(offset), code: code, cause: cause}
		out, result, ok := QuoteInvalid(raw, validation.At(m.key, m), l, int(budget))
		if !ok {
			return
		}
		if result != code || len(out) > int(budget) || len(out) > l.MaxFieldBytes {
			t.Fatal("invalid result/bounds")
		}
		d := wire.NewDecoder(raw, l)
		for {
			v, found, e := d.Next()
			if e != nil || !found {
				t.Fatal("invented instance")
			}
			if v.Offset == int(offset) {
				size := (8 + len(v.Data) + 3) &^ 3
				if v.Key != m.key || v.Flags != m.flags || !bytes.Equal(out, raw[v.Offset:v.Offset+size]) {
					t.Fatal("changed bytes")
				}
				break
			}
		}
	})
}
