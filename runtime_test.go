package diameter

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/testkit"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

const appID uint32 = 16777312

var serverOrigin = pdu.Origin{Host: "hss.example", Realm: "example"}
var draOrigin = pdu.Origin{Host: "dra.example", Realm: "example"}

func testConfig(t *testing.T) NodeConfig {
	t.Helper()
	bp, err := base.Profile()
	if err != nil {
		t.Fatal(err)
	}
	sp, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		t.Fatal(err)
	}
	return NodeConfig{Origin: pdu.Origin{Host: "client.example", Realm: "example"}, Bundle: bundle, Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 2048, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16}, Resources: ResourcePlan{NodeBytes: 128 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096, Classes: []ClassConfig{{"sms", 64 << 20, 16}}, MaxPorts: 8, MaxPortOutstanding: 16, AdapterBytesPerConnection: 4096, MaxDiscardBytes: 4096}, Timeouts: Timeouts{Handshake: time.Second, FrameRead: time.Second, FrameWrite: time.Second, Request: 3 * time.Second, Watchdog: 30 * time.Second, ShutdownTimeout: 3 * time.Second, DPAWait: time.Second, CloseReserve: time.Second, ClosingGrace: time.Second}, Capabilities: LocalCapabilities{HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-test", AuthApplicationID: []uint32{appID}}, Peers: []PeerConfig{{ID: "dra", Origin: draOrigin, MaxPending: 4, MaxBytes: 16 << 20, Commands: []CommandPolicy{{Command: pdu.CommandKey{ApplicationID: appID, Code: 8388647, Request: true}, Class: "sms", DestinationRealm: "example", Sources: pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}, ErrorAgents: []pdu.Origin{draOrigin}}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second}}}}}
}
func newTestNode(t *testing.T) (*Node, *codec.Codec) {
	t.Helper()
	cfg := testConfig(t)
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
	})
	return n, co
}
func ctxTest(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func readMessage(ctx context.Context, s *testkit.Stream, co *codec.Codec) (pdu.Message, error) {
	prefix, err := s.ReadPrefix(ctx, time.Second)
	if err != nil {
		return pdu.Message{}, err
	}
	_, size, err := wire.Header(prefix[:], co.Limits())
	if err != nil {
		return pdu.Message{}, err
	}
	b := make([]byte, size)
	if err = s.ReadFrameInto(ctx, b); err != nil {
		return pdu.Message{}, err
	}
	return co.Decode(b)
}
func writeMessage(ctx context.Context, s *testkit.Stream, co *codec.Codec, m pdu.Message) error {
	b, err := co.Append(nil, m)
	if err != nil {
		return err
	}
	r := s.WriteFrame(ctx, b)
	if r.Err != nil {
		return r.Err
	}
	if r.Bytes != len(b) {
		return io.ErrShortWrite
	}
	return nil
}
func connectTest(t *testing.T, n *Node, co *codec.Codec) (*Peer, *testkit.Stream, *testkit.Stream) {
	t.Helper()
	a, b := testkit.Pipe()
	t.Cleanup(func() { _ = b.Close() })
	ctx := ctxTest(t)
	errs := make(chan error, 1)
	go func() {
		req, err := readMessage(ctx, b, co)
		if err == nil {
			var ans pdu.Message
			ans, err = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake-dra", AuthApplicationID: []uint32{appID}}, draOrigin)
			if err == nil {
				err = writeMessage(ctx, b, co, ans)
			}
		}
		errs <- err
	}()
	p, err := n.Connect(ctx, "dra", a)
	if err != nil {
		t.Fatal(err)
	}
	if err = <-errs; err != nil {
		t.Fatal(err)
	}
	return p, a, b
}
func requestBody() *s6c.SRR {
	return &s6c.SRR{SessionID: "client.example;1;1", AuthSessionState: 1, MSISDN: avp.Some([]byte{0x21, 0x43, 0xf5}), SCAddress: []byte{0x21, 0x43, 0xf5}}
}
func answerFor(co *codec.Codec, req pdu.Message) (pdu.Message, error) {
	return co.CompleteAnswer(req, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, serverOrigin)
}
func portTest(t *testing.T, n *Node, slots int) *CompletionPort {
	t.Helper()
	p, err := n.NewCompletionPort(CompletionLimits{slots, 64 << 20})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := p.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return p
}
func waitFor(t *testing.T, f func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !f() {
		if time.Now().After(deadline) {
			t.Fatal("condition did not settle")
		}
		time.Sleep(time.Millisecond)
	}
}
func TestRuntimeHandshakeSRRSRAAndBudgetTransfer(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	if !p.Info().Ready {
		t.Fatal(p.Info())
	}
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	body := requestBody()
	admit, cancel := context.WithCancel(ctx)
	id, err := p.Submit(admit, body, RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	body.MSISDN.Value[0] = 0xff
	body.SessionID = "changed"
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	if req.Body.(*s6c.SRR).MSISDN.Value[0] != 0x21 {
		t.Fatal("snapshot aliases input")
	}
	if n.Resources().Pending != 1 || n.Resources().OutboundBytes == 0 {
		t.Fatal(n.Resources())
	}
	if _, err = p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrBackpressure) {
		t.Fatal("port did not reserve completion", err)
	}
	answer, err := answerFor(co, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return n.Resources().Pending == 0 })
	s := n.Resources()
	if s.OutboundBytes == 0 || s.Outstanding != 1 {
		t.Fatal("result not charged", s)
	}
	// A completed duplicate does not acquire another A or change the chosen result.
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	completion, err := port.Receive(ctx)
	if err != nil || completion.RequestID != id || completion.Result.Outcome != AnswerReceived || !completion.Result.Delivery.MayHaveReachedPeer {
		t.Fatal(completion, err)
	}
	if s = n.Resources(); s.OutboundBytes != 0 || s.Outstanding != 0 || s.Pending != 0 {
		t.Fatal("leak", s)
	}
}
func TestUnmatchedAndMatchedInvalidAnswers(t *testing.T) {
	for _, kind := range []string{"wrong-header", "wrong-session", "wrong-source", "error-agent", "normal-agent"} {
		t.Run(kind, func(t *testing.T) {
			n, co := newTestNode(t)
			p, _, remote := connectTest(t, n, co)
			port := portTest(t, n, 1)
			ctx := ctxTest(t)
			_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
			if err != nil {
				t.Fatal(err)
			}
			req, err := readMessage(ctx, remote, co)
			if err != nil {
				t.Fatal(err)
			}
			answer, err := answerFor(co, req)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "wrong-header":
				answer.Header.EndToEnd++
			case "wrong-session":
				answer.Body.(*s6c.SRA).SessionID = "wrong"
			case "wrong-source":
				answer.Body.(*s6c.SRA).OriginHost = "evil.example"
			case "normal-agent":
				answer.Body.(*s6c.SRA).OriginHost = draOrigin.Host
			case "error-agent":
				answer, err = co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3002}, draOrigin)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = writeMessage(ctx, remote, co, answer); err != nil {
				t.Fatal(err)
			}
			if kind == "wrong-header" {
				if n.Resources().Pending != 1 {
					t.Fatal("wrong header consumed pending")
				}
				answer, err = answerFor(co, req)
				if err != nil {
					t.Fatal(err)
				}
				if err = writeMessage(ctx, remote, co, answer); err != nil {
					t.Fatal(err)
				}
			}
			got, err := port.Receive(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "wrong-session" || kind == "wrong-source" || kind == "normal-agent" {
				if !errors.Is(got.Result.Err, validation.ErrInvalidAnswer) {
					t.Fatal(got)
				}
			} else if got.Result.Outcome != AnswerReceived {
				t.Fatal(got)
			}
		})
	}
}
func TestCancelQueuedAndInProgress(t *testing.T) {
	n, co := newTestNode(t)
	p, local, remote := connectTest(t, n, co)
	port := portTest(t, n, 2)
	ctx := ctxTest(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	local.Hooks(func(ctx context.Context, _ []byte) *transport.WriteResult {
		once.Do(func() { close(entered) })
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return &transport.WriteResult{Err: ctx.Err()}
		}
	}, nil)
	first, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	second, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	_ = n.Cancel(second)
	_ = n.Cancel(second)
	got, err := port.Receive(ctx)
	if err != nil || got.RequestID != second || got.Result.Outcome != NotWritten || !errors.Is(got.Result.Err, context.Canceled) {
		t.Fatal(got, err)
	}
	_ = n.Cancel(first)
	n.mu.Lock()
	if port.count != 0 || n.requests[first] == nil {
		t.Error("in-flight buffer released before writer exit")
	}
	n.mu.Unlock()
	close(release)
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	got, err = port.Receive(ctx)
	if err != nil || got.RequestID != first || got.Result.Outcome != WrittenNoAnswer || !errors.Is(got.Result.Err, context.Canceled) {
		t.Fatal(got, err)
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}
func TestAnswerWinsBeforeWriterReturnAndClose(t *testing.T) {
	n, co := newTestNode(t)
	p, local, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	wrote := make(chan struct{})
	release := make(chan struct{})
	local.Hooks(nil, func(ctx context.Context, b []byte) *transport.WriteResult {
		close(wrote)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return &transport.WriteResult{Bytes: len(b)}
	})
	id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	<-wrote
	answer, err := answerFor(co, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return n.requests[id].selected })
	_ = n.Cancel(id)
	_ = n.Close()
	close(release)
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != AnswerReceived || got.Result.Err != nil {
		t.Fatal(got, err)
	}
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}
func TestNodeCloseLeavesChargedCompletionsAndWaitBarrier(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	_ = n.Close()
	_ = n.Close()
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if s := n.Resources(); s.Outstanding != 1 || s.OutboundBytes == 0 || s.Pending != 0 {
		t.Fatal(s)
	}
	got, err := port.Receive(ctx)
	if err != nil || got.RequestID != id || got.Result.Outcome != WrittenNoAnswer || !errors.Is(got.Result.Err, ErrClosed) {
		t.Fatal(got, err)
	}
	if _, err = p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
func TestDeadlineAndPortClose(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 2)
	ctx := ctxTest(t)
	id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	n.mu.Lock()
	deadline := n.requests[id].deadline
	n.mu.Unlock()
	n.tick(deadline)
	got, err := port.Receive(ctx)
	if err != nil || !errors.Is(got.Result.Err, context.DeadlineExceeded) {
		t.Fatal(got, err)
	}
	_, err = p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	if err = port.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}

