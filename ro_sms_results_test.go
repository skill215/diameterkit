package diameter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func smsResultPair(kind string, code uint32, child bool) (*ro.CCR, *ro.CCA) {
	q, a := roBody(), roAnswer(code)
	if kind != "iec" {
		q, a = roECURBody(kind == "term"), roECURAnswer(kind == "term", code)
	}
	if child {
		a.ResultCode = 2001
		a.MultipleServicesCreditControl = []ro.AnswerMSCC{{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(code)}}
	}
	return q, a
}

func TestRoSMSResultsNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		counter := uint32(1000)
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, code := range []uint32{4010, 4011, 5030} {
				for _, child := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%d/child%v", kind, code, child), func(t *testing.T) {
						q, a := smsResultPair(kind, code, child)
						id, err := p.Submit(ctxTest(t), q, RequestOptions{}, port)
						if err != nil {
							t.Fatal(err)
						}
						r := routeRead(t, co, remote)
						if kind == "term" && r.Body.(*ro.CCR).MultipleServicesCreditControl[0].UsedServiceUnit[0].CCServiceSpecificUnits.Value != 1 {
							t.Fatal("usage erased")
						}
						ans, err := co.CompleteAnswer(r, a, serverOrigin)
						if err != nil {
							t.Fatal(err)
						}
						if err = writeWire(ctxTest(t), remote, co, ans); err != nil {
							t.Fatal(err)
						}
						got := routeCompletion(t, n, port, id)
						if got.Outcome != AnswerReceived || !got.Validation.Complete {
							t.Fatal(got)
						}
						v := got.Answer.Body.(*ro.CCA)
						if v.ResultCode != a.ResultCode || child && (v.MultipleServicesCreditControl[0].ResultCode.Value != code || v.MultipleServicesCreditControl[0].GrantedServiceUnit.Present) {
							t.Fatal("flattened result", v)
						}
						counter++
						inReq, err := co.CompleteRequest(q, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: counter, EndToEnd: counter + 1000})
						if err != nil {
							t.Fatal(err)
						}
						if err = writeWire(ctxTest(t), remote, co, inReq); err != nil {
							t.Fatal(err)
						}
						in, err := p.Receive(ctxTest(t))
						if err != nil {
							t.Fatal(err)
						}
						if res, err := in.Respond(ctxTest(t), a); err != nil || !res.Committed {
							t.Fatal(res, err)
						}
						answer := routeRead(t, co, remote)
						if err = co.ValidateAnswer(inReq, answer, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
							t.Fatal(err)
						}
						if _, err = in.Respond(ctxTest(t), a); !errors.Is(err, ErrInboundDone) {
							t.Fatal(err)
						}
						waitFor(t, func() bool { return n.Resources().Outstanding == 0 && n.Resources().InboundOutstanding == 0 })
					})
				}
			}
		}
		assertNoAbortFollowup(t, remote)
	})
}

func TestRoSMSRefundNetworkEvidenceAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, roConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for i, token := range []avp.Optional[[]byte]{{}, avp.Some([]byte{}), avp.Some([]byte{0, 255, 0, 1})} {
			id, err := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port)
			if err != nil {
				t.Fatal(err)
			}
			r := routeRead(t, co, remote)
			a := roAnswer(2001)
			a.MultipleServicesCreditControl[0].RefundInformation = token
			ans, err := co.CompleteAnswer(r, a, serverOrigin)
			if err != nil {
				t.Fatal(err)
			}
			if err = writeWire(ctxTest(t), remote, co, ans); err != nil {
				t.Fatal(err)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
			leaf, err := ro.RefundInformationForService(r.Body.(*ro.CCR), got.Answer.Body.(*ro.CCA), 1, 4)
			if err != nil || leaf.Present != token.Present || !bytes.Equal(leaf.Value, token.Value) {
				t.Fatal(leaf, err)
			}
			if err = ro.CheckRefundInformationEcho(r.Body.(*ro.CCR), got.Answer.Body.(*ro.CCA), 1, leaf, 4); err != nil {
				t.Fatal(err)
			}
			inReq, err := co.CompleteRequest(roBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(700 + i), EndToEnd: uint32(1700 + i)})
			if err != nil {
				t.Fatal(err)
			}
			if err = writeWire(ctxTest(t), remote, co, inReq); err != nil {
				t.Fatal(err)
			}
			in, err := p.Receive(ctxTest(t))
			if err != nil {
				t.Fatal(err)
			}
			big := roAnswer(2001)
			big.MultipleServicesCreditControl[0].RefundInformation = avp.Some(make([]byte, 2048))
			if res, err := in.Respond(ctxTest(t), big); err == nil || res.Committed {
				t.Fatal("oversize token committed", res, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if res, err := in.Respond(ctx, a); !errors.Is(err, context.Canceled) || res.Committed {
				t.Fatal(res, err)
			}
			if res, err := in.Respond(ctxTest(t), a); err != nil || !res.Committed {
				t.Fatal(res, err)
			}
			answer := routeRead(t, co, remote)
			if err = co.ValidateAnswer(inReq, answer, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
				t.Fatal(err)
			}
			if _, err = ro.RefundInformationForService(inReq.Body.(*ro.CCR), answer.Body.(*ro.CCA), 1, 4); err != nil {
				t.Fatal(err)
			}
			if _, err = in.Respond(ctxTest(t), a); !errors.Is(err, ErrInboundDone) {
				t.Fatal(err)
			}
			waitFor(t, func() bool {
				return n.Resources().Outstanding == 0 && n.Resources().InboundOutstanding == 0 && n.Resources().InboundBytes == 0
			})
		}
		assertNoAbortFollowup(t, remote)
	})
}

func TestRoSMSRefundCancelLateAnswerAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, closePeer := range []bool{false, true} {
			t.Run(fmt.Sprint(closePeer), func(t *testing.T) {
				n, co := newInboundNode(t, roConfig(t))
				p, remote := networkConnectNode(t, n, co, tls)
				port := portTest(t, n, 1)
				id, err := p.Submit(ctxTest(t), roBody(), RequestOptions{}, port)
				if err != nil {
					t.Fatal(err)
				}
				r := routeRead(t, co, remote)
				if closePeer {
					var wg sync.WaitGroup
					wg.Add(2)
					go func() { defer wg.Done(); _ = n.Cancel(id) }()
					go func() { defer wg.Done(); _ = p.Close() }()
					wg.Wait()
				} else if err = n.Cancel(id); err != nil {
					t.Fatal(err)
				}
				got := routeCompletion(t, n, port, id)
				if got.Answer.Body != nil {
					t.Fatal("cancel invented original answer", got)
				}
				if _, err = ro.RefundInformationForService(r.Body.(*ro.CCR), nil, 1, 4); !errors.Is(err, validation.ErrContextRequired) {
					t.Fatal(err)
				}
				if !closePeer {
					a, err := co.CompleteAnswer(r, roAnswer(2001), serverOrigin)
					if err != nil {
						t.Fatal(err)
					}
					if err = writeWire(ctxTest(t), remote, co, a); err != nil {
						t.Fatal(err)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
				defer cancel()
				if _, err = port.Receive(ctx); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal("second completion", err)
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
			})
		}
	})
}
