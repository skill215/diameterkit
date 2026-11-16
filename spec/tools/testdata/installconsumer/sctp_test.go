//go:build linux && !386

// Copied into an external consumer; never compiled as part of the core module.
package installconsumer_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	adapter "github.com/skill215/diameterkit/adapters/sctp"
	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
	"net"
	"testing"
	"time"
)

func TestInstalledSCTPConsumer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg := adapter.DefaultConfig()
	requirements, err := cfg.Requirements()
	if err != nil || requirements.BufferBytes == 0 {
		t.Fatal("requirements", requirements, err)
	}
	socket, err := cfg.SocketConfig()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := socket.Listen("sctp4", &upstream.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}}})
	if err != nil {
		t.Fatal(err)
	} // Native SCTP must not be skipped.
	defer listener.Close()
	raw, err := socket.DialContext(ctx, "sctp4", nil, listener.Addr().(*upstream.SCTPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	remote, err := listener.AcceptSCTP()
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	a, err := adapter.New(raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	z, err := adapter.New(remote, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	if a.Info().BufferBytes != requirements.BufferBytes {
		t.Fatal("reserve mismatch")
	}
	for _, direction := range [][2]transport.Conn{{a, z}, {z, a}} {
		for _, size := range []int{20, 256, 4096} {
			frame := bytes.Repeat([]byte{7}, size)
			binary.BigEndian.PutUint32(frame[:4], 1<<24|uint32(size))
			if r := direction[0].WriteFrame(ctx, frame); r.Err != nil || r.Bytes != size {
				t.Fatal(r)
			}
			for {
				_, err = direction[1].ReadPrefix(ctx, time.Second)
				var event *transport.Event
				if errors.As(err, &event) && !event.Terminal {
					continue
				}
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			got := make([]byte, size)
			for {
				err = direction[1].ReadFrameInto(ctx, got)
				var event *transport.Event
				if errors.As(err, &event) && !event.Terminal {
					continue
				}
				break
			}
			if err != nil || !bytes.Equal(got, frame) {
				t.Fatal("frame mismatch", err)
			}
		}
	}
}
