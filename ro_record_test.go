// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func recordConfig(t *testing.T, family string) NodeConfig {
	c := roRecoveryConfig(t, family)
	c.RoRecovery = RoRecoveryLimits{MaxRecordBytes: 65536, MaxTargets: 4, MaxTotalAttempts: 8, MaxSegments: 4, MaxAttemptsPerSegment: 4, MaxReplayAge: time.Hour, MaxResponseWait: time.Minute}
	c.Resources.RoRecoverySlots = 4
	c.Resources.RoRecoveryClasses = []IdentifierClass{{"sms", 4}}
	c.Resources.IdentifierBytes += 4*512 + 256
	return c
}
func recordOptions() RoRecoveryOptions {
	return RoRecoveryOptions{PolicyID: "sms-ro-recovery-v1", Targets: []RoRecoveryTarget{{ID: "ocs-a", Peer: "dra", Retention: time.Hour}}, Limits: RoRecoveryLimits{MaxRecordBytes: 65536, MaxTargets: 4, MaxTotalAttempts: 8, MaxSegments: 4, MaxAttemptsPerSegment: 4, MaxReplayAge: time.Minute, MaxResponseWait: time.Second}}
}
func prepareRecord(t *testing.T, n *Node, body pdu.Body, opts RoRecoveryOptions) RoRecoveryRequest {
	t.Helper()
	r, e := n.PrepareRoRecoveryRequest(ctxTest(t), body, opts)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.CheckIntegrity(); e != nil {
		t.Fatal(e)
	}
	return r
}
func copyRecord(t *testing.T, r RoRecoveryRequest) RoRecoveryRequest {
	t.Helper()
	raw, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	var out RoRecoveryRequest
	if e = json.Unmarshal(raw, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func assertRecordIdle(t *testing.T, n *Node, held int) {
	t.Helper()
	info := n.RoRecoveryInfo()
	if info.Held != held || info.Reserved != 0 || n.Resources().OutboundBytes != 0 {
		t.Fatalf("info=%+v resources=%+v", info, n.Resources())
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.seq != 0 || len(n.requests) != 0 || len(n.ports) != 0 {
		t.Fatal("preparation allocated a request/port", n.seq, len(n.requests), len(n.ports))
	}
	for _, p := range n.peers {
		if p.account.UsedSlots != 0 || p.preparing != 0 || p.conn != nil {
			t.Fatal("preparation selected a peer connection")
		}
	}
}
func TestRoRecordAllProfilesAndOwnedData(t *testing.T) {
	for _, family := range []string{"standard", "sms", "money", "combined"} {
		kinds := []string{"iec", "initial", "term"}
		if family == "money" {
			kinds = []string{"general", "balance", "refund"}
		}
		if family == "combined" {
			kinds = []string{"iec", "initial", "term", "general", "balance", "refund"}
		}
		for _, kind := range kinds {
			t.Run(family+"/"+kind, func(t *testing.T) {
				n, _ := newInboundNode(t, recordConfig(t, family))
				body := roRecoveryBody(t, family, kind)
				opts := recordOptions()
				r := prepareRecord(t, n, body, opts)
				assertRecordIdle(t, n, 1)
				h, _, e := wire.Header(r.Frame, n.cfg.Limits)
				if e != nil || h.HopByHop != 0 || h.Flags != pdu.RequestFlag|pdu.ProxiableFlag {
					t.Fatal(h, e)
				}
				expected, e := n.engine.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, DestinationHost: serverOrigin.Host, DestinationRealm: serverOrigin.Realm, EndToEnd: h.EndToEnd})
				if e != nil {
					t.Fatal(e)
				}
				encoded, e := n.engine.Append(nil, expected)
				if e != nil || !bytes.Equal(encoded, r.Frame) {
					t.Fatal("original content changed", e)
				}
				roundtrip := copyRecord(t, r)
				if e = roundtrip.CheckIntegrity(); e != nil {
					t.Fatal("application JSON roundtrip", e)
				}
				opts.Targets[0].ID = "changed"
				r.Targets[0].Policy.Sources.Terminals[0].Host = "changed.example"
				r.Frame[len(r.Frame)-1] ^= 1
				if roundtrip.CheckIntegrity() != nil || r.CheckIntegrity() == nil {
					t.Fatal("mutable copies/checksum")
				}
				guard := n.roRecovery.slots[n.roRecovery.byE2E[h.EndToEnd]]
				if guard.digest != roundtrip.Digest || guard.recordID != roundtrip.RecordID {
					t.Fatal("guard retained application alias")
				}
				if n.peers["dra"].policies[h.Key()].Sources.Terminals[0].Host != "hss.example" && !sameOrigin(n.peers["dra"].policies[h.Key()].Sources.Terminals[0], serverOrigin) {
					t.Fatal("policy mutated")
				}
				if info := n.IdentifierInfo(); info.ApplicationActive != 0 || info.Active != 0 || info.Reserved != 0 {
					t.Fatal(info)
				}
			})
		}
	}
}

func TestRoRecordTargetsBoundsAndHopState(t *testing.T) {
	for _, family := range []string{"standard", "sms", "money", "combined"} {
		t.Run(family, func(t *testing.T) {
			c := recordConfig(t, family)
			target := pdu.Origin{Host: strings.Repeat("a", 50) + "." + strings.Repeat("b", 50) + ".example", Realm: "backup.example"}
			second := &c.Peers[1].Commands[0]
			second.DestinationHost = target.Host
			second.DestinationRealm = target.Realm
			second.Sources.Terminals = []pdu.Origin{target}
			n, _ := newInboundNode(t, c)
			o := recordOptions()
			o.Targets = append(o.Targets, RoRecoveryTarget{ID: "ocs-b", Peer: "dra2", Retention: time.Hour})
			kind := "iec"
			if family == "money" {
				kind = "refund"
			}
			r := prepareRecord(t, n, roRecoveryBody(t, family, kind), o)
			frame, e := roTargetFrame(r.Frame, target, n.cfg.Limits, r.MaxFrameBytes)
			if e != nil || len(frame) != r.MaxFrameBytes || r.MaxFrameBytes <= len(r.Frame) {
				t.Fatal(len(frame), r.MaxFrameBytes, e)
			}
			msg, e := n.engine.Decode(frame)
			if e != nil {
				t.Fatal(e)
			}
			op, _ := n.cfg.Bundle.Lookup(msg.Header.Key())
			env, e := op.Envelope(msg.Body)
			if e != nil || env.DestinationHost.Value != target.Host || env.DestinationRealm.Value != target.Realm || binary.BigEndian.Uint32(frame[16:20]) != binary.BigEndian.Uint32(r.Frame[16:20]) {
				t.Fatal(env, e)
			}
			back, e := roTargetFrame(frame, serverOrigin, n.cfg.Limits, len(r.Frame))
			if e != nil || !bytes.Equal(back, r.Frame) {
				t.Fatal("retarget rewrote non-destination bytes", e)
			}
			before := n.IdentifierInfo().ApplicationAllocated
			tooSmall := o
			tooSmall.Limits.MaxRecordBytes = roRecordMetadata(r) + len(r.Frame)
			if _, e = n.PrepareRoRecoveryRequest(ctxTest(t), roRecoveryBody(t, family, kind), tooSmall); !errors.Is(e, wire.ErrLimit) {
				t.Fatal("longest target ignored", e)
			}
			if n.IdentifierInfo().ApplicationAllocated != before {
				t.Fatal("oversize target consumed E2E")
			}
			assertRecordIdle(t, n, 1)
		})
	}
	for _, hop := range []string{"proxy", "route"} {
		t.Run(hop, func(t *testing.T) {
			n, _ := newInboundNode(t, recordConfig(t, "standard"))
			b := roBody()
			if hop == "proxy" {
				b.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{1}}}
			} else {
				b.RouteRecord = []string{"proxy.example"}
			}
			if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), b, recordOptions()); !errors.Is(e, validation.ErrEnvelope) {
				t.Fatal(e)
			}
			if n.IdentifierInfo().ApplicationAllocated != 0 {
				t.Fatal("hop state consumed E2E")
			}
			assertRecordIdle(t, n, 0)
		})
	}
}

