// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded service-unit group omissions on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Preserve the unsupported classification of the existing unit-shape rule.
type ecurUnitOmission struct{ *profileGroupOmission }

func (*ecurUnitOmission) Error() string { return validation.ErrUnsupported.Error() }
func (*ecurUnitOmission) Unwrap() error {
	return errors.Join(validation.ErrMissing, validation.ErrUnsupported)
}

// Both groups require one UInt64 child. Its eight zero octets are a diagnostic
// placeholder, independent of any requested amount or observed usage.
func missingECURUnit(initial bool) error {
	const units = "\x00\x00\x00\x00\x00\x00\x01\xa1\x40\x08\x00\x00" // 417
	code := uint32(446)
	plan := "\x00\x00\x00\x00\x00\x00\x01\xbe\x40\x00\x01\x00" + units
	if initial {
		code = 437
		plan = "\x00\x00\x00\x00\x00\x00\x01\xb5\x40\x00\x01\x00" + units
	}
	group := &profileGroupOmission{key: avp.Key{Code: code}, flags: 64, plan: plan}
	mscc := &ecurUnitOmission{&profileGroupOmission{key: avp.Key{Code: 456}, flags: 64, child: group}}
	return &validation.FieldError{Path: []avp.Key{mscc.key, group.key}, Rule: "wire", Cause: mscc}
}
