package diameter

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
)

func scapAdvertisementFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/codec/b78-scap-" + name + ".hex")
	if e != nil {
		t.Fatal(e)
	}
	b, e = hex.DecodeString(strings.TrimSpace(string(b)))
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func scapReadHandshake(ctx context.Context, t *testing.T, n *Node, c transport.Conn, co *codec.Codec) (pdu.Message, error) {
	t.Helper()
	if n.peers["dra"].scap1Advertisement != SCAP1LegacyDualAdvertisement {
		return readWire(ctx, c, co)
	}
	prefix, e := c.ReadPrefix(ctx, time.Second)
	if e != nil {
		return pdu.Message{}, e
	}
	_, size, e := wire.Header(prefix[:], co.Limits())
	if e != nil {
		return pdu.Message{}, e
	}
	b := make([]byte, size)
	if e = c.ReadFrameInto(ctx, b); e != nil {
		return pdu.Message{}, e
	}
	// Observe the actual outbound wire, including both groups, before projection.
	if bytes.Count(b, []byte(scap1DualGroup)) != 1 || bytes.Count(b, []byte(scap1StrictGroup)) != 1 {
		return pdu.Message{}, ErrProtocol
	}
	b, e = projectSCAP1Advertisement(b, co.Limits(), false)
	if e != nil {
		return pdu.Message{}, e
	}
	return co.Decode(b)
}
func TestSCAP1LegacyAdvertisementProjection(t *testing.T) {
	n, co := newInboundNode(t, scapRuntimeConfig(t))
	for _, suffix := range []string{"cer", "cea"} {
		t.Run(suffix, func(t *testing.T) {
			strict := scapAdvertisementFixture(t, "strict-"+suffix)
			dual := scapAdvertisementFixture(t, "dual-"+suffix)
			if _, e := co.Decode(dual); e == nil {
				t.Fatal("standard codec accepted dual VSAI")
			}
			owned := append([]byte(nil), dual...)
			out, e := projectSCAP1Advertisement(owned, co.Limits(), false)
			if e != nil || !bytes.Equal(out, strict) || &out[0] != &owned[0] {
				t.Fatal("receive projection", e)
			}
			m, e := co.Decode(out)
			if e != nil {
				t.Fatal(e)
			}
			var vs []pdu.VendorSpecificApplicationID
			if r, ok := m.Body.(*base.CER); ok {
				vs = r.VendorSpecificApplicationID
			} else {
				vs = m.Body.(*base.CEA).VendorSpecificApplicationID
			}
			if !hasSCAP1Accounting(vs) || len(vs) != 1 || vs[0].AuthApplicationID.Present {
				t.Fatal("dual group grants auth", vs)
			}
			b := make([]byte, len(strict), len(dual))
			copy(b, strict)
			out, e = projectSCAP1Advertisement(b, co.Limits(), true)
			if e != nil || !bytes.Equal(out, dual) || &out[0] != &b[0] {
				t.Fatal("send projection", e)
			}
			clear(owned)
			if !hasSCAP1Accounting(vs) {
				t.Fatal("borrowed capabilities")
			}
		})
	}
	_ = n
	for _, name := range []string{"orphan", "duplicate", "reversed", "split", "wrong-vendor", "wrong-auth", "wrong-acct", "child-flags", "outer-flags", "child-order", "extra", "width", "strict-twice"} {
		t.Run(name, func(t *testing.T) {
			b, e := projectSCAP1Advertisement(scapAdvertisementFixture(t, name+"-cer"), co.Limits(), false)
			if e == nil {
				_, e = co.Decode(b)
			}
			if e == nil {
				t.Fatal("malformed advertisement accepted")
			}
		})
	}
	for _, name := range []string{"frame", "avps", "repeated", "field", "depth"} {
		t.Run(name, func(t *testing.T) {
			b := scapAdvertisementFixture(t, "dual-cer")
			l := co.Limits()
			switch name {
			case "frame":
				l.MaxFrameBytes = len(b) - 4
				l.MaxFieldBytes = l.MaxFrameBytes
			case "avps":
				l.MaxAVPs = 10
			case "repeated":
				l.MaxRepeated = 1
			case "field":
				l.MaxFieldBytes = 35
			case "depth":
				l.MaxDepth = 1
			}
			if name == "depth" { // One child depth is permitted; zero is invalid configuration.
				if _, e := projectSCAP1Advertisement(b, l, false); e != nil {
					t.Fatal(e)
				}
				l.MaxDepth = 0
			}
			if _, e := projectSCAP1Advertisement(b, l, false); e == nil {
				t.Fatal("limit bypass")
			}
		})
	}
	b := scapAdvertisementFixture(t, "strict-cer")
	if _, e := projectSCAP1Advertisement(b[:len(b):len(b)], co.Limits(), true); !errors.Is(e, wire.ErrLimit) {
		t.Fatal("unreserved TX growth", e)
	}
}
func TestSCAP1LegacyAdvertisementConfiguration(t *testing.T) {
	for name, mut := range map[string]func(*NodeConfig){
		"unknown": func(c *NodeConfig) { c.Peers[0].SCAP1Advertisement = 99 },
		"no accounting policy": func(c *NodeConfig) {
			c.Peers[0].Commands = inboundConfig(t).Peers[0].Commands
			c.Peers[0].InboundCommands = nil
		},
		"duplicate local": func(c *NodeConfig) {
			c.Capabilities.VendorSpecificApplicationID = append(c.Capabilities.VendorSpecificApplicationID, scapCapability())
		},
		"other vendor193 auth": func(c *NodeConfig) {
			c.Capabilities.VendorSpecificApplicationID = append(c.Capabilities.VendorSpecificApplicationID, pdu.VendorSpecificApplicationID{VendorID: 193, AuthApplicationID: avp.Some(uint32(4))})
		},
		"repeat budget": func(c *NodeConfig) { c.Limits.MaxRepeated = 1 },
		"field budget":  func(c *NodeConfig) { c.Limits.MaxFieldBytes = 35 },
	} {
		t.Run(name, func(t *testing.T) {
			c := scapRuntimeConfig(t)
			c.Peers[0].SCAP1Advertisement = SCAP1LegacyDualAdvertisement
			mut(&c)
			n, e := NewNode(c)
			if n != nil {
				n.Close()
			}
			if !errors.Is(e, ErrConfig) {
				t.Fatal(e)
			}
		})
	}
	c := scapRuntimeConfig(t)
	c.Peers[0].SCAP1Advertisement = SCAP1LegacyDualAdvertisement
	n, _ := newInboundNode(t, c)
	c.Peers[0].SCAP1Advertisement = SCAP1StandardAdvertisement
	if n.peers["dra"].scap1Advertisement != SCAP1LegacyDualAdvertisement {
		t.Fatal("config alias")
	}
	if n.cfg.Resources.TransportBytes == 0 {
		t.Fatal("missing control reservation")
	}
}
func TestSCAP1LegacyAdvertisementNetwork(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, passive := range []bool{false, true} {
			for _, dual := range []bool{false, true} {
				name := "strict"
				if dual {
					name = "dual"
				}
				if passive {
					name += "/accept"
				} else {
					name += "/connect"
				}
				t.Run(name, func(t *testing.T) {
					cfg := scapRuntimeConfig(t)
					cfg.Peers[0].SCAP1Advertisement = SCAP1LegacyDualAdvertisement
					n, co := newInboundNode(t, cfg)
					p, remote, e := scapHandshake(t, n, co, tls, passive, LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{scapCapability()}}, dual)
					if e != nil || !p.Info().Ready {
						t.Fatal(e)
					}
					port := portTest(t, n, 1)
					id, e := p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
					if e != nil {
						t.Fatal(e)
					}
					req := routeRead(t, co, remote)
					ans, e := co.CompleteAnswer(req, scapAnswer(2001), serverOrigin)
					if e != nil {
						t.Fatal(e)
					}
					if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
						t.Fatal(e)
					}
					if r := routeCompletion(t, n, port, id); r.Outcome != AnswerReceived {
						t.Fatal(r)
					}
					for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
						if _, e = n.NewRoute(RouteConfig{Command: scapBody(t).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: mode, MaxAttempts: 2}); !errors.Is(e, ErrUnsupported) {
							t.Fatal("compat grants replay", e)
						}
					}
				})
			}
		}
	})
}