func TestRoRecordConfigurationAndAdmissionDenials(t *testing.T) {
	for _, change := range []func(*NodeConfig){
		func(c *NodeConfig) { c.RoRecovery = RoRecoveryLimits{} }, func(c *NodeConfig) { c.RoRecovery.MaxTargets = 17 }, func(c *NodeConfig) { c.RoRecovery.MaxRecordBytes = 4<<20 + 1 },
		func(c *NodeConfig) { c.RoRecovery.MaxTotalAttempts = 257 }, func(c *NodeConfig) { c.RoRecovery.MaxSegments = 9 }, func(c *NodeConfig) { c.RoRecovery.MaxAttemptsPerSegment = 17 }, func(c *NodeConfig) { c.RoRecovery.MaxResponseWait = 2 * time.Hour },
		func(c *NodeConfig) { c.Resources.IdentifierBytes-- }, func(c *NodeConfig) { c.Resources.RoRecoverySlots = 0 }, func(c *NodeConfig) { c.Resources.RoRecoverySlots = 4097 }, func(c *NodeConfig) { c.Resources.RoRecoveryClasses[0].ID = "absent" }, func(c *NodeConfig) { c.Resources.RoRecoveryClasses[0].Slots = 5 },
	} {
		c := recordConfig(t, "standard")
		change(&c)
		n, e := NewNode(c)
		if n != nil {
			n.Close()
		}
		if !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
	}
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	for _, change := range []func(*RoRecoveryOptions){
		func(o *RoRecoveryOptions) { o.PolicyID = "" }, func(o *RoRecoveryOptions) { o.PolicyID = strings.Repeat("x", 129) }, func(o *RoRecoveryOptions) { o.Targets = nil }, func(o *RoRecoveryOptions) { o.Targets[0].Peer = "missing" }, func(o *RoRecoveryOptions) { o.Targets[0].Retention = time.Second },
		func(o *RoRecoveryOptions) {
			o.Targets = append(o.Targets, RoRecoveryTarget{ID: "other", Peer: "dra2", Retention: time.Hour})
		}, func(o *RoRecoveryOptions) { o.Limits.MaxReplayAge = 2 * time.Hour }, func(o *RoRecoveryOptions) { o.Limits.MaxRecordBytes = 65537 }, func(o *RoRecoveryOptions) { o.Limits.MaxSegments = 0 },
	} {
		o := recordOptions()
		change(&o)
		if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), o); !errors.Is(e, ErrConfig) {
			t.Fatal(e)
		}
	}
	b := roBody()
	b.RequestedAction = avp.Some(uint32(3))
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), b, recordOptions()); e == nil {
		t.Fatal("unknown action")
	}
	b = roBody()
	b.OriginHost = "foreign.example"
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), b, recordOptions()); !errors.Is(e, validation.ErrEnvelope) {
		t.Fatal(e)
	}
	b = roBody()
	b.DestinationHost = avp.Some("foreign.example")
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), b, recordOptions()); !errors.Is(e, validation.ErrEnvelope) {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := n.PrepareRoRecoveryRequest(ctx, roBody(), recordOptions()); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if n.IdentifierInfo().ApplicationAllocated != 0 {
		t.Fatal("invalid input consumed identifier")
	}
	assertRecordIdle(t, n, 0)
}

