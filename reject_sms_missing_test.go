// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed isolated combined SMS condition diagnostics and acceptance.
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
	"math/bits"
	"testing"
)

func smsPath(code uint32) []missingSelector {
	return lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: code})
}
func smsMinimum(t *testing.T, code uint32) []byte {
	width := 0
	if code == 2011 || code == 2012 {
		width = 4
	}
	part := func(c uint32, data []byte) []byte {
		return invalidPart(t, avp.Key{VendorID: 10415, Code: c}, 192, data)
	}
	return part(873, part(2000, part(code, make([]byte, width))))
}
func smsOmit(t *testing.T, raw []byte, code uint32) []byte {
	out := append(append([]byte(nil), raw[:20]...), omitSelected(t, raw[20:], smsPath(code), false)...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func TestSMSMissingCombinations(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term", "iec"} {
		lists := [][]uint32{{2018, 2010, 2017, 2011, 2015}}
		if kind == "term" {
			lists = append(lists, []uint32{2014, 2012})
		}
		for _, codes := range lists {
			original := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 11310, 21310))
			for mask := uint8(0); int(mask) < 1<<len(codes); mask++ {
				raw := append([]byte(nil), original...)
				var last uint32
				for i, code := range codes {
					if mask&(1<<i) != 0 {
						raw = smsOmit(t, raw, code)
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
				q, ok := missingavp.Quote(cause, n.cfg.Limits, 40)
				if ok != (bits.OnesCount8(mask) == 1) || ok && !bytes.Equal(q, smsMinimum(t, last)) {
					t.Fatal(kind, mask, cause, q)
				}
				if !ok {
					if _, e := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(e, ErrUnsupported) {
						t.Fatal("multi omission", e)
					}
				}
			}
		}
	}
}
func TestSMSMissingPrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"common-mms", "common-term", "mms-term", "common-multi", "originator"} {
		req := combinedDuplicateRequest(t, n, co, "term", 11311, 21311)
		raw := reqBytes(t, co, req)
		var want []byte
		switch kind {
		case "common-mms":
			raw = smsOmit(t, raw, 2018)
			raw = recipientOmit(t, raw, 1202)
			want = smsMinimum(t, 2018)
		case "common-term":
			raw = smsOmit(t, raw, 2018)
			raw = smsOmit(t, raw, 2014)
			want = smsMinimum(t, 2018)
		case "mms-term":
			raw = recipientOmit(t, raw, 1202)
			raw = smsOmit(t, raw, 2014)
			want = mmsMinimum(t, 1202)
		case "common-multi":
			raw = smsOmit(t, raw, 2018)
			raw = smsOmit(t, raw, 2010)
			raw = smsOmit(t, raw, 2014)
		case "originator":
			req = combinedDuplicateRequest(t, n, co, "iec", 11311, 21311)
			raw = missingRaw(t, co, req, smsPath(2008), false)
		}
		_, cause := co.Decode(raw)
		q, ok := missingavp.Quote(cause, n.cfg.Limits, 40)
		if ok != (want != nil) || ok && !bytes.Equal(q, want) {
			t.Fatal(kind, cause, q)
		}
		if kind == "originator" && !errors.Is(cause, validation.ErrUnsupported) {
			t.Fatal(cause)
		}
	}
}
func TestSMSMissingValueBoundaries(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012} {
		req := combinedDuplicateRequest(t, n, co, "term", 11312, 21312)
		value := []byte{0}
		wantErr := validation.ErrField
		if code == 2011 || code == 2012 {
			value = []byte{0, 0, 0}
			wantErr = avp.ErrValue
		}
		if code == 2015 || code == 2014 {
			value = []byte{0, 255}
			wantErr = nil
		}
		raw, _ := invalidRaw(t, co, req, smsPath(code), value, avp.Key{})
		_, cause := co.Decode(raw)
		if wantErr == nil {
			if cause != nil {
				t.Fatal("opaque value", code, cause)
			}
		} else if !errors.Is(cause, wantErr) {
			t.Fatal(code, cause)
		}
		if _, ok := missingavp.Quote(cause, n.cfg.Limits, 40); ok {
			t.Fatal("present value became absent")
		}
		if code != 2011 && code != 2012 {
			body := req.Body.(*combined.CCR)
			s := &body.ServiceInformation.SMSInformation
			if code == 2018 || code == 2017 {
				s.ClientAddressWire = avp.Some([]byte{})
				s.SMSCAddress = avp.Some([]byte{})
			}
			if code == 2010 {
				s.RecipientSCCPAddress = avp.Some([]byte{})
			}
			if code == 2015 {
				s.SMUserDataHeader = avp.Some([]byte{})
			}
			if code == 2014 {
				s.SMStatus = avp.Some([]byte{})
			}
			if _, e := co.Decode(reqBytes(t, co, req)); e != nil {
				t.Fatal("present empty", code, e)
			}
		}
	}
}
func TestRejectSMSMissingMultipleClosed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, codes := range [][]uint32{{2018, 2010}, {2014, 2012}} {
			t.Run(fmt.Sprint(codes), func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, "term", 11313, 21313))
				for _, code := range codes {
					raw = smsOmit(t, raw, code)
				}
				sendRejectRaw(t, remote, raw)
				waitFor(t, func() bool { return !p.Info().Ready })
				if raw, e := readAdapter(ctxTest(t), remote); e == nil {
					t.Fatalf("multiple reply %x", raw)
				}
				closeDuplicateNode(t, n)
			})
		}
	})
}

func TestSMSMissingBounds(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, code := range []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012} {
		req := combinedDuplicateRequest(t, n, co, "term", 11311, 21311)
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

func TestRejectSMSMissingDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012} {
			t.Run(fmt.Sprint(code), func(t *testing.T) {
				cfg := combinedConfig(t)
				cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				req := combinedDuplicateRequest(t, n, co, "term", 11314, 21314)
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

func TestRejectSMSMissing(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, code := range []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012} {
			for _, kind := range []string{"initial", "term", "iec"} {
				if (code == 2014 || code == 2012) && kind != "term" {
					continue
				}
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
							combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 11300, 21300))
							var e error
							held, e = p.Receive(ctxTest(t))
							if e != nil {
								t.Fatal(e)
							}
						}
						before := n.Resources()
						req := combinedDuplicateRequest(t, n, co, kind, 11301, 21301)
						raw := missingRaw(t, co, req, smsPath(code), false)
						_, cause := co.Decode(raw)
						if !errors.Is(cause, validation.ErrMissing) {
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
						follow := combinedDuplicateRequest(t, n, co, kind, 11302, 21302)
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
