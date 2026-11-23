// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Money client-wire value modes, variable quotes and public consumption boundaries.
// SPDX-AI-Date: 2026-11-23

package combined

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestMoneyClientModes(t *testing.T) {
	valid := []byte{0, 1, 127, 0, 0, 1}
	values := [][]byte{nil, valid, []byte("127.0.0.1"), bytes.Repeat([]byte{0}, 1024)}
	for i := range valid {
		b := append([]byte{}, valid...)
		b[i] ^= 1
		values = append(values, b)
	}
	for _, n := range []int{1, 3, 4, 5, 7, 8, 15, 16, 17} {
		values = append(values, bytes.Repeat([]byte{1}, n))
	}
	for _, kind := range []string{"general", "balance"} {
		for sample, data := range values {
			for _, fault := range []string{"plain", "absent", "no-mms", "parameter", "unit", "money", "action", "node", "later-mms"} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%s/%d", kind, sample, fault, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						m.ServiceInformation.SMSInformation.ClientAddressWire = avp.Some(data)
						switch fault {
						case "absent":
							m.ServiceInformation.SMSInformation.ClientAddressWire.Present = false
						case "no-mms":
							m.ServiceInformation.MMSInformation.Present = false
						case "parameter":
							m.Timezone = avp.Some(int32(1))
						case "unit":
							m.MultipleServicesCreditControl.RequestedServiceUnit.Present = false
						case "money":
							m.MultipleServicesCreditControl.RequestedServiceUnit.Value.Money = avp.Some(Money{})
						case "action":
							m.RequestedAction = avp.Some(uint32(3))
						case "node":
							m.ServiceInformation.SMSInformation.SMSNode = 0
						case "later-mms":
							m.ServiceInformation.MMSInformation.Value.Priority = avp.Some(int32(1))
						}
						e := validateCCRSubset(m, mode)
						want := (fault == "plain" || fault == "later-mms") && !bytes.Equal(data, valid) && mode == validation.Receive
						var proof interface{ DiameterMoneyClientValueV1() string }
						if errors.As(e, &proof) != want {
							t.Fatal(e)
						}
						if (e == nil) != (fault == "plain" && bytes.Equal(data, valid)) {
							t.Fatal("acceptance", e)
						}
						var f *validation.FieldError
						if want {
							if !errors.Is(e, validation.ErrField) || !errors.Is(e, validation.ErrUnsupported) || !errors.As(e, &f) || len(f.Path) != 3 {
								t.Fatal(e)
							}
							for i, c := range []uint32{873, 2000, 2018} {
								if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
									t.Fatal(f.Path)
								}
							}
							*m = CCR{}
							if proof.DiameterMoneyClientValueV1() != "combined-money-client-wire-v1" {
								t.Fatal("retained state")
							}
						} else if fault == "plain" && e != nil {
							if !errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrField) || !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 2018}) {
								t.Fatal("legacy classification", e)
							}
						}
					})
				}
			}
		}
	}
}
func TestMoneyClientOtherFamilies(t *testing.T) {
	for _, kind := range []string{"refund", "iec", "initial", "term"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, kind)
			m.ServiceInformation.SMSInformation.ClientAddressWire = avp.Some([]byte{1})
			e := validateCCRSubset(m, mode)
			var proof interface{ DiameterMoneyClientValueV1() string }
			if errors.As(e, &proof) {
				t.Fatal(kind, e)
			}
		}
	}
}
