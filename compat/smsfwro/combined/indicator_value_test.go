// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined indicator value acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestIndicatorValueModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for mask := 0; mask < 32; mask++ {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%d/%d", kind, value, mask, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						m.MultipleServicesIndicator = avp.Optional[int32]{Present: mask&1 != 0, Value: value}
						m.UserEquipmentInfo.Present = mask&2 != 0
						m.ServiceIdentifier.Present = mask&4 != 0
						m.Timezone.Present = mask&8 != 0
						m.RefundInformation.Present = mask&16 != 0
						// Test this existing rule directly; absent EVENT markers select money at the caller.
						err := validateSMSRequest(m, mode)
						valid := mask == 3 && value == 1
						want := mask == 3 && value != 1 && mode == validation.Receive
						var proof interface{ DiameterIndicatorValueV1() string }
						var field *validation.FieldError
						if (err == nil) != valid || errors.As(err, &proof) != want {
							t.Fatal("scope", err)
						}
						if want {
							if !errors.Is(err, validation.ErrField) || errors.Is(err, validation.ErrUnsupported) || !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 455}) {
								t.Fatal(err)
							}
							*m = CCR{}
							if proof.DiameterIndicatorValueV1() != "combined-sms-indicator-v1" {
								t.Fatal("retained input")
							}
						} else if !valid && err != validation.ErrField {
							t.Fatal("legacy classification", err)
						}
					})
				}
			}
		}
	}
}
func TestIndicatorValueMoney(t *testing.T) {
	for _, kind := range []string{"general", "balance", "refund"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, kind)
			if err := validateCCRSubset(m, mode); err != nil {
				t.Fatal(err)
			}
		}
	}
}
