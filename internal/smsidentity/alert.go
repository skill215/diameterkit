// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package smsidentity

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// MaxAlertTokenBytes is a selected library limit, not a 3GPP IMSI length rule.
const MaxAlertTokenBytes = 256

// AlertToken preserves UTF8 bytes exactly; it performs no identity lookup or
// Unicode/case normalization. Shared by Router TFA output and later ALR input.
func AlertToken(s string) error {
	if len(s) == 0 || len(s) > MaxAlertTokenBytes || avp.UTF8(s) != nil {
		return validation.ErrField
	}
	return nil
}
