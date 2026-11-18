// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated IEC Originator-Interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestInterfaceValueModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
			for _, present := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%t/%d", kind, value, present, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						s := &m.ServiceInformation.SMSInformation
						s.OriginatorInterface.InterfaceType = value
						s.OriginatorSCCPAddress = avp.Optional[[]byte]{Present: present, Value: []byte{}}
						e := validateCCRSubset(m, mode)
						var proof interface{ DiameterInterfaceValueV1() string }
						want := (kind == "iec" && present && value != 1 || (kind == "initial" || kind == "term") && !present && value != 3) && mode == validation.Receive
						if errors.As(e, &proof) != want {
							t.Fatal(e)
						}
						if want {
							if !errors.Is(e, validation.ErrUnsupported) || !errors.Is(e, validation.ErrField) || errors.Is(e, validation.ErrMissing) {
								t.Fatal(e)
							}
							var f *validation.FieldError
							if !errors.As(e, &f) || len(f.Path) != 4 {
								t.Fatal(e)
							}
							for i, code := range []uint32{873, 2000, 2009, 2006} {
								if f.Path[i] != (avp.Key{VendorID: 10415, Code: code}) {
									t.Fatal(f.Path)
								}
							}
							*m = CCR{}
							name := "combined-iec-originator-interface-v1"
							if kind != "iec" {
								name = "combined-ecur-originator-interface-v1"
							}
							if proof.DiameterInterfaceValueV1() != name {
								t.Fatal("borrowed state")
							}
						} else if kind == "iec" && present && value != 1 {
							var f *validation.FieldError
							if !errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrField) || !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0].Code != 2008 {
								t.Fatal("old mode path", e)
							}
						}
					})
				}
			}
		}
	}
}
