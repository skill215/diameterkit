// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated subscription value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import "github.com/skill215/diameterkit/validation"

// Static evidence retains no subscription identity or request.
type subscriptionDataValue struct{}

func (subscriptionDataValue) Error() string { return validation.ErrField.Error() }
func (subscriptionDataValue) Unwrap() error { return validation.ErrField }
func (subscriptionDataValue) DiameterSubscriptionDataValueV1() string {
	return "combined-subscription-data-nul-v1"
}
