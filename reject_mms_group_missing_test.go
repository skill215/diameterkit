// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional whole MMS group omission and bounded diagnostic acceptance.
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
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func mmsGroupPath() []missingSelector { return recipientPath(1201)[:2] }
func mmsGroupMinimum(t *testing.T) []byte {
	t.Helper()
	out := make([]byte, 180)
	w := wire.NewEncoder(out, wire.DefaultLimits())
	key := func(code uint32) avp.Key { return avp.Key{VendorID: 10415, Code: code} }
	service := w.BeginGroup(key(873), 192)
	mms := w.BeginGroup(key(877), 192)
	originator := w.BeginGroup(key(886), 192)
	w.Uint32(key(899), 192, 0)
	w.Bytes(key(897), 192, nil)
	w.EndGroup(originator)
	recipient := w.BeginGroup(key(1201), 192)
	w.Uint32(key(899), 192, 0)
	w.Bytes(key(897), 192, nil)
	w.EndGroup(recipient)
	w.Uint32(key(1202), 192, 0)
	w.Uint32(key(1209), 192, 0)
	w.Uint32(key(1212), 192, 0)
	w.Bytes(key(1210), 192, nil)
	w.Uint32(key(1216), 192, 0)
	w.EndGroup(mms)
	w.EndGroup(service)
	if w.Err() != nil || w.Len() != len(out) {
		t.Fatal(w.Err(), w.Len())
	}
	return out
}
func TestMMSGroupMissingBounds(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	want := mmsGroupMinimum(t)
	for _, kind := range []string{"initial", "term", "iec"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11510, 21510)
		raw := missingRaw(t, co, req, mmsGroupPath(), false)
		for _, negative := range []string{"exact", "bytes", "field", "frame", "avps", "depth", "outer", "group", "bare"} {
			t.Run(kind+"/"+negative, func(t *testing.T) {
				_, cause := co.Decode(raw)
				var f *validation.FieldError
				if !errors.As(cause, &f) {
					t.Fatal(cause)
				}
				l := n.cfg.Limits
				l.MaxFrameBytes = 208
				l.MaxFieldBytes = 180
				l.MaxAVPs = 14
				l.MaxRepeated = 1
				l.MaxDepth = 4
				budget := 180
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
				case "group":
					f.Path[1].Code++
				case "bare":
					cause = &validation.FieldError{Path: f.Path, Cause: validation.ErrMissing}
				}
				q, ok := missingavp.Quote(cause, l, budget)
				if ok != (negative == "exact") || ok && !bytes.Equal(q, want) {
					t.Fatal(cause, q)
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
		req.Body.(*combined.CCR).ServiceInformation.MMSInformation = avp.Optional[combined.MMSInformation]{}
		_, cause := co.Snapshot(req)
		if !errors.Is(cause, validation.ErrMissing) {
			t.Fatal(cause)
		}
		if _, ok := missingavp.Quote(cause, n.cfg.Limits, 180); ok {
			t.Fatal("send evidence")
		}
	}
}
func TestMMSGroupMissingPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term", "iec"} {
		for _, fault := range []string{"plain", "sms", "sequence", "subscriptions", "marker", "later", "empty"} {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 11511, 21511)
				raw := missingRaw(t, co, req, mmsGroupPath(), false)
				want := mmsGroupMinimum(t)
				switch fault {
				case "sms":
					raw = smsOmit(t, raw, 2018)
					want = smsMinimum(t, 2018)
				case "sequence":
					raw, _ = invalidRaw(t, co, req, lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2}, avp.Key{})
					raw = mmsGroupOmit(t, raw, mmsGroupPath())
					want = nil
				case "subscriptions":
					raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
					want = nil
				case "marker":
					raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 455}))
					want = nil
				case "later":
					raw, _ = invalidRaw(t, co, req, originatorInterfacePath(), []byte{0, 0, 0, 99}, avp.Key{})
					raw = mmsGroupOmit(t, raw, mmsGroupPath())
				case "empty":
					raw, _ = invalidRaw(t, co, req, mmsGroupPath(), nil, avp.Key{})
					want = bytes.Replace(recipientMinimum(t), []byte{0, 0, 4, 177}, []byte{0, 0, 3, 118}, 1)
				}
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("invalid accepted")
				}
				q, ok := missingavp.Quote(cause, n.cfg.Limits, 180)
				if ok != (want != nil) || ok && !bytes.Equal(q, want) {
					t.Fatal(cause, q)
				}
			})
		}
	}
	for _, kind := range []string{"general", "balance", "refund"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11512, 21512)
		raw := reqBytes(t, co, req)
		if kind != "refund" {
			raw = mmsGroupOmit(t, raw, mmsGroupPath())
		}
		_, cause := co.Decode(raw)
		if _, ok := missingavp.Quote(cause, n.cfg.Limits, 180); ok {
			t.Fatal("money proof", kind, cause)
		}
	}
}
func TestRejectMMSGroupMissingDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"initial", "term", "iec"} {
			t.Run(kind, func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, kind, 11514, 21514)
				sendRejectRaw(t, remote, missingRaw(t, co, req, mmsGroupPath(), false))
				waitFor(t, func() bool { return !p.Info().Ready })
				if raw, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("default reply %x", raw)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func mmsGroupOmit(t *testing.T, raw []byte, path []missingSelector) []byte {
	t.Helper()
	out := append(append([]byte(nil), raw[:20]...), omitSelected(t, raw[20:], path, false)...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}

func TestRejectMMSGroupMissing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"initial", "term", "iec"} {
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
						combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11500, 21500))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedDuplicateRequest(t, n, co, kind, 11501, 21501)
					raw := missingRaw(t, co, req, mmsGroupPath(), false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					want := mmsGroupMinimum(t)
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
					follow := combinedDuplicateRequest(t, n, co, kind, 11502, 21502)
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
