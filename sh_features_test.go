// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"sync"
	"testing"
	"time"
)

func featureWire(bits uint32, m bool) []sh.SupportedFeatures {
	return []sh.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: bits, Mandatory: m}}
}
func peerEvidence(t *testing.T, cmd uint32, requester, responder pdu.Origin, rb, ab uint32) validation.Context {
	t.Helper()
	c, e := sh.PeerFeatures(sh.PeerFeatureEvidence{CommandCode: cmd, Requester: requester, Responder: responder, RequesterKnown: true, ResponderKnown: true, RequesterFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: rb}}, ResponderFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: ab}}})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestShFeaturesInboundFreezeAndReply(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, ab := range []uint32{0, 1, 9} {
			req := shRequest(t, n, co, 306, uint32(2000+i))
			req.Body.(*sh.UDR).SupportedFeatures = featureWire(9, true)
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if r := in.Validation(); r.State != validation.NeedsApplication || r.Complete {
				t.Fatal(r)
			}
			answer := shAnswer(306, 5011, true).(*sh.UDA)
			if ab == 9 {
				answer = shAnswer(306, 2001, false).(*sh.UDA)
			}
			answer.SupportedFeatures = featureWire(ab, false)
			if _, e = in.Respond(ctxTest(t), answer); !errors.Is(e, validation.ErrContextRequired) {
				t.Fatal(e)
			}
			// The mutable public view must never become feature or association evidence.
			in.Message().Body.(*sh.UDR).SupportedFeatures[0].FeatureList = 0
			wrong := peerEvidence(t, 306, gmscOrigin, serverOrigin, 9, ab)
			if r := in.Check(wrong); r.Complete {
				t.Fatal("DRA/wrong local binding accepted")
			}
			ctx := peerEvidence(t, 306, gmscOrigin, n.cfg.Origin, 9, ab)
			if r := in.Check(ctx); !r.Complete {
				t.Fatal(r)
			}
			if r := in.Check(peerEvidence(t, 306, gmscOrigin, n.cfg.Origin, 9, ab|16)); r.Complete {
				t.Fatal("changed frozen evidence")
			}
			if r := in.Validation(); !r.Complete {
				t.Fatal("failed check destroyed prior verdict")
			}
			canceled, cancel := context.WithCancel(ctxTest(t))
			cancel()
			if _, e = in.Respond(canceled, answer); !errors.Is(e, context.Canceled) {
				t.Fatal(e)
			}
			answer.SupportedFeatures[0].Mandatory = true
			if _, e = in.Respond(ctxTest(t), answer); e == nil {
				t.Fatal("answer M1 sent")
			}
			answer.SupportedFeatures[0].Mandatory = false
			if r, e := in.Respond(ctxTest(t), answer); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			duplicateAnswer(t, n, co, remote, req)
			if r := in.Check(ctx); !r.Complete {
				t.Fatal(r)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestShFeaturesOutboundBindingAndCompletion(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		b := subscription(11, 0)
		b.DataReference = []uint32{11, 12}
		b.SupportedFeatures = featureWire(1, true)
		if _, e := p.Submit(ctxTest(t), b, RequestOptions{}, port); !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(e)
		}
		no := peerEvidence(t, 308, n.cfg.Origin, serverOrigin, 1, 0)
		if _, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: no}, port); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
		wrong := peerEvidence(t, 308, n.cfg.Origin, gmscOrigin, 1, 1)
		if _, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: wrong}, port); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
		yes := peerEvidence(t, 308, n.cfg.Origin, serverOrigin, 1, 1)
		id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: yes}, port)
		if e != nil {
			t.Fatal(e)
		}
		b.SupportedFeatures[0].FeatureList = 0
		req := routeRead(t, co, remote)
		if req.Body.(*sh.SNR).SupportedFeatures[0].FeatureList != 1 {
			t.Fatal("caller mutation")
		}
		answer := shAnswer(308, 2001, false).(*sh.SNA)
		answer.SupportedFeatures = featureWire(1, false)
		a, e := co.CompleteAnswer(req, answer, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		done, e := port.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		r := done.Result
		if done.RequestID != id || r.Err != nil || !r.Validation.Complete {
			t.Fatal(r)
		}
		if e := n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(ctxTest(t), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("completed twice", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestShFeaturesUnknownAnswerEvidenceRetained(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		if _, e := p.Submit(ctxTest(t), shBody(306), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		answer := shAnswer(306, 2001, false).(*sh.UDA)
		answer.SupportedFeatures = featureWire(1, false)
		a, e := co.CompleteAnswer(req, answer, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		done, e := port.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		r := done.Result
		if r.Err != nil || r.Answer.Body == nil || r.Validation.State != validation.NeedsApplication {
			t.Fatal(r)
		}
	})
}
func TestShFeaturesCheckCloseBarrier(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		req := shRequest(t, n, co, 308, 2200)
		b := req.Body.(*sh.SNR)
		b.DataReference = []uint32{11, 25}
		b.SupportedFeatures = featureWire(1, true)
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		ctx := peerEvidence(t, 308, gmscOrigin, n.cfg.Origin, 1, 1)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			for range 100 {
				in.Check(ctx)
			}
		}()
		go func() {
			defer wg.Done()
			for range 100 {
				in.Validation()
			}
		}()
		go func() { defer wg.Done(); n.Close() }()
		wg.Wait()
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if r := in.Check(ctx); r.State != validation.Expired {
			t.Fatal(r)
		}
		if n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}

// M0 and M1 offers with identical original identity are different requests.
func TestShFeaturesDuplicateMandatoryConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		req := shRequest(t, n, co, 306, 2300)
		b := req.Body.(*sh.UDR)
		b.SupportedFeatures = featureWire(1, true)
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		ctx := peerEvidence(t, 306, gmscOrigin, n.cfg.Origin, 1, 1)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		conflict := req
		conflict.Header.HopByHop++
		b.SupportedFeatures[0].Mandatory = false
		if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
			t.Fatal(e)
		}
		a := duplicateAnswer(t, n, co, remote, conflict)
		if a.Header.Flags&pdu.ErrorFlag == 0 {
			t.Fatal("conflict reused")
		}
		answer := shAnswer(306, 2001, false).(*sh.UDA)
		answer.SupportedFeatures = featureWire(1, false)
		if r, e := in.Respond(ctxTest(t), answer); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		// Original M1 association, despite mutation of the sender's separate request.
		b.SupportedFeatures[0].Mandatory = true
		duplicateAnswer(t, n, co, remote, req)
		req.Header.HopByHop += 2
		if e = writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		a = duplicateAnswer(t, n, co, remote, req)
		if a.Body.(*sh.UDA).ResultCode != avp.Some(uint32(2001)) {
			t.Fatal(a)
		}
	})
}

