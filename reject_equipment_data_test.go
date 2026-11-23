// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Equipment data variable quotes, public consumption and mixed-error boundaries.
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

func equipmentDataSamples() [][]byte {
	return [][]byte{{0}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)}
}
func equipmentDataPath() []missingSelector { return lengthPath(avp.Key{Code: 458}, avp.Key{Code: 460}) }
func equipmentDataRaw(t *testing.T, raw []byte, kind string, data []byte) []byte {
	t.Helper()
	b, _ := invalidEdit(t, raw[20:], equipmentDataPath(), data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func equipmentDataQuote(t *testing.T, data []byte) []byte {
	return invalidPart(t, avp.Key{Code: 458}, 64, invalidPart(t, avp.Key{Code: 460}, 64, data))
}
func TestRejectEquipmentData(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			{
				for sample, data := range equipmentDataSamples() {

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
							req := combinedDuplicateRequest(t, n, co, kind, 13101, 23101)

							raw := equipmentDataRaw(t, reqBytes(t, co, req), kind, data)
							want := equipmentDataQuote(t, data)
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
							field.Path[0] = avp.Key{Code: 458}
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13102, 23102)
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

func TestRejectEquipmentDataEmergencyBudget(t *testing.T) {
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
		combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, "iec", 12920, 22920))
		held, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		before := n.Resources()
		raw := equipmentDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13121, 23121)), "iec", bytes.Repeat([]byte{0}, 1024))
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

func equipmentDataFault(t *testing.T, raw []byte, kind, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "type", "type-only":
		edit(lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), equipmentValueData(1))
		if fault == "type-only" {
			edit(equipmentDataPath(), []byte("id"))
		}
		return raw
	case "missing-type":
		return mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}))
	case "duplicate-value":
		return unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 458}), invalidPart(t, avp.Key{Code: 460}, 64, []byte{0}))
	case "earlier-nul":
		edit(lengthPath(avp.Key{Code: 443}, avp.Key{Code: 444}), []byte{0})
		return raw
	}
	return equipmentFault(t, raw, kind, fault)
}
func TestEquipmentDataPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "interface", "type", "type-only", "extra", "mandatory-extra", "missing-value", "missing-type", "duplicate-type", "duplicate-value", "type-width", "earlier-width", "earlier-nul", "later-width", "reordered-width"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw := equipmentDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13110, 23110)), kind, []byte{1, 0, 2})
				raw = equipmentDataFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				var proof interface{ DiameterEquipmentDataValueV1() string }
				want := fault == "plain" || fault == "sequence" || fault == "subscriptions" || fault == "sms" || fault == "mms" || fault == "timestamp" || fault == "interface" || fault == "later-width"
				if errors.As(cause, &proof) != want {
					t.Fatal("order", cause)
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if want || fault == "type-only" || fault == "earlier-nul" {
					expected = 5004
				}
				if fault == "type-width" || fault == "earlier-width" || fault == "reordered-width" {
					expected = 5014
				}
				if fault == "mandatory-extra" {
					expected = 5001
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				if want && !bytes.Equal(out, equipmentDataQuote(t, []byte{1, 0, 2})) {
					t.Fatal("wrong instance")
				}
				if fault == "type-only" && !bytes.Equal(out, equipmentValueQuote(t, 1)) {
					t.Fatal("type priority")
				}
				if fault == "earlier-nul" && !bytes.Equal(out, subscriptionDataQuote(t, []byte{0})) {
					t.Fatal("earlier subscription instance")
				}
				if fault == "extra" && (!errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported)) {
					t.Fatal("mixed extra", cause)
				}
				if fault == "type" && (!errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrField)) {
					t.Fatal("mixed type/value", cause)
				}
			})
		}
	}
}
func TestRejectEquipmentDataEarlierOrClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, fault := range []string{"default", "type", "extra", "earlier-nul"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 13111, 23111), equipmentDataPath(), []byte{0}, avp.Key{})
					raw = equipmentDataFault(t, raw, kind, fault)
					sendRejectRaw(t, remote, raw)
					if fault == "earlier-nul" {
						h, _, e := wire.Header(raw, cfg.Limits)
						if e != nil {
							t.Fatal(e)
						}
						ans := readReject(t, remote, h, 5004)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, subscriptionDataQuote(t, []byte{0})) {
							t.Fatal("earlier subscription quote")
						}
						noDuplicateDelivery(t, p)
						return
					}
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected reply %x", b)
					}
				})
			}
		}
	})
}
