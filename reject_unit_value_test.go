// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated Reporting-Reason and requested-unit value acceptance.
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

func unitValuePath(kind string) []missingSelector {
	if kind == "term" {
		return lengthPath(avp.Key{Code: 456}, avp.Key{VendorID: 10415, Code: 872})
	}
	return lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}, avp.Key{Code: 417})
}
func unitValueData(kind string, value uint64) []byte {
	if kind == "term" {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(value))
		return b
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, value)
	return b
}
func unitValueQuote(t *testing.T, kind string, value uint64) []byte {
	data := unitValueData(kind, value)
	var leaf []byte
	if kind == "term" {
		leaf = invalidPart(t, avp.Key{VendorID: 10415, Code: 872}, 192, data)
	} else {
		leaf = invalidPart(t, avp.Key{Code: 437}, 64, invalidPart(t, avp.Key{Code: 417}, 64, data))
	}
	return invalidPart(t, avp.Key{Code: 456}, 64, leaf)
}
func TestRejectUnitValue(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			{
				for _, value := range []uint64{0, ^uint64(0)} {
					for _, emergency := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/emergency-%t", kind, value, emergency), func(t *testing.T) {
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
								combinedSend(t, co, remote, combinedDuplicateRequest(t, n, co, kind, 12230, 22300))
								var e error
								held, e = p.Receive(ctxTest(t))
								if e != nil {
									t.Fatal(e)
								}
							}
							before := n.Resources()
							req := combinedDuplicateRequest(t, n, co, kind, 12301, 22301)
							data := unitValueData(kind, value)
							raw, _ := invalidRaw(t, co, req, unitValuePath(kind), data, avp.Key{})
							want := unitValueQuote(t, kind, value)
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
							field.Path[0] = avp.Key{Code: 456}
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
							follow := combinedDuplicateRequest(t, n, co, kind, 12302, 22302)
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

func TestUnitValueWire(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, value := range []uint64{0, 1, 2, 3, 0x7fffffff, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)} {
			t.Run(fmt.Sprintf("%s/%d", kind, value), func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12310, 22310), unitValuePath(kind), unitValueData(kind, value), avp.Key{})
				_, cause := co.Decode(raw)
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 32)
				want := value != 1
				if kind == "term" {
					want = uint32(value) != 2
				}
				if ok != want || ok && (code != 5004 || !bytes.Equal(out, unitValueQuote(t, kind, value))) || !want && cause != nil {
					t.Fatal(cause, code, out)
				}
			})
		}
	}
}
func unitValueFault(t *testing.T, raw []byte, kind, fault string) []byte {
	t.Helper()
	edit := func(path []missingSelector, data []byte) {
		b, _ := invalidEdit(t, raw[20:], path, data, avp.Key{})
		raw = append(raw[:20:20], b...)
		wire.Put24(raw[1:4], uint32(len(raw)))
	}
	switch fault {
	case "sequence":
		edit(lengthPath(avp.Key{Code: 416}), []byte{0, 0, 0, 2})
	case "subscriptions":
		raw = mmsGroupOmit(t, raw, []missingSelector{{avp.Key{Code: 443}, 0}})
	case "sms":
		raw = smsOmit(t, raw, 2018)
	case "mms":
		raw = mmsGroupOmit(t, raw, mmsGroupPath())
	case "timestamp":
		edit(recipientPath(1202), []byte{0, 0, 0, 0})
	case "interface":
		p := ecurInterfacePath(true)
		if kind == "iec" {
			p = lengthPath(avp.Key{VendorID: 10415, Code: 873}, avp.Key{VendorID: 10415, Code: 2000}, avp.Key{VendorID: 10415, Code: 2009}, avp.Key{VendorID: 10415, Code: 2006})
		}
		edit(p, []byte{0, 0, 0, 9})
	case "selected-width":
		edit(unitValuePath(kind), []byte{1})
	case "missing-unit":
		c := uint32(437)
		if kind == "term" {
			c = 446
		}
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: c}))
	case "missing-reason":
		raw = mmsGroupOmit(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{VendorID: 10415, Code: 872}))
	case "later-sms":
		if kind == "term" {
			raw = smsOmit(t, raw, 2014)
		} else {
			body := invalidPart(t, avp.Key{VendorID: 10415, Code: 2014}, 192, []byte{0, 0, 0, 0})
			raw = unitAppendRaw(t, raw, smsPath(0)[:2], body)
		}
	case "forbidden-unit":
		c := uint32(446)
		if kind == "term" {
			c = 437
		}
		body := invalidPart(t, avp.Key{Code: c}, 64, invalidPart(t, avp.Key{Code: 417}, 64, unitValueData("initial", 1)))
		raw = unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 456}), body)
	case "money", "standard-money":
		c := uint32(420)
		if fault == "standard-money" {
			c = 413
		}
		value := invalidPart(t, avp.Key{Code: 447}, 64, unitValueData("initial", 1))
		value = append(value, invalidPart(t, avp.Key{Code: 429}, 64, []byte{0, 0, 0, 0})...)
		money := invalidPart(t, avp.Key{Code: 445}, 64, value)
		money = append(money, invalidPart(t, avp.Key{Code: 425}, 64, []byte{0, 0, 3, 72})...)
		raw = unitAppendRaw(t, raw, lengthPath(avp.Key{Code: 456}, avp.Key{Code: 437}), invalidPart(t, avp.Key{Code: c}, 64, money))
	case "used-width":
		edit(lengthPath(avp.Key{Code: 456}, avp.Key{Code: 446}, avp.Key{Code: 417}), []byte{1})
	}
	return raw
}

