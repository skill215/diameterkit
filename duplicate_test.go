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
	"sync/atomic"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
)

func duplicateConfig(t *testing.T) NodeConfig {
	c := rejectConfig(t)
	c.Resources.NodeBytes = 1 << 30
	c.Resources.DuplicateClasses = []ClassConfig{{ID: "sms-cache", Bytes: 128 << 20, Slots: 8}}
	p := &c.Peers[0].InboundCommands[0]
	p.NoCache = false
	p.Duplicate = DuplicatePolicy{Class: "sms-cache", MaxProcessing: p.ExecutionTimeout, ReplayRetention: time.Second, MaxArrivals: 4, MaxCanonicalBytes: 2048, MaxResultBytes: 700, MaxEnvelopeBytes: 300}
	return c
}
func otherArrival(t *testing.T, co *codec.Codec, req pdu.Message, hbh uint32) pdu.Message {
	t.Helper()
	m, e := co.Snapshot(req)
	if e != nil {
		t.Fatal(e)
	}
	m.Header.HopByHop = hbh
	m.Header.Flags ^= pdu.RetransmittedFlag
	s := m.Body.(*s6c.SRR)
	s.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{byte(hbh), 4, 5}}}
	return m
}
func noDuplicateDelivery(t *testing.T, p *Peer) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if in, e := p.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) || in != nil {
		t.Fatal("duplicate delivered", in, e)
	}
}
func duplicateAnswer(t *testing.T, n *Node, co *codec.Codec, remote transport.Conn, req pdu.Message) pdu.Message {
	t.Helper()
	a, e := readWire(ctxTest(t), remote, co)
	if e != nil {
		t.Fatal(e)
	}
	if e = co.ValidateAnswer(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
		t.Fatal(e)
	}
	return a
}
func duplicateConnect(t *testing.T, n *Node, co *codec.Codec, tls bool, id PeerID) (*Peer, transport.Conn) {
	t.Helper()
	a, b := networkAdapters(t, tls)
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		m, e := readWire(ctx, b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(m, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "duplicate-dra", AuthApplicationID: []uint32{appID}}, n.peers[id].origin)
			if e == nil {
				e = writeWire(ctx, b, co, ans)
			}
		}
		ch <- e
	}()
	p, e := n.Connect(ctx, id, a)
	if e != nil {
		t.Fatal(e)
	}
	if e = <-ch; e != nil {
		t.Fatal(e)
	}
	return p, b
}
func twoDuplicatePeers(c NodeConfig) NodeConfig {
	p := c.Peers[0]
	p.ID = "dra2"
	p.Origin = pdu.Origin{Host: "dra2.example", Realm: "example"}
	p.InboundCommands = append([]InboundPolicy(nil), p.InboundCommands...)
	c.Peers = append(c.Peers, p)
	return c
}
func TestDuplicateConfiguration(t *testing.T) {
	for _, mutate := range []func(*NodeConfig){
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].Duplicate.MaxArrivals = 0 },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].Duplicate.ReplayRetention = 0 },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].Duplicate.MaxProcessing = time.Second },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].Duplicate.MaxEnvelopeBytes = 400 },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].Duplicate.MaxCanonicalBytes = 4096 },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure },
		func(c *NodeConfig) { c.Resources.DuplicateClasses = nil },
		func(c *NodeConfig) { c.Peers[0].InboundCommands[0].NoCache = true },
	} {
		c := duplicateConfig(t)
		mutate(&c)
		if _, e := NewNode(c); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
	}
}
func TestDuplicateProcessingCompletedAndEnvelope(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 110)
		before := n.Resources()
		dup := otherArrival(t, co, req, 111)
		dup.Body.(*s6c.SRR).OriginHost = strings.ToUpper(serverOrigin.Host)
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if n.Resources().DuplicateRecords != 1 || n.Resources().DuplicateBytes != before.DuplicateBytes {
			t.Fatal(n.Resources())
		}
		in.Message().Body.(*s6c.SRR).SessionID = "mutated-view"
		result, e := in.Respond(ctxTest(t), inboundFailure())
		if e != nil || !result.Committed {
			t.Fatal(result, e)
		}
		// The original and duplicate may be queued in either order; associate by HBH.
		for range 2 {
			a, e := readWire(ctxTest(t), remote, co)
			if e != nil {
				t.Fatal(e)
			}
			r := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				r = dup
			}
			if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
		late := otherArrival(t, co, req, 112)
		late.Header.Flags &^= pdu.RetransmittedFlag
		if e := writeWire(ctxTest(t), remote, co, late); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, late)
		noDuplicateDelivery(t, p)
	})
}
func TestDuplicateConflictDoesNotReplaceOrRefresh(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 120)
		n.mu.Lock()
		r := in.duplicate
		until := r.keepUntil
		n.mu.Unlock()
		bad := otherArrival(t, co, req, 121)
		bad.Body.(*s6c.SRR).SessionID += "different"
		if e := writeWire(ctxTest(t), remote, co, bad); e != nil {
			t.Fatal(e)
		}
		readReject(t, remote, bad.Header, 5012)
		noDuplicateDelivery(t, p)
		n.mu.Lock()
		same := r.keepUntil.Equal(until) && !r.committed
		n.mu.Unlock()
		if !same {
			t.Fatal("conflict replaced record")
		}
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, req)
		good := otherArrival(t, co, req, 122)
		if e := writeWire(ctxTest(t), remote, co, good); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, remote, good)
	})
}
func TestDuplicateAcrossPeersAndDisconnectedPrimary(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, twoDuplicatePeers(duplicateConfig(t)))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		in, req := receiveInbound(t, n, p, a, co, 130)
		_ = a.Close()
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		dup := otherArrival(t, co, req, 131)
		if e := writeWire(ctxTest(t), b, co, dup); e != nil {
			t.Fatal(e)
		}
		noDuplicateDelivery(t, p2)
		result, e := in.Respond(ctxTest(t), inboundFailure())
		if !result.Committed || result.PrimaryWrite.State != ResponseNotWritten || e == nil {
			t.Fatal(result, e)
		}
		duplicateAnswer(t, n, co, b, dup)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
