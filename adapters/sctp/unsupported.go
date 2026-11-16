//go:build !linux || 386

package sctp

import (
	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

// New closes an owned non-nil socket and reports the upstream platform error.
func New(raw *upstream.SCTPConn, cfg Config) (transport.Conn, error) {
	if raw != nil {
		_ = raw.Close()
	}
	_, e := upstream.MessageIORequirements(*SocketConfig().Messages)
	return nil, e
}
