// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS template value acceptance.
// SPDX-AI-Date: 2026-11-19

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

func smsValuePath(node bool) []missingSelector {
	code := uint32(2007)
	if node {
		code = 2016
	}
	return smsPath(code)
}
func smsValueQuote(t *testing.T, node bool, value int32) []byte {
	code := uint32(2007)
	if node {
		code = 2016
	}
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(value))
	b := invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, data[:])
	return invalidPart(t, avp.Key{VendorID: 10415, Code: 873}, 192, invalidPart(t, avp.Key{VendorID: 10415, Code: 2000}, 192, b))
}
func TestSMSValueWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, origin := range []bool{false, true} {
			for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
				t.Run(fmt.Sprintf("%s/%t/%d", kind, origin, value), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 12010, 22010)
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], uint32(value))
					raw, _ := invalidRaw(t, co, req, smsValuePath(origin), data[:], avp.Key{})
					_, cause := co.Decode(raw)
					expected := int32(0)
					q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 40)
					if ok != (value != expected) || ok && (code != 5004 || !bytes.Equal(q, smsValueQuote(t, origin, value))) {
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
func TestRejectSMSValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, fault := range []string{"default-node", "default-message", "both"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					if fault == "both" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12212, 22212), smsValuePath(fault != "default-message"), []byte{255, 255, 255, 255}, avp.Key{})
					if fault == "both" {
						b, _ := invalidEdit(t, raw[20:], smsValuePath(false), []byte{0, 0, 0, 1}, avp.Key{})
						raw = append(raw[:20:20], b...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					sendRejectRaw(t, remote, raw)
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected reply %x", b)
					}
				})
			}
		}
	})
}
func TestRejectSMSValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, origin := range []bool{false, true} {
				for _, value := range []int32{-1, 4} {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12220, 22200))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12201, 22201)
							var data [4]byte
							binary.BigEndian.PutUint32(data[:], uint32(value))
							raw, _ := invalidRaw(t, co, req, smsValuePath(origin), data[:], avp.Key{})
							want := smsValueQuote(t, origin, value)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12202, 22202)
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

func TestSMSValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, node := range []bool{false, true} {
			for _, fault := range []string{"plain", "sequence", "subscription", "marker", "identity", "both", "selected-width", "counterpart-width", "later-sms", "later-mms", "later-time", "later-unit"} {
				t.Run(fmt.Sprintf("%s/%t/%s", kind, node, fault), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 12211, 22211)
					raw, _ := invalidRaw(t, co, req, smsValuePath(node), []byte{255, 255, 255, 255}, avp.Key{})
					expected := uint32(0)
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
					case "subscription":
						raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
					case "marker":
						edit(lengthPath(avp.Key{Code: 455}), []byte{0, 0, 0, 0})
					case "identity":
						edit(lengthPath(avp.Key{Code: 458}, avp.Key{Code: 460}), []byte("other"))
					case "both":
						edit(smsValuePath(!node), []byte{0, 0, 0, 1})
					case "selected-width":
						edit(smsValuePath(node), []byte{1})
						expected = 5014
					case "counterpart-width":
						edit(smsValuePath(!node), []byte{0})
						expected = 5014
					case "later-sms":
						raw = smsOmit(t, raw, 2018)
						expected = 5004
					case "later-mms":
						raw = mmsGroupOmit(t, raw, mmsGroupPath())
						expected = 5004
					case "later-time":
						edit(recipientPath(1202), []byte{0, 0, 0, 0})
						expected = 5004
					case "later-unit":
						c := uint32(437)
						if kind == "term" {
							c = 446
						}
						raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: c}))
						expected = 5004
					}
					_, cause := co.Decode(raw)
					if cause == nil {
						t.Fatal("accepted")
					}
					q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					want := smsValueQuote(t, node, -1)
					if fault == "marker" {
						expected = 5004
						want = invalidPart(t, avp.Key{Code: 455}, 64, []byte{0, 0, 0, 0})
					}
					if ok != (expected != 0) || code != expected || code == 5004 && !bytes.Equal(q, want) {
						t.Fatal(cause, code, q)
					}
				})
			}
		}
	}
}
