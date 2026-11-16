// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6 (Codex)
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Combined-profile duplicate acceptance tests generated with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

var combinedDuplicateKinds = []string{"iec", "general", "balance", "refund", "initial", "term"}

func combinedDuplicateConfig(t *testing.T) NodeConfig {
	t.Helper()
	c := combinedConfig(t)
	d := duplicateConfig(t)
	c.Resources.NodeBytes = 1 << 30
	c.Resources.DuplicateClasses = d.Resources.DuplicateClasses
	c.Resources.InboundClasses[0].Bytes = 256 << 20
	c.Peers[0].MaxInboundBytes = 256 << 20
	ip := &c.Peers[0].InboundCommands[0]
	ip.NoCache = false
	ip.Duplicate = d.Peers[0].InboundCommands[0].Duplicate
	ip.Duplicate.ReplayRetention = 5 * time.Second
	return c
}

func combinedDuplicateRequest(t *testing.T, n *Node, co *codec.Codec, kind string, hbh, e2e uint32) pdu.Message {
	t.Helper()
	r, err := co.CompleteRequest(combinedBody(t, kind), pdu.RequestEnvelope{
		OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm,
		DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm,
		HopByHop: hbh, EndToEnd: e2e,
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Body.(*combined.CCR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "first-dra.example", ProxyState: []byte{0, 255}}}
	return r
}

func combinedArrival(t *testing.T, co *codec.Codec, request pdu.Message, hbh uint32, retransmit bool) pdu.Message {
	t.Helper()
	r, err := co.Snapshot(request)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.HopByHop = hbh
	r.Header.Flags &^= pdu.RetransmittedFlag
	if retransmit {
		r.Header.Flags |= pdu.RetransmittedFlag
	}
	r.Body.(*combined.CCR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "next-dra.example", ProxyState: []byte{byte(hbh), 0, 254}}}
	return r
}

func combinedDuplicateResult(t *testing.T, n *Node, co *codec.Codec, request, answer pdu.Message, code uint32) {
	t.Helper()
	if v := co.ValidateAnswerContext(request, answer, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, validation.Context{}); !v.Complete {
		t.Fatal(v)
	}
	var proxies []pdu.ProxyInfo
	switch body := answer.Body.(type) {
	case *combined.CCA:
		if code == 3004 || body.ResultCode != code {
			t.Fatal("wrong cached normal result", body.ResultCode, code)
		}
		proxies = body.ProxyInfo
	case *pdu.ErrorAnswer:
		if code != 3004 || body.ResultCode != code {
			t.Fatal("wrong cached E result", body.ResultCode, code)
		}
		proxies = body.ProxyInfo
	default:
		t.Fatal("wrong cached answer type")
	}
	want := request.Body.(*combined.CCR).ProxyInfo
	if len(proxies) != len(want) {
		t.Fatal("wrong proxy chain length")
	}
	for i := range want {
		if proxies[i].ProxyHost != want[i].ProxyHost || !bytes.Equal(proxies[i].ProxyState, want[i].ProxyState) {
			t.Fatal("cached first-arrival proxy leaked")
		}
	}
}

func combinedSend(t *testing.T, co *codec.Codec, remote transport.Conn, r pdu.Message) {
	t.Helper()
	if err := writeWire(ctxTest(t), remote, co, r); err != nil {
		t.Fatal(err)
	}
}

