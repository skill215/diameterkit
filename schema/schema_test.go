// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package schema_test

import (
	"errors"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestExplicitFrozenBundle(t *testing.T) {
	bp, err := base.Profile()
	if err != nil {
		t.Fatal(err)
	}
	sp, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	a, err := schema.Compile(bp, sp)
	if err != nil {
		t.Fatal(err)
	}
	b, err := schema.Compile(sp, bp)
	if err != nil {
		t.Fatal(err)
	}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("profile order affects fingerprint")
	}
	if _, err = schema.Compile(bp, bp); !errors.Is(err, schema.ErrDefinition) {
		t.Fatal("ambiguous definition accepted")
	}
	if _, err = schema.Compile(schema.Profile{}); !errors.Is(err, schema.ErrDefinition) {
		t.Fatal("empty profile accepted")
	}
	op, ok := a.Lookup((&s6c.SRR{}).CommandKey())
	if !ok {
		t.Fatal("missing operation")
	}
	op.Decode = nil
	original, _ := a.Lookup(op.Key)
	if original.Decode == nil {
		t.Fatal("mutable table escaped")
	}
	if _, err = schema.NewProfile("broken", "1", bp.Digest(), op); !errors.Is(err, schema.ErrDefinition) {
		t.Fatal("missing operation accepted")
	}
	answer, _ := a.Lookup((&s6c.SRA{}).CommandKey())
	profile, err := schema.NewProfile("owned-error-table", "2", sp.Digest(), answer)
	if err != nil {
		t.Fatal(err)
	}
	answer.Error.Decode = nil
	bundle, err := schema.Compile(profile)
	if err != nil {
		t.Fatal(err)
	}
	frozen, _ := bundle.Lookup(answer.Key)
	if frozen.Error.Decode == nil {
		t.Fatal("error operations retained caller pointer")
	}
	frozen.Error.Decode = nil
	fresh, _ := bundle.Lookup(answer.Key)
	if fresh.Error.Decode == nil {
		t.Fatal("error operations escaped through lookup")
	}
	fresh.Envelope = nil
	if _, err = schema.NewProfile("missing-envelope", "1", sp.Digest(), fresh); !errors.Is(err, schema.ErrDefinition) {
		t.Fatal(err)
	}
}

// Recovery permission is part of the frozen operation identity. A route cannot
// silently gain replay authority when its wire fields happen to stay the same.
func TestTransportReplayPermission(t *testing.T) {
	sp, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	b, err := schema.Compile(sp)
	if err != nil {
		t.Fatal(err)
	}
	op, _ := b.Lookup((&s6c.SRR{}).CommandKey())
	if !op.TransportReplay {
		t.Fatal("SRR lacks audited replay permission")
	}
	makeBundle := func() *schema.Bundle {
		p, e := schema.NewProfile("replay-permission", "1", sp.Digest(), op)
		if e != nil {
			t.Fatal(e)
		}
		b, e := schema.Compile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	allowed := makeBundle()
	op.TransportReplay = false
	denied := makeBundle()
	if allowed.Fingerprint() == denied.Fingerprint() {
		t.Fatal("permission missing from fingerprint")
	}
	frozen, _ := allowed.Lookup(op.Key)
	if !frozen.TransportReplay {
		t.Fatal("caller mutation escaped freeze")
	}
	op.TransportReplay = true
	op.Key.Request = false
	if _, e := schema.NewProfile("bad-answer", "1", sp.Digest(), op); !errors.Is(e, schema.ErrDefinition) {
		t.Fatal(e)
	}
	op.Key.Request = true
	op.Key.ApplicationID = 0
	if _, e := schema.NewProfile("bad-control", "1", sp.Digest(), op); !errors.Is(e, schema.ErrDefinition) {
		t.Fatal(e)
	}
}

func TestAnswerRelationFingerprintAndFreeze(t *testing.T) {
	sp, e := s6c.Profile()
	if e != nil {
		t.Fatal(e)
	}
	b, e := schema.Compile(sp)
	if e != nil {
		t.Fatal(e)
	}
	op, _ := b.Lookup((&s6c.SRA{}).CommandKey())
	makeBundle := func() *schema.Bundle {
		p, e := schema.NewProfile("relation", "1", sp.Digest(), op)
		if e != nil {
			t.Fatal(e)
		}
		b, e := schema.Compile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	op.CheckAnswer = nil
	without := makeBundle()
	op.CheckAnswer = func(pdu.Body, pdu.Body) error { return validationSentinel }
	with := makeBundle()
	if without.Fingerprint() == with.Fingerprint() {
		t.Fatal("relation omitted from fingerprint")
	}
	op.CheckAnswer = nil
	frozen, _ := with.Lookup(op.Key)
	if frozen.CheckAnswer == nil || frozen.CheckAnswer(nil, nil) != validationSentinel {
		t.Fatal("mutable operation escaped")
	}
	op, _ = b.Lookup((&s6c.SRR{}).CommandKey())
	op.CheckAnswer = frozen.CheckAnswer
	if _, e = schema.NewProfile("bad-relation", "1", sp.Digest(), op); !errors.Is(e, schema.ErrDefinition) {
		t.Fatal(e)
	}
}

var validationSentinel = errors.New("test relation")

func TestContextOperationContractFingerprint(t *testing.T) {
	p, _ := base.Profile()
	b, _ := schema.Compile(p)
	op, _ := b.Lookup((&base.DWR{}).CommandKey())
	op.CheckContext = func(pdu.Body, pdu.Body, validation.Context, validation.Mode) error { return nil }
	op.ContextName = "ShPeerFeatures"
	if _, e := schema.NewProfile(p.Name(), p.Version(), p.Digest(), op); !errors.Is(e, schema.ErrDefinition) {
		t.Fatal("NoContext conflict", e)
	}
	op.NoContext = false
	valid, e := schema.NewProfile(p.Name(), p.Version(), p.Digest(), op)
	if e != nil {
		t.Fatal(e)
	}
	x, _ := schema.Compile(valid)
	op.ContextName = "DifferentNamedScope"
	other, e := schema.NewProfile(p.Name(), p.Version(), p.Digest(), op)
	if e != nil {
		t.Fatal(e)
	}
	y, _ := schema.Compile(other)
	if x.Fingerprint() == y.Fingerprint() {
		t.Fatal("context scope absent")
	}
	op.ContextName = ""
	if _, e = schema.NewProfile(p.Name(), p.Version(), p.Digest(), op); !errors.Is(e, schema.ErrDefinition) {
		t.Fatal(e)
	}
}
