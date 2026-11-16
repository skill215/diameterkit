// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package codec_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func newCodec(t testing.TB, cfg codec.Config) *codec.Codec {
	t.Helper()
	bp, err := base.Profile()
	if err != nil {
		t.Fatal(err)
	}
	sp, err := s6c.Profile()
	if err != nil {
		t.Fatal(err)
	}
	b, err := schema.Compile(bp, sp)
	if err != nil {
		t.Fatal(err)
	}
	c, err := codec.New(b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func fixture(t testing.TB, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", "codec", name+".hex"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := hex.DecodeString(strings.Join(strings.Fields(string(b)), ""))
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func TestCandidateManifest(t *testing.T) {
	var entries []struct {
		File, SHA256 string
		Bytes        int
	}
	b, err := os.ReadFile("../testdata/codec/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &entries); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b := fixture(t, strings.TrimSuffix(e.File, ".hex"))
		sum := sha256.Sum256(b)
		if len(b) != e.Bytes || hex.EncodeToString(sum[:]) != e.SHA256 {
			t.Fatalf("fixture changed: %s", e.File)
		}
	}
}
func TestIndependentCandidatePairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, name := range []string{"base-dwr", "base-dwa", "v-srr-01", "v-sra-01", "v-sra-02"} {
		t.Run(name, func(t *testing.T) {
			input := fixture(t, name)
			m, err := c.Decode(input)
			if err != nil {
				t.Fatal(err)
			}
			if m.Header.HopByHop != 0x51525354 || m.Header.EndToEnd != 0x61626364 {
				t.Fatal("pair IDs lost")
			}
			switch b := m.Body.(type) {
			case *base.DWR:
				if !b.OriginStateID.Present || b.OriginStateID.Value != 0 {
					t.Fatal("present zero lost")
				}
			case *base.DWA:
				if b.ResultCode != 2001 {
					t.Fatal("result")
				}
			case *s6c.SRR:
				s, err := avp.DecodeTBCD(b.MSISDN.Value)
				if err != nil || s != "12025550123" || !b.SMRPMTI.Present || b.SMRPMTI.Value != 0 {
					t.Fatalf("request values %s: %v", s, err)
				}
			case *s6c.SRA:
				if name == "v-sra-01" {
					if !b.ResultCode.Present || b.ResultCode.Value != 2001 || b.UserName.Value != "001010123456789" || b.ServingNode.Value.MMEName != avp.Some("mme.example") {
						t.Fatal("success fields")
					}
				} else if b.ResultCode.Present || b.ExperimentalResult.Value.VendorID != 10415 || b.ExperimentalResult.Value.ExperimentalResultCode != 5001 {
					t.Fatal("failure namespace")
				}
			default:
				t.Fatalf("unexpected body %T", b)
			}
			clear(input) // Decode must own everything it retained.
			got, err := c.Append(nil, m)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, fixture(t, name)) {
				t.Fatal("golden mismatch after input reuse")
			}
		})
	}
}
func TestCompleteRequestCopiesAndRejectsConflicts(t *testing.T) {
	c := newCodec(t, codec.Config{})
	m, err := c.Decode(fixture(t, "v-srr-01"))
	if err != nil {
		t.Fatal(err)
	}
	b := m.Body.(*s6c.SRR)
	b.OriginHost = ""
	b.OriginRealm = ""
	b.DestinationRealm = ""
	b.DestinationHost = avp.Optional[string]{Value: "ignored.invalid"}
	b.Extra = []avp.Raw{{Key: avp.Key{VendorID: 555, Code: 701}, Flags: avp.VendorFlag, Data: []byte{1, 2, 3}}}
	if err = c.ValidateBody(b); err != nil {
		t.Fatal(err)
	}
	if err = c.Validate(m); err == nil {
		t.Fatal("incomplete envelope accepted for send")
	}
	env := pdu.RequestEnvelope{OriginHost: "gmsc.example", OriginRealm: "example.net", DestinationRealm: "example.net", DestinationHost: "hss.example", HopByHop: 7, EndToEnd: 9}
	complete, err := c.CompleteRequest(b, env)
	if err != nil {
		t.Fatal(err)
	}
	if b.OriginHost != "" || b.DestinationHost.Present {
		t.Fatal("input changed")
	}
	out := complete.Body.(*s6c.SRR)
	b.MSISDN.Value[0] = 0xff
	b.Extra[0].Data[0] = 99
	if out.MSISDN.Value[0] == 0xff || out.Extra[0].Data[0] == 99 || complete.Header.HopByHop != 7 {
		t.Fatal("clone/IDs")
	}
	env.OriginHost = "other.example"
	if _, err = c.CompleteRequest(out, env); !errors.Is(err, validation.ErrEnvelope) {
		t.Fatalf("conflict: %v", err)
	}
}
func addRaw(t *testing.T, c *codec.Codec, name string, r avp.Raw) []byte {
	t.Helper()
	input := fixture(t, name)
	w := wire.NewEncoder(nil, c.Limits())
	w.Bytes(r.Key, r.Flags, r.Data)
	if w.Err() != nil {
		t.Fatal(w.Err())
	}
	tail := make([]byte, w.Len())
	w = wire.NewEncoder(tail, c.Limits())
	w.Bytes(r.Key, r.Flags, r.Data)
	input = append(input, tail...)
	wire.Put24(input[1:4], uint32(len(input)))
	return input
}
func TestRawVendorParentAndDiagnosticBoundaries(t *testing.T) {
	c := newCodec(t, codec.Config{})
	// Unknown M=0 is opaque even when its bytes resemble a malformed child AVP.
	raw := avp.Raw{Key: avp.Key{VendorID: 555, Code: 701}, Flags: avp.VendorFlag, Data: []byte{0, 0, 1}}
	data := addRaw(t, c, "v-srr-01", raw)
	m, err := c.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	b := m.Body.(*s6c.SRR)
	if len(b.Extra) != 1 || !bytes.Equal(b.Extra[0].Data, raw.Data) {
		t.Fatal("raw lost")
	}
	raw.Flags |= avp.MandatoryFlag
	if _, err = c.Decode(addRaw(t, c, "v-srr-01", raw)); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatalf("unknown M: %v", err)
	}
	// A known MME child placed at the root must not populate a ServingNode.
	raw = avp.Raw{Key: avp.Key{VendorID: 10415, Code: 2402}, Flags: 0xc0, Data: []byte("mme.example")}
	if _, err = c.Decode(addRaw(t, c, "v-sra-01", raw)); !errors.Is(err, validation.ErrUnsupported) {
		t.Fatalf("wrong parent: %v", err)
	}
	data = fixture(t, "v-sra-01")
	data[164] = '!'
	_, err = c.Decode(data)
	var field *validation.FieldError
	if err == nil || !errors.As(err, &field) || len(field.Path) != 2 || field.Path[0].Code != 2401 || field.Path[1].Code != 2402 {
		t.Fatalf("nested path: %v", err)
	}
	if strings.Contains(err.Error(), "mme.example") {
		t.Fatal("diagnostic leaks data")
	}
}
func TestRejectMalformedAndBounded(t *testing.T) {
	c := newCodec(t, codec.Config{})
	tests := map[string]func([]byte) []byte{
		"short header":    func(b []byte) []byte { return b[:19] },
		"truncated":       func(b []byte) []byte { return b[:len(b)-1] },
		"trailing":        func(b []byte) []byte { return append(b, 0) },
		"version":         func(b []byte) []byte { b[0] = 2; return b },
		"request E":       func(b []byte) []byte { b[4] |= pdu.ErrorFlag; return b },
		"wrong P":         func(b []byte) []byte { b[4] &= ^pdu.ProxiableFlag; return b },
		"short AVP":       func(b []byte) []byte { wire.Put24(b[25:28], 7); return b },
		"oversized AVP":   func(b []byte) []byte { wire.Put24(b[25:28], 0xffffff); return b },
		"missing Session": func(b []byte) []byte { b = append(b[:20], b[48:]...); wire.Put24(b[1:4], uint32(len(b))); return b },
		"duplicate Session": func(b []byte) []byte {
			b = append(b, append([]byte(nil), b[20:48]...)...)
			wire.Put24(b[1:4], uint32(len(b)))
			return b
		},
		"bad u32":        func(b []byte) []byte { wire.Put24(b[53:56], 11); return b },
		"missing vendor": func(b []byte) []byte { clear(b[148:152]); return b },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Decode(mutate(fixture(t, "v-srr-01"))); err == nil {
				t.Fatal("malformed accepted")
			}
		})
	}
	for _, change := range []func(*codec.Limits){func(l *codec.Limits) { l.MaxFrameBytes = 100; l.MaxFieldBytes = 100 }, func(l *codec.Limits) { l.MaxFieldBytes = 4 }, func(l *codec.Limits) { l.MaxAVPs = 3; l.MaxRepeated = 3 }} {
		l := wire.DefaultLimits()
		change(&l)
		limited := newCodec(t, codec.Config{Limits: l})
		if _, err := limited.Decode(fixture(t, "v-srr-01")); !errors.Is(err, wire.ErrLimit) {
			t.Fatalf("limit: %v", err)
		}
	}
}
func TestResultSelectionAndIdentity(t *testing.T) {
	c := newCodec(t, codec.Config{})
	cases := map[string]func(*s6c.SRA){
		"both": func(b *s6c.SRA) {
			b.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})
		},
		"neither":        func(b *s6c.SRA) { b.ResultCode.Present = false },
		"missing IMSI":   func(b *s6c.SRA) { b.UserName.Present = false },
		"invalid IMSI":   func(b *s6c.SRA) { b.UserName.Value = "imsi" },
		"incomplete MME": func(b *s6c.SRA) { b.ServingNode.Value.MMERealm = avp.Optional[string]{} },
		"other result":   func(b *s6c.SRA) { b.ResultCode.Value = 5001 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			m, err := c.Decode(fixture(t, "v-sra-01"))
			if err != nil {
				t.Fatal(err)
			}
			change(m.Body.(*s6c.SRA))
			if name == "invalid IMSI" {
				req := decoded(t, c, "v-srr-01")
				b := req.Body.(*s6c.SRR)
				ctx, e := s6c.RoutingRole(s6c.RoutingRoleEvidence{Role: s6c.RoutingFromGMSC, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}})
				if e != nil {
					t.Fatal(e)
				}
				if r := c.ValidateAnswerContext(req, m, routingSources(), ctx); r.Complete {
					t.Fatal("ordinary HSS accepted token")
				}
				return
			}
			if _, err = c.Append(nil, m); err == nil {
				t.Fatal("invalid branch accepted")
			}
		})
	}
	m, _ := c.Decode(fixture(t, "v-sra-02"))
	m.Body.(*s6c.SRA).UserName = avp.Some("001010123456789")
	if err := c.Validate(m); err == nil {
		t.Fatal("failure contains success identity")
	}
	e := fixture(t, "v-sra-02")
	e[4] |= pdu.ErrorFlag
	if _, err := c.Decode(e); !errors.Is(err, validation.ErrMissing) {
		t.Fatalf("E grammar must require its own Result-Code: %v", err)
	}
}

