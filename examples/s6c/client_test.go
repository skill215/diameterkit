package main

import (
	"context"
	"errors"
	diameter "github.com/skill215/diameterkit"
	app "github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	client "github.com/skill215/diameterkit/client/s6c"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func TestRoutingClient(t *testing.T) {
	for _, async := range []bool{false, true} {
		for _, code := range []uint32{2001, 5012, 5001, 3004} {
			t.Run(fmtClientName(async, code), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				p, e := openPair(ctx)
				if e != nil {
					t.Fatal(e)
				}
				defer func() {
					if e := p.close(); e != nil {
						t.Error(e)
					}
					if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
						t.Error("resource leak")
					}
				}()
				route, e := p.left.node.NewRoute(diameter.RouteConfig{Command: (&app.SRR{}).CommandKey(), Peers: []diameter.PeerID{remotePeer}, ExchangePolicy: diameter.SingleAttempt, MaxAttempts: 1})
				if e != nil {
					t.Fatal(e)
				}
				if !async {
					if e := p.left.port.Close(ctx); e != nil {
						t.Fatal(e)
					}
				}
				c, e := client.New(route)
				if e != nil {
					t.Fatal(e)
				}
				session, e := p.left.node.NewSessionID()
				if e != nil {
					t.Fatal(e)
				}
				q := app.RoutingQuery{SessionID: session, MSISDN: avp.Some("12025550123"), SCAddress: "12025550100"}
				server := make(chan error, 1)
				go func() {
					in, e := p.right.peer.Receive(ctx)
					if e != nil {
						server <- e
						return
					}
					a, e := newAnswer(code)
					if e == nil {
						_, e = in.Respond(ctx, a)
					}
					server <- e
				}()
				var reply client.RoutingReply
				if async {
					id, e := c.SubmitRouting(ctx, q, diameter.RequestOptions{}, p.left.port)
					if e != nil {
						t.Fatal(e)
					}
					done, e := p.left.port.Receive(ctx)
					if e != nil || done.RequestID != id {
						t.Fatal(done, e)
					}
					r, e := app.NewSRR(q)
					if e != nil {
						t.Fatal(e)
					}
					reply, e = c.InterpretRouting(app.RoutingContext{Request: r}, done.Result)
					if e != nil {
						t.Fatal(e)
					}
				} else {
					reply, e = c.QueryRouting(ctx, q, diameter.RequestOptions{})
					if e != nil {
						t.Fatal(e)
					}
				}
				if e := <-server; e != nil {
					t.Fatal(e)
				}
				if reply.Routing == nil || reply.Exchange.Outcome != diameter.AnswerReceived || reply.Routing.Protocol.Code != code {
					t.Fatal(reply)
				}
				if code == 2001 && (reply.Routing.Outcome != app.RoutingSuccess || len(reply.Routing.Candidates) != 1) {
					t.Fatal(reply)
				}
				// An interpretation failure retains the exact exchange and never submits again.
				bad := &app.SRR{SessionID: "wrong", AuthSessionState: 1, MSISDN: avp.Some([]byte{0x21}), SCAddress: []byte{0x21}}
				failed, e := c.InterpretRouting(app.RoutingContext{Request: bad}, reply.Exchange)
				if !errors.Is(e, validation.ErrInvalidAnswer) || failed.Exchange.Answer.Body != reply.Exchange.Answer.Body || failed.Routing != nil {
					t.Fatal(failed, e)
				}
			})
		}
	}
}
func fmtClientName(async bool, code uint32) string {
	if async {
		return "submit-" + fmtCode(code)
	}
	return "query-" + fmtCode(code)
}
func fmtCode(code uint32) string {
	switch code {
	case 2001:
		return "success"
	case 5012:
		return "base"
	case 5001:
		return "experimental"
	default:
		return "error"
	}
}

func TestRoutingClientCanceledQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p, e := openPair(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := p.close(); e != nil {
			t.Error(e)
		}
		if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
			t.Error("resource leak")
		}
	}()
	if e = p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	route, e := p.left.node.NewRoute(diameter.RouteConfig{Command: (&app.SRR{}).CommandKey(), Peers: []diameter.PeerID{remotePeer}, ExchangePolicy: diameter.SingleAttempt, MaxAttempts: 1})
	if e != nil {
		t.Fatal(e)
	}
	c, e := client.New(route)
	if e != nil {
		t.Fatal(e)
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	reply, e := c.QueryRouting(canceled, app.RoutingQuery{SessionID: "canceled;1", MSISDN: avp.Some("123"), SCAddress: "456"}, diameter.RequestOptions{})
	if !errors.Is(e, context.Canceled) || reply.Routing != nil || reply.Exchange.Outcome != diameter.NotWritten {
		t.Fatal(reply, e)
	}
}