func TestRoRecordIntegrityRejectsMutation(t *testing.T) {
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	r := prepareRecord(t, n, roBody(), recordOptions())
	changes := []func(*RoRecoveryRequest){
		func(r *RoRecoveryRequest) { r.Version++ }, func(r *RoRecoveryRequest) { r.RecordID[0] ^= 1 }, func(r *RoRecoveryRequest) { r.WireFingerprint = "" }, func(r *RoRecoveryRequest) { r.ContextName += "changed" }, func(r *RoRecoveryRequest) { r.PolicyID += "changed" },
		func(r *RoRecoveryRequest) { r.CreatedAt = r.CreatedAt.Add(time.Second) }, func(r *RoRecoveryRequest) { r.ReplayUntil = r.ReplayUntil.Add(time.Second) }, func(r *RoRecoveryRequest) { r.Limits.MaxTotalAttempts++ }, func(r *RoRecoveryRequest) { r.Limits.MaxSegments++ }, func(r *RoRecoveryRequest) { r.Limits.MaxReplayAge += time.Second }, func(r *RoRecoveryRequest) { r.Limits.MaxResponseWait += time.Second },
		func(r *RoRecoveryRequest) { r.Targets[0].TargetID += "changed" }, func(r *RoRecoveryRequest) { r.Targets[0].Retention += time.Second }, func(r *RoRecoveryRequest) { r.Targets[0].Policy.Class = "other" }, func(r *RoRecoveryRequest) { r.Targets[0].Policy.Timeout += time.Second }, func(r *RoRecoveryRequest) { r.Targets[0].Policy.MaxAnswerBytes += 4 }, func(r *RoRecoveryRequest) {
			r.Targets[0].Policy.Sources.ErrorAgents = append(r.Targets[0].Policy.Sources.ErrorAgents, pdu.Origin{Host: "other.example", Realm: "example"})
		},
		func(r *RoRecoveryRequest) { r.Frame[12] = 1 }, func(r *RoRecoveryRequest) { r.Frame[16] ^= 1 }, func(r *RoRecoveryRequest) { r.Frame[4] |= pdu.RetransmittedFlag }, func(r *RoRecoveryRequest) { r.Frame[len(r.Frame)-1] ^= 1 }, func(r *RoRecoveryRequest) { r.MaxFrameBytes += 4 },
	}
	for i, change := range changes {
		v := copyRecord(t, r)
		change(&v)
		if e := v.CheckIntegrity(); e == nil {
			t.Fatal("mutation accepted", i)
		}
	}
	// A checksum is deliberately not a MAC. Even a recalculated checksum cannot
	// change the original identity/digest retained by the current Node's guard.
	v := copyRecord(t, r)
	v.PolicyID = "different-policy"
	v.Digest = roRecordDigest(v)
	if v.CheckIntegrity() != nil {
		t.Fatal("checksum presented as authentication")
	}
	slot := n.roRecovery.byE2E[binary.BigEndian.Uint32(r.Frame[16:20])]
	if n.roRecovery.slots[slot].digest == v.Digest {
		t.Fatal("guard mutated")
	}
}

