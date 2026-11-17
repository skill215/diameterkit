// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Generated missing-AVP diagnostics over real TCP and TLS.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Each selector names one direct-parent occurrence; no recursive key search.
type missingSelector struct {
	key        avp.Key
	occurrence int
}

func omitSelected(t *testing.T, data []byte, path []missingSelector, markParent bool) []byte {
	t.Helper()
	d := wire.NewDecoder(data, wire.DefaultLimits())
	out := []byte{}
	seen := 0
	changed := false
	for {
		v, ok, err := d.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
		n := (int(wire.Uint24(data[v.Offset+5:v.Offset+8])) + 3) &^ 3
		part := append([]byte(nil), data[v.Offset:v.Offset+n]...)
		if v.Key == path[0].key {
			if seen == path[0].occurrence {
				changed = true
				if len(path) == 1 {
					seen++
					continue
				}
				head := 8
				if v.Key.VendorID != 0 {
					head = 12
				}
				part = append(part[:head], omitSelected(t, v.Data, path[1:], false)...)
				if markParent {
					part[4] |= avp.MandatoryFlag
				}
				wire.Put24(part[5:8], uint32(len(part)))
			}
			seen++
		}
		out = append(out, part...)
	}
	if !changed {
		t.Fatal("fixture selector not found", path)
	}
	return out
}

func missingRaw(t *testing.T, co *codec.Codec, req pdu.Message, path []missingSelector, mark bool) []byte {
	t.Helper()
	raw := reqBytes(t, co, req)
	out := append(append([]byte(nil), raw[:20]...), omitSelected(t, raw[20:], path, mark)...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}

func assertMissingChain(t *testing.T, data []byte, path []missingSelector, width int, firstFlags byte) {
	t.Helper()
	d := wire.NewDecoder(data, wire.DefaultLimits())
	for i, s := range path {
		v, ok, e := d.Next()
		if e != nil || !ok || v.Key != s.key || d.Remaining() != 0 {
			t.Fatal("wrong diagnostic path", v, e)
		}
		if i == 0 && v.Flags != firstFlags {
			t.Fatal("wrong parent instance flags", v.Flags)
		}
		if i == len(path)-1 {
			if !bytes.Equal(v.Data, make([]byte, width)) {
				t.Fatal("nonzero or wrong-width placeholder", v)
			}
			return
		}
		d, e = d.Child(v.Data)
		if e != nil {
			t.Fatal(e)
		}
	}
}

func TestRejectDeepMissingGenerated(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, tc := range []struct {
			name     string
			combined bool
			path     []missingSelector
			width    int
			flags    byte
		}{
			{"vendor", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{Code: 266}, 0}}, 4, 0xc0},
			{"feature-id", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{VendorID: 10415, Code: 629}, 0}}, 4, 0xc0},
			{"feature-list", false, []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 1}, {avp.Key{VendorID: 10415, Code: 630}, 0}}, 4, 0xc0},
			{"application", true, []missingSelector{{avp.Key{Code: 258}, 0}}, 4, 0x40},
			{"time", true, []missingSelector{{avp.Key{Code: 55}, 0}}, 4, 0x40},
			{"subscription", true, []missingSelector{{avp.Key{Code: 443}, 1}, {avp.Key{Code: 450}, 0}}, 4, 0x40},
			{"money", true, []missingSelector{{avp.Key{Code: 456}, 0}, {avp.Key{Code: 437}, 0}, {avp.Key{Code: 413}, 0}, {avp.Key{Code: 445}, 0}, {avp.Key{Code: 447}, 0}}, 8, 0x40},
			{"exponent", true, []missingSelector{{avp.Key{Code: 456}, 0}, {avp.Key{Code: 437}, 0}, {avp.Key{Code: 413}, 0}, {avp.Key{Code: 445}, 0}, {avp.Key{Code: 429}, 0}}, 4, 0x40},
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
					req = combinedDuplicateRequest(t, n, co, "general", 9601, 9602)
				} else {
					req = inboundRequest(t, n, co, 9601)
					req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}, {VendorID: 10415, FeatureListID: 2, FeatureList: 2}}
				}
				raw := missingRaw(t, co, req, tc.path, !tc.combined)
				_, err := co.Decode(raw)
				if !errors.Is(err, validation.ErrMissing) {
					t.Fatal(err)
				}
				if bytes.Contains([]byte(err.Error()), []byte("001010123456789")) {
					t.Fatal("subscriber leaked into error")
				}
				quote, ok := missingavp.Quote(err, cfg.Limits, 1024)
				if !ok {
					t.Fatal("no proven diagnostic", err)
				}
				assertMissingChain(t, quote, tc.path, tc.width, tc.flags)
				if _, e := n.rejectFrame(raw, 20, rejectDecode, err); e == nil {
					t.Fatal("reply budget bypass")
				}
				sendRejectRaw(t, remote, raw)
				answer := readReject(t, remote, req.Header, 5005)
				if !answer.FailedAVP.Present || !answer.SessionID.Present || len(answer.ProxyInfo) != 1 {
					t.Fatal(answer)
				}
				assertMissingChain(t, answer.FailedAVP.Value.Data, tc.path, tc.width, tc.flags)
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
				if !p.Info().Ready {
					t.Fatal(p.Info())
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, e := p.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("invalid request delivered", e)
				}
			})
		}
	})
}

func TestRejectDeepMissingFailsClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, kind := range []string{"default-policy", "duplicate-session", "bad-proxy", "reserved-parent"} {
			t.Run(kind, func(t *testing.T) {
				cfg := rejectConfig(t)
				if kind == "default-policy" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, tls)
				req := inboundRequest(t, n, co, 9603)
				req.Body.(*s6c.SRR).SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}
				path := []missingSelector{{avp.Key{VendorID: 10415, Code: 628}, 0}, {avp.Key{VendorID: 10415, Code: 630}, 0}}

				raw := missingRaw(t, co, req, path, false)
				d := wire.NewDecoder(raw[20:], cfg.Limits)
				for {
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					start := 20 + v.Offset
					switch kind {
					case "duplicate-session":
						if v.Key == (avp.Key{Code: 263}) {
							n := (int(wire.Uint24(raw[start+5:start+8])) + 3) &^ 3
							raw = append(raw, raw[start:start+n]...)
						}
					case "bad-proxy":
						if v.Key == (avp.Key{Code: 284}) {
							raw[start+8] = 0xff
						}
					case "reserved-parent":
						if v.Key == (avp.Key{VendorID: 10415, Code: 628}) {
							raw[start+4] |= 1
						}
					}
				}
				wire.Put24(raw[1:4], uint32(len(raw)))
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if b, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatal(fmt.Sprintf("unsafe answer: %x", b))
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
			})
		}
	})
}

func TestMissingUnreviewedWholeGroupRemainsUnsupported(t *testing.T) {
	cfg := combinedConfig(t)
	n, co := newInboundNode(t, cfg)
	req := combinedDuplicateRequest(t, n, co, "general", 9605, 9606)
	// RequestMSCC has contextual requirements and no reviewed minimum grammar.
	path := []missingSelector{{avp.Key{Code: 456}, 0}}
	raw := missingRaw(t, co, req, path, false)
	_, err := co.Decode(raw)
	if !errors.Is(err, validation.ErrMissing) {
		t.Fatal(err)
	}
	if b, ok := missingavp.Quote(err, cfg.Limits, 1024); ok {
		t.Fatal("invented Grouped payload", b)
	}
	if _, err = n.rejectFrame(raw, 1024, rejectDecode, err); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}
