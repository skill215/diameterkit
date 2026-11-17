// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit UTF8 minimum without inferring subscription domain length.
// SPDX-AI-Date: 2026-11-17

package dcca

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestVariableSubscriptionMinimum(t *testing.T) {
	l := wire.DefaultLimits()
	raw := make([]byte, 20)
	w := wire.NewEncoder(raw, l)
	w.Uint32(avp.Key{Code: 450}, 64, 1)
	_, err := decodeSubscriptionID(wire.NewDecoder(raw[:12], l))
	q, ok := missingavp.Quote(err, l, 12)
	if !ok || !errors.Is(err, validation.ErrMissing) || len(q) != 12 {
		t.Fatal(err, q)
	}
	d := wire.NewDecoder(q, l)
	v, found, e := d.Next()
	if e != nil || !found || v.Key != (avp.Key{Code: 444}) || len(v.Data) != 1 || v.Data[0] != 0 || wire.Uint24(q[5:8]) != 9 {
		t.Fatal(v, e)
	}
	w.Bytes(avp.Key{Code: 444}, 64, nil)
	_, err = decodeSubscriptionID(wire.NewDecoder(raw, l))
	if err == nil || errors.Is(err, validation.ErrMissing) {
		t.Fatal("present empty", err)
	}
	// One diagnostic byte does not assert that one character is a valid IMSI.
	if err := validateSubscriptionID(&SubscriptionID{SubscriptionIDType: 1, SubscriptionIDData: "1"}, validation.Receive); err == nil {
		t.Fatal("domain minimum widened")
	}
}