func TestDuplicatePolicyConflictAndIndependentSource(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, conflict := range []bool{false, true} {
			c := twoDuplicatePeers(duplicateConfig(t))
			c.Peers[1].InboundCommands[0].Sources = append(c.Peers[1].InboundCommands[0].Sources, pdu.Origin{Host: "extra.example", Realm: "example"})
			if conflict {
				c.Peers[1].InboundCommands[0].Duplicate.ReplayRetention += time.Second
			}
			n, co := newInboundNode(t, c)
			p, a := duplicateConnect(t, n, co, tls, "dra")
			p2, b := duplicateConnect(t, n, co, tls, "dra2")
			in, req := receiveInbound(t, n, p, a, co, 140)
			dup := otherArrival(t, co, req, 141)
			if e := writeWire(ctxTest(t), b, co, dup); e != nil {
				t.Fatal(e)
			}
			noDuplicateDelivery(t, p2)
			if conflict {
				readReject(t, b, dup.Header, 5012)
			}
			if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
				t.Fatal(e)
			}
			duplicateAnswer(t, n, co, a, req)
			if !conflict {
				duplicateAnswer(t, n, co, b, dup)
			}
		}
	})
}
func TestDuplicateAdmissionAndRetention(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := duplicateConfig(t)
		c.Resources.DuplicateClasses[0].Slots = 1
		n, co := newInboundNode(t, c)
		p, a := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, a, co, 150)
		n.mu.Lock()
		until := in.duplicate.keepUntil
		n.mu.Unlock()
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
		// A new identity cannot evict the promised result, even with all I/W free.
		other := inboundRequest(t, n, co, 151)
		if e := writeWire(ctxTest(t), a, co, other); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return !p.Info().Ready })
		if !errors.Is(p.Info().Cause, ErrBackpressure) || n.Resources().DuplicateRecords != 1 {
			t.Fatal(p.Info(), n.Resources())
		}
		n.tick(until)
		if n.Resources().DuplicateRecords != 0 || n.Resources().DuplicateBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestDuplicateArrivalLimitDoesNotCancelLogicalWork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := twoDuplicatePeers(duplicateConfig(t))
		for i := range c.Peers {
			c.Peers[i].InboundCommands[0].Duplicate.MaxArrivals = 1
		}
		n, co := newInboundNode(t, c)
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		in, req := receiveInbound(t, n, p, a, co, 160)
		dup := otherArrival(t, co, req, 161)
		if e := writeWire(ctxTest(t), b, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return !p2.Info().Ready })
		if !errors.Is(p2.Info().Cause, ErrBackpressure) {
			t.Fatal(p2.Info())
		}
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
	})
}
func TestDuplicateTimeoutCachedAndNeverRefreshed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := duplicateConfig(t)
		ip := &c.Peers[0].InboundCommands[0]
		ip.ExecutionTimeout = 80 * time.Millisecond
		ip.Duplicate.MaxProcessing = ip.ExecutionTimeout
		n, co := newInboundNode(t, c)
		p, a := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, a, co, 170)
		dup := otherArrival(t, co, req, 171)
		if e := writeWire(ctxTest(t), a, co, dup); e != nil {
			t.Fatal(e)
		}
		for range 2 {
			ans, e := readWire(ctxTest(t), a, co)
			if e != nil {
				t.Fatal(e)
			}
			if v, ok := ans.Body.(*pdu.ErrorAnswer); !ok || v.ResultCode != 5012 {
				t.Fatal(ans)
			}
		}
		if r, e := in.Respond(ctxTest(t), inboundFailure()); r.Committed || !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(r, e)
		}
		late := otherArrival(t, co, req, 172)
		if e := writeWire(ctxTest(t), a, co, late); e != nil {
			t.Fatal(e)
		}
		readReject(t, a, late.Header, 5012)
		noDuplicateDelivery(t, p)
	})
}
func TestDuplicateResultLimitAllowsRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := duplicateConfig(t)
		c.Peers[0].InboundCommands[0].Duplicate.MaxResultBytes = 180
		n, co := newInboundNode(t, c)
		p, a := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, a, co, 180)
		big := inboundFailure()
		big.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}, Data: make([]byte, 200)}}
		if r, e := in.Respond(ctxTest(t), big); r.Committed || !errors.Is(e, wire.ErrLimit) {
			t.Fatal(r, e)
		}
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
	})
}
func TestDuplicateCacheBorrowSurvivesExpiryAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		a, b := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: a, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		p := connectInboundTransport(t, n, co, barrier, b)
		in, req := receiveInbound(t, n, p, b, co, 190)
		n.mu.Lock()
		until := in.duplicate.keepUntil
		n.mu.Unlock()
		ch := make(chan ResponseResult, 1)
		go func() { r, _ := in.Respond(ctxTest(t), inboundFailure()); ch <- r }()
		duplicateAnswer(t, n, co, b, req)
		<-barrier.entered
		n.tick(until)
		_ = n.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if e := n.Wait(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal("cache released during write borrow", n.Resources())
		}
		once.Do(func() { close(barrier.release) })
		if r := <-ch; !r.Committed || r.PrimaryWrite.State != ResponseWritten {
			t.Fatal(r)
		}
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().DuplicateRecords != 0 {
			t.Fatal(n.Resources())
		}
	})
}

