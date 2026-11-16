package tls

import (
	"context"
	"strings"

	"github.com/skill215/diameterkit/internal/streamtransport"
	"github.com/skill215/diameterkit/transport"
)

// Dialer includes the TLS handshake. The caller keeps certificate pools, keys,
// session caches and callback data immutable/concurrency-safe as for Config.
type Dialer struct {
	network, address string
	config           Config
}

func NewDialer(network, address string, cfg Config) (*Dialer, error) {
	if err := streamtransport.ValidateEndpoint(network, address); err != nil {
		return nil, err
	}
	config, err := cloneConfig(cfg)
	if err != nil {
		return nil, err
	}
	cfg.TLS = config
	return &Dialer{strings.Clone(network), strings.Clone(address), cfg}, nil
}
func (d *Dialer) Info() transport.Info { return transport.Info{BufferBytes: BufferBytes} }
func (d *Dialer) Dial(ctx context.Context) (transport.Conn, error) {
	raw, err := streamtransport.DialTCP(ctx, d.network, d.address)
	if err != nil {
		return nil, err
	}
	conn, err := Client(ctx, raw, d.config)
	// Do not return an interface holding a typed nil *Conn on failure.
	if err != nil {
		return nil, err
	}
	return conn, nil
}

var _ transport.Dialer = (*Dialer)(nil)
