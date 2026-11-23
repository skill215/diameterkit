// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Subscription data variable quotes, public consumption and mixed-error boundaries.
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

func subscriptionDataSamples() [][]byte {
	return [][]byte{{0}, {1, 0}, {1, 0, 2}, bytes.Repeat([]byte{0}, 17), bytes.Repeat([]byte{0}, 1024)}
}
func subscriptionDataPath(index int) []missingSelector {
	return []missingSelector{{key: avp.Key{Code: 443}, occurrence: index}, {key: avp.Key{Code: 444}}}
}
func subscriptionDataRaw(t *testing.T, raw []byte, index int, data []byte) []byte {
	t.Helper()
	b, _ := invalidEdit(t, raw[20:], subscriptionDataPath(index), data, avp.Key{})
	raw = append(raw[:20:20], b...)
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func subscriptionDataQuote(t *testing.T, data []byte) []byte {
	return invalidPart(t, avp.Key{Code: 443}, 64, invalidPart(t, avp.Key{Code: 444}, 64, data))
}
func TestRejectSubscriptionData(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
			for _, index := range []int{0, 1} {
				for sample, data := range subscriptionDataSamples() {

					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%d/emergency-%t", kind, index, sample, emergency), func(t *testing.T) {
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
							req := combinedDuplicateRequest(t, n, co, kind, 13301, 23301)

							raw := subscriptionDataRaw(t, reqBytes(t, co, req), index, data)
							want := subscriptionDataQuote(t, data)
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
							field.Path[0] = avp.Key{Code: 443}
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
							follow := combinedDuplicateRequest(t, n, co, kind, 13302, 23302)
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

func TestRejectSubscriptionDataEmergencyBudget(t *testing.T) {
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
		raw := subscriptionDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13321, 23321)), 1, bytes.Repeat([]byte{0}, 1024))
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

// Edit exactly one repeated group, preserving every other original root AVP.
func subscriptionGroupEdit(t *testing.T, raw []byte, index int, change func([]byte) []byte) []byte {
	t.Helper()
	d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
	body := []byte{}
	seen := 0
	for {
		v, more, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !more {
			break
		}
		head := 8
		if v.Key.VendorID != 0 {
			head = 12
		}
		size := (head + len(v.Data) + 3) &^ 3
		part := append([]byte{}, raw[20+v.Offset:20+v.Offset+size]...)
		if v.Key == (avp.Key{Code: 443}) {
			if seen == index {
				part = change(part)
			}
			seen++
		}
		body = append(body, part...)
	}
	if seen <= index {
		t.Fatal("missing group")
	}
	out := append(raw[:20:20], body...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func subscriptionDataFault(t *testing.T, raw []byte, index int, fault string) []byte {
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	path := subscriptionDataPath(index)
	switch fault {
	case "type":
		path[1].key.Code = 450
		edit(path, addressValueData(-1))
		return raw
	case "type-width":
		path[1].key.Code = 450
		edit(path, []byte{1})
		return raw
	case "earlier-width":
		edit(lengthPath(avp.Key{Code: 416}), []byte{1})
		return raw
	case "later-width", "reordered-width":
		edit(lengthPath(avp.Key{Code: 55}), []byte{1})
		if fault == "reordered-width" {
			var part []byte
			raw = subscriptionGroupEdit(t, raw, index, func(b []byte) []byte { part = b; return nil })
			raw = append(raw, part...)
			wire.Put24(raw[1:4], uint32(len(raw)))
		}
		return raw
	case "sequence":
		edit(lengthPath(avp.Key{Code: 415}), addressValueData(9))
		return raw
	case "missing-data":
		return mmsGroupOmit(t, raw, path)
	case "missing-type":
		path[1].key.Code = 450
		return mmsGroupOmit(t, raw, path)
	case "duplicate-type", "duplicate-data", "extra", "mandatory-extra":
		var extra []byte
		switch fault {
		case "duplicate-type":
			extra = invalidPart(t, avp.Key{Code: 450}, 64, addressValueData(0))
		case "duplicate-data":
			extra = invalidPart(t, avp.Key{Code: 444}, 64, []byte{0})
		default:
			flags := byte(0)
			if fault == "mandatory-extra" {
				flags = 64
			}
			extra = invalidPart(t, avp.Key{Code: 9999}, flags, nil)
		}
		return subscriptionGroupEdit(t, raw, index, func(b []byte) []byte { return invalidPart(t, avp.Key{Code: 443}, 64, append(b[8:], extra...)) })
	}
	return raw
}
func TestSubscriptionDataPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, index := range []int{0, 1} {
			for _, fault := range []string{"plain", "type", "sequence", "earlier-width", "later-width", "reordered-width", "type-width", "missing-type", "missing-data", "duplicate-type", "duplicate-data", "extra", "mandatory-extra"} {
				t.Run(fmt.Sprintf("%s/%d/%s", kind, index, fault), func(t *testing.T) {
					raw := subscriptionDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13310, 23310)), index, []byte{1, 0, 2})
					raw = subscriptionDataFault(t, raw, index, fault)
					_, cause := co.Decode(raw)
					var proof interface{ DiameterSubscriptionDataValueV1() string }
					want := fault == "plain" || fault == "type" || fault == "sequence" || fault == "later-width"
					if errors.As(cause, &proof) != want {
						t.Fatal("order", cause)
					}
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					expected := uint32(0)
					if want {
						expected = 5004
					}
					if fault == "type-width" || fault == "earlier-width" || fault == "reordered-width" {
						expected = 5014
					}
					if fault == "mandatory-extra" {
						expected = 5001
					}
					if ok != (expected != 0) || code != expected || want && !bytes.Equal(out, subscriptionDataQuote(t, []byte{1, 0, 2})) {
						t.Fatal(cause, code, out)
					}
					if fault == "extra" && (!errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported)) {
						t.Fatal("mixed", cause)
					}
				})
			}
		}
	}
}
func TestSubscriptionDataGroupOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, move := range []bool{false, true} {
			for _, mixed := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%t/%t", kind, move, mixed), func(t *testing.T) {
					raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13312, 23312))
					raw = subscriptionDataRaw(t, raw, 0, []byte{1, 0, 2})
					raw = subscriptionDataRaw(t, raw, 1, []byte{0})
					if mixed {
						raw = subscriptionDataFault(t, raw, 0, "extra")
					}
					if move {
						var part []byte
						raw = subscriptionGroupEdit(t, raw, 0, func(b []byte) []byte { part = b; return nil })
						raw = append(raw, part...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					_, cause := co.Decode(raw)
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					if mixed && !move {
						if ok || code != 0 || out != nil || !errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrUnsupported) {
							t.Fatal("mixed first", cause)
						}
						return
					}
					data := []byte{1, 0, 2}
					if move {
						data = []byte{0}
					}
					if !ok || code != 5004 || !bytes.Equal(out, subscriptionDataQuote(t, data)) {
						t.Fatal("first instance", cause, code, out)
					}
				})
			}
		}
	}
}
func TestRejectSubscriptionDataClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, index := range []int{0, 1} {
			for _, fault := range []string{"default", "extra", "earlier-extra"} {
				t.Run(fmt.Sprintf("%d/%s", index, fault), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw := subscriptionDataRaw(t, reqBytes(t, co, combinedDuplicateRequest(t, n, co, "iec", 13311, 23311)), index, []byte{0})
					if fault == "extra" {
						raw = subscriptionDataFault(t, raw, index, "extra")
					}
					if fault == "earlier-extra" {
						raw = subscriptionDataFault(t, raw, 0, "extra")
					}
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
