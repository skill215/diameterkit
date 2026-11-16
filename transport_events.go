package diameter

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/skill215/diameterkit/transport"
)

// Two fixed fact snapshots plus the in-flight event; conservatively charged
// per physical candidate even though the last Peer snapshot is shared.
const transportEventReserve uint64 = 2048

// TransportFacts is a fixed diagnostic snapshot of this physical connection.
// Notifications are not Diameter liveness or proof of remote processing.
type TransportFacts struct {
	ConnectionID                               ConnectionID
	Events, PathChanges, SendFailures, Unknown uint64
	Last                                       transport.Event
	LastPath                                   transport.Event
}

func incrementSaturated(v *uint64) {
	if *v != math.MaxUint64 {
		*v++
	}
}

// receiveEvent runs only between socket operations, never while holding a
// borrowed notification buffer. It neither refreshes Tw nor completes attempts
// from SEND_FAILED payloads. Context zero deliberately provides diagnostics only.
func (n *Node) receiveEvent(c *connection, err error) (bool, error) {
	var e *transport.Event
	if !errors.As(err, &e) {
		return false, err
	}
	n.mu.Lock()
	if c.closed {
		cause := c.cause
		if cause == nil {
			cause = ErrClosed
		}
		n.mu.Unlock()
		return false, cause
	}
	if e.Sequence == 0 || e.Sequence <= c.transport.Last.Sequence {
		n.mu.Unlock()
		n.fail(c, ErrProtocol)
		return false, ErrProtocol
	}
	f := &c.transport
	f.ConnectionID = c.id
	incrementSaturated(&f.Events)
	f.Last = *e
	switch e.Kind {
	case transport.EventPathChange:
		incrementSaturated(&f.PathChanges)
		f.LastPath = *e
	case transport.EventSendFailure:
		incrementSaturated(&f.SendFailures)
	case transport.EventUnknown:
		incrementSaturated(&f.Unknown)
	}
	// A losing/old candidate cannot overwrite the active connection's facts.
	if c.peer.conn == c || c.peer.conn == nil && c.peer.chosen == nil {
		c.peer.transport = *f
	}
	if e.Terminal {
		n.stopConnectionReasonLocked(c, e, StopTransport)
	}
	n.mu.Unlock()
	if e.Terminal {
		_ = c.io.Close()
		n.notify()
		return false, e
	}
	return true, nil
}
func (n *Node) readPrefix(ctx context.Context, c *connection) ([20]byte, error) {
	for {
		p, e := c.io.ReadPrefix(ctx, n.cfg.Timeouts.FrameRead)
		if e == nil {
			return p, nil
		}
		more, e := n.receiveEvent(c, e)
		if !more {
			return [20]byte{}, e
		}
	}
}
func (n *Node) readFrameInto(ctx context.Context, c *connection, dst []byte) error {
	for {
		e := c.io.ReadFrameInto(ctx, dst)
		if e == nil {
			return nil
		}
		more, e := n.receiveEvent(c, e)
		if !more {
			return e
		}
	}
}
func (n *Node) discardFrame(ctx context.Context, c *connection, size int) error {
	for {
		e := c.io.DiscardFrame(ctx, size)
		if e == nil {
			return nil
		}
		more, e := n.receiveEvent(c, e)
		if !more {
			return e
		}
	}
}

// Only application frames may use the post-handshake unordered mode. Mode is
// selected per call under the state gate, not stored as a mutable socket option.
func (n *Node) writeApplication(ctx context.Context, c *connection, b []byte) transport.WriteResult {
	if w, ok := c.io.(transport.MessageWriter); ok {
		n.mu.Lock()
		ordered := !c.sendBarrierUntil.IsZero()
		expired := ordered && !time.Now().Before(c.sendBarrierUntil)
		n.mu.Unlock()
		if expired {
			return transport.WriteResult{Err: context.DeadlineExceeded}
		}
		return w.WriteFrameOrdered(ctx, b, ordered)
	}
	return c.io.WriteFrame(ctx, b)
}
