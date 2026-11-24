// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded combined money MMS exclusion and acceptance.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func TestMoneyMMSExclusions(t *testing.T) {
	for _, kind := range []string{"general", "balance", "refund"} {
		for mask := 0; mask < 64; mask++ {
			t.Run(fmt.Sprintf("%s/%02x", kind, mask), func(t *testing.T) {
				m := actionRequest(t, kind)
				mm := m.ServiceInformation.MMSInformation.Value
				mm.RecipientAddress = avp.Optional[Address]{Present: mask&1 != 0, Value: mm.OriginatorAddress}
				mm.SubmissionTime.Present = mask&2 != 0
				mm.Priority.Present = mask&4 != 0
				mm.UserDataLength.Present = mask&8 != 0
				mm.SMSKey = avp.Optional[[]byte]{Present: mask&16 != 0, Value: []byte("sms-0001")}
				mm.StatusReportRequest.Present = mask&32 != 0
				m.ServiceInformation.MMSInformation = avp.Some(mm)
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					err := validateCCRSubset(m, mode)
					if kind == "refund" {
						if err != validation.ErrField {
							t.Fatalf("%v refund MMS group: %v", mode, err)
						}
						continue
					}
					if mask == 0 {
						if err != nil {
							t.Fatal(mode, err)
						}
						continue
					}
					var f *validation.FieldError
					if !errors.Is(err, validation.ErrUnsupported) || !errors.As(err, &f) {
						t.Fatal(mode, err)
					}
					var proof interface{ DiameterMoneyMMSKeyExclusionV1() }
					want := mask == 16 && mode == validation.Receive
					if errors.As(err, &proof) != want {
						t.Fatal(mode, err)
					}
					path := []uint32{877}
					if want {
						path = []uint32{873, 877, 1210}
					}
					if len(f.Path) != len(path) {
						t.Fatal(mode, f.Path)
					}
					for i, c := range path {
						if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
							t.Fatal(mode, f.Path)
						}
					}
				}
			})
		}
	}
}
