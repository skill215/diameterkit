// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package schema combines immutable, explicitly supplied protocol implementations.
// It neither imports applications nor discovers them through global registration.
package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"sort"
	"strings"
)

var ErrDefinition = errors.New("invalid or conflicting protocol definition")

// Command is the bounded wire-codec operation contract. Callback implementations
// are trusted, immutable and concurrency-safe. Envelope views borrow validated
// bodies; Clone/completion produce owned values. Canonical consumes validated raw
// requests and returns owned comparison bytes; nil means caching is unsupported.
type Command struct {
	// ContextName bounds the report scope (at most 64 bytes). CheckContext is a
	// compiled pure rule: request is the original request; answer is nil during
	// request checks. For an answer, mode describes the answer's send/receive side.
	// It must check all application-context relations against the supplied original.
	ContextName  string
	CheckContext func(request, answer pdu.Body, ctx validation.Context, mode validation.Mode) error
	// NoContext certifies this slice needs no additional application evidence.
	// It is mutually exclusive with CheckContext; neither present means wire-only.
	NoContext       bool
	TransportReplay bool // Audited fixed-terminal transport recovery, not business retry.
	Canonical       func([]byte, wire.Limits, int) ([]byte, error)
	Key             pdu.CommandKey
	Proxiable       bool
	Decode          func(*wire.Decoder) (pdu.Body, error)
	Encode          func(*wire.Encoder, pdu.Body) error
	Validate        func(pdu.Body, validation.Mode) error
	Clone           func(pdu.Body) (pdu.Body, error)
	CompleteRequest func(pdu.Body, pdu.RequestEnvelope) (pdu.Body, error)
	DecodeBound     func(int, wire.Limits) (uint64, error)
	Envelope        func(pdu.Body) (pdu.Envelope, error)
	CompleteAnswer  func(pdu.Body, pdu.AnswerEnvelope) (pdu.Body, error)
	// CheckAnswer checks application-specific original-request relations on normal answers.
	// Both bodies are validated, bounded and borrowed; no mutation or I/O is permitted.
	CheckAnswer func(request, answer pdu.Body) error
	Error       *ErrorOperations
}

// ErrorOperations is the separately compiled E grammar for an answer command.
// Its callbacks obey the same ownership/concurrency rules as Command. Application
// restrictions belong in these fixed callbacks, not in a dynamic result plugin.
type ErrorOperations struct {
	Decode         func(*wire.Decoder) (pdu.Body, error)
	Encode         func(*wire.Encoder, pdu.Body) error
	Validate       func(pdu.Body, validation.Mode) error
	Clone          func(pdu.Body) (pdu.Body, error)
	DecodeBound    func(int, wire.Limits) (uint64, error)
	Envelope       func(pdu.Body) (pdu.Envelope, error)
	CompleteAnswer func(pdu.Body, pdu.AnswerEnvelope) (pdu.Body, error)
}
type Profile struct {
	name, version, digest string
	commands              []Command
}

func NewProfile(name, version, digest string, commands ...Command) (Profile, error) {
	if name == "" || version == "" || len(commands) == 0 {
		return Profile{}, ErrDefinition
	}
	b, err := hex.DecodeString(digest)
	if err != nil || len(b) != sha256.Size {
		return Profile{}, ErrDefinition
	}
	seen := map[pdu.CommandKey]bool{}
	frozen := append([]Command(nil), commands...)
	for i, c := range frozen {
		if c.Key.Code == 0 || c.Key.Code > wire.MaxWireSize || seen[c.Key] || c.Decode == nil || c.Encode == nil || c.Validate == nil || c.Clone == nil || c.DecodeBound == nil || c.Key.Request && c.CompleteRequest == nil {
			return Profile{}, ErrDefinition
		}
		if len(c.ContextName) > 64 {
			return Profile{}, ErrDefinition
		}
		if (c.ContextName == "") != (c.CheckContext == nil) || c.CheckContext != nil && c.NoContext {
			return Profile{}, ErrDefinition
		}
		if c.CheckAnswer != nil && c.Key.Request {
			return Profile{}, ErrDefinition
		}
		if c.TransportReplay && (!c.Key.Request || c.Key.ApplicationID == 0) {
			return Profile{}, ErrDefinition
		}
		seen[c.Key] = true
		if c.Envelope == nil || !c.Key.Request && c.CompleteAnswer == nil {
			return Profile{}, ErrDefinition
		}
		if c.Error != nil {
			e := *c.Error
			if c.Key.Request || e.Decode == nil || e.Encode == nil || e.Validate == nil || e.Clone == nil || e.DecodeBound == nil || e.Envelope == nil || e.CompleteAnswer == nil {
				return Profile{}, ErrDefinition
			}
			frozen[i].Error = &e
		}
	}
	return Profile{name, version, strings.ToLower(digest), frozen}, nil
}
func (p Profile) Name() string    { return p.name }
func (p Profile) Version() string { return p.version }
func (p Profile) Digest() string  { return p.digest }

type Bundle struct {
	commands    map[pdu.CommandKey]Command
	fingerprint string
}

func Compile(profiles ...Profile) (*Bundle, error) {
	if len(profiles) == 0 {
		return nil, ErrDefinition
	}
	b := &Bundle{commands: make(map[pdu.CommandKey]Command)}
	var identities []string
	names := make(map[string]bool)
	for _, p := range profiles {
		if len(p.commands) == 0 || names[p.name] {
			return nil, ErrDefinition
		}
		names[p.name] = true
		for _, c := range p.commands {
			identities = append(identities, fmt.Sprintf("ops8/%q/%q/%s/%d/%d/%t/%t/E=%t/C=%t/R=%t/N=%t/A=%t/X=%q", p.name, p.version, p.digest, c.Key.ApplicationID, c.Key.Code, c.Key.Request, c.Proxiable, c.Error != nil, c.Canonical != nil, c.TransportReplay, c.NoContext, c.CheckAnswer != nil, c.ContextName))
			if _, ok := b.commands[c.Key]; ok {
				return nil, fmt.Errorf("%w: command %v", ErrDefinition, c.Key)
			}
			b.commands[c.Key] = c
		}
	}
	sort.Strings(identities)
	sum := sha256.Sum256([]byte(strings.Join(identities, "\n")))
	b.fingerprint = hex.EncodeToString(sum[:])
	return b, nil
}
func (b *Bundle) Lookup(k pdu.CommandKey) (Command, bool) {
	if b == nil {
		return Command{}, false
	}
	c, ok := b.commands[k]
	if c.Error != nil {
		copy := *c.Error
		c.Error = &copy
	}
	return c, ok
}
func (b *Bundle) Fingerprint() string {
	if b == nil {
		return ""
	}
	return b.fingerprint
}
