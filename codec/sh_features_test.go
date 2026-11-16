package codec_test

import (
	"bytes"
	"errors"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func sf(bits uint32, m bool) []sh.SupportedFeatures {
	return []sh.SupportedFeatures{{VendorID: 10415, FeatureListID: 1, FeatureList: bits, Mandatory: m}}
}
func evidence(t *testing.T, cmd, requester, responder uint32) validation.Context {
	t.Helper()
	c, e := sh.PeerFeatures(sh.PeerFeatureEvidence{CommandCode: cmd, Requester: pdu.Origin{Host: "as.example", Realm: "example.net"}, Responder: pdu.Origin{Host: "hss.example", Realm: "example.net"}, RequesterKnown: true, ResponderKnown: true, RequesterFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: requester}}, ResponderFeatures: []sh.FeatureSupport{{VendorID: 10415, FeatureListID: 1, FeatureList: responder}}})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestShFeature5011TriggerAndFullSet(t *testing.T) {
	c := shCodec(t, codec.Config{})
	origin := pdu.Origin{Host: "hss.example", Realm: "example.net"}
	src := pdu.AnswerSources{Terminals: []pdu.Origin{origin}}
	for _, tc := range []struct {
		name     string
		req, ans uint32
		m, omit  bool
		result   uint32
		valid    bool
	}{
		{"mandatory-none", 1, 0, true, false, 5011, true}, {"partial", 9, 1, true, false, 5011, true},
		{"all-supported", 9, 9, true, false, 2001, true}, {"offer", 9, 1, false, false, 2001, true},
		{"false-5011-offer", 9, 1, false, false, 5011, false}, {"false-5011-all", 9, 9, true, false, 5011, false},
		{"plain-5011", 0, 0, false, true, 5011, false}, {"missed-5011", 9, 1, true, false, 2001, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := decoded(t, c, "v-sh-udr-01")
			b := r.Body.(*sh.UDR)
			if !tc.omit {
				b.SupportedFeatures = sf(tc.req, tc.m)
			}
			ab := shResult(tc.result, tc.result == 5011, 306).(*sh.UDA)
			ab.SupportedFeatures = sf(tc.ans, false)
			a, e := c.CompleteAnswer(r, ab, origin)
			if !tc.valid {
				if e == nil {
					e = c.ValidateAnswer(r, a, src)
				}
				if e == nil {
					t.Fatal("invalid trigger accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			raw := encoded(t, c, a)
			a, e = c.DecodeAnswer(r, raw, src)
			if e != nil {
				t.Fatal(e)
			}
			if report := c.ValidateAnswerContext(r, a, src, validation.Context{}); report.State != validation.NeedsApplication {
				t.Fatal(report)
			}
			if report := c.ValidateAnswerContext(r, a, src, evidence(t, 306, tc.req, tc.ans)); !report.Complete {
				t.Fatal(report)
			}
			if report := c.ValidateAnswerContext(r, a, src, evidence(t, 306, tc.req, tc.ans|16)); report.Complete {
				t.Fatal("intersection mistaken for full set")
			}
			// Exact zero bitmaps survive decoding and are not treated as missing children.
			if a.Body.(*sh.UDA).SupportedFeatures[0].FeatureList != tc.ans {
				t.Fatal("bitmap lost")
			}
		})
	}
}
func TestShMixedReferencesAndMandatoryCanonical(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, refs := range [][]uint32{{11, 12}, {11, 25}, {12, 25}, {11, 12, 25}} {
		for _, action := range []uint32{0, 1} {
			r := decoded(t, c, "d21-sub11-r")
			b := r.Body.(*sh.SNR)
			b.DataReference = refs
			b.SubsReqType = action
			b.SupportedFeatures = sf(1, true)
			b.SendDataIndication = avp.Optional[uint32]{}
			if refs[len(refs)-1] == 25 {
				b.OneTimeNotification = avp.Some(uint32(0))
			}
			raw := encoded(t, c, r)
			decoded, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			if report := c.ValidateContext(decoded, validation.Context{}); report.State != validation.NeedsApplication {
				t.Fatal(report)
			}
			if report := c.ValidateContext(decoded, evidence(t, 308, 1, 1)); !report.Complete {
				t.Fatal(report)
			}
			orig, e := c.Canonical(raw, 8192)
			if e != nil {
				t.Fatal(e)
			}
			// Change only outer M in actual wire: tolerated receive but significant to
			// original-identity conflict detection and required-feature failure semantics.
			for i := 20; i < len(raw); {
				n := int(raw[i+5])<<16 | int(raw[i+6])<<8 | int(raw[i+7])
				if bytes.Equal(raw[i:i+4], []byte{0, 0, 2, 116}) {
					raw[i+4] &^= 64
				}
				i += (n + 3) &^ 3
			}
			m0, e := c.Decode(raw)
			if e != nil {
				t.Fatal(e)
			}
			changed, e := c.Canonical(raw, 8192)
			if e != nil || bytes.Equal(orig, changed) {
				t.Fatal("M erased", e)
			}
			if e = c.Validate(m0); e == nil {
				t.Fatal("actual use sent with M0")
			}
			if refs[len(refs)-1] == 25 {
				b.SendDataIndication = avp.Some(uint32(0))
				if e = c.Validate(r); e == nil {
					t.Fatal("SendData0 accepted")
				}
				b.SendDataIndication = avp.Optional[uint32]{}
			}
			b.DataReference = []uint32{11, 11}
			if e = c.Validate(r); !errors.Is(e, validation.ErrDuplicate) {
				t.Fatal(e)
			}
		}
	}
}
func TestShFeatureEvidenceBoundsAndBinding(t *testing.T) {
	c := shCodec(t, codec.Config{})
	r := decoded(t, c, "v-sh-udr-01")
	r.Body.(*sh.UDR).SupportedFeatures = sf(1, true)
	ctx := evidence(t, 306, 1, 0)
	if report := c.ValidateContext(r, ctx); !report.Complete {
		t.Fatal(report)
	} // received required feature may be refused by 5011
	r.Body.(*sh.UDR).DestinationHost = avp.Some("dra.example")
	if report := c.ValidateContext(r, ctx); !errors.Is(report.Err(), validation.ErrContext) {
		t.Fatal(report)
	}
	r.Body.(*sh.UDR).DestinationHost = avp.Optional[string]{}
	if report := c.ValidateContext(r, evidence(t, 308, 1, 0)); !errors.Is(report.Err(), validation.ErrContext) {
		t.Fatal(report)
	}
	b := r.Body.(*sh.UDR)
	b.SupportedFeatures = append(sf(1, true), sf(0, false)...)
	if e := c.Validate(r); !errors.Is(e, validation.ErrDuplicate) {
		t.Fatal(e)
	}
	for _, tc := range []sh.PeerFeatureEvidence{{}, {CommandCode: 305}, {CommandCode: 306, Requester: pdu.Origin{Host: "as.example", Realm: "example.net"}, Responder: pdu.Origin{Host: "hss.example", Realm: "example.net"}, RequesterFeatures: []sh.FeatureSupport{{}}}} {
		if _, e := sh.PeerFeatures(tc); !errors.Is(e, validation.ErrContext) {
			t.Fatal(e)
		}
	}
}

func TestShFeatureFrozenPairs(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, set := range [][]string{
		{"v25-sh-need-r", "v25-sh-need-a", "f25-sh-5011", "f25-sh-base5001"},
		{"b20-sh-partial-r", "b20-sh-partial-a", "b20-sh-partial-f"},
		{"v25-sh-offer-r", "v25-sh-offer-a"}, {"v25-sh-plain-r", "v25-sh-plain-a"},
		{"v25-mix-snr-r", "v25-mix-snr-a", "f25-mix-snr"}, {"v25-mix-pnr-r", "v25-mix-pnr-a", "f25-mix-pnr"},
	} {
		t.Run(set[0], func(t *testing.T) {
			req := decoded(t, c, set[0])
			origin := pdu.Origin{Host: "hss.example", Realm: "example.net"}
			if req.Header.CommandCode == 309 {
				origin.Host = "as.example"
			}
			for i, name := range set {
				raw := fixture(t, name)
				m, e := c.Decode(raw)
				if e != nil {
					t.Fatal(name, e)
				}
				encodedRaw := encoded(t, c, m)
				again, e := c.Decode(encodedRaw)
				if e != nil || !bytes.Equal(encodedRaw, encoded(t, c, again)) {
					t.Fatal(name, "unstable reencoding", e)
				}
				if i == 0 {
					a, e := c.Canonical(raw, 8192)
					if e != nil {
						t.Fatal(e)
					}
					b, e := c.Canonical(encodedRaw, 8192)
					if e != nil || !bytes.Equal(a, b) {
						t.Fatal("canonical changed", e)
					}
				}
				if i > 0 {
					if e = c.ValidateAnswer(req, m, pdu.AnswerSources{Terminals: []pdu.Origin{origin}}); e != nil {
						t.Fatal(name, e)
					}
				}
			}
		})
	}
}

func TestShMultiPullRemainsOutsideSlice(t *testing.T) {
	c := shCodec(t, codec.Config{})
	if _, e := c.Decode(fixture(t, "v25-sh-partial-r")); !errors.Is(e, validation.ErrDuplicate) {
		t.Fatal(e)
	}
}

func TestShFeatureParentChildrenAndUninterpretedExtension(t *testing.T) {
	c := shCodec(t, codec.Config{})
	// Independent TLV mutation of the frozen M1 group; do not use our encoder
	// to manufacture malformed child presence or widths.
	original := fixture(t, "v25-sh-need-r")
	at := 20
	for {
		n := int(original[at+5])<<16 | int(original[at+6])<<8 | int(original[at+7])
		if bytes.Equal(original[at:at+4], []byte{0, 0, 2, 116}) {
			break
		}
		at += (n + 3) &^ 3
	}
	for _, kind := range []string{"missing", "short", "duplicate"} {
		b := append([]byte(nil), original...)
		end := at + 56
		child := at + 40
		switch kind {
		case "missing":
			b = append(b[:child], b[end:]...)
			b[at+7] = 40
		case "short":
			b[child+7] = 15 // leave padding, but payload width is three
		case "duplicate":
			copyChild := append([]byte(nil), b[child:end]...)
			b = append(append(append([]byte(nil), b[:end]...), copyChild...), b[end:]...)
			b[at+7] = 72
		}
		b[1], b[2], b[3] = byte(len(b)>>16), byte(len(b)>>8), byte(len(b))
		if _, e := c.Decode(b); e == nil {
			t.Fatal(kind, "accepted")
		}
	}
	r := decoded(t, c, "v25-sh-need-r")
	b := r.Body.(*sh.SNR)
	b.SupportedFeatures[0].Extra = []avp.Raw{{Key: avp.Key{VendorID: 99999, Code: 1}, Flags: 128, Data: []byte{1}}}
	if e := c.Validate(r); e != nil {
		t.Fatal(e)
	}
	if report := c.ValidateContext(r, evidence(t, 308, 1, 1)); report.State != validation.NeedsApplication {
		t.Fatal("unknown feature child certified", report)
	}
	b.SupportedFeatures[0].Extra[0].Flags |= 64
	if e := c.Validate(r); !errors.Is(e, validation.ErrUnsupported) {
		t.Fatal(e)
	}
}
