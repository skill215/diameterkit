// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Conditional EVENT Requested-Action omission and bounded acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func actionRequest(t *testing.T, kind string) *CCR {
	t.Helper()
	switch kind {
	case "initial":
		return NewInitial(ECURInput{})
	case "term":
		return NewTermination(TerminationInput{})
	case "iec":
		return NewIEC(IECInput{})
	case "general":
		return NewDirectDebit(GeneralInput{})
	case "balance":
		return NewBalanceCheck(GeneralInput{})
	case "refund":
		m, e := NewRefund(RefundInput{RefundInformation: avp.Some([]byte{})})
		if e != nil {
			t.Fatal(e)
		}
		return m
	}
	t.Fatal(kind)
	return nil
}
func TestActionMissingModes(t *testing.T) {
	for _, kind := range []string{"initial", "term", "iec", "general", "balance", "refund"} {
		for _, value := range []uint32{0, 1, 2, 3, ^uint32(0)} {
			for _, present := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%t/%d", kind, value, present, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						m.RequestedAction = avp.Optional[uint32]{Present: present, Value: value}
						event := kind != "initial" && kind != "term"
						proof := event && !present && mode == validation.Receive
						valid := !event && !present || present && (kind == "iec" && value == 0 || (kind == "general" || kind == "balance") && (value == 0 || value == 2) || kind == "refund" && value == 1)
						err := validateCCRSubset(m, mode)
						if (err == nil) != valid || errors.Is(err, validation.ErrMissing) != proof {
							t.Fatal(err)
						}
						unsupported := event && !present || !event && present || kind == "iec" && value != 0 || event && value > 2
						if !valid && unsupported && !errors.Is(err, validation.ErrUnsupported) {
							t.Fatal("old classification", err)
						}
						q, ok := missingavp.Quote(err, wire.DefaultLimits(), 12)
						if ok != proof || ok && len(q) != 12 {
							t.Fatal(err, q)
						}
						if proof {
							if !errors.Is(err, validation.ErrUnsupported) {
								t.Fatal(err)
							}
							var field *validation.FieldError
							if !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 436}) {
								t.Fatal(err)
							}
							*m = CCR{}
							if q, ok := missingavp.Quote(err, wire.DefaultLimits(), 12); !ok || len(q) != 12 {
								t.Fatal("borrowed request")
							}
						}
					})
				}
			}
		}
	}
}
