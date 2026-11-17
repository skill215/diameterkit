// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/internal/groupavp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// FailurePolicy selects a compiled mapping, not arbitrary result codes/callbacks.
// The zero value preserves the earlier bounded close/terminate behavior.
type FailurePolicy uint8

const (
	CloseOnFailure FailurePolicy = iota
	// BaseErrorFailures enables the bounded mappings documented in spec/batch-07.md.
	BaseErrorFailures
)

// ProtocolFailure names an explicit local processing failure, never a success.
// FailureOverloaded requires a request explicitly addressed to the local host.
type ProtocolFailure uint8

const (
	FailureContextUnavailable ProtocolFailure = iota + 1
	FailureOverloaded
	FailureDeadlineExceeded
)

type rejectReason uint8

const (
	rejectSource rejectReason = iota + 1
	rejectDestination
	rejectOverload
	rejectDeadline
	rejectCommand
	rejectApplication
	rejectDecode
)

func rejectionReserve(cfg NodeConfig, e *codecengine.Engine) (uint64, error) {
	r := cfg.Resources
	if r.RejectSlots == 0 {
		if r.EmergencyRequestMax != 0 || r.EmergencyReplyMax != 0 {
			return 0, ErrConfig
		}
		return 0, nil
	}
	if r.RejectSlots < 0 || r.RejectSlots > 64 || r.EmergencyRequestMax < 20 || r.EmergencyRequestMax%4 != 0 || r.EmergencyRequestMax > cfg.Limits.MaxFrameBytes || r.EmergencyReplyMax < 20 || r.EmergencyReplyMax%4 != 0 || r.EmergencyReplyMax > cfg.Limits.MaxFrameBytes {
		return 0, ErrConfig
	}
	rb, err := wire.DecodeBound(r.EmergencyRequestMax, cfg.Limits, 4096)
	if err != nil {
		return 0, err
	}
	ab, err := wire.DecodeBound(r.EmergencyReplyMax, cfg.Limits, 4096)
	if err != nil {
		return 0, err
	}
	for _, pc := range cfg.Peers {
		for _, ip := range pc.InboundCommands {
			h := pdu.Header{Flags: pdu.RequestFlag | pdu.ProxiableFlag, ApplicationID: ip.Command.ApplicationID, CommandCode: ip.Command.Code}
			b, err := e.DecodeBound(h, r.EmergencyRequestMax)
			if err != nil {
				return 0, ErrConfig
			}
			rb = max(rb, b)
			h.Flags = pdu.ProxiableFlag | pdu.ErrorFlag
			b, err = e.DecodeBound(h, r.EmergencyReplyMax)
			if err != nil {
				return 0, ErrConfig
			}
			ab = max(ab, b)
		}
	}
	if rb > (math.MaxUint64-131072)/8 || ab > (math.MaxUint64-131072-8*rb)/8 {
		return 0, ErrConfig
	}
	b := 8*(rb+ab) + 131072
	if b > math.MaxUint64/uint64(r.RejectSlots) {
		return 0, ErrConfig
	}
	return b, nil
}

