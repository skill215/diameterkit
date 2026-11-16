// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestRoSMSResultsFrozenPairs(t *testing.T) {
	c := roCodec(t)
	for _, kind := range []struct{ name, request string }{{"iec", "v-iec-ccr-01"}, {"initial", "v-sms-ccr-01"}, {"term", "v-term-ccr-01"}} {
		for _, code := range []uint32{4010, 4011, 5030} {
			for _, branch := range []struct {
				name       string
				top, child uint32
			}{{"top", code, 0}, {"child", 2001, code}, {"both", code, code}, {"top-failure-child-success", code, 2001}} {
				t.Run(fmt.Sprintf("%s/%s/%d", kind.name, branch.name, code), func(t *testing.T) {
					r := decoded(t, c, kind.request)
					raw := fixture(t, fmt.Sprintf("b71-%s-%s-%d", kind.name, branch.name, code))
					a, err := c.Decode(raw)
					if err != nil {
						t.Fatal(err)
					}
					if err := c.ValidateAnswer(r, a, roSources()); err != nil {
						t.Fatal(err)
					}
					if report := c.ValidateAnswerContext(r, a, roSources(), validation.Context{}); !report.Complete {
						t.Fatal(report)
					}
					v := a.Body.(*ro.CCA)
					if v.ResultCode != branch.top {
						t.Fatal("top result flattened", v)
					}
					if branch.child == 0 {
						if len(v.MultipleServicesCreditControl) != 0 {
							t.Fatal("invented parent")
						}
					} else {
						s := v.MultipleServicesCreditControl[0]
						if !s.ResultCode.Present || s.ResultCode.Value != branch.child || s.GrantedServiceUnit.Present {
							t.Fatal("child result/grant", s)
						}
					}
					out := encoded(t, c, a)
					again, err := c.Decode(out)
					if err != nil || !reflect.DeepEqual(a, again) {
						t.Fatal("semantic roundtrip", err)
					}
					clear(raw)
					if !bytes.Equal(out, encoded(t, c, a)) {
						t.Fatal("borrowed wire input")
					}
				})
			}
		}
	}
	for _, name := range []string{"child-unselected", "child-rating", "duplicate-parent", "top-token"} {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Decode(fixture(t, "b71-"+name)); !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(err)
			}
		})
	}
	if err := c.ValidateAnswer(decoded(t, c, "v-iec-ccr-01"), decoded(t, c, "b71-wrong-service"), roSources()); !errors.Is(err, validation.ErrInvalidAnswer) {
		t.Fatal(err)
	}
}

