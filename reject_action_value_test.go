// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded conditional EVENT Requested-Action value diagnostics.
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

func TestActionValueWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, value := range []uint32{0, 1, 2, 3, ^uint32(0)} {
			t.Run(fmt.Sprintf("%s/%d", kind, value), func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 11810, 21810)
				var data [4]byte
				binary.BigEndian.PutUint32(data[:], value)
				var raw, want []byte
				event := kind != "initial" && kind != "term"
				if event {
					raw, want = invalidRaw(t, co, req, actionPath(), data[:], avp.Key{})
				} else {
					raw = reqBytes(t, co, req)
					want = invalidPart(t, avp.Key{Code: 436}, 64, data[:])
					raw = append(raw, want...)
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				_, cause := co.Decode(raw)
				q, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 12)
				proof := !event || kind == "iec" && value != 0 || value > 2
				expected := uint32(5004)
				if !event {
					expected = 5008
				}
				if ok != proof || ok && (code != expected || !bytes.Equal(q, want)) {
					t.Fatal(cause, code, q)
				}
				if proof {
					if !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrField) != event || errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					reply, e := n.rejectFrame(raw, 4096, rejectDecode, cause)
					if e != nil || len(reply) == 0 {
						t.Fatal(e)
					}
					if _, e := n.rejectFrame(raw, len(reply), rejectDecode, cause); e != nil {
						t.Fatal("exact answer budget", e)
					}
					if _, e := n.rejectFrame(raw, len(reply)-1, rejectDecode, cause); e == nil {
						t.Fatal("short answer budget")
					}
					req.Body.(*combined.CCR).RequestedAction = avp.Some(value)
					_, sendError := co.Snapshot(req)
					if !errors.Is(sendError, validation.ErrUnsupported) || errors.Is(sendError, validation.ErrField) {
						t.Fatal("send classification changed", sendError)
					}
				} else if cause != nil {
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal("unreviewed acquired reply", e)
					}
				}
			})
		}
	}
}
func TestActionValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "general", "balance", "refund"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "later-unit", "later-interface", "length", "duplicate", "trailing", "sms", "mms", "timestamp"} {
			if kind != "iec" && (fault == "sms" || fault == "mms" || fault == "timestamp") {
				continue
			}
			t.Run(kind+"/"+fault, func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 11811, 21811)
				raw, want := invalidRaw(t, co, req, actionPath(), []byte{0, 0, 0, 3}, avp.Key{})
				edit := func(path []missingSelector, data []byte) {
					body, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
					raw = append(raw[:20:20], body...)
					wire.Put24(raw[1:4], uint32(len(raw)))
				}
				expected := uint32(5004)
				switch fault {
				case "sequence":
					edit(lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2})
					expected = 0
				case "subscriptions":
					raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
					expected = 0
				case "later-unit":
					raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}))
				case "later-interface":
					edit(originatorInterfacePath(), []byte{0, 0, 0, 99})
				case "length":
					edit(actionPath(), []byte{1})
					expected = 5014
				case "duplicate":
					raw = append(raw, invalidPart(t, avp.Key{Code: 436}, 64, []byte{0, 0, 0, 99})...)
					wire.Put24(raw[1:4], uint32(len(raw)))
					expected = 0
				case "trailing":
					raw = append(raw, 1)
					wire.Put24(raw[1:4], uint32(len(raw)))
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
}
func TestRejectActionValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "general", "balance", "refund"} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/emergency-%t", kind, emergency), func(t *testing.T) {
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
						combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11800, 21800))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedDuplicateRequest(t, n, co, kind, 11801, 21801)
					raw, want := invalidRaw(t, co, req, actionPath(), []byte{0, 0, 0, 3}, avp.Key{})
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
					field.Path[0] = avp.Key{Code: 436}
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
					follow := combinedDuplicateRequest(t, n, co, kind, 11802, 21802)
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
