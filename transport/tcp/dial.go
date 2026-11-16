// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package tcp

import (
	"context"
	"strings"

	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
)

// Dialer connects to one frozen endpoint. Node supplies the absolute deadline;
// no internal retry, fallback to another configured peer or listener is created.
type Dialer struct{ network, address string }

func NewDialer(network, address string) (*Dialer, error) {
	if err := streamtransport.ValidateEndpoint(network, address); err != nil {
		return nil, err
	}
	return &Dialer{strings.Clone(network), strings.Clone(address)}, nil
}
func (d *Dialer) Info() transport.Info { return transport.Info{BufferBytes: BufferBytes} }
func (d *Dialer) Dial(ctx context.Context) (transport.Conn, error) {
	raw, err := streamtransport.DialTCP(ctx, d.network, d.address)
	if err != nil {
		return nil, err
	}
	return New(raw)
}

var _ transport.Dialer = (*Dialer)(nil)
