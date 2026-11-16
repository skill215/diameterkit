// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func sources() pdu.AnswerSources {
	return pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "hss.example", Realm: "example.net"}}, ErrorAgents: []pdu.Origin{{Host: "dra-a.example", Realm: "example.net"}, {Host: "dra-a.example", Realm: "example"}}}
}
func decoded(t testing.TB, c *codec.Codec, name string) pdu.Message {
	t.Helper()
	m, err := c.Decode(fixture(t, name))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func encoded(t testing.TB, c *codec.Codec, m pdu.Message) []byte {
	t.Helper()
	b, err := c.Append(nil, m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEFixedPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	pairs := []struct{ req, answer string }{{"b2-srr", "e28-sra-3002"}, {"b2-srr", "e28-sra-5012"}, {"b2-srr-proxy", "b2-sra-e-proxy"}, {"b2-srr-proxy", "b2-sra-normal-proxy"}, {"v-srr-01", "b2-sra-e-5001"}, {"v-srr-01", "b2-sra-e-5014-badlength"}}
	for _, pair := range pairs {
		t.Run(pair.answer, func(t *testing.T) {
			req := decoded(t, c, pair.req)
			input := fixture(t, pair.answer)
			m, err := c.DecodeAnswer(req, input, sources())
			if err != nil {
				t.Fatal(err)
			}
			clear(input)
			if !bytes.Equal(encoded(t, c, m), fixture(t, pair.answer)) {
				t.Fatal("owned E/normal golden mismatch")
			}
			if err := c.ValidateAnswer(req, m, sources()); err != nil {
				t.Fatal(err)
			}
			if b, ok := m.Body.(*pdu.ErrorAnswer); ok {
				if b.AuthSessionState.Present {
					t.Fatal("E incorrectly requires/fills normal state")
				}
				if pair.answer == "b2-sra-e-5014-badlength" && !b.FailedAVP.Value.Incomplete {
					t.Fatal("malformed quotation must report incomplete")
				}
			}
		})
	}
	req := decoded(t, c, "b2-srr")
	for name, want := range map[string]error{"n28-sra-e-nosession": validation.ErrInvalidAnswer, "n28-sra-e-noresult": validation.ErrInvalidAnswer, "n28-sra-e-e2e": validation.ErrUnmatchedAnswer} {
		m, err := c.DecodeAnswer(req, fixture(t, name), sources())
		if !errors.Is(err, want) || m.Body != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	// Generic E syntax permits absent Session; association with this request does not.
	if _, err := c.Decode(fixture(t, "n28-sra-e-nosession")); err != nil {
		t.Fatal(err)
	}
}

func TestHeaderMatchingVersusInvalidAnswer(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "b2-srr-proxy")
	mismatches := map[string]func(*pdu.Header){"HBH": func(h *pdu.Header) { h.HopByHop++ }, "E2E": func(h *pdu.Header) { h.EndToEnd++ }, "App": func(h *pdu.Header) { h.ApplicationID++ }, "Command": func(h *pdu.Header) { h.CommandCode++ }, "request": func(h *pdu.Header) { h.Flags |= pdu.RequestFlag }}
	for name, mutate := range mismatches {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, "b2-sra-normal-proxy")
			mutate(&a.Header)
			if pdu.MatchesAnswerHeader(req.Header, a.Header) {
				t.Fatal("false header candidate")
			}
			if err := c.ValidateAnswer(req, a, sources()); !errors.Is(err, validation.ErrUnmatchedAnswer) {
				t.Fatal(err)
			}
		})
	}
	invalids := map[string]func(*pdu.Message){"P": func(a *pdu.Message) { a.Header.Flags &= ^pdu.ProxiableFlag }, "T": func(a *pdu.Message) { a.Header.Flags |= pdu.RetransmittedFlag }, "Session": func(a *pdu.Message) { a.Body.(*s6c.SRA).SessionID = "other" }, "Proxy order": func(a *pdu.Message) { p := a.Body.(*s6c.SRA).ProxyInfo; p[0], p[1] = p[1], p[0] }, "Proxy state": func(a *pdu.Message) { a.Body.(*s6c.SRA).ProxyInfo[0].ProxyState[0] = 1 }, "Proxy missing": func(a *pdu.Message) { a.Body.(*s6c.SRA).ProxyInfo = nil }, "ErrorAgent normal": func(a *pdu.Message) { a.Body.(*s6c.SRA).OriginHost = "dra-a.example" }, "Realm": func(a *pdu.Message) { a.Body.(*s6c.SRA).OriginRealm = "evil.example" }}
	for name, mutate := range invalids {
		t.Run(name, func(t *testing.T) {
			a := decoded(t, c, "b2-sra-normal-proxy")
			mutate(&a)
			if !pdu.MatchesAnswerHeader(req.Header, a.Header) {
				t.Fatal("body/P/T must be checked after identifying candidate")
			}
			if err := c.ValidateAnswer(req, a, sources()); !errors.Is(err, validation.ErrInvalidAnswer) {
				t.Fatal(err)
			}
		})
	}
	a := decoded(t, c, "b2-sra-e-proxy")
	b := a.Body.(*pdu.ErrorAnswer)
	b.OriginHost = "evil.example"
	b.ErrorReportingHost = avp.Some("hss.example")
	if err := c.ValidateAnswer(req, a, sources()); !errors.Is(err, validation.ErrSource) {
		t.Fatal("reporter must not authorize Origin", err)
	}
	a = decoded(t, c, "b2-sra-normal-proxy")
	a.Body.(*s6c.SRA).OriginHost = "HSS.EXAMPLE"
	if err := c.ValidateAnswer(req, a, sources()); err != nil {
		t.Fatal("ASCII identity comparison", err)
	}
	raw := fixture(t, "b2-sra-e-proxy")
	raw[4] |= 0x0f
	if _, err := c.DecodeAnswer(req, raw, sources()); err != nil {
		t.Fatal("reserved receive flags", err)
	}
}

