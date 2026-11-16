package diameter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func scapBody(t *testing.T) *scap1.ACR {
	t.Helper()
	r, e := scap1.NewEvent(scap1.EventInput{SessionID: "ctf.example;scap1;1", SubscriptionMSISDN: avp.Some("15551234567"), Timestamp: avp.Some(avp.Time(3908988800)), DataCodingScheme: avp.Some("8"), OriginatorMSISDN: avp.Some("15551234567"), OriginatorVMSC: avp.Some("15550000001"), TrafficCase: avp.Some("MO"), DestinationMSISDN: avp.Some("15557654321"), ConcatenatedMessages: avp.Some("1"), UserDataLength: avp.Some("140")})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func scapAnswer(code uint32) *scap1.ACA { return &scap1.ACA{ResultCode: code, AccountingRecordType: 1} }
func scapCapability() pdu.VendorSpecificApplicationID {
	return pdu.VendorSpecificApplicationID{VendorID: 193, AcctApplicationID: avp.Some(uint32(19302))}
}
func scapRuntimeConfig(t *testing.T) NodeConfig {
	c := rejectConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	cp, e := scap1.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(bp, sp, cp)
	if e != nil {
		t.Fatal(e)
	}
	c.Capabilities.AuthApplicationID = nil
	c.Capabilities.VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{scapCapability()}
	out, in := c.Peers[0].Commands[0], c.Peers[0].InboundCommands[0]
	out.Command = scapBody(t).CommandKey()
	in.Command = out.Command
	out.ApplicationBinding = SCAP1Accounting
	in.ApplicationBinding = SCAP1Accounting
	out.DestinationHost = serverOrigin.Host
	in.Sources = []pdu.Origin{gmscOrigin}
	c.Peers[0].Commands = []CommandPolicy{out}
	c.Peers[0].InboundCommands = []InboundPolicy{in}
	return c
}
func TestSCAP1RuntimeNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, scapRuntimeConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, body := range []pdu.Body{scapAnswer(2001), scapAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			q := scapBody(t)
			id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.ServiceParameterInfo[2].ParameterValue[0] = 99
			r := routeRead(t, co, remote)
			if r.Body.(*scap1.ACR).ServiceParameterInfo[2].ParameterValue[0] != '1' {
				t.Fatal("submit alias")
			}
			if r.Header.ApplicationID != 0 || r.Header.CommandCode != 271 || r.Header.EndToEnd&0x80000000 != 0 {
				t.Fatal("SCAP header/identifier allocation", r.Header)
			}
			if n.Resources().Outstanding != 1 || p.Info().Pending != 1 || n.Resources().RejectOutstanding != 0 {
				t.Fatal("not ordinary pending", n.Resources(), p.Info())
			}
			dw := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 9901, EndToEnd: 9902}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
			if e = writeWire(ctxTest(t), remote, co, dw); e != nil {
				t.Fatal(e)
			}
			dwa := routeRead(t, co, remote)
			if _, ok := dwa.Body.(*base.DWA); !ok {
				t.Fatal("control namespace lost", dwa)
			}

			a, e := co.CompleteAnswer(r, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived || !got.Validation.Complete {
				t.Fatal(got)
			}
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e := remote.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic charging/retry", e)
		}
	})
}
func TestSCAP1RuntimeInboundPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, scapRuntimeConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, body := range []pdu.Body{scapAnswer(2001), scapAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			r, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(500 + i), EndToEnd: uint32(900 + i)})
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
			if !in.Validation().Complete {
				t.Fatal(in.Validation())
			}
			in.Message().Body.(*scap1.ACR).AccountingRecordNumber = 77
			bad := scapAnswer(2001)
			bad.AccountingRecordNumber = 1
			if got, e := in.Respond(ctxTest(t), bad); e == nil || got.Committed {
				t.Fatal("bad preparation committed", got, e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if got, e := in.Respond(ctx, body); got.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(got, e)
			}
			if got, e := in.Respond(ctxTest(t), body); e != nil || !got.Committed {
				t.Fatal(got, e)
			}
			a := routeRead(t, co, remote)
			if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestSCAP1RuntimeCancelLateInvalidAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, scapRuntimeConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		a, e := co.CompleteAnswer(r, scapAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = port.Receive(quiet)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		id, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = co.CompleteAnswer(r, scapAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		// Use a complete wire projection: static send rejects wrong record numbers,
		// receive must first decode and then reject against the original request.
		raw, e := co.Append(nil, a)
		if e != nil {
			t.Fatal(e)
		}
		raw = replaceAVP32(t, raw, 485, 1)
		if result := remote.WriteFrame(ctxTest(t), raw); result.Err != nil {
			t.Fatal(result.Err)
		}
		got = routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		id, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, remote)
		n.Close()
		got = routeCompletion(t, n, port, id)
		if got.Err == nil {
			t.Fatal("close succeeded request", got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
