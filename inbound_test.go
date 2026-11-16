package diameter

import (
	"context"
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
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func inboundConfig(t *testing.T) NodeConfig {
	cfg := testConfig(t)
	cfg.Resources.NodeBytes = 256 << 20
	cfg.Resources.InboundClasses = []ClassConfig{{ID: "sms", Bytes: 64 << 20, Slots: 4}}
	cfg.Peers[0].MaxInbound = 4
	cfg.Peers[0].MaxInboundBytes = 64 << 20
	cfg.Peers[0].InboundCommands = []InboundPolicy{{Command: pdu.CommandKey{ApplicationID: appID, Code: 8388647, Request: true}, Class: "sms", Sources: []pdu.Origin{serverOrigin}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, ExecutionTimeout: 2 * time.Second, NoCache: true}}
	return cfg
}
func newInboundNode(t *testing.T, cfg NodeConfig) (*Node, *codec.Codec) {
	t.Helper()
	n, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	co, err := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := n.Wait(ctx); err != nil {
			t.Error(err)
		}
		if n.Resources().InboundOutstanding != 0 || n.Resources().InboundBytes != 0 || n.Resources().RejectOutstanding != 0 || n.Resources().RejectBytes != 0 || n.Resources().DuplicateRecords != 0 || n.Resources().DuplicateBytes != 0 {
			t.Error(n.Resources())
		}
	})
	return n, co
}
func inboundRequest(t *testing.T, n *Node, co *codec.Codec, hbh uint32) pdu.Message {
	t.Helper()
	body := requestBody()
	body.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{1, 2, 3}}}
	req, err := co.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationRealm: n.cfg.Origin.Realm, DestinationHost: n.cfg.Origin.Host, HopByHop: hbh, EndToEnd: hbh + 100})
	if err != nil {
		t.Fatal(err)
	}
	return req
}
func inboundFailure() *s6c.SRA {
	return &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}
}
func receiveInbound(t *testing.T, n *Node, p *Peer, remote transport.Conn, co *codec.Codec, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req := inboundRequest(t, n, co, hbh)
	if err := writeWire(ctxTest(t), remote, co, req); err != nil {
		t.Fatal(err)
	}
	in, err := p.Receive(ctxTest(t))
	if err != nil {
		t.Fatal(err)
	}
	return in, req
}
func TestInboundRequestAnswerOwnership(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 11)
		before := n.Resources()
		if before.InboundOutstanding != 1 || before.InboundBytes == 0 || before.Outstanding != 0 {
			t.Fatal(before)
		}
		view := in.Message()
		v := view.Body.(*s6c.SRR)
		v.SessionID = "mutated"
		v.ProxyInfo[0].ProxyState[0] = 99
		view.Header.HopByHop = 123
		r, err := in.Respond(ctxTest(t), inboundFailure())
		if err != nil || !r.Committed || r.PrimaryWrite.State != ResponseWritten {
			t.Fatal(r, err)
		}
		select {
		case <-in.Done():
		default:
			t.Fatal("submission not settled")
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
		if in.Info().ConnectionID != p.Info().ConnectionID || in.Info().RequestID == (RequestID{}) {
			t.Fatal(in.Info())
		}
		if _, err = in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, ErrInboundDone) {
			t.Fatal(err)
		}
		if n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestInboundValidationAndCanceledRespondCanRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 12)
		for _, body := range []pdu.Body{nil, &s6c.SRA{ResultCode: avp.Some(uint32(2001)), AuthSessionState: 1}, &pdu.ErrorAnswer{ResultCode: 5012, ErrorMessage: avp.Some(strings.Repeat("a", 1500))}, &s6c.SRA{SessionID: "wrong", AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}} {
			if r, err := in.Respond(ctxTest(t), body); err == nil || r.Committed {
				t.Fatal(r, err)
			}
			select {
			case <-in.Done():
				t.Fatal("invalid response consumed right")
			default:
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if r, err := in.Respond(ctx, inboundFailure()); !errors.Is(err, context.Canceled) || r.Committed {
			t.Fatal(r, err)
		}
		r, err := in.Respond(ctxTest(t), &pdu.ErrorAnswer{ResultCode: 5012})
		if err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if ans.Header.Flags&pdu.ErrorFlag == 0 {
			t.Fatal(ans)
		}
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
	})
}
func TestInboundReceiveCancellationAndExecutionDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := inboundConfig(t)
		cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 60 * time.Millisecond
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := p.Receive(ctx); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		in, _ := receiveInbound(t, n, p, remote, co, 13)
		select {
		case <-in.Done():
		case <-time.After(time.Second):
			t.Fatal("execution deadline")
		}
		if r, err := in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, context.DeadlineExceeded) || r.Committed {
			t.Fatal(r, err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		// A caller abandoning Receive does not stop reader progress; queued work also expires.
		req := inboundRequest(t, n, co, 14)
		if err := writeWire(ctxTest(t), remote, co, req); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 1 })
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if !p.Info().Ready {
			t.Fatal("timeout poisoned connection")
		}
	})
}
func TestInboundSourceDestinationAndLengthFailClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"origin", "realm", "host", "length"} {
			t.Run(kind, func(t *testing.T) {
				n, co := newInboundNode(t, inboundConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 15)
				v := req.Body.(*s6c.SRR)
				switch kind {
				case "origin":
					v.OriginHost = "rogue.example"
				case "realm":
					v.DestinationRealm = "wrong.example"
				case "host":
					v.DestinationHost = avp.Some("wrong.example")
				case "length":
					v.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}, Data: make([]byte, 1100)}}
				}
				if err := writeWire(ctxTest(t), remote, co, req); err != nil {
					t.Fatal(err)
				}
				waitFor(t, func() bool { return !p.Info().Ready })
				if kind == "origin" && !errors.Is(p.Info().Cause, validation.ErrSource) {
					t.Fatal(p.Info().Cause)
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()
				if _, err := p.Receive(ctx); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
			})
		}
	})
}
func TestInboundBudgetPartitionAndOverflow(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := inboundConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, _ := receiveInbound(t, n, p, remote, co, 16)
		port := portTest(t, n, 1)
		if _, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); err != nil {
			t.Fatal("inbound borrowed outbound partition", err)
		}
		req, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		ans, err := answerFor(co, req)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctxTest(t), remote, co, ans); err != nil {
			t.Fatal(err)
		}
		if _, err = port.Receive(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
		// Saturated inbound admission never waits on the consumer or sends an unbudgeted error.
		if err = writeWire(ctxTest(t), remote, co, inboundRequest(t, n, co, 17)); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return !p.Info().Ready })
		if n.Resources().InboundOutstanding != 1 {
			t.Fatal(n.Resources())
		}
		select {
		case <-in.Done():
			t.Fatal("disconnect ended logical submission")
		default:
		}
		r, err := in.Respond(ctxTest(t), inboundFailure())
		if !r.Committed || r.PrimaryWrite.State != ResponseNotWritten || err == nil {
			t.Fatal(r, err)
		}
	})
}
func TestInboundNoCacheIsExplicitAndDoesNotSuppressArrivals(t *testing.T) {
	cfg := inboundConfig(t)
	cfg.Peers[0].InboundCommands[0].NoCache = false
	if _, err := NewNode(cfg); !errors.Is(err, ErrConfig) {
		t.Fatal(err)
	}
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		a, req := receiveInbound(t, n, p, remote, co, 18)
		if err := writeWire(ctxTest(t), remote, co, req); err != nil {
			t.Fatal(err)
		}
		b, err := p.Receive(ctxTest(t))
		if err != nil {
			t.Fatal(err)
		}
		if a.Info().RequestID == b.Info().RequestID {
			t.Fatal("unimplemented dedup claimed")
		}
		for _, in := range []*Inbound{a, b} {
			if _, err = in.Respond(ctxTest(t), inboundFailure()); err != nil {
				t.Fatal(err)
			}
			if _, err = readWire(ctxTest(t), remote, co); err != nil {
				t.Fatal(err)
			}
		}
	})
}
func TestInboundConcurrentRespondAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, closeNode := range []bool{false, true} {
			n, co := newInboundNode(t, inboundConfig(t))
			p, remote := networkConnectNode(t, n, co, tls)
			in, _ := receiveInbound(t, n, p, remote, co, 19)
			type response struct {
				r   ResponseResult
				err error
			}
			ch := make(chan response, 8)
			var wg sync.WaitGroup
			start := make(chan struct{})
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					r, e := in.Respond(ctxTest(t), inboundFailure())
					ch <- response{r, e}
				}()
			}
			close(start)
			if closeNode {
				_ = n.Close()
			}
			wg.Wait()
			close(ch)
			commits := 0
			for result := range ch {
				if result.r.Committed {
					commits++
				} else if result.err == nil {
					t.Fatal(result)
				}
			}
			if commits > 1 || !closeNode && commits != 1 {
				t.Fatal("commits", commits)
			}
			if !closeNode {
				if _, err := readWire(ctxTest(t), remote, co); err != nil {
					t.Fatal(err)
				}
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestInboundShutdownDrainsResponseBeforeDPR(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 20)
		done := make(chan error, 1)
		go func() { done <- n.Shutdown(ctxTest(t)) }()
		waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
		if r, err := in.Respond(ctxTest(t), inboundFailure()); err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		ans, err := readWire(ctxTest(t), remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
		dpr, err := readWire(ctxTest(t), remote, co)
		if err != nil || dpr.Header.CommandCode != 282 {
			t.Fatal(dpr, err)
		}
		dpa, err := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctxTest(t), remote, co, dpa); err != nil {
			t.Fatal(err)
		}
		if err = <-done; err != nil {
			t.Fatal(err)
		}
		if err = n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
	})
}
func TestInboundPeerCloseAndNodeCloseBoundaries(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		delivered, _ := receiveInbound(t, n, p, remote, co, 21)
		if err := writeWire(ctxTest(t), remote, co, inboundRequest(t, n, co, 22)); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		_ = p.Close()
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 1 })
		if _, err := p.Receive(ctxTest(t)); !errors.Is(err, ErrClosed) {
			t.Fatal(err)
		}
		select {
		case <-delivered.Done():
			t.Fatal("Peer.Close discarded delivered work")
		default:
		}
		_ = n.Close()
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
		select {
		case <-delivered.Done():
		default:
			t.Fatal("Node.Close left live handle")
		}
		if r, err := delivered.Respond(ctxTest(t), inboundFailure()); r.Committed || !errors.Is(err, ErrClosed) {
			t.Fatal(r, err)
		}
	})
}

