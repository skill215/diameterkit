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
	"github.com/skill215/diameterkit/wire"
)

var gmscOrigin = pdu.Origin{Host: "gmsc.example", Realm: "example"}
var iwmscOrigin = pdu.Origin{Host: "iwmsc.example", Realm: "example"}

func reportBody() *s6c.RDR {
	return &s6c.RDR{SessionID: "gmsc.example;report;1", AuthSessionState: 1, UserIdentifier: s6c.ReportUserIdentifier{MSISDN: avp.Some([]byte{0x21, 0x43, 0xf5})}, SCAddress: []byte{0x21, 0x43, 0xf5}, SMDeliveryOutcome: s6c.SMDeliveryOutcome{MMESMDeliveryOutcome: avp.Some(s6c.MMESMDeliveryOutcome{SMDeliveryCause: avp.Some(uint32(0))})}, RDRFlags: avp.Some(uint32(1))}
}
func alertBody() *s6c.ALR {
	return &s6c.ALR{SessionID: "hss.example;alert;1", AuthSessionState: 1, UserIdentifier: s6c.AlertUserIdentifier{MSISDN: avp.Some([]byte{0x21, 0x43, 0xf5})}, SCAddress: []byte{0x21, 0x43, 0xf5}, MaximumUEAvailabilityTime: avp.Some(avp.Time(0))}
}
func reportAlertConfig(t *testing.T, cache bool) NodeConfig {
	c := rejectConfig(t)
	if cache {
		c = duplicateConfig(t)
	}
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	c.Peers[0].Commands = nil
	c.Peers[0].InboundCommands = nil
	for _, body := range []pdu.Body{reportBody(), alertBody()} {
		p, q := out, in
		p.Command = body.CommandKey()
		q.Command = p.Command
		p.DestinationHost = serverOrigin.Host
		p.Sources = pdu.AnswerSources{Terminals: []pdu.Origin{serverOrigin}, ErrorAgents: []pdu.Origin{draOrigin}}
		q.Sources = []pdu.Origin{gmscOrigin}
		if p.Command.Code == 8388648 {
			p.DestinationHost = iwmscOrigin.Host
			p.Sources.Terminals = []pdu.Origin{iwmscOrigin}
			q.Sources = []pdu.Origin{serverOrigin}
		}
		c.Peers[0].Commands = append(c.Peers[0].Commands, p)
		c.Peers[0].InboundCommands = append(c.Peers[0].InboundCommands, q)
	}
	return c
}
func reportAnswer(alert bool, code uint32) pdu.Body {
	if alert {
		return &s6c.ALA{AuthSessionState: 1, ResultCode: avp.Some(code)}
	}
	return &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(code)}
}
func reportRequest(t *testing.T, n *Node, co *codec.Codec, alert bool, hbh uint32) pdu.Message {
	t.Helper()
	var body pdu.Body = reportBody()
	origin := gmscOrigin
	if alert {
		body = alertBody()
		origin = serverOrigin
	}
	req, err := co.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: origin.Host, OriginRealm: origin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 10000})
	if err != nil {
		t.Fatal(err)
	}
	return req
}
func sendReportRequest(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, alert bool, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	req := reportRequest(t, n, co, alert, hbh)
	if err := writeWire(ctxTest(t), remote, co, req); err != nil {
		t.Fatal(err)
	}
	in, err := p.Receive(ctxTest(t))
	if err != nil {
		t.Fatal(err)
	}
	return in, req
}