func TestShFeaturesCloseWaitsForCheckBorrower(t *testing.T) {
	cfg := shConfig(t, false)
	entered, release := make(chan struct{}), make(chan struct{})
	var commands []schema.Command
	for code := uint32(306); code <= 309; code++ {
		for _, request := range []bool{true, false} {
			op, _ := cfg.Bundle.Lookup(pdu.CommandKey{ApplicationID: sh.ApplicationID, Code: code, Request: request})
			if code == 308 && request {
				original := op.CheckContext
				op.CheckContext = func(r, a pdu.Body, c validation.Context, m validation.Mode) error {
					if !c.Empty() {
						close(entered)
						<-release
					}
					return original(r, a, c, m)
				}
			}
			commands = append(commands, op)
		}
	}
	profile, _ := sh.Profile()
	wrapped, e := schema.NewProfile(profile.Name(), profile.Version(), profile.Digest(), commands...)
	if e != nil {
		t.Fatal(e)
	}
	// Retain the other profiles required by the test config's base and control policy.
	baseProfile, _ := base.Profile()
	cfg.Bundle, e = schema.Compile(baseProfile, wrapped)
	if e != nil {
		t.Fatal(e)
	}
	n, co := newInboundNode(t, cfg)
	p, remote := networkConnectNode(t, n, co, false)
	req := shRequest(t, n, co, 308, 2400)
	req.Body.(*sh.SNR).SupportedFeatures = featureWire(1, true)
	if e = writeWire(ctxTest(t), remote, co, req); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	ctx := peerEvidence(t, 308, gmscOrigin, n.cfg.Origin, 1, 1)
	result := make(chan validation.Report, 1)
	go func() { result <- in.Check(ctx) }()
	<-entered
	before := n.Resources()
	if e = n.Close(); e != nil {
		t.Fatal(e)
	}
	if after := n.Resources(); after.InboundBytes != before.InboundBytes || after.InboundOutstanding != 1 {
		t.Fatal("borrowed reservation released", before, after)
	}
	deadline, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if e = n.Wait(deadline); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("Wait ignored validator borrower", e)
	}
	close(release)
	if r := <-result; r.State != validation.Expired {
		t.Fatal(r)
	}
	if e = n.Wait(ctxTest(t)); e != nil {
		t.Fatal(e)
	}
	if n.Resources().InboundOutstanding != 0 {
		t.Fatal(n.Resources())
	}
}

func TestShFeaturesIndependentMixedNotification(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		req := shRequest(t, n, co, 309, 2500)
		body := req.Body.(*sh.PNR)
		body.SupportedFeatures = featureWire(1, true)
		data, e := shdata.MarshalNotification(shdata.NotificationData{IMSUserState: avp.Some(uint8(1)), Reachability: avp.Some(shdata.Reachability{MME: avp.Some(uint8(0))})}, shdata.DefaultLimits())
		if e != nil {
			t.Fatal(e)
		}
		body.UserData = data
		if e = writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		ctx, e := sh.PeerFeatures(sh.PeerFeatureEvidence{CommandCode: 309, Requester: serverOrigin, Responder: n.cfg.Origin, RequesterKnown: true, ResponderKnown: true, RequesterFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}, ResponderFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}, NotificationReferences: []uint32{11, 25}})
		if e != nil {
			t.Fatal(e)
		}
		parsed, e := shdata.Parse(in.Message().Body.(*sh.PNR).UserData, shdata.DefaultLimits())
		if e != nil {
			t.Fatal(e)
		}
		if r, e := parsed.CheckNotification([]uint32{11, 25}); e != nil || !r.Complete {
			t.Fatal(r, e)
		}
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		// The same declared mixture cannot switch to MSISDN-only identity.
		bad := req
		badBody := *body
		badBody.UserIdentity = sh.UserIdentity{MSISDN: avp.Some([]byte{0x21, 0xf3})}
		bad.Body = &badBody
		if r := co.ValidateContext(bad, ctx); !errors.Is(r.Err(), validation.ErrUnsupported) {
			t.Fatal(r)
		}
		answer := shAnswer(309, 2001, false).(*sh.PNA)
		answer.SupportedFeatures = featureWire(1, false)
		if r, e := in.Respond(ctxTest(t), answer); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
