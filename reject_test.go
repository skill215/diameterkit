package diameter

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
)

func rejectConfig(t *testing.T) NodeConfig {
	cfg := inboundConfig(t)
	cfg.Resources.NodeBytes = 512 << 20
	cfg.Resources.TransportBytes = 128 << 20
	cfg.Resources.RejectSlots = 2
	cfg.Resources.EmergencyRequestMax = 1024
	cfg.Resources.EmergencyReplyMax = 1024
	cfg.Peers[0].UnknownRequests = BaseErrorFailures
	cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
	return cfg
}
func sendRejectRaw(t *testing.T, remote transport.Conn, b []byte) {
	t.Helper()
	r := remote.WriteFrame(ctxTest(t), b)
	if r.Err != nil || r.Bytes != len(b) {
		t.Fatal(r)
	}
}
func readReject(t *testing.T, remote transport.Conn, req pdu.Header, code uint32) *pdu.ErrorAnswer {
	t.Helper()
	b, err := readAdapter(ctxTest(t), remote)
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := wire.Header(b, wire.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !pdu.MatchesAnswerHeader(req, h) || h.Flags != req.Flags&pdu.ProxiableFlag|pdu.ErrorFlag {
		t.Fatal("bad rejection header", h, req)
	}
	ops := base.NewErrorOperations(h.Key(), base.ErrorOptions{})
	body, err := ops.Decode(wire.NewDecoder(b[20:], wire.DefaultLimits()))
	if err != nil {
		t.Fatal(err)
	}
	answer := body.(*pdu.ErrorAnswer)
	if answer.ResultCode != code || answer.OriginHost != "client.example" || answer.OriginRealm != "example" || answer.ErrorMessage.Present {
		t.Fatal(answer)
	}
	return answer
}
func reqBytes(t *testing.T, co *codec.Codec, req pdu.Message) []byte {
	t.Helper()
	b, err := co.Append(nil, req)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func withoutAVP(t *testing.T, b []byte, key avp.Key) []byte {
	t.Helper()
	out := append([]byte(nil), b[:20]...)
	d := wire.NewDecoder(b[20:], wire.DefaultLimits())
	for {
		v, ok, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !ok {
			break
		}
		if v.Key == key {
			continue
		}
		start := 20 + v.Offset
		length := (int(wire.Uint24(b[start+5:start+8])) + 3) &^ 3
		out = append(out, b[start:start+length]...)
	}
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func TestRejectPolicyConfiguration(t *testing.T) {
	for _, mutate := range []func(*NodeConfig){
		func(c *NodeConfig) {
			c.Resources.RejectSlots = 0
			c.Resources.EmergencyRequestMax = 0
			c.Resources.EmergencyReplyMax = 0
		},
		func(c *NodeConfig) { c.Resources.RejectSlots = 65 }, func(c *NodeConfig) { c.Resources.EmergencyReplyMax = 19 },
		func(c *NodeConfig) { c.Resources.TransportBytes = 1 }, func(c *NodeConfig) { c.Peers[0].UnknownRequests = FailurePolicy(99) },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].FailurePolicy = FailurePolicy(99) },
	} {
		cfg := rejectConfig(t)
		mutate(&cfg)
		if _, err := NewNode(cfg); !errors.Is(err, ErrConfig) {
			t.Fatal(err)
		}
	}
}
func TestRejectSourceAndDestination(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, which := range []string{"source", "host", "realm"} {
			t.Run(which, func(t *testing.T) {
				n, co := newInboundNode(t, rejectConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 71)
				body := req.Body.(*s6c.SRR)
				switch which {
				case "source":
					body.OriginHost = "rogue.example"
				case "host":
					body.DestinationHost = avp.Some("elsewhere.example")
				case "realm":
					body.DestinationRealm = "elsewhere.example"
				}
				req.Header.Flags |= pdu.RetransmittedFlag
				sendRejectRaw(t, remote, reqBytes(t, co, req))
				ans := readReject(t, remote, req.Header, 3002)
				if !ans.SessionID.Present || ans.SessionID.Value != body.SessionID || !pdu.EqualProxies(ans.ProxyInfo, body.ProxyInfo) {
					t.Fatal("lost envelope")
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				if !p.Info().Ready {
					t.Fatal(p.Info())
				}
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
				if _, err := p.Receive(ctx); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
				cancel()
			})
		}
	})
}
func TestRejectUnsupportedApplicationAndCommand(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, unknownApp := range []bool{false, true} {
			n, co := newInboundNode(t, rejectConfig(t))
			p, remote := networkConnectNode(t, n, co, tls)
			req := inboundRequest(t, n, co, 72)
			raw := reqBytes(t, co, req)
			code := uint32(3001)
			if unknownApp {
				binary.BigEndian.PutUint32(raw[8:12], 123456)
				code = 3007
			} else {
				wire.Put24(raw[5:8], 8388999)
			}
			h, _, _ := wire.Header(raw, n.cfg.Limits)
			sendRejectRaw(t, remote, raw)
			ans := readReject(t, remote, h, code)
			if ans.SessionID.Value != req.Body.(*s6c.SRR).SessionID || len(ans.ProxyInfo) != 1 {
				t.Fatal(ans)
			}
			waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
			if !p.Info().Ready || n.Resources().InboundOutstanding != 0 {
				t.Fatal(p.Info(), n.Resources())
			}
		}
	})
}
func TestRejectOverloadHasIndependentReserve(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		held, _ := receiveInbound(t, n, p, remote, co, 73)
		before := n.Resources()
		req := inboundRequest(t, n, co, 74)
		sendRejectRaw(t, remote, reqBytes(t, co, req))
		readReject(t, remote, req.Header, 3004)
		waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
		after := n.Resources()
		if after.InboundBytes != before.InboundBytes || after.InboundOutstanding != 1 || after.FixedBytes != before.FixedBytes {
			t.Fatal(before, after)
		}
		// Fixed Base controls remain independent of both application and reject usage.
		dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 999, EndToEnd: 1000}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
		if err := writeWire(ctxTest(t), remote, co, dwr); err != nil {
			t.Fatal(err)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil || ans.Header.CommandCode != 280 {
			t.Fatal(ans, err)
		}
		if _, err = held.Respond(ctxTest(t), inboundFailure()); err != nil {
			t.Fatal(err)
		}
		if _, err = readWire(ctxTest(t), remote, co); err != nil {
			t.Fatal(err)
		}
	})
}
func TestRejectOverloadRealmOnlyFailsClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		_, _ = receiveInbound(t, n, p, remote, co, 75)
		req := inboundRequest(t, n, co, 76)
		req.Body.(*s6c.SRR).DestinationHost = avp.Optional[string]{}
		sendRejectRaw(t, remote, reqBytes(t, co, req))
		waitFor(t, func() bool { return !p.Info().Ready })
		if _, err := readAdapter(ctxTest(t), remote); err == nil {
			t.Fatal("illegitimate 3004")
		}
		waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
	})
}
func TestRejectMissingEnvelopeLeaf(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{263, 264, 296, 283, 277} {
			n, co := newInboundNode(t, rejectConfig(t))
			p, remote := networkConnectNode(t, n, co, tls)
			req := inboundRequest(t, n, co, 77)
			raw := withoutAVP(t, reqBytes(t, co, req), avp.Key{Code: code})
			sendRejectRaw(t, remote, raw)
			ans := readReject(t, remote, req.Header, 5005)
			if ans.SessionID.Present != (code != 263) || len(ans.ProxyInfo) != 1 || !ans.FailedAVP.Present {
				t.Fatal(ans)
			}
			d := wire.NewDecoder(ans.FailedAVP.Value.Data, n.cfg.Limits)
			v, ok, err := d.Next()
			if err != nil || !ok || v.Key != (avp.Key{Code: code}) {
				t.Fatal(v, err)
			}
			if !p.Info().Ready {
				t.Fatal(p.Info())
			}
		}
	})
}
func TestRejectUntrustedEnvelopeAndBoundaryClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"duplicate-session", "invalid-session", "bad-proxy", "bad-length"} {
			t.Run(kind, func(t *testing.T) {
				n, co := newInboundNode(t, rejectConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 78)
				raw := reqBytes(t, co, req)
				// Use an unsupported command so only the minimal rejection parser runs.
				wire.Put24(raw[5:8], 8388999)
				d := wire.NewDecoder(raw[20:], n.cfg.Limits)
				for {
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					start := 20 + v.Offset
					if v.Key == (avp.Key{Code: 263}) {
						if kind == "duplicate-session" {
							size := (int(wire.Uint24(raw[start+5:start+8])) + 3) &^ 3
							raw = append(raw, raw[start:start+size]...)
						}
						if kind == "invalid-session" {
							raw[start+8] = 0xff
						}
						if kind == "bad-length" {
							wire.Put24(raw[start+5:start+8], 0xffffff)
						}
					}
					if v.Key == (avp.Key{Code: 284}) && kind == "bad-proxy" {
						binary.BigEndian.PutUint32(raw[start+8:start+12], 99999)
					}
				}
				wire.Put24(raw[1:4], uint32(len(raw)))
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if _, err := readAdapter(ctxTest(t), remote); err == nil {
					t.Fatal("unsafe rejection")
				}
				waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
			})
		}
	})
}
func TestRejectPublicFailuresAndCanceledRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, f := range []ProtocolFailure{FailureContextUnavailable, FailureOverloaded, FailureDeadlineExceeded} {
			in, req := receiveInbound(t, n, p, remote, co, uint32(80+i))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if r, e := in.Reject(ctx, f); r.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(r, e)
			}
			if r, e := in.Reject(ctxTest(t), ProtocolFailure(99)); r.Committed || !errors.Is(e, ErrUnsupported) {
				t.Fatal(r, e)
			}
			r, err := in.Reject(ctxTest(t), f)
			if err != nil || !r.Committed || r.PrimaryWrite.State != ResponseWritten {
				t.Fatal(r, err)
			}
			code := uint32(5012)
			if f == FailureOverloaded {
				code = 3004
			}
			readReject(t, remote, req.Header, code)
			if _, err = in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, ErrInboundDone) {
				t.Fatal(err)
			}
		}
	})
}
func TestRejectDeadlineAutomaticAndQueuedWork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, delivered := range []bool{true, false} {
			cfg := rejectConfig(t)
			cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 60 * time.Millisecond
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, tls)
			req := inboundRequest(t, n, co, 83)
			sendRejectRaw(t, remote, reqBytes(t, co, req))
			var in *Inbound
			if delivered {
				var err error
				in, err = p.Receive(ctxTest(t))
				if err != nil {
					t.Fatal(err)
				}
			}
			ans := readReject(t, remote, req.Header, 5012)
			if ans.SessionID.Value != req.Body.(*s6c.SRR).SessionID {
				t.Fatal(ans)
			}
			if delivered {
				select {
				case <-in.Done():
				default:
					t.Fatal("not decided")
				}
				if _, err := in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
			if !p.Info().Ready {
				t.Fatal(p.Info())
			}
		}
	})
}
func TestRejectRespondCommitRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 84)
		ch := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); r, _ := in.Respond(ctxTest(t), inboundFailure()); ch <- r }()
		go func() { defer wg.Done(); r, _ := in.Reject(ctxTest(t), FailureContextUnavailable); ch <- r }()
		wg.Wait()
		close(ch)
		committed := 0
		for r := range ch {
			if r.Committed {
				committed++
			}
		}
		if committed != 1 {
			t.Fatal(committed)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}

