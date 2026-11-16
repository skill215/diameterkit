// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6 (Codex)
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded mixed-application acceptance tests generated with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

var mixedNames = [4]string{"ro", "s6c", "sh", "sgd"}
var mixedCommands = [4]pdu.CommandKey{(&ro.CCR{}).CommandKey(), (&s6c.SRR{}).CommandKey(), (&sh.UDR{}).CommandKey(), (&sgd.OFR{}).CommandKey()}

type mixedMetrics struct {
	pressureMetrics
	Stalled                                                    string
	ClassWindow, HealthyCompleted, ClassRefusals, PortRefusals int
}

func mixedPair(t *testing.T, ctx context.Context, window int) (*pair, [5]*diameter.CompletionPort) {
	t.Helper()
	var profiles []schema.Profile
	for _, factory := range []func() (schema.Profile, error){base.Profile, ro.Profile, s6c.Profile, sh.Profile, sgd.Profile} {
		p, err := factory()
		if err != nil {
			t.Fatal(err)
		}
		profiles = append(profiles, p)
	}
	bundle, err := schema.Compile(profiles...)
	if err != nil {
		t.Fatal(err)
	}
	origins := [2]pdu.Origin{{Host: "client.example", Realm: "example"}, {Host: "server.example", Realm: "example"}}
	var cfg [2]diameter.NodeConfig
	for i := range cfg {
		c, err := endpointConfig(origins[i], origins[1-i], i == 0)
		if err != nil {
			t.Fatal(err)
		}
		c.Bundle = bundle
		c.Timeouts.Request = 10 * time.Second
		c.Capabilities.AuthApplicationID = []uint32{ro.ApplicationID}
		c.Capabilities.VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{
			sh.Application(),
			{VendorID: 10415, AuthApplicationID: avp.Some(mixedCommands[1].ApplicationID)},
			{VendorID: 10415, AuthApplicationID: avp.Some(mixedCommands[3].ApplicationID)},
		}
		c.Resources.NodeBytes = 1 << 30
		c.Resources.MaxPorts, c.Resources.MaxPortOutstanding = 5, window
		c.Resources.Classes, c.Resources.InboundClasses = nil, nil
		c.Peers[0].Commands, c.Peers[0].InboundCommands = nil, nil
		// One spare peer slot ensures a class refusal cannot be explained by
		// the shared peer limit. The spare port likewise isolates class admission.
		c.Peers[0].MaxPending, c.Peers[0].MaxBytes = 4*window+1, 256<<20
		if i == 1 {
			c.Peers[0].MaxInbound, c.Peers[0].MaxInboundBytes = 4*window+1, 256<<20
		}
		for j, key := range mixedCommands {
			class := diameter.ClassID(mixedNames[j])
			c.Resources.Classes = append(c.Resources.Classes, diameter.ClassConfig{ID: class, Slots: window, Bytes: 64 << 20})
			if i == 0 {
				c.Peers[0].Commands = append(c.Peers[0].Commands, diameter.CommandPolicy{
					Command: key, Class: class, DestinationHost: origins[1].Host, DestinationRealm: origins[1].Realm,
					Sources: pdu.AnswerSources{Terminals: []pdu.Origin{origins[1]}}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 10 * time.Second,
				})
			} else {
				// Four S6c inbound reservations require 72,400,896 bytes
				// with these decode limits; 64 MiB would test byte exhaustion
				// instead of the intended independent slot saturation.
				c.Resources.InboundClasses = append(c.Resources.InboundClasses, diameter.ClassConfig{ID: class, Slots: window, Bytes: 128 << 20})
				c.Peers[0].InboundCommands = append(c.Peers[0].InboundCommands, diameter.InboundPolicy{
					Command: key, Class: class, Sources: []pdu.Origin{origins[0]}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024, ExecutionTimeout: 10 * time.Second, NoCache: true,
				})
			}
		}
		cfg[i] = c
	}
	p, err := connectPair(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var ports [5]*diameter.CompletionPort
	t.Cleanup(func() {
		_ = p.left.node.Close()
		_ = p.right.node.Close()
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, port := range ports {
			if port != nil {
				if err := port.Close(cleanup); err != nil {
					t.Error(err)
				}
			}
		}
		if err := p.close(); err != nil {
			t.Error(err)
		}
		if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
			t.Error("mixed resources retained after Close/Wait")
		}
	})
	if err := p.left.port.Close(ctx); err != nil {
		t.Fatal(err)
	}
	for i := range ports {
		ports[i], err = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: window, MaxOutstandingBytes: 64 << 20})
		if err != nil {
			t.Fatal(err)
		}
	}
	return p, ports
}