// The command binding accepts realm-only local routing and keeps the adjacent
// DRA identity separate from the explicitly permitted terminal Origin.
func TestInboundRealmOnlyDestination(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		req := inboundRequest(t, n, co, 23)
		req.Body.(*s6c.SRR).DestinationHost = avp.Optional[string]{}
		if err := writeWire(ctxTest(t), remote, co, req); err != nil {
			t.Fatal(err)
		}
		in, err := p.Receive(ctxTest(t))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = in.Respond(ctxTest(t), inboundFailure()); err != nil {
			t.Fatal(err)
		}
		if _, err = readWire(ctxTest(t), remote, co); err != nil {
			t.Fatal(err)
		}
	})
}

// Fault injection delays return AFTER a real TCP/TLS write. The library must
// preserve the borrow/charge until return, even after ctx cancellation or Close.
type inboundWriteBarrier struct {
	transport.Conn
	request bool
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *inboundWriteBarrier) WriteFrame(ctx context.Context, frame []byte) transport.WriteResult {
	r := b.Conn.WriteFrame(ctx, frame)
	if len(frame) >= 20 && frame[8] == 1 && (frame[4]&pdu.RequestFlag != 0) == b.request {
		b.once.Do(func() { close(b.entered); <-b.release })
	}
	return r
}
func connectInboundTransport(t *testing.T, n *Node, co *codec.Codec, a, b transport.Conn) *Peer {
	t.Helper()
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		req, err := readWire(ctx, b, co)
		if err == nil {
			var ans pdu.Message
			ans, err = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "network-dra", AuthApplicationID: n.cfg.Capabilities.AuthApplicationID, VendorSpecificApplicationID: n.cfg.Capabilities.VendorSpecificApplicationID}, draOrigin)
			if err == nil {
				err = writeWire(ctx, b, co, ans)
			}
		}
		ch <- err
	}()
	p, err := n.Connect(ctx, "dra", a)
	if err != nil {
		t.Fatal(err)
	}
	if err = <-ch; err != nil {
		t.Fatal(err)
	}
	return p
}
func TestInboundCommittedWriteBorrowSurvivesCancellationAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		defer release.Do(func() { close(barrier.release) })
		p := connectInboundTransport(t, n, co, barrier, b)
		in, _ := receiveInbound(t, n, p, b, co, 30)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch := make(chan ResponseResult, 1)
		go func() { r, _ := in.Respond(ctx, inboundFailure()); ch <- r }()
		select {
		case <-barrier.entered:
		case <-time.After(time.Second):
			t.Fatal("writer did not enter")
		}
		cancel()
		_ = n.Close()
		select {
		case <-ch:
			t.Fatal("returned before writer released borrow")
		default:
		}
		if n.Resources().InboundOutstanding != 1 {
			t.Fatal(n.Resources())
		}
		wait, cancelWait := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancelWait()
		if err := n.Wait(wait); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		release.Do(func() { close(barrier.release) })
		r := <-ch
		if !r.Committed || r.PrimaryWrite.State != ResponseWritten || r.PrimaryWrite.Err != nil {
			t.Fatal(r)
		}
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
	})
}
func TestInboundClosingSuppressesQueuedAnswer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
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
		in, _ := receiveInbound(t, n, p, b, co, 31)
		ch := make(chan ResponseResult, 1)
		go func() { r, _ := in.Respond(ctxTest(t), inboundFailure()); ch <- r }()
		select {
		case <-in.Done():
		case <-time.After(time.Second):
			t.Fatal("commit")
		}
		n.mu.Lock()
		n.enterClosingLocked(p.conn, time.Now().Add(time.Second), ErrRemoteDPR)
		n.mu.Unlock()
		r := <-ch
		if !r.Committed || r.PrimaryWrite.State != ResponseNotWritten || !errors.Is(r.PrimaryWrite.Err, ErrRemoteDPR) {
			t.Fatal(r)
		}
		release.Do(func() { close(barrier.release) })
		_ = n.Close()
		if err := n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
		if _, err := port.Receive(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
	})
}
func TestInboundDisconnectRetainsLogicalWorkWithoutSendingOnReconnect(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, _ := receiveInbound(t, n, p, remote, co, 32)
		old := in.Info().ConnectionID
		_ = remote.Close()
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		// New connection enters REOPEN; it must not inherit the old arrival's response.
		p, newRemote := networkConnectNode(t, n, co, tls)
		if p.Info().ConnectionID == old {
			t.Fatal("same generation")
		}
		r, err := in.Respond(ctxTest(t), inboundFailure())
		if !r.Committed || err == nil || r.PrimaryWrite.State != ResponseNotWritten {
			t.Fatal(r, err)
		}
		// Only Base watchdog traffic is permitted here; no old SRA may be injected.
		msg, err := readWire(ctxTest(t), newRemote, co)
		if err != nil || msg.Header.CommandCode != 280 {
			t.Fatal(msg, err)
		}
	})
}
func TestInboundQueueSurvivesTransportDisconnect(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		if err := writeWire(ctxTest(t), remote, co, inboundRequest(t, n, co, 33)); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.inboundQueue) == 1 })
		_ = remote.Close()
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		in, err := p.Receive(ctxTest(t))
		if err != nil {
			t.Fatal(err)
		}
		r, err := in.Respond(ctxTest(t), inboundFailure())
		if !r.Committed || err == nil || r.PrimaryWrite.State != ResponseNotWritten {
			t.Fatal(r, err)
		}
	})
}

