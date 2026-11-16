package sh

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"strings"
)

// FeatureSupport names one complete bitmap. An empty known set and missing
// evidence are distinct. Unknown vendors/lists never imply Notif-Eff support.
type FeatureSupport struct{ VendorID, FeatureListID, FeatureList uint32 }

// PeerFeatureEvidence applies to exactly one command pair and two end terminals.
// The application obtains this evidence from its deployment/terminal contract,
// never by substituting a neighbouring DRA's CER. No expiry/cache is maintained.
type PeerFeatureEvidence struct {
	CommandCode                          uint32
	Requester, Responder                 pdu.Origin
	RequesterKnown, ResponderKnown       bool
	RequesterFeatures, ResponderFeatures []FeatureSupport
	// NotificationReferences explicitly describes a PNR's opaque User-Data use.
	// It does not certify its XML. Empty leaves XML interpretation outside this scope.
	NotificationReferences []uint32
}

// PeerFeatures constructs the only Sh context kind. It owns bounded evidence and
// never learns or modifies support from messages or from previous calls.
func PeerFeatures(e PeerFeatureEvidence) (validation.Context, error) {
	bad := func() (validation.Context, error) { return validation.Context{}, validation.ErrContext }
	if e.CommandCode < 306 || e.CommandCode > 309 {
		return bad()
	}
	for _, o := range []pdu.Origin{e.Requester, e.Responder} {
		if len(o.Host) > 253 || len(o.Realm) > 253 || avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
			return bad()
		}
	}
	s := contextvalue.Sh{Command: e.CommandCode, RequesterHost: strings.Clone(e.Requester.Host), RequesterRealm: strings.Clone(e.Requester.Realm), ResponderHost: strings.Clone(e.Responder.Host), ResponderRealm: strings.Clone(e.Responder.Realm), RequesterKnown: e.RequesterKnown, ResponderKnown: e.ResponderKnown}
	for i, fs := range [][]FeatureSupport{e.RequesterFeatures, e.ResponderFeatures} {
		known := e.RequesterKnown
		if i == 1 {
			known = e.ResponderKnown
		}
		if len(fs) > 4 || !known && len(fs) > 0 {
			return bad()
		}
		for j, f := range fs {
			for _, prior := range fs[:j] {
				if prior.VendorID == f.VendorID && prior.FeatureListID == f.FeatureListID {
					return bad()
				}
			}
			v := contextvalue.Feature{Vendor: f.VendorID, List: f.FeatureListID, Bits: f.FeatureList}
			if i == 0 {
				s.Requester[j] = v
				s.RequesterCount++
			} else {
				s.Responder[j] = v
				s.ResponderCount++
			}
		}
	}
	if len(e.NotificationReferences) > 3 || len(e.NotificationReferences) > 0 && e.CommandCode != 309 {
		return bad()
	}
	for _, ref := range e.NotificationReferences {
		b := refBit(ref)
		if b == 0 || s.NotificationRefs&b != 0 {
			return bad()
		}
		s.NotificationRefs |= b
	}
	return contextvalue.NewSh(s), nil
}
func refBit(ref uint32) uint8 {
	switch ref {
	case 11:
		return 1
	case 12:
		return 2
	case 25:
		return 4
	}
	return 0
}
func hasNotif(fs []SupportedFeatures, mandatory bool) bool {
	for _, f := range fs {
		if f.VendorID == 10415 && f.FeatureListID == 1 && f.FeatureList&1 != 0 && (!mandatory || f.Mandatory) {
			return true
		}
	}
	return false
}
func featureShape(fs []SupportedFeatures, answer bool, mode validation.Mode) error {
	for i, f := range fs {
		if answer && f.Mandatory && mode != validation.Receive {
			return validation.At(key(628), validation.ErrField)
		}
		for _, prior := range fs[:i] {
			if prior.VendorID == f.VendorID && prior.FeatureListID == f.FeatureListID {
				return validation.At(key(628), validation.ErrDuplicate)
			}
		}
	}
	return nil
}
func features(b pdu.Body) ([]SupportedFeatures, pdu.Origin, string, string, bool) {
	switch m := b.(type) {
	case *UDR:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, m.DestinationHost.Value, m.DestinationRealm, false
	case *PUR:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, m.DestinationHost.Value, m.DestinationRealm, false
	case *SNR:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, m.DestinationHost.Value, m.DestinationRealm, len(m.DataReference) > 1
	case *PNR:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, m.DestinationHost, m.DestinationRealm, false
	case *UDA:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, "", "", false
	case *PUA:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, "", "", false
	case *SNA:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, "", "", false
	case *PNA:
		return m.SupportedFeatures, pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, "", "", false
	}
	return nil, pdu.Origin{}, "", "", false
}
func featureResult(a pdu.Body) (bool, bool) {
	var exp avp.Optional[ExperimentalResult]
	var code avp.Optional[uint32]
	switch m := a.(type) {
	case *UDA:
		exp, code = m.ExperimentalResult, m.ResultCode
	case *PUA:
		exp, code = m.ExperimentalResult, m.ResultCode
	case *SNA:
		exp, code = m.ExperimentalResult, m.ResultCode
	case *PNA:
		exp, code = m.ExperimentalResult, m.ResultCode
	}
	return exp.Present && exp.Value.VendorID == 10415 && exp.Value.ExperimentalResultCode == 5011, code.Present && code.Value == 5001
}
func deficit(req, ans []SupportedFeatures) bool {
	for _, f := range req {
		if !f.Mandatory {
			continue
		}
		var bits uint32
		for _, a := range ans {
			if a.VendorID == f.VendorID && a.FeatureListID == f.FeatureListID {
				bits = a.FeatureList
			}
		}
		if f.FeatureList&^bits != 0 {
			return true
		}
	}
	return false
}
func checkFeatureAnswer(r, a pdu.Body) error {
	req, _, _, _, _ := features(r)
	ans, _, _, _, _ := features(a)
	is5011, unknownAVP := featureResult(a)
	if is5011 && (len(ans) == 0 || !deficit(req, ans)) {
		return validation.ErrInvalidAnswer
	}
	// AVP_UNSUPPORTED has separate base diagnostics; it does not recognize 628.
	if len(req) > 0 && !unknownAVP {
		if len(ans) == 0 {
			return validation.ErrInvalidAnswer
		}
		if deficit(req, ans) != is5011 {
			return validation.ErrInvalidAnswer
		}
	}
	return nil
}
func checkUDAFeatures(r *UDR, a *UDA) error { return checkFeatureAnswer(r, a) }
func checkPUAFeatures(r *PUR, a *PUA) error { return checkFeatureAnswer(r, a) }
func checkSNAFeatures(r *SNR, a *SNA) error { return checkFeatureAnswer(r, a) }
func checkPNAFeatures(r *PNR, a *PNA) error { return checkFeatureAnswer(r, a) }
func support(fs [4]contextvalue.Feature, n uint8, v, l uint32) uint32 {
	for _, f := range fs[:n] {
		if f.Vendor == v && f.List == l {
			return f.Bits
		}
	}
	return 0
}
func equalSet(wire []SupportedFeatures, fs [4]contextvalue.Feature, n uint8) bool {
	if len(wire) != int(n) {
		return false
	}
	for _, w := range wire {
		found := false
		for _, f := range fs[:n] {
			if w.VendorID == f.Vendor && w.FeatureListID == f.List && w.FeatureList == f.Bits {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func checkPeerFeatures(request, answer pdu.Body, ctx validation.Context, mode validation.Mode) error {
	// Standalone answers cannot supply their own original-request evidence.
	if request == nil {
		return validation.ErrContextRequired
	}
	req, origin, dh, dr, mixed := features(request)
	ans, ao, _, _, _ := features(answer)
	s, present := ctx.Sh()
	if !present {
		if !ctx.Empty() {
			return validation.ErrContext
		}
		if len(req) == 0 && len(ans) == 0 {
			return nil
		}
		return validation.ErrContextRequired
	}
	if request.CommandKey().Code != s.Command || !strings.EqualFold(origin.Host, s.RequesterHost) || !strings.EqualFold(origin.Realm, s.RequesterRealm) || !strings.EqualFold(dr, s.ResponderRealm) || dh != "" && !strings.EqualFold(dh, s.ResponderHost) {
		return validation.ErrContext
	}
	if answer != nil && (!strings.EqualFold(ao.Host, s.ResponderHost) || !strings.EqualFold(ao.Realm, s.ResponderRealm)) {
		return validation.ErrContext
	}
	if !s.RequesterKnown || !s.ResponderKnown {
		return validation.ErrContextRequired
	}
	for _, f := range req {
		if f.FeatureList&^support(s.Requester, s.RequesterCount, f.VendorID, f.FeatureListID) != 0 {
			return validation.ErrContext
		}
		if len(f.Extra) > 0 {
			return validation.ErrContextRequired
		}
	}
	if s.NotificationRefs != 0 && s.NotificationRefs&(s.NotificationRefs-1) != 0 {
		mixed = true
	}
	if mixed {
		if pnr, ok := request.(*PNR); ok && !pnr.UserIdentity.PublicIdentity.Present {
			return validation.ErrUnsupported
		}
		if !hasNotif(req, answer == nil && mode != validation.Receive) {
			return validation.ErrField
		}
		if support(s.Requester, s.RequesterCount, 10415, 1)&1 == 0 {
			return validation.ErrUnsupported
		}
		// Receiving a mandatory unsupported feature still permits a 5011 response.
		if answer == nil && mode == validation.Send && support(s.Responder, s.ResponderCount, 10415, 1)&1 == 0 {
			return validation.ErrUnsupported
		}
	}
	if answer != nil {
		_, unknownAVP := featureResult(answer)
		if !unknownAVP && !equalSet(ans, s.Responder, s.ResponderCount) {
			return validation.ErrContext
		}
		for _, f := range ans {
			if len(f.Extra) > 0 {
				return validation.ErrContextRequired
			}
		}
	}
	return nil
}
