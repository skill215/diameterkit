// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional whole MMS group omission and bounded diagnostic acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestMMSGroupMissingModes(t *testing.T) {
	for _, kind := range []string{"initial", "term", "iec"} {
		for _, state := range []string{"present", "absent", "empty"} {
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				t.Run(fmt.Sprintf("%s/%s/%d", kind, state, mode), func(t *testing.T) {
					m := NewIEC(IECInput{})
					if kind == "initial" {
						m = NewInitial(ECURInput{})
					}
					if kind == "term" {
						m = NewTermination(TerminationInput{})
					}
					if state == "absent" {
						m.ServiceInformation.MMSInformation = avp.Optional[MMSInformation]{}
					}
					if state == "empty" {
						m.ServiceInformation.MMSInformation = avp.Some(MMSInformation{})
					}
					err := validateCCRSubset(m, mode)
					valid := state == "present"
					proof := state == "absent" && mode == validation.Receive
					if (err == nil) != valid || !valid && !errors.Is(err, validation.ErrMissing) {
						t.Fatal(err)
					}
					q, ok := missingavp.Quote(err, wire.DefaultLimits(), 180)
					if ok != proof || ok && len(q) != 180 {
						t.Fatal(err, q)
					}
					if !valid && !proof {
						var f *validation.FieldError
						if !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0].Code != 877 {
							t.Fatal("old path", err)
						}
					}
					if proof {
						*m = CCR{}
						if _, ok := missingavp.Quote(err, wire.DefaultLimits(), 180); !ok {
							t.Fatal("borrowed model")
						}
					}
				})
			}
		}
	}
	for _, kind := range []string{"general", "balance", "refund"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			t.Run(fmt.Sprintf("%s/%d", kind, mode), func(t *testing.T) {
				m := NewDirectDebit(GeneralInput{})
				if kind == "balance" {
					m = NewBalanceCheck(GeneralInput{})
				}
				if kind == "refund" {
					var err error
					m, err = NewRefund(RefundInput{RefundInformation: avp.Some([]byte{})})
					if err != nil {
						t.Fatal(err)
					}
				}
				m.ServiceInformation.MMSInformation = avp.Optional[MMSInformation]{}
				err := validateCCRSubset(m, mode)
				if _, ok := missingavp.Quote(err, wire.DefaultLimits(), 180); ok {
					t.Fatal("money proof", err)
				}
				if kind == "refund" {
					if err != nil {
						t.Fatal(err)
					}
				} else if !errors.Is(err, validation.ErrField) {
					t.Fatal(err)
				}
			})
		}
	}
}
