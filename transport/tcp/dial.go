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
