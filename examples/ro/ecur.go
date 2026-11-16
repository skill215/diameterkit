package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
)

// Completed endpoint fields come from the explicit example configuration, not
// the adjacent peer's advertised business role or a hidden session lookup.
func newInitial(n *diameter.Node, ctf, ocs pdu.Origin) (*ro.CCR, error) {
	r, e := newDebit(n)
	if e != nil {
		return nil, e
	}
	r.CCRequestType = ro.InitialRequest
	r.RequestedAction = avp.Optional[uint32]{}
	r.OriginHost, r.OriginRealm = ctf.Host, ctf.Realm
	r.DestinationHost, r.DestinationRealm = avp.Some(ocs.Host), ocs.Realm
	return r, nil
}

// used is an application-selected synthetic measurement. This constructor does
// not measure SMS delivery, consume a grant, choose a refund or allocate a new
// Session-Id. Fresh owned groups avoid aliasing the original request.
func newTermination(initial *ro.CCR, used uint64) (*ro.CCR, error) {
	if initial == nil || len(initial.MultipleServicesCreditControl) != 1 || len(initial.MultipleServicesCreditControl[0].ServiceIdentifier) != 1 {
		return nil, fmt.Errorf("missing initial service")
	}
	r := &ro.CCR{SessionID: initial.SessionID, OriginHost: initial.OriginHost, OriginRealm: initial.OriginRealm,
		DestinationHost: initial.DestinationHost, DestinationRealm: initial.DestinationRealm, AuthApplicationID: ro.ApplicationID,
		ServiceContextID: initial.ServiceContextID, CCRequestType: ro.TerminationRequest, CCRequestNumber: 1, MultipleServicesIndicator: avp.Some(uint32(1)),
		MultipleServicesCreditControl: []ro.RequestMSCC{{ServiceIdentifier: []uint32{initial.MultipleServicesCreditControl[0].ServiceIdentifier[0]}, UsedServiceUnit: []ro.ServiceUnit{{CCServiceSpecificUnits: avp.Some(used)}}, ReportingReason: []uint32{ro.ReportingFinal}}},
		ServiceInformation:            avp.Some(ro.ServiceInformation{SMSInformation: avp.Some(ro.SMSInformation{SMSNode: avp.Some(uint32(1)), SMMessageType: avp.Some(uint32(0))})})}
	if e := ro.CheckECURContinuation(initial, r); e != nil {
		return nil, e
	}
	return r, nil
}

func termResultName(result string) string {
	switch result {
	case "2001":
		return "base:2001"
	case "5002":
		return "base:5002"
	case "5012":
		return "base:5012"
	case "3004":
		return "E:3004"
	default:
		return ""
	}
}
func newECURAnswer(term bool, result string) (pdu.Body, error) {
	if term {
		if termResultName(result) == "" {
			return nil, fmt.Errorf("unsupported termination result")
		}
		if result == "3004" {
			return &pdu.ErrorAnswer{ResultCode: 3004}, nil
		}
		code := uint32(2001)
		if result == "5002" {
			code = 5002
		}
		if result == "5012" {
			code = 5012
		}
		return &ro.CCA{AuthApplicationID: ro.ApplicationID, CCRequestType: ro.TerminationRequest, CCRequestNumber: 1, ResultCode: code}, nil
	}
	if resultName(result) == "" {
		return nil, fmt.Errorf("unsupported initial result")
	}
	body, e := newAnswer(result)
	if e != nil {
		return nil, e
	}
	if a, ok := body.(*ro.CCA); ok {
		a.CCRequestType = ro.InitialRequest
		for i := range a.MultipleServicesCreditControl {
			a.MultipleServicesCreditControl[i].RefundInformation = avp.Optional[[]byte]{}
		}
	}
	return body, nil
}

// This is example application policy AFTER validated request/answer association.
// Neither this predicate nor CheckECURContinuation establishes delivery/accounting
// truth. An incomplete/failed/uncertain exchange must never reach this decision.
func canTerminate(answer pdu.Body, used uint64) bool {
	a, ok := answer.(*ro.CCA)
	if !ok || a.CCRequestType != ro.InitialRequest || a.CCRequestNumber != 0 || a.ResultCode != 2001 || a.ExperimentalResult.Present || len(a.MultipleServicesCreditControl) != 1 {
		return false
	}
	s := a.MultipleServicesCreditControl[0]
	return len(s.ServiceIdentifier) == 1 && s.ServiceIdentifier[0] == 1 && s.ResultCode.Present && s.ResultCode.Value == 2001 && s.GrantedServiceUnit.Present && s.GrantedServiceUnit.Value.CCServiceSpecificUnits.Present && s.GrantedServiceUnit.Value.CCServiceSpecificUnits.Value >= used
}

type ecurReceipt struct {
	Header pdu.Header
	ID     diameter.RequestID
	Answer pdu.Body
}

