// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package wire_test

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/wire"
	"math"
	"testing"
)

func TestSigned32ExactWidthAndBudget(t *testing.T) {
	for _, n := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		lim := wire.DefaultLimits()
		b := make([]byte, 12)
		w := wire.NewEncoder(b, lim)
		w.Int32(avp.Key{Code: 429}, 64, n)
		if w.Err() != nil || binary.BigEndian.Uint32(b[8:]) != uint32(n) {
			t.Fatal(n, w.Err())
		}
		got, e := wire.I32(b[8:])
		if e != nil || got != n {
			t.Fatal(got, e)
		}
		for _, dst := range [][]byte{nil, b} {
			lim.MaxFieldBytes = 3
			w = wire.NewEncoder(dst, lim)
			w.Int32(avp.Key{Code: 429}, 64, n)
			if !errors.Is(w.Err(), wire.ErrLimit) {
				t.Fatal(w.Err())
			}
		}
	}
	for _, n := range []int{0, 1, 3, 5, 8} {
		if _, e := wire.I32(make([]byte, n)); !errors.Is(e, avp.ErrValue) {
			t.Fatal(n, e)
		}
	}
}
