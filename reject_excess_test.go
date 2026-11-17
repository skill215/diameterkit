// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: First-excess AVP automatic rejection over TCP and TLS.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Insert two distinct extra instances immediately after the selected original.
// The expected quote must be the first extra, never the original or last extra.
func addExcess(t *testing.T, src []byte, path []missingSelector, payload []byte, marked bool) ([]byte, []byte) {
	t.Helper()
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
				if len(path) == 1 {
					out = append(out, part...)
					want = make([]byte, (head+len(payload)+3)&^3)
					w := wire.NewEncoder(want, wire.DefaultLimits())
					w.Bytes(v.Key, v.Flags, payload)
					if w.Err() != nil {
						t.Fatal(w.Err())
					}
					for i := head + len(payload); i < len(want); i++ {
						want[i] = 0xa5
					}
					out = append(out, want...)
					last := append([]byte(nil), want...)
					if len(payload) > 0 {
						last[head] ^= 0xff
					}
					out = append(out, last...)
					seen++
					continue
				}
				var nested []byte
				nested, want = addExcess(t, v.Data, path[1:], payload, false)
				part = append(part[:head], nested...)
				wire.Put24(part[5:8], uint32(len(part)))
				if marked {
					part[4] |= avp.MandatoryFlag
				}
			}
			seen++
		}
		out = append(out, part...)
	}
	if want == nil {
		t.Fatal("selector missing", path)
	}
	return out, want
}

func excessRaw(t *testing.T, co *codec.Codec, req pdu.Message, path []missingSelector, payload []byte, mark bool) ([]byte, []byte) {
	t.Helper()
	b := reqBytes(t, co, req)
	body, want := addExcess(t, b[20:], path, payload, mark)
	out := append(append([]byte(nil), b[:20]...), body...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out, want
}

func assertExcessPath(t *testing.T, b []byte, path []missingSelector, want []byte, marked bool) {
	t.Helper()
	d := wire.NewDecoder(b, wire.DefaultLimits())
	for i, s := range path {
		v, ok, e := d.Next()
		if e != nil || !ok || v.Key != s.key || d.Remaining() != 0 {
			t.Fatal("wrong quote hierarchy", v, e)
		}
		if i == len(path)-1 {
			if !bytes.Equal(b, want) {
				t.Fatalf("quote %x, want first excess %x", b, want)
			}
			return
		}
		if marked && i == 0 && v.Flags != 0xc0 {
			t.Fatal("wrong parent flags", v.Flags)
		}
		b = v.Data
		d, e = d.Child(b)
		if e != nil {
			t.Fatal(e)
		}
	}
}

func TestRejectExcessGenerated(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, tc := range []struct {
			name     string
			combined bool
			path     []missingSelector
			payload  []byte
			marked   bool
		}{
			{"top-scalar", false, []missingSelector{{avp.Key{Code: 277}, 0}}, []byte{0, 0, 0, 7}, false},
			{"binary-padding", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 3300}, 0}}, []byte{0, 255, 1}, false},
			{"second-parent", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{VendorID: 10415, Code: 630}, 0}}, []byte{0x11, 0x22, 0x33, 0x44}, true},
			{"non-vendor-child", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{Code: 266}, 0}}, []byte{0, 0, 0, 9}, true},
			{"deep-money", true, []missingSelector{{avp.Key{Code: 456}, 0}, {avp.Key{Code: 437}, 0}, {avp.Key{Code: 413}, 0}, {avp.Key{Code: 445}, 0}, {avp.Key{Code: 447}, 0}}, []byte{0, 1, 2, 3, 4, 5, 6, 7}, false},
			{"grouped-opaque", true, []missingSelector{{avp.Key{Code: 456}, 0}, {avp.Key{Code: 437}, 0}, {avp.Key{Code: 413}, 0}}, []byte{0xff}, false},
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
					req = combinedDuplicateRequest(t, n, co, "general", 9701, 9702)
				} else {
					req = inboundRequest(t, n, co, 9701)
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
				}
				raw, want := excessRaw(t, co, req, tc.path, tc.payload, tc.marked)
				_, err := co.Decode(raw)
				if !errors.Is(err, validation.ErrDuplicate) {
					t.Fatal(err)
				}
				q, ok := excessavp.Quote(raw[20:], err, cfg.Limits, 1024)
				if !ok {
					t.Fatal("no diagnostic", err)
				}
				assertExcessPath(t, q, tc.path, want, tc.marked)
				if _, e := n.rejectFrame(raw, 20, rejectDecode, err); e == nil {
					t.Fatal("reply budget bypass")
				}
				direct, e := n.rejectFrame(raw, 1024, rejectDecode, err)
				if e != nil {
					t.Fatal(e)
				}
				decoded, e := co.Decode(direct)
				if e != nil {
					t.Fatal("selected E grammar", e)
				}
				if decoded.Body.(*pdu.ErrorAnswer).FailedAVP.Value.Incomplete != (tc.name == "grouped-opaque") {
					t.Fatal("diagnostic completeness changed")
				}
				sendRejectRaw(t, remote, raw)
				ans := readReject(t, remote, req.Header, 5009)
				if !ans.FailedAVP.Present || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
					t.Fatal(ans)
				}
				assertExcessPath(t, ans.FailedAVP.Value.Data, tc.path, want, tc.marked)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
				if !p.Info().Ready {
					t.Fatal(p.Info())
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, e = p.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("delivered invalid request", e)
				}
			})
		}
	})
}

