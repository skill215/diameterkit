// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS template value diagnostics.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Static branch evidence contains no request storage or subscriber values.
type smsValue bool

func (smsValue) Error() string { return validation.ErrUnsupported.Error() }
func (smsValue) Unwrap() error { return errors.Join(validation.ErrUnsupported, validation.ErrField) }
func (node smsValue) DiameterSMSValueV1() string {
	if node {
		return "combined-sms-node-v1"
	}
	return "combined-sm-message-type-v1"
}
func invalidSMSValue(node bool) error {
	code := uint32(2007)
	if node {
		code = 2016
	}
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: code}}, Rule: "wire", Cause: smsValue(node)}
}
