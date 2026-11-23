// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated MMS SMSKey NUL-value diagnostics.
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

func TestMMSKeyModes(t *testing.T) {
	samples := [][]byte{nil, {}, []byte("key"), {0}, {0, 1}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 1024)}
	for i, data := range samples {
		for _, present := range []bool{false, true} {
			for _, extra := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%d/%t/%t/%d", i, present, extra, mode), func(t *testing.T) {
						m := &MMSInformation{SMSKey: avp.Optional[[]byte]{Present: present, Value: data}}
						if extra {
							m.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}}}
						}
						e := validateMMSInformationSubset(m, mode)
						bad := present && bytes.IndexByte(data, 0) >= 0
						var proof interface{ DiameterMMSKeyValueV1() string }
						want := bad && !extra && mode == validation.Receive
						if errors.As(e, &proof) != want || errors.Is(e, validation.ErrField) != bad || errors.Is(e, validation.ErrUnsupported) != (extra && !bad) || (e == nil) != (!bad && !extra) {
							t.Fatal(e)
						}
						if bad {
							var f *validation.FieldError
							if !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 1210}) {
								t.Fatal(e)
							}
						}
						*m = MMSInformation{}
						if want && proof.DiameterMMSKeyValueV1() != "combined-mms-key-nul-v1" {
							t.Fatal("retained state")
						}
					})
				}
			}
		}
	}
}
