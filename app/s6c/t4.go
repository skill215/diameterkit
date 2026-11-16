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

type T4Procedure uint8

const (
	// T4DeliveryReport asserts a report following T4 Submit Trigger, GMSC -> HSS.
	// Selected MME/MSC/SGSN/IP-SM-GW outcome parents describe delivery nodes,
	// not the Diameter requester role; each supplied parent keeps its own cause.
	T4DeliveryReport T4Procedure = iota + 1
	// T4HSSAlert asserts HSS -> IWMSC, with no available MSISDN/Alert MSISDN.
	T4HSSAlert
	// T4Routing asserts GMSC -> HSS routing for a T4 device trigger without MSISDN.
	T4Routing
)

// T4NoMSISDNEvidence selects one T4 procedure with a known IMSI and no usable
// MSISDN. The application supplies that fact; the library does not discover it
// from a digit string, a DRA connection or previous exchanges. SCAddress is
// decimal international digits without '+' or TON/NPI. Each exchange supplies
// its own evidence: a successful report never automatically authorizes an alert.
type T4NoMSISDNEvidence struct {
	Procedure            T4Procedure
	Requester, Responder pdu.Origin
	IMSI, SCAddress      string
}

// T4NoMSISDN owns a bounded, immutable purpose/identity/endpoint assertion.
// It does not store MWD, verify delivery history, resolve users or send requests.
func T4NoMSISDN(e T4NoMSISDNEvidence) (validation.Context, error) {
	command := uint32(8388649)
	switch e.Procedure {
	case T4DeliveryReport:
	case T4HSSAlert:
		command = 8388648
	case T4Routing:
		command = 8388647
	default:
		return validation.Context{}, validation.ErrContext
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if len(e.IMSI) < 5 || len(e.IMSI) > 15 {
		return validation.Context{}, validation.ErrContext
	}
	for _, c := range e.IMSI {
		if c < '0' || c > '9' {
			return validation.Context{}, validation.ErrContext
		}
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	return contextvalue.NewS6cT4(contextvalue.SMSIdentity{Command: command, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UserName: strings.Clone(e.IMSI), SCAddress: strings.Clone(e.SCAddress)}), nil
}
func checkRoutingContext(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	r, ok := request.(*SRR)
	if !ok {
		return validation.ErrContextRequired
	}
	if a, ok := answer.(*SRA); ok {
		if err := checkSRAIPSMGWSelection(a); err != nil {
			return err
		}
		if err := validateSRADiagnostics(a, validation.Send); err != nil {
			return err
		}
	}
	if _, ok := ctx.RouterRouting(); ok {
		return checkSpecialRouting(r, answer, ctx)
	}
	if _, ok := ctx.MCCMNCRouting(); ok {
		return checkSpecialRouting(r, answer, ctx)
	}
	if r.SMDeliveryNotIntended.Present {
		if ctx.Empty() {
			return validation.ErrContextRequired
		}
		return validation.ErrContext
	}
	if a, ok := answer.(*SRA); ok && a.UserName.Present && imsi(a.UserName.Value) != nil {
		if ctx.Empty() {
			return validation.ErrContextRequired
		}
		return validation.At(avp.Key{Code: 1}, validation.ErrField)
	}
	if r.SMSMICorrelationID.Present {
		return checkIMSRouting(r, answer, ctx)
	}
	if _, ok := ctx.Routing(); ok {
		return checkRoutingRole(r, answer, ctx)
	}
	if ctx.Empty() {
		if r.UserName.Present || isAbsentSRA(answer) {
			return validation.ErrContextRequired
		}
		return nil
	}
	return checkT4NoMSISDN(request, answer, ctx)
}
func checkReportContext(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	r, ok := request.(*RDR)
	if !ok {
		return validation.ErrContextRequired
	}
	if r.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Present {
		if err := validateIPSMGWOutcomeSubset(&r.SMDeliveryOutcome.IPSMGWSMDeliveryOutcome.Value, validation.Send); err != nil {
			return validation.At(smsKey(3316), validation.At(smsKey(3320), err))
		}
	}

	if r.SMSMICorrelationID.Present {
		return checkIMSReport(r, answer, ctx)
	}

	if ctx.Empty() {
		if r.UserIdentifier.UserName.Present {
			return validation.ErrContextRequired
		}
		return nil // Original ordinary MSISDN wire slice only.
	}
	return checkT4NoMSISDN(request, answer, ctx)
}
func checkT4NoMSISDN(request, answer pdu.Body, ctx validation.Context) error {
	e, ok := ctx.S6cT4()
	if !ok {
		return validation.ErrContext
	}
	var host, realm, dest, dr string
	var user avp.Optional[string]
	var msisdn avp.Optional[[]byte]
	var sc []byte
	switch r := request.(type) {
	case *SRR:
		if e.Command != 8388647 || !r.DestinationHost.Present {
			return validation.ErrContext
		}
		host, realm, dest, dr = r.OriginHost, r.OriginRealm, r.DestinationHost.Value, r.DestinationRealm
		user, msisdn, sc = r.UserName, r.MSISDN, r.SCAddress
	case *RDR:
		if e.Command != 8388649 || !r.DestinationHost.Present {
			return validation.ErrContext
		}
		host, realm, dest, dr = r.OriginHost, r.OriginRealm, r.DestinationHost.Value, r.DestinationRealm
		user, msisdn, sc = r.UserIdentifier.UserName, r.UserIdentifier.MSISDN, r.SCAddress
	case *ALR:
		if e.Command != 8388648 {
			return validation.ErrContext
		}
		if r.SMSGMSCAlertEvent.Present || r.ServingNode.Present {
			return validation.ErrContext
		}
		host, realm, dest, dr = r.OriginHost, r.OriginRealm, r.DestinationHost, r.DestinationRealm
		user, msisdn, sc = r.UserIdentifier.UserName, r.UserIdentifier.MSISDN, r.SCAddress
	default:
		return validation.ErrType
	}
	if !strings.EqualFold(host, e.RequesterHost) || !strings.EqualFold(realm, e.RequesterRealm) || !strings.EqualFold(dest, e.ResponderHost) || !strings.EqualFold(dr, e.ResponderRealm) {
		return validation.ErrContext
	}
	if !user.Present || msisdn.Present || user.Value != e.UserName {
		return validation.ErrContext
	}
	digits, err := avp.DecodeTBCD(sc)
	if err != nil || digits != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	switch a := answer.(type) {
	case *SRA:
		if e.Command != 8388647 {
			return validation.ErrType
		}
		host, realm = a.OriginHost, a.OriginRealm
	case *RDA:
		if e.Command != 8388649 {
			return validation.ErrType
		}
		host, realm = a.OriginHost, a.OriginRealm
	case *ALA:
		if e.Command != 8388648 {
			return validation.ErrType
		}
		host, realm = a.OriginHost, a.OriginRealm
	default:
		return validation.ErrType
	}
	if !strings.EqualFold(host, e.ResponderHost) || !strings.EqualFold(realm, e.ResponderRealm) {
		return validation.ErrContext
	}
	return nil
}
