// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

var crossOrigin = pdu.Origin{Host: "secondary-long-ocs.example", Realm: "secondary.example"}

func crossConfig(t *testing.T, family string) NodeConfig {
	c := recordConfig(t, family)
	for i := range c.Peers {
		c.Peers[i].MaxBytes = 128 << 20
	}
	c.RoTargets = []RoTargetConfig{
		{TargetID: "ocs-a", Peers: []PeerID{"dra", "dra2"}, Destination: serverOrigin, ReplayDomain: "charging-domain", EvidenceID: "deployment-primary", Retention: time.Hour, SessionEvidenceID: "primary-session-continuity"},
		{TargetID: "ocs-b", Peers: []PeerID{"dra", "dra2"}, Destination: crossOrigin, ReplayDomain: "charging-domain", EvidenceID: "deployment-secondary", Retention: time.Hour, SessionEvidenceID: "secondary-session-continuity"},
	}
	return c
}
func crossOptions() RoRecoveryOptions {
	o := recordOptions()
	o.PolicyID = RoSMSFWRecoveryPolicy
	o.SessionEvidenceID = "application-session-record"
	o.Targets = []RoRecoveryTarget{{ID: "ocs-a", Peer: "dra", Retention: time.Hour}, {ID: "ocs-b", Peer: "dra", Retention: time.Hour}}
	return o
}
func crossPlan() RoRecoveryPlan {
	p := trPlan(2, 500*time.Millisecond)
	p.Peers = []PeerID{"dra"}
	p.Alternates = []RoRecoveryPath{{TargetID: "ocs-b", Peers: []PeerID{"dra"}}}
	return p
}
func crossReply(t *testing.T, co *codec.Codec, remote transport.Conn, req pdu.Message, origin pdu.Origin, code uint32) {
	t.Helper()
	var body pdu.Body
	if code >= 3000 && code < 4000 {
		body = &pdu.ErrorAnswer{ResultCode: code}
	} else {
		body = roRecoveryAnswer(req.Body, code)
	}
	ans, e := co.CompleteAnswer(req, body, origin)
	if e != nil {
		t.Fatal(e)
	}
	if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
		t.Fatal(e)
	}
}
func crossIdentity(t *testing.T, n *Node, co *codec.Codec, r RoRecoveryRequest, m pdu.Message) {
	t.Helper()
	b, e := co.Append(nil, m)
	if e != nil {
		t.Fatal(e)
	}
	expected, e := roTargetFrame(r.Frame, crossOrigin, n.cfg.Limits, r.Limits.MaxRecordBytes)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(b[16:], expected[16:]) {
		t.Fatal("non-destination original bytes changed")
	}
}
func TestRoTargetsSameDRAProfilesAndLateAnswers(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, family := range []string{"standard", "sms", "money", "combined"} {
			kinds := []string{"iec", "initial", "term"}
			if family == "money" {
				kinds = []string{"general", "balance", "refund"}
			}
			if family == "combined" {
				kinds = []string{"iec", "initial", "term", "general", "balance", "refund"}
			}
			for _, kind := range kinds {
				for _, oldWins := range []bool{false, true} {
					n, co := newInboundNode(t, crossConfig(t, family))
					_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
					port := portTest(t, n, 1)
					r := prepareRecord(t, n, roRecoveryBody(t, family, kind), crossOptions())
					id := submitSegment(t, n, r, crossPlan(), segmentProgress(2), port)
					first := routeRead(t, co, remote)
					i := crossExpire(t, n, id)
					if i.TargetID != "ocs-a" {
						t.Fatal(i)
					}
					next, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt)
					if e != nil {
						t.Fatal(e)
					}
					second := routeRead(t, co, remote)
					crossIdentity(t, n, co, r, second)
					if second.Header.HopByHop == first.Header.HopByHop || second.Header.Flags&pdu.RetransmittedFlag == 0 || trInfo(t, n, id).TargetID != "ocs-b" {
						t.Fatal("target or retransmission header")
					}
					again, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt)
					if e != nil || again != next {
						t.Fatal(e)
					}
					winner := uint32(2)
					if oldWins {
						crossReply(t, co, remote, first, serverOrigin, 2001)
						winner = 1
					} else {
						crossReply(t, co, remote, second, crossOrigin, 2001)
					}
					got := routeCompletion(t, n, port, id)
					if got.SelectedAttempt.Number != winner || got.Attempts[0].ConnectionID != got.Attempts[1].ConnectionID || got.Attempts[1].TargetID != "ocs-b" || got.RoRecovery.CanContinue {
						t.Fatal(got)
					}
					segmentIdle(t, n)
				}
			}
		}
	})
}
func TestRoTargetsAuthorityAndSessionGates(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"domain", "session", "target-session", "policy", "path", "repeat", "foreign"} {
			c := crossConfig(t, "standard")
			o := crossOptions()
			body := roRecoveryBody(t, "standard", "term")
			p := crossPlan()
			switch kind {
			case "domain":
				c.RoTargets[1].ReplayDomain = "unrelated"
			case "session":
				o.SessionEvidenceID = ""
			case "target-session":
				c.RoTargets[1].SessionEvidenceID = ""
			case "policy":
				o.PolicyID = RoTransientRecoveryPolicy
			case "path":
				c.RoTargets[1].Peers = []PeerID{"dra2"}
				o.Targets[1].Peer = "dra2"
			case "repeat":
				p.Alternates[0].TargetID = "ocs-a"
			case "foreign":
				p.Alternates[0].TargetID = "foreign"
			}
			n, co := newInboundNode(t, c)
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, body, o)
			if _, e := n.SubmitRoRecovery(ctxTest(t), r, p, segmentProgress(2), port); e == nil {
				t.Fatal("unauthorized plan", kind)
			}
			if n.Resources().Pending != 0 || n.seq != 0 {
				t.Fatal("failed authority consumed admission")
			}
			p.Alternates = nil
			id := submitSegment(t, n, r, p, segmentProgress(2), port)
			req := routeRead(t, co, remote)
			crossReply(t, co, remote, req, serverOrigin, 2001)
			routeCompletion(t, n, port, id)
		}
	})
}
func TestRoTargetsConstructorBoundsAndOwnedConfig(t *testing.T) {
	for _, mutate := range []func(*NodeConfig){func(c *NodeConfig) { c.RoTargets[1].TargetID = "ocs-a" }, func(c *NodeConfig) { c.RoTargets[1].EvidenceID = "" }, func(c *NodeConfig) { c.RoTargets[1].Peers = []PeerID{"missing"} }, func(c *NodeConfig) { c.RoTargets[1].Peers = []PeerID{"dra", "dra"} }, func(c *NodeConfig) { c.RoTargets[1].SessionEvidenceID = strings.Repeat("x", 129) }, func(c *NodeConfig) { c.RoTargets[1].Retention = 0 }} {
		c := crossConfig(t, "standard")
		mutate(&c)
		if n, e := NewNode(c); e == nil {
			n.Close()
			t.Fatal("invalid target configuration accepted")
		}
	}
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := crossConfig(t, "standard")
		n, co := newInboundNode(t, c)
		c.RoTargets[1].Peers[0] = "missing"
		c.RoTargets[1].Destination = serverOrigin
		c.RoTargets[1].EvidenceID = "changed"
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), crossOptions())
		p := crossPlan()
		p.TargetID = "ocs-b"
		p.Alternates = nil
		id := submitSegment(t, n, r, p, segmentProgress(2), port)
		req := routeRead(t, co, remote)
		crossIdentity(t, n, co, r, req)
		if req.Header.Flags&pdu.RetransmittedFlag != 0 {
			t.Fatal("first alternate selection set T")
		}
		r.Targets[1].AllowedPeers[0] = "missing"
		r.Targets[1].Policy.Sources.Terminals[0] = serverOrigin
		crossReply(t, co, remote, req, crossOrigin, 2001)
		got := routeCompletion(t, n, port, id)
		if got.Err != nil {
			t.Fatal(got)
		}
	})
}
func TestRoTargetsSourceIsolation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, crossConfig(t, "standard"))
		_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), crossOptions())
		id := submitSegment(t, n, r, crossPlan(), segmentProgress(2), port)
		routeRead(t, co, remote)
		i := crossExpire(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		crossReply(t, co, remote, req, serverOrigin, 2001)
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.RoRecovery.CanContinue {
			t.Fatal(got)
		}
	})
}
func TestRoTargetsErrorAndNoAnswerSegments(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, errorAnswer := range []bool{false, true} {
			n, co := newInboundNode(t, crossConfig(t, "standard"))
			_, remote := roRecoveryNetwork(t, n, co, "dra", tls)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), crossOptions())
			p := crossPlan()
			p.MaxAttempts = 1
			p.Alternates = nil
			p.ErrorContinuation = RoErrorTransient
			id := submitSegment(t, n, r, p, segmentProgress(1), port)
			old := routeRead(t, co, remote)
			if errorAnswer {
				crossReply(t, co, remote, old, serverOrigin, 3004)
			} else {
				n.Cancel(id)
			}
			got := routeCompletion(t, n, port, id)
			p.Trigger = got.RoRecovery.NextTrigger
			p.TargetID = "ocs-b"
			id = submitSegment(t, n, r, p, nextSegmentProgress(got, 1), port)
			next := routeRead(t, co, remote)
			crossIdentity(t, n, co, r, next)
			if (next.Header.Flags&pdu.RetransmittedFlag != 0) == errorAnswer {
				t.Fatal("wrong cross-segment T")
			}
			crossReply(t, co, remote, old, serverOrigin, 2001)
			crossReply(t, co, remote, next, crossOrigin, 2001)
			got = routeCompletion(t, n, port, id)
			if got.Err != nil || !got.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
				t.Fatal(got)
			}
		}
	})
}
func TestRoTargetsVisitedAndAttemptBudgets(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, crossConfig(t, "standard"))
		peer, remote := roRecoveryNetwork(t, n, co, "dra", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), crossOptions())
		p := crossPlan()
		p.MaxAttempts = 3
		p.Peers = []PeerID{"dra", "dra2"}
		id := submitSegment(t, n, r, p, segmentProgress(3), port)
		routeRead(t, co, remote)
		i := crossExpire(t, n, id)
		n.mu.Lock()
		limit := peer.account.LimitSlots
		peer.account.LimitSlots = 1
		n.mu.Unlock()
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		if trInfo(t, n, id).AttemptsUsed != 1 {
			t.Fatal("rejected target consumed attempt")
		}
		n.mu.Lock()
		peer.account.LimitSlots = limit
		n.mu.Unlock()
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, remote)
		roRecoveryNetwork(t, n, co, "dra2", tls)
		i = crossExpire(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrNoEligibleAlternate) {
			t.Fatal("revisited frozen path", e)
		}
		// A now-ready earlier target cannot be revisited after switching terminals.
		crossReply(t, co, remote, second, crossOrigin, 2001)
		got := routeCompletion(t, n, port, id)
		if len(got.Attempts) != 2 {
			t.Fatal(got)
		}
		segmentIdle(t, n)
	})
}
func TestRoTargetsOldReaderAndSuccessorBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := crossConfig(t, "standard")
		c.RoTargets[1].Peers = []PeerID{"dra2"}
		o := crossOptions()
		o.Targets[1].Peer = "dra2"
		n, co := newInboundNode(t, c)
		a, b := networkAdapters(t, tls)
		barrier := &roRecoveryBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		roRecoveryConnect(t, n, co, "dra", barrier, b)
		_, other := roRecoveryNetwork(t, n, co, "dra2", tls)
		port := portTest(t, n, 1)
		r := prepareRecord(t, n, roBody(), o)
		p := crossPlan()
		p.Alternates[0].Peers = []PeerID{"dra2"}
		id := submitSegment(t, n, r, p, segmentProgress(2), port)
		first := routeRead(t, co, b)
		crossReply(t, co, b, first, serverOrigin, 4012)
		<-barrier.entered
		i := crossExpire(t, n, id)
		if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
			t.Fatal(e)
		}
		second := routeRead(t, co, other)
		crossReply(t, co, other, second, crossOrigin, 2001)
		waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, e := port.Receive(ctx)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) || n.Resources().OutboundBytes == 0 {
			t.Fatal("released old A before reader", e)
		}
		once.Do(func() { close(barrier.release) })
		got := routeCompletion(t, n, port, id)
		if got.SelectedAttempt.Number != 2 {
			t.Fatal(got)
		}
		segmentIdle(t, n)
	})
}
func TestRoTargetsWriterRetirementAndCancel(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, cancelNow := range []bool{false, true} {
			n, co := newInboundNode(t, crossConfig(t, "standard"))
			a, b := networkAdapters(t, tls)
			barrier := &roRecoveryBarrier{Conn: a, write: true, entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			defer once.Do(func() { close(barrier.release) })
			roRecoveryConnect(t, n, co, "dra", barrier, b)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), crossOptions())
			id := submitSegment(t, n, r, crossPlan(), segmentProgress(2), port)
			first := routeRead(t, co, b)
			<-barrier.entered
			i := trExpire(t, n, id)
			if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); !errors.Is(e, ErrAttemptNotReplayable) {
				t.Fatal(e)
			}
			if cancelNow {
				n.Cancel(id)
			}
			once.Do(func() { close(barrier.release) })
			if cancelNow {
				got := routeCompletion(t, n, port, id)
				if len(got.Attempts) != 1 {
					t.Fatal(got)
				}
			} else {
				i = roAwait(t, n, id)
				if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
					t.Fatal(e)
				}
				routeRead(t, co, b)
				crossReply(t, co, b, first, serverOrigin, 2001)
				got := routeCompletion(t, n, port, id)
				if got.SelectedAttempt.Number != 1 {
					t.Fatal(got)
				}
			}
			segmentIdle(t, n)
		}
	})
}

