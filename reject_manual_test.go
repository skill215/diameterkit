// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Manual Proxy/VSAI exact-instance diagnostics and TCP/TLS safety boundaries.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/internal/groupavp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func manualConfig(t *testing.T) NodeConfig {
	cfg := mixedDuplicateConfig(t, false)
	cfg.Resources.RejectSlots = 2
	cfg.Resources.EmergencyRequestMax = 4096
	cfg.Resources.EmergencyReplyMax = 4096
	for i := range cfg.Peers {
		for j := range cfg.Peers[i].InboundCommands {
			cfg.Peers[i].InboundCommands[j].FailurePolicy = BaseErrorFailures
		}
	}
	return cfg
}
func manualQuote(raw []byte, e error, l wire.Limits, budget int) ([]byte, uint32, bool) {
	if q, code, ok := groupavp.Quote(raw, e, l, budget); ok {
		return q, code, true
	}
	if q, ok := excessavp.Quote(raw, e, l, budget); ok {
		return q, 5009, true
	}
	if q, code, ok := excessavp.QuoteInvalid(raw, e, l, budget); ok {
		return q, code, true
	}
	if q, ok := missingavp.Quote(e, l, budget); ok {
		return q, 5005, true
	}
	return nil, 0, false
}
func manualFrame(t *testing.T, co *codec.Codec, req pdu.Message, kind string) ([]byte, []byte, []missingSelector, uint32) {
	path := lengthPath(avp.Key{Code: 260}, avp.Key{Code: 266})
	var raw, want []byte
	code := uint32(5014)
	switch kind {
	case "auth-short":
		path[1].key.Code = 258
	case "vendor-long":
	case "duplicate-vendor", "duplicate-auth":
		if kind == "duplicate-auth" {
			path[1].key.Code = 258
		}
		raw, want = excessRaw(t, co, req, path, []byte{0, 0, 0, 9}, false)
		code = 5009
	case "unknown":
		raw, want = invalidRaw(t, co, req, path[:1], []byte{1, 255, 2}, avp.Key{VendorID: 55555, Code: 999999})
		path[1].key = avp.Key{VendorID: 55555, Code: 999999}
		code = 5001
	case "missing-vendor":
		original := reqBytes(t, co, req)
		d := wire.NewDecoder(original[20:], wire.DefaultLimits())
		for {
			v, ok, e := d.Next()
			if e != nil || !ok {
				t.Fatal("missing VSAI", e)
			}
			if v.Key.Code == 260 {
				child := append(make([]byte, 20), v.Data...)
				child = withoutAVP(t, child, avp.Key{Code: 266})
				raw, _ = invalidRaw(t, co, req, path[:1], child[20:], avp.Key{})
				break
			}
		}
		want = invalidPart(t, avp.Key{Code: 266}, 64, make([]byte, 4))
		code = 5005
	}
	if raw == nil {
		data := []byte{1, 2, 3}
		if kind == "vendor-long" {
			data = make([]byte, 5)
		}
		raw, want = invalidRaw(t, co, req, path, data, avp.Key{})
	}
	return raw, want, path, code
}
func TestRejectManualVSAI(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, emergency := range []bool{false, true} {
			for _, kind := range []string{"vendor-short", "vendor-long", "auth-short", "duplicate-vendor", "duplicate-auth", "unknown", "missing-vendor"} {
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
					raw, want, path, code := manualFrame(t, co, req, kind)
					_, cause := co.Decode(raw)
					q, result, ok := manualQuote(raw[20:], cause, cfg.Limits, 1024)
					if !ok || result != code {
						t.Fatal("lost manual evidence", cause, result)
					}
					assertExcessPath(t, q, path, want, false)
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, code)
					if !ans.SessionID.Present || !base.EqualProxies(ans.ProxyInfo, req.Body.(*sh.UDR).ProxyInfo) {
						t.Fatal("envelope")
					}
					assertExcessPath(t, ans.FailedAVP.Value.Data, path, want, false)
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

func TestRejectManualProxyClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"host", "duplicate-host", "duplicate-state", "unknown", "missing-host", "missing-state"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := inboundRequest(t, n, co, 10211)
				b := req.Body.(*s6c.SRR)
				b.ProxyInfo = append(b.ProxyInfo, pdu.ProxyInfo{ProxyHost: "second.example", ProxyState: []byte{0, 255, 2}})
				path := []missingSelector{{avp.Key{Code: 284}, 1}, {avp.Key{Code: 280}, 0}}
				var raw, want []byte
				code := uint32(5004)
				switch kind {
				case "host":
					raw, want = invalidRaw(t, co, req, path, []byte("bad_name"), avp.Key{})
				case "duplicate-host", "duplicate-state":
					if kind == "duplicate-state" {
						path[1].key.Code = 33
					}
					raw, want = excessRaw(t, co, req, path, []byte{1, 255, 3}, false)
					code = 5009
				case "unknown":
					path[1].key = avp.Key{VendorID: 55555, Code: 999999}
					raw, want = invalidRaw(t, co, req, path[:1], []byte{1, 255, 3}, path[1].key)
					code = 5001
				default:
					data := invalidPart(t, avp.Key{Code: 33}, 64, []byte{0, 1})
					if kind == "missing-state" {
						data = invalidPart(t, avp.Key{Code: 280}, 64, []byte("second.example"))
					}
					raw, _ = invalidRaw(t, co, req, path[:1], data, avp.Key{})
					code = 5005
					want = invalidPart(t, avp.Key{Code: 280}, 64, []byte{0})
					clear(want[9:])
					if kind == "missing-state" {
						path[1].key.Code = 33
						want = invalidPart(t, avp.Key{Code: 33}, 64, nil)
					}
				}
				_, cause := co.Decode(raw)
				q, result, ok := manualQuote(raw[20:], cause, cfg.Limits, 1024)
				if code == 0 {
					if ok || !errors.Is(cause, validation.ErrMissing) {
						t.Fatal("invented variable omission", cause)
					}
				} else {
					if !ok || result != code {
						t.Fatal("manual instance", cause, result)
					}
					assertExcessPath(t, q, path, want, false)
				}
				if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); e == nil {
					t.Fatal("fabricated proxy envelope")
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if _, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatal("unsafe reply")
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestManualRepeatedVSAI(t *testing.T) {
	cfg := manualConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	// CER uses the same manual VSAI decoder, repeated at the top level. This is
	// offline codec evidence; handshake-specific answer mapping remains separate.
	b := &base.CER{OriginHost: "a.example", OriginRealm: "example", HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, VendorID: 0, ProductName: "test", VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{sh.Application(), {VendorID: 193, AcctApplicationID: avp.Some(uint32(19302))}}}
	req := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257}, Body: b}
	for _, code := range []uint32{266, 259} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			path := []missingSelector{{avp.Key{Code: 260}, 1}, {avp.Key{Code: code}, 0}}
			for _, duplicate := range []bool{false, true} {
				var raw, want []byte
				if duplicate {
					raw, want = excessRaw(t, co, req, path, []byte{1, 2, 3, 4}, false)
				} else {
					raw, want = invalidRaw(t, co, req, path, []byte{1, 2, 3}, avp.Key{})
				}
				_, cause := co.Decode(raw)
				q, result, ok := manualQuote(raw[20:], cause, cfg.Limits, 1024)
				expected := uint32(5014)
				if duplicate {
					expected = 5009
				}
				if !ok || result != expected {
					t.Fatal(cause, result)
				}
				assertExcessPath(t, q, path, want, false)
				frozen := append([]byte(nil), q...)
				clear(raw)
				if !bytes.Equal(q, frozen) {
					t.Fatal("borrowed quote")
				}
			}
		})
	}
}

