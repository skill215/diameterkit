// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata

// Fixed child order from the pinned ShDataType_Rel18.xsd. Opaque known children
// preserve their outer order/multiplicity but make no claim about their content.
// No runtime schema loading, reflection, descendant-name search or plugin hooks.
type shape uint8

const (
	opaque shape = iota
	rootShape
	ims
	ext1
	ext2
	ext3
	sms
	smsExt
	smsSBI
	reach
	reachExt
	reachAMF
	numberLeaf
	scLeaf
	sbiLeaf
	mmeLeaf
	sgsnLeaf
	amfLeaf
	nonAMFLeaf
	imsStateLeaf
	scscfLeaf
)

type child struct {
	name   string
	shape  shape
	repeat bool
}

func children(s shape) ([]child, bool) {
	switch s {
	case rootShape:
		return []child{{"PublicIdentifiers", opaque, false}, {"RepositoryData", opaque, true}, {"Sh-IMS-Data", ims, false}, {"CSLocationInformation", opaque, false}, {"PSLocationInformation", opaque, false}, {"CSUserState", opaque, false}, {"PSUserState", opaque, false}, {"Extension", opaque, false}}, true
	case ims:
		return []child{{"SCSCFName", scscfLeaf, false}, {"IFCs", opaque, false}, {"IMSUserState", imsStateLeaf, false}, {"ChargingInformation", opaque, false}, {"Extension", ext1, false}}, true
	case ext1:
		return []child{{"PSIActivation", opaque, false}, {"Extension", ext2, false}}, false
	case ext2:
		return []child{{"DSAI", opaque, true}, {"Extension", ext3, false}}, false
	case ext3:
		return []child{{"ServiceLevelTraceInfo", opaque, false}, {"IPv4Address", opaque, false}, {"IPv6Prefix", opaque, false}, {"IPv6InterfaceIdentifier", opaque, false}, {"ServicePriorityLevel", opaque, false}, {"UEReachabilityForIP", reach, false}, {"SMSRegistrationInfo", sms, false}, {"Extension", opaque, false}}, false
	case sms:
		return []child{{"IP-SM-GW-Number", numberLeaf, false}, {"Extension", smsExt, false}}, true
	case smsExt:
		return []child{{"SCAddress", scLeaf, false}, {"Extension", smsSBI, false}}, true
	case smsSBI:
		return []child{{"IP-SM-GW-SBISupportIndication", sbiLeaf, false}}, false
	case reach:
		return []child{{"UEIPReachabilityMME", mmeLeaf, false}, {"Extension", reachExt, false}}, true
	case reachExt:
		return []child{{"UEIPReachabilitySGSN", sgsnLeaf, false}, {"Extension", reachAMF, false}}, true
	case reachAMF:
		return []child{{"UEIPReachabilityAMF3GPP", amfLeaf, false}, {"UEIPReachabilityAMFnon3GPP", nonAMFLeaf, false}, {"Extension", opaque, false}}, true
	}
	return nil, false
}
