// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded combined money MMS exclusion and acceptance.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// The general and balance projections use the same template. Its explicit
// general-mms-key negative fixture contains an otherwise valid SMSKey.
type moneyMMSKeyExclusion struct{}

func (moneyMMSKeyExclusion) Error() string                   { return validation.ErrUnsupported.Error() }
func (moneyMMSKeyExclusion) Unwrap() error                   { return validation.ErrUnsupported }
func (moneyMMSKeyExclusion) DiameterMoneyMMSKeyExclusionV1() {}
func excludedMoneyMMSKey() error {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 877}, {VendorID: 10415, Code: 1210}}, Rule: "wire", Cause: moneyMMSKeyExclusion{}}
}
