// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codecengine

import (
	"github.com/skill215/diameterkit/validation"
)

func (e *Engine) Canonical(src []byte, maxBytes int) ([]byte, error) {
	m, err := e.Decode(src)
	if err != nil {
		return nil, err
	}
	op, ok := e.bundle.Lookup(m.Header.Key())
	if !ok || op.Canonical == nil {
		return nil, validation.ErrUnsupported
	}
	return op.Canonical(src, e.limits, maxBytes)
}
