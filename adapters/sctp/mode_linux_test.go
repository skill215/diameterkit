//go:build linux && !386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"unsafe"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

func modeSockets(t *testing.T, client, server Config) (*upstream.SCTPConn, *upstream.SCTPConn) {
	t.Helper()
	sc, e := server.SocketConfig()
	if e != nil {
		t.Fatal(e)
	}
	l, e := sc.Listen("sctp4", &upstream.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}}})
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	cc, e := client.SocketConfig()
	if e != nil {
		t.Fatal(e)
	}
	a, e := cc.DialContext(testContext(t), "sctp4", nil, l.Addr().(*upstream.SCTPAddr))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { a.Close() })
	b, e := l.AcceptSCTP()
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { b.Close() })
	return a, b
}
func rawModeRead(t *testing.T, raw *upstream.SCTPConn, want []byte) upstream.SndRcvInfo {
	t.Helper()
	ctx := testContext(t)
	data, notify, oob := make([]byte, len(want)+8), make([]byte, 4096), make([]byte, 128)
	for {
		r, e := raw.ReadEvent(ctx, data, notify, oob)
		if e != nil {
			t.Fatal(e)
		}
		if r.Kind == upstream.ReadNotification {
			continue
		}
		if r.Kind != upstream.ReadData || !r.InfoPresent || !r.EndOfRecord() || !bytes.Equal(data[:r.N], want) {
			t.Fatal(r)
		}
		return r.Info
	}
}

func TestSCTPModeActualStreamsAndPPID(t *testing.T) {
	for _, policy := range []PPIDPolicy{PPID46, PPID0Compatibility} {
		for _, single := range []bool{false, true} {
			t.Run(fmt.Sprintf("ppid%d/single%v", policy, single), func(t *testing.T) {
				client, server := DefaultConfig(), DefaultConfig()
				client.StreamMode = OrderedMultiStream
				client.PPIDPolicy = policy
				client.OutboundStreams, client.MaxInboundStreams = 5, 2
				server.OutboundStreams, server.MaxInboundStreams = 2, 3
				if single {
					server.MaxInboundStreams = 1
				}
				raw, peer := modeSockets(t, client, server)
				c, e := New(raw, client)
				if e != nil {
					t.Fatal(e)
				}
				defer c.Close()
				info := c.(ModeReporter).SCTPMode()
				out := uint16(3)
				if single {
					out = 1
				}
				if info.OutboundStreams != out || info.InboundStreams != 2 || info.SingleStreamFallback != single {
					t.Fatal(info)
				}
				seen := map[uint16]bool{}
				for j := 0; j < 8; j++ {
					b := modeFrame(fmt.Sprintf("session;%d", j))
					res := c.(transport.MessageWriter).WriteFrameOrdered(testContext(t), b, false)
					if res.Err != nil || res.Bytes != len(b) || res.Uncertain {
						t.Fatal(res)
					}
					r := rawModeRead(t, peer, b)
					expected, _ := sendInfo(client, out, b, false)
					if r.Stream != expected.SID || r.PPID != expected.PPID || r.Flags&upstream.SCTP_UNORDERED != 0 {
						t.Fatal(r, expected)
					}
					seen[r.Stream] = true
				}
				if !single && (!seen[1] || !seen[2]) {
					t.Fatal("did not use local outstream beyond local inbound range", seen)
				}
				// Replies are independent of the sender's stream; receive accepts
				// legal stream1 even when our only local outbound stream is0.
				for _, ppid := range []uint32{46, 0} {
					if ppid == 0 && policy == PPID46 {
						continue
					}
					b := modeFrame("incoming")
					if n, e := peer.WriteMsgContext(testContext(t), b, upstream.SndInfo{SID: 1, PPID: ppid}); e != nil || n != len(b) {
						t.Fatal(n, e)
					}
					got, e := read(testContext(t), c)
					if e != nil || !bytes.Equal(got, b) {
						t.Fatal(e)
					}
				}
				b := modeFrame("barrier")
				if r := c.WriteFrame(testContext(t), b); r.Err != nil {
					t.Fatal(r)
				}
				r := rawModeRead(t, peer, b)
				if r.Stream != 0 || r.Flags&upstream.SCTP_UNORDERED != 0 {
					t.Fatal(r)
				}
			})
		}
	}
}

