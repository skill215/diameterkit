// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined User-Equipment-Info-Type value evidence.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/validation"
)

// Static evidence only; no request bytes or equipment identity is retained.
type equipmentTypeValue struct{}

func (equipmentTypeValue) Error() string { return validation.ErrUnsupported.Error() }
func (equipmentTypeValue) Unwrap() error {
	return errors.Join(validation.ErrUnsupported, validation.ErrField)
}
func (equipmentTypeValue) DiameterEquipmentTypeValueV1() string { return "combined-equipment-type-v1" }
