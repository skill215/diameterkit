// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated INITIAL/TERMINATION interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// A static branch discriminator, not an interface value or retained request.
type ecurInterfaceValue bool

func (ecurInterfaceValue) Error() string { return validation.ErrUnsupported.Error() }
func (ecurInterfaceValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (origin ecurInterfaceValue) DiameterInterfaceValueV1() string {
	if origin {
		return "combined-ecur-originator-interface-v1"
	}
	return "combined-ecur-destination-interface-v1"
}
func invalidECURInterface(origin bool) error {
	code := uint32(2002)
	if origin {
		code = 2009
	}
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: code}, {VendorID: 10415, Code: 2006}}, Rule: "wire", Cause: ecurInterfaceValue(origin)}
}

type excludedECUROriginator struct{}

func (excludedECUROriginator) Error() string                     { return validation.ErrUnsupported.Error() }
func (excludedECUROriginator) Unwrap() error                     { return validation.ErrUnsupported }
func (excludedECUROriginator) DiameterExcludedECUROriginatorV1() {}

type excludedInitialStatus struct{}

func (excludedInitialStatus) Error() string                    { return validation.ErrUnsupported.Error() }
func (excludedInitialStatus) Unwrap() error                    { return validation.ErrUnsupported }
func (excludedInitialStatus) DiameterExcludedInitialStatusV1() {}
