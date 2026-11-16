package codecengine_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestFrozenRequestAndHeaderBeforeBody(t *testing.T) {
	p, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := schema.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	reqOp, _ := bundle.Lookup((&s6c.SRR{}).CommandKey())
	answerOp, _ := bundle.Lookup((&s6c.SRA{}).CommandKey())
	original := answerOp.Decode
	var calls atomic.Int32
	answerOp.Decode = func(d *wire.Decoder) (pdu.Body, error) { calls.Add(1); return original(d) }
	custom, err := schema.NewProfile("snapshot-test", "2", p.Digest(), reqOp, answerOp)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err = schema.Compile(custom)
	if err != nil {
		t.Fatal(err)
	}
	e, err := codecengine.New(bundle, wire.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.CompleteRequest(&s6c.SRR{SessionID: "s;1", AuthSessionState: 1, MSISDN: avp.Some([]byte{0x21}), SCAddress: []byte{0x43}, ProxyInfo: []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0}}}}, pdu.RequestEnvelope{OriginHost: "local.example", OriginRealm: "example", DestinationRealm: "example", HopByHop: 1, EndToEnd: 2})
	if err != nil {
		t.Fatal(err)
	}
	policy := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "hss.example", Realm: "example"}}}
	snapshot, err := e.PrepareRequest(r, policy)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := e.RequestBound(r, policy)
	if err != nil || bound != snapshot.Bound() {
		t.Fatal(bound, err)
	}
	answer, err := e.CompleteAnswer(r, &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, policy.Terminals[0])
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := e.Append(nil, answer)
	if err != nil {
		t.Fatal(err)
	}
	wrong := append([]byte(nil), bytes...)
	wrong[15]++
	if _, err = snapshot.DecodeAnswerOwned(wrong); !errors.Is(err, validation.ErrUnmatchedAnswer) || calls.Load() != 0 {
		t.Fatal("unmatched head decoded body", err)
	}
	policy.Terminals[0].Host = "evil.example"
	r.Body.(*s6c.SRR).SessionID = "changed"
	r.Body.(*s6c.SRR).ProxyInfo[0].ProxyState[0] = 9
	if !snapshot.MatchesHeader(answer.Header) || snapshot.Header().HopByHop != 1 {
		t.Fatal("header snapshot")
	}
	if _, err = snapshot.DecodeAnswer(bytes); err != nil {
		t.Fatal("caller mutation changed private evidence", err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 10 {
				if err := snapshot.CheckAnswer(answer); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	// A correct header with invalid bytes is classified as matched-invalid; no body escapes.
	malformed := append([]byte(nil), bytes...)
	wire.Put24(malformed[25:28], 7)
	out, err := snapshot.DecodeAnswerOwned(malformed)
	if !errors.Is(err, validation.ErrInvalidAnswer) || out.Body != nil {
		t.Fatal(err)
	}
}

func TestRuntimeBoundOverflowAndIncompletePair(t *testing.T) {
	p, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	b, err := schema.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := b.Lookup((&s6c.SRR{}).CommandKey())
	a, _ := b.Lookup((&s6c.SRA{}).CommandKey())
	r.DecodeBound = func(int, wire.Limits) (uint64, error) { return ^uint64(0), nil }
	p, err = schema.NewProfile("overflow", "1", p.Digest(), r, a)
	if err != nil {
		t.Fatal(err)
	}
	b, err = schema.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	e, err := codecengine.New(b, wire.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.CompleteRequest(&s6c.SRR{SessionID: "s;1", AuthSessionState: 1, MSISDN: avp.Some([]byte{0x21}), SCAddress: []byte{0x43}}, pdu.RequestEnvelope{OriginHost: "a.example", OriginRealm: "example", DestinationRealm: "example"})
	if err != nil {
		t.Fatal(err)
	}
	policy := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "b.example", Realm: "example"}}}
	if _, err = e.RequestBound(m, policy); !errors.Is(err, wire.ErrLimit) {
		t.Fatal("wrapped reservation", err)
	}
	if _, err = e.CompleteAnswerBound(m, &pdu.ErrorAnswer{ResultCode: 5012}, policy.Terminals[0]); !errors.Is(err, wire.ErrLimit) {
		t.Fatal("wrapped answer reservation", err)
	}
	p, err = schema.NewProfile("one-sided", "1", p.Digest(), r)
	if err != nil {
		t.Fatal(err)
	}
	b, err = schema.Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	e, err = codecengine.New(b, wire.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.PrepareRequest(m, policy); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("pending evidence without answer operations", err)
	}
}
