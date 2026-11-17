// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Common redirect operation contract fingerprint migration.
// SPDX-AI-Date: 2026-11-17

package schema_test

import (
	"crypto/sha256"
	"fmt"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/schema"
	"testing"
)

func TestRedirectOperationFingerprint(t *testing.T) {
	p, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := b.Lookup((&s6c.SRA{}).CommandKey())
	p, e = schema.NewProfile("redirect-contract", "1", p.Digest(), c)
	if e != nil {
		t.Fatal(e)
	}
	b, e = schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	identity := fmt.Sprintf("/%q/%q/%s/%d/%d/%t/%t/E=%t/C=%t/R=%t/N=%t/A=%t/X=%q", p.Name(), p.Version(), p.Digest(), c.Key.ApplicationID, c.Key.Code, c.Key.Request, c.Proxiable, c.Error != nil, c.Canonical != nil, c.TransportReplay, c.NoContext, c.CheckAnswer != nil, c.ContextName)
	old := fmt.Sprintf("%x", sha256.Sum256([]byte("ops8"+identity)))
	current := fmt.Sprintf("%x", sha256.Sum256([]byte("ops9"+identity)))
	if b.Fingerprint() == old || b.Fingerprint() != current {
		t.Fatal("shared E semantics omitted from fingerprint")
	}
}