func TestRoRecordGuardQuotasAndExpiry(t *testing.T) {
	c := recordConfig(t, "standard")
	c.Resources.Classes = append(c.Resources.Classes, ClassConfig{ID: "other", Bytes: 1 << 20, Slots: 2})
	c.Resources.NodeBytes += 1 << 20
	c.Resources.RoRecoveryClasses = []IdentifierClass{{"sms", 1}, {"other", 1}}
	c.Resources.IdentifierBytes += 256
	n, _ := newInboundNode(t, c)
	r := prepareRecord(t, n, roBody(), recordOptions())
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrIdentifiers) {
		t.Fatal("class quota", e)
	}
	info := n.RoRecoveryInfo()
	info.Classes[0].Slots = 100
	if n.RoRecoveryInfo().Classes[0].Slots != 1 {
		t.Fatal("info alias")
	}
	n.mu.Lock()
	slot := n.roRecovery.byE2E[binary.BigEndian.Uint32(r.Frame[16:20])]
	until := n.roRecovery.slots[slot].keepUntil
	n.roRecovery.expire(until.Add(-time.Nanosecond))
	if !n.roRecovery.slots[slot].occupied {
		t.Fatal("early eviction")
	}
	n.roRecovery.expire(until)
	if n.roRecovery.slots[slot].occupied {
		t.Fatal("no exact expiry")
	}
	n.mu.Unlock()
	prepareRecord(t, n, roBody(), recordOptions())
	assertRecordIdle(t, n, 1)
}

