package shdata

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"

	"github.com/skill215/diameterkit/avp"
)

const openData = "<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension>"
const closeData = "</Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>"

// MarshalSMSUpdate constructs Ref24 update data. Number="" explicitly encodes
// deregistration. SCAddress is rejected even when present with an empty value.
// Nonempty numbers must be 1..15 ASCII digits, first digit 1..9. No coercion,
// PublicIdentifiers, unknown extension or implicit XML from application state.
func MarshalSMSUpdate(s SMSRegistration, limits Limits) ([]byte, error) {
	if s.SCAddress.Present {
		return nil, ErrArgument
	}
	return marshalSMS(s, limits)
}

// MarshalSMSPull constructs Ref24 pull data. SCAddress preserves absent versus
// explicit empty; nonempty text uses the same finite numeric subset as Number.
// Parsing can preserve other XSD strings with an incomplete interpretation report.
func MarshalSMSPull(s SMSRegistration, limits Limits) ([]byte, error) { return marshalSMS(s, limits) }
func marshalSMS(s SMSRegistration, l Limits) ([]byte, error) {
	if !l.valid() {
		return nil, ErrArgument
	}
	if s.Number != "" && !number(s.Number) || s.SCAddress.Present && s.SCAddress.Value != "" && !number(s.SCAddress.Value) {
		return nil, ErrUnsupported
	}
	if s.SBI.Present && s.SBI.Value > 1 {
		return nil, ErrStructure
	}
	w := boundedXML{max: l.MaxBytes}
	w.tag(openData + "<SMSRegistrationInfo>")
	w.leaf("IP-SM-GW-Number", s.Number)
	if s.SCAddress.Present || s.SBI.Present {
		w.tag("<Extension>")
		if s.SCAddress.Present {
			w.leaf("SCAddress", s.SCAddress.Value)
		}
		if s.SBI.Present {
			w.tag("<Extension>")
			w.leaf("IP-SM-GW-SBISupportIndication", strconv.Itoa(int(s.SBI.Value)))
			w.tag("</Extension>")
		}
		w.tag("</Extension>")
	}
	w.tag("</SMSRegistrationInfo>" + closeData)
	return w.finish(l)
}

// MarshalReachability constructs Ref25 data with at least one explicit zero
// marker. A missing marker is not an unreachable boolean. No subscription or
// MME/SGSN/AMF action is performed, and the four markers are never merged.
func MarshalReachability(r Reachability, l Limits) ([]byte, error) {
	if !l.valid() {
		return nil, ErrArgument
	}
	w := boundedXML{max: l.MaxBytes}
	w.tag(openData)
	if err := writeReachability(&w, r); err != nil {
		return nil, err
	}
	w.tag(closeData)
	return w.finish(l)
}
func writeReachability(w *boundedXML, r Reachability) error {
	if !r.any() {
		return ErrStructure
	}
	for _, v := range []avp.Optional[uint8]{r.MME, r.SGSN, r.AMF3GPP, r.AMFNon3GPP} {
		if v.Present && v.Value != 0 {
			return ErrStructure
		}
	}

	w.tag("<UEReachabilityForIP>")
	if r.MME.Present {
		w.leaf("UEIPReachabilityMME", "0")
	}
	if r.SGSN.Present || r.AMF3GPP.Present || r.AMFNon3GPP.Present {
		w.tag("<Extension>")
		if r.SGSN.Present {
			w.leaf("UEIPReachabilitySGSN", "0")
		}
		if r.AMF3GPP.Present || r.AMFNon3GPP.Present {
			w.tag("<Extension>")
			if r.AMF3GPP.Present {
				w.leaf("UEIPReachabilityAMF3GPP", "0")
			}
			if r.AMFNon3GPP.Present {
				w.leaf("UEIPReachabilityAMFnon3GPP", "0")
			}
			w.tag("</Extension>")
		}
		w.tag("</Extension>")
	}
	w.tag("</UEReachabilityForIP>")
	return w.err
}

type boundedXML struct {
	bytes.Buffer
	max int
	err error
}

func (w *boundedXML) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if len(b) > w.max-w.Len() {
		w.err = ErrLimit
		return 0, w.err
	}
	return w.Buffer.Write(b)
}
func (w *boundedXML) tag(s string) {
	if w.err == nil {
		_, w.err = w.Write([]byte(s))
	}
}
func (w *boundedXML) leaf(name, value string) {
	w.tag("<" + name + ">")
	if w.err == nil {
		w.err = xml.EscapeText(io.Writer(w), []byte(value))
	}
	w.tag("</" + name + ">")
}
func (w *boundedXML) finish(l Limits) ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	// Enforce the same explicit token/depth/text bounds as reception.
	doc, err := Parse(w.Bytes(), l)
	if err != nil {
		return nil, err
	}
	return doc.raw, nil
}
