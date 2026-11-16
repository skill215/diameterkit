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
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

// MOInterface identifies the configured end-to-end interface, never a DRA hop.
// SGd and Gdd cannot be distinguished by their shared Application-Id.
type MOInterface uint8

const (
	MOInterfaceSGd MOInterface = iota + 1
	MOInterfaceGdd
)

// MOForwardingEvidence binds an ordinary known-IMSI sender to an SMS-IWMSC.
// SenderMSISDN includes presence; decimal digits preserve leading zeroes.
// Endpoint identities are bounded to253 bytes. This is an application assertion,
// not verification of a subscriber database, physical interface or subscription.
type MOForwardingEvidence struct {
	Interface            MOInterface
	Requester, Responder pdu.Origin
	SenderIMSI           string
	SenderMSISDN         avp.Optional[string]
	SCAddress            string
}

// MOForwarding supplies owned evidence for OFR/OFA. Normal experimental5001
// requires SGd and an original MSISDN. Explicit Gdd keeps the received answer
// owned but reports ErrUnsupported; it does not silently reinterpret the result.
func MOForwarding(e MOForwardingEvidence) (validation.Context, error) {
	if e.Interface != MOInterfaceSGd && e.Interface != MOInterfaceGdd {
		return validation.Context{}, validation.ErrContext
	}
	// Reuse only the bounded identity validation/storage of T4; kind7 selects
	// separate semantics and never asserts an External-Identifier mapping.
	v, err := T4Charging(T4ChargingEvidence{Requester: e.Requester, Responder: e.Responder, SenderIMSI: e.SenderIMSI, SenderMSISDN: e.SenderMSISDN, SCAddress: e.SCAddress})
	if err != nil {
		return validation.Context{}, err
	}
	identity, _ := v.T4Charging()
	identity.Interface = uint8(e.Interface)
	return contextvalue.NewMOForwarding(identity), nil
}

func moUnknown(a *OFA) bool {
	return a.ExperimentalResult.Present && a.ExperimentalResult.Value.VendorID == 10415 && a.ExperimentalResult.Value.ExperimentalResultCode == 5001
}

func checkMOForwarding(r *OFR, answer pdu.Body, e contextvalue.T4Charging) error {
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
	n, err := avp.DecodeTBCD(r.SCAddress)
	if err != nil || n != e.SCAddress {
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
	if moUnknown(a) && e.Interface != uint8(MOInterfaceSGd) {
		return validation.At(avp.Key{Code: 297}, validation.ErrUnsupported)
	}
	if a.ExternalIdentifier.Present {
		return validation.At(smsKey(3111), validation.ErrUnsupported)
	}
	return nil
}
