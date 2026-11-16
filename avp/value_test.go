// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package avp_test

import (
	"github.com/skill215/diameterkit/avp"
	"testing"
)

func TestTBCDNumbers(t *testing.T) {
	for _, s := range []string{"0", "12", "12025550123", "001010123456789"} {
		b, err := avp.EncodeTBCD(s)
		if err != nil {
			t.Fatal(err)
		}
		got, err := avp.DecodeTBCD(b)
		if err != nil || got != s {
			t.Fatalf("%q: %q %v", s, got, err)
		}
	}
	for _, b := range [][]byte{nil, {0xff}, {0xfa}, {0xf1, 0x23}, {0xa1}, {0x21, 0x43, 0x65, 0x87, 0x09, 0x21, 0x43, 0x65}} {
		if avp.CheckTBCD(b) == nil {
			t.Errorf("accepted %x", b)
		}
	}
	for _, s := range []string{"", "+123", "1234567890123456", "１２"} {
		if _, err := avp.EncodeTBCD(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
}
func TestIdentityAndUTF8(t *testing.T) {
	for _, s := range []string{"hss.example", "HSS-1.example", "xn--fsqu00a.example"} {
		if err := avp.Identity(s); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range []string{"", "a..b", "-a.example", "a_.example", "例子.example", "a.example."} {
		if avp.Identity(s) == nil {
			t.Errorf("accepted %q", s)
		}
	}
	if avp.UTF8("号码") != nil || avp.UTF8(string([]byte{255})) == nil {
		t.Fatal("UTF8")
	}
}
