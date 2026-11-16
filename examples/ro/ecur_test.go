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
)

func initialFor(t *testing.T, p *pair) *ro.CCR {
	t.Helper()
	r, e := newInitial(p.left.node, p.left.origin, p.right.origin)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func termFor(t *testing.T, r *ro.CCR) *ro.CCR {
	t.Helper()
	term, e := newTermination(r, 1)
	if e != nil {
		t.Fatal(e)
	}
	return term
}
func noECURFollowup(t *testing.T, p *pair) {
	t.Helper()
	ctx, cancel := context.WithTimeout(exampleContext(t), 20*time.Millisecond)
	defer cancel()
	if in, e := p.right.peer.Receive(ctx); in != nil || !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("automatic follow-up", in, e)
	}
}
func TestRoPublicECURFlow(t *testing.T) {
	for _, term := range []string{"2001", "5002", "5012", "3004"} {
		t.Run("term="+term, func(t *testing.T) {
			var out bytes.Buffer
			r, e := runECUR(exampleContext(t), "2001", term, &out)
			if e != nil {
				t.Fatal(e)
			}
			if !r.TermSent || r.Initial != resultName("2001") || r.Termination != termResultName(term) || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
				t.Fatal(r)
			}
			if !strings.Contains(out.String(), "ECUR INITIAL="+resultName("2001")+"; TERMINATION="+termResultName(term)) {
				t.Fatal(out.String())
			}
		})
	}
	for _, initial := range []string{"2001/4012", "4012", "5012", "3004"} {
		t.Run("initial="+initial, func(t *testing.T) {
			var out bytes.Buffer
			r, e := runECUR(exampleContext(t), initial, "2001", &out)
			if e != nil {
				t.Fatal(e)
			}
			if r.TermSent || r.Initial != resultName(initial) || r.Termination != "not-sent" || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
				t.Fatal(r)
			}
			if !strings.Contains(out.String(), "TERMINATION=not-sent") {
				t.Fatal(out.String())
			}
		})
	}
}
func TestRoPublicECURContinuation(t *testing.T) {
	p := examplePair(t)
	initial := initialFor(t, p)
	for name, mut := range map[string]func(*ro.CCR){
		"session":          func(r *ro.CCR) { r.SessionID += ";different" },
		"ctf host":         func(r *ro.CCR) { r.OriginHost = "other.example" },
		"ctf realm":        func(r *ro.CCR) { r.OriginRealm = "other.example" },
		"ocs host":         func(r *ro.CCR) { r.DestinationHost = avp.Some("other.example") },
		"ocs realm":        func(r *ro.CCR) { r.DestinationRealm = "other.example" },
		"realm only":       func(r *ro.CCR) { r.DestinationHost = avp.Optional[string]{} },
		"service":          func(r *ro.CCR) { r.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2 },
		"type UPDATE":      func(r *ro.CCR) { r.CCRequestType = 2 },
		"number":           func(r *ro.CCR) { r.CCRequestNumber = 2 },
		"context":          func(r *ro.CCR) { r.ServiceContextID = "other@example.net" },
		"requested action": func(r *ro.CCR) { r.RequestedAction = avp.Some(ro.DirectDebiting) },
		"RSU in TERM":      func(r *ro.CCR) { r.MultipleServicesCreditControl[0].RequestedServiceUnit = avp.Some(ro.ServiceUnit{}) },
		"missing USU":      func(r *ro.CCR) { r.MultipleServicesCreditControl[0].UsedServiceUnit = nil },
		"missing FINAL":    func(r *ro.CCR) { r.MultipleServicesCreditControl[0].ReportingReason = nil },
	} {
		t.Run(name, func(t *testing.T) {
			r := termFor(t, initial)
			mut(r)
			if e := ro.CheckECURContinuation(initial, r); e == nil {
				t.Fatal("invalid continuation accepted")
			}
		})
	}
	if initial.MultipleServicesCreditControl[0].ServiceIdentifier[0] != 1 || len(initial.MultipleServicesCreditControl[0].UsedServiceUnit) != 0 {
		t.Fatal("termination builder aliased INITIAL")
	}
	zero, e := newTermination(initial, 0)
	if e != nil {
		t.Fatal(e)
	}
	if !zero.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Present {
		t.Fatal("zero usage lost presence")
	}
	// Pure relation checking can pass without an INITIAL answer. It is not a grant.
	if e := ro.CheckECURContinuation(initial, zero); e != nil {
		t.Fatal(e)
	}
	for _, result := range []string{"2001/4012", "4012", "5012", "3004"} {
		a, e := newECURAnswer(false, result)
		if e != nil {
			t.Fatal(e)
		}
		if canTerminate(a, 1) {
			t.Fatal("failed answer authorized TERM", result)
		}
	}
	a, e := newECURAnswer(false, "2001")
	if e != nil {
		t.Fatal(e)
	}
	if canTerminate(a, 2) || !canTerminate(a, 1) {
		t.Fatal("application grant bound")
	}
}
func TestRoPublicECURBudgetsAndAssociation(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	initial := initialFor(t, p)
	term := termFor(t, initial)
	id, e := p.left.peer.Submit(ctx, initial, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	firstHeader := in.Message().Header
	if r := p.left.node.Resources(); r.Outstanding != 1 || r.Pending != 1 || r.OutboundBytes == 0 {
		t.Fatal(r)
	}
	a, e := newECURAnswer(false, "2001")
	if e != nil {
		t.Fatal(e)
	}
	if r, e := in.Respond(ctx, a); e != nil || !r.Committed {
		t.Fatal(r, e)
	}
	if _, e := p.left.peer.Submit(ctx, term, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal("unconsumed INITIAL lost charge", e)
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id || c.Result.Err != nil || !c.Result.Validation.Complete || !canTerminate(c.Result.Answer.Body, 1) {
		t.Fatal(c, e)
	}
	if r := p.left.node.Resources(); r.Outstanding != 0 || r.Pending != 0 || r.OutboundBytes != 0 {
		t.Fatal(r)
	}
	noECURFollowup(t, p)
	// A finite independent byte limit can reject the next exchange after INITIAL.
	if e := p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 1})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := p.left.peer.Submit(ctx, term, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
		t.Fatal(e)
	}
	if r := p.left.node.Resources(); r.Outstanding != 0 || r.OutboundBytes != 0 {
		t.Fatal(r)
	}
	noECURFollowup(t, p)
	if e := p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
	if e != nil {
		t.Fatal(e)
	}
	id2, e := p.left.peer.Submit(ctx, term, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		t.Fatal(e)
	}
	term.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value = 99
	in, e = p.right.peer.Receive(ctx)
	if e != nil {
		t.Fatal(e)
	}
	m := in.Message()
	r := m.Body.(*ro.CCR)
	if id == id2 || firstHeader.HopByHop == m.Header.HopByHop || firstHeader.EndToEnd == m.Header.EndToEnd || r.SessionID != initial.SessionID || r.MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value != 1 {
		t.Fatal("exchange identity or snapshot mismatch")
	}
	if e := ro.CheckECURContinuation(initial, r); e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"old INITIAL answer", "wrong service", "TERM grant", "TERM token"} {
		t.Run(kind, func(t *testing.T) {
			body, e := newECURAnswer(true, "2001")
			if e != nil {
				t.Fatal(e)
			}
			a := body.(*ro.CCA)
			switch kind {
			case "old INITIAL answer":
				body, e = newECURAnswer(false, "2001")
				if e != nil {
					t.Fatal(e)
				}
			case "wrong service":
				a.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{2}, ResultCode: avp.Some(uint32(2001))}}
			case "TERM grant":
				a.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(uint32(2001)), GrantedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))})}}
			case "TERM token":
				a.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(uint32(2001)), RefundInformation: avp.Some([]byte{1})}}
			}
			if x, e := in.Respond(ctx, body); x.Committed || e == nil {
				t.Fatal(x, e)
			}
		})
	}
	a, e = newECURAnswer(true, "5002")
	if e != nil {
		t.Fatal(e)
	}
	if r, e := in.Respond(ctx, a); !r.Committed || e != nil {
		t.Fatal(r, e)
	}
	c, e = p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id2 || c.Result.Err != nil || !c.Result.Validation.Complete || c.Result.Answer.Body.(*ro.CCA).ResultCode != 5002 {
		t.Fatal(c, e)
	}
	noECURFollowup(t, p)
}
func TestRoPublicECURCancelAndClose(t *testing.T) {
	for _, termStep := range []bool{false, true} {
		for _, closeNode := range []bool{false, true} {
			t.Run(fmt.Sprintf("term=%t/close=%t", termStep, closeNode), func(t *testing.T) {
				p := examplePair(t)
				ctx := exampleContext(t)
				initial := initialFor(t, p)
				request := initial
				if termStep {
					if _, e := exchangeECUR(ctx, p, initial, nil, "2001"); e != nil {
						t.Fatal(e)
					}
					request = termFor(t, initial)
				}
				id, e := p.left.peer.Submit(ctx, request, diameter.RequestOptions{}, p.left.port)
				if e != nil {
					t.Fatal(e)
				}
				in, e := p.right.peer.Receive(ctx)
				if e != nil {
					t.Fatal(e)
				}
				if closeNode {
					if e := p.left.node.Close(); e != nil {
						t.Fatal(e)
					}
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
					return
				}
				if e := p.left.node.Cancel(id); e != nil {
					t.Fatal(e)
				}
				c, e := p.left.port.Receive(ctx)
				if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, context.Canceled) {
					t.Fatal(c, e)
				}
				a, e := newECURAnswer(termStep, "2001")
				if e != nil {
					t.Fatal(e)
				}
				if r, e := in.Respond(ctx, a); e != nil || !r.Committed {
					t.Fatal(r, e)
				}
				noECURFollowup(t, p)
				quiet, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
				if _, e := p.left.port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("late answer completed twice", e)
				}
			})
		}
	}
}
func TestRoPublicECURCanceledAndInvalidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := runECUR(ctx, "2001", "2001", io.Discard); e == nil {
		t.Fatal("canceled run succeeded")
	}
	for _, v := range [][2]string{{"5002", "2001"}, {"2001", "4012"}} {
		if _, e := runECUR(exampleContext(t), v[0], v[1], io.Discard); e == nil {
			t.Fatal(v)
		}
	}
}
