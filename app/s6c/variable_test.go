// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Selected TBCD omission versus present empty field.
// SPDX-AI-Date: 2026-11-17

package s6c

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestVariableMSISDNMinimum(t *testing.T) {
	l := wire.DefaultLimits()
	_, err := decodeUserIdentifier(wire.NewDecoder(nil, l))
	q, ok := missingavp.Quote(err, l, 16)
	if !ok || !errors.Is(err, validation.ErrMissing) || len(q) != 16 {
		t.Fatal(err, q)
	}
	d := wire.NewDecoder(q, l)
	v, found, err := d.Next()
	if err != nil || !found || v.Key != (avp.Key{VendorID: 10415, Code: 701}) || v.Flags != 192 || len(v.Data) != 1 || v.Data[0] != 0 || wire.Uint24(q[5:8]) != 13 {
		t.Fatal(v, err)
	}
	if _, ok := missingavp.Quote(validation.At(v.Key, validation.ErrMissing), l, 16); ok {
		t.Fatal("bare proof")
	}
	empty := make([]byte, 12)
	w := wire.NewEncoder(empty, l)
	w.Bytes(v.Key, 192, nil)
	_, err = decodeUserIdentifier(wire.NewDecoder(empty, l))
	if err == nil || errors.Is(err, validation.ErrMissing) {
		t.Fatal("present empty", err)
	}
}
