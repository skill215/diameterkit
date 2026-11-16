// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package s6c

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// RoutingOriginRole describes the end requester, never the adjacent DRA.
type RoutingOriginRole uint8

const (
	RoutingFromGMSC RoutingOriginRole = iota + 1
	RoutingFromIPSMGW
)

// RoutingRoleEvidence is bounded deployment evidence for an ordinary MSISDN SRR.
// UNRI is an external fact used only for the IP-SM-GW Absent User identity rule.
// It is not a bit in SRR-Flags and is not inferred from an address or a timer.
// Unknown UNRI can be supplied at submission; a 5550 answer without IMSI then remains incomplete.
type RoutingRoleEvidence struct {
	Role                 RoutingOriginRole
	Requester, Responder pdu.Origin
	UNRI                 avp.Optional[bool]
}

// RoutingRole owns its endpoint strings; it performs no HSS lookup or MWD update.
func RoutingRole(e RoutingRoleEvidence) (validation.Context, error) {
	if e.Role != RoutingFromGMSC && e.Role != RoutingFromIPSMGW || e.Role == RoutingFromGMSC && e.UNRI.Present {
		return validation.Context{}, validation.ErrContext
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	r := contextvalue.Routing{Role: uint8(e.Role), RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UNRIKnown: e.UNRI.Present}
	if e.UNRI.Present {
		r.UNRI = e.UNRI.Value
	}
	return contextvalue.NewRouting(r), nil
}
func isAbsentSRA(answer pdu.Body) bool {
	a, ok := answer.(*SRA)
	return ok && a.ExperimentalResult.Present && a.ExperimentalResult.Value.VendorID == 10415 && a.ExperimentalResult.Value.ExperimentalResultCode == 5550
}
func checkRoutingRole(r *SRR, answer pdu.Body, ctx validation.Context) error {
	e, ok := ctx.Routing()
	if !ok {
		return validation.ErrContext
	}
	if r.UserName.Present || !r.MSISDN.Present || !r.DestinationHost.Present {
		return validation.ErrContext
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*SRA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if isAbsentSRA(answer) && e.Role == uint8(RoutingFromIPSMGW) && !a.UserName.Present {
		if !e.UNRIKnown {
			return validation.ErrContextRequired
		}
		if !e.UNRI {
			return validation.At(avp.Key{Code: 1}, validation.ErrMissing)
		}
	}
	return nil
}
func hasSRADiagnostics(m *SRA) bool {
	return m.MMEAbsentUserDiagnosticSM.Present || m.MSCAbsentUserDiagnosticSM.Present || m.SGSNAbsentUserDiagnosticSM.Present
}
func validateSRADiagnostics(m *SRA, mode validation.Mode) error {
	for _, d := range []struct {
		code uint32
		v    avp.Optional[uint32]
		mask uint16
	}{
		{3313, m.MMEAbsentUserDiagnosticSM, 1 << 13},
		{3314, m.MSCAbsentUserDiagnosticSM, 1<<0 | 1<<1 | 1<<2 | 1<<3 | 1<<4 | 1<<9 | 1<<13},
		{3315, m.SGSNAbsentUserDiagnosticSM, 1<<2 | 1<<5 | 1<<6 | 1<<7 | 1<<8 | 1<<10 | 1<<13},
	} {
		if !d.v.Present {
			continue
		}
		if d.v.Value > 255 {
			return validation.At(smsKey(d.code), validation.ErrField)
		}
		if mode != validation.Receive && (d.v.Value > 13 || d.mask&(1<<d.v.Value) == 0) {
			return validation.At(smsKey(d.code), validation.ErrUnsupported)
		}
	}
	return nil
}
