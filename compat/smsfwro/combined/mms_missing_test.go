// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Selected isolated combined MMS omissions and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestMMSMissingModes(t *testing.T) {
	for _, code := range []uint32{1202, 1209, 1212, 1210, 1216} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := NewIEC(IECInput{})
			mm := &m.ServiceInformation.MMSInformation.Value
			switch code {
			case 1202:
				mm.SubmissionTime = avp.Optional[avp.Time]{}
			case 1209:
				mm.Priority = avp.Optional[int32]{}
			case 1212:
				mm.UserDataLength = avp.Optional[uint32]{}
			case 1210:
				mm.SMSKey = avp.Optional[[]byte]{}
			case 1216:
				mm.StatusReportRequest = avp.Optional[int32]{}
			}
			e := validateCCRSubset(m, mode)
			if !errors.Is(e, validation.ErrMissing) {
				t.Fatal(code, mode, e)
			}
			q, ok := missingavp.Quote(e, wire.DefaultLimits(), 40)
			if ok != (mode == validation.Receive) {
				t.Fatal(code, mode, e, q)
			}
			var f *validation.FieldError
			if !errors.As(e, &f) {
				t.Fatal(e)
			}
			if mode != validation.Receive && (len(f.Path) != 1 || f.Path[0].Code != 877) {
				t.Fatal("local path", f)
			}
			*mm = MMSInformation{}
			if mode == validation.Receive {
				if _, ok := missingavp.Quote(e, wire.DefaultLimits(), 40); !ok {
					t.Fatal("borrowed values")
				}
			}
		}
	}
}
