// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Manual Proxy/VSAI exact-instance diagnostics and TCP/TLS safety boundaries.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/wire"
)

func groupFrame(t *testing.T, co *codec.Codec, req pdu.Message, kind string) ([]byte, []byte, uint32) {
	vendor := invalidPart(t, avp.Key{Code: 266}, 64, []byte{0, 0, 40, 175})
	auth := invalidPart(t, avp.Key{Code: 258}, 64, []byte{1, 0, 0, 1})
	acct := invalidPart(t, avp.Key{Code: 259}, 0, []byte{0, 0, 0, 9})
	data := vendor
	code := uint32(5009)
	want := append(append([]byte{}, auth...), acct...)
	if kind == "neither" {
		code = 5005
		want = append(invalidPart(t, avp.Key{Code: 258}, 64, make([]byte, 4)), invalidPart(t, avp.Key{Code: 259}, 64, make([]byte, 4))...)
	} else {
		if kind == "both-reverse" {
			want = append(append([]byte{}, acct...), auth...)
		}
		data = append(data, want...)
	}
	data = append(data, invalidPart(t, avp.Key{Code: 999999}, 0, []byte{0, 255})...)
	raw, _ := invalidRaw(t, co, req, lengthPath(avp.Key{Code: 260}), data, avp.Key{})
	return raw, want, code
}
func assertGroupPair(t *testing.T, q, want []byte) {
	t.Helper()
	d := wire.NewDecoder(q, wire.DefaultLimits())
	v, ok, e := d.Next()
	if e != nil || !ok || v.Key != (avp.Key{Code: 260}) || d.Remaining() != 0 || !bytes.Equal(v.Data, want) {
		t.Fatalf("pair %x want %x (%v)", q, want, e)
	}
}
func TestRejectGroupVSAI(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, emergency := range []bool{false, true} {
			for _, kind := range []string{"both", "both-reverse", "neither"} {
				t.Run(fmt.Sprintf("%s/emergency-%t", kind, emergency), func(t *testing.T) {
					cfg := manualConfig(t)
					if emergency {
						cfg.Resources.InboundClasses[0].Slots = 1
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := mixedDuplicateRequest(t, n, co, 2, 10201, 20201, gmscOrigin)
					var held *Inbound
					if emergency {
						first := mixedDuplicateRequest(t, n, co, 2, 10200, 20200, gmscOrigin)
						combinedSend(t, co, remote, first)
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					raw, want, code := groupFrame(t, co, req, kind)
					_, cause := co.Decode(raw)
					q, result, ok := manualQuote(raw[20:], cause, cfg.Limits, 1024)
					if !ok || result != code {
						t.Fatal("lost manual evidence", cause, result)
					}
					assertGroupPair(t, q, want)
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, code)
					if !ans.SessionID.Present || !base.EqualProxies(ans.ProxyInfo, req.Body.(*sh.UDR).ProxyInfo) {
						t.Fatal("envelope")
					}
					assertGroupPair(t, ans.FailedAVP.Value.Data, want)
					waitFor(t, func() bool {
						return n.Resources().RejectOutstanding == 0 && n.Resources().InboundOutstanding == before.InboundOutstanding
					})
					after := n.Resources()
					if after.InboundBytes != before.InboundBytes {
						t.Fatal("retained or changed charge")
					}
					if !p.Info().Ready {
						t.Fatal("closed safe request")
					}
					noDuplicateDelivery(t, p)
					if held != nil {
						if _, e := held.Respond(ctxTest(t), mixedDuplicateAnswer(held.Message())); e != nil {
							t.Fatal(e)
						}
						if _, e := readWire(ctxTest(t), remote, co); e != nil {
							t.Fatal(e)
						}
					}
					req.Header.HopByHop++
					req.Header.EndToEnd++
					combinedSend(t, co, remote, req)
					in, e := p.Receive(ctxTest(t))
					if e != nil {
						t.Fatal(e)
					}
					if _, e = in.Reject(ctxTest(t), FailureContextUnavailable); e != nil {
						t.Fatal(e)
					}
					readReject(t, remote, req.Header, 5012)
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}
