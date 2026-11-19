// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated money-template SMS-Node value diagnostics.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Static evidence retains no request or subscriber bytes.
type moneyNodeValue struct{}

func (moneyNodeValue) Error() string { return validation.ErrUnsupported.Error() }
func (moneyNodeValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (moneyNodeValue) DiameterMoneyNodeValueV1() string { return "combined-money-sms-node-v1" }
func invalidMoneyNode() error {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2016}}, Rule: "wire", Cause: moneyNodeValue{}}
}
