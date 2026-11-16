package wire_test

import (
	"encoding/binary"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/wire"
	"math"
	"testing"
)

func TestUnsigned64ExactWidthAndBudgets(t *testing.T) {
	for _, n := range []uint64{0, 1, math.MaxUint32, 1<<53 + 1, 1 << 63, math.MaxUint64} {
		limits := wire.DefaultLimits()
		measure := wire.NewEncoder(nil, limits)
		measure.Uint64(avp.Key{Code: 417}, 64, n)
		if measure.Err() != nil || measure.Len() != 16 {
			t.Fatal(measure.Len(), measure.Err())
		}
		b := make([]byte, 16)
		e := wire.NewEncoder(b, limits)
		e.Uint64(avp.Key{Code: 417}, 64, n)
		if e.Err() != nil || binary.BigEndian.Uint64(b[8:]) != n {
			t.Fatal(e.Err(), b)
		}
		d := wire.NewDecoder(b, limits)
		v, ok, err := d.Next()
		if err != nil || !ok {
			t.Fatal(err)
		}
		got, err := wire.U64(v.Data)
		if err != nil || got != n {
			t.Fatal(got, err)
		}
		for _, dst := range [][]byte{nil, make([]byte, 16)} {
			limits.MaxFieldBytes = 7
			e := wire.NewEncoder(dst, limits)
			e.Uint64(avp.Key{Code: 417}, 64, n)
			if !errors.Is(e.Err(), wire.ErrLimit) {
				t.Fatal(e.Err())
			}
		}
	}
	for _, size := range []int{0, 1, 4, 7, 9, 16} {
		if _, err := wire.U64(make([]byte, size)); !errors.Is(err, avp.ErrValue) {
			t.Fatal(size, err)
		}
	}
}
