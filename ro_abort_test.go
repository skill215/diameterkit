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

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func abortBody() *abort.ASR {
	return &abort.ASR{SessionID: "ctf.example;abort;1", AuthApplicationID: 4, OriginStateID: avp.Some(uint32(9)), ProxyInfo: []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255}}}}
}
func abortConfig(t *testing.T, cache bool) NodeConfig {
	t.Helper()
	c := roConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	rp, e := ro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	ap, e := abort.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(bp, sp, rp, ap)
	if e != nil {
		t.Fatal(e)
	}
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	out.Command = abortBody().CommandKey()
	in.Command = out.Command
	in.Sources = []pdu.Origin{serverOrigin}
	c.Peers[0].Commands = append(c.Peers[0].Commands, out)
	c.Peers[0].InboundCommands = []InboundPolicy{in}
	if cache {
		d := duplicateConfig(t)
		c.Resources.NodeBytes = d.Resources.NodeBytes
		c.Resources.DuplicateClasses = d.Resources.DuplicateClasses
		c.Peers[0].InboundCommands[0].NoCache = false
		c.Peers[0].InboundCommands[0].Duplicate = d.Peers[0].InboundCommands[0].Duplicate
	}
	return c
}
func abortRequest(t *testing.T, n *Node, co *codec.Codec, hbh uint32) pdu.Message {
	t.Helper()
	m, e := co.CompleteRequest(abortBody(), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 1000})
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func receiveAbort(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r := abortRequest(t, n, co, hbh)
	if e := writeWire(ctxTest(t), remote, co, r); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, r
}
func assertNoAbortFollowup(t *testing.T, remote transport.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("unexpected automatic request/answer", e)
	}
}
func TestRoAbortNetworkPairsAndExplicitDecision(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, abortConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for i, body := range []pdu.Body{&abort.ASA{ResultCode: 2001}, &abort.ASA{ResultCode: 5002}, &abort.ASA{ResultCode: 5012}, &pdu.ErrorAnswer{ResultCode: 3004}} {
			id, e := p.Submit(ctxTest(t), abortBody(), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			r := routeRead(t, co, remote)
			a, e := co.CompleteAnswer(r, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
			in, req := receiveAbort(t, n, p, co, remote, uint32(600+i))
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			select {
			case <-in.Done():
				t.Fatal("automatic decision before application response")
			default:
			}
			if result, e := in.Respond(ctxTest(t), &abort.ASA{}); !errors.Is(e, validation.ErrUnsupported) || result.Committed {
				t.Fatal(result, e)
			}
			in.Message().Body.(*abort.ASR).SessionID = "changed-view"
			if result, e := in.Respond(ctxTest(t), body); e != nil || !result.Committed {
				t.Fatal(result, e)
			}
			answer := duplicateAnswer(t, n, co, remote, req)
			if a, ok := answer.Body.(*abort.ASA); ok && a.OriginStateID.Present {
				t.Fatal("remote state echoed")
			}
			if _, e = in.Respond(ctxTest(t), body); !errors.Is(e, ErrInboundDone) {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().Outstanding == 0 })
		}
		if _, e := n.NewRoute(RouteConfig{Command: abortBody().CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 2 * time.Second}}); !errors.Is(e, ErrUnsupported) {
			t.Fatal("ASR gained replay", e)
		}
		assertNoAbortFollowup(t, remote)
	})
}
func abortArrival(t *testing.T, co *codec.Codec, r pdu.Message, hbh uint32) pdu.Message {
	t.Helper()
	m, e := co.Snapshot(r)
	if e != nil {
		t.Fatal(e)
	}
	m.Header.HopByHop = hbh
	m.Header.Flags |= pdu.RetransmittedFlag
	m.Body.(*abort.ASR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{byte(hbh)}}}
	return m
}
func TestRoAbortDuplicateContentAndCommittedDecision(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, abortConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, r := receiveAbort(t, n, p, co, remote, 610)
		before := n.Resources()
		dup := abortArrival(t, co, r, 611)
		if e := writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		bad := abortArrival(t, co, r, 612)
		bad.Body.(*abort.ASR).OriginStateID = avp.Some(uint32(10))
		if e := writeWire(ctxTest(t), remote, co, bad); e != nil {
			t.Fatal(e)
		}
		readReject(t, remote, bad.Header, 5012)
		if n.Resources().DuplicateRecords != 1 || n.Resources().DuplicateBytes != before.DuplicateBytes {
			t.Fatal(n.Resources())
		}
		if result, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 2001}); e != nil || !result.Committed {
			t.Fatal(result, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			req := r
			if a.Header.HopByHop == dup.Header.HopByHop {
				req = dup
			}
			if e := co.ValidateAnswer(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		late := abortArrival(t, co, r, 613)
		if e := writeWire(ctxTest(t), remote, co, late); e != nil {
			t.Fatal(e)
		}
		a := duplicateAnswer(t, n, co, remote, late)
		if a.Body.(*abort.ASA).ResultCode != 2001 {
			t.Fatal("committed result changed")
		}
		noDuplicateDelivery(t, p)
		if _, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 5002}); !errors.Is(e, ErrInboundDone) {
			t.Fatal("business redecision replaced cached ASA", e)
		}
	})
}
func TestRoAbortCancellationExpiryAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, mode := range []string{"response-cancel", "execution-expiry", "close-race"} {
			t.Run(mode, func(t *testing.T) {
				cfg := abortConfig(t, false)
				if mode == "execution-expiry" {
					cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 100 * time.Millisecond
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				in, r := receiveAbort(t, n, p, co, remote, 620)
				switch mode {
				case "response-cancel":
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					if result, e := in.Respond(ctx, &abort.ASA{ResultCode: 5012}); !errors.Is(e, context.Canceled) || result.Committed {
						t.Fatal(result, e)
					}
					if result, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 5002}); e != nil || !result.Committed {
						t.Fatal(result, e)
					}
					duplicateAnswer(t, n, co, remote, r)
				case "execution-expiry":
					readReject(t, remote, r.Header, 5012)
					if _, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 2001}); !errors.Is(e, context.DeadlineExceeded) {
						t.Fatal(e)
					}
				case "close-race":
					var wg sync.WaitGroup
					wg.Add(2)
					go func() { defer wg.Done(); _, _ = in.Respond(context.Background(), &abort.ASA{ResultCode: 5012}) }()
					go func() { defer wg.Done(); _ = p.Close() }()
					wg.Wait()
					if e := p.Wait(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
					if _, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 2001}); e == nil {
						t.Fatal("closed inbound reopened")
					}
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
			})
		}
	})
}
func TestRoAbortOutboundLateAnswerAndAuthority(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, mode := range []string{"cancel", "wrong-terminal", "wrong-session"} {
			t.Run(mode, func(t *testing.T) {
				n, co := newInboundNode(t, abortConfig(t, false))
				p, remote := networkConnectNode(t, n, co, tls)
				port := portTest(t, n, 1)
				id, e := p.Submit(ctxTest(t), abortBody(), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				r := routeRead(t, co, remote)
				origin := serverOrigin
				if mode == "wrong-terminal" {
					origin = pdu.Origin{Host: "other.example", Realm: "example"}
				}
				a, e := co.CompleteAnswer(r, &abort.ASA{ResultCode: 2001}, origin)
				if e != nil {
					t.Fatal(e)
				}
				if mode == "wrong-session" {
					a.Body.(*abort.ASA).SessionID += "other"
				}
				if mode == "cancel" {
					if e = n.Cancel(id); e != nil {
						t.Fatal(e)
					}
					got := routeCompletion(t, n, port, id)
					if !errors.Is(got.Err, context.Canceled) {
						t.Fatal(got)
					}
				}
				if e = writeWire(ctxTest(t), remote, co, a); e != nil {
					t.Fatal(e)
				}
				if mode != "cancel" {
					got := routeCompletion(t, n, port, id)
					if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
						t.Fatal(got)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
				defer cancel()
				if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("second completion", e)
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
				assertNoAbortFollowup(t, remote)
			})
		}
	})
}
func TestRoAbortIndependentInFlightTermination(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, first := range []string{"asa", "cca", "cancel-term"} {
			t.Run(first, func(t *testing.T) {
				n, co := newInboundNode(t, abortConfig(t, false))
				p, remote := networkConnectNode(t, n, co, tls)
				port := portTest(t, n, 1)
				term := roECURBody(true)
				term.SessionID = abortBody().SessionID
				id, e := p.Submit(ctxTest(t), term, RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				request := routeRead(t, co, remote)
				in, asr := receiveAbort(t, n, p, co, remote, 630)
				cca, e := co.CompleteAnswer(request, roECURAnswer(true, 2001), serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				respond := func() {
					t.Helper()
					result, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 2001})
					if e != nil || !result.Committed {
						t.Fatal(result, e)
					}
					duplicateAnswer(t, n, co, remote, asr)
				}
				if first == "asa" {
					respond()
					if n.Resources().Outstanding != 1 {
						t.Fatal("ASA completed TERM")
					}
				}
				if first == "cancel-term" {
					if e = n.Cancel(id); e != nil {
						t.Fatal(e)
					}
					got := routeCompletion(t, n, port, id)
					if !errors.Is(got.Err, context.Canceled) {
						t.Fatal(got)
					}
					respond()
				} else {
					if e = writeWire(ctxTest(t), remote, co, cca); e != nil {
						t.Fatal(e)
					}
					got := routeCompletion(t, n, port, id)
					if got.Outcome != AnswerReceived {
						t.Fatal(got)
					}
					if first == "cca" {
						respond()
					}
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 && n.Resources().InboundOutstanding == 0 })
				assertNoAbortFollowup(t, remote)
			})
		}
	})
}
func TestRoAbortInboundSourceDestinationAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, abortConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, mutate := range []func(*abort.ASR){func(r *abort.ASR) { r.OriginHost = "unknown.example" }, func(r *abort.ASR) { r.DestinationHost = "other.example" }, func(r *abort.ASR) { r.DestinationRealm = "other.example" }} {
			r := abortRequest(t, n, co, uint32(640+i))
			mutate(r.Body.(*abort.ASR))
			if e := writeWire(ctxTest(t), remote, co, r); e != nil {
				t.Fatal(e)
			}
			// Source and destination violations use the configured finite failure policy.
			a := routeRead(t, co, remote)
			if _, ok := a.Body.(*pdu.ErrorAnswer); !ok {
				t.Fatalf("%T", a.Body)
			}
			noDuplicateDelivery(t, p)
		}
		in, r := receiveAbort(t, n, p, co, remote, 650)
		if n.Resources().InboundBytes == 0 {
			t.Fatal("uncharged inbound")
		}
		big := &abort.ASA{ResultCode: 5012, ErrorMessage: avp.Some(fmt.Sprintf("%02048d", 1))}
		if result, e := in.Respond(ctxTest(t), big); e == nil || result.Committed {
			t.Fatal("oversize accepted", result, e)
		}
		if result, e := in.Respond(ctxTest(t), &abort.ASA{ResultCode: 5012}); e != nil || !result.Committed {
			t.Fatal(result, e)
		}
		duplicateAnswer(t, n, co, remote, r)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().InboundBytes == 0 })
	})
}

