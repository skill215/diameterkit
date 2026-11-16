// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package s6c

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

type RoutingOutcome uint8

const (
	RoutingUnsupported RoutingOutcome = iota
	RoutingSuccess
	RoutingRejected
)

type ResultNamespace uint8

const (
	BaseResult ResultNamespace = iota + 1
	ExperimentalResultNamespace
	ErrorResult
)

type RoutingProtocolResult struct {
	Namespace      ResultNamespace
	VendorID, Code uint32
}
type IdentityKind uint8

const (
	IdentityUnresolved IdentityKind = iota
	IdentityIMSI
	IdentityRouterToken
	IdentityMCCMNC
)

type SubscriberIdentity struct {
	Kind  IdentityKind
	Value avp.Optional[string]
}
type NodeKind uint8

const (
	NodeMME NodeKind = iota + 1
	NodeMSC
	NodeSGSN
	NodeIPSMGW
)

type CandidateSource uint8

const (
	PrimaryNode CandidateSource = iota + 1
	AdditionalNode
)

type AccessKind uint8

const (
	NeedsResolution AccessKind = iota
	DiameterAccess
	MAPAccess
	SBIAccess
)

type CompanionNode struct {
	Kind        NodeKind
	Host, Realm avp.Optional[string]
}

// A combined MSC/MME is one candidate with a Companion, never two routes.
type RouteCandidate struct {
	Kind                NodeKind
	Source              CandidateSource
	Access              AccessKind
	Number, Host, Realm avp.Optional[string]
	Companion           avp.Optional[CompanionNode]
	SBI                 avp.Optional[uint32]
}
type NodeDiagnostic struct {
	Node  NodeKind
	Value uint32
}
type RoutingResult struct {
	Outcome     RoutingOutcome
	Protocol    RoutingProtocolResult
	Identity    SubscriberIdentity
	Candidates  []RouteCandidate
	Diagnostics []NodeDiagnostic
	MWDStatus   avp.Optional[uint32]
	Issues      []error
	// Answer owns all selected and raw fields. No candidate grants retry.
	Answer pdu.Body
}
type RoutingContext struct {
	Request    *SRR
	Validation validation.Context
}

