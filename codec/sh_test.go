package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func shCodec(t *testing.T, config codec.Config) *codec.Codec {
	t.Helper()
	profiles := []schema.Profile{}
	for _, makeProfile := range []func() (schema.Profile, error){base.Profile, s6c.Profile, sgd.Profile, sh.Profile} {
		p, e := makeProfile()
		if e != nil {
			t.Fatal(e)
		}
		profiles = append(profiles, p)
	}
	b, e := schema.Compile(profiles...)
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(b, config)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestShGoldenPairs(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, pair := range [][3]string{{"v-sh-udr-01", "v-sh-uda-01", "v-sh-uda-02"}, {"v-sh-pur-01", "v-sh-pua-01", "v-sh-pua-02"}, {"v-sh-snr-01", "v-sh-sna-01", "v-sh-sna-02"}, {"v-sh-pnr-01", "v-sh-pna-01", "v-sh-pna-02"}} {
		t.Run(pair[0], func(t *testing.T) {
			req := decoded(t, c, pair[0])
			origin := pdu.Origin{Host: "hss.example", Realm: "example.net"}
			if req.Header.CommandCode == 309 {
				origin.Host = "as.example"
			}
			src := pdu.AnswerSources{Terminals: []pdu.Origin{origin}}
			for i, name := range pair {
				raw := fixture(t, name)
				m, e := c.Decode(raw)
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(encoded(t, c, m), raw) {
					t.Fatal("fixture re-encoding", name)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), fixture(t, name)) {
					t.Fatal("borrowed data")
				}
				n, e := c.Measure(m)
				if e != nil || n != len(fixture(t, name)) {
					t.Fatal(n, e)
				}
				if b, e := c.DecodeBound(m.Header, n); e != nil || b < uint64(n) {
					t.Fatal(b, e)
				}
				if i == 0 {
					continue
				}
				if e = c.ValidateAnswer(req, m, src); e != nil {
					t.Fatal(e)
				}
				if r := c.ValidateAnswerContext(req, m, src, validation.Context{}); !r.Complete {
					t.Fatal(r)
				}
			}
		})
	}
}
func shResult(code uint32, experimental bool, command uint32) pdu.Body {
	rc := avp.Some(code)
	ex := avp.Optional[sh.ExperimentalResult]{}
	if experimental {
		rc = avp.Optional[uint32]{}
		ex = avp.Some(sh.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
	}
	switch command {
	case 306:
		a := &sh.UDA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
		if !experimental && code == 2001 {
			a.UserData = avp.Some([]byte("<Sh-Data/>"))
		}
		return a
	case 307:
		return &sh.PUA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	case 308:
		return &sh.SNA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	default:
		return &sh.PNA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: rc, ExperimentalResult: ex}
	}
}
func TestShResultNamespacesAndCommandBranches(t *testing.T) {
	c := shCodec(t, codec.Config{})
	origin := pdu.Origin{Host: "peer.example", Realm: "example.net"}
	allowed := map[uint32]map[uint32]bool{306: {5001: true, 5101: true, 5102: true, 4100: true}, 307: {5001: true, 5101: true, 5103: true, 4101: true}, 308: {5001: true, 5101: true, 5104: true}, 309: {5001: true, 5008: true, 5100: true, 5107: true}}
	for _, name := range []string{"v-sh-udr-01", "v-sh-pur-01", "v-sh-snr-01", "v-sh-pnr-01"} {
		req := decoded(t, c, name)
		for _, code := range []uint32{0, 2001, 4100, 4101, 5001, 5002, 5008, 5011, 5100, 5101, 5102, 5103, 5104, 5105, 5106, 5107, 5108} {
			_, e := c.CompleteAnswer(req, shResult(code, true, req.Header.CommandCode), origin)
			if (e == nil) != allowed[req.Header.CommandCode][code] {
				t.Fatalf("%s experiment %d: %v", name, code, e)
			}
		}
		for _, code := range []uint32{2001, 5012} {
			a, e := c.CompleteAnswer(req, shResult(code, false, req.Header.CommandCode), origin)
			if e != nil {
				t.Fatal(e)
			}
			if e = c.ValidateAnswer(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{origin}}); e != nil {
				t.Fatal(e)
			}
		}
		a, e := c.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, origin)
		if e != nil {
			t.Fatal(e)
		}
		if a.Header.Flags&pdu.ErrorFlag == 0 {
			t.Fatal("lost E")
		}
		if e = c.ValidateAnswer(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{origin}}); e != nil {
			t.Fatal(e)
		}
		if e = c.ValidateAnswer(req, a, pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "other.example", Realm: origin.Realm}}}); !errors.Is(e, validation.ErrSource) {
			t.Fatal(e)
		}
	}
	a := decoded(t, c, "v-sh-uda-01")
	b := a.Body.(*sh.UDA)
	b.ExperimentalResult = avp.Some(sh.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})
	if e := c.Validate(a); !errors.Is(e, validation.ErrField) {
		t.Fatal(e)
	}
	b.ResultCode = avp.Optional[uint32]{}
	b.UserData = avp.Optional[[]byte]{}
	b.ExperimentalResult.Value.VendorID = 0
	if e := c.Validate(a); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
