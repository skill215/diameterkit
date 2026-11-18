// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated IEC Originator-SCCP omission under a reviewed mixed condition.
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

func TestOriginatorMissingModes(t *testing.T) {
	for _, typ := range []int32{-1, 0, 1, 2, 3, 99} {
		for _, present := range []bool{false, true} {
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				t.Run(fmt.Sprintf("%d/%t/%d", typ, present, mode), func(t *testing.T) {
					m := NewIEC(IECInput{})
					s := &m.ServiceInformation.SMSInformation
					s.OriginatorInterface.InterfaceType = typ
					s.OriginatorSCCPAddress = avp.Optional[[]byte]{}
					if present {
						s.OriginatorSCCPAddress = avp.Some([]byte("originator"))
					}
					e := validateCCRSubset(m, mode)
					valid := present && typ == 1
					proof := !present && typ == 1 && mode == validation.Receive
					if (e == nil) != valid || (!valid && !errors.Is(e, validation.ErrUnsupported)) || errors.Is(e, validation.ErrMissing) != proof {
						t.Fatal(e)
					}
					q, ok := missingavp.Quote(e, wire.DefaultLimits(), 36)
					if ok != proof || ok && len(q) != 36 {
						t.Fatal(e, q)
					}
					if !valid && !proof {
						var f *validation.FieldError
						if !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0].Code != 2008 {
							t.Fatal("old path", e)
						}
					}
					if proof {
						*s = SMSInformation{}
						if q, ok := missingavp.Quote(e, wire.DefaultLimits(), 36); !ok || len(q) != 36 {
							t.Fatal("borrowed request")
						}
					}
				})
			}
		}
	}
	for _, kind := range []string{"initial", "term"} {
		for _, typ := range []int32{1, 3} {
			for _, present := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%t/%d", kind, typ, present, mode), func(t *testing.T) {
						m := NewInitial(ECURInput{})
						if kind == "term" {
							m = NewTermination(TerminationInput{})
						}
						s := &m.ServiceInformation.SMSInformation
						s.OriginatorInterface.InterfaceType = typ
						if present {
							s.OriginatorSCCPAddress = avp.Some([]byte{})
						}
						e := validateCCRSubset(m, mode)
						valid := !present && typ == 3
						if (e == nil) != valid || !valid && !errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrMissing) {
							t.Fatal("nonIEC", e)
						}
						if _, ok := missingavp.Quote(e, wire.DefaultLimits(), 36); ok {
							t.Fatal("nonIEC proof")
						}
					})
				}
			}
		}
	}
}
