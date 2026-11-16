//go:build linux && !386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"context"
	"encoding/binary"
	"errors"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
	upstream "github.com/skill215/sctp"
	"sync"
	"syscall"
	"testing"
	"time"
)

const appID uint32 = 16777312

var serverOrigin = pdu.Origin{Host: "hss.example", Realm: "example"}
var draOrigin = pdu.Origin{Host: "dra.example", Realm: "example"}

func nodeConfig(t *testing.T) diameter.NodeConfig {
	t.Helper()
	bp, err := base.Profile()
	if err != nil {
		t.Fatal(err)
	}
	sp, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		t.Fatal(err)
	}
	cfg := diameter.NodeConfig{Origin: pdu.Origin{Host: "client.example", Realm: "example"}, Bundle: bundle, Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 2048, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16}, Resources: diameter.ResourcePlan{NodeBytes: 128 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096, Classes: []diameter.ClassConfig{{ID: "sms", Bytes: 64 << 20, Slots: 16}}, MaxPorts: 8, MaxPortOutstanding: 16, AdapterBytesPerConnection: 4096, MaxDiscardBytes: 4096}, Timeouts: diameter.Timeouts{Handshake: time.Second, FrameRead: time.Second, FrameWrite: time.Second, Request: 3 * time.Second, Watchdog: 30 * time.Second, ShutdownTimeout: 3 * time.Second, DPAWait: time.Second, CloseReserve: time.Second, ClosingGrace: time.Second}, Capabilities: diameter.LocalCapabilities{HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-test", AuthApplicationID: []uint32{appID}}, Peers: []diameter.PeerConfig{{ID: "dra", Origin: draOrigin, MaxPending: 4, MaxBytes: 16 << 20, Commands: []diameter.CommandPolicy{{Command: pdu.CommandKey{ApplicationID: appID, Code: 8388647, Request: true}, Class: "sms", DestinationRealm: "example", Sources: pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}, ErrorAgents: []pdu.Origin{draOrigin}}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second}}}}}
	info, e := DefaultConfig().Requirements()
	if e != nil {
		t.Fatal(e)
	}
	cfg.Resources.AdapterBytesPerConnection = info.BufferBytes
	return cfg
}
func ownNode(t *testing.T, cfg diameter.NodeConfig) *diameter.Node {
	t.Helper()
	n, e := diameter.NewNode(cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		n.Close()
		if e := n.Wait(testContext(t)); e != nil {
			t.Error(e)
		}
	})
	return n
}
func nodePair(t *testing.T) (*diameter.Node, *diameter.Peer, *diameter.Node, *diameter.Peer) {
	t.Helper()
	client := nodeConfig(t)
	server := nodeConfig(t)
	server.Origin = serverOrigin
	client.Peers[0].Origin = serverOrigin
	server.Peers[0].Origin = client.Origin
	server.Peers[0].Commands = nil
	server.Resources.NodeBytes = 256 << 20
	server.Resources.InboundClasses = []diameter.ClassConfig{{ID: "sms", Bytes: 64 << 20, Slots: 4}}
	server.Peers[0].MaxInbound = 4
	server.Peers[0].MaxInboundBytes = 64 << 20
	server.Peers[0].InboundCommands = []diameter.InboundPolicy{{Command: pdu.CommandKey{ApplicationID: appID, Code: 8388647, Request: true}, Class: "sms", Sources: []pdu.Origin{client.Origin}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, ExecutionTimeout: 2 * time.Second, NoCache: true}}
	n, m := ownNode(t, client), ownNode(t, server)
	a, b := adapters(t)
	ch := make(chan error, 1)
	go func() { _, e := m.Accept(testContext(t), "dra", b); ch <- e }()
	p, e := n.Connect(testContext(t), "dra", a)
	if e != nil {
		t.Fatal(e)
	}
	if e = <-ch; e != nil {
		t.Fatal(e)
	}
	q, _ := m.Peer("dra")
	if !p.Info().Ready || !q.Info().Ready {
		t.Fatal(p.Info(), q.Info())
	}
	return n, p, m, q
}
func requestBody() *s6c.SRR {
	return &s6c.SRR{SessionID: "client.example;1;1", AuthSessionState: 1, MSISDN: avp.Some([]byte{0x21, 0x43, 0xf5}), SCAddress: []byte{0x21, 0x43, 0xf5}}
}
func success() *s6c.SRA {
	return &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"), ServingNode: avp.Some(s6c.ServingNode{MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example"), MMENumber: avp.Some([]byte{0x21, 0x43, 0xf5})})}
}
func waitFor(t *testing.T, f func() bool) {
	t.Helper()
	ctx := testContext(t)
	for !f() {
		select {
		case <-ctx.Done():
			t.Fatal("did not settle")
		case <-time.After(time.Millisecond):
		}
	}
}
func TestSCTPNodesBaseS6cSuccessFailureAndShutdown(t *testing.T) {
	n, p, m, q := nodePair(t)
	for _, body := range []pdu.Body{success(), &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, &pdu.ErrorAnswer{ResultCode: 3002}} {
		done := make(chan error, 1)
		go func() {
			in, e := q.Receive(testContext(t))
			if e == nil {
				var r diameter.ResponseResult
				r, e = in.Respond(testContext(t), body)
				if e == nil && !r.Committed {
					e = errors.New("not committed")
				}
			}
			done <- e
		}()
		r, e := p.Request(testContext(t), requestBody(), diameter.RequestOptions{})
		if e != nil || r.Outcome != diameter.AnswerReceived {
			t.Fatal(r, e)
		}
		if e = <-done; e != nil {
			t.Fatal(e)
		}
	}
	if p.Info().Transport.Events == 0 || p.Info().Transport.Last.Kind != transport.EventAssociationUp {
		t.Fatal(p.Info())
	}
	if e := n.Shutdown(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if e := n.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if p.Info().LastDPAResult != 2001 {
		t.Fatal(p.Info())
	}
	m.Close()
	if e := m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 || m.Resources().InboundBytes != 0 || m.Resources().InboundOutstanding != 0 || p.Info().Candidates != 0 || q.Info().Candidates != 0 {
		t.Fatal(n.Resources(), m.Resources(), p.Info(), q.Info())
	}
}
func TestSCTPNodesCancelLateAnswerAndCloseRace(t *testing.T) {
	n, p, m, q := nodePair(t)
	port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 2, MaxOutstandingBytes: 64 << 20})
	if e != nil {
		t.Fatal(e)
	}
	defer port.Close(testContext(t))
	id, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := q.Receive(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	if e = n.Cancel(id); e != nil {
		t.Fatal(e)
	}
	c, e := port.Receive(testContext(t))
	if e != nil || c.Result.Outcome != diameter.WrittenNoAnswer {
		t.Fatal(c, e)
	}
	if _, e = in.Respond(testContext(t), success()); e != nil {
		t.Fatal(e)
	}
	id, e = p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
	if e != nil {
		t.Fatal(e)
	}
	in, e = q.Receive(testContext(t))
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	wg.Go(func() { n.Cancel(id) })
	wg.Go(func() { in.Respond(testContext(t), success()) })
	wg.Go(func() { n.Close() })
	wg.Go(func() { m.Close() })
	wg.Wait()
	c, e = port.Receive(testContext(t))
	if e != nil || c.RequestID != id || !c.Result.Delivery.MayHaveReachedPeer {
		t.Fatal(c, e)
	}
	if e = n.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	if e = m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("duplicate completion", e)
	}
	if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 || m.Resources().InboundBytes != 0 || m.Resources().InboundOutstanding != 0 || p.Info().Candidates != 0 || q.Info().Candidates != 0 {
		t.Fatal(n.Resources(), m.Resources())
	}
}
func TestSCTPNodeReserveRejectedBeforeHandshake(t *testing.T) {
	cfg := nodeConfig(t)
	cfg.Resources.AdapterBytesPerConnection--
	n := ownNode(t, cfg)
	a, b := adapters(t)
	if _, e := n.Connect(testContext(t), "dra", a); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	if _, e := prefix(testContext(t), b); e == nil {
		t.Fatal("CER sent or socket leaked")
	}
	p, _ := n.Peer("dra")
	if p.Info().Candidates != 0 {
		t.Fatal(p.Info())
	}
}
func TestSCTPNodeAssociationTerminationSettlesPending(t *testing.T) {
	n, p, m, q := nodePair(t)
	port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 64 << 20})
	if e != nil {
		t.Fatal(e)
	}
	defer port.Close(testContext(t))
	id, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = q.Receive(testContext(t)); e != nil {
		t.Fatal(e)
	}
	m.Close()
	if e = m.Wait(testContext(t)); e != nil {
		t.Fatal(e)
	}
	r, e := port.Receive(testContext(t))
	if e != nil || r.RequestID != id || r.Result.Outcome != diameter.WrittenNoAnswer {
		t.Fatal(r, e)
	}
	waitFor(t, func() bool { return p.Info().Candidates == 0 })
	f := p.Info()
	eventTermination := (f.Transport.Last.Kind == transport.EventShutdown || f.Transport.Last.Kind == transport.EventAssociationLost) && f.Transport.Last.Terminal
	var eventCause *transport.Event
	eventTermination = eventTermination || errors.As(f.Cause, &eventCause) && eventCause.Terminal
	// A writer may observe kernel shutdown before the notification reader wins.
	if f.Ready || f.Watchdog != diameter.Down || (!eventTermination && !errors.Is(f.Cause, syscall.ESHUTDOWN)) {
		t.Fatal(f)
	}
	if n.Resources().Outstanding != 0 {
		t.Fatal(n.Resources())
	}
}
func TestSCTPHandshakeCancellationReleasesCandidate(t *testing.T) {
	for _, accept := range []bool{false, true} {
		n := ownNode(t, nodeConfig(t))
		a, b := adapters(t)
		defer b.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		var e error
		if accept {
			_, e = n.Accept(ctx, "dra", a)
		} else {
			_, e = n.Connect(ctx, "dra", a)
		}
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		p, _ := n.Peer("dra")
		if p.Info().Candidates != 0 || p.Info().Ready {
			t.Fatal(p.Info())
		}
	}
}

