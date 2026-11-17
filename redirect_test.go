// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: TCP/TLS redirect ownership, bounded response, no automatic routing and cached replay.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"context"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"strings"
	"testing"
	"time"
)

func redirectBody() *pdu.ErrorAnswer {
	return &pdu.ErrorAnswer{ResultCode: 3006, RedirectHost: []string{"aaa://new.example;transport=tcp", "aaas://secure.example:5658;transport=sctp;protocol=diameter"}, RedirectHostUsage: avp.Some(uint32(6)), RedirectMaxCacheTime: avp.Some(uint32(3600))}
}

func TestRedirectRouteDoesNotFollow(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			for _, agent := range []bool{false, true} {
				t.Run(fmt.Sprintf("mode-%d/agent-%t", mode, agent), func(t *testing.T) {
					n, co := newInboundNode(t, routeConfig(t))
					_, a := duplicateConnect(t, n, co, encrypted, "dra")
					_, b := duplicateConnect(t, n, co, encrypted, "dra2")
					r := newRouteTest(t, n, mode)
					port := portTest(t, n, 1)
					id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
					if e != nil {
						t.Fatal(e)
					}
					req := routeRead(t, co, a)
					origin := serverOrigin
					if agent {
						origin = draOrigin
					}
					ans, e := co.CompleteAnswer(req, redirectBody(), origin)
					if e != nil {
						t.Fatal(e)
					}
					if e = writeWire(ctxTest(t), a, co, ans); e != nil {
						t.Fatal(e)
					}
					got := routeCompletion(t, n, port, id)
					if got.Err != nil || got.Outcome != AnswerReceived || len(got.Attempts) != 1 || got.Answer.Body.(*pdu.ErrorAnswer).RedirectHost[0] != redirectBody().RedirectHost[0] {
						t.Fatal(got)
					}
					// A new explicit submission still follows the configured route, not the hint.
					id, e = r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
					if e != nil {
						t.Fatal(e)
					}
					next := routeRead(t, co, a)
					routeReply(t, co, a, next)
					if got = routeCompletion(t, n, port, id); got.Outcome != AnswerReceived {
						t.Fatal(got)
					}
					// Idle read deadlines can retire a framed transport; probe only after
					// the second explicit exchange, when no further use is required.
					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
					defer cancel()
					if _, e = b.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
						t.Fatal("automatic alternate request", e)
					}
					ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
					defer cancel2()
					if _, e = a.ReadPrefix(ctx2, time.Second); !errors.Is(e, context.DeadlineExceeded) {
						t.Fatal("automatic original retry", e)
					}
					if e = port.Close(ctxTest(t)); e != nil {
						t.Fatal(e)
					}
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}

func TestRedirectInboundCache(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for app, name := range duplicateProfileNames {
			t.Run(name, func(t *testing.T) {
				n, co := newInboundNode(t, mixedDuplicateConfig(t, true))
				p, a := crossDuplicateConnect(t, n, co, encrypted, "dra")
				p2, b := crossDuplicateConnect(t, n, co, encrypted, "dra2")
				req := mixedDuplicateRequest(t, n, co, app, 10101, 20101, gmscOrigin)
				combinedSend(t, co, a, req)
				in, e := p.Receive(ctxTest(t))
				if e != nil {
					t.Fatal(e)
				}
				mark := markDuplicate(n, in)
				before := n.Resources()
				dup := mixedDuplicateArrival(t, co, req, 10102)
				combinedSend(t, co, b, dup)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
				noDuplicateDelivery(t, p2)
				body := redirectBody()
				result, e := in.Respond(ctxTest(t), body)
				if e != nil || !result.Committed {
					t.Fatal(result, e)
				}
				body.RedirectHost[0] = "aaa://mutated.example"
				body.RedirectHostUsage.Value = 0
				body.RedirectMaxCacheTime.Value = 0
				check := func(m pdu.Message) {
					t.Helper()
					v, ok := m.Body.(*pdu.ErrorAnswer)
					if !ok || v.ResultCode != 3006 || v.RedirectHost[0] != redirectBody().RedirectHost[0] || v.RedirectHostUsage.Value != 6 || v.RedirectMaxCacheTime.Value != 3600 {
						t.Fatal("redirect snapshot changed", m)
					}
				}
				check(duplicateAnswer(t, n, co, a, req))
				check(duplicateAnswer(t, n, co, b, dup))
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				hit := mixedDuplicateArrival(t, co, req, 10103)
				combinedSend(t, co, b, hit)
				check(duplicateAnswer(t, n, co, b, hit))
				assertDuplicateMark(t, n, mark, true)
				noDuplicateDelivery(t, p2)
				if s := n.Resources(); s.DuplicateRecords != 1 || s.DuplicateBytes != before.DuplicateBytes {
					t.Fatal(s)
				}
				// Redirect cache hints never alter the configured duplicate retention clock.
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRedirectRespondFailureIsUncommitted(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"missing-host", "missing-ttl", "uri", "usage", "repeat", "reply"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				in, req := receiveInbound(t, n, p, remote, co, 10111)
				bad := redirectBody()
				switch kind {
				case "missing-host":
					bad.RedirectHost = nil
				case "missing-ttl":
					bad.RedirectMaxCacheTime.Present = false
				case "uri":
					bad.RedirectHost[0] = "https://bad.example"
				case "usage":
					bad.RedirectHostUsage.Value = 7
				case "repeat":
					bad.RedirectHost = make([]string, 17)
					for i := range bad.RedirectHost {
						bad.RedirectHost[i] = "aaa://h.example"
					}
				case "reply":
					bad.RedirectHost = make([]string, 8)
					for i := range bad.RedirectHost {
						bad.RedirectHost[i] = "aaa://" + strings.Repeat("a", 60) + "." + strings.Repeat("b", 60) + ".example"
					}
				}
				before := n.Resources()
				result, e := in.Respond(ctxTest(t), bad)
				if e == nil || result.Committed {
					t.Fatal("invalid redirect committed", result, e)
				}
				if s := n.Resources(); s.InboundOutstanding != 1 || s.InboundBytes != before.InboundBytes {
					t.Fatal("changed live reservation", s)
				}
				if result, e = in.Respond(ctxTest(t), redirectBody()); e != nil || !result.Committed {
					t.Fatal(result, e)
				}
				ans := duplicateAnswer(t, n, co, remote, req)
				if ans.Body.(*pdu.ErrorAnswer).ResultCode != 3006 {
					t.Fatal(ans)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRedirectInvalidAnswerStops(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"uri", "missing-ttl", "source", "repeat", "reply"} {
			t.Run(kind, func(t *testing.T) {
				n, co := newInboundNode(t, routeConfig(t))
				_, a := duplicateConnect(t, n, co, encrypted, "dra")
				_, b := duplicateConnect(t, n, co, encrypted, "dra2")
				r := newRouteTest(t, n, TransportFailover)
				port := portTest(t, n, 1)
				id, e := r.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				req := routeRead(t, co, a)
				sender, e := codec.New(n.cfg.Bundle, codec.Config{})
				if e != nil {
					t.Fatal(e)
				}
				body := redirectBody()
				origin := serverOrigin
				if kind == "source" {
					origin.Host = "untrusted.example"
				}
				if kind == "repeat" {
					body.RedirectHost = make([]string, 17)
					for i := range body.RedirectHost {
						body.RedirectHost[i] = "aaa://h.example"
					}
				}
				if kind == "reply" {
					body.RedirectHost = make([]string, 8)
					for i := range body.RedirectHost {
						body.RedirectHost[i] = "aaa://" + strings.Repeat("a", 60) + "." + strings.Repeat("b", 60) + ".example"
					}
				}
				ans, e := sender.CompleteAnswer(req, body, origin)
				if e != nil {
					t.Fatal(e)
				}
				raw := reqBytes(t, sender, ans)
				if kind == "uri" || kind == "missing-ttl" {
					d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
					out := append([]byte(nil), raw[:20]...)
					for {
						v, ok, e := d.Next()
						if e != nil {
							t.Fatal(e)
						}
						if !ok {
							break
						}
						size := (8 + len(v.Data) + 3) &^ 3
						if kind == "missing-ttl" && v.Key == (avp.Key{Code: 262}) {
							continue
						}
						part := append([]byte(nil), raw[20+v.Offset:20+v.Offset+size]...)
						if kind == "uri" && v.Key == (avp.Key{Code: 292}) {
							part[8] = 'h'
						}
						out = append(out, part...)
					}
					raw = out
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				sendRejectRaw(t, a, raw)
				got := routeCompletion(t, n, port, id)
				if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil || len(got.Attempts) != 1 || got.StopReason != StopProtocol {
					t.Fatal(got)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()
				if _, e = b.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("invalid redirect followed", e)
				}
				if e = port.Close(ctxTest(t)); e != nil {
					t.Fatal(e)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}