func TestRejectEmergencyExhaustionAndWriteBorrow(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.RejectSlots = 1
		n, co := newInboundNode(t, cfg)
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := connectInboundTransport(t, n, co, barrier, b)
		req := inboundRequest(t, n, co, 90)
		raw := reqBytes(t, co, req)
		wire.Put24(raw[5:8], 8388999)
		h, _, _ := wire.Header(raw, n.cfg.Limits)
		sendRejectRaw(t, b, raw)
		readReject(t, b, h, 3001)
		select {
		case <-barrier.entered:
		case <-time.After(time.Second):
			t.Fatal("no writer")
		}
		if n.Resources().RejectOutstanding != 1 || n.Resources().RejectBytes == 0 || n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
		// The first frame has been written but its I/O borrower has not returned.
		// A second refusal cannot steal that slot or grow an emergency queue.
		h.HopByHop++
		wire.PutHeader(raw, h, len(raw))
		sendRejectRaw(t, b, raw)
		waitFor(t, func() bool { return !p.Info().Ready })
		if !errors.Is(p.Info().Cause, ErrBackpressure) || n.Resources().RejectOutstanding != 1 {
			t.Fatal(p.Info(), n.Resources())
		}
		_ = n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()
		if err := n.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		release.Do(func() { close(barrier.release) })
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
		if n.Resources().RejectOutstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRejectEmergencyLimitsDoNotTruncateEnvelope(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, replyLimit := range []bool{false, true} {
			cfg := rejectConfig(t)
			if replyLimit {
				cfg.Resources.EmergencyReplyMax = 128
			} else {
				cfg.Resources.EmergencyRequestMax = 128
			}
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, tls)
			req := inboundRequest(t, n, co, 91)
			if replyLimit {
				req.Body.(*s6c.SRR).ProxyInfo[0].ProxyState = make([]byte, 300)
			}
			raw := reqBytes(t, co, req)
			wire.Put24(raw[5:8], 8388999)
			sendRejectRaw(t, remote, raw)
			waitFor(t, func() bool { return !p.Info().Ready })
			if _, err := readAdapter(ctxTest(t), remote); err == nil {
				t.Fatal("truncated envelope sent")
			}
			waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
		}
	})
}
func TestRejectTimeoutWinsDuringResponsePreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 80 * time.Millisecond
		reqKey := cfg.Peers[0].InboundCommands[0].Command
		ansKey := reqKey
		ansKey.Request = false
		reqOp, _ := cfg.Bundle.Lookup(reqKey)
		ansOp, _ := cfg.Bundle.Lookup(ansKey)
		original := ansOp.CompleteAnswer
		entered, release := make(chan struct{}), make(chan struct{})
		var once, releaseOnce sync.Once
		defer releaseOnce.Do(func() { close(release) })
		ansOp.CompleteAnswer = func(body pdu.Body, env pdu.AnswerEnvelope) (pdu.Body, error) {
			once.Do(func() { close(entered); <-release })
			return original(body, env)
		}
		profile, err := schema.NewProfile("reject-race", "1", strings.Repeat("0", 64), reqOp, ansOp)
		if err != nil {
			t.Fatal(err)
		}
		bp, err := base.Profile()
		if err != nil {
			t.Fatal(err)
		}
		cfg.Bundle, err = schema.Compile(bp, profile)
		if err != nil {
			t.Fatal(err)
		}
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 92)
		ch := make(chan error, 1)
		go func() {
			r, e := in.Respond(ctxTest(t), inboundFailure())
			if r.Committed {
				e = errors.New("late application won")
			}
			ch <- e
		}()
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("preparer")
		}
		readReject(t, remote, req.Header, 5012)
		if n.Resources().InboundOutstanding != 1 {
			t.Fatal("released before preparation exit", n.Resources())
		}
		waitFor(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			return !in.writing && in.frame == nil
		})
		// A completed write retains its evidence while the losing preparer still
		// borrows the charge; its former write deadline must no longer fire.
		n.mu.Lock()
		oldDeadline := in.writeDeadline
		if n.nextDeadlineLocked().Equal(oldDeadline) {
			t.Error("completed write remains on timer")
		}
		n.mu.Unlock()
		n.tick(oldDeadline.Add(time.Nanosecond))
		n.mu.Lock()
		settled := in.result.PrimaryWrite
		n.mu.Unlock()
		if settled.State != ResponseWritten || settled.Err != nil {
			t.Fatal("settled write relabeled", settled)
		}
		releaseOnce.Do(func() { close(release) })
		if err = <-ch; !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
