// Package sctp adapts exclusive one-to-one Linux SCTP message sockets to the
// Diameter transport contract. It is an independent, pure-Go module.
package sctp

import (
	"errors"
	"time"

	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

// StateReserve covers the adapter value (including prefix/discard/one event),
// constructor option/status copies and bounded synchronization bookkeeping.
// Variable notification/OOB buffers are added, not hidden in this allowance.
const StateReserve uint64 = 4096

var ErrConfig = errors.New("invalid SCTP adapter configuration or socket options")

type Config struct {
	NotificationMax           int
	NotificationWindow        time.Duration
	MaxNotificationsPerWindow uint32
	StreamMode                StreamMode
	PPIDPolicy                PPIDPolicy
	// Zero selects 1 outbound stream in default mode, 16 in ordered mode,
	// and 16 maximum inbound streams. Actual negotiated counts govern I/O.
	OutboundStreams, MaxInboundStreams uint16
}

func DefaultConfig() Config {
	return Config{NotificationMax: 4096, NotificationWindow: time.Second, MaxNotificationsPerWindow: 128}
}

type StreamMode uint8

const (
	UnorderedStream0 StreamMode = iota
	OrderedMultiStream
)

type PPIDPolicy uint8

const (
	// PPID46 sends and accepts only clear-text Diameter PPID46.
	PPID46 PPIDPolicy = iota
	// PPID0Compatibility sends 0 and accepts 0 or 46. It must be selected
	// explicitly for the configured endpoint; received data never changes it.
	// Neither policy accepts DTLS PPID47.
	PPID0Compatibility
)

// ModeInfo is an immutable association snapshot. A single-stream fallback has
// ordered head-of-line blocking. Even multi-stream mode retains receive level0
// partial-delivery blocking; this is not a latency or session-order guarantee.
type ModeInfo struct {
	StreamMode                      StreamMode
	PPIDPolicy                      PPIDPolicy
	InboundStreams, OutboundStreams uint16
	SingleStreamFallback            bool
}

// ModeReporter is optionally available on the transport.Conn returned by this
// adapter. It exposes negotiated counts without granting raw socket ownership.
type ModeReporter interface{ SCTPMode() ModeInfo }

// SocketConfig returns the default socket configuration. Use Config.SocketConfig
// for explicit modes and stream requests before transferring ownership to New.
func SocketConfig() upstream.SocketConfig {
	c, _ := DefaultConfig().SocketConfig()
	return c
}

func (c Config) SocketConfig() (upstream.SocketConfig, error) {
	if c.StreamMode > OrderedMultiStream || c.PPIDPolicy > PPID0Compatibility {
		return upstream.SocketConfig{}, ErrConfig
	}
	m := upstream.MessageOptions{ReceiveInfo: upstream.ReceiveInfoSNDRCV,
		Events:  upstream.SCTP_EVENT_ASSOCIATION | upstream.SCTP_EVENT_ADDRESS | upstream.SCTP_EVENT_SEND_FAILURE | upstream.SCTP_EVENT_SHUTDOWN | upstream.SCTP_EVENT_PARTIAL_DELIVERY,
		NoDelay: true, AllowFragmentation: true}
	out, in := c.OutboundStreams, c.MaxInboundStreams
	if out == 0 {
		out = 1
		if c.StreamMode == OrderedMultiStream {
			out = 16
		}
	}
	if in == 0 {
		in = 16
	}
	return upstream.SocketConfig{InitMsg: upstream.InitMsg{NumOstreams: out, MaxInstreams: in}, Messages: &m}, nil
}

// Requirements is a no-I/O query to use before NewNode. Control RX/TX and
// emergency slots are reserved separately by Node and must not be counted twice.
// Kernel queues, Go runtime/allocator/stacks and caller configuration are not RSS
// bounded by this logical ledger. Prebuilt sockets are charged from takeover.
func (c Config) Requirements() (transport.Info, error) {
	if _, e := c.SocketConfig(); e != nil {
		return transport.Info{}, e
	}
	setup := DefaultSetupConfig()
	setup.Transport = c
	r, e := upstream.SetupRequirements(setup.upstreamConfig())
	if e != nil {
		return transport.Info{}, configError(e)
	}
	local, e := c.localBytes(r.IORequirements)
	if e != nil {
		return transport.Info{}, e
	}
	return transport.Info{BufferBytes: r.ConnectionBytes + local}, nil
}
func (c Config) localBytes(r upstream.IORequirements) (uint64, error) {
	if _, e := c.SocketConfig(); e != nil {
		return 0, e
	}
	if c.NotificationMax < r.NotificationMin || c.NotificationMax > 1<<20 || c.NotificationWindow <= 0 || c.MaxNotificationsPerWindow == 0 {
		return 0, ErrConfig
	}
	return StateReserve + uint64(c.NotificationMax) + uint64(r.OOBMin), nil
}