func TestCombinedDuplicateTemplatesAndCachedResults(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range combinedDuplicateKinds {
			for _, code := range []uint32{2001, 5012, 3004} {
				t.Run(fmt.Sprintf("%s/%d", kind, code), func(t *testing.T) {
					n, co := newInboundNode(t, combinedDuplicateConfig(t))
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := combinedDuplicateRequest(t, n, co, kind, 8101, 18101)
					combinedSend(t, co, remote, req)
					in, err := p.Receive(ctxTest(t))
					if err != nil {
						t.Fatal(err)
					}
					dup := combinedArrival(t, co, req, 8102, true)
					combinedSend(t, co, remote, dup)
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
					noDuplicateDelivery(t, p)
					before := n.Resources()
					n.mu.Lock()
					record := in.duplicate
					until, deadline := record.keepUntil, record.deadline
					n.mu.Unlock()
					// Change the second Subscription-Id and its required UEI copy,
					// preserving template validity while changing canonical content.
					conflict := combinedArrival(t, co, req, 8103, false)
					conflict.Body.(*combined.CCR).SubscriptionID[1].SubscriptionIDData[0] = '9'
					if b := conflict.Body.(*combined.CCR); b.UserEquipmentInfo.Present {
						b.UserEquipmentInfo.Value.UserEquipmentInfoValue[0] = '9'
					}
					combinedSend(t, co, remote, conflict)
					readReject(t, remote, conflict.Header, 5012)
					noDuplicateDelivery(t, p)
					if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes || s.DuplicateConflicts != before.DuplicateConflicts+1 {
						t.Fatal("conflict changed cache accounting", before, s)
					}
					n.mu.Lock()
					unchanged := record.keepUntil.Equal(until) && record.deadline.Equal(deadline) && !record.committed
					n.mu.Unlock()
					if !unchanged {
						t.Fatal("processing conflict replaced or refreshed record")
					}
					// The delivered mutable view cannot alter private answer authority.
					view := in.Message().Body.(*combined.CCR)
					view.CCRequestNumber = 99
					view.ProxyInfo[0].ProxyState[0] = 99
					bad := combinedAnswer(req.Body.(*combined.CCR), 2001)
					bad.CCRequestNumber = 99
					if r, err := in.Respond(ctxTest(t), bad); err == nil || r.Committed {
						t.Fatal("invalid response committed", r, err)
					}
					var answer pdu.Body = combinedAnswer(req.Body.(*combined.CCR), code)
					if code == 3004 {
						answer = &pdu.ErrorAnswer{ResultCode: code}
					}
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					if r, err := in.Respond(ctx, answer); !errors.Is(err, context.Canceled) || r.Committed {
						t.Fatal(r, err)
					}
					if r, err := in.Respond(ctxTest(t), answer); err != nil || !r.Committed {
						t.Fatal(r, err)
					}
					// Committed result bytes must no longer alias caller-owned bodies.
					switch b := answer.(type) {
					case *combined.CCA:
						b.ResultCode = 4012
					case *pdu.ErrorAnswer:
						b.ResultCode = 3002
					}
					want := map[uint32]pdu.Message{req.Header.HopByHop: req, dup.Header.HopByHop: dup}
					for range 2 {
						a := routeRead(t, co, remote)
						r, ok := want[a.Header.HopByHop]
						if !ok {
							t.Fatal("wrong or duplicate arrival answer")
						}
						delete(want, a.Header.HopByHop)
						combinedDuplicateResult(t, n, co, r, a, code)
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
					late := combinedArrival(t, co, req, 8104, false)
					combinedSend(t, co, remote, late)
					combinedDuplicateResult(t, n, co, late, routeRead(t, co, remote), code)
					conflict.Header.HopByHop = 8105
					combinedSend(t, co, remote, conflict)
					readReject(t, remote, conflict.Header, 5012)
					noDuplicateDelivery(t, p)
					n.mu.Lock()
					unchanged = record.keepUntil.Equal(until) && record.deadline.Equal(deadline) && record.committed
					n.mu.Unlock()
					if !unchanged || n.Resources().DuplicateConflicts != before.DuplicateConflicts+2 {
						t.Fatal("cached hit/conflict refreshed or replaced record")
					}
				})
			}
		}
	})
}