// Rebuild one containing area without depending on production quote walkers.
func unitAppendRaw(t *testing.T, raw []byte, path []missingSelector, extra []byte) []byte {
	t.Helper()
	var walk func([]byte, []missingSelector) []byte
	walk = func(src []byte, p []missingSelector) []byte {
		d := wire.NewDecoder(src, wire.DefaultLimits())
		var out []byte
		for {
			v, ok, e := d.Next()
			if e != nil {
				t.Fatal(e)
			}
			if !ok {
				break
			}
			data := v.Data
			if v.Key == p[0].key {
				if len(p) == 1 {
					data = append(append([]byte{}, data...), extra...)
				} else {
					data = walk(data, p[1:])
				}
			}
			out = append(out, invalidPart(t, v.Key, v.Flags, data)...)
		}
		return out
	}
	body := walk(raw[20:], path)
	out := append(raw[:20:20], body...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func TestUnitValuePrecedence(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		faults := []string{"plain", "sequence", "subscriptions", "sms", "mms", "timestamp", "interface", "selected-width", "missing-unit", "forbidden-unit", "later-sms"}
		if kind == "term" {
			faults = append(faults, "missing-reason", "used-width")
		} else {
			faults = append(faults, "money", "standard-money")
		}
		for _, fault := range faults {
			t.Run(kind+"/"+fault, func(t *testing.T) {
				raw, _ := invalidRaw(t, co, combinedDuplicateRequest(t, n, co, kind, 12311, 22311), unitValuePath(kind), unitValueData(kind, 0), avp.Key{})
				raw = unitValueFault(t, raw, kind, fault)
				_, cause := co.Decode(raw)
				if cause == nil {
					t.Fatal("accepted")
				}
				out, code, ok := excessavp.QuoteInvalid(raw[20:], cause, n.cfg.Limits, 4096)
				expected := uint32(0)
				if fault == "plain" || fault == "later-sms" || fault == "interface" {
					expected = 5004
				}
				if fault == "selected-width" || fault == "used-width" {
					expected = 5014
				}
				if ok != (expected != 0) || code != expected {
					t.Fatal(cause, code, out)
				}
				if fault == "plain" || fault == "later-sms" {
					if !bytes.Equal(out, unitValueQuote(t, kind, 0)) {
						t.Fatal("order")
					}
				}
				if fault == "interface" {
					var proof interface{ DiameterUnitValueV1() string }
					if errors.As(cause, &proof) {
						t.Fatal("earlier interface displaced")
					}
				}
			})
		}
	}
}
func TestRejectUnitValueFailurePolicy(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			faults := []string{"default", "forbidden-unit", "missing-unit"}
			if kind == "term" {
				faults = append(faults, "missing-reason")
			} else {
				faults = append(faults, "money", "standard-money")
			}
			for _, fault := range faults {
				t.Run(kind+"/"+fault, func(t *testing.T) {
					cfg := combinedConfig(t)
					cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
					if fault == "default" {
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
					}
					n, co := newInboundNode(t, cfg)
					defer closeDuplicateNode(t, n)
					p, remote := networkConnectNode(t, n, co, encrypted)
					req := combinedDuplicateRequest(t, n, co, kind, 12312, 22312)
					raw, _ := invalidRaw(t, co, req, unitValuePath(kind), unitValueData(kind, 0), avp.Key{})
					raw = unitValueFault(t, raw, kind, fault)
					sendRejectRaw(t, remote, raw)
					if kind == "initial" && fault == "missing-unit" {
						ans := readReject(t, remote, req.Header, 5005)
						if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, unitValueQuote(t, "initial", 0)) {
							t.Fatal(ans.FailedAVP)
						}
						noDuplicateDelivery(t, p)
						return
					}
					waitFor(t, func() bool { return !p.Info().Ready })
					if b, e := readAdapter(ctxTest(t), remote); e == nil {
						t.Fatalf("unexpected reply %x", b)
					}
				})
			}
		}
	})
}
