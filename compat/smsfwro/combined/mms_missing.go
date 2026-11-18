// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Selected isolated combined MMS omissions and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Only called at the existing SMS Receive condition. Multiple missing fields
// retain the original coarse failure; no missing sibling or business value is inferred.
func missingSMSMMSField(m *MMSInformation) error {
	fields := [...]struct {
		present bool
		code    uint32
	}{
		{m.RecipientAddress.Present, 1201}, {m.SubmissionTime.Present, 1202},
		{m.Priority.Present, 1209}, {m.UserDataLength.Present, 1212},
		{m.SMSKey.Present, 1210}, {m.StatusReportRequest.Present, 1216},
	}
	count := 0
	var code uint32
	for _, f := range fields {
		if !f.present {
			count++
			code = f.code
		}
	}
	if count != 1 {
		return nil
	}
	if code == 1201 {
		return missingSMSRecipientAddress()
	}
	width := 4
	kind := ""
	if code == 1210 {
		width = 0
		kind = "octets"
	}
	leaf := &profileOmission{key: avp.Key{VendorID: 10415, Code: code}, flags: 192, width: width, wireType: kind}
	mms := &profileOmission{key: avp.Key{VendorID: 10415, Code: 877}, flags: 192, child: leaf}
	service := &profileOmission{key: avp.Key{VendorID: 10415, Code: 873}, flags: 192, child: mms}
	return &validation.FieldError{Path: []avp.Key{service.key, mms.key, leaf.key}, Rule: "wire", Cause: service}
}