func exchangeECUR(ctx context.Context, p *pair, request, initial *ro.CCR, result string) (ecurReceipt, error) {
	var receipt ecurReceipt
	term := request.CCRequestType == ro.TerminationRequest
	if term {
		if e := ro.CheckECURContinuation(initial, request); e != nil {
			return receipt, e
		}
	}
	id, e := p.left.peer.Submit(ctx, request, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		return receipt, e
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		return receipt, e
	}
	if !in.Validation().Complete {
		return receipt, fmt.Errorf("incomplete inbound ECUR validation")
	}
	message := in.Message()
	received, ok := message.Body.(*ro.CCR)
	if !ok || received.SessionID != request.SessionID || received.CCRequestType != request.CCRequestType || received.CCRequestNumber != request.CCRequestNumber {
		return receipt, fmt.Errorf("ECUR request identity mismatch")
	}
	if term {
		if e := ro.CheckECURContinuation(initial, received); e != nil {
			return receipt, e
		}
	}
	answer, e := newECURAnswer(term, result)
	if e != nil {
		return receipt, e
	}
	if r, e := in.Respond(ctx, answer); e != nil || !r.Committed {
		return receipt, errors.Join(e, fmt.Errorf("ECUR answer not settled"))
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil {
		return receipt, e
	}
	if c.RequestID != id || c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
		return receipt, fmt.Errorf("ECUR completion: %+v", c.Result)
	}
	if a, ok := c.Result.Answer.Body.(*ro.CCA); ok {
		if a.SessionID != request.SessionID || a.CCRequestType != request.CCRequestType || a.CCRequestNumber != request.CCRequestNumber {
			return receipt, fmt.Errorf("ECUR answer identity mismatch")
		}
		// Compare the explicitly selected synthetic result, without merging levels.
		expected, ok := answer.(*ro.CCA)
		if !ok {
			return receipt, fmt.Errorf("expected ECUR E answer")
		}
		if a.ResultCode != expected.ResultCode || a.ExperimentalResult.Present || len(a.MultipleServicesCreditControl) != len(expected.MultipleServicesCreditControl) {
			return receipt, fmt.Errorf("ECUR result mismatch")
		}
		if len(a.MultipleServicesCreditControl) == 1 {
			s, want := a.MultipleServicesCreditControl[0], expected.MultipleServicesCreditControl[0]
			if len(s.ServiceIdentifier) != 1 || s.ServiceIdentifier[0] != 1 || s.ResultCode != want.ResultCode || s.RefundInformation.Present || s.GrantedServiceUnit.Present != want.GrantedServiceUnit.Present {
				return receipt, fmt.Errorf("ECUR service mismatch")
			}
			if s.GrantedServiceUnit.Present && s.GrantedServiceUnit.Value.CCServiceSpecificUnits != want.GrantedServiceUnit.Value.CCServiceSpecificUnits {
				return receipt, fmt.Errorf("ECUR grant mismatch")
			}
		}
	} else if a, ok := c.Result.Answer.Body.(*pdu.ErrorAnswer); !ok || result != "3004" || a.ResultCode != 3004 || !a.SessionID.Present || a.SessionID.Value != request.SessionID {
		return receipt, fmt.Errorf("ECUR E result mismatch")
	}
	return ecurReceipt{Header: message.Header, ID: id, Answer: c.Result.Answer.Body}, nil
}

type ecurReport struct {
	Initial, Termination string
	TermSent             bool
	Left, Right          diameter.ResourceSnapshot
}

func runECUR(ctx context.Context, initialResult, termResult string, out io.Writer) (r ecurReport, err error) {
	if resultName(initialResult) == "" || termResultName(termResult) == "" {
		return r, fmt.Errorf("invalid ECUR result selection")
	}
	p, err := openPair(ctx)
	if err != nil {
		return r, err
	}
	defer func() {
		err = errors.Join(err, p.close())
		r.Left, r.Right = p.left.node.Resources(), p.right.node.Resources()
		if r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
			err = errors.Join(err, fmt.Errorf("resources retained after ECUR cleanup"))
		}
	}()
	initial, err := newInitial(p.left.node, p.left.origin, p.right.origin)
	if err != nil {
		return r, err
	}
	first, err := exchangeECUR(ctx, p, initial, nil, initialResult)
	if err != nil {
		return r, err
	}
	r.Initial = resultName(initialResult)
	r.Termination = "not-sent"
	// Explicit example application choice; no hidden Node-owned business session.
	const used uint64 = 1
	if canTerminate(first.Answer, used) {
		term, e := newTermination(initial, used)
		if e != nil {
			return r, e
		}
		last, e := exchangeECUR(ctx, p, term, initial, termResult)
		if e != nil {
			return r, e
		}
		if first.ID == last.ID || first.Header.HopByHop == last.Header.HopByHop || first.Header.EndToEnd == last.Header.EndToEnd {
			return r, fmt.Errorf("TERMINATION reused INITIAL exchange identity")
		}
		r.TermSent = true
		r.Termination = termResultName(termResult)
	}
	if err = p.left.node.Shutdown(ctx); err != nil {
		return r, err
	}
	if err = p.right.node.Shutdown(ctx); err != nil {
		return r, err
	}
	_, err = fmt.Fprintf(out, "ECUR INITIAL=%s; TERMINATION=%s\n", r.Initial, r.Termination)
	return r, err
}
