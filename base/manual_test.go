// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Manual decoder limits, domain exclusions and bounded evidence fuzzing.
// SPDX-AI-Date: 2026-11-17

package base_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/internal/excessavp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func manualPart(k uint32, data []byte) []byte {
	b := make([]byte, (8+len(data)+3)&^3)
	w := wire.NewEncoder(b, wire.DefaultLimits())
	w.Bytes(avp.Key{Code: k}, 64, data)
	return b
}
func TestManualLeafLimits(t *testing.T) {
	for _, key := range []uint32{266, 258, 259, 280, 33} {
		t.Run(string(rune(key)), func(t *testing.T) {
			data := []byte{0, 0, 0, 1}
			if key == 280 {
				data = []byte("proxy.example")
			}
			raw := append(manualPart(key, data), manualPart(key, data)...)
			l := wire.DefaultLimits()
			l.MaxRepeated = 1
			var e error
			if key == 280 || key == 33 {
				_, e = base.DecodeProxy(wire.NewDecoder(raw, l))
			} else {
				_, e = base.DecodeVSAI(wire.NewDecoder(raw, l))
			}
			if !errors.Is(e, wire.ErrLimit) {
				t.Fatal("repetition priority", e)
			}
			if q, ok := excessavp.Quote(raw, e, l, 1024); ok || q != nil {
				t.Fatal("resource guessed as5009")
			}
		})
	}
	// Conditional Auth/Acct pairing is not a single-leaf value or omission.
	for _, apps := range [][]byte{nil, append(manualPart(258, make([]byte, 4)), manualPart(259, make([]byte, 4))...)} {
		raw := append(manualPart(266, make([]byte, 4)), apps...)
		_, e := base.DecodeVSAI(wire.NewDecoder(raw, wire.DefaultLimits()))
		if !errors.Is(e, validation.ErrField) {
			t.Fatal(e)
		}
		if _, _, ok := excessavp.QuoteInvalid(raw, e, wire.DefaultLimits(), 1024); ok {
			t.Fatal("invented conditional value")
		}
		if _, ok := missingavp.Quote(e, wire.DefaultLimits(), 1024); ok {
			t.Fatal("invented conditional omission")
		}
	}
	// Binary Proxy-State may be empty, and optional opaque extensions survive.
	raw := append(manualPart(280, []byte("proxy.example")), manualPart(33, nil)...)
	extra := manualPart(999999, []byte{0, 255})
	extra[4] = 0
	raw = append(raw, extra...)
	p, e := base.DecodeProxy(wire.NewDecoder(raw, wire.DefaultLimits()))
	if e != nil || len(p.ProxyState) != 0 || len(p.Extra) != 1 {
		t.Fatal(p, e)
	}
}

func FuzzManualDiagnostics(f *testing.F) {
	for _, raw := range [][]byte{manualPart(266, []byte{1, 2, 3}), manualPart(280, []byte("bad_name")), append(manualPart(33, nil), manualPart(33, []byte{1})...), manualPart(999999, []byte{255})} {
		f.Add(raw, false, uint16(1024))
		f.Add(raw, true, uint16(1024))
	}
	f.Fuzz(func(t *testing.T, raw []byte, proxy bool, budget uint16) {
		l := wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 2048, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 32}
		if len(raw) > l.MaxFrameBytes-20 {
			return
		}
		var e error
		if proxy {
			_, e = base.DecodeProxy(wire.NewDecoder(raw, l))
		} else {
			_, e = base.DecodeVSAI(wire.NewDecoder(raw, l))
		}
		if e == nil {
			return
		}
		q, code, ok := excessavp.QuoteInvalid(raw, e, l, int(budget))
		if !ok {
			q, ok = excessavp.Quote(raw, e, l, int(budget))
			code = 5009
		}
		if !ok {
			return
		}
		if len(q) > int(budget) || len(q) > l.MaxFieldBytes {
			t.Fatal("unbounded diagnostic")
		}
		d := wire.NewDecoder(raw, l)
		found := false
		for {
			v, yes, err := d.Next()
			if err != nil || !yes {
				break
			}
			h := 8
			if v.Key.VendorID != 0 {
				h = 12
			}
			size := (h + len(v.Data) + 3) &^ 3
			if bytes.Equal(q, raw[v.Offset:v.Offset+size]) {
				found = true
				if code == 5014 && len(v.Data) == 4 {
					t.Fatal("valid width")
				}
				break
			}
		}
		if !found {
			t.Fatal("invented instance")
		}
		saved := append([]byte(nil), q...)
		clear(raw)
		if !bytes.Equal(q, saved) {
			t.Fatal("borrowed diagnostic")
		}
	})
}
