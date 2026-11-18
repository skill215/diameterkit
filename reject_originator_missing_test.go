// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated IEC Originator-SCCP omission under a reviewed mixed condition.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func originatorInterfacePath() []missingSelector {
	return append(smsPath(2009), missingSelector{avp.Key{VendorID: 10415, Code: 2006}, 0})
}
func TestOriginatorMissingWireMatrix(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, typ := range []int32{-1, 0, 1, 2, 3, 99} {
		for _, present := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", typ, present), func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, "iec", 11410, 21410)
				var value [4]byte
				binary.BigEndian.PutUint32(value[:], uint32(typ))
				raw, _ := invalidRaw(t, co, req, originatorInterfacePath(), value[:], avp.Key{})
				if !present {
					raw = smsOmit(t, raw, 2008)
				}
				_, cause := co.Decode(raw)
				valid := present && typ == 1
				proof := !present && typ == 1
				if (cause == nil) != valid || (!valid && !errors.Is(cause, validation.ErrUnsupported)) || errors.Is(cause, validation.ErrMissing) != proof {
					t.Fatal(cause)
				}
				q, ok := missingavp.Quote(cause, n.cfg.Limits, 36)
				if ok != proof || ok && !bytes.Equal(q, smsMinimum(t, 2008)) {
					t.Fatal(q, cause)
				}
				if !valid && !proof {
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
			})
		}
	}
}
func TestOriginatorMissingPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"common", "mms", "action", "timestamp", "empty", "nul"} {
		req := combinedDuplicateRequest(t, n, co, "iec", 11411, 21411)
		raw := reqBytes(t, co, req)
		var want []byte
		switch kind {
		case "common":
			raw = smsOmit(t, raw, 2018)
			raw = smsOmit(t, raw, 2008)
			want = smsMinimum(t, 2018)
		case "mms":
			raw = recipientOmit(t, raw, 1202)
			raw = smsOmit(t, raw, 2008)
			want = mmsMinimum(t, 1202)
		case "action":
			raw, _ = invalidRaw(t, co, req, lengthPath(avp.Key{Code: 436}), []byte{0, 0, 0, 1}, avp.Key{})
			raw = smsOmit(t, raw, 2008)
		case "timestamp":
			raw, _ = invalidRaw(t, co, req, recipientPath(1202), []byte{0, 0, 0, 0}, avp.Key{})
			raw = smsOmit(t, raw, 2008)
		case "empty":
			raw, _ = invalidRaw(t, co, req, smsPath(2008), nil, avp.Key{})
		case "nul":
			raw, _ = invalidRaw(t, co, req, smsPath(2008), []byte{0}, avp.Key{})
		}
		_, cause := co.Decode(raw)
		q, ok := missingavp.Quote(cause, n.cfg.Limits, 40)
		if ok != (want != nil) || ok && !bytes.Equal(q, want) {
			t.Fatal(kind, cause, q)
		}
		switch kind {
		case "action":
			if !errors.Is(cause, validation.ErrUnsupported) || errors.Is(cause, validation.ErrMissing) {
				t.Fatal(cause)
			}
		case "timestamp", "nul":
			if !errors.Is(cause, validation.ErrField) {
				t.Fatal(cause)
			}
		case "empty":
			if cause != nil {
				t.Fatal("present empty", cause)
			}
		}
	}
}
func TestRejectOriginatorMissingMixedConditions(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"interface-only", "both", "action"} {
			t.Run(kind, func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "iec", 11412, 21412)
				path := originatorInterfacePath()
				if kind == "action" {
					path = lengthPath(avp.Key{Code: 436})
				}
				raw, want := invalidRaw(t, co, req, path, []byte{0, 0, 0, 2}, avp.Key{})
				if kind != "interface-only" {
					raw = smsOmit(t, raw, 2008)
				}
				sendRejectRaw(t, remote, raw)
				// Action validation precedes the missing Originator condition.
				// Batch118 maps only that earlier selected value failure.
				if kind == "action" {
					answer := readReject(t, remote, req.Header, 5004)
					if !answer.FailedAVP.Present || !bytes.Equal(answer.FailedAVP.Value.Data, want) {
						t.Fatal("first failure")
					}
					noDuplicateDelivery(t, p)
					closeDuplicateNode(t, n)
					return
				}
				waitFor(t, func() bool { return !p.Info().Ready })
				if b, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("ambiguous failure replied %x", b)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestOriginatorMissingBounds(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{2008} {
		req := combinedDuplicateRequest(t, n, co, "iec", 11411, 21411)
		raw := missingRaw(t, co, req, smsPath(code), false)
		want := smsMinimum(t, code)
		for _, negative := range []string{"exact", "bytes", "field", "frame", "avps", "depth", "outer", "middle", "leaf", "bare"} {
			t.Run(fmt.Sprintf("%d/%s", code, negative), func(t *testing.T) {
				_, cause := co.Decode(raw)
				var f *validation.FieldError
				if !errors.As(cause, &f) {
					t.Fatal(cause)
				}
				l := n.cfg.Limits
				l.MaxFrameBytes = len(want) + 28
				l.MaxFieldBytes = len(want)
				l.MaxAVPs = 4
				l.MaxRepeated = 1
				l.MaxDepth = 3
				budget := len(want)
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
				case "outer":
					f.Path[0].Code = 263
				case "middle":
					f.Path[1].Code++
				case "leaf":
					f.Path[2].Code++
				case "bare":
					cause = &validation.FieldError{Path: f.Path, Cause: validation.ErrMissing}
				}
				q, ok := missingavp.Quote(cause, l, budget)
				if ok != (negative == "exact") || ok && !bytes.Equal(q, want) {
					t.Fatal(negative, q, cause)
				}
				if negative == "outer" {
					if _, ok := missingDiagnostic(cause); ok {
						t.Fatal("downgrade")
					}
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
			})
		}
	}
}

func TestRejectOriginatorMissingDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{2008} {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "iec", 11414, 21414)
				sendRejectRaw(t, remote, missingRaw(t, co, req, smsPath(code), false))
				waitFor(t, func() bool { return !p.Info().Ready })
				if raw, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("default reply %x", raw)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRejectOriginatorMissing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{2008} {
			for _, kind := range []string{"iec"} {
				for _, emergency := range []bool{false, true} {
					t.Run(fmt.Sprintf("%d/%s/emergency-%t", code, kind, emergency), func(t *testing.T) {
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
							combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11400, 21400))
							var e error
							held, e = p.Receive(ctxTest(t))
							if e != nil {
								t.Fatal(e)
							}
						}
						before := n.Resources()
						req := combinedDuplicateRequest(t, n, co, kind, 11401, 21401)
						raw := missingRaw(t, co, req, smsPath(code), false)
						_, cause := co.Decode(raw)
						if !errors.Is(cause, validation.ErrMissing) || !errors.Is(cause, validation.ErrUnsupported) {
							t.Fatal(cause)
						}
						want := smsMinimum(t, code)
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
						field.Path[0] = avp.Key{VendorID: 10415, Code: 873}
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
						follow := combinedDuplicateRequest(t, n, co, kind, 11402, 21402)
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
	})
}
