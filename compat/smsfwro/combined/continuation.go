// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package combined

import (
	"bytes"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// CheckECURContinuation validates two caller-supplied complete request bodies:
// the no-UPDATE sequence, same Session/CTF/fixed OCS/ServiceContext, ordered
// subscriptions, SMSKey and originating/recipient MMS addresses. It neither
// retains input nor checks a prior CCA, usage against a grant, delivery success,
// time ordering, routing changes or financial state. Each answer must separately
// be validated against its own request. Do not mutate inputs concurrently.
func CheckECURContinuation(initial, termination *CCR) error {
	if initial == nil || termination == nil {
		return validation.ErrType
	}
	if e := validateCCR(initial, validation.Send); e != nil {
		return e
	}
	if e := validateCCR(termination, validation.Send); e != nil {
		return e
	}
	if initial.CCRequestType != 1 || termination.CCRequestType != 3 {
		return at(416, validation.ErrUnsupported)
	}
	if initial.SessionID != termination.SessionID {
		return at(263, validation.ErrContext)
	}
	if initial.ServiceContextID != termination.ServiceContextID {
		return at(461, validation.ErrContext)
	}
	if !strings.EqualFold(initial.OriginHost, termination.OriginHost) || !strings.EqualFold(initial.OriginRealm, termination.OriginRealm) {
		return at(264, validation.ErrContext)
	}
	if !initial.DestinationHost.Present || !termination.DestinationHost.Present {
		return at(293, validation.ErrContextRequired)
	}
	if !strings.EqualFold(initial.DestinationHost.Value, termination.DestinationHost.Value) || !strings.EqualFold(initial.DestinationRealm, termination.DestinationRealm) {
		return at(293, validation.ErrContext)
	}
	for i := range initial.SubscriptionID {
		if !bytes.Equal(initial.SubscriptionID[i].SubscriptionIDData, termination.SubscriptionID[i].SubscriptionIDData) {
			return at(443, validation.ErrContext)
		}
	}
	a, b := initial.ServiceInformation.MMSInformation.Value, termination.ServiceInformation.MMSInformation.Value
	if !bytes.Equal(a.SMSKey.Value, b.SMSKey.Value) || !bytes.Equal(a.OriginatorAddress.AddressData, b.OriginatorAddress.AddressData) || !bytes.Equal(a.RecipientAddress.Value.AddressData, b.RecipientAddress.Value.AddressData) {
		return vendorAt(877, validation.ErrContext)
	}
	return nil
}
