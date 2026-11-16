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
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
)

func exampleContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func examplePair(t *testing.T, mode diameter.SCAP1Advertisement) *pair {
	t.Helper()
	p, e := openPair(exampleContext(t), mode, mode)
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
func TestSCAP1PublicExample(t *testing.T) {
	for _, mode := range []diameter.SCAP1Advertisement{diameter.SCAP1StandardAdvertisement, diameter.SCAP1LegacyDualAdvertisement} {
		for _, code := range []uint32{2001, 5012, 3004} {
			t.Run(fmt.Sprintf("mode%d/result%d", mode, code), func(t *testing.T) {
				var out bytes.Buffer
				r, e := run(exampleContext(t), mode, code, &out)
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
}
func TestSCAP1PublicConfiguration(t *testing.T) {
	for name, mut := range map[string]func(*diameter.NodeConfig){
		"no local capability": func(c *diameter.NodeConfig) { c.Capabilities.VendorSpecificApplicationID = nil },
		"wrong vendor":        func(c *diameter.NodeConfig) { c.Capabilities.VendorSpecificApplicationID[0].VendorID = 10415 },
		"dual local capability": func(c *diameter.NodeConfig) {
			c.Capabilities.VendorSpecificApplicationID[0].AuthApplicationID = avp.Some(uint32(19302))
		},
		"outbound missing binding": func(c *diameter.NodeConfig) { c.Peers[0].Commands[0].ApplicationBinding = diameter.HeaderApplication },
		"inbound missing binding": func(c *diameter.NodeConfig) {
			c.Peers[0].InboundCommands[0].ApplicationBinding = diameter.HeaderApplication
		},
		"unfixed terminal":            func(c *diameter.NodeConfig) { c.Peers[0].Commands[0].DestinationHost = "" },
		"unknown advertisement":       func(c *diameter.NodeConfig) { c.Peers[0].SCAP1Advertisement = 99 },
		"insufficient control budget": func(c *diameter.NodeConfig) { c.Resources.TransportBytes = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			cfg, e := endpointConfig(pdu.Origin{Host: "left.example", Realm: "example"}, pdu.Origin{Host: "right.example", Realm: "example"}, diameter.SCAP1LegacyDualAdvertisement)
			if e != nil {
				t.Fatal(e)
			}
			mut(&cfg)
			n, e := diameter.NewNode(cfg)
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
func TestSCAP1PublicBudgetLifetimeCancelAndClose(t *testing.T) {
	for _, mode := range []diameter.SCAP1Advertisement{diameter.SCAP1StandardAdvertisement, diameter.SCAP1LegacyDualAdvertisement} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			p := examplePair(t, mode)
			ctx := exampleContext(t)
			body, e := newEvent(p.left.node)
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
			in.Message().Body.(*scap1.ACR).SessionID = "modified-by-handler"
			if r, e := in.Respond(ctx, &scap1.ACA{ResultCode: 2001, AccountingRecordType: 1, AccountingRecordNumber: 99}); r.Committed || e == nil {
				t.Fatal(r, e)
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			answer := &scap1.ACA{ResultCode: 2001, AccountingRecordType: 1}
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
			if a, ok := completed.Result.Answer.Body.(*scap1.ACA); !ok || a.SessionID != body.SessionID {
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
			body, e = newEvent(p.left.node)
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
			body, e = newEvent(p.left.node)
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
		})
	}
}
func TestSCAP1PublicAdvertisementMismatch(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprint(reverse), func(t *testing.T) {
			a, b := diameter.SCAP1StandardAdvertisement, diameter.SCAP1LegacyDualAdvertisement
			if reverse {
				a, b = b, a
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			p, e := openPair(ctx, a, b)
			if p != nil {
				_ = p.close()
			}
			if e == nil {
				t.Fatal("strict peer accepted legacy advertisement")
			}
		})
	}
}
func TestSCAP1PublicCanceledAndInvalidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := run(ctx, diameter.SCAP1StandardAdvertisement, 2001, io.Discard); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := run(context.Background(), diameter.SCAP1StandardAdvertisement, 2002, io.Discard); e == nil {
		t.Fatal("unselected example result")
	}
}
