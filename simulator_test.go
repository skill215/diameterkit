// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
	"time"
)

func TestSimulatorShMigrationNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		for _, code := range []uint32{306, 307, 308} {
			req := shRequest(t, n, co, code, 7700+code)
			if code == 307 {
				req.Body.(*sh.PUR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra-one.example", ProxyState: []byte{0, 255}}}
			}
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if code == 306 {
				bad := shAnswer(code, 2001, false).(*sh.UDA)
				bad.UserData.Present = false
				if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrMissing) {
					t.Fatal(r, e)
				}
			}
			duplicate, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			duplicate.Header.HopByHop += 100
			duplicate.Header.Flags |= pdu.RetransmittedFlag
			if code == 307 {
				duplicate.Body.(*sh.PUR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra-two.example", ProxyState: []byte{1, 254}}}
			}
			if e = writeWire(ctxTest(t), remote, co, duplicate); e != nil {
				t.Fatal(e)
			}
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
			noDuplicateDelivery(t, p)
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			if r, e := in.Respond(canceled, shAnswer(code, 2001, false)); r.Committed || !errors.Is(e, context.Canceled) {
				t.Fatal(r, e)
			}
			if r, e := in.Respond(ctxTest(t), shAnswer(code, 2001, false)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			for range 2 {
				a := routeRead(t, co, remote)
				r := req
				if a.Header.HopByHop == duplicate.Header.HopByHop {
					r = duplicate
				}
				if v := co.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, validation.Context{}); !v.Complete {
					t.Fatal(v)
				}
			}
			duplicate.Header.HopByHop++
			if e = writeWire(ctxTest(t), remote, co, duplicate); e != nil {
				t.Fatal(e)
			}
			duplicateAnswer(t, n, co, remote, duplicate)
			noDuplicateDelivery(t, p)
			waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
		}
	})
}
func TestSimulatorSMSMigrationNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		for _, mt := range []bool{false, true} {
			req := smsRequest(t, n, co, mt, false, 7800)
			if !mt {
				b := req.Body.(*sgd.OFR)
				b.UserIdentifier.UserName = avp.Some("460310000000000")
				b.SMRPUI = []byte{0, 255, 0x30, 0x42, 0x30, 0x44}
			}
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if !mt {
				if r, e := in.Respond(ctxTest(t), &sgd.OFA{AuthSessionState: 0, ResultCode: avp.Some(uint32(2001))}); r.Committed || !errors.Is(e, validation.ErrField) {
					t.Fatal(r, e)
				}
			}
			if r, e := in.Respond(ctxTest(t), smsAnswer(mt, 2001)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			duplicateAnswer(t, n, co, remote, req)
			waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
		}
		// Inject the old POC's illegal Auth-Session-State through raw transport;
		// codec send would reject it before I/O. Invalid matched answer completes once.
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), smsBody(false), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		answer, e := co.CompleteAnswer(req, smsAnswer(false, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := co.Append(nil, answer)
		if e != nil {
			t.Fatal(e)
		}
		sendRejectRaw(t, remote, replaceAVP32(t, raw, 277, 0))
		got := routeCompletion(t, n, port, id)
		if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
			t.Fatal(got)
		}
		if e = writeWire(ctxTest(t), remote, co, answer); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("second completion", e)
		}
		waitFor(t, func() bool { return n.Resources().OutboundBytes == 0 })
	})
}
func TestSimulatorSRAFailureMigrationNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveInbound(t, n, p, remote, co, 7900)
		bad := &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012)), UserName: avp.Some("123499999999"), ServingNode: avp.Some(s6c.ServingNode{MMEName: avp.Some("test-mme-name"), MMERealm: avp.Some("test-mme-realm"), MMENumber: avp.Some([]byte{0x21, 0x43, 0x65, 0x87, 0x09})})}
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrField) {
			t.Fatal(r, e)
		}
		bad.UserName.Present = false
		bad.ServingNode.Present = false
		if r, e := in.Respond(ctxTest(t), bad); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		duplicateAnswer(t, n, co, remote, req)
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
	})
}
