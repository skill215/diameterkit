// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Combined cardinality diagnostics, wire precedence and TCP/TLS resource acceptance.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/internal/cardinalityavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

// Retain selected top-level repeated instances without reordering other AVPs.
func retainCardinality(t *testing.T, raw []byte, key avp.Key, mask uint16) []byte {
	t.Helper()
	out := append([]byte(nil), raw[:20]...)
	d := wire.NewDecoder(raw[20:], wire.DefaultLimits())
	index := 0
	for {
		before := d.Remaining()
		v, ok, e := d.Next()
		if e != nil {
			t.Fatal(e)
		}
		if !ok {
			break
		}
		size := before - d.Remaining()
		keep := true
		if v.Key == key {
			keep = mask&(1<<index) != 0
			index++
		}
		if keep {
			out = append(out, raw[20+v.Offset:20+v.Offset+size]...)
		}
	}
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}
func cardinalityMinimum(t *testing.T, group, typ, value avp.Key, flags byte, count int) []byte {
	t.Helper()
	emit := func(w *wire.Encoder) {
		for i := 0; i < count; i++ {
			g := w.BeginGroup(group, flags)
			w.Uint32(typ, flags, 0)
			w.Bytes(value, flags, nil)
			w.EndGroup(g)
		}
	}
	w := wire.NewEncoder(nil, wire.DefaultLimits())
	emit(w)
	out := make([]byte, w.Len())
	w = wire.NewEncoder(out, wire.DefaultLimits())
	emit(w)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	return out
}
func combinedCardinalityRequest(t *testing.T, n *Node, co *codec.Codec, id uint32) pdu.Message {
	return combinedDuplicateRequest(t, n, co, "general", id, id+10000)
}

