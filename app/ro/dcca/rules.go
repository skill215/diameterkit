// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package dcca implements a bounded RFC4006 single-service EVENT binding.
// Its top-level units and explicit non-3GPP service context are independent of
// the SMS/MSCC and legacy bindings. Choose one CCR profile per Bundle.
// Rating, balance policy, refund eligibility and persistence belong to the caller.
package dcca

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

const (
	ApplicationID  uint32 = 4
	EventRequest   uint32 = 4
	DirectDebiting uint32 = 0
	RefundAccount  uint32 = 1
	CheckBalance   uint32 = 2
	EnoughCredit   uint32 = 0
	NoCredit       uint32 = 1
)

func at(code uint32, err error) error { return validation.At(avp.Key{Code: code}, err) }

// Known unselected credit-control fields cannot become supported by clearing M.
// Quoted Failed-AVP diagnostics are validated separately by the base codec.
func extras(v []avp.Raw) error {
	for _, x := range v {
		k := x.Key
		if k.VendorID == 0 && (k.Code >= 411 && k.Code <= 461 || k.Code == 258 || k.Code == 268 || k.Code == 297 || k.Code == 1292) || k.VendorID == 10415 && (k.Code >= 2000 && k.Code <= 2040 || k.Code == 873 || k.Code == 628 || k.Code == 3408 || k.Code == 3409) {
			return validation.At(k, validation.ErrUnsupported)
		}
	}
	return nil
}
func validateValueSubset(v *UnitValue, _ validation.Mode) error      { return extras(v.Extra) }
func validateMoneySubset(v *Money, _ validation.Mode) error          { return extras(v.Extra) }
func validateCostSubset(v *CostInformation, _ validation.Mode) error { return extras(v.Extra) }
func validateUnitSubset(v *ServiceUnit, _ validation.Mode) error {
	// This slice chooses one unit domain. It never sums or converts money/events.
	if v.CCMoney.Present == v.CCServiceSpecificUnits.Present {
		return validation.ErrUnsupported
	}
	return extras(v.Extra)
}
func validateSubscriptionSubset(v *SubscriptionID, _ validation.Mode) error {
	if v.SubscriptionIDType > 1 {
		return at(450, validation.ErrUnsupported)
	}
	if len(v.SubscriptionIDData) == 0 || len(v.SubscriptionIDData) > 15 || v.SubscriptionIDType == 1 && len(v.SubscriptionIDData) < 5 {
		return at(444, validation.ErrField)
	}
	for _, c := range v.SubscriptionIDData {
		if c < '0' || c > '9' {
			return at(444, validation.ErrField)
		}
	}
	return extras(v.Extra)
}
func validateCCRSubset(r *CCR, _ validation.Mode) error {
	if r.CCRequestType != EventRequest {
		return at(416, validation.ErrUnsupported)
	}
	if r.CCRequestNumber != 0 {
		return at(415, validation.ErrUnsupported)
	}
	if r.RequestedAction > CheckBalance {
		return at(436, validation.ErrUnsupported)
	}
	i := strings.LastIndexByte(r.ServiceContextID, '@')
	if i < 1 || avp.Identity(r.ServiceContextID[i+1:]) != nil {
		return at(461, validation.ErrField)
	}
	// The generic binding must not be used to suppress 3GPP-specific obligations.
	domain := strings.ToLower(r.ServiceContextID[i+1:])
	if domain == "3gpp.org" || strings.HasSuffix(domain, ".3gpp.org") {
		return at(461, validation.ErrUnsupported)
	}
	if len(r.SubscriptionID) == 2 && r.SubscriptionID[0].SubscriptionIDType == r.SubscriptionID[1].SubscriptionIDType {
		return at(443, validation.ErrDuplicate)
	}
	return extras(r.Extra)
}
func validateCCASubset(a *CCA, mode validation.Mode) error {
	if mode != validation.Receive && (a.CCRequestType != EventRequest || a.CCRequestNumber != 0) {
		return validation.ErrUnsupported
	}
	if mode != validation.Receive && a.CheckBalanceResult.Present && a.CheckBalanceResult.Value > NoCredit {
		return at(422, validation.ErrUnsupported)
	}
	if a.DirectDebitingFailureHandling.Present && a.DirectDebitingFailureHandling.Value > 1 {
		return at(428, validation.ErrUnsupported)
	}
	switch a.ResultCode {
	case 2001, 4010, 4011, 4012, 5030, 5011:
	case 5031:
		if !a.FailedAVP.Present {
			return at(279, validation.ErrMissing)
		}
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		if err := base.ValidateFailureDiagnostic(a.ResultCode, a.FailedAVP.Present, mode); err != nil {
			return err
		}
	default:
		return at(268, validation.ErrUnsupported)
	}
	return extras(a.Extra)
}
func validateCCARelation(r *CCR, a *CCA) error {
	if a.CCRequestType != r.CCRequestType {
		return at(416, validation.ErrInvalidAnswer)
	}
	if a.CCRequestNumber != r.CCRequestNumber {
		return at(415, validation.ErrInvalidAnswer)
	}
	if a.CheckBalanceResult.Present && r.RequestedAction != CheckBalance {
		return at(422, validation.ErrInvalidAnswer)
	}
	if a.GrantedServiceUnit.Present && r.RequestedAction != DirectDebiting {
		return at(431, validation.ErrInvalidAnswer)
	}
	if a.DirectDebitingFailureHandling.Present && r.RequestedAction != DirectDebiting {
		return at(428, validation.ErrInvalidAnswer)
	}
	return nil
}
func checkOneTimeEvent(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	if !ctx.Empty() {
		return validation.ErrContext
	}
	r, ok := request.(*CCR)
	if !ok {
		return validation.ErrContextRequired
	}
	a, ok := answer.(*CCA)
	if !ok {
		return nil
	}
	if a.CheckBalanceResult.Present && a.CheckBalanceResult.Value > NoCredit {
		return at(422, validation.ErrUnsupported)
	}
	if a.ResultCode == 2001 {
		if r.RequestedAction == CheckBalance && !a.CheckBalanceResult.Present {
			return at(422, validation.ErrMissing)
		}
		if r.RequestedAction == DirectDebiting && !a.GrantedServiceUnit.Present {
			return at(431, validation.ErrMissing)
		}
	}
	return nil
}
func validateEventError(a *pdu.ErrorAnswer, _ validation.Mode) error {
	if a.ExperimentalResult.Present {
		return at(297, validation.ErrUnsupported)
	}
	if a.ResultCode == 5031 && !a.FailedAVP.Present {
		return at(279, validation.ErrMissing)
	}
	return extras(a.Extra)
}
