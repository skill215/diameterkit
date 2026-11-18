// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit combined subscription-count omission evidence.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/validation"
)

type subscriptionOmission struct{ count int }

func (*subscriptionOmission) Error() string { return validation.ErrMissing.Error() }
func (e *subscriptionOmission) Unwrap() error {
	if e.count == 0 {
		return validation.ErrMissing
	}
	// Preserve the previous handwritten subset failure for a one-entry request.
	return errors.Join(validation.ErrMissing, validation.ErrUnsupported)
}
func (e *subscriptionOmission) DiameterMissingCardinalityV1() (string, int) {
	return "combined-subscriptions-v1", e.count
}

// Nil preserves the original decoder/subset validation fallback.
func diagnoseCCRSubscriptions(m *CCR) error {
	if len(m.SubscriptionID) >= 2 {
		return nil
	}
	for _, v := range m.SubscriptionID {
		if v.SubscriptionIDType < 0 || v.SubscriptionIDType > 1 || validateSubscriptionIDSubset(&v, validation.Receive) != nil {
			return nil
		}
	}
	return at(443, &subscriptionOmission{count: len(m.SubscriptionID)})
}
