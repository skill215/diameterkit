// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package streamtransport

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

const ListenerBufferBytes uint64 = 1024

// TCPListener owns its socket and serializes deadline reset with the prior
// cancellation callback. A canceled Accept cannot poison the following call.
type TCPListener struct {
	raw       *net.TCPListener
	mu        sync.Mutex
	closeOnce sync.Once
	closeErr  error
}

func NewTCPListener(raw *net.TCPListener) (*TCPListener, error) {
	if raw == nil {
		return nil, errors.New("nil TCP listener")
	}
	if err := raw.SetDeadline(time.Time{}); err != nil {
		_ = raw.Close()
		return nil, err
	}
	return &TCPListener{raw: raw}, nil
}
func (l *TCPListener) Addr() net.Addr { return l.raw.Addr() }
func (l *TCPListener) Close() error {
	l.closeOnce.Do(func() { l.closeErr = l.raw.Close() })
	return l.closeErr
}
func (l *TCPListener) Accept(ctx context.Context) (*net.TCPConn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline, _ := ctx.Deadline()
	if err := l.raw.SetDeadline(deadline); err != nil {
		return nil, err
	}
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { _ = l.raw.SetDeadline(time.Now()); close(done) })
	raw, err := l.raw.AcceptTCP()
	if !stop() {
		<-done
	}
	if ctx.Err() != nil {
		err = ctx.Err()
	} else if !deadline.IsZero() && !time.Now().Before(deadline) {
		err = context.DeadlineExceeded
	}
	if err != nil && raw != nil {
		_ = raw.Close()
		raw = nil
	}
	return raw, err
}
