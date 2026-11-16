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
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
	"io"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func testContext(t testing.TB) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func sockets(t testing.TB) (*upstream.SCTPConn, *upstream.SCTPConn) {
	t.Helper()
	cfg := SocketConfig()
	cfg.InitMsg.NumOstreams = 4
	l, e := cfg.Listen("sctp4", &upstream.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}}})
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	c, e := cfg.DialContext(testContext(t), "sctp4", nil, l.Addr().(*upstream.SCTPAddr))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { c.Close() })
	s, e := l.AcceptSCTP()
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	return c, s
}
func adapters(t testing.TB) (transport.Conn, transport.Conn) {
	a, b := sockets(t)
	x, e := New(a, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	y, e := New(b, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	return x, y
}
func data(size int) []byte {
	b := bytes.Repeat([]byte{9}, size)
	binary.BigEndian.PutUint32(b[:4], 1<<24|uint32(size))
	return b
}
func prefix(ctx context.Context, c transport.Conn) ([20]byte, error) {
	for {
		p, e := c.ReadPrefix(ctx, time.Second)
		var v *transport.Event
		if errors.As(e, &v) && !v.Terminal {
			continue
		}
		return p, e
	}
}
func read(ctx context.Context, c transport.Conn) ([]byte, error) {
	p, e := prefix(ctx, c)
	if e != nil {
		return nil, e
	}
	b := make([]byte, int(binary.BigEndian.Uint32(p[:4])&0xffffff))
	for {
		e = c.ReadFrameInto(ctx, b)
		var v *transport.Event
		if errors.As(e, &v) && !v.Terminal {
			continue
		}
		return b, e
	}
}
func TestRealSCTPFramesDiscardAndMetadata(t *testing.T) {
	raw, peer := sockets(t)
	c, e := New(raw, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	ctx := testContext(t)
	for i, size := range []int{20, 47, 8192, 100003} {
		b := data(size)
		// Incoming stream1 is valid, independent of our outgoing stream0.
		n, e := peer.WriteMsgContext(ctx, b, upstream.SndInfo{SID: 1, PPID: 46})
		if e != nil || n != size {
			t.Fatal(n, e)
		}
		if i == 2 {
			if _, e = prefix(ctx, c); e != nil {
				t.Fatal(e)
			}
			if e = c.DiscardFrame(ctx, size); e != nil {
				t.Fatal(e)
			}
		} else {
			got, e := read(ctx, c)
			if e != nil || !bytes.Equal(b, got) {
				t.Fatal(size, e)
			}
		}
	}
	b := data(40)
	r := c.(transport.MessageWriter).WriteFrameOrdered(ctx, b, false)
	if r.Err != nil || r.Bytes != len(b) || r.Uncertain {
		t.Fatal(r)
	}
	ledger, _ := upstream.SetupRequirements(DefaultSetupConfig().upstreamConfig())
	req := ledger.IORequirements
	d := make([]byte, 128)
	notify := make([]byte, 4096)
	oob := make([]byte, req.OOBMin)
	for {
		v, e := peer.ReadEvent(ctx, d, notify, oob)
		if e != nil {
			t.Fatal(e)
		}
		if v.Kind == upstream.ReadNotification {
			continue
		}
		if !v.InfoPresent || v.Info.Stream != 0 || v.Info.PPID != 46 || v.Info.Context != 0 || v.Info.Flags&upstream.SCTP_UNORDERED == 0 || v.RecvFlags&syscall.MSG_EOR == 0 || !bytes.Equal(d[:v.N], b) {
			t.Fatal(v)
		}
		break
	}
}
func TestRealSCTPInvalidMessageBoundary(t *testing.T) {
	for _, kind := range []string{"short", "early_eor", "missing_eor", "ppid", "two_frames"} {
		t.Run(kind, func(t *testing.T) {
			a, b := sockets(t)
			c, e := New(a, DefaultConfig())
			if e != nil {
				t.Fatal(e)
			}
			msg := data(32)
			ppid := uint32(46)
			switch kind {
			case "short":
				msg = msg[:19]
			case "early_eor":
				msg = msg[:25]
			case "missing_eor":
				msg = append(msg, 1)
			case "ppid":
				ppid = 47
			case "two_frames":
				msg = append(msg, data(20)...)
			}
			if _, e = b.WriteMsgContext(testContext(t), msg, upstream.SndInfo{PPID: ppid}); e != nil {
				t.Fatal(e)
			}
			if _, e = read(testContext(t), c); !errors.Is(e, transport.ErrFrameState) {
				t.Fatal(e)
			}
			if r := c.WriteFrame(testContext(t), data(20)); !errors.Is(r.Err, net.ErrClosed) {
				t.Fatal(r)
			}
		})
	}
}
func TestRealSCTPReadCancellationCloseAndDeadline(t *testing.T) {
	for _, action := range []string{"cancel", "close", "deadline"} {
		t.Run(action, func(t *testing.T) {
			c, _ := adapters(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if action == "deadline" {
				ctx, cancel = context.WithTimeout(ctx, 30*time.Millisecond)
				defer cancel()
			}
			done := make(chan error, 1)
			go func() { _, e := prefix(ctx, c); done <- e }()
			if action == "cancel" {
				cancel()
			}
			if action == "close" {
				var wg sync.WaitGroup
				for range 4 {
					wg.Go(func() { c.Close() })
				}
				wg.Wait()
			}
			select {
			case e := <-done:
				if e == nil || action == "cancel" && !errors.Is(e, context.Canceled) || action == "deadline" && !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal(e)
				}
			case <-time.After(time.Second):
				t.Fatal("read retained")
			}
		})
	}
}
func TestRealSCTPCompletedContextsAndFrameDeadline(t *testing.T) {
	a, b := adapters(t)
	for range 25 {
		ctx, cancel := context.WithCancel(testContext(t))
		r := a.WriteFrame(ctx, data(80))
		cancel()
		if r.Err != nil {
			t.Fatal(r)
		}
		ctx, cancel = context.WithCancel(testContext(t))
		got, e := read(ctx, b)
		cancel()
		if e != nil || len(got) != 80 {
			t.Fatal(e)
		}
	}
	if r := a.WriteFrame(testContext(t), data(40)); r.Err != nil {
		t.Fatal(r)
	}
	ctx, cancel := context.WithTimeout(testContext(t), 30*time.Millisecond)
	defer cancel()
	if _, e := prefix(ctx, b); e != nil {
		t.Fatal(e)
	}
	<-ctx.Done()
	if e := b.ReadFrameInto(testContext(t), make([]byte, 40)); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
func TestRealSCTPShutdownNotificationSealsWriter(t *testing.T) {
	a, b := sockets(t)
	c, e := New(a, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	b.Close()
	_, e = prefix(testContext(t), c)
	var v *transport.Event
	if !errors.As(e, &v) || !v.Terminal || v.Kind != transport.EventShutdown {
		t.Fatalf("%T %v", e, e)
	}
	if r := c.WriteFrame(testContext(t), data(20)); !errors.Is(r.Err, net.ErrClosed) {
		t.Fatal(r)
	}
}
func TestResourceLedgerAndConstructor(t *testing.T) {
	cfg := DefaultConfig()
	info, e := cfg.Requirements()
	if e != nil {
		t.Fatal(e)
	}
	ledger, _ := upstream.SetupRequirements(DefaultSetupConfig().upstreamConfig())
	req := ledger.IORequirements
	// Includes the concrete adapter value, one transient Event and constructor
	// snapshots; buffer arrays are already within sizeof(conn), not added twice.
	state := unsafe.Sizeof(conn{}) + unsafe.Sizeof(transport.Event{}) + unsafe.Sizeof(upstream.Status{}) + unsafe.Sizeof(upstream.MessageOptionState{})
	if uint64(state) > StateReserve || info.BufferBytes != ledger.ConnectionBytes+StateReserve+uint64(cfg.NotificationMax+req.OOBMin) {
		t.Fatal(state, info)
	}
	for _, bad := range []Config{{}, {NotificationMax: 1, NotificationWindow: time.Second, MaxNotificationsPerWindow: 1}, {NotificationMax: 1<<20 + 1, NotificationWindow: time.Second, MaxNotificationsPerWindow: 1}, {NotificationMax: 4096, MaxNotificationsPerWindow: 1}, {NotificationMax: 4096, NotificationWindow: time.Second}} {
		if _, e := bad.Requirements(); e == nil {
			t.Fatal(bad)
		}
	}
	a, _ := sockets(t)
	if _, e := New(a, Config{}); e == nil {
		t.Fatal("invalid config")
	}
	if _, e := a.GetStatus(); e == nil {
		t.Fatal("failed constructor leaked socket")
	}
	t.Logf("state=%d upstream=%d bound=%d", state, ledger.ConnectionBytes, info.BufferBytes)
}

// Deterministic fragments/notifications test the same adapter below ReadEvent;
// native tests above prove the syscall boundary, no fake kernel claim here.
type step struct {
	r            upstream.ReadResult
	data, notify []byte
	err          error
}
type scripted struct {
	steps                 []step
	calls, closed, writes int
	written               int
	writeErr              error
}

func (f *scripted) ReadEvent(ctx context.Context, d, n, o []byte) (upstream.ReadResult, error) {
	if e := ctx.Err(); e != nil {
		return upstream.ReadResult{}, e
	}
	if len(f.steps) == 0 {
		return upstream.ReadResult{}, io.EOF
	}
	s := f.steps[0]
	f.steps = f.steps[1:]
	f.calls++
	copy(d, s.data)
	copy(n, s.notify)
	return s.r, s.err
}
func (f *scripted) WriteMsgContext(context.Context, []byte, upstream.SndInfo) (int, error) {
	f.writes++
	return f.written, f.writeErr
}
func (f *scripted) Close() error { f.closed++; return nil }
func scriptedConn(steps ...step) (*conn, *scripted) {
	f := &scripted{steps: steps}
	cfg := DefaultConfig()
	i, _ := cfg.Requirements()
	return &conn{socket: f, config: cfg, info: i, notification: make([]byte, 4096), oob: make([]byte, 64), inbound: 16, outbound: 1, association: 7}, f
}
func fragment(b []byte, eor bool) step {
	flags := 0
	if eor {
		flags = syscall.MSG_EOR
	}
	return step{r: upstream.ReadResult{Kind: upstream.ReadData, N: len(b), InfoPresent: true, Consumed: true, RecvFlags: flags, Info: upstream.SndRcvInfo{AssocID: 7, PPID: 46}}, data: b}
}
func notify(kind uint16, body int) step {
	b := make([]byte, 8+body)
	binary.NativeEndian.PutUint16(b, kind)
	binary.NativeEndian.PutUint32(b[4:], uint32(len(b)))
	return step{r: upstream.ReadResult{Kind: upstream.ReadNotification, N: len(b), Consumed: true}, notify: b}
}
func unknown() step { return notify(0xffff, 0) }
func TestContinuationInEveryReceiveStage(t *testing.T) {
	for _, discard := range []bool{false, true} {
		b := data(50)
		c, f := scriptedConn(fragment(b[:7], false), unknown(), fragment(b[7:20], false), fragment(b[20:27], false), unknown(), fragment(b[27:], true))
		var event *transport.Event
		if _, e := c.ReadPrefix(testContext(t), time.Second); !errors.As(e, &event) || event.Progress != 7 {
			t.Fatal(e, event)
		}
		d := c.deadline
		if _, e := c.ReadPrefix(testContext(t), 2*time.Second); e != nil {
			t.Fatal(e)
		}
		if c.deadline != d {
			t.Fatal("deadline refreshed")
		}
		dst := make([]byte, 50)
		readBody := func() error {
			if discard {
				return c.DiscardFrame(testContext(t), 50)
			}
			return c.ReadFrameInto(testContext(t), dst)
		}
		if e := readBody(); !errors.As(e, &event) || event.Progress != 27 {
			t.Fatal(e, event)
		}
		if _, e := c.ReadPrefix(testContext(t), time.Second); e == nil {
			t.Fatal("phase reset")
		}
		if e := readBody(); e != nil {
			t.Fatal(e)
		}
		if !discard && !bytes.Equal(dst, b) {
			t.Fatal("continuation overwrote body")
		}
		if c.deadline != (time.Time{}) || f.closed != 0 {
			t.Fatal("cursor not cleared")
		}
	}
}
func TestNotificationLimitAndTypedTerminal(t *testing.T) {
	c, f := scriptedConn(unknown(), unknown())
	c.config.MaxNotificationsPerWindow = 1
	_, _ = c.ReadPrefix(testContext(t), time.Second)
	if _, e := c.ReadPrefix(testContext(t), time.Second); !errors.Is(e, transport.ErrNotificationLimit) || f.closed != 1 {
		t.Fatal(e, f.closed)
	}
	for _, kind := range []uint16{uint16(upstream.SCTP_ASSOC_CHANGE), uint16(upstream.SCTP_SHUTDOWN_EVENT), uint16(upstream.SCTP_PARTIAL_DELIVERY_EVENT)} {
		n := notify(kind, 4)
		if kind == uint16(upstream.SCTP_ASSOC_CHANGE) {
			n = notify(kind, 12)
			binary.NativeEndian.PutUint16(n.notify[8:], uint16(upstream.SCTP_RESTART))
			binary.NativeEndian.PutUint32(n.notify[16:], 7)
		} else if kind == uint16(upstream.SCTP_PARTIAL_DELIVERY_EVENT) {
			n = notify(kind, 16)
			binary.NativeEndian.PutUint32(n.notify[12:], 7)
		} else {
			binary.NativeEndian.PutUint32(n.notify[8:], 7)
		}
		c, f = scriptedConn(n)
		_, e := c.ReadPrefix(testContext(t), time.Second)
		var event *transport.Event
		if !errors.As(e, &event) || !event.Terminal || f.closed != 1 {
			t.Fatal(e, event)
		}
		if r := c.WriteFrame(testContext(t), data(20)); !errors.Is(r.Err, net.ErrClosed) || f.writes != 0 {
			t.Fatal(r)
		}
	}
	// Path failure and SEND_FAILED remain owned diagnostics, including unknown
	// numeric states and a borrowed upstream payload tail that we never retain.
	path := notify(uint16(upstream.SCTP_PEER_ADDR_CHANGE), 140)
	binary.NativeEndian.PutUint16(path.notify[8:], syscall.AF_INET)
	binary.BigEndian.PutUint16(path.notify[10:], 3868)
	copy(path.notify[12:], []byte{192, 0, 2, 1})
	binary.NativeEndian.PutUint32(path.notify[136:], 999)
	binary.NativeEndian.PutUint32(path.notify[144:], 7)
	failed := notify(uint16(upstream.SCTP_SEND_FAILED), 48)
	binary.NativeEndian.PutUint32(failed.notify[8:], 123)
	binary.BigEndian.PutUint32(failed.notify[20:], 46)
	binary.NativeEndian.PutUint32(failed.notify[44:], 7)
	c, f = scriptedConn(path, failed)
	_, e := c.ReadPrefix(testContext(t), time.Second)
	var event *transport.Event
	if !errors.As(e, &event) || event.Kind != transport.EventPathChange || event.State != 999 || event.Address.String() != "192.0.2.1" || event.Port != 3868 || event.Terminal {
		t.Fatal(e, event)
	}
	saved := *event
	_, e = c.ReadPrefix(testContext(t), time.Second)
	if !errors.As(e, &event) || event.Kind != transport.EventSendFailure || event.Code != 123 || event.PPID != 46 || event.Terminal || f.closed != 0 {
		t.Fatal(e, event)
	}
	clear(c.notification)
	if saved.Address.String() != "192.0.2.1" || event.Code != 123 {
		t.Fatal("borrowed notification")
	}

}
func TestFragmentsNeverRefreshDeadlineOrBorrowAfterReturn(t *testing.T) {
	b := data(40)
	c, _ := scriptedConn(fragment(b[:20], false), unknown(), fragment(b[20:], true))
	if _, e := c.ReadPrefix(testContext(t), time.Millisecond); e != nil {
		t.Fatal(e)
	}
	c.deadline = time.Now().Add(-time.Second)
	if e := c.ReadFrameInto(testContext(t), make([]byte, 40)); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
func TestWriteEvidenceAndNoTailRetry(t *testing.T) {
	for _, v := range []struct {
		n int
		e error
	}{{0, context.Canceled}, {7, io.ErrClosedPipe}, {7, nil}, {20, nil}, {20, io.ErrClosedPipe}} {
		c, f := scriptedConn()
		f.written = v.n
		f.writeErr = v.e
		r := c.WriteFrame(testContext(t), data(20))
		if r.Bytes != v.n || r.Uncertain || f.writes != 1 {
			t.Fatal(r, f.writes)
		}
		if v.n != 20 || v.e != nil {
			if r.Err == nil || f.closed != 1 {
				t.Fatal(r)
			}
		} else if r.Err != nil {
			t.Fatal(r)
		}
	}
	// Receiver-local failure wins before raw Close wakes an already blocked send.
	c, _ := scriptedConn()
	b := &blockedWriter{entered: make(chan struct{}), closed: make(chan struct{})}
	c.socket = b
	done := make(chan transport.WriteResult, 1)
	go func() { done <- c.WriteFrame(testContext(t), data(20)) }()
	<-b.entered
	c.failure(transport.ErrNotificationLimit)
	r := <-done
	if !errors.Is(r.Err, transport.ErrNotificationLimit) || r.Bytes != 0 || r.Uncertain {
		t.Fatal(r)
	}

}
func TestRealSCTPBlockedWriteCancellationAndClose(t *testing.T) {
	for _, action := range []string{"cancel", "close", "deadline"} {
		t.Run(action, func(t *testing.T) {
			a, _ := sockets(t)
			if e := a.SetWriteBuffer(4096); e != nil {
				t.Fatal(e)
			}
			c, e := New(a, DefaultConfig())
			if e != nil {
				t.Fatal(e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if action == "deadline" {
				ctx, cancel = context.WithTimeout(ctx, 60*time.Millisecond)
				defer cancel()
			}
			entered := make(chan struct{})
			done := make(chan transport.WriteResult, 1)
			go func() {
				close(entered)
				for {
					r := c.WriteFrame(ctx, data(2048))
					if r.Err != nil {
						done <- r
						return
					}
				}
			}()
			<-entered
			if action != "deadline" {
				time.Sleep(20 * time.Millisecond)
				if action == "cancel" {
					cancel()
				} else {
					c.Close()
				}
			}
			select {
			case r := <-done:
				if r.Err == nil || r.Bytes < 0 || r.Bytes > 2048 || r.Uncertain {
					t.Fatal(r)
				}
				if action == "deadline" && !errors.Is(r.Err, context.DeadlineExceeded) {
					t.Fatal(r)
				}
				if action == "cancel" && !errors.Is(r.Err, context.Canceled) {
					t.Fatal(r)
				}
			case <-time.After(time.Second):
				t.Fatal("writer retained")
			}
		})
	}
}
func TestNotificationBoundsAndUpstreamFragmentErrors(t *testing.T) {
	for _, kind := range []string{"missing_info", "changed_stream", "changed_ssn", "truncation", "oversized_notification", "malformed_notification"} {
		t.Run(kind, func(t *testing.T) {
			b := data(40)
			next := fragment(b[20:], true)
			switch kind {
			case "missing_info":
				next.r.InfoPresent = false
			case "changed_stream":
				next.r.Info.Stream = 1
			case "changed_ssn":
				next.r.Info.SSN = 1
			case "truncation":
				next.r.RecvFlags |= syscall.MSG_TRUNC
			case "oversized_notification":
				next = step{r: upstream.ReadResult{Kind: upstream.ReadNotification, RequiredNotificationBytes: 8192}, err: upstream.ErrNotificationTooLarge}
			case "malformed_notification":
				next = notify(uint16(upstream.SCTP_ASSOC_CHANGE), 0)
			}
			// These failures are now detected by upstream E. Feed its public
			// result/error contract, rather than reimplementing that validator.
			switch kind {
			case "missing_info", "changed_stream", "changed_ssn", "truncation":
				next.r.CloseRequired = true
				next.err = upstream.ErrMessageBoundary
			}
			c, f := scriptedConn(fragment(b[:20], false), next)
			if _, e := c.ReadPrefix(testContext(t), time.Second); e != nil {
				t.Fatal(e)
			}
			if e := c.ReadFrameInto(testContext(t), make([]byte, 40)); e == nil || f.closed != 1 {
				t.Fatal(e, f.closed)
			}
			if len(c.notification) != 4096 || f.calls != 2 {
				t.Fatal("grew or retried")
			}
		})
	}
}

type blockedWriter struct {
	entered, closed chan struct{}
	once            sync.Once
}

func (b *blockedWriter) ReadEvent(context.Context, []byte, []byte, []byte) (upstream.ReadResult, error) {
	return upstream.ReadResult{}, io.EOF
}
func (b *blockedWriter) WriteMsgContext(context.Context, []byte, upstream.SndInfo) (int, error) {
	close(b.entered)
	<-b.closed
	return 0, net.ErrClosed
}
func (b *blockedWriter) Close() error { b.once.Do(func() { close(b.closed) }); return nil }