func TestRejectExcessClosedBoundaries(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"default", "session", "manual-proxy", "resource-repeats", "bad-parent-flags", "bad-length", "reply-budget"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				if kind == "default" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				if kind == "resource-repeats" {
					cfg.Limits.MaxRepeated = 1
				}
				if kind == "reply-budget" {
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 64
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 9703)
				path := []missingSelector{{avp.Key{Code: 277}, 0}}
				if kind == "session" {
					path = []missingSelector{{avp.Key{Code: 263}, 0}}
				}
				if kind == "manual-proxy" {
					path = []missingSelector{{avp.Key{Code: 284}, 0}, {avp.Key{Code: 280}, 0}}
				}
				if kind == "bad-parent-flags" {
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}
					path = []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 0}, {avp.Key{VendorID: 10415, Code: 630}, 0}}
				}
				raw, _ := excessRaw(t, co, req, path, []byte{0, 0, 0, 1}, false)
				if kind == "bad-parent-flags" || kind == "bad-length" {
					d := wire.NewDecoder(raw[20:], cfg.Limits)
					seen := 0
					for {
						v, ok, e := d.Next()
						if e != nil {
							t.Fatal(e)
						}
						if !ok {
							break
						}
						if v.Key == path[0].key {
							seen++
							start := 20 + v.Offset
							if kind == "bad-parent-flags" {
								raw[start+4] |= 1
								break
							}
							if seen == 2 {
								wire.Put24(raw[start+5:start+8], 0xffffff)
								break
							}
						}
					}
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

func TestRejectExcessEmergencyReserve(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, tls)
		held, _ := receiveInbound(t, n, p, remote, co, 9705)
		before := n.Resources()
		req := inboundRequest(t, n, co, 9706)
		path := []missingSelector{{avp.Key{Code: 277}, 0}}
		raw, want := excessRaw(t, co, req, path, []byte{0, 0, 0, 7}, false)
		sendRejectRaw(t, remote, raw)
		ans := readReject(t, remote, req.Header, 5009)
		assertExcessPath(t, ans.FailedAVP.Value.Data, path, want, false)
		waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
		after := n.Resources()
		if after.InboundOutstanding != 1 || after.InboundBytes != before.InboundBytes || after.FixedBytes != before.FixedBytes {
			t.Fatal("consumed held reservation", before, after)
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
