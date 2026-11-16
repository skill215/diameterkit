package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func legacyBody() *smsfwro.CCR {
	return smsfwro.NewIEC(smsfwro.IECInput{SessionID: "ctf.example;legacy;1", ServiceContextID: "legacy-sms@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", Timestamp: 3908988800, ProtocolID: []byte{0, 255}, UserDataHeader: []byte{0, 255}, SMSKey: "key"})
}
func legacyAnswer(code uint32) *smsfwro.CCA {
	return &smsfwro.CCA{AuthApplicationID: 4, CCRequestType: 4, ResultCode: code}
}
func legacyConfig(t *testing.T) NodeConfig {
	c := roConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	p, e := smsfwro.Profile()
	if e != nil {
		t.Fatal(e)
	}
	c.Bundle, e = schema.Compile(bp, sp, p)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestSMSFWIECNetworkPairs(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, body := range []pdu.Body{legacyAnswer(2001), legacyAnswer(4012), legacyAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			q := legacyBody()
			id, e := p.Submit(ctxTest(t), q, RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			q.ServiceInformation.SMSInformation.SMUserDataHeader[0] = 99
			r := routeRead(t, co, remote)
			if r.Body.(*smsfwro.CCR).ServiceInformation.SMSInformation.SMUserDataHeader[0] != 0 {
				t.Fatal("submit alias")
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
func TestSMSFWIECInboundPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		for i, body := range []pdu.Body{legacyAnswer(2001), legacyAnswer(4012), legacyAnswer(5012), &pdu.ErrorAnswer{ResultCode: 3004}} {
			r, e := co.CompleteRequest(legacyBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: uint32(500 + i), EndToEnd: uint32(900 + i)})
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
			in.Message().Body.(*smsfwro.CCR).CCRequestNumber = 77
			bad := legacyAnswer(2001)
			bad.CCRequestNumber = 1
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
func TestSMSFWIECCancelLateInvalidAndClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, legacyConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), legacyBody(), RequestOptions{}, port)
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
		a, e := co.CompleteAnswer(r, legacyAnswer(2001), serverOrigin)
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
		id, e = p.Submit(ctxTest(t), legacyBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = co.CompleteAnswer(r, legacyAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		// Use a complete wire projection: static send rejects wrong request numbers,
		// receive must first decode and then reject against the original request.
		raw, e := co.Append(nil, a)
		if e != nil {
			t.Fatal(e)
		}
		raw = replaceAVP32(t, raw, 415, 1)
		if result := remote.WriteFrame(ctxTest(t), raw); result.Err != nil {
			t.Fatal(result.Err)
		}
		got = routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		id, e = p.Submit(ctxTest(t), legacyBody(), RequestOptions{}, port)
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
func replaceAVP32(t *testing.T, b []byte, code uint32, value uint32) []byte {
	t.Helper()
	for off := 20; off < len(b); {
		n := int(b[off+5])<<16 | int(b[off+6])<<8 | int(b[off+7])
		c := uint32(b[off])<<24 | uint32(b[off+1])<<16 | uint32(b[off+2])<<8 | uint32(b[off+3])
		if c == code && b[off+4]&128 == 0 && n == 12 {
			b[off+8] = byte(value >> 24)
			b[off+9] = byte(value >> 16)
			b[off+10] = byte(value >> 8)
			b[off+11] = byte(value)
			return b
		}
		off += (n + 3) &^ 3
	}
	t.Fatal("AVP not found")
	return nil
}
