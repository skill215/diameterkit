// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

type ipReportPurpose bool

func eachIPReportPurpose(t *testing.T, f func(*testing.T, bool, ipReportPurpose)) {
	t.Helper()
	for _, tc := range []struct {
		name string
		t4   ipReportPurpose
	}{{"msisdn", false}, {"t4_imsi", true}} {
		t.Run(tc.name, func(t *testing.T) { eachNetwork(t, func(t *testing.T, tls bool) { f(t, tls, tc.t4) }) })
	}
}
func (p ipReportPurpose) body(mask int) *s6c.RDR {
	b := ipDeliveryReportBody(mask)
	if p {
		b.UserIdentifier = s6c.ReportUserIdentifier{UserName: avp.Some("001010123456789")}
	}
	return b
}
func (p ipReportPurpose) context(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	if p {
		return t4Context(t, false, requester, responder)
	}
	return validation.Context{}
}
func (p ipReportPurpose) request(t *testing.T, n *Node, co *codec.Codec, mask int, hbh uint32) pdu.Message {
	t.Helper()
	if !p {
		return ipDeliveryReportRequest(t, n, co, mask, hbh)
	}
	r, e := co.CompleteRequest(p.body(mask), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 1000})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func (p ipReportPurpose) receive(t *testing.T, n *Node, peer *Peer, co *codec.Codec, remote transport.Conn, mask int, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	if !p {
		return ipDeliveryReportReceive(t, n, peer, co, remote, mask, hbh)
	}
	r := p.request(t, n, co, mask, hbh)
	if err := writeWire(ctxTest(t), remote, co, r); err != nil {
		t.Fatal(err)
	}
	in, err := peer.Receive(ctxTest(t))
	if err != nil {
		t.Fatal(err)
	}
	return in, r
}
func (p ipReportPurpose) check(t *testing.T, in *Inbound, responder pdu.Origin) {
	t.Helper()
	if p {
		if r := in.Validation(); r.State != validation.NeedsApplication {
			t.Fatal(r)
		}
		if r, err := in.Respond(ctxTest(t), ipDeliveryReportAnswer(2001)); r.Committed || !errors.Is(err, validation.ErrContextRequired) {
			t.Fatal(r, err)
		}
		if r := in.Check(t4Context(t, true, gmscOrigin, responder)); !errors.Is(r.Err(), validation.ErrContext) {
			t.Fatal("alert cannot authorize report", r)
		}
		if r := in.Check(p.context(t, draOrigin, responder)); !errors.Is(r.Err(), validation.ErrContext) {
			t.Fatal("DRA is not requester", r)
		}
	}
	if r := in.Check(p.context(t, gmscOrigin, responder)); !r.Complete {
		t.Fatal(r)
	}
	if p {
		e := s6c.T4NoMSISDNEvidence{Procedure: s6c.T4DeliveryReport, Requester: gmscOrigin, Responder: responder, IMSI: "001010123456780", SCAddress: "12345"}
		wrong, err := s6c.T4NoMSISDN(e)
		if err != nil {
			t.Fatal(err)
		}
		if r := in.Check(wrong); !errors.Is(r.Err(), validation.ErrContext) {
			t.Fatal("frozen identity changed", r)
		}
	}
}
