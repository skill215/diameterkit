// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group exclusions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func TestECURUnitModes(t *testing.T) {
	for _, op := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, variant := range []string{"both", "status", "value", "missing", "money", "extras"} {
			t.Run(op+"/"+variant, func(t *testing.T) {
				m := actionRequest(t, op)
				u := &m.MultipleServicesCreditControl
				u.UsedServiceUnit.Present = true
				u.RequestedServiceUnit.Present = true
				switch variant {
				case "status":
					m.ServiceInformation.SMSInformation.SMStatus.Present = op != "term"
				case "value":
					if op == "term" {
						u.ReportingReason.Value = 9
					} else {
						u.RequestedServiceUnit.Value.CCServiceSpecificUnits = 9
					}
				case "missing":
					if op == "term" {
						u.UsedServiceUnit.Present = false
					} else {
						u.RequestedServiceUnit.Present = false
					}
				case "money":
					u.RequestedServiceUnit.Value.StandardMoney.Present = true
				case "extras":
					m.Extra = []avp.Raw{{Key: avp.Key{Code: 999}}}
				}
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					err := validateCCRSubset(m, mode)
					var proof interface{ DiameterExcludedECURUnitV1() }
					want := mode == validation.Receive && variant == "both" && (op == "initial" || op == "term")
					if errors.As(err, &proof) != want {
						t.Fatal(mode, err)
					}
					if want {
						code := uint32(446)
						if op == "term" {
							code = 437
						}
						var f *validation.FieldError
						if !errors.Is(err, validation.ErrUnsupported) || !errors.As(err, &f) || len(f.Path) != 2 || f.Path[0] != (avp.Key{Code: 456}) || f.Path[1] != (avp.Key{Code: code}) {
							t.Fatal(err)
						}
					}
				}
			})
		}
	}
}
