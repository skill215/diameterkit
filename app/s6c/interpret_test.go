package s6c_test

import (
	"errors"
	s "github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"strconv"
	"testing"
)

func number(v string) avp.Optional[[]byte] { b, _ := avp.EncodeTBCD(v); return avp.Some(b) }
func request(t *testing.T) *s.SRR {
	t.Helper()
	r, e := s.NewSRR(s.RoutingQuery{SessionID: "test;1", MSISDN: avp.Some("12025550123"), SCAddress: "12025550100", Fields: s.SRR{OriginHost: "gmsc.example", OriginRealm: "example", DestinationHost: avp.Some("hss.example"), DestinationRealm: "example", SRRFlags: avp.Some(uint32(1)), SMRPSMEA: avp.Some([]byte{2, 0x91, 0x21})}})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func answer() *s.SRA {
	return &s.SRA{SessionID: "test;1", AuthSessionState: 1, OriginHost: "hss.example", OriginRealm: "example", ResultCode: avp.Some(uint32(2001)), UserName: avp.Some("001010123456789"), ServingNode: avp.Some(s.ServingNode{MSCNumber: number("12025550199")})}
}
func TestNewSRROwnershipAndErrors(t *testing.T) {
	q := s.RoutingQuery{SessionID: "test;1", MSISDN: avp.Some("12345"), SCAddress: "67890", Fields: s.SRR{SMRPMTI: avp.Some(uint32(0)), SMRPSMEA: avp.Some([]byte{2, 0x91, 0x21})}}
	r, e := s.NewSRR(q)
	if e != nil {
		t.Fatal(e)
	}
	q.Fields.SMRPSMEA.Value[2] = 0xff
	if r.SMRPSMEA.Value[2] != 0x21 || !r.SMRPMTI.Present || r.SMRPMTI.Value != 0 {
		t.Fatal(r)
	}
	for _, bad := range []string{"", "+123", "12x", "1234567890123456"} {
		q.MSISDN = avp.Some(bad)
		if _, e = s.NewSRR(q); e == nil {
			t.Fatal(bad)
		}
	}
	q.MSISDN = avp.Some("123")
	q.Fields.SessionID = "conflict"
	if _, e = s.NewSRR(q); !errors.Is(e, validation.ErrEnvelope) {
		t.Fatal(e)
	}
}
func TestInterpretCandidates(t *testing.T) {
	for _, tc := range []struct {
		name      string
		n         s.ServingNode
		kind      s.NodeKind
		access    s.AccessKind
		companion bool
	}{
		{"msc", s.ServingNode{MSCNumber: number("123")}, s.NodeMSC, s.MAPAccess, false},
		{"combined", s.ServingNode{MSCNumber: number("123"), MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example")}, s.NodeMSC, s.MAPAccess, true},
		{"mme", s.ServingNode{MMENumber: number("123"), MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example")}, s.NodeMME, s.DiameterAccess, false},
		{"sgsn-map", s.ServingNode{SGSNNumber: number("123")}, s.NodeSGSN, s.MAPAccess, false},
		{"sgsn-diameter", s.ServingNode{SGSNNumber: number("123"), SGSNName: avp.Some("sgsn.example"), SGSNRealm: avp.Some("example")}, s.NodeSGSN, s.DiameterAccess, false},
		{"ip-number", s.ServingNode{IPSMGWNumber: number("123")}, s.NodeIPSMGW, s.NeedsResolution, false},
		{"ip-name", s.ServingNode{IPSMGWNumber: number("123"), IPSMGWName: avp.Some("ip.example")}, s.NodeIPSMGW, s.NeedsResolution, false},
		{"sbi-zero", s.ServingNode{IPSMGWNumber: number("123"), IPSMGWSBISupportIndicator: avp.Some(uint32(0))}, s.NodeIPSMGW, s.NeedsResolution, false},
		{"sbi-one", s.ServingNode{IPSMGWNumber: number("123"), IPSMGWSBISupportIndicator: avp.Some(uint32(1))}, s.NodeIPSMGW, s.SBIAccess, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := answer()
			a.ServingNode = avp.Some(tc.n)
			r, e := s.InterpretSRA(s.RoutingContext{Request: request(t)}, a)
			if e != nil || r.Outcome != s.RoutingSuccess || len(r.Candidates) != 1 {
				t.Fatal(r, e)
			}
			n := r.Candidates[0]
			if n.Kind != tc.kind || n.Access != tc.access || n.Companion.Present != tc.companion || n.Number.Value != "123" || n.Source != s.PrimaryNode {
				t.Fatal(n)
			}
		})
	}
	t.Run("additional-owned", func(t *testing.T) {
		a := answer()
		a.AdditionalServingNode = avp.Some(s.AdditionalServingNode{SGSNNumber: number("456")})
		r, e := s.InterpretSRA(s.RoutingContext{Request: request(t)}, a)
		if e != nil || len(r.Candidates) != 2 || r.Candidates[1].Source != s.AdditionalNode {
			t.Fatal(r, e)
		}
		a.AdditionalServingNode.Value.SGSNNumber.Value[0] = 0xff
		if r.Answer.(*s.SRA).AdditionalServingNode.Value.SGSNNumber.Value[0] == 0xff {
			t.Fatal("borrowed")
		}
	})
}
func TestInterpretBoundaries(t *testing.T) {
	for _, name := range []string{"realm", "reserved-sbi", "missing-node", "invalid-tbcd", "unrequested-additional", "conflict", "missing-result", "unknown-vendor", "wrong-session", "missing-context", "reserved-diagnostic"} {
		t.Run(name, func(t *testing.T) {
			r, a := request(t), answer()
			unsupported := false
			switch name {
			case "realm":
				a.ServingNode = avp.Some(s.ServingNode{IPSMGWNumber: number("123"), IPSMGWRealm: avp.Some("example")})
				unsupported = true
			case "reserved-sbi":
				a.ServingNode = avp.Some(s.ServingNode{IPSMGWNumber: number("123"), IPSMGWSBISupportIndicator: avp.Some(uint32(2))})
				unsupported = true
			case "missing-node":
				a.ServingNode = avp.Optional[s.ServingNode]{}
			case "invalid-tbcd":
				a.ServingNode.Value.MSCNumber = avp.Some([]byte{0xff})
			case "unrequested-additional":
				r.SRRFlags = avp.Optional[uint32]{}
				a.AdditionalServingNode = avp.Some(s.AdditionalServingNode{SGSNNumber: number("123")})
			case "conflict":
				a.ExperimentalResult = avp.Some(s.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})
			case "missing-result":
				a.ResultCode = avp.Optional[uint32]{}
			case "unknown-vendor":
				a.ResultCode = avp.Optional[uint32]{}
				a.ExperimentalResult = avp.Some(s.ExperimentalResult{VendorID: 99, ExperimentalResultCode: 5001})
			case "wrong-session":
				a.SessionID = "other"
			case "missing-context", "reserved-diagnostic":
				a.ResultCode = avp.Optional[uint32]{}
				a.ExperimentalResult = avp.Some(s.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5550})
				a.ServingNode = avp.Optional[s.ServingNode]{}
				unsupported = true
				if name == "reserved-diagnostic" {
					a.MMEAbsentUserDiagnosticSM = avp.Some(uint32(255))
				}
			}
			out, e := s.InterpretSRA(s.RoutingContext{Request: r}, a)
			if unsupported {
				if e != nil || out.Outcome != s.RoutingUnsupported || len(out.Candidates) != 0 || len(out.Issues) == 0 || out.Answer == nil {
					t.Fatal(out, e)
				}
			} else if e == nil {
				t.Fatal("accepted", out)
			}
		})
	}
}
func TestInterpretResults(t *testing.T) {
	for _, code := range []uint32{5001, 5550, 5552, 5556, 5557} {
		t.Run(strconv.FormatUint(uint64(code), 10), func(t *testing.T) {
			r, a := request(t), answer()
			a.ResultCode = avp.Optional[uint32]{}
			a.ExperimentalResult = avp.Some(s.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
			a.ServingNode = avp.Optional[s.ServingNode]{}
			a.UserName = avp.Optional[string]{}
			ctx, e := s.RoutingRole(s.RoutingRoleEvidence{Role: s.RoutingFromGMSC, Requester: pdu.Origin{Host: r.OriginHost, Realm: r.OriginRealm}, Responder: pdu.Origin{Host: "hss.example", Realm: "example"}})
			if e != nil {
				t.Fatal(e)
			}
			if code == 5550 {
				a.MSCAbsentUserDiagnosticSM = avp.Some(uint32(0))
				a.MWDStatus = avp.Some(uint32(0))
			}
			out, e := s.InterpretSRA(s.RoutingContext{Request: r, Validation: ctx}, a)
			if e != nil || out.Outcome != s.RoutingRejected || out.Protocol.Namespace != s.ExperimentalResultNamespace || out.Protocol.Code != code {
				t.Fatal(out, e)
			}
			if code == 5550 && (len(out.Diagnostics) != 1 || out.Diagnostics[0].Value != 0 || !out.MWDStatus.Present) {
				t.Fatal(out)
			}
		})
	}
	r, a := request(t), answer()
	a.ResultCode = avp.Some(uint32(5012))
	a.UserName = avp.Optional[string]{}
	a.ServingNode = avp.Optional[s.ServingNode]{}
	out, e := s.InterpretSRA(s.RoutingContext{Request: r}, a)
	if e != nil || out.Protocol.Namespace != s.BaseResult || out.Outcome != s.RoutingRejected {
		t.Fatal(out, e)
	}
	ea := &pdu.ErrorAnswer{Command: pdu.CommandKey{ApplicationID: 16777312, Code: 8388647}, SessionID: avp.Some(r.SessionID), OriginHost: "hss.example", OriginRealm: "example", ResultCode: 3004}
	out, e = s.InterpretSRA(s.RoutingContext{Request: r}, ea)
	if e != nil || out.Protocol.Namespace != s.ErrorResult || out.Outcome != s.RoutingRejected {
		t.Fatal(out, e)
	}
}
