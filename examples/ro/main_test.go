// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func exampleContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func examplePair(t *testing.T) *pair {
	t.Helper()
	p, e := openPair(exampleContext(t))
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
func TestRoPublicExample(t *testing.T) {
	for _, code := range []string{"2001", "2001/4012", "4012", "5012", "3004"} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			var out bytes.Buffer
			r, e := run(exampleContext(t), code, &out)
			if e != nil {
				t.Fatal(e)
			}
			if r.Answer != resultName(code) || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
				t.Fatal(r)
			}
			if !strings.Contains(out.String(), "ctf->ocs result="+resultName(code)) {
				t.Fatal(out.String())
			}
		})
	}
}
func TestRoPublicConfiguration(t *testing.T) {
	for name, mut := range map[string]func(*diameter.NodeConfig){
		"wrong binding":          func(c *diameter.NodeConfig) { c.Peers[0].Commands[0].ApplicationBinding = diameter.SCAP1Accounting },
		"insufficient transport": func(c *diameter.NodeConfig) { c.Resources.TransportBytes = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			c, e := endpointConfig(pdu.Origin{Host: "ctf.example", Realm: "example"}, pdu.Origin{Host: "ocs.example", Realm: "example"}, true)
			if e != nil {
				t.Fatal(e)
			}
			mut(&c)
			n, e := diameter.NewNode(c)
			if n != nil {
				n.Close()
				_ = n.Wait(exampleContext(t))
			}
			if !errors.Is(e, diameter.ErrConfig) {
				t.Fatal(e)
			}
		})
	}
}
func TestRoPublicResponseRules(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	body, e := newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	id, e := p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	for name, mut := range map[string]func(*ro.CCA){
		"wrong type":           func(a *ro.CCA) { a.CCRequestType = ro.InitialRequest },
		"wrong number":         func(a *ro.CCA) { a.CCRequestNumber = 1 },
		"wrong service":        func(a *ro.CCA) { a.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2 },
		"missing child result": func(a *ro.CCA) { a.MultipleServicesCreditControl[0].ResultCode = avp.Optional[uint32]{} },
		"missing grant": func(a *ro.CCA) {
			a.MultipleServicesCreditControl[0].GrantedServiceUnit = avp.Optional[ro.ServiceUnit]{}
		},
		"unselected experimental pair": func(a *ro.CCA) {
			a.ExperimentalResult = avp.Some(ro.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})
		},
	} {
		t.Run(name, func(t *testing.T) {
			a, e := newAnswer("2001")
			if e != nil {
				t.Fatal(e)
			}
			mut(a.(*ro.CCA))
			if r, e := in.Respond(ctx, a); r.Committed || e == nil {
				t.Fatal(r, e)
			}
		})
	}
	a, e := newAnswer("2001/4012")
	if e != nil {
		t.Fatal(e)
	}
	if r, e := in.Respond(ctx, a); !r.Committed || e != nil {
		t.Fatal(r, e)
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id || c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
		t.Fatal(c, e)
	}
	if e := checkAnswer(c.Result.Answer.Body, "2001/4012", body.SessionID); e != nil {
		t.Fatal(e)
	}
	// Neither success at the command level nor a service denial triggers follow-up.
	quiet, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if in, e := p.right.peer.Receive(quiet); in != nil || !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("automatic business retry", in, e)
	}
}
func TestRoPublicRequestBoundary(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	for name, mut := range map[string]func(*ro.CCR){
		"missing explicit zero action": func(r *ro.CCR) { r.RequestedAction = avp.Optional[uint32]{} },
		"unsupported action":           func(r *ro.CCR) { r.RequestedAction = avp.Some(uint32(1)) },
		"wrong context":                func(r *ro.CCR) { r.ServiceContextID = "other@example.net" },
		"missing SMS information":      func(r *ro.CCR) { r.ServiceInformation = avp.Optional[ro.ServiceInformation]{} },
		"wrong number":                 func(r *ro.CCR) { r.CCRequestNumber = 1 },
		"wrong host":                   func(r *ro.CCR) { r.DestinationHost = avp.Some("other.example") },
		"wrong realm":                  func(r *ro.CCR) { r.DestinationRealm = "other.example" },
	} {
		t.Run(name, func(t *testing.T) {
			r, e := newDebit(p.left.node)
			if e != nil {
				t.Fatal(e)
			}
			mut(r)
			if _, e := p.left.peer.Submit(ctx, r, diameter.RequestOptions{}, p.left.port); e == nil {
				t.Fatal("invalid request admitted")
			}
			if r := p.left.node.Resources(); r.Pending != 0 || r.Outstanding != 0 || r.OutboundBytes != 0 {
				t.Fatal(r)
			}
		})
	}
	quiet, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if in, e := p.right.peer.Receive(quiet); in != nil || !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(in, e)
	}
	port, e := p.right.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
	if e != nil {
		t.Fatal(e)
	}
	p.right.port = port
	r, e := newDebit(p.right.node)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := p.right.peer.Submit(ctx, r, diameter.RequestOptions{}, port); !errors.Is(e, diameter.ErrUnsupported) {
		t.Fatal("reverse CCR authorized", e)
	}
	if e := exchange(ctx, &p.left, &p.right, "2001"); e != nil {
		t.Fatal(e)
	}
}
func TestRoPublicExactUnits(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	r, e := newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	const exact uint64 = (1 << 53) + 1
	r.MultipleServicesCreditControl[0].RequestedServiceUnit = avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(exact)})
	id, e := p.left.peer.Submit(ctx, r, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if in.Message().Body.(*ro.CCR).MultipleServicesCreditControl[0].RequestedServiceUnit.Value.CCServiceSpecificUnits.Value != exact {
		t.Fatal("request units rounded")
	}
	a, e := newAnswer("2001")
	if e != nil {
		t.Fatal(e)
	}
	a.(*ro.CCA).MultipleServicesCreditControl[0].GrantedServiceUnit = avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(exact)})
	if x, e := in.Respond(ctx, a); e != nil || !x.Committed {
		t.Fatal(x, e)
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id || c.Result.Err != nil {
		t.Fatal(c, e)
	}
	if c.Result.Answer.Body.(*ro.CCA).MultipleServicesCreditControl[0].GrantedServiceUnit.Value.CCServiceSpecificUnits.Value != exact {
		t.Fatal("answer units rounded")
	}
}
func TestRoPublicTerminalSources(t *testing.T) {
	for _, answerSource := range []bool{false, true} {
		t.Run(fmt.Sprint(answerSource), func(t *testing.T) {
			origins := [2]pdu.Origin{{Host: "ctf.example", Realm: "example"}, {Host: "ocs.example", Realm: "example"}}
			var cfg [2]diameter.NodeConfig
			for i := range cfg {
				var e error
				cfg[i], e = endpointConfig(origins[i], origins[1-i], i == 0)
				if e != nil {
					t.Fatal(e)
				}
			}
			other := pdu.Origin{Host: "not-authorized.example", Realm: "example"}
			if answerSource {
				cfg[0].Peers[0].Commands[0].Sources.Terminals = []pdu.Origin{other}
			} else {
				cfg[1].Peers[0].InboundCommands[0].Sources = []pdu.Origin{other}
			}
			ctx := exampleContext(t)
			p, e := connectPair(ctx, cfg)
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() {
				if e := p.close(); e != nil {
					t.Error(e)
				}
				for _, ep := range []*endpoint{&p.left, &p.right} {
					if r := ep.node.Resources(); r != (diameter.ResourceSnapshot{}) {
						t.Error(r)
					}
				}
			})
			r, e := newDebit(p.left.node)
			if e != nil {
				t.Fatal(e)
			}
			id, e := p.left.peer.Submit(ctx, r, diameter.RequestOptions{}, p.left.port)
			if e != nil {
				t.Fatal(e)
			}
			if !answerSource {
				quiet, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
				defer cancel()
				in, e := p.right.peer.Receive(quiet)
				if in != nil || e == nil {
					t.Fatal(in, e)
				}
				_ = p.left.node.Cancel(id)
				return
			}
			in, e := p.right.peer.Receive(ctx)
			if e != nil {
				t.Fatal(e)
			}
			a, e := newAnswer("2001")
			if e != nil {
				t.Fatal(e)
			}
			if x, e := in.Respond(ctx, a); !x.Committed || e != nil {
				t.Fatal(x, e)
			}
			c, e := p.left.port.Receive(ctx)
			if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, validation.ErrInvalidAnswer) || c.Result.Answer.Body != nil {
				t.Fatal(c, e)
			}
		})
	}
}
func TestRoPublicCanceledAndInvalidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := run(ctx, "2001", io.Discard); e == nil {
		t.Fatal("canceled run succeeded")
	}
	if _, e := run(exampleContext(t), "2002", io.Discard); e == nil {
		t.Fatal("unselected result accepted")
	}
}
func TestRoPublicBudgetLifetimeCancelAndClose(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	body, e := newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	admission, stopAdmission := context.WithCancel(ctx)
	id, e := p.left.peer.Submit(admission, body, diameter.RequestOptions{}, p.left.port)
	stopAdmission() // Submit context controls admission, not the accepted request.
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	before := p.left.node.Resources()
	if before.Outstanding != 1 || before.OutboundBytes == 0 || before.Pending != 1 || p.right.node.Resources().InboundOutstanding != 1 {
		t.Fatal(before, p.right.node.Resources())
	}
	if _, e = p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	after := p.left.node.Resources()
	if before.OutboundBytes != after.OutboundBytes || after.Outstanding != 1 {
		t.Fatal(before, after)
	}
	waitOnly, stopWait := context.WithTimeout(ctx, 10*time.Millisecond)
	_, e = p.left.port.Receive(waitOnly)
	stopWait()
	if !errors.Is(e, context.DeadlineExceeded) || p.left.node.Resources().Outstanding != 1 {
		t.Fatal("completion wait canceled the request", e)
	}
	// The application-visible view is independent of the original response envelope.
	in.Message().Body.(*ro.CCR).SessionID = "modified-by-handler"
	if r, e := in.Respond(ctx, &ro.CCA{AuthApplicationID: ro.ApplicationID, CCRequestType: ro.EventRequest, CCRequestNumber: 0, ResultCode: 2001}); r.Committed || e == nil {
		t.Fatal(r, e)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	answer, e := newAnswer("2001")
	if e != nil {
		t.Fatal(e)
	}
	if r, e := in.Respond(canceled, answer); r.Committed || !errors.Is(e, context.Canceled) {
		t.Fatal(r, e)
	}
	if r, e := in.Respond(ctx, answer); !r.Committed || e != nil {
		t.Fatal(r, e)
	}
	// Even after responding, an unconsumed completion still occupies the port.
	if _, e = p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	completed, e := p.left.port.Receive(ctx)
	if e != nil || completed.RequestID != id || completed.Result.Err != nil {
		t.Fatal(completed, e)
	}
	if a, ok := completed.Result.Answer.Body.(*ro.CCA); !ok || a.SessionID != body.SessionID {
		t.Fatal("response used mutable view", completed.Result.Answer)
	}
	if r := p.left.node.Resources(); r.OutboundBytes != 0 || r.Outstanding != 0 || r.Pending != 0 {
		t.Fatal(r)
	}
	// Bytes independently reject, even with an available completion slot.
	if e = p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 1})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal("byte budget", e)
	}
	if r := p.left.node.Resources(); r.Outstanding != 0 || r.OutboundBytes != 0 {
		t.Fatal(r)
	}
	quiet, stop := context.WithTimeout(ctx, 20*time.Millisecond)
	unexpected, e := p.right.peer.Receive(quiet)
	stop()
	if unexpected != nil || !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("rejected submit reached responder", unexpected, e)
	}
	if e = p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
	if e != nil {
		t.Fatal(e)
	}
	body, e = newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	id, e = p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e = p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if e = p.left.node.Cancel(id); e != nil {
		t.Fatal(e)
	}
	completed, e = p.left.port.Receive(ctx)
	if e != nil || completed.RequestID != id || !errors.Is(completed.Result.Err, context.Canceled) {
		t.Fatal(completed, e)
	}
	if r, e := in.Respond(ctx, answer); !r.Committed || e != nil {
		t.Fatal(r, e)
	}
	quiet, stop = context.WithTimeout(ctx, 20*time.Millisecond)
	_, e = p.left.port.Receive(quiet)
	stop()
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("late answer completed twice", e)
	}
	body, e = newDebit(p.left.node)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, p.left.port); e != nil {
		t.Fatal(e)
	}
	if _, e = p.right.peer.Receive(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.node.Close()
	if e = p.left.node.Wait(ctx); e != nil {
		t.Fatal(e)
	}
	// Wait joins I/O; it does not consume this retained completion for the caller.
	if r := p.left.node.Resources(); r.Ports != 1 || r.Outstanding != 1 || r.FixedBytes == 0 || r.OutboundBytes == 0 || r.Pending != 0 {
		t.Fatal(r)
	}
	if e = p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	if r := p.left.node.Resources(); r != (diameter.ResourceSnapshot{}) {
		t.Fatal(r)
	}
}

// HeaderApplication capabilities are negotiated at handshake; NewNode does not
// infer them from profiles or require them as SCAP1Accounting does.
func TestRoPublicMissingNegotiatedApplication(t *testing.T) {
	for _, missing := range []int{0, 1} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			origins := [2]pdu.Origin{{Host: "ctf.example", Realm: "example"}, {Host: "ocs.example", Realm: "example"}}
			var configs [2]diameter.NodeConfig
			for i := range configs {
				c, e := endpointConfig(origins[i], origins[1-i], i == 0)
				if e != nil {
					t.Fatal(e)
				}
				configs[i] = c
			}
			configs[missing].Capabilities.AuthApplicationID = nil
			p, e := connectPair(exampleContext(t), configs)
			if p != nil {
				_ = p.close()
			}
			if e == nil {
				t.Fatal("missing Ro capability negotiated successfully")
			}
		})
	}
}
