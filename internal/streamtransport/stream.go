// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package streamtransport shares the bounded TCP/TLS frame cursor.
package streamtransport

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/skill215/diameterkit/transport"
)

// BufferBytes is the conservative adapter bookkeeping/scratch reserve. It excludes
// net/crypto/tls storage, kernel buffers and the caller's frame, not a heap bound.
const BufferBytes uint64 = 4096

// Conn has one reader and one writer; Close may run concurrently with either.
type Conn struct {
	io        net.Conn
	closeConn func() error
	tls       bool
	once      sync.Once
	closeErr  error
	prefix    [20]byte
	size      int
	deadline  time.Time
	scratch   [256]byte
}

func New(conn net.Conn, closeConn func() error, encrypted bool) *Conn {
	return &Conn{io: conn, closeConn: closeConn, tls: encrypted}
}
func (s *Conn) Info() transport.Info { return transport.Info{BufferBytes: BufferBytes} }
func (s *Conn) Close() error         { s.once.Do(func() { s.closeErr = s.closeConn() }); return s.closeErr }
func earlier(a, b time.Time) time.Time {
	if a.IsZero() || !b.IsZero() && b.Before(a) {
		return b
	}
	return a
}

// deadlineScope serializes first-byte deadline updates against cancellation. stop
// joins any active callback before clearing the direction's socket deadline.
type deadlineScope struct {
	mu       sync.Mutex
	ctx      context.Context
	set      func(time.Time) error
	deadline time.Time
	stop     func() bool
	done     chan struct{}
}

func scope(ctx context.Context, set func(time.Time) error, d time.Time) (*deadlineScope, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a := &deadlineScope{ctx: ctx, set: set, done: make(chan struct{})}
	if err := a.update(d); err != nil {
		return nil, err
	}
	a.stop = context.AfterFunc(ctx, func() {
		a.mu.Lock()
		_ = a.set(time.Now())
		a.mu.Unlock()
		close(a.done)
	})
	return a, nil
}
func (a *deadlineScope) update(d time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if t, ok := a.ctx.Deadline(); ok {
		d = earlier(d, t)
	}
	a.deadline = earlier(a.deadline, d)
	d = a.deadline
	if a.ctx.Err() != nil {
		d = time.Now()
	}
	return a.set(d)
}
func (a *deadlineScope) finish() {
	if !a.stop() {
		<-a.done
	}
	_ = a.set(time.Time{})
}
func (a *deadlineScope) check() error {
	if err := a.ctx.Err(); err != nil {
		return err
	}
	if !a.deadline.IsZero() && !time.Now().Before(a.deadline) {
		return context.DeadlineExceeded
	}
	return nil
}
func (a *deadlineScope) failure(err error) error {
	if cause := a.ctx.Err(); cause != nil {
		return errors.Join(cause, err)
	}
	if !a.deadline.IsZero() && !time.Now().Before(a.deadline) {
		return errors.Join(context.DeadlineExceeded, err)
	}
	return err
}
func (s *Conn) failed(a *deadlineScope, err error) error { _ = s.Close(); return a.failure(err) }
func (s *Conn) ReadPrefix(ctx context.Context, timeout time.Duration) ([20]byte, error) {
	if s.size != 0 || timeout <= 0 {
		return [20]byte{}, transport.ErrFrameState
	}
	a, err := scope(ctx, s.io.SetReadDeadline, time.Time{})
	if err != nil {
		return [20]byte{}, err
	}
	defer a.finish()
	if _, err = io.ReadFull(s.io, s.prefix[:1]); err != nil {
		return [20]byte{}, s.failed(a, err)
	}
	s.deadline = time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok {
		s.deadline = earlier(s.deadline, d)
	}
	if err = a.update(s.deadline); err != nil {
		return [20]byte{}, s.failed(a, err)
	}
	if _, err = io.ReadFull(s.io, s.prefix[1:]); err != nil {
		return [20]byte{}, s.failed(a, err)
	}
	if err = a.check(); err != nil {
		return [20]byte{}, s.failed(a, err)
	}
	s.size = int(binary.BigEndian.Uint32(s.prefix[:4]) & 0xffffff)
	if s.size < 20 {
		return [20]byte{}, s.failed(a, transport.ErrFrameState)
	}
	return s.prefix, nil
}
func (s *Conn) ReadFrameInto(ctx context.Context, dst []byte) error {
	if s.size < 20 || len(dst) != s.size {
		return transport.ErrFrameState
	}
	a, err := scope(ctx, s.io.SetReadDeadline, s.deadline)
	if err != nil {
		_ = s.Close()
		return err
	}
	defer a.finish()
	// Even a header-only frame may not bypass an expired total frame deadline.
	if !time.Now().Before(s.deadline) {
		return s.failed(a, context.DeadlineExceeded)
	}
	copy(dst, s.prefix[:])
	if _, err = io.ReadFull(s.io, dst[20:]); err != nil {
		return s.failed(a, err)
	}
	if err = a.check(); err != nil {
		return s.failed(a, err)
	}
	s.size = 0
	return nil
}
func (s *Conn) DiscardFrame(ctx context.Context, size int) error {
	if s.size < 20 || size != s.size {
		return transport.ErrFrameState
	}
	a, err := scope(ctx, s.io.SetReadDeadline, s.deadline)
	if err != nil {
		_ = s.Close()
		return err
	}
	defer a.finish()
	if !time.Now().Before(s.deadline) {
		return s.failed(a, context.DeadlineExceeded)
	}
	for remaining := size - 20; remaining > 0; {
		n := min(remaining, len(s.scratch))
		if _, err = io.ReadFull(s.io, s.scratch[:n]); err != nil {
			return s.failed(a, err)
		}
		remaining -= n
	}
	if err = a.check(); err != nil {
		return s.failed(a, err)
	}
	s.size = 0
	return nil
}
func (s *Conn) WriteFrame(ctx context.Context, b []byte) transport.WriteResult {
	// Validate boundaries only; Diameter semantics belong to the caller's codec.
	if len(b) < 20 || len(b) > 0xffffff || int(binary.BigEndian.Uint32(b[:4])&0xffffff) != len(b) {
		return transport.WriteResult{Err: transport.ErrFrameState}
	}
	a, err := scope(ctx, s.io.SetWriteDeadline, time.Time{})
	if err != nil {
		return transport.WriteResult{Err: err}
	}
	defer a.finish()
	n, err := s.io.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	if err != nil {
		err = s.failed(a, err)
	}
	// TLS reports completed plaintext bytes; an incomplete encrypted record may
	// also have reached the socket. Do not claim definitely-zero delivery on error.
	return transport.WriteResult{Bytes: n, Uncertain: s.tls && err != nil, Err: err}
}

var _ transport.Conn = (*Conn)(nil)
