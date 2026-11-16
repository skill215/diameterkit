package diameter

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
)

func TestSCAP1BindingConfiguration(t *testing.T) {
	cases := map[string]func(*NodeConfig){
		"absent local": func(c *NodeConfig) { c.Capabilities.VendorSpecificApplicationID = nil },
		"wrong vendor": func(c *NodeConfig) { c.Capabilities.VendorSpecificApplicationID[0].VendorID = 10415 },
		"auth instead": func(c *NodeConfig) {
			c.Capabilities.VendorSpecificApplicationID[0] = pdu.VendorSpecificApplicationID{VendorID: 193, AuthApplicationID: avp.Some(uint32(19302))}
		},
		"dual local": func(c *NodeConfig) {
			c.Capabilities.VendorSpecificApplicationID[0].AuthApplicationID = avp.Some(uint32(19302))
		},
		"no outbound optin": func(c *NodeConfig) { c.Peers[0].Commands[0].ApplicationBinding = HeaderApplication },
		"no inbound optin":  func(c *NodeConfig) { c.Peers[0].InboundCommands[0].ApplicationBinding = HeaderApplication },
		"other control":     func(c *NodeConfig) { c.Peers[0].Commands[0].Command.Code = 280 },
		"other application": func(c *NodeConfig) { c.Peers[0].Commands[0].Command.ApplicationID = 19302 },
		"unknown binding":   func(c *NodeConfig) { c.Peers[0].Commands[0].ApplicationBinding = ApplicationBinding(99) },
		"realm only":        func(c *NodeConfig) { c.Peers[0].Commands[0].DestinationHost = "" },
		"source mismatch":   func(c *NodeConfig) { c.Peers[0].Commands[0].Sources.Terminals = []pdu.Origin{draOrigin} },
		"multiple terminals": func(c *NodeConfig) {
			c.Peers[0].Commands[0].Sources.Terminals = append(c.Peers[0].Commands[0].Sources.Terminals, draOrigin)
		},
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			c := scapRuntimeConfig(t)
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
	n, _ := newInboundNode(t, c)
	c.Capabilities.VendorSpecificApplicationID[0].AcctApplicationID.Value = 4
	c.Peers[0].Commands[0].ApplicationBinding = HeaderApplication
	if !hasSCAP1Accounting(n.cer().VendorSpecificApplicationID) || n.peers["dra"].policies[scapBody(t).CommandKey()].ApplicationBinding != SCAP1Accounting {
		t.Fatal("caller config alias")
	}
	for _, mode := range []ExchangePolicy{TransportFailover, ExplicitRetransmit} {
		if _, e := n.NewRoute(RouteConfig{Command: scapBody(t).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: mode, MaxAttempts: 2}); !errors.Is(e, ErrUnsupported) {
			t.Fatal("SCAP replay", mode, e)
		}
	}
}

// Append the actual old second VSAI shape (both Auth and Acct) to a valid control
// frame independently of the strict local encoder. No production grammar override.
func scapDualAdvertisement(raw []byte) []byte {
	b := append([]byte(nil), raw...)
	b = append(b, 0, 0, 1, 4, 64, 0, 0, 44, 0, 0, 1, 10, 64, 0, 0, 12, 0, 0, 0, 193,
		0, 0, 1, 2, 64, 0, 0, 12, 0, 0, 75, 102, 0, 0, 1, 3, 64, 0, 0, 12, 0, 0, 75, 102)
	b[1], b[2], b[3] = byte(len(b)>>16), byte(len(b)>>8), byte(len(b))
	return b
}

func scapHandshake(t *testing.T, n *Node, co *codec.Codec, encrypted, passive bool, remoteCaps LocalCapabilities, dual bool) (*Peer, transport.Conn, error) {
	t.Helper()
	a, b := networkAdapters(t, encrypted)
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		if passive {
			m := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 111, EndToEnd: 222}, Body: &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "scap-peer", AuthApplicationID: remoteCaps.AuthApplicationID, AcctApplicationID: remoteCaps.AcctApplicationID, VendorSpecificApplicationID: remoteCaps.VendorSpecificApplicationID}}
			raw, e := co.Append(nil, m)
			if e == nil {
				if dual {
					raw = scapDualAdvertisement(raw)
				}
				e = b.WriteFrame(ctx, raw).Err
			}
			if e == nil {
				var reply pdu.Message
				reply, e = scapReadHandshake(ctx, t, n, b, co)
				if e == nil {
					if ca, ok := reply.Body.(*base.CEA); !ok || (ca.ResultCode != 2001 && ca.ResultCode != 5010) {
						e = ErrProtocol
					}
				}
			}
			ch <- e
			return
		}
		req, e := scapReadHandshake(ctx, t, n, b, co)
		if e == nil {
			// The request must explicitly advertise exactly the selected strict capability.
			ce, ok := req.Body.(*base.CER)
			if !ok || !hasSCAP1Accounting(ce.VendorSpecificApplicationID) {
				ch <- ErrProtocol
				return
			}
			var reply pdu.Message
			reply, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "scap-peer", AuthApplicationID: remoteCaps.AuthApplicationID, AcctApplicationID: remoteCaps.AcctApplicationID, VendorSpecificApplicationID: remoteCaps.VendorSpecificApplicationID}, draOrigin)
			if e == nil {
				var raw []byte
				raw, e = co.Append(nil, reply)
				if e == nil {
					if dual {
						raw = scapDualAdvertisement(raw)
					}
					e = b.WriteFrame(ctx, raw).Err
				}
			}
		}
		ch <- e
	}()
	var p *Peer
	var err error
	if passive {
		p, err = n.Accept(ctx, "dra", a)
	} else {
		p, err = n.Connect(ctx, "dra", a)
	}
	if e := <-ch; e != nil && !(dual && passive && err != nil) {
		t.Fatal("test peer handshake", e)
	}
	return p, b, err
}

