// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated equipment value NUL diagnostics with bounded original quotes.
// SPDX-AI-Date: 2026-11-23

package combined

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestEquipmentDataModes(t *testing.T) {
	for _, typ := range []int32{-2147483648, -1, 0, 1, 2147483647} {
		for i, data := range [][]byte{nil, {}, []byte("id"), {0}, {0, 1}, {1, 0, 2}, bytes.Repeat([]byte{0}, 1024)} {
			for _, extra := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%d/%d/%t/%d", typ, i, extra, mode), func(t *testing.T) {
						m := &UserEquipmentInfo{UserEquipmentInfoType: typ, UserEquipmentInfoValue: data}
						if extra {
							m.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}}}
						}
						e := validateUserEquipmentInfoSubset(m, mode)
						bad := bytes.IndexByte(data, 0) >= 0
						var proof interface{ DiameterEquipmentDataValueV1() string }
						want := typ == 0 && bad && !extra && mode == validation.Receive
						if errors.As(e, &proof) != want {
							t.Fatal(e)
						}
						if typ == 0 && bad {
							var f *validation.FieldError
							if !errors.Is(e, validation.ErrField) || errors.Is(e, validation.ErrUnsupported) || !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{Code: 460}) {
								t.Fatal("classification/path", e)
							}
						}
						if typ != 0 && bad && (!errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrField)) {
							t.Fatal("mixed", e)
						}
						if (e == nil) != (typ == 0 && !bad && !extra) {
							t.Fatal("acceptance", e)
						}
						*m = UserEquipmentInfo{}
						if want && proof.DiameterEquipmentDataValueV1() != "combined-equipment-data-nul-v1" {
							t.Fatal("retained state")
						}
					})
				}
			}
		}
	}
}
