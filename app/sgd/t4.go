package sgd

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// T4ChargingEvidence explicitly selects SMS-IWMSC -> MTC-IWF MO forwarding
// for charging. It is an application assertion, not a lookup or SCTP proof.
// SenderIMSI and the optional SenderMSISDN must match the original request,
// including absence. SCAddress is the configured MTC-IWF international number.
// Numbers are decimal digits, without '+' or TON/NPI. ExternalIdentifier, when
// present, is the application's expected exact mapping for this sender.
// Leave that mapping absent when it is not yet known: a normal answer then
// remains owned but its validation needs application evidence. Inbound.Check
// freezes successful evidence, so supply the mapping before preparing an answer.
type T4ChargingEvidence struct {
	Requester, Responder pdu.Origin
	SenderIMSI           string
	SenderMSISDN         avp.Optional[string]
	SCAddress            string
	ExternalIdentifier   avp.Optional[string]
}

// T4Charging creates fixed, owned evidence for one original OFR/OFA exchange.
// Endpoint fields and the supported ASCII external identifier are each bounded
// to253 bytes; this is this slice's resource limit, not a protocol NAI maximum.
// Ordinary SGd/Gdd with no charging identifier continues to accept zero Context.
func T4Charging(e T4ChargingEvidence) (validation.Context, error) {
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if imsi(e.SenderIMSI) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	if e.SenderMSISDN.Present {
		if _, err := avp.EncodeTBCD(e.SenderMSISDN.Value); err != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if e.ExternalIdentifier.Present && externalIdentifierASCII(e.ExternalIdentifier.Value) != nil {
		return validation.Context{}, validation.ErrContext
	}
	v := contextvalue.T4Charging{RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), IMSI: strings.Clone(e.SenderIMSI), SCAddress: strings.Clone(e.SCAddress), MSISDNPresent: e.SenderMSISDN.Present, ExternalKnown: e.ExternalIdentifier.Present}
	if e.SenderMSISDN.Present {
		v.MSISDN = strings.Clone(e.SenderMSISDN.Value)
	}
	if e.ExternalIdentifier.Present {
		v.ExternalIdentifier = strings.Clone(e.ExternalIdentifier.Value)
	}
	return contextvalue.NewT4Charging(v), nil
}

// RFC4282 ASCII dot-string with escaped characters and a multi-label DNS realm.
// Preserve escapes, case and leading zeroes; do not use naive Split("@").
// Unicode preparation/IDNA, controls and long semantic identities are not
// implemented by this bounded slice. Wire UTF8 and field limits are separate.
func externalIdentifierASCII(s string) error {
	if len(s) > 253 {
		return validation.ErrUnsupported
	}
	for i := range len(s) {
		if s[i] >= 128 {
			return validation.ErrUnsupported
		}
	}
	part := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case '\\':
			i++
			if i == len(s) || s[i] < 32 || s[i] > 126 {
				return validation.ErrField
			}
			part = true
		case '.':
			if !part {
				return validation.ErrField
			}
			part = false
		case '@':
			if !part {
				return validation.ErrField
			}
			realm := s[i+1:]
			if !strings.Contains(realm, ".") || avp.Identity(realm) != nil {
				return validation.ErrField
			}
			return nil
		default:
			if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", rune(b))) {
				return validation.ErrField
			}
			part = true
		}
	}
	return validation.ErrField
}

func checkT4Charging(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	r, ok := request.(*OFR)
	if !ok {
		return validation.ErrContextRequired
	}
	if r.SMSMICorrelationID.Present {
		return checkIMSMO(r, answer, ctx)
	}
	if e, present := ctx.MOForwarding(); present {
		return checkMOForwarding(r, answer, e)
	}
	e, present := ctx.T4Charging()
	if !present {
		if !ctx.Empty() {
			return validation.ErrContext
		}
		if a, ok := answer.(*OFA); ok && (a.ExternalIdentifier.Present || moUnknown(a)) {
			return validation.ErrContextRequired
		}
		return nil // Existing ordinary MO slice, not implicit T4 role/charging proof.
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !r.DestinationHost.Present || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if !r.UserIdentifier.UserName.Present || r.UserIdentifier.UserName.Value != e.IMSI || r.UserIdentifier.MSISDN.Present != e.MSISDNPresent {
		return validation.ErrContext
	}
	if e.MSISDNPresent {
		n, err := avp.DecodeTBCD(r.UserIdentifier.MSISDN.Value)
		if err != nil || n != e.MSISDN {
			return validation.ErrContext
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
	if moUnknown(a) {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	if !a.ExternalIdentifier.Present {
		return validation.At(smsKey(3111), validation.ErrMissing)
	}
	if err := externalIdentifierASCII(a.ExternalIdentifier.Value); err != nil {
		return validation.At(smsKey(3111), err)
	}
	if !e.ExternalKnown {
		return validation.ErrContextRequired
	}
	if a.ExternalIdentifier.Value != e.ExternalIdentifier {
		return validation.At(smsKey(3111), validation.ErrField)
	}
	return nil
}
