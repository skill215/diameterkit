// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated equipment value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import "github.com/skill215/diameterkit/validation"

// Static evidence retains no equipment identity or request.
type equipmentDataValue struct{}

func (equipmentDataValue) Error() string { return validation.ErrField.Error() }
func (equipmentDataValue) Unwrap() error { return validation.ErrField }
func (equipmentDataValue) DiameterEquipmentDataValueV1() string {
	return "combined-equipment-data-nul-v1"
}
