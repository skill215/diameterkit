// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional whole MMS group omission and bounded diagnostic acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Called only at the existing SMS Receive whole-MMS omission check.
// The V2 plan includes the structural Originator-Address and all six conditional fields, including the
// Recipient-Address group. Empty octets and zero scalars are minimum examples,
// never inferred address, timestamp, priority, key or status values.
func missingSMSMMSGroup() error {
	const plan = "\x00\x00\x28\xaf\x00\x00\x03\x6d\xc0\x00\x07\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x76\xc0\x00\x02\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x83\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x81\xc0\x00\x00\x05" +
		"\x00\x00\x28\xaf\x00\x00\x04\xb1\xc0\x00\x02\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x83\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x03\x81\xc0\x00\x00\x05" +
		"\x00\x00\x28\xaf\x00\x00\x04\xb2\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x04\xb9\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x04\xbc\xc0\x04\x00\x00" +
		"\x00\x00\x28\xaf\x00\x00\x04\xba\xc0\x00\x00\x05" +
		"\x00\x00\x28\xaf\x00\x00\x04\xc0\xc0\x04\x00\x00"
	mms := &profileGroupOmission{key: avp.Key{VendorID: 10415, Code: 877}, flags: 192, plan: plan}
	service := &profileGroupOmission{key: avp.Key{VendorID: 10415, Code: 873}, flags: 192, child: mms}
	return &validation.FieldError{Path: []avp.Key{service.key, mms.key}, Rule: "wire", Cause: service}
}
