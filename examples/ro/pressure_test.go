// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/pdu"
)

type pressureMetrics struct {
	Window, Waves, Requests, Refusals, Canceled int
	ClientLogicalPeak, ServerLogicalPeak        uint64
	ClientOutstandingPeak, ServerInboundPeak    int
	SampledHeapAlloc, ProcessRSSHighWater       uint64
}

func logicalBytes(s diameter.ResourceSnapshot) uint64 {
	return s.FixedBytes + s.OutboundBytes + s.InboundBytes + s.DuplicateBytes + s.RejectBytes
}

func (m *pressureMetrics) sample(t testing.TB, p *pair) {
	t.Helper()
	a, b := p.left.node.Resources(), p.right.node.Resources()
	m.ClientLogicalPeak = max(m.ClientLogicalPeak, logicalBytes(a))
	m.ServerLogicalPeak = max(m.ServerLogicalPeak, logicalBytes(b))
	m.ClientOutstandingPeak = max(m.ClientOutstandingPeak, a.Outstanding)
	m.ServerInboundPeak = max(m.ServerInboundPeak, b.InboundOutstanding)
	if logicalBytes(a) > 1<<30 || logicalBytes(b) > 1<<30 || a.Outstanding > m.Window || b.InboundOutstanding > m.Window {
		t.Fatal("configured logical bounds exceeded", a, b)
	}
	var heap runtime.MemStats
	runtime.ReadMemStats(&heap)
	m.SampledHeapAlloc = max(m.SampledHeapAlloc, heap.HeapAlloc)
	if runtime.GOOS == "linux" {
		status, e := os.ReadFile("/proc/self/status")
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, line := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(line, "VmHWM:") {
				fields := strings.Fields(line)
				v, e := strconv.ParseUint(fields[1], 10, 64)
				if e != nil {
					t.Fatal(e)
				}
				m.ProcessRSSHighWater = max(m.ProcessRSSHighWater, v*1024)
				found = true
			}
		}
		if !found || m.ProcessRSSHighWater > 1<<30 {
			t.Fatal("process RSS guard", m.ProcessRSSHighWater)
		}
	}
}

func pressurePair(t testing.TB, ctx context.Context, window int) *pair {
	t.Helper()
	origins := [2]pdu.Origin{{Host: "ctf.example", Realm: "example"}, {Host: "ocs.example", Realm: "example"}}
	var cfg [2]diameter.NodeConfig
	for i := range cfg {
		c, e := endpointConfig(origins[i], origins[1-i], i == 0)
		if e != nil {
			t.Fatal(e)
		}
		c.Resources.NodeBytes = 1 << 30
		c.Resources.MaxPortOutstanding = window
		c.Resources.Classes[0].Slots, c.Resources.Classes[0].Bytes = window, 256<<20
		c.Peers[0].MaxPending, c.Peers[0].MaxBytes = window, 256<<20
		if i == 1 {
			c.Resources.InboundClasses[0].Slots, c.Resources.InboundClasses[0].Bytes = window, 256<<20
			c.Peers[0].MaxInbound, c.Peers[0].MaxInboundBytes = window, 256<<20
		}
		cfg[i] = c
	}
	p, e := connectPair(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := p.close(); e != nil {
			t.Error(e)
		}
		if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
			t.Error("resources retained after Wait")
		}
	})
	if e = p.left.port.Close(ctx); e != nil {
		t.Fatal(e)
	}
	p.left.port, e = p.left.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: window, MaxOutstandingBytes: 256 << 20})
	if e != nil {
		t.Fatal(e)
	}
	return p
}

func pressureWait(t testing.TB, ctx context.Context, predicate func() bool) {
	t.Helper()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for !predicate() {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-tick.C:
		}
	}
}

