// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package s6c

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/validation"
)

func TestSMEAInternationalDecimal(t *testing.T) {
	for n := 1; n <= 15; n++ {
		digits := "123456789012345"[:n]
		b, e := EncodeSMEA(digits)
		if e != nil {
			t.Fatal(e)
		}
		got, e := DecodeSMEA(b)
		if e != nil || got != digits || b[0] != byte(n) || b[1] != 0x91 {
			t.Fatal(n, b, got, e)
		}
		if len(b) != 2+(n+1)/2 {
			t.Fatal(b)
		}
		second, _ := EncodeSMEA(digits)
		b[2] = 0
		if bytes.Equal(b, second) {
			t.Fatal("shared builder buffer")
		}
	}
	for _, s := range []string{"", strings.Repeat("1", 16), "+123", "12a", "１２３"} {
		if _, e := EncodeSMEA(s); !errors.Is(e, validation.ErrField) {
			t.Fatal(s, e)
		}
	}
	for _, b := range [][]byte{nil, {1, 0x91}, {0, 0x91, 0xf1}, {1, 0x11, 0xf1}, {1, 0x91, 0x21}, {2, 0x91, 0xf1}, {1, 0x91, 0xfa}, {1, 0x91, 0xf1, 0x00}} {
		if _, e := DecodeSMEA(b); !errors.Is(e, validation.ErrField) {
			t.Fatal(b, e)
		}
	}
	if _, e := DecodeSMEA([]byte{3, 0xa1, 0x21, 0xf3}); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	b, e := EncodeSMEA("12345")
	if e != nil || !bytes.Equal(b, []byte{5, 0x91, 0x21, 0x43, 0xf5}) {
		t.Fatal(b, e)
	}
}