func TestRejectCardinalityCombined(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, mask := range []uint16{0, 1, 2} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("mask-%d/emergency-%t", mask, emergency), func(t *testing.T) {
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
						combinedSend(t, co, remote, combinedCardinalityRequest(t, n, co, 10900))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := combinedCardinalityRequest(t, n, co, 10901)
					raw := retainCardinality(t, reqBytes(t, co, req), avp.Key{Code: 443}, mask)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}
					if mask != 0 && !errors.Is(cause, validation.ErrUnsupported) {
						t.Fatal("lost previous sentinel")
					}
					count := 1
					if mask == 0 {
						count = 2
					}
					want := cardinalityMinimum(t, avp.Key{Code: 443}, avp.Key{Code: 450}, avp.Key{Code: 444}, 64, count)
					q, ok := cardinalityavp.Quote(raw[20:], cause, cfg.Limits, len(want))
					if !ok || !bytes.Equal(q, want) {
						t.Fatal(q, cause)
					}
					if _, ok := cardinalityavp.Quote(raw[20:], cause, cfg.Limits, len(want)-1); ok {
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
					field.Path[0] = avp.Key{Code: 443}
					sendRejectRaw(t, remote, raw)
					ans := readReject(t, remote, req.Header, 5005)
					if !ans.FailedAVP.Present || !bytes.Equal(ans.FailedAVP.Value.Data, want) || !ans.SessionID.Present || len(ans.ProxyInfo) != 1 {
						t.Fatal("diagnostic/envelope")
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
					follow := combinedCardinalityRequest(t, n, co, 10902)
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
func TestCardinalityCombinedDecodeBoundaries(t *testing.T) {
	cfg := combinedConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"unknown-type", "duplicate-type", "wrong-order", "empty-group", "excess", "bare", "send-one", "send-zero"} {
		t.Run(kind, func(t *testing.T) {
			req := combinedCardinalityRequest(t, n, co, 10910)
			body := req.Body.(*combined.CCR)
			if kind == "send-one" || kind == "send-zero" {
				body.SubscriptionID = body.SubscriptionID[:1]
				if kind == "send-zero" {
					body.SubscriptionID = nil
				}
				_, err := co.Snapshot(req)
				if err == nil || cardinalityavp.HasEvidence(err) {
					t.Fatal("send scope", err)
				}
				return
			}
			raw := reqBytes(t, co, req)
			switch kind {
			case "unknown-type":
				raw, _ = invalidRaw(t, co, req, lengthPath(avp.Key{Code: 443}, avp.Key{Code: 450}), []byte{0, 0, 0, 42}, avp.Key{})
				raw = retainCardinality(t, raw, avp.Key{Code: 443}, 1)
			case "duplicate-type":
				path := []missingSelector{{avp.Key{Code: 443}, 1}, {avp.Key{Code: 450}, 0}}
				raw, _ = invalidRaw(t, co, req, path, []byte{0, 0, 0, 0}, avp.Key{})
			case "wrong-order":
				body.SubscriptionID[0], body.SubscriptionID[1] = body.SubscriptionID[1], body.SubscriptionID[0]
				// Encode the valid frame first, then exchange its two actual group bytes.
				original := reqBytes(t, co, combinedCardinalityRequest(t, n, co, 10910))
				d := wire.NewDecoder(original[20:], cfg.Limits)
				var parts [][]byte
				var offsets []int
				for {
					before := d.Remaining()
					v, ok, e := d.Next()
					if e != nil {
						t.Fatal(e)
					}
					if !ok {
						break
					}
					if v.Key.Code == 443 {
						parts = append(parts, append([]byte(nil), original[20+v.Offset:20+v.Offset+before-d.Remaining()]...))
						offsets = append(offsets, 20+v.Offset)
					}
				}
				raw = append([]byte(nil), original[:offsets[0]]...)
				raw = append(raw, parts[1]...)
				raw = append(raw, original[offsets[0]+len(parts[0]):offsets[1]]...)
				raw = append(raw, parts[0]...)
				raw = append(raw, original[offsets[1]+len(parts[1]):]...)
			case "empty-group":
				raw, _ = invalidRaw(t, co, req, lengthPath(avp.Key{Code: 443}), nil, avp.Key{})
			case "excess":
				raw, _ = excessRaw(t, co, req, []missingSelector{{avp.Key{Code: 443}, 1}}, nil, false)
			case "bare":
				raw = retainCardinality(t, raw, avp.Key{Code: 443}, 0)
			}
			_, cause := co.Decode(raw)
			if kind == "bare" {
				cause = validation.At(avp.Key{Code: 443}, validation.ErrMissing)
			}
			if cause == nil {
				t.Fatal("accepted invalid input")
			}
			if _, ok := cardinalityavp.Quote(raw[20:], cause, cfg.Limits, 4096); ok {
				t.Fatal("guessed quantity", cause)
			}
			if kind == "excess" && !errors.Is(cause, validation.ErrDuplicate) {
				t.Fatal(cause)
			}
			if kind == "empty-group" && !errors.Is(cause, validation.ErrMissing) {
				t.Fatal(cause)
			}
		})
	}
}
func TestRejectCardinalityCombinedDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := combinedConfig(t)
		cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, encrypted)
		req := combinedCardinalityRequest(t, n, co, 10920)
		raw := retainCardinality(t, reqBytes(t, co, req), avp.Key{Code: 443}, 0)
		sendRejectRaw(t, remote, raw)
		waitFor(t, func() bool { return !p.Info().Ready })
		if frame, e := readAdapter(ctxTest(t), remote); e == nil {
			t.Fatalf("default replied %x", frame)
		}
		closeDuplicateNode(t, n)
	})
}

func TestCardinalityCombinedTemplates(t *testing.T) {
	cfg := combinedConfig(t)
	n, co := newInboundNode(t, cfg)
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, mask := range []uint16{0, 1, 2} {
			t.Run(fmt.Sprintf("%s/%d", kind, mask), func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 10930, 20930)
				raw := retainCardinality(t, reqBytes(t, co, req), avp.Key{Code: 443}, mask)
				_, cause := co.Decode(raw)
				count := 1
				if mask == 0 {
					count = 2
				}
				q, ok := cardinalityavp.Quote(raw[20:], cause, cfg.Limits, 56)
				want := cardinalityMinimum(t, avp.Key{Code: 443}, avp.Key{Code: 450}, avp.Key{Code: 444}, 64, count)
				if !ok || !bytes.Equal(q, want) {
					t.Fatal(kind, q, cause)
				}
			})
		}
	}
}
