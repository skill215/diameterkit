// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Multiple-Services-Indicator value evidence.
// SPDX-AI-Date: 2026-11-19

package combined

import "github.com/skill215/diameterkit/validation"

// Static evidence only; no request, subscriber value or persistent state.
type indicatorValue struct{}

func (indicatorValue) Error() string                    { return validation.ErrField.Error() }
func (indicatorValue) Unwrap() error                    { return validation.ErrField }
func (indicatorValue) DiameterIndicatorValueV1() string { return "combined-sms-indicator-v1" }
