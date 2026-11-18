// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional EVENT Requested-Action omission and bounded acceptance.
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
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func actionPath() []missingSelector { return lengthPath(avp.Key{Code: 436}) }
func actionMinimum(t *testing.T) []byte {
	return invalidPart(t, avp.Key{Code: 436}, 64, []byte{0, 0, 0, 0})
}
func TestActionMissingWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, value := range []uint32{0, 1, 2, 3, ^uint32(0)} {
			for _, present := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/%t", kind, value, present), func(t *testing.T) {
					req := combinedDuplicateRequest(t, n, co, kind, 11610, 21610)
					raw := reqBytes(t, co, req)
					event := kind != "initial" && kind != "term"
					var data [4]byte
					binary.BigEndian.PutUint32(data[:], value)
					if event {
						if present {
							raw, _ = invalidRaw(t, co, req, actionPath(), data[:], avp.Key{})
						} else {
							raw = missingRaw(t, co, req, actionPath(), false)
						}
					} else if present {
						raw = append(raw, invalidPart(t, avp.Key{Code: 436}, 64, data[:])...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
					_, cause := co.Decode(raw)
					proof := event && !present
					valid := !event && !present || present && (kind == "iec" && value == 0 || (kind == "general" || kind == "balance") && (value == 0 || value == 2) || kind == "refund" && value == 1)
					if (cause == nil) != valid || errors.Is(cause, validation.ErrMissing) != proof {
						t.Fatal(cause)
					}
					q, ok := missingavp.Quote(cause, n.cfg.Limits, 12)
					if ok != proof || ok && !bytes.Equal(q, actionMinimum(t)) {
						t.Fatal(cause, q)
					}
					if proof && !errors.Is(cause, validation.ErrUnsupported) {
						t.Fatal(cause)
					}
					if !valid && !proof && !(present && event && (kind == "iec" && value != 0 || value > 2)) {
						if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
							t.Fatal("present value acquired reply", e)
						}
					}
				})
			}
		}
	}
}
func TestActionMissingBounds(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "general", "balance", "refund"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11611, 21611)
		raw := missingRaw(t, co, req, actionPath(), false)
		for _, negative := range []string{"exact", "bytes", "field", "frame", "avps", "depth", "path", "bare"} {
			t.Run(kind+"/"+negative, func(t *testing.T) {
				_, cause := co.Decode(raw)
				var f *validation.FieldError
				if !errors.As(cause, &f) {
					t.Fatal(cause)
				}
				l := n.cfg.Limits
				l.MaxFrameBytes = 40
				l.MaxFieldBytes = 12
				l.MaxAVPs = 2
				l.MaxRepeated = 1
				l.MaxDepth = 1
				budget := 12
				switch negative {
				case "bytes":
					budget--
				case "field":
					l.MaxFieldBytes--
				case "frame":
					l.MaxFrameBytes--
				case "avps":
					l.MaxAVPs--
				case "depth":
					l.MaxDepth--
				case "path":
					f.Path[0].Code = 263
				case "bare":
					cause = &validation.FieldError{Path: f.Path, Cause: validation.ErrMissing}
				}
				q, ok := missingavp.Quote(cause, l, budget)
				if ok != (negative == "exact") || ok && !bytes.Equal(q, actionMinimum(t)) {
					t.Fatal(cause, q)
				}
				if negative == "path" {
					if _, ok := missingDiagnostic(cause); ok {
						t.Fatal("downgrade")
					}
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
			})
		}
		req.Body.(*combined.CCR).RequestedAction = avp.Optional[uint32]{}
		_, cause := co.Snapshot(req)
		if !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrMissing) {
			t.Fatal("send classification", cause)
		}
		if _, ok := missingavp.Quote(cause, n.cfg.Limits, 12); ok {
			t.Fatal("send evidence")
		}
	}
}
func TestActionMissingPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "general", "balance", "refund"} {
		for _, fault := range []string{"plain", "sequence", "subscriptions", "later-unit", "later-interface", "length", "sms", "mms", "timestamp"} {
			if kind != "iec" && (fault == "sms" || fault == "mms" || fault == "timestamp") {
				continue
			}
			t.Run(kind+"/"+fault, func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 11612, 21612)
				raw := missingRaw(t, co, req, actionPath(), false)
				want := actionMinimum(t)
				switch fault {
				case "sequence":
					raw, _ = invalidRaw(t, co, req, lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2}, avp.Key{})
					raw = mmsGroupOmit(t, raw, actionPath())
					want = nil
				case "subscriptions":
					raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
					want = nil
				case "later-unit":
					raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}))
				case "later-interface":
					raw, _ = invalidRaw(t, co, req, originatorInterfacePath(), []byte{0, 0, 0, 99}, avp.Key{})
					raw = mmsGroupOmit(t, raw, actionPath())
				case "length":
					raw, _ = invalidRaw(t, co, req, actionPath(), []byte{0}, avp.Key{})
					want = nil
				case "sms":
					raw = smsOmit(t, raw, 2018)
					want = smsMinimum(t, 2018)
				case "mms":
					raw = mmsGroupOmit(t, raw, mmsGroupPath())
					want = mmsGroupMinimum(t)
				case "timestamp":
					raw, _ = invalidRaw(t, co, req, recipientPath(1202), []byte{0, 0, 0, 0}, avp.Key{})
					raw = mmsGroupOmit(t, raw, actionPath())
					want = nil
				}
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("invalid accepted")
				}
				q, ok := missingavp.Quote(cause, n.cfg.Limits, 180)
				if ok != (want != nil) || ok && !bytes.Equal(q, want) {
					t.Fatal(cause, q)
				}
				if fault == "length" {
					if _, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 180); !ok || code != 5014 {
						t.Fatal("length order", cause, code)
					}
				}
			})
		}
	}
}
func TestRejectActionMissingClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "general", "balance", "refund"} {
			for _, invalid := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/invalid-%t", kind, invalid), func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure

					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := combinedDuplicateRequest(t, n, co, kind, 11614, 21614)
					raw := missingRaw(t, co, req, actionPath(), false)
					if invalid {
						raw, _ = invalidRaw(t, co, req, actionPath(), []byte{0, 0, 0, 3}, avp.Key{})
					}
					sendRejectRaw(t, remote, raw)
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("closed case replied %x", b)
					}
					closeDuplicateNode(t, n)
				})
			}
		}
	})
}

func TestRejectActionMissing(t *testing.T) {
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
						combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11600, 21600))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedDuplicateRequest(t, n, co, kind, 11601, 21601)
					raw := missingRaw(t, co, req, actionPath(), false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) || !errors.Is(cause, validation.ErrUnsupported) {
						t.Fatal(cause)
					}
					want := actionMinimum(t)
					q, ok := missingavp.Quote(cause, cfg.Limits, len(want))
					if !ok || !bytes.Equal(q, want) {
						t.Fatal(q, cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, len(want)-1); ok {
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
					ans := readReject(t, remote, req.Header, 5005)
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
					follow := combinedDuplicateRequest(t, n, co, kind, 11602, 21602)
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