func TestRoAbortCommittedResultSurvivesPrimaryLoss(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, twoDuplicatePeers(abortConfig(t, true)))
		p, a := networkConnectNode(t, n, co, tls)
		local, b := networkAdapters(t, tls)
		ch := make(chan error, 1)
		ctx := ctxTest(t)
		go func() {
			r, e := readWire(ctx, b, co)
			if e == nil {
				var answer pdu.Message
				answer, e = co.CompleteAnswer(r, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "second-dra", AuthApplicationID: n.cfg.Capabilities.AuthApplicationID}, n.peers["dra2"].origin)
				if e == nil {
					e = writeWire(ctx, b, co, answer)
				}
			}
			ch <- e
		}()
		p2, e := n.Connect(ctx, "dra2", local)
		if e != nil {
			t.Fatal(e)
		}
		if e = <-ch; e != nil {
			t.Fatal(e)
		}
		in, r := receiveAbort(t, n, p, co, a, 660)
		_ = a.Close()
		waitFor(t, func() bool { return p.Info().Candidates == 0 })
		dup := abortArrival(t, co, r, 661)
		if e = writeWire(ctx, b, co, dup); e != nil {
			t.Fatal(e)
		}
		noDuplicateDelivery(t, p2)
		result, e := in.Respond(ctx, &abort.ASA{ResultCode: 2001})
		if !result.Committed || result.PrimaryWrite.State != ResponseNotWritten || e == nil {
			t.Fatal(result, e)
		}
		after := duplicateAnswer(t, n, co, b, dup)
		if after.Body.(*abort.ASA).ResultCode != 2001 {
			t.Fatal("lost committed decision")
		}
		if _, e = in.Respond(ctx, &abort.ASA{ResultCode: 5002}); !errors.Is(e, ErrInboundDone) {
			t.Fatal("transport failure reopened business decision", e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
