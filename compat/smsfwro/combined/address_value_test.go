// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated combined Address-Type value acceptance.
// SPDX-AI-Date: 2026-11-19

package combined

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestAddressValueModes(t *testing.T) {
	for _, value := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
		for _, data := range [][]byte{nil, []byte("123"), {0}, {'1', 0, '2'}} {
			for extra := 0; extra < 4; extra++ {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%d/%x/%d/%d", value, data, extra, mode), func(t *testing.T) {
						m := &Address{AddressType: value, AddressData: data}
						if extra > 0 {
							flags := byte(0)
							key := avp.Key{Code: 9999}
							if extra == 2 {
								flags = 64
							}
							if extra == 3 {
								key = avp.Key{VendorID: 10415, Code: 899}
								flags = 192
							}
							m.Extra = []avp.Raw{{Key: key, Flags: flags}}
						}
						err := validateAddress(m, mode)
						want := value != 1 && cstring(data) == nil && extra == 0 && mode == validation.Receive
						valid := value == 1 && cstring(data) == nil && extra == 0
						var proof interface{ DiameterAddressTypeValueV1() string }
						var f *validation.FieldError
						if (err == nil) != valid || errors.As(err, &proof) != want {
							t.Fatal("scope", err)
						}
						if want {
							if !errors.Is(err, validation.ErrUnsupported) || !errors.Is(err, validation.ErrField) || !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 899}) {
								t.Fatal(err)
							}
							*m = Address{}
							if proof.DiameterAddressTypeValueV1() != "combined-address-type-v1" {
								t.Fatal("retained data")
							}
						}
						if value != 1 && extra < 2 && !want {
							if !errors.Is(err, validation.ErrUnsupported) || errors.Is(err, validation.ErrField) || !errors.As(err, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 899}) {
								t.Fatal("legacy first failure", err)
							}
						}
					})
				}
			}
		}
	}
}
func TestAddressValueBuilderModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term", "general", "balance"} {
		for _, recipient := range []bool{false, true} {
			if recipient && (kind == "general" || kind == "balance") {
				continue
			}
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				m := actionRequest(t, kind)
				m.SessionID = "address-mode;1"
				m.ServiceContextID = "sms@example"
				m.OriginHost = "ctf.example"
				m.OriginRealm = "example"
				m.DestinationRealm = "example"
				code := uint32(886)
				a := &m.ServiceInformation.MMSInformation.Value.OriginatorAddress
				if recipient {
					a = &m.ServiceInformation.MMSInformation.Value.RecipientAddress.Value
					code = 1201
				}
				a.AddressType = -1
				err := validateCCR(m, mode)
				var f *validation.FieldError
				var proof interface{ DiameterAddressTypeValueV1() string }
				if !errors.As(err, &f) || len(f.Path) != 4 || errors.As(err, &proof) != (mode == validation.Receive) {
					t.Fatal(kind, recipient, err)
				}
				for i, c := range []uint32{873, 877, code, 899} {
					if f.Path[i] != (avp.Key{VendorID: 10415, Code: c}) {
						t.Fatal(f.Path)
					}
				}
			}
		}
	}
}
