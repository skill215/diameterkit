// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated SMS address value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

type smsDataValue struct{}

func (smsDataValue) Error() string                  { return validation.ErrField.Error() }
func (smsDataValue) Unwrap() error                  { return validation.ErrField }
func (smsDataValue) DiameterSMSDataValueV1() string { return "combined-sms-data-nul-v1" }

// Named evidence retains no address bytes and does not parse client wire octets.
func diagnoseSMSData(m *SMSInformation) error {
	if len(m.Extra) != 0 {
		return nil
	}
	count := 0
	var code uint32
	for i, v := range [...]avp.Optional[[]byte]{m.OriginatorSCCPAddress, m.RecipientSCCPAddress, m.SMSCAddress} {
		if v.Present && cstring(v.Value) != nil {
			count++
			code = [...]uint32{2008, 2010, 2017}[i]
		}
	}
	if count != 1 {
		return nil
	}
	return vendorAt(code, smsDataValue{})
}
