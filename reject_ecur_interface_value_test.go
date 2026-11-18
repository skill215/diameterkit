// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: INITIAL/TERMINATION interface value wire and network acceptance.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func ecurInterfacePath(origin bool) []missingSelector {
	group := uint32(2002)
	if origin {
		group = 2009
	}
	return lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: group}, avp.Key{VendorID: 10415, Code: 2006})
}
func ecurInterfaceQuote(t *testing.T, origin bool, value int32) []byte {
	group := uint32(2002)
	if origin {
		group = 2009
	}
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(value))
	b := invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, data[:])
	for _, code := range []uint32{group, 2000, 873} {
		b = invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, b)
	}
	return b
}
func ecurAddAddress(t *testing.T, raw []byte) []byte {
	t.Helper()
	// Rebuild the known SMS fixture using the independent test encoder.
	var walk func([]byte, []missingSelector) []byte
	walk = func(src []byte, path []missingSelector) []byte {
		d := wire.NewDecoder(src, wire.DefaultLimits())
		var out []byte
		for {
			v, ok, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !ok {
				break
			}
			data := v.Data
			if v.Key == path[0].key {
				if len(path) == 1 {
					data = append(append([]byte{}, data...), invalidPart(t, avp.Key{VendorID: 10415, Code: 2008}, 192, nil)...)
				} else {
					data = walk(data, path[1:])
				}
			}
			out = append(out, invalidPart(t, v.Key, v.Flags, data)...)
		}
		return out
	}
	b := walk(raw[20:], lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}))
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func TestECURInterfaceValueWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term"} {
		for _, origin := range []bool{false, true} {
			for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
				t.Run(fmt.Sprintf("%s/%t/%d", kind, origin, value), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 12010, 22010)
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], uint32(value))
					raw, _ := invalidRaw(t, co, req, ecurInterfacePath(origin), data[:], avp.Key{})
					_, cause := co.Decode(raw)
					expected := int32(2)
					if origin {
						expected = 3
					}
					q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 52)
					if ok != (value != expected) || ok && (code != 5004 || !bytes.Equal(q, ecurInterfaceQuote(t, origin, value))) {
						t.Fatal(cause, code, q)
					}
					if value == expected && cause != nil {
						t.Fatal(cause)
					}
				})
			}
		}
	}
}
func TestECURInterfaceValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term"} {
		for _, origin := range []bool{false, true} {
			for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "action", "address", "both", "selected-length", "counterpart-length", "counterpart-duplicate", "counterpart-missing", "counterpart-trailing", "later-unit"} {
				t.Run(fmt.Sprintf("%s/%t/%s", kind, origin, fault), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 12011, 22011)
					raw, _ := invalidRaw(t, co, req, ecurInterfacePath(origin), []byte{0, 0, 0, 4}, avp.Key{})
					expected := uint32(0)
					want := ecurInterfaceQuote(t, origin, 4)
					edit := func(path []missingSelector, data []byte) {
						b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
						raw = append(raw[:20:20], b...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					switch fault {
					case "plain":
						expected = 5004
					case "sequence":
						edit(lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2})
					case "subscriptions":
						raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
					case "sms":
						raw = smsOmit(t, raw, 2018)
					case "mms":
						raw = mmsGroupOmit(t, raw, mmsGroupPath())
					case "timestamp":
						edit(recipientPath(1202), []byte{0, 0, 0, 0})
					case "action":
						raw = append(raw, invalidPart(t, avp.Key{Code: 436}, 64, []byte{0, 0, 0, 0})...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					case "address":
						raw = ecurAddAddress(t, raw)
					case "both":
						edit(ecurInterfacePath(!origin), []byte{0, 0, 0, 4})
					case "selected-length":
						edit(ecurInterfacePath(origin), []byte{1})
						expected = 5014
					case "counterpart-length":
						edit(ecurInterfacePath(!origin), []byte{1})
						expected = 5014
					case "counterpart-duplicate", "counterpart-trailing":
						group := uint32(2009)
						value := byte(3)
						if origin {
							group = 2002
							value = 2
						}
						leaf := invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, []byte{0, 0, 0, value})
						if fault == "counterpart-duplicate" {
							leaf = append(leaf, leaf...)
						} else {
							leaf = append(leaf, 1)
						}
						edit(smsPath(group), leaf)
					case "counterpart-missing":
						raw = mmsGroupOmit(t, raw, ecurInterfacePath(!origin))
					case "later-unit":
						code := uint32(437)
						if kind == "term" {
							code = 446
						}
						raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: code}))
						expected = 5004
					}
					_, cause := co.Decode(raw)
					if cause == nil {
						t.Fatal("invalid accepted")
					}
					q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					if ok != (expected != 0) || code != expected || code == 5004 && !bytes.Equal(q, want) {
						t.Fatal(cause, code, q)
					}
				})
			}
		}
	}
}
func TestRejectECURInterfaceValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"initial", "term"} {
			for _, fault := range []string{"default-origin", "default-dest", "both", "address", "action"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default-origin" || fault == "default-dest" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := combinedDuplicateRequest(t, n, co, kind, 12012, 22012)
					origin := fault != "default-dest"
					raw, _ := invalidRaw(t, co, req, ecurInterfacePath(origin), []byte{0, 0, 0, 4}, avp.Key{})
					switch fault {
					case "both":
						b, _ := invalidEdit(t, raw[20:], ecurInterfacePath(false), []byte{0, 0, 0, 4}, avp.Key{})
						raw = append(raw[:20:20], b...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					case "address":
						raw = ecurAddAddress(t, raw)
					case "action":
						raw = append(raw, invalidPart(t, avp.Key{Code: 436}, 64, []byte{0, 0, 0, 0})...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					sendRejectRaw(t, remote, raw)
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unreviewed/default reply %x", b)
					}
				})
			}
		}
	})
}

func TestRejectECURInterfaceValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"initial", "term"} {
			for _, origin := range []bool{false, true} {
				for _, value := range []int32{-1, 0, 4} {
					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%t/%d/emergency-%t", kind, origin, value, emergency), func(t *testing.T) {
							cfg := combinedConfig(t)
							cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
							cfg.Resources.RejectSlots = 2
							cfg.Resources.EmergencyRequestMax = 4096
							cfg.Resources.EmergencyReplyMax = 4096
							if emergency {
								cfg.Resources.InboundClasses[0].Slots = 1
							}
							n, co := newInboundNode(t, cfg)
							p, remote := networkConnectNode(t, n, co, encrypted)
							var held *Inbound
							if emergency {
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12000, 22000))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12001, 22001)
							var data [4]byte
							binary.BigEndian.PutUint32(data[:], uint32(value))
							raw, _ := invalidRaw(t, co, req, ecurInterfacePath(origin), data[:], avp.Key{})
							want := ecurInterfaceQuote(t, origin, value)
							_, cause := co.Decode(raw)
							if !errors.Is(cause, validation.ErrField) || !errors.Is(cause, validation.ErrUnsupported) {
								t.Fatal(cause)
							}

							q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, cfg.Limits, len(want))
							if !ok || code != 5004 || !bytes.Equal(q, want) {
								t.Fatal(q, cause)
							}
							if _, _, ok := excessavp.QuoteInvalid(raw[20:], cause, cfg.Limits, len(want)-1); ok {
								t.Fatal("quote budget")
							}
							if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
								t.Fatal("answer budget")
							}
							var field *validation.FieldError
							if !errors.As(cause, &field) {
								t.Fatal(cause)
							}
							field.Path[0] = avp.Key{Code: 263}
							if _, ok := missingDiagnostic(cause); ok {
								t.Fatal("downgrade")
							}
							if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
								t.Fatal("altered path", e)
							}
							field.Path[0] = avp.Key{VendorID: 10415, Code: 873}
							reply, e := n.rejectFrame(raw, 4096, rejectDecode, cause)
							if e != nil {
								t.Fatal(e)
							}
							if _, e = n.rejectFrame(raw, len(reply), rejectDecode, cause); e != nil {
								t.Fatal("exact reply bound", e)
							}
							if _, e = n.rejectFrame(raw, len(reply)-1, rejectDecode, cause); e == nil {
								t.Fatal("short reply bound")
							}
							sendRejectRaw(t, remote, raw)
							ans := readReject(t, remote, req.Header, 5004)
							if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
								t.Fatal("diagnostic/envelope")
							}
							body := req.Body.(*combined.CCR)
							if ans.SessionID.Value != body.SessionID || ans.ProxyInfo[0].ProxyHost != body.ProxyInfo[0].ProxyHost || !bytes.Equal(ans.ProxyInfo[0].ProxyState, body.ProxyInfo[0].ProxyState) {
								t.Fatal("envelope contents")
							}
							clear(raw)
							if !bytes.Equal(q, want) {
								t.Fatal("borrowed output")
							}
							waitFor(t, func() bool {
								return n.Resources().RejectOutstanding == 0 && n.Resources().InboundOutstanding == before.InboundOutstanding
							})
							if n.Resources().InboundBytes != before.InboundBytes {
								t.Fatal("held charge")
							}
							noDuplicateDelivery(t, p)
							if held != nil {
								if _, e := held.Reject(ctxTest(t), FailureContextUnavailable); e != nil {
									t.Fatal(e)
								}
								readReject(t, remote, held.Message().Header, 5012)
							}
							follow := combinedDuplicateRequest(t, n, co, kind, 12002, 22002)
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
			}
		}
	})
}
