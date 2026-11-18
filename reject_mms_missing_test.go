// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Selected isolated combined MMS omissions and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"math/bits"
	"testing"
)

func mmsMinimum(t *testing.T, code uint32) []byte {
	width := 4
	if code == 1210 {
		width = 0
	}
	part := func(c uint32, data []byte) []byte {
		return invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, data)
	}
	return part(873, part(877, part(code, make([]byte, width))))
}
func TestMMSMissingCombinations(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	codes := []uint32{1201, 1202, 1209, 1212, 1210, 1216}
	for _, kind := range []string{"initial", "term", "iec"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11210, 21210)
		original := reqBytes(t, co, req)
		for mask := uint8(0); mask < 64; mask++ {
			raw := append([]byte(nil), original...)
			var last uint32
			for i, code := range codes {
				if mask&(1<<i) != 0 {
					raw = recipientOmit(t, raw, code)
					last = code
				}
			}
			_, cause := co.Decode(raw)
			if mask == 0 {
				if cause != nil {
					t.Fatal(kind, cause)
				}
				continue
			}
			if !errors.Is(cause, validation.ErrMissing) {
				t.Fatal(kind, mask, cause)
			}
			q, ok := missingavp.Quote(cause, n.cfg.Limits, 128)
			if ok != (bits.OnesCount8(mask) == 1) {
				t.Fatal(kind, mask, cause, q)
			}
			if ok {
				want := mmsMinimum(t, last)
				if last == 1201 {
					want = recipientMinimum(t)
				}
				if !bytes.Equal(q, want) {
					t.Fatal(kind, mask, q)
				}
			} else {
				if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
					t.Fatal("multiple omission", mask, e)
				}
			}
		}
	}
}
func TestMMSMissingBounds(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{1202, 1209, 1212, 1210, 1216} {
		req := combinedDuplicateRequest(t, n, co, "initial", 11211, 21211)
		raw := missingRaw(t, co, req, recipientPath(code), false)
		want := mmsMinimum(t, code)
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
func TestMMSMissingValueBoundaries(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{1202, 1209, 1212, 1210, 1216} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			req := combinedDuplicateRequest(t, n, co, "initial", 11212, 21212)
			value := []byte{0, 0, 0}
			if code == 1210 {
				value = []byte{0}
			}
			raw, _ := invalidRaw(t, co, req, recipientPath(code), value, avp.Key{})
			_, cause := co.Decode(raw)
			wantError := avp.ErrValue
			if code == 1210 {
				wantError = validation.ErrField
			}
			if !errors.Is(cause, wantError) {
				t.Fatal("present invalid", code, cause)
			}
			if _, ok := missingavp.Quote(cause, n.cfg.Limits, 40); ok {
				t.Fatal("invalid became absent")
			}
			if code == 1210 {
				raw, _ = invalidRaw(t, co, req, recipientPath(code), nil, avp.Key{})
				if _, e := co.Decode(raw); e != nil {
					t.Fatal("empty SMSKey changed", e)
				}
			}
		})
	}
	req := combinedDuplicateRequest(t, n, co, "initial", 11213, 21213)
	raw := missingRaw(t, co, req, lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: 2018}), false)
	raw = recipientOmit(t, raw, 1202)
	_, cause := co.Decode(raw)
	if _, ok := missingavp.Quote(cause, n.cfg.Limits, 40); ok {
		t.Fatal("earlier SMS failure lost precedence")
	}
}
func TestRejectMMSMissingDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{1202, 1209, 1212, 1210, 1216} {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "initial", 11214, 21214)
				sendRejectRaw(t, remote, missingRaw(t, co, req, recipientPath(code), false))
				waitFor(t, func() bool { return !p.Info().Ready })
				if raw, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("default reply %x", raw)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRejectMMSMissing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{1202, 1209, 1212, 1210, 1216} {
			for _, kind := range []string{"initial", "term", "iec"} {
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
							combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11200, 21200))
							var e error
							held, e = p.Receive(ctxTest(t))
							if e != nil {
								t.Fatal(e)
							}
						}
						before := n.Resources()
						req := combinedDuplicateRequest(t, n, co, kind, 11201, 21201)
						raw := missingRaw(t, co, req, recipientPath(code), false)
						_, cause := co.Decode(raw)
						if !errors.Is(cause, validation.ErrMissing) {
							t.Fatal(cause)
						}
						want := mmsMinimum(t, code)
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
						follow := combinedDuplicateRequest(t, n, co, kind, 11202, 21202)
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
