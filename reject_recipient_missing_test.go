// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed conditional Recipient-Address omission and bounded acceptance.
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

func recipientPath(code uint32) []missingSelector {
	return lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 877}, avp.Key{VendorID: 10415, Code: code})
}
func recipientMinimum(t *testing.T) []byte {
	part := func(code uint32, data []byte) []byte {
		return invalidPart(t, avp.Key{VendorID: 10415, Code: code}, 192, data)
	}
	return part(873, part(877, part(1201, append(part(899, make([]byte, 4)), part(897, nil)...))))
}
func recipientOmit(t *testing.T, raw []byte, code uint32) []byte {
	out := append(append([]byte(nil), raw[:20]...), omitSelected(t, raw[20:], recipientPath(code), false)...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func TestRecipientMissingBounds(t *testing.T) {
	cfg := combinedConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	want := recipientMinimum(t)
	for _, kind := range []string{"initial", "term", "iec"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11110, 21110)
		raw := missingRaw(t, co, req, recipientPath(1201), false)
		_, cause := co.Decode(raw)
		if !errors.Is(cause, validation.ErrMissing) {
			t.Fatal(kind, cause)
		}
		for _, negative := range []string{"exact", "bytes", "field", "frame", "avps", "depth", "outer-path", "middle-path", "leaf-path", "bare"} {
			t.Run(kind+"/"+negative, func(t *testing.T) {
				_, cause := co.Decode(raw)
				l := cfg.Limits
				budget := 64
				l.MaxFrameBytes = 92
				l.MaxFieldBytes = 64
				l.MaxAVPs = 6
				l.MaxRepeated = 1
				l.MaxDepth = 4
				var f *validation.FieldError
				if !errors.As(cause, &f) {
					t.Fatal(cause)
				}
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
				case "outer-path":
					f.Path[0].Code = 263
				case "middle-path":
					f.Path[1].Code++
				case "leaf-path":
					f.Path[2].Code++
				case "bare":
					cause = &validation.FieldError{Path: f.Path, Cause: validation.ErrMissing}
				}
				q, ok := missingavp.Quote(cause, l, budget)
				if ok != (negative == "exact") || ok && !bytes.Equal(q, want) {
					t.Fatal(q, cause)
				}
				if negative == "outer-path" {
					if _, ok := missingDiagnostic(cause); ok {
						t.Fatal("downgrade")
					}
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
			})
		}
		body := req.Body.(*combined.CCR)
		body.ServiceInformation.MMSInformation.Value.RecipientAddress = avp.Optional[combined.Address]{}
		_, err := co.Snapshot(req)
		if !errors.Is(err, validation.ErrMissing) {
			t.Fatal("send", err)
		}
		if _, ok := missingavp.Quote(err, cfg.Limits, 64); ok {
			t.Fatal("send proof")
		}
	}
}
func TestRecipientMissingScope(t *testing.T) {
	cfg := combinedConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"general", "balance", "refund"} {
		req := combinedDuplicateRequest(t, n, co, kind, 11120, 21120)
		if _, e := co.Decode(reqBytes(t, co, req)); e != nil {
			t.Fatal("money scope", kind, e)
		}
	}
	for _, kind := range []string{"empty-group", "wrong-type", "nul-value", "empty-value", "duplicate-group", "timestamp", "mms-absent", "sms-field"} {
		t.Run(kind, func(t *testing.T) {
			req := combinedDuplicateRequest(t, n, co, "initial", 11121, 21121)
			path := recipientPath(1201)
			raw := reqBytes(t, co, req)
			switch kind {
			case "empty-group":
				raw, _ = invalidRaw(t, co, req, path, nil, avp.Key{})
			case "wrong-type":
				raw, _ = invalidRaw(t, co, req, append(path, missingSelector{avp.Key{VendorID: 10415, Code: 899}, 0}), []byte{0, 0, 0, 2}, avp.Key{})
			case "nul-value":
				raw, _ = invalidRaw(t, co, req, append(path, missingSelector{avp.Key{VendorID: 10415, Code: 897}, 0}), []byte{0}, avp.Key{})
			case "empty-value":
				raw, _ = invalidRaw(t, co, req, append(path, missingSelector{avp.Key{VendorID: 10415, Code: 897}, 0}), nil, avp.Key{})
			case "duplicate-group":
				raw, _ = excessRaw(t, co, req, path, nil, false)
			case "timestamp":
				raw, _ = invalidRaw(t, co, req, recipientPath(1202), []byte{0, 0, 0, 0}, avp.Key{})
			case "mms-absent":
				raw = missingRaw(t, co, req, path[:2], false)
			case "sms-field":
				raw = missingRaw(t, co, req, lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: 2018}), false)
			}
			_, cause := co.Decode(raw)
			if kind == "empty-value" {
				if cause != nil {
					t.Fatal("zero octets changed", cause)
				}
				return
			}
			if cause == nil {
				t.Fatal("invalid accepted")
			}
			q, ok := missingavp.Quote(cause, cfg.Limits, 1024)
			if ok && bytes.Equal(q, recipientMinimum(t)) {
				t.Fatal("conditional proof outside scope")
			}
			switch kind {
			case "empty-group":
				if !errors.Is(cause, validation.ErrMissing) {
					t.Fatal(cause)
				}
			case "wrong-type":
				if !errors.Is(cause, validation.ErrUnsupported) {
					t.Fatal(cause)
				}
			case "nul-value", "timestamp":
				if !errors.Is(cause, validation.ErrField) {
					t.Fatal(cause)
				}
			case "duplicate-group":
				if !errors.Is(cause, validation.ErrDuplicate) {
					t.Fatal(cause)
				}
			}
		})
	}
}
func TestRejectRecipientMissingClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{0, 1202, 1209, 1212, 1210, 1216} {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				if code == 0 {
					cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "initial", 11130, 21130)
				raw := missingRaw(t, co, req, recipientPath(1201), false)
				if code != 0 {
					raw = recipientOmit(t, raw, code)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, 4096); ok {
						t.Fatal("multiple omission inferred")
					}
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal(e)
					}
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if b, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("unexpected reply %x", b)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestRejectRecipientMissing(t *testing.T) {
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
						combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11100, 21100))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedDuplicateRequest(t, n, co, kind, 11101, 21101)
					raw := missingRaw(t, co, req, recipientPath(1201), false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					want := recipientMinimum(t)
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
					follow := combinedDuplicateRequest(t, n, co, kind, 11102, 21102)
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
