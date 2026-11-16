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
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
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
func TestShPublicExample(t *testing.T) {
	for _, code := range []uint32{2001, 5012, 3004} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			var out bytes.Buffer
			r, e := run(exampleContext(t), code, &out)
			if e != nil {
				t.Fatal(e)
			}
			if r.Answers != ([2]uint32{code, code}) || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
				t.Fatal(r)
			}
			if !strings.Contains(out.String(), fmt.Sprintf("left->right result=%d; right->left result=%d", code, code)) {
				t.Fatal(out.String())
			}
		})
	}
}
func TestShPublicConfiguration(t *testing.T) {
	for name, mut := range map[string]func(*diameter.NodeConfig){
		"wrong binding":          func(c *diameter.NodeConfig) { c.Peers[0].Commands[0].ApplicationBinding = diameter.SCAP1Accounting },
		"insufficient transport": func(c *diameter.NodeConfig) { c.Resources.TransportBytes = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			c, e := endpointConfig(pdu.Origin{Host: "left.example", Realm: "example"}, pdu.Origin{Host: "right.example", Realm: "example"})
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
func TestShPublicResponseAndXMLBoundary(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	body, e := newPull(p.left.node)
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
	// Each invalid preparation leaves the inbound available for a valid answer.
	for name, mut := range map[string]func(*sh.UDA){
		"missing success data": func(a *sh.UDA) { a.UserData = avp.Optional[[]byte]{} },
		"failure with data":    func(a *sh.UDA) { a.ResultCode = avp.Some(uint32(5012)) },
		"wrong message VSAI":   func(a *sh.UDA) { a.VendorSpecificApplicationID.VendorID = 193 },
	} {
		t.Run(name, func(t *testing.T) {
			a, e := newAnswer(2001)
			if e != nil {
				t.Fatal(e)
			}
			mut(a.(*sh.UDA))
			r, e := in.Respond(ctx, a)
			if e == nil || r.Committed {
				t.Fatal(r, e)
			}
		})
	}
	// Diameter owns opaque bytes: XML rejection is an explicit application check.
	a, e := newAnswer(2001)
	if e != nil {
		t.Fatal(e)
	}
	a.(*sh.UDA).UserData = avp.Some([]byte("<Sh-Data>"))
	if r, e := in.Respond(ctx, a); e != nil || !r.Committed {
		t.Fatal(r, e)
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil || c.RequestID != id || c.Result.Err != nil || !c.Result.Validation.Complete {
		t.Fatal(c, e)
	}
	raw := c.Result.Answer.Body.(*sh.UDA).UserData.Value
	if _, e := shdata.Parse(raw, dataLimits()); !errors.Is(e, shdata.ErrXML) {
		t.Fatal(e)
	}
	good, e := newAnswer(2001)
	if e != nil {
		t.Fatal(e)
	}
	raw = good.(*sh.UDA).UserData.Value
	limits := dataLimits()
	limits.MaxBytes = 16
	limits.MaxTextBytes = 16
	limits.MaxOpaqueBytes = 16
	if _, e := shdata.Parse(raw, limits); !errors.Is(e, shdata.ErrLimit) {
		t.Fatal(e)
	}
}
func TestShPublicCanceledAndInvalidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := run(ctx, 2001, io.Discard); e == nil {
		t.Fatal("canceled run succeeded")
	}
	if _, e := run(exampleContext(t), 2002, io.Discard); e == nil {
		t.Fatal("unselected result accepted")
	}
}
func TestShPublicBudgetLifetimeCancelAndClose(t *testing.T) {
	p := examplePair(t)
	ctx := exampleContext(t)
	body, e := newPull(p.left.node)
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
	in.Message().Body.(*sh.UDR).SessionID = "modified-by-handler"
	if r, e := in.Respond(ctx, &sh.UDA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: avp.Some(uint32(2001))}); r.Committed || e == nil {
		t.Fatal(r, e)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	answer, e := newAnswer(2001)
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
	if a, ok := completed.Result.Answer.Body.(*sh.UDA); !ok || a.SessionID != body.SessionID {
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
	body, e = newPull(p.left.node)
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
	body, e = newPull(p.left.node)
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
func TestShPublicMissingNegotiatedApplication(t *testing.T) {
	for _, missing := range []int{0, 1} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			origins := [2]pdu.Origin{{Host: "left.example", Realm: "example"}, {Host: "right.example", Realm: "example"}}
			var configs [2]diameter.NodeConfig
			for i := range configs {
				c, e := endpointConfig(origins[i], origins[1-i])
				if e != nil {
					t.Fatal(e)
				}
				configs[i] = c
			}
			configs[missing].Capabilities.VendorSpecificApplicationID = nil
			p, e := connectPair(exampleContext(t), configs)
			if p != nil {
				_ = p.close()
			}
			if e == nil {
				t.Fatal("missing Sh capability negotiated successfully")
			}
		})
	}
}
