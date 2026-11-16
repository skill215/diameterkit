package main

import (
	"errors"
	"reflect"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/compat/smsfwro/legacy420"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

// The three mutually exclusive 4/272 bindings are selected before Node creation.
// They are not inferred from AVPs, and a recovery record cannot change binding.
func compatProfile(binding string) (schema.Profile, error) {
	switch binding {
	case "combined":
		return combined.Profile()
	case "smsfwro":
		return smsfwro.Profile()
	case "legacy420":
		return legacy420.Profile()
	}
	return schema.Profile{}, diameter.ErrConfig
}

func validBoundCompat(binding, action, scenario string) bool {
	if scenario != "error" && scenario != "restart" && scenario != "no-answer" && scenario != "no-answer-import" {
		return false
	}
	switch binding {
	case "combined":
		return validCompatOptions("compat", action)
	case "smsfwro":
		return action == "debit" || action == "initial" || action == "term"
	case "legacy420":
		return action == "debit" || action == "balance" || action == "refund"
	}
	return false
}

func validCompatSelection(flow, binding, action, scenario string) bool {
	if flow != "compat" {
		return binding == "combined"
	}
	return validBoundCompat(binding, action, scenario)
}

func boundCompatRequest(binding, action, session string) (pdu.Body, error) {
	if !validBoundCompat(binding, action, "error") {
		return nil, diameter.ErrConfig
	}
	if binding == "combined" {
		return compatRequest(action, session)
	}
	if binding == "smsfwro" {
		ec := smsfwro.ECURInput{SessionID: session, ServiceContextID: "legacy@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: "15551234567", RecipientMSISDN: "15557654321", SMSKey: "synthetic-1", ProtocolID: []byte{0, 255}}
		switch action {
		case "debit":
			return smsfwro.NewIEC(smsfwro.IECInput{SessionID: session, ServiceContextID: ec.ServiceContextID, SubscriptionMSISDN: ec.SubscriptionMSISDN, SubscriptionIMSI: ec.SubscriptionIMSI, SMSCAddress: ec.SMSCAddress, OriginatorSCCPAddress: "15550000002", RecipientSCCPAddress: ec.RecipientSCCPAddress, OriginatorMSISDN: ec.OriginatorMSISDN, RecipientMSISDN: ec.RecipientMSISDN, SMSKey: ec.SMSKey, DestinationInterface: 2, ProtocolID: ec.ProtocolID}), nil
		case "initial":
			return smsfwro.NewInitial(ec), nil
		case "term":
			return smsfwro.NewTermination(smsfwro.TerminationInput{ECURInput: ec, UsedServiceSpecificUnits: 0, Status: []byte{0, 255}}), nil
		}
	}
	v := legacy420.GeneralInput{SessionID: session, ServiceContextID: "legacy@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", OriginatorMSISDN: "15551234567", ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 0, ProtocolID: []byte{0, 255}}
	switch action {
	case "debit":
		return legacy420.NewDirectDebit(v), nil
	case "balance":
		return legacy420.NewBalanceCheck(v), nil
	case "refund":
		return legacy420.NewRefund(legacy420.RefundInput{SessionID: session, ServiceContextID: v.ServiceContextID, SubscriptionMSISDN: v.SubscriptionMSISDN, SubscriptionIMSI: v.SubscriptionIMSI, ValueDigits: v.ValueDigits, Exponent: v.Exponent, CurrencyCode: v.CurrencyCode, ProtocolID: v.ProtocolID, RefundInformation: avp.Some([]byte{0, 255})})
	}
	return nil, diameter.ErrConfig
}

func boundCompatAnswer(q pdu.Body, code uint32) pdu.Body {
	if code == 3004 {
		return &pdu.ErrorAnswer{ResultCode: code}
	}
	switch q := q.(type) {
	case *combined.CCR:
		return compatAnswer(q, code)
	case *smsfwro.CCR:
		return &smsfwro.CCA{AuthApplicationID: 4, CCRequestType: q.CCRequestType, CCRequestNumber: q.CCRequestNumber, ResultCode: code}
	case *legacy420.CCR:
		a := &legacy420.CCA{AuthApplicationID: 4, CCRequestType: q.CCRequestType, CCRequestNumber: q.CCRequestNumber, ResultCode: code}
		if code == 2001 && q.RequestedAction == 2 {
			a.CheckBalanceResult = avp.Some(uint32(1))
		}
		return a
	}
	return nil
}

func checkCompatAnswer(binding, action string, code uint32, body pdu.Body) error {
	var actual uint32
	var balance avp.Optional[uint32]
	matched := false
	switch a := body.(type) {
	case *combined.CCA:
		actual, balance, matched = a.ResultCode, a.CheckBalanceResult, binding == "combined" && code != 3004
	case *smsfwro.CCA:
		actual, matched = a.ResultCode, binding == "smsfwro" && code != 3004
	case *legacy420.CCA:
		actual, balance, matched = a.ResultCode, a.CheckBalanceResult, binding == "legacy420" && code != 3004
	case *pdu.ErrorAnswer:
		actual, matched = a.ResultCode, code == 3004
	}
	if !matched || actual != code {
		return errors.New("answer binding or result changed")
	}
	if action == "balance" && code == 2001 && (!balance.Present || balance.Value != 1) {
		return errors.New("NO_CREDIT lost")
	}
	return nil
}

// Normalize only the two explicitly authorized destination fields in a copy.
// All other fields (including opaque bytes and presence) must remain identical.
func sameCompatRequest(first, second pdu.Body) bool {
	var normalized pdu.Body
	var host avp.Optional[string]
	var realm string
	switch a := first.(type) {
	case *combined.CCR:
		b, ok := second.(*combined.CCR)
		if !ok {
			return false
		}
		v := *b
		host, realm = v.DestinationHost, v.DestinationRealm
		v.DestinationHost, v.DestinationRealm = a.DestinationHost, a.DestinationRealm
		normalized = &v
	case *smsfwro.CCR:
		b, ok := second.(*smsfwro.CCR)
		if !ok {
			return false
		}
		v := *b
		host, realm = v.DestinationHost, v.DestinationRealm
		v.DestinationHost, v.DestinationRealm = a.DestinationHost, a.DestinationRealm
		normalized = &v
	case *legacy420.CCR:
		b, ok := second.(*legacy420.CCR)
		if !ok {
			return false
		}
		v := *b
		host, realm = v.DestinationHost, v.DestinationRealm
		v.DestinationHost, v.DestinationRealm = a.DestinationHost, a.DestinationRealm
		normalized = &v
	}
	return host.Present && host.Value == backupOrigin.Host && realm == backupOrigin.Realm && reflect.DeepEqual(first, normalized)
}

func closeCompatTargets(primary, backup *pair) error {
	err := errors.Join(primary.close(), backup.close())
	for _, n := range []*diameter.Node{primary.left.node, primary.right.node, backup.right.node} {
		if n.Resources() != (diameter.ResourceSnapshot{}) {
			err = errors.Join(err, errors.New("compat resources retained"))
		}
	}
	return err
}
