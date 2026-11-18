// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed conditional Recipient-Address omission and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Only the existing SMS receive rule, after its earlier checks and with all
// other conditional MMS fields present, constructs this diagnostic.
// The V2 plan is Recipient-Address -> Int32 Address-Type, empty octets Address-Data.
// It describes minimum sizes, not a valid address or a missing business value.
func missingSMSRecipientAddress() error {
	const plan = "\x00\x00\x28\xaf\x00\x00\x04\xb1\xc0\x00\x02\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x83\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x81\xc0\x00\x00\x05"
	leaf := &profileGroupOmission{key: avp.Key{VendorID: 10415, Code: 1201}, flags: 192, plan: plan}
	mms := &profileGroupOmission{key: avp.Key{VendorID: 10415, Code: 877}, flags: 192, child: leaf}
	service := &profileGroupOmission{key: avp.Key{VendorID: 10415, Code: 873}, flags: 192, child: mms}
	return &validation.FieldError{Path: []avp.Key{service.key, mms.key, leaf.key}, Rule: "wire", Cause: service}
}