func mixedRequest(t *testing.T, n *diameter.Node, app int) (pdu.Body, string) {
	t.Helper()
	if app == 0 {
		r, err := newDebit(n)
		if err != nil {
			t.Fatal(err)
		}
		return r, r.SessionID
	}
	session, err := n.NewSessionID()
	if err != nil {
		t.Fatal(err)
	}
	switch app {
	case 1:
		r, err := s6c.NewSRR(s6c.RoutingQuery{SessionID: session, MSISDN: avp.Some("12025550123"), SCAddress: "12025550100"})
		if err != nil {
			t.Fatal(err)
		}
		return r, session
	case 2:
		return &sh.UDR{SessionID: session, VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, UserIdentity: sh.UserIdentity{PublicIdentity: avp.Some("sip:alice@example.net")}, DataReference: 24}, session
	default:
		sc, err := avp.EncodeTBCD("12025550100")
		if err != nil {
			t.Fatal(err)
		}
		return &sgd.OFR{SessionID: session, AuthSessionState: 1, UserIdentifier: sgd.UserIdentifier{UserName: avp.Some("001010123456789")}, SCAddress: sc, SMRPUI: []byte{1, 0, 255, 4}}, session
	}
}

func mixedAnswer(t *testing.T, app int) pdu.Body {
	t.Helper()
	switch app {
	case 0:
		a, err := newAnswer("2001")
		if err != nil {
			t.Fatal(err)
		}
		return a
	case 1:
		return &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}
	case 2:
		return &sh.UDA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}
	default:
		return &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), SMRPUI: avp.Some([]byte{0, 255, 2})}
	}
}

type mixedExchange struct {
	app      int
	id       diameter.RequestID
	session  string
	in       *diameter.Inbound
	canceled bool
}

