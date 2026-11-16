package diameter

import (
	"bytes"
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"sync"
	"testing"
	"time"
)

func imsMTBody() *sgd.TFR {
	b := smsBody(true).(*sgd.TFR)
	b.UserName = "001010000000001"
	b.MMENumberForMTSMS = avp.Optional[[]byte]{}
	b.SMSMICorrelationID = avp.Some(sgd.MTSMSMICorrelationID{OriginatingSIPURI: "sip:alice@example.net", DestinationSIPURI: "sip:bob@example.net"})
	return b
}
func imsMTUnknownAnswer() *sgd.TFA {
	return &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}
}
func imsMTRuntimeEvidence(requester, responder pdu.Origin) sgd.IMSMTForwardingEvidence {
	return sgd.IMSMTForwardingEvidence{Requester: requester, Responder: responder, HSSID: "001010000000001", OriginatingSIPURI: "sip:alice@example.net", DestinationSIPURI: "sip:bob@example.net", SCAddress: "123"}
}
func imsMTRuntimeContext(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, e := sgd.IMSMTForwarding(imsMTRuntimeEvidence(requester, responder))
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func imsMTFailure() *sgd.TFA {
	return &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555}), SMDeliveryFailureCause: avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 0})}
}
func receiveIMSMT(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r, e := co.CompleteRequest(imsMTBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 10000})
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
	return in, r
}
func TestSGDIMSMTDuplicateCacheAndIdentityConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSMT(t, n, p, co, remote, 5101)
		ctx := imsMTRuntimeContext(t, gmscOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*sgd.TFR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 5}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), imsMTFailure()); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			a := routeRead(t, co, remote)
			original := req
			if a.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if r := co.ValidateAnswerContext(original, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
				t.Fatal(r)
			}
		}
		dup.Header.HopByHop++
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*sgd.TFA); a.ExperimentalResult.Value.ExperimentalResultCode != 5555 {
			t.Fatal(a)
		}
		for i := range 5 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(5110 + i)
			b := conflict.Body.(*sgd.TFR)
			switch i {
			case 0:
				b.UserName = "001010000000002"
			case 1:
				b.SMSMICorrelationID.Value.OriginatingSIPURI = "sip:%61lice@example.net"
			case 2:
				b.SCAddress[0] = 0x43
			case 3:
				b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net"
			case 4:
				b.SMDeliveryTimer = avp.Some(uint32(31))
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
func TestSGDIMSMTBudgetAndCheckClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := smsConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSMT(t, n, p, co, remote, 5201)
		other, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		other.Header.HopByHop++
		other.Header.EndToEnd++
		if e = writeWire(ctxTest(t), remote, co, other); e != nil {
			t.Fatal(e)
		}
		if a, ok := duplicateAnswer(t, n, co, remote, other).Body.(*pdu.ErrorAnswer); !ok || a.ResultCode != 3004 {
			t.Fatal(a)
		}
		ctx := imsMTRuntimeContext(t, gmscOrigin, n.cfg.Origin)
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), imsMTFailure()) }()
		close(start)
		_ = n.Close()
		wg.Wait()
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestSGDIMSMTInboundFreezeAndPreparation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := receiveIMSMT(t, n, p, co, remote, 6001)
		if r := in.Validation(); r.State != validation.NeedsApplication {
			t.Fatal(r)
		}
		if r, e := in.Respond(ctxTest(t), imsMTFailure()); r.Committed || !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(r, e)
		}
		if r := in.Check(moRuntimeContext(t, moRuntimeEvidence(gmscOrigin, n.cfg.Origin, sgd.MOInterfaceSGd))); r.Complete {
			t.Fatal("ordinary role bypass")
		}
		if r := in.Check(imsMORuntimeContext(t, gmscOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("MO role bypass")
		}
		if r := in.Check(imsMTRuntimeContext(t, draOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("DRA role bypass")
		}
		in.Message().Body.(*sgd.TFR).SMSMICorrelationID.Value.OriginatingSIPURI = "sip:changed@example.net"
		ctx := imsMTRuntimeContext(t, gmscOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		e := imsMTRuntimeEvidence(gmscOrigin, n.cfg.Origin)
		e.DestinationSIPURI = "sip:changed@example.net"
		wrong, err := sgd.IMSMTForwarding(e)
		if err != nil {
			t.Fatal(err)
		}
		if r := in.Check(wrong); r.Complete {
			t.Fatal("changed frozen evidence")
		}
		bad := imsMTFailure()
		bad.ExperimentalResult.Value.ExperimentalResultCode = 5550
		bad.SMDeliveryFailureCause.Present = false
		bad.AbsentUserDiagnosticSM = avp.Some(uint32(13))
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		if r, e := in.Respond(ctxTest(t), imsMTUnknownAnswer()); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, imsMTFailure()); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if r, e := in.Respond(ctxTest(t), imsMTFailure()); !r.Committed || e != nil {
			t.Fatal(r, e)
		}
		a := duplicateAnswer(t, n, co, remote, req)
		if r := co.ValidateAnswerContext(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
			t.Fatal(r)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
	})
}
func TestSGDIMSMTOutboundReportsLateAnswerAndNoReplay(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, policy := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			if _, e := n.NewRoute(RouteConfig{Command: imsMTBody().CommandKey(), Peers: []PeerID{p.id}, ExchangePolicy: policy, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Second}}); !errors.Is(e, ErrUnsupported) {
				t.Fatal(e)
			}
		}
		for _, ctx := range []validation.Context{{}, imsMTRuntimeContext(t, n.cfg.Origin, draOrigin)} {
			if _, e := p.Submit(ctxTest(t), imsMTBody(), RequestOptions{Validation: ctx}, port); e == nil {
				t.Fatal("missing binding")
			}
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		ctx := imsMTRuntimeContext(t, n.cfg.Origin, serverOrigin)
		for _, body := range []pdu.Body{smsAnswer(true, 2001), imsMTFailure(), &sgd.TFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5552})}, smsAnswer(true, 5012), &pdu.ErrorAnswer{ResultCode: 3004}, imsMTUnknownAnswer()} {
			b := imsMTBody()
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:changed@example.net"
			b.SMRPUI[0] = 255
			req := routeRead(t, co, remote)
			a, e := co.CompleteAnswer(req, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if got.Outcome != AnswerReceived {
				t.Fatal(got)
			}
			if f, ok := body.(*sgd.TFA); ok && f.ExperimentalResult.Present && f.ExperimentalResult.Value.ExperimentalResultCode == 5001 {
				if !errors.Is(got.Validation.Err(), validation.ErrUnsupported) {
					t.Fatal(got)
				}
			} else if !got.Validation.Complete {
				t.Fatal(got)
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			short, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("second completion", e)
			}
			cancel()
		}
		id, e := p.Submit(ctxTest(t), imsMTBody(), RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, e := co.CompleteAnswer(req, smsAnswer(true, 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(short); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		// No automatic forwarding/retry follows either TFA or cancellation.
		short2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel2()
		if _, e = remote.ReadPrefix(short2, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic send", e)
		}
	})
}
func TestSGDIMSMTRejectedInboundReleases(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		req, e := co.CompleteRequest(imsMTBody(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 6101, EndToEnd: 6201})
		if e != nil {
			t.Fatal(e)
		}
		raw, e := co.Append(nil, req)
		if e != nil {
			t.Fatal(e)
		}
		// Independently replace the URI scheme in a structurally valid UTF8 field.
		marker := []byte("sip:alice@example.net")
		at := bytes.Index(raw, marker)
		if at < 0 {
			t.Fatal("missing URI")
		}
		copy(raw[at:], []byte("tel:"))
		sendRejectRaw(t, remote, raw)
		waitFor(t, func() bool { return p.Info().Protocol == ProtocolClosed })
		if info := p.Info(); !errors.Is(info.Cause, validation.ErrUnsupported) || info.Ready {
			t.Fatal(info)
		}
		if _, e = remote.ReadPrefix(ctxTest(t), time.Second); e == nil || errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("expected close without response", e)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 && n.Resources().InboundOutstanding == 0 })
		n.mu.Lock()
		queued := len(p.inboundQueue)
		n.mu.Unlock()
		if queued != 0 {
			t.Fatal("unsupported inbound delivered", queued)
		}
	})
}
