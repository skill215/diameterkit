// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Cross-peer combined and mixed-profile volatile duplicate acceptance.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"strings"
	"testing"
	"time"
)

var duplicateProfileNames = []string{"combined", "s6c", "sh", "sgd"}
var duplicateOtherOrigin = pdu.Origin{Host: "second-ctf.example", Realm: "example"}

// Each remote explicitly advertises the selected Bundle applications. Peer
// identities differ while an admitted request may retain the same origin.
func crossDuplicateConnect(t *testing.T, n *Node, co *codec.Codec, encrypted bool, id PeerID) (*Peer, transport.Conn) {
	t.Helper()
	a, b := networkAdapters(t, encrypted)
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		m, e := readWire(ctx, b, co)
		if e == nil {
			var ans pdu.Message
			ans, e = co.CompleteAnswer(m, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "duplicate-peer", AuthApplicationID: n.cfg.Capabilities.AuthApplicationID, VendorSpecificApplicationID: n.cfg.Capabilities.VendorSpecificApplicationID}, n.peers[id].origin)
			if e == nil {
				e = writeWire(ctx, b, co, ans)
			}
		}
		ch <- e
	}()
	p, e := n.Connect(ctx, id, a)
	if e != nil {
		t.Fatal(e)
	}
	if e = <-ch; e != nil {
		t.Fatal(e)
	}
	return p, b
}

type duplicateMark struct {
	record          *duplicateRecord
	deadline, until time.Time
	canonical       []byte
}

func markDuplicate(n *Node, in *Inbound) duplicateMark {
	n.mu.Lock()
	defer n.mu.Unlock()
	r := in.duplicate
	return duplicateMark{r, r.deadline, r.keepUntil, append([]byte(nil), r.canonical...)}
}
func assertDuplicateMark(t *testing.T, n *Node, m duplicateMark, committed bool) {
	t.Helper()
	n.mu.Lock()
	r := m.record
	ok := n.duplicates[r.key] == r && r.deadline.Equal(m.deadline) && r.keepUntil.Equal(m.until) && r.committed == committed && bytes.Equal(r.canonical, m.canonical)
	n.mu.Unlock()
	if !ok {
		t.Fatal("record replaced, refreshed or changed")
	}
}
func closeDuplicateNode(t *testing.T, n *Node) {
	t.Helper()
	conflicts := n.Resources().DuplicateConflicts
	_ = n.Close()
	if e := n.Wait(ctxTest(t)); e != nil {
		t.Fatal(e)
	}
	if s := n.Resources(); s != (ResourceSnapshot{DuplicateConflicts: conflicts}) {
		t.Fatal("retained charged resources", s)
	}
}

