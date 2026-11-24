// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS client variable quotes, public consumption and mixed-error boundaries.
// SPDX-AI-Date: 2026-11-24

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func smsClientSamples() [][]byte {
	return [][]byte{nil, []byte("different"), {0}, {1, 0, 2}, bytes.Repeat([]byte{1}, 1024), bytes.Repeat([]byte{0}, 1024)}
}
func smsClientRaw(t *testing.T, raw []byte, kind string, code uint32, data []byte) []byte {
	b, _ := invalidEdit(t, raw[20:], smsPath(code), data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func smsClientQuote(t *testing.T, code uint32, data []byte) []byte {
	return smsDataQuote(t, code, data)
}
func TestRejectSMSClient(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, selected := range []uint32{2018} {
				for sample, data := range smsClientSamples() {

					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%d/emergency-%t", kind, selected, sample, emergency), func(t *testing.T) {
							cfg := combinedConfig(t)
							cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
							cfg.Peers[0].InboundCommands[0].MaxRequestBytes = 4096
							cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 4096
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12290, 22900))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 13501, 23401)

							raw := smsClientRaw(t, reqBytes(t, co, req), kind, selected, data)
							want := smsClientQuote(t, selected, data)
							_, cause := co.Decode(raw)
							if !errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported) {
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13502, 23402)
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

func TestRejectSMSClientEmergencyBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := combinedConfig(t)
		cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
		cfg.Peers[0].InboundCommands[0].MaxRequestBytes = 4096
		cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 4096
		cfg.Resources.RejectSlots = 2
		cfg.Resources.EmergencyRequestMax = 4096
		cfg.Resources.EmergencyReplyMax = 256
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		defer closeDuplicateNode(t, n)
		p, remote := networkConnectNode(t, n, co, encrypted)
		combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, "iec", 13520, 22920))
		held, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		before := n.Resources()
		raw := smsClientRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13521, 23521)), "iec", 2018, bytes.Repeat([]byte{0}, 1024))
		if len(raw) > cfg.Resources.EmergencyRequestMax {
			t.Fatal("request budget")
		}
		_, cause := co.Decode(raw)
		if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); e != nil {
			t.Fatal("large reply", e)
		}
		if _, e := n.rejectFrame(raw, cfg.Resources.EmergencyReplyMax, rejectDecode, cause); e == nil {
			t.Fatal("short reply")
		}
		sendRejectRaw(t, remote, raw)
		waitFor(t, func() bool { return !p.Info().Ready })
		if b, e := readAdapter(ctxTest(t), remote); e == nil {
			t.Fatalf("truncated answer %x", b)
		}
		waitFor(t, func() bool { return n.Resources().RejectOutstanding == 0 })
		if n.Resources().InboundOutstanding != before.InboundOutstanding || n.Resources().InboundBytes != before.InboundBytes {
			t.Fatal("held charge")
		}
		if result, e := held.Reject(ctxTest(t), FailureContextUnavailable); !result.Committed || e == nil {
			t.Fatal("closed connection completion", result, e)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}

func TestSMSClientPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, data := range [][]byte{{0}, []byte("different")} {
			for _, fault := range []string{"plain", "sequence", "node", "type", "missing", "later-mms", "later-action", "earlier-width", "interface-width", "mms-width", "smsc-nul", "earlier-nul", "extra", "mandatory-extra", "root-extra", "duplicate"} {
				t.Run(fmt.Sprintf("%s/%x/%s", kind, data, fault), func(t *testing.T) {
					raw := smsClientRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13510, 23510)), kind, 2018, data)
					edit := func(path []missingSelector, value []byte) {
						b, _ := invalidEdit(t, raw[20:], path, value, avp.Key{})
						raw = append(raw[:20:20], b...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					switch fault {
					case "sequence":
						edit(lengthPath(avp.Key{Code: 415}), addressValueData(9))
					case "node":
						edit(smsPath(2016), addressValueData(9))
					case "type":
						edit(smsPath(2007), addressValueData(9))
					case "missing":
						raw = mmsGroupOmit(t, raw, smsPath(2011))
					case "later-mms":
						raw = mmsGroupOmit(t, raw, recipientPath(1202)[:2])
					case "later-action":
						if kind == "iec" {
							edit(lengthPath(avp.Key{Code: 436}), addressValueData(9))
						} else {
							raw = smsClientAppendRoot(raw, invalidPart(t, avp.Key{Code: 436}, 64, addressValueData(9)))
						}
					case "earlier-width":
						edit(lengthPath(avp.Key{Code: 416}), []byte{1})
					case "interface-width":
						edit(append(smsPath(2009), missingSelector{key: avp.Key{VendorID: 10415, Code: 2006}}), []byte{1})
					case "mms-width":
						edit(recipientPath(1202), []byte{1})
					case "smsc-nul":
						edit(smsPath(2017), []byte{0})
					case "earlier-nul":
						raw = subscriptionDataRaw(t, raw, 0, []byte{0})
					case "extra", "mandatory-extra":
						flags := byte(0)
						if fault == "mandatory-extra" {
							flags = 64
						}
						raw = unitAppendRaw(t, raw, smsPath(2018)[:2], invalidPart(t, avp.Key{Code: 9999}, flags, nil))
					case "root-extra":
						raw = smsClientAppendRoot(raw, invalidPart(t, avp.Key{Code: 9999}, 0, nil))
					case "duplicate":
						raw = unitAppendRaw(t, raw, smsPath(2018)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: 2018}, 192, data))
					}
					_, cause := co.Decode(raw)
					var proof interface{ DiameterSMSClientValueV1() string }
					want := fault == "plain" || fault == "later-mms" || fault == "later-action" || fault == "root-extra"
					if errors.As(cause, &proof) != want {
						t.Fatal("order", cause)
					}
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					expected := uint32(0)
					if want || fault == "node" || fault == "type" || fault == "smsc-nul" || fault == "earlier-nul" {
						expected = 5004
					}
					if fault == "earlier-width" || fault == "interface-width" || fault == "mms-width" {
						expected = 5014
					}
					if fault == "mandatory-extra" {
						expected = 5001
					}
					if ok != (expected != 0) || code != expected {
						t.Fatal(cause, code, out)
					}
					if want && !bytes.Equal(out, smsClientQuote(t, 2018, data)) {
						t.Fatal("wrong leaf")
					}
					if fault == "smsc-nul" && !bytes.Equal(out, smsDataQuote(t, 2017, []byte{0})) {
						t.Fatal("group before request")
					}
				})
			}
		}
	}
}
func TestRejectSMSClientDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, data := range [][]byte{{0}, []byte("different")} {
			cfg := combinedConfig(t)
			cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
			n, co := newInboundNode(t, cfg)
			p, remote := networkConnectNode(t, n, co, encrypted)
			raw := smsClientRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13511, 23511)), "iec", 2018, data)
			sendRejectRaw(t, remote, raw)
			waitFor(t, func() bool { return !p.Info().Ready })
			if b, e := readAdapter(ctxTest(t), remote); e == nil {
				t.Fatalf("unexpected %x", b)
			}
			closeDuplicateNode(t, n)
		}
	})
}

func smsClientAppendRoot(raw, extra []byte) []byte {
	raw = append(raw, extra...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
