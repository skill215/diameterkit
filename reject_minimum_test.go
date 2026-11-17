// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Whole-group 5005 diagnostics over TCP/TLS and emergency slots.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func minimumExample(t *testing.T, kind string) ([]missingSelector, []byte) {
	part := func(code, vendor uint32, data []byte) []byte {
		flags := byte(64)
		if vendor != 0 {
			flags = 192
		}
		return invalidPart(t, avp.Key{Code: code, VendorID: vendor}, flags, data)
	}
	switch kind {
	case "money-value":
		children := append(part(447, 0, make([]byte, 8)), part(429, 0, make([]byte, 4))...)
		return lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}, avp.Key{Code: 413}, avp.Key{Code: 445}), part(456, 0, part(437, 0, part(413, 0, part(445, 0, children))))
	case "vendor-application":
		return lengthPath(avp.Key{Code: 260}), part(260, 0, append(part(266, 0, make([]byte, 4)), part(259, 0, make([]byte, 4))...))
	case "requested-unit":
		children := append(part(611, 193, make([]byte, 4)), part(612, 193, part(617, 193, make([]byte, 8)))...)
		return lengthPath(avp.Key{Code: 606, VendorID: 193}), part(606, 193, children)
	case "unit-value":
		return lengthPath(avp.Key{Code: 606, VendorID: 193}, avp.Key{Code: 612, VendorID: 193}), part(606, 193, part(612, 193, part(617, 193, make([]byte, 8))))
	}
	panic(kind)
}
func TestRejectMinimumGroups(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"money-value", "vendor-application", "requested-unit", "unit-value"} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/emergency-%t", kind, emergency), func(t *testing.T) {
					cfg := scapRuntimeConfig(t)
					if kind == "money-value" {
						cfg = combinedConfig(t)
					}
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					cfg.Resources.RejectSlots = 2
					cfg.Resources.EmergencyRequestMax = 4096
					cfg.Resources.EmergencyReplyMax = 4096
					if emergency {
						cfg.Resources.InboundClasses[0].Slots = 1
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					makeRequest := func(id uint32) pdu.Message {
						if kind == "money-value" {
							return combinedDuplicateRequest(t, n, co, "general", id, id+10000)
						}
						req, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: id, EndToEnd: id + 10000})
						if e != nil {
							t.Fatal(e)
						}
						req.Body.(*scap1.ACR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{0, 255}}}
						return req
					}
					var held *Inbound
					if emergency {
						combinedSend(t, co, remote, makeRequest(10400))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := makeRequest(10401)
					path, want := minimumExample(t, kind)
					raw := missingRaw(t, co, req, path, false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					q, ok := missingavp.Quote(cause, cfg.Limits, len(want))
					if !ok || !bytes.Equal(q, want) {
						t.Fatalf("minimum %x want %x (%v)", q, want, cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, len(want)-1); ok {
						t.Fatal("quote budget")
					}
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5005)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
						t.Fatal("diagnostic/envelope")
					}
					clear(raw)
					if !bytes.Equal(q, want) {
						t.Fatal("borrowed")
					}
					waitFor(t, func() bool {
						return n.Resources().RejectOutstanding == 0 && n.Resources().InboundOutstanding == before.InboundOutstanding
					})
					if n.Resources().InboundBytes != before.InboundBytes {
						t.Fatal("changed held charge")
					}
					noDuplicateDelivery(t, p)
					if held != nil {
						if _, e := held.Reject(ctxTest(t), FailureContextUnavailable); e != nil {
							t.Fatal(e)
						}
						readReject(t, remote, held.Message().Header, 5012)
					}
					follow := makeRequest(10402)
					combinedSend(t, co, remote, follow)
					in, e := p.Receive(ctxTest(t))
					if e != nil {
						t.Fatal(e)
					}
					if _, e = in.Reject(ctxTest(t), FailureContextUnavailable); e != nil {
						t.Fatal(e)
					}
					readReject(t, remote, follow.Header, 5012)
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}
