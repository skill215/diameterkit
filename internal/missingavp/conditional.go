// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Strict consumer validation for one reviewed conditional minimum.
// SPDX-AI-Date: 2026-11-18

package missingavp

import "github.com/skill215/diameterkit/avp"

// A named conditional proof cannot grant arbitrary widths to ordinary OctetStrings.
// Like other omission evidence, callbacks are trusted schema code, not wire
// authentication. Unsupported producers must fail closed, including old consumers.
type conditionalEvidence interface {
	DiameterMissingConditionalV1() (string, uint32)
}

func conditionalMinimum(proof conditionalEvidence, key avp.Key, flags byte, width int) bool {
	rule, condition := proof.DiameterMissingConditionalV1()
	return rule == "s6c-srr-gprs-smea-v1" && condition&1 != 0 &&
		key == (avp.Key{VendorID: 10415, Code: 3309}) && flags == 192 && width == 3
}
