// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS address value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestSMSDataModes(t *testing.T) {
	states := []avp.Optional[[]byte]{{}, avp.Some([]byte{}), avp.Some([]byte("ok")), avp.Some([]byte{0})}
	for a := range states {
		for b := range states {
			for c := range states {
				for _, extra := range []bool{false, true} {
					for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
						t.Run(fmt.Sprintf("%d/%d/%d/%t/%d", a, b, c, extra, mode), func(t *testing.T) {
							m := &SMSInformation{OriginatorSCCPAddress: states[a], RecipientSCCPAddress: states[b], SMSCAddress: states[c], ClientAddressWire: avp.Some([]byte{0})}
							if extra {
								m.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}}}
							}
							count := 0
							var code uint32
							for i, state := range []int{a, b, c} {
								if state == 3 {
									count++
									code = []uint32{2008, 2010, 2017}[i]
								}
							}
							e := validateSMSInformationSubset(m, mode)
							var proof interface{ DiameterSMSDataValueV1() string }
							want := mode == validation.Receive && count == 1 && !extra
							if errors.As(e, &proof) != want {
								t.Fatal("proof", e)
							}
							var field *validation.FieldError
							if count > 0 {
								if !errors.Is(e, validation.ErrField) || errors.Is(e, validation.ErrUnsupported) {
									t.Fatal("classification", e)
								}
								if errors.As(e, &field) != want {
									t.Fatal("path boundary", e)
								}
								if want && (len(field.Path) != 1 || field.Path[0] != (avp.Key{VendorID: 10415, Code: code})) {
									t.Fatal("path", field)
								}
							}
							if (e == nil) != (count == 0 && !extra) {
								t.Fatal("acceptance", e)
							}
							*m = SMSInformation{}
							if want && proof.DiameterSMSDataValueV1() != "combined-sms-data-nul-v1" {
								t.Fatal("retained state")
							}
						})
					}
				}
			}
		}
	}
}