func TestSCAP1RuntimeCapabilities(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cases := []struct {
			name     string
			caps     LocalCapabilities
			dual, ok bool
		}{
			{"accounting", LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{scapCapability()}}, false, true},
			{"missing", LocalCapabilities{AuthApplicationID: []uint32{appID}}, false, false},
			{"auth only", LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 193, AuthApplicationID: avp.Some(uint32(19302))}}}, false, false},
			{"unscoped acct", LocalCapabilities{AcctApplicationID: []uint32{19302}}, false, false},
			{"wrong vendor", LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 10415, AcctApplicationID: avp.Some(uint32(19302))}}}, false, false},
			{"old dual", LocalCapabilities{VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{scapCapability()}}, true, false},
		}
		for _, tc := range cases {
			for _, passive := range []bool{false, true} {
				name := tc.name + "/connect"
				if passive {
					name = tc.name + "/accept"
				}
				t.Run(name, func(t *testing.T) {
					n, co := newInboundNode(t, scapRuntimeConfig(t))
					p, remote, e := scapHandshake(t, n, co, encrypted, passive, tc.caps, tc.dual)
					if !tc.ok {
						if e == nil || n.peers["dra"].Info().Ready {
							t.Fatal("missing strict capability accepted", e)
						}
						return
					}
					if e != nil || !p.Info().Ready {
						t.Fatal(e)
					}
					port := portTest(t, n, 1)
					id, e := p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
					if e != nil {
						t.Fatal(e)
					}
					r := routeRead(t, co, remote)
					ans, e := co.CompleteAnswer(r, scapAnswer(2001), serverOrigin)
					if e != nil {
						t.Fatal(e)
					}
					if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
						t.Fatal(e)
					}
					if got := routeCompletion(t, n, port, id); got.Outcome != AnswerReceived {
						t.Fatal(got)
					}
				})
			}
		}
	})
}

