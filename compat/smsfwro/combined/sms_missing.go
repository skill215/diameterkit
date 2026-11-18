// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed isolated combined SMS condition diagnostics and acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

type smsOmissionField struct {
	present bool
	code    uint32
}

// Called only by the corresponding existing Receive condition. The fixed list
// describes that condition, not every optional field in SMSInformation.
func missingSMSCondition(fields []smsOmissionField) error {
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
	width := 0
	kind := "octets"
	switch code {
	case 2011, 2012:
		width = 4
		kind = ""
	case 2018, 2010, 2017, 2015, 2014:
	default:
		return nil
	}
	leaf := &profileOmission{key: avp.Key{VendorID: 10415, Code: code}, flags: 192, width: width, wireType: kind}
	sms := &profileOmission{key: avp.Key{VendorID: 10415, Code: 2000}, flags: 192, child: leaf}
	service := &profileOmission{key: avp.Key{VendorID: 10415, Code: 873}, flags: 192, child: sms}
	return &validation.FieldError{Path: []avp.Key{service.key, sms.key, leaf.key}, Rule: "wire", Cause: service}
}
