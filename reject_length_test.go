// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded typed-length 5014 rejection over TCP and TLS.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/wire"
)

func lengthPath(keys ...avp.Key) []missingSelector {
	p := make([]missingSelector, len(keys))
	for i, k := range keys {
		p[i].key = k
	}
	return p
}
func lengthAddressRequest(t *testing.T, n *Node, co *codec.Codec) pdu.Message {
	t.Helper()
	b := roBody()
	b.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{0, 255, 1}}}
	b.ServiceInformation.Value.SMSInformation.Value.RecipientInfo = []ro.RecipientInfo{{RecipientSCCPAddress: avp.Some(avp.Address{Family: 8, Data: []byte("15551234567")})}}
	m, e := co.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 10001, EndToEnd: 20001})
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func lengthAddressPath() []missingSelector {
	return lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: 2026}, avp.Key{VendorID: 10415, Code: 2010})
}

func TestRejectLengthGenerated(t *testing.T) {
	key := func(code uint32) avp.Key { return avp.Key{Code: code} }
	deep := func(code uint32) []missingSelector {
		return lengthPath(key(456), key(437), key(413), key(445), key(code))
	}
	cases := []struct {
		name, profile string
		path          []missingSelector
		data          []byte
	}{
		{"u32-empty", "s6c", lengthPath(key(277)), nil},
		{"u32-short-padding", "s6c", lengthPath(key(277)), []byte{1, 2, 3}},
		{"u32-long", "s6c", lengthPath(key(277)), []byte{1, 2, 3, 4, 5}},
		{"tbcd-empty", "s6c", lengthPath(avp.Key{VendorID: 10415, Code: 3300}), nil},
		{"tbcd-long", "s6c", lengthPath(avp.Key{VendorID: 10415, Code: 3300}), bytes.Repeat([]byte{0x21}, 9)},
		{"second-parent-vendor", "s6c", []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{VendorID: 10415, Code: 629}, 0}}, []byte{1, 2, 3}},
		{"i32-short-deep", "combined", deep(429), []byte{1, 2, 3}},
		{"i32-long-deep", "combined", deep(429), make([]byte, 5)},
		{"i64-short-deep", "combined", deep(447), make([]byte, 7)},
		{"i64-long-deep", "combined", deep(447), make([]byte, 9)},
		{"u64-short", "combined", lengthPath(key(456), key(437), key(417)), make([]byte, 7)},
		{"u64-long", "combined", lengthPath(key(456), key(437), key(417)), make([]byte, 9)},
		{"time-empty", "combined", lengthPath(key(55)), nil},
		{"time-long", "combined", lengthPath(key(55)), make([]byte, 5)},
		{"address-empty", "ro", lengthAddressPath(), nil},
		{"address-family-short", "ro", lengthAddressPath(), []byte{0}},
		{"address-ipv4-short", "ro", lengthAddressPath(), []byte{0, 1, 192, 0, 2}},
		{"address-ipv4-long", "ro", lengthAddressPath(), []byte{0, 1, 192, 0, 2, 1, 2}},
		{"address-ipv6-short", "ro", lengthAddressPath(), append([]byte{0, 2}, make([]byte, 15)...)},
		{"address-ipv6-long", "ro", lengthAddressPath(), append([]byte{0, 2}, make([]byte, 17)...)},
	}
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := rejectConfig(t)
				if tc.profile == "combined" {
					cfg = combinedConfig(t)
				}
				if tc.profile == "ro" {
					cfg = roConfig(t)
				}
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				var req pdu.Message
				switch tc.profile {
				case "combined":
					req = combinedDuplicateRequest(t, n, co, "general", 10001, 20001)
				case "ro":
					req = lengthAddressRequest(t, n, co)
				default:
					req = inboundRequest(t, n, co, 10001)
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
				}
				raw, want := invalidRaw(t, co, req, tc.path, tc.data, avp.Key{})
				_, cause := co.Decode(raw)
				if !errors.Is(cause, avp.ErrValue) {
					t.Fatal("lost sentinel", cause)
				}
				quote, code, ok := excessavp.QuoteInvalid(raw[20:], cause, cfg.Limits, 1024)
				if !ok || code != 5014 {
					t.Fatal("missing typed evidence", cause, code)
				}
				assertExcessPath(t, quote, tc.path, want, false)
				if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
					t.Fatal("reply budget bypass")
				}
				sendRejectRaw(t, remote, raw)
				ans := readReject(t, remote, req.Header, 5014)
				if !ans.SessionID.Present || len(ans.ProxyInfo) != 1 || ans.FailedAVP.Value.Incomplete {
					t.Fatal("lost envelope or incomplete typed quote")
				}
				assertExcessPath(t, ans.FailedAVP.Value.Data, tc.path, want, false)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
				if !p.Info().Ready {
					t.Fatal("closed peer")
				}
				noDuplicateDelivery(t, p)
				// A following valid request proves rejection did not retain or poison work.
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
	})
}

