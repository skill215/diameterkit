// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"testing"
)

func TestNotificationCombinationsAndExactData(t *testing.T) {
	for mask := 1; mask < 8; mask++ {
		n := NotificationData{}
		var refs []uint32
		if mask&1 != 0 {
			n.IMSUserState = avp.Some(uint8(1))
			refs = append(refs, 11)
		}
		if mask&2 != 0 {
			n.SCSCFName = avp.Some("")
			refs = append(refs, 12)
		}
		if mask&4 != 0 {
			n.Reachability = avp.Some(Reachability{MME: avp.Some(uint8(0))})
			refs = append(refs, 25)
		}
		raw, e := MarshalNotification(n, DefaultLimits())
		if e != nil {
			t.Fatal(e)
		}
		d, e := Parse(raw, DefaultLimits())
		if e != nil {
			t.Fatal(e)
		}
		if r, e := d.CheckNotification(refs); e != nil || !r.Complete {
			t.Fatal(r, e)
		}
		clear(raw)
		if r, e := d.CheckNotification(refs); e != nil || !r.Complete {
			t.Fatal("borrowed", r, e)
		}
		if len(refs) > 1 {
			if r, e := d.CheckNotification(refs[:1]); e != nil || r.Complete {
				t.Fatal("extra data accepted", r, e)
			}
		}
		l := DefaultLimits()
		l.MaxElements = 1
		if _, e = MarshalNotification(n, l); !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, n := range []NotificationData{{}, {IMSUserState: avp.Some(uint8(4))}, {Reachability: avp.Some(Reachability{})}, {SCSCFName: avp.Some("https://invalid")}} {
		if _, e := MarshalNotification(n, DefaultLimits()); e == nil {
			t.Fatal(n)
		}
	}
	d, e := Parse([]byte("<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState></Sh-IMS-Data></Sh-Data>"), DefaultLimits())
	if e != nil {
		t.Fatal(e)
	}
	for _, refs := range [][]uint32{nil, {24}, {11, 11}, {11, 12, 25, 11}} {
		if _, e = d.CheckNotification(refs); !errors.Is(e, ErrArgument) {
			t.Fatal(refs, e)
		}
	}
	if r, e := d.CheckNotification([]uint32{11, 25}); e != nil || r.Complete {
		t.Fatal("missing data", r, e)
	}
}
