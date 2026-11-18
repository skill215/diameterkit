// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated IEC Originator-SCCP omission under a reviewed mixed condition.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// The earlier subset classified the mixed condition as unsupported. Keep that
// sentinel even when an isolated Receive omission also gains missing evidence.
type iecOriginatorOmission struct{ *profileOmission }

func (*iecOriginatorOmission) Error() string { return validation.ErrUnsupported.Error() }
func (*iecOriginatorOmission) Unwrap() error {
	return errors.Join(validation.ErrMissing, validation.ErrUnsupported)
}

// Called only for IEC, absent Originator-SCCP and existing interface type1,
// after the request's earlier checks. Static metadata owns no request values.
func missingIECOriginator() error {
	leaf := &profileOmission{key: avp.Key{VendorID: 10415, Code: 2008}, flags: 192, wireType: "octets"}
	sms := &profileOmission{key: avp.Key{VendorID: 10415, Code: 2000}, flags: 192, child: leaf}
	service := &iecOriginatorOmission{&profileOmission{key: avp.Key{VendorID: 10415, Code: 873}, flags: 192, child: sms}}
	return &validation.FieldError{Path: []avp.Key{service.key, sms.key, leaf.key}, Rule: "wire", Cause: service}
}
