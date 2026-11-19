// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Service-Identifier/Timezone isolation, mixed errors and wire-order acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestMoneyParameterModes(t *testing.T) {
	for _, kind := range []string{"general", "balance"} {
		for _, id := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for _, zone := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
				for _, fault := range []string{"plain", "money", "no-standard", "refund", "no-service", "no-timezone", "no-client", "no-mms"} {
					for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
						t.Run(fmt.Sprintf("%s/%d/%d/%s/%d", kind, id, zone, fault, mode), func(t *testing.T) {
							m := actionRequest(t, kind)
							m.ServiceIdentifier = avp.Some(id)
							m.Timezone = avp.Some(zone)
							switch fault {
							case "money":
								m.MultipleServicesCreditControl.RequestedServiceUnit.Value.Money = avp.Some(Money{})
							case "no-standard":
								m.MultipleServicesCreditControl.RequestedServiceUnit.Value.StandardMoney.Present = false
							case "refund":
								m.RefundInformation = avp.Some([]byte{})
							case "no-service":
								m.ServiceIdentifier.Present = false
							case "no-timezone":
								m.Timezone.Present = false
							case "no-client":
								m.ServiceInformation.SMSInformation.ClientAddressWire.Present = false
							case "no-mms":
								m.ServiceInformation.MMSInformation.Present = false
							}
							err := validateCCRSubset(m, mode)
							valid := fault == "plain" && id == 1 && zone == 0
							want := fault == "plain" && ((id != 1 && zone == 0) || (id == 1 && zone != 0)) && mode == validation.Receive
							var proof interface{ DiameterMoneyParameterValueV1() string }
							if (err == nil) != valid || errors.As(err, &proof) != want {
								t.Fatal(err)
							}
							if !valid {
								if !errors.Is(err, validation.ErrField) || errors.Is(err, validation.ErrUnsupported) {
									t.Fatal("classification", err)
								}
								var f *validation.FieldError
								if want {
									k := avp.Key{Code: 439}
									name := "combined-money-service-identifier-v1"
									if zone != 0 {
										k = avp.Key{VendorID: 10425, Code: 23}
										name = "combined-money-timezone-v1"
									}
									if !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != k {
										t.Fatal(err)
									}
									*m = CCR{}
									if proof.DiameterMoneyParameterValueV1() != name {
										t.Fatal("retained state")
									}
								} else if err != validation.ErrField {
									t.Fatal("legacy bare error", err)
								}
							}
						})
					}
				}
			}
		}
	}
}
func TestMoneyParameterRefundBoundary(t *testing.T) {
	for _, zone := range []bool{false, true} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, "refund")
			if zone {
				m.Timezone = avp.Some(int32(-1))
			} else {
				m.ServiceIdentifier = avp.Some(int32(0))
			}
			e := validateCCRSubset(m, mode)
			var proof interface{ DiameterMoneyParameterValueV1() string }
			if e != validation.ErrField || errors.As(e, &proof) {
				t.Fatal(e)
			}
		}
	}
}