func TestPassiveCEAWriteBarrierAndRemoteDPR(t *testing.T) {
	n, co := newTestNode(t)
	local, remote := testkit.Pipe()
	defer remote.Close()
	ctx := ctxTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	local.Hooks(func(ctx context.Context, _ []byte) *transport.WriteResult {
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return &transport.WriteResult{Err: ctx.Err()}
		}
	}, nil)
	acceptResult := make(chan error, 1)
	go func() { _, err := n.Accept(ctx, "dra", local); acceptResult <- err }()
	cer := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 11, EndToEnd: 12}, Body: &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake", AuthApplicationID: []uint32{appID}}}
	if err := writeMessage(ctx, remote, co, cer); err != nil {
		t.Fatal(err)
	}
	<-entered
	p, _ := n.Peer("dra")
	port := portTest(t, n, 1)
	if p.Info().Ready {
		t.Fatal("Ready before full CEA write")
	}
	if _, err := p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrNotReady) {
		t.Fatal(err)
	}
	close(release)
	cea, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	if err = <-acceptResult; err != nil {
		t.Fatal(err)
	}
	if cea.Body.(*base.CEA).ResultCode != 2001 || !p.Info().Ready {
		t.Fatal(cea, p.Info())
	}
	local.Hooks(nil, nil)
	// Existing fully written pending may still win while responding to DPR.
	_, err = p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	dpr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 282, HopByHop: 21, EndToEnd: 22}, Body: &base.DPR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, DisconnectCause: 0}}
	if err = writeMessage(ctx, remote, co, dpr); err != nil {
		t.Fatal(err)
	}
	dpa, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	if dpa.Header.HopByHop != 21 || p.Info().Ready || p.Info().Protocol != Closing {
		t.Fatal(dpa, p.Info())
	}
	answer, err := answerFor(co, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != AnswerReceived {
		t.Fatal(got, err)
	}
	n.mu.Lock()
	deadline := p.conn.closingDeadline
	n.mu.Unlock()
	n.tick(deadline)
	waitFor(t, func() bool { return p.Info().Protocol == ProtocolClosed })
}
func TestHandshakeRejectsIdentityAndNoCommonApplication(t *testing.T) {
	for _, bad := range []string{"identity", "capability", "wrong-command"} {
		t.Run(bad, func(t *testing.T) {
			n, co := newTestNode(t)
			a, b := testkit.Pipe()
			defer b.Close()
			ctx := ctxTest(t)
			errs := make(chan error, 1)
			go func() {
				req, err := readMessage(ctx, b, co)
				if err == nil {
					body := &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake", AuthApplicationID: []uint32{appID}}
					origin := draOrigin
					if bad == "identity" {
						origin.Host = "evil.example"
					}
					if bad == "capability" {
						body.AuthApplicationID = []uint32{4}
					}
					var answer pdu.Message
					answer, err = co.CompleteAnswer(req, body, origin)
					if bad == "wrong-command" {
						answer = pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
					}
					if err == nil {
						err = writeMessage(ctx, b, co, answer)
					}
				}
				errs <- err
			}()
			if _, err := n.Connect(ctx, "dra", a); err == nil {
				t.Fatal("bad handshake accepted")
			}
			<-errs
			p, _ := n.Peer("dra")
			if p.Info().Ready || p.Info().Protocol != ProtocolClosed {
				t.Fatal(p.Info())
			}
		})
	}
}
func TestWatchdogSingleProbeAndRecovery(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	n.mu.Lock()
	c := p.conn
	deadline := c.tw
	n.mu.Unlock()
	n.tick(deadline)
	dwr, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	if dwr.Header.CommandCode != 280 {
		t.Fatal(dwr)
	}
	dwa, err := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	wrong := dwa
	wrong.Header.EndToEnd++
	if err = writeMessage(ctx, remote, co, wrong); err != nil {
		t.Fatal(err)
	}
	n.mu.Lock()
	if c.probe == nil {
		t.Fatal("wrong DWA retired probe")
	}
	deadline = c.tw
	n.mu.Unlock()
	n.tick(deadline)
	if p.Info().Watchdog != Suspect || p.Info().Ready {
		t.Fatal(p.Info())
	}
	if err = writeMessage(ctx, remote, co, dwa); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return p.Info().Ready })
	n.mu.Lock()
	if c.probe != nil || c.probeQueued {
		t.Error("late writer reinstated probe")
	}
	n.mu.Unlock()
}
func TestConnectionGenerationDoesNotInheritPending(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	oldReq, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	oldConn := p.Info().ConnectionID
	_ = remote.Close()
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != WrittenNoAnswer {
		t.Fatal(got, err)
	}
	waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn == nil })
	p, _, remote = connectTest(t, n, co)
	if p.Info().ConnectionID == oldConn || p.Info().Ready || p.Info().Watchdog != Reopen {
		t.Fatal(p.Info())
	}
	// Reopen needs three matched DWA, never a reused old application answer.
	oldAnswer, err := answerFor(co, oldReq)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, oldAnswer); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if i > 0 {
			n.mu.Lock()
			due := p.conn.tw
			n.mu.Unlock()
			n.tick(due)
		}
		dwr, err := readMessage(ctx, remote, co)
		if err != nil {
			t.Fatal(err)
		}
		dwa, err := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeMessage(ctx, remote, co, dwa); err != nil {
			t.Fatal(err)
		}
		want := i + 1
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn.numDWA == want })
	}
	if !p.Info().Ready || n.Resources().Pending != 0 {
		t.Fatal(p.Info(), n.Resources())
	}
}
func TestShortWritePreservesEvidenceAndClosesConnection(t *testing.T) {
	for _, uncertain := range []bool{false, true} {
		t.Run(map[bool]string{false: "partial", true: "uncertain"}[uncertain], func(t *testing.T) {
			n, co := newTestNode(t)
			p, local, _ := connectTest(t, n, co)
			port := portTest(t, n, 1)
			ctx := ctxTest(t)
			local.Hooks(func(context.Context, []byte) *transport.WriteResult {
				return &transport.WriteResult{Bytes: 7, Uncertain: uncertain, Err: io.ErrUnexpectedEOF}
			}, nil)
			_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
			if err != nil {
				t.Fatal(err)
			}
			got, err := port.Receive(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := PartlyWritten
			if uncertain {
				want = WriteUncertain
			}
			if got.Result.Outcome != want || got.Result.Delivery.Bytes != 7 || !got.Result.Delivery.MayHaveReachedPeer {
				t.Fatal(got)
			}
			waitFor(t, func() bool { return !p.Info().Ready })
			if n.Resources().OutboundBytes != 0 {
				t.Fatal(n.Resources())
			}
		})
	}
}
func TestOversizedAnswerRejectedBeforeReadingBody(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	h := req.Header
	h.Flags &^= pdu.RequestFlag
	prefix := make([]byte, 20)
	wire.PutHeader(prefix, h, 2048)
	r := remote.WriteFrame(ctx, prefix)
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	got, err := port.Receive(ctx)
	if err != nil || !errors.Is(got.Result.Err, wire.ErrLimit) {
		t.Fatal(got, err)
	}
	waitFor(t, func() bool { return !p.Info().Ready })
}
func TestConcurrentCancelAnswerCloseExactlyOnce(t *testing.T) {
	for i := 0; i < 25; i++ {
		n, co := newTestNode(t)
		p, _, remote := connectTest(t, n, co)
		port := portTest(t, n, 1)
		ctx := ctxTest(t)
		id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
		if err != nil {
			t.Fatal(err)
		}
		req, err := readMessage(ctx, remote, co)
		if err != nil {
			t.Fatal(err)
		}
		answer, err := answerFor(co, req)
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() { defer wg.Done(); _ = n.Cancel(id) }()
		go func() { defer wg.Done(); _ = writeMessage(ctx, remote, co, answer) }()
		go func() { defer wg.Done(); _ = n.Close() }()
		wg.Wait()
		got, err := port.Receive(ctx)
		if err != nil || got.RequestID != id {
			t.Fatal(got, err)
		}
		if err = n.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		n.mu.Lock()
		count := port.count
		n.mu.Unlock()
		if count != 0 || n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal("duplicate completion/leak", count, n.Resources())
		}
	}
}
func TestConstructionAdmissionAndIdentifierLimits(t *testing.T) {
	cfg := testConfig(t)
	cfg.Resources.TransportBytes = 1
	if n, err := NewNode(cfg); err == nil {
		n.Close()
		t.Fatal("unfunded transport")
	}
	n, co := newTestNode(t)
	p, _, _ := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := p.Submit(canceled, requestBody(), RequestOptions{}, port); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	body := requestBody()
	body.DestinationRealm = "other.example"
	if _, err := p.Submit(ctx, body, RequestOptions{}, port); !errors.Is(err, validation.ErrEnvelope) {
		t.Fatal(err)
	}
	body = requestBody()
	body.DestinationHost = avp.Some("unbound.example")
	if _, err := p.Submit(ctx, body, RequestOptions{}, port); !errors.Is(err, validation.ErrEnvelope) {
		t.Fatal("realm-only binding bypass", err)
	}
	n.mu.Lock()
	n.appCount = 1 << 31
	n.mu.Unlock()
	if _, err := p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrIdentifiers) {
		t.Fatal(err)
	}
	if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}

