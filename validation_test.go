package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func TestInboundValidationPrivateEvidence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 81)
		before := n.Resources()
		r := in.Validation()
		if !r.Complete || r.State != validation.Ready || len(r.CheckedRules) != 3 {
			t.Fatal(r)
		}
		in.Message().Body.(*s6c.SRR).MSISDN = avp.Optional[[]byte]{}
		r.CheckedRules[0] = "caller mutation"
		r.Complete = false
		var wg sync.WaitGroup
		for range 16 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 20 {
					r := in.Check(validation.Context{})
					if !r.Complete || r.State != validation.Ready || r.CheckedRules[0] != "WIRE-STATIC" {
						t.Error(r)
					}
					r.CheckedRules[0] = "independent"
				}
			}()
		}
		wg.Wait()
		if after := n.Resources(); after.InboundBytes != before.InboundBytes || after.InboundOutstanding != before.InboundOutstanding {
			t.Fatal(after, before)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := in.Respond(ctx, inboundFailure()); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if !in.Validation().Complete {
			t.Fatal("cancelled preparation changed verdict")
		}
		if _, err := in.Respond(ctxTest(t), &s6c.SRA{}); err == nil {
			t.Fatal("invalid answer committed")
		}
		if !in.Validation().Complete {
			t.Fatal("bad answer changed request verdict")
		}
		if _, err := in.Respond(ctxTest(t), inboundFailure()); err != nil {
			t.Fatal(err)
		}
		ans := routeRead(t, co, remote)
		if err := co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
		if !in.Validation().Complete {
			t.Fatal("commit erased historical verdict")
		}
	})
}

func TestInboundValidationExpiryClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, closeNode := range []bool{false, true} {
			cfg := inboundConfig(t)
			cfg.Peers[0].InboundCommands[0].ExecutionTimeout = 40 * time.Millisecond
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, tls)
			in, _ := receiveInbound(t, n, p, remote, co, 82)
			if closeNode {
				var wg sync.WaitGroup
				start := make(chan struct{})
				for range 8 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-start
						for range 20 {
							r := in.Validation()
							if r.State != validation.Ready && r.State != validation.Expired {
								t.Error(r)
							}
						}
					}()
				}
				close(start)
				_ = n.Close()
				wg.Wait()
			} else {
				select {
				case <-in.Done():
				case <-time.After(time.Second):
					t.Fatal("expiry")
				}
			}
			r := in.Check(validation.Context{})
			if r.Complete || r.State != validation.Expired || len(r.LocalErrors) != 1 || !errors.Is(r.LocalErrors[0], validation.ErrExpired) {
				t.Fatal(r)
			}
			if _, err := in.Respond(ctxTest(t), inboundFailure()); err == nil {
				t.Fatal("expired response")
			}
		}
	})
}

func TestRuntimeAnswerValidationReports(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, inboundConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		for _, kind := range []string{"normal", "E", "session", "cancel"} {
			id, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{Validation: validation.Context{}}, port)
			if err != nil {
				t.Fatal(err)
			}
			req := routeRead(t, co, remote)
			if kind == "cancel" {
				_ = n.Cancel(id)
			} else {
				ans, e := answerFor(co, req)
				if e != nil {
					t.Fatal(e)
				}
				if kind == "E" {
					ans, e = co.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3002}, draOrigin)
					if e != nil {
						t.Fatal(e)
					}
				}
				if kind == "session" {
					ans.Body.(*s6c.SRA).SessionID = "another-session"
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
			}
			c, e := port.Receive(ctxTest(t))
			if e != nil || c.RequestID != id {
				t.Fatal(c, e)
			}
			r := c.Result.Validation
			switch kind {
			case "normal", "E":
				if !r.Complete || r.State != validation.Ready || c.Result.Answer.Body == nil || c.Result.Err != nil || len(r.CheckedRules) != 3 {
					t.Fatal(c)
				}
			case "session":
				if r.Complete || r.State != validation.Rejected || len(r.CheckedRules) != 1 || len(r.Violations) != 1 || !errors.Is(r.Violations[0], validation.ErrEnvelope) || c.Result.Answer.Body != nil || !errors.Is(c.Result.Err, validation.ErrInvalidAnswer) {
					t.Fatal(c)
				}
			case "cancel":
				if r.State != validation.Unchecked || r.Complete || c.Result.Answer.Body != nil {
					t.Fatal(c)
				}
			}
		}
	})
}

func unauditedConfig(t *testing.T, cfg NodeConfig, request bool) NodeConfig {
	t.Helper()
	sp, _ := s6c.Profile()
	sb, _ := schema.Compile(sp)
	req, _ := sb.Lookup((&s6c.SRR{}).CommandKey())
	ans, _ := sb.Lookup((&s6c.SRA{}).CommandKey())
	if request {
		req.NoContext = false
		req.CheckContext = nil
		req.ContextName = ""
	} else {
		ans.NoContext = false
		ans.CheckContext = nil
		ans.ContextName = ""
	}
	sp, err := schema.NewProfile(sp.Name(), sp.Version(), sp.Digest(), req, ans)
	if err != nil {
		t.Fatal(err)
	}
	bp, _ := base.Profile()
	cfg.Bundle, err = schema.Compile(bp, sp)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestRuntimeContextAuditAdmission(t *testing.T) {
	if n, err := NewNode(unauditedConfig(t, inboundConfig(t), true)); err == nil {
		_ = n.Close()
		t.Fatal("unaudited inbound accepted")
	}
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, unauditedConfig(t, testConfig(t), true))
		p, _ := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		before := n.Resources()
		id, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if id != (RequestID{}) || !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(id, err)
		}
		after := n.Resources()
		if after.Outstanding != before.Outstanding || after.OutboundBytes != before.OutboundBytes {
			t.Fatal(after, before)
		}
	})
}

func TestRuntimeUnsupportedAnswerReport(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, unauditedConfig(t, inboundConfig(t), false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		if _, err := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); err != nil {
			t.Fatal(err)
		}
		req := routeRead(t, co, remote)
		routeReply(t, co, remote, req)
		c, err := port.Receive(ctxTest(t))
		if err != nil {
			t.Fatal(err)
		}
		r := c.Result.Validation
		if c.Result.Err != nil || c.Result.Outcome != AnswerReceived || c.Result.Answer.Body == nil || r.Complete || len(r.LocalErrors) != 1 || !errors.Is(r.LocalErrors[0], validation.ErrUnsupported) || len(r.MissingContext) != 0 {
			t.Fatal(c)
		}
		// Incoming request can be checked, but an unaudited normal answer cannot commit.
		in, _ := receiveInbound(t, n, p, remote, co, 83)
		if _, err = in.Respond(ctxTest(t), inboundFailure()); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
		if !in.Validation().Complete {
			t.Fatal("failed Respond consumed verdict")
		}
	})
}
