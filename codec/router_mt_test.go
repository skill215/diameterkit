package codec_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func routerMTEvidence(r pdu.Message) sgd.RetransmissionEvidence {
	e := retransmissionEvidence(r, 0)
	e.Target = sgd.RetransmissionToRouter
	e.RouterAlertSupported = avp.Some(true)
	return e
}
func TestRouterMTPairs(t *testing.T) {
	c := sgdCodec(t)
	for _, rn := range []string{"r", "mme-r", "sgsn-r"} {
		t.Run(rn, func(t *testing.T) {
			r := decoded(t, c, "b75-router-mt-"+rn)
			e := routerMTEvidence(r)
			ctx := retransmissionContext(t, e)
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
			if v := c.ValidateContext(r, ctx); !v.Complete {
				t.Fatal(v)
			}
			if v := c.ValidateContext(r, validation.Context{}); v.State != validation.NeedsApplication {
				t.Fatal(v)
			}
			for _, n := range []string{"wait", "numeric", "unicode", "decomposed", "max", "unknown", "zero", "success", "failure", "e"} {
				t.Run(n, func(t *testing.T) {
					a := decoded(t, c, "b75-router-mt-"+n)
					for _, v := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
						if !v.Complete {
							t.Fatal(v)
						}
					}
					raw := encoded(t, c, a)
					b, err := c.Decode(raw)
					if err != nil {
						t.Fatal(err)
					}
					clear(raw)
					if !reflect.DeepEqual(a, b) {
						t.Fatal("not owned or semantic round trip")
					}
					sz, err := c.Measure(a)
					if err != nil {
						t.Fatal(err)
					}
					if bound, err := c.DecodeBound(a.Header, sz); err != nil || bound < uint64(sz) {
						t.Fatal(bound, err)
					}
					a.Header.EndToEnd++
					if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
						t.Fatal(err)
					}
				})
			}
			for _, tc := range []struct {
				n     string
				state validation.State
				send  error
			}{{"equal", validation.NeedsApplication, validation.ErrUnsupported}, {"late", validation.Rejected, validation.ErrField}, {"missing-token", validation.Rejected, validation.ErrMissing}} {
				a := decoded(t, c, "b75-router-mt-"+tc.n)
				if v := c.ValidateAnswerContext(r, a, src, ctx); v.State != tc.state {
					t.Fatal(tc.n, v)
				}
				if v := c.ValidateResponseContext(r, a, src, ctx); !errors.Is(v.Err(), tc.send) {
					t.Fatal(tc.n, v)
				}
			}
		})
	}
}
func TestRouterMTWireRejections(t *testing.T) {
	c := sgdCodec(t)
	for _, n := range []string{"both-r", "noaddress-r", "long", "empty", "nul", "invalid-utf8", "missing-name", "msisdn", "external", "lmsi", "duplicate-name", "duplicate-token", "notime"} {
		t.Run(n, func(t *testing.T) {
			if _, err := c.Decode(fixture(t, "b75-router-mt-"+n)); err == nil {
				t.Fatal("accepted malformed or unsupported wire")
			}
		})
	}
	a := decoded(t, c, "b75-router-mt-wait")
	a.Body.(*sgd.TFA).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3102}, Flags: 192}}
	if err := c.Validate(a); !errors.Is(err, validation.ErrDuplicate) {
		t.Fatal(err)
	}
	a = decoded(t, c, "b75-router-mt-wait")
	a.Body.(*sgd.TFA).ExperimentalResult.Present = false
	a.Body.(*sgd.TFA).ResultCode = avp.Some(uint32(2001))
	if err := c.Validate(a); err == nil {
		t.Fatal("success with time")
	}
}
func TestRouterMTFactsAndPurpose(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "b75-router-mt-r")
	a := decoded(t, c, "b75-router-mt-wait")
	base := routerMTEvidence(r)
	src := pdu.AnswerSources{Terminals: []pdu.Origin{base.Responder}}
	// An absent Optional's stale value must not change frozen context identity.
	absent := base
	absent.RouterAlertSupported = avp.Optional[bool]{Value: true}
	zero := absent
	zero.RouterAlertSupported.Value = false
	if !reflect.DeepEqual(retransmissionContext(t, absent), retransmissionContext(t, zero)) {
		t.Fatal("absent Alert capability retained stale value")
	}
	for _, tc := range []struct {
		n        string
		edit     func(*sgd.RetransmissionEvidence)
		req, ans validation.State
	}{
		{"alert unknown", func(e *sgd.RetransmissionEvidence) { e.RouterAlertSupported.Present = false }, validation.NeedsApplication, validation.NeedsApplication},
		{"alert false", func(e *sgd.RetransmissionEvidence) { e.RouterAlertSupported.Value = false }, validation.Rejected, validation.Rejected},
		{"PLMN unknown", func(e *sgd.RetransmissionEvidence) { e.SamePLMN.Present = false }, validation.NeedsApplication, validation.NeedsApplication},
		{"PLMN false", func(e *sgd.RetransmissionEvidence) { e.SamePLMN.Value = false }, validation.Rejected, validation.Rejected},
		{"power unknown", func(e *sgd.RetransmissionEvidence) { e.PowerSaving.Present = false }, validation.Ready, validation.NeedsApplication},
		{"power false", func(e *sgd.RetransmissionEvidence) { e.PowerSaving.Value = false }, validation.Ready, validation.Rejected},
		{"era unknown", func(e *sgd.RetransmissionEvidence) { e.RequestedEra.Present = false }, validation.Ready, validation.NeedsApplication},
		{"wrong endpoint", func(e *sgd.RetransmissionEvidence) { e.Responder.Host = "dra.example" }, validation.Rejected, validation.Rejected},
	} {
		t.Run(tc.n, func(t *testing.T) {
			e := base
			tc.edit(&e)
			ctx := retransmissionContext(t, e)
			if v := c.ValidateContext(r, ctx); v.State != tc.req {
				t.Fatal(v)
			}
			if v := c.ValidateAnswerContext(r, a, src, ctx); v.State != tc.ans {
				t.Fatal(v)
			}
		})
	}
	for _, target := range []sgd.RetransmissionTarget{0, sgd.RetransmissionToMME, sgd.RetransmissionToSGSN} {
		e := base
		e.Target = target
		if _, err := sgd.Retransmission(e); !errors.Is(err, validation.ErrContext) {
			t.Fatal(target, err)
		}
	}
	for _, rn := range []string{"r", "mme-r"} {
		r := decoded(t, c, "b75-router-mt-"+rn)
		e := base
		e.Target = sgd.RetransmissionToMME
		e.RouterAlertSupported.Present = false
		ctx := retransmissionContext(t, e)
		if v := c.ValidateAnswerContext(r, a, src, ctx); v.Complete {
			t.Fatal("direct target accepted Router token")
		}
		if rn == "r" {
			if v := c.ValidateContext(r, ctx); v.Complete {
				t.Fatal("direct target without access number")
			}
		}
	}
	// No maximum needs no power/PLMN/Alert assertion, but missing access number still needs Router purpose.
	r = decoded(t, c, "b75-router-mt-nomax-r")
	e := base
	e.SamePLMN.Present = false
	e.RouterAlertSupported.Present = false
	e.PowerSaving.Present = false
	e.RequestedEra.Present = false
	ctx := retransmissionContext(t, e)
	if v := c.ValidateContext(r, ctx); !v.Complete {
		t.Fatal(v)
	}
	if v := c.ValidateContext(r, validation.Context{}); v.State != validation.NeedsApplication {
		t.Fatal(v)
	}
	if v := c.ValidateAnswerContext(r, a, src, ctx); v.Complete {
		t.Fatal("time without original max")
	}
	if v := c.ValidateAnswerContext(r, decoded(t, c, "b75-router-mt-success"), src, ctx); !v.Complete {
		t.Fatal(v)
	}
	if n := reflect.TypeFor[validation.Context]().Size() + 4*253 + 3*15 + 253; n > 2048 {
		t.Fatal("fixed context budget", n)
	}
}
func TestRouterMTTokenToAlert(t *testing.T) {
	c := sgdCodec(t)
	r := decoded(t, c, "b75-router-mt-r")
	a := decoded(t, c, "b75-router-mt-unicode")
	e := routerMTEvidence(r)
	if v := c.ValidateAnswerContext(r, a, pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}, retransmissionContext(t, e)); !v.Complete {
		t.Fatal(v)
	}
	token := a.Body.(*sgd.TFA).UserIdentifier.Value.UserName
	sc, err := avp.DecodeTBCD(r.Body.(*sgd.TFR).SMSGMSCAddress.Value)
	if err != nil {
		t.Fatal(err)
	}
	alert := decoded(t, c, "b74-router-reachable")
	b := alert.Body.(*s6c.ALR)
	b.OriginHost = e.Responder.Host
	b.OriginRealm = e.Responder.Realm
	b.DestinationHost = e.Requester.Host
	b.DestinationRealm = e.Requester.Realm
	b.UserIdentifier.UserName = avp.Some(token)
	b.SCAddress = bytes.Clone(r.Body.(*sgd.TFR).SMSGMSCAddress.Value)
	ctx := routerContext(t, s6c.RouterAlertEvidence{Router: e.Responder, GMSC: e.Requester, PriorAlertToken: token, PriorGMSCAddress: sc})
	if v := c.ValidateContext(alert, ctx); !v.Complete {
		t.Fatal(v)
	}
	b.UserIdentifier.UserName = avp.Some("令牌-e\u0301")
	if v := c.ValidateContext(alert, ctx); v.Complete {
		t.Fatal("token normalized")
	}
}
