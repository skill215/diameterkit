// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional S6c SRR omission over TCP/TLS and emergency slots.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestRejectConditionalSMEA(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, tc := range []struct {
			name  string
			app   int
			path  []missingSelector
			width int
			flags byte
		}{
			{"s6c-smea", 1, lengthPath(avp.Key{VendorID: 10415, Code: 3309}), 3, 192},
		} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/emergency-%t", tc.name, emergency), func(t *testing.T) {
					cfg := manualConfig(t)
					cfg.Resources.RejectSlots = 2
					cfg.Resources.EmergencyRequestMax = 4096
					cfg.Resources.EmergencyReplyMax = 4096
					if emergency {
						cfg.Resources.InboundClasses[0].Slots = 1
					}
					n, co := newInboundNode(t, cfg)
					p, remote := networkConnectNode(t, n, co, encrypted)
					makeRequest := func(id uint32) pdu.Message {
						req := mixedDuplicateRequest(t, n, co, tc.app, id, id+10000, gmscOrigin)
						body := req.Body.(*s6c.SRR)
						body.SRRFlags = avp.Some(uint32(1))
						body.SMRPSMEA = avp.Some([]byte{1, 0x91, 0xf1})
						return req
					}
					var held *Inbound
					if emergency {
						combinedSend(t, co, remote, makeRequest(10800))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := makeRequest(10801)
					path := tc.path
					raw := missingRaw(t, co, req, path, false)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					q, ok := missingavp.Quote(cause, cfg.Limits, 1024)
					assertMissingChain(t, q, path, tc.width, tc.flags)
					want := append([]byte(nil), q...)
					if !ok {
						t.Fatalf("minimum %x want %x (%v)", q, want, cause)
					}
					if _, ok := missingavp.Quote(cause, cfg.Limits, len(want)-1); ok {
						t.Fatal("quote budget")
					}
					if _, e := n.rejectFrame(raw, 20, rejectDecode, cause); e == nil {
						t.Fatal("reply budget")
					}
					if _, ok := missingDiagnostic(cause); ok {
						t.Fatal("conditional proof fell back")
					}
					var field *validation.FieldError
					if !errors.As(cause, &field) {
						t.Fatal(cause)
					}
					field.Path[0] = avp.Key{Code: 263}
					if _, err := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(err, ErrUnsupported) {
						t.Fatal("tampered evidence", err)
					}
					field.Path[0] = path[0].key

					empty, _ := invalidRaw(t, co, req, path, nil, avp.Key{})
					_, emptyErr := co.Decode(empty)
					if errors.Is(emptyErr, validation.ErrMissing) || (emptyErr == nil) != (tc.width == 0) {
						t.Fatal("present-empty semantics", emptyErr)
					}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5005)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
						t.Fatal("diagnostic/envelope")
					}
					if wire.Uint24(q[5:8]) != 15 || len(q) != 16 || q[15] != 0 {
						t.Fatal("conditional length/padding", q)
					}

					clear(raw)
					if !bytes.Equal(q, want) {
						t.Fatal("borrowed")
					}
					waitFor(t, func() bool {
						return n.Resources().RejectOutstanding == 0 && n.Resources().InboundOutstanding == before.InboundOutstanding
					})
					if n.Resources().InboundBytes != before.InboundBytes {
						t.Fatal("changed held charge")
					}
					noDuplicateDelivery(t, p)
					if held != nil {
						if _, e := held.Reject(ctxTest(t), FailureContextUnavailable); e != nil {
							t.Fatal(e)
						}
						readReject(t, remote, held.Message().Header, 5012)
					}
					follow := makeRequest(10802)
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

