// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded isolated MMS SMSKey NUL-value diagnostics.
// SPDX-AI-Date: 2026-11-23

package combined

import "github.com/skill215/diameterkit/validation"

// Static evidence retains no key, address or request data.
type mmsKeyValue struct{}

func (mmsKeyValue) Error() string                 { return validation.ErrField.Error() }
func (mmsKeyValue) Unwrap() error                 { return validation.ErrField }
func (mmsKeyValue) DiameterMMSKeyValueV1() string { return "combined-mms-key-nul-v1" }
