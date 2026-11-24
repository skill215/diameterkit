// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS client wire NUL and equality diagnostics with original bounded quotes.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Static evidence retains neither client bytes nor the comparison address.
type smsClientValue bool

func (smsClientValue) Error() string { return validation.ErrField.Error() }
func (smsClientValue) Unwrap() error { return validation.ErrField }
func (v smsClientValue) DiameterSMSClientValueV1() string {
	if v {
		return "combined-sms-client-nul-v1"
	}
	return "combined-sms-client-mismatch-v1"
}
func invalidSMSClient(nul bool) error {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2018}}, Rule: "wire", Cause: smsClientValue(nul)}
}
