// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded conditional EVENT Requested-Action value diagnostics.
// SPDX-AI-Date: 2026-11-18

package excessavp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type actionMetadata struct {
	rule  string
	cause error
}

func (m *actionMetadata) Error() string                 { return "action value" }
func (m *actionMetadata) Unwrap() error                 { return m.cause }
func (m *actionMetadata) DiameterActionValueV1() string { return m.rule }
func actionWord(code, value uint32) []byte {
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b, code)
	b[4] = 64
	b[7] = 12
	binary.BigEndian.PutUint32(b[8:], value)
	return b
}
func actionFixture(iec bool) []byte {
	b := append(actionWord(416, 4), actionWord(436, 3)...)
	if iec {
		b = append(b, actionWord(455, 1)...)
		b = append(b, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
	}
	return b
}
func actionCause(iec bool) *validation.FieldError {
	rule := "combined-money-event-v1"
	if iec {
		rule = "combined-iec-event-v1"
	}
	return validation.At(avp.Key{Code: 436}, &actionMetadata{rule: rule, cause: errors.Join(validation.ErrField, validation.ErrUnsupported)}).(*validation.FieldError)
}
func TestActionValueQuoteControls(t *testing.T) {
	for _, iec := range []bool{false, true} {
		label := "money"
		if iec {
			label = "iec"
		}
		for _, bad := range []string{"exact", "reordered", "bytes", "negative-budget", "field", "frame", "avps", "depth", "repeat", "path", "nested-path", "bare", "cause", "rule", "branch", "valid", "type", "no-type", "duplicate-type", "duplicate-action", "duplicate-marker", "marker-value", "missing-marker", "action-flags", "vendor", "length", "trailing", "empty"} {
			t.Run(label+"/"+bad, func(t *testing.T) {
				raw := actionFixture(iec)
				cause := actionCause(iec)
				m := cause.Cause.(*actionMetadata)
				limits := wire.DefaultLimits()
				budget := 12
				switch bad {
				case "reordered":
					raw = append(append([]byte{}, raw[12:24]...), append(raw[:12:12], raw[24:]...)...)
				case "bytes":
					budget = 11
				case "negative-budget":
					budget = -1
				case "field":
					limits.MaxFieldBytes = 11
				case "frame":
					limits.MaxFrameBytes = len(raw) + 19
				case "avps":
					limits.MaxAVPs = 1
				case "depth":
					limits.MaxDepth = 0
				case "repeat":
					limits.MaxRepeated = 0
				case "path":
					cause.Path[0].Code = 263
				case "nested-path":
					cause.Path = append(cause.Path, avp.Key{Code: 436})
				case "bare":
					cause.Cause = validation.ErrField
				case "cause":
					m.cause = validation.ErrUnsupported
				case "rule":
					m.rule = "unreviewed"
				case "branch":
					if iec {
						m.rule = "combined-money-event-v1"
					} else {
						m.rule = "combined-iec-event-v1"
					}
				case "valid":
					binary.BigEndian.PutUint32(raw[20:], 0)
				case "type":
					binary.BigEndian.PutUint32(raw[8:], 1)
				case "no-type":
					raw = raw[12:]
				case "duplicate-type":
					raw = append(raw, actionWord(416, 4)...)
				case "duplicate-action":
					raw = append(raw, actionWord(436, 99)...)
				case "duplicate-marker":
					raw = append(raw, actionWord(455, 1)...)
					raw = append(raw, actionWord(455, 1)...)
				case "marker-value":
					if iec {
						binary.BigEndian.PutUint32(raw[32:], 2)
					} else {
						raw = append(raw, actionWord(455, 2)...)
					}
				case "missing-marker":
					if iec {
						raw = raw[:36]
					} else {
						raw = append(raw, []byte{0, 0, 1, 202, 64, 0, 0, 8}...)
					}
				case "action-flags":
					raw[16] = 0
				case "vendor":
					raw[16] = 192
				case "length":
					raw[19] = 11
				case "trailing":
					raw = append(raw, 1)
				case "empty":
					raw = nil
				}
				out, code, ok := QuoteInvalid(raw, cause, limits, budget)
				want := bad == "exact" || bad == "reordered"
				if ok != want || !ok && (out != nil || code != 0) {
					t.Fatal("outcome", code, ok, out)
				}
				if ok {
					if code != 5004 || !bytes.Equal(out, actionWord(436, 3)) {
						t.Fatal("wrong quote", out)
					}
					clear(raw)
					if !bytes.Equal(out, actionWord(436, 3)) {
						t.Fatal("borrowed bytes")
					}
				}
			})
		}
	}
}
func TestActionValueExactBudget(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 44
	l.MaxFieldBytes = 12
	l.MaxAVPs = 2
	l.MaxDepth = 1
	l.MaxRepeated = 1
	out, code, ok := QuoteInvalid(actionFixture(false), actionCause(false), l, 12)
	if !ok || code != 5004 || len(out) != 12 {
		t.Fatal(code, ok)
	}
}

// This oracle checks the emitted instance and the named branch against all
// top-level selectors. Existing fuzz targets retain their original assertions.
func fuzzActionValue(t *testing.T, raw []byte, budget uint16) {
	t.Helper()
	for _, iec := range []bool{false, true} {
		l := wire.DefaultLimits()
		l.MaxFrameBytes = 4096
		l.MaxFieldBytes = 2048
		l.MaxAVPs = 64
		l.MaxRepeated = 64
		out, code, ok := QuoteInvalid(raw, actionCause(iec), l, int(budget))
		if !ok {
			if out != nil || code != 0 {
				t.Fatal("partial quote")
			}
			continue
		}
		if code != 5004 || len(out) != 12 || len(out) > int(budget) || !bytes.Equal(out[:8], actionWord(436, 0)[:8]) {
			t.Fatal("invalid quote")
		}
		value := binary.BigEndian.Uint32(out[8:])
		if value == 0 || !iec && value <= 2 {
			t.Fatal("valid action quoted")
		}
		d := wire.NewDecoder(raw, l)
		actions, kinds, markers, equipment := 0, 0, 0, 0
		for {
			v, found, err := d.Next()
			if err != nil {
				t.Fatal("malformed source accepted")
			}
			if !found {
				break
			}
			switch v.Key {
			case avp.Key{Code: 436}:
				actions++
				if !bytes.Equal(out, raw[v.Offset:v.Offset+12]) {
					t.Fatal("not exact instance")
				}
			case avp.Key{Code: 416}:
				kinds++
				if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 4 {
					t.Fatal("not EVENT")
				}
			case avp.Key{Code: 455}:
				markers++
				if len(v.Data) != 4 || binary.BigEndian.Uint32(v.Data) != 1 {
					t.Fatal("bad selector")
				}
			case avp.Key{Code: 458}:
				equipment++
			}
		}
		if actions != 1 || kinds != 1 || iec && (markers != 1 || equipment != 1) || !iec && (markers != 0 || equipment != 0) {
			t.Fatal("ambiguous branch")
		}
	}
}
