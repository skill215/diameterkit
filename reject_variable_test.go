// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Selected variable-field 5005 diagnostics over TCP/TLS and emergency slots.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestVariableOfflineCommands(t *testing.T) {
	cfg := manualConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	host := []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}
	for _, body := range []pdu.Body{
		&base.CER{OriginHost: "peer.example", OriginRealm: "example", HostIPAddress: host, VendorID: 0, ProductName: ""},
		&base.CEA{ResultCode: 2001, OriginHost: "peer.example", OriginRealm: "example", HostIPAddress: host, VendorID: 0, ProductName: ""},
		&sgd.TFR{SessionID: "peer.example;variable", AuthSessionState: 1, OriginHost: "peer.example", OriginRealm: "example", DestinationHost: "hss.example", DestinationRealm: "example", UserName: "001010123456789", SCAddress: []byte{0x21, 0xf3}, SMRPUI: []byte{}},
	} {
		t.Run(fmt.Sprintf("%T", body), func(t *testing.T) {
			k := body.CommandKey()
			flags := byte(0)
			if _, ok := body.(*base.CER); ok {
				flags = pdu.RequestFlag
			}
			if _, ok := body.(*sgd.TFR); ok {
				flags = pdu.RequestFlag | pdu.ProxiableFlag
			}
			req := pdu.Message{Header: pdu.Header{Flags: flags, CommandCode: k.Code, ApplicationID: k.ApplicationID}, Body: body}
			raw := reqBytes(t, co, req)
			if _, e := co.Decode(raw); e != nil {
				t.Fatal("present empty", e)
			}
			key := avp.Key{Code: 269}
			f := byte(64)
			if _, ok := body.(*sgd.TFR); ok {
				key = avp.Key{VendorID: 10415, Code: 3301}
				f = 192
			}
			path := lengthPath(key)
			raw = missingRaw(t, co, req, path, false)
			_, e := co.Decode(raw)
			q, ok := missingavp.Quote(e, cfg.Limits, 32)
			if !ok {
				t.Fatal(e)
			}
			assertMissingChain(t, q, path, 0, f)
			// TFR SC-Address has no opt-in even though OFR's same AVP does.
			if _, ok := body.(*sgd.TFR); ok {
				raw = missingRaw(t, co, req, lengthPath(avp.Key{VendorID: 10415, Code: 3300}), false)
				_, e = co.Decode(raw)
				if !errors.Is(e, validation.ErrMissing) {
					t.Fatal(e)
				}
				if _, ok := missingavp.Quote(e, cfg.Limits, 1024); ok {
					t.Fatal("inferred field by key")
				}
				if _, e = n.rejectFrame(raw, 1024, rejectDecode, e); !errors.Is(e, ErrUnsupported) {
					t.Fatal("unsupported omission mapped", e)
				}
			}
		})
	}
}

func TestRejectVariableFields(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, tc := range []struct {
			name  string
			app   int
			path  []missingSelector
			width int
			flags byte
		}{
			{"service-context", 0, lengthPath(avp.Key{Code: 461}), 1, 64},
			{"second-subscription", 0, []missingSelector{{avp.Key{Code: 443}, 1}, {avp.Key{Code: 444}, 0}}, 0, 64},
			{"s6c-sc", 1, lengthPath(avp.Key{VendorID: 10415, Code: 3300}), 1, 192},
			{"destination-realm", 1, lengthPath(avp.Key{Code: 283}), 1, 64},
			{"sgd-ui", 3, lengthPath(avp.Key{VendorID: 10415, Code: 3301}), 0, 192},
			{"sgd-sc", 3, lengthPath(avp.Key{VendorID: 10415, Code: 3300}), 1, 192},
		} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/emergency-%t", tc.name, emergency), func(t *testing.T) {
					cfg := manualConfig(t)
					cfg.Resources.RejectSlots = 2
					cfg.Resources.EmergencyRequestMax = 4096
					cfg.Resources.EmergencyReplyMax = 4096
					if emergency {
						cfg.Resources.InboundClasses[0].Slots = 1
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					makeRequest := func(id uint32) pdu.Message {
						return mixedDuplicateRequest(t, n, co, tc.app, id, id+10000, gmscOrigin)
					}
					var held *Inbound
					if emergency {
						combinedSend(t, co, remote, makeRequest(10500))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := makeRequest(10501)
					path := tc.path
					raw := missingRaw(t, co, req, path, false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					q, ok := missingavp.Quote(cause, cfg.Limits, 1024)
					assertMissingChain(t, q, path, tc.width, tc.flags)
					want := append([]byte(nil), q...)
					if !ok {
						t.Fatalf("minimum %x want %x (%v)", q, want, cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, len(want)-1); ok {
						t.Fatal("quote budget")
					}
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					if tc.name == "destination-realm" {
						if _, ok := missingDiagnostic(cause); ok {
							t.Fatal("explicit proof fell back")
						}
					}
					empty, _ := invalidRaw(t, co, req, path, nil, avp.Key{})
					_, emptyErr := co.Decode(empty)
					if errors.Is(emptyErr, validation.ErrMissing) || (emptyErr == nil) != (tc.width == 0) {
						t.Fatal("present-empty semantics", emptyErr)
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5005)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
						t.Fatal("diagnostic/envelope")
					}
					if len(path) == 1 && tc.width == 1 {
						h := 8
						if path[0].key.VendorID != 0 {
							h = 12
						}
						if wire.Uint24(q[5:8]) != uint32(h+1) || !bytes.Equal(q[len(q)-3:], []byte{0, 0, 0}) {
							t.Fatal("leaf length/padding")
						}
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
					follow := makeRequest(10502)
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
