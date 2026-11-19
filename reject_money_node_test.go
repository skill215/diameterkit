// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated money SMS-Node values, mixed errors and validation order.
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

func TestRejectMoneyNode(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance", "refund"} {
			{
				for _, value := range []int32{-2147483648, -1, 0, 1, 2147483647} {
					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/emergency-%t", kind, value, emergency), func(t *testing.T) {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12270, 22700))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12701, 22701)
							data := addressValueData(value)
							raw, _ := invalidRaw(t, co, req, smsValuePath(true), data, avp.Key{})
							want := smsValueQuote(t, true, value)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12702, 22702)
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

func moneyNodeFault(t *testing.T, raw []byte, kind, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "sequence":
		edit(lengthPath(avp.Key{Code: 415}), addressValueData(1))
	case "subscription":
		edit([]missingSelector{{key: avp.Key{Code: 443}, occurrence: 0}, {key: avp.Key{Code: 450}}}, addressValueData(9))
	case "action":
		edit(lengthPath(avp.Key{Code: 436}), addressValueData(3))
	case "missing-action":
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 436}))
	case "missing-unit":
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}))
	case "used":
		raw = unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 456}), invalidPart(t, avp.Key{Code: 446}, 64, invalidPart(t, avp.Key{Code: 417}, 64, make([]byte, 8))))
	case "reason":
		raw = unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 456}), invalidPart(t, avp.Key{VendorID: 10415, Code: 872}, 192, addressValueData(2)))
	case "selected-width":
		edit(smsValuePath(true), []byte{1})
	case "message-width":
		edit(smsValuePath(false), []byte{1})
	case "message-value":
		edit(smsValuePath(false), addressValueData(-99))
	case "later-refund":
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 1292}))
	case "later-service":
		edit(lengthPath(avp.Key{Code: 439}), addressValueData(0))
	case "later-client":
		edit(smsPath(2018), []byte{1})
	case "later-mms":
		raw = mmsGroupOmit(t, raw, mmsGroupPath())
	case "later-address":
		edit(addressValuePath(886, 899), addressValueData(-1))
	case "extra", "mandatory-extra":
		flags := byte(0)
		if fault == "mandatory-extra" {
			flags = 64
		}
		raw = unitAppendRaw(t, raw, smsPath(0)[:2], invalidPart(t, avp.Key{Code: 9999}, flags, nil))
	default:
		var c uint32
		if _, e := fmt.Sscanf(fault, "forbidden-%d", &c); e == nil {
			data := []byte{}
			if c == 2011 || c == 2012 {
				data = make([]byte, 4)
			}
			raw = unitAppendRaw(t, raw, smsPath(0)[:2], invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, data))
		}
	}
	return raw
}
func TestMoneyNodePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance", "refund"} {
		faults := []string{"plain", "sequence", "subscription", "action", "missing-action", "missing-unit", "used", "reason", "selected-width", "message-width", "message-value", "extra", "mandatory-extra"}
		for _, c := range []uint32{2008, 2010, 2017, 2011, 2015, 2014, 2012} {
			faults = append(faults, fmt.Sprintf("forbidden-%d", c))
		}
		if kind == "refund" {
			faults = append(faults, "later-refund")
		} else {
			faults = append(faults, "later-service", "later-client", "later-mms", "later-address")
		}
		for _, fault := range faults {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12711, 22711), smsValuePath(true), addressValueData(-1), avp.Key{})
				raw = moneyNodeFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				var proof interface{ DiameterMoneyNodeValueV1() string }
				want := fault == "plain" || fault == "message-value" || fault == "later-refund" || fault == "later-service" || fault == "later-client" || fault == "later-mms"
				if errors.As(cause, &proof) != want {
					t.Fatal("order", cause)
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if want || fault == "action" || fault == "later-address" {
					expected = 5004
				}
				if fault == "mandatory-extra" {
					expected = 5001
				}
				if fault == "selected-width" || fault == "message-width" {
					expected = 5014
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				if want && !bytes.Equal(out, smsValueQuote(t, true, -1)) {
					t.Fatal("wrong instance")
				}
				if fault == "later-address" && !bytes.Equal(out, addressValueQuote(t, 886, -1)) {
					t.Fatal("structural group must precede root rule")
				}
			})
		}
	}
}
func TestRejectMoneyNodeClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"general", "balance", "refund"} {
			for _, fault := range []string{"default", "missing-unit", "forbidden-2008", "forbidden-2010", "forbidden-2017", "forbidden-2011", "forbidden-2015", "forbidden-2014", "forbidden-2012"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12712, 22712), smsValuePath(true), addressValueData(-1), avp.Key{})
					raw = moneyNodeFault(t, raw, kind, fault)
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
