package s6c

import (
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/internal/smsidentity"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// MaxRouterAlertTokenBytes bounds this library's selected UTF8 token subset.
// This is a local evidence/storage limit, not a normative IMSI length rule.
const MaxRouterAlertTokenBytes = smsidentity.MaxAlertTokenBytes

// RouterAlertEvidence asserts the Router -> GMSC endpoint leg. PriorAlertToken
// is the exact User-Identifier/User-Name previously returned by this Router in
// the TFA; PriorGMSCAddress is the decimal SMS-GMSC-Address from the prior TFR.
// The application authenticates/retains that history and supplies its mapping.
// Neither the adjacent DRA nor the new Serving-Node identifies these endpoints.
type RouterAlertEvidence struct {
	Router, GMSC                      pdu.Origin
	PriorAlertToken, PriorGMSCAddress string
}

// RouterAlert owns one bounded assertion for ALR/ALA only, including the named
// legacy SGd Alert binding. It does not inspect a TFA, generate identifiers,
// verify external history, forward a request or schedule SMS delivery.
func RouterAlert(e RouterAlertEvidence) (validation.Context, error) {
	for _, o := range []pdu.Origin{e.Router, e.GMSC} {
		if avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if alertToken(e.PriorAlertToken) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.PriorGMSCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	return contextvalue.NewRouterAlert(contextvalue.SMSIdentity{
		Command:       8388648,
		RequesterHost: strings.Clone(e.Router.Host), RequesterRealm: strings.Clone(e.Router.Realm),
		ResponderHost: strings.Clone(e.GMSC.Host), ResponderRealm: strings.Clone(e.GMSC.Realm),
		UserName: strings.Clone(e.PriorAlertToken), SCAddress: strings.Clone(e.PriorGMSCAddress),
	}), nil
}

func alertToken(s string) error {
	return smsidentity.AlertToken(s)
}

func checkRouterAlert(r *ALR, answer pdu.Body, ctx validation.Context) error {
	e, ok := ctx.RouterAlert()
	if !ok || r.SMSMICorrelationID.Present || !r.UserIdentifier.UserName.Present || r.UserIdentifier.MSISDN.Present {
		return validation.ErrContext
	}
	if !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) || r.UserIdentifier.UserName.Value != e.UserName {
		return validation.ErrContext
	}
	if !r.SMSGMSCAlertEvent.Present {
		return validation.At(smsKey(3333), validation.ErrMissing)
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer != nil {
		a, ok := answer.(*ALA)
		if !ok || a == nil {
			return validation.ErrType
		}
		if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
			return validation.ErrContext
		}
	}
	return nil
}
