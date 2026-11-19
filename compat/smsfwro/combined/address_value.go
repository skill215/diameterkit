// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Address-Type value evidence.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/validation"
)

// Static evidence; no address, parent group or request bytes are retained.
type addressTypeValue struct{}

func (addressTypeValue) Error() string { return validation.ErrUnsupported.Error() }
func (addressTypeValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (addressTypeValue) DiameterAddressTypeValueV1() string { return "combined-address-type-v1" }