func TestSCAP1RuntimeMixedPeerWithoutCapability(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := scapRuntimeConfig(t)
		ordinary := inboundConfig(t)
		cfg.Capabilities.AuthApplicationID = []uint32{appID}
		cfg.Peers[0].Commands = append(cfg.Peers[0].Commands, ordinary.Peers[0].Commands[0])
		n, co := newInboundNode(t, cfg)
		p, remote, e := scapHandshake(t, n, co, encrypted, false, LocalCapabilities{AuthApplicationID: []uint32{appID}}, false)
		if e != nil || !p.Info().Ready {
			t.Fatal(e)
		}
		port := portTest(t, n, 1)
		before := n.Resources()
		if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrNotReady) {
			t.Fatal("numeric/peer readiness bypass", e)
		}
		if after := n.Resources(); after.Outstanding != before.Outstanding || after.OutboundBytes != before.OutboundBytes {
			t.Fatal(after)
		}
		r, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 700, EndToEnd: 701})
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, r); e != nil {
			t.Fatal(e)
		}
		a := routeRead(t, co, remote)
		ea, ok := a.Body.(*pdu.ErrorAnswer)
		if !ok || ea.ResultCode != 3007 {
			t.Fatal("missing accounting capability", a)
		}
		if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		// An ordinary configured application remains usable on the same connection.
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r = routeRead(t, co, remote)
		a, e = answerFor(co, r)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); got.Outcome != AnswerReceived {
			t.Fatal(got)
		}
	})
}

func TestSCAP1RuntimeRouteBudgetAndShutdown(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, scapRuntimeConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		route, e := n.NewRoute(RouteConfig{Command: scapBody(t).CommandKey(), Peers: []PeerID{"dra"}, ExchangePolicy: SingleAttempt, MaxAttempts: 1})
		if e != nil {
			t.Fatal(e)
		}
		id, e := route.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrBackpressure) {
			t.Fatal("port budget bypass", e)
		}
		done := make(chan error, 1)
		go func() { done <- p.Shutdown(ctxTest(t)) }()
		waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
		if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrClosed) {
			t.Fatal("draining admission", e)
		}
		a, e := co.CompleteAnswer(r, scapAnswer(2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, a); e != nil {
			t.Fatal(e)
		}
		dpr := routeRead(t, co, remote)
		if dpr.Header.ApplicationID != 0 || dpr.Header.CommandCode != 282 {
			t.Fatal("ACR hijacked close control", dpr)
		}
		dpa, e := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, dpa); e != nil {
			t.Fatal(e)
		}
		if e = <-done; e != nil {
			t.Fatal(e)
		}
		got := routeCompletion(t, n, port, id)
		if got.Outcome != AnswerReceived || len(got.Attempts) != 1 {
			t.Fatal(got)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}

func TestSCAP1RuntimeUnknownAppZeroAndInboundBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := scapRuntimeConfig(t)
		cfg.Peers[0].MaxInbound = 1
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, encrypted)
		makeReq := func(h uint32) pdu.Message {
			r, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: h, EndToEnd: h + 100})
			if e != nil {
				t.Fatal(e)
			}
			return r
		}
		r := makeReq(10)
		if e := writeWire(ctxTest(t), remote, co, r); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		if n.Resources().InboundOutstanding != 1 || n.Resources().Pending != 0 {
			t.Fatal("inbound did not use application budget", n.Resources())
		}
		if e = writeWire(ctxTest(t), remote, co, makeReq(11)); e != nil {
			t.Fatal(e)
		}
		a := routeRead(t, co, remote)
		if ea, ok := a.Body.(*pdu.ErrorAnswer); !ok || ea.ResultCode != 3004 {
			t.Fatal("inbound budget bypass", a)
		}
		// Unregistered App0 command is still rejected, even while SCAP1 is enabled.
		raw, e := co.Append(nil, makeReq(12))
		if e != nil {
			t.Fatal(e)
		}
		raw[5], raw[6], raw[7] = 0, 3, 231
		sendRejectRaw(t, remote, raw)
		// Generic E for an unknown command cannot be decoded through a registered
		// command operation; inspect only the bounded header and Result-Code.
		prefix, e := remote.ReadPrefix(ctxTest(t), time.Second)
		if e != nil {
			t.Fatal(e)
		}
		h, size, e := wire.Header(prefix[:], co.Limits())
		if e != nil {
			t.Fatal(e)
		}
		buf := make([]byte, size)
		if e = remote.ReadFrameInto(ctxTest(t), buf); e != nil {
			t.Fatal(e)
		}
		if h.ApplicationID != 0 || h.CommandCode != 999 || h.Flags&pdu.ErrorFlag == 0 {
			t.Fatal(h)
		}
		found := false
		for off := 20; off < len(buf); {
			l := int(buf[off+5])<<16 | int(buf[off+6])<<8 | int(buf[off+7])
			if binary.BigEndian.Uint32(buf[off:]) == 268 {
				found = binary.BigEndian.Uint32(buf[off+8:]) == 3001
			}
			off += (l + 3) &^ 3
		}
		if !found {
			t.Fatal("unknown App0 accepted")
		}
		if got, e := in.Respond(ctxTest(t), scapAnswer(2001)); e != nil || !got.Committed {
			t.Fatal(got, e)
		}
		routeRead(t, co, remote)
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = p.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("overload produced second business delivery", e)
		}
	})
}

