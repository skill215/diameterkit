// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit SCAP1 ordered omission diagnostics and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package scap1

import "github.com/skill215/diameterkit/validation"

type parameterOmission struct{ count int }

func (*parameterOmission) Error() string { return validation.ErrMissing.Error() }
func (*parameterOmission) Unwrap() error { return validation.ErrMissing }
func (e *parameterOmission) DiameterMissingCardinalityV1() (string, int) {
	return "scap1-parameters-v1", e.count
}

// Nil preserves the original decoder minimum and its validation fallback.
func diagnoseACRParameters(m *ACR) error {
	types := [...]int32{0, 1, 5, 8, 14, 16, 17, 18, 2, 3, 13}
	if len(m.ServiceParameterInfo) >= len(types) {
		return nil
	}
	previous := -1
	for _, v := range m.ServiceParameterInfo {
		if validateServiceParameterSubset(&v, validation.Receive) != nil {
			return nil
		}
		position := -1
		for i, typ := range types {
			if v.ParameterType == typ {
				position = i
				break
			}
		}
		if position <= previous {
			return nil
		}
		previous = position
	}
	return at(607, 193, &parameterOmission{count: len(m.ServiceParameterInfo)})
}