type crossSecondWriter struct {
	transport.Conn
	entered, release chan struct{}
	count            int
}

func (b *crossSecondWriter) WriteFrame(ctx context.Context, p []byte) transport.WriteResult {
	r := b.Conn.WriteFrame(ctx, p)
	if len(p) > 12 && p[8] == 0 && p[9] == 0 && p[10] == 0 && p[11] == 4 {
		b.count++
		if b.count == 2 {
			close(b.entered)
			<-b.release
		}
	}
	return r
}
func TestRoTargetsSuccessorWriterAndCloseBarrier(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, end := range []string{"old-answer", "cancel", "node", "peer"} {
			n, co := newInboundNode(t, crossConfig(t, "standard"))
			a, b := networkAdapters(t, tls)
			barrier := &crossSecondWriter{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
			var once sync.Once
			defer once.Do(func() { close(barrier.release) })
			peer := roRecoveryConnect(t, n, co, "dra", barrier, b)
			port := portTest(t, n, 1)
			r := prepareRecord(t, n, roBody(), crossOptions())
			id := submitSegment(t, n, r, crossPlan(), segmentProgress(2), port)
			first := routeRead(t, co, b)
			i := crossExpire(t, n, id)
			if _, e := n.Retransmit(ctxTest(t), id, i.CurrentAttempt); e != nil {
				t.Fatal(e)
			}
			routeRead(t, co, b)
			<-barrier.entered
			switch end {
			case "old-answer":
				crossReply(t, co, b, first, serverOrigin, 2001)
			case "cancel":
				n.Cancel(id)
			case "node":
				n.Close()
			case "peer":
				peer.Close()
			}
			waitFor(t, func() bool { i, _ := n.RequestInfo(id); return i.Phase == Completing })
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			_, e := port.Receive(ctx)
			cancel()
			if !errors.Is(e, context.DeadlineExceeded) || n.Resources().OutboundBytes == 0 {
				t.Fatal("early cross-target Q release", e)
			}
			once.Do(func() { close(barrier.release) })
			got := routeCompletion(t, n, port, id)
			if !got.Attempts[1].Delivery.MayHaveReachedPeer || got.Attempts[1].TargetID != "ocs-b" {
				t.Fatal(got)
			}
			if end == "old-answer" && got.SelectedAttempt.Number != 1 {
				t.Fatal(got)
			}
			segmentIdle(t, n)
		}
	})
}
func TestRoTargetsFrozenAuthorityAndLargestFrame(t *testing.T) {
	c := crossConfig(t, "standard")
	n, _ := newInboundNode(t, c)
	r := prepareRecord(t, n, roBody(), crossOptions())
	port := portTest(t, n, 1)
	for _, mutate := range []func(*RoRecoveryRequest){func(r *RoRecoveryRequest) { r.SessionEvidenceID = "forged" }, func(r *RoRecoveryRequest) { r.Targets[1].EvidenceID = "forged" }, func(r *RoRecoveryRequest) { r.Targets[1].Policy.Sources.Terminals[0] = serverOrigin }, func(r *RoRecoveryRequest) { r.Version = 1 }, func(r *RoRecoveryRequest) { r.Targets[1].AllowedPeers = []PeerID{"dra2"} }} {
		bad := copyRecord(t, r)
		mutate(&bad)
		bad.Digest = roRecordDigest(bad)
		if _, e := n.SubmitRoRecovery(ctxTest(t), bad, crossPlan(), segmentProgress(2), port); e == nil {
			t.Fatal("mutated authority admitted")
		}
	}
	// A target too long for its original configured frame limit fails before E2E allocation.
	c = crossConfig(t, "standard")
	for i := range c.Peers {
		c.Peers[i].Commands[0].MaxRequestBytes = 512
	}
	c.RoTargets[1].Destination = pdu.Origin{Host: strings.Repeat(strings.Repeat("a", 60)+".", 4) + "t", Realm: strings.Repeat(strings.Repeat("b", 60)+".", 4) + "t"}
	n, _ = newInboundNode(t, c)
	before := n.IdentifierInfo().ApplicationAllocated
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), crossOptions()); e == nil {
		t.Fatal("oversized alternate admitted")
	}
	if n.IdentifierInfo().ApplicationAllocated != before || n.RoRecoveryInfo().Held != 0 {
		t.Fatal("failed preparation reserved identity")
	}
}

func crossExpire(t *testing.T, n *Node, id RequestID) RequestInfo {
	t.Helper()
	trExpire(t, n, id)
	return roAwait(t, n, id)
}
