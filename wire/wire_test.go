// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package wire_test

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestVendorGroupedMeasureMatchesEncodeAtLimit(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxFieldBytes = 12
	encode := func(dst []byte) (int, error) {
		w := wire.NewEncoder(dst, l)
		m := w.BeginGroup(avp.Key{VendorID: 10415, Code: 2401}, 0xc0)
		w.Uint32(avp.Key{Code: 266}, 0x40, 10415)
		w.EndGroup(m)
		return w.Len(), w.Err()
	}
	n, err := encode(nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 24 {
		t.Fatal(n)
	}
	if n2, err := encode(make([]byte, n)); err != nil || n2 != n {
		t.Fatalf("measure/write disagree: %d %v", n2, err)
	}
	l.MaxFieldBytes--
	if _, err = encode(nil); !errors.Is(err, wire.ErrLimit) {
		t.Fatal(err)
	}
}
func TestSharedDepthAndCount(t *testing.T) {
	l := wire.DefaultLimits()
	l.MaxDepth = 1
	l.MaxAVPs = 2
	l.MaxRepeated = 2
	b := []byte{0, 0, 1, 41, 64, 0, 0, 20, 0, 0, 1, 10, 64, 0, 0, 12, 0, 0, 40, 175}
	d := wire.NewDecoder(b, l)
	v, _, err := d.Next()
	if err != nil {
		t.Fatal(err)
	}
	child, err := d.Child(v.Data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = child.Child(nil); !errors.Is(err, wire.ErrLimit) {
		t.Fatal("depth", err)
	}
	if _, ok, err := child.Next(); !ok || err != nil {
		t.Fatal(err)
	}
	another, err := d.Child(v.Data)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = another.Next(); !errors.Is(err, wire.ErrLimit) {
		t.Fatal("shared count", err)
	}
}
func TestDecodeBoundAndConfig(t *testing.T) {
	l := wire.DefaultLimits()
	a, err := wire.DecodeBound(20, l, 1024)
	if err != nil {
		t.Fatal(err)
	}
	b, err := wire.DecodeBound(200, l, 1024)
	if err != nil || b <= a {
		t.Fatal("not monotonic")
	}
	for _, n := range []int{-1, 19, l.MaxFrameBytes + 1} {
		if _, err := wire.DecodeBound(n, l, 1024); err == nil {
			t.Fatal("invalid bound")
		}
	}
	l.MaxDepth = 65
	if l.Validate() == nil {
		t.Fatal("unbounded depth")
	}
}
