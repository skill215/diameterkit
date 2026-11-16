// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/testkit"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
)

func TestLocalShutdownDrainsAnswerThenDPR(t *testing.T) {
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
	done := make(chan error, 1)
	go func() { done <- p.Shutdown(ctx) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	if p.Info().Protocol != IOpen || p.Info().Ready {
		t.Fatal(p.Info())
	}
	if _, err = p.Submit(ctx, requestBody(), RequestOptions{}, port); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	answer, err := answerFor(co, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, answer); err != nil {
		t.Fatal(err)
	}
	dpr, err := readMessage(ctx, remote, co)
	if err != nil || dpr.Header.CommandCode != 282 {
		t.Fatal(dpr, err)
	}
	dpa, err := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, dpa); err != nil {
		t.Fatal(err)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if p.Info().Lifecycle != Closed || p.Info().LastDPAResult != 2001 {
		t.Fatal(p.Info())
	}
	// Peer cleanup does not consume the application-owned completion port.
	if n.Resources().Outstanding != 1 {
		t.Fatal(n.Resources())
	}
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != AnswerReceived {
		t.Fatal(got, err)
	}
}
func TestDrainAllowsOnlyAdmittedFirstWrites(t *testing.T) {
	n, co := newTestNode(t)
	p, local, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	port := portTest(t, n, 2)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	local.Hooks(func(ctx context.Context, _ []byte) *transport.WriteResult {
		once.Do(func() {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
			}
		})
		return nil
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
	done := make(chan error, 1)
	go func() { done <- p.Shutdown(ctx) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	close(release)
	for range 2 {
		req, err := readMessage(ctx, remote, co)
		if err != nil || req.Header.CommandCode != 8388647 {
			t.Fatal(req, err)
		}
		ans, err := answerFor(co, req)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeMessage(ctx, remote, co, ans); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []RequestID{first, second} {
		got, err := port.Receive(ctx)
		if err != nil || got.RequestID != id || got.Result.Outcome != AnswerReceived {
			t.Fatal(got, err)
		}
	}
	dpr, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	dpa, err := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	_ = writeMessage(ctx, remote, co, dpa)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
}
func TestDrainDeadlineRevokesQueuedWriteWithoutInterleavingDPR(t *testing.T) {
	n, co := newTestNode(t)
	p, local, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	port := portTest(t, n, 2)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	local.Hooks(func(ctx context.Context, _ []byte) *transport.WriteResult {
		once.Do(func() {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
			}
		})
		return nil
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
	done := make(chan error, 1)
	go func() { done <- p.Shutdown(ctx) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	n.tick(p.Info().DrainDeadline)
	got, err := port.Receive(ctx)
	if err != nil || got.RequestID != second || got.Result.Outcome != NotWritten || !errors.Is(got.Result.Err, ErrShutdown) {
		t.Fatal(got, err)
	}
	close(release)
	req, err := readMessage(ctx, remote, co)
	if err != nil || req.Header.CommandCode != 8388647 {
		t.Fatal(req, err)
	}
	dpr, err := readMessage(ctx, remote, co)
	if err != nil || dpr.Header.CommandCode != 282 {
		t.Fatal(dpr, err)
	}
	dpa, err := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	_ = writeMessage(ctx, remote, co, dpa)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	got, err = port.Receive(ctx)
	if err != nil || got.RequestID != first || got.Result.Outcome != WrittenNoAnswer {
		t.Fatal(got, err)
	}
}
func TestFastErrorDPAClosesBeforeDPRWriterReturns(t *testing.T) {
	n, co := newTestNode(t)
	p, local, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	local.Hooks(nil, func(ctx context.Context, b []byte) *transport.WriteResult {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return &transport.WriteResult{Bytes: len(b)}
	})
	done := make(chan error, 1)
	go func() { done <- p.Shutdown(ctx) }()
	dpr, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	dpa, err := co.CompleteAnswer(dpr, &pdu.ErrorAnswer{ResultCode: 5012}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, dpa); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return p.Info().Protocol == ProtocolClosed })
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if p.Info().LastDPAResult != 5012 || p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestCrossedDPRAndUnmatchedControlDoNotRefreshDeadline(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	done := make(chan error, 1)
	go func() { done <- p.Shutdown(ctx) }()
	localDPR, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	n.mu.Lock()
	c := p.conn
	original := c.closingDeadline
	n.mu.Unlock()
	wrong, err := co.CompleteAnswer(localDPR, &base.DPA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	wrong.Header.HopByHop++
	if err = writeMessage(ctx, remote, co, wrong); err != nil {
		t.Fatal(err)
	}
	dpr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 282, HopByHop: 61, EndToEnd: 62}, Body: &base.DPR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, DisconnectCause: 0}}
	if err = writeMessage(ctx, remote, co, dpr); err != nil {
		t.Fatal(err)
	}
	reply, err := readMessage(ctx, remote, co)
	if err != nil || reply.Header.HopByHop != 61 {
		t.Fatal(reply, err)
	}
	n.mu.Lock()
	if c.closingDeadline.After(original) || c.disconnect == nil {
		t.Error("cross DPR refreshed deadline/cleared local match")
	}
	n.mu.Unlock()
	// Duplicate DPR must not enqueue a second DPA.
	if err = writeMessage(ctx, remote, co, dpr); err != nil {
		t.Fatal(err)
	}
	good, err := co.CompleteAnswer(localDPR, &base.DPA{ResultCode: 2001}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, good); err != nil {
		t.Fatal(err)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if p.Info().LastDPAResult != 2001 {
		t.Fatal(p.Info())
	}
}
func TestShutdownCancellationAndRepeatedDeadlines(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	port := portTest(t, n, 1)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	call, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- n.Shutdown(call) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	first := p.Info().ShutdownDeadline
	secondDone := make(chan error, 1)
	go func() { secondDone <- p.Shutdown(context.Background()) }()
	waitFor(t, func() bool { return !p.Info().ShutdownDeadline.After(first) })
	cancel()
	if err = <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	<-secondDone
	if p.Info().Lifecycle != Closed || p.Info().Ready {
		t.Fatal(p.Info())
	}
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != WrittenNoAnswer {
		t.Fatal(got, err)
	}
}
func TestShutdownWaitingForDPAHasBoundedExit(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	ctx := ctxTest(t)
	done := make(chan error, 1)
	go func() { done <- n.Shutdown(ctx) }()
	if _, err := readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	n.mu.Lock()
	c := p.conn
	deadline := c.closingDeadline
	n.mu.Unlock()
	n.tick(deadline)
	if err := <-done; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if err := n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}
func TestImmediateCloseUpgradesShutdown(t *testing.T) {
	n, co := newTestNode(t)
	p, _, _ := connectTest(t, n, co)
	ctx := ctxTest(t)
	done := make(chan error, 1)
	go func() { done <- n.Shutdown(ctx) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	_ = n.Close()
	if err := n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	<-done
	if p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}

func TestShutdownTotalDeadlineStopsPartialFrameBeforeDPR(t *testing.T) {
	n, co := newTestNode(t)
	p, _, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	ctx := ctxTest(t)
	_, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = remote.ReadPrefix(ctx, n.cfg.Timeouts.FrameRead); err != nil {
		t.Fatal(err)
	} // writer still borrows body
	done := make(chan error, 1)
	go func() { done <- n.Shutdown(ctx) }()
	waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
	n.tick(p.Info().ShutdownDeadline)
	if err = <-done; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != PartlyWritten || got.Result.Delivery.Bytes != 20 {
		t.Fatal(got, err)
	}
	if n.Resources().OutboundBytes != 0 || p.Info().Candidates != 0 {
		t.Fatal(n.Resources(), p.Info())
	}
}
func TestPeerShutdownDoesNotCloseOtherPeerAndNodeSharesDeadline(t *testing.T) {
	cfg := testConfig(t)
	second := cfg.Peers[0]
	second.ID = "other"
	cfg.Peers = append(cfg.Peers, second)
	cfg.Resources.TransportBytes = 64 << 20
	cfg.Resources.NodeBytes = 256 << 20
	n, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	co, err := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ctxTest(t)
	defer func() { _ = n.Close(); _ = n.Wait(ctx) }()
	first, _, remote1 := connectTest(t, n, co)
	a, b := testkit.Pipe()
	defer b.Close()
	hs := make(chan error, 1)
	go func() {
		req, e := readMessage(ctx, b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(req, n.cea(2001), draOrigin)
			if e == nil {
				e = writeMessage(ctx, b, co, ans)
			}
		}
		hs <- e
	}()
	other, err := n.Connect(ctx, "other", a)
	if err != nil {
		t.Fatal(err)
	}
	if err = <-hs; err != nil {
		t.Fatal(err)
	}
	port := portTest(t, n, 2)
	for _, p := range []*Peer{first, other} {
		if _, err = p.Submit(ctx, requestBody(), RequestOptions{}, port); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = readMessage(ctx, remote1, co); err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, b, co); err != nil {
		t.Fatal(err)
	}
	peerDone := make(chan error, 1)
	go func() { peerDone <- first.Shutdown(ctx) }()
	waitFor(t, func() bool { return first.Info().Lifecycle == Draining })
	if !other.Info().Ready {
		t.Fatal("peer shutdown affected neighbor")
	}
	nodeDone := make(chan error, 1)
	go func() { nodeDone <- n.Shutdown(ctx) }()
	waitFor(t, func() bool { return other.Info().Lifecycle == Draining })
	n.mu.Lock()
	total := n.shutdownDeadline
	n.mu.Unlock()
	if first.Info().ShutdownDeadline.After(total) || other.Info().ShutdownDeadline.After(total) {
		t.Fatal("per-peer deadline extension")
	}
	n.tick(total)
	<-nodeDone
	<-peerDone
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = port.Receive(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if n.Resources().OutboundBytes != 0 {
		t.Fatal(n.Resources())
	}
}
func TestShutdownCandidateCEACannotPublishReady(t *testing.T) {
	n, co := newTestNode(t)
	a, b := testkit.Pipe()
	defer b.Close()
	ctx := ctxTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	a.Hooks(nil, func(ctx context.Context, frame []byte) *transport.WriteResult {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return &transport.WriteResult{Bytes: len(frame)}
	})
	accept := make(chan error, 1)
	go func() { _, err := n.Accept(ctx, "dra", a); accept <- err }()
	if err := writeMessage(ctx, b, co, peerCER(n)); err != nil {
		t.Fatal(err)
	}
	if _, err := readMessage(ctx, b, co); err != nil {
		t.Fatal(err)
	}
	<-entered
	if err := n.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-accept; err == nil {
		t.Fatal("installed candidate after shutdown")
	}
	p, _ := n.Peer("dra")
	if p.Info().Ready || p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}

func TestShutdownReturnsAtConfiguredBoundWhileWaitTracksBorrower(t *testing.T) {
	cfg := testConfig(t)
	cfg.Timeouts.ShutdownTimeout = 60 * time.Millisecond
	cfg.Timeouts.CloseReserve = 30 * time.Millisecond
	cfg.Timeouts.ClosingGrace = 20 * time.Millisecond
	cfg.Timeouts.DPAWait = 20 * time.Millisecond
	n, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	co, err := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ctxTest(t)
	p, local, remote := connectTest(t, n, co)
	port := portTest(t, n, 1)
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	defer func() {
		releaseOnce.Do(func() { close(release) })
		_ = n.Close()
		if err := n.Wait(ctx); err != nil {
			t.Error(err)
		}
	}()
	// Fault injection: the bytes are already written, but the adapter delays its
	// return (and thus its borrow retirement) even after Close. Never release early.
	local.Hooks(nil, func(_ context.Context, b []byte) *transport.WriteResult {
		close(entered)
		<-release
		return &transport.WriteResult{Bytes: len(b)}
	})
	if _, err = p.Submit(ctx, requestBody(), RequestOptions{}, port); err != nil {
		t.Fatal(err)
	}
	if _, err = readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	<-entered
	done := make(chan error, 1)
	go func() { done <- n.Shutdown(context.Background()) }()
	select {
	case err = <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("configured shutdown bound ignored")
	}
	n.mu.Lock()
	stillOwned := n.total.UsedSlots == 1 && len(n.requests) == 1
	n.mu.Unlock()
	if !stillOwned {
		t.Fatal("borrowed request released")
	}
	select {
	case <-n.done:
		t.Fatal("Wait barrier passed live borrower")
	default:
	}
	releaseOnce.Do(func() { close(release) })
	if err = n.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := port.Receive(ctx)
	if err != nil || got.Result.Outcome != WrittenNoAnswer {
		t.Fatal(got, err)
	}
}
