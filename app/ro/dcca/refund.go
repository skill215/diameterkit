// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package dcca

import (
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// CheckRefundRelation checks caller-supplied completed request bodies for a new
// refund of the same single service, subscriber set, service context and OCS.
// A different Session-Id is required: a refund is not retransmission of a debit.
// It does not require a successful original answer (unknown delivery is not proof
// of no debit), compare amounts, authorize a refund, or verify external history.
// It retains no input. 3GPP 2022 and private 1292 tokens are outside this binding.
func CheckRefundRelation(debit, refund *CCR) error {
	if debit == nil || refund == nil {
		return validation.ErrType
	}
	if e := validateCCR(debit, validation.Send); e != nil {
		return e
	}
	if e := validateCCR(refund, validation.Send); e != nil {
		return e
	}
	if debit.RequestedAction != DirectDebiting || refund.RequestedAction != RefundAccount {
		return at(436, validation.ErrContext)
	}
	if debit.SessionID == refund.SessionID {
		return at(263, validation.ErrContext)
	}
	if debit.ServiceContextID != refund.ServiceContextID {
		return at(461, validation.ErrContext)
	}
	if debit.ServiceIdentifier != refund.ServiceIdentifier {
		return at(439, validation.ErrContext)
	}
	if !strings.EqualFold(debit.OriginHost, refund.OriginHost) || !strings.EqualFold(debit.OriginRealm, refund.OriginRealm) {
		return at(264, validation.ErrContext)
	}
	if !debit.DestinationHost.Present || !refund.DestinationHost.Present {
		return at(293, validation.ErrContextRequired)
	}
	if !strings.EqualFold(debit.DestinationHost.Value, refund.DestinationHost.Value) || !strings.EqualFold(debit.DestinationRealm, refund.DestinationRealm) {
		return at(293, validation.ErrContext)
	}
	if len(debit.SubscriptionID) != len(refund.SubscriptionID) {
		return at(443, validation.ErrContext)
	}
	for _, a := range debit.SubscriptionID {
		matched := false
		for _, b := range refund.SubscriptionID {
			if a.SubscriptionIDType == b.SubscriptionIDType && a.SubscriptionIDData == b.SubscriptionIDData {
				matched = true
			}
		}
		if !matched {
			return at(443, validation.ErrContext)
		}
	}
	// Uninterpreted extensions cannot supply evidence of historical identity.
	if len(debit.Extra) != 0 || len(refund.Extra) != 0 {
		return validation.ErrUnsupported
	}
	for _, r := range []*CCR{debit, refund} {
		for _, s := range r.SubscriptionID {
			if len(s.Extra) != 0 {
				return at(443, validation.ErrUnsupported)
			}
		}
	}
	return nil
}
