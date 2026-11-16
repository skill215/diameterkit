// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func errorRecord(t *testing.T, n *Node, body pdu.Body, policy string) RoRecoveryRequest {
	o := recordOptions()
	o.PolicyID = policy
	return prepareRecord(t, n, body, o)
}
func errorPlan(mode RoErrorContinuation) RoRecoveryPlan {
	p := segmentPlan(1)
	p.ErrorContinuation = mode
	if mode == RoErrorSMSFWBeforeTx {
		p.ErrorContinuationUntil = time.Now().Add(time.Second)
	}
	return p
}
func errorReply(t *testing.T, co *codec.Codec, remote transport.Conn, req pdu.Message, code uint32) {
	t.Helper()
	if code >= 3000 && code < 4000 {
		a, e := co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: code}, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
	} else {
		baseCode := code
		if code == 4002 {
			baseCode = 5012
		}
		body := roRecoveryAnswer(req.Body, baseCode)
		switch a := body.(type) {
		case *ro.CCA:
			a.FailedAVP = ratingDiagnostic()
		case *smsfwro.CCA:
			a.FailedAVP = ratingDiagnostic()
		case *legacy420.CCA:
			a.FailedAVP = ratingDiagnostic()
		case *combined.CCA:
			a.FailedAVP = ratingDiagnostic()
		}
		a, e := co.CompleteAnswer(req, body, serverOrigin)
		if e != nil {
			t.Fatalf("code %d: %v", code, e)
		}
		if code == 4002 {
			raw, e := co.Append(nil, a)
			if e != nil {
				t.Fatal(e)
			}
			d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
			for {
				v, ok, e := d.Next()
				if e != nil {
					t.Fatal(e)
				}
				if !ok {
					break
				}
				if v.Key == (avp.Key{Code: 268}) {
					binary.BigEndian.PutUint32(v.Data, 4002)
				}
			}
			sendRejectRaw(t, remote, raw)
		} else if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
	}
}
func TestRoErrorContinuationPolicyAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 2)
		r := errorRecord(t, n, roBody(), RoTransientRecoveryPolicy)
		for _, p := range []RoRecoveryPlan{errorPlan(RoErrorSMSFWBeforeTx), errorPlan(99), func() RoRecoveryPlan {
			p := errorPlan(RoErrorTransient)
			p.ErrorContinuationUntil = time.Now()
			return p
		}()} {
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, segmentProgress(1), port); !errors.Is(e, ErrConfig) {
				t.Fatal(e)
			}
		}
		p := errorPlan(RoErrorTransient)
		id := submitSegment(t, n, r, p, segmentProgress(1), port)
		req := routeRead(t, co, remote)
		changed := p
		changed.ErrorContinuation = RoErrorFinal
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, changed, segmentProgress(1), port); !errors.Is(e, ErrRoRecoveryConflict) {
			t.Fatal(e)
		}
		errorReply(t, co, remote, req, 3004)
		got := routeCompletion(t, n, port, id)
		if !got.RoRecovery.CanContinue || got.RoRecovery.NextTrigger != RoRecoveryErrorAnswer {
			t.Fatal(got)
		}
		p.Trigger = RoRecoveryNoAnswer
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, nextSegmentProgress(got, 1), port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal(e)
		}
		// Caller-owned completion mutation cannot forge a no-answer trigger or erase history.
		progress := nextSegmentProgress(got, 1)
		progress.Delivery = DeliveryEvidence{}
		p.Trigger = RoRecoveryErrorAnswer
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, progress, port); !errors.Is(e, ErrRoRecoveryProgress) {
			t.Fatal(e)
		}
		id = submitSegment(t, n, r, p, nextSegmentProgress(got, 1), port)
		routeRead(t, co, remote)
		n.Cancel(id)
		routeCompletion(t, n, port, id)
	})
}
func TestRoErrorContinuationCodeMatrix(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, family := range []string{"standard", "sms", "money", "combined"} {
			for _, mode := range []RoErrorContinuation{RoErrorFinal, RoErrorTransient, RoErrorSMSFWBeforeTx} {
				for _, code := range []uint32{3002, 3004, 3005, 4002, 5001, 5004, 5005, 5008, 5012, 2001, 4012, 5011} {
					n, co := newInboundNode(t, recordConfig(t, family))
					_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
					port := portTest(t, n, 1)
					body := roRecoveryBody(t, family, "iec")
					if family == "money" {
						body = roRecoveryBody(t, family, "refund")
					}
					r := errorRecord(t, n, body, RoSMSFWRecoveryPolicy)
					p := errorPlan(mode)
					id := submitSegment(t, n, r, p, segmentProgress(1), port)
					req := routeRead(t, co, remote)
					errorReply(t, co, remote, req, code)
					got := routeCompletion(t, n, port, id)
					transient := code == 3002 || code == 3004 || code == 3005
					extra := code == 4002 || code == 5001 || code == 5004 || code == 5005 || code == 5008 || code == 5012
					want := mode != RoErrorFinal && transient || mode == RoErrorSMSFWBeforeTx && extra
					if got.Err != nil || got.Outcome != AnswerReceived || got.RoRecovery.CanContinue != want {
						t.Fatalf("%s mode%d code%d: %+v recovery%+v", family, mode, code, got, got.RoRecovery)
					}
					if code == 4002 {
						if got.Validation.Complete || !errors.Is(got.Validation.Err(), validation.ErrUnsupported) {
							t.Fatal("4002 semantic claim", got.Validation)
						}
						if _, e := co.Append(nil, got.Answer); !errors.Is(e, validation.ErrUnsupported) {
							t.Fatal("4002 send enabled", e)
						}
					}
					if !want {
						p.Trigger = RoRecoveryErrorAnswer
						if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, nextSegmentProgress(got, 1), port); !errors.Is(e, ErrRoRecoveryFinal) {
							t.Fatal(e)
						}
					}
					segmentIdle(t, n)
				}
			}
		}
	})
}
func TestRoErrorContinuationOriginalIdentityTAndFinal(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, a := roRecoveryNetwork(t, n, co, "dra", tls)
		_, b := roRecoveryNetwork(t, n, co, "dra2", tls)
		port := portTest(t, n, 2)
		r := errorRecord(t, n, roBody(), RoSMSFWRecoveryPolicy)
		p := errorPlan(RoErrorSMSFWBeforeTx)
		id := submitSegment(t, n, r, p, segmentProgress(1), port)
		first := routeRead(t, co, a)
		errorReply(t, co, a, first, 5012)
		got := routeCompletion(t, n, port, id)
		p.Trigger = RoRecoveryErrorAnswer
		p.MaxAttempts = 2
		p.ResponseWait = 500 * time.Millisecond
		id = submitSegment(t, n, r, p, nextSegmentProgress(got, 2), port)
		second := routeRead(t, co, a)
		encoded, e := co.Append(nil, second)
		if e != nil {
			t.Fatal(e)
		}
		if second.Header.Flags&pdu.RetransmittedFlag != 0 || second.Header.HopByHop == first.Header.HopByHop || !bytes.Equal(encoded[16:], r.Frame[16:]) {
			t.Fatal("identity or T after error")
		}
		// A previous segment's late answer cannot complete the new local exchange.
		errorReply(t, co, a, first, 2001)
		i := trExpire(t, n, id)
		// Reading the frame at the remote endpoint does not join the local
		// writer. Tr deliberately leaves a writing attempt active until that
		// borrower settles; observe the replay gate before requesting recovery.
		waitFor(t, func() bool { return trInfo(t, n, id).ReplayEligible })
		i = trInfo(t, n, id)
		if _, e = n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		third := routeRead(t, co, b)
		routeSameRequest(t, co, second, third)
		errorReply(t, co, b, third, 2001)
		got = routeCompletion(t, n, port, id)
		if got.RoRecovery.CanContinue || got.RoRecovery.NextTrigger != 0 || !got.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
			t.Fatal(got)
		}
		p.MaxAttempts = 1
		if _, e = n.SubmitRoRecovery(ctxTest(t), r, p, nextSegmentProgress(got, 1), port); !errors.Is(e, ErrRoRecoveryFinal) {
			t.Fatal(e)
		}
	})
}
func TestRoErrorContinuationNoWriteKeepsResetLineage(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		_, a := roRecoveryNetwork(t, n, co, "dra", tls)
		local, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: local, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		peer := roRecoveryConnect(t, n, co, "dra2", barrier, b)
		port := portTest(t, n, 2)
		r := errorRecord(t, n, roBody(), RoTransientRecoveryPolicy)
		p := errorPlan(RoErrorTransient)
		id := submitSegment(t, n, r, p, segmentProgress(1), port)
		req := routeRead(t, co, a)
		errorReply(t, co, a, req, 3004)
		first := routeCompletion(t, n, port, id)
		blockPort := portTest(t, n, 1)
		block, e := peer.Submit(ctxTest(t), roBody(), RequestOptions{}, blockPort)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, b)
		<-barrier.entered
		p.Trigger = RoRecoveryErrorAnswer
		p.Peers = []PeerID{"dra2"}
		id = submitSegment(t, n, r, p, nextSegmentProgress(first, 1), port)
		n.Cancel(id)
		second := routeCompletion(t, n, port, id)
		if second.Outcome != NotWritten || !second.RoRecovery.CumulativeDelivery.MayHaveReachedPeer || second.RoRecovery.NextTrigger != RoRecoveryNoAnswer {
			t.Fatal(second)
		}
		p.Trigger = RoRecoveryNoAnswer
		p.Peers = []PeerID{"dra"}
		id = submitSegment(t, n, r, p, nextSegmentProgress(second, 1), port)
		req = routeRead(t, co, a)
		if req.Header.Flags&pdu.RetransmittedFlag != 0 {
			t.Fatal("historical delivery resurrected pre-error T")
		}
		errorReply(t, co, a, req, 2001)
		routeCompletion(t, n, port, id)
		n.Cancel(block)
		once.Do(func() { close(barrier.release) })
		routeCompletion(t, n, blockPort, block)
		segmentIdle(t, n)
	})
}
func TestRoErrorContinuationBusinessWindow(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{5012, 3004} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			local, remote := networkAdapters(t, tls)
			barrier := &roRecoveryBarrier{Conn: local, entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			defer once.Do(func() { close(barrier.release) })
			roRecoveryConnect(t, n, co, "dra", barrier, remote)
			port := portTest(t, n, 1)
			r := errorRecord(t, n, roBody(), RoSMSFWRecoveryPolicy)
			p := errorPlan(RoErrorSMSFWBeforeTx)
			p.ErrorContinuationUntil = time.Now().Add(20 * time.Millisecond)
			id := submitSegment(t, n, r, p, segmentProgress(1), port)
			req := routeRead(t, co, remote)
			errorReply(t, co, remote, req, code)
			<-barrier.entered
			waitFor(t, func() bool { return !time.Now().Before(p.ErrorContinuationUntil) })
			once.Do(func() { close(barrier.release) })
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.CanContinue != (code == 3004) {
				t.Fatal("post-Tx code eligibility", got)
			}
		}
	})
}
func TestRoErrorContinuationWriterRetirementAndWindowCapture(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		local, remote := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: local, write: true, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, remote)
		port := portTest(t, n, 2)
		r := errorRecord(t, n, roBody(), RoSMSFWRecoveryPolicy)
		p := errorPlan(RoErrorSMSFWBeforeTx)
		p.ErrorContinuationUntil = time.Now().Add(100 * time.Millisecond)
		id := submitSegment(t, n, r, p, segmentProgress(1), port)
		req := routeRead(t, co, remote)
		<-barrier.entered
		errorReply(t, co, remote, req, 5012)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		next := segmentProgress(1)
		next.SegmentNo = 2
		next.ReservedBefore = 1
		next.ReservedThrough = 2
		if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, next, port); !errors.Is(e, ErrRoRecoveryBusy) {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return !time.Now().Before(p.ErrorContinuationUntil) })
		once.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if !got.RoRecovery.CanContinue || !got.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
			t.Fatal("writer retirement lost answer-window evidence", got)
		}
	})
}
func TestRoErrorContinuationInvalidAndExperimentalResults(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"source", "request-number", "experimental"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			r := errorRecord(t, n, roBody(), RoSMSFWRecoveryPolicy)
			p := errorPlan(RoErrorSMSFWBeforeTx)
			id := submitSegment(t, n, r, p, segmentProgress(1), port)
			req := routeRead(t, co, remote)
			body := roAnswer(5012)
			origin := serverOrigin
			switch kind {
			case "source":
				origin.Host = "foreign.example"
			case "request-number":
				// Mutate received bytes after valid local construction.
			case "experimental":
				// Append received-only experimental bytes after normal construction.
			}
			ans, e := co.CompleteAnswer(req, body, origin)
			if e != nil {
				t.Fatal(e)
			}
			if kind == "request-number" {
				raw, e := co.Append(nil, ans)
				if e != nil {
					t.Fatal(e)
				}
				d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
				for {
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					if v.Key == (avp.Key{Code: 415}) {
						binary.BigEndian.PutUint32(v.Data, 99)
					}
				}
				sendRejectRaw(t, remote, raw)
			} else if kind == "experimental" {
				sendRejectRaw(t, remote, resultWire(t, co, ans, true, 10415, false))
			} else if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.CanContinue {
				t.Fatal(kind, got)
			}
			p.Trigger = RoRecoveryErrorAnswer
			if _, e = n.SubmitRoRecovery(ctxTest(t), r, p, nextSegmentProgress(got, 1), port); !errors.Is(e, ErrRoRecoveryFinal) {
				t.Fatal(e)
			}
		}
	})
}
func TestRoErrorContinuationLimitsAndDisabledPolicy(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"slots", "attempts", "unknown-policy"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			opts := recordOptions()
			opts.PolicyID = RoSMSFWRecoveryPolicy
			switch kind {
			case "slots":
				opts.Limits.MaxSegments = 1
			case "attempts":
				opts.Limits.MaxSegments = 1
				opts.Limits.MaxTotalAttempts = 1
				opts.Limits.MaxAttemptsPerSegment = 1
			case "unknown-policy":
				opts.PolicyID = "operator-name-is-not-permission"
			}
			r := prepareRecord(t, n, roBody(), opts)
			p := errorPlan(RoErrorTransient)
			if kind == "unknown-policy" {
				if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, segmentProgress(1), port); !errors.Is(e, ErrConfig) {
					t.Fatal(e)
				}
				p.ErrorContinuation = RoErrorFinal
			}
			id := submitSegment(t, n, r, p, segmentProgress(1), port)
			req := routeRead(t, co, remote)
			errorReply(t, co, remote, req, 3004)
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.CanContinue || got.RoRecovery.NextTrigger != 0 {
				t.Fatal(got)
			}
		}
	})
}

