// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package ro

import (
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// CheckECURContinuation checks the relation between caller-supplied, completed
// INITIAL and TERMINATION request bodies for the selected no-UPDATE SMS path.
// It checks the same Session, CTF/OCS endpoints, service context and service key,
// with request numbers 0 then 1. It does not retain either input, validate a prior
// grant, assert delivery, meter usage, allocate IDs or send a request. Callers
// validate each answer against its own request using the codec/runtime, and use
// a fresh exchange for TERMINATION. Do not mutate inputs concurrently.
func CheckECURContinuation(initial, termination *CCR) error {
	if initial == nil || termination == nil {
		return validation.ErrType
	}
	if err := validateCCR(initial, validation.Send); err != nil {
		return err
	}
	if err := validateCCR(termination, validation.Send); err != nil {
		return err
	}
	if initial.CCRequestType != InitialRequest || termination.CCRequestType != TerminationRequest {
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
	// Realm-only routing cannot prove the same OCS; a selected terminal is required.
	if !initial.DestinationHost.Present || !termination.DestinationHost.Present {
		return at(293, validation.ErrContextRequired)
	}
	if !strings.EqualFold(initial.DestinationHost.Value, termination.DestinationHost.Value) || !strings.EqualFold(initial.DestinationRealm, termination.DestinationRealm) {
		return at(293, validation.ErrContext)
	}
	if initial.MultipleServicesCreditControl[0].ServiceIdentifier[0] != termination.MultipleServicesCreditControl[0].ServiceIdentifier[0] {
		return at(456, at(439, validation.ErrContext))
	}
	return nil
}
