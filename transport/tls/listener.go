package tls

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
	"net"
)

var errServerCertificate = errors.New("TLS listener requires a server certificate or callback")

const ListenerBufferBytes = streamtransport.ListenerBufferBytes

// NewListener takes ownership of raw, including on failure, and freezes the
// top-level TLS configuration. Shared certificate/callback data follows Config.
// TLS runs only in Incoming.Establish, after socket admission by the caller.
type Listener struct {
	raw    *streamtransport.TCPListener
	config Config
}

func NewListener(raw *net.TCPListener, cfg Config) (*Listener, error) {
	config, err := cloneConfig(cfg)
	if err == nil && len(config.Certificates) == 0 && config.GetCertificate == nil && config.GetConfigForClient == nil {
		err = errServerCertificate
	}
	if err != nil {
		if raw != nil {
			_ = raw.Close()
		}
		return nil, err
	}
	l, err := streamtransport.NewTCPListener(raw)
	if err != nil {
		return nil, err
	}
	cfg.TLS = config
	return &Listener{l, cfg}, nil
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
	return &incoming{raw, l.config}, nil
}

type incoming struct {
	raw    *net.TCPConn
	config Config
}

func (i *incoming) Info() transport.Info { return transport.Info{BufferBytes: BufferBytes} }
func (i *incoming) Close() error         { return i.raw.Close() }
func (i *incoming) Establish(ctx context.Context) (transport.Conn, error) {
	c, err := Server(ctx, i.raw, i.config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

var _ transport.Listener = (*Listener)(nil)
