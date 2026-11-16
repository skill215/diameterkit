// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// AlertOriginRole is supplied by the application's end-terminal deployment
// contract. It never describes the adjacent DRA or the new serving node.
type AlertOriginRole uint8

const (
	AlertFromHSS AlertOriginRole = iota + 1
	AlertFromMME
	AlertFromSGSN
)

type AlertRoleEvidence struct {
	Role                 AlertOriginRole
	Requester, Responder pdu.Origin
}

// AlertRole owns one bounded ALR/ALA endpoint-role binding. It performs no I/O
// or history lookup and does not authorize endpoints outside runtime policy.
func AlertRole(e AlertRoleEvidence) (validation.Context, error) {
	if e.Role < AlertFromHSS || e.Role > AlertFromSGSN {
		return validation.Context{}, validation.ErrContext
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	return contextvalue.NewAlert(contextvalue.Alert{Role: uint8(e.Role), RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm)}), nil
}
func validateReportIdentity(m *ReportUserIdentifier, _ validation.Mode) error {
	return validateSMSRequestIdentity(m.UserName, m.MSISDN, m.Extra)
}
func validateAlertIdentity(m *AlertUserIdentifier, _ validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3111), smsKey(2400)); err != nil {
		return err
	}
	if m.UserName.Present == m.MSISDN.Present {
		return validation.ErrField
	}
	if m.UserName.Present {
		// Wire syntax alone cannot distinguish IMSI from a Router-issued token.
		// The named role/purpose context validates that semantic distinction.
		return validation.At(avp.Key{Code: 1}, alertToken(m.UserName.Value))
	}
	return nil
}
func validateSMSRequestIdentity(user avp.Optional[string], msisdn avp.Optional[[]byte], extra []avp.Raw) error {
	if err := unsupportedFields(extra, smsKey(3111), smsKey(2400)); err != nil {
		return err
	}
	if user.Present == msisdn.Present {
		return validation.ErrField
	}
	if user.Present {
		s := user.Value
		if len(s) < 5 || len(s) > 15 {
			return validation.At(avp.Key{Code: 1}, validation.ErrField)
		}
		for _, b := range s {
			if b < '0' || b > '9' {
				return validation.At(avp.Key{Code: 1}, validation.ErrUnsupported)
			}
		}
	}
	return nil
}
func validateALRSubset(m *ALR, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3325), smsKey(3326), smsKey(3327)); err != nil {
		return err
	}
	if m.SMSMICorrelationID.Present {
		if !m.UserIdentifier.MSISDN.Present || m.UserIdentifier.UserName.Present {
			return validation.At(smsKey(3102), validation.ErrField)
		}
		digits, err := avp.DecodeTBCD(m.UserIdentifier.MSISDN.Value)
		if err != nil || digits != "000000000000000" {
			return validation.At(smsKey(3102), validation.At(smsKey(701), validation.ErrField))
		}
	}

	if m.UserIdentifier.MSISDN.Present {
		if m.SMSGMSCAlertEvent.Present || m.ServingNode.Present {
			return validation.ErrField
		}
		return nil // Existing bounded HSS wire shape; not a proof of deployment role.
	}
	if !m.SMSGMSCAlertEvent.Present {
		// HSS T4 is distinguishable only with explicit purpose evidence.
		if m.ServingNode.Present {
			return validation.At(smsKey(2401), validation.ErrField)
		}
		return nil
	}
	f := m.SMSGMSCAlertEvent.Value
	if f&3 == 0 {
		return validation.At(smsKey(3333), validation.ErrUnsupported)
	}
	if mode != validation.Receive && f&^uint32(3) != 0 {
		return validation.At(smsKey(3333), validation.ErrUnsupported)
	}
	if m.MaximumUEAvailabilityTime.Present {
		return validation.At(smsKey(3329), validation.ErrUnsupported)
	}
	if m.ServingNode.Present {
		n := m.ServingNode.Value
		if n.IPSMGWNumber.Present {
			return validation.At(smsKey(2401), validation.ErrUnsupported)
		}
		if f&2 == 0 || n.MSCNumber.Present {
			return validation.At(smsKey(2401), validation.ErrField)
		}
		// ServingNode's generated rule validates either full MME or SGSN shapes.
		// Cross-RAT moves are legal; the old sender's role must not choose the new node.
	}
	return nil
}
func checkAlertRole(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	r, ok := request.(*ALR)
	if !ok {
		return validation.ErrContextRequired
	}
	if r.SMSMICorrelationID.Present {
		return checkIMSAlert(r, answer, ctx)
	}
	if _, ok := ctx.RouterAlert(); ok {
		return checkRouterAlert(r, answer, ctx)
	}

	if _, ok := ctx.S6cT4(); ok {
		return checkT4NoMSISDN(request, answer, ctx)
	}
	e, present := ctx.Alert()
	if !present {
		if !ctx.Empty() {
			return validation.ErrContext
		}
		if r.UserIdentifier.MSISDN.Present {
			return nil
		}
		return validation.ErrContextRequired
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if (e.Role == uint8(AlertFromHSS)) != r.UserIdentifier.MSISDN.Present {
		return validation.ErrContext
	}
	if r.UserIdentifier.UserName.Present && !r.SMSGMSCAlertEvent.Present {
		return validation.At(smsKey(3333), validation.ErrMissing)
	}
	if r.UserIdentifier.UserName.Present {
		if err := validateSMSRequestIdentity(r.UserIdentifier.UserName, r.UserIdentifier.MSISDN, r.UserIdentifier.Extra); err != nil {
			return validation.At(smsKey(3102), err)
		}
	}
	if answer != nil {
		a, ok := answer.(*ALA)
		if !ok {
			return validation.ErrType
		}
		if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
			return validation.ErrContext
		}
	}
	return nil
}
