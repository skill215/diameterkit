//go:build linux && !386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/skill215/diameterkit/transport"
	upstream "github.com/skill215/sctp"
)

type messageSocket interface {
	ReadEvent(context.Context, []byte, []byte, []byte) (upstream.ReadResult, error)
	WriteMsgContext(context.Context, []byte, upstream.SndInfo) (int, error)
	Close() error
}
type conn struct {
	socket            messageSocket
	config            Config
	info              transport.Info
	mu                sync.Mutex
	cause             error
	closed            bool
	once              sync.Once
	closeErr          error
	prefix            [20]byte
	scratch           [256]byte
	notification, oob []byte
	phase             uint8 // 0 prefix,1 classified,2 body,3 discard
	offset, size      int
	eor               bool
	deadline          time.Time
	timeout           time.Duration
	sequence          uint64
	window            time.Time
	notifications     uint32
	inbound           uint16
	outbound          uint16
	association       int32
}

// New takes exclusive ownership even on failure. The caller must relinquish
// all I/O, options, deadlines and raw-fd access. It verifies message mode/options
// and negotiated streams, clears legacy socket timeouts, and takes no address
// copies. Use Config.Requirements before Node construction; Node checks Info
// before CER/CEA. Use NewDialer/Listen for Node-managed setup admission.
func New(raw *upstream.SCTPConn, cfg Config) (transport.Conn, error) {
	if raw == nil {
		return nil, ErrConfig
	}
	fail := func(e error) (transport.Conn, error) { _ = raw.Close(); return nil, e }
	info, e := cfg.Requirements()
	if e != nil {
		return fail(e)
	}
	state, e := raw.GetMessageOptions()
	if e != nil {
		return fail(e)
	}
	if state.Options != *SocketConfig().Messages {
		return fail(ErrConfig)
	}
	status, e := raw.GetStatus()
	if e != nil {
		return fail(e)
	}
	if status.State != upstream.SCTP_ESTABLISHED || status.Instreams == 0 || status.Ostreams == 0 {
		return fail(ErrConfig)
	}
	if e = raw.SetSocketTimeouts(0, 0); e != nil {
		return fail(e)
	}
	if e = raw.SetDeadline(time.Time{}); e != nil {
		return fail(e)
	}
	req, _ := upstream.MessageIORequirements(state.Options)
	return &conn{socket: raw, config: cfg, info: info, notification: make([]byte, cfg.NotificationMax), oob: make([]byte, req.OOBMin), inbound: status.Instreams, outbound: status.Ostreams, association: int32(status.AssocID)}, nil
}
func (s *conn) Info() transport.Info { return s.info }
func (s *conn) SCTPMode() ModeInfo {
	return ModeInfo{StreamMode: s.config.StreamMode, PPIDPolicy: s.config.PPIDPolicy, InboundStreams: s.inbound, OutboundStreams: s.outbound, SingleStreamFallback: s.config.StreamMode == OrderedMultiStream && s.outbound == 1}
}

// seal chooses the first adapter-level terminal cause before waking the other
// direction. A writer's later net.ErrClosed cannot downgrade a local quota or
// framing failure into retryable transport loss. Node-local causes still win at
// the endpoint's own state gate.
func (s *conn) seal(cause error) error {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		s.cause = cause
	}
	first := s.cause
	s.mu.Unlock()
	s.once.Do(func() { s.closeErr = s.socket.Close() })
	return first
}
func (s *conn) Close() error { s.seal(net.ErrClosed); return s.closeErr }
func (s *conn) closedError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		return nil
	}
	return errors.Join(net.ErrClosed, s.cause)
}
func (s *conn) failure(e error) error { return s.seal(e) }
func earlier(a, b time.Time) time.Time {
	if a.IsZero() || !b.IsZero() && b.Before(a) {
		return b
	}
	return a
}

