package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func TestSimulatorInputDecisions(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, tc := range []struct {
		name string
		want error
	}{
		{"mme-ofr-original", wire.ErrFrame}, {"mme-ofr-header-fixed", validation.ErrField}, {"mme-ofr-host-invalid", avp.ErrValue}, {"mme-alr-original", wire.ErrFrame}, {"mme-ofr-host-fixed", validation.ErrField}, {"mme-alr-both-identities", validation.ErrField},
		{"mme-sra-5012", validation.ErrField}, {"poc-ofa-state0", validation.ErrField}, {"sh-uda-2001", validation.ErrMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := c.Decode(fixture(t, "b77-sim-"+tc.name))
			if !errors.Is(e, tc.want) {
				t.Fatal(e, "want", tc.want)
			}
		})
	}
	for _, n := range []string{"s6a-316", "s6a-318"} {
		t.Run(n, func(t *testing.T) {
			if _, e := c.Decode(fixture(t, "b77-sim-"+n)); e == nil {
				t.Fatal("deferred S6a admitted")
			}
		})
	}
}
func TestSimulatorAcceptedPairs(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, tc := range []struct{ request, answer, host string }{
		{"v-srr-01", "mme-sra-2001", "hss.example"}, {"v-srr-01", "mme-sra-failure-fixed", "hss.example"},
		{"v-tfr-01", "mme-tfa-2001", "mme.example"}, {"v-tfr-01", "mme-tfa-5012", "mme.example"},
		{"v-rdr-01", "mme-rda-2001", "hss.example"}, {"v-rdr-01", "mme-rda-5012", "hss.example"},
		{"v-ofr-01", "poc-ofa-state1", "iwmsc.example"},
		{"v-sh-udr-01", "sh-uda-5012", "hss.example"}, {"v-sh-udr-01", "sh-uda-fixed", "hss.example"},
		{"v-sh-pur-01", "sh-pua-2001", "hss.example"}, {"v-sh-pur-01", "sh-pua-5012", "hss.example"},
		{"v-sh-snr-01", "sh-sna-2001", "hss.example"}, {"v-sh-snr-01", "sh-sna-5012", "hss.example"},
		{"v-rdr-01", "sh-rda-vsai", "hss.example"},
	} {
		t.Run(tc.answer, func(t *testing.T) {
			r, a := decoded(t, c, tc.request), decoded(t, c, "b77-sim-"+tc.answer)
			src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: tc.host, Realm: "example.net"}}}
			if report := c.ValidateAnswerContext(r, a, src, validation.Context{}); !report.Complete {
				t.Fatal(report)
			}
			if tc.answer == "sh-rda-vsai" {
				if e := c.Validate(a); e == nil {
					t.Fatal("SMS ignored VSAI retransmitted")
				}
				return
			}
			raw := encoded(t, c, a)
			owned, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			clear(raw)
			if !bytes.Equal(encoded(t, c, a), encoded(t, c, owned)) {
				t.Fatal("borrowed answer")
			}
			a.Header.EndToEnd++
			if e := c.ValidateAnswer(r, a, src); !errors.Is(e, validation.ErrUnmatchedAnswer) {
				t.Fatal(e)
			}
		})
	}
}
func TestSimulatorRequestMigration(t *testing.T) {
	c := shCodec(t, codec.Config{})
	r := decoded(t, c, "b77-sim-mme-ofr-fixed")
	b := r.Body.(*sgd.OFR)
	if b.UserIdentifier.UserName.Value != "460310000000000" {
		t.Fatal("unexpected configured identity")
	}
	sc, e := avp.DecodeTBCD(b.SCAddress)
	if e != nil || sc != "138888888" {
		t.Fatal(sc, e)
	}
	if !bytes.Equal(b.SMRPUI, []byte{0, 255, 0x30, 0x42, 0x30, 0x44}) {
		t.Fatal("TPDU reinterpreted")
	}
	if v := c.ValidateContext(r, validation.Context{}); !v.Complete {
		t.Fatal(v)
	}
	r = decoded(t, c, "b77-sim-mme-alr-fixed")
	a := r.Body.(*s6c.ALR)
	if v := c.ValidateContext(r, validation.Context{}); v.State != validation.NeedsApplication {
		t.Fatal(v)
	}
	ctx, e := s6c.AlertRole(s6c.AlertRoleEvidence{Role: s6c.AlertFromMME, Requester: pdu.Origin{Host: a.OriginHost, Realm: a.OriginRealm}, Responder: pdu.Origin{Host: a.DestinationHost, Realm: a.DestinationRealm}})
	if e != nil {
		t.Fatal(e)
	}
	if v := c.ValidateContext(r, ctx); !v.Complete {
		t.Fatal(v)
	}
	// The library does not silently remove '+' or infer an international number.
	if _, e := avp.EncodeTBCD("+138888888"); e == nil {
		t.Fatal("implicit number rewriting")
	}
}
func TestSimulatorOriginalEnvelopeRepair(t *testing.T) {
	c := shCodec(t, codec.Config{})
	r := decoded(t, c, "v-sh-pur-01")
	r.Body.(*sh.PUR).ProxyInfo = []pdu.ProxyInfo{{ProxyHost: "dra.example", ProxyState: []byte{0, 255}}}
	a := decoded(t, c, "b77-sim-sh-pua-no-proxy")
	src := pdu.AnswerSources{Terminals: []pdu.Origin{{Host: "hss.example", Realm: "example.net"}}}
	if e := c.ValidateAnswer(r, a, src); e == nil {
		t.Fatal("lost proxy accepted")
	}
	corrected, e := c.CompleteAnswer(r, &sh.PUA{AuthSessionState: 1, VendorSpecificApplicationID: sh.Application(), ResultCode: avp.Some(uint32(2001))}, src.Terminals[0])
	if e != nil {
		t.Fatal(e)
	}
	if v := c.ValidateAnswerContext(r, corrected, src, validation.Context{}); !v.Complete {
		t.Fatal(v)
	}
	if len(corrected.Body.(*sh.PUA).ProxyInfo) != 1 {
		t.Fatal("proxy not recovered from original")
	}
	// Old destination-copy only works when it is the authorized responder.
	src.Terminals[0].Host = "actual-hss.example"
	if e := c.ValidateAnswer(decoded(t, c, "v-srr-01"), decoded(t, c, "b77-sim-mme-sra-2001"), src); !errors.Is(e, validation.ErrSource) {
		t.Fatal(e)
	}
}
