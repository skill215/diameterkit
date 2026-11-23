// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: MMS SMSKey variable quotes, public consumption and mixed-error boundaries.
// SPDX-AI-Date: 2026-11-23

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

func mmsKeySamples() [][]byte {
	return [][]byte{{0}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)}
}
func mmsKeyPath() []missingSelector { return recipientPath(1210) }
func mmsKeyRaw(t *testing.T, raw []byte, kind string, data []byte) []byte {
	t.Helper()
	if kind == "general" || kind == "balance" {
		return unitAppendRaw(t, raw, mmsGroupPath(), invalidPart(t, avp.Key{VendorID: 10415, Code: 1210}, 192, data))
	}
	body, _ := invalidEdit(t, raw[20:], mmsKeyPath(), data, avp.Key{})
	raw = append(raw[:20:20], body...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func mmsKeyQuote(t *testing.T, data []byte) []byte {
	leaf := invalidPart(t, avp.Key{VendorID: 10415, Code: 1210}, 192, data)
	return invalidPart(t, avp.Key{VendorID: 10415, Code: 873}, 192, invalidPart(t, avp.Key{VendorID: 10415, Code: 877}, 192, leaf))
}
func TestRejectMMSKey(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term", "general", "balance"} {
			{
				for sample, data := range mmsKeySamples() {

					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/emergency-%t", kind, sample, emergency), func(t *testing.T) {
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
							req := combinedDuplicateRequest(t, n, co, kind, 13001, 23001)

							raw := mmsKeyRaw(t, reqBytes(t, co, req), kind, data)
							want := mmsKeyQuote(t, data)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13002, 23002)
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

func TestRejectMMSKeyEmergencyBudget(t *testing.T) {
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
		combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, "general", 12920, 22920))
		held, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		before := n.Resources()
		raw := mmsKeyRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "general", 13021, 23021)), "general", bytes.Repeat([]byte{0}, 1024))
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

func mmsKeyFault(t *testing.T, raw []byte, kind, fault string) []byte {
	switch fault {
	case "sequence", "subscriptions", "sms", "timestamp", "interface":
		return unitValueFault(t, raw, kind, fault)
	case "extra", "mandatory-extra":
		flags := byte(0)
		if fault == "mandatory-extra" {
			flags = 64
		}
		return unitAppendRaw(t, raw, mmsGroupPath(), invalidPart(t, avp.Key{Code: 9999}, flags, []byte{1}))
	case "missing-originator":
		return mmsGroupOmit(t, raw, recipientPath(886))
	case "duplicate-key":
		return unitAppendRaw(t, raw, mmsGroupPath(), invalidPart(t, avp.Key{VendorID: 10415, Code: 1210}, 192, []byte{0}))
	}
	path := mmsKeyPath()
	data := []byte("key")
	switch fault {
	case "plain":
		return raw
	case "address":
		path = addressValuePath(886, 899)
		data = addressValueData(0)
	case "address-nul":
		path = addressValuePath(886, 897)
		data = []byte{0}
	case "scalar-width":
		path = recipientPath(1202)
		data = []byte{1}
	case "earlier-width":
		path = smsPath(2016)
		data = []byte{1}
	case "earlier-nul":
		path = smsPath(2017)
		data = []byte{0}
	}
	b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func TestMMSKeyPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "timestamp", "interface", "extra", "mandatory-extra", "missing-originator", "duplicate-key", "address", "address-nul", "scalar-width", "earlier-width", "earlier-nul", "valid"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw := mmsKeyRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13010, 23010)), kind, []byte{1, 0, 2})
				raw = mmsKeyFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				var proof interface{ DiameterMMSKeyValueV1() string }
				want := fault == "plain" || fault == "sequence" || fault == "subscriptions" || fault == "sms" || fault == "timestamp" || fault == "interface"
				if errors.As(cause, &proof) != want {
					t.Fatal("order", cause)
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if want || fault == "address" || fault == "address-nul" {
					expected = 5004
				}
				if fault == "scalar-width" || fault == "earlier-width" {
					expected = 5014
				}
				if fault == "mandatory-extra" {
					expected = 5001
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				if fault == "address-nul" && !bytes.Equal(out, addressDataQuote(t, 886, []byte{0})) {
					t.Fatal("earlier address instance")
				}
				if want && !bytes.Equal(out, mmsKeyQuote(t, []byte{1, 0, 2})) {
					t.Fatal("wrong instance")
				}
				if fault == "extra" && (!errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported)) {
					t.Fatal("mixed", cause)
				}
				if (cause == nil) != (fault == "valid") {
					t.Fatal("acceptance", cause)
				}
			})
		}
	}
}
func TestRejectMMSKeyEarlierOrClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, fault := range []string{"default", "extra", "address-nul", "earlier-nul"} {
			t.Run(fault, func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				if fault == "default" {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				n, co := newInboundNode(t, cfg)
				defer closeDuplicateNode(t, n)
				p, remote := networkConnectNode(t, n, co, encrypted)
				raw := mmsKeyRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13012, 23012)), "iec", []byte{0})
				if fault != "default" {
					raw = mmsKeyFault(t, raw, "iec", fault)
				}
				sendRejectRaw(t, remote, raw)
				if fault == "address-nul" {
					h, _, e := wire.Header(raw, cfg.Limits)
					if e != nil {
						t.Fatal(e)
					}
					ans := readReject(t, remote, h, 5004)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, addressDataQuote(t, 886, []byte{0})) {
						t.Fatal("earlier address quote")
					}
					noDuplicateDelivery(t, p)
					return
				}
				waitFor(t, func() bool { return !p.Info().Ready })
				if b, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("unexpected answer %x", b)
				}
			})
		}
	})
}