func TestRoRecordDefaultIdentifierRolloverAndCollision(t *testing.T) {
	c := recordConfig(t, "standard")
	c.Identifiers.Space = SequenceSpace{Prefix: 9, CounterBits: 2}
	n, _ := newInboundNode(t, c)
	r := prepareRecord(t, n, roBody(), recordOptions())
	id := binary.BigEndian.Uint32(r.Frame[16:20])
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.identifiers.appActive != 0 {
		t.Fatal("prepared record is not an active exchange")
	}
	other, e := n.nextE2E(false)
	if e != nil || other == id {
		t.Fatal(other, e)
	}
	n.retireIdentifier(&identifierLease{slot: -1})
	// Simulate a completed round and reuse guard elapsed while ReplayUntil remains.
	n.identifiers.appRetired = time.Now().Add(-2 * n.identifiers.config.ReuseGuard)
	if _, e = n.nextE2E(false); !errors.Is(e, ErrIdentifierCollision) {
		t.Fatal("guard did not protect rollover", e)
	}
	if n.appCount != 1 || n.identifiers.appActive != 0 {
		t.Fatal("collision scanned or acquired lease", n.appCount, n.identifiers.appActive)
	}
	if v, e := n.nextE2E(true); e != nil || v == id {
		t.Fatal("recovery collided with controls", v, e)
	} else {
		n.retireIdentifier(&identifierLease{slot: -1, control: true})
	}
}

func TestRoRecordCustomIdentifierCollisionsAndRollback(t *testing.T) {
	var candidate atomic.Uint32
	candidate.Store(17)
	c := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) { return candidate.Load(), nil }), 4)
	c.Resources.IdentifierBytes += 4*512 + 256
	n, _ := newInboundNode(t, c)
	prepareRecord(t, n, roBody(), recordOptions())
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrIdentifierCollision) {
		t.Fatal(e)
	}
	assertRecordIdle(t, n, 1)
	if info := n.IdentifierInfo(); info.Active != 0 || info.Reserved != 0 || info.Quarantine != 0 {
		t.Fatal("duplicate guard charge", info)
	}
	n.mu.Lock()
	l, e := n.identifiers.reserve("sms", time.Now())
	if e != nil {
		t.Fatal(e)
	}
	n.identifiers.sourceBusy = false
	if e = n.installApplicationIdentifier(l, 17); !errors.Is(e, ErrIdentifierCollision) {
		t.Fatal("ordinary request bypassed recovery guard", e)
	}
	n.abandonIdentifier(l)
	l, e = n.identifiers.reserve("sms", time.Now())
	if e != nil {
		t.Fatal(e)
	}
	n.identifiers.sourceBusy = false
	if e = n.installApplicationIdentifier(l, 18); e != nil {
		t.Fatal(e)
	}
	n.retireIdentifier(l)
	n.mu.Unlock()
	candidate.Store(18)
	if _, e = n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrIdentifierCollision) {
		t.Fatal("ordinary quarantine bypass", e)
	}
	candidate.Store(0x80000000)
	if _, e = n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrIdentifierRange) {
		t.Fatal(e)
	}
	candidate.Store(19)
	prepareRecord(t, n, roBody(), recordOptions())
	assertRecordIdle(t, n, 2)
}

