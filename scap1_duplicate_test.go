// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func scapDuplicateConfig(t *testing.T) NodeConfig {
	c := scapRuntimeConfig(t)
	d := duplicateConfig(t)
	c.Resources.NodeBytes = d.Resources.NodeBytes
	c.Resources.DuplicateClasses = d.Resources.DuplicateClasses
	p := &c.Peers[0].InboundCommands[0]
	p.NoCache = false
	p.Duplicate = d.Peers[0].InboundCommands[0].Duplicate
	p.Duplicate.MaxCanonicalBytes = 2048
	return c
}
func TestSCAP1MultiArrivalCacheAndConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{2001, 5012, 3004} {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				cfg := scapDuplicateConfig(t)
				cfg.Peers[0].SCAP1Advertisement = SCAP1LegacyDualAdvertisement
				n, co := newInboundNode(t, cfg)
				p, remote, e := scapHandshake(t, n, co, tls, false, LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{scapCapability()}}, true)
				if e != nil {
					t.Fatal(e)
				}
				req, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7801, EndToEnd: 17801})
				if e != nil {
					t.Fatal(e)
				}
				req.Body.(*scap1.ACR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra-one.example", ProxyState: []byte{0, 255}}}
				if e = writeWire(ctxTest(t), remote, co, req); e != nil {
					t.Fatal(e)
				}
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				dup, e := co.Snapshot(req)
				if e != nil {
					t.Fatal(e)
				}
				dup.Header.HopByHop++
				dup.Header.Flags |= pdu.RetransmittedFlag
				dup.Body.(*scap1.ACR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra-two.example", ProxyState: []byte{1, 254}}}
				if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
					t.Fatal(e)
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
				noDuplicateDelivery(t, p)
				before := n.Resources()
				n.mu.Lock()
				until := in.duplicate.keepUntil
				n.mu.Unlock()
				conflict, e := co.Snapshot(dup)
				if e != nil {
					t.Fatal(e)
				}
				conflict.Header.HopByHop++
				conflict.Body.(*scap1.ACR).ServiceParameterInfo[6].ParameterValue = []byte("changed-one-parent")
				if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
					t.Fatal(e)
				}
				readReject(t, remote, conflict.Header, 5012)
				noDuplicateDelivery(t, p)
				after := n.Resources()
				if after.DuplicateBytes != before.DuplicateBytes || after.DuplicateRecords != 1 || after.DuplicateConflicts != before.DuplicateConflicts+1 {
					t.Fatal(before, after)
				}
				n.mu.Lock()
				same := in.duplicate.keepUntil.Equal(until) && !in.duplicate.committed
				n.mu.Unlock()
				if !same {
					t.Fatal("conflict changed record")
				}
				in.Message().Body.(*scap1.ACR).AccountingRecordNumber = 99
				bad := scapAnswer(2001)
				bad.AccountingRecordNumber = 99
				if r, e := in.Respond(ctxTest(t), bad); r.Committed || e == nil {
					t.Fatal(r, e)
				}
				var answer pdu.Body = scapAnswer(code)
				if code == 3004 {
					answer = &pdu.ErrorAnswer{ResultCode: code}
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				if r, e := in.Respond(ctx, answer); r.Committed || !errors.Is(e, context.Canceled) {
					t.Fatal(r, e)
				}
				if r, e := in.Respond(ctxTest(t), answer); !r.Committed || e != nil {
					t.Fatal(r, e)
				}
				expected := map[uint32]pdu.Message{req.Header.HopByHop: req, dup.Header.HopByHop: dup}
				for range 2 {
					a := routeRead(t, co, remote)
					r, ok := expected[a.Header.HopByHop]
					if !ok {
						t.Fatal("duplicate/wrong arrival answer", a.Header)
					}
					delete(expected, a.Header.HopByHop)
					if v := co.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, validation.Context{}); !v.Complete {
						t.Fatal(v)
					}
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				dup.Header.HopByHop += 10
				if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
					t.Fatal(e)
				}
				duplicateAnswer(t, n, co, remote, dup)
				noDuplicateDelivery(t, p)
				n.Close()
				if e = n.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				if r := n.Resources(); r.InboundBytes != 0 || r.DuplicateBytes != 0 || r.DuplicateRecords != 0 {
					t.Fatal(r)
				}
			})
		}
	})
}
func TestSCAP1MultiArrivalAdmissionAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"arrival-limit", "cache-limit", "active-close"} {
			t.Run(kind, func(t *testing.T) {
				cfg := scapDuplicateConfig(t)
				cfg.Resources.DuplicateClasses[0].Slots = 1
				cfg.Peers[0].InboundCommands[0].Duplicate.MaxArrivals = 1
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				req, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 7805, EndToEnd: 17805})
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, req); e != nil {
					t.Fatal(e)
				}
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				if kind != "active-close" {
					r, e := co.Snapshot(req)
					if e != nil {
						t.Fatal(e)
					}
					r.Header.HopByHop++
					if kind == "cache-limit" {
						r.Header.EndToEnd++
					}
					if e = writeWire(ctxTest(t), remote, co, r); e != nil {
						t.Fatal(e)
					}
					// C/arrival exhaustion follows the existing generic fail-closed contract;
					// unlike pre-body I-budget refusal it does not generate E/3004.
					waitFor(t, func() bool { return !p.Info().Ready })
					if !errors.Is(p.Info().Cause, ErrBackpressure) {
						t.Fatal(p.Info())
					}
					if r := n.Resources(); r.DuplicateRecords != 1 {
						t.Fatal(r)
					}
				}
				start := make(chan struct{})
				var wg sync.WaitGroup
				wg.Add(2)
				go func() { defer wg.Done(); <-start; _, _ = in.Respond(ctxTest(t), scapAnswer(2001)) }()
				go func() { defer wg.Done(); <-start; _ = n.Close() }()
				close(start)
				wg.Wait()
				if e = n.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				if r := n.Resources(); r.InboundBytes != 0 || r.DuplicateBytes != 0 || r.DuplicateRecords != 0 {
					t.Fatal(r)
				}
			})
		}
	})
}
func TestSCAP1LegacyReconnectDropsCapability(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := scapRuntimeConfig(t)
		cfg.Peers[0].SCAP1Advertisement = SCAP1LegacyDualAdvertisement
		cfg.Capabilities.AuthApplicationID = []uint32{appID}
		cfg.Peers[0].Commands = append(cfg.Peers[0].Commands, inboundConfig(t).Peers[0].Commands[0])
		n, co := newInboundNode(t, cfg)
		p, remote, e := scapHandshake(t, n, co, tls, false, LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{scapCapability()}}, true)
		if e != nil {
			t.Fatal(e)
		}
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, remote)
		old := p.Info().ConnectionID
		remote.Close()
		r := routeCompletion(t, n, port, id)
		if r.Err == nil || !r.Delivery.MayHaveReachedPeer {
			t.Fatal(r)
		}
		waitFor(t, func() bool { return p.Info().Candidates == 0 && p.Info().Protocol == ProtocolClosed })
		p, second, e := scapHandshake(t, n, co, tls, false, LocalCapabilities{AuthApplicationID: []uint32{appID}}, false)
		if e != nil {
			t.Fatal(e)
		}
		finishReopen(t, n, p, co, second)
		if p.Info().ConnectionID == old {
			t.Fatal("old generation")
		}
		if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrNotReady) {
			t.Fatal("old capability reused", e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = second.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic replay", e)
		}
	})
}