// rejectionEnvelope reads only provably framed top-level AVPs and complete Proxy
// groups. It never repairs an invalid/duplicate Session or invents a Proxy value.
func rejectionEnvelope(src []byte, l wire.Limits) (pdu.Header, pdu.Envelope, error) {
	h, size, err := wire.Header(src, l)
	if err != nil || size != len(src) || h.Flags&pdu.RequestFlag == 0 {
		return h, pdu.Envelope{}, ErrProtocol
	}
	d := wire.NewDecoder(src[20:], l)
	var env pdu.Envelope
	for {
		v, ok, err := d.Next()
		if err != nil {
			return h, env, err
		}
		if !ok {
			break
		}
		switch v.Key {
		case avp.Key{Code: 263}:
			if env.SessionID.Present || len(v.Data) == 0 || avp.UTF8(string(v.Data)) != nil {
				return h, env, validation.ErrEnvelope
			}
			env.SessionID = avp.Some(string(v.Data))
		case avp.Key{Code: 293}:
			if env.DestinationHost.Present || avp.Identity(string(v.Data)) != nil {
				return h, env, validation.ErrEnvelope
			}
			env.DestinationHost = avp.Some(string(v.Data))
		case avp.Key{Code: 284}:
			if len(env.ProxyInfo) >= l.MaxRepeated {
				return h, env, wire.ErrLimit
			}
			child, err := d.Child(v.Data)
			if err != nil {
				return h, env, err
			}
			proxy, err := base.DecodeProxy(child)
			if err != nil {
				return h, env, err
			}
			// Send validation is deliberate: trusted receive-only extras that cannot be
			// echoed legally are not dropped or rewritten to make rejection succeed.
			if err = base.ValidateProxy(proxy, validation.Send); err != nil {
				return h, env, err
			}
			env.ProxyInfo = append(env.ProxyInfo, *proxy)
		}
	}
	return h, env, nil
}
func missingDiagnostic(cause error) ([]byte, bool) {
	// A failed explicit proof must not fall back to the older untyped examples.
	var explicit interface {
		DiameterMissingAVPV1() (avp.Key, byte, int, error)
	}
	if errors.As(cause, &explicit) {
		return nil, false
	}
	var group1 interface {
		DiameterMissingGroupV1() (avp.Key, byte, string, error)
	}
	var group2 interface {
		DiameterMissingGroupV2() (avp.Key, byte, string, error)
	}
	if errors.As(cause, &group1) || errors.As(cause, &group2) {
		return nil, false
	}
	var field *validation.FieldError
	if !errors.Is(cause, validation.ErrMissing) || !errors.As(cause, &field) || len(field.Path) != 1 {
		return nil, false
	}
	k := field.Path[0]
	if k.VendorID != 0 {
		return nil, false
	}
	size := 8
	switch k.Code {
	case 263, 264, 296, 283:
	case 277:
		size = 12
	default:
		return nil, false
	}
	b := make([]byte, size)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	w.Bytes(k, avp.MandatoryFlag, b[8:])
	return b, true
}
func publicFailureBody(request pdu.Message, n *Node, f ProtocolFailure) (pdu.Body, error) {
	switch f {
	case FailureContextUnavailable, FailureDeadlineExceeded:
		return &pdu.ErrorAnswer{ResultCode: 5012}, nil
	case FailureOverloaded:
		op, ok := n.cfg.Bundle.Lookup(request.Header.Key())
		if !ok {
			return nil, ErrUnsupported
		}
		env, err := op.Envelope(request.Body)
		if err != nil {
			return nil, err
		}
		if !env.DestinationHost.Present || !sameOrigin(pdu.Origin{Host: env.DestinationHost.Value, Realm: n.cfg.Origin.Realm}, n.cfg.Origin) {
			return nil, ErrUnsupported
		}
		return &pdu.ErrorAnswer{ResultCode: 3004}, nil
	default:
		return nil, ErrUnsupported
	}
}

// Reject uses the same preparation, validation and single commit gate as Respond.
// It accepts only named failures from the enabled policy; it cannot send success.
func (in *Inbound) Reject(ctx context.Context, failure ProtocolFailure) (ResponseResult, error) {
	if in.policy.FailurePolicy != BaseErrorFailures || failure < FailureContextUnavailable || failure > FailureDeadlineExceeded {
		return ResponseResult{}, ErrUnsupported
	}
	return in.respond(ctx, nil, failure)
}

func (n *Node) rejectFrame(raw []byte, maxBytes int, reason rejectReason, cause error) ([]byte, error) {
	h, env, err := rejectionEnvelope(raw, n.cfg.Limits)
	if err != nil {
		return nil, err
	}
	code := uint32(0)
	var diagnostic []byte
	switch reason {
	case rejectSource, rejectDestination:
		code = 3002
	case rejectOverload:
		if !env.DestinationHost.Present || !sameOrigin(pdu.Origin{Host: env.DestinationHost.Value, Realm: n.cfg.Origin.Realm}, n.cfg.Origin) {
			return nil, ErrUnsupported
		}
		code = 3004
	case rejectDeadline:
		code = 5012
	case rejectCommand:
		code = 3001
	case rejectApplication:
		code = 3007
	case rejectDecode:
		var ok bool
		if diagnostic, code, ok = groupavp.Quote(raw[20:], cause, n.cfg.Limits, maxBytes); ok {
			break
		}
		if diagnostic, code, ok = excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, maxBytes); ok {
			break
		}
		if errors.Is(cause, validation.ErrDuplicate) {
			diagnostic, ok = excessavp.Quote(raw[20:], cause, n.cfg.Limits, maxBytes)
			if !ok {
				return nil, ErrUnsupported
			}
			code = 5009
			break
		}
		diagnostic, ok = missingavp.Quote(cause, n.cfg.Limits, maxBytes)
		if !ok {
			diagnostic, ok = missingDiagnostic(cause)
		}
		if !ok {
			return nil, ErrUnsupported
		}
		code = 5005
	default:
		return nil, ErrUnsupported
	}
	h.Flags = h.Flags&pdu.ProxiableFlag | pdu.ErrorFlag
	body := &pdu.ErrorAnswer{Command: h.Key(), SessionID: env.SessionID, OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, ResultCode: code, ProxyInfo: env.ProxyInfo}
	if diagnostic != nil {
		body.FailedAVP = avp.Some(pdu.FailedAVP{Data: diagnostic})
	}
	// Known commands retain their selected E grammar. Unknown commands use only
	// the fixed Base grammar, never an inferred application decoder or profile.
	var ops *schema.ErrorOperations
	if op, ok := n.cfg.Bundle.Lookup(h.Key()); ok {
		ops = op.Error
	} else {
		ops = base.NewErrorOperations(h.Key(), base.ErrorOptions{})
	}
	if ops == nil {
		return nil, ErrUnsupported
	}
	if err = ops.Validate(body, validation.Send); err != nil {
		return nil, err
	}
	size := wire.NewEncoder(nil, n.cfg.Limits)
	if err = ops.Encode(size, body); err != nil {
		return nil, err
	}
	if size.Err() != nil {
		return nil, size.Err()
	}
	if size.Len() > maxBytes-20 {
		return nil, wire.ErrLimit
	}
	frame := make([]byte, 20+size.Len())
	wire.PutHeader(frame, h, len(frame))
	w := wire.NewEncoder(frame[20:], n.cfg.Limits)
	if err = ops.Encode(w, body); err != nil {
		return nil, err
	}
	if w.Err() != nil {
		return nil, w.Err()
	}
	return frame, nil
}