func TestCombinedCrossPeerResults(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range combinedDuplicateKinds {
			for _, code := range []uint32{2001, 5012, 3004} {
				for _, disconnect := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%d/disconnect-%t", kind, code, disconnect), func(t *testing.T) {
						cfg := twoDuplicatePeers(combinedDuplicateConfig(t))
						cfg.Timeouts.Request = 10 * time.Second
						cfg.Timeouts.Watchdog = 2 * time.Minute
						for i := range cfg.Peers {
							p := &cfg.Peers[i].InboundCommands[0]
							p.ExecutionTimeout = 5 * time.Second
							p.Duplicate.MaxProcessing = p.ExecutionTimeout
							p.Duplicate.ReplayRetention = 30 * time.Second
						}
						n, co := newInboundNode(t, cfg)
						p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
						p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
						req := combinedDuplicateRequest(t, n, co, kind, 9901, 19901)
						combinedSend(t, co, a, req)
						in, e := p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
						mark := markDuplicate(n, in)
						before := n.Resources()
						dup := combinedArrival(t, co, req, 9902, true)
						dup.Body.(*combined.CCR).OriginHost = strings.ToUpper(gmscOrigin.Host)
						combinedSend(t, co, b, dup)
						waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
						noDuplicateDelivery(t, p2)
						if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes {
							t.Fatal("cross-peer double charge", s)
						}
						conflict := combinedArrival(t, co, req, 9903, false)
						cb := conflict.Body.(*combined.CCR)
						cb.SubscriptionID[1].SubscriptionIDData[0] = '9'
						if cb.UserEquipmentInfo.Present {
							cb.UserEquipmentInfo.Value.UserEquipmentInfoValue[0] = '9'
						}
						combinedSend(t, co, b, conflict)
						readReject(t, b, conflict.Header, 5012)
						assertDuplicateMark(t, n, mark, false)
						if disconnect {
							_ = a.Close()
							waitFor(t, func() bool { return p.Info().Candidates == 0 })
						}
						var answer pdu.Body = combinedAnswer(req.Body.(*combined.CCR), code)
						if code == 3004 {
							answer = &pdu.ErrorAnswer{ResultCode: code}
						}
						r, e := in.Respond(ctxTest(t), answer)
						if !r.Committed {
							t.Fatal(r, e)
						}
						if disconnect {
							if e == nil || r.PrimaryWrite.State != ResponseNotWritten {
								t.Fatal("primary disconnect hidden", r, e)
							}
						} else {
							if e != nil {
								t.Fatal(e)
							}
							combinedDuplicateResult(t, n, co, req, routeRead(t, co, a), code)
						}
						combinedDuplicateResult(t, n, co, dup, routeRead(t, co, b), code)
						waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
						late := combinedArrival(t, co, req, 9904, false)
						combinedSend(t, co, b, late)
						combinedDuplicateResult(t, n, co, late, routeRead(t, co, b), code)
						conflict.Header.HopByHop = 9905
						combinedSend(t, co, b, conflict)
						readReject(t, b, conflict.Header, 5012)
						noDuplicateDelivery(t, p2)
						assertDuplicateMark(t, n, mark, true)
						if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes || s.DuplicateConflicts != 2 {
							t.Fatal("cache state changed", s)
						}
						closeDuplicateNode(t, n)
					})
				}
			}
		}
	})
}

func TestCombinedCrossPeerAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"source", "policy", "arrivals"} {
			t.Run(kind, func(t *testing.T) {
				cfg := twoDuplicatePeers(combinedDuplicateConfig(t))
				if kind == "source" {
					cfg.Peers[1].InboundCommands[0].Sources = []pdu.Origin{duplicateOtherOrigin}
				}
				if kind == "policy" {
					cfg.Peers[1].InboundCommands[0].Duplicate.ReplayRetention += time.Second
				}
				if kind == "arrivals" {
					for i := range cfg.Peers {
						cfg.Peers[i].InboundCommands[0].Duplicate.MaxArrivals = 1
					}
				}
				n, co := newInboundNode(t, cfg)
				p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
				p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
				req := combinedDuplicateRequest(t, n, co, "refund", 9911, 19911)
				combinedSend(t, co, a, req)
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				mark := markDuplicate(n, in)
				before := n.Resources()
				dup := combinedArrival(t, co, req, 9912, true)
				combinedSend(t, co, b, dup)
				switch kind {
				case "source":
					readReject(t, b, dup.Header, 3002)
				case "policy":
					readReject(t, b, dup.Header, 5012)
				default:
					waitFor(t, func() bool { return !p2.Info().Ready })
					if !errors.Is(p2.Info().Cause, ErrBackpressure) {
						t.Fatal(p2.Info())
					}
				}
				noDuplicateDelivery(t, p2)
				assertDuplicateMark(t, n, mark, false)
				if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes {
					t.Fatal(s)
				}
				if _, e = in.Respond(ctxTest(t), combinedAnswer(req.Body.(*combined.CCR), 2001)); e != nil {
					t.Fatal(e)
				}
				combinedDuplicateResult(t, n, co, req, routeRead(t, co, a), 2001)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				late := combinedArrival(t, co, req, 9913, false)
				combinedSend(t, co, a, late)
				combinedDuplicateResult(t, n, co, late, routeRead(t, co, a), 2001)
				assertDuplicateMark(t, n, mark, true)
				closeDuplicateNode(t, n)
			})
		}
	})
}

