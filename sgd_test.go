// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func smsBody(mt bool) pdu.Body {
	if mt {
		return &sgd.TFR{SessionID: "gmsc.example;mt;1", AuthSessionState: 1, UserName: "001010123456789", SCAddress: []byte{0x21, 0xf3}, SMRPUI: []byte{4, 0, 255, 1}, MMENumberForMTSMS: avp.Some([]byte{0x21, 0xf5}), SMDeliveryTimer: avp.Some(uint32(30)), SMDeliveryStartTime: avp.Some(avp.Time(0))}
	}
	return &sgd.OFR{SessionID: "mme.example;mo;1", AuthSessionState: 1, UserIdentifier: sgd.UserIdentifier{UserName: avp.Some("001010123456789"), MSISDN: avp.Some([]byte{0x21, 0xf4})}, SCAddress: []byte{0x21, 0xf3}, SMRPUI: []byte{1, 0, 255, 1}}
}
func smsAnswer(mt bool, code uint32) pdu.Body {
	if mt {
		return &sgd.TFA{AuthSessionState: 1, ResultCode: avp.Some(code)}
	}
	return &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(code)}
}
func smsConfig(t *testing.T, cache bool) NodeConfig {
	t.Helper()
	c := rejectConfig(t)
	if cache {
		c = duplicateConfig(t)
	}
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	gp, e := sgd.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(bp, sp, gp)
	if e != nil {
		t.Fatal(e)
	}
	c.Capabilities.AuthApplicationID = append(c.Capabilities.AuthApplicationID, 16777313)
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	c.Peers[0].Commands = nil
	c.Peers[0].InboundCommands = nil
	for _, mt := range []bool{false, true} {
		p, q := out, in
		p.Command = smsBody(mt).CommandKey()
		q.Command = p.Command
		p.DestinationHost = serverOrigin.Host
		q.Sources = []pdu.Origin{gmscOrigin}
		c.Peers[0].Commands = append(c.Peers[0].Commands, p)
		c.Peers[0].InboundCommands = append(c.Peers[0].InboundCommands, q)
	}
	return c
}
func smsRequest(t *testing.T, n *Node, co *codec.Codec, mt, sgsn bool, hbh uint32) pdu.Message {
	t.Helper()
	body := smsBody(mt)
	if mt && sgsn {
		b := body.(*sgd.TFR)
		b.SGSNNumber = b.MMENumberForMTSMS
		b.MMENumberForMTSMS = avp.Optional[[]byte]{}
	}
	r, e := co.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 10000})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func sendSMS(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, mt, sgsn bool, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r := smsRequest(t, n, co, mt, sgsn, hbh)
	if e := writeWire(ctxTest(t), remote, co, r); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, r
}
func TestSGDIndependentExchangesAndResults(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 4)
		ids := map[RequestID]uint32{}
		for _, mt := range []bool{false, true} {
			body := smsBody(mt)
			id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			ids[id] = body.CommandKey().Code
		}
		reqs := []pdu.Message{routeRead(t, co, remote), routeRead(t, co, remote)}
		if reqs[0].Header.HopByHop == reqs[1].Header.HopByHop || reqs[0].Header.EndToEnd == reqs[1].Header.EndToEnd {
			t.Fatal("shared pending identifiers")
		}
		for i := 1; i >= 0; i-- {
			r := reqs[i]
			a, e := co.CompleteAnswer(r, smsAnswer(r.Header.CommandCode == 8388646, 2001), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
		}
		for range 2 {
			r, e := port.Receive(ctxTest(t))
			if e != nil || r.Result.Outcome != AnswerReceived || r.Result.Answer.Header.CommandCode != ids[r.RequestID] || !r.Result.Validation.Complete {
				t.Fatal(r, e)
			}
			delete(ids, r.RequestID)
		}
		// The library reports the result; none of these answers initiates another SMS.
		for i, tc := range []struct {
			mt   bool
			body pdu.Body
		}{
			{false, &sgd.OFA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), SMRPUI: avp.Some([]byte{1, 0, 255})}},
			{false, &sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555}), SMDeliveryFailureCause: avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 3})}},
			{false, smsAnswer(false, 5012)}, {false, &pdu.ErrorAnswer{ResultCode: 3004}},
			{true, &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555}), SMDeliveryFailureCause: avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 0})}},
			{true, &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5550}), AbsentUserDiagnosticSM: avp.Some(uint32(13))}},
			{true, smsAnswer(true, 2001)}, {true, smsAnswer(true, 5012)}, {true, &pdu.ErrorAnswer{ResultCode: 3004}},
		} {
			in, req := sendSMS(t, n, p, co, remote, tc.mt, tc.mt, uint32(800+i))
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			if r, e := in.Respond(ctxTest(t), tc.body); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			ans := routeRead(t, co, remote)
			if e := co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("unexpected follow-up", e)
		}
	})
}
func TestSGDPrivateTargetAndPrecommitRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := sendSMS(t, n, p, co, remote, true, true, 820)
		// The application view must not change the original SGSN association.
		b := in.Message().Body.(*sgd.TFR)
		b.MMENumberForMTSMS = b.SGSNNumber
		b.SGSNNumber = avp.Optional[[]byte]{}
		b.SessionID = "changed"
		wrong := &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5557})}
		if r, e := in.Respond(ctxTest(t), wrong); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(canceled, smsAnswer(true, 2001)); r.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), smsAnswer(true, 2001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		ans := routeRead(t, co, remote)
		if e := co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
	})
}
func TestSGDInvalidMatchedTargetCompletesOnce(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		body := smsBody(true).(*sgd.TFR)
		body.SGSNNumber = body.MMENumberForMTSMS
		body.MMENumberForMTSMS = avp.Optional[[]byte]{}
		id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		changed, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		b := changed.Body.(*sgd.TFR)
		b.MMENumberForMTSMS = b.SGSNNumber
		b.SGSNNumber = avp.Optional[[]byte]{}
		ans, e := co.CompleteAnswer(changed, &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5557})}, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id || r.Result.Answer.Body != nil || !errors.Is(r.Result.Err, validation.ErrInvalidAnswer) {
			t.Fatal(r, e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestSGDDuplicateTPDUAndArrivalReplies(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, mt := range []bool{false, true} {
			in, req := sendSMS(t, n, p, co, remote, mt, mt, uint32(840+i*10))
			dup, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			dup.Header.HopByHop++
			dup.Header.Flags |= pdu.RetransmittedFlag
			proxy := []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255, 1}}}
			if mt {
				dup.Body.(*sgd.TFR).ProxyInfo = proxy
			} else {
				dup.Body.(*sgd.OFR).ProxyInfo = proxy
			}
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			noDuplicateDelivery(t, p)
			if r, e := in.Respond(ctxTest(t), smsAnswer(mt, 2001)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			for range 2 {
				ans := routeRead(t, co, remote)
				r := req
				if ans.Header.HopByHop == dup.Header.HopByHop {
					r = dup
				}
				if e = co.ValidateAnswer(r, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
					t.Fatal(e)
				}
			}
			dup.Header.HopByHop++
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			duplicateAnswer(t, n, co, remote, dup)
			noDuplicateDelivery(t, p)
			if mt {
				dup.Body.(*sgd.TFR).SMRPUI[0] ^= 1
			} else {
				dup.Body.(*sgd.OFR).SMRPUI[0] ^= 1
			}
			dup.Header.HopByHop++
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			a := duplicateAnswer(t, n, co, remote, dup)
			if b, ok := a.Body.(*pdu.ErrorAnswer); !ok || b.ResultCode != 5012 {
				t.Fatal(a)
			}
			noDuplicateDelivery(t, p)
		}
	})
}
func TestSGDBudgetAndCloseRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := smsConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, _ := sendSMS(t, n, p, co, remote, false, false, 870)
		req := smsRequest(t, n, co, true, false, 871)
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		a := duplicateAnswer(t, n, co, remote, req)
		if b, ok := a.Body.(*pdu.ErrorAnswer); !ok || b.ResultCode != 3004 {
			t.Fatal(a)
		}
		if r, e := in.Respond(ctxTest(t), smsAnswer(false, 2001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		routeRead(t, co, remote)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		in, _ = sendSMS(t, n, p, co, remote, true, true, 872)
		start := make(chan struct{})
		results := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() { defer wg.Done(); <-start; r, _ := in.Respond(ctxTest(t), smsAnswer(true, 2001)); results <- r }()
		}
		close(start)
		_ = n.Close()
		wg.Wait()
		close(results)
		committed := 0
		for r := range results {
			if r.Committed {
				committed++
			}
		}
		if committed > 1 {
			t.Fatal("multiple commits")
		}
		if e := n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundOutstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestSGDCancelLateAnswerAndNoReplay(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		ctx := ctxTest(t)
		id, e := p.Submit(ctx, smsBody(true), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		r, e := port.Receive(ctxTest(t))
		if e != nil || r.RequestID != id || !errors.Is(r.Result.Err, context.Canceled) {
			t.Fatal(r, e)
		}
		a, e := co.CompleteAnswer(req, smsAnswer(true, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel2()
		if _, e = port.Receive(ctx2); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late completion", e)
		}
		for _, mt := range []bool{false, true} {
			for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
				if _, e = n.NewRoute(RouteConfig{Command: smsBody(mt).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: mode, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 2 * time.Second}}); e == nil {
					t.Fatal("unaudited replay enabled")
				}
			}
		}
	})
}
