package codec_test

import (
	"bytes"
	"github.com/skill215/diameterkit/codec"
	"testing"
)

func FuzzCanonical(f *testing.F) {
	c := newCodec(f, codec.Config{})
	for _, name := range []string{"v-srr-01", "b2-srr-proxy"} {
		f.Add(fixture(f, name))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 4096 {
			return
		}
		out, err := c.Canonical(b, 8192)
		if err != nil {
			return
		}
		if len(out) > 8192 || len(out) < 12 {
			t.Fatal("unbounded canonical output")
		}
		again, err := c.Canonical(b, 8192)
		if err != nil || !bytes.Equal(out, again) {
			t.Fatal("nondeterministic canonicalization", err)
		}
		changed := append([]byte(nil), b...)
		changed[4] ^= 0x10
		changed[12] ^= 0xff
		same, err := c.Canonical(changed, 8192)
		if err != nil || !bytes.Equal(out, same) {
			t.Fatal("HBH/T changed comparison", err)
		}
	})
}
