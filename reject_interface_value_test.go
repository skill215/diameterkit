// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated IEC Originator-Interface value diagnostics.
// SPDX-AI-Date: 2026-11-18

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

func interfaceValueQuote(t *testing.T, value int32) []byte {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(value))
	b := invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, data[:])
	for _, code := range []uint32{2009, 2000, 873} {
		b = invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, b)
	}
	return b
}
func TestInterfaceValueWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term", "general", "balance", "refund"} {
		for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
			presence := []bool{false}
			if kind == "iec" {
				presence = []bool{false, true}
			}
			for _, present := range presence {
				t.Run(fmt.Sprintf("%s/%d/%t", kind, value, present), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 11910, 21910)
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], uint32(value))
					raw, _ := invalidRaw(t, co, req, originatorInterfacePath(), data[:], avp.Key{})
					if kind == "iec" && !present {
						raw = smsOmit(t, raw, 2008)
					}
					_, cause := co.Decode(raw)
					q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 52)
					proof := kind == "iec" && present && value != 1
					if ok != proof || ok && (code != 5004 || !bytes.Equal(q, interfaceValueQuote(t, value))) {
						t.Fatal(cause, code, q)
					}
					if !proof {
						return
					}
					if !errors.Is(cause, validation.ErrUnsupported) || !errors.Is(cause, validation.ErrField) || errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					var field *validation.FieldError
					if !errors.As(cause, &field) || len(field.Path) != 4 {
						t.Fatal(cause)
					}
					for i, code := range []uint32{873, 2000, 2009, 2006} {
						if field.Path[i] != (avp.Key{VendorID: 10415, Code: code}) {
							t.Fatal(field.Path)
						}
					}
					req.Body.(*combined.CCR).ServiceInformation.SMSInformation.OriginatorInterface.InterfaceType = value
					_, sendErr := co.Snapshot(req)
					if !errors.Is(sendErr, validation.ErrUnsupported) || errors.Is(sendErr, validation.ErrField) {
						t.Fatal("send changed", sendErr)
					}
				})
			}
		}
	}
}
func TestInterfaceValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, fault := range []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "action", "address-nul", "both", "length", "duplicate-leaf", "duplicate-parent", "trailing", "later-unit"} {
		t.Run(fault, func(t *testing.T) {
			req := combinedDuplicateRequest(t, n, co, "iec", 11911, 21911)
			raw, _ := invalidRaw(t, co, req, originatorInterfacePath(), []byte{0, 0, 0, 2}, avp.Key{})
			want := interfaceValueQuote(t, 2)
			expected := uint32(5004)
			edit := func(path []missingSelector, data []byte) {
				b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
				raw = append(raw[:20:20], b...)
				wire.Put24(raw[1:4], uint32(len(raw)))
			}
			switch fault {
			case "sequence":
				edit(lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2})
				expected = 0
			case "subscriptions":
				raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
				expected = 0
			case "sms":
				raw = smsOmit(t, raw, 2018)
				expected = 0
			case "mms":
				raw = mmsGroupOmit(t, raw, mmsGroupPath())
				expected = 0
			case "timestamp":
				edit(recipientPath(1202), []byte{0, 0, 0, 0})
				expected = 0
			case "action":
				edit(actionPath(), []byte{0, 0, 0, 3})
				want = invalidPart(t, avp.Key{Code: 436}, 64, []byte{0, 0, 0, 3})
			case "address-nul":
				edit(smsPath(2008), []byte{0})
				expected = 0
			case "both":
				raw = smsOmit(t, raw, 2008)
				expected = 0
			case "length":
				edit(originatorInterfacePath(), []byte{1})
				expected = 5014
			case "duplicate-leaf":
				edit(smsPath(2009), append(invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, []byte{0, 0, 0, 2}), invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, []byte{0, 0, 0, 99})...))
				expected = 0
			case "duplicate-parent":
				// Replace the whole SMS group payload with itself plus another interface.
				d := wire.NewDecoder(raw[20:], n.cfg.Limits)
				var service []byte
				for {
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					if v.Key == (avp.Key{VendorID: 10415, Code: 873}) {
						service = v.Data
					}
				}
				d = wire.NewDecoder(service, n.cfg.Limits)
				var sms []byte
				for {
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					if v.Key == (avp.Key{VendorID: 10415, Code: 2000}) {
						sms = v.Data
					}
				}
				extra := invalidPart(t, avp.Key{VendorID: 10415, Code: 2009}, 192, invalidPart(t, avp.Key{VendorID: 10415, Code: 2006}, 192, []byte{0, 0, 0, 99}))
				edit(lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}), append(append([]byte{}, sms...), extra...))
				expected = 0
			case "trailing":
				raw = append(raw, 1)
				wire.Put24(raw[1:4], uint32(len(raw)))
				expected = 0
			case "later-unit":
				raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}))
			}
			_, cause := co.Decode(raw)
			if cause == nil {
				t.Fatal("invalid accepted")
			}
			q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
			if ok != (expected != 0) || code != expected || code == 5004 && !bytes.Equal(q, want) {
				t.Fatal(cause, code, q)
			}
		})
	}
}
func TestRejectInterfaceValueClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"default", "mixed", "nonIEC"} {
			for _, value := range []int32{-1, 2} {
				t.Run(fmt.Sprintf("%s/%d", kind, value), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					if kind != "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					template := "iec"
					if kind == "nonIEC" {
						template = "initial"
					}
					req := combinedDuplicateRequest(t, n, co, template, 11912, 21912)
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], uint32(value))
					raw, _ := invalidRaw(t, co, req, originatorInterfacePath(), data[:], avp.Key{})
					if kind == "mixed" {
						raw = smsOmit(t, raw, 2008)
					}
					sendRejectRaw(t, remote, raw)
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unreviewed/default reply %x", b)
					}
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}
func TestRejectInterfaceValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, value := range []int32{-1, 0, 2} {
			kind := "iec"
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/emergency-%t", value, emergency), func(t *testing.T) {
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
						combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11900, 21900))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedDuplicateRequest(t, n, co, kind, 11901, 21901)
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], uint32(value))
					raw, _ := invalidRaw(t, co, req, originatorInterfacePath(), data[:], avp.Key{})
					want := interfaceValueQuote(t, value)
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
					follow := combinedDuplicateRequest(t, n, co, kind, 11902, 21902)
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
	})
}
