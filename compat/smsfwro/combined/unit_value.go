// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS unit value diagnostics.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// A static discriminator, never a retained request or unit value.
type unitValue bool

func (unitValue) Error() string { return validation.ErrUnsupported.Error() }
func (unitValue) Unwrap() error { return errors.Join(validation.ErrUnsupported, validation.ErrField) }
func (reason unitValue) DiameterUnitValueV1() string {
	if reason {
		return "combined-termination-reporting-reason-v1"
	}
	return "combined-sms-requested-units-v1"
}
func invalidUnitValue(reason bool) error {
	path := []avp.Key{{Code: 456}, {Code: 437}, {Code: 417}}
	if reason {
		path = []avp.Key{{Code: 456}, {VendorID: 10415, Code: 872}}
	}
	return &validation.FieldError{Path: path, Rule: "wire", Cause: unitValue(reason)}
}
