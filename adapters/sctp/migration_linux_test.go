//go:build linux && !386

package sctp

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

func terminalRecord(restart bool) step {
	n := notify(uint16(upstream.SCTP_PARTIAL_DELIVERY_EVENT), 16)
	binary.NativeEndian.PutUint32(n.notify[8:], upstream.PartialDeliveryAborted)
	binary.NativeEndian.PutUint32(n.notify[12:], 7)
	if restart {
		n = notify(uint16(upstream.SCTP_ASSOC_CHANGE), 12)
		binary.NativeEndian.PutUint16(n.notify[8:], uint16(upstream.SCTP_RESTART))
		binary.NativeEndian.PutUint32(n.notify[16:], 7)
		n.err = upstream.ErrAssociationChanged
	}
	n.r.CloseRequired = true
	return n
}
func TestUpstreamTerminalRecordsPreserveFacts(t *testing.T) {
	for _, restart := range []bool{false, true} {
		name := "partial-abort"
		if restart {
			name = "restart"
		}
		for _, phase := range []string{"prefix", "body", "discard"} {
			t.Run(name+"/"+phase, func(t *testing.T) {
				b := data(40)
				progress := 20
				if phase == "prefix" {
					progress = 7
				}
				c, f := scriptedConn(fragment(b[:progress], false), terminalRecord(restart))
				_, e := c.ReadPrefix(testContext(t), time.Second)
				if phase != "prefix" {
					if e != nil {
						t.Fatal(e)
					}
					if phase == "body" {
						e = c.ReadFrameInto(testContext(t), make([]byte, 40))
					} else {
						e = c.DiscardFrame(testContext(t), 40)
					}
				}
				var ev *transport.Event
				expected := transport.EventPartialDeliveryAborted
				if restart {
					expected = transport.EventAssociationRestart
				}
				if !errors.As(e, &ev) || !ev.Terminal || ev.Kind != expected || ev.Association != 7 || ev.Progress != progress || ev.Sequence != 1 || !errors.Is(e, transport.ErrAssociationClosed) || f.closed != 1 {
					t.Fatal(e, ev, f.closed)
				}
				if restart && !errors.Is(e, upstream.ErrAssociationChanged) {
					t.Fatal("lost native cause", e)
				}
				saved := *ev
				clear(c.notification)
				if saved != *ev {
					t.Fatal("borrowed notification")
				}
				if r := c.WriteFrame(testContext(t), data(20)); !errors.Is(r.Err, transport.ErrAssociationClosed) || f.writes != 0 {
					t.Fatal(r, f.writes)
				}
			})
		}
	}
}
func TestUpstreamInvalidTerminalRecordCannotContinue(t *testing.T) {
	for _, kind := range []string{"nonterminal", "wrong-restart-type", "malformed", "unconsumed", "failed-consumption"} {
		t.Run(kind, func(t *testing.T) {
			n := terminalRecord(true)
			switch kind {
			case "nonterminal":
				n = unknown()
				n.r.CloseRequired = true
			case "wrong-restart-type":
				binary.NativeEndian.PutUint16(n.notify[8:], uint16(upstream.SCTP_COMM_UP))
			case "malformed":
				binary.NativeEndian.PutUint32(n.notify[4:], 8)
			case "unconsumed":
				n.r.Consumed = false
			case "failed-consumption":
				n.err = errors.Join(upstream.ErrAssociationChanged, upstream.ErrMessageBoundary)
			}
			c, f := scriptedConn(n)
			_, e := c.ReadPrefix(testContext(t), time.Second)
			var ev *transport.Event
			if e == nil || errors.As(e, &ev) || f.closed != 1 {
				t.Fatal(e, ev, f.closed)
			}
			if _, e = c.ReadPrefix(testContext(t), time.Second); !errors.Is(e, net.ErrClosed) || f.calls != 1 {
				t.Fatal(e, f.calls)
			}
		})
	}
}
func TestUpstreamRestartWriteLatchPreservesDelivery(t *testing.T) {
	for _, n := range []int{0, 7, 20} {
		c, f := scriptedConn()
		f.written = n
		f.writeErr = upstream.ErrAssociationChanged
		r := c.WriteFrame(testContext(t), data(20))
		if r.Bytes != n || r.Uncertain || !errors.Is(r.Err, upstream.ErrAssociationChanged) || !errors.Is(r.Err, transport.ErrAssociationClosed) || f.writes != 1 || f.closed != 1 {
			t.Fatal(r, f.writes, f.closed)
		}
	}
	c, _ := scriptedConn()
	c.failure(transport.ErrNotificationLimit)
	if r := c.WriteFrame(testContext(t), data(20)); !errors.Is(r.Err, transport.ErrNotificationLimit) || errors.Is(r.Err, transport.ErrAssociationClosed) {
		t.Fatal("resource cause downgraded", r)
	}
}
func TestUpstreamListenerDeadlineAndOwnedBind(t *testing.T) {
	l := listenTest(t)
	first := l.Addr().(*upstream.SCTPAddr)
	port := first.Port
	first.Port = 1
	clear(first.IPAddrs[0].IP)
	second := l.Addr().(*upstream.SCTPAddr)
	if second.Port != port || !second.IPAddrs[0].IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Fatal(second)
	}
	if e := l.raw.SetDeadline(time.Now().Add(-time.Second)); e != nil {
		t.Fatal(e)
	}
	if _, e := l.Accept(testContext(t)); !errors.Is(e, os.ErrDeadlineExceeded) {
		t.Fatal(e)
	}
	if e := l.raw.SetDeadline(time.Time{}); e != nil {
		t.Fatal(e)
	}
	c, e := dialerTest(t, l).Dial(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	i, e := l.Accept(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	s, e := i.Establish(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	l.Close()
	if r := c.WriteFrame(testContext(t), data(32)); r.Err != nil {
		t.Fatal(r)
	}
	if _, e = read(testContext(t), s); e != nil {
		t.Fatal(e)
	}
	wildcard, e := Listen(testContext(t), "sctp4", ":0", DefaultSetupConfig())
	if e != nil {
		t.Fatal(e)
	}
	defer wildcard.Close()
	a := wildcard.Addr().(*upstream.SCTPAddr)
	if a.Port == 0 || len(a.IPAddrs) != 1 || !a.IPAddrs[0].IP.IsUnspecified() {
		t.Fatal(a)
	}
}
func TestUpstreamSocketTimeoutTakeover(t *testing.T) {
	a, b := sockets(t)
	if e := a.SetSocketTimeouts(time.Millisecond, time.Millisecond); e != nil {
		t.Fatal(e)
	}
	if e := a.SetDeadline(time.Now().Add(-time.Second)); e != nil {
		t.Fatal(e)
	}
	c, e := New(a, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	// After old SO/read deadlines would have expired, the taken-over socket must
	// still wait for the caller's deadline and accept newly arriving data.
	done := make(chan error, 1)
	go func() { _, e := read(testContext(t), c); done <- e }()
	select {
	case e := <-done:
		t.Fatal("inherited timeout", e)
	case <-time.After(20 * time.Millisecond):
	}
	if _, e = b.WriteMsgContext(context.Background(), data(32), upstream.SndInfo{PPID: 46}); e != nil {
		t.Fatal(e)
	}
	select {
	case e := <-done:
		if e != nil {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("read retained")
	}
}