func TestInboundS6CResultBranchesAndBaseControl(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := inboundConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		bodies := []pdu.Body{
			&s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"), ServingNode: avp.Some(s6c.ServingNode{MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example"), MMENumber: avp.Some([]byte{0x21, 0x43, 0xf5})})},
			&s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))},
			inboundFailure(), &pdu.ErrorAnswer{ResultCode: 3002},
		}
		for i, body := range bodies {
			in, req := receiveInbound(t, n, p, remote, co, uint32(40+i))
			// With the sole inbound slot occupied, Base DWR/DWA must still make progress.
			dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 99, EndToEnd: 199}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
			if err := writeWire(ctxTest(t), remote, co, dwr); err != nil {
				t.Fatal(err)
			}
			dwa, err := readWire(ctxTest(t), remote, co)
			if err != nil {
				t.Fatal(err)
			}
			if err = co.ValidateAnswer(dwr, dwa, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
				t.Fatal(err)
			}
			r, err := in.Respond(ctxTest(t), body)
			if err != nil || !r.Committed {
				t.Fatal(r, err)
			}
			ans, err := readWire(ctxTest(t), remote, co)
			if err != nil {
				t.Fatal(err)
			}
			if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
				t.Fatal(err)
			}
		}
	})
}
func TestInboundPreparationBorrowAndSinglePreparer(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, action := range []string{"cancel", "close", "deadline"} {
			t.Run(action, func(t *testing.T) {
				cfg := inboundConfig(t)
				if action == "deadline" {
					cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 80 * time.Millisecond
				}
				reqKey := cfg.Peers[0].InboundCommands[0].Command
				ansKey := reqKey
				ansKey.Request = false
				req, _ := cfg.Bundle.Lookup(reqKey)
				ans, _ := cfg.Bundle.Lookup(ansKey)
				original := ans.CompleteAnswer
				entered, release := make(chan struct{}), make(chan struct{})
				var once, releaseOnce sync.Once
				defer releaseOnce.Do(func() { close(release) })
				ans.CompleteAnswer = func(body pdu.Body, env pdu.AnswerEnvelope) (pdu.Body, error) {
					once.Do(func() { close(entered); <-release })
					return original(body, env)
				}
				sp, err := schema.NewProfile("test-inbound-barrier", "1", strings.Repeat("0", 64), req, ans)
				if err != nil {
					t.Fatal(err)
				}
				bp, err := base.Profile()
				if err != nil {
					t.Fatal(err)
				}
				cfg.Bundle, err = schema.Compile(bp, sp)
				if err != nil {
					t.Fatal(err)
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				in, _ := receiveInbound(t, n, p, remote, co, 50)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				type answer struct {
					r ResponseResult
					e error
				}
				ch := make(chan answer, 1)
				go func() { r, e := in.Respond(ctx, inboundFailure()); ch <- answer{r, e} }()
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("preparer")
				}
				if _, err = in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, ErrResponseBusy) {
					t.Fatal(err)
				}
				want := context.Canceled
				switch action {
				case "cancel":
					cancel()
				case "close":
					want = ErrClosed
					_ = n.Close()
				case "deadline":
					want = context.DeadlineExceeded
					select {
					case <-in.Done():
					case <-time.After(time.Second):
						t.Fatal("deadline")
					}
				}
				if n.Resources().InboundOutstanding != 1 {
					t.Fatal("released preparing borrow", n.Resources())
				}
				if action == "close" {
					wait, c := context.WithTimeout(context.Background(), 15*time.Millisecond)
					if err = n.Wait(wait); !errors.Is(err, context.DeadlineExceeded) {
						t.Fatal(err)
					}
					c()
				}
				releaseOnce.Do(func() { close(release) })
				result := <-ch
				if result.r.Committed || !errors.Is(result.e, want) {
					t.Fatal(result)
				}
				if action == "cancel" {
					if _, err = in.Respond(ctxTest(t), inboundFailure()); err != nil {
						t.Fatal(err)
					}
					if _, err = readWire(ctxTest(t), remote, co); err != nil {
						t.Fatal(err)
					}
				} else {
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				}
			})
		}
	})
}

