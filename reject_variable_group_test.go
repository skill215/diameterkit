// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Variable-child whole-group 5005 diagnostics over TCP/TLS and emergency slots.
// SPDX-AI-Date: 2026-11-17

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/app/ro/dcca"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func minimumVariableExample(t *testing.T, kind string) ([]missingSelector, []byte) {
	part := func(code, vendor uint32, data []byte) []byte {
		flags := byte(64)
		head := 8
		if vendor != 0 {
			flags = 192
			head = 12
		}
		b := invalidPart(t, avp.Key{Code: code, VendorID: vendor}, flags, data)
		clear(b[head+len(data):])
		return b
	}
	switch kind {
	case "dcca":
		return lengthPath(avp.Key{Code: 443}), part(443, 0, append(part(450, 0, make([]byte, 4)), part(444, 0, make([]byte, 1))...))
	case "scap":
		return lengthPath(avp.Key{VendorID: 193, Code: 553}), part(553, 193, append(part(555, 193, make([]byte, 4)), part(554, 193, nil)...))
	case "address":
		return lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 877}, avp.Key{VendorID: 10415, Code: 886}), part(873, 10415, part(877, 10415, part(886, 10415, append(part(899, 10415, make([]byte, 4)), part(897, 10415, nil)...))))
	}
	panic(kind)
}

type minimumVariableFallbackProof struct{}

func (*minimumVariableFallbackProof) Error() string { return "invalid explicit minimum" }
func (*minimumVariableFallbackProof) Unwrap() error { return validation.ErrMissing }
func (*minimumVariableFallbackProof) DiameterMissingGroupV2() (avp.Key, byte, string, error) {
	return avp.Key{Code: 283}, 64, "invalid", nil
}
func TestMinimumVariableLegacyFallback(t *testing.T) {
	cause := validation.At(avp.Key{Code: 283}, &minimumVariableFallbackProof{})
	if _, ok := missingDiagnostic(cause); ok {
		t.Fatal("invalid explicit tree became legacy identity placeholder")
	}
	if _, ok := missingDiagnostic(validation.At(avp.Key{Code: 283}, validation.ErrMissing)); !ok {
		t.Fatal("legacy bare behavior changed")
	}
}
func TestRejectMinimumVariableGroups(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"dcca", "scap", "address"} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/emergency-%t", kind, emergency), func(t *testing.T) {
					cfg := scapRuntimeConfig(t)
					if kind == "dcca" {
						cfg = dccaConfig(t)
					}
					if kind == "address" {
						cfg = combinedConfig(t)
					}
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					cfg.Resources.RejectSlots = 2
					cfg.Resources.EmergencyRequestMax = 4096
					cfg.Resources.EmergencyReplyMax = 4096
					if emergency {
						cfg.Resources.InboundClasses[0].Slots = 1
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					makeRequest := func(id uint32) pdu.Message {
						if kind == "address" {
							return combinedDuplicateRequest(t, n, co, "general", id, id+10000)
						}
						if kind == "dcca" {
							b := dccaBody(dcca.CheckBalance)
							b.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{0, 255}}}
							req, e := co.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: serverOrigin.Host, OriginRealm: serverOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: id, EndToEnd: id + 10000})
							if e != nil {
								t.Fatal(e)
							}
							return req
						}
						req, e := co.CompleteRequest(scapBody(t), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: id, EndToEnd: id + 10000})
						if e != nil {
							t.Fatal(e)
						}
						req.Body.(*scap1.ACR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{0, 255}}}
						return req
					}
					var held *Inbound
					if emergency {
						combinedSend(t, co, remote, makeRequest(10600))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := makeRequest(10601)
					path, want := minimumVariableExample(t, kind)
					raw := missingRaw(t, co, req, path, false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					q, ok := missingavp.Quote(cause, cfg.Limits, len(want))
					if !ok || !bytes.Equal(q, want) {
						t.Fatalf("minimum %x want %x (%v)", q, want, cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, len(want)-1); ok {
						t.Fatal("quote budget")
					}
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					empty, _ := invalidRaw(t, co, req, path, nil, avp.Key{})
					_, emptyErr := co.Decode(empty)
					var field *validation.FieldError
					if !errors.Is(emptyErr, validation.ErrMissing) || !errors.As(emptyErr, &field) || len(field.Path) != len(path)+1 {
						t.Fatal("present group confused with whole absence", emptyErr)
					}
					eq, ok := missingavp.Quote(emptyErr, cfg.Limits, 1024)
					if !ok || bytes.Equal(eq, want) {
						t.Fatal("invented variable child in present group")
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5005)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
						t.Fatal("diagnostic/envelope")
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
					follow := makeRequest(10602)
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

func TestRejectMinimumVariableBoundaries(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"unreviewed-sms", "default-policy"} {
			t.Run(kind, func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				if kind == "default-policy" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				template := "general"
				req := combinedDuplicateRequest(t, n, co, template, 10610, 20610)
				var path []missingSelector
				switch kind {
				case "unreviewed-sms":
					path = lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000})
				default:
					path, _ = minimumVariableExample(t, "address")
				}
				raw := missingRaw(t, co, req, path, false)
				_, cause := co.Decode(raw)
				if !errors.Is(cause, validation.ErrMissing) {
					t.Fatal(kind, cause)
				}
				q, ok := missingavp.Quote(cause, cfg.Limits, 1024)
				if ok != (kind == "default-policy") {
					t.Fatal("boundary", kind, q, cause)
				}
				if kind != "default-policy" {
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if b, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("unsafe reply %x", b)
				}
				waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
				closeDuplicateNode(t, n)
			})
		}
	})
}
