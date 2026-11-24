// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated Reporting-Reason and requested-unit value acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func unitModeCheck(t *testing.T, m *CCR, mode validation.Mode, want, valid, reason bool, legacy, excluded uint32) {
	t.Helper()
	err := validateCCRSubset(m, mode)
	var proof interface{ DiameterUnitValueV1() string }
	if errors.As(err, &proof) != want || (err == nil) != valid {
		t.Fatal("scope/acceptance", err)
	}
	if want {
		var f *validation.FieldError
		path := []avp.Key{{Code: 456}, {Code: 437}, {Code: 417}}
		name := "combined-sms-requested-units-v1"
		if reason {
			path = []avp.Key{{Code: 456}, {VendorID: 10415, Code: 872}}
			name = "combined-termination-reporting-reason-v1"
		}
		if !errors.Is(err, validation.ErrField) || !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrMissing) || !errors.As(err, &f) || len(f.Path) != len(path) {
			t.Fatal(err)
		}
		for i, k := range path {
			if f.Path[i] != k {
				t.Fatal(f.Path)
			}
		}
		*m = CCR{}
		if proof.DiameterUnitValueV1() != name {
			t.Fatal("retained state")
		}
	} else if excluded != 0 {
		var f *validation.FieldError
		var marker interface{ DiameterExcludedECURUnitV1() }
		if !errors.As(err, &marker) || !errors.Is(err, validation.ErrUnsupported) || !errors.As(err, &f) || len(f.Path) != 2 || f.Path[0] != (avp.Key{Code: 456}) || f.Path[1] != (avp.Key{Code: excluded}) {
			t.Fatal(err)
		}
	} else if !valid {
		var f *validation.FieldError
		if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{Code: legacy}) {
			t.Fatal("legacy path/class", err)
		}
	}
}
func TestUnitValueModes(t *testing.T) {
	for _, value := range []int32{-2147483648, -1, 0, 1, 2, 3, 2147483647} {
		for mask := 0; mask < 8; mask++ {
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				t.Run(fmt.Sprintf("reason/%d/%d/%d", value, mask, mode), func(t *testing.T) {
					m := actionRequest(t, "term")
					u := &m.MultipleServicesCreditControl
					u.RequestedServiceUnit.Present = mask&1 != 0
					u.UsedServiceUnit.Present = mask&2 != 0
					u.ReportingReason = avp.Optional[int32]{Present: mask&4 != 0, Value: value}
					var excluded uint32
					if mode == validation.Receive && mask == 7 && value == 2 {
						excluded = 437
					}
					unitModeCheck(t, m, mode, mode == validation.Receive && mask == 6 && value != 2, mask == 6 && value == 2, true, 456, excluded)
				})
			}
		}
	}
	for _, kind := range []string{"iec", "initial"} {
		for _, value := range []uint64{0, 1, 2, 1 << 32, 1 << 63, ^uint64(0)} {
			for mask := 0; mask < 32; mask++ {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%d/%d", kind, value, mask, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						u := &m.MultipleServicesCreditControl
						u.RequestedServiceUnit.Present = mask&1 != 0
						u.UsedServiceUnit.Present = mask&2 != 0
						u.ReportingReason.Present = mask&4 != 0
						r := &u.RequestedServiceUnit.Value
						r.CCServiceSpecificUnits = value
						r.Money.Present = mask&8 != 0
						r.StandardMoney.Present = mask&16 != 0
						legacy := uint32(456)
						if mask&7 == 1 {
							legacy = 437
						}
						var excluded uint32
						if kind == "initial" && mode == validation.Receive && mask == 3 && value == 1 {
							excluded = 446
						}
						unitModeCheck(t, m, mode, mode == validation.Receive && mask == 1 && value != 1, mask == 1 && value == 1, false, legacy, excluded)
					})
				}
			}
		}
	}
}
func TestUnitValueMoneyBoundary(t *testing.T) {
	for _, kind := range []string{"general", "balance", "refund"} {
		for _, value := range []uint64{0, 1, 2, ^uint64(0)} {
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				t.Run(fmt.Sprintf("%s/%d/%d", kind, value, mode), func(t *testing.T) {
					m := actionRequest(t, kind)
					m.MultipleServicesCreditControl.RequestedServiceUnit.Value.CCServiceSpecificUnits = value
					err := validateCCRSubset(m, mode)
					var proof interface{ DiameterUnitValueV1() string }
					if err != nil || errors.As(err, &proof) {
						t.Fatal("money changed", err)
					}
				})
			}
		}
	}
}