type recordedWrite struct {
	code  uint32
	flags uint16
	sid   uint16
	ppid  uint32
}
type recordingSocket struct {
	messageSocket
	mu     sync.Mutex
	writes []recordedWrite
}

func (r *recordingSocket) WriteMsgContext(ctx context.Context, b []byte, i upstream.SndInfo) (int, error) {
	r.mu.Lock()
	if len(r.writes) < 32 {
		r.writes = append(r.writes, recordedWrite{code: binary.BigEndian.Uint32(b[4:8]) & 0xffffff, flags: i.Flags, sid: i.SID, ppid: i.PPID})
	}
	r.mu.Unlock()
	return r.messageSocket.WriteMsgContext(ctx, b, i)
}
func TestSCTPResponderOrderedBarrierAndTimeout(t *testing.T) {
	for _, respond := range []bool{true, false} {
		testSCTPBarrier(t, DefaultConfig(), respond)
	}
}
func testSCTPBarrier(t *testing.T, mode Config, respond bool) {
	cfg := nodeConfig(t)
	cfg.Timeouts.Handshake = 300 * time.Millisecond
	n := ownNode(t, cfg)
	rawA, rawB := modeSockets(t, mode, mode)
	a, err := New(rawA, mode)
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(rawB, mode)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	concrete := a.(*conn)
	rec := &recordingSocket{messageSocket: concrete.socket}
	concrete.socket = rec
	co, e := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if e != nil {
		t.Fatal(e)
	}
	send := func(m pdu.Message) error {
		frame, e := co.Append(nil, m)
		if e != nil {
			return e
		}
		return b.WriteFrame(testContext(t), frame).Err
	}
	recv := func() (pdu.Message, error) {
		frame, e := read(testContext(t), b)
		if e != nil {
			return pdu.Message{}, e
		}
		return co.Decode(frame)
	}
	cer := &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: cfg.Capabilities.HostIPAddress, ProductName: "sctp-test", AuthApplicationID: []uint32{appID}}
	if e = send(pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 11, EndToEnd: 12}, Body: cer}); e != nil {
		t.Fatal(e)
	}
	p, e := n.Accept(testContext(t), "dra", a)
	if e != nil {
		t.Fatal(e)
	}
	cea, e := recv()
	if e != nil || cea.Header.CommandCode != 257 {
		t.Fatal(cea, e)
	}
	dwr, e := recv()
	if e != nil || dwr.Header.CommandCode != 280 {
		t.Fatal(dwr, e)
	}
	if !p.Info().SendBarrier {
		t.Fatal("barrier removed by transport events")
	}
	if !respond {
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		if !errors.Is(p.Info().Cause, context.DeadlineExceeded) {
			t.Fatal(p.Info())
		}
	} else {
		// An admitted application frame before DWA is still ordered stream0.
		port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 2, MaxOutstandingBytes: 64 << 20})
		if e != nil {
			t.Fatal(e)
		}
		first, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = recv(); e != nil {
			t.Fatal(e)
		}
		dwa, e := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = send(dwa); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return !p.Info().SendBarrier })
		second, e := p.Submit(testContext(t), requestBody(), diameter.RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = recv(); e != nil {
			t.Fatal(e)
		}
		n.Cancel(first)
		n.Cancel(second)
		port.Close(testContext(t))
	}
	rec.mu.Lock()
	writes := append([]recordedWrite(nil), rec.writes...)
	rec.mu.Unlock()
	ppid := uint32(46)
	if mode.PPIDPolicy == PPID0Compatibility {
		ppid = 0
	}
	if len(writes) < 2 || writes[0] != (recordedWrite{code: 257, ppid: ppid}) || writes[1] != (recordedWrite{code: 280, ppid: ppid}) {
		t.Fatal(writes)
	}
	if respond {
		if len(writes) != 4 || writes[2] != (recordedWrite{code: 8388647, ppid: ppid}) {
			t.Fatal(writes)
		}
		post := writes[3]
		if post.code != 8388647 || post.ppid != ppid {
			t.Fatal(post)
		}
		if mode.StreamMode == OrderedMultiStream {
			if post.flags != 0 || post.sid == 0 || post.sid >= concrete.outbound {
				t.Fatal(post)
			}
		} else if post.flags != upstream.SCTP_UNORDERED || post.sid != 0 {
			t.Fatal(post)
		}
	}

}