func TestShRequiredVSAIAndIdentityBoundaries(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, tc := range []struct {
		name   string
		mutate func(*sh.UDR)
		want   error
	}{
		{"wrong vendor", func(m *sh.UDR) { m.VendorSpecificApplicationID.VendorID = 0 }, validation.ErrField},
		{"wrong app", func(m *sh.UDR) { m.VendorSpecificApplicationID.AuthApplicationID.Value = 16777312 }, validation.ErrField},
		{"accounting instead", func(m *sh.UDR) {
			m.VendorSpecificApplicationID.AuthApplicationID = avp.Optional[uint32]{}
			m.VendorSpecificApplicationID.AcctApplicationID = avp.Some(sh.ApplicationID)
		}, validation.ErrField},
		{"no identity", func(m *sh.UDR) { m.UserIdentity = sh.UserIdentity{} }, validation.ErrField},
		{"two identities", func(m *sh.UDR) { m.UserIdentity.MSISDN = avp.Some([]byte{0x21, 0xf3}) }, validation.ErrField},
		{"ordinary MSISDN", func(m *sh.UDR) { m.UserIdentity = sh.UserIdentity{MSISDN: avp.Some([]byte{0x21, 0xf3})} }, nil},
		{"wrong ref", func(m *sh.UDR) { m.DataReference = 25 }, validation.ErrUnsupported},
		{"feature M0", func(m *sh.UDR) { m.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 628}, Flags: 128}} }, validation.ErrDuplicate},
		{"external M0", func(m *sh.UDR) {
			m.UserIdentity.Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 3111}, Flags: 128, Data: []byte("a@example.net")}}
		}, validation.ErrUnsupported},
		{"private M0", func(m *sh.UDR) { m.Extra = []avp.Raw{{Key: avp.Key{Code: 1}, Data: []byte("private@example.net")}} }, validation.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := decoded(t, c, "v-sh-udr-01")
			tc.mutate(m.Body.(*sh.UDR))
			if e := c.Validate(m); !errors.Is(e, tc.want) {
				t.Fatal(e)
			}
		})
	}
	for _, tc := range []struct {
		value string
		pass  bool
	}{
		{"sip:alice@example.net", true}, {"sip:a%2Fb@EXAMPLE.net", true}, {"sip:a&team@example.net", true}, {"tel:+12025550123", true},
		{"sip:alice", false}, {"sip:@example.net", false}, {"sip:a%GG@example.net", false}, {"sip:用户@example.net", false}, {"sip:a@例子.net", false},
		{"sip:a:password@example.net", false}, {"sip:a@example.net:5060", false}, {"sip:a@example.net;transport=tcp", false}, {"sip:a@[::1]", false}, {"sip:a@999.999.1.1", false}, {"tel:+1202-555-0123", false}, {"tel:+123;ext=1", false}, {"tel:+0123", false},
	} {
		m := decoded(t, c, "v-sh-udr-01")
		m.Body.(*sh.UDR).UserIdentity.PublicIdentity.Value = tc.value
		if e := c.Validate(m); (e == nil) != tc.pass {
			t.Fatal(tc.value, e)
		}
	}
}
func TestShOpaqueDataAndBudgets(t *testing.T) {
	c := shCodec(t, codec.Config{})
	// XML validation is explicitly NOT performed. This includes binary bytes,
	// malformed XML and external entity text, which is never evaluated/fetched.
	for _, data := range [][]byte{nil, {}, {0, 255, 254}, []byte("<broken"), []byte(`<!DOCTYPE x SYSTEM "file:///not-read"><Sh-Data/>`), []byte("<Sh-Data>用户&amp;data</Sh-Data>")} {
		req := decoded(t, c, "v-sh-pur-01")
		req.Body.(*sh.PUR).UserData = data
		raw := encoded(t, c, req)
		m, e := c.Decode(raw)
		if e != nil || !bytes.Equal(m.Body.(*sh.PUR).UserData, data) {
			t.Fatal(e)
		}
		if r := c.ValidateContext(m, validation.Context{}); !r.Complete {
			t.Fatal(r)
		}
		saved := append([]byte(nil), data...)
		clear(raw)
		if !bytes.Equal(m.Body.(*sh.PUR).UserData, saved) {
			t.Fatal("borrowed opaque input")
		}
	}
	req := decoded(t, c, "v-sh-pur-01")
	req.Body.(*sh.PUR).UserData = make([]byte, 513*1024)
	prefix := []byte{1, 2, 3}
	out, e := c.Append(prefix, req)
	if !errors.Is(e, wire.ErrLimit) || !bytes.Equal(out, prefix) {
		t.Fatal("failed append changed prefix", e)
	}
	uda := decoded(t, c, "v-sh-uda-01")
	uda.Body.(*sh.UDA).UserData = avp.Optional[[]byte]{}
	if e = c.Validate(uda); !errors.Is(e, validation.ErrMissing) {
		t.Fatal(e)
	}
	uda.Body.(*sh.UDA).UserData = avp.Some([]byte{})
	if e = c.Validate(uda); e != nil {
		t.Fatal("present empty is not absent", e)
	}
	for _, name := range []string{"v-sh-uda-02", "v-sh-sna-02"} {
		m := decoded(t, c, name)
		switch b := m.Body.(type) {
		case *sh.UDA:
			b.UserData = avp.Some([]byte("data"))
		case *sh.SNA:
			b.UserData = avp.Some([]byte("data"))
		}
		if e = c.Validate(m); !errors.Is(e, validation.ErrField) {
			t.Fatal(e)
		}
	}
}
func TestShOneTimeAndExpiryPresence(t *testing.T) {
	c := shCodec(t, codec.Config{})
	m := decoded(t, c, "v-sh-snr-01")
	r := m.Body.(*sh.SNR)
	if !r.OneTimeNotification.Present || r.OneTimeNotification.Value != 0 {
		t.Fatal("zero lost")
	}
	for _, v := range []avp.Time{0, 0xfffffff0, 0x10} {
		r.ExpiryTime = avp.Some(v)
		d, e := c.Decode(encoded(t, c, m))
		if e != nil || d.Body.(*sh.SNR).ExpiryTime != r.ExpiryTime {
			t.Fatal(e)
		}
	}
	a, e := c.CompleteAnswer(m, shResult(2001, false, 308), pdu.Origin{Host: "hss.example", Realm: "example.net"})
	if e != nil || a.Body.(*sh.SNA).ExpiryTime.Present {
		t.Fatal("expiry silently copied", e)
	}
	r.OneTimeNotification = avp.Optional[uint32]{}
	if e = c.Validate(m); !errors.Is(e, validation.ErrMissing) {
		t.Fatal(e)
	}
	r.OneTimeNotification = avp.Some(uint32(0))
	r.SubsReqType = 2
	if e = c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	r.SubsReqType = 0
	r.SendDataIndication = avp.Some(uint32(0))
	if e = c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
	r.SendDataIndication = avp.Optional[uint32]{}
	raw := rewriteAVP(t, encoded(t, c, m), []uint32{712}, func(v []byte) []byte { v[4] |= 64; return v })
	received, e := c.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.CompleteAnswer(received, shResult(2001, false, 308), pdu.Origin{Host: "hss.example", Realm: "example.net"}); e != nil {
		t.Fatal(e)
	}
}
func TestShIndependentWireNegatives(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, tc := range []struct {
		name, file string
		path       []uint32
		edit       func([]byte) []byte
	}{
		{"no VSAI", "v-sh-udr-01", []uint32{260}, func([]byte) []byte { return nil }},
		{"no VSAI vendor", "v-sh-uda-01", []uint32{260, 266}, func([]byte) []byte { return nil }},
		{"wrong auth app", "v-sh-pnr-01", []uint32{260, 258}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[8:], 16777312); return b }},
		{"duplicate VSAI", "v-sh-pua-01", []uint32{260}, func(b []byte) []byte { return append(b, b...) }},
		{"no public identity", "v-sh-udr-01", []uint32{700, 601}, func([]byte) []byte { return nil }},
		{"bad public UTF8", "v-sh-pur-01", []uint32{700, 601}, func(b []byte) []byte { b[12] = 255; return b }},
		{"no UDR ref", "v-sh-udr-01", []uint32{703}, func([]byte) []byte { return nil }},
		{"multi ref unsupported", "v-sh-udr-01", []uint32{703}, func(b []byte) []byte { return append(b, b...) }},
		{"no PUR data", "v-sh-pur-01", []uint32{702}, func([]byte) []byte { return nil }},
		{"no PNR data", "v-sh-pnr-01", []uint32{702}, func([]byte) []byte { return nil }},
		{"no PNR host", "v-sh-pnr-01", []uint32{293}, func([]byte) []byte { return nil }},
		{"no one time", "v-sh-snr-01", []uint32{712}, func([]byte) []byte { return nil }},
		{"expiry width3", "v-sh-sna-01", []uint32{709}, func(b []byte) []byte { wire.Put24(b[5:8], 15); return b }},
		{"duplicate result", "v-sh-pna-01", []uint32{268}, func(b []byte) []byte { return append(b, b...) }},
		{"wrong experimental vendor", "v-sh-pna-02", []uint32{297, 266}, func(b []byte) []byte { binary.BigEndian.PutUint32(b[8:], 0); return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, e := c.Decode(rewriteAVP(t, fixture(t, tc.file), tc.path, tc.edit)); e == nil {
				t.Fatal("accepted negative")
			}
		})
	}
}
func TestShCanonicalPreservesXMLAndPublicIdentity(t *testing.T) {
	c := shCodec(t, codec.Config{})
	req := decoded(t, c, "v-sh-pur-01")
	raw := encoded(t, c, req)
	before, e := c.Canonical(raw, 4096)
	if e != nil {
		t.Fatal(e)
	}
	req.Header.HopByHop++
	req.Header.Flags |= pdu.RetransmittedFlag
	after, e := c.Canonical(encoded(t, c, req), 4096)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("transport identity", e)
	}
	b := req.Body.(*sh.PUR)
	b.UserData = append(b.UserData, ' ')
	after, e = c.Canonical(encoded(t, c, req), 4096)
	if e != nil || bytes.Equal(before, after) {
		t.Fatal("XML whitespace collapsed", e)
	}
	req = decoded(t, c, "v-sh-pur-01")
	req.Body.(*sh.PUR).UserIdentity.PublicIdentity.Value = "sip:alice@EXAMPLE.net"
	after, e = c.Canonical(encoded(t, c, req), 4096)
	if e != nil || bytes.Equal(before, after) {
		t.Fatal("public URI folded", e)
	}
	profile, _ := sh.Profile()
	bundle, _ := schema.Compile(profile)
	for _, body := range []pdu.Body{&sh.UDR{}, &sh.PUR{}, &sh.SNR{}, &sh.PNR{}} {
		op, _ := bundle.Lookup(body.CommandKey())
		if op.NoContext || op.CheckContext == nil || op.TransportReplay || op.Canonical == nil {
			t.Fatal(op.Key)
		}
	}
}