func TestRoRecordCancellationCloseAndSourceBusy(t *testing.T) {
	for _, stop := range []string{"cancel", "close", "source-error"} {
		t.Run(stop, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			c := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) {
				close(entered)
				<-release
				if stop == "source-error" {
					return 0, fmt.Errorf("source failed")
				}
				return 42, nil
			}), 4)
			c.Resources.IdentifierBytes += 4*512 + 256
			n, _ := newInboundNode(t, c)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, e := n.PrepareRoRecoveryRequest(ctx, roBody(), recordOptions()); done <- e }()
			<-entered
			info := n.RoRecoveryInfo()
			if info.Reserved != 1 || info.Held != 0 || n.Resources().OutboundBytes == 0 {
				t.Fatal(info)
			}
			if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrIdentifiers) {
				t.Fatal("concurrent source", e)
			}
			switch stop {
			case "cancel":
				cancel()
			case "close":
				n.Close()
				short, finish := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer finish()
				if e := n.Wait(short); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("Wait passed live preparer", e)
				}
			}
			close(release)
			e := <-done
			want := context.Canceled
			if stop == "close" {
				want = ErrClosed
			}
			if stop == "source-error" {
				want = ErrIdentifierSourceFailed
			}
			if !errors.Is(e, want) {
				t.Fatal(e)
			}
			if stop == "close" {
				if e = n.Wait(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
			}
			assertRecordIdle(t, n, 0)
		})
	}
}

func TestRoRecordConcurrentPreparationAndBudget(t *testing.T) {
	n, _ := newInboundNode(t, recordConfig(t, "standard"))
	var wg sync.WaitGroup
	out := make(chan RoRecoveryRequest, 20)
	errs := make(chan error, 20)
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := roBody()
			body.SessionID = fmt.Sprintf("ctf.example;%d", i)
			r, e := n.PrepareRoRecoveryRequest(ctxTest(t), body, recordOptions())
			if e != nil {
				errs <- e
			} else {
				out <- r
			}
		}()
	}
	wg.Wait()
	close(out)
	close(errs)
	count := 0
	ids := map[uint32]bool{}
	for r := range out {
		id := binary.BigEndian.Uint32(r.Frame[16:20])
		if ids[id] {
			t.Fatal("duplicate ID")
		}
		ids[id] = true
		count++
	}
	if count != 4 {
		t.Fatal("guard capacity", count)
	}
	for e := range errs {
		if !errors.Is(e, ErrIdentifiers) && !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
	}
	assertRecordIdle(t, n, 4)
	c := recordConfig(t, "standard")
	c.Resources.Classes[0].Bytes = 1
	limited, _ := newInboundNode(t, c)
	if _, e := limited.PrepareRoRecoveryRequest(ctxTest(t), roBody(), recordOptions()); !errors.Is(e, ErrBackpressure) {
		t.Fatal(e)
	}
	assertRecordIdle(t, limited, 0)
	if limited.IdentifierInfo().ApplicationAllocated != 0 {
		t.Fatal("budget rejection consumed ID")
	}
}

func TestRoRecordOrdinarySubmitCannotReusePreparedIdentity(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		var candidate atomic.Uint32
		candidate.Store(77)
		c := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) { return candidate.Load(), nil }), 4)
		c.Resources.IdentifierBytes += 4*512 + 256
		n, co := newInboundNode(t, c)
		p, remote := roRecoveryNetwork(t, n, co, "dra", encrypted)
		record := prepareRecord(t, n, roBody(), recordOptions())
		port := portTest(t, n, 1)
		if _, e := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port); !errors.Is(e, ErrIdentifierCollision) {
			t.Fatal("prepared ID reused by ordinary Submit", e)
		}
		if n.Resources().Pending != 0 || n.Resources().Outstanding != 0 {
			t.Fatal(n.Resources())
		}
		candidate.Store(78)
		id, e := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if req.Header.EndToEnd != 78 || req.Header.EndToEnd == binary.BigEndian.Uint32(record.Frame[16:20]) {
			t.Fatal("Prepare wrote network or duplicate E2E", req.Header)
		}
		roRecoveryReply(t, co, remote, req, 2001)
		got := routeCompletion(t, n, port, id)
		if got.Outcome != AnswerReceived {
			t.Fatal(got)
		}
		if info := n.RoRecoveryInfo(); info.Held != 1 || info.Reserved != 0 {
			t.Fatal(info)
		}
	})
}

