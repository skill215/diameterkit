// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated money SMS-Node values, mixed errors and validation order.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestMoneyNodeModes(t *testing.T) {
	for _, kind := range []string{"general", "balance", "refund"} {
		for _, v := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for _, msg := range []int32{-2147483648, 0, 2, 2147483647} {
				for extra := 0; extra < 8; extra++ {
					for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
						t.Run(fmt.Sprintf("%s/%d/%d/%d/%d", kind, v, msg, extra, mode), func(t *testing.T) {
							m := actionRequest(t, kind)
							s := &m.ServiceInformation.SMSInformation
							s.SMSNode = v
							s.SMMessageType = msg
							switch extra {
							case 1:
								s.OriginatorSCCPAddress = avp.Some([]byte{})
							case 2:
								s.RecipientSCCPAddress = avp.Some([]byte{})
							case 3:
								s.SMSCAddress = avp.Some([]byte{})
							case 4:
								s.ReplyPathRequested = avp.Some(int32(0))
							case 5:
								s.SMUserDataHeader = avp.Some([]byte{})
							case 6:
								s.SMStatus = avp.Some([]byte{})
							case 7:
								s.DischargeTime = avp.Some(avp.Time(0))
							}
							e := validateCCRSubset(m, mode)
							want := v != 2 && extra == 0 && mode == validation.Receive
							var proof interface{ DiameterMoneyNodeValueV1() string }
							var f *validation.FieldError
							if (e == nil) != (v == 2 && extra == 0) || errors.As(e, &proof) != want {
								t.Fatal(e)
							}
							if e != nil {
								if !errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrField) != want || !errors.As(e, &f) {
									t.Fatal(e)
								}
								if want {
									if len(f.Path) != 3 {
										t.Fatal(f.Path)
									}
									for i, c := range []uint32{873, 2000, 2016} {
										if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
											t.Fatal(f.Path)
										}
									}
									*m = CCR{}
									if proof.DiameterMoneyNodeValueV1() != "combined-money-sms-node-v1" {
										t.Fatal("retained state")
									}
								} else if len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 2000}) {
									t.Fatal("legacy path", e)
								}
							}
						})
					}
				}
			}
		}
	}
}