func TestRejectLengthEmergency(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, tbcd := range []bool{false, true} {
			t.Run(fmt.Sprintf("tbcd-%t", tbcd), func(t *testing.T) {
				cfg := rejectConfig(t)
				cfg.Resources.InboundClasses[0].Slots = 1
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				held, _ := receiveInbound(t, n, p, remote, co, 10011)
				before := n.Resources()
				req := inboundRequest(t, n, co, 10012)
				path := lengthPath(avp.Key{Code: 277})
				data := []byte{1, 2, 3}
				if tbcd {
					path = lengthPath(avp.Key{VendorID: 10415, Code: 3300})
					data = nil
				}
				raw, want := invalidRaw(t, co, req, path, data, avp.Key{})
				sendRejectRaw(t, remote, raw)
				ans := readReject(t, remote, req.Header, 5014)
				assertExcessPath(t, ans.FailedAVP.Value.Data, path, want, false)
				waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
				after := n.Resources()
				if after.InboundOutstanding != 1 || after.InboundBytes != before.InboundBytes || after.FixedBytes != before.FixedBytes {
					t.Fatal("held charge changed")
				}
				if _, e := held.Respond(ctxTest(t), inboundFailure()); e != nil {
					t.Fatal(e)
				}
				if _, e := readWire(ctxTest(t), remote, co); e != nil {
					t.Fatal(e)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRejectLengthClosedBoundaries(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"default", "session", "proxy", "reply", "field", "count", "depth", "repeat", "ancestor-flags", "truncated", "bad-header", "bare", "address-zero", "address-other-empty", "address-domain"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				if kind == "default" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				if kind == "reply" {
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 64
				}
				if kind == "field" {
					cfg.Limits.MaxFieldBytes = 128
				}
				if kind == "count" {
					cfg.Limits.MaxAVPs = 16
				}
				if kind == "depth" {
					cfg.Limits.MaxDepth = 1
				}
				if kind == "repeat" {
					cfg.Limits.MaxRepeated = 1
				}
				address := kind == "address-zero" || kind == "address-other-empty" || kind == "address-domain"
				if address {
					cfg = roConfig(t)
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				sender, e := codec.New(cfg.Bundle, codec.Config{})
				if e != nil {
					t.Fatal(e)
				}
				req := inboundRequest(t, n, sender, 10021)
				path := lengthPath(avp.Key{Code: 277})
				data := []byte{1, 2, 3}
				if address {
					req = lengthAddressRequest(t, n, sender)
					path = lengthAddressPath()
					data = []byte{0, 0}
					if kind == "address-other-empty" {
						data[1] = 99
					}
					if kind == "address-domain" {
						data = []byte{0, 1, 192, 0, 2, 1}
					}
				}
				if kind == "field" {
					data = make([]byte, 129)
				}
				if kind == "count" || kind == "repeat" || kind == "ancestor-flags" || kind == "depth" {
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
					path = []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{VendorID: 10415, Code: 629}, 0}}
				}
				raw, _ := invalidRaw(t, sender, req, path, data, avp.Key{})
				if kind == "session" || kind == "proxy" {
					bad := lengthPath(avp.Key{Code: 263})
					payload := []byte{255}
					if kind == "proxy" {
						bad = lengthPath(avp.Key{Code: 284}, avp.Key{Code: 280})
						payload = []byte("bad_name")
					}
					body, _ := invalidEdit(t, raw[20:], bad, payload, avp.Key{})
					raw = append(raw[:20:20], body...)
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				if kind == "ancestor-flags" {
					d := wire.NewDecoder(raw[20:], cfg.Limits)
					seen := 0
					for {
						v, ok, e := d.Next()
						if e != nil || !ok {
							t.Fatal("missing parent")
						}
						if v.Key == path[0].key {
							seen++
							if seen == 2 {
								raw[20+v.Offset+4] |= 1
								break
							}
						}
					}
				}
				if kind == "truncated" {
					raw = raw[:len(raw)-4]
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				if kind == "bad-header" {
					d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
					for {
						v, ok, e := d.Next()
						if e != nil || !ok {
							t.Fatal("missing leaf")
						}
						if v.Key == path[0].key {
							wire.Put24(raw[20+v.Offset+5:20+v.Offset+8], 7)
							break
						}
					}
				}
				if kind == "bare" {
					if _, e = n.rejectFrame(raw, 1024, rejectDecode, avp.ErrValue); e == nil {
						t.Fatal("guessed bare error")
					}
					closeDuplicateNode(t, n)
					return
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if _, e = readAdapter(ctxTest(t), remote); e == nil {
					t.Fatal("unsafe response")
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}
