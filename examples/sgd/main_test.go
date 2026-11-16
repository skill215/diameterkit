package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func exampleContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func trackPair(t *testing.T, p *pair) *pair {
	t.Helper()
	t.Cleanup(func() {
		if e := p.close(); e != nil {
			t.Error(e)
		}
		for _, ep := range []*endpoint{&p.left, &p.right} {
			if r := ep.node.Resources(); r != (diameter.ResourceSnapshot{}) {
				t.Errorf("retained: %+v", r)
			}
		}
	})
	return p
}
func examplePair(t *testing.T, iface string) *pair {
	t.Helper()
	p, e := openPair(exampleContext(t), iface)
	if e != nil {
		t.Fatal(e)
	}
	return trackPair(t, p)
}
func configs(t *testing.T, iface string) [2]diameter.NodeConfig {
	t.Helper()
	host := "mme.example"
	if iface == "gdd" {
		host = "sgsn.example"
	}
	origins := [2]pdu.Origin{{Host: host, Realm: "example"}, {Host: "sms.example", Realm: "example"}}
	var c [2]diameter.NodeConfig
	for i := range c {
		var e error
		c[i], e = endpointConfig(origins[i], origins[1-i], i == 0)
		if e != nil {
			t.Fatal(e)
		}
	}
	return c
}
func direction(p *pair, mt bool) (*endpoint, *endpoint) {
	if mt {
		return &p.right, &p.left
	}
	return &p.left, &p.right
}
func sessionOf(b pdu.Body) string {
	switch r := b.(type) {
	case *sgd.OFR:
		return r.SessionID
	case *sgd.TFR:
		return r.SessionID
	}
	return ""
}
func requestEdit(b pdu.Body, what string) {
	switch r := b.(type) {
	case *sgd.OFR:
		switch what {
		case "host":
			r.DestinationHost = avp.Some("wrong.example")
		case "realm":
			r.DestinationRealm = "wrong.example"
		case "identity":
			r.UserIdentifier.UserName = avp.Optional[string]{}
		case "tpdu":
			r.SMRPUI = make([]byte, 201)
		case "view":
			r.SessionID = "changed"
		}
	case *sgd.TFR:
		switch what {
		case "host":
			r.DestinationHost = "wrong.example"
		case "realm":
			r.DestinationRealm = "wrong.example"
		case "identity":
			r.UserName = ""
		case "tpdu":
			r.SMRPUI = make([]byte, 201)
		case "view":
			r.SessionID = "changed"
		}
	}
}
func TestSGDPublicExample(t *testing.T) {
	for _, iface := range []string{"sgd", "gdd"} {
		for _, code := range []uint32{2001, 5012, 5555, 3004} {
			t.Run(fmt.Sprintf("%s/%d", iface, code), func(t *testing.T) {
				var out bytes.Buffer
				r, e := run(exampleContext(t), iface, code, &out)
				if e != nil {
					t.Fatal(e)
				}
				if r.Answers != ([2]string{resultName(code), resultName(code)}) || r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
					t.Fatal(r)
				}
				if !strings.Contains(out.String(), iface+" MO result="+resultName(code)+"; MT result="+resultName(code)) {
					t.Fatal(out.String())
				}
			})
		}
	}
}
func TestSGDPublicDirectionAndDestination(t *testing.T) {
	for _, iface := range []string{"sgd", "gdd"} {
		for _, mt := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/MT=%t", iface, mt), func(t *testing.T) {
				p := examplePair(t, iface)
				from, to := direction(p, mt)
				ctx := exampleContext(t)
				wrong, e := newRequest(from.node, !mt, iface == "gdd")
				if e != nil {
					t.Fatal(e)
				}
				if _, e = from.peer.Submit(ctx, wrong, diameter.RequestOptions{}, from.port); !errors.Is(e, diameter.ErrUnsupported) {
					t.Fatal("wrong direction", e)
				}
				for _, kind := range []string{"host", "realm", "identity", "tpdu"} {
					t.Run(kind, func(t *testing.T) {
						b, e := newRequest(from.node, mt, iface == "gdd")
						if e != nil {
							t.Fatal(e)
						}
						requestEdit(b, kind)
						if _, e = from.peer.Submit(ctx, b, diameter.RequestOptions{}, from.port); e == nil {
							t.Fatal("invalid request admitted")
						}
						if r := from.node.Resources(); r.Pending != 0 || r.Outstanding != 0 || r.OutboundBytes != 0 {
							t.Fatal(r)
						}
					})
				}
				quiet, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
				if in, e := to.peer.Receive(quiet); in != nil || !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal(in, e)
				}
				if e := exchange(ctx, from, to, mt, iface == "gdd", 2001); e != nil {
					t.Fatal(e)
				}
			})
		}
	}
}
func TestSGDPublicResponseRules(t *testing.T) {
	for _, iface := range []string{"sgd", "gdd"} {
		for _, mt := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/MT=%t", iface, mt), func(t *testing.T) {
				p := examplePair(t, iface)
				from, to := direction(p, mt)
				ctx := exampleContext(t)
				b, e := newRequest(from.node, mt, iface == "gdd")
				if e != nil {
					t.Fatal(e)
				}
				id, e := from.peer.Submit(ctx, b, diameter.RequestOptions{}, from.port)
				if e != nil {
					t.Fatal(e)
				}
				in, e := to.peer.Receive(ctx)
				if e != nil {
					t.Fatal(e)
				}
				for _, kind := range []string{"wrong command", "missing cause", "wrong direction cause", "mixed namespace"} {
					t.Run(kind, func(t *testing.T) {
						a, e := newAnswer(mt, 5555)
						if e != nil {
							t.Fatal(e)
						}
						if kind == "wrong command" {
							a, e = newAnswer(!mt, 2001)
							if e != nil {
								t.Fatal(e)
							}
						} else {
							var cause *avp.Optional[sgd.SMDeliveryFailureCause]
							var rc *avp.Optional[uint32]
							switch a := a.(type) {
							case *sgd.OFA:
								cause = &a.SMDeliveryFailureCause
								rc = &a.ResultCode
							case *sgd.TFA:
								cause = &a.SMDeliveryFailureCause
								rc = &a.ResultCode
							}
							switch kind {
							case "missing cause":
								*cause = avp.Optional[sgd.SMDeliveryFailureCause]{}
							case "wrong direction cause":
								v := uint32(0)
								if mt {
									v = 3
								}
								*cause = avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: v})
							case "mixed namespace":
								*rc = avp.Some(uint32(5012))
							}
						}
						if r, e := in.Respond(ctx, a); r.Committed || e == nil {
							t.Fatal(r, e)
						}
					})
				}
				if mt && iface == "gdd" {
					a := &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5557})}
					if r, e := in.Respond(ctx, a); r.Committed || e == nil {
						t.Fatal("MME-only result accepted for SGSN", r, e)
					}
				}
				a, e := newAnswer(mt, 5555)
				if e != nil {
					t.Fatal(e)
				}
				if r, e := in.Respond(ctx, a); !r.Committed || e != nil {
					t.Fatal(r, e)
				}
				c, e := from.port.Receive(ctx)
				if e != nil || c.RequestID != id || c.Result.Err != nil || !c.Result.Validation.Complete {
					t.Fatal(c, e)
				}
				if e := checkAnswer(c.Result.Answer.Body, mt, 5555, sessionOf(b)); e != nil {
					t.Fatal(e)
				}
			})
		}
	}
}
func TestSGDPublicTerminalSources(t *testing.T) {
	for _, iface := range []string{"sgd", "gdd"} {
		for _, mt := range []bool{false, true} {
			for _, answerSource := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/MT=%t/answer=%t", iface, mt, answerSource), func(t *testing.T) {
					cfg := configs(t, iface)
					fromIndex, toIndex := 0, 1
					if mt {
						fromIndex, toIndex = 1, 0
					}
					outsider := pdu.Origin{Host: "not-authorized.example", Realm: "example"}
					if answerSource {
						cfg[fromIndex].Peers[0].Commands[0].Sources.Terminals = []pdu.Origin{outsider}
					} else {
						cfg[toIndex].Peers[0].InboundCommands[0].Sources = []pdu.Origin{outsider}
					}
					ctx := exampleContext(t)
					p, e := connectPair(ctx, cfg)
					if e != nil {
						t.Fatal(e)
					}
					trackPair(t, p)
					from, to := direction(p, mt)
					b, e := newRequest(from.node, mt, iface == "gdd")
					if e != nil {
						t.Fatal(e)
					}
					id, e := from.peer.Submit(ctx, b, diameter.RequestOptions{}, from.port)
					if e != nil {
						t.Fatal(e)
					}
					if !answerSource {
						quiet, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
						defer cancel()
						in, e := to.peer.Receive(quiet)
						if in != nil || e == nil {
							t.Fatal("unauthorized source delivered", in, e)
						}
						_ = from.node.Cancel(id)
						return
					}
					in, e := to.peer.Receive(ctx)
					if e != nil {
						t.Fatal(e)
					}
					a, e := newAnswer(mt, 2001)
					if e != nil {
						t.Fatal(e)
					}
					if r, e := in.Respond(ctx, a); !r.Committed || e != nil {
						t.Fatal(r, e)
					}
					c, e := from.port.Receive(ctx)
					if e != nil || c.RequestID != id || !errors.Is(c.Result.Err, validation.ErrInvalidAnswer) || c.Result.Answer.Body != nil {
						t.Fatal(c, e)
					}
				})
			}
		}
	}
}
func TestSGDPublicConfigurationAndNegotiation(t *testing.T) {
	for _, missing := range []int{0, 1} {
		t.Run(fmt.Sprintf("missing capability %d", missing), func(t *testing.T) {
			c := configs(t, "sgd")
			c[missing].Capabilities.VendorSpecificApplicationID = nil
			p, e := connectPair(exampleContext(t), c)
			if p != nil {
				_ = p.close()
			}
			if e == nil {
				t.Fatal("missing capability accepted")
			}
		})
	}
	for _, kind := range []string{"binding", "transport"} {
		t.Run(kind, func(t *testing.T) {
			c := configs(t, "sgd")[0]
			if kind == "binding" {
				c.Peers[0].Commands[0].ApplicationBinding = diameter.SCAP1Accounting
			} else {
				c.Resources.TransportBytes = 1
			}
			n, e := diameter.NewNode(c)
			if n != nil {
				n.Close()
				_ = n.Wait(exampleContext(t))
			}
			if !errors.Is(e, diameter.ErrConfig) {
				t.Fatal(e)
			}
		})
	}
}
func TestSGDPublicCanceledAndInvalidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := run(ctx, "sgd", 2001, io.Discard); e == nil {
		t.Fatal("canceled run succeeded")
	}
	for _, tc := range []struct {
		iface string
		code  uint32
	}{{"bad", 2001}, {"sgd", 2002}} {
		if _, e := run(exampleContext(t), tc.iface, tc.code, io.Discard); e == nil {
			t.Fatal(tc)
		}
	}
}