func TestManualProofBounds(t *testing.T) {
	cfg := manualConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	req := mixedDuplicateRequest(t, n, co, 2, 10231, 20231, gmscOrigin)
	for _, kind := range []string{"valid", "budget", "field", "count", "depth", "frame", "bare", "path", "prefix", "flags", "valid-width", "truncated"} {
		t.Run(kind, func(t *testing.T) {
			raw, _, _, _ := manualFrame(t, co, req, "vendor-short")
			_, cause := co.Decode(raw)
			l := cfg.Limits
			budget := 1024
			switch kind {
			case "budget":
				budget = 19
			case "field":
				l.MaxFieldBytes = 16
			case "count":
				l.MaxAVPs = 2
				l.MaxRepeated = 2
			case "depth":
				l.MaxDepth = 1
			case "frame":
				l.MaxFrameBytes = 40
				l.MaxFieldBytes = 40
			case "bare":
				cause = validation.At(avp.Key{Code: 260}, validation.At(avp.Key{Code: 266}, avp.ErrValue))
			case "path":
				var field *validation.FieldError
				if !errors.As(cause, &field) {
					t.Fatal(cause)
				}
				copy := *field
				copy.Path = append([]avp.Key(nil), field.Path...)
				copy.Path[0].Code++
				cause = &copy
			case "prefix":
				raw = append(append(raw[:20:20], invalidPart(t, avp.Key{Code: 999998}, 0, []byte{1})...), raw[20:]...)
			case "flags", "valid-width", "truncated":
				d := wire.NewDecoder(raw[20:], cfg.Limits)
				for {
					v, ok, e := d.Next()
					if e != nil || !ok {
						t.Fatal("VSAI not found")
					}
					if v.Key.Code == 260 {
						pos := 20 + v.Offset
						if kind == "flags" {
							raw[pos+4] ^= 0x40
						}
						if kind == "valid-width" {
							wire.Put24(raw[pos+8+5:pos+8+8], 12)
						}
						if kind == "truncated" {
							raw = raw[:pos+8+9]
						}
						break
					}
				}
			}
			q, result, ok := manualQuote(raw[20:], cause, l, budget)
			if ok != (kind == "valid") || !ok && (q != nil || result != 0) {
				t.Fatal("unsafe proof accepted", kind, result)
			}
		})
	}
}

func TestRejectManualVSAIClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"default", "reply", "repeat", "field", "count", "depth", "session"} {
			t.Run(kind, func(t *testing.T) {
				cfg := manualConfig(t)
				// Isolate this policy from cross-peer duplicate-domain equality checks.
				cfg.Peers = cfg.Peers[:1]
				policy := cfg.Peers[0].InboundCommands[2]
				policy.NoCache = true
				policy.Duplicate = DuplicatePolicy{}
				cfg.Peers[0].InboundCommands = []InboundPolicy{policy}
				cfg.Resources.DuplicateClasses = nil
				cfg.Capabilities.AuthApplicationID = nil
				cfg.Capabilities.VendorSpecificApplicationID = []pdu.VendorSpecificApplicationID{sh.Application()}
				switch kind {
				case "default":
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				case "reply":
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 64
				case "repeat":
					cfg.Limits.MaxRepeated = 1
				case "field":
					cfg.Limits.MaxFieldBytes = 128
				case "count":
					cfg.Limits.MaxAVPs = 12
					cfg.Limits.MaxRepeated = 12
				case "depth":
					cfg.Limits.MaxDepth = 1
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				sender, e := codec.New(cfg.Bundle, codec.Config{})
				if e != nil {
					t.Fatal(e)
				}
				req := mixedDuplicateRequest(t, n, sender, 2, 10241, 20241, gmscOrigin)
				if kind == "count" {
					for i := 0; i < 12; i++ {
						req.Body.(*sh.UDR).Extra = append(req.Body.(*sh.UDR).Extra, avp.Raw{Key: avp.Key{Code: uint32(999900 + i)}, Data: []byte{1}})
					}
				}
				selected := "vendor-short"
				if kind == "repeat" {
					selected = "duplicate-vendor"
				}
				raw, _, _, _ := manualFrame(t, sender, req, selected)
				if kind == "both-apps" || kind == "neither-app" {
					data := invalidPart(t, avp.Key{Code: 266}, 64, []byte{0, 0, 40, 175})
					if kind == "both-apps" {
						data = append(data, invalidPart(t, avp.Key{Code: 258}, 64, []byte{1, 0, 0, 1})...)
						data = append(data, invalidPart(t, avp.Key{Code: 259}, 64, []byte{1, 0, 0, 1})...)
					}
					raw, _ = invalidRaw(t, sender, req, lengthPath(avp.Key{Code: 260}), data, avp.Key{})
				}
				if kind == "field" {
					raw, _ = invalidRaw(t, sender, req, lengthPath(avp.Key{Code: 260}, avp.Key{Code: 266}), make([]byte, 129), avp.Key{})
				}
				if kind == "session" {
					body, _ := invalidEdit(t, raw[20:], lengthPath(avp.Key{Code: 263}), []byte{255}, avp.Key{})
					raw = append(raw[:20:20], body...)
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if _, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatal("unsafe manual reply")
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}
