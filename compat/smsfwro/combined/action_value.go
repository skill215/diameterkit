// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded conditional EVENT Requested-Action value diagnostics.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Only the selected Receive branch creates this static proof. No action value,
// request storage or financial state is retained. Keep the earlier unsupported
// classification while identifying this reviewed field-value diagnostic.
type requestedActionValue string

func (requestedActionValue) Error() string { return validation.ErrUnsupported.Error() }
func (requestedActionValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (e requestedActionValue) DiameterActionValueV1() string { return string(e) }
func invalidRequestedAction(rule string) error {
	return validation.At(avp.Key{Code: 436}, requestedActionValue(rule))
}
