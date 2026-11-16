// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package abort implements the bounded Ro ASR/ASA protocol (App 4, command 274).
// Compose Profile with one chosen CCR profile. The application decides whether
// a session was terminated and explicitly selects the ASA result. This package
// neither owns sessions nor sends CCR TERMINATION or persists charging state.
package abort

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

const (
	ApplicationID    uint32 = 4
	Success          uint32 = 2001
	UnknownSessionID uint32 = 5002
	UnableToComply   uint32 = 5012
)

// Known fields outside this slice cannot gain support by clearing M. Other
// optional extensions follow the common owned-raw rules; diagnostics quote AVPs
// and deliberately do not pass through this top-level field filter.
func extras(fields []avp.Raw) error {
	for _, f := range fields {
		k := f.Key
		if k.VendorID == 0 && (k.Code >= 411 && k.Code <= 461 || k.Code == 258 || k.Code == 268 || k.Code == 297 || k.Code == 292 || k.Code == 261 || k.Code == 262) || k.VendorID == 10415 && (k.Code == 873 || k.Code == 628 || k.Code == 3408 || k.Code == 3409) {
			return validation.At(k, validation.ErrUnsupported)
		}
	}
	return nil
}

func validateASRSubset(m *ASR, _ validation.Mode) error { return extras(m.Extra) }

func validateASASubset(m *ASA, _ validation.Mode) error {
	switch m.ResultCode {
	case Success, UnknownSessionID, UnableToComply:
	default:
		return validation.At(avp.Key{Code: 268}, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}

func validateAbortError(m *pdu.ErrorAnswer, _ validation.Mode) error {
	if m.ExperimentalResult.Present {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
