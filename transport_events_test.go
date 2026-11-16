package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/internal/testkit"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"net/netip"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func TestTransportFactsAreBoundedAndNotLiveness(t *testing.T) {
	if ioStopReason(syscall.ESHUTDOWN) != StopTransport {
		t.Fatal("SCTP send shutdown misclassified")
	}
	n, co := newTestNode(t)
	p, _, _ := connectTest(t, n, co)
	n.mu.Lock()
	c := p.conn
	before := c.tw
	n.mu.Unlock()
	for i, kind := range []transport.EventKind{transport.EventAssociationUp, transport.EventPathChange, transport.EventSendFailure, transport.EventUnknown} {
		e := &transport.Event{Sequence: uint64(i + 1), Kind: kind, Address: netip.MustParseAddr("192.0.2.1"), State: 1, Context: 0, Progress: 27}
		more, err := n.receiveEvent(c, e)
		if !more || err != nil {
			t.Fatal(more, err)
		}
	}
	n.mu.Lock()
	after := c.tw
	n.mu.Unlock()
	f := p.Info()
	if before != after || !f.Ready || f.Transport.Events != 4 || f.Transport.PathChanges != 1 || f.Transport.SendFailures != 1 || f.Transport.Unknown != 1 || f.Transport.LastPath.Address.String() != "192.0.2.1" || f.Transport.ConnectionID != f.ConnectionID {
		t.Fatal(f)
	}
	used := 2*unsafe.Sizeof(TransportFacts{}) + unsafe.Sizeof(transport.Event{})
	if uint64(used) > transportEventReserve {
		t.Fatal(used, transportEventReserve)
	}
	// Snapshots have owned values; changing one cannot change Peer state.
	f.Transport.LastPath.State = 999
	if p.Info().Transport.LastPath.State == 999 {
		t.Fatal("borrowed facts")
	}
	// A valid message cannot clear an expired barrier while its timer is delayed.
	n.mu.Lock()
	c.sendBarrierUntil = time.Now().Add(-time.Second)
	n.aliveLocked(c, false)
	closed, cause := c.closed, c.cause
	n.mu.Unlock()
	if !closed || !errors.Is(cause, context.DeadlineExceeded) {
		t.Fatal(closed, cause)
	}
}
func TestTransportTerminalEventSettlesOnceAndRetires(t *testing.T) {
	for _, kind := range []transport.EventKind{transport.EventAssociationLost, transport.EventAssociationRestart, transport.EventShutdown, transport.EventPartialDeliveryAborted} {
		t.Run((&transport.Event{Kind: kind}).Error()+string(rune('0'+kind)), func(t *testing.T) {
			n, co := newTestNode(t)
			p, _, remote := connectTest(t, n, co)
			port := portTest(t, n, 1)
			id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = readMessage(ctxTest(t), remote, co); e != nil {
				t.Fatal(e)
			}
			n.mu.Lock()
			c := p.conn
			n.mu.Unlock()
			more, e := n.receiveEvent(c, &transport.Event{Sequence: 1, Kind: kind, Terminal: true})
			if more || !errors.Is(e, transport.ErrAssociationClosed) {
				t.Fatal(more, e)
			}
			result, e := port.Receive(ctxTest(t))
			if e != nil || result.RequestID != id || result.Result.Outcome != WrittenNoAnswer {
				t.Fatal(result, e)
			}
			waitFor(t, func() bool { return p.Info().Candidates == 0 })
			if p.Info().Ready || p.Info().Watchdog != Down || n.Resources().Outstanding != 0 {
				t.Fatal(p.Info(), n.Resources())
			}
			if more, _ = n.receiveEvent(c, &transport.Event{Sequence: 2, Kind: transport.EventAssociationUp}); more {
				t.Fatal("revived retired connection")
			}
			if p.Info().Transport.Last.Kind != kind {
				t.Fatal("late event replaced terminal facts")
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal(e)
			}
		})
	}
}

// The continuation is exercised through the core helpers at all phases; the
// adapter separately tests byte offsets and never retaining a target slice.
type eventTransport struct {
	transport.Conn
	prefixEvents, bodyEvents, discardEvents int
	sequence                                uint64
}

func (e *eventTransport) event(n *int) error {
	if *n == 0 {
		return nil
	}
	*n--
	e.sequence++
	return &transport.Event{Kind: transport.EventPathChange, Sequence: e.sequence}
}
func (e *eventTransport) ReadPrefix(ctx context.Context, d time.Duration) ([20]byte, error) {
	if err := e.event(&e.prefixEvents); err != nil {
		return [20]byte{}, err
	}
	return e.Conn.ReadPrefix(ctx, d)
}
func (e *eventTransport) ReadFrameInto(ctx context.Context, b []byte) error {
	if err := e.event(&e.bodyEvents); err != nil {
		return err
	}
	return e.Conn.ReadFrameInto(ctx, b)
}
func (e *eventTransport) DiscardFrame(ctx context.Context, n int) error {
	if err := e.event(&e.discardEvents); err != nil {
		return err
	}
	return e.Conn.DiscardFrame(ctx, n)
}
func TestTransportContinuationHandshakeBodyAndDiscard(t *testing.T) {
	n, co := newTestNode(t)
	a, b := testkit.Pipe()
	defer b.Close()
	wrapped := &eventTransport{Conn: a, prefixEvents: 2, bodyEvents: 2, discardEvents: 2}
	ctx := ctxTest(t)
	done := make(chan error, 1)
	go func() {
		req, e := readMessage(ctx, b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(req, n.cea(2001), draOrigin)
			if e == nil {
				e = writeMessage(ctx, b, co, ans)
			}
		}
		if e == nil {
			req, e = readMessage(ctx, b, co)
		}
		if e == nil {
			var ans pdu.Message
			ans, e = answerFor(co, req)
			if e == nil {
				// First unmatched HBH forces discard under the same event-aware wrapper.
				unmatched := ans
				unmatched.Header.HopByHop++
				e = writeMessage(ctx, b, co, unmatched)
				if e == nil {
					e = writeMessage(ctx, b, co, ans)
				}
			}
		}
		done <- e
	}()
	p, e := n.Connect(ctx, "dra", wrapped)
	if e != nil {
		t.Fatal(e)
	}
	r, e := p.Request(ctx, requestBody(), RequestOptions{})
	if e != nil || r.Outcome != AnswerReceived {
		t.Fatal(r, e)
	}
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if p.Info().Transport.Events != 6 || p.Info().Transport.PathChanges != 6 {
		t.Fatal(p.Info())
	}
}
