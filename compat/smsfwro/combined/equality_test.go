// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Equality rules across the named SMS templates and validation modes.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"
	"slices"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func TestCombinedEqualityModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, tc := range []struct {
			name              string
			sub, ue           []byte
			event, submission avp.Time
			path              []avp.Key
		}{
			{"equal-empty", nil, []byte{}, 0, 0, nil},
			{"equal-opaque", []byte{0xff}, []byte{0xff}, ^avp.Time(0), ^avp.Time(0), nil},
			{"time", nil, nil, 0, ^avp.Time(0), []avp.Key{{VendorID: 10415, Code: 1202}}},
			{"identity", nil, []byte{0xff}, 0, 0, []avp.Key{{Code: 458}, {Code: 460}}},
			{"both", []byte{0xff}, nil, ^avp.Time(0), 0, []avp.Key{{Code: 458}, {Code: 460}}},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				m := actionRequest(t, kind)
				m.SubscriptionID[1].SubscriptionIDData = tc.sub
				m.UserEquipmentInfo.Value.UserEquipmentInfoValue = tc.ue
				m.EventTimestamp = tc.event
				m.ServiceInformation.MMSInformation.Value.SubmissionTime = avp.Some(tc.submission)
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					err := validateCCRSubset(m, mode)
					if tc.path == nil {
						if err != nil {
							t.Fatal(mode, err)
						}
						continue
					}
					var f *validation.FieldError
					if !errors.As(err, &f) || f.Cause != validation.ErrField || !slices.Equal(f.Path, tc.path) {
						t.Fatal(mode, err)
					}
				}
			})
		}
	}
}
