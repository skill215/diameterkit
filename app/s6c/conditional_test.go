// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional SRR omission proof and unchanged validation semantics.
// SPDX-AI-Date: 2026-11-18

package s6c

import (
	"errors"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestConditionalSRRSMEA(t *testing.T) {
	for _, tc := range []struct {
		name           string
		flags          avp.Optional[uint32]
		smea           avp.Optional[[]byte]
		missing, proof bool
	}{
		{"absent-flags", avp.Optional[uint32]{}, avp.Optional[[]byte]{}, false, false},
		{"clear-condition", avp.Some(uint32(2)), avp.Optional[[]byte]{}, false, false},
		{"required", avp.Some(uint32(1)), avp.Optional[[]byte]{}, true, true},
		{"reserved-receive", avp.Some(uint32(0x80000001)), avp.Optional[[]byte]{}, true, true},
		{"minimum-valid", avp.Some(uint32(1)), avp.Some([]byte{1, 0x91, 0xf1}), false, false},
		{"present-empty", avp.Some(uint32(1)), avp.Some([]byte{}), false, false},
		{"short", avp.Some(uint32(1)), avp.Some([]byte{1, 0x91}), false, false},
		{"unsupported-ton", avp.Some(uint32(1)), avp.Some([]byte{1, 0xa1, 0xf1}), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &SRR{MSISDN: avp.Some([]byte{0x21, 0xf3}), SRRFlags: tc.flags, SMRPSMEA: tc.smea}
			err := validateSRRSubset(m, validation.Receive)
			if errors.Is(err, validation.ErrMissing) != tc.missing {
				t.Fatal(err)
			}
			q, ok := missingavp.Quote(err, wire.DefaultLimits(), 16)
			if ok != tc.proof {
				t.Fatal(err, q)
			}
			if tc.proof {
				if len(q) != 16 || wire.Uint24(q[5:8]) != 15 {
					t.Fatal(q)
				}
				// Replacing caller-visible values cannot rewrite the private evidence.
				m.SRRFlags.Value = 0
				if _, ok := missingavp.Quote(err, wire.DefaultLimits(), 16); !ok {
					t.Fatal("borrowed flags")
				}
			}
			if tc.name == "present-empty" || tc.name == "short" {
				if !errors.Is(err, validation.ErrField) {
					t.Fatal(err)
				}
			}
			if tc.name == "unsupported-ton" && !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(err)
			}
			if tc.name == "absent-flags" || tc.name == "clear-condition" || tc.name == "minimum-valid" {
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	for _, mode := range []validation.Mode{validation.Build, validation.Send} {
		m := &SRR{MSISDN: avp.Some([]byte{0x21, 0xf3}), SRRFlags: avp.Some(uint32(1))}
		err := validateSRRSubset(m, mode)
		if !errors.Is(err, validation.ErrMissing) {
			t.Fatal(err)
		}
		if _, ok := missingavp.Quote(err, wire.DefaultLimits(), 16); ok {
			t.Fatal("local construction gained receive evidence")
		}
	}
	m := &SRR{SRRFlags: avp.Some(uint32(1))}
	if _, ok := missingavp.Quote(validateSRRSubset(m, validation.Receive), wire.DefaultLimits(), 16); ok {
		t.Fatal("identity ambiguity inferred")
	}
}