// read consumes one fragment or yields one owned event. No dst or notification
// tail survives return. Context timers end before the upstream borrower retires.
func (s *conn) read(ctx context.Context, dst []byte) (int, error) {
	if e := s.closedError(); e != nil {
		return 0, e
	}
	if ctx == nil {
		return 0, s.failure(ErrConfig)
	}
	if err := ctx.Err(); err != nil {
		return 0, s.failure(err)
	}
	if !s.deadline.IsZero() {
		if !time.Now().Before(s.deadline) {
			return 0, s.failure(context.DeadlineExceeded)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, s.deadline)
		defer cancel()
	}
	r, e := s.socket.ReadEvent(ctx, dst, s.notification, s.oob)
	// Restart and partial-abort may deliver valid raw notifications together
	// with a terminal marker. Project their facts before handling generic errors;
	// a malformed/failed notification must never become a continuation.
	if r.Kind == upstream.ReadNotification && r.Consumed && (e == nil || errors.Is(e, upstream.ErrAssociationChanged) && !errors.Is(e, upstream.ErrMessageBoundary)) {
		if r.N < 8 || r.N > len(s.notification) {
			return 0, s.failure(transport.ErrFrameState)
		}
		return 0, s.event(s.notification[:r.N], e, r.CloseRequired)
	}
	if e != nil {
		if errors.Is(e, upstream.ErrNotificationTooLarge) {
			e = errors.Join(transport.ErrNotificationLimit, e)
		}
		return 0, s.failure(e)
	}
	if r.CloseRequired {
		return 0, s.failure(transport.ErrFrameState)
	}
	if r.Kind == upstream.ReadEOF {
		return 0, s.failure(io.EOF)
	}
	if r.Kind != upstream.ReadData || !r.Consumed || r.N <= 0 || r.N > len(dst) {
		return 0, s.failure(transport.ErrFrameState)
	}
	if r.Info.PPID != 46 && !(s.config.PPIDPolicy == PPID0Compatibility && r.Info.PPID == 0) {
		return 0, s.failure(transport.ErrFrameState)
	}
	if s.offset == 0 {
		s.deadline = time.Now().Add(s.timeout)
		if d, ok := ctx.Deadline(); ok {
			s.deadline = earlier(s.deadline, d)
		}
	}

	s.offset += r.N
	s.eor = r.EndOfRecord()
	return r.N, nil
}
func (s *conn) ReadPrefix(ctx context.Context, timeout time.Duration) ([20]byte, error) {
	if s.phase != 0 || timeout <= 0 {
		return [20]byte{}, transport.ErrFrameState
	}
	if s.offset == 0 {
		s.timeout = timeout
	}
	for s.offset < 20 {
		_, e := s.read(ctx, s.prefix[s.offset:])
		if e != nil {
			return [20]byte{}, e
		}
		if s.eor && s.offset < 20 {
			return [20]byte{}, s.failure(transport.ErrFrameState)
		}
	}
	s.size = int(binary.BigEndian.Uint32(s.prefix[:4]) & 0xffffff)
	if s.size < 20 || s.eor && s.size != 20 || !s.eor && s.size == 20 {
		return [20]byte{}, s.failure(transport.ErrFrameState)
	}
	s.phase = 1
	return s.prefix, nil
}
func (s *conn) finish() error {
	if s.offset != s.size || !s.eor {
		return s.failure(transport.ErrFrameState)
	}
	s.offset = 0
	s.size = 0
	s.phase = 0
	s.deadline = time.Time{}
	s.eor = false
	return nil
}
func (s *conn) checkBody(ctx context.Context) error {
	if e := s.closedError(); e != nil {
		return e
	}
	if ctx == nil {
		return s.failure(ErrConfig)
	}
	if e := ctx.Err(); e != nil {
		return s.failure(e)
	}
	if !time.Now().Before(s.deadline) {
		return s.failure(context.DeadlineExceeded)
	}
	return nil
}
func (s *conn) ReadFrameInto(ctx context.Context, dst []byte) error {
	if s.phase != 1 && s.phase != 2 || len(dst) != s.size {
		return transport.ErrFrameState
	}
	if e := s.checkBody(ctx); e != nil {
		return e
	}
	if s.phase == 1 {
		copy(dst, s.prefix[:])
		s.phase = 2
	}
	for s.offset < s.size {
		if _, e := s.read(ctx, dst[s.offset:]); e != nil {
			return e
		}
		if s.eor && s.offset != s.size {
			return s.failure(transport.ErrFrameState)
		}
	}
	return s.finish()
}
func (s *conn) DiscardFrame(ctx context.Context, size int) error {
	if s.phase != 1 && s.phase != 3 || size != s.size {
		return transport.ErrFrameState
	}
	if e := s.checkBody(ctx); e != nil {
		return e
	}
	s.phase = 3
	for s.offset < s.size {
		if _, e := s.read(ctx, s.scratch[:min(len(s.scratch), s.size-s.offset)]); e != nil {
			return e
		}
		if s.eor && s.offset != s.size {
			return s.failure(transport.ErrFrameState)
		}
	}
	return s.finish()
}
func (s *conn) WriteFrame(ctx context.Context, b []byte) transport.WriteResult {
	return s.WriteFrameOrdered(ctx, b, true)
}
func (s *conn) WriteFrameOrdered(ctx context.Context, b []byte, ordered bool) transport.WriteResult {
	if e := s.closedError(); e != nil {
		return transport.WriteResult{Err: e}
	}
	if ctx == nil {
		return transport.WriteResult{Err: ErrConfig}
	}
	if e := ctx.Err(); e != nil {
		return transport.WriteResult{Err: e}
	}
	if len(b) < 20 || len(b) > 0xffffff || int(binary.BigEndian.Uint32(b[:4])&0xffffff) != len(b) {
		return transport.WriteResult{Err: transport.ErrFrameState}
	}
	i, e := sendInfo(s.config, s.outbound, b, ordered)
	if e != nil {
		return transport.WriteResult{Err: e}
	}
	// One call, one message. Context=0 intentionally disables attempt association.
	n, e := s.socket.WriteMsgContext(ctx, b, i)
	// The writer may see the restart latch before the reader commits its event.
	// Preserve a transport cause without inventing a notification sequence.
	if errors.Is(e, upstream.ErrAssociationChanged) {
		e = errors.Join(transport.ErrAssociationClosed, e)
	}
	uncertain := n < 0 || n > len(b)
	if uncertain {
		n = max(0, min(n, len(b)))
	}
	if e == nil && (n != len(b) || uncertain) {
		e = io.ErrShortWrite
	}
	if e != nil {
		e = s.failure(e)
	}
	return transport.WriteResult{Bytes: n, Uncertain: uncertain, Err: e}
}
func (s *conn) event(b []byte, cause error, closeRequired bool) error {
	now := time.Now()
	if s.window.IsZero() || now.Sub(s.window) >= s.config.NotificationWindow {
		s.window = now
		s.notifications = 0
	}
	if s.notifications >= s.config.MaxNotificationsPerWindow {
		return s.failure(transport.ErrNotificationLimit)
	}
	s.notifications++
	n, e := upstream.ParseNotification(b)
	if e != nil {
		return s.failure(e)
	}
	if s.sequence == math.MaxUint64 {
		return s.failure(transport.ErrNotificationLimit)
	}
	s.sequence++
	v := &transport.Event{Sequence: s.sequence, Progress: s.offset, Type: n.Header.Type, Flags: n.Header.Flags}
	switch n.Kind {
	case upstream.NotificationAssociationChange:
		a := n.Association
		v.Association = int32(a.AssocID)
		v.State = int64(a.State)
		v.Code = uint32(a.Error)
		v.InboundStreams = a.InboundStreams
		v.OutboundStreams = a.OutboundStreams
		switch a.State {
		case upstream.SCTP_COMM_UP:
			v.Kind = transport.EventAssociationUp
		case upstream.SCTP_RESTART:
			v.Kind = transport.EventAssociationRestart
			v.Terminal = true
		case upstream.SCTP_COMM_LOST, upstream.SCTP_CANT_STR_ASSOC, upstream.SCTP_SHUTDOWN_COMP:
			v.Kind = transport.EventAssociationLost
			v.Terminal = true
		}
	case upstream.NotificationPeerAddressChange:
		a := n.PeerAddress
		v.Kind = transport.EventPathChange
		v.Association = int32(a.AssocID)
		v.State = int64(a.State)
		v.Code = uint32(a.Error)
		v.Address = a.Address.IP
		v.Port = a.Address.Port
		v.ScopeID = a.Address.ScopeID
	case upstream.NotificationSendFailure:
		a := n.SendFailure
		v.Kind = transport.EventSendFailure
		v.Association = int32(a.AssocID)
		v.Code = a.Error
		v.Stream = a.Info.Stream
		v.PPID = a.Info.PPID
		v.Context = a.Info.Context
	case upstream.NotificationShutdown:
		v.Kind = transport.EventShutdown
		v.Terminal = true
		v.Association = int32(n.Shutdown.AssocID)
	case upstream.NotificationPartialDelivery:
		a := n.PartialDelivery
		v.Association = int32(a.AssocID)
		v.State = int64(a.Indication)
		if a.Indication == upstream.PartialDeliveryAborted {
			v.Kind = transport.EventPartialDeliveryAborted
			v.Terminal = true
		}
	}
	if n.Kind != upstream.NotificationUnknown && v.Association != s.association {
		return s.failure(transport.ErrFrameState)
	}
	if v.Kind == transport.EventAssociationUp && (v.InboundStreams != s.inbound || v.OutboundStreams != s.outbound) {
		return s.failure(transport.ErrFrameState)
	}
	if closeRequired && !v.Terminal || cause != nil && v.Kind != transport.EventAssociationRestart {
		return s.failure(errors.Join(transport.ErrFrameState, cause))
	}
	if v.Terminal {
		return s.failure(errors.Join(v, cause))
	}
	return v
}

var _ transport.Conn = (*conn)(nil)
