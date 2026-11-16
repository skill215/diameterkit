//go:build linux && !386

package sctp

import (
	"bytes"
	"context"
	"errors"
	"github.com/skill215/diameterkit/transport"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func listenTest(t *testing.T) *Listener {
	t.Helper()
	l, e := Listen(testContext(t), "sctp4", "127.0.0.1:0", DefaultSetupConfig())
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { l.Close() })
	return l
}
func dialerTest(t *testing.T, l *Listener) *Dialer {
	t.Helper()
	d, e := NewDialer("sctp4", "", l.Addr().String(), DefaultSetupConfig())
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func TestSetupInputBoundsAndLedger(t *testing.T) {
	cfg := DefaultSetupConfig()
	r, e := cfg.Requirements()
	if e != nil {
		t.Fatal(e)
	}
	if r.BufferBytes != 52620 || r.ConnectionBufferBytes != 99676 {
		t.Fatal(r)
	}
	size := unsafe.Sizeof(Listener{}) + unsafe.Sizeof(incoming{}) + unsafe.Sizeof(Dialer{})
	if uint64(size) > SetupStateReserve {
		t.Fatal(size)
	}
	t.Logf("setup_values=%d listener=%d candidate=%d", size, r.BufferBytes, r.ConnectionBufferBytes)
	for _, test := range []struct{ network, local, remote string }{
		{"tcp", "", "127.0.0.1:3868"}, {"sctp4", "", ""}, {"sctp4", "", ":3868"}, {"sctp4", "", "localhost:diameter"}, {"sctp4", "", "localhost:0"}, {"sctp4", "", "localhost:+1"}, {"sctp4", "", "localhost:65536"}, {"sctp4", "", "127.0.0.1/:3868"}, {"sctp4", "", "127.0.0.1/127.0.0.1:3868/"},
		{"sctp4", "", strings.Repeat("a", 4097) + ":9"}, {"sctp4", "", strings.Repeat("127.0.0.1/", 16) + "127.0.0.1:9"},
		{"sctp6", "", "[::1%" + strings.Repeat("z", 257) + "]:9"},
	} {
		if d, e := NewDialer(test.network, test.local, test.remote, cfg); d != nil || !errors.Is(e, ErrConfig) {
			t.Fatal(test, e)
		}
	}
	tiny := cfg
	tiny.MaxAddresses = 1
	if _, e := NewDialer("sctp4", "127.0.0.1:0", "127.0.0.1:9", tiny); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
	for _, bad := range []SetupConfig{{}, {Transport: DefaultConfig(), MaxAddresses: 65, MaxAddressBytes: 4096}, {Transport: DefaultConfig(), MaxAddresses: 1, MaxZoneBytes: 4097, MaxAddressBytes: 4096}} {
		if _, e := bad.Requirements(); e == nil {
			t.Fatal(bad)
		}
	}
	for _, test := range []struct{ network, local, remote string }{{"sctp4", "127.0.0.1:0", "localhost:3868"}, {"sctp", "", "127.0.0.1/127.0.0.2:3868"}, {"sctp6", "[::1]:0", "[::1%1]:3868"}} {
		if _, e := NewDialer(test.network, test.local, test.remote, cfg); e != nil {
			t.Fatal(test, e)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d, e := NewDialer("sctp4", "", "localhost:3868", cfg)
	if e != nil {
		t.Fatal(e)
	}
	if c, e := d.Dial(ctx); c != nil || !errors.Is(e, context.Canceled) {
		t.Fatal(c, e)
	}
	if l, e := Listen(ctx, "sctp4", "127.0.0.1:0", cfg); l != nil || !errors.Is(e, context.Canceled) {
		t.Fatal(l, e)
	}
	if _, e := d.Dial(nil); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}

}
func TestListenerCancelDeadlineReuseAndOneOwner(t *testing.T) {
	l := listenTest(t)
	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		started := time.Now()
		in, e := l.Accept(ctx)
		cancel()
		if in != nil || !errors.Is(e, context.DeadlineExceeded) || time.Since(started) > time.Second {
			t.Fatal(in, e)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan error, 1)
	go func() { _, e := l.Accept(ctx); out <- e }()
	waitFor(t, func() bool { return l.accepting.Load() })
	if _, e := l.Accept(testContext(t)); !errors.Is(e, ErrConcurrentAccept) {
		t.Fatal(e)
	}
	cancel()
	if e := <-out; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	d := dialerTest(t, l)
	clientCtx, clientCancel := context.WithCancel(testContext(t))
	c, e := d.Dial(clientCtx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	clientCancel()
	i, e := l.Accept(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	serverCtx, serverCancel := context.WithCancel(testContext(t))
	s, e := i.Establish(serverCtx)
	serverCancel()
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if i.Info().BufferBytes != d.Info().BufferBytes || s.Info().BufferBytes > d.Info().BufferBytes {
		t.Fatal(i.Info(), s.Info())
	}
	if _, e = i.Establish(testContext(t)); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
	// Cancellation after handoff cannot poison established data I/O.
	time.Sleep(40 * time.Millisecond)
	b := data(80)
	if r := c.WriteFrame(testContext(t), b); r.Err != nil {
		t.Fatal(r)
	}
	got, e := read(testContext(t), s)
	if e != nil || !bytes.Equal(b, got) {
		t.Fatal(e)
	}
}
func TestListenerCloseUnblocksAndIncomingCloseRace(t *testing.T) {
	l := listenTest(t)
	out := make(chan error, 1)
	go func() { _, e := l.Accept(testContext(t)); out <- e }()
	waitFor(t, func() bool { return l.accepting.Load() })
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() { l.Close() })
	}
	wg.Wait()
	select {
	case e := <-out:
		if !errors.Is(e, net.ErrClosed) {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("Accept retained")
	}
	for range 20 {
		l := listenTest(t)
		d := dialerTest(t, l)
		c, e := d.Dial(testContext(t))
		if e != nil {
			t.Fatal(e)
		}
		i, e := l.Accept(testContext(t))
		if e != nil {
			t.Fatal(e)
		}
		type result struct {
			c transport.Conn
			e error
		}
		done := make(chan result, 1)
		go func() { c, e := i.Establish(testContext(t)); done <- result{c, e} }()
		i.Close()
		r := <-done
		if r.c != nil {
			if w := r.c.WriteFrame(testContext(t), data(20)); w.Err == nil {
				t.Fatal("closed Incoming handed out usable socket")
			}
			r.c.Close()
		}
		if _, e = i.Establish(testContext(t)); !errors.Is(e, net.ErrClosed) {
			t.Fatal(e)
		}
		c.Close()
		l.Close()
	}
}
func TestListenerCanceledArrivalAndFreshAccept(t *testing.T) {
	for range 20 {
		// A canceled Accept may leave this round's arrival queued. Do not
		// feed its subsequently closed association to the next round: that
		// can finish with a kernel error before the next cancel even runs.
		l := listenTest(t)
		d := dialerTest(t, l)
		ctx, cancel := context.WithCancel(context.Background())
		out := make(chan transport.Incoming, 1)
		errs := make(chan error, 1)
		go func() { i, e := l.Accept(ctx); out <- i; errs <- e }()
		type result struct {
			c transport.Conn
			e error
		}
		dial := make(chan result, 1)
		go func() { c, e := d.Dial(testContext(t)); dial <- result{c, e} }()
		cancel()
		i, e := <-out, <-errs
		if i != nil {
			i.Close()
		} else if !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		r := <-dial
		if r.c != nil {
			r.c.Close()
		}
		l.Close()
		if l.accepting.Load() {
			t.Fatal("accept owner retained")
		}
	}
	// Verify an actual fresh accept and exchange after all canceled owners
	// have returned. Same-listener reuse without a racing arrival is covered
	// by TestListenerCancelDeadlineReuseAndOneOwner.
	l := listenTest(t)
	c, e := dialerTest(t, l).Dial(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	i, e := l.Accept(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	s, e := i.Establish(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	frame := data(80)
	if r := c.WriteFrame(testContext(t), frame); r.Err != nil {
		t.Fatal(r.Err)
	}
	got, e := read(testContext(t), s)
	if e != nil || !bytes.Equal(got, frame) {
		t.Fatal("fresh exchange", e)
	}
}
