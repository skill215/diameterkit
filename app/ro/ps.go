// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package ro

import "github.com/skill215/diameterkit/validation"

func validateEquipmentSubset(m *UserEquipmentInfo, _ validation.Mode) error {
	if m.UserEquipmentInfoType != 0 {
		return at(459, validation.ErrUnsupported)
	}
	if len(m.UserEquipmentInfoValue) != 16 {
		return at(460, validation.ErrUnsupported)
	}
	for _, c := range m.UserEquipmentInfoValue {
		if c < '0' || c > '9' {
			return at(460, validation.ErrUnsupported)
		}
	}
	return extras(m.Extra)
}
func roPLMN(b []byte) bool {
	return b[0]&15 <= 9 && b[0]>>4 <= 9 && b[1]&15 <= 9 && (b[1]>>4 <= 9 || b[1]>>4 == 15) && b[2]&15 <= 9 && b[2]>>4 <= 9
}
func validatePSSubset(m *PSInformation, mode validation.Mode) error {
	if m.RATType.Present {
		b := m.RATType.Value
		if len(b) != 1 {
			return smsAt(21, validation.ErrField)
		}
		if b[0] < 1 || b[0] > 9 {
			return smsAt(21, validation.ErrUnsupported)
		}
	}
	if m.UserLocationInfo.Present {
		b := m.UserLocationInfo.Value
		if len(b) == 0 {
			return smsAt(22, validation.ErrField)
		}
		n := 0
		switch b[0] {
		case 128:
			n = 6
		case 129:
			n = 8
		default:
			return smsAt(22, validation.ErrUnsupported)
		}
		if len(b) != n || !roPLMN(b[1:4]) {
			return smsAt(22, validation.ErrField)
		}
		if b[0] == 129 && b[4]&0xf0 != 0 && mode != validation.Receive {
			return smsAt(22, validation.ErrUnsupported)
		}
	}
	if m.MSTimeZone.Present {
		b := m.MSTimeZone.Value
		if len(b) != 2 || b[0]>>4 > 9 {
			return smsAt(23, validation.ErrField)
		}
		if b[1]&3 == 3 || mode != validation.Receive && b[1]&0xfc != 0 {
			return smsAt(23, validation.ErrUnsupported)
		}
	}
	return extras(m.Extra)
}