func TestSGDPublicBudgetLifetimeCancelAndClose(t *testing.T) {
	for _, iface := range []string{"sgd", "gdd"} {
		for _, mt := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/MT=%t", iface, mt), func(t *testing.T) {
				p := examplePair(t, iface)
				from, to := direction(p, mt)
				ctx := exampleContext(t)
				body, e := newRequest(from.node, mt, iface == "gdd")
				if e != nil {
					t.Fatal(e)
				}
				admission, stopAdmission := context.WithCancel(ctx)
				id, e := from.peer.Submit(admission, body, diameter.RequestOptions{}, from.port)
				stopAdmission() // Submit context controls admission, not the accepted request.
				if e != nil {
					t.Fatal(e)
				}
				in, e := to.peer.Receive(ctx)
				if e != nil {
					t.Fatal(e)
				}
				before := from.node.Resources()
				if before.Outstanding != 1 || before.OutboundBytes == 0 || before.Pending != 1 || to.node.Resources().InboundOutstanding != 1 {
					t.Fatal(before, to.node.Resources())
				}
				if _, e = from.peer.Submit(ctx, body, diameter.RequestOptions{}, from.port); !errors.Is(e, diameter.ErrBackpressure) {
					t.Fatal(e)
				}
				after := from.node.Resources()
				if before.OutboundBytes != after.OutboundBytes || after.Outstanding != 1 {
					t.Fatal(before, after)
				}
				waitOnly, stopWait := context.WithTimeout(ctx, 10*time.Millisecond)
				_, e = from.port.Receive(waitOnly)
				stopWait()
				if !errors.Is(e, context.DeadlineExceeded) || from.node.Resources().Outstanding != 1 {
					t.Fatal("completion wait canceled the request", e)
				}
				// The application-visible view is independent of the original response envelope.
				requestEdit(in.Message().Body, "view")
				if r, e := in.Respond(ctx, &pdu.ErrorAnswer{ResultCode: 2001}); r.Committed || e == nil {
					t.Fatal(r, e)
				}
				canceled, cancel := context.WithCancel(ctx)
				cancel()
				answer, e := newAnswer(mt, 2001)
				if e != nil {
					t.Fatal(e)
				}
				if r, e := in.Respond(canceled, answer); r.Committed || !errors.Is(e, context.Canceled) {
					t.Fatal(r, e)
				}
				if r, e := in.Respond(ctx, answer); !r.Committed || e != nil {
					t.Fatal(r, e)
				}
				// Even after responding, an unconsumed completion still occupies the port.
				if _, e = from.peer.Submit(ctx, body, diameter.RequestOptions{}, from.port); !errors.Is(e, diameter.ErrBackpressure) {
					t.Fatal(e)
				}
				completed, e := from.port.Receive(ctx)
				if e != nil || completed.RequestID != id || completed.Result.Err != nil {
					t.Fatal(completed, e)
				}
				if e := checkAnswer(completed.Result.Answer.Body, mt, 2001, sessionOf(body)); e != nil {
					t.Fatal(e)
				}

				if r := from.node.Resources(); r.OutboundBytes != 0 || r.Outstanding != 0 || r.Pending != 0 {
					t.Fatal(r)
				}
				// Bytes independently reject, even with an available completion slot.
				if e = from.port.Close(ctx); e != nil {
					t.Fatal(e)
				}
				from.port, e = from.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 1})
				if e != nil {
					t.Fatal(e)
				}
				if _, e = from.peer.Submit(ctx, body, diameter.RequestOptions{}, from.port); !errors.Is(e, diameter.ErrBackpressure) {
					t.Fatal("byte budget", e)
				}
				if r := from.node.Resources(); r.Outstanding != 0 || r.OutboundBytes != 0 {
					t.Fatal(r)
				}
				quiet, stop := context.WithTimeout(ctx, 20*time.Millisecond)
				unexpected, e := to.peer.Receive(quiet)
				stop()
				if unexpected != nil || !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("rejected submit reached responder", unexpected, e)
				}
				if e = from.port.Close(ctx); e != nil {
					t.Fatal(e)
				}
				from.port, e = from.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
				if e != nil {
					t.Fatal(e)
				}
				body, e = newRequest(from.node, mt, iface == "gdd")
				if e != nil {
					t.Fatal(e)
				}
				id, e = from.peer.Submit(ctx, body, diameter.RequestOptions{}, from.port)
				if e != nil {
					t.Fatal(e)
				}
				in, e = to.peer.Receive(ctx)
				if e != nil {
					t.Fatal(e)
				}
				if e = from.node.Cancel(id); e != nil {
					t.Fatal(e)
				}
				completed, e = from.port.Receive(ctx)
				if e != nil || completed.RequestID != id || !errors.Is(completed.Result.Err, context.Canceled) {
					t.Fatal(completed, e)
				}
				if r, e := in.Respond(ctx, answer); !r.Committed || e != nil {
					t.Fatal(r, e)
				}
				quiet, stop = context.WithTimeout(ctx, 20*time.Millisecond)
				_, e = from.port.Receive(quiet)
				stop()
				if !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("late answer completed twice", e)
				}
				body, e = newRequest(from.node, mt, iface == "gdd")
				if e != nil {
					t.Fatal(e)
				}
				if _, e = from.peer.Submit(ctx, body, diameter.RequestOptions{}, from.port); e != nil {
					t.Fatal(e)
				}
				if _, e = to.peer.Receive(ctx); e != nil {
					t.Fatal(e)
				}
				from.node.Close()
				if e = from.node.Wait(ctx); e != nil {
					t.Fatal(e)
				}
				// Wait joins I/O; it does not consume this retained completion for the caller.
				if r := from.node.Resources(); r.Ports != 1 || r.Outstanding != 1 || r.FixedBytes == 0 || r.OutboundBytes == 0 || r.Pending != 0 {
					t.Fatal(r)
				}
				if e = from.port.Close(ctx); e != nil {
					t.Fatal(e)
				}
				if r := from.node.Resources(); r != (diameter.ResourceSnapshot{}) {
					t.Fatal(r)
				}
			})
		}
	}
}
