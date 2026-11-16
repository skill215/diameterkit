package tcp

import (
	"context"
	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
	"net"
)

const ListenerBufferBytes = streamtransport.ListenerBufferBytes

// Listener wraps a caller-bound TCP listener; NewListener transfers exclusive
// ownership even on failure. Bind addresses and kernel backlog remain external.
type Listener struct{ raw *streamtransport.TCPListener }

func NewListener(raw *net.TCPListener) (*Listener, error) {
	l, err := streamtransport.NewTCPListener(raw)
	if err != nil {
		return nil, err
	}
	return &Listener{l}, nil
}
func (l *Listener) Info() transport.ListenerInfo {
	return transport.ListenerInfo{BufferBytes: ListenerBufferBytes, ConnectionBufferBytes: BufferBytes}
}
func (l *Listener) Addr() net.Addr { return l.raw.Addr() }
func (l *Listener) Close() error   { return l.raw.Close() }
func (l *Listener) Accept(ctx context.Context) (transport.Incoming, error) {
	raw, err := l.raw.Accept(ctx)
	if err != nil {
		return nil, err
	}
	return &incoming{raw}, nil
}

type incoming struct{ raw *net.TCPConn }

func (i *incoming) Info() transport.Info { return transport.Info{BufferBytes: BufferBytes} }
func (i *incoming) Close() error         { return i.raw.Close() }
func (i *incoming) Establish(ctx context.Context) (transport.Conn, error) {
	if err := ctx.Err(); err != nil {
		_ = i.Close()
		return nil, err
	}
	return New(i.raw)
}

var _ transport.Listener = (*Listener)(nil)