func TestS6cReportAlertIndependentExchanges(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 4)
		ids := map[RequestID]uint32{}
		for _, body := range []pdu.Body{reportBody(), alertBody()} {
			id, err := p.Submit(ctxTest(t), body, RequestOptions{}, port)
			if err != nil {
				t.Fatal(err)
			}
			ids[id] = body.CommandKey().Code
		}
		reqs := map[uint32]pdu.Message{}
		for range 2 {
			r := routeRead(t, co, remote)
			reqs[r.Header.CommandCode] = r
		}
		if reqs[8388648].Header.HopByHop == reqs[8388649].Header.HopByHop || reqs[8388648].Header.EndToEnd == reqs[8388649].Header.EndToEnd {
			t.Fatal("shared exchange identifiers")
		}
		rdr := reqs[8388649]
		rda, err := co.CompleteAnswer(rdr, reportAnswer(false, 2001), serverOrigin)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := co.Append(nil, rda)
		if err != nil {
			t.Fatal(err)
		}
		wire.Put24(raw[5:8], 8388648) // Correct RDR HBH/E2E, wrong command: discard, not completion.
		sendRejectRaw(t, remote, raw)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, err = port.Receive(ctx)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		// Answer alert first: RDA/ALA do not share pending or imply future activity.
		ala, err := co.CompleteAnswer(reqs[8388648], reportAnswer(true, 2001), iwmscOrigin)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range []pdu.Message{ala, rda} {
			if err = writeWire(ctxTest(t), remote, co, a); err != nil {
				t.Fatal(err)
			}
		}
		for range 2 {
			c, e := port.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			want, ok := ids[c.RequestID]
			if !ok || c.Result.Answer.Header.CommandCode != want || c.Result.Outcome != AnswerReceived || !c.Result.Validation.Complete {
				t.Fatal(c)
			}
			delete(ids, c.RequestID)
		}
		if len(ids) != 0 {
			t.Fatal(ids)
		}
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, err = remote.ReadPrefix(ctx, time.Second)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("unexpected automatic follow-up", err)
		}
	})
}

func TestS6cReportAlertInboundBranches(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, tc := range []struct {
			alert bool
			body  pdu.Body
		}{
			{false, &s6c.RDA{AuthSessionState: 1, ResultCode: avp.Some(uint32(2001)), UserIdentifier: avp.Some(s6c.UserIdentifier{MSISDN: []byte{0x21, 0xf6}})}},
			{false, &s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}},
			{false, &s6c.RDA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5558})}},
			{false, reportAnswer(false, 5012)}, {false, &pdu.ErrorAnswer{ResultCode: 3004}},
			{true, reportAnswer(true, 2001)}, {true, reportAnswer(true, 5012)}, {true, &pdu.ErrorAnswer{ResultCode: 3004}},
		} {
			in, req := sendReportRequest(t, n, p, co, remote, tc.alert, uint32(700+i))
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			if r, err := in.Respond(ctxTest(t), tc.body); err != nil || !r.Committed {
				t.Fatal(r, err)
			}
			ans := routeRead(t, co, remote)
			if err := co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
				t.Fatal(err)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}