func TestControlProgressWithFullCompletionPort(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	answer, err := answerFor(co, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return n.Resources().Pending == 0 })
	before := n.Resources().OutboundBytes
	dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 100, EndToEnd: 101}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
	if err = writeMessage(ctx, remote, co, dwr); err != nil {
		t.Fatal(err)
	}
	dwa, err := readMessage(ctx, remote, co)
	if err != nil || dwa.Header.CommandCode != 280 || dwa.Header.HopByHop != 100 {
		t.Fatal(dwa, err)
	}
	if n.Resources().OutboundBytes != before || n.Resources().Outstanding != 1 {
		t.Fatal("control borrowed business charge", n.Resources())
	}
}
func TestPeerWindowAndBudgetRollback(t *testing.T) {
	cfg := testConfig(t)
	// Exercise the four-slot limit, independently of the growing Profile decode bound.
	cfg.Peers[0].MaxBytes = 64 << 20
	n, co := newInboundNode(t, cfg)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 8)
	ctx := ctxTest(t)
	var ids []RequestID
	for i := 0; i < 4; i++ {
		id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
		if _, err = readMessage(ctx, remote, co); err != nil {
			t.Fatal(err)
		}
	}
	before := n.Resources()
	if _, err := p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrBackpressure) {
		t.Fatal(err)
	}
	if n.Resources() != before {
		t.Fatal("failed admission charged resources")
	}
	for _, id := range ids {
		_ = n.Cancel(id)
	}
	waitFor(t, func() bool { return n.Resources().Pending == 0 })
	// Retired completions keep class/port slots, but release the peer window.
	id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	_ = n.Cancel(id)
	for i := 0; i < 5; i++ {
		if _, err = port.Receive(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}
func TestPortCloseWaitsForBorrowedWriter(t *testing.T) {
	n, co := newTestNode(t)
	p, local, _ := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	local.Hooks(func(ctx context.Context, _ []byte) *transport.WriteResult {
		close(entered)
		select {
		case <-release:
			return &transport.WriteResult{Err: io.EOF}
		case <-ctx.Done():
			return &transport.WriteResult{Err: ctx.Err()}
		}
	}, nil)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	immediate, cancel := context.WithCancel(ctx)
	cancel()
	if err = port.Close(immediate); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if n.Resources().OutboundBytes == 0 || n.Resources().Outstanding != 1 {
		t.Fatal("early release", n.Resources())
	}
	close(release)
	if err = port.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}
func TestSynchronousRequestUsesSameCancellationEvidence(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	call, cancel := context.WithCancel(ctx)
	results := make(chan Result, 1)
	go func() { r, _ := p.Request(call, requestBody(), RequestOptions{}); results <- r }()
	if _, err := readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	cancel()
	r := <-results
	if !errors.Is(r.Err, context.Canceled) || r.Outcome != WrittenNoAnswer {
		t.Fatal(r)
	}
	if n.Resources().Outstanding != 0 || n.Resources().Ports != 0 {
		t.Fatal(n.Resources())
	}
}
func TestWrongHeaderDoesNotAllocateOrDecodeBody(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	h := req.Header
	h.Flags &^= pdu.RequestFlag
	h.EndToEnd++
	raw := make([]byte, 1024)
	wire.PutHeader(raw, h, len(raw)) // invalid AVP body, trustworthy boundary
	before := n.Resources()
	r := remote.WriteFrame(ctx, raw)
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	if n.Resources() != before || !p.Info().Ready {
		t.Fatal(n.Resources(), p.Info())
	}
	_ = n.Cancel(id)
	if _, err = port.Receive(ctx); err != nil {
		t.Fatal(err)
	}
}
func TestMalformedMatchedHeaderCompletesInvalid(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	h := req.Header
	h.Flags &^= pdu.RequestFlag
	h.Flags |= pdu.RetransmittedFlag
	raw := make([]byte, 20)
	wire.PutHeader(raw, h, 20)
	_ = remote.WriteFrame(ctx, raw)
	got, err := port.Receive(ctx)
	if err != nil || !errors.Is(got.Result.Err, validation.ErrInvalidAnswer) {
		t.Fatal(got, err)
	}
}

func TestOversizedRequestRejectedBeforeAdmission(t *testing.T) {
	n, co := newTestNode(t)
	p, _, _ := connectTest(t, n, co)
	port := portTest(t, n, 1)
	body := requestBody()
	body.Extra = []avp.Raw{{Key: avp.Key{Code: 999999}, Data: make([]byte, 1500)}}
	n.mu.Lock()
	before := n.appCount
	n.mu.Unlock()
	if _, err := p.Submit(ctxTest(t), body, RequestOptions{}, port); !errors.Is(err, wire.ErrLimit) {
		t.Fatal(err)
	}
	n.mu.Lock()
	after := n.appCount
	n.mu.Unlock()
	if before != after || n.Resources().OutboundBytes != 0 {
		t.Fatal("oversize consumed admission", n.Resources())
	}
}
func TestStoppedNodeReleasesFixedReserveAfterPortClose(t *testing.T) {
	n, _ := newTestNode(t)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_ = n.Close()
	if err := n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if n.Resources().FixedBytes == 0 {
		t.Fatal("port ring prematurely released")
	}
	if err := port.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if n.Resources().FixedBytes != 0 {
		t.Fatal(n.Resources())
	}
}
