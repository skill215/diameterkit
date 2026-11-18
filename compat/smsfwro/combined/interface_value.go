// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated IEC Originator-Interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Only the isolated IEC Receive branch emits this static proof. No decoded
// interface value, address or request buffer is retained.
type iecInterfaceValue struct{}

func (iecInterfaceValue) Error() string { return validation.ErrUnsupported.Error() }
func (iecInterfaceValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (iecInterfaceValue) DiameterInterfaceValueV1() string {
	return "combined-iec-originator-interface-v1"
}
func invalidIECInterface() error {
	return &validation.FieldError{Path: []avp.Key{
		{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000},
		{VendorID: 10415, Code: 2009}, {VendorID: 10415, Code: 2006},
	}, Rule: "wire", Cause: iecInterfaceValue{}}
}
