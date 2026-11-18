// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed conditional Recipient-Address omission and bounded acceptance.
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

func TestRecipientMissingModes(t *testing.T) {
	for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
		m := NewIEC(IECInput{})
		m.ServiceInformation.MMSInformation.Value.RecipientAddress = avp.Optional[Address]{}
		err := validateCCRSubset(m, mode)
		if !errors.Is(err, validation.ErrMissing) {
			t.Fatal(mode, err)
		}
		q, ok := missingavp.Quote(err, wire.DefaultLimits(), 64)
		if ok != (mode == validation.Receive) {
			t.Fatal(mode, err, q)
		}
		var field *validation.FieldError
		if !errors.As(err, &field) {
			t.Fatal(err)
		}
		if mode != validation.Receive && (len(field.Path) != 1 || field.Path[0].Code != 877) {
			t.Fatal("changed local path", field)
		}
		if mode == validation.Receive {
			m.ServiceInformation.MMSInformation.Value.RecipientAddress = avp.Some(Address{AddressType: 1, AddressData: []byte("later")})
			if q, ok := missingavp.Quote(err, wire.DefaultLimits(), 64); !ok || len(q) != 64 {
				t.Fatal("borrowed request", q)
			}
		}
	}
}
