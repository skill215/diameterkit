// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Redirect E grammar, independent wire, ownership, limits and source validation.
// SPDX-AI-Date: 2026-11-17

package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"strings"
	"testing"
)

func redirectU32(n uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, n); return b }
func redirectWire(t testing.TB, hosts []string, usage, ttl avp.Optional[uint32]) []byte {
	t.Helper()
	b := fixture(t, "b2-sra-e-proxy")
	// Independent wire edit: no ErrorAnswer encoder constructs the redirect AVPs.
	d := wire.NewDecoder(b[20:], wire.DefaultLimits())
	for {
		v, ok, e := d.Next()
		if e != nil || !ok {
			t.Fatal("missing result", e)
		}
		if v.Key == (avp.Key{Code: 268}) {
			copy(b[20+v.Offset+8:], redirectU32(3006))
			break
		}
	}
	appendAVP := func(code uint32, data []byte) {
		part := make([]byte, (8+len(data)+3)&^3)
		binary.BigEndian.PutUint32(part, code)
		part[4] = 0x40
		wire.Put24(part[5:8], uint32(8+len(data)))
		copy(part[8:], data)
		b = append(b, part...)
	}
	for _, host := range hosts {
		appendAVP(292, []byte(host))
	}
	if usage.Present {
		appendAVP(261, redirectU32(usage.Value))
	}
	if ttl.Present {
		appendAVP(262, redirectU32(ttl.Value))
	}
	wire.Put24(b[1:4], uint32(len(b)))
	return b
}

func TestRedirectWireAndOwnership(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "b2-srr-proxy")
	for usage := -1; usage <= 6; usage++ {
		for _, ttl := range []uint32{0, 60, ^uint32(0)} {
			t.Run(fmt.Sprintf("usage-%d/ttl-%d", usage, ttl), func(t *testing.T) {
				u := avp.Optional[uint32]{}
				if usage >= 0 {
					u = avp.Some(uint32(usage))
				}
				hosts := []string{"aaa://first.example:3868;transport=tcp", "AAAS://Second.Example;transport=sctp;protocol=diameter"}
				raw := redirectWire(t, hosts, u, avp.Some(ttl))
				m, e := c.DecodeAnswer(req, raw, sources())
				if e != nil {
					t.Fatal(e)
				}
				clear(raw)
				b := m.Body.(*pdu.ErrorAnswer)
				if b.ResultCode != 3006 || len(b.RedirectHost) != 2 || b.RedirectHost[1] != hosts[1] || b.RedirectHostUsage != u || b.RedirectMaxCacheTime != avp.Some(ttl) {
					t.Fatal("lost redirect fields", b)
				}
				clone, e := c.Snapshot(m)
				if e != nil {
					t.Fatal(e)
				}
				b.RedirectHost[0] = "aaa://changed.example"
				if clone.Body.(*pdu.ErrorAnswer).RedirectHost[0] != hosts[0] {
					t.Fatal("aliased host slice")
				}
				wireBytes := encoded(t, c, clone)
				if _, e = c.DecodeAnswer(req, wireBytes, sources()); e != nil {
					t.Fatal(e)
				}
			})
		}
	}
	for _, u := range []avp.Optional[uint32]{{}, avp.Some(uint32(0))} {
		raw := redirectWire(t, []string{"aaa://host.example"}, u, avp.Optional[uint32]{})
		if _, e := c.DecodeAnswer(req, raw, sources()); e != nil {
			t.Fatal("default no-cache", e)
		}
	}
	proto := &pdu.ErrorAnswer{ResultCode: 3006, RedirectHost: []string{"aaa://owned.example"}, RedirectHostUsage: avp.Optional[uint32]{Value: 99}, RedirectMaxCacheTime: avp.Optional[uint32]{Value: 99}}
	m, e := c.CompleteAnswer(req, proto, pdu.Origin{Host: "hss.example", Realm: "example.net"})
	if e != nil {
		t.Fatal(e)
	}
	proto.RedirectHost[0] = "aaa://changed.example"
	b := m.Body.(*pdu.ErrorAnswer)
	if b.RedirectHost[0] != "aaa://owned.example" || b.RedirectHostUsage.Value != 0 || b.RedirectMaxCacheTime.Value != 0 {
		t.Fatal("clone or absent option storage")
	}
}

