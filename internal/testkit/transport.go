// Package testkit provides a bounded stream fake. It is never a production transport.
package testkit

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

type WriteHook func(context.Context, []byte) *transport.WriteResult

// Stream uses net.Pipe for deterministic backpressure and real borrowed buffers.
type Stream struct {
	conn          net.Conn
	prefix        [20]byte
	size          int
	deadline      time.Time
	mu            sync.Mutex
	before, after WriteHook
}

func Pipe() (*Stream, *Stream)         { a, b := net.Pipe(); return &Stream{conn: a}, &Stream{conn: b} }
func (s *Stream) Info() transport.Info { return transport.Info{BufferBytes: 4096} }
func (s *Stream) Close() error         { return s.conn.Close() }
func (s *Stream) Hooks(before, after WriteHook) {
	s.mu.Lock()
	s.before = before
	s.after = after
	s.mu.Unlock()
}
func minDeadline(a, b time.Time) time.Time {
	if a.IsZero() || !b.IsZero() && b.Before(a) {
		return b
	}
	return a
}

// stop waits for an in-flight cancellation callback before clearing its deadline.
// A previous operation can never reset the next operation's deadline.
func (s *Stream) scope(ctx context.Context, read bool, deadline time.Time) func() {
	set := s.conn.SetWriteDeadline
	if read {
		set = s.conn.SetReadDeadline
	}
	if d, ok := ctx.Deadline(); ok {
		deadline = minDeadline(deadline, d)
	}
	_ = set(deadline)
	finished := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { _ = set(time.Now()); close(finished) })
	return func() {
		if !stop() {
			<-finished
		}
		_ = set(time.Time{})
	}
}
func (s *Stream) ReadPrefix(ctx context.Context, timeout time.Duration) ([20]byte, error) {
	if s.size != 0 {
		return [20]byte{}, errors.New("unfinished prefix")
	}
	done := s.scope(ctx, true, time.Time{})
	defer done()
	if _, err := io.ReadFull(s.conn, s.prefix[:1]); err != nil {
		return [20]byte{}, err
	}
	s.deadline = time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok {
		s.deadline = minDeadline(s.deadline, d)
	}
	_ = s.conn.SetReadDeadline(s.deadline)
	if _, err := io.ReadFull(s.conn, s.prefix[1:]); err != nil {
		return [20]byte{}, err
	}
	s.size = int(binary.BigEndian.Uint32(s.prefix[:4]) & 0xffffff)
	return s.prefix, nil
}
func (s *Stream) ReadFrameInto(ctx context.Context, dst []byte) error {
	if s.size < 20 || len(dst) != s.size {
		return errors.New("frame target mismatch")
	}
	done := s.scope(ctx, true, s.deadline)
	defer done()
	copy(dst, s.prefix[:])
	_, err := io.ReadFull(s.conn, dst[20:])
	if err == nil {
		s.size = 0
	}
	return err
}
func (s *Stream) DiscardFrame(ctx context.Context, n int) error {
	if n != s.size || n < 20 {
		return errors.New("discard boundary mismatch")
	}
	done := s.scope(ctx, true, s.deadline)
	defer done()
	var scratch [256]byte
	for remain := n - 20; remain > 0; {
		take := min(remain, len(scratch))
		if _, err := io.ReadFull(s.conn, scratch[:take]); err != nil {
			return err
		}
		remain -= take
	}
	s.size = 0
	return nil
}
func (s *Stream) WriteFrame(ctx context.Context, b []byte) transport.WriteResult {
	s.mu.Lock()
	before, after := s.before, s.after
	s.mu.Unlock()
	if before != nil {
		if r := before(ctx, b); r != nil {
			return *r
		}
	}
	done := s.scope(ctx, false, time.Time{})
	defer done()
	n, err := s.conn.Write(b)
	result := transport.WriteResult{Bytes: n, Err: err}
	if after != nil {
		if r := after(ctx, b); r != nil {
			return *r
		}
	}
	return result
}
