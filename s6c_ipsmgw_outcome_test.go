package diameter

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
)

func ipDeliveryReportBody(mask int) *s6c.RDR {
	b := reportBody()
	b.SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
	if mask&1 != 0 {
		b.SMDeliveryOutcome.MMESMDeliveryOutcome = avp.Some(s6c.MMESMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(13))})
	}
	if mask&2 != 0 {
		b.SMDeliveryOutcome.MSCSMDeliveryOutcome = avp.Some(s6c.MSCSMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(0))})
	}
	if mask&4 != 0 {
		b.SMDeliveryOutcome.SGSNSMDeliveryOutcome = avp.Some(s6c.SGSNSMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(5))})
	}
	b.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome = avp.Some(s6c.IPSMGWSMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(1)), AbsentUserDiagnosticSM: avp.Some(uint32(11))})
	return b
}
func ipDeliveryReportRequest(t *testing.T, n *Node, co *codec.Codec, mask int, hbh uint32) pdu.Message {
	t.Helper()
	b := ipDeliveryReportBody(mask)
	req, e := co.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 1000})
	if e != nil {
		t.Fatal(e)
	}
	return req
}
func ipDeliveryReportReceive(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, mask int, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req := ipDeliveryReportRequest(t, n, co, mask, hbh)
	if e := writeWire(ctxTest(t), remote, co, req); e != nil {
		t.Fatal(e)
	}
	in, e := p.Receive(ctxTest(t))
	if e != nil {
		t.Fatal(e)
	}
	return in, req
}
func ipDeliveryReportAnswer(result uint32) *s6c.RDA {
	b := &s6c.RDA{AuthSessionState: 1}
	if result == 2001 {
		b.ResultCode = avp.Some(result)
		b.UserIdentifier = avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf6}})
	} else {
		b.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: result})
	}
	return b
}
func TestS6cIPDeliveryReportExchanges(t *testing.T) {
	eachIPReportPurpose(t, func(t *testing.T, tls bool, purpose ipReportPurpose) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		if purpose {
			for _, ctx := range []validation.Context{{}, purpose.context(t, n.cfg.Origin, draOrigin)} {
				if _, err := p.Submit(ctxTest(t), purpose.body(0), RequestOptions{Validation: ctx}, port); err == nil {
					t.Fatal("missing binding")
				}
			}
			if n.Resources().OutboundBytes != 0 || n.Resources().Outstanding != 0 {
				t.Fatal(n.Resources())
			}
		}
		for _, policy := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			if _, err := n.NewRoute(RouteConfig{Command: purpose.body(0).CommandKey(), Peers: []PeerID{p.id}, ExchangePolicy: policy, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Second}}); !errors.Is(err, ErrUnsupported) {
				t.Fatal(err)
			}
		}
		for _, mask := range []int{0, 1, 2, 4, 5, 6} {
			for _, code := range []uint32{2001, 5001, 5558, 5012, 3004} {
				body := purpose.body(mask)
				id, e := p.Submit(ctxTest(t), body, RequestOptions{Validation: purpose.context(t, n.cfg.Origin, serverOrigin)}, port)
				if e != nil {
					t.Fatal(mask, e)
				}
				// Caller changes cannot rewrite the admitted report.
				body.SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
				if purpose {
					body.UserIdentifier.UserName.Value = "001010123456780"
				}
				req := routeRead(t, co, remote)
				if r := co.ValidateContext(req, purpose.context(t, n.cfg.Origin, serverOrigin)); !r.Complete {
					t.Fatal(r)
				}
				if o := req.Body.(*s6c.RDR).SMDeliveryOutcome; !o.IPSMGWSMDeliveryOutcome.Present {
					t.Fatal("request mutation escaped")
				}
				var response pdu.Body = ipDeliveryReportAnswer(code)
				if code == 5012 {
					response = reportAnswer(false, 5012)
				}
				if code == 3004 {
					response = &pdu.ErrorAnswer{ResultCode: 3004}
				}
				ans, e := co.CompleteAnswer(req, response, serverOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
					t.Fatal(e)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || got.Err != nil || !got.Validation.Complete {
					t.Fatal(got)
				}
				if code == 3004 {
					if a, ok := got.Answer.Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
						t.Fatal(a)
					}
				} else {
					b := got.Answer.Body.(*s6c.RDA)
					if code == 2001 {
						if !b.UserIdentifier.Present {
							t.Fatal(b)
						}
					} else if code == 5012 {
						if !b.ResultCode.Present || b.ResultCode.Value != 5012 || b.ExperimentalResult.Present {
							t.Fatal(b)
						}
					} else if b.ExperimentalResult.Value.ExperimentalResultCode != code || b.ResultCode.Present {
						t.Fatal(b)
					}
				}
				if n.Resources().OutboundBytes != 0 {
					t.Fatal(n.Resources())
				}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e := remote.ReadPrefix(ctx, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic alert or retry", e)
		}
	})
}
func TestS6cIPDeliveryReportInboundBudgetAndRetry(t *testing.T) {
	eachIPReportPurpose(t, func(t *testing.T, tls bool, purpose ipReportPurpose) {
		cfg := reportAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := purpose.receive(t, n, p, co, remote, 6, 900)
		purpose.check(t, in, n.cfg.Origin)
		before := n.Resources()
		if before.InboundOutstanding != 1 || before.InboundBytes == 0 {
			t.Fatal(before)
		}
		blocked := purpose.request(t, n, co, 4, 901)
		if e := writeWire(ctxTest(t), remote, co, blocked); e != nil {
			t.Fatal(e)
		}
		if a := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); a.ResultCode != 3004 {
			t.Fatal(a)
		}
		in.Message().Body.(*s6c.RDR).SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
		if !in.Check(purpose.context(t, gmscOrigin, n.cfg.Origin)).Complete {
			t.Fatal(in.Validation())
		}
		bad := ipDeliveryReportAnswer(5558)
		bad.UserIdentifier = ipDeliveryReportAnswer(2001).UserIdentifier
		if r, e := in.Respond(ctxTest(t), bad); e == nil || r.Committed {
			t.Fatal(r, e)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if r, e := in.Respond(ctx, ipDeliveryReportAnswer(2001)); !errors.Is(e, context.Canceled) || r.Committed {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), ipDeliveryReportAnswer(2001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		in, req = purpose.receive(t, n, p, co, remote, 4, 902)
		purpose.check(t, in, n.cfg.Origin)
		if r, e := in.Respond(ctxTest(t), ipDeliveryReportAnswer(5001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
	})
}
func TestS6cIPDeliveryReportDuplicateParents(t *testing.T) {
	eachIPReportPurpose(t, func(t *testing.T, tls bool, purpose ipReportPurpose) {
		n, co := newInboundNode(t, reportAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := purpose.receive(t, n, p, co, remote, 6, 920)
		purpose.check(t, in, n.cfg.Origin)
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*s6c.RDR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{4, 5}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), ipDeliveryReportAnswer(5558)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			r := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				r = dup
			}
			if report := co.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, purpose.context(t, gmscOrigin, n.cfg.Origin)); !report.Complete {
				t.Fatal(report)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*s6c.RDA); a.ExperimentalResult.Value.ExperimentalResultCode != 5558 {
			t.Fatal(a)
		}
		for i := range 4 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(930 + i)
			o := &conflict.Body.(*s6c.RDR).SMDeliveryOutcome
			if i == 0 {
				o.MSCSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Present = false
			} else if i == 1 {
				o.IPSMGWSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value = 12
			} else if i == 2 {
				if purpose {
					conflict.Body.(*s6c.RDR).UserIdentifier.UserName.Value = "001010123456780"
				} else {
					conflict.Body.(*s6c.RDR).UserIdentifier.MSISDN.Value[0] = 0x43
				}
			} else {
				conflict.Body.(*s6c.RDR).SCAddress[0] = 0x43
			}
			if e = writeWire(ctxTest(t), remote, co, conflict); e != nil {
				t.Fatal(e)
			}
			if a, ok := routeRead(t, co, remote).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 5012 {
				t.Fatal(a)
			}
		}
		noDuplicateDelivery(t, p)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		if n.Resources().DuplicateRecords != 1 {
			t.Fatal(n.Resources())
		}
	})
}
func TestS6cIPDeliveryReportCancelClose(t *testing.T) {
	eachIPReportPurpose(t, func(t *testing.T, tls bool, purpose ipReportPurpose) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), purpose.body(6), RequestOptions{Validation: purpose.context(t, n.cfg.Origin, serverOrigin)}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); got.Outcome == AnswerReceived {
			t.Fatal(got)
		}
		ans, e := co.CompleteAnswer(req, ipDeliveryReportAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(ctx)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late completion", e)
		}
		in, _ := purpose.receive(t, n, p, co, remote, 5, 940)
		purpose.check(t, in, n.cfg.Origin)
		start := make(chan struct{})
		results := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r, _ := in.Respond(ctxTest(t), ipDeliveryReportAnswer(2001))
				results <- r
			}()
		}
		wg.Add(1)
		go func() { defer wg.Done(); <-start; in.Check(purpose.context(t, gmscOrigin, n.cfg.Origin)) }()
		close(start)
		_ = n.Close()
		wg.Wait()
		close(results)
		commits := 0
		for r := range results {
			if r.Committed {
				commits++
			}
		}
		if commits > 1 {
			t.Fatal("multiple commits")
		}
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}

