// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package transport defines the bounded frame transport used by the endpoint.
// Receive methods may return an owned *Event continuation. See Event for the
// required resume/terminal behavior; stream adapters return only frames/errors.
package transport

import (
	"context"
	"errors"
	"time"
)

// ErrFrameState reports a cursor/length misuse before stream bytes are consumed.
// Invalid on-wire lengths also return it and terminate the stream.
var ErrFrameState = errors.New("invalid stream frame cursor or length")

// Info declares fixed adapter-owned bookkeeping/scratch storage, excluding
// caller-owned frames, standard-library TLS/net storage and kernel buffers.
// It is a logical reserve, not a process heap/RSS bound.
// Constructors must validate this against the configured per-connection reserve.
type Info struct{ BufferBytes uint64 }

type WriteResult struct {
	Bytes int
	// Uncertain means the adapter cannot establish how much was accepted.
	Uncertain bool
	Err       error
}

// Conn has one reader and one writer. Close is idempotent and must promptly
// unblock all I/O. A method must stop borrowing input/output before returning.
// Non-nil connections passed to Connect/Accept transfer exclusive ownership even
// when admission fails. No method retains an entire caller frame after return.
type Conn interface {
	Info() Info
	// ReadPrefix consumes exactly one 20-byte prefix. The context bounds the
	// whole operation; frameTimeout starts at the first byte, not while idle.
	// The same absolute frame deadline continues through ReadFrameInto/Discard.
	ReadPrefix(context.Context, time.Duration) ([20]byte, error)
	// ReadFrameInto fills exactly dst, including the consumed prefix. It must
	// not allocate a second full frame. Only one prefix may be outstanding.
	ReadFrameInto(context.Context, []byte) error
	// DiscardFrame consumes the remaining bytes using bounded fixed scratch.
	DiscardFrame(context.Context, int) error
	WriteFrame(context.Context, []byte) WriteResult
	Close() error
}

// Dialer owns a fixed endpoint and immutable connection configuration. Info
// declares the maximum adapter reserve before Dial is called. Dial includes TLS
// when applicable, honors ctx and relinquishes all I/O before returning. Every
// non-nil returned Conn transfers ownership, even alongside an error. A Dialer
// must not start retries or retain a returned Conn. Implementations are trusted;
// a callback ignoring cancellation cannot be forcibly killed by the library.
type Dialer interface {
	Info() Info
	Dial(context.Context) (Conn, error)
}

// MessageWriter optionally selects per-frame ordering on a message transport.
// WriteFrame remains ordered; the endpoint may request unordered application
// traffic only after its capability-exchange barrier. Control stays ordered.
// Implementations still have exactly one writer and cannot retain frame bytes.
type MessageWriter interface {
	WriteFrameOrdered(context.Context, []byte, bool) WriteResult
}
