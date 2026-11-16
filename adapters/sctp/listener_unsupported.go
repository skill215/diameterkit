//go:build !linux || 386

package sctp

import (
	"context"
	"github.com/skill215/diameterkit/transport"
	"net"
)

// Listener is unavailable on platforms rejected by the pinned upstream.
type Listener struct{}

func Listen(ctx context.Context, network, address string, cfg SetupConfig) (*Listener, error) {
	_, e := cfg.Requirements()
	return nil, e
}
func (*Listener) Info() transport.ListenerInfo { return transport.ListenerInfo{} }
func (*Listener) Addr() net.Addr               { return nil }
func (*Listener) Close() error                 { return nil }
func (*Listener) Accept(context.Context) (transport.Incoming, error) {
	_, e := DefaultConfig().Requirements()
	return nil, e
}
