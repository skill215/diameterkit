// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package transport

import (
	"errors"
	"net/netip"
)

var (
	ErrAssociationClosed = errors.New("transport association terminated")
	ErrNotificationLimit = errors.New("transport notification budget exceeded")
)

type EventKind uint8

const (
	EventUnknown EventKind = iota
	EventAssociationUp
	EventAssociationLost
	EventAssociationRestart
	EventShutdown
	EventPartialDeliveryAborted
	EventPathChange
	EventSendFailure
)

// Event is a synchronous, owned receive continuation, never a completed frame.
// The caller processes it then resumes the same read method, context, target
// length and (for ReadFrameInto) unchanged destination. Progress is cumulative.
// There is no payload tail, queue, callback or borrowed data. Sequence is local
// to this connection; the endpoint attaches its own ConnectionID.
// Terminal events require Close and forbid continuation. The adapter seals
// write admission before returning one, even before the endpoint handles it.
// This implements error solely to preserve Conn's two-stage API: callers must
// recognize *Event before treating a receive result as an I/O failure.
type Event struct {
	Kind                            EventKind
	Sequence                        uint64
	Progress                        int
	Terminal                        bool
	Type, Flags                     uint16
	State                           int64
	Code                            uint32
	Association                     int32
	Address                         netip.Addr
	Port                            uint16
	ScopeID                         uint32
	InboundStreams, OutboundStreams uint16
	Stream                          uint16
	PPID, Context                   uint32
}

func (e *Event) Error() string { return "transport receive event" }
func (e *Event) Unwrap() error {
	if e.Terminal {
		return ErrAssociationClosed
	}
	return nil
}
