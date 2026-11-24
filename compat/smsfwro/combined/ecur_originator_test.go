// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Originator-SCCP-Address exclusion on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestECUROriginatorModes(t *testing.T) {
	for _, op := range []string{"initial", "term"} {
		for _, fault := range []string{"isolated", "orig", "dest", "both", "action", "unit", "status", "extra"} {
			t.Run(op+"/"+fault, func(t *testing.T) {
				m := actionRequest(t, op)
				s := &m.ServiceInformation.SMSInformation
				s.OriginatorSCCPAddress = avp.Some([]byte{})
				switch fault {
				case "orig":
					s.OriginatorInterface.InterfaceType = 1
				case "dest":
					s.DestinationInterface.InterfaceType = 1
				case "both":
					s.OriginatorInterface.InterfaceType = 1
					s.DestinationInterface.InterfaceType = 1
				case "action":
					m.RequestedAction = avp.Some(uint32(0))
				case "unit":
					m.MultipleServicesCreditControl.RequestedServiceUnit.Present = !m.MultipleServicesCreditControl.RequestedServiceUnit.Present
				case "status":
					s.SMStatus.Present = !s.SMStatus.Present
				case "extra":
					m.Extra = []avp.Raw{{Key: avp.Key{Code: 999}}}
				}
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					err := validateCCRSubset(m, mode)
					var marker interface{ DiameterExcludedECUROriginatorV1() }
					var f *validation.FieldError
					if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &f) {
						t.Fatal(mode, err)
					}
					want := mode == validation.Receive && fault == "isolated"
					if errors.As(err, &marker) != want {
						t.Fatal(mode, err)
					}
					if want {
						if len(f.Path) != 3 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 873}) || f.Path[1] != (avp.Key{VendorID: 10415, Code: 2000}) || f.Path[2] != (avp.Key{VendorID: 10415, Code: 2008}) {
							t.Fatal(mode, err)
						}
					} else {
						leaf := avp.Key{VendorID: 10415, Code: 2008}
						if fault == "action" {
							leaf = avp.Key{Code: 436}
						}
						if len(f.Path) != 1 || f.Path[0] != leaf {
							t.Fatal(mode, err)
						}
					}
				}
			})
		}
	}
}
