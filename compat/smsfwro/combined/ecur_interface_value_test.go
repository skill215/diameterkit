// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: INITIAL/TERMINATION interface value acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestECURInterfaceValueModes(t *testing.T) {
	for _, kind := range []string{"initial", "term"} {
		for _, origin := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
			for _, dest := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
				for _, address := range []bool{false, true} {
					for _, action := range []bool{false, true} {
						for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
							t.Run(fmt.Sprintf("%s/%d/%d/%t/%t/%d", kind, origin, dest, address, action, mode), func(t *testing.T) {
								m := actionRequest(t, kind)
								s := &m.ServiceInformation.SMSInformation
								s.OriginatorInterface.InterfaceType = origin
								s.DestinationInterface.InterfaceType = dest
								s.OriginatorSCCPAddress = avp.Optional[[]byte]{Present: address}
								m.RequestedAction = avp.Optional[uint32]{Present: action}
								err := validateCCRSubset(m, mode)
								isolated := (origin != 3 && dest == 2) || (origin == 3 && dest != 2)
								want := !address && !action && isolated && mode == validation.Receive
								var proof interface{ DiameterInterfaceValueV1() string }
								if errors.As(err, &proof) != want {
									t.Fatal(err)
								}
								if want {
									var field *validation.FieldError
									code := uint32(2002)
									name := "combined-ecur-destination-interface-v1"
									if origin != 3 {
										code = 2009
										name = "combined-ecur-originator-interface-v1"
									}
									if !errors.Is(err, validation.ErrUnsupported) || !errors.Is(err, validation.ErrField) || errors.Is(err, validation.ErrMissing) || !errors.As(err, &field) || len(field.Path) != 4 {
										t.Fatal(err)
									}
									for i, c := range []uint32{873, 2000, code, 2006} {
										if field.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
											t.Fatal(field.Path)
										}
									}
									*m = CCR{}
									if proof.DiameterInterfaceValueV1() != name {
										t.Fatal("retained request")
									}
								} else {
									valid := !address && !action && origin == 3 && dest == 2
									if valid != (err == nil) {
										t.Fatal("acceptance changed", err)
									}
									if !valid && (!errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField)) {
										t.Fatal("classification changed", err)
									}
									if !valid && !action {
										var field *validation.FieldError
										if !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{VendorID: 10415, Code: 2008}) {
											t.Fatal("legacy path", err)
										}
									}
								}
							})
						}
					}
				}
			}
		}
	}
}