func TestRoErrorContinuationCloseWhileRetiring(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, end := range []string{"peer", "node", "port"} {
			n, co := newInboundNode(t, recordConfig(t, "standard"))
			local, remote := networkAdapters(t, tls)
			barrier := &roRecoveryBarrier{Conn: local, write: true, entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			defer once.Do(func() { close(barrier.release) })
			peer := roRecoveryConnect(t, n, co, "dra", barrier, remote)
			port := portTest(t, n, 2)
			r := errorRecord(t, n, roBody(), RoTransientRecoveryPolicy)
			p := errorPlan(RoErrorTransient)
			id := submitSegment(t, n, r, p, segmentProgress(1), port)
			req := routeRead(t, co, remote)
			<-barrier.entered
			errorReply(t, co, remote, req, 3004)
			waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
			var closed chan error
			switch end {
			case "peer":
				peer.Close()
			case "node":
				n.Close()
			case "port":
				closed = make(chan error, 1)
				go func() { closed <- port.Close(ctxTest(t)) }()
				waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return port.closed })
			}
			once.Do(func() { close(barrier.release) })
			if end == "port" {
				if e := <-closed; e != nil {
					t.Fatal(e)
				}
				segmentIdle(t, n)
			} else {
				got := routeCompletion(t, n, port, id)
				if got.RoRecovery.CanContinue {
					t.Fatal("close granted continuation", got)
				}
			}
			n.mu.Lock()
			slot, exists := n.roRecovery.byE2E[req.Header.EndToEnd]
			allowed := exists && n.roRecovery.slots[slot].continuable
			n.mu.Unlock()
			if !exists && end != "node" {
				t.Fatal("guard released before its retention deadline")
			}
			if allowed {
				t.Fatal("closing guard remains continuable")
			}
		}
	})
}

