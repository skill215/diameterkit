// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded 5004/5001 automatic rejection over TCP and TLS.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"context"
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
	"time"
)

func invalidPart(t *testing.T, k avp.Key, flags byte, data []byte) []byte {
	t.Helper()
	head := 8
	if k.VendorID != 0 {
		head = 12
	}
	b := make([]byte, (head+len(data)+3)&^3)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	w.Bytes(k, flags, data)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	for i := head + len(data); i < len(b); i++ {
		b[i] = 0xa5
	}
	return b
}

// Replace one exact instance, or append distinct unknown Mandatory instances to
// the selected parent. Later instances must never replace the first failure.
func invalidEdit(t *testing.T, src []byte, path []missingSelector, payload []byte, unknown avp.Key) ([]byte, []byte) {
	t.Helper()
	if len(path) == 0 {
		first := invalidPart(t, unknown, avp.MandatoryFlag|func() byte {
			if unknown.VendorID != 0 {
				return avp.VendorFlag
			}
			return 0
		}(), payload)
		last := invalidPart(t, unknown, first[4], []byte{9, 8, 7})
		out := append(append(append([]byte(nil), src...), first...), last...)
		return out, first
	}
	d := wire.NewDecoder(src, wire.DefaultLimits())
	var out, want []byte
	seen := 0
	for {
		v, ok, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !ok {
			break
		}
		head := 8
		if v.Key.VendorID != 0 {
			head = 12
		}
		size := (head + len(v.Data) + 3) &^ 3
		part := append([]byte(nil), src[v.Offset:v.Offset+size]...)
		if v.Key == path[0].key {
			if seen == path[0].occurrence {
				if len(path) == 1 && unknown == (avp.Key{}) {
					part = invalidPart(t, v.Key, v.Flags, payload)
					want = part
				} else {
					var nested []byte
					nested, want = invalidEdit(t, v.Data, path[1:], payload, unknown)
					part = append(part[:head], nested...)
					wire.Put24(part[5:8], uint32(len(part)))
				}
			}
			seen++
		}
		out = append(out, part...)
	}
	if want == nil {
		t.Fatal("missing selector", path)
	}
	return out, want
}

func invalidRaw(t *testing.T, co *codec.Codec, req pdu.Message, path []missingSelector, data []byte, unknown avp.Key) ([]byte, []byte) {
	t.Helper()
	raw := reqBytes(t, co, req)
	body, want := invalidEdit(t, raw[20:], path, data, unknown)
	out := append(raw[:20:20], body...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out, want
}

func TestRejectInvalidGenerated(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, tc := range []struct {
			name     string
			combined bool
			path     []missingSelector
			payload  []byte
			unknown  avp.Key
			code     uint32
		}{
			{"enum", false, []missingSelector{{avp.Key{Code: 277}, 0}}, []byte{0, 0, 0, 7}, avp.Key{}, 5004},
			{"utf8-padding", false, []missingSelector{{avp.Key{Code: 1}, 0}}, []byte{0xff, 0, 1}, avp.Key{}, 5004},
			{"identity", false, []missingSelector{{avp.Key{Code: 264}, 0}}, []byte("bad_name"), avp.Key{}, 5004},
			{"tbcd", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 3300}, 0}}, []byte{0xaa}, avp.Key{}, 5004},
			{"second-route-record", false, []missingSelector{{avp.Key{Code: 282}, 1}}, []byte("bad_name"), avp.Key{}, 5004},
			{"unknown-top", false, nil, []byte{0, 255, 1}, avp.Key{Code: 999999}, 5001},
			{"unknown-vendor-second-parent", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}}, []byte{0, 255, 1}, avp.Key{VendorID: 55555, Code: 999999}, 5001},
			{"unknown-deep", true, []missingSelector{{avp.Key{Code: 456}, 0}, {avp.Key{Code: 437}, 0}, {avp.Key{Code: 413}, 0}, {avp.Key{Code: 445}, 0}}, []byte{0xff}, avp.Key{Code: 999999}, 5001},
		} {
			t.Run(tc.name, func(t *testing.T) {
				cfg := rejectConfig(t)
				if tc.combined {
					c := combinedConfig(t)
					cfg.Bundle = c.Bundle
					cfg.Peers = c.Peers
					cfg.Capabilities = c.Capabilities
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				var req pdu.Message
				if tc.combined {
					req = combinedDuplicateRequest(t, n, co, "general", 9801, 9802)
				} else {
					req = inboundRequest(t, n, co, 9801)
					if tc.name == "utf8-padding" {
						req.Body.(*s6c.SRR).UserName = avp.Some("001010123456789")
						req.Body.(*s6c.SRR).MSISDN = avp.Optional[[]byte]{}
					}
					req.Body.(*s6c.SRR).RouteRecord = []string{"first.example", "second.example"}
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
				}
				raw, want := invalidRaw(t, co, req, tc.path, tc.payload, tc.unknown)
				path := append([]missingSelector(nil), tc.path...)
				if tc.code == 5001 {
					path = append(path, missingSelector{tc.unknown, 0})
				}
				_, err := co.Decode(raw)
				if err == nil {
					t.Fatal("accepted invalid request")
				}
				if tc.code == 5001 && !errors.Is(err, validation.ErrUnsupported) {
					t.Fatal(err)
				}
				q, code, ok := excessavp.QuoteInvalid(raw[20:], err, cfg.Limits, 1024)
				if !ok || code != tc.code {
					t.Fatal("missing evidence", err, code)
				}
				assertExcessPath(t, q, path, want, false)
				if _, e := n.rejectFrame(raw, 20, rejectDecode, err); e == nil {
					t.Fatal("reply budget bypass")
				}
				direct, e := n.rejectFrame(raw, 1024, rejectDecode, err)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = co.Decode(direct); e != nil {
					t.Fatal("selected E grammar", e)
				}
				sendRejectRaw(t, remote, raw)
				ans := readReject(t, remote, req.Header, tc.code)
				if !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
					t.Fatal("lost envelope")
				}
				assertExcessPath(t, ans.FailedAVP.Value.Data, path, want, false)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
				if !p.Info().Ready {
					t.Fatal("peer closed")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, e = p.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("invalid delivery", e)
				}
			})
		}
	})
}

