package diameter

import (
	"errors"
	"time"
)

// ErrResponseWaitExpired ends a Ro segment whose final allowed attempt reached
// its response deadline. It does not imply that the connection has failed.
var ErrResponseWaitExpired = errors.New("Ro attempt response wait expired")

// One immutable duration per segment. Acceptance is the queue commit, not the
// start/end of WriteFrame and not the application business timeout.
func startRoResponseWait(x *exchange, now time.Time) {
	if x.group == nil || x.group.recovery == nil {
		return
	}
	x.acceptedAt = now
	if wait := x.group.recovery.responseWait; wait > 0 {
		x.responseWaitUntil = now.Add(wait)
	}
}

// Returns true when scheduling has stopped. A writer holding Q records expiry
// but remains AttemptActive; its ordinary frame-write deadline/connection failure
// settles delivery evidence before a successor may borrow the shared frame.
// The caller checks the logical deadline first; expiry never closes a healthy
// connection, changes watchdog state, frees A or automatically sends another CCR.
func (n *Node) expireRoResponseWaitLocked(x *exchange, now time.Time) bool {
	if x.selected || x.stopped || x.responseWaitUntil.IsZero() || !x.responseWaitExpired && now.Before(x.responseWaitUntil) {
		return false
	}
	x.responseWaitExpired = true
	if x.writing {
		return false
	}
	n.stopAttemptLocked(x, ErrResponseWaitExpired, StopResponseWait)
	return true
}