func TestRoRecordFrozenBeforeCustomSourceAndStatePreservation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	c := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) { close(entered); <-release; return 91, nil }), 4)
	c.Resources.IdentifierBytes += 4*512 + 256
	n, _ := newInboundNode(t, c)
	b := roBody()
	b.OriginStateID = avp.Some(uint32(123))
	opts := recordOptions()
	out := make(chan RoRecoveryRequest, 1)
	errs := make(chan error, 1)
	go func() { r, e := n.PrepareRoRecoveryRequest(ctxTest(t), b, opts); out <- r; errs <- e }()
	<-entered
	b.SessionID = "changed"
	b.OriginStateID = avp.Some(uint32(999))
	b.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 99
	opts.Targets[0].ID = "changed"
	close(release)
	r := <-out
	if e := <-errs; e != nil {
		t.Fatal(e)
	}
	msg, e := n.engine.Decode(r.Frame)
	if e != nil {
		t.Fatal(e)
	}
	envOp, _ := n.cfg.Bundle.Lookup(msg.Header.Key())
	env, e := envOp.Envelope(msg.Body)
	if e != nil || env.SessionID.Value == "changed" || r.Targets[0].TargetID != "ocs-a" {
		t.Fatal(env, e)
	}
	// Compare with the full originally built request, including Origin-State and
	// deep unit/service content; no envelope field is silently refreshed.
	original := roBody()
	original.OriginStateID = avp.Some(uint32(123))
	want, e := n.engine.CompleteRequest(original, pdu.RequestEnvelope{OriginHost: n.cfg.Origin.Host, OriginRealm: n.cfg.Origin.Realm, DestinationHost: serverOrigin.Host, DestinationRealm: serverOrigin.Realm, EndToEnd: 91})
	if e != nil {
		t.Fatal(e)
	}
	encoded, e := n.engine.Append(nil, want)
	if e != nil || !bytes.Equal(encoded, r.Frame) {
		t.Fatal("input alias/state rewrite", e)
	}
	assertRecordIdle(t, n, 1)
}

func TestRoRecordPreflightDoesNotCallIdentifierSource(t *testing.T) {
	var calls atomic.Int32
	c := customIdentifiers(recordConfig(t, "standard"), idSource(func() (uint32, error) { calls.Add(1); return 31, nil }), 4)
	c.Resources.IdentifierBytes += 4*512 + 256
	target := pdu.Origin{Host: "second.example", Realm: "example"}
	c.Peers[1].Commands[0].DestinationHost = target.Host
	c.Peers[1].Commands[0].DestinationRealm = target.Realm
	c.Peers[1].Commands[0].Sources.Terminals = []pdu.Origin{target}
	c.Peers[1].Commands[0].MaxRequestBytes = 20
	n, _ := newInboundNode(t, c)
	o := recordOptions()
	o.Targets = append(o.Targets, RoRecoveryTarget{ID: "second", Peer: "dra2", Retention: time.Hour})
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), o); !errors.Is(e, wire.ErrLimit) {
		t.Fatal(e)
	}
	o = recordOptions()
	o.Limits.MaxReplayAge = time.Nanosecond
	o.Limits.MaxResponseWait = time.Nanosecond
	if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), roBody(), o); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	if calls.Load() != 0 {
		t.Fatal("preflight failure called source", calls.Load())
	}
	assertRecordIdle(t, n, 0)
}
