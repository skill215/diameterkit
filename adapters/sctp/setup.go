// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"context"
	"errors"
	"time"

	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

var ErrConcurrentAccept = errors.New("SCTP listener allows one Accept caller")

// SetupStateReserve covers only adapter setup wrappers and synchronization.
// Upstream address/config/socket costs come from its versioned resource query.
const SetupStateReserve uint64 = 4096

// SetupConfig delegates combined local+remote N/Z/D bounds and owned setup to
// upstream. Every field is explicit; defaults retain the original adapter policy.
type SetupConfig struct {
	Transport                                   Config
	MaxAddresses, MaxZoneBytes, MaxAddressBytes int
}

func DefaultSetupConfig() SetupConfig {
	return SetupConfig{Transport: DefaultConfig(), MaxAddresses: 16, MaxZoneBytes: 256, MaxAddressBytes: 4096}
}
func (c SetupConfig) upstreamConfig() upstream.SetupConfig {
	socket, _ := c.Transport.SocketConfig() // Requirements validates before any I/O.
	return upstream.SetupConfig{Socket: socket, Limits: upstream.AddressLimits{MaxAddresses: c.MaxAddresses, MaxZoneBytes: c.MaxZoneBytes, MaxAddressBytes: c.MaxAddressBytes}}
}

// Requirements is a no-I/O upstream ledger query plus adapter-owned storage.
// Listener and every physical candidate are charged independently. Upstream
// scratch is included in its totals; it is never added a second time.
func (c SetupConfig) Requirements() (transport.ListenerInfo, error) {
	if _, e := c.Transport.SocketConfig(); e != nil {
		return transport.ListenerInfo{}, e
	}
	r, e := upstream.SetupRequirements(c.upstreamConfig())
	if e != nil {
		return transport.ListenerInfo{}, configError(e)
	}
	local, e := c.Transport.localBytes(r.IORequirements)
	if e != nil {
		return transport.ListenerInfo{}, e
	}
	return transport.ListenerInfo{BufferBytes: r.ListenerBytes + SetupStateReserve, ConnectionBufferBytes: r.DialerBytes + max(r.EstablishBytes, r.ConnectionBytes) + SetupStateReserve + local}, nil
}
func configError(e error) error {
	if errors.Is(e, upstream.ErrInvalidConfig) {
		return errors.Join(ErrConfig, e)
	}
	return e
}
func setupContext(ctx context.Context) error {
	if ctx == nil {
		return ErrConfig
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if d, ok := ctx.Deadline(); ok && !time.Now().Before(d) {
		return context.DeadlineExceeded
	}
	return nil
}

// Dialer is only the transport.Dialer wrapper. Upstream freezes/validates inputs
// and resolves/connects within the same context; Node owns admission and retry.
type Dialer struct {
	raw  *upstream.Dialer
	cfg  Config
	info transport.Info
}

func NewDialer(network, local, remote string, cfg SetupConfig) (*Dialer, error) {
	r, e := cfg.Requirements()
	if e != nil {
		return nil, e
	}
	raw, e := upstream.NewDialer(network, local, remote, cfg.upstreamConfig())
	if e != nil {
		return nil, configError(e)
	}
	return &Dialer{raw: raw, cfg: cfg.Transport, info: transport.Info{BufferBytes: r.ConnectionBufferBytes}}, nil
}
func (d *Dialer) Info() transport.Info { return d.info }
func (d *Dialer) Dial(ctx context.Context) (transport.Conn, error) {
	raw, e := d.raw.DialContext(ctx)
	if e != nil {
		return nil, configError(e)
	}
	c, e := New(raw, d.cfg)
	if e != nil {
		return nil, e
	}
	// The adapter handoff follows the upstream socket handoff; do not publish a
	// wrapper after this caller's context expired during adapter construction.
	if e = setupContext(ctx); e != nil {
		c.Close()
		return nil, e
	}
	return c, nil
}

var _ transport.Dialer = (*Dialer)(nil)
