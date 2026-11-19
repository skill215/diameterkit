// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Service-Identifier/Timezone isolation, mixed errors and wire-order acceptance.
// SPDX-AI-Date: 2026-11-19

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

func moneyParameterKey(zone bool) avp.Key {
	if zone {
		return avp.Key{VendorID: 10425, Code: 23}
	}
	return avp.Key{Code: 439}
}
func moneyParameterPath(zone bool) []missingSelector { return lengthPath(moneyParameterKey(zone)) }
func moneyParameterQuote(t *testing.T, zone bool, v int32) []byte {
	flags := byte(64)
	if zone {
		flags = 192
	}
	return invalidPart(t, moneyParameterKey(zone), flags, addressValueData(v))
}
func TestRejectMoneyParameter(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance"} {
			for _, zone := range []bool{false, true} {
				for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
					if (!zone && value == 1) || (zone && value == 0) {
						continue
					}
					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%t/%d/emergency-%t", kind, zone, value, emergency), func(t *testing.T) {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12280, 22800))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12801, 22801)
							data := addressValueData(value)
							raw, _ := invalidRaw(t, co, req, moneyParameterPath(zone), data, avp.Key{})
							want := moneyParameterQuote(t, zone, value)
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
							field.Path[0] = moneyParameterKey(zone)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12802, 22802)
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

func moneyParameterFault(t *testing.T, raw []byte, kind string, zone bool, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "both":
		edit(moneyParameterPath(!zone), addressValueData(-1))
	case "no-service":
		raw = mmsGroupOmit(t, raw, moneyParameterPath(false))
	case "no-timezone":
		raw = mmsGroupOmit(t, raw, moneyParameterPath(true))
	case "no-client":
		raw = smsOmit(t, raw, 2018)
	case "no-mms":
		raw = mmsGroupOmit(t, raw, mmsGroupPath())
	case "no-standard":
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}, avp.Key{Code: 413}))
	case "money":
		raw = unitValueFault(t, raw, "general", "money")
	case "refund":
		raw = append(raw, invalidPart(t, avp.Key{Code: 1292}, 64, nil)...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	case "node":
		edit(smsValuePath(true), addressValueData(0))
	case "selected-width":
		edit(moneyParameterPath(zone), []byte{1})
	case "counterpart-width":
		edit(moneyParameterPath(!zone), []byte{1})
	case "later-client":
		edit(smsPath(2018), []byte{0, 0, 0})
	case "later-mms":
		raw = unitAppendRaw(t, raw, mmsGroupPath(), invalidPart(t, avp.Key{VendorID: 10415, Code: 1209}, 192, addressValueData(-1)))
	case "client-width":
		edit(smsPath(2016), []byte{1})
	default:
		raw = moneyNodeFault(t, raw, kind, fault)
	}
	return raw
}
func TestMoneyParameterPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance"} {
		for _, zone := range []bool{false, true} {
			for _, fault := range []string{"plain", "both", "sequence", "subscription", "action", "missing-action", "missing-unit", "used", "reason", "node", "forbidden-2008", "no-service", "no-timezone", "no-client", "no-mms", "no-standard", "money", "refund", "selected-width", "counterpart-width", "client-width", "later-client", "later-mms", "later-address", "extra", "mandatory-extra"} {
				t.Run(fmt.Sprintf("%s/%t/%s", kind, zone, fault), func(t *testing.T) {
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12811, 22811), moneyParameterPath(zone), addressValueData(-1), avp.Key{})
					raw = moneyParameterFault(t, raw, kind, zone, fault)
					_, cause := co.Decode(raw)
					if cause == nil {
						t.Fatal("accepted")
					}
					var proof interface{ DiameterMoneyParameterValueV1() string }
					want := fault == "plain" || fault == "later-client" || fault == "later-mms"
					if errors.As(cause, &proof) != want {
						t.Fatal("order", cause)
					}
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					expected := uint32(0)
					if want || fault == "action" || fault == "node" || fault == "later-address" {
						expected = 5004
					}
					if fault == "selected-width" || fault == "counterpart-width" || fault == "client-width" {
						expected = 5014
					}
					if fault == "mandatory-extra" {
						expected = 5001
					}
					if ok != (expected != 0) || code != expected {
						t.Fatal(cause, code, out)
					}
					if want && !bytes.Equal(out, moneyParameterQuote(t, zone, -1)) {
						t.Fatal("wrong instance")
					}
					if fault == "node" && !bytes.Equal(out, smsValueQuote(t, true, 0)) {
						t.Fatal("earlier node")
					}
					if fault == "later-address" && !bytes.Equal(out, addressValueQuote(t, 886, -1)) {
						t.Fatal("earlier group rule")
					}
				})
			}
		}
	}
}
func TestRejectMoneyParameterClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance"} {
			for _, zone := range []bool{false, true} {
				for _, fault := range []string{"default", "both", "no-service", "no-timezone", "no-client", "no-mms", "no-standard", "money", "refund"} {
					t.Run(fmt.Sprintf("%s/%t/%s", kind, zone, fault), func(t *testing.T) {
						cfg := combinedConfig(t)
						cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
						if fault == "default" {
							cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
						}
						n, co := newInboundNode(t, cfg)
						defer closeDuplicateNode(t, n)
						p, remote := networkConnectNode(t, n, co, encrypted)
						raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12812, 22812), moneyParameterPath(zone), addressValueData(-1), avp.Key{})
						raw = moneyParameterFault(t, raw, kind, zone, fault)
						sendRejectRaw(t, remote, raw)
						waitFor(t, func() bool { return !p.Info().Ready })
						if b, e := readAdapter(ctxTest(t), remote); e == nil {
							t.Fatalf("unexpected reply %x", b)
						}
					})
				}
			}
		}
	})
}
