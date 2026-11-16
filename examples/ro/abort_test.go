package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func abortPairFor(t *testing.T, cfg [2]diameter.NodeConfig) *pair {
	t.Helper()
	p, e := connectPair(exampleContext(t), cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := p.close(); e != nil {
			t.Error(e)
		}
		for _, ep := range []*endpoint{&p.left, &p.right} {
			if r := ep.node.Resources(); r != (diameter.ResourceSnapshot{}) {
				t.Errorf("retained resources: %+v", r)
			}
		}
	})
	return p
}
func abortConfigFor(t *testing.T) [2]diameter.NodeConfig {
	t.Helper()
	cfg, e := abortConfigs()
	if e != nil {
		t.Fatal(e)
	}
	return cfg
}
func abortRequestFor() *abort.ASR {
	return &abort.ASR{SessionID: knownAbortSession, AuthApplicationID: abort.ApplicationID,
		OriginStateID: avp.Some(uint32(7)), ProxyInfo: []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255}}}}
}
func receiveAbortFor(t *testing.T, p *pair) (diameter.RequestID, *diameter.Inbound) {
	t.Helper()
	id, e := p.left.peer.Submit(exampleContext(t), abortRequestFor(), diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(exampleContext(t))
	if e != nil {
		t.Fatal(e)
	}
	if !in.Validation().Complete {
		t.Fatal(in.Validation())
	}
	return id, in
}
func TestRoPublicAbortFlow(t *testing.T) {
	for _, result := range []string{"2001", "5002", "5012", "3004"} {
		t.Run(result, func(t *testing.T) {
			var out bytes.Buffer
			r, e := runAbort(exampleContext(t), knownAbortSession, result, &out)
			if e != nil || r.Answer != abortResultName(result) || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
				t.Fatal(r, e)
			}
			if !strings.Contains(out.String(), "ocs->ctf abort="+r.Answer) {
				t.Fatal(out.String())
			}
		})
	}
	t.Run("unknown session", func(t *testing.T) {
		r, e := runAbort(exampleContext(t), "ctf.example;unknown", "2001", io.Discard)
		if e != nil || r.Answer != "base:5002" {
			t.Fatal(r, e)
		}
	})
}
func TestRoPublicAbortConfiguration(t *testing.T) {
	for _, missing := range []int{0, 1} {
		t.Run(fmt.Sprintf("missing Auth4/%d", missing), func(t *testing.T) {
			cfg := abortConfigFor(t)
			cfg[missing].Capabilities.AuthApplicationID = nil
			p, e := connectPair(exampleContext(t), cfg)
			if p != nil {
				_ = p.close()
			}
			if e == nil {
				t.Fatal("profile inferred advertised capability")
			}
		})
	}
	for _, wrong := range []string{"binding", "transport"} {
		t.Run(wrong, func(t *testing.T) {
			cfg := abortConfigFor(t)
			if wrong == "binding" {
				cfg[0].Peers[0].Commands[0].ApplicationBinding = diameter.SCAP1Accounting
			} else {
				cfg[0].Resources.TransportBytes = 1
			}
			n, e := diameter.NewNode(cfg[0])
			if n != nil {
				_ = n.Close()
				_ = n.Wait(exampleContext(t))
			}
			if !errors.Is(e, diameter.ErrConfig) {
				t.Fatal(e)
			}
		})
	}
}
func TestRoPublicAbortRequestAndDirection(t *testing.T) {
	p := abortPairFor(t, abortConfigFor(t))
	ctx := exampleContext(t)
	for name, mut := range map[string]func(*abort.ASR){
		"missing session":      func(r *abort.ASR) { r.SessionID = "" },
		"wrong app":            func(r *abort.ASR) { r.AuthApplicationID = 0 },
		"wrong host":           func(r *abort.ASR) { r.DestinationHost = "other.example" },
		"wrong realm":          func(r *abort.ASR) { r.DestinationRealm = "other.example" },
		"CCR number smuggling": func(r *abort.ASR) { r.Extra = []avp.Raw{{Key: avp.Key{Code: 415}, Data: []byte{0, 0, 0, 0}}} },
	} {
		t.Run(name, func(t *testing.T) {
			r := abortRequestFor()
			mut(r)
			if _, e := p.left.peer.Submit(ctx, r, diameter.RequestOptions{}, p.left.port); e == nil {
				t.Fatal("invalid ASR admitted")
			}
			if r := p.left.node.Resources(); r.Outstanding != 0 || r.Pending != 0 || r.OutboundBytes != 0 {
				t.Fatal(r)
			}
		})
	}
	// Loaded CCR definitions and Auth4 do not authorize either CCR or reverse ASR.
	ccr, e := newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.left.peer.Submit(ctx, ccr, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrUnsupported) {
		t.Fatal(e)
	}
	p.right.port, e = p.right.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.right.peer.Submit(ctx, abortRequestFor(), diameter.RequestOptions{}, p.right.port); !errors.Is(e, diameter.ErrUnsupported) {
		t.Fatal("reverse ASR", e)
	}
	if _, e = p.left.node.NewRoute(diameter.RouteConfig{Command: abortRequestFor().CommandKey(), Peers: []diameter.PeerID{remotePeer}, ExchangePolicy: diameter.ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: diameter.ReplayDomain{Target: p.right.origin, Retention: 2 * time.Second}}); !errors.Is(e, diameter.ErrUnsupported) {
		t.Fatal("ASR granted replay", e)
	}
	if _, e = exchangeAbort(ctx, p, knownAbortSession, "2001"); e != nil {
		t.Fatal(e)
	}
}
func TestRoPublicAbortRespondAndBudget(t *testing.T) {
	p := abortPairFor(t, abortConfigFor(t))
	ctx := exampleContext(t)
	id, in := receiveAbortFor(t, p)
	before := p.left.node.Resources()
	if before.Outstanding != 1 || before.Pending != 1 || before.OutboundBytes == 0 || p.right.node.Resources().InboundOutstanding != 1 {
		t.Fatal(before)
	}
	if _, e := p.left.peer.Submit(ctx, abortRequestFor(), diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	wait, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	if _, e := p.left.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	cancel()
	if p.left.node.Resources().Outstanding != 1 {
		t.Fatal("wait consumed request")
	}
	for name, answer := range map[string]pdu.Body{
		"missing result":        &abort.ASA{},
		"unselected result":     &abort.ASA{ResultCode: 4012},
		"normal protocol error": &abort.ASA{ResultCode: 3004},
		"charging field":        &abort.ASA{ResultCode: 2001, Extra: []avp.Raw{{Key: avp.Key{Code: 416}, Data: []byte{0, 0, 0, 1}}}},
		"experimental E":        &pdu.ErrorAnswer{ResultCode: 3004, ExperimentalResult: avp.Some(pdu.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})},
	} {
		t.Run(name, func(t *testing.T) {
			if r, e := in.Respond(ctx, answer); r.Committed || e == nil {
				t.Fatal(r, e)
			}
		})
	}
	stop, stopCancel := context.WithCancel(ctx)
	stopCancel()
	if r, e := in.Respond(stop, &abort.ASA{ResultCode: 2001}); r.Committed || !errors.Is(e, context.Canceled) {
		t.Fatal(r, e)
	}
	// The response envelope retains the actual request, not this mutable view.
	view := in.Message().Body.(*abort.ASR)
	view.SessionID = "changed-view"
	view.ProxyInfo[0].ProxyState[0] = 99
	if r, e := in.Respond(ctx, &abort.ASA{ResultCode: 2001}); e != nil || !r.Committed {
		t.Fatal(r, e)
	}
	if _, e := p.left.peer.Submit(ctx, abortRequestFor(), diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal("unconsumed completion lost charge", e)
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id || c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
		t.Fatal(c, e)
	}
	if e := checkAbortAnswer(c.Result.Answer.Body, "base:2001", knownAbortSession); e != nil {
		t.Fatal(e)
	}
	a := c.Result.Answer.Body.(*abort.ASA)
	if a.OriginHost != p.right.origin.Host || len(a.ProxyInfo) != 1 || !bytes.Equal(a.ProxyInfo[0].ProxyState, []byte{0, 255}) {
		t.Fatal("envelope corruption", a)
	}
	if _, e := in.Respond(ctx, &abort.ASA{ResultCode: 5002}); !errors.Is(e, diameter.ErrInboundDone) {
		t.Fatal(e)
	}
	if r := p.left.node.Resources(); r.Outstanding != 0 || r.OutboundBytes != 0 || r.Pending != 0 {
		t.Fatal(r)
	}
	if e := p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 1})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := p.left.peer.Submit(ctx, abortRequestFor(), diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	if r := p.left.node.Resources(); r.Outstanding != 0 || r.OutboundBytes != 0 {
		t.Fatal(r)
	}
}
func TestRoPublicAbortTerminalSources(t *testing.T) {
	for _, answerSource := range []bool{false, true} {
		t.Run(fmt.Sprint(answerSource), func(t *testing.T) {
			cfg := abortConfigFor(t)
			other := pdu.Origin{Host: "unauthorized.example", Realm: "example"}
			if answerSource {
				cfg[0].Peers[0].Commands[0].Sources.Terminals = []pdu.Origin{other}
			} else {
				cfg[1].Peers[0].InboundCommands[0].Sources = []pdu.Origin{other}
			}
			p := abortPairFor(t, cfg)
			ctx := exampleContext(t)
			id, e := p.left.peer.Submit(ctx, abortRequestFor(), diameter.RequestOptions{}, p.left.port)
			if e != nil {
				t.Fatal(e)
			}
			if !answerSource {
				wait, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
				defer cancel()
				if in, e := p.right.peer.Receive(wait); in != nil || e == nil {
					t.Fatal(in, e)
				}
				_ = p.left.node.Cancel(id)
				return
			}
			in, e := p.right.peer.Receive(ctx)
			if e != nil {
				t.Fatal(e)
			}
			if r, e := in.Respond(ctx, &abort.ASA{ResultCode: 2001}); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			c, e := p.left.port.Receive(ctx)
			if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, validation.ErrInvalidAnswer) || c.Result.Answer.Body != nil {
				t.Fatal(c, e)
			}
		})
	}
}
func TestRoPublicAbortCancelClose(t *testing.T) {
	for _, mode := range []string{"cancel late answer", "close retained result", "respond close race"} {
		t.Run(mode, func(t *testing.T) {
			p := abortPairFor(t, abortConfigFor(t))
			ctx := exampleContext(t)
			id, in := receiveAbortFor(t, p)
			switch mode {
			case "cancel late answer":
				if e := p.left.node.Cancel(id); e != nil {
					t.Fatal(e)
				}
				c, e := p.left.port.Receive(ctx)
				if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, context.Canceled) {
					t.Fatal(c, e)
				}
				if r, e := in.Respond(ctx, &abort.ASA{ResultCode: 2001}); e != nil || !r.Committed {
					t.Fatal(r, e)
				}
				wait, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
				if _, e := p.left.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("second completion", e)
				}
			case "close retained result":
				_ = p.left.node.Close()
				if e := p.left.node.Wait(ctx); e != nil {
					t.Fatal(e)
				}
				if r := p.left.node.Resources(); r.Outstanding != 1 || r.Pending != 0 || r.OutboundBytes == 0 {
					t.Fatal(r)
				}
				if e := p.left.port.Close(ctx); e != nil {
					t.Fatal(e)
				}
				if r := p.left.node.Resources(); r != (diameter.ResourceSnapshot{}) {
					t.Fatal(r)
				}
			case "respond close race":
				var wg sync.WaitGroup
				wg.Add(2)
				go func() { defer wg.Done(); _, _ = in.Respond(ctx, &abort.ASA{ResultCode: 5012}) }()
				go func() { defer wg.Done(); _ = p.right.node.Close() }()
				wg.Wait()
				if e := p.right.node.Wait(ctx); e != nil {
					t.Fatal(e)
				}
				if _, e := in.Respond(ctx, &abort.ASA{ResultCode: 2001}); e == nil {
					t.Fatal("closed response reopened")
				}
				c, e := p.left.port.Receive(ctx)
				if e != nil || c.RequestID != id {
					t.Fatal(c, e)
				}
				wait, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
				if _, e := p.left.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("second completion", e)
				}
			}
		})
	}
}
func TestRoPublicAbortDoesNotOwnCCR(t *testing.T) {
	cfg := abortConfigFor(t)
	// This test explicitly enables the separate CTF->OCS CCR path to observe that
	// successful ASA does not finish an outstanding CCR or create a TERM.
	ocs, e := endpointConfig(cfg[0].Origin, cfg[1].Origin, false)
	if e != nil {
		t.Fatal(e)
	}
	ctf, e := endpointConfig(cfg[1].Origin, cfg[0].Origin, true)
	if e != nil {
		t.Fatal(e)
	}
	cfg[0].Peers[0].InboundCommands = ocs.Peers[0].InboundCommands
	cfg[0].Peers[0].MaxInbound = ocs.Peers[0].MaxInbound
	cfg[0].Peers[0].MaxInboundBytes = ocs.Peers[0].MaxInboundBytes
	cfg[0].Resources.InboundClasses = ocs.Resources.InboundClasses
	cfg[1].Peers[0].Commands = ctf.Peers[0].Commands
	p := abortPairFor(t, cfg)
	ctx := exampleContext(t)
	p.right.port, e = p.right.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
	if e != nil {
		t.Fatal(e)
	}
	initial, e := newInitial(p.right.node, p.right.origin, p.left.origin)
	if e != nil {
		t.Fatal(e)
	}
	initial.SessionID = knownAbortSession
	id, e := p.right.peer.Submit(ctx, initial, diameter.RequestOptions{}, p.right.port)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := p.left.peer.Receive(ctx); e != nil {
		t.Fatal(e)
	}
	if _, e := exchangeAbort(ctx, p, knownAbortSession, "2001"); e != nil {
		t.Fatal(e)
	}
	if r := p.right.node.Resources(); r.Outstanding != 1 || r.Pending != 1 {
		t.Fatal("ASA completed CCR", r)
	}
	for _, from := range []bool{true, false} {
		wait, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
		if from {
			if _, e := p.right.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("CCR auto completed", e)
			}
		} else {
			if in, e := p.left.peer.Receive(wait); in != nil || !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("automatic TERM", in, e)
			}
		}
		cancel()
	}
	if e := p.right.node.Cancel(id); e != nil {
		t.Fatal(e)
	}
	c, e := p.right.port.Receive(ctx)
	if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, context.Canceled) {
		t.Fatal(c, e)
	}
}
func TestRoPublicAbortInvalidAndCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := runAbort(ctx, knownAbortSession, "2001", io.Discard); e == nil {
		t.Fatal("canceled run")
	}
	for _, v := range [][2]string{{"", "2001"}, {knownAbortSession, "4012"}} {
		if _, e := runAbort(exampleContext(t), v[0], v[1], io.Discard); e == nil {
			t.Fatal(v)
		}
	}
	for _, v := range [][4]string{
		{"abort", "2001", "5002", knownAbortSession}, {"abort", "4012", "2001", knownAbortSession},
		{"abort", "2001", "2001", ""}, {"iec", "2001", "2001", "different"}, {"ecur", "2001", "2001", "different"},
	} {
		if validFlowOptions(v[0], v[1], v[2], v[3]) {
			t.Fatal("silently ignored argument", v)
		}
	}
}
