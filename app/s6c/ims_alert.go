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

// IMSHSSAlertEvidence declares an IMS no-MSISDN retry alert from HSS to IWMSC.
// The application supplies the destination URI and SC from its own procedure;
// this value contains no MWD history or permission to initiate future requests.
type IMSHSSAlertEvidence struct {
	Requester, Responder         pdu.Origin
	DestinationSIPURI, SCAddress string
}

// IMSHSSAlert owns bounded evidence for ALR/ALA only. SCAddress is decimal
// international digits without '+' or TON/NPI. The message must explicitly
// contain the fifteen-zero dummy MSISDN; this factory never inserts it.
func IMSHSSAlert(e IMSHSSAlertEvidence) (validation.Context, error) {
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if reportSIPURI(e.DestinationSIPURI) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	return contextvalue.NewS6cIMSAlert(contextvalue.SMSIdentity{Command: 8388648, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), SCAddress: strings.Clone(e.SCAddress), DestinationSIPURI: strings.Clone(e.DestinationSIPURI)}), nil
}
func validateAlertCorrelation(m *AlertSMSMICorrelationID, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3325), smsKey(3326)); err != nil {
		return err
	}
	if mode != validation.Receive {
		return validation.At(smsKey(3327), reportSIPURI(m.DestinationSIPURI))
	}
	return nil
}
func checkIMSAlert(r *ALR, answer pdu.Body, ctx validation.Context) error {
	if err := validateAlertCorrelation(&r.SMSMICorrelationID.Value, validation.Send); err != nil {
		return validation.At(smsKey(3324), err)
	}
	if ctx.Empty() {
		return validation.ErrContextRequired
	}
	e, ok := ctx.S6cIMSAlert()
	if !ok {
		return validation.ErrContext
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) || r.SMSMICorrelationID.Value.DestinationSIPURI != e.DestinationSIPURI {
		return validation.ErrContext
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*ALA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	return nil
}