// This writer emits only a prefix through the already established stream. Its
// receiving peer must reserve before the body and close without waiting for it.
func TestInboundPartialFrameCloseAndPrefixBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, oversize := range []bool{false, true} {
			n, co := newInboundNode(t, inboundConfig(t))
			a, raw, _ := networkStream(t, tls)
			ctx := ctxTest(t)
			ch := make(chan error, 1)
			go func() {
				b, e := codec.ReadFrame(raw, 4096)
				if e == nil {
					var req pdu.Message
					req, e = co.Decode(b)
					if e == nil {
						var ans pdu.Message
						ans, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "prefix-peer", AuthApplicationID: []uint32{appID}}, draOrigin)
						if e == nil {
							b, e = co.Append(nil, ans)
							if e == nil {
								_, e = raw.Write(b)
							}
						}
					}
				}
				ch <- e
			}()
			p, err := n.Connect(ctx, "dra", a)
			if err != nil {
				t.Fatal(err)
			}
			if err = <-ch; err != nil {
				t.Fatal(err)
			}
			req := inboundRequest(t, n, co, 60)
			b, err := co.Append(nil, req)
			if err != nil {
				t.Fatal(err)
			}
			if oversize {
				b[1] = 0
				b[2] = 8
				b[3] = 0
			}
			rawSend(t, raw, b[:20])
			if oversize {
				waitFor(t, func() bool { return !p.Info().Ready })
				if !errors.Is(p.Info().Cause, wire.ErrLimit) || n.Resources().InboundOutstanding != 0 {
					t.Fatal(p.Info(), n.Resources())
				}
			} else {
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 1 })
				_ = n.Close()
				if err = n.Wait(ctx); err != nil {
					t.Fatal(err)
				}
				if n.Resources().InboundOutstanding != 0 {
					t.Fatal(n.Resources())
				}
			}
		}
	})
}
