// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"sync"
	"testing"
	"time"
)

func imsMOBody() *sgd.OFR {
	b := smsBody(false).(*sgd.OFR)
	b.UserIdentifier.UserName = avp.Optional[string]{}
	b.UserIdentifier.MSISDN = avp.Some([]byte{0, 0, 0, 0, 0, 0, 0, 0xf0})
	b.SMSMICorrelationID = avp.Some(sgd.MOSMSMICorrelationID{HSSID: "001010000000001", OriginatingSIPURI: "sip:alice@example.net", DestinationSIPURI: "sip:bob@example.net"})
	b.SMDeliveryOutcome = avp.Some(sgd.MOSMDeliveryOutcome{IPSMGWSMDeliveryOutcome: sgd.IPSMGWSMDeliveryOutcome{SMDeliveryCause: 0}})
	return b
}
func imsMORuntimeEvidence(requester, responder pdu.Origin) sgd.IMSMOForwardingEvidence {
	return sgd.IMSMOForwardingEvidence{Requester: requester, Responder: responder, HSSID: "001010000000001", OriginatingSIPURI: "sip:alice@example.net", DestinationSIPURI: "sip:bob@example.net", SCAddress: "123"}
}
func imsMORuntimeContext(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, e := sgd.IMSMOForwarding(imsMORuntimeEvidence(requester, responder))
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func imsMOFailure() *sgd.OFA {
	return &sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555}), SMDeliveryFailureCause: avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: 4})}
}
func TestSGDIMSMODuplicateCacheAndIdentityConflict(t *testing.T) {
	eachIMSMOSource(t, func(t *testing.T, tls bool, source imsMOSource) {
		n, co := newInboundNode(t, smsConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := source.receive(t, n, p, co, remote, 5101)
		ctx := source.context(t, gmscOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		dup.Body.(*sgd.OFR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "other-dra.example", ProxyState: []byte{1, 5}}}
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 2 })
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), imsMOFailure()); e != nil || !r.Committed {
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
		if a := duplicateAnswer(t, n, co, remote, dup).Body.(*sgd.OFA); a.ExperimentalResult.Value.ExperimentalResultCode != 5555 {
			t.Fatal(a)
		}
		for i := range 7 {
			conflict, e := co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			conflict.Header.HopByHop = uint32(5110 + i)
			b := conflict.Body.(*sgd.OFR)
			switch i {
			case 0:
				b.SMSMICorrelationID.Value.HSSID = "001010000000002"
			case 1:
				b.SMSMICorrelationID.Value.OriginatingSIPURI = "sip:%61lice@example.net"
			case 2:
				b.SCAddress[0] = 0x43
			case 3:
				b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:bob@Example.net"
			case 4:
				b.SMDeliveryOutcome.Value.IPSMGWSMDeliveryOutcome.SMDeliveryCause = 1
			case 5:
				b.UserIdentifier.UserName = avp.Some("001010123456780")
				if !source.imsi.Present {
					b.UserIdentifier.MSISDN = avp.Optional[[]byte]{}
				}
			case 6:
				b.UserIdentifier.UserName = avp.Some("001010123456789")
				b.UserIdentifier.MSISDN = avp.Some([]byte{0x43, 0xf5})
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
func TestSGDIMSMOBudgetAndCheckClose(t *testing.T) {
	eachIMSMOSource(t, func(t *testing.T, tls bool, source imsMOSource) {
		cfg := smsConfig(t, false)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := source.receive(t, n, p, co, remote, 5201)
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
		ctx := source.context(t, gmscOrigin, n.cfg.Origin)
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; in.Check(ctx) }()
		go func() { defer wg.Done(); <-start; in.Respond(ctxTest(t), imsMOFailure()) }()
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
func TestSGDIMSMOInboundFreezeAndPreparation(t *testing.T) {
	eachIMSMOSource(t, func(t *testing.T, tls bool, source imsMOSource) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		in, req := source.receive(t, n, p, co, remote, 6001)
		if r := in.Validation(); r.State != validation.NeedsApplication {
			t.Fatal(r)
		}
		if r, e := in.Respond(ctxTest(t), imsMOFailure()); r.Committed || !errors.Is(e, validation.ErrContextRequired) {
			t.Fatal(r, e)
		}
		if r := in.Check(moRuntimeContext(t, moRuntimeEvidence(gmscOrigin, n.cfg.Origin, sgd.MOInterfaceSGd))); r.Complete {
			t.Fatal("ordinary role bypass")
		}
		if r := in.Check(source.context(t, draOrigin, n.cfg.Origin)); r.Complete {
			t.Fatal("DRA role bypass")
		}
		in.Message().Body.(*sgd.OFR).SMSMICorrelationID.Value.OriginatingSIPURI = "sip:changed@example.net"
		ctx := source.context(t, gmscOrigin, n.cfg.Origin)
		if r := in.Check(ctx); !r.Complete {
			t.Fatal(r)
		}
		e := source.evidence(gmscOrigin, n.cfg.Origin)
		e.DestinationSIPURI = "sip:changed@example.net"
		wrong, err := sgd.IMSMOForwarding(e)
		if err != nil {
			t.Fatal(err)
		}
		if r := in.Check(wrong); r.Complete {
			t.Fatal("changed frozen evidence")
		}
		wrongSource := source.evidence(gmscOrigin, n.cfg.Origin)
		wrongSource.SenderIMSI = avp.Some("001010123456780")
		other, err := sgd.IMSMOForwarding(wrongSource)
		if err != nil {
			t.Fatal(err)
		}
		if r := in.Check(other); r.Complete {
			t.Fatal("source changed after freeze")
		}
		bad := imsMOFailure()
		bad.ExternalIdentifier = avp.Some("device@example.net")
		if r, e := in.Respond(ctxTest(t), bad); r.Committed || !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(r, e)
		}
		wantUnknown := error(validation.ErrUnsupported)
		if source.imsi.Present && !source.msisdn.Present {
			wantUnknown = validation.ErrMissing
		}
		if r, e := in.Respond(ctxTest(t), moUnknownAnswer()); r.Committed || !errors.Is(e, wantUnknown) {
			t.Fatal(r, e)
		}
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, e := in.Respond(canceled, imsMOFailure()); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if r, e := in.Respond(ctxTest(t), imsMOFailure()); !r.Committed || e != nil {
			t.Fatal(r, e)
		}
		a := duplicateAnswer(t, n, co, remote, req)
		if r := co.ValidateAnswerContext(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}, ctx); !r.Complete {
			t.Fatal(r)
		}
		waitFor(t, func() bool { return n.Resources().InboundBytes == 0 })
	})
}
func TestSGDIMSMOOutboundReportsLateAnswerAndNoReplay(t *testing.T) {
	eachIMSMOSource(t, func(t *testing.T, tls bool, source imsMOSource) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 1)
		for _, policy := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
			if _, e := n.NewRoute(RouteConfig{Command: source.body().CommandKey(), Peers: []PeerID{p.id}, ExchangePolicy: policy, MaxAttempts: 2, MaxReplayAge: time.Second, ReplayDomain: ReplayDomain{Target: serverOrigin, Retention: time.Second}}); !errors.Is(e, ErrUnsupported) {
				t.Fatal(e)
			}
		}
		for _, ctx := range []validation.Context{{}, source.context(t, n.cfg.Origin, draOrigin)} {
			if _, e := p.Submit(ctxTest(t), source.body(), RequestOptions{Validation: ctx}, port); e == nil {
				t.Fatal("missing binding")
			}
		}
		if n.Resources().Outstanding != 0 || n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
		ctx := source.context(t, n.cfg.Origin, serverOrigin)
		for _, body := range []pdu.Body{smsAnswer(false, 2001), imsMOFailure(), &sgd.OFA{AuthSessionState: 1, ExperimentalResult: avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5552})}, smsAnswer(false, 5012), &pdu.ErrorAnswer{ResultCode: 3004}, moUnknownAnswer()} {
			b := source.body()
			id, e := p.Submit(ctxTest(t), b, RequestOptions{Validation: ctx}, port)
			if e != nil {
				t.Fatal(e)
			}
			b.SMSMICorrelationID.Value.DestinationSIPURI = "sip:changed@example.net"
			b.SMRPUI[0] = 255
			b.UserIdentifier.UserName = avp.Some("001010123456700")
			b.UserIdentifier.MSISDN = avp.Some([]byte{0x43, 0xf5})
			req := routeRead(t, co, remote)
			if r := co.ValidateContext(req, ctx); !r.Complete {
				t.Fatal("borrowed source identity", r)
			}
			unknown, isOFA := body.(*sgd.OFA)
			missingMSISDN := isOFA && unknown.ExperimentalResult.Present && unknown.ExperimentalResult.Value.ExperimentalResultCode == 5001 && source.imsi.Present && !source.msisdn.Present
			wireBody := body
			if missingMSISDN {
				wireBody = smsAnswer(false, 2001)
			}
			a, e := co.CompleteAnswer(req, wireBody, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if missingMSISDN {
				// Inject an invalid pair while preserving a structurally valid answer.
				// CompleteAnswer correctly refuses to construct this relation.
				a.Body.(*sgd.OFA).ResultCode = avp.Optional[uint32]{}
				a.Body.(*sgd.OFA).ExperimentalResult = unknown.ExperimentalResult
			}
			if e = writeWire(ctxTest(t), remote, co, a); e != nil {
				t.Fatal(e)
			}
			got := routeCompletion(t, n, port, id)
			if missingMSISDN {
				if !errors.Is(got.Err, validation.ErrInvalidAnswer) || !errors.Is(got.Err, validation.ErrMissing) || got.Answer.Body != nil || got.StopReason != StopProtocol {
					t.Fatal(got)
				}
			} else if got.Outcome != AnswerReceived || got.Err != nil {
				t.Fatal(got)
			}
			if f, ok := body.(*sgd.OFA); ok && f.ExperimentalResult.Present && f.ExperimentalResult.Value.ExperimentalResultCode == 5001 {
				want := error(validation.ErrUnsupported)
				if missingMSISDN {
					want = validation.ErrMissing
				}
				if !errors.Is(got.Validation.Err(), want) {
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
		id, e := p.Submit(ctxTest(t), source.body(), RequestOptions{Validation: ctx}, port)
		if e != nil {
			t.Fatal(e)
		}
		req := routeRead(t, co, remote)
		_ = n.Cancel(id)
		routeCompletion(t, n, port, id)
		a, e := co.CompleteAnswer(req, smsAnswer(false, 2001), serverOrigin)
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
		// No automatic forwarding/retry follows either OFA or cancellation.
		short2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel2()
		if _, e = remote.ReadPrefix(short2, time.Second); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("automatic send", e)
		}
	})
}
func TestSGDIMSMORejectedInboundReleases(t *testing.T) {
	eachIMSMOSource(t, func(t *testing.T, tls bool, source imsMOSource) {
		n, co := newInboundNode(t, smsConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		req, e := co.CompleteRequest(source.body(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 6101, EndToEnd: 6201})
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
