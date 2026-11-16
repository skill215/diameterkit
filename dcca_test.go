package diameter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/ro/dcca"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func dccaBody(action uint32) *dcca.CCR {
	return &dcca.CCR{SessionID: fmt.Sprintf("ctf.example;event;%d", action), AuthApplicationID: 4, ServiceContextID: "event@example.net", CCRequestType: 4, RequestedAction: action, ServiceIdentifier: 7, SubscriptionID: []dcca.SubscriptionID{{SubscriptionIDType: 0, SubscriptionIDData: "15550000001"}}}
}
func dccaAnswer(action, code uint32) *dcca.CCA {
	a := &dcca.CCA{ResultCode: code, AuthApplicationID: 4, CCRequestType: 4}
	if code == 2001 {
		switch action {
		case dcca.CheckBalance:
			a.CheckBalanceResult = avp.Some(dcca.NoCredit)
		case dcca.DirectDebiting:
			a.GrantedServiceUnit = avp.Some(dcca.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(0))})
		}
	}
	return a
}
func dccaConfig(t *testing.T) NodeConfig {
	t.Helper()
	cfg := roConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	dp, e := dcca.Profile()
	if e != nil {
		t.Fatal(e)
	}
	cfg.Bundle, e = schema.Compile(bp, sp, dp)
	if e != nil {
		t.Fatal(e)
	}
	cfg.Peers[0].InboundCommands[0].Sources = []pdu.Origin{serverOrigin}
	return cfg
}
func TestDCCANetworkOperationsAndNoAutomaticRefund(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, dccaConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, action := range []uint32{dcca.CheckBalance, dcca.DirectDebiting, dcca.RefundAccount} {
			for _, code := range []uint32{2001, 4010, 4011, 4012, 5030, 5012} {
				body := dccaBody(action)
				id, e := p.Submit(ctxTest(t), body, RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				body.SubscriptionID[0].SubscriptionIDData = "9"
				req := routeRead(t, co, remote)
				if req.Body.(*dcca.CCR).SubscriptionID[0].SubscriptionIDData != "15550000001" {
					t.Fatal("submit alias")
				}
				a, e := co.CompleteAnswer(req, dccaAnswer(action, code), serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, a); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || !got.Validation.Complete || got.Answer.Body.(*dcca.CCA).ResultCode != code {
					t.Fatal(got)
				}
				if action == dcca.CheckBalance && code == 2001 && got.Answer.Body.(*dcca.CCA).CheckBalanceResult.Value != dcca.NoCredit {
					t.Fatal("NO_CREDIT flattened")
				}
			}
		}
		if _, e := n.NewRoute(RouteConfig{Command: dccaBody(0).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: ExplicitRetransmit, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: 2 * time.Second}}); !errors.Is(e, ErrUnsupported) {
			t.Fatal("new binding gained recovery", e)
		}
		if _, e := n.PrepareRoRecoveryRequest(ctxTest(t), dccaBody(0), RoRecoveryOptions{}); !errors.Is(e, ErrUnsupported) {
			t.Fatal("record import capability broadened", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		assertNoAbortFollowup(t, remote)
	})
}
func TestDCCANetworkConditionalReportsAndActionAssociation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, mode := range []string{"missing", "unknown", "wrong-action", "wrong-number", "error"} {
			t.Run(mode, func(t *testing.T) {
				n, co := newInboundNode(t, dccaConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				port := portTest(t, n, 1)
				action := dcca.CheckBalance
				if mode == "wrong-action" {
					action = dcca.RefundAccount
				}
				id, e := p.Submit(ctxTest(t), dccaBody(action), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, remote)
				body := dccaAnswer(action, 2001)
				if mode == "missing" {
					body.CheckBalanceResult = avp.Optional[uint32]{}
				}
				ans, e := co.CompleteAnswer(req, body, serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				raw := reqBytes(t, co, ans)
				switch mode {
				case "unknown":
					raw = replaceAVP32(t, raw, 422, 99)
				case "wrong-number":
					raw = replaceAVP32(t, raw, 415, 1)
				case "wrong-action":
					// A structurally valid balance result received against a refund request.
					ans.Body.(*dcca.CCA).CheckBalanceResult = avp.Some(uint32(0))
					raw = reqBytes(t, co, ans)
				case "error":
					ans, e = co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, draOrigin)
					if e != nil {
						t.Fatal(e)
					}
					raw = reqBytes(t, co, ans)
				}
				sendRejectRaw(t, remote, raw)
				got := routeCompletion(t, n, port, id)
				switch mode {
				case "missing", "unknown":
					if got.Outcome != AnswerReceived || got.Answer.Body == nil || got.Validation.Complete {
						t.Fatal("lost owned incomplete answer", got)
					}
					v := got.Answer.Body.(*dcca.CCA).CheckBalanceResult
					if mode == "missing" && v.Present || mode == "unknown" && (!v.Present || v.Value != 99) {
						t.Fatal(v)
					}
				case "wrong-action", "wrong-number":
					if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
						t.Fatal(got)
					}
				case "error":
					if got.Outcome != AnswerReceived || !got.Validation.Complete {
						t.Fatal(got)
					}
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
				assertNoAbortFollowup(t, remote)
			})
		}
	})
}
func TestDCCAInboundResponsePreparationAndBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, dccaConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, action := range []uint32{dcca.CheckBalance, dcca.DirectDebiting, dcca.RefundAccount} {
			r, e := co.CompleteRequest(dccaBody(action), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(700 + i), EndToEnd: uint32(1700 + i)})
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, r); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if n.Resources().InboundBytes == 0 || !in.Validation().Complete {
				t.Fatal(n.Resources(), in.Validation())
			}
			if action != dcca.RefundAccount {
				bad := dccaAnswer(action, 2001)
				bad.CheckBalanceResult = avp.Optional[uint32]{}
				bad.GrantedServiceUnit = avp.Optional[dcca.ServiceUnit]{}
				if res, e := in.Respond(ctxTest(t), bad); e == nil || res.Committed {
					t.Fatal("conditional failure committed", res, e)
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if res, e := in.Respond(ctx, dccaAnswer(action, 2001)); !errors.Is(e, context.Canceled) || res.Committed {
				t.Fatal(res, e)
			}
			big := dccaAnswer(action, 2001)
			big.ErrorMessage = avp.Some(fmt.Sprintf("%02048d", 1))
			if res, e := in.Respond(ctxTest(t), big); e == nil || res.Committed {
				t.Fatal("answer cap bypass", res, e)
			}
			if res, e := in.Respond(ctxTest(t), dccaAnswer(action, 2001)); e != nil || !res.Committed {
				t.Fatal(res, e)
			}
			duplicateAnswer(t, n, co, remote, r)
			if _, e = in.Respond(ctxTest(t), dccaAnswer(action, 5012)); !errors.Is(e, ErrInboundDone) {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().InboundBytes == 0 })
		}
	})
}
func TestDCCACancelLateAnswerAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, closePeer := range []bool{false, true} {
			t.Run(fmt.Sprint(closePeer), func(t *testing.T) {
				n, co := newInboundNode(t, dccaConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				port := portTest(t, n, 1)
				id, e := p.Submit(ctxTest(t), dccaBody(dcca.RefundAccount), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, remote)
				if closePeer {
					var wg sync.WaitGroup
					wg.Add(2)
					go func() { defer wg.Done(); _ = n.Cancel(id) }()
					go func() { defer wg.Done(); _ = p.Close() }()
					wg.Wait()
				} else {
					if e = n.Cancel(id); e != nil {
						t.Fatal(e)
					}
				}
				got := routeCompletion(t, n, port, id)
				if got.Answer.Body != nil {
					t.Fatal("made refund successful", got)
				}
				if !closePeer {
					a, e := co.CompleteAnswer(req, dccaAnswer(dcca.RefundAccount, 2001), serverOrigin)
					if e != nil {
						t.Fatal(e)
					}
					if e = writeWire(ctxTest(t), remote, co, a); e != nil {
						t.Fatal(e)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
				defer cancel()
				if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("second completion", e)
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
			})
		}
	})
}
