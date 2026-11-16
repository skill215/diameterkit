// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

// Run the same network lifecycle contract for all selected source shapes.
type imsMOSource struct{ imsi, msisdn avp.Optional[string] }

func eachIMSMOSource(t *testing.T, f func(*testing.T, bool, imsMOSource)) {
	t.Helper()
	for _, tc := range []struct {
		name   string
		source imsMOSource
	}{
		{"dummy", imsMOSource{}},
		{"imsi", imsMOSource{imsi: avp.Some("001010123456789")}},
		{"imsi_msisdn", imsMOSource{imsi: avp.Some("001010123456789"), msisdn: avp.Some("0012025550123")}},
	} {
		t.Run(tc.name, func(t *testing.T) { eachNetwork(t, func(t *testing.T, tls bool) { f(t, tls, tc.source) }) })
	}
}
func (s imsMOSource) body() *sgd.OFR {
	b := imsMOBody()
	if s.imsi.Present {
		b.UserIdentifier.UserName = s.imsi
		b.UserIdentifier.MSISDN = avp.Optional[[]byte]{}
		if s.msisdn.Present {
			v, err := avp.EncodeTBCD(s.msisdn.Value)
			if err != nil {
				panic(err)
			}
			b.UserIdentifier.MSISDN = avp.Some(v)
		}
	}
	return b
}
func (s imsMOSource) evidence(requester, responder pdu.Origin) sgd.IMSMOForwardingEvidence {
	e := imsMORuntimeEvidence(requester, responder)
	e.SenderIMSI = s.imsi
	e.SenderMSISDN = s.msisdn
	return e
}
func (s imsMOSource) context(t *testing.T, requester, responder pdu.Origin) validation.Context {
	t.Helper()
	v, err := sgd.IMSMOForwarding(s.evidence(requester, responder))
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func (s imsMOSource) receive(t *testing.T, n *Node, p *Peer, co *codec.Codec, remote transport.Conn, hbh uint32) (*Inbound, pdu.Message) {
	t.Helper()
	r, e := co.CompleteRequest(s.body(), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: hbh, EndToEnd: hbh + 10000})
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
