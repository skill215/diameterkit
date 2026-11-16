// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package tls performs a bounded TLS handshake on an established TCP socket,
// then exposes the shared Diameter stream frame contract. Its optional Dialer
// includes TCP/TLS establishment; Listener defers TLS until socket admission.
// Neither API loads certificate files.
package tls

import (
	"context"
	stdtls "crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
)

const BufferBytes = streamtransport.BufferBytes

// Config requires a finite handshake timeout. Node.Dial/Serve additionally bound
// transport and CER/CEA with one absolute deadline. Standalone callers own CER/CEA
// timeout. TLS is cloned; callers must keep shared certificate/pool/callback data
// immutable as required by crypto/tls.Config. Callbacks must honor cancellation.
type Config struct {
	TLS              *stdtls.Config
	HandshakeTimeout time.Duration
}

// Conn exposes negotiated TLS state without granting raw socket/deadline access.
// BufferBytes excludes crypto/tls buffers, certificates, caches and kernel storage.
type Conn struct {
	stream *streamtransport.Conn
	secure *stdtls.Conn
}

func (c *Conn) ConnectionState() stdtls.ConnectionState { return c.secure.ConnectionState() }

// Client takes exclusive socket ownership even on failure. ServerName/trust and
// optional mutual authentication are explicit in Config.TLS; there is no fallback.
func Client(ctx context.Context, conn *net.TCPConn, cfg Config) (*Conn, error) {
	return handshake(ctx, conn, cfg, true)
}
func Server(ctx context.Context, conn *net.TCPConn, cfg Config) (*Conn, error) {
	return handshake(ctx, conn, cfg, false)
}
func handshake(ctx context.Context, raw *net.TCPConn, cfg Config, client bool) (*Conn, error) {
	if raw == nil {
		return nil, errors.New("nil TLS TCP connection")
	}
	fail := func(err error) (*Conn, error) { _ = raw.Close(); return nil, err }
	config, err := cloneConfig(cfg)
	if err != nil {
		return fail(err)
	}
	if err := raw.SetDeadline(time.Time{}); err != nil {
		return fail(err)
	}
	var secure *stdtls.Conn
	if client {
		secure = stdtls.Client(raw, config)
	} else {
		secure = stdtls.Server(raw, config)
	}
	hctx, cancel := context.WithTimeout(ctx, cfg.HandshakeTimeout)
	defer cancel()
	if err := secure.HandshakeContext(hctx); err != nil {
		return fail(err)
	}
	if err := hctx.Err(); err != nil {
		return fail(err)
	}
	// Close the underlying socket directly: TLS close_notify may otherwise wait on
	// a blocked writer. Diameter DPR/DPA already owns orderly application shutdown.
	return &Conn{stream: streamtransport.New(secure, raw.Close, true), secure: secure}, nil
}

func (c *Conn) Info() transport.Info { return c.stream.Info() }
func (c *Conn) ReadPrefix(ctx context.Context, timeout time.Duration) ([20]byte, error) {
	return c.stream.ReadPrefix(ctx, timeout)
}
func (c *Conn) ReadFrameInto(ctx context.Context, dst []byte) error {
	return c.stream.ReadFrameInto(ctx, dst)
}
func (c *Conn) DiscardFrame(ctx context.Context, size int) error {
	return c.stream.DiscardFrame(ctx, size)
}
func (c *Conn) WriteFrame(ctx context.Context, b []byte) transport.WriteResult {
	return c.stream.WriteFrame(ctx, b)
}
func (c *Conn) Close() error { return c.stream.Close() }

var _ transport.Conn = (*Conn)(nil)

func cloneConfig(cfg Config) (*stdtls.Config, error) {
	if cfg.TLS == nil || cfg.HandshakeTimeout <= 0 {
		return nil, errors.New("TLS config and positive handshake timeout required")
	}
	config := cfg.TLS.Clone()
	if config.MinVersion == 0 {
		config.MinVersion = stdtls.VersionTLS12
	}
	if config.MinVersion < stdtls.VersionTLS12 || config.MaxVersion != 0 && config.MaxVersion < config.MinVersion {
		return nil, errors.New("TLS version range must start at TLS 1.2 or later")
	}
	return config, nil
}