func TestSCTPModePPIDRejectionAndShortWrite(t *testing.T) {
	for _, tc := range []struct {
		policy PPIDPolicy
		ppid   uint32
	}{{PPID46, 0}, {PPID46, 47}, {PPID0Compatibility, 47}, {PPID0Compatibility, 99}} {
		t.Run(fmt.Sprintf("%d/%d", tc.policy, tc.ppid), func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.PPIDPolicy = tc.policy
			a, b := modeSockets(t, cfg, cfg)
			c, e := New(a, cfg)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = b.WriteMsgContext(testContext(t), modeFrame("bad"), upstream.SndInfo{PPID: tc.ppid}); e != nil {
				t.Fatal(e)
			}
			if _, e = prefix(testContext(t), c); !errors.Is(e, transport.ErrFrameState) {
				t.Fatal(e)
			}
			if r := c.WriteFrame(testContext(t), modeFrame("late")); r.Err == nil {
				t.Fatal(r)
			}
		})
	}
	c, f := scriptedConn()
	c.config.StreamMode = OrderedMultiStream
	c.config.PPIDPolicy = PPID0Compatibility
	c.outbound = 3
	f.written = 7
	r := c.WriteFrameOrdered(testContext(t), modeFrame("short"), false)
	if r.Err == nil || r.Bytes != 7 || f.writes != 1 || f.closed != 1 {
		t.Fatal(r, f)
	}
}

func TestSCTPModeConfigBudgetAndFreeze(t *testing.T) {
	base, e := DefaultConfig().Requirements()
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []StreamMode{UnorderedStream0, OrderedMultiStream} {
		c := DefaultConfig()
		c.StreamMode = mode
		c.PPIDPolicy = PPID0Compatibility
		r, e := c.Requirements()
		if e != nil || r != base {
			t.Fatal(r, e)
		}
		s := DefaultSetupConfig()
		s.Transport = c
		b, e := s.Requirements()
		normal, _ := DefaultSetupConfig().Requirements()
		if e != nil || b != normal {
			t.Fatal(b, e)
		}
		if uint64(unsafe.Sizeof(conn{})+unsafe.Sizeof(ModeInfo{})+unsafe.Sizeof(transport.Event{})+unsafe.Sizeof(upstream.Status{})+unsafe.Sizeof(upstream.MessageOptionState{})) > StateReserve {
			t.Fatal("state reserve")
		}
		socket, e := c.SocketConfig()
		if e != nil {
			t.Fatal(e)
		}
		out := uint16(1)
		if mode == OrderedMultiStream {
			out = 16
		}
		if socket.InitMsg.NumOstreams != out || socket.InitMsg.MaxInstreams != 16 {
			t.Fatal(socket)
		}
		d, e := NewDialer("sctp4", "", "127.0.0.1:3868", s)
		if e != nil {
			t.Fatal(e)
		}
		s.Transport.StreamMode = 99
		if d.cfg.StreamMode != mode || d.cfg.PPIDPolicy != PPID0Compatibility {
			t.Fatal("borrowed setup", d.cfg)
		}
	}
	for _, bad := range []Config{func() Config { c := DefaultConfig(); c.StreamMode = 99; return c }(), func() Config { c := DefaultConfig(); c.PPIDPolicy = 99; return c }()} {
		if _, e := bad.SocketConfig(); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
		if _, e := bad.Requirements(); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
		s := DefaultSetupConfig()
		s.Transport = bad
		if _, e := NewDialer("sctp4", "", "127.0.0.1:3868", s); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
		if _, e := Listen(testContext(t), "sctp4", "127.0.0.1:0", s); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
	}
}

func TestSCTPModeManagedS6cCancelAndShutdown(t *testing.T) {
	for _, mode := range []StreamMode{UnorderedStream0, OrderedMultiStream} {
		for _, policy := range []PPIDPolicy{PPID46, PPID0Compatibility} {
			t.Run(fmt.Sprintf("mode%d/ppid%d", mode, policy), func(t *testing.T) {
				s := DefaultSetupConfig()
				s.Transport.StreamMode = mode
				s.Transport.PPIDPolicy = policy
				s.Transport.OutboundStreams = 4
				s.Transport.MaxInboundStreams = 3
				l, e := Listen(testContext(t), "sctp4", "127.0.0.1:0", s)
				if e != nil {
					t.Fatal(e)
				}
				defer l.Close()
				d, e := NewDialer("sctp4", "", l.Addr().String(), s)
				if e != nil {
					t.Fatal(e)
				}
				a, b := managedConfigs(t)
				n, m := ownNode(t, a), ownNode(t, b)
				q, _ := m.Peer("dra")
				run := startManaged(t, func(ctx context.Context) error { return m.Serve(ctx, "dra", l) })
				p, e := n.Dial(testContext(t), "dra", d)
				if e != nil {
					t.Fatal(e)
				}
				waitFor(t, func() bool { return q.Info().Ready && !q.Info().SendBarrier })
				managedRequest(t, p, q)
				port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 2, MaxOutstandingBytes: 64 << 20})
				if e != nil {
					t.Fatal(e)
				}
				defer port.Close(testContext(t))
				id, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				in, e := q.Receive(testContext(t))
				if e != nil {
					t.Fatal(e)
				}
				if e = n.Cancel(id); e != nil {
					t.Fatal(e)
				}
				completion, e := port.Receive(testContext(t))
				if e != nil || completion.Result.Outcome != diameter.WrittenNoAnswer {
					t.Fatal(completion, e)
				}
				if _, e = in.Respond(testContext(t), success()); e != nil {
					t.Fatal(e)
				}
				managedRequest(t, p, q) // late answer did not consume new exchange
				run.cancel()
				if e = run.wait(t); !errors.Is(e, context.Canceled) {
					t.Fatal(e)
				}
				if e = n.Shutdown(testContext(t)); e != nil {
					t.Fatal(e)
				}
				if e = n.Wait(testContext(t)); e != nil {
					t.Fatal(e)
				}
				m.Close()
				if e = m.Wait(testContext(t)); e != nil {
					t.Fatal(e)
				}
				if p.Info().LastDPAResult != 2001 || n.Resources().Outstanding != 0 || m.Resources().InboundOutstanding != 0 || p.Info().Candidates != 0 || q.Info().Candidates != 0 {
					t.Fatal(p.Info(), q.Info(), n.Resources(), m.Resources())
				}
			})
		}
	}
}