func TestBoundedMixedApplicationPressure(t *testing.T) {
	for _, window := range []int{1, 4} {
		for stalled, name := range mixedNames {
			t.Run(fmt.Sprintf("window%d/%s", window, name), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				p, ports := mixedPair(t, ctx, window)
				m := mixedMetrics{pressureMetrics: pressureMetrics{Window: 4 * window, Waves: 8}, Stalled: name, ClassWindow: window}
				seen := map[diameter.RequestID]bool{}
				submit := func(apps []int) []*mixedExchange {
					var work []*mixedExchange
					var bodies []pdu.Body
					for _, app := range apps {
						for range window {
							body, session := mixedRequest(t, p.left.node, app)
							work = append(work, &mixedExchange{app: app, session: session})
							bodies = append(bodies, body)
						}
					}
					errs := make([]error, len(work))
					var wg sync.WaitGroup
					for i, w := range work {
						wg.Add(1)
						go func() {
							defer wg.Done()
							w.id, errs[i] = p.left.peer.Submit(ctx, bodies[i], diameter.RequestOptions{}, ports[w.app])
						}()
					}
					wg.Wait()
					bySession := map[string]*mixedExchange{}
					for i, w := range work {
						if errs[i] != nil {
							t.Fatal(errs[i])
						}
						bySession[w.session] = w
					}
					for received := range work {
						in, err := p.right.peer.Receive(ctx)
						if err != nil {
							t.Fatalf("receive %d/%d after %d requests: %v; client=%+v server=%+v; client resources=%+v server resources=%+v", received, len(work), m.Requests, err, p.left.peer.Info(), p.right.peer.Info(), p.left.node.Resources(), p.right.node.Resources())
						}
						var session string
						switch b := in.Message().Body.(type) {
						case *ro.CCR:
							session = b.SessionID
						case *s6c.SRR:
							session = b.SessionID
						case *sh.UDR:
							session = b.SessionID
						case *sgd.OFR:
							session = b.SessionID
						}
						w := bySession[session]
						if w == nil || w.in != nil || in.Message().Body.CommandKey() != mixedCommands[w.app] || !in.Validation().Complete {
							t.Fatal("unexpected/duplicate/incomplete mixed inbound")
						}
						w.in = in
					}
					m.Requests += len(work)
					m.sample(t, p)
					return work
				}
				respond := func(work []*mixedExchange) {
					for _, w := range work {
						r, err := w.in.Respond(ctx, mixedAnswer(t, w.app))
						if err != nil || !r.Committed {
							t.Fatal("mixed response", err, r)
						}
					}
				}
				consume := func(work []*mixedExchange, healthy bool) {
					wanted := map[diameter.RequestID]*mixedExchange{}
					for _, w := range work {
						wanted[w.id] = w
					}
					for _, slot := range work {
						c, err := ports[slot.app].Receive(ctx)
						if err != nil {
							t.Fatal(err)
						}
						w := wanted[c.RequestID]
						if w == nil || w.app != slot.app || seen[c.RequestID] {
							t.Fatal("wrong/missing/duplicate completion")
						}
						seen[c.RequestID] = true
						if w.canceled {
							if !errors.Is(c.Result.Err, context.Canceled) {
								t.Fatal("canceled exchange changed outcome", c.Result)
							}
							continue
						}
						if c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
							t.Fatal(c.Result)
						}
						var session string
						var code uint32
						switch b := c.Result.Answer.Body.(type) {
						case *ro.CCA:
							session, code = b.SessionID, b.ResultCode
							if err := checkAnswer(b, "2001", w.session); err != nil {
								t.Fatal(err)
							}
						case *s6c.SRA:
							session, code = b.SessionID, b.ResultCode.Value
						case *sh.UDA:
							session, code = b.SessionID, b.ResultCode.Value
						case *sgd.OFA:
							session, code = b.SessionID, b.ResultCode.Value
						}
						want := uint32(2001)
						if w.app == 1 || w.app == 2 {
							want = 5012
						}
						if session != w.session || code != want || c.Result.Answer.Body.CommandKey() != mixedAnswer(t, w.app).CommandKey() {
							t.Fatal("mixed answer association/result")
						}
						if healthy {
							m.HealthyCompleted++
						}
					}
				}
				refuse := func() {
					body, _ := mixedRequest(t, p.left.node, stalled)
					for _, port := range []*diameter.CompletionPort{ports[stalled], ports[4]} {
						before := p.left.node.Resources()
						if _, err := p.left.peer.Submit(ctx, body, diameter.RequestOptions{}, port); !errors.Is(err, diameter.ErrBackpressure) {
							t.Fatal("saturated class admitted", err)
						}
						after := p.left.node.Resources()
						if before.Outstanding != after.Outstanding || before.OutboundBytes != after.OutboundBytes {
							t.Fatal("refusal retained outbound resources")
						}
						m.Refusals++
					}
					m.PortRefusals++
					m.ClassRefusals++
				}
				var healthyApps []int
				for app := range mixedNames {
					if app != stalled {
						healthyApps = append(healthyApps, app)
					}
				}
				for wave := 0; wave < m.Waves; wave++ {
					all := submit([]int{0, 1, 2, 3})
					if p.left.node.Resources().Outstanding != 4*window || p.right.node.Resources().InboundOutstanding != 4*window {
						t.Fatal("mixed window not saturated")
					}
					var slow, healthy []*mixedExchange
					for _, w := range all {
						if w.app == stalled {
							slow = append(slow, w)
						} else {
							healthy = append(healthy, w)
						}
					}
					respond(healthy)
					consume(healthy, true)
					pressureWait(t, ctx, func() bool {
						return p.right.node.Resources().InboundOutstanding == window && p.left.node.Resources().Outstanding == window
					})
					refuse()
					// Other classes complete a fresh full wave while this handler
					// deliberately retains its inbound requests without answering.
					healthy = submit(healthyApps)
					respond(healthy)
					consume(healthy, true)
					for i, w := range slow {
						if wave%2 == 1 && i%2 == 0 {
							if err := p.left.node.Cancel(w.id); err != nil {
								t.Fatal(err)
							}
							w.canceled = true
							m.Canceled++
						}
					}
					respond(slow)
					pressureWait(t, ctx, func() bool {
						return p.left.node.Resources().Pending == 0 && p.right.node.Resources().InboundOutstanding == 0
					})
					refuse()
					// Completed but unconsumed outcomes still retain class/port
					// capacity, while the other classes continue independently.
					healthy = submit(healthyApps)
					respond(healthy)
					consume(healthy, true)
					consume(slow, false)
					pressureWait(t, ctx, func() bool {
						a, b := p.left.node.Resources(), p.right.node.Resources()
						return a.Outstanding == 0 && a.OutboundBytes == 0 && b.InboundOutstanding == 0 && b.InboundBytes == 0
					})
				}
				if len(seen) != m.Requests || m.Requests != 80*window || m.HealthyCompleted != 72*window || m.Refusals != 32 {
					t.Fatal("incomplete mixed workload", m)
				}
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("MIXED_PRESSURE %s", data)
			})
		}
	}
}