func mixedDuplicateBody(t *testing.T, app int) pdu.Body {
	t.Helper()
	switch app {
	case 0:
		return combinedBody(t, "general")
	case 1:
		return requestBody()
	case 2:
		return shBody(306)
	default:
		return &sgd.OFR{SessionID: "ctf.example;mixed;sgd", AuthSessionState: 1, UserIdentifier: sgd.UserIdentifier{UserName: avp.Some("001010123456789")}, SCAddress: []byte{0x21, 0x43, 0xf5}, SMRPUI: []byte{0, 255, 1}}
	}
}
func mixedDuplicateConfig(t *testing.T, shared bool) NodeConfig {
	t.Helper()
	cfg := combinedDuplicateConfig(t)
	cfg.Timeouts.Request = 10 * time.Second
	cfg.Timeouts.Watchdog = 2 * time.Minute
	var profiles []schema.Profile
	for _, factory := range []func() (schema.Profile, error){base.Profile, combined.Profile, s6c.Profile, sh.Profile, sgd.Profile} {
		p, e := factory()
		if e != nil {
			t.Fatal(e)
		}
		profiles = append(profiles, p)
	}
	var e error
	cfg.Bundle, e = schema.Compile(profiles...)
	if e != nil {
		t.Fatal(e)
	}
	cfg.Resources.NodeBytes = 2 << 30
	cfg.Resources.InboundClasses = []ClassConfig{{ID: "mixed-inbound", Bytes: 512 << 20, Slots: 16}}
	cfg.Resources.DuplicateClasses = nil
	cfg.Capabilities.AuthApplicationID = []uint32{4}
	cfg.Capabilities.VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{sh.Application(), {VendorID: 10415, AuthApplicationID: avp.Some(uint32(16777312))}, {VendorID: 10415, AuthApplicationID: avp.Some(uint32(16777313))}}
	peer := &cfg.Peers[0]
	policy := peer.InboundCommands[0]
	peer.InboundCommands = nil
	peer.Commands = nil
	peer.MaxInbound = 16
	peer.MaxInboundBytes = 512 << 20
	for app, name := range duplicateProfileNames {
		class := ClassID("cache-" + name)
		if shared {
			class = "shared-cache"
		}
		if !shared || app == 0 {
			cfg.Resources.DuplicateClasses = append(cfg.Resources.DuplicateClasses, ClassConfig{ID: class, Bytes: 128 << 20, Slots: 8})
		}
		p := policy
		p.Command = mixedDuplicateBody(t, app).CommandKey()
		p.Class = "mixed-inbound"
		p.Sources = []pdu.Origin{gmscOrigin, duplicateOtherOrigin}
		p.ExecutionTimeout = 5 * time.Second
		p.Duplicate.Class = class
		p.Duplicate.MaxProcessing = p.ExecutionTimeout
		p.Duplicate.ReplayRetention = 30 * time.Second
		peer.InboundCommands = append(peer.InboundCommands, p)
	}
	return twoDuplicatePeers(cfg)
}
func mixedDuplicateRequest(t *testing.T, n *Node, co *codec.Codec, app int, hbh, end uint32, origin pdu.Origin) pdu.Message {
	t.Helper()
	m, e := co.CompleteRequest(mixedDuplicateBody(t, app), pdu.RequestEnvelope{OriginHost: origin.Host, OriginRealm: origin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: end})
	if e != nil {
		t.Fatal(e)
	}
	setMixedProxy(t, m.Body, hbh)
	return m
}
func setMixedProxy(t *testing.T, body pdu.Body, hbh uint32) {
	t.Helper()
	v := []pdu.ProxyInfo{{ProxyHost: fmt.Sprintf("hop-%d.example", hbh), ProxyState: []byte{byte(hbh), 0, 255}}}
	switch b := body.(type) {
	case *combined.CCR:
		b.ProxyInfo = v
	case *s6c.SRR:
		b.ProxyInfo = v
	case *sh.UDR:
		b.ProxyInfo = v
	case *sgd.OFR:
		b.ProxyInfo = v
	default:
		t.Fatal("unknown fixture")
	}
}
func mixedDuplicateArrival(t *testing.T, co *codec.Codec, req pdu.Message, hbh uint32) pdu.Message {
	t.Helper()
	m, e := co.Snapshot(req)
	if e != nil {
		t.Fatal(e)
	}
	m.Header.HopByHop = hbh
	m.Header.Flags ^= pdu.RetransmittedFlag
	setMixedProxy(t, m.Body, hbh)
	return m
}
func mixedDuplicateAnswer(req pdu.Message) pdu.Body {
	switch b := req.Body.(type) {
	case *combined.CCR:
		return combinedAnswer(b, 2001)
	case *s6c.SRR:
		return inboundFailure()
	case *sh.UDR:
		return shAnswer(306, 5012, false)
	default:
		return &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), SMRPUI: avp.Some([]byte{0, 255, 2})}
	}
}
func mixedDuplicateResult(t *testing.T, n *Node, co *codec.Codec, remote transport.Conn, req pdu.Message) {
	t.Helper()
	a := duplicateAnswer(t, n, co, remote, req)
	var got, want []pdu.ProxyInfo
	switch b := a.Body.(type) {
	case *combined.CCA:
		if b.ResultCode != 2001 {
			t.Fatal(b)
		}
		got = b.ProxyInfo
	case *s6c.SRA:
		if !b.ExperimentalResult.Present || b.ExperimentalResult.Value.ExperimentalResultCode != 5001 {
			t.Fatal(b)
		}
		got = b.ProxyInfo
	case *sh.UDA:
		if !b.ResultCode.Present || b.ResultCode.Value != 5012 {
			t.Fatal(b)
		}
		got = b.ProxyInfo
	case *sgd.OFA:
		if !b.ResultCode.Present || b.ResultCode.Value != 2001 {
			t.Fatal(b)
		}
		got = b.ProxyInfo
	default:
		t.Fatal("wrong profile result", a)
	}
	switch b := req.Body.(type) {
	case *combined.CCR:
		want = b.ProxyInfo
	case *s6c.SRR:
		want = b.ProxyInfo
	case *sh.UDR:
		want = b.ProxyInfo
	case *sgd.OFR:
		want = b.ProxyInfo
	}
	if len(got) != 1 || len(want) != 1 || got[0].ProxyHost != want[0].ProxyHost || !bytes.Equal(got[0].ProxyState, want[0].ProxyState) {
		t.Fatal("cross-arrival proxy leak")
	}
}