// Canonical tests exercise original raw frames, not an encode/decode round-trip
// as a substitute for unknown-field ordering or original presence evidence.
func canonicalParts(t *testing.T, b []byte) [][]byte {
	t.Helper()
	d := wire.NewDecoder(b[20:], wire.DefaultLimits())
	var out [][]byte
	for {
		v, ok, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !ok {
			break
		}
		start := 20 + v.Offset
		size := (int(wire.Uint24(b[start+5:start+8])) + 3) &^ 3
		out = append(out, append([]byte(nil), b[start:start+size]...))
	}
	return out
}
func joinCanonicalParts(b []byte, parts [][]byte) []byte {
	out := append([]byte(nil), b[:20]...)
	for _, p := range parts {
		out = append(out, p...)
	}
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func TestCanonicalOriginalEvidenceAndLimits(t *testing.T) {
	n, co := newInboundNode(t, duplicateConfig(t))
	req := inboundRequest(t, n, co, 200)
	raw := reqBytes(t, co, req)
	want, e := co.Canonical(raw, 2048)
	if e != nil {
		t.Fatal(e)
	}
	altered := otherArrival(t, co, req, 201)
	altered.Body.(*s6c.SRR).OriginHost = strings.ToUpper(serverOrigin.Host)
	altered.Body.(*s6c.SRR).RouteRecord = []string{"other.example"}
	b := reqBytes(t, co, altered)
	parts := canonicalParts(t, b)
	for i := 1; i < len(parts)/2; i++ {
		j := len(parts) - i
		parts[i], parts[j] = parts[j], parts[i]
	}
	b = joinCanonicalParts(b, parts)
	for off := 20; off < len(b); {
		b[off+4] &^= avp.MandatoryFlag
		b[off+4] |= 1
		size := int(wire.Uint24(b[off+5 : off+8]))
		for i := size; i < (size+3)&^3; i++ {
			b[off+i] = 0x7f
		}
		off += (size + 3) &^ 3
	}
	b[4] |= 0x0f
	got, e := co.Canonical(b, 2048)
	if e != nil || !bytes.Equal(want, got) {
		t.Fatal("approved normalization", e)
	}
	if _, e = co.Canonical(raw, 12); !errors.Is(e, wire.ErrLimit) {
		t.Fatal(e)
	}
	clear(raw)
	if bytes.Equal(want, make([]byte, len(want))) {
		t.Fatal("canonical borrows caller bytes")
	}
}
func TestCanonicalUnknownBarriersPresenceAndRepeatedOrder(t *testing.T) {
	n, co := newInboundNode(t, duplicateConfig(t))
	req := inboundRequest(t, n, co, 210)
	raw := reqBytes(t, co, req)
	want, e := co.Canonical(raw, 2048)
	if e != nil {
		t.Fatal(e)
	}
	changed := otherArrival(t, co, req, 211)
	changed.Body.(*s6c.SRR).SMRPMTI = avp.Some(uint32(0))
	b := reqBytes(t, co, changed)
	got, e := co.Canonical(b, 2048)
	if e != nil || bytes.Equal(got, want) {
		t.Fatal("presence lost", e)
	}
	req.Body.(*s6c.SRR).Extra = []avp.Raw{{Key: avp.Key{Code: 9999}, Data: []byte{1, 2}}}
	b = reqBytes(t, co, req)
	parts := canonicalParts(t, b)
	barrier := parts[len(parts)-1]
	parts = append(parts[:2], append([][]byte{barrier}, parts[2:len(parts)-1]...)...)
	b = joinCanonicalParts(b, parts)
	a, e := co.Canonical(b, 2048)
	if e != nil {
		t.Fatal(e)
	}
	parts[1], parts[3] = parts[3], parts[1]
	different, e := co.Canonical(joinCanonicalParts(b, parts), 2048)
	if e != nil || bytes.Equal(a, different) {
		t.Fatal("sorted across unknown barrier", e)
	}
	// Group order and same-key top-level repeat order are significant.
	s := req.Body.(*s6c.SRR)
	s.Extra = nil
	s.SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
	b = reqBytes(t, co, req)
	a, e = co.Canonical(b, 2048)
	if e != nil {
		t.Fatal(e)
	}
	s.SupportedFeatures[0], s.SupportedFeatures[1] = s.SupportedFeatures[1], s.SupportedFeatures[0]
	different, e = co.Canonical(reqBytes(t, co, req), 2048)
	if e != nil || bytes.Equal(a, different) {
		t.Fatal("repeat order lost", e)
	}
}

func TestCanonicalIgnoreScopeAndNestedSequence(t *testing.T) {
	n, co := newInboundNode(t, duplicateConfig(t))
	req := inboundRequest(t, n, co, 220)
	raw := reqBytes(t, co, req)
	want, e := co.Canonical(raw, 2048)
	if e != nil {
		t.Fatal(e)
	}
	// Only the fixed SMS top-level 260 is ignored, even if its opaque body is not a group.
	b := append([]byte(nil), raw...)
	extra := make([]byte, 12)
	w := wire.NewEncoder(extra, wire.DefaultLimits())
	w.Bytes(avp.Key{Code: 260}, avp.MandatoryFlag, []byte{7})
	b = append(b, extra...)
	wire.Put24(b[1:4], uint32(len(b)))
	got, e := co.Canonical(b, 2048)
	if e != nil || !bytes.Equal(want, got) {
		t.Fatal("SMS VSAI", e)
	}
	req.Body.(*s6c.SRR).Extra = []avp.Raw{{Key: avp.Key{VendorID: 999, Code: 284}, Flags: avp.VendorFlag, Data: []byte{1, 2, 3}}}
	got, e = co.Canonical(reqBytes(t, co, req), 2048)
	if e != nil || bytes.Equal(want, got) {
		t.Fatal("stripped vendor extension", e)
	}
	req.Body.(*s6c.SRR).Extra = nil
	req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 2}}
	raw = reqBytes(t, co, req)
	want, e = co.Canonical(raw, 2048)
	if e != nil {
		t.Fatal(e)
	}
	parts := canonicalParts(t, raw)
	for _, p := range parts {
		if wire.Uint24(p[1:4]) == 628 {
			a := append([]byte(nil), p[12:24]...)
			copy(p[12:28], p[24:40])
			copy(p[28:40], a)
			break
		}
	}
	got, e = co.Canonical(joinCanonicalParts(raw, parts), 2048)
	if e != nil || bytes.Equal(want, got) {
		t.Fatal("group sequence collapsed", e)
	}
	// Base control comparisons preserve their VSAI; they never enter runtime cache.
	cer := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257}, Body: &base.CER{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "canonical", AuthApplicationID: []uint32{appID}}}
	first := reqBytes(t, co, cer)
	x, e := co.Canonical(first, 2048)
	if e != nil {
		t.Fatal(e)
	}
	cer.Body.(*base.CER).VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(appID)}}
	y, e := co.Canonical(reqBytes(t, co, cer), 2048)
	if e != nil || bytes.Equal(x, y) {
		t.Fatal("Base VSAI ignored", e)
	}
}
func TestDuplicateConcurrentFirstArrivalAndRespondRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, twoDuplicatePeers(duplicateConfig(t)))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		req := inboundRequest(t, n, co, 230)
		dup := otherArrival(t, co, req, 231)
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i, c := range []transport.Conn{a, b} {
			wg.Add(1)
			go func(i int, c transport.Conn) {
				defer wg.Done()
				m := req
				if i == 1 {
					m = dup
				}
				errs <- writeWire(ctxTest(t), c, co, m)
			}(i, c)
		}
		wg.Wait()
		for range 2 {
			if e := <-errs; e != nil {
				t.Fatal(e)
			}
		}
		waitFor(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			for _, r := range n.duplicates {
				if len(r.arrivals) == 2 {
					return true
				}
			}
			return false
		})
		n.mu.Lock()
		owner := p
		if len(p2.inboundQueue) == 1 {
			owner = p2
		}
		queued := len(p.inboundQueue) + len(p2.inboundQueue)
		n.mu.Unlock()
		if queued != 1 {
			t.Fatal(queued)
		}
		in, e := owner.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		results := make(chan ResponseResult, 2)
		for range 2 {
			go func() { r, _ := in.Respond(ctxTest(t), inboundFailure()); results <- r }()
		}
		wins := 0
		for range 2 {
			if (<-results).Committed {
				wins++
			}
		}
		if wins != 1 {
			t.Fatal(wins)
		}
		duplicateAnswer(t, n, co, a, req)
		duplicateAnswer(t, n, co, b, dup)
	})
}
func TestDuplicateRetiredRecordAllowsNewWork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, a := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, a, co, 240)
		n.mu.Lock()
		until := in.duplicate.keepUntil
		n.mu.Unlock()
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		n.tick(until)
		if e := writeWire(ctxTest(t), a, co, req); e != nil {
			t.Fatal(e)
		}
		again, e := p.Receive(ctxTest(t))
		if e != nil || again.Info().RequestID == in.Info().RequestID {
			t.Fatal(again, e)
		}
	})
}
func TestDuplicateEnvelopeBoundAndPerArrivalSource(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, oversize := range []bool{false, true} {
			n, co := newInboundNode(t, twoDuplicatePeers(duplicateConfig(t)))
			p, a := duplicateConnect(t, n, co, tls, "dra")
			p2, b := duplicateConnect(t, n, co, tls, "dra2")
			in, req := receiveInbound(t, n, p, a, co, 250)
			dup := otherArrival(t, co, req, 251)
			if oversize {
				dup.Body.(*s6c.SRR).ProxyInfo[0].ProxyState = make([]byte, 350)
			} else {
				dup.Body.(*s6c.SRR).OriginRealm = "wrong.example"
			}
			if e := writeWire(ctxTest(t), b, co, dup); e != nil {
				t.Fatal(e)
			}
			if oversize {
				waitFor(t, func() bool { return !p2.Info().Ready })
			} else {
				readReject(t, b, dup.Header, 3002)
			}
			if n.Resources().DuplicateRecords != 1 {
				t.Fatal(n.Resources())
			}
			if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
				t.Fatal(e)
			}
			duplicateAnswer(t, n, co, a, req)
		}
	})
}
func TestDuplicateSuppressedUndeliveredPeerClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, twoDuplicatePeers(duplicateConfig(t)))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		req := inboundRequest(t, n, co, 260)
		if e := writeWire(ctxTest(t), a, co, req); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.inboundQueue) == 1 })
		_ = p.Close()
		dup := otherArrival(t, co, req, 261)
		if e := writeWire(ctxTest(t), b, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return !p2.Info().Ready })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}
