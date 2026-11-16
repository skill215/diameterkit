// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Missing diagnostic limits, provenance and fuzz acceptance.
// SPDX-AI-Date: 2026-11-16

package missingavp

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// Synthetic metadata producers exercise the consumer independently of codegen.
type omission struct {
	key   avp.Key
	flags byte
	width int
	child error
}

func (*omission) Error() string { return validation.ErrMissing.Error() }
func (*omission) Unwrap() error { return validation.ErrMissing }
func (e *omission) DiameterMissingAVPV1() (avp.Key, byte, int, error) {
	return e.key, e.flags, e.width, e.child
}
func Required(key avp.Key, flags byte, width int) error {
	return validation.At(key, &omission{key: key, flags: flags, width: width})
}
func Within(v wire.View, err error) error {
	var proof evidence
	if !errors.As(err, &proof) {
		return validation.At(v.Key, err)
	}
	f := validation.At(v.Key, err).(*validation.FieldError)
	f.Cause = &omission{key: v.Key, flags: v.Flags, child: proof.(error)}
	return f
}

func TestMissingQuoteBoundsAndProvenance(t *testing.T) {
	leaf := avp.Key{VendorID: 10415, Code: 630}
	parent := wire.View{Key: avp.Key{VendorID: 10415, Code: 628}, Flags: 0xc0, Data: []byte("must not retain")}
	newError := func() error { return Within(parent, Required(leaf, 0x80, 4)) }
	for _, tc := range []struct {
		name   string
		change func(*wire.Limits, *int, *error)
		ok     bool
	}{
		{"exact", func(l *wire.Limits, n *int, e *error) { *n = 28; l.MaxDepth = 2 }, true},
		{"bytes", func(l *wire.Limits, n *int, e *error) { *n = 27 }, false},
		{"field", func(l *wire.Limits, n *int, e *error) { l.MaxFieldBytes = 27 }, false},
		{"depth", func(l *wire.Limits, n *int, e *error) { l.MaxDepth = 1 }, false},
		{"count", func(l *wire.Limits, n *int, e *error) { l.MaxAVPs = 2; l.MaxRepeated = 2 }, false},
		{"invalid-limits", func(l *wire.Limits, n *int, e *error) { l.MaxDepth = 65 }, false},
		{"unproven", func(l *wire.Limits, n *int, e *error) {
			*e = validation.At(parent.Key, validation.At(leaf, validation.ErrMissing))
		}, false},
		{"changed-path", func(l *wire.Limits, n *int, e *error) { (*e).(*validation.FieldError).Path[0].Code++ }, false},
		{"extra-path", func(l *wire.Limits, n *int, e *error) { *e = validation.At(parent.Key, *e) }, false},
		{"wrapped", func(l *wire.Limits, n *int, e *error) { *e = fmt.Errorf("decode: %w", *e) }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := wire.DefaultLimits()
			n := 1024
			err := newError()
			tc.change(&l, &n, &err)
			b, ok := Quote(err, l, n)
			if ok != tc.ok {
				t.Fatalf("quote=%x ok=%v", b, ok)
			}
			if !ok {
				return
			}
			if !errors.Is(err, validation.ErrMissing) || bytes.Contains(b, parent.Data) {
				t.Fatal("lost classification or leaked data")
			}
			d := wire.NewDecoder(b, l)
			v, found, e := d.Next()
			if e != nil || !found || v.Key != parent.Key || v.Flags != parent.Flags || d.Remaining() != 0 {
				t.Fatal(v, e)
			}
			child, e := d.Child(v.Data)
			if e != nil {
				t.Fatal(e)
			}
			v, found, e = child.Next()
			if e != nil || !found || v.Key != leaf || v.Flags != 0x80 || !bytes.Equal(v.Data, make([]byte, 4)) || child.Remaining() != 0 {
				t.Fatal(v, e)
			}
		})
	}
	for _, width := range []int{-1, 0, 1, 3, 9} {
		if _, ok := Quote(Required(leaf, 0x80, width), wire.DefaultLimits(), 1024); ok {
			t.Fatal("guessed width", width)
		}
	}
}

func FuzzMissingQuote(f *testing.F) {
	f.Add([]byte{0, 0, 4, 2, 64, 64})
	f.Add([]byte{3, 0x80, 8, 5, 64, 64})
	f.Add([]byte{63, 0xc0, 8, 64, 64, 64})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) < 6 {
			return
		}
		depth := int(b[0] % 65)
		flags := b[1]
		width := int(b[2] % 10)
		key := avp.Key{Code: 447}
		if flags&0x80 != 0 {
			key.VendorID = 10415
		}
		err := Required(key, flags, width)
		for i := 0; i < depth; i++ {
			err = Within(wire.View{Key: key, Flags: flags}, err)
		}
		l := wire.DefaultLimits()
		l.MaxDepth = int(b[3] % 66)
		l.MaxAVPs = int(b[4])
		l.MaxRepeated = 1
		maxBytes := int(b[5]) * 16
		quoted, ok := Quote(err, l, maxBytes)
		if !ok {
			return
		}
		if len(quoted) > maxBytes || depth+1 > l.MaxDepth || depth+2 > l.MaxAVPs {
			t.Fatal("unbounded quote")
		}
		d := wire.NewDecoder(quoted, l)
		for i := 0; i <= depth; i++ {
			v, found, e := d.Next()
			if e != nil || !found || v.Key != key || v.Flags != flags || d.Remaining() != 0 {
				t.Fatal("invalid chain", e)
			}
			if i == depth {
				if len(v.Data) != width || !bytes.Equal(v.Data, make([]byte, width)) {
					t.Fatal("invalid placeholder")
				}
				break
			}
			d, e = d.Child(v.Data)
			if e != nil {
				t.Fatal(e)
			}
		}
	})
}
