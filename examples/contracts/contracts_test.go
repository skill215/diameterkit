// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package contracts_test exercises selected public bindings without repository
// helpers or fixtures. It can be copied unchanged into an external Go module.
package contracts_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/app/ro/dcca"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/compat/sgdalert"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

// These signatures prove accessibility only, not recovery policy or admission.
var (
	_ func(*diameter.Node, context.Context, pdu.Body, diameter.RoRecoveryOptions) (diameter.RoRecoveryRequest, error)                                                               = (*diameter.Node).PrepareRoRecoveryRequest
	_ func(*diameter.Node, context.Context, diameter.RoRecoveryRequest, diameter.RoRecoveryPlan, diameter.RoRecoveryProgress, *diameter.CompletionPort) (diameter.RequestID, error) = (*diameter.Node).SubmitRoRecovery
)

func profileFor(t *testing.T, f func() (schema.Profile, error)) schema.Profile {
	t.Helper()
	p, e := f()
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func codecFor(t *testing.T, f func() (schema.Profile, error)) *codec.Codec {
	t.Helper()
	p := profileFor(t, f)
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, codec.Config{})
	if e != nil {
		t.Fatal(e)
	}
	return c
}

var origin = pdu.Origin{Host: "ctf.example", Realm: "example"}
var target = pdu.Origin{Host: "ocs.example", Realm: "example"}

func requestFor(t *testing.T, c *codec.Codec, b pdu.Body, hbh uint32) pdu.Message {
	t.Helper()
	m, e := c.CompleteRequest(b, pdu.RequestEnvelope{OriginHost: origin.Host, OriginRealm: origin.Realm, DestinationHost: target.Host, DestinationRealm: target.Realm, HopByHop: hbh, EndToEnd: hbh + 100})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := c.Append(nil, m)
	if e != nil {
		t.Fatal(e)
	}
	m, e = c.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func answerFor(t *testing.T, c *codec.Codec, r pdu.Message, b pdu.Body) pdu.Message {
	t.Helper()
	a, e := c.CompleteAnswer(r, b, target)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := c.Append(nil, a)
	if e != nil {
		t.Fatal(e)
	}
	src := pdu.AnswerSources{Terminals: []pdu.Origin{target}}
	a, e = c.DecodeAnswer(r, raw, src)
	if e != nil {
		t.Fatal(e)
	}
	if v := c.ValidateAnswerContext(r, a, src, validation.Context{}); !v.Complete {
		t.Fatal(v)
	}
	return a
}
func TestPublicProfilesAndBindingIsolation(t *testing.T) {
	factories := map[string]func() (schema.Profile, error){
		"base": base.Profile, "sh": sh.Profile, "s6c": s6c.Profile, "sgd": sgd.Profile,
		"ro": ro.Profile, "dcca": dcca.Profile, "abort": abort.Profile,
		"smsfwro": smsfwro.Profile, "legacy420": legacy420.Profile, "combined": combined.Profile,
		"scap1": scap1.Profile, "sgdalert": sgdalert.Profile,
	}
	for name, f := range factories {
		t.Run(name, func(t *testing.T) {
			p := profileFor(t, f)
			if p.Name() == "" || p.Version() == "" || len(p.Digest()) != 64 {
				t.Fatal("missing profile identity")
			}
			b, e := schema.Compile(p)
			if e != nil || len(b.Fingerprint()) != 64 {
				t.Fatal(e)
			}
		})
	}
	names := []string{"ro", "dcca", "smsfwro", "legacy420", "combined"}
	for i, name := range names {
		for _, other := range names[i+1:] {
			t.Run(name+" conflicts "+other, func(t *testing.T) {
				if _, e := schema.Compile(profileFor(t, factories[name]), profileFor(t, factories[other])); !errors.Is(e, schema.ErrDefinition) {
					t.Fatal("same 4/272 inferred by payload", e)
				}
			})
		}
		b, e := schema.Compile(profileFor(t, base.Profile), profileFor(t, factories[name]), profileFor(t, abort.Profile))
		if e != nil {
			t.Fatal(e)
		}
		op, ok := b.Lookup((&abort.ASR{}).CommandKey())
		if !ok || op.TransportReplay {
			t.Fatal("abort gained CCR replay")
		}
	}
}
func dccaRequest(action uint32, session string) *dcca.CCR {
	return &dcca.CCR{SessionID: session, AuthApplicationID: 4, CCRequestType: 4, RequestedAction: action,
		ServiceContextID: "event@example.net", ServiceIdentifier: 7,
		SubscriptionID: []dcca.SubscriptionID{{SubscriptionIDType: 0, SubscriptionIDData: "15550000001"}}}
}
func TestPublicDCCAOperationsAndNewRefund(t *testing.T) {
	c := codecFor(t, dcca.Profile)
	for _, action := range []uint32{dcca.DirectDebiting, dcca.RefundAccount, dcca.CheckBalance} {
		for _, code := range []uint32{2001, 4012, 3004} {
			t.Run(fmt.Sprintf("action%d/result%d", action, code), func(t *testing.T) {
				r := requestFor(t, c, dccaRequest(action, "ctf.example;operation"), 1)
				a := &dcca.CCA{AuthApplicationID: 4, CCRequestType: 4, ResultCode: code}
				if code == 2001 {
					if action == dcca.CheckBalance {
						a.CheckBalanceResult = avp.Some(dcca.NoCredit)
					}
					if action == dcca.DirectDebiting {
						a.GrantedServiceUnit = avp.Some(dcca.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(0))})
					}
				}
				var body pdu.Body = a
				if code == 3004 {
					body = &pdu.ErrorAnswer{ResultCode: 3004}
				}
				result := answerFor(t, c, r, body)
				if action == dcca.CheckBalance && code == 2001 && result.Body.(*dcca.CCA).CheckBalanceResult.Value != dcca.NoCredit {
					t.Fatal("balance success became credit grant")
				}
			})
		}
	}
	debit := requestFor(t, c, dccaRequest(dcca.DirectDebiting, "ctf.example;debit"), 2).Body.(*dcca.CCR)
	refund := requestFor(t, c, dccaRequest(dcca.RefundAccount, "ctf.example;refund"), 3).Body.(*dcca.CCR)
	if e := dcca.CheckRefundRelation(debit, refund); e != nil {
		t.Fatal(e)
	}
	refund.SessionID = debit.SessionID
	if e := dcca.CheckRefundRelation(debit, refund); !errors.Is(e, validation.ErrContext) {
		t.Fatal("refund became retransmission", e)
	}
	// DCCA does not silently accept the standard 3GPP SMS service context.
	invalid := dccaRequest(dcca.CheckBalance, "ctf.example;bad")
	invalid.ServiceContextID = ro.SMSServiceContextID
	if e := c.ValidateBody(invalid); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
