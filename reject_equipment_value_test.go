// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined equipment-type value acceptance.
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

func equipmentValueData(value int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(value))
	return b
}
func equipmentValueQuote(t *testing.T, value int32) []byte {
	return invalidPart(t, avp.Key{Code: 458}, 64, invalidPart(t, avp.Key{Code: 459}, 64, equipmentValueData(value)))
}
func TestRejectEquipmentValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			{
				for _, value := range []int32{-2147483648, -1, 1, 2, 2147483647} {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12250, 22500))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12501, 22501)
							data := equipmentValueData(value)
							raw, _ := invalidRaw(t, co, req, lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), data, avp.Key{})
							want := equipmentValueQuote(t, value)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12502, 22502)
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

func equipmentFault(t *testing.T, raw []byte, kind, fault string) []byte {
	t.Helper()
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
	}
	switch fault {
	case "sequence", "subscriptions", "sms", "mms", "timestamp", "interface":
		return unitValueFault(t, raw, kind, fault)
	case "nul":
		edit(lengthPath(avp.Key{Code: 458}, avp.Key{Code: 460}), []byte{'A', 0, 'B'})
	case "extra", "mandatory-extra":
		flags := byte(0)
		if fault == "mandatory-extra" {
			flags = 64
		}
		return unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 458}), invalidPart(t, avp.Key{Code: 9999}, flags, []byte{1}))
	case "missing-value":
		return mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 458}, avp.Key{Code: 460}))
	case "duplicate-type":
		return unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 458}), invalidPart(t, avp.Key{Code: 459}, 64, equipmentValueData(1)))
	case "type-width":
		edit(lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), []byte{1})
	case "earlier-width":
		edit(lengthPath(avp.Key{Code: 443}, avp.Key{Code: 450}), []byte{1})
	case "later-width", "reordered-width":
		edit(smsValuePath(true), []byte{1})
		if fault == "reordered-width" {
			d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
			var rest, group []byte
			for {
				v, ok, e := d.Next()
				if e != nil {
					t.Fatal(e)
				}
				if !ok {
					break
				}
				part := invalidPart(t, v.Key, v.Flags, v.Data)
				if v.Key == (avp.Key{Code: 458}) {
					group = part
				} else {
					rest = append(rest, part...)
				}
			}
			raw = append(raw[:20:20], append(rest, group...)...)
		}
	}
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func TestEquipmentValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "interface", "nul", "extra", "mandatory-extra", "missing-value", "duplicate-type", "type-width", "earlier-width", "later-width", "reordered-width"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12510, 22510), lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), equipmentValueData(-1), avp.Key{})
				raw = equipmentFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(5004)
				switch fault {
				case "nul", "extra", "missing-value", "duplicate-type":
					expected = 0
				case "mandatory-extra":
					expected = 5001
				case "type-width", "earlier-width", "reordered-width":
					expected = 5014
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				var proof interface{ DiameterEquipmentTypeValueV1() string }
				if errors.As(cause, &proof) != (expected == 5004) {
					t.Fatal("precedence", cause)
				}
				if expected == 5004 && !bytes.Equal(out, equipmentValueQuote(t, -1)) {
					t.Fatal("wrong instance")
				}
				if fault == "nul" || fault == "extra" {
					if !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrField) {
						t.Fatal("mixed classification", cause)
					}
				}
				if fault == "missing-value" && !errors.Is(cause, validation.ErrMissing) {
					t.Fatal(cause)
				}
				if fault == "duplicate-type" && !errors.Is(cause, validation.ErrDuplicate) {
					t.Fatal(cause)
				}
			})
		}
	}
}
func TestRejectEquipmentValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, fault := range []string{"default", "nul", "extra"} {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12511, 22511), lengthPath(avp.Key{Code: 458}, avp.Key{Code: 459}), equipmentValueData(-1), avp.Key{})
					raw = equipmentFault(t, raw, kind, fault)
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
