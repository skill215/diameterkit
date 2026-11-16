package diameter

import (
	"context"
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

func imsHSSAlertBody() *s6c.ALR {
	b := alertBody()
	b.UserIdentifier = s6c.AlertUserIdentifier{MSISDN: avp.Some([]byte{0, 0, 0, 0, 0, 0, 0, 0xf0})}
	b.SMSMICorrelationID = avp.Some(s6c.AlertSMSMICorrelationID{DestinationSIPURI: "sip:bob@example.net"})
	return b
}
func imsHSSAlertContext(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	ctx, err := s6c.IMSHSSAlert(s6c.IMSHSSAlertEvidence{Requester: requester, Responder: responder, DestinationSIPURI: "sip:bob@example.net", SCAddress: "12345"})
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}
func imsHSSAlertRequest(t *testing.T, n *Node, co *codec.Codec, hbh uint32) pdu.Message {
	t.Helper()
	r, err := co.CompleteRequest(imsHSSAlertBody(), pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 1000})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func imsHSSAlertReceive(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r := imsHSSAlertRequest(t, n, co, hbh)
	if err := writeWire(ctxTest(t), remote, co, r); err != nil {
		t.Fatal(err)
	}
	in, err := p.Receive(ctxTest(t))
	if err != nil {
		t.Fatal(err)
	}
	if in.Validation().State != validation.NeedsApplication {
		t.Fatal(in.Validation())
	}
	return in, r
}
func TestS6cIMSAlertExchangesAndAdmission(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, ctx := range []validation.Context{{}, imsHSSAlertContext(t, n.cfg.Origin, draOrigin), imsDeliveryContext(t, n.cfg.Origin, iwmscOrigin), t4Context(t, true, n.cfg.Origin, iwmscOrigin)} {
			if _, err := p.Submit(ctxTest(t), imsHSSAlertBody(), RequestOptions{Validation: ctx}, port); err == nil {
				t.Fatal("unauthorized alert")
			}
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		for _, tm := range []avp.Optional[avp.Time]{{}, avp.Some(avp.Time(0)), avp.Some(avp.Time(0xffffffff))} {
			for _, answer := range []pdu.Body{reportAnswer(true, 2001), reportAnswer(true, 5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
				b := imsHSSAlertBody()
				b.MaximumUEAvailabilityTime = tm
				id, err := p.Submit(ctxTest(t), b, RequestOptions{Validation: imsHSSAlertContext(t, n.cfg.Origin, iwmscOrigin)}, port)
				if err != nil {
					t.Fatal(err)
				}
				b.UserIdentifier.MSISDN.Value[0] = 1
				b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:edited@example.net"
				r := routeRead(t, co, remote)
				rb := r.Body.(*s6c.ALR)
				if rb.MaximumUEAvailabilityTime != tm || rb.UserIdentifier.MSISDN.Value[0] != 0 || rb.SMSMICorrelationID.Value.DestinationSIPURI != "sip:bob@example.net" {
					t.Fatal(rb)
				}
				origin := iwmscOrigin
				if _, ok := answer.(*pdu.ErrorAnswer); ok {
					origin = draOrigin
				}
				a, err := co.CompleteAnswer(r, answer, origin)
				if err != nil {
					t.Fatal(err)
				}
				if err = writeWire(ctxTest(t), remote, co, a); err != nil {
					t.Fatal(err)
				}
				got := routeCompletion(t, n, port, id)
				if got.Outcome != AnswerReceived || !got.Validation.Complete {
					t.Fatal(got)
				}
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err := remote.ReadPrefix(ctx, time.Second); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("automatic retry or SMS", err)
		}
	})
}
func TestS6cIMSAlertInboundFreezeBudgetAndRetry(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := reportAlertConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := imsHSSAlertReceive(t, n, p, co, remote, 33001)
		if r, err := in.Respond(ctxTest(t), reportAnswer(true, 2001)); r.Committed || !errors.Is(err, validation.ErrContextRequired) {
			t.Fatal(r, err)
		}
		role, err := s6c.AlertRole(s6c.AlertRoleEvidence{Role: s6c.AlertFromHSS, Requester: serverOrigin, Responder: n.cfg.Origin})
		if err != nil {
			t.Fatal(err)
		}
		for _, ctx := range []validation.Context{role, imsDeliveryContext(t, serverOrigin, n.cfg.Origin), imsHSSAlertContext(t, draOrigin, n.cfg.Origin), t4Context(t, true, serverOrigin, n.cfg.Origin)} {
			if r := in.Check(ctx); r.Complete {
				t.Fatal(r)
			}
		}
		blocked := imsHSSAlertRequest(t, n, co, 33002)
		if err := writeWire(ctxTest(t), remote, co, blocked); err != nil {
			t.Fatal(err)
		}
		if a := duplicateAnswer(t, n, co, remote, blocked).Body.(*pdu.ErrorAnswer); a.ResultCode != 3004 {
			t.Fatal(a)
		}
		b := in.Message().Body.(*s6c.ALR)
		b.UserIdentifier.MSISDN.Value[0] = 1
		b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:edited@example.net"
		b.SCAddress[0] = 0x99
		ctx := imsHSSAlertContext(t, serverOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		if r := in.Check(imsHSSAlertContext(t, draOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("changed frozen evidence")
		}
		if !in.Validation().Complete {
			t.Fatal(in.Validation())
		}
		bad := &s6c.ALA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}
		if r, err := in.Respond(ctxTest(t), bad); err == nil || r.Committed {
			t.Fatal(r, err)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if r, err := in.Respond(canceled, reportAnswer(true, 2001)); r.Committed || !errors.Is(err, context.Canceled) {
			t.Fatal(r, err)
		}
		if r, err := in.Respond(ctxTest(t), reportAnswer(true, 5012)); err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
	})
}
func TestS6cIMSAlertDuplicatesAndURIConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := imsHSSAlertReceive(t, n, p, co, remote, 33100)
		dup, err := co.Snapshot(req)
		if err != nil {
			t.Fatal(err)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*s6c.ALR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{4, 5}}}
		if err = writeWire(ctxTest(t), remote, co, dup); err != nil {
			t.Fatal(err)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r := in.Check(imsHSSAlertContext(t, serverOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		if r, err := in.Respond(ctxTest(t), reportAnswer(true, 2001)); err != nil || !r.Committed {
			t.Fatal(r, err)
		}
		seen := map[uint32]bool{}
		for range 2 {
			a := routeRead(t, co, remote)
			r := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				r = dup
			}
			if seen[a.Header.HopByHop] {
				t.Fatal("same arrival answered twice")
			}
			seen[a.Header.HopByHop] = true
			if err := co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); err != nil {
				t.Fatal(err)
			}
		}
		dup.Header.HopByHop++
		if err = writeWire(ctxTest(t), remote, co, dup); err != nil {
			t.Fatal(err)
		}
		duplicateAnswer(t, n, co, remote, dup)
		for i := range 2 {
			conflict, err := co.Snapshot(req)
			if err != nil {
				t.Fatal(err)
			}
			conflict.Header.HopByHop = uint32(33110 + i)
			b := conflict.Body.(*s6c.ALR)
			if i == 0 {
				b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:%62ob@example.net"
			} else {
				b.MaximumUEAvailabilityTime.Present = false
			}
			if err = writeWire(ctxTest(t), remote, co, conflict); err != nil {
				t.Fatal(err)
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

func TestS6cIMSAlertCancelClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), imsHSSAlertBody(), RequestOptions{Validation: imsHSSAlertContext(t, n.cfg.Origin, iwmscOrigin)}, port)
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
		ans, e := co.CompleteAnswer(req, reportAnswer(true, 2001), iwmscOrigin)
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
		in, _ := imsHSSAlertReceive(t, n, p, co, remote, 33200)
		if r := in.Check(imsHSSAlertContext(t, serverOrigin, n.cfg.Origin)); !r.Complete {
			t.Fatal(r)
		}
		start := make(chan struct{})
		results := make(chan ResponseResult, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r, _ := in.Respond(ctxTest(t), reportAnswer(true, 2001))
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
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}

func TestS6cIMSAlertAndReportSeparatePending(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, reportAlertConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 4)
		ids := map[RequestID]uint32{}
		for _, body := range []pdu.Body{imsDeliveryReportBody(0), imsHSSAlertBody()} {
			ctx := imsDeliveryContext(t, n.cfg.Origin, serverOrigin)
			if body.CommandKey().Code == 8388648 {
				ctx = imsHSSAlertContext(t, n.cfg.Origin, iwmscOrigin)
			}
			id, err := p.Submit(ctxTest(t), body, RequestOptions{Validation: ctx}, port)
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
