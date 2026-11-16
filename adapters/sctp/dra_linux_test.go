//go:build linux && !386

package sctp

import (
	"fmt"
	"math"
	"testing"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
)

func TestSCTPRelayAndRequestTargets(t *testing.T) {
	for _, passive := range []bool{false, true} {
		t.Run(fmt.Sprintf("passive=%v", passive), func(t *testing.T) {
			cfg := nodeConfig(t)
			cfg.Peers[0].RequestTargets = []diameter.RequestTarget{
				{ID: "one", Command: requestBody().CommandKey(), Destination: pdu.Origin{Host: "hss1.example", Realm: "example"}},
				{ID: "two", Command: requestBody().CommandKey(), Destination: pdu.Origin{Host: "hss2.other", Realm: "other"}},
			}
			n := ownNode(t, cfg)
			co, e := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
			if e != nil {
				t.Fatal(e)
			}
			a, b := adapters(t)
			send := func(m pdu.Message) error {
				frame, e := co.Append(nil, m)
				if e != nil {
					return e
				}
				return b.WriteFrame(testContext(t), frame).Err
			}
			recv := func() (pdu.Message, error) {
				frame, e := read(testContext(t), b)
				if e != nil {
					return pdu.Message{}, e
				}
				return co.Decode(frame)
			}
			var p *diameter.Peer
			if passive {
				cer := &base.CER{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm, HostIPAddress: cfg.Capabilities.HostIPAddress, ProductName: "relay", AuthApplicationID: []uint32{math.MaxUint32}}
				if e = send(pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257, HopByHop: 11, EndToEnd: 12}, Body: cer}); e != nil {
					t.Fatal(e)
				}
				p, e = n.Accept(testContext(t), "dra", a)
				if e != nil {
					t.Fatal(e)
				}
				cea, e := recv()
				if e != nil || cea.Body.(*base.CEA).ResultCode != 2001 {
					t.Fatal(cea, e)
				}
				dwr, e := recv()
				if e != nil || dwr.Header.CommandCode != 280 {
					t.Fatal(dwr, e)
				}
				if !p.Info().SendBarrier {
					t.Fatal("relay skipped handshake barrier")
				}
				dwa, e := co.CompleteAnswer(dwr, &base.DWA{ResultCode: 2001}, draOrigin)
				if e != nil {
					t.Fatal(e)
				}
				if e = send(dwa); e != nil {
					t.Fatal(e)
				}
				waitFor(t, func() bool { return !p.Info().SendBarrier })
			} else {
				ch := make(chan error, 1)
				go func() {
					req, e := recv()
					if e == nil {
						var ans pdu.Message
						ans, e = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: cfg.Capabilities.HostIPAddress, ProductName: "relay", AuthApplicationID: []uint32{math.MaxUint32}}, draOrigin)
						if e == nil {
							e = send(ans)
						}
					}
					ch <- e
				}()
				p, e = n.Connect(testContext(t), "dra", a)
				if e != nil {
					t.Fatal(e)
				}
				if e = <-ch; e != nil {
					t.Fatal(e)
				}
			}
			port, e := n.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 2, MaxOutstandingBytes: 64 << 20})
			if e != nil {
				t.Fatal(e)
			}
			defer port.Close(testContext(t))
			for _, target := range cfg.Peers[0].RequestTargets {
				if _, e = p.Submit(testContext(t), requestBody(), diameter.RequestOptions{TargetID: target.ID}, port); e != nil {
					t.Fatal(e)
				}
			}
			seen := map[string]bool{}
			for range 2 {
				req, e := recv()
				if e != nil {
					t.Fatal(e)
				}
				r := req.Body.(*s6c.SRR)
				origin := pdu.Origin{Host: r.DestinationHost.Value, Realm: r.DestinationRealm}
				seen[origin.Host] = true
				ans, e := co.CompleteAnswer(req, success(), origin)
				if e != nil {
					t.Fatal(e)
				}
				if e = send(ans); e != nil {
					t.Fatal(e)
				}
			}
			if !seen["hss1.example"] || !seen["hss2.other"] {
				t.Fatal(seen)
			}
			for range 2 {
				r, e := port.Receive(testContext(t))
				if e != nil || r.Result.Err != nil {
					t.Fatal(r, e)
				}
			}
			waitFor(t, func() bool {
				return n.Resources().Outstanding == 0 && n.Resources().OutboundBytes == 0 && n.Resources().Pending == 0
			})
		})
	}
}
