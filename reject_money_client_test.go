// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Money client-wire value modes, variable quotes and public consumption boundaries.
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

func moneyClientSamples() [][]byte {
	return [][]byte{nil, {0}, {1, 2, 3}, {1, 2, 3, 4}, {0, 1, 127, 0, 0}, {0, 1, 127, 0, 0, 2}, {0, 1, 127, 0, 0, 1, 0}, []byte("127.0.0.1"), bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{1}, 1024)}
}
func moneyClientQuote(t *testing.T, data []byte) []byte {
	leaf := invalidPart(t, avp.Key{VendorID: 10415, Code: 2018}, 192, data)
	return invalidPart(t, avp.Key{VendorID: 10415, Code: 873}, 192, invalidPart(t, avp.Key{VendorID: 10415, Code: 2000}, 192, leaf))
}
func TestRejectMoneyClient(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance"} {
			{
				for sample, data := range moneyClientSamples() {

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
							req := combinedDuplicateRequest(t, n, co, kind, 12901, 22901)

							raw, _ := invalidRaw(t, co, req, smsPath(2018), data, avp.Key{})
							want := moneyClientQuote(t, data)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12902, 22902)
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

func moneyClientFault(t *testing.T, raw []byte, kind, fault string) []byte {
	if fault == "service" || fault == "timezone" {
		b, _ := invalidEdit(t, raw[20:], moneyParameterPath(fault == "timezone"), addressValueData(-1), avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
		return raw
	}
	return moneyParameterFault(t, raw, kind, false, fault)
}
func TestMoneyClientPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance"} {
		for _, fault := range []string{"plain", "sequence", "subscription", "action", "missing-action", "missing-unit", "used", "reason", "node", "forbidden-2008", "no-service", "no-timezone", "no-client", "no-mms", "no-standard", "money", "refund", "service", "timezone", "selected-width", "counterpart-width", "client-width", "later-mms", "later-address", "extra", "mandatory-extra"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12911, 22911), smsPath(2018), []byte{1, 2, 3}, avp.Key{})
				raw = moneyClientFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				var proof interface{ DiameterMoneyClientValueV1() string }
				want := fault == "plain" || fault == "later-mms"
				if errors.As(cause, &proof) != want {
					t.Fatal("order", cause)
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if want || fault == "action" || fault == "node" || fault == "service" || fault == "timezone" || fault == "later-address" {
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
				if want && !bytes.Equal(out, moneyClientQuote(t, []byte{1, 2, 3})) {
					t.Fatal("wrong instance")
				}
				if (fault == "service" || fault == "timezone") && !bytes.Equal(out, moneyParameterQuote(t, fault == "timezone", -1)) {
					t.Fatal("earlier parameter")
				}
				if fault == "later-address" && !bytes.Equal(out, addressValueQuote(t, 886, -1)) {
					t.Fatal("earlier group")
				}
			})
		}
	}
}
func TestRejectMoneyClientClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance"} {
			for _, fault := range []string{"default", "no-service", "no-timezone", "no-client", "no-mms", "no-standard", "money", "refund"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					cfg.Peers[0].InboundCommands[0].MaxRequestBytes = 4096
					cfg.Peers[0].InboundCommands[0].MaxAnswerBytes = 4096
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12912, 22912), smsPath(2018), []byte{1}, avp.Key{})
					raw = moneyClientFault(t, raw, kind, fault)
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
func TestRejectMoneyClientEmergencyBudget(t *testing.T) {
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
		raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, "general", 12921, 22921), smsPath(2018), bytes.Repeat([]byte{1}, 1024), avp.Key{})
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
