package diameter

import "github.com/skill215/diameterkit/pdu"

// ApplicationBinding chooses a reviewed mapping between a command's header and
// its negotiated capability. It neither registers a profile nor advertises it.
type ApplicationBinding uint8

const (
	// HeaderApplication retains the existing nonzero header-application matching.
	HeaderApplication ApplicationBinding = iota
	// SCAP1Accounting enables only the legacy App0/271 command with a strictly
	// negotiated Vendor193/Acct19302 VSAI on both sides. Auth19302, unscoped Acct19302,
	// another Vendor or a dual Auth/Acct VSAI do not qualify. It grants no replay.
	// Outbound requires one fixed destination host/realm and matching terminal.
	SCAP1Accounting
)

func hasSCAP1Accounting(vs []pdu.VendorSpecificApplicationID) bool {
	for _, v := range vs {
		if v.VendorID == 193 && !v.AuthApplicationID.Present && v.AcctApplicationID.Present && v.AcctApplicationID.Value == 19302 {
			return true
		}
	}
	return false
}
func validApplicationBinding(k pdu.CommandKey, b ApplicationBinding, local LocalCapabilities) bool {
	switch b {
	case HeaderApplication:
		return k.ApplicationID != 0
	case SCAP1Accounting:
		return k.ApplicationID == 0 && k.Code == 271 && hasSCAP1Accounting(local.VendorSpecificApplicationID)
	default:
		return false
	}
}

// Published with the established connection and immutable thereafter. The SCAP1
// bit is separate from the numeric header map: it never makes arbitrary App0
// commands available or turns an Auth advertisement into an Acct capability.
type negotiatedApplications struct {
	byID            map[uint32]bool
	scap1Accounting bool
}

func (a *negotiatedApplications) allows(k pdu.CommandKey, b ApplicationBinding) bool {
	if a == nil {
		return false
	}
	if b == SCAP1Accounting {
		return k.ApplicationID == 0 && k.Code == 271 && a.scap1Accounting
	}
	return b == HeaderApplication && k.ApplicationID != 0 && a.byID[k.ApplicationID]
}
func (a *negotiatedApplications) usable(p *Peer) bool {
	if !p.scap1Accounting {
		return len(a.byID) > 0
	}
	// A peer opting into SCAP1 needs at least one usable configured command. The
	// legacy numeric map alone could otherwise count a wrong-kind19302 match.
	for k, v := range p.policies {
		if a.allows(k, v.ApplicationBinding) {
			return true
		}
	}
	for k, v := range p.inboundPolicies {
		if a.allows(k, v.ApplicationBinding) {
			return true
		}
	}
	return false
}
func (n *Node) negotiateApplications(p *Peer, auth, acct []uint32, vs []pdu.VendorSpecificApplicationID) *negotiatedApplications {
	return &negotiatedApplications{byID: n.commonApps(auth, acct, vs), scap1Accounting: p.scap1Accounting && hasSCAP1Accounting(n.cfg.Capabilities.VendorSpecificApplicationID) && hasSCAP1Accounting(vs)}
}