func TestDuplicateReplayBorrowAndClosing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, twoDuplicatePeers(duplicateConfig(t)))
		p, a := duplicateConnect(t, n, co, tls, "dra")
		in, req := receiveInbound(t, n, p, a, co, 270)
		n.mu.Lock()
		until := in.duplicate.keepUntil
		n.mu.Unlock()
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
		// Put the actual replay writer behind a bounded test return barrier.
		c, d := networkAdapters(t, tls)
		barrier := &inboundWriteBarrier{Conn: c, entered: make(chan struct{}), release: make(chan struct{})}
		var once sync.Once
		defer once.Do(func() { close(barrier.release) })
		ch := make(chan error, 1)
		go func() {
			r, e := readWire(ctxTest(t), d, co)
			if e == nil {
				var ans pdu.Message
				ans, e = co.CompleteAnswer(r, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "replay-peer", AuthApplicationID: []uint32{appID}}, n.peers["dra2"].origin)
				if e == nil {
					e = writeWire(ctxTest(t), d, co, ans)
				}
			}
			ch <- e
		}()
		if _, e := n.Connect(ctxTest(t), "dra2", barrier); e != nil {
			t.Fatal(e)
		}
		if e := <-ch; e != nil {
			t.Fatal(e)
		}
		dup := otherArrival(t, co, req, 271)
		if e := writeWire(ctxTest(t), d, co, dup); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, d, dup)
		<-barrier.entered
		n.tick(until)
		_ = n.Close()
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal("replay cache borrow released", n.Resources())
		}
		once.Do(func() { close(barrier.release) })
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().DuplicateRecords != 0 {
			t.Fatal(n.Resources())
		}
	})
}

