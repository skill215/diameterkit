// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded SM-Status exclusion on combined ECUR INITIAL requests.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func TestInitialStatusModes(t *testing.T) {
	for _, op := range []string{"initial", "iec", "term"} {
		for mask := 0; mask < 4; mask++ {
			for _, extra := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/extra-%t", op, mask, extra), func(t *testing.T) {
					m := actionRequest(t, op)
					s := &m.ServiceInformation.SMSInformation
					s.SMStatus = avp.Optional[[]byte]{Present: mask&1 != 0, Value: []byte{0}}
					s.DischargeTime = avp.Optional[avp.Time]{Present: mask&2 != 0}
					if extra {
						m.Extra = []avp.Raw{{Key: avp.Key{Code: 999}}}
					}
					for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
						err := validateCCRSubset(m, mode)
						var marker interface{ DiameterExcludedInitialStatusV1() }
						want := op == "initial" && mask == 1 && !extra && mode == validation.Receive
						if errors.As(err, &marker) != want {
							t.Fatal(mode, err)
						}
						if op == "term" || mask == 0 {
							continue
						}
						var f *validation.FieldError
						if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &f) {
							t.Fatal(mode, err)
						}
						if want {
							if len(f.Path) != 3 || f.Path[2] != (avp.Key{VendorID: 10415, Code: 2014}) {
								t.Fatal(mode, err)
							}
						} else if len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 2000}) {
							t.Fatal(mode, err)
						}
					}
				})
			}
		}
	}
}