func TestRejectManualWinnerSuppressesTimeout(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 80 * time.Millisecond
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 93)
		if _, err := in.Respond(ctxTest(t), inboundFailure()); err != nil {
			t.Fatal(err)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
		n.tick(in.Info().ExecutionDeadline.Add(time.Second))
		// Prove no stale timeout frame precedes the next Base response.
		dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 500, EndToEnd: 600}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
		if err = writeWire(ctxTest(t), remote, co, dwr); err != nil {
			t.Fatal(err)
		}
		ans, err = readWire(ctxTest(t), remote, co)
		if err != nil || ans.Header.CommandCode != 280 {
			t.Fatal(ans, err)
		}
	})
}
func TestRejectClosingSuppressesQueuedAutomaticAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, request: true, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := connectInboundTransport(t, n, co, barrier, b)
		port := portTest(t, n, 1)
		if _, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); err != nil {
			t.Fatal(err)
		}
		if _, err := readWire(ctxTest(t), b, co); err != nil {
			t.Fatal(err)
		}
		select {
		case <-barrier.entered:
		case <-time.After(time.Second):
			t.Fatal("writer")
		}
		raw := reqBytes(t, co, inboundRequest(t, n, co, 94))
		wire.Put24(raw[5:8], 8388999)
		sendRejectRaw(t, b, raw)
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.conn.answers) == 1 })
		n.mu.Lock()
		n.enterClosingLocked(p.conn, time.Now().Add(time.Second), ErrRemoteDPR)
		n.mu.Unlock()
		if n.Resources().RejectOutstanding != 0 {
			t.Fatal(n.Resources())
		}
		release.Do(func() { close(barrier.release) })
		_ = n.Close()
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
	})
}
func TestRejectOverloadChecksBindingAndMalformedRequestFirst(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, missing := range []bool{false, true} {
			cfg := rejectConfig(t)
			cfg.Resources.InboundClasses[0].Slots = 1
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, tls)
			_, _ = receiveInbound(t, n, p, remote, co, 95)
			req := inboundRequest(t, n, co, 96)
			code := uint32(3002)
			if !missing {
				req.Body.(*s6c.SRR).OriginHost = "rogue.example"
			}
			raw := reqBytes(t, co, req)
			if missing {
				raw = withoutAVP(t, raw, avp.Key{Code: 263})
				code = 5005
			}
			sendRejectRaw(t, remote, raw)
			readReject(t, remote, req.Header, code)
			if !p.Info().Ready {
				t.Fatal(p.Info())
			}
			waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
		}
	})
}
func TestRejectUnknownBaseCommand(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		raw := reqBytes(t, co, inboundRequest(t, n, co, 97))
		binary.BigEndian.PutUint32(raw[8:12], 0)
		wire.Put24(raw[5:8], 9999)
		h, _, _ := wire.Header(raw, n.cfg.Limits)
		sendRejectRaw(t, remote, raw)
		readReject(t, remote, h, 3001)
		if !p.Info().Ready {
			t.Fatal(p.Info())
		}
	})
}

func TestRejectEmergencyPrefixClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		a, remote, _ := networkStream(t, tls)
		ch := make(chan error, 1)
		go func() {
			b, err := codec.ReadFrame(remote, 4096)
			if err == nil {
				var req, ans pdu.Message
				req, err = co.Decode(b)
				if err == nil {
					ans, err = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "prefix-peer", AuthApplicationID: []uint32{appID}}, draOrigin)
				}
				if err == nil {
					b, err = co.Append(nil, ans)
				}
				if err == nil {
					_, err = remote.Write(b)
				}
			}
			ch <- err
		}()
		if _, err := n.Connect(ctxTest(t), "dra", a); err != nil {
			t.Fatal(err)
		}
		if err := <-ch; err != nil {
			t.Fatal(err)
		}
		raw := reqBytes(t, co, inboundRequest(t, n, co, 98))
		wire.Put24(raw[5:8], 8388999)
		// A complete prefix reserves the emergency slot before waiting for body.
		rawSend(t, remote, raw[:20])
		waitFor(t, func() bool { return n.Resources().RejectOutstanding == 1 })
		if n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
		_ = n.Close()
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
		if n.Resources().RejectOutstanding != 0 || n.Resources().RejectBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestRejectNeverAnswersAnAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		raw := reqBytes(t, co, inboundRequest(t, n, co, 99))
		raw[4] &^= pdu.RequestFlag
		wire.Put24(raw[5:8], 8388999)
		binary.BigEndian.PutUint32(raw[8:12], 9999)
		sendRejectRaw(t, remote, raw)
		dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 501, EndToEnd: 601}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
		if err := writeWire(ctxTest(t), remote, co, dwr); err != nil {
			t.Fatal(err)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil || ans.Header.CommandCode != 280 || !p.Info().Ready || n.Resources().RejectOutstanding != 0 {
			t.Fatal(ans, err, p.Info(), n.Resources())
		}
	})
}