func TestSCAP1RuntimeDirectionPolicies(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, inOnly := range []bool{false, true} {
			name := "outbound-only"
			if inOnly {
				name = "inbound-only"
			}
			t.Run(name, func(t *testing.T) {
				cfg := scapRuntimeConfig(t)
				if inOnly {
					cfg.Peers[0].Commands = nil
				} else {
					cfg.Peers[0].InboundCommands = nil
					cfg.Peers[0].MaxInbound = 0
					cfg.Peers[0].MaxInboundBytes = 0
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				r, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 900, EndToEnd: 901})
				if e != nil {
					t.Fatal(e)
				}
				if e = writeWire(ctxTest(t), remote, co, r); e != nil {
					t.Fatal(e)
				}
				if !inOnly {
					a := routeRead(t, co, remote)
					if ea, ok := a.Body.(*pdu.ErrorAnswer); !ok || ea.ResultCode != 3001 {
						t.Fatal("capable but no inbound policy", a)
					}
				} else {
					in, e := p.Receive(ctxTest(t))
					if e != nil {
						t.Fatal(e)
					}
					if got, e := in.Respond(ctxTest(t), scapAnswer(2001)); e != nil || !got.Committed {
						t.Fatal(got, e)
					}
					routeRead(t, co, remote)
					port := portTest(t, n, 1)
					if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrUnsupported) {
						t.Fatal("inbound grants outbound", e)
					}
				}
			})
		}
	})
}

func TestSCAP1RuntimeReconnectRechecksCapability(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := scapRuntimeConfig(t)
		ordinary := inboundConfig(t)
		cfg.Capabilities.AuthApplicationID = []uint32{appID}
		cfg.Peers[0].Commands = append(cfg.Peers[0].Commands, ordinary.Peers[0].Commands[0])
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		id, e := p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		routeRead(t, co, remote)
		oldConnection := p.Info().ConnectionID
		remote.Close()
		got := routeCompletion(t, n, port, id)
		if got.Err == nil || !got.Delivery.MayHaveReachedPeer {
			t.Fatal(got)
		}
		waitFor(t, func() bool { return p.Info().Candidates == 0 && p.Info().Protocol == ProtocolClosed })
		p, second, e := scapHandshake(t, n, co, encrypted, false, LocalCapabilities{AuthApplicationID: []uint32{appID}}, false)
		if e != nil {
			t.Fatal(e)
		}
		finishReopen(t, n, p, co, second)
		if p.Info().ConnectionID == oldConnection {
			t.Fatal("generation reused")
		}
		if _, e = p.Submit(ctxTest(t), scapBody(t), RequestOptions{}, port); !errors.Is(e, ErrNotReady) {
			t.Fatal("old accounting capability leaked across reconnect", e)
		}
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, e = second.ReadPrefix(quiet, time.Second)
		cancel()
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("reconnect replayed business request", e)
		}
		if n.Resources().Outstanding != 0 || n.Resources().Pending != 0 {
			t.Fatal(n.Resources())
		}
	})
}
