//go:build !linux || 386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

// New closes an owned non-nil socket and reports the upstream platform error.
func New(raw *upstream.SCTPConn, cfg Config) (transport.Conn, error) {
	if raw != nil {
		_ = raw.Close()
	}
	_, e := upstream.MessageIORequirements(*SocketConfig().Messages)
	return nil, e
}
