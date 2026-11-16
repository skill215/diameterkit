//go:build linux && !386

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/transport"
)

func TestBoundedSCTPPressure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var pairs [4][2]transport.Conn
	var reserved uint64
	for i := range pairs {
		a, b := adapters(t)
		pairs[i] = [2]transport.Conn{a, b}
		reserved += a.Info().BufferBytes + b.Info().BufferBytes
		t.Cleanup(func() { a.Close(); b.Close() })
	}
	if reserved == 0 || reserved > 1<<20 {
		t.Fatal("fixed transport reserve", reserved)
	}
	errs := make(chan error, len(pairs))
	var wg sync.WaitGroup
	for _, p := range pairs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 128; i++ {
				frame := data([]int{20, 256, 4096}[i%3])
				if r := p[0].WriteFrame(ctx, frame); r.Err != nil || r.Bytes != len(frame) {
					errs <- fmt.Errorf("write: %+v", r)
					return
				}
				got, e := read(ctx, p[1])
				if e != nil || !bytes.Equal(got, frame) {
					errs <- fmt.Errorf("read mismatch: %v", e)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	for _, p := range pairs {
		for _, c := range p {
			if e := c.Close(); e != nil {
				t.Error(e)
			}
			if r := c.WriteFrame(ctx, data(20)); r.Err == nil {
				t.Error("closed transport accepted write")
			}
		}
	}
	t.Logf("SCTP_PRESSURE connections=8 frames=512 fixed_buffer_bytes=%d", reserved)
}

// One-way loopback send plus owned frame receive, not a remote network RTT.
func BenchmarkSCTPFrame(b *testing.B) {
	for _, size := range []int{256, 4096} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			a, z := adapters(b)
			defer a.Close()
			defer z.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			frame := data(size)
			b.SetBytes(int64(size))
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if r := a.WriteFrame(ctx, frame); r.Err != nil {
					b.Fatal(r.Err)
				}
				got, e := read(ctx, z)
				if e != nil || !bytes.Equal(frame, got) {
					b.Fatal("frame receive", e)
				}
			}
			b.StopTimer()
		})
	}
}
