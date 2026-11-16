// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package ro

import (
	"bytes"
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// RefundInformationForService returns an owned copy of the selected service's
// 10415/2022 value from a known original IEC answer. Absent with nil error means
// the selected parent is known and has no token. Present with an empty Value is
// still present. A nil original/answer, or no matching answer parent, cannot
// establish absence and returns ErrContextRequired. Errors never return a token.
//
// Both bodies must come from a pair already validated by codec.ValidateAnswer
// or the runtime, including header/connection and authorized source association.
// This helper additionally checks body/session/proxy, explicit original OCS,
// and the single service key. It cannot authenticate caller-supplied history.
// maxBytes bounds the token copy (0 permits only absent/empty, up to MaxWireSize).
// No body is retained; do not mutate inputs concurrently.
//
// This checks the 2022 echo condition only, not refund eligibility, amounts or
// the complete SMS refund procedure. It does not enable a refund CCR or convert
// private 1292 tokens. An unsuccessful original result is not proof of no debit.
func RefundInformationForService(original *CCR, answer *CCA, serviceID uint32, maxBytes int) (avp.Optional[[]byte], error) {
	v, err := refundInformation(original, answer, serviceID, maxBytes)
	if err != nil {
		return avp.Optional[[]byte]{}, err
	}
	if v.Present {
		v.Value = append([]byte(nil), v.Value...)
	}
	return v, nil
}

// CheckRefundInformationEcho checks the presence and exact bytes proposed for
// 10415/2022 in the specified refund service parent against the original IEC.
// It has the same pair-validation precondition and scope as
// RefundInformationForService. It checks a proposed leaf, not a refund message,
// new-operation identity, financial authorization or eligibility to send it.
// Missing information is never silently replaced with an absent token.
func CheckRefundInformationEcho(original *CCR, answer *CCA, serviceID uint32, echo avp.Optional[[]byte], maxBytes int) error {
	v, err := refundInformation(original, answer, serviceID, maxBytes)
	if err != nil {
		return err
	}
	if echo.Present && len(echo.Value) > maxBytes {
		return wire.ErrLimit
	}
	if echo.Present != v.Present || v.Present && !bytes.Equal(v.Value, echo.Value) {
		return at(456, smsAt(2022, validation.ErrContext))
	}
	return nil
}

// The returned leaf borrows from answer. Only the public accessor copies it.
func refundInformation(original *CCR, answer *CCA, serviceID uint32, maxBytes int) (avp.Optional[[]byte], error) {
	var zero avp.Optional[[]byte]
	if maxBytes < 0 || maxBytes > wire.MaxWireSize {
		return zero, wire.ErrConfig
	}
	if original == nil || answer == nil {
		return zero, validation.ErrContextRequired
	}
	// Reject ambiguous parents before traversal/copy. Missing answer parent does
	// not mean the corresponding service was explicitly returned without 2022.
	if len(original.MultipleServicesCreditControl) != 1 || len(answer.MultipleServicesCreditControl) > 1 {
		return zero, at(456, validation.ErrUnsupported)
	}
	if len(answer.MultipleServicesCreditControl) == 1 {
		token := answer.MultipleServicesCreditControl[0].RefundInformation
		if token.Present && len(token.Value) > maxBytes {
			return zero, wire.ErrLimit
		}
	}
	if err := validateCCR(original, validation.Receive); err != nil {
		return zero, err
	}
	if original.CCRequestType != EventRequest || original.RequestedAction.Value != DirectDebiting {
		return zero, at(416, validation.ErrUnsupported)
	}
	if err := validateCCA(answer, validation.Receive); err != nil {
		return zero, err
	}
	if err := validateCCARelation(original, answer); err != nil {
		return zero, err
	}
	if original.SessionID != answer.SessionID {
		return zero, at(263, validation.ErrInvalidAnswer)
	}
	if !pdu.EqualProxies(original.ProxyInfo, answer.ProxyInfo) {
		return zero, at(284, validation.ErrInvalidAnswer)
	}
	if !original.DestinationHost.Present {
		return zero, at(293, validation.ErrContextRequired)
	}
	if !strings.EqualFold(original.DestinationHost.Value, answer.OriginHost) || !strings.EqualFold(original.DestinationRealm, answer.OriginRealm) {
		return zero, validation.ErrSource
	}
	if err := checkSMSCreditControl(original, answer, validation.Context{}, validation.Receive); err != nil {
		return zero, err
	}
	if original.MultipleServicesCreditControl[0].ServiceIdentifier[0] != serviceID {
		return zero, at(456, at(439, validation.ErrContext))
	}
	if len(answer.MultipleServicesCreditControl) == 0 {
		return zero, at(456, validation.ErrContextRequired)
	}
	v := answer.MultipleServicesCreditControl[0].RefundInformation
	if !v.Present {
		v.Value = nil
	}
	return v, nil
}
