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

// Receive evidence for the two template-specific MSCC group exclusions.
type excludedECURUnit struct{}

func (excludedECURUnit) Error() string               { return validation.ErrUnsupported.Error() }
func (excludedECURUnit) Unwrap() error               { return validation.ErrUnsupported }
func (excludedECURUnit) DiameterExcludedECURUnitV1() {}

// The generated group validators have already checked child cardinality and extras.
func isolatedECURUnit(m *CCR) bool {
	u := &m.MultipleServicesCreditControl
	r := &u.RequestedServiceUnit.Value
	if len(m.Extra) != 0 || !u.RequestedServiceUnit.Present || !u.UsedServiceUnit.Present || r.Money.Present || r.StandardMoney.Present {
		return false
	}
	s := &m.ServiceInformation.SMSInformation
	switch m.CCRequestType {
	case 1:
		return r.CCServiceSpecificUnits == 1 && !u.ReportingReason.Present && !s.SMStatus.Present && !s.DischargeTime.Present
	case 3:
		return u.ReportingReason.Present && u.ReportingReason.Value == 2 && s.SMStatus.Present && s.DischargeTime.Present
	default:
		return false
	}
}