func TestRoSMSRefundPresenceOwnershipAndEcho(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name    string
		present bool
		token   []byte
	}{{"absent", false, nil}, {"empty", true, nil}, {"binary", true, []byte{0, 255, 0, 1}}, {"failure", true, []byte{0, 255}}} {
		t.Run(tc.name, func(t *testing.T) {
			r, a := decoded(t, c, "v-iec-ccr-01"), decoded(t, c, "b71-token-"+tc.name)
			if err := c.ValidateAnswer(r, a, roSources()); err != nil {
				t.Fatal(err)
			}
			q, v := r.Body.(*ro.CCR), a.Body.(*ro.CCA)
			got, err := ro.RefundInformationForService(q, v, 1, len(tc.token))
			if err != nil || got.Present != tc.present || !bytes.Equal(got.Value, tc.token) {
				t.Fatal(got, err)
			}
			if err = ro.CheckRefundInformationEcho(q, v, 1, got, len(tc.token)); err != nil {
				t.Fatal(err)
			}
			wrong := got
			wrong.Present = !got.Present
			if err = ro.CheckRefundInformationEcho(q, v, 1, wrong, 8); !errors.Is(err, validation.ErrContext) {
				t.Fatal("absence equals empty", err)
			}
			if tc.present {
				if err = ro.CheckRefundInformationEcho(q, v, 1, avp.Some([]byte{0, 255, 0, 2}), 8); !errors.Is(err, validation.ErrContext) {
					t.Fatal("token mismatch", err)
				}
				if len(tc.token) > 0 {
					if _, err = ro.RefundInformationForService(q, v, 1, len(tc.token)-1); !errors.Is(err, wire.ErrLimit) {
						t.Fatal(err)
					}
					got.Value[0] = 77
					if v.MultipleServicesCreditControl[0].RefundInformation.Value[0] != 0 {
						t.Fatal("output aliases answer")
					}
					v.MultipleServicesCreditControl[0].RefundInformation.Value[0] = 88
					if got.Value[0] != 77 {
						t.Fatal("answer aliases output")
					}
				}
			} else {
				v.MultipleServicesCreditControl[0].RefundInformation.Value = []byte{1, 2}
				got, err = ro.RefundInformationForService(q, v, 1, 0)
				if err != nil || got.Present || got.Value != nil {
					t.Fatal("absent residual retained", got, err)
				}
				if err = ro.CheckRefundInformationEcho(q, v, 1, avp.Optional[[]byte]{Value: []byte{1}}, 0); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRoSMSRefundRejectsUnknownAndUnrelatedEvidence(t *testing.T) {
	c := roCodec(t)
	for _, tc := range []struct {
		name   string
		change func(**ro.CCR, **ro.CCA, *uint32, *int)
		want   error
	}{
		{"unknown-request", func(q **ro.CCR, _ **ro.CCA, _ *uint32, _ *int) { *q = nil }, validation.ErrContextRequired},
		{"unknown-answer", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { *a = nil }, validation.ErrContextRequired},
		{"unknown-parent", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).ResultCode = 4011
			(*a).MultipleServicesCreditControl = nil
		}, validation.ErrContextRequired},
		{"negative-bound", func(_ **ro.CCR, _ **ro.CCA, _ *uint32, n *int) { *n = -1 }, wire.ErrConfig},
		{"over-wire-bound", func(_ **ro.CCR, _ **ro.CCA, _ *uint32, n *int) { *n = wire.MaxWireSize + 1 }, wire.ErrConfig},
		{"zero-bound", func(_ **ro.CCR, _ **ro.CCA, _ *uint32, n *int) { *n = 0 }, wire.ErrLimit},
		{"other-selected-service", func(_ **ro.CCR, _ **ro.CCA, s *uint32, _ *int) { *s = 2 }, validation.ErrContext},
		{"other-answer-service", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).MultipleServicesCreditControl[0].ServiceIdentifier[0] = 2
		}, validation.ErrInvalidAnswer},
		{"duplicate-answer-parent", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).MultipleServicesCreditControl = append((*a).MultipleServicesCreditControl, (*a).MultipleServicesCreditControl[0])
		}, validation.ErrUnsupported},
		{"duplicate-request-parent", func(q **ro.CCR, _ **ro.CCA, _ *uint32, _ *int) {
			(*q).MultipleServicesCreditControl = append((*q).MultipleServicesCreditControl, (*q).MultipleServicesCreditControl[0])
		}, validation.ErrUnsupported},
		{"no-service-key", func(q **ro.CCR, _ **ro.CCA, _ *uint32, _ *int) {
			(*q).MultipleServicesCreditControl[0].ServiceIdentifier = nil
		}, validation.ErrUnsupported},
		{"session", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { (*a).SessionID += "different" }, validation.ErrInvalidAnswer},
		{"number", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { (*a).CCRequestNumber++ }, validation.ErrInvalidAnswer},
		{"type", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { (*a).CCRequestType = ro.InitialRequest }, validation.ErrInvalidAnswer},
		{"other-host", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { (*a).OriginHost = "other.example" }, validation.ErrSource},
		{"other-realm", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) { (*a).OriginRealm = "other.example" }, validation.ErrSource},
		{"realm-only", func(q **ro.CCR, _ **ro.CCA, _ *uint32, _ *int) { (*q).DestinationHost.Present = false }, validation.ErrContextRequired},
		{"proxy", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "proxy.example", ProxyState: []byte{1}}}
		}, validation.ErrInvalidAnswer},
		{"unknown-child-result", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).MultipleServicesCreditControl[0].ResultCode.Present = false
		}, validation.ErrUnsupported},
		{"features", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).SupportedFeatures = []ro.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 1}}
		}, validation.ErrUnsupported},
		{"experimental-conflict", func(_ **ro.CCR, a **ro.CCA, _ *uint32, _ *int) {
			(*a).ExperimentalResult = avp.Some(ro.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5011})
		}, validation.ErrResultConflict},
		{"refund-not-original", func(q **ro.CCR, _ **ro.CCA, _ *uint32, _ *int) { (*q).RequestedAction.Value = 1 }, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := decoded(t, c, "v-iec-ccr-01").Body.(*ro.CCR)
			a := decoded(t, c, "b71-token-binary").Body.(*ro.CCA)
			s, n := uint32(1), 64
			tc.change(&q, &a, &s, &n)
			got, err := ro.RefundInformationForService(q, a, s, n)
			if !errors.Is(err, tc.want) || got.Present || got.Value != nil {
				t.Fatal(got, err, "want", tc.want)
			}
			if err = ro.CheckRefundInformationEcho(q, a, s, avp.Some([]byte{0, 255, 0, 1}), n); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	q := decoded(t, c, "v-iec-ccr-01").Body.(*ro.CCR)
	a := decoded(t, c, "b71-token-binary").Body.(*ro.CCA)
	// A diagnostic copy of 2022 is not the selected service's value.
	diagnostic := decoded(t, c, "b71-token-absent").Body.(*ro.CCA)
	diagnostic.ResultCode = 4010
	diagnostic.FailedAVP = avp.Some(pdu.FailedAVP{Data: []byte{0, 0, 7, 230, 192, 0, 0, 14, 0, 0, 40, 175, 0, 255, 0, 0}})
	if leaf, err := ro.RefundInformationForService(q, diagnostic, 1, 4); err != nil || leaf.Present {
		t.Fatal("diagnostic became a service token", leaf, err)
	}
	q.DestinationHost.Value = strings.ToUpper(q.DestinationHost.Value)
	q.DestinationRealm = strings.ToUpper(q.DestinationRealm)
	if _, err := ro.RefundInformationForService(q, a, 1, wire.MaxWireSize); err != nil {
		t.Fatal(err)
	}
	if err := ro.CheckRefundInformationEcho(q, a, 1, avp.Some(make([]byte, 5)), 4); !errors.Is(err, wire.ErrLimit) {
		t.Fatal(err)
	}
	initial := decoded(t, c, "v-sms-ccr-01").Body.(*ro.CCR)
	if _, err := ro.RefundInformationForService(initial, a, 1, 4); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("not original IEC", err)
	}
	// Body-level extraction cannot replace common header/source association.
	r, answer := decoded(t, c, "v-iec-ccr-01"), decoded(t, c, "b71-token-binary")
	answer.Header.EndToEnd++
	if err := c.ValidateAnswer(r, answer, roSources()); !errors.Is(err, validation.ErrUnmatchedAnswer) {
		t.Fatal(err)
	}
	refund := decoded(t, c, "v-iec-ccr-01")
	refund.Body.(*ro.CCR).RequestedAction.Value = 1
	if _, err := c.Append(nil, refund); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("SMS refund send unexpectedly enabled", err)
	}
}