func TestDuplicateReconnectOnlyReplaysNewArrival(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		c := duplicateConfig(t)
		c.Peers[0].InboundCommands[0].Duplicate.ReplayRetention = time.Minute
		n, co := newInboundNode(t, c)
		p, a := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, a, co, 280)
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
		old := p.Info().ConnectionID
		_ = a.Close()
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		p, b := networkConnectNode(t, n, co, tls)
		if p.Info().ConnectionID == old || p.Info().Ready {
			t.Fatal(p.Info())
		}
		for i := 0; i < 3; i++ {
			if i > 0 {
				n.mu.Lock()
				due := p.conn.tw
				n.mu.Unlock()
				n.tick(due)
			}
			dwr, e := readWire(ctxTest(t), b, co)
			if e != nil || dwr.Header.CommandCode != 280 {
				t.Fatal("unsolicited old application answer", dwr, e)
			}
			dwa, e := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), b, co, dwa); e != nil {
				t.Fatal(e)
			}
			want := i + 1
			waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return p.conn.numDWA == want })
		}
		if !p.Info().Ready {
			t.Fatal(p.Info())
		}
		dup := otherArrival(t, co, req, 281)
		if e := writeWire(ctxTest(t), b, co, dup); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, b, dup)
		noDuplicateDelivery(t, p)
	})
}
func TestDuplicateResultVariantsAndCanceledRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, duplicateConfig(t))
		p, a := networkConnectNode(t, n, co, tls)
		for i, body := range []pdu.Body{&s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"), ServingNode: avp.Some(s6c.ServingNode{MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example"), MMENumber: avp.Some([]byte{0x21, 0x43, 0xf5})})}, &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 3002}, nil} {
			in, req := receiveInbound(t, n, p, a, co, uint32(290+i))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if r, e := in.Respond(ctx, body); r.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(r, e)
			}
			var r ResponseResult
			var e error
			if body == nil {
				r, e = in.Reject(ctxTest(t), FailureContextUnavailable)
			} else {
				r, e = in.Respond(ctxTest(t), body)
			}
			if e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			original := duplicateAnswer(t, n, co, a, req)
			dup := otherArrival(t, co, req, uint32(390+i))
			if e = writeWire(ctxTest(t), a, co, dup); e != nil {
				t.Fatal(e)
			}
			replay := duplicateAnswer(t, n, co, a, dup)
			x, _, e := splitResult(reqBytes(t, co, original), n.cfg.Limits)
			if e != nil {
				t.Fatal(e)
			}
			y, _, e := splitResult(reqBytes(t, co, replay), n.cfg.Limits)
			if e != nil || !bytes.Equal(x, y) {
				t.Fatal("result changed", e)
			}
		}
	})
}

