// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// These are deliberately untrusted received bytes. The send API must not be
// used to manufacture the unresolved experimental pair being tested.
func resultWire(t *testing.T, c *codec.Codec, a pdu.Message, exp bool, vendor uint32, features bool) []byte {
	t.Helper()
	raw, e := c.Append(nil, a)
	if e != nil {
		t.Fatal(e)
	}
	b := make([]byte, 256)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	if exp {
		mark := w.BeginGroup(avp.Key{Code: 297}, 64)
		w.Uint32(avp.Key{Code: 266}, 64, vendor)
		w.Uint32(avp.Key{Code: 298}, 64, 5011)
		w.EndGroup(mark)
	}
	if features {
		mark := w.BeginGroup(avp.Key{VendorID: 10415, Code: 628}, 128)
		w.Uint32(avp.Key{Code: 266}, 64, 10415)
		w.Uint32(avp.Key{VendorID: 10415, Code: 629}, 128, 1)
		w.Uint32(avp.Key{VendorID: 10415, Code: 630}, 128, 0x80000000)
		w.EndGroup(mark)
	}
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	raw = append(raw, b[:w.Len()]...)
	wire.PutHeader(raw, a.Header, len(raw))
	return raw
}
func ratingDiagnostic() avp.Optional[pdu.FailedAVP] {
	return avp.Some(pdu.FailedAVP{Data: []byte{0, 0, 1, 205, 64, 0, 0, 12, 0, 0, 0, 0}})
}
func TestRoResultSafetyNetworkDelivery(t *testing.T) {
	for _, standard := range []bool{false, true} {
		name := "combined"
		if standard {
			name = "standard"
		}
		t.Run(name, func(t *testing.T) {
			eachNetwork(t, func(t *testing.T, encrypted bool) {
				cfg := combinedConfig(t)
				if standard {
					cfg = roConfig(t)
				}
				n, co := newInboundNode(t, cfg)
				p, remote := networkConnectNode(t, n, co, encrypted)
				port := portTest(t, n, 1)
				kinds := []string{"iec", "initial", "term", "general", "balance", "refund"}
				if standard {
					kinds = []string{"iec"}
				}
				for _, kind := range kinds {
					for _, scenario := range []string{"rating", "version", "conflict", "feature-failure", "unknown-vendor", "features-only", "e-version"} {
						var req pdu.Body = combinedBody(t, kind)
						if standard {
							req = roBody()
						}
						id, e := p.Submit(ctxTest(t), req, RequestOptions{}, port)
						if e != nil {
							t.Fatal(e)
						}
						r := routeRead(t, co, remote)
						code := uint32(2001)
						switch scenario {
						case "rating":
							code = 5031
						case "version", "e-version":
							code = 5011
						case "feature-failure":
							code = 5012
						}
						var body pdu.Body
						if standard {
							a := roAnswer(code)
							if scenario == "rating" {
								a.FailedAVP = ratingDiagnostic()
							}
							body = a
						} else {
							q := r.Body.(*combined.CCR)
							a := combinedAnswer(q, code)
							if scenario == "rating" {
								a.FailedAVP = ratingDiagnostic()
							}
							body = a
						}
						if scenario == "e-version" {
							body = &pdu.ErrorAnswer{ResultCode: 5011}
						}
						a, e := co.CompleteAnswer(r, body, serverOrigin)
						if e != nil {
							t.Fatal(kind, scenario, e)
						}
						exp := scenario == "conflict" || scenario == "feature-failure" || scenario == "unknown-vendor"
						vendor := uint32(10415)
						if scenario == "unknown-vendor" {
							vendor = 999
						}
						raw := resultWire(t, co, a, exp, vendor, exp || scenario == "features-only")
						sendRejectRaw(t, remote, raw)
						clear(raw)
						got := routeCompletion(t, n, port, id)
						if got.Outcome != AnswerReceived || got.Answer.Body == nil || got.Err != nil {
							t.Fatal(kind, scenario, got)
						}
						switch scenario {
						case "conflict":
							if got.Validation.Complete || !errors.Is(got.Validation.Err(), validation.ErrResultConflict) {
								t.Fatal(got)
							}
						case "feature-failure", "unknown-vendor", "features-only":
							if got.Validation.Complete || !errors.Is(got.Validation.Err(), validation.ErrUnsupported) {
								t.Fatal(got)
							}
						default:
							if !got.Validation.Complete {
								t.Fatal(got)
							}
						}
						if exp {
							switch x := got.Answer.Body.(type) {
							case *ro.CCA:
								if x.ExperimentalResult.Value.VendorID != vendor || x.ExperimentalResult.Value.ExperimentalResultCode != 5011 || x.ResultCode != code {
									t.Fatal(x)
								}
							case *combined.CCA:
								if x.ExperimentalResult.Value.VendorID != vendor || x.ExperimentalResult.Value.ExperimentalResultCode != 5011 || x.ResultCode != code {
									t.Fatal(x)
								}
							}
						}
					}
				}
				waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
				quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, e := remote.ReadPrefix(quiet, time.Second); !errors.Is(e, context.DeadlineExceeded) {
					t.Fatal("result triggered auto-retry", e)
				}
			})
		})
	}
}
func TestRoResultSafetyNetworkInvalidAndLate(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		port := portTest(t, n, 1)
		for _, scenario := range []string{"missing-top", "missing-diagnostic", "wrong-number", "e-experimental"} {
			id, e := p.Submit(ctxTest(t), combinedBody(t, "iec"), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			r := routeRead(t, co, remote)
			var body pdu.Body = combinedAnswer(r.Body.(*combined.CCR), 2001)
			if scenario == "e-experimental" {
				body = &pdu.ErrorAnswer{ResultCode: 5012}
			}
			a, e := co.CompleteAnswer(r, body, serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			raw := resultWire(t, co, a, scenario != "missing-diagnostic", 10415, true)
			switch scenario {
			case "missing-top":
				raw = withoutAVP(t, raw, avp.Key{Code: 268})
			case "missing-diagnostic":
				raw = replaceAVP32(t, raw, 268, 5031)
			case "wrong-number":
				raw = replaceAVP32(t, raw, 415, 1)
			}
			sendRejectRaw(t, remote, raw)
			got := routeCompletion(t, n, port, id)
			if !errors.Is(got.Err, validation.ErrInvalidAnswer) || got.Answer.Body != nil {
				t.Fatal(scenario, got)
			}
		}
		id, e := p.Submit(ctxTest(t), combinedBody(t, "general"), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		r := routeRead(t, co, remote)
		if e = n.Cancel(id); e != nil {
			t.Fatal(e)
		}
		if got := routeCompletion(t, n, port, id); !errors.Is(got.Err, context.Canceled) {
			t.Fatal(got)
		}
		a, e := co.CompleteAnswer(r, combinedAnswer(r.Body.(*combined.CCR), 2001), serverOrigin)
		if e != nil {
			t.Fatal(e)
		}
		sendRejectRaw(t, remote, resultWire(t, co, a, true, 10415, true))
		quiet, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(quiet); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late failure completed twice", e)
		}
		n.Close()
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 })
	})
}
func TestRoResultSafetyNetworkRespond(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		n, co := newInboundNode(t, combinedConfig(t))
		p, remote := networkConnectNode(t, n, co, encrypted)
		r, e := co.CompleteRequest(combinedBody(t, "iec"), pdu.RequestEnvelope{OriginHost: gmscOrigin.Host, OriginRealm: gmscOrigin.Realm, DestinationHost: n.cfg.Origin.Host, DestinationRealm: n.cfg.Origin.Realm, HopByHop: 400, EndToEnd: 500})
		if e != nil {
			t.Fatal(e)
		}
		if e = writeWire(ctxTest(t), remote, co, r); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		bad := combinedAnswer(r.Body.(*combined.CCR), 5031)
		if got, e := in.Respond(ctxTest(t), bad); got.Committed || !errors.Is(e, validation.ErrMissing) {
			t.Fatal(got, e)
		}
		for _, body := range []pdu.Body{&combined.CCA{ResultCode: 2001, AuthApplicationID: 4, CCRequestType: 4, ExperimentalResult: avp.Some(combined.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})}, &pdu.ErrorAnswer{ResultCode: 5012, ExperimentalResult: avp.Some(pdu.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})}} {
			if got, e := in.Respond(ctxTest(t), body); got.Committed || !errors.Is(e, validation.ErrUnsupported) {
				t.Fatal("pairing committed", got, e)
			}
		}
		bad.FailedAVP = ratingDiagnostic()
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if got, e := in.Respond(canceled, bad); got.Committed || !errors.Is(e, context.Canceled) {
			t.Fatal(got, e)
		}
		if got, e := in.Respond(ctxTest(t), bad); e != nil || !got.Committed {
			t.Fatal(got, e)
		}
		a := routeRead(t, co, remote)
		if e = co.ValidateAnswer(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		if a.Body.(*combined.CCA).ResultCode != 5031 || !a.Body.(*combined.CCA).FailedAVP.Present {
			t.Fatal(a)
		}
		waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
	})
}
