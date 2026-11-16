package codec_test

import (
	"github.com/skill215/diameterkit/codec"
	"testing"
)

// Setup/fixture I/O is excluded. Append reuses a caller-owned destination;
// Decode and Canonical include their owned output allocations.
func BenchmarkCodec(b *testing.B) {
	for _, name := range []string{"base-dwr", "v-srr-01"} {
		b.Run(name, func(b *testing.B) {
			c := newCodec(b, codec.Config{})
			frame := fixture(b, name)
			msg, e := c.Decode(frame)
			if e != nil {
				b.Fatal(e)
			}
			b.Run("Decode", func(b *testing.B) {
				b.SetBytes(int64(len(frame)))
				b.ReportAllocs()
				for b.Loop() {
					if _, e := c.Decode(frame); e != nil {
						b.Fatal(e)
					}
				}
			})
			b.Run("AppendReuse", func(b *testing.B) {
				dst := make([]byte, 0, len(frame))
				b.SetBytes(int64(len(frame)))
				b.ReportAllocs()
				for b.Loop() {
					if _, e := c.Append(dst[:0], msg); e != nil {
						b.Fatal(e)
					}
				}
			})
			if name == "v-srr-01" {
				b.Run("Canonical", func(b *testing.B) {
					b.SetBytes(int64(len(frame)))
					b.ReportAllocs()
					for b.Loop() {
						if _, e := c.Canonical(frame, 8192); e != nil {
							b.Fatal(e)
						}
					}
				})
			}
		})
	}
}
