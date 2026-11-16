package s6c

import (
	"github.com/skill215/diameterkit/internal/smsidentity"
	"strings"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// IMSDeliveryReportEvidence asserts one MSISDN-less IMS retry report to a known
// HSS. The application supplies the identities and purpose, not a history store.
// SCAddress is decimal international digits, without '+' or TON/NPI.
type IMSDeliveryReportEvidence struct {
	Requester, Responder                pdu.Origin
	HSSID, DestinationSIPURI, SCAddress string
}

// IMSDeliveryReport owns bounded evidence for RDR/RDA only. It neither resolves
// HSS addresses nor verifies allocation against subscribers or schedules retry.
// URI matching is byte-exact; the selected ASCII SIP subset is at most 256 bytes.
func IMSDeliveryReport(e IMSDeliveryReportEvidence) (validation.Context, error) {
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return validation.Context{}, validation.ErrContext
		}
	}
	if hssID(e.HSSID) != nil || reportSIPURI(e.DestinationSIPURI) != nil {
		return validation.Context{}, validation.ErrContext
	}
	if _, err := avp.EncodeTBCD(e.SCAddress); err != nil {
		return validation.Context{}, validation.ErrContext
	}
	return contextvalue.NewS6cIMSReport(contextvalue.SMSIdentity{Command: 8388649, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), UserName: strings.Clone(e.HSSID), SCAddress: strings.Clone(e.SCAddress), DestinationSIPURI: strings.Clone(e.DestinationSIPURI)}), nil
}
func hssID(s string) error        { return smsidentity.HSSID(s) }
func reportSIPURI(s string) error { return smsidentity.SIPURI(s) }
func validateReportCorrelation(m *ReportSMSMICorrelationID, mode validation.Mode) error {
	if err := unsupportedFields(m.Extra, smsKey(3325), smsKey(3326)); err != nil {
		return err
	}
	if mode != validation.Receive {
		return validation.At(smsKey(3327), reportSIPURI(m.DestinationSIPURI))
	}
	return nil
}
func checkIMSReport(r *RDR, answer pdu.Body, ctx validation.Context) error {
	if err := validateReportCorrelation(&r.SMSMICorrelationID.Value, validation.Send); err != nil {
		return validation.At(smsKey(3324), err)
	}
	if ctx.Empty() {
		return validation.ErrContextRequired
	}
	e, ok := ctx.S6cIMSReport()
	if !ok {
		return validation.ErrContext
	}
	if !r.DestinationHost.Present || !strings.EqualFold(r.OriginHost, e.RequesterHost) || !strings.EqualFold(r.OriginRealm, e.RequesterRealm) || !strings.EqualFold(r.DestinationHost.Value, e.ResponderHost) || !strings.EqualFold(r.DestinationRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	if r.UserIdentifier.UserName.Value != e.UserName || r.SMSMICorrelationID.Value.DestinationSIPURI != e.DestinationSIPURI {
		return validation.ErrContext
	}
	sc, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || sc != e.SCAddress {
		return validation.ErrContext
	}
	if answer == nil {
		return nil
	}
	a, ok := answer.(*RDA)
	if !ok {
		return validation.ErrType
	}
	if !strings.EqualFold(a.OriginHost, e.ResponderHost) || !strings.EqualFold(a.OriginRealm, e.ResponderRealm) {
		return validation.ErrContext
	}
	return nil
}
