// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

// App0/271 needs a separate reviewed runtime dispatch/capability binding. Adding
// the offline profile alone must not advertise19302, reserve control resources
// for business ACR, or enable a route that silently treats App0 as a capability.
func TestSCAP1OfflineProfileDoesNotEnableRuntime(t *testing.T) {
	cfg := testConfig(t)
	bp, e := base.Profile()
	if e != nil {
		t.Fatal(e)
	}
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	cp, e := scap1.Profile()
	if e != nil {
		t.Fatal(e)
	}
	cfg.Bundle, e = schema.Compile(bp, sp, cp)
	if e != nil {
		t.Fatal(e)
	}
	n, e := NewNode(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer n.Close()
	cer := n.cer()
	if len(cer.VendorSpecificApplicationID) != 0 || len(cer.AcctApplicationID) != 0 || len(cer.AuthApplicationID) != 1 || cer.AuthApplicationID[0] != appID {
		t.Fatal("offline profile auto advertised", cer)
	}
	key := pdu.CommandKey{ApplicationID: 0, Code: 271, Request: true}
	cfg.Peers[0].Commands[0].Command = key
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			bad.Close()
		}
		t.Fatal("unreviewed App0 outbound binding", e)
	}
	cfg = inboundConfig(t)
	cfg.Bundle = n.cfg.Bundle
	cfg.Peers[0].InboundCommands[0].Command = key
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			bad.Close()
		}
		t.Fatal("unreviewed App0 inbound binding", e)
	}
}
