// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Address data variable quotes, public consumption and mixed-error boundaries.
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

func addressDataSamples() [][]byte {
	return [][]byte{{0}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)}
}
func addressDataPath(group uint32) []missingSelector { return addressValuePath(group, 897) }
func addressDataRaw(t *testing.T, raw []byte, group uint32, data []byte) []byte {
	t.Helper()
	b, _ := invalidEdit(t, raw[20:], addressDataPath(group), data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func addressDataQuote(t *testing.T, group uint32, data []byte) []byte {
	b := invalidPart(t, avp.Key{VendorID: 10415, Code: 897}, 192, data)
	for _, c := range []uint32{group, 877, 873} {
		b = invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, b)
	}
	return b
}
func TestRejectAddressData(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, sampleCase := range addressValueCases {
			kind, group := sampleCase.kind, sampleCase.group
			{
				for sample, data := range addressDataSamples() {

					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%d/emergency-%t", kind, group, sample, emergency), func(t *testing.T) {
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
							req := combinedDuplicateRequest(t, n, co, kind, 13201, 23201)

							raw := addressDataRaw(t, reqBytes(t, co, req), group, data)
							want := addressDataQuote(t, group, data)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13202, 23202)
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

func TestRejectAddressDataEmergencyBudget(t *testing.T) {
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
		raw := addressDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13221, 23221)), 886, bytes.Repeat([]byte{0}, 1024))
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

func addressDataFault(t *testing.T, raw []byte, kind string, group uint32, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "type", "type-only":
		edit(addressValuePath(group, 899), addressValueData(-1))
		if fault == "type-only" {
			edit(addressDataPath(group), []byte("id"))
		}
		return raw
	case "missing-type":
		return mmsGroupOmit(t, raw, addressValuePath(group, 899))
	case "duplicate-data":
		return unitAppendRaw(t, raw, recipientPath(group), invalidPart(t, avp.Key{VendorID: 10415, Code: 897}, 192, []byte{0}))
	case "earlier-nul":
		edit(lengthPath(avp.Key{Code: 443}, avp.Key{Code: 444}), []byte{0})
		return raw
	}
	return addressFault(t, raw, kind, group, fault)
}
func TestAddressDataPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, sampleCase := range addressValueCases {
		kind, group := sampleCase.kind, sampleCase.group
		for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "timestamp", "interface", "type", "type-only", "extra", "mandatory-extra", "missing-data", "missing-type", "duplicate-type", "duplicate-data", "type-width", "earlier-width", "earlier-nul", "later-width", "reordered-width"} {
			// Money templates have no MMS submission-time field to mutate.
			if (kind == "general" || kind == "balance") && (fault == "timestamp" || fault == "later-width" || fault == "reordered-width") {
				continue
			}
			t.Run(fmt.Sprintf("%s/%d/%s", kind, group, fault), func(t *testing.T) {
				raw := addressDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13210, 23210)), group, []byte{1, 0, 2})
				raw = addressDataFault(t, raw, kind, group, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				var proof interface{ DiameterAddressDataValueV1() string }
				want := fault == "plain" || fault == "sequence" || fault == "subscriptions" || fault == "sms" || fault == "timestamp" || fault == "interface" || fault == "later-width"
				if errors.As(cause, &proof) != want {
					t.Fatal("order", cause)
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if want || fault == "type-only" {
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
				if want && !bytes.Equal(out, addressDataQuote(t, group, []byte{1, 0, 2})) {
					t.Fatal("wrong instance")
				}
				if fault == "type-only" && !bytes.Equal(out, addressValueQuote(t, group, -1)) {
					t.Fatal("type priority")
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
func TestRejectAddressDataClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, sampleCase := range addressValueCases {
			kind, group := sampleCase.kind, sampleCase.group
			for _, fault := range []string{"default", "type", "extra", "earlier-nul"} {
				t.Run(fmt.Sprintf("%s/%d/%s", kind, group, fault), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 13211, 23211), addressDataPath(group), []byte{0}, avp.Key{})
					raw = addressDataFault(t, raw, kind, group, fault)
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

func TestAddressDataGroupOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, move := range []bool{false, true} {
			for _, mixed := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%t/%t", kind, move, mixed), func(t *testing.T) {
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12612, 22612), addressValuePath(886, 897), []byte{1, 0, 2}, avp.Key{})
					body, _ := invalidEdit(t, raw[20:], addressValuePath(1201, 897), []byte{0}, avp.Key{})
					raw = append(raw[:20:20], body...)
					wire.Put24(raw[1:4], uint32(len(raw)))
					if mixed {
						raw = addressDataFault(t, raw, kind, 886, "type")
					}
					if move {
						raw = addressMoveLast(t, raw, 886)
					}
					_, cause := co.Decode(raw)
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					if mixed && !move {
						if ok || !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrField) {
							t.Fatal("mixed first", cause)
						}
						return
					}
					group, v := uint32(886), []byte{1, 0, 2}
					if move {
						group, v = 1201, []byte{0}
					}
					if !ok || code != 5004 || !bytes.Equal(out, addressDataQuote(t, group, v)) {
						t.Fatal("group order", cause, code, out)
					}
				})
			}
		}
	}
}