func TestBoundedRoPressure(t *testing.T) {
	for _, window := range []int{1, 8, 16} {
		t.Run(fmt.Sprint(window), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			p := pressurePair(t, ctx, window)
			m := pressureMetrics{Window: window, Waves: 16}
			for wave := 0; wave < m.Waves; wave++ {
				requests := make([]*ro.CCR, window)
				ids, errs := make([]diameter.RequestID, window), make([]error, window)
				var wg sync.WaitGroup
				for i := range window {
					var e error
					requests[i], e = newDebit(p.left.node)
					if e != nil {
						t.Fatal(e)
					}
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						ids[i], errs[i] = p.left.peer.Submit(ctx, requests[i], diameter.RequestOptions{}, p.left.port)
					}(i)
				}
				wg.Wait()
				sessions := map[string]diameter.RequestID{}
				for i, e := range errs {
					if e != nil {
						t.Fatal(e)
					}
					sessions[requests[i].SessionID] = ids[i]
				}
				incoming := make([]*diameter.Inbound, window)
				for i := range window {
					var e error
					incoming[i], e = p.right.peer.Receive(ctx)
					if e != nil {
						t.Fatal(e)
					}
				}
				m.sample(t, p)
				if p.left.node.Resources().Outstanding != window || p.right.node.Resources().InboundOutstanding != window {
					t.Fatal("window not saturated")
				}
				refuse := func() {
					q, e := newDebit(p.left.node)
					if e != nil {
						t.Fatal(e)
					}
					if _, e = p.left.peer.Submit(ctx, q, diameter.RequestOptions{}, p.left.port); !errors.Is(e, diameter.ErrBackpressure) {
						t.Fatal("full port admitted", e)
					}
					m.Refusals++
				}
				refuse()
				canceled := map[diameter.RequestID]bool{}
				for i, in := range incoming {
					id, ok := sessions[in.Message().Body.(*ro.CCR).SessionID]
					if !ok {
						t.Fatal("unexpected session")
					}
					if wave%2 == 1 && i%2 == 0 {
						if e := p.left.node.Cancel(id); e != nil {
							t.Fatal(e)
						}
						canceled[id] = true
						m.Canceled++
					}
					a, e := newAnswer("2001")
					if e != nil {
						t.Fatal(e)
					}
					if _, e = in.Respond(ctx, a); e != nil {
						t.Fatal(e)
					}
				}
				pressureWait(t, ctx, func() bool {
					return p.left.node.Resources().Pending == 0 && p.right.node.Resources().InboundOutstanding == 0
				})
				m.sample(t, p)
				refuse() // Completed results still own the bounded port until consumed.
				seen := map[diameter.RequestID]bool{}
				for range window {
					c, e := p.left.port.Receive(ctx)
					if e != nil {
						t.Fatal(e)
					}
					if seen[c.RequestID] {
						t.Fatal("duplicate completion")
					}
					seen[c.RequestID] = true
					if canceled[c.RequestID] {
						if !errors.Is(c.Result.Err, context.Canceled) {
							t.Fatal(c.Result.Err)
						}
					} else if c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
						t.Fatal(c.Result)
					}
				}
				for _, id := range ids {
					if !seen[id] {
						t.Fatal("missing completion")
					}
				}
				pressureWait(t, ctx, func() bool {
					s := p.left.node.Resources()
					r := p.right.node.Resources()
					return s.Outstanding == 0 && s.OutboundBytes == 0 && r.InboundOutstanding == 0 && r.InboundBytes == 0
				})
				m.Requests += window
			}
			if e := p.close(); e != nil {
				t.Fatal(e)
			}
			m.sample(t, p)
			data, e := json.Marshal(m)
			if e != nil {
				t.Fatal(e)
			}
			t.Logf("PRESSURE %s", data)
		})
	}
}

func BenchmarkRoRoundTrip(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	p := pressurePair(b, ctx, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if e := exchange(ctx, &p.left, &p.right, "2001"); e != nil {
			b.Fatal(e)
		}
	}
	b.StopTimer()
}