func TestSCTPModeCloseWriteRace(t *testing.T) {
	for range 8 {
		cfg := DefaultConfig()
		cfg.StreamMode = OrderedMultiStream
		cfg.PPIDPolicy = PPID0Compatibility
		a, b := modeSockets(t, cfg, cfg)
		c, e := New(a, cfg)
		if e != nil {
			t.Fatal(e)
		}
		defer b.Close()
		ctx, cancel := context.WithCancel(testContext(t))
		var wg sync.WaitGroup
		var result transport.WriteResult
		wg.Go(func() { result = c.(transport.MessageWriter).WriteFrameOrdered(ctx, modeFrame("race"), false) })
		wg.Go(func() { cancel(); c.Close() })
		wg.Wait()
		if result.Bytes < 0 || result.Bytes > len(modeFrame("race")) || result.Err == nil && result.Bytes != len(modeFrame("race")) {
			t.Fatal(result)
		}
	}
}

func TestSCTPModeResponderBarrier(t *testing.T) {
	for _, mode := range []StreamMode{UnorderedStream0, OrderedMultiStream} {
		for _, policy := range []PPIDPolicy{PPID46, PPID0Compatibility} {
			for _, respond := range []bool{true, false} {
				t.Run(fmt.Sprintf("mode%d/ppid%d/respond%v", mode, policy, respond), func(t *testing.T) {
					cfg := DefaultConfig()
					cfg.StreamMode = mode
					cfg.PPIDPolicy = policy
					testSCTPBarrier(t, cfg, respond)
				})
			}
		}
	}
}

func TestSCTPModeAdmissionBudget(t *testing.T) {
	for _, mode := range []StreamMode{UnorderedStream0, OrderedMultiStream} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			s := DefaultSetupConfig()
			s.Transport.StreamMode = mode
			s.Transport.PPIDPolicy = PPID0Compatibility
			r, e := s.Requirements()
			if e != nil {
				t.Fatal(e)
			}
			a, b := managedConfigs(t)
			a.Resources.AdapterBytesPerConnection = r.ConnectionBufferBytes - 1
			n := ownNode(t, a)
			raw, e := NewDialer("sctp4", "", "127.0.0.1:3868", s)
			if e != nil {
				t.Fatal(e)
			}
			d := &countedDialer{Dialer: raw}
			if _, e = n.Dial(testContext(t), "dra", d); !errors.Is(e, diameter.ErrBackpressure) || d.calls.Load() != 0 {
				t.Fatal(e, d.calls.Load())
			}
			b.Resources.ListenerBytesPerPeer = r.BufferBytes - 1
			m := ownNode(t, b)
			l, e := Listen(testContext(t), "sctp4", "127.0.0.1:0", s)
			if e != nil {
				t.Fatal(e)
			}
			if e = m.Serve(testContext(t), "dra", l); !errors.Is(e, diameter.ErrBackpressure) {
				t.Fatal(e)
			}
			if _, e = l.Accept(testContext(t)); !errors.Is(e, net.ErrClosed) {
				t.Fatal(e)
			}
			p, _ := n.Peer("dra")
			q, _ := m.Peer("dra")
			if p.Info().Candidates != 0 || q.Info().Candidates != 0 {
				t.Fatal(p.Info(), q.Info())
			}
		})
	}
}
