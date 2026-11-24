// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded SM-Status exclusion on combined ECUR INITIAL requests.
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

type initialStatusProof struct{}

func (initialStatusProof) Error() string                    { return "excluded initial status" }
func (initialStatusProof) Unwrap() error                    { return validation.ErrUnsupported }
func (initialStatusProof) DiameterExcludedInitialStatusV1() {}
func initialStatusCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2014}}, Cause: initialStatusProof{}}
}
func initialStatusArea(payload []byte, units uint64) []byte {
	return append(ecurSMSArea(1, 2014, payload, 3, 2), unitPart(456, unitPart(437, unitPart(417, binary.BigEndian.AppendUint64(nil, units))))...)
}
func TestInitialStatusConsumer(t *testing.T) {
	payload := []byte{0, 255, 0}
	for _, test := range []string{"exact", "budget", "field", "frame", "count", "depth", "event", "sequence", "action", "no-units", "wrong-units", "duplicate-units", "used", "reporting", "bare", "path", "malformed-tail"} {
		t.Run(test, func(t *testing.T) {
			src := initialStatusArea(payload, 1)
			l := wire.DefaultLimits()
			budget := 40
			var cause error = initialStatusCause()
			switch test {
			case "budget":
				budget--
			case "field":
				l.MaxFieldBytes = 39
			case "frame":
				l.MaxFrameBytes = len(src) + 19
				l.MaxFieldBytes = l.MaxFrameBytes
			case "count":
				l.MaxAVPs = 22
				l.MaxRepeated = 22
			case "depth":
				l.MaxDepth = 3
			case "event":
				src[11] = 4
			case "sequence":
				src[23] = 1
			case "action":
				src = append(src, actionWord(436, 0)...)
			case "no-units":
				src = ecurSMSArea(1, 2014, payload, 3, 2)
			case "wrong-units":
				src = initialStatusArea(payload, 0)
			case "duplicate-units":
				src = append(src, unitPart(456, nil)...)
			case "used", "reporting":
				child := unitPart(446, nil)
				if test == "reporting" {
					child = interfacePart(872, make([]byte, 4))
				}
				body := append(unitPart(437, unitPart(417, binary.BigEndian.AppendUint64(nil, 1))), child...)
				src = append(ecurSMSArea(1, 2014, payload, 3, 2), unitPart(456, body)...)
			case "path":
				f := initialStatusCause()
				f.Path[2].Code = 2008
				cause = f
			case "bare":
				cause = validation.ErrUnsupported
			case "malformed-tail":
				src = append(src, 1)
			}
			q, c, ok := QuoteInvalid(src, cause, l, budget)
			if ok != (test == "exact") {
				t.Fatal(c, ok)
			}
			if ok {
				clear(src)
				if c != 5008 || !bytes.Equal(q, interfacePart(873, interfacePart(2000, interfacePart(2014, payload)))) {
					t.Fatal(c, q)
				}
			} else if len(q) != 0 || c != 0 {
				t.Fatal("partial quote")
			}
		})
	}
}