func TestRedirectInvalidFields(t *testing.T) {
	c := newCodec(t, codec.Config{})
	req := decoded(t, c, "b2-srr-proxy")
	for _, kind := range []string{"hosts-missing", "usage-unknown", "ttl-missing", "uri", "host-long", "duplicate-usage", "duplicate-ttl", "usage-width", "ttl-width", "wrong-result", "missing-E", "vendor-host", "unknown-M", "source", "session", "proxy"} {
		t.Run(kind, func(t *testing.T) {
			hosts := []string{"aaa://host.example"}
			u := avp.Some(uint32(1))
			ttl := avp.Some(uint32(30))
			switch kind {
			case "hosts-missing":
				hosts = nil
			case "usage-unknown":
				u.Value = 7
			case "ttl-missing":
				ttl.Present = false
			case "uri":
				hosts[0] = "https://host.example"
			case "host-long":
				hosts[0] = "aaa://" + strings.Repeat("a", avp.MaxDiameterURIBytes)
			}
			raw := redirectWire(t, hosts, u, ttl)
			switch kind {
			case "duplicate-usage":
				raw = append(raw, rawAVP(261, redirectU32(1))...)
			case "duplicate-ttl":
				raw = append(raw, rawAVP(262, redirectU32(30))...)
			case "usage-width":
				raw = replaceTop(t, raw, 261, rawAVP(261, []byte{1}))
			case "ttl-width":
				raw = replaceTop(t, raw, 262, rawAVP(262, []byte{1}))
			case "wrong-result":
				raw = replaceTop(t, raw, 268, rawAVP(268, redirectU32(5012)))
			case "missing-E":
				raw[4] &^= pdu.ErrorFlag
			case "vendor-host":
				part := make([]byte, 16)
				binary.BigEndian.PutUint32(part, 292)
				part[4] = 0xc0
				wire.Put24(part[5:8], 16)
				binary.BigEndian.PutUint32(part[8:], 10415)
				raw = replaceTop(t, raw, 292, part)
			case "unknown-M":
				raw = append(raw, rawAVP(999999, []byte{1})...)
			case "source":
				raw = replaceTop(t, raw, 264, rawAVP(264, []byte("evil.example")))
			case "session":
				raw = replaceTop(t, raw, 263, rawAVP(263, []byte("other-session")))
			case "proxy":
				raw = replaceTop(t, raw, 284, rawAVP(284, rawAVP(280, []byte("wrong.example"))))
			}
			wire.Put24(raw[1:4], uint32(len(raw)))
			m, e := c.DecodeAnswer(req, raw, sources())
			if !errors.Is(e, validation.ErrInvalidAnswer) || m.Body != nil {
				t.Fatal("invalid redirect accepted", e)
			}
		})
	}
	for _, code := range []uint32{292, 261, 262} {
		proto := &pdu.ErrorAnswer{ResultCode: 3006, RedirectHost: []string{"aaa://host.example"}, Extra: []avp.Raw{{Key: avp.Key{Code: code}, Data: []byte{1}}}}
		if _, e := c.CompleteAnswer(req, proto, pdu.Origin{Host: "hss.example", Realm: "example.net"}); e == nil {
			t.Fatal("raw field shadow", code)
		}
	}
}

func TestRedirectBounds(t *testing.T) {
	for _, kind := range []string{"repeat", "field", "count", "frame", "fixed-host-cap"} {
		t.Run(kind, func(t *testing.T) {
			limits := wire.DefaultLimits()
			hosts := []string{"aaa://host.example", "aaas://other.example"}
			switch kind {
			case "repeat":
				limits.MaxRepeated = 1
			case "field":
				limits.MaxFieldBytes = 16
			case "count":
				limits.MaxAVPs = 6
				limits.MaxRepeated = 6
			case "frame":
				limits.MaxFrameBytes = 160
				limits.MaxFieldBytes = 160
			case "fixed-host-cap":
				hosts = make([]string, base.MaxRedirectHosts+1)
				for i := range hosts {
					hosts[i] = "aaa://host.example"
				}
				limits.MaxRepeated = base.MaxRedirectHosts + 1
			}
			c := newCodec(t, codec.Config{Limits: limits})
			raw := redirectWire(t, hosts, avp.Optional[uint32]{}, avp.Optional[uint32]{})
			m, e := c.Decode(raw)
			if !errors.Is(e, wire.ErrLimit) || m.Body != nil {
				t.Fatal("decode limit", e)
			}
			unlimited := newCodec(t, codec.Config{})
			req := decoded(t, unlimited, "b2-srr-proxy")
			proto := &pdu.ErrorAnswer{ResultCode: 3006, RedirectHost: hosts}
			if _, e = c.CompleteAnswer(req, proto, pdu.Origin{Host: "hss.example", Realm: "example.net"}); !errors.Is(e, wire.ErrLimit) {
				t.Fatal("build limit", e)
			}
		})
	}
	// Exact fixed cap remains valid when the caller's wire budget admits it.
	hosts := make([]string, base.MaxRedirectHosts)
	for i := range hosts {
		hosts[i] = "aaa://host.example"
	}
	c := newCodec(t, codec.Config{})
	raw := redirectWire(t, hosts, avp.Optional[uint32]{}, avp.Optional[uint32]{})
	m, e := c.Decode(raw)
	if e != nil || len(m.Body.(*pdu.ErrorAnswer).RedirectHost) != len(hosts) {
		t.Fatal(e)
	}
	encodedBytes := encoded(t, c, m)
	if !bytes.Contains(encodedBytes, []byte(hosts[0])) {
		t.Fatal("lost hosts")
	}
}
