package sgd

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// IMSMTForwardingEvidence selects one SMS-GMSC -> IP-SM-GW delivery to an IMS
// originator and destination without MSISDN. HSSID identifies the destination
// HSS, not the IP-SM-GW
// endpoint. Selection/history and both public identities come from the application;
// no previous SRR/OFR exchange is required or consulted by this library.
type IMSMTForwardingEvidence struct {
	Requester, Responder                                   pdu.Origin
	HSSID, OriginatingSIPURI, DestinationSIPURI, SCAddress string
}

// IMSMTForwarding owns bounded, separate MT purpose evidence. Endpoint identities
// are at most 253 bytes, selected ASCII SIP URIs at most 256 each. A domain-separated
// digest binds the ordered raw URIs; the owned request keeps the full strings.
// SCAddress is international decimal digits, without '+' or TOA. No routing lookup,
// SIP/TPDU processing, delivery scheduling, persistence or retry is performed.
func IMSMTForwarding(e IMSMTForwardingEvidence) (validation.Context, error) {
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
	return contextvalue.NewIMSMT(contextvalue.SMSIdentity{Command: 8388646, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UserName: strings.Clone(e.HSSID), SCAddress: strings.Clone(e.SCAddress), URIsDigest: smsidentity.MTURIsDigest(e.OriginatingSIPURI, e.DestinationSIPURI)}), nil
}
func validateMTCorrelation(m *MTSMSMICorrelationID, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, 3325); err != nil {
		return err
	}
	if mode != validation.Receive {
		if err := smsidentity.SIPURI(m.OriginatingSIPURI); err != nil {
			return validation.At(smsKey(3326), err)
		}
		return validation.At(smsKey(3327), smsidentity.SIPURI(m.DestinationSIPURI))
	}
	return nil
}
func validateIMSMTRequest(r *TFR) error {
	// MME/SGSN power-saving/retransmission evidence does not authorize IMS MT.
	if r.MaximumRetransmissionTime.Present {
		return validation.At(smsKey(3330), validation.ErrUnsupported)
	}
	if r.SMSGMSCAddress.Present {
		return validation.At(smsKey(3332), validation.ErrUnsupported)
	}
	return validation.At(smsKey(3324), validateMTCorrelation(&r.SMSMICorrelationID.Value, validation.Send))
}
func checkIMSMT(r *TFR, answer pdu.Body, ctx validation.Context) error {
	if err := validateIMSMTRequest(r); err != nil {
		return err
	}
	if ctx.Empty() {
		return validation.ErrContextRequired
	}
	e, ok := ctx.IMSMT()
	if !ok {
		return validation.ErrContext
	}
	m := r.SMSMICorrelationID.Value
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) || r.UserName != e.UserName || smsidentity.MTURIsDigest(m.OriginatingSIPURI, m.DestinationSIPURI) != e.URIsDigest {
		return validation.ErrContext
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*TFA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if a.ExperimentalResult.Present {
		switch a.ExperimentalResult.Value.ExperimentalResultCode {
		case 5552, 5555:
		default:
			return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
		}
	}
	if a.AbsentUserDiagnosticSM.Present {
		return validation.At(smsKey(3322), validation.ErrUnsupported)
	}
	if a.EPSLocationInformation.Present {
		return validation.At(smsKey(1496), validation.ErrUnsupported)
	}
	if a.RequestedRetransmissionTime.Present {
		return validation.At(smsKey(3331), validation.ErrUnsupported)
	}
	return nil
}