// queueAutomaticLocked atomically shares Respond's decision and writer slot.
func (n *Node) queueAutomaticLocked(in *Inbound, frame []byte, terminal error) {
	if in.decided {
		return
	}
	n.commitDuplicateLocked(in, in.timeoutResult, frame[4], false)
	in.decided = true
	in.terminal = terminal
	in.result.Committed = true
	close(in.done)
	in.peer.inboundQueue = removeInbound(in.peer.inboundQueue, in)
	in.timeoutFrame = nil
	in.frame = frame
	in.result.PrimaryWrite.Delivery.FrameBytes = len(frame)
	in.writeDeadline = time.Now().Add(n.cfg.Timeouts.FrameWrite)
	if !n.mayWriteAnswerLocked(in, time.Now()) {
		if in.result.PrimaryWrite.Err == nil {
			in.result.PrimaryWrite.Err = ErrNotReady
		}
		n.releaseInboundLocked(in)
	} else {
		in.conn.enqueueAnswer(in)
		in.conn.signal()
	}
	n.notify()
}
func (n *Node) expireInboundLocked(in *Inbound) {
	if in.decided {
		return
	}
	if in.duplicate != nil && in.duplicate.owner != in {
		if owner := in.duplicate.owner; owner != nil {
			n.expireInboundLocked(owner)
		} else {
			n.endInboundLocked(in, context.DeadlineExceeded)
		}
		return
	}
	if len(in.timeoutFrame) > 0 && in.policy.FailurePolicy == BaseErrorFailures {
		n.queueAutomaticLocked(in, in.timeoutFrame, context.DeadlineExceeded)
	} else {
		n.endInboundLocked(in, context.DeadlineExceeded)
	}
}
func (n *Node) readEmergency(c *connection, h pdu.Header, size int, reason rejectReason, policy *inboundPolicy) error {
	n.mu.Lock()
	if n.closed || n.shutting || !c.peer.readyLocked() || c.peer.conn != c {
		n.mu.Unlock()
		return ErrNotReady
	}
	r := n.cfg.Resources
	if size > r.EmergencyRequestMax || n.rejectBound == 0 {
		n.mu.Unlock()
		return wire.ErrLimit
	}
	if n.seq == math.MaxUint64 {
		n.mu.Unlock()
		return ErrIdentifiers
	}
	charge, err := budget.Reserve(n.rejectBound, 1, "emergency-reject", &n.rejectAccount)
	if err != nil {
		n.mu.Unlock()
		return ErrBackpressure
	}
	n.seq++
	id := RequestID{n.boot, n.seq}
	in := &Inbound{emergency: true, node: n, peer: c.peer, conn: c, charge: charge, busy: true, info: InboundInfo{id, c.peer.id, c.id, time.Now().Add(n.cfg.Timeouts.FrameRead)}, done: make(chan struct{}), finished: make(chan struct{})}
	n.inbounds[id] = in
	n.mu.Unlock()
	n.notify()
	raw := make([]byte, size)
	err = n.readFrameInto(c.ctx, c, raw)
	readErr := err
	var frame []byte
	if err == nil {
		var cause error
		if policy != nil {
			var request pdu.Message
			request, cause = n.engine.DecodeOwned(raw)
			if cause != nil {
				reason = rejectDecode
			} else {
				cause = checkInboundBinding(request, *policy, n)
				if errors.Is(cause, validation.ErrSource) {
					reason = rejectSource
				} else if cause != nil {
					reason = rejectDestination
				}
			}
		}
		frame, err = n.rejectFrame(raw, r.EmergencyReplyMax, reason, cause)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	in.busy = false
	if err == nil && (in.decided || c.closed || n.closed) {
		err = ErrClosed
	}
	if err != nil {
		n.endInboundLocked(in, err)
		n.releaseInboundLocked(in)
		if readErr != nil {
			return transportFailure{readErr}
		}
		return err
	}
	n.queueAutomaticLocked(in, frame, ErrInboundDone)
	return nil
}
