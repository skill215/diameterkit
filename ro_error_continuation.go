// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import "github.com/skill215/diameterkit/wire"

// These frozen record policy names enable only the finite error sets below.
// They do not certify OCS deduplication, session continuity or business success.
const (
	RoTransientRecoveryPolicy = "ro-transient-errors-v1"
	RoSMSFWRecoveryPolicy     = "smsfw-ro-recovery-v1"
)

// RoErrorContinuation is selected before segment admission. The application
// chooses its business phase; the library does not run Tx or implement DDFH.
type RoErrorContinuation uint8

const (
	RoErrorFinal         RoErrorContinuation = iota // Default: all answers terminate recovery.
	RoErrorTransient                                // 3002/3004/3005 only; late/replay phases use this set.
	RoErrorSMSFWBeforeTx                            // Also 4002/5001/5004/5005/5008/5012; named compatibility only.
)

func validRoErrorContinuation(policy string, mode RoErrorContinuation) bool {
	switch mode {
	case RoErrorFinal:
		return true
	case RoErrorTransient:
		return policy == RoTransientRecoveryPolicy || policy == RoSMSFWRecoveryPolicy
	case RoErrorSMSFWBeforeTx:
		return policy == RoSMSFWRecoveryPolicy
	default:
		return false
	}
}

func permitsRoError(mode RoErrorContinuation, code uint32) bool {
	if mode != RoErrorTransient && mode != RoErrorSMSFWBeforeTx {
		return false
	}
	switch code {
	case 3002, 3004, 3005:
		return true
	case 4002, 5001, 5004, 5005, 5008, 5012:
		return mode == RoErrorSMSFWBeforeTx
	default:
		return false
	}
}

// Extract only after the normal/E decoder, original-request and source checks
// succeed. Scan the bounded frame without encoding a body or retaining bytes.
// Experimental namespaces, duplicate result fields and nested codes never grant
// continuation. Wire parsing alone is deliberately not an authorization API.
func roValidatedBaseResult(frame []byte, limits wire.Limits) uint32 {
	if len(frame) < 20 {
		return 0
	}
	d := wire.NewDecoder(frame[20:], limits)
	var code uint32
	found := false
	for {
		v, ok, err := d.Next()
		if err != nil {
			return 0
		}
		if !ok {
			return code
		}
		if v.Key.VendorID != 0 {
			continue
		}
		switch v.Key.Code {
		case 297:
			return 0
		case 268:
			if found {
				return 0
			}
			found = true
			code, err = wire.U32(v.Data)
			if err != nil {
				return 0
			}
		}
	}
}
