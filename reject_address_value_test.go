// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Address-Type value acceptance.
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

var addressValueCases = []struct {
	kind  string
	group uint32
}{{"iec", 886}, {"iec", 1201}, {"initial", 886}, {"initial", 1201}, {"term", 886}, {"term", 1201}, {"general", 886}, {"balance", 886}}

func addressValuePath(group, leaf uint32) []missingSelector {
	return append(recipientPath(group), missingSelector{key: avp.Key{VendorID: 10415, Code: leaf}})
}
func addressValueData(value int32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(value))
	return b
}
func addressValueQuote(t *testing.T, group uint32, value int32) []byte {
	part := func(c uint32, b []byte) []byte { return invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, b) }
	return part(873, part(877, part(group, part(899, addressValueData(value)))))
}
func TestRejectAddressValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, sample := range addressValueCases {
			kind, group := sample.kind, sample.group
			{
				for _, value := range []int32{-2147483648, -1, 0, 2, 2147483647} {
					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%d/emergency-%t", kind, group, value, emergency), func(t *testing.T) {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12260, 22600))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12601, 22601)
							data := addressValueData(value)
							raw, _ := invalidRaw(t, co, req, addressValuePath(group, 899), data, avp.Key{})
							want := addressValueQuote(t, group, value)
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12602, 22602)
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

func addressMoveLast(t *testing.T, raw []byte, group uint32) []byte {
	t.Helper()
	area := raw[20:]
	var part []byte
	for _, key := range []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 877}, {VendorID: 10415, Code: group}} {
		d := wire.NewDecoder(area, wire.DefaultLimits())
		found := false
		for {
			v, ok, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !ok {
				break
			}
			if v.Key == key {
				area = v.Data
				part = invalidPart(t, v.Key, v.Flags, v.Data)
				found = true
				break
			}
		}
		if !found {
			t.Fatal("missing reorder group")
		}
	}
	raw = mmsGroupOmit(t, raw, recipientPath(group))
	return unitAppendRaw(t, raw, recipientPath(group)[:2], part)
}
func addressFault(t *testing.T, raw []byte, kind string, group uint32, fault string) []byte {
	t.Helper()
	edit := func(path []missingSelector, b []byte) {
		body, _ := invalidEdit(t, raw[20:], path, b, avp.Key{})
		raw = append(raw[:20:20], body...)
	}
	switch fault {
	case "sequence", "subscriptions", "sms", "timestamp", "interface":
		return unitValueFault(t, raw, kind, fault)
	case "nul":
		edit(addressValuePath(group, 897), []byte{'a', 0, 'b'})
	case "extra", "mandatory-extra":
		flags := byte(0)
		if fault == "mandatory-extra" {
			flags = 64
		}
		return unitAppendRaw(t, raw, recipientPath(group), invalidPart(t, avp.Key{Code: 9999}, flags, []byte{1}))
	case "missing-data":
		return mmsGroupOmit(t, raw, addressValuePath(group, 897))
	case "duplicate-type":
		return unitAppendRaw(t, raw, recipientPath(group), invalidPart(t, avp.Key{VendorID: 10415, Code: 899}, 192, addressValueData(2)))
	case "type-width":
		edit(addressValuePath(group, 899), []byte{1})
	case "earlier-width":
		edit(smsValuePath(true), []byte{1})
	case "later-width", "reordered-width":
		edit(recipientPath(1202), []byte{1})
		if fault == "reordered-width" {
			wire.Put24(raw[1:4], uint32(len(raw)))
			return addressMoveLast(t, raw, group)
		}
	}
	wire.Put24(raw[1:4], uint32(len(raw)))
	return raw
}
func TestAddressValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, group := range []uint32{886, 1201} {
			for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "timestamp", "interface", "nul", "extra", "mandatory-extra", "missing-data", "duplicate-type", "type-width", "earlier-width", "later-width", "reordered-width"} {
				t.Run(fmt.Sprintf("%s/%d/%s", kind, group, fault), func(t *testing.T) {
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12610, 22610), addressValuePath(group, 899), addressValueData(-1), avp.Key{})
					raw = addressFault(t, raw, kind, group, fault)
					_, cause := co.Decode(raw)
					if cause == nil {
						t.Fatal("accepted")
					}
					out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
					expected := uint32(5004)
					switch fault {
					case "nul", "extra", "missing-data", "duplicate-type":
						expected = 0
					case "mandatory-extra":
						expected = 5001
					case "type-width", "earlier-width", "reordered-width":
						expected = 5014
					}
					if ok != (expected != 0) || code != expected {
						t.Fatal(cause, code, out)
					}
					var proof interface{ DiameterAddressTypeValueV1() string }
					if errors.As(cause, &proof) != (expected == 5004) {
						t.Fatal("precedence", cause)
					}
					if expected == 5004 && !bytes.Equal(out, addressValueQuote(t, group, -1)) {
						t.Fatal("wrong instance")
					}
					if fault == "nul" || fault == "extra" {
						if !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrField) {
							t.Fatal("mixed", cause)
						}
					}
					if fault == "missing-data" && !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					if fault == "duplicate-type" && !errors.Is(cause, validation.ErrDuplicate) {
						t.Fatal(cause)
					}
				})
			}
		}
	}
}
func TestAddressValueGroupOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, move := range []bool{false, true} {
			for _, mixed := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%t/%t", kind, move, mixed), func(t *testing.T) {
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12612, 22612), addressValuePath(886, 899), addressValueData(-1), avp.Key{})
					body, _ := invalidEdit(t, raw[20:], addressValuePath(1201, 899), addressValueData(2), avp.Key{})
					raw = append(raw[:20:20], body...)
					wire.Put24(raw[1:4], uint32(len(raw)))
					if mixed {
						raw = addressFault(t, raw, kind, 886, "nul")
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
					group, v := uint32(886), int32(-1)
					if move {
						group, v = 1201, 2
					}
					if !ok || code != 5004 || !bytes.Equal(out, addressValueQuote(t, group, v)) {
						t.Fatal("group order", cause, code, out)
					}
				})
			}
		}
	}
}
func TestRejectAddressValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, sample := range addressValueCases {
			for _, fault := range []string{"default", "nul", "extra"} {
				t.Run(fmt.Sprintf("%s/%d/%s", sample.kind, sample.group, fault), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, sample.kind, 12611, 22611), addressValuePath(sample.group, 899), addressValueData(-1), avp.Key{})
					raw = addressFault(t, raw, sample.kind, sample.group, fault)
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