func TestCompleteNormalAndErrorAnswers(t *testing.T) {
	c := newCodec(t, codec.Config{})
	origin := pdu.Origin{Host: "hss.example", Realm: "example.net"}
	for _, isError := range []bool{false, true} {
		req := decoded(t, c, "b2-srr-proxy")
		req.Header.Flags |= pdu.RetransmittedFlag
		var body pdu.Body = &s6c.SRA{ResultCode: avp.Some(uint32(5012)), AuthSessionState: 1}
		if isError {
			body = &pdu.ErrorAnswer{ResultCode: 5012}
		}
		bound, err := c.CompleteAnswerBound(req, body, origin)
		if err != nil || bound == 0 {
			t.Fatal(bound, err)
		}
		out, err := c.CompleteAnswer(req, body, origin)
		if err != nil {
			t.Fatal(err)
		}
		if out.Header.Flags&(pdu.RequestFlag|pdu.RetransmittedFlag) != 0 || out.Header.HopByHop != req.Header.HopByHop || out.Header.EndToEnd != req.Header.EndToEnd {
			t.Fatal("answer flags/identifiers")
		}
		if err := c.ValidateAnswer(req, out, sources()); err != nil {
			t.Fatal(err)
		}
		if isError {
			if body.(*pdu.ErrorAnswer).SessionID.Present || out.Body.(*pdu.ErrorAnswer).OriginStateID.Present {
				t.Fatal("input mutation or Origin-State copied")
			}
		} else if body.(*s6c.SRA).SessionID != "" {
			t.Fatal("input mutated")
		}
		before := encoded(t, c, out)
		req.Body.(*s6c.SRR).ProxyInfo[0].ProxyState[0] = 9
		if !bytes.Equal(before, encoded(t, c, out)) {
			t.Fatal("reply aliases request proxy")
		}
	}
	req := decoded(t, c, "b2-srr-proxy")
	for _, body := range []pdu.Body{&s6c.SRA{SessionID: "conflict", AuthSessionState: 1, ResultCode: avp.Some(uint32(5012))}, &pdu.ErrorAnswer{ResultCode: 5012, OriginHost: "spoof.example"}, &pdu.ErrorAnswer{ResultCode: 5012, ProxyInfo: []pdu.ProxyInfo{{ProxyHost: "other.example", ProxyState: []byte{1}}}}} {
		if m, err := c.CompleteAnswer(req, body, origin); !errors.Is(err, validation.ErrEnvelope) || m.Body != nil {
			t.Fatal("conflicting answer", err)
		}
	}
	if _, err := c.CompleteAnswer(decoded(t, c, "v-sra-01"), &pdu.ErrorAnswer{ResultCode: 5012}, origin); !errors.Is(err, validation.ErrType) {
		t.Fatal("answer to an answer", err)
	}
	// A failed build does not consume anything: the same input can then succeed.
	b := &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(uint32(5001))}
	if _, err := c.CompleteAnswer(req, b, origin); !errors.Is(err, validation.ErrMissing) {
		t.Fatal(err)
	}
	b.ResultCode.Value = 5012
	if _, err := c.CompleteAnswer(req, b, origin); err != nil {
		t.Fatal(err)
	}
}

