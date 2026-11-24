// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS address variable quotes, public consumption and mixed-error boundaries.
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

func smsDataSamples() [][]byte { return [][]byte{{0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 1024)} }
func smsDataRaw(t *testing.T, raw []byte, kind string, code uint32, data []byte) []byte {
	t.Helper()
	if kind == "general" || kind == "balance" || kind == "refund" || (code == 2008 && kind != "iec") {
		return unitAppendRaw(t, raw, smsPath(code)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, data))
	}
	b, _ := invalidEdit(t, raw[20:], smsPath(code), data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func smsDataQuote(t *testing.T, code uint32, data []byte) []byte {
	leaf := invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, data)
	return invalidPart(t, avp.Key{VendorID: 10415, Code: 873}, 192, invalidPart(t, avp.Key{VendorID: 10415, Code: 2000}, 192, leaf))
}
func TestRejectSMSData(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
			for _, selected := range []uint32{2008, 2010, 2017} {
				for sample, data := range smsDataSamples() {

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
							req := combinedDuplicateRequest(t, n, co, kind, 13401, 23401)

							raw := smsDataRaw(t, reqBytes(t, co, req), kind, selected, data)
							want := smsDataQuote(t, selected, data)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13402, 23402)
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

func TestRejectSMSDataEmergencyBudget(t *testing.T) {
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
		raw := smsDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "general", 13421, 23421)), "general", 2017, bytes.Repeat([]byte{0}, 1024))
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

func smsDataFault(t *testing.T, raw []byte, kind string, code uint32, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "sequence":
		edit(lengthPath(avp.Key{Code: 415}), addressValueData(9))
	case "node":
		edit(smsPath(2016), addressValueData(9))
	case "client-nul":
		if kind == "refund" {
			return unitAppendRaw(t, raw, smsPath(code)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: 2018}, 192, []byte{0}))
		}
		edit(smsPath(2018), []byte{0})
	case "earlier-nul":
		return subscriptionDataRaw(t, raw, 0, []byte{0})
	case "earlier-width":
		edit(lengthPath(avp.Key{Code: 416}), []byte{1})
	case "scalar-width":
		edit(smsPath(2016), []byte{1})
	case "interface-width":
		edit(append(smsPath(2009), missingSelector{key: avp.Key{VendorID: 10415, Code: 2006}}), []byte{1})
	case "missing-required":
		return mmsGroupOmit(t, raw, smsPath(2016))
	case "duplicate-data":
		return unitAppendRaw(t, raw, smsPath(code)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, []byte{0}))
	case "extra", "mandatory-extra":
		flags := byte(0)
		if fault == "mandatory-extra" {
			flags = 64
		}
		return unitAppendRaw(t, raw, smsPath(code)[:2], invalidPart(t, avp.Key{Code: 9999}, flags, nil))
	case "other-nul":
		other := uint32(2017)
		if code == 2017 {
			other = 2010
		}
		return smsDataRaw(t, raw, kind, other, []byte{0})
	}
	return raw
}
func TestSMSDataPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, code := range []uint32{2008, 2010, 2017} {
			for _, fault := range []string{"plain", "sequence", "node", "client-nul", "earlier-nul", "earlier-width", "scalar-width", "interface-width", "missing-required", "duplicate-data", "extra", "mandatory-extra", "other-nul"} {
				t.Run(fmt.Sprintf("%s/%d/%s", kind, code, fault), func(t *testing.T) {
					raw := smsDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13410, 23410)), kind, code, []byte{1, 0, 2})
					raw = smsDataFault(t, raw, kind, code, fault)
					_, cause := co.Decode(raw)
					var proof interface{ DiameterSMSDataValueV1() string }
					want := fault == "plain" || fault == "sequence" || fault == "node" || fault == "client-nul"
					if errors.As(cause, &proof) != want {
						t.Fatal("order", cause)
					}
					out, result, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					expected := uint32(0)
					if want || fault == "earlier-nul" {
						expected = 5004
					}
					if fault == "earlier-width" || fault == "scalar-width" || fault == "interface-width" {
						expected = 5014
					}
					if fault == "mandatory-extra" {
						expected = 5001
					}
					if ok != (expected != 0) || result != expected || want && !bytes.Equal(out, smsDataQuote(t, code, []byte{1, 0, 2})) {
						t.Fatal(cause, result, out)
					}
					if fault == "earlier-nul" && !bytes.Equal(out, subscriptionDataQuote(t, []byte{0})) {
						t.Fatal("earlier instance")
					}
					if fault == "extra" || fault == "other-nul" {
						var f *validation.FieldError
						if !errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported) || !errors.As(cause, &f) || len(f.Path) != 2 {
							t.Fatal("mixed path", cause)
						}
					}
				})
			}
		}
	}
}
func TestSMSDataGroupOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{2008, 2010, 2017} {
		for _, move := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", code, move), func(t *testing.T) {
				raw := smsDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13413, 23413)), "iec", code, []byte{0})
				b, _ := invalidEdit(t, raw[20:], recipientPath(1202), []byte{1}, avp.Key{})
				raw = append(raw[:20:20], b...)
				wire.Put24(raw[1:4], uint32(len(raw)))
				if move {
					d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
					var sms []byte
					for {
						v, more, e := d.Next()
						if e != nil {
							t.Fatal(e)
						}
						if !more {
							break
						}
						if v.Key == (avp.Key{VendorID: 10415, Code: 873}) {
							child, e := d.Child(v.Data)
							if e != nil {
								t.Fatal(e)
							}
							for {
								x, yes, e := child.Next()
								if e != nil {
									t.Fatal(e)
								}
								if !yes {
									break
								}
								if x.Key == (avp.Key{VendorID: 10415, Code: 2000}) {
									sms = invalidPart(t, x.Key, x.Flags, x.Data)
								}
							}
						}
					}
					raw = mmsGroupOmit(t, raw, smsPath(code)[:2])
					raw = unitAppendRaw(t, raw, smsPath(code)[:1], sms)
				}
				_, cause := co.Decode(raw)
				out, result, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(5004)
				if move {
					expected = 5014
				}
				if !ok || result != expected || !move && !bytes.Equal(out, smsDataQuote(t, code, []byte{0})) {
					t.Fatal("group order", cause, result, out)
				}
			})
		}
	}
}
func TestRejectSMSDataClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{2008, 2010, 2017} {
			for _, fault := range []string{"default", "extra", "other-nul"} {
				t.Run(fmt.Sprintf("%d/%s", code, fault), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw := smsDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13411, 23411)), "iec", code, []byte{0})
					raw = smsDataFault(t, raw, "iec", code, fault)
					sendRejectRaw(t, remote, raw)
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected answer %x", b)
					}
				})
			}
		}
	})
}
