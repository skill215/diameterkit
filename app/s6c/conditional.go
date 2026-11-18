// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Private S6c SRR flag-conditioned omission evidence.
// SPDX-AI-Date: 2026-11-18

package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// Only the Receive branch of the reviewed SRR condition constructs this proof.
// It retains the decoded flag word, never subscriber data or request buffers.
type srrSMEAMissing struct{ flags uint32 }

func (*srrSMEAMissing) Error() string { return validation.ErrMissing.Error() }
func (*srrSMEAMissing) Unwrap() error { return validation.ErrMissing }
func (*srrSMEAMissing) DiameterMissingAVPV1() (avp.Key, byte, int, error) {
	// DecodeSMEA's selected syntax needs a count, TOA and at least one digit byte.
	// Zero-filled diagnostics are minimum-size examples, not valid addresses.
	return smsKey(3309), 192, 3, nil
}
func (e *srrSMEAMissing) DiameterMissingConditionalV1() (string, uint32) {
	return "s6c-srr-gprs-smea-v1", e.flags
}
