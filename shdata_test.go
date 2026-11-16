// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"testing"

	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
)

func TestShDataExplicitApplicationChecksOverTCPAndTLS(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		limits := shdata.DefaultLimits()
		reg, e := shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: "12025550999"}, limits)
		if e != nil {
			t.Fatal(e)
		}
		dereg, e := shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: ""}, limits)
		if e != nil {
			t.Fatal(e)
		}
		reach, e := shdata.MarshalReachability(shdata.Reachability{SGSN: avp.Some(uint8(0)), AMF3GPP: avp.Some(uint8(0))}, limits)
		if e != nil {
			t.Fatal(e)
		}
		pull, e := shdata.MarshalSMSPull(shdata.SMSRegistration{Number: "12025550999", SCAddress: avp.Some("12025550100"), SBI: avp.Some(uint8(0))}, limits)
		if e != nil {
			t.Fatal(e)
		}
		for i, tc := range []struct {
			command  uint32
			data     []byte
			use      shdata.Use
			complete bool
		}{
			{307, reg, shdata.SMSUpdate, true}, {307, dereg, shdata.SMSUpdate, true},
			{307, pull, shdata.SMSUpdate, false}, {309, reach, shdata.ReachabilityNotification, true},
			{309, []byte("<Sh-Data/>"), shdata.ReachabilityNotification, false},
			{309, []byte("<broken"), shdata.ReachabilityNotification, false},
		} {
			req := shRequest(t, n, co, tc.command, uint32(1800+i))
			switch b := req.Body.(type) {
			case *sh.PUR:
				b.UserData = tc.data
			case *sh.PNR:
				b.UserData = tc.data
			}
			if e = writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if !in.Validation().Complete {
				t.Fatal("AVP report unexpectedly includes XML", in.Validation())
			}
			view := in.Message()
			var payload []byte
			switch b := view.Body.(type) {
			case *sh.PUR:
				payload = b.UserData
			case *sh.PNR:
				payload = b.UserData
			}
			d, e := shdata.Parse(payload, limits)
			complete := false
			if e == nil {
				r, e := d.Check(tc.use)
				if e != nil {
					t.Fatal(e)
				}
				complete = r.Complete
			}
			if complete != tc.complete {
				t.Fatalf("case %d helper complete=%v err=%v", i, complete, e)
			}
			if d != nil {
				clear(payload)
				if !bytes.Equal(d.Raw(), tc.data) {
					t.Fatal("borrowed application view")
				}
			}
			// This is an explicit application decision, not an automatic XML-to-Diameter
			// error mapping or a persistent registration/subscription update.
			code := uint32(5012)
			if complete {
				code = 2001
			}
			if r, e := in.Respond(ctxTest(t), shAnswer(tc.command, code, false)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			duplicateAnswer(t, n, co, remote, req)
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
		// Pull data passes as owned bytes through a real request/answer completion.
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), shBody(306), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		body := shAnswer(306, 2001, false).(*sh.UDA)
		body.UserData = avp.Some(pull)
		ans, e := co.CompleteAnswer(req, body, serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
			t.Fatal(e)
		}
		done, e := port.Receive(ctxTest(t))
		if e != nil || done.RequestID != id || done.Result.Outcome != AnswerReceived {
			t.Fatal(done, e)
		}
		d, e := shdata.Parse(done.Result.Answer.Body.(*sh.UDA).UserData.Value, limits)
		if e != nil {
			t.Fatal(e)
		}
		r, e := d.Check(shdata.SMSPull)
		if e != nil || !r.Complete || d.Data().SMS.Value.SCAddress.Value != "12025550100" {
			t.Fatal(r, e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