type impostor struct{}

func (*impostor) CommandKey() pdu.CommandKey { return (&s6c.SRR{}).CommandKey() }
func TestTypeAndAppendRollback(t *testing.T) {
	c := newCodec(t, codec.Config{})
	var typedNil *s6c.SRR
	for _, body := range []pdu.Body{nil, typedNil, &impostor{}} {
		if err := c.ValidateBody(body); !errors.Is(err, validation.ErrType) {
			t.Fatalf("bad type: %v", err)
		}
	}
	m, _ := c.Decode(fixture(t, "v-srr-01"))
	m.Body.(*s6c.SRR).MSISDN = avp.Some([]byte{255})
	dst := make([]byte, 3, 1000)
	copy(dst, "abc")
	got, err := c.Append(dst, m)
	if err == nil || len(got) != 3 || string(got) != "abc" || &got[0] != &dst[0] {
		t.Fatal("rollback")
	}
}

func TestOperationFailureAndBudgetBeforeClone(t *testing.T) {
	bp, err := base.Profile()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := schema.Compile(bp)
	if err != nil {
		t.Fatal(err)
	}
	op, _ := bundle.Lookup((&base.DWR{}).CommandKey())
	encode := op.Encode
	calls := 0
	op.Encode = func(w *wire.Encoder, b pdu.Body) error {
		calls++
		if err := encode(w, b); err != nil {
			return err
		}
		if calls == 2 {
			return io.ErrClosedPipe
		}
		return nil
	}
	profile, err := schema.NewProfile("failure-test", "1", bp.Digest(), op)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err = schema.Compile(profile)
	if err != nil {
		t.Fatal(err)
	}
	c, err := codec.New(bundle, codec.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280}, Body: &base.DWR{OriginHost: "a.example", OriginRealm: "example"}}
	dst := make([]byte, 3, 1000)
	copy(dst, "abc")
	got, err := c.Append(dst, m)
	if !errors.Is(err, io.ErrClosedPipe) || calls != 2 || len(got) != 3 || string(got) != "abc" {
		t.Fatal("second-pass rollback", err)
	}
	op.Encode = encode
	completed := false
	complete := op.CompleteRequest
	op.CompleteRequest = func(b pdu.Body, env pdu.RequestEnvelope) (pdu.Body, error) { completed = true; return complete(b, env) }
	profile, err = schema.NewProfile("budget-test", "1", bp.Digest(), op)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err = schema.Compile(profile)
	if err != nil {
		t.Fatal(err)
	}
	limits := wire.DefaultLimits()
	limits.MaxFieldBytes = 20
	c, err = codec.New(bundle, codec.Config{Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	m.Body.(*base.DWR).Extra = []avp.Raw{{Key: avp.Key{Code: 999}, Data: make([]byte, 21)}}
	if _, err = c.CompleteRequest(m.Body, pdu.RequestEnvelope{}); !errors.Is(err, wire.ErrLimit) || completed {
		t.Fatal("clone before reservation bound", err)
	}
	m.Body.(*base.DWR).Extra = nil
	completed = false
	if _, err = c.CompleteRequest(m.Body, pdu.RequestEnvelope{OriginHost: strings.Repeat("x", 1<<20)}); !errors.Is(err, validation.ErrEnvelope) || completed {
		t.Fatal("invalid envelope copied before validation", err)
	}
	if _, err = c.CompleteRequest(m.Body, pdu.RequestEnvelope{DestinationRealm: "example"}); !errors.Is(err, validation.ErrEnvelope) {
		t.Fatal("unsupported envelope", err)
	}
}

type shortWriter struct {
	calls int
	n     int
	err   error
}

func (w *shortWriter) Write(b []byte) (int, error) { w.calls++; return w.n, w.err }
func TestFrameIO(t *testing.T) {
	c := newCodec(t, codec.Config{})
	b := fixture(t, "base-dwr")
	r := bytes.NewReader(append(append([]byte{}, b...), b...))
	if _, err := c.Read(r); err != nil {
		t.Fatal(err)
	}
	if r.Len() != len(b) {
		t.Fatal("consumed next frame")
	}
	if _, err := c.Read(bytes.NewReader(b[:len(b)-1])); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	r = bytes.NewReader(b)
	if _, err := codec.ReadFrame(r, 20); !errors.Is(err, wire.ErrLimit) || r.Len() != len(b)-20 {
		t.Fatal("must reject at header before body read")
	}
	for _, w := range []*shortWriter{{n: 3}, {n: len(b), err: io.ErrClosedPipe}} {
		n, err := codec.WriteFrame(w, b, 1<<20)
		if n != w.n || err == nil || w.calls != 1 {
			t.Fatal("write evidence/tail retry")
		}
	}
}
func TestBasePairsAndConcurrentCodec(t *testing.T) {
	c := newCodec(t, codec.Config{})
	ip, err := avp.IPAddress(netip.MustParseAddr("127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	bodies := []pdu.Body{&base.CER{OriginHost: "a.example", OriginRealm: "example", HostIPAddress: []avp.Address{ip}, VendorID: 0, ProductName: "diameterkit", AuthApplicationID: []uint32{16777312}}, &base.CEA{ResultCode: 2001, OriginHost: "b.example", OriginRealm: "example", HostIPAddress: []avp.Address{ip}, ProductName: "peer"}, &base.DPR{OriginHost: "a.example", OriginRealm: "example", DisconnectCause: 0}, &base.DPA{ResultCode: 2001, OriginHost: "b.example", OriginRealm: "example"}}
	for _, body := range bodies {
		k := body.CommandKey()
		flags := byte(0)
		if k.Request {
			flags = pdu.RequestFlag
		}
		m := pdu.Message{Header: pdu.Header{Flags: flags, CommandCode: k.Code}, Body: body}
		b, err := c.Append(nil, m)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = c.Decode(b); err != nil {
			t.Fatal(err)
		}
	}
	input := fixture(t, "v-sra-01")
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				m, err := c.Decode(input)
				if err != nil {
					t.Error(err)
					return
				}
				if _, err = c.Append(nil, m); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
}
func FuzzDecode(f *testing.F) {
	c := newCodec(f, codec.Config{})
	for _, name := range []string{"base-dwr", "v-srr-01", "v-sra-01", "v-sra-02"} {
		f.Add(fixture(f, name))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		m, err := c.Decode(b)
		if err == nil {
			out, err := c.Append(nil, m)
			if err == nil {
				if _, err = c.Decode(out); err != nil {
					t.Fatalf("accepted output does not decode: %v", err)
				}
			}
		}
	})
}
