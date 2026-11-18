// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded conditional EVENT Requested-Action value diagnostics.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestActionValueModes(t *testing.T) {
	for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, value := range []uint32{0, 1, 2, 3, ^uint32(0)} {
			for _, present := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%t/%d", kind, value, present, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						m.RequestedAction = avp.Optional[uint32]{Present: present, Value: value}
						err := validateCCRSubset(m, mode)
						event := kind != "initial" && kind != "term"
						want := mode == validation.Receive && present && event && (kind == "iec" && value != 0 || value > 2)
						var proof interface{ DiameterActionValueV1() string }
						if errors.As(err, &proof) != want {
							t.Fatal("proof scope", err)
						}
						if !want {
							return
						}
						rule := "combined-money-event-v1"
						if kind == "iec" {
							rule = "combined-iec-event-v1"
						}
						if proof.DiameterActionValueV1() != rule || !errors.Is(err, validation.ErrField) || !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrMissing) {
							t.Fatal(err)
						}
						var field *validation.FieldError
						if !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 436}) {
							t.Fatal(err)
						}
						*m = CCR{}
						if proof.DiameterActionValueV1() != rule {
							t.Fatal("borrowed request")
						}
					})
				}
			}
		}
	}
}
