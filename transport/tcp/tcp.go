// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package tcp adapts established TCP sockets to the Diameter frame contract.
// Its optional Dialer establishes one socket; Node owns reconnect policy.
package tcp

import (
	"errors"
	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
	"net"
	"time"
)

const BufferBytes = streamtransport.BufferBytes

// New takes exclusive ownership, including on failure. Callers must not retain
// socket I/O or deadline access. Close directly closes the socket (no half-close).
func New(conn *net.TCPConn) (transport.Conn, error) {
	if conn == nil {
		return nil, errors.New("nil TCP connection")
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return streamtransport.New(conn, conn.Close, false), nil
}