func TestRoErrorContinuationBoundedRoundsAndAdmissionRollback(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, recordConfig(t, "standard"))
		peer, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		opts := recordOptions()
		opts.PolicyID = RoTransientRecoveryPolicy
		opts.Limits.MaxSegments = 3
		r := prepareRecord(t, n, roBody(), opts)
		p := errorPlan(RoErrorTransient)
		progress := segmentProgress(2)
		for round := uint32(1); round <= 3; round++ {
			id := submitSegment(t, n, r, p, progress, port)
			req := routeRead(t, co, remote)
			if req.Header.Flags&pdu.RetransmittedFlag != 0 {
				t.Fatal("error continuation T")
			}
			errorReply(t, co, remote, req, 3004)
			got := routeCompletion(t, n, port, id)
			if got.RoRecovery.SegmentNo != round || got.RoRecovery.ReservedThrough != round*2 || got.RoRecovery.CanContinue != (round < 3) {
				t.Fatal(got)
			}
			if round == 3 {
				break
			}
			progress = nextSegmentProgress(got, 2)
			p.Trigger = RoRecoveryErrorAnswer
			n.mu.Lock()
			limit := peer.account.LimitBytes
			peer.account.LimitBytes = 1
			n.mu.Unlock()
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, progress, port); !errors.Is(e, ErrBackpressure) {
				t.Fatal(e)
			}
			n.mu.Lock()
			peer.account.LimitBytes = limit
			slot := n.roRecovery.byE2E[req.Header.EndToEnd]
			g := n.roRecovery.slots[slot]
			n.mu.Unlock()
			if g.segmentNo != round || g.reservedThrough != round*2 || !g.continuable {
				t.Fatal("rejected admission consumed progress", g)
			}
		}
		segmentIdle(t, n)
	})
}