func rawAVP(code uint32, data []byte) []byte {
	n := 8 + len(data)
	b := make([]byte, (n+3)&^3)
	binary.BigEndian.PutUint32(b, code)
	b[4] = 0x40
	wire.Put24(b[5:8], uint32(n))
	copy(b[8:], data)
	return b
}
func appendBytes(frame []byte, tail []byte) []byte {
	b := append(append([]byte(nil), frame...), tail...)
	wire.Put24(b[1:4], uint32(len(b)))
	return b
}
func findTop(t testing.TB, b []byte, code uint32) (int, int) {
	t.Helper()
	for at := 20; at < len(b); {
		n := int(wire.Uint24(b[at+5 : at+8]))
		if binary.BigEndian.Uint32(b[at:]) == code && b[at+4]&0x80 == 0 {
			return at, n
		}
		at += (n + 3) &^ 3
	}
	t.Fatalf("missing %d", code)
	return 0, 0
}
func replaceTop(t testing.TB, b []byte, code uint32, replacement []byte) []byte {
	at, n := findTop(t, b, code)
	out := append(append(append([]byte(nil), b[:at]...), replacement...), b[at+((n+3)&^3):]...)
	wire.Put24(out[1:4], uint32(len(out)))
	return out
}

func TestErrorGrammarAndDiagnostics(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, code := range []uint32{2001, 4001, 6000} {
		b := replaceTop(t, fixture(t, "e28-sra-3002"), 268, rawAVP(268, []byte{byte(code >> 24), byte(code >> 16), byte(code >> 8), byte(code)}))
		if _, err := c.Decode(b); !errors.Is(err, validation.ErrField) {
			t.Fatal(code, err)
		}
	}
	for _, code := range []uint32{5001, 5004, 5007, 5008, 5009, 5014} {
		b := replaceTop(t, fixture(t, "e28-sra-3002"), 268, rawAVP(268, []byte{0, 0, byte(code >> 8), byte(code)}))
		if _, err := c.Decode(b); !errors.Is(err, validation.ErrMissing) {
			t.Fatal(code, err)
		}
	}
	missing := replaceTop(t, fixture(t, "e28-sra-3002"), 268, rawAVP(268, []byte{0, 0, 0x13, 0x8d}))
	if _, err := c.Decode(missing); err != nil {
		t.Fatal("5005 receive SHOULD is not MUST", err)
	}
	// Supplementary experimental namespace cannot replace/overwrite the base failure.
	extra := rawAVP(297, append(rawAVP(266, []byte{0, 0, 0, 99}), rawAVP(298, []byte{0, 0, 7, 209})...))
	m, err := c.Decode(appendBytes(fixture(t, "e28-sra-3002"), extra))
	if err != nil {
		t.Fatal(err)
	}
	eb := m.Body.(*pdu.ErrorAnswer)
	if eb.ResultCode != 3002 || eb.ExperimentalResult.Value.VendorID != 99 || eb.ExperimentalResult.Value.ExperimentalResultCode != 2001 {
		t.Fatal("namespaces merged")
	}
	negatives := map[string][]byte{
		"duplicate result":   appendBytes(fixture(t, "e28-sra-3002"), rawAVP(268, []byte{0, 0, 0x0b, 0xba})),
		"destination":        appendBytes(fixture(t, "e28-sra-3002"), rawAVP(283, []byte("example.net"))),
		"bad text":           appendBytes(fixture(t, "e28-sra-3002"), rawAVP(281, []byte{0xc3, 0x28})),
		"NUL text":           appendBytes(fixture(t, "e28-sra-3002"), rawAVP(281, []byte{'a', 0, 'b'})),
		"empty failed":       replaceTop(t, fixture(t, "b2-sra-e-5001"), 279, rawAVP(279, nil)),
		"duplicate failed":   appendBytes(fixture(t, "b2-sra-e-5001"), rawAVP(279, rawAVP(99999, []byte{1}))),
		"split experimental": appendBytes(fixture(t, "e28-sra-3002"), append(rawAVP(297, rawAVP(266, []byte{0, 0, 0, 99})), rawAVP(297, rawAVP(298, []byte{0, 0, 0, 1}))...)),
	}
	for name, b := range negatives {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Decode(b); err == nil {
				t.Fatal("malformed envelope accepted")
			}
		})
	}
	// Quoted short headers, unknown M=1 and bad UTF8 stay diagnostic evidence.
	for _, quote := range [][]byte{{0, 0, 1, 8}, rawAVP(99999, []byte{0xff}), rawAVP(281, []byte{0xff, 0})} {
		b := replaceTop(t, fixture(t, "b2-sra-e-5014-badlength"), 279, rawAVP(279, quote))
		m, err := c.Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		got := m.Body.(*pdu.ErrorAnswer).FailedAVP.Value
		if !bytes.Equal(got.Data, quote) {
			t.Fatal("quote changed")
		}
		if len(quote) < 8 {
			if !got.Incomplete {
				t.Fatal("short quote not reported")
			}
			if _, err = c.Append(nil, m); err == nil {
				t.Fatal("short quotation is not a standard send builder")
			}
		}
	}
	// This batch does not silently treat known application extensions as opaque
	// successes of the common E slice; their E-specific semantics remain explicit.
	knownApp := avp.Raw{Key: avp.Key{VendorID: 10415, Code: 628}, Flags: 0x80, Data: []byte{1}}
	if _, err = c.Decode(addRaw(t, c, "e28-sra-3002", knownApp)); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("known unimplemented E field accepted as opaque", err)
	}
	redirect := replaceTop(t, fixture(t, "e28-sra-3002"), 268, rawAVP(268, []byte{0, 0, 0x0b, 0xbe}))
	if _, err = c.Decode(redirect); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("redirect requires its own fields", err)
	}
}