func TestCombinedDuplicateCrossTemplateIdentity(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for i, original := range combinedDuplicateKinds {
			t.Run(original, func(t *testing.T) {
				n, co := newInboundNode(t, combinedDuplicateConfig(t))
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, original, 8201, 18201)
				combinedSend(t, co, remote, req)
				in, err := p.Receive(ctxTest(t))
				if err != nil {
					t.Fatal(err)
				}
				for j, other := range combinedDuplicateKinds {
					if i != j {
						conflict := combinedDuplicateRequest(t, n, co, other, uint32(8210+j), req.Header.EndToEnd)
						combinedSend(t, co, remote, conflict)
						readReject(t, remote, conflict.Header, 5012)
					}
				}
				noDuplicateDelivery(t, p)
				if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateConflicts != 5 {
					t.Fatal(s)
				}
				if _, err = in.Respond(ctxTest(t), combinedAnswer(req.Body.(*combined.CCR), 2001)); err != nil {
					t.Fatal(err)
				}
				combinedDuplicateResult(t, n, co, req, routeRead(t, co, remote), 2001)
			})
		}
	})
}

func TestCombinedDuplicateAdmissionRetentionAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, mode := range []string{"arrivals", "cache", "expiry", "close"} {
			t.Run(mode, func(t *testing.T) {
				cfg := combinedDuplicateConfig(t)
				cfg.Resources.DuplicateClasses[0].Slots = 1
				cfg.Peers[0].InboundCommands[0].Duplicate.MaxArrivals = 1
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "refund", 8301, 18301)
				combinedSend(t, co, remote, req)
				in, err := p.Receive(ctxTest(t))
				if err != nil {
					t.Fatal(err)
				}
				n.mu.Lock()
				until := in.duplicate.keepUntil
				n.mu.Unlock()
				if mode == "cache" || mode == "expiry" {
					if _, err = in.Respond(ctxTest(t), combinedAnswer(req.Body.(*combined.CCR), 2001)); err != nil {
						t.Fatal(err)
					}
					combinedDuplicateResult(t, n, co, req, routeRead(t, co, remote), 2001)
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				}
				switch mode {
				case "arrivals", "cache":
					other := combinedArrival(t, co, req, 8302, true)
					if mode == "cache" {
						other.Header.EndToEnd++
					}
					combinedSend(t, co, remote, other)
					waitFor(t, func() bool { return !p.Info().Ready })
					if !errors.Is(p.Info().Cause, ErrBackpressure) || n.Resources().DuplicateRecords != 1 {
						t.Fatal(p.Info(), n.Resources())
					}
					// Neither a new identity nor another arrival evicts the promised record.
					n.tick(until)
					waitFor(t, func() bool { return n.Resources().DuplicateRecords == 0 && n.Resources().DuplicateBytes == 0 })
				case "expiry":
					// Advance the existing timer input, without a wall-clock sleep.
					n.tick(until)
					waitFor(t, func() bool { return n.Resources().DuplicateRecords == 0 && n.Resources().DuplicateBytes == 0 })
					late := combinedArrival(t, co, req, 8303, false)
					combinedSend(t, co, remote, late)
					again, err := p.Receive(ctxTest(t))
					if err != nil {
						t.Fatal(err)
					}
					if again.Info().RequestID == in.Info().RequestID {
						t.Fatal("expired cache reused logical work")
					}
					if _, err = again.Respond(ctxTest(t), combinedAnswer(req.Body.(*combined.CCR), 5012)); err != nil {
						t.Fatal(err)
					}
					combinedDuplicateResult(t, n, co, late, routeRead(t, co, remote), 5012)
				case "close":
					start := make(chan struct{})
					var wg sync.WaitGroup
					wg.Add(2)
					go func() {
						defer wg.Done()
						<-start
						_, _ = in.Respond(ctxTest(t), combinedAnswer(req.Body.(*combined.CCR), 2001))
					}()
					go func() { defer wg.Done(); <-start; _ = n.Close() }()
					close(start)
					wg.Wait()
				}
				_ = n.Close()
				if err = n.Wait(ctxTest(t)); err != nil {
					t.Fatal(err)
				}
				if s := n.Resources(); s != (ResourceSnapshot{}) {
					t.Fatal("retained resources", s)
				}
			})
		}
	})
}
