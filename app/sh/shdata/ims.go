// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata

import (
	"strconv"
	"strings"

	"github.com/skill215/diameterkit/avp"
)

// scscfName selects the ordinary host-only SIP URI used by the frozen examples.
// Full SIP URI syntax, ports/parameters/IPv6/IDNA are not inferred from xs:anyURI.
func scscfName(s string) bool {
	if !strings.HasPrefix(s, "sip:") {
		return false
	}
	host := s[4:]
	if avp.Identity(host) != nil {
		return false
	}
	last := host[strings.LastIndexByte(host, '.')+1:]
	return last[0] >= 'a' && last[0] <= 'z' || last[0] >= 'A' && last[0] <= 'Z'
}

// MarshalIMSUserState constructs exactly one Ref11 value: 0 not registered,
// 1 registered, 2 registered with unregistered services, 3 authentication pending.
// It does not compute this state from subscriptions or suppress transient states.
func MarshalIMSUserState(state uint8, l Limits) ([]byte, error) {
	if !l.valid() {
		return nil, ErrArgument
	}
	if state > 3 {
		return nil, ErrStructure
	}
	return marshalIMS("IMSUserState", strconv.Itoa(int(state)), l)
}

// MarshalSCSCFName constructs exactly one Ref12 value. Empty is explicitly
// present unavailable/removed data; it does not erase a library registration.
// Nonempty values select sip:DNS-host, preserving case and original spelling.
// Parse can retain other XML strings with an incomplete URI interpretation report.
func MarshalSCSCFName(name string, l Limits) ([]byte, error) {
	if !l.valid() {
		return nil, ErrArgument
	}
	if name != "" && !scscfName(name) {
		return nil, ErrUnsupported
	}
	return marshalIMS("SCSCFName", name, l)
}
func marshalIMS(name, value string, l Limits) ([]byte, error) {
	w := boundedXML{max: l.MaxBytes}
	w.tag("<Sh-Data><Sh-IMS-Data>")
	w.leaf(name, value)
	w.tag("</Sh-IMS-Data></Sh-Data>")
	return w.finish(l)
}

// NotificationData is the finite Ref11/12/25 combination; presence is explicit.
// It carries no subscription lookup or automatic one-time/continuous lifecycle.
type NotificationData struct {
	IMSUserState avp.Optional[uint8]
	SCSCFName    avp.Optional[string]
	Reachability avp.Optional[Reachability]
}

func MarshalNotification(n NotificationData, l Limits) ([]byte, error) {
	if !l.valid() {
		return nil, ErrArgument
	}
	if !n.IMSUserState.Present && !n.SCSCFName.Present && !n.Reachability.Present {
		return nil, ErrArgument
	}
	if n.IMSUserState.Present && n.IMSUserState.Value > 3 {
		return nil, ErrStructure
	}
	if n.SCSCFName.Present && n.SCSCFName.Value != "" && !scscfName(n.SCSCFName.Value) {
		return nil, ErrUnsupported
	}
	w := boundedXML{max: l.MaxBytes}
	w.tag("<Sh-Data><Sh-IMS-Data>")
	if n.SCSCFName.Present {
		w.leaf("SCSCFName", n.SCSCFName.Value)
	}
	if n.IMSUserState.Present {
		w.leaf("IMSUserState", strconv.Itoa(int(n.IMSUserState.Value)))
	}
	if n.Reachability.Present {
		w.tag("<Extension><Extension><Extension>")
		if err := writeReachability(&w, n.Reachability.Value); err != nil {
			return nil, err
		}
		w.tag("</Extension></Extension></Extension>")
	}
	w.tag("</Sh-IMS-Data></Sh-Data>")
	return w.finish(l)
}

// CheckNotification checks exactly the supplied data classes, not the entire
// subscription set: a notification may contain only data that actually changed.
// Feature negotiation is a separate explicit Diameter context check.
func (d *Document) CheckNotification(references []uint32) (Report, error) {
	if d == nil || len(references) == 0 || len(references) > 3 {
		return Report{}, ErrArgument
	}
	var bits uint8
	for _, ref := range references {
		var b uint8
		switch ref {
		case 11:
			b = 1
		case 12:
			b = 2
		case 25:
			b = 4
		default:
			return Report{}, ErrArgument
		}
		if bits&b != 0 {
			return Report{}, ErrArgument
		}
		bits |= b
	}
	return d.checkUses(0, bits)
}