func TestS6cRDRReceiveOnlyFlagsRespond(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		req := reportRequest(t, n, co, false, 720)
		raw, err := co.Append(nil, req)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for off := 20; off < len(raw); {
			size := int(binary.BigEndian.Uint32(raw[off+4:off+8]) & 0xffffff)
			if binary.BigEndian.Uint32(raw[off:off+4]) == 3323 {
				binary.BigEndian.PutUint32(raw[off+12:], 0x80000001)
				found = true
				break
			}
			off += (size + 3) &^ 3
		}
		if !found {
			t.Fatal("flags missing")
		}
		sendRejectRaw(t, remote, raw)
		in, err := p.Receive(ctxTest(t))
		if err != nil {
			t.Fatal(err)
		}
		if in.Message().Body.(*s6c.RDR).RDRFlags.Value != 0x80000001 {
			t.Fatal("lost receive flags")
		}
		in.Message().Body.(*s6c.RDR).SMDeliveryOutcome = s6c.SMDeliveryOutcome{}
		in.Message().Body.(*s6c.RDR).SessionID = "mutated"
		if !in.Check(validation.Context{}).Complete {
			t.Fatal(in.Validation())
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err = in.Respond(ctx, reportAnswer(false, 2001)); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err = in.Respond(ctxTest(t), reportAnswer(true, 2001)); err == nil {
			t.Fatal("wrong body committed")
		}
		if r, e := in.Respond(ctxTest(t), reportAnswer(false, 2001)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		ans := routeRead(t, co, remote)
		if err = co.ValidateAnswer(req, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestS6cReportAlertDuplicateArrivals(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, alert := range []bool{false, true} {
			in, req := sendReportRequest(t, n, p, co, remote, alert, uint32(730+i*10))
			dup, err := co.Snapshot(req)
			if err != nil {
				t.Fatal(err)
			}
			dup.Header.HopByHop++
			dup.Header.Flags |= pdu.RetransmittedFlag
			proxies := []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{1, 2, 3}}}
			if alert {
				dup.Body.(*s6c.ALR).ProxyInfo = proxies
			} else {
				dup.Body.(*s6c.RDR).ProxyInfo = proxies
			}
			if err = writeWire(ctxTest(t), remote, co, dup); err != nil {
				t.Fatal(err)
			}
			noDuplicateDelivery(t, p)
			if r, e := in.Respond(ctxTest(t), reportAnswer(alert, 2001)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			for range 2 {
				ans := routeRead(t, co, remote)
				original := req
				if ans.Header.HopByHop == dup.Header.HopByHop {
					original = dup
				}
				if err = co.ValidateAnswer(original, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
					t.Fatal(err)
				}
			}
			dup.Header.HopByHop++
			if err = writeWire(ctxTest(t), remote, co, dup); err != nil {
				t.Fatal(err)
			}
			duplicateAnswer(t, n, co, remote, dup)
			noDuplicateDelivery(t, p)
			// Changing a nested identity under the same E2E is a content conflict.
			if alert {
				dup.Body.(*s6c.ALR).UserIdentifier.MSISDN.Value[0] ^= 1
			} else {
				dup.Body.(*s6c.RDR).UserIdentifier.MSISDN.Value[0] ^= 1
			}
			dup.Header.HopByHop++
			if err = writeWire(ctxTest(t), remote, co, dup); err != nil {
				t.Fatal(err)
			}
			conflict := routeRead(t, co, remote)
			if a, ok := conflict.Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 5012 {
				t.Fatal(conflict)
			}
			noDuplicateDelivery(t, p)
		}
	})
}

func TestS6cReportAlertBudgetAndReject(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := reportAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, _ := sendReportRequest(t, n, p, co, remote, false, 760)
		req := reportRequest(t, n, co, true, 761)
		if err := writeWire(ctxTest(t), remote, co, req); err != nil {
			t.Fatal(err)
		}
		ans := duplicateAnswer(t, n, co, remote, req)
		if a, ok := ans.Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(ans)
		}
		if !in.Validation().Complete || n.Resources().InboundOutstanding != 1 {
			t.Fatal(n.Resources())
		}
		if r, err := in.Reject(ctxTest(t), FailureContextUnavailable); err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		a := routeRead(t, co, remote)
		if a.Body.(*pdu.ErrorAnswer).ResultCode != 5012 {
			t.Fatal(a)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		in, _ = sendReportRequest(t, n, p, co, remote, true, 762)
		if r, err := in.Respond(ctxTest(t), reportAnswer(true, 2001)); err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		routeRead(t, co, remote)
	})
}

func TestS6cReportAlertCloseRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, alert := range []bool{false, true} {
			n, co := newInboundNode(t, reportAlertConfig(t, false))
			p, remote := networkConnectNode(t, n, co, tls)
			in, _ := sendReportRequest(t, n, p, co, remote, alert, 770)
			start := make(chan struct{})
			results := make(chan ResponseResult, 2)
			var wg sync.WaitGroup
			for range 2 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					r, _ := in.Respond(ctxTest(t), reportAnswer(alert, 2001))
					results <- r
				}()
			}
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
			if err := n.Wait(ctxTest(t)); err != nil {
				t.Fatal(err)
			}
			if n.Resources().InboundOutstanding != 0 {
				t.Fatal(n.Resources())
			}
		}
	})
}

func TestS6cReportAlertNoTransportReplay(t *testing.T) {
	cfg := twoDuplicatePeers(reportAlertConfig(t, false))
	n, _ := newInboundNode(t, cfg)
	for _, body := range []pdu.Body{reportBody(), alertBody()} {
		for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			target := serverOrigin
			if body.CommandKey().Code == 8388648 {
				target = iwmscOrigin
			}
			_, err := n.NewRoute(RouteConfig{Command: body.CommandKey(), Peers: []PeerID{"dra", "dra2"}, ExchangePolicy: mode, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: target, Retention: 2 * time.Second}})
			if err == nil {
				t.Fatal("unaudited replay enabled")
			}
		}
	}
}