func TestConditionalSMEADecodeBoundaries(t *testing.T) {
	cfg := manualConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	key := avp.Key{VendorID: 10415, Code: 3309}
	flagKey := avp.Key{VendorID: 10415, Code: 3310}
	for _, kind := range []string{"absent-flags", "clear-condition", "reserved-flags", "short-flags", "duplicate-flags", "present-empty", "short-value", "bare", "identity-missing"} {
		t.Run(kind, func(t *testing.T) {
			req := mixedDuplicateRequest(t, n, co, 1, 10810, 20810, gmscOrigin)
			body := req.Body.(*s6c.SRR)
			body.SRRFlags = avp.Some(uint32(1))
			body.SMRPSMEA = avp.Some([]byte{1, 0x91, 0xf1})
			if kind == "absent-flags" {
				body.SRRFlags = avp.Optional[uint32]{}
			}
			if kind == "clear-condition" {
				body.SRRFlags = avp.Some(uint32(2))
			}
			raw := missingRaw(t, co, req, lengthPath(key), false)
			switch kind {
			case "reserved-flags", "short-flags", "duplicate-flags":
				if kind == "duplicate-flags" {
					raw, _ = excessRaw(t, co, req, lengthPath(flagKey), []byte{0, 0, 0, 1}, false)
				} else {
					value := []byte{0x80, 0, 0, 1}
					if kind == "short-flags" {
						value = []byte{1}
					}
					raw, _ = invalidRaw(t, co, req, lengthPath(flagKey), value, avp.Key{})
				}
				raw = append(raw[:20], omitSelected(t, raw[20:], lengthPath(key), false)...)
				wire.Put24(raw[1:4], uint32(len(raw)))
			case "present-empty", "short-value":
				value := []byte{}
				if kind == "short-value" {
					value = []byte{1, 0x91}
				}
				raw, _ = invalidRaw(t, co, req, lengthPath(key), value, avp.Key{})
			case "identity-missing":
				for _, identity := range []avp.Key{{Code: 1}, {VendorID: 10415, Code: 701}} {
					// Both alternatives are optional on the wire; remove whichever exists.
					d := wire.NewDecoder(raw[20:], cfg.Limits)
					found := false
					for {
						v, ok, e := d.Next()
						if e != nil {
							t.Fatal(e)
						}
						if !ok {
							break
						}
						if v.Key == identity {
							found = true
						}
					}
					if found {
						raw = append(raw[:20], omitSelected(t, raw[20:], lengthPath(identity), false)...)
						wire.Put24(raw[1:4], uint32(len(raw)))
					}
				}
			}
			_, cause := co.Decode(raw)
			if kind == "bare" {
				cause = validation.At(key, validation.ErrMissing)
			}
			q, ok := missingavp.Quote(cause, cfg.Limits, 16)
			if ok != (kind == "reserved-flags") {
				t.Fatal("condition proof", q, cause)
			}
			if kind == "absent-flags" || kind == "clear-condition" {
				if cause != nil {
					t.Fatal(cause)
				}
				return
			}
			if cause == nil {
				t.Fatal("accepted invalid request")
			}
			if kind == "short-flags" || kind == "duplicate-flags" {
				want := uint32(5014)
				if kind == "duplicate-flags" {
					want = 5009
				}
				_, code, ok := manualQuote(raw[20:], cause, cfg.Limits, 4096)
				if !ok || code != want {
					t.Fatal("prior decode failure lost", code, cause)
				}
			} else if kind != "reserved-flags" {
				if _, err := n.rejectFrame(raw, 4096, rejectDecode, cause); !errors.Is(err, ErrUnsupported) {
					t.Fatal("guessed conditional reply", err)
				}
			}
		})
	}
}

func TestRejectConditionalSMEADefaultPolicy(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := rejectConfig(t)
		cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, encrypted)
		req := inboundRequest(t, n, co, 10820)
		body := req.Body.(*s6c.SRR)
		body.SRRFlags = avp.Some(uint32(1))
		body.SMRPSMEA = avp.Some([]byte{1, 0x91, 0xf1})
		raw := missingRaw(t, co, req, lengthPath(avp.Key{VendorID: 10415, Code: 3309}), false)
		sendRejectRaw(t, remote, raw)
		waitFor(t, func() bool { return !p.Info().Ready })
		if frame, err := readAdapter(ctxTest(t), remote); err == nil {
			t.Fatalf("default policy replied %x", frame)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 && n.Resources().RejectOutstanding == 0 })
		closeDuplicateNode(t, n)
	})
}
