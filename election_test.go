package diameter

import (
	"errors"
	"github.com/skill215/diameterkit/codec"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/testkit"
	"github.com/skill215/diameterkit/pdu"
)

func peerCER(n *Node) pdu.Message {
	return pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 41, EndToEnd: 42}, Body: &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake", AuthApplicationID: []uint32{appID}}}
}
func TestSimultaneousConnectionsElectOneDirection(t *testing.T) {
	// Upper-case spelling must have no effect on the RFC identity comparison.
	lowCfg := testConfig(t)
	highCfg := testConfig(t)
	lowCfg.Origin.Host = "Alpha.example"
	highCfg.Origin.Host = "Zulu.example"
	lowCfg.Peers[0].Origin = highCfg.Origin
	highCfg.Peers[0].Origin = lowCfg.Origin
	a, err := NewNode(lowCfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewNode(highCfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := ctxTest(t)
	defer func() {
		_ = a.Close()
		_ = b.Close()
		if err := a.Wait(ctx); err != nil {
			t.Error(err)
		}
		if err := b.Wait(ctx); err != nil {
			t.Error(err)
		}
	}()
	ao, bi := testkit.Pipe()
	bo, ai := testkit.Pipe()
	results := make(chan error, 4)
	go func() { _, e := a.Connect(ctx, "dra", ao); results <- e }()
	go func() { _, e := b.Connect(ctx, "dra", bo); results <- e }()
	ap, _ := a.Peer("dra")
	bp, _ := b.Peer("dra")
	waitFor(t, func() bool { return ap.Info().Candidates == 1 && bp.Info().Candidates == 1 })
	go func() { _, e := a.Accept(ctx, "dra", ai); results <- e }()
	go func() { _, e := b.Accept(ctx, "dra", bi); results <- e }()
	wins, losses := 0, 0
	for range 4 {
		e := <-results
		if e == nil {
			wins++
		} else {
			losses++
		}
	}
	if wins != 2 || losses != 2 || ap.Info().Protocol != IOpen || bp.Info().Protocol != ROpen || !ap.Info().Ready || !bp.Info().Ready {
		t.Fatal(wins, losses, ap.Info(), bp.Info())
	}
	waitFor(t, func() bool { return ap.Info().Candidates == 1 && bp.Info().Candidates == 1 })
	// A later third connection cannot hot-replace the active generation.
	before := ap.Info().ConnectionID
	x, y := testkit.Pipe()
	defer y.Close()
	if _, err = a.Accept(ctx, "dra", x); !errors.Is(err, ErrNotReady) {
		t.Fatal(err)
	}
	if ap.Info().ConnectionID != before {
		t.Fatal("replaced open generation")
	}
}
func TestOutgoingFailurePromotesValidatedPassive(t *testing.T) {
	n, co := newTestNode(t)
	ctx := ctxTest(t)
	out, outRemote := testkit.Pipe()
	defer outRemote.Close()
	active := make(chan error, 1)
	go func() { _, err := n.Connect(ctx, "dra", out); active <- err }()
	if _, err := readMessage(ctx, outRemote, co); err != nil {
		t.Fatal(err)
	}
	in, inRemote := testkit.Pipe()
	defer inRemote.Close()
	passive := make(chan error, 1)
	go func() { _, err := n.Accept(ctx, "dra", in); passive <- err }()
	if err := writeMessage(ctx, inRemote, co, peerCER(n)); err != nil {
		t.Fatal(err)
	}
	p, _ := n.Peer("dra")
	waitFor(t, func() bool {
		n.mu.Lock()
		defer n.mu.Unlock()
		for _, c := range p.candidates {
			if c != nil && !c.initiator && c.prefixResult != nil {
				return true
			}
		}
		return false
	})
	// Both physical slots are occupied until their borrowers retire.
	third, other := testkit.Pipe()
	defer other.Close()
	if _, err := n.Accept(ctx, "dra", third); err == nil {
		t.Fatal("third candidate accepted")
	}
	_ = outRemote.Close()
	if err := <-active; err == nil {
		t.Fatal("failed active succeeded")
	}
	cea, err := readMessage(ctx, inRemote, co)
	if err != nil || cea.Header.CommandCode != 257 {
		t.Fatal(cea, err)
	}
	if err = <-passive; err != nil {
		t.Fatal(err)
	}
	if p.Info().Protocol != ROpen || !p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestPassiveEOFLeavesOutgoingHandshakeIntact(t *testing.T) {
	n, co := newTestNode(t)
	ctx := ctxTest(t)
	out, remote := testkit.Pipe()
	defer remote.Close()
	active := make(chan error, 1)
	go func() { _, e := n.Connect(ctx, "dra", out); active <- e }()
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	in, other := testkit.Pipe()
	passive := make(chan error, 1)
	go func() { _, e := n.Accept(ctx, "dra", in); passive <- e }()
	if err = writeMessage(ctx, other, co, peerCER(n)); err != nil {
		t.Fatal(err)
	}
	_ = other.Close()
	if err = <-passive; err == nil {
		t.Fatal("dead passive succeeded")
	}
	cea, err := co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake", AuthApplicationID: []uint32{appID}}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, cea); err != nil {
		t.Fatal(err)
	}
	if err = <-active; err != nil {
		t.Fatal(err)
	}
	p, _ := n.Peer("dra")
	if !p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestShutdownDuringElectionCannotInstallCandidate(t *testing.T) {
	n, co := newTestNode(t)
	ctx := ctxTest(t)
	out, remote := testkit.Pipe()
	defer remote.Close()
	active := make(chan error, 1)
	go func() { _, e := n.Connect(ctx, "dra", out); active <- e }()
	if _, err := readMessage(ctx, remote, co); err != nil {
		t.Fatal(err)
	}
	in, other := testkit.Pipe()
	defer other.Close()
	passive := make(chan error, 1)
	go func() { _, e := n.Accept(ctx, "dra", in); passive <- e }()
	if err := writeMessage(ctx, other, co, peerCER(n)); err != nil {
		t.Fatal(err)
	}
	if err := n.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-active; err == nil {
		t.Fatal("active installed after shutdown")
	}
	if err := <-passive; err == nil {
		t.Fatal("passive installed after shutdown")
	}
	p, _ := n.Peer("dra")
	if p.Info().Candidates != 0 || p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestElectionCandidateUsesOriginalHandshakeDeadline(t *testing.T) {
	n, co := newTestNode(t)
	ctx := ctxTest(t)
	a, b := testkit.Pipe()
	defer b.Close()
	done := make(chan error, 2)
	go func() { _, e := n.Connect(ctx, "dra", a); done <- e }()
	if _, err := readMessage(ctx, b, co); err != nil {
		t.Fatal(err)
	}
	c, d := testkit.Pipe()
	defer d.Close()
	go func() { _, e := n.Accept(ctx, "dra", c); done <- e }()
	p, _ := n.Peer("dra")
	waitFor(t, func() bool { return p.Info().Candidates == 2 })
	n.mu.Lock()
	deadline := p.handshakeDeadline
	for _, candidate := range p.candidates {
		if candidate.handshakeDeadline.After(deadline) {
			t.Error("election refreshed deadline")
		}
	}
	n.mu.Unlock()
	_ = p.Close()
	<-done
	<-done
}

func TestUnverifiedPassiveIdentityCannotWinElection(t *testing.T) {
	n, co := newTestNode(t)
	ctx := ctxTest(t)
	out, remote := testkit.Pipe()
	defer remote.Close()
	active := make(chan error, 1)
	go func() { _, e := n.Connect(ctx, "dra", out); active <- e }()
	req, err := readMessage(ctx, remote, co)
	if err != nil {
		t.Fatal(err)
	}
	in, other := testkit.Pipe()
	defer other.Close()
	passive := make(chan error, 1)
	go func() { _, e := n.Accept(ctx, "dra", in); passive <- e }()
	cer := peerCER(n)
	cer.Body.(*base.CER).OriginHost = "aaa.example"
	if err = writeMessage(ctx, other, co, cer); err != nil {
		t.Fatal(err)
	}
	if err = <-passive; !errors.Is(err, ErrProtocol) {
		t.Fatal(err)
	}
	cea, err := co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "fake", AuthApplicationID: []uint32{appID}}, draOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeMessage(ctx, remote, co, cea); err != nil {
		t.Fatal(err)
	}
	if err = <-active; err != nil {
		t.Fatal(err)
	}
	p, _ := n.Peer("dra")
	if !p.Info().Ready {
		t.Fatal(p.Info())
	}
}
func TestElectionTimeoutClosesBothCandidates(t *testing.T) {
	cfg := testConfig(t)
	cfg.Timeouts.Handshake = 80 * time.Millisecond
	n, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	co, err := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ctxTest(t)
	defer func() { _ = n.Close(); _ = n.Wait(ctx) }()
	a, b := testkit.Pipe()
	defer b.Close()
	done := make(chan error, 2)
	go func() { _, e := n.Connect(ctx, "dra", a); done <- e }()
	if _, err = readMessage(ctx, b, co); err != nil {
		t.Fatal(err)
	}
	c, d := testkit.Pipe()
	defer d.Close()
	go func() { _, e := n.Accept(ctx, "dra", c); done <- e }()
	if err = writeMessage(ctx, d, co, peerCER(n)); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		select {
		case err = <-done:
			if err == nil {
				t.Fatal("timed out candidate installed")
			}
		case <-ctx.Done():
			t.Fatal("candidate did not retire")
		}
	}
	p, _ := n.Peer("dra")
	if p.Info().Ready || p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}