// InterpretSRA applies the current fixed Profile's body, pair and context rules.
// First validate header correlation and source authority with the codec/runtime.
// Named evidence requires explicit request endpoints. Unsupported/context-missing
// answers retain evidence but expose no routes. This function publishes nothing.
func InterpretSRA(c RoutingContext, answer pdu.Body) (RoutingResult, error) {
	var out RoutingResult
	if err := validateSRR(c.Request, validation.Build); err != nil {
		return out, err
	}
	if a, ok := answer.(*pdu.ErrorAnswer); ok {
		ops := base.NewErrorOperations(pdu.CommandKey{ApplicationID: 16777312, Code: 8388647}, base.ErrorOptions{DiagnosticGroups: profileDiagnosticGroups(), AuthSessionState: avp.Some(uint32(1)), IgnoreVSAI: true, UnsupportedFields: profileErrorUnsupported()})
		if err := ops.Validate(a, validation.Receive); err != nil {
			return out, err
		}
		if !a.SessionID.Present || a.SessionID.Value != c.Request.SessionID || !pdu.EqualProxies(c.Request.ProxyInfo, a.ProxyInfo) {
			return out, validation.ErrInvalidAnswer
		}
		out.Answer, _ = ops.Clone(a)
		out.Outcome, out.Protocol = RoutingRejected, RoutingProtocolResult{Namespace: ErrorResult, Code: a.ResultCode}
		return out, nil
	}
	a, ok := answer.(*SRA)
	if !ok || a == nil {
		return out, validation.ErrType
	}
	if a.ResultCode.Present == a.ExperimentalResult.Present {
		if a.ResultCode.Present {
			return out, validation.ErrResultConflict
		}
		return out, validation.ErrMissing
	}
	if err := validateSRA(a, validation.Receive); err != nil {
		return out, err
	}
	if a.SessionID != c.Request.SessionID || !pdu.EqualProxies(c.Request.ProxyInfo, a.ProxyInfo) {
		return out, validation.ErrInvalidAnswer
	}
	if err := checkSRANodeRequest(c.Request, a); err != nil {
		return out, err
	}
	a = cloneSRA(a)
	out.Answer, out.MWDStatus = a, a.MWDStatus
	out.Identity.Value = a.UserName
	if a.ResultCode.Present {
		out.Protocol = RoutingProtocolResult{Namespace: BaseResult, Code: a.ResultCode.Value}
	} else {
		out.Protocol = RoutingProtocolResult{ExperimentalResultNamespace, a.ExperimentalResult.Value.VendorID, a.ExperimentalResult.Value.ExperimentalResultCode}
	}
	for _, d := range []struct {
		kind  NodeKind
		value avp.Optional[uint32]
	}{{NodeMME, a.MMEAbsentUserDiagnosticSM}, {NodeMSC, a.MSCAbsentUserDiagnosticSM}, {NodeSGSN, a.SGSNAbsentUserDiagnosticSM}} {
		if d.value.Present {
			out.Diagnostics = append(out.Diagnostics, NodeDiagnostic{d.kind, d.value.Value})
		}
	}
	if err := checkRoutingContext(c.Request, a, c.Validation, validation.Receive); err != nil {
		if errors.Is(err, validation.ErrContextRequired) || errors.Is(err, validation.ErrUnsupported) {
			out.Issues = []error{err}
			return out, nil
		}
		return out, err
	}
	if a.UserName.Present {
		out.Identity.Kind = IdentityIMSI
		if _, ok := c.Validation.RouterRouting(); ok {
			out.Identity.Kind = IdentityRouterToken
		}
		if _, ok := c.Validation.MCCMNCRouting(); ok {
			out.Identity.Kind = IdentityMCCMNC
		}
	}
	out.Outcome = RoutingRejected
	if out.Protocol.Namespace != BaseResult || out.Protocol.Code != 2001 {
		return out, nil
	}
	out.Outcome = RoutingSuccess
	if a.ServingNode.Present {
		out.Candidates = append(out.Candidates, routingCandidate(a.ServingNode.Value, PrimaryNode))
	}
	if a.AdditionalServingNode.Present {
		n := a.AdditionalServingNode.Value
		out.Candidates = append(out.Candidates, routingCandidate(ServingNode{SGSNName: n.SGSNName, SGSNRealm: n.SGSNRealm, SGSNNumber: n.SGSNNumber, MMEName: n.MMEName, MMERealm: n.MMERealm, MMENumber: n.MMENumber, MSCNumber: n.MSCNumber}, AdditionalNode))
	}
	return out, nil
}

func routingCandidate(n ServingNode, source CandidateSource) RouteCandidate {
	c := RouteCandidate{Source: source}
	var number avp.Optional[[]byte]
	switch {
	case n.IPSMGWNumber.Present:
		c.Kind, c.Host, c.Realm, c.SBI, number = NodeIPSMGW, n.IPSMGWName, n.IPSMGWRealm, n.IPSMGWSBISupportIndicator, n.IPSMGWNumber
		c.Access = NeedsResolution
		if c.SBI.Present && c.SBI.Value == 1 {
			c.Access = SBIAccess
		}
	case n.SGSNNumber.Present:
		c.Kind, c.Host, c.Realm, number = NodeSGSN, n.SGSNName, n.SGSNRealm, n.SGSNNumber
		c.Access = MAPAccess
		if c.Host.Present {
			c.Access = DiameterAccess
		}
	case n.MSCNumber.Present:
		c.Kind, c.Access, number = NodeMSC, MAPAccess, n.MSCNumber
		if n.MMEName.Present {
			c.Companion = avp.Some(CompanionNode{NodeMME, n.MMEName, n.MMERealm})
		}
	default:
		c.Kind, c.Access, c.Host, c.Realm, number = NodeMME, DiameterAccess, n.MMEName, n.MMERealm, n.MMENumber
	}
	if number.Present {
		s, _ := avp.DecodeTBCD(number.Value)
		c.Number = avp.Some(s)
	} // validated above
	return c
}
