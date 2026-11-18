// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit SCAP1 ordered omission diagnostics and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/internal/cardinalityavp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"math/bits"
	"testing"
)

func scapCardinalityRequest(t *testing.T, n *Node, co *codec.Codec, id uint32) pdu.Message {
	t.Helper()
	body := scapBody(t)
	body.ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte("scap-cardinality")}}
	req, e := co.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: id, EndToEnd: id + 10000})
	if e != nil {
		t.Fatal(e)
	}
	return req
}
func TestCardinalitySCAPAllSubsets(t *testing.T) {
	n, co := newInboundNode(t, scapRuntimeConfig(t))
	defer closeDuplicateNode(t, n)
	req := scapCardinalityRequest(t, n, co, 11030)
	original := reqBytes(t, co, req)
	for mask := uint16(0); mask < 2047; mask++ {
		raw := retainCardinality(t, original, avp.Key{Code: 607, VendorID: 193}, mask)
		_, cause := co.Decode(raw)
		want := cardinalityMinimum(t, avp.Key{Code: 607, VendorID: 193}, avp.Key{Code: 608, VendorID: 193}, avp.Key{Code: 609, VendorID: 193}, 192, 11-bits.OnesCount16(mask))
		q, ok := cardinalityavp.Quote(raw[20:], cause, n.cfg.Limits, len(want))
		if !errors.Is(cause, validation.ErrMissing) || !ok || !bytes.Equal(q, want) {
			t.Fatalf("subset %d: %x %v", mask, q, cause)
		}
		if _, ok := cardinalityavp.Quote(raw[20:], cause, n.cfg.Limits, len(want)-1); ok {
			t.Fatalf("subset budget %d", mask)
		}
		if _, ok := cardinalityavp.Quote(original[20:], cause, n.cfg.Limits, 440); ok {
			t.Fatal("count mismatch")
		}
	}
	if _, e := co.Decode(original); e != nil {
		t.Fatal("full valid request", e)
	}
}
func TestCardinalitySCAPDecodeBoundaries(t *testing.T) {
	n, co := newInboundNode(t, scapRuntimeConfig(t))
	defer closeDuplicateNode(t, n)
	key := avp.Key{Code: 607, VendorID: 193}
	for _, kind := range []string{"unknown-type", "negative-type", "duplicate-type", "wrong-order", "full-wrong-order", "full-wrong-value", "empty-group", "excess", "bare", "send-ten", "send-zero"} {
		t.Run(kind, func(t *testing.T) {
			req := scapCardinalityRequest(t, n, co, 11010)
			body := req.Body.(*scap1.ACR)
			if kind == "send-ten" || kind == "send-zero" {
				body.ServiceParameterInfo = body.ServiceParameterInfo[:10]
				if kind == "send-zero" {
					body.ServiceParameterInfo = nil
				}
				_, e := co.Snapshot(req)
				if !errors.Is(e, validation.ErrMissing) || cardinalityavp.HasEvidence(e) {
					t.Fatal("send scope", e)
				}
				return
			}
			raw := reqBytes(t, co, req)
			switch kind {
			case "unknown-type", "negative-type", "duplicate-type", "wrong-order", "full-wrong-order":
				index := 0
				value := []byte{0, 0, 0, 42}
				if kind == "negative-type" {
					value = []byte{255, 255, 255, 255}
				}
				if kind == "duplicate-type" {
					index = 1
					value = []byte{0, 0, 0, 0}
				}
				if kind == "wrong-order" || kind == "full-wrong-order" {
					value = []byte{0, 0, 0, 5}
				}
				raw, _ = invalidRaw(t, co, req, []missingSelector{{key, index}, {avp.Key{Code: 608, VendorID: 193}, 0}}, value, avp.Key{})
				if kind != "full-wrong-order" {
					raw = retainCardinality(t, raw, key, 3)
				}
			case "full-wrong-value":
				raw, _ = invalidRaw(t, co, req, []missingSelector{{key, 0}, {avp.Key{Code: 609, VendorID: 193}, 0}}, []byte("2"), avp.Key{})
			case "empty-group":
				raw, _ = invalidRaw(t, co, req, lengthPath(key), nil, avp.Key{})
			case "excess":
				raw, _ = excessRaw(t, co, req, []missingSelector{{key, 10}}, nil, false)
			case "bare":
				raw = retainCardinality(t, raw, key, 0)
			}
			_, cause := co.Decode(raw)
			if kind == "bare" {
				cause = validation.At(key, validation.ErrMissing)
			}
			if cause == nil || cardinalityavp.HasEvidence(cause) {
				t.Fatal("unreviewed omission", cause)
			}
			if _, ok := cardinalityavp.Quote(raw[20:], cause, n.cfg.Limits, 440); ok {
				t.Fatal("guessed cardinality")
			}
			switch kind {
			case "full-wrong-order":
				if !errors.Is(cause, validation.ErrUnsupported) {
					t.Fatal(cause)
				}
			case "full-wrong-value":
				if !errors.Is(cause, validation.ErrField) {
					t.Fatal(cause)
				}
			case "excess":
				if !errors.Is(cause, validation.ErrDuplicate) {
					t.Fatal(cause)
				}
			default:
				if !errors.Is(cause, validation.ErrMissing) {
					t.Fatal(cause)
				}
			}
		})
	}
}

func TestRejectCardinalitySCAP(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, mask := range []uint16{0, 1, 0x7fe, 0x3ff} {
			for _, emergency := range []bool{false, true} {
				t.Run(fmt.Sprintf("mask-%d/emergency-%t", mask, emergency), func(t *testing.T) {
					cfg := scapRuntimeConfig(t)
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
						combinedSend(t, co, remote, scapCardinalityRequest(t, n, co, 11000))
						var e error
						held, e = p.Receive(ctxTest(t))
						if e != nil {
							t.Fatal(e)
						}
					}
					before := n.Resources()
					req := scapCardinalityRequest(t, n, co, 11001)
					raw := retainCardinality(t, reqBytes(t, co, req), avp.Key{Code: 607, VendorID: 193}, mask)
					_, cause := co.Decode(raw)
					if !errors.Is(cause, validation.ErrMissing) {
						t.Fatal(cause)
					}

					count := 11 - bits.OnesCount16(mask)
					want := cardinalityMinimum(t, avp.Key{Code: 607, VendorID: 193}, avp.Key{Code: 608, VendorID: 193}, avp.Key{Code: 609, VendorID: 193}, 192, count)
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
					field.Path[0] = avp.Key{Code: 607, VendorID: 193}
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
					follow := scapCardinalityRequest(t, n, co, 11002)
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

func TestRejectCardinalitySCAPDefault(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		cfg := scapRuntimeConfig(t)
		cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
		n, co := newInboundNode(t, cfg)
		p, remote := networkConnectNode(t, n, co, encrypted)
		req := scapCardinalityRequest(t, n, co, 11020)
		raw := retainCardinality(t, reqBytes(t, co, req), avp.Key{Code: 607, VendorID: 193}, 0)
		sendRejectRaw(t, remote, raw)
		waitFor(t, func() bool { return !p.Info().Ready })
		if frame, e := readAdapter(ctxTest(t), remote); e == nil {
			t.Fatalf("default replied %x", frame)
		}
		closeDuplicateNode(t, n)
	})
}
