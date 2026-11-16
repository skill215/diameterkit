// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codecengine_test

import (
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestOwnedFrameAdoptionAndClone(t *testing.T) {
	p, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	b, err := schema.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	e, err := codecengine.New(b, wire.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	body := &s6c.SRR{SessionID: "a;1", AuthSessionState: 1, OriginHost: "a.example", OriginRealm: "example", DestinationRealm: "example", MSISDN: avp.Some([]byte{0x21}), SCAddress: []byte{0x43}, SMRPMTI: avp.Some(uint32(0))}
	frame, err := e.Append(nil, pdu.Message{Header: pdu.Header{Flags: 0xc0, CommandCode: 8388647, ApplicationID: 16777312}, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.DecodeOwned(frame)
	if err != nil {
		t.Fatal(err)
	}
	decoded := m.Body.(*s6c.SRR)
	// Compare addresses without mutating an adopted frame or importing unsafe.
	found := false
	for i := range frame {
		if &frame[i] == &decoded.MSISDN.Value[0] {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("owned entry copied payload")
	}
	op, _ := b.Lookup(body.CommandKey())
	cloned, err := op.Clone(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if &cloned.(*s6c.SRR).MSISDN.Value[0] == &decoded.MSISDN.Value[0] {
		t.Fatal("clone aliases owned frame")
	}
}