func TestDiagnosticGlobalBudgets(t *testing.T) {
	quote := rawAVP(99999, []byte{1})
	for range 8 {
		quote = rawAVP(284, quote)
	}
	b := replaceTop(t, fixture(t, "b2-sra-e-5014-badlength"), 279, rawAVP(279, quote))
	l := wire.DefaultLimits()
	l.MaxDepth = 9
	l.MaxFieldBytes = len(quote)
	c := newCodec(t, codec.Config{Limits: l})
	m, err := c.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Append(nil, m); err != nil {
		t.Fatal("send and receive must share diagnostic depth budget", err)
	}
	for _, change := range []func(*wire.Limits){func(l *wire.Limits) { l.MaxDepth = 8 }, func(l *wire.Limits) { l.MaxFieldBytes-- }, func(l *wire.Limits) { l.MaxAVPs = 13; l.MaxRepeated = 13 }} {
		limit := l
		change(&limit)
		limited := newCodec(t, codec.Config{Limits: limit})
		if _, err := limited.Decode(b); !errors.Is(err, wire.ErrLimit) {
			t.Fatalf("decode diagnostic budget: %v", err)
		}
		if _, err := limited.Append(nil, m); !errors.Is(err, wire.ErrLimit) {
			t.Fatalf("encode diagnostic budget: %v", err)
		}
	}
}

func TestCommonNormalFieldsAndSMSIgnore(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "v-srr-01")
	b := req.Body.(*s6c.SRR)
	b.RouteRecord = []string{"dra-a.example", "dra-b.example"}
	b.SupportedFeatures = []s6c.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: 0}, {VendorID: 10415, FeatureListID: 1, FeatureList: 7}}
	input := encoded(t, c, req)
	round, err := c.Decode(input)
	if err != nil {
		t.Fatal(err)
	}
	rb := round.Body.(*s6c.SRR)
	if len(rb.SupportedFeatures) != 2 || rb.SupportedFeatures[0].FeatureList != 0 || rb.SupportedFeatures[1].FeatureList != 7 || rb.RouteRecord[1] != "dra-b.example" {
		t.Fatal("ordered repeated groups/zero lost")
	}
	// Inner VSAI is deliberately malformed; the SMS receive rule ignores semantics.
	input = appendBytes(input, rawAVP(260, []byte{1, 2}))
	r, err := c.Decode(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Append(nil, r); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatal("SMS must not send ignored VSAI", err)
	}
	if _, err = c.CompleteAnswer(r, &pdu.ErrorAnswer{ResultCode: 3002}, pdu.Origin{Host: "dra-a.example", Realm: "example.net"}); err != nil {
		t.Fatal("ignored VSAI prevents answering", err)
	}
	// Known fields cannot be smuggled into answers through M=0 raw extensions.
	normal := decoded(t, c, "v-sra-01")
	normal.Body.(*s6c.SRA).Extra = []avp.Raw{{Key: avp.Key{Code: 293}, Data: []byte("hss.example")}}
	if _, err = c.Append(nil, normal); !errors.Is(err, validation.ErrField) {
		t.Fatal(err)
	}
	normal = decoded(t, c, "v-sra-01")
	normal.Body.(*s6c.SRA).ErrorMessage = avp.Some("invalid\x00text")
	if err := c.Validate(normal); err == nil {
		t.Fatal("normal error text NUL")
	}
}

