// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated address value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import "github.com/skill215/diameterkit/validation"

// Static evidence retains no address identity or request.
type addressDataValue struct{}

func (addressDataValue) Error() string { return validation.ErrField.Error() }
func (addressDataValue) Unwrap() error { return validation.ErrField }
func (addressDataValue) DiameterAddressDataValueV1() string {
	return "combined-address-data-nul-v1"
}
