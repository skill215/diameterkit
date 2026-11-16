// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func roBody() *ro.CCR {
	return &ro.CCR{SessionID: "ctf.example;iec;1", AuthApplicationID: 4, ServiceContextID: ro.SMSServiceContextID, CCRequestType: ro.EventRequest, CCRequestNumber: 0, RequestedAction: avp.Some(ro.DirectDebiting), MultipleServicesIndicator: avp.Some(uint32(1)),
		MultipleServicesCreditControl: []ro.RequestMSCC{{ServiceIdentifier: []uint32{1}, RequestedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))})}},
		ServiceInformation:            avp.Some(ro.ServiceInformation{SMSInformation: avp.Some(ro.SMSInformation{SMSNode: avp.Some(uint32(1)), SMMessageType: avp.Some(uint32(0))})})}
}
func roAnswer(code uint32) *ro.CCA {
	a := &ro.CCA{AuthApplicationID: 4, CCRequestType: ro.EventRequest, CCRequestNumber: 0, ResultCode: code}
	if code == 2001 {
		a.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(uint32(2001)), GrantedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))}), RefundInformation: avp.Some([]byte{0, 255})}}
	}
	return a
}
func roConfig(t *testing.T) NodeConfig {
	c := rejectConfig(t)
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
	c.Bundle, e = schema.Compile(bp, sp, rp)
	if e != nil {
		t.Fatal(e)
	}
	c.Capabilities.AuthApplicationID = append(c.Capabilities.AuthApplicationID, 4)
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	out.Command = roBody().CommandKey()
	in.Command = out.Command
	out.DestinationHost = serverOrigin.Host
	in.Sources = []pdu.Origin{gmscOrigin}
	c.Peers[0].Commands = []CommandPolicy{out}
	c.Peers[0].InboundCommands = []InboundPolicy{in}
	return c
}
func TestRoIECNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for i, body := range []pdu.Body{roAnswer(2001), roAnswer(4012), roAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			request := roBody()
			id, e := p.Submit(ctxTest(t), request, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			request.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 99
			req := routeRead(t, co, remote)
			if req.Body.(*ro.CCR).MultipleServicesCreditControl[0].ServiceIdentifier[0] != 1 {
				t.Fatal("submit alias")
			}
			ans, e := co.CompleteAnswer(req, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
			inbound, e := co.CompleteRequest(roBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(900 + i), EndToEnd: uint32(1900 + i)})
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, inbound); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			in.Message().Body.(*ro.CCR).MultipleServicesCreditControl[0].ServiceIdentifier[0] = 22
			if result, e := in.Respond(ctxTest(t), body); e != nil || !result.Committed {
				t.Fatal(result, e)
			}
			answer := routeRead(t, co, remote)
			if e = co.ValidateAnswer(inbound, answer, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().Outstanding == 0 })
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic follow-up", e)
		}
	})
}
func TestRoIECInvalidServiceCompletesOnce(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		changed, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		changed.Body.(*ro.CCR).MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2
		a := roAnswer(2001)
		a.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2
		answer, e := co.CompleteAnswer(changed, a, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, answer); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestRoIECCancelLateAnswerDoesNotReviveRequest(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		answer, e := co.CompleteAnswer(req, roAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, answer); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late completion", e)
		}
		// A subsequently configured recovery route cannot revive a completed Peer request.
		if _, e = n.NewRoute(RouteConfig{Command: roBody().CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 2 * time.Second}}); e != nil {
			t.Fatal(e)
		}
		if _, e = n.Retransmit(ctxTest(t), id, AttemptID{id, 1}); !errors.Is(e, ErrRequestNotActive) {
			t.Fatal("completed request revived", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}

func TestRoIECInboundPreparationRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		req, e := co.CompleteRequest(roBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 990, EndToEnd: 1990})
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
		wrong := roAnswer(2001)
		wrong.MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2
		if got, e := in.Respond(ctxTest(t), wrong); got.Committed || !errors.Is(e, validation.ErrInvalidAnswer) {
			t.Fatal(got, e)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if got, e := in.Respond(ctx, roAnswer(2001)); got.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(got, e)
		}
		body := roAnswer(2001)
		body.MultipleServicesCreditControl[0].GrantedServiceUnit.Value.CCServiceSpecificUnits = avp.Some(^uint64(0))
		if got, e := in.Respond(ctxTest(t), body); e != nil || !got.Committed {
			t.Fatal(got, e)
		}
		body.MultipleServicesCreditControl[0].RefundInformation.Value[0] = 9
		ans := routeRead(t, co, remote)
		a := ans.Body.(*ro.CCA)
		if a.MultipleServicesCreditControl[0].RefundInformation.Value[0] != 0 || a.MultipleServicesCreditControl[0].GrantedServiceUnit.Value.CCServiceSpecificUnits.Value != ^uint64(0) {
			t.Fatal("answer snapshot")
		}
		if e = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
