// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS template value acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestSMSValueModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, node := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			for _, message := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%d/%d", kind, node, message, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						m.ServiceInformation.SMSInformation.SMSNode = node
						m.ServiceInformation.SMSInformation.SMMessageType = message
						err := validateCCRSubset(m, mode)
						sms := kind == "iec" || kind == "initial" || kind == "term"
						want := sms && ((node != 0 && message == 0) || (node == 0 && message != 0)) && mode == validation.Receive
						var proof interface{ DiameterSMSValueV1() string }
						if errors.As(err, &proof) != want {
							t.Fatal(err)
						}
						if want {
							code := uint32(2007)
							name := "combined-sm-message-type-v1"
							if node != 0 {
								code = 2016
								name = "combined-sms-node-v1"
							}
							var f *validation.FieldError
							if !errors.Is(err, validation.ErrUnsupported) || !errors.Is(err, validation.ErrField) || errors.Is(err, validation.ErrMissing) || !errors.As(err, &f) || len(f.Path) != 3 {
								t.Fatal(err)
							}
							for i, c := range []uint32{873, 2000, code} {
								if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
									t.Fatal(f.Path)
								}
							}
							*m = CCR{}
							if proof.DiameterSMSValueV1() != name {
								t.Fatal("retained state")
							}
						} else if sms {
							if (err == nil) != (node == 0 && message == 0) {
								t.Fatal("acceptance changed", err)
							}
							if err != nil {
								var f *validation.FieldError
								if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 2000}) {
									t.Fatal("legacy classification", err)
								}
							}
						} else if (err == nil) != (node == 2) {
							t.Fatal("money acceptance changed", err)
						}
					})
				}
			}
		}
	}
}
