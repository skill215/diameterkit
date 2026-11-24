// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Attribution of prohibited ECUR Requested-Action in each validation mode.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestECURActionModes(t *testing.T) {
	for _, kind := range []string{"initial", "term"} {
		for _, fault := range []string{"isolated", "interface", "units", "extra", "timestamp"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				m := actionRequest(t, kind)
				m.RequestedAction = avp.Some(^uint32(0))
				switch fault {
				case "interface":
					m.ServiceInformation.SMSInformation.DestinationInterface.InterfaceType = 9
				case "units":
					m.MultipleServicesCreditControl.RequestedServiceUnit.Present = !m.MultipleServicesCreditControl.RequestedServiceUnit.Present
				case "extra":
					m.Extra = []avp.Raw{{Key: avp.Key{Code: 999}}}
				case "timestamp":
					m.EventTimestamp++
				}
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					err := validateCCRSubset(m, mode)
					var proof interface{ DiameterExcludedECURActionV1() }
					if errors.As(err, &proof) != (mode == validation.Receive && fault == "isolated") {
						t.Fatal(mode, err)
					}
					var f *validation.FieldError
					if !errors.As(err, &f) || len(f.Path) != 1 {
						t.Fatal(mode, err)
					}
					if fault == "timestamp" {
						if f.Path[0] != (avp.Key{VendorID: 10415, Code: 1202}) || !errors.Is(err, validation.ErrField) {
							t.Fatal(mode, err)
						}
						continue
					}
					if f.Path[0] != (avp.Key{Code: 436}) || !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) {
						t.Fatal(mode, err)
					}
				}
			})
		}
	}
}