func TestMixedDuplicateIdentity(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for app, name := range duplicateProfileNames {
			t.Run(name, func(t *testing.T) {
				n, co := newInboundNode(t, mixedDuplicateConfig(t, true))
				p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
				p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
				req := mixedDuplicateRequest(t, n, co, app, 9921, 19921, gmscOrigin)
				combinedSend(t, co, a, req)
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				mark := markDuplicate(n, in)
				before := n.Resources()
				for phase := range 2 {
					for other := range duplicateProfileNames {
						if other == app {
							continue
						}
						conflict := mixedDuplicateRequest(t, n, co, other, uint32(9930+phase*10+other), req.Header.EndToEnd, gmscOrigin)
						combinedSend(t, co, b, conflict)
						readReject(t, b, conflict.Header, 5012)
					}
					noDuplicateDelivery(t, p2)
					assertDuplicateMark(t, n, mark, phase == 1)
					if phase == 0 {
						if _, e = in.Respond(ctxTest(t), mixedDuplicateAnswer(req)); e != nil {
							t.Fatal(e)
						}
						mixedDuplicateResult(t, n, co, a, req)
					}
				}
				if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes || s.DuplicateConflicts != 6 {
					t.Fatal(s)
				}
				dup := mixedDuplicateArrival(t, co, req, 9950)
				combinedSend(t, co, b, dup)
				mixedDuplicateResult(t, n, co, b, dup)
				noDuplicateDelivery(t, p2)
				// Different authorized origin, same E2E, creates independent logical work.
				independent := mixedDuplicateRequest(t, n, co, app, 9951, req.Header.EndToEnd, duplicateOtherOrigin)
				combinedSend(t, co, b, independent)
				next, e := p2.Receive(ctxTest(t))
				if e != nil || next.Info().RequestID == in.Info().RequestID {
					t.Fatal("independent origin collapsed", e)
				}
				if n.Resources().DuplicateRecords != 2 {
					t.Fatal(n.Resources())
				}
				if _, e = next.Respond(ctxTest(t), mixedDuplicateAnswer(independent)); e != nil {
					t.Fatal(e)
				}
				mixedDuplicateResult(t, n, co, b, independent)
				assertDuplicateMark(t, n, mark, true)
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestMixedDuplicateConcurrentArrivals(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, shared := range []bool{false, true} {
			for held, name := range duplicateProfileNames {
				t.Run(fmt.Sprintf("shared-%t/held-%s", shared, name), func(t *testing.T) {
					n, co := newInboundNode(t, mixedDuplicateConfig(t, shared))
					p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
					p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
					requests := make([]pdu.Message, 4)
					arrivals := make([]pdu.Message, 4)
					work := make([]*Inbound, 4)
					marks := make([]duplicateMark, 4)
					for app := range duplicateProfileNames {
						requests[app] = mixedDuplicateRequest(t, n, co, app, uint32(10001+app), uint32(20001+app), gmscOrigin)
						combinedSend(t, co, a, requests[app])
						var e error
						work[app], e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
						marks[app] = markDuplicate(n, work[app])
					}
					before := n.Resources()
					if before.InboundOutstanding != 4 || before.DuplicateRecords != 4 {
						t.Fatal(before)
					}
					for app := range requests {
						arrivals[app] = mixedDuplicateArrival(t, co, requests[app], uint32(10011+app))
						combinedSend(t, co, b, arrivals[app])
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 8 })
					noDuplicateDelivery(t, p)
					noDuplicateDelivery(t, p2)
					if s := n.Resources(); s.DuplicateRecords != 4 || s.DuplicateBytes != before.DuplicateBytes {
						t.Fatal("processing duplicates double charged", s)
					}
					for app := range requests {
						if app == held {
							continue
						}
						if _, e := work[app].Respond(ctxTest(t), mixedDuplicateAnswer(requests[app])); e != nil {
							t.Fatal(e)
						}
						mixedDuplicateResult(t, n, co, a, requests[app])
						mixedDuplicateResult(t, n, co, b, arrivals[app])
						assertDuplicateMark(t, n, marks[app], true)
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
					assertDuplicateMark(t, n, marks[held], false)
					// Three committed profiles replay while the fourth logical request still
					// owns both processing arrivals. A stalled application cannot replace them.
					for app := range requests {
						if app == held {
							continue
						}
						hit := mixedDuplicateArrival(t, co, requests[app], uint32(10021+app))
						combinedSend(t, co, b, hit)
						mixedDuplicateResult(t, n, co, b, hit)
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
					assertDuplicateMark(t, n, marks[held], false)
					if _, e := work[held].Respond(ctxTest(t), mixedDuplicateAnswer(requests[held])); e != nil {
						t.Fatal(e)
					}
					mixedDuplicateResult(t, n, co, a, requests[held])
					mixedDuplicateResult(t, n, co, b, arrivals[held])
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
					for app := range requests {
						assertDuplicateMark(t, n, marks[app], true)
					}
					if s := n.Resources(); s.DuplicateRecords != 4 || s.DuplicateBytes != before.DuplicateBytes || s.DuplicateConflicts != 0 {
						t.Fatal(s)
					}
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}

func TestMixedDuplicateCacheBudgets(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, shared := range []bool{false, true} {
			for _, bytesLimit := range []bool{false, true} {
				t.Run(fmt.Sprintf("shared-%t/bytes-%t", shared, bytesLimit), func(t *testing.T) {
					cfg := mixedDuplicateConfig(t, shared)
					// Obtain compiled per-profile reservation sizes before selecting exact byte
					// boundaries; no private account is modified after Node construction.
					probe, e := NewNode(cfg)
					if e != nil {
						t.Fatal(e)
					}
					bounds := make([]uint64, 4)
					var total uint64
					for app := range duplicateProfileNames {
						bounds[app] = probe.peers["dra"].inboundPolicies[mixedDuplicateBody(t, app).CommandKey()].cacheBound
						total += bounds[app]
					}
					closeDuplicateNode(t, probe)
					for i := range cfg.Resources.DuplicateClasses {
						c := &cfg.Resources.DuplicateClasses[i]
						if bytesLimit {
							c.Bytes = bounds[i]
							if shared {
								c.Bytes = total
							}
						} else {
							c.Slots = 1
							if shared {
								c.Slots = 4
							}
						}
					}
					n, co := newInboundNode(t, cfg)
					p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
					p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
					reqs := make([]pdu.Message, 4)
					marks := make([]duplicateMark, 4)
					var used uint64
					admit := func(app int) {
						req := mixedDuplicateRequest(t, n, co, app, uint32(9960+app), uint32(19960+app), gmscOrigin)
						reqs[app] = req
						combinedSend(t, co, a, req)
						in, e := p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
						marks[app] = markDuplicate(n, in)
						used += bounds[app]
						if _, e = in.Respond(ctxTest(t), mixedDuplicateAnswer(req)); e != nil {
							t.Fatal(e)
						}
						mixedDuplicateResult(t, n, co, a, req)
						waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
						if n.Resources().DuplicateBytes != used {
							t.Fatal("wrong reservation sum", n.Resources(), used)
						}
					}
					admit(0)
					if shared {
						for app := 1; app < 4; app++ {
							admit(app)
						}
					}
					// Replays still fit when no new cache record can be admitted.
					hit := mixedDuplicateArrival(t, co, reqs[0], 9970)
					combinedSend(t, co, b, hit)
					mixedDuplicateResult(t, n, co, b, hit)
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
					fresh := mixedDuplicateRequest(t, n, co, 0, 9971, 29971, gmscOrigin)
					combinedSend(t, co, b, fresh)
					waitFor(t, func() bool { return !p2.Info().Ready })
					if !errors.Is(p2.Info().Cause, ErrBackpressure) {
						t.Fatal(p2.Info())
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
					if n.Resources().DuplicateBytes != used {
						t.Fatal("refusal evicted result", n.Resources())
					}
					noDuplicateDelivery(t, p2)
					// Other classes admit independent work after the first class is full.
					if !shared {
						for app := 1; app < 4; app++ {
							admit(app)
						}
					}
					if s := n.Resources(); s.DuplicateRecords != 4 || s.DuplicateBytes != total {
						t.Fatal(s, total)
					}
					for app := range reqs {
						hit = mixedDuplicateArrival(t, co, reqs[app], uint32(9980+app))
						combinedSend(t, co, a, hit)
						mixedDuplicateResult(t, n, co, a, hit)
						assertDuplicateMark(t, n, marks[app], true)
					}
					noDuplicateDelivery(t, p)
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
					latest := marks[0].until
					for _, m := range marks {
						if m.until.After(latest) {
							latest = m.until
						}
					}
					n.tick(latest)
					waitFor(t, func() bool { return n.Resources().DuplicateRecords == 0 && n.Resources().DuplicateBytes == 0 })
					// Explicit timer expiry, not a wall-clock soak: retained identity is new work.
					again := mixedDuplicateArrival(t, co, reqs[0], 9990)
					combinedSend(t, co, a, again)
					next, e := p.Receive(ctxTest(t))
					if e != nil {
						t.Fatal(e)
					}
					if markDuplicate(n, next).record == marks[0].record {
						t.Fatal("expired record reused")
					}
					if _, e = next.Respond(ctxTest(t), mixedDuplicateAnswer(again)); e != nil {
						t.Fatal(e)
					}
					mixedDuplicateResult(t, n, co, a, again)
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}