func smsDebit() *ro.CCR {
	return &ro.CCR{SessionID: "ctf.example;iec", AuthApplicationID: 4, ServiceContextID: ro.SMSServiceContextID, CCRequestType: ro.EventRequest,
		RequestedAction: avp.Some(ro.DirectDebiting), MultipleServicesIndicator: avp.Some(uint32(1)),
		MultipleServicesCreditControl: []ro.RequestMSCC{{ServiceIdentifier: []uint32{1}, RequestedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))})}},
		ServiceInformation:            avp.Some(ro.ServiceInformation{SMSInformation: avp.Some(ro.SMSInformation{SMSNode: avp.Some(uint32(1)), SMMessageType: avp.Some(uint32(0))})})}
}
func TestPublicSMSRefundTokenBoundary(t *testing.T) {
	c := codecFor(t, ro.Profile)
	for _, v := range []struct {
		name  string
		token avp.Optional[[]byte]
	}{
		{"absent", avp.Optional[[]byte]{}}, {"empty", avp.Some([]byte{})}, {"binary", avp.Some([]byte{0, 255})},
	} {
		t.Run(v.name, func(t *testing.T) {
			r := requestFor(t, c, smsDebit(), 5)
			a := answerFor(t, c, r, &ro.CCA{AuthApplicationID: 4, CCRequestType: ro.EventRequest, ResultCode: 2001,
				MultipleServicesCreditControl: []ro.AnswerMSCC{{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(uint32(2001)), GrantedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))}), RefundInformation: v.token}}})
			q, ans := r.Body.(*ro.CCR), a.Body.(*ro.CCA)
			got, e := ro.RefundInformationForService(q, ans, 1, 16)
			if e != nil || got.Present != v.token.Present || !bytes.Equal(got.Value, v.token.Value) {
				t.Fatal(got, e)
			}
			if e := ro.CheckRefundInformationEcho(q, ans, 1, got, 16); e != nil {
				t.Fatal(e)
			}
			wrong := got
			wrong.Present = !got.Present
			if e := ro.CheckRefundInformationEcho(q, ans, 1, wrong, 16); !errors.Is(e, validation.ErrContext) {
				t.Fatal(e)
			}
			if len(got.Value) > 0 {
				got.Value[0] = 99
				if ans.MultipleServicesCreditControl[0].RefundInformation.Value[0] != 0 {
					t.Fatal("borrowed token")
				}
			}
		})
	}
	if _, e := ro.RefundInformationForService(nil, nil, 1, 16); !errors.Is(e, validation.ErrContextRequired) {
		t.Fatal("unknown treated as absent", e)
	}
	r := smsDebit()
	r.RequestedAction = avp.Some(uint32(1))
	if e := c.ValidateBody(r); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal("2022 helper enabled standard SMS refund", e)
	}
}
func TestPublicLegacyBuilders(t *testing.T) {
	general := legacy420.GeneralInput{SessionID: "ctf.example;legacy", ServiceContextID: "legacy@example.net",
		SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", OriginatorMSISDN: "15551234567",
		ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 0, ProtocolID: []byte{0, 255}}
	for _, binding := range []string{"legacy420", "combined"} {
		t.Run(binding, func(t *testing.T) {
			c := codecFor(t, legacy420.Profile)
			var debit, balance pdu.Body = legacy420.NewDirectDebit(general), legacy420.NewBalanceCheck(general)
			if binding == "combined" {
				c = codecFor(t, combined.Profile)
				debit = combined.NewDirectDebit(general)
				balance = combined.NewBalanceCheck(general)
			}
			requestFor(t, c, debit, 7)
			requestFor(t, c, balance, 8)
			input := legacy420.RefundInput{SessionID: "ctf.example;refund", ServiceContextID: general.ServiceContextID,
				SubscriptionMSISDN: general.SubscriptionMSISDN, SubscriptionIMSI: general.SubscriptionIMSI,
				ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ProtocolID: []byte{0, 255}}
			if binding == "legacy420" {
				if _, e := legacy420.NewRefund(input); !errors.Is(e, validation.ErrMissing) {
					t.Fatal(e)
				}
				input.RefundInformation = avp.Some([]byte{})
				r, e := legacy420.NewRefund(input)
				if e != nil {
					t.Fatal(e)
				}
				requestFor(t, c, r, 9)
				if !r.RefundInformation.Present || r.MultipleServicesCreditControl.RequestedServiceUnit.Money.Value.UnitValue.ValueDigits != -123 {
					t.Fatal("explicit token/money changed")
				}
			} else {
				if _, e := combined.NewRefund(input); !errors.Is(e, validation.ErrMissing) {
					t.Fatal(e)
				}
				input.RefundInformation = avp.Some([]byte{0, 255})
				r, e := combined.NewRefund(input)
				if e != nil {
					t.Fatal(e)
				}
				input.RefundInformation.Value[0] = 99
				requestFor(t, c, r, 9)
				if r.RefundInformation.Value[0] != 0 {
					t.Fatal("borrowed private token")
				}
			}
		})
	}
	ec := smsfwro.ECURInput{SessionID: "ctf.example;ecur", ServiceContextID: "legacy@example.net",
		SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001",
		RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "sms-1", ProtocolID: []byte{0, 255}}
	for _, binding := range []string{"smsfwro", "combined"} {
		t.Run(binding+" ECUR", func(t *testing.T) {
			term := smsfwro.TerminationInput{ECURInput: ec, UsedServiceSpecificUnits: 0, Status: []byte{0, 255}}
			if binding == "smsfwro" {
				c := codecFor(t, smsfwro.Profile)
				i := requestFor(t, c, smsfwro.NewInitial(ec), 10)
				q := requestFor(t, c, smsfwro.NewTermination(term), 11)
				if e := smsfwro.CheckECURContinuation(i.Body.(*smsfwro.CCR), q.Body.(*smsfwro.CCR)); e != nil {
					t.Fatal(e)
				}
			} else {
				c := codecFor(t, combined.Profile)
				i := requestFor(t, c, combined.NewInitial(ec), 10)
				q := requestFor(t, c, combined.NewTermination(term), 11)
				if e := combined.CheckECURContinuation(i.Body.(*combined.CCR), q.Body.(*combined.CCR)); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
func TestPublicRecoveryValuesNotAuthorization(t *testing.T) {
	if e := (diameter.RoRecoveryRequest{}).CheckIntegrity(); !errors.Is(e, diameter.ErrRoRecoveryRecord) {
		t.Fatal(e)
	}
	terminal := diameter.RoRecoveryResult{RecordID: [16]byte{1}, SegmentNo: 2, ReservedThrough: 4}
	checkpoint := terminal.Checkpoint()
	if checkpoint.Unknown || checkpoint.NextTrigger != 0 || checkpoint.ReservedThrough != 4 {
		t.Fatal("terminal became replay permission", checkpoint)
	}
	// Import flags are supplied by the application. This value is NOT submitted.
	claim := diameter.RoRecoveryImport{Checkpoint: checkpoint}
	if claim.SenderFenced || claim.ProgressConfirmed || claim.ClockConfirmed {
		t.Fatal("manufactured historical proof")
	}
}
