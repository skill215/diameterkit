// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Named group-constraint evidence without retained request bytes.
// SPDX-AI-Date: 2026-11-17

package base

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

type groupConstraint struct {
	key   avp.Key
	rule  string
	cause error
}

func (e *groupConstraint) Error() string { return e.rule }
func (e *groupConstraint) Unwrap() error { return e.cause }
func (e *groupConstraint) DiameterGroupAVPV1() (avp.Key, byte, int, string, error) {
	return e.key, 0, 0, e.rule, nil
}
func constraintFailure(rule string, key avp.Key, cause error) error {
	proof := &groupConstraint{key: key, rule: rule, cause: cause}
	if key == (avp.Key{}) {
		proof.cause = errors.Join(validation.ErrField, cause)
		return &validation.FieldError{Rule: "wire", Cause: proof}
	}
	return validation.At(key, proof)
}
