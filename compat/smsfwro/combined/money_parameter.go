// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated Service-Identifier and Timezone value diagnostics.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Static branch evidence carries neither the request nor financial data.
type moneyParameterValue bool

func (moneyParameterValue) Error() string { return validation.ErrField.Error() }
func (moneyParameterValue) Unwrap() error { return validation.ErrField }
func (zone moneyParameterValue) DiameterMoneyParameterValueV1() string {
	if zone {
		return "combined-money-timezone-v1"
	}
	return "combined-money-service-identifier-v1"
}
func invalidMoneyParameter(zone bool) error {
	k := avp.Key{Code: 439}
	if zone {
		k = avp.Key{VendorID: 10425, Code: 23}
	}
	return validation.At(k, moneyParameterValue(zone))
}
