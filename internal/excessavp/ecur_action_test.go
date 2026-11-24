// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Requested-Action exclusion for combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type ecurActionMetadata struct{}

func (ecurActionMetadata) Error() string                 { return "excluded action" }
func (ecurActionMetadata) Unwrap() error                 { return validation.ErrUnsupported }
func (ecurActionMetadata) DiameterExcludedECURActionV1() {}
func ecurActionFixture(kind, number uint32) []byte {
	return bytes.Join([][]byte{actionWord(436, 17), actionWord(415, number), actionWord(455, 1), actionWord(416, kind), {0, 0, 1, 202, 64, 0, 0, 8}}, nil)
}
func TestECURActionSelectors(t *testing.T) {
	raw := ecurActionFixture(1, 0)
	proof := validation.At(avp.Key{Code: 436}, ecurActionMetadata{})
	for _, name := range []string{"valid", "term", "event", "number", "absent", "duplicate", "trailing", "flags", "vendor", "indicator", "equip-flags", "count", "frame", "field", "budget", "path", "bare"} {
		t.Run(name, func(t *testing.T) {
			b := bytes.Clone(raw)
			l := wire.DefaultLimits()
			budget := 12
			cause := proof
			switch name {
			case "term":
				b[23] = 1
				b[47] = 3
			case "event":
				b[47] = 4
			case "number":
				b[23] = 1
			case "absent":
				b = b[12:]
			case "duplicate":
				b = append(b, b[:12]...)
			case "trailing":
				b = append(b, 0)
			case "flags":
				b[4] = 0
			case "vendor":
				b[4] |= 128
			case "indicator":
				b[35] = 2
			case "equip-flags":
				b[52] = 0
			case "count":
				l.MaxAVPs = 4
				l.MaxRepeated = 4
			case "frame":
				l.MaxFrameBytes = len(b) + 19
				l.MaxFieldBytes = 12
			case "field":
				l.MaxFieldBytes = 11
			case "budget":
				budget = 11
			case "path":
				cause = validation.At(avp.Key{Code: 415}, ecurActionMetadata{})
			case "bare":
				cause = validation.At(avp.Key{Code: 436}, validation.ErrUnsupported)
			}
			q, c, ok := QuoteInvalid(b, cause, l, budget)
			if ok != (name == "valid" || name == "term") {
				t.Fatal(ok, c)
			}
			if ok {
				if c != 5008 || !bytes.Equal(q, b[:12]) {
					t.Fatal(c, q)
				}
			} else if q != nil || c != 0 {
				t.Fatal(q, c)
			}
		})
	}
}
func fuzzECURAction(t *testing.T, raw []byte, budget uint16) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	proof := validation.At(avp.Key{Code: 436}, ecurActionMetadata{})
	q, c, ok := QuoteInvalid(raw, proof, l, int(budget))
	if !ok {
		if c != 0 || q != nil {
			t.Fatal("partial result")
		}
		return
	}
	if c != 5008 || len(q) != 12 || int(budget) < len(q) {
		t.Fatal(c, q)
	}
	d := wire.NewDecoder(raw, l)
	var found int
	var kind, number uint32
	for {
		v, more, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		switch v.Key {
		case avp.Key{Code: 436}:
			found++
			if !bytes.Equal(q, raw[v.Offset:v.Offset+12]) {
				t.Fatal("wrong action")
			}
		case avp.Key{Code: 416}:
			kind = binary.BigEndian.Uint32(v.Data)
		case avp.Key{Code: 415}:
			number = binary.BigEndian.Uint32(v.Data)
		}
	}
	if found != 1 || !(kind == 1 && number == 0 || kind == 3 && number == 1) {
		t.Fatal("wrong selector")
	}
}
