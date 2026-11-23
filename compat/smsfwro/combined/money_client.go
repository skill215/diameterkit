// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded money ClientAddressWire value diagnostics without content reinterpretation.
// SPDX-AI-Date: 2026-11-23

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Static evidence retains no request or client bytes.
type moneyClientValue struct{}

func (moneyClientValue) Error() string { return validation.ErrUnsupported.Error() }
func (moneyClientValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (moneyClientValue) DiameterMoneyClientValueV1() string { return "combined-money-client-wire-v1" }
func invalidMoneyClient() error {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2018}}, Rule: "wire", Cause: moneyClientValue{}}
}