func TestDuplicateReplayAssemblyRechecksClosing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := twoDuplicatePeers(duplicateConfig(t))
		key := cfg.Peers[0].InboundCommands[0].Command
		reqOp, _ := cfg.Bundle.Lookup(key)
		key.Request = false
		ansOp, _ := cfg.Bundle.Lookup(key)
		original := ansOp.Decode
		var armed atomic.Bool
		entered, release := make(chan struct{}), make(chan struct{})
		var once, releaseOnce sync.Once
		defer releaseOnce.Do(func() { close(release) })
		ansOp.Decode = func(d *wire.Decoder) (pdu.Body, error) {
			if armed.Load() {
				once.Do(func() { close(entered); <-release })
			}
			return original(d)
		}
		sp, e := schema.NewProfile("replay-closing", "1", strings.Repeat("1", 64), reqOp, ansOp)
		if e != nil {
			t.Fatal(e)
		}
		bp, e := base.Profile()
		if e != nil {
			t.Fatal(e)
		}
		cfg.Bundle, e = schema.Compile(bp, sp)
		if e != nil {
			t.Fatal(e)
		}
		n, co := newInboundNode(t, cfg)
		p, a := duplicateConnect(t, n, co, tls, "dra")
		p2, b := duplicateConnect(t, n, co, tls, "dra2")
		in, req := receiveInbound(t, n, p, a, co, 300)
		if _, e = in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		duplicateAnswer(t, n, co, a, req)
		armed.Store(true)
		dup := otherArrival(t, co, req, 301)
		if e = writeWire(ctxTest(t), b, co, dup); e != nil {
			t.Fatal(e)
		}
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("no assembly")
		}
		n.mu.Lock()
		n.enterClosingLocked(p2.conn, time.Now().Add(100*time.Millisecond), ErrRemoteDPR)
		n.mu.Unlock()
		releaseOnce.Do(func() { close(release) })
		if _, e = readAdapter(ctxTest(t), b); e == nil {
			t.Fatal("application answer started after Closing")
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal("closing destroyed promised cache", n.Resources())
		}
	})
}
