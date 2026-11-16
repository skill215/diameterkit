package wire_test

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/wire"
	"math"
	"testing"
)

func TestSigned64ExactWidthAndBudget(t *testing.T) {
	for _, n := range []int64{math.MinInt64, -(1<<53 + 1), -1, 0, 1, 1<<53 + 1, math.MaxInt64} {
		lim := wire.DefaultLimits()
		b := make([]byte, 16)
		w := wire.NewEncoder(b, lim)
		w.Int64(avp.Key{Code: 447}, 64, n)
		if w.Err() != nil || binary.BigEndian.Uint64(b[8:]) != uint64(n) {
			t.Fatal(n, w.Err())
		}
		got, e := wire.I64(b[8:])
		if e != nil || got != n {
			t.Fatal(got, e)
		}
		for _, dst := range [][]byte{nil, b} {
			lim.MaxFieldBytes = 7
			w = wire.NewEncoder(dst, lim)
			w.Int64(avp.Key{Code: 447}, 64, n)
			if !errors.Is(w.Err(), wire.ErrLimit) {
				t.Fatal(w.Err())
			}
		}
	}
	for _, n := range []int{0, 1, 4, 7, 9, 16} {
		if _, e := wire.I64(make([]byte, n)); !errors.Is(e, avp.ErrValue) {
			t.Fatal(n, e)
		}
	}
}
