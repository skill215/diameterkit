// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined equipment-type value acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestEquipmentValueModes(t *testing.T) {
	for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
		for _, data := range [][]byte{nil, {}, []byte("123"), {0}, {'1', 0, '2'}} {
			for extra := 0; extra < 4; extra++ {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%d/%x/%d/%d", value, data, extra, mode), func(t *testing.T) {
						m := &UserEquipmentInfo{UserEquipmentInfoType: value, UserEquipmentInfoValue: data}
						if extra > 0 {
							flags := byte(0)
							code := uint32(9999)
							if extra == 2 {
								flags = 64
							}
							if extra == 3 {
								code = 459
							}
							m.Extra = []avp.Raw{{Key: avp.Key{Code: code}, Flags: flags}}
						}
						err := validateUserEquipmentInfo(m, mode)
						want := value != 0 && cstring(data) == nil && extra == 0 && mode == validation.Receive
						valid := value == 0 && cstring(data) == nil && extra == 0
						var proof interface{ DiameterEquipmentTypeValueV1() string }
						var field *validation.FieldError
						if (err == nil) != valid || errors.As(err, &proof) != want {
							t.Fatal("scope", err)
						}
						if want {
							if !errors.Is(err, validation.ErrUnsupported) || !errors.Is(err, validation.ErrField) || !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 459}) {
								t.Fatal(err)
							}
							*m = UserEquipmentInfo{}
							if proof.DiameterEquipmentTypeValueV1() != "combined-equipment-type-v1" {
								t.Fatal("retained input")
							}
						}
						if value != 0 && extra < 2 && !want {
							if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &field) || len(field.Path) != 1 || field.Path[0] != (avp.Key{Code: 459}) {
								t.Fatal("legacy first failure", err)
							}
						}
					})
				}
			}
		}
	}
}
func TestEquipmentValueBuilderModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, kind)
			m.SessionID = "equipment-mode;1"
			m.ServiceContextID = "sms@example"
			m.OriginHost = "ctf.example"
			m.OriginRealm = "example"
			m.DestinationRealm = "example"
			m.UserEquipmentInfo.Value.UserEquipmentInfoType = -1
			err := validateCCR(m, mode)
			var f *validation.FieldError
			var proof interface{ DiameterEquipmentTypeValueV1() string }
			if !errors.As(err, &f) || len(f.Path) != 2 || f.Path[0] != (avp.Key{Code: 458}) || f.Path[1] != (avp.Key{Code: 459}) || errors.As(err, &proof) != (mode == validation.Receive) {
				t.Fatal(err)
			}
		}
	}
}
