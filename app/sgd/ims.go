// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sgd

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// IMSMOForwardingEvidence selects IP-SM-GW -> SMS-IWMSC after a temporary direct
// delivery failure. Source identities are explicit: known IMSI with optional real
// MSISDN, or neither identity (the request then uses the fifteen-zero dummy).
// HSSID identifies the destination HSS, not the sender or the adjacent DRA.
// HSSID and the request outcome come from the destination IMS network's failure;
// this factory asserts purpose, not verification of SIP history or subscribers.
// Absent SenderIMSI means unavailable; absent SenderMSISDN means no usable number,
// not an unknown value to fill in later. A present MSISDN is never the dummy.
type IMSMOForwardingEvidence struct {
	Requester, Responder                                   pdu.Origin
	HSSID, OriginatingSIPURI, DestinationSIPURI, SCAddress string
	SenderIMSI, SenderMSISDN                               avp.Optional[string]
}

// IMSMOForwarding owns bounded evidence for OFR/OFA. Each selected ASCII SIP URI
// is at most 256 bytes; their ordered raw bytes are bound by a SHA-256 digest.
// Original URI values remain in the owned request. No lookup, MWD, retry or SMS
// forwarding is performed. SCAddress is international digits without '+' or TOA.
func IMSMOForwarding(e IMSMOForwardingEvidence) (validation.Context, error) {
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if smsidentity.HSSID(e.HSSID) != nil || smsidentity.SIPURI(e.OriginatingSIPURI) != nil || smsidentity.SIPURI(e.DestinationSIPURI) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	if e.SenderIMSI.Present && imsi(e.SenderIMSI.Value) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if e.SenderMSISDN.Present {
		if !e.SenderIMSI.Present || e.SenderMSISDN.Value == "000000000000000" {
			return validation.Context{}, validation.ErrContext
		}
		if _, err := avp.EncodeTBCD(e.SenderMSISDN.Value); err != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	var senderIMSI, senderMSISDN string
	if e.SenderIMSI.Present {
		senderIMSI = strings.Clone(e.SenderIMSI.Value)
	}
	if e.SenderMSISDN.Present {
		senderMSISDN = strings.Clone(e.SenderMSISDN.Value)
	}
	return contextvalue.NewIMSMO(contextvalue.SMSIdentity{Command: 8388645, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UserName: strings.Clone(e.HSSID), SCAddress: strings.Clone(e.SCAddress), URIsDigest: smsidentity.URIsDigest(e.OriginatingSIPURI, e.DestinationSIPURI)}, senderIMSI, senderMSISDN, e.SenderMSISDN.Present), nil
}
func validateMOCorrelation(m *MOSMSMICorrelationID, mode validation.Mode) error {
	if err := smsidentity.HSSID(m.HSSID); err != nil {
		return validation.At(smsKey(3325), err)
	}
	if mode != validation.Receive {
		if err := smsidentity.SIPURI(m.OriginatingSIPURI); err != nil {
			return validation.At(smsKey(3326), err)
		}
		return validation.At(smsKey(3327), smsidentity.SIPURI(m.DestinationSIPURI))
	}
	return nil
}
func validateMOOutcome(m *MOSMDeliveryOutcome, _ validation.Mode) error {
	return unsupportedFields(m.Extra, 3317, 3318, 3319, 3336, 3337, 3321, 3322)
}
func validateIMSOutcome(m *IPSMGWSMDeliveryOutcome, mode validation.Mode) error {
	if m.AbsentUserDiagnosticSM.Present {
		if m.SMDeliveryCause != 1 {
			return validation.At(smsKey(3322), validation.ErrField)
		}
		if m.AbsentUserDiagnosticSM.Value > 255 {
			return validation.At(smsKey(3322), validation.ErrField)
		}
	}
	if mode != validation.Receive {
		// Only the temporary-failure paths selected for this forwarding procedure.
		if m.SMDeliveryCause > 1 {
			return validation.At(smsKey(3321), validation.ErrUnsupported)
		}
		if m.AbsentUserDiagnosticSM.Present && m.AbsentUserDiagnosticSM.Value != 11 && m.AbsentUserDiagnosticSM.Value != 12 {
			return validation.At(smsKey(3322), validation.ErrUnsupported)
		}
	}
	return nil
}
func checkIMSMO(r *OFR, answer pdu.Body, ctx validation.Context) error {
	if err := validateMOCorrelation(&r.SMSMICorrelationID.Value, validation.Send); err != nil {
		return validation.At(smsKey(3324), err)
	}
	if err := validateIMSOutcome(&r.SMDeliveryOutcome.Value.IPSMGWSMDeliveryOutcome, validation.Send); err != nil {
		return validation.At(smsKey(3316), validation.At(smsKey(3320), err))
	}
	if ctx.Empty() {
		return validation.ErrContextRequired
	}
	e, ok := ctx.IMSMO()
	if !ok {
		return validation.ErrContext
	}
	m := r.SMSMICorrelationID.Value
	if !r.DestinationHost.Present || !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) || m.HSSID != e.UserName || smsidentity.URIsDigest(m.OriginatingSIPURI, m.DestinationSIPURI) != e.URIsDigest {
		return validation.ErrContext
	}
	// The factory stores only available source identities; the dummy is wire-only.
	senderIMSI, senderMSISDN, msisdnPresent, _ := ctx.IMSMOSender()
	if r.UserIdentifier.UserName.Present != (senderIMSI != "") || (r.UserIdentifier.UserName.Present && r.UserIdentifier.UserName.Value != senderIMSI) {
		return validation.ErrContext
	}
	if senderIMSI != "" {
		if r.UserIdentifier.MSISDN.Present != msisdnPresent {
			return validation.ErrContext
		}
		if msisdnPresent {
			digits, err := avp.DecodeTBCD(r.UserIdentifier.MSISDN.Value)
			if err != nil || digits != senderMSISDN {
				return validation.ErrContext
			}
		}
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*OFA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	// Real MSISDN presence does not authorize ordinary SGd unknown-MSISDN or
	// T4 charging results under this IMS purpose.
	if moUnknown(a) {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	if a.ExternalIdentifier.Present {
		return validation.At(smsKey(3111), validation.ErrUnsupported)
	}
	return nil
}
