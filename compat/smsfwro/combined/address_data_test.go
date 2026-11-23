// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Isolated address value NUL diagnostics with bounded original quotes.
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

func TestAddressDataModes(t *testing.T) {
	for _, typ := range []int32{-2147483648, -1, 0, 1, 2147483647} {
		for i, data := range [][]byte{nil, {}, []byte("id"), {0}, {0, 1}, {1, 0, 2}, bytes.Repeat([]byte{0}, 1024)} {
			for _, extra := range []bool{false, true} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%d/%d/%t/%d", typ, i, extra, mode), func(t *testing.T) {
						m := &Address{AddressType: typ, AddressData: data}
						if extra {
							m.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}}}
						}
						e := validateAddressSubset(m, mode)
						bad := bytes.IndexByte(data, 0) >= 0
						var proof interface{ DiameterAddressDataValueV1() string }
						want := typ == 1 && bad && !extra && mode == validation.Receive
						if errors.As(e, &proof) != want {
							t.Fatal(e)
						}
						if typ == 1 && bad {
							var f *validation.FieldError
							if !errors.Is(e, validation.ErrField) || errors.Is(e, validation.ErrUnsupported) || !errors.As(e, &f) || len(f.Path) != 1 || f.Path[0] != (avp.Key{VendorID: 10415, Code: 897}) {
								t.Fatal("classification/path", e)
							}
						}
						if typ != 1 && bad && (!errors.Is(e, validation.ErrUnsupported) || errors.Is(e, validation.ErrField)) {
							t.Fatal("mixed", e)
						}
						if (e == nil) != (typ == 1 && !bad && !extra) {
							t.Fatal("acceptance", e)
						}
						*m = Address{}
						if want && proof.DiameterAddressDataValueV1() != "combined-address-data-nul-v1" {
							t.Fatal("retained state")
						}
					})
				}
			}
		}
	}
}
