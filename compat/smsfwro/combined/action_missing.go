// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional EVENT Requested-Action omission and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Preserve the mixed presence/value rule's earlier unsupported classification.
type requestedActionOmission struct{ *profileOmission }

func (*requestedActionOmission) Error() string { return validation.ErrUnsupported.Error() }
func (*requestedActionOmission) Unwrap() error {
	return errors.Join(validation.ErrMissing, validation.ErrUnsupported)
}

// The four zero bytes describe the minimum UInt32 width. They do not select a
// debit, refund or balance operation and retain no request or financial values.
func missingRequestedAction() error {
	leaf := &requestedActionOmission{&profileOmission{key: avp.Key{Code: 436}, flags: 64, width: 4}}
	return &validation.FieldError{Path: []avp.Key{leaf.key}, Rule: "wire", Cause: leaf}
}