func TestRejectInvalidClosedBoundaries(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"default", "session", "proxy", "manual-unsupported", "reply-budget", "resource-repeat", "bad-parent-flags", "truncated"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				if kind == "default" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				if kind == "reply-budget" {
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 64
				}
				if kind == "resource-repeat" {
					cfg.Limits.MaxRepeated = 1
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 9803)
				path := []missingSelector{{avp.Key{Code: 277}, 0}}
				payload := []byte{0, 0, 0, 7}
				unknown := avp.Key{}
				switch kind {
				case "session":
					path[0].key.Code = 263
					payload = []byte{0xff}
				case "proxy":
					path = []missingSelector{{avp.Key{Code: 284}, 0}, {avp.Key{Code: 280}, 0}}
					payload = []byte("bad_name")
				case "manual-unsupported":
					req.Body.(*s6c.SRR).SMRPMTI = avp.Some(uint32(0))
					path[0].key = avp.Key{VendorID: 10415, Code: 3308}
					payload = []byte{0, 0, 0, 2}
				case "resource-repeat", "bad-parent-flags":
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
					path = []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}}
					unknown = avp.Key{Code: 999999}
				}
				encoder := co
				if kind == "resource-repeat" {
					var e error
					encoder, e = codec.New(cfg.Bundle, codec.Config{})
					if e != nil {
						t.Fatal(e)
					}
				}
				raw, _ := invalidRaw(t, encoder, req, path, payload, unknown)
				if kind == "bad-parent-flags" {
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
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if _, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatal("unsafe response")
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
			})
		}
	})
}

func TestRejectInvalidEmergencyAndOptional(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, code := range []uint32{5004, 5001} {
			t.Run(func() string {
				if code == 5004 {
					return "value"
				}
				return "unsupported"
			}(), func(t *testing.T) {
				cfg := rejectConfig(t)
				cfg.Resources.InboundClasses[0].Slots = 1
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				held, _ := receiveInbound(t, n, p, remote, co, 9805)
				before := n.Resources()
				req := inboundRequest(t, n, co, 9806)
				path := []missingSelector{{avp.Key{Code: 277}, 0}}
				unknown := avp.Key{}
				if code == 5001 {
					path = nil
					unknown = avp.Key{Code: 999999}
				}
				raw, want := invalidRaw(t, co, req, path, []byte{0, 0, 0, 7}, unknown)
				sendRejectRaw(t, remote, raw)
				ans := readReject(t, remote, req.Header, code)
				if code == 5001 {
					path = append(path, missingSelector{unknown, 0})
				}
				assertExcessPath(t, ans.FailedAVP.Value.Data, path, want, false)
				waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
				after := n.Resources()
				if after.InboundOutstanding != 1 || after.InboundBytes != before.InboundBytes || after.FixedBytes != before.FixedBytes {
					t.Fatal("held reservation changed")
				}
				if _, e := held.Respond(ctxTest(t), inboundFailure()); e != nil {
					t.Fatal(e)
				}
				if _, e := readWire(ctxTest(t), remote, co); e != nil {
					t.Fatal(e)
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
			})
		}
		// A non-M unknown value stays opaque and is delivered normally.
		n, co := newInboundNode(t, rejectConfig(t))
		p, remote := networkConnectNode(t, n, co, tls)
		req := inboundRequest(t, n, co, 9807)
		raw := reqBytes(t, co, req)
		raw = append(raw, invalidPart(t, avp.Key{VendorID: 55555, Code: 999999}, avp.VendorFlag, []byte{255})...)
		wire.Put24(raw[1:4], uint32(len(raw)))
		sendRejectRaw(t, remote, raw)
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		if _, e = readWire(ctxTest(t), remote, co); e != nil {
			t.Fatal(e)
		}
	})
}