func TestS6cIPDeliveryUnknownReceiveAndAdmission(t *testing.T) {
	eachIPReportPurpose(t, func(t *testing.T, tls bool, purpose ipReportPurpose) {
		for i, field := range []uint32{3321, 3322, 3322} {
			n, co := newInboundNode(t, reportAlertConfig(t, false))
			p, remote := networkConnectNode(t, n, co, tls)
			// Local unsupported values fail before admission or writes.
			port := portTest(t, n, 1)
			b := purpose.body(0)
			b.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value.AbsentUserDiagnosticSM.Value = 13
			if _, err := p.Submit(ctxTest(t), b, RequestOptions{Validation: purpose.context(t, n.cfg.Origin, serverOrigin)}, port); !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(err)
			}
			if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
				t.Fatal(n.Resources())
			}
			value := uint32(3)
			if i == 1 {
				value = 0
			}
			if i == 2 {
				value = 13
			}
			req := purpose.request(t, n, co, 0, uint32(31901+i))
			raw, err := co.Append(nil, req)
			if err != nil {
				t.Fatal(err)
			}
			// Project only the remote leaf, keeping valid TLV widths and framing.
			found := false
			for off := 20; off < len(raw); {
				size := int(binary.BigEndian.Uint32(raw[off+4:off+8]) & 0xffffff)
				if binary.BigEndian.Uint32(raw[off:]) == 3316 {
					for node := off + 12; node < off+size; {
						ns := int(binary.BigEndian.Uint32(raw[node+4:node+8]) & 0xffffff)
						if binary.BigEndian.Uint32(raw[node:]) == 3320 {
							for leaf := node + 12; leaf < node+ns; {
								ls := int(binary.BigEndian.Uint32(raw[leaf+4:leaf+8]) & 0xffffff)
								if binary.BigEndian.Uint32(raw[leaf:]) == field {
									binary.BigEndian.PutUint32(raw[leaf+12:], value)
									found = true
								}
								leaf += (ls + 3) &^ 3
							}
						}
						node += (ns + 3) &^ 3
					}
				}
				off += (size + 3) &^ 3
			}
			if !found {
				t.Fatal("missing leaf")
			}
			// Codec receive preserves the raw enum/diagnostic; full validation rejects it.
			decoded, err := co.Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			report := co.ValidateContext(decoded, purpose.context(t, gmscOrigin, n.cfg.Origin))
			if !errors.Is(report.Err(), validation.ErrUnsupported) || report.State != validation.Rejected {
				t.Fatal(report)
			}
			body := decoded.Body.(*s6c.RDR).SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value
			if field == 3321 && body.SMDeliveryCause.Value != value || field == 3322 && (!body.AbsentUserDiagnosticSM.Present || body.AbsentUserDiagnosticSM.Value != value) {
				t.Fatal(body)
			}
			// The existing inbound policy closes on an unsupported complete request;
			// it does not deliver an actionable Inbound or invent a business response.
			sendRejectRaw(t, remote, raw)
			waitFor(t, func() bool { return p.Info().Protocol == ProtocolClosed })
			if info := p.Info(); !errors.Is(info.Cause, validation.ErrUnsupported) || info.Ready {
				t.Fatal(info)
			}
			if _, err := remote.ReadPrefix(ctxTest(t), time.Second); err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("expected peer close without a response", err)
			}
			waitFor(t, func() bool { return n.Resources().InboundBytes == 0 && n.Resources().InboundOutstanding == 0 })
			n.mu.Lock()
			queued := len(p.inboundQueue)
			n.mu.Unlock()
			if queued != 0 {
				t.Fatal("unsupported request delivered", queued)
			}
		}
	})
}
