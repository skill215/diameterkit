package codec_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
)

func TestShSubscriptionFrozenPairs(t *testing.T) {
	c := shCodec(t, codec.Config{})
	src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "hss.example", Realm: "example.net"}}}
	for _, name := range []string{"d21-sub11", "b19-sub12", "b19-unsub25", "b19-unsub11", "b19-unsub12"} {
		t.Run(name, func(t *testing.T) {
			req := decoded(t, c, name+"-r")
			for _, suffix := range []string{"-r", "-a", "-f"} {
				raw := fixture(t, name+suffix)
				m := decoded(t, c, name+suffix)
				if !bytes.Equal(encoded(t, c, m), raw) {
					t.Fatal("frozen reencode")
				}
				if suffix != "-r" {
					if r := c.ValidateAnswerContext(req, m, src, validation.Context{}); !r.Complete {
						t.Fatal(r)
					}
				}
				if a, ok := m.Body.(*sh.SNA); ok && a.UserData.Present {
					d, e := shdata.Parse(a.UserData.Value, shdata.DefaultLimits())
					if e != nil {
						t.Fatal(e)
					}
					use := shdata.IMSUserStateData
					if req.Body.(*sh.SNR).DataReference[0] == 12 {
						use = shdata.SCSCFNameData
					}
					if r, e := d.Check(use); e != nil || !r.Complete {
						t.Fatal(r, e)
					}
				}
			}
		})
	}
}
func TestShSubscriptionPresenceMatrix(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, ref := range []uint32{11, 12, 25} {
		for _, action := range []uint32{0, 1} {
			for _, send := range []avp.Optional[uint32]{{}, avp.Some(uint32(0)), avp.Some(uint32(1)), avp.Some(uint32(2))} {
				m := decoded(t, c, "d21-sub11-r")
				b := m.Body.(*sh.SNR)
				b.DataReference = []uint32{ref}
				b.SubsReqType = action
				b.SendDataIndication = send
				if ref == 25 {
					b.OneTimeNotification = avp.Some(uint32(0))
				}
				valid := !send.Present || ref != 25 && send.Value <= 1
				e := c.Validate(m)
				if (e == nil) != valid {
					t.Fatalf("ref%d action%d send%+v: %v", ref, action, send, e)
				}
				if !valid {
					continue
				}
				// Expiry is an absolute raw Time for both actions, never a local timer or
				// automatic unsubscribe flag. Explicit fields are not silently cleared.
				b.ExpiryTime = avp.Some(avp.Time(0))
				round, e := c.Decode(encoded(t, c, m))
				if e != nil {
					t.Fatal(e)
				}
				r := round.Body.(*sh.SNR)
				if r.SendDataIndication != send || !r.ExpiryTime.Present || r.SubsReqType != action {
					t.Fatal(r)
				}
			}
		}
	}
	for _, ref := range []uint32{11, 12} {
		m := decoded(t, c, "d21-sub11-r")
		b := m.Body.(*sh.SNR)
		b.DataReference = []uint32{ref}
		b.OneTimeNotification = avp.Some(uint32(0))
		if e := c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
		b.OneTimeNotification = avp.Optional[uint32]{}
		b.UserIdentity = sh.UserIdentity{MSISDN: avp.Some([]byte{0x21, 0xf3})}
		if e := c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(e)
		}
	}
	for _, v := range []uint32{0, 24, 26} {
		m := decoded(t, c, "d21-sub11-r")
		m.Body.(*sh.SNR).DataReference = []uint32{v}
		if e := c.Validate(m); !errors.Is(e, validation.ErrUnsupported) {
			t.Fatal(v, e)
		}
	}
	// M0 feature data and duplicate references cannot enable mixed subscriptions.
	m := decoded(t, c, "d21-sub11-r")
	m.Body.(*sh.SNR).Extra = []avp.Raw{{Key: avp.Key{VendorID: 10415, Code: 628}, Flags: 128}}
	if e := c.Validate(m); e == nil {
		t.Fatal("features silently enabled")
	}
}
func TestShSubscriptionResultAndAssociation(t *testing.T) {
	c := shCodec(t, codec.Config{})
	origin := pdu.Origin{Host: "hss.example", Realm: "example.net"}
	src := pdu.AnswerSources{Terminals: []pdu.Origin{origin}}
	for _, name := range []string{"d21-sub11", "b19-sub12", "b19-unsub25", "b19-unsub11", "b19-unsub12"} {
		req := decoded(t, c, name+"-r")
		for _, exp := range []uint32{5001, 5101, 5104} {
			a, e := c.CompleteAnswer(req, shResult(exp, true, 308), origin)
			if e != nil {
				t.Fatal(e)
			}
			if e = c.ValidateAnswer(req, a, src); e != nil {
				t.Fatal(e)
			}
		}
		for _, exp := range []uint32{5002, 5105, 5106} {
			if _, e := c.CompleteAnswer(req, shResult(exp, true, 308), origin); !errors.Is(e, validation.ErrUnsupported) {
				t.Fatal(exp, e)
			}
		}
		for _, withData := range []bool{false, true} {
			b := shResult(2001, false, 308).(*sh.SNA)
			if withData {
				b.UserData = avp.Some([]byte("<Sh-Data/>"))
			}
			a, e := c.CompleteAnswer(req, b, origin)
			if e != nil {
				t.Fatal(e)
			}
			if a.Body.(*sh.SNA).ExpiryTime.Present {
				t.Fatal("expiry filled in")
			}
			if e = c.ValidateAnswer(req, a, src); e != nil {
				t.Fatal(e)
			}
			// Optional User-Data remains opaque and is not promoted to a MUST even for
			// SendData=1. The application separately interprets its actual data/reference.
			a.Header.HopByHop++
			if e = c.ValidateAnswer(req, a, src); e == nil {
				t.Fatal("wrong request association")
			}
		}
		a, e := c.CompleteAnswer(req, &pdu.ErrorAnswer{ResultCode: 3004}, origin)
		if e != nil {
			t.Fatal(e)
		}
		if e = c.ValidateAnswer(req, a, src); e != nil {
			t.Fatal(e)
		}
	}
}
func TestShSubscriptionIndependentWireNegatives(t *testing.T) {
	c := shCodec(t, codec.Config{})
	raw := fixture(t, "d21-sub11-r")
	for _, tc := range []struct {
		name   string
		path   []uint32
		mutate func([]byte) []byte
	}{
		{"send-short", []uint32{710}, func(v []byte) []byte { return v[:15] }},
		{"send-value", []uint32{710}, func(v []byte) []byte { binary.BigEndian.PutUint32(v[12:], 2); return v }},
		{"duplicate-send", []uint32{710}, func(v []byte) []byte { return append(v, v...) }},
		{"duplicate-reference", []uint32{703}, func(v []byte) []byte { return append(v, v...) }},
		{"action-outside", []uint32{705}, func(v []byte) []byte { binary.BigEndian.PutUint32(v[12:], 2); return v }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := rewriteAVP(t, raw, tc.path, tc.mutate)
			if _, e := c.Decode(bad); e == nil {
				t.Fatal("accepted malformed request")
			}
		})
	}
	receiveM1 := rewriteAVP(t, raw, []uint32{710}, func(v []byte) []byte { v[4] |= 64; return v })
	if _, e := c.Decode(receiveM1); e != nil {
		t.Fatal("known M bit receive tolerance", e)
	}
	r := decoded(t, c, "d21-sub11-r")
	a, e := c.Canonical(encoded(t, c, r), 4096)
	if e != nil {
		t.Fatal(e)
	}
	r.Body.(*sh.SNR).SendDataIndication = avp.Some(uint32(0))
	b, e := c.Canonical(encoded(t, c, r), 4096)
	if e != nil || bytes.Equal(a, b) {
		t.Fatal("lost SendData value", e)
	}
	r.Body.(*sh.SNR).SendDataIndication = avp.Optional[uint32]{}
	z, e := c.Canonical(encoded(t, c, r), 4096)
	if e != nil || bytes.Equal(z, b) {
		t.Fatal("absent equals zero", e)
	}
	p, e := sh.Profile()
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := schema.Compile(p)
	if e != nil {
		t.Fatal(e)
	}
	op, ok := bundle.Lookup(r.Body.CommandKey())
	if !ok {
		t.Fatal("missing command")
	}
	if op.TransportReplay || op.NoContext || op.CheckContext == nil {
		t.Fatal("scope/replay audit changed")
	}
}
