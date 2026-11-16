// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/sh/shdata"
)

func TestIMSStatePresenceAndValues(t *testing.T) {
	l := shdata.DefaultLimits()
	for state := uint8(0); state <= 3; state++ {
		raw, e := shdata.MarshalIMSUserState(state, l)
		if e != nil {
			t.Fatal(e)
		}
		d := parse(t, raw)
		check(t, d, shdata.IMSUserStateData, true)
		if !d.Data().IMSUserState.Present || d.Data().IMSUserState.Value != state {
			t.Fatal(d.Data())
		}
		check(t, d, shdata.ReachabilityNotification, false)
		check(t, d, shdata.SMSPull, false)
		if state == 1 && !bytes.Equal(raw, fixture(t, "x-sh-cont-01")) {
			t.Fatal("frozen state XML")
		}
	}
	for _, state := range []uint8{4, 255} {
		if _, e := shdata.MarshalIMSUserState(state, l); !errors.Is(e, shdata.ErrStructure) {
			t.Fatal(e)
		}
	}
	base := string(fixture(t, "x-sh-cont-01"))
	for _, value := range []string{"4", "-1", "256", "", "1.0", "１", "+ 1"} {
		if _, e := shdata.Parse([]byte(strings.Replace(base, ">1<", ">"+value+"<", 1)), l); !errors.Is(e, shdata.ErrStructure) {
			t.Fatal(value, e)
		}
	}
	for _, value := range []string{" +001 ", "03", "-0"} {
		d := parse(t, []byte(strings.Replace(base, ">1<", ">"+value+"<", 1)))
		check(t, d, shdata.IMSUserStateData, true)
	}
	check(t, parse(t, []byte("<Sh-Data/>")), shdata.IMSUserStateData, false)
}
func TestSCSCFNamePresenceRepresentationAndBudget(t *testing.T) {
	l := shdata.DefaultLimits()
	for _, name := range []string{"sip:scscf.example.net", "sip:SCscf.example.net", ""} {
		raw, e := shdata.MarshalSCSCFName(name, l)
		if e != nil {
			t.Fatal(e)
		}
		d := parse(t, raw)
		check(t, d, shdata.SCSCFNameData, true)
		if !d.Data().SCSCFName.Present || d.Data().SCSCFName.Value != name {
			t.Fatal(d.Data())
		}
		if name == "sip:scscf.example.net" && !bytes.Equal(raw, fixture(t, "b19-scscf-12")) {
			t.Fatal("frozen name")
		}
		if name == "" && !bytes.Equal(raw, fixture(t, "b19-scscf-empty")) {
			t.Fatal("empty name")
		}
	}
	base := string(fixture(t, "b19-scscf-12"))
	for _, name := range []string{"sips:scscf.example.net", "sip:scscf.example.net:5060", "sip:user@scscf.example.net", "sip:[::1]", "sip:域.example.net", "sip:192.0.2.1", " sip:scscf.example.net ", "A&amp;B"} {
		d := parse(t, []byte(strings.Replace(base, "sip:scscf.example.net", name, 1)))
		check(t, d, shdata.SCSCFNameData, false)
		if _, e := shdata.MarshalSCSCFName(name, l); !errors.Is(e, shdata.ErrUnsupported) {
			t.Fatal(name, e)
		}
	}
	check(t, parse(t, []byte("<Sh-Data/>")), shdata.SCSCFNameData, false)
	l.MaxTextBytes = 3
	if b, e := shdata.MarshalSCSCFName("sip:scscf.example.net", l); b != nil || !errors.Is(e, shdata.ErrLimit) {
		t.Fatal(e)
	}
	l = shdata.DefaultLimits()
	l.MaxDepth = 2
	if _, e := shdata.MarshalIMSUserState(0, l); !errors.Is(e, shdata.ErrLimit) {
		t.Fatal(e)
	}
}
func TestIMSPathsAndSingleReferenceScope(t *testing.T) {
	raw := fixture(t, "x-sh-multi-01")
	d := parse(t, raw)
	if !d.Data().SCSCFName.Present || !d.Data().IMSUserState.Present || len(d.Fragments()) != 0 {
		t.Fatal(d.Data())
	}
	// Recognizing both leaves does not authorize a mixed subscription or claim a
	// document contains only the independently selected reference.
	check(t, d, shdata.IMSUserStateData, false)
	check(t, d, shdata.SCSCFNameData, false)
	for _, s := range []string{
		`<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState><SCSCFName/></Sh-IMS-Data></Sh-Data>`,
		`<Sh-Data><Sh-IMS-Data><SCSCFName/><SCSCFName/></Sh-IMS-Data></Sh-Data>`,
		`<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState><IMSUserState>0</IMSUserState></Sh-IMS-Data></Sh-Data>`,
		prefix + `<IMSUserState>1</IMSUserState>` + suffix,
		`<Sh-Data><Sh-IMS-Data><SCSCFName><x/></SCSCFName></Sh-IMS-Data></Sh-Data>`,
	} {
		if _, e := shdata.Parse([]byte(s), shdata.DefaultLimits()); !errors.Is(e, shdata.ErrStructure) {
			t.Fatal(s, e)
		}
	}
	d = parse(t, []byte(`<Sh-Data><Sh-IMS-Data><x:IMSUserState xmlns:x="urn:x">1</x:IMSUserState></Sh-IMS-Data></Sh-Data>`))
	check(t, d, shdata.IMSUserStateData, false)
	if d.Data().IMSUserState.Present {
		t.Fatal("namespace confused")
	}
	// Escaping is interpreted without changing raw identity or mutating text.
	s := strings.Replace(string(fixture(t, "b19-scscf-12")), "sip:scscf.example.net", "sip:&#115;cscf.example.net", 1)
	d = parse(t, []byte(s))
	check(t, d, shdata.SCSCFNameData, true)
	if d.Data().SCSCFName.Value != "sip:scscf.example.net" || !bytes.Equal(d.Raw(), []byte(s)) {
		t.Fatal(d.Data())
	}
}
