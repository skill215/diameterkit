// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined indicator value acceptance.
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

func indicatorValueData(value int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(value))
	return b
}
func TestRejectIndicatorValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			{
				for _, value := range []int32{-2147483648, -1, 0, 2, 2147483647} {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12240, 22400))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12401, 22401)
							data := indicatorValueData(value)
							raw, _ := invalidRaw(t, co, req, lengthPath(avp.Key{Code: 455}), data, avp.Key{})
							want := invalidPart(t, avp.Key{Code: 455}, 64, indicatorValueData(value))
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
							field.Path[0] = avp.Key{Code: 455}
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12402, 22402)
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

func indicatorFault(t *testing.T, raw []byte, kind, fault string) []byte {
	t.Helper()
	switch fault {
	case "sequence", "subscriptions", "sms", "mms", "timestamp", "interface":
		return unitValueFault(t, raw, kind, fault)
	case "missing-indicator":
		return mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 455}))
	case "missing-equipment":
		return mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 458}))
	case "service":
		raw = append(raw, invalidPart(t, avp.Key{Code: 439}, 64, indicatorValueData(1))...)
	case "timezone":
		raw = append(raw, invalidPart(t, avp.Key{VendorID: 10425, Code: 23}, 192, indicatorValueData(0))...)
	case "refund":
		raw = append(raw, invalidPart(t, avp.Key{Code: 1292}, 64, []byte{1})...)
	case "width":
		b, _ := invalidEdit(t, raw[20:], lengthPath(avp.Key{Code: 455}), []byte{1}, avp.Key{})
		raw = append(raw[:20:20], b...)
	case "equipment-width":
		b, _ := invalidEdit(t, raw[20:], lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), []byte{1}, avp.Key{})
		raw = append(raw[:20:20], b...)
	}
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func TestIndicatorValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "interface", "missing-indicator", "missing-equipment", "service", "timezone", "refund", "width", "equipment-width"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12410, 22410), lengthPath(avp.Key{Code: 455}), indicatorValueData(0), avp.Key{})
				raw = indicatorFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				switch fault {
				case "plain", "sms", "mms", "timestamp", "interface":
					expected = 5004
				case "width", "equipment-width":
					expected = 5014
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				if expected == 5004 && !bytes.Equal(out, invalidPart(t, avp.Key{Code: 455}, 64, indicatorValueData(0))) {
					t.Fatal("order")
				}
				if fault == "missing-indicator" || fault == "missing-equipment" || fault == "service" || fault == "timezone" || fault == "refund" {
					if !errors.Is(cause, validation.ErrField) {
						t.Fatal("mixed classification", cause)
					}
					var proof interface{ DiameterIndicatorValueV1() string }
					if errors.As(cause, &proof) {
						t.Fatal("mixed proof")
					}
				}
			})
		}
	}
}
func TestRejectIndicatorValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, fault := range []string{"default", "missing-indicator", "missing-equipment", "service", "timezone", "refund"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12411, 22411), lengthPath(avp.Key{Code: 455}), indicatorValueData(0), avp.Key{})
					raw = indicatorFault(t, raw, kind, fault)
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