func TestBaseControlCompletionAndVSAI(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "base-dwr")
	for _, body := range []pdu.Body{&base.DWA{ResultCode: 2001}, &pdu.ErrorAnswer{ResultCode: 5012}} {
		m, err := c.CompleteAnswer(req, body, pdu.Origin{Host: "peer.example", Realm: "example.net"})
		if err != nil {
			t.Fatal(err)
		}
		if m.Header.Flags&pdu.ProxiableFlag != 0 {
			t.Fatal("base P")
		}
		if err = c.ValidateAnswer(req, m, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "peer.example", Realm: "example.net"}}}); err != nil {
			t.Fatal(err)
		}
	}
	request := &base.CER{OriginHost: "peer.example", OriginRealm: "example.net", HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "test", VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(uint32(16777312))}}}
	m := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 257}, Body: request}
	got := decodedFrame(t, c, encoded(t, c, m)).Body.(*base.CER)
	if got.VendorSpecificApplicationID[0].AuthApplicationID.Value != 16777312 {
		t.Fatal("VSAI")
	}
	request.VendorSpecificApplicationID[0].AcctApplicationID = avp.Some(uint32(4))
	if err := c.Validate(m); err == nil {
		t.Fatal("VSAI permits Auth and Acct together")
	}
}
func decodedFrame(t testing.TB, c *codec.Codec, b []byte) pdu.Message {
	t.Helper()
	m, err := c.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestSnapshotAndCompletionBounds(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "b2-srr-proxy")
	snapshot, err := c.Snapshot(req)
	if err != nil {
		t.Fatal(err)
	}
	before := encoded(t, c, snapshot)
	req.Body.(*s6c.SRR).MSISDN.Value[0] = 0xff
	req.Body.(*s6c.SRR).ProxyInfo[0].ProxyState[0] = 4
	if !bytes.Equal(before, encoded(t, c, snapshot)) {
		t.Fatal("snapshot aliases caller")
	}
	n, err := c.Measure(snapshot)
	if err != nil || n != len(before) {
		t.Fatal(n, err)
	}
	bound, err := c.DecodeBound(snapshot.Header, n)
	if err != nil || bound < uint64(n) {
		t.Fatal(bound, err)
	}
	limited := wire.DefaultLimits()
	limited.MaxFrameBytes = 120
	limited.MaxFieldBytes = 120
	c2 := newCodec(t, codec.Config{Limits: limited})
	if _, err := c2.Snapshot(snapshot); !errors.Is(err, wire.ErrLimit) {
		t.Fatal("snapshot limit", err)
	}
	for _, s := range []pdu.AnswerSources{{}, {Terminals: make([]pdu.Origin, 65)}, {Terminals: []pdu.Origin{{Host: "hss.example", Realm: ""}}}} {
		if err := c.ValidateAnswer(snapshot, decoded(t, c, "b2-sra-normal-proxy"), s); !errors.Is(err, validation.ErrSource) {
			t.Fatal("unbounded/empty source policy", err)
		}
	}
	ebody := &pdu.ErrorAnswer{ResultCode: 5012, ErrorMessage: avp.Optional[string]{Value: string(make([]byte, 1<<20))}}
	out, err := c.CompleteAnswer(snapshot, ebody, pdu.Origin{Host: "hss.example", Realm: "example.net"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Body.(*pdu.ErrorAnswer).ErrorMessage.Value != "" {
		t.Fatal("absent field retains unbudgeted storage")
	}
}

func FuzzDecodeAnswer(f *testing.F) {
	c := newCodec(f, codec.Config{})
	req := decoded(f, c, "v-srr-01")
	for _, name := range []string{"v-sra-01", "v-sra-02", "b2-sra-e-5001", "b2-sra-e-5014-badlength"} {
		f.Add(fixture(f, name))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		m, err := c.DecodeAnswer(req, b, sources())
		if err != nil {
			if m.Body != nil {
				t.Fatal("partial invalid answer")
			}
			return
		}
		if err = c.ValidateAnswer(req, m, sources()); err != nil {
			t.Fatal(err)
		}
	})
}
