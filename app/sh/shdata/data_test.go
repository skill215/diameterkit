// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
)

const prefix = "<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension>"
const suffix = "</Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>"

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("../../../testdata/shdata/" + name + ".xml")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func parse(t *testing.T, b []byte) *shdata.Document {
	t.Helper()
	d, e := shdata.Parse(b, shdata.DefaultLimits())
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func check(t *testing.T, d *shdata.Document, use shdata.Use, want bool) {
	t.Helper()
	r, e := d.Check(use)
	if e != nil || r.Scope != shdata.Scope || r.Complete != want {
		t.Fatalf("report=%+v err=%v", r, e)
	}
}
func zero() avp.Optional[uint8] { return avp.Some(uint8(0)) }
func TestFrozenCandidatesAndConstruction(t *testing.T) {
	l := shdata.DefaultLimits()
	cases := []struct {
		name  string
		use   shdata.Use
		build func() ([]byte, error)
	}{
		{"x-sh-reg-01", shdata.SMSUpdate, func() ([]byte, error) {
			return shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: "12025550999"}, l)
		}},
		{"x-sh-pull-01", shdata.SMSPull, func() ([]byte, error) {
			return shdata.MarshalSMSPull(shdata.SMSRegistration{Number: "12025550999", SCAddress: avp.Some("12025550100")}, l)
		}},
		{"x-sh-dereg-01", shdata.SMSUpdate, func() ([]byte, error) { return shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: ""}, l) }},
		{"x-sh-mme-01", shdata.ReachabilityNotification, func() ([]byte, error) { return shdata.MarshalReachability(shdata.Reachability{MME: zero()}, l) }},
		{"x-sh-sgsn-01", shdata.ReachabilityNotification, func() ([]byte, error) { return shdata.MarshalReachability(shdata.Reachability{SGSN: zero()}, l) }},
		{"x-sh-amf-01", shdata.ReachabilityNotification, func() ([]byte, error) {
			return shdata.MarshalReachability(shdata.Reachability{AMF3GPP: zero(), AMFNon3GPP: zero()}, l)
		}},
		{"x-sh-sbi-01", shdata.SMSUpdate, func() ([]byte, error) {
			return shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: "12025550999", SBI: avp.Some(uint8(1))}, l)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := fixture(t, c.name)
			d := parse(t, raw)
			check(t, d, c.use, true)
			built, e := c.build()
			if e != nil || !bytes.Equal(raw, built) {
				t.Fatalf("construction %s %v", built, e)
			}
			if !bytes.Equal(d.Raw(), raw) {
				t.Fatal("changed raw")
			}
		})
	}
	d := parse(t, fixture(t, "x-sh-ext-01"))
	check(t, d, shdata.SMSUpdate, false)
	if d.Data().SMS.Value.Number != "12025550999" || len(d.Fragments()) != 1 {
		t.Fatal(d.Data(), d.Fragments())
	}
	f := d.Fragments()[0]
	if f.Namespace != "urn:example:ext" || f.Local != "IP-SM-GW-Number" || !bytes.HasPrefix(d.Raw()[f.Start:f.End], []byte("<x:IP-SM-GW-Number")) {
		t.Fatal(f)
	}
	for _, name := range []string{"x23-ident", "x23-repo-delete", "x23-repo-data", "x23-ifc", "x23-dsai"} {
		t.Run(name, func(t *testing.T) {
			d := parse(t, fixture(t, name))
			check(t, d, shdata.SMSUpdate, false)
			if len(d.Fragments()) == 0 {
				t.Fatal("silently claimed interpretation")
			}
		})
	}
}
func TestDeepStructuralNegatives(t *testing.T) {
	reg := string(fixture(t, "x-sh-reg-01"))
	sbi := string(fixture(t, "x-sh-sbi-01"))
	mme := string(fixture(t, "x-sh-mme-01"))
	amf := string(fixture(t, "x-sh-amf-01"))
	cases := map[string]string{
		"missing-number":        strings.Replace(reg, "<IP-SM-GW-Number>12025550999</IP-SM-GW-Number>", "", 1),
		"duplicate-number":      strings.Replace(reg, "</SMSRegistrationInfo>", "<IP-SM-GW-Number/></SMSRegistrationInfo>", 1),
		"missing-parent":        strings.Replace(strings.Replace(reg, "<Extension>", "", 1), "</Extension>", "", 1),
		"extra-parent":          strings.Replace(reg, "<SMSRegistrationInfo>", "<Extension><SMSRegistrationInfo>", 1),
		"wrong-namespace":       strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns="urn:wrong">`, 1),
		"wrong-sbi-spelling":    strings.ReplaceAll(sbi, "IP-SM-GW-SBISupportIndication", "IP-SM-GW-SBI-SupportIndication"),
		"sbi-2":                 strings.Replace(sbi, ">1<", ">2<", 1),
		"mme-1":                 strings.Replace(mme, ">0<", ">1<", 1),
		"empty-enum":            strings.Replace(mme, ">0<", "><", 1),
		"wrong-sgsn-parent":     strings.ReplaceAll(mme, "UEIPReachabilityMME", "UEIPReachabilitySGSN"),
		"wrong-amf-parent":      strings.ReplaceAll(mme, "UEIPReachabilityMME", "UEIPReachabilityAMF3GPP"),
		"reversed-amf":          strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(amf, "UEIPReachabilityAMF3GPP", "TEMP"), "UEIPReachabilityAMFnon3GPP", "UEIPReachabilityAMF3GPP"), "TEMP", "UEIPReachabilityAMFnon3GPP"),
		"wrong-top-order":       prefix + "<SMSRegistrationInfo><IP-SM-GW-Number/></SMSRegistrationInfo><UEReachabilityForIP/>" + suffix,
		"duplicate-container":   strings.Replace(reg, "</SMSRegistrationInfo>", "</SMSRegistrationInfo><SMSRegistrationInfo><IP-SM-GW-Number/></SMSRegistrationInfo>", 1),
		"wildcard-forbidden":    strings.Replace(sbi, "</IP-SM-GW-SBISupportIndication>", `</IP-SM-GW-SBISupportIndication><x:extra xmlns:x="urn:x"/>`, 1),
		"wildcard-before-field": strings.Replace(reg, "<SMSRegistrationInfo>", `<SMSRegistrationInfo><x:extra xmlns:x="urn:x"/>`, 1),
		"unqualified-unknown":   strings.Replace(reg, "</SMSRegistrationInfo>", "<Unknown/></SMSRegistrationInfo>", 1),
		"child-in-leaf":         strings.Replace(reg, "12025550999", "<child/>", 1),
		"mixed-text":            strings.Replace(reg, "<SMSRegistrationInfo>", "<SMSRegistrationInfo>unexpected", 1),
		"sc-21-unicode":         strings.Replace(string(fixture(t, "x-sh-pull-01")), "12025550100", strings.Repeat("界", 21), 1),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			d, e := shdata.Parse([]byte(s), shdata.DefaultLimits())
			if d != nil || e == nil {
				t.Fatal("accepted invalid structure", e)
			}
		})
	}
}
func TestPresenceAndUseReports(t *testing.T) {
	for _, raw := range []string{"<Sh-Data/>", prefix + "<UEReachabilityForIP/>" + suffix} {
		d := parse(t, []byte(raw))
		check(t, d, shdata.SMSUpdate, false)
		check(t, d, shdata.ReachabilityNotification, false)
	}
	d := parse(t, fixture(t, "x-sh-dereg-01"))
	if !d.Data().SMS.Present || d.Data().SMS.Value.Number != "" {
		t.Fatal(d.Data())
	}
	check(t, d, shdata.SMSUpdate, true)
	check(t, d, shdata.ReachabilityNotification, false)
	d = parse(t, fixture(t, "x-sh-pull-01"))
	check(t, d, shdata.SMSPull, true)
	check(t, d, shdata.SMSUpdate, false)
	raw := strings.Replace(string(fixture(t, "x-sh-pull-01")), "12025550100", "", 1)
	d = parse(t, []byte(raw))
	if !d.Data().SMS.Value.SCAddress.Present || d.Data().SMS.Value.SCAddress.Value != "" {
		t.Fatal(d.Data())
	}
	check(t, d, shdata.SMSPull, true)
	for _, s := range []string{"0", "+0", "-0", "0000", " \t+000\r\n"} {
		d = parse(t, []byte(strings.Replace(string(fixture(t, "x-sh-mme-01")), ">0<", ">"+s+"<", 1)))
		if d.Data().Reachability.Value.MME != zero() {
			t.Fatal(d.Data())
		}
		check(t, d, shdata.ReachabilityNotification, true)
	}
	for _, s := range []string{"-1", "256", "0 0", "+", "--0", "+-0", "1.0", "０", "\u00a00"} {
		_, e := shdata.Parse([]byte(strings.Replace(string(fixture(t, "x-sh-mme-01")), ">0<", ">"+s+"<", 1)), shdata.DefaultLimits())
		if !errors.Is(e, shdata.ErrStructure) {
			t.Fatalf("enum %q: %v", s, e)
		}
	}
	d = parse(t, []byte(strings.Replace(string(fixture(t, "x-sh-sbi-01")), ">1<", ">0<", 1)))
	if !d.Data().SMS.Value.SBI.Present || d.Data().SMS.Value.SBI.Value != 0 {
		t.Fatal(d.Data())
	}
	if _, e := d.Check(0); !errors.Is(e, shdata.ErrArgument) {
		t.Fatal(e)
	}
	for _, n := range []string{"+1202", "01202", "１２", " 1202", "1202 ", "1234567890123456"} {
		d = parse(t, []byte(strings.Replace(string(fixture(t, "x-sh-reg-01")), "12025550999", n, 1)))
		if d.Data().SMS.Value.Number != n {
			t.Fatal("coerced number")
		}
		check(t, d, shdata.SMSUpdate, false)
		if _, e := shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: n}, shdata.DefaultLimits()); !errors.Is(e, shdata.ErrUnsupported) {
			t.Fatal(n, e)
		}
	}
}
func TestXMLSecurityAndNamespaces(t *testing.T) {
	reg := string(fixture(t, "x-sh-reg-01"))
	invalid := []string{
		`<Sh-Data><x:1 xmlns:x="urn:x"/></Sh-Data>`,
		`<Sh-Data><x:e xmlns:x="urn:x" x:1="bad"/></Sh-Data>`,
		"", reg + reg, "junk" + reg, reg + "junk", "&#32;" + reg, "<![CDATA[ ]]>" + reg,
		`<!DOCTYPE Sh-Data SYSTEM "file:///etc/passwd">` + reg,
		`<!DOCTYPE Sh-Data [<!ENTITY bomb "xxx">]>` + reg,
		strings.Replace(reg, "12025550999", "&unknown;", 1),
		strings.Replace(reg, "12025550999", "&#0;", 1),
		strings.Replace(reg, "12025550999", "\x00", 1),
		strings.Replace(reg, "12025550999", "\xc0\x80", 1),
		strings.Replace(reg, "12025550999", "\xed\xa0\x80", 1),
		strings.Replace(reg, "12025550999", "\xf4\x90\x80\x80", 1),
		strings.Replace(reg, "<Sh-Data>", "<Sh-Data a='1' a='2'>", 1),
		strings.Replace(reg, "<Sh-Data>", "<Sh-Data x:a='1'>", 1),
		strings.Replace(reg, "</SMSRegistrationInfo>", "<x:Foo/></SMSRegistrationInfo>", 1),
		strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:x="urn:x" xmlns:y="urn:x" x:a="1" y:a="2">`, 1),
		strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:xml="urn:bad">`, 1),
		strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:x="http://www.w3.org/XML/1998/namespace">`, 1),
		strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:xmlns="urn:x">`, 1),
		strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:x="">`, 1),
		strings.Replace(reg, "</Sh-Data>", "</sh-data>", 1),
		strings.Replace(reg, "</Sh-Data>", "", 1),
		`<?xml version="1.0" encoding="UTF-16"?>` + reg,
		`<?xml?>` + reg, ` <?xml version="1.0"?>` + reg,
		`<?XML version="1.0"?>` + reg, reg + `<?xml version="1.0"?>`,
		strings.Replace(reg, "<SMSRegistrationInfo>", `<SMSRegistrationInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true">`, 1),
	}
	for i, s := range invalid {
		d, e := shdata.Parse([]byte(s), shdata.DefaultLimits())
		if d != nil || e == nil {
			t.Fatalf("invalid XML %d accepted", i)
		}
	}
	for _, head := range []string{"", `<?xml version="1.0"?>`, "\xef\xbb\xbf<?xml version='1.0' encoding='utf-8' standalone='yes'?>", "<!--comment--><?app test?>"} {
		d := parse(t, []byte(head+reg))
		check(t, d, shdata.SMSUpdate, true)
	}
	hints := strings.Replace(reg, "<Sh-Data>", `<Sh-Data xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://127.0.0.1:1/never-fetch" xsi:schemaLocation="urn:x file:///never-read">`, 1)
	check(t, parse(t, []byte(hints)), shdata.SMSUpdate, true)
	// A qualified same-local-name at an actual wildcard position is never read as
	// a standard leaf. Namespace rebinding cannot smuggle it into the selected path.
	s := strings.Replace(reg, "</SMSRegistrationInfo>", `<x:IP-SM-GW-Number xmlns:x="urn:outer"><x:child xmlns:x="urn:inner">中文&amp;é</x:child><IP-SM-GW-Number>bad</IP-SM-GW-Number></x:IP-SM-GW-Number></SMSRegistrationInfo>`, 1)
	d := parse(t, []byte(s))
	if d.Data().SMS.Value.Number != "12025550999" {
		t.Fatal(d.Data())
	}
	check(t, d, shdata.SMSUpdate, false)
	// Resolved namespace and prefix stack restoration are both exercised here.
	s = strings.Replace(reg, "</SMSRegistrationInfo>", `<x:e xmlns:x="urn:x"/><x:e/></SMSRegistrationInfo>`, 1)
	if _, e := shdata.Parse([]byte(s), shdata.DefaultLimits()); !errors.Is(e, shdata.ErrXML) {
		t.Fatal(e)
	}
}
func TestResourceBoundsAndOwnership(t *testing.T) {
	raw := fixture(t, "x-sh-ext-01")
	d := parse(t, raw)
	u := d.Usage()
	l := shdata.DefaultLimits()
	for _, c := range []struct {
		name  string
		set   func(*shdata.Limits, int)
		value int
	}{
		{"bytes", func(l *shdata.Limits, n int) { l.MaxBytes = n; l.MaxTextBytes = n; l.MaxOpaqueBytes = n }, u.Bytes},
		{"depth", func(l *shdata.Limits, n int) { l.MaxDepth = n }, u.MaxDepth},
		{"elements", func(l *shdata.Limits, n int) { l.MaxElements = n }, u.Elements},
		{"attributes", func(l *shdata.Limits, n int) { l.MaxAttributes = n }, u.Attributes},
		{"text", func(l *shdata.Limits, n int) { l.MaxTextBytes = n }, u.TextBytes},
		{"opaque", func(l *shdata.Limits, n int) { l.MaxOpaqueBytes = n }, u.OpaqueBytes},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := l
			c.set(&a, c.value)
			if _, e := shdata.Parse(raw, a); e != nil {
				t.Fatal("exact", e)
			}
			if c.value > 1 {
				c.set(&a, c.value-1)
				if got, e := shdata.Parse(raw, a); got != nil || !errors.Is(e, shdata.ErrLimit) {
					t.Fatal("over", e)
				}
			}
		})
	}
	for _, a := range []shdata.Limits{{}, {MaxBytes: -1}, {MaxBytes: 1 << 21, MaxDepth: 1, MaxElements: 1, MaxAttributes: 1, MaxTextBytes: 1, MaxOpaqueBytes: 1}} {
		if _, e := shdata.Parse(raw, a); !errors.Is(e, shdata.ErrArgument) {
			t.Fatal(e)
		}
	}
	orig := bytes.Clone(raw)
	clear(raw)
	copyRaw := d.Raw()
	clear(copyRaw)
	data := d.Data()
	data.SMS.Value.Number = "changed"
	f := d.Fragments()
	f[0].Start = 0
	r, _ := d.Check(shdata.SMSUpdate)
	r.Issues[0].Path = "changed"
	if !bytes.Equal(d.Raw(), orig) || d.Data().SMS.Value.Number != "12025550999" || d.Fragments()[0].Start == 0 {
		t.Fatal("borrowed mutable data")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				check(t, d, shdata.SMSUpdate, false)
				if !bytes.Equal(d.Raw(), orig) {
					t.Error("concurrent raw")
				}
			}
		})
	}
	wg.Wait()
	// All limits apply inside uninterpreted subtrees too, before retaining them.
	for _, s := range []string{`<x:e xmlns:x="urn:x">` + strings.Repeat("<n>", 70) + strings.Repeat("</n>", 70) + `</x:e>`, `<x:e xmlns:x="urn:x">` + strings.Repeat("<n/>", 600) + `</x:e>`, `<x:e xmlns:x="urn:x">` + strings.Repeat("界", 12000) + `</x:e>`} {
		_, e := shdata.Parse([]byte("<Sh-Data>"+s+"</Sh-Data>"), l)
		if !errors.Is(e, shdata.ErrLimit) {
			t.Fatal(e)
		}
	}
}
func TestUnicodeEscapesAndRawIdentity(t *testing.T) {
	base := string(fixture(t, "x-sh-pull-01"))
	for _, text := range []string{strings.Repeat("界", 20), "A&amp;B", "é", "e\u0301"} {
		s := strings.Replace(base, "12025550100", text, 1)
		d := parse(t, []byte(s))
		check(t, d, shdata.SMSPull, false)
		if !bytes.Equal(d.Raw(), []byte(s)) {
			t.Fatal("normalized raw")
		}
		if text == "A&amp;B" && d.Data().SMS.Value.SCAddress.Value != "A&B" {
			t.Fatal(d.Data())
		}
	}
	reg := string(fixture(t, "x-sh-reg-01"))
	escaped := strings.Replace(reg, "12025550999", "&#49;2025550999", 1)
	d := parse(t, []byte(escaped))
	check(t, d, shdata.SMSUpdate, true)
	if d.Data().SMS.Value.Number != "12025550999" || bytes.Equal(d.Raw(), []byte(reg)) {
		t.Fatal("lost raw escaped representation")
	}
	a := fixture(t, "x-sh-dereg-01")
	b := []byte(strings.Replace(string(a), "<IP-SM-GW-Number></IP-SM-GW-Number>", "<IP-SM-GW-Number/>", 1))
	da, db := parse(t, a), parse(t, b)
	if da.Data() != db.Data() || bytes.Equal(da.Raw(), db.Raw()) {
		t.Fatal("semantic and byte identities conflated")
	}
}
func TestBuilderRejectionsAndPresence(t *testing.T) {
	l := shdata.DefaultLimits()
	if _, e := shdata.MarshalSMSUpdate(shdata.SMSRegistration{SCAddress: avp.Some("")}, l); !errors.Is(e, shdata.ErrArgument) {
		t.Fatal(e)
	}
	if _, e := shdata.MarshalSMSPull(shdata.SMSRegistration{SBI: avp.Some(uint8(2))}, l); !errors.Is(e, shdata.ErrStructure) {
		t.Fatal(e)
	}
	if _, e := shdata.MarshalReachability(shdata.Reachability{}, l); !errors.Is(e, shdata.ErrStructure) {
		t.Fatal(e)
	}
	if _, e := shdata.MarshalReachability(shdata.Reachability{MME: avp.Some(uint8(1))}, l); !errors.Is(e, shdata.ErrStructure) {
		t.Fatal(e)
	}
	for _, s := range []string{`1<&`, "1234567890123456", "１２", "+12"} {
		if _, e := shdata.MarshalSMSPull(shdata.SMSRegistration{SCAddress: avp.Some(s)}, l); !errors.Is(e, shdata.ErrUnsupported) {
			t.Fatal(s, e)
		}
	}
	b, e := shdata.MarshalSMSPull(shdata.SMSRegistration{SCAddress: avp.Some(""), SBI: zero()}, l)
	if e != nil {
		t.Fatal(e)
	}
	d := parse(t, b)
	check(t, d, shdata.SMSPull, true)
	if !d.Data().SMS.Value.SBI.Present || !d.Data().SMS.Value.SCAddress.Present {
		t.Fatal(d.Data())
	}
	b, e = shdata.MarshalReachability(shdata.Reachability{MME: zero(), SGSN: zero(), AMF3GPP: zero(), AMFNon3GPP: zero()}, l)
	if e != nil {
		t.Fatal(e)
	}
	d = parse(t, b)
	check(t, d, shdata.ReachabilityNotification, true)
	if !d.Data().Reachability.Value.AMFNon3GPP.Present {
		t.Fatal(d.Data())
	}
	for _, modify := range []func(*shdata.Limits){func(l *shdata.Limits) { l.MaxBytes = 100; l.MaxTextBytes = 50; l.MaxOpaqueBytes = 50 }, func(l *shdata.Limits) { l.MaxDepth = 2 }, func(l *shdata.Limits) { l.MaxTextBytes = 2 }, func(l *shdata.Limits) { l.MaxElements = 2 }} {
		a := l
		modify(&a)
		b, e := shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: "12025550999"}, a)
		if b != nil || !errors.Is(e, shdata.ErrLimit) {
			t.Fatal("builder bounds", e)
		}
	}
}
func FuzzParse(f *testing.F) {
	for _, s := range []string{`<Sh-Data><Sh-IMS-Data><IMSUserState>3</IMSUserState></Sh-IMS-Data></Sh-Data>`, `<Sh-Data><Sh-IMS-Data><SCSCFName>sip:scscf.example.net</SCSCFName></Sh-IMS-Data></Sh-Data>`, `<Sh-Data><Sh-IMS-Data><SCSCFName/></Sh-IMS-Data></Sh-Data>`, "<Sh-Data/>", prefix + "<SMSRegistrationInfo><IP-SM-GW-Number/></SMSRegistrationInfo>" + suffix, prefix + "<UEReachabilityForIP><UEIPReachabilityMME>0</UEIPReachabilityMME></UEReachabilityForIP>" + suffix, `<Sh-Data><x:e xmlns:x="urn:x"/></Sh-Data>`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		d, e := shdata.Parse(b, shdata.DefaultLimits())
		if e != nil {
			if d != nil {
				t.Fatal("partial result")
			}
			return
		}
		if !bytes.Equal(d.Raw(), b) {
			t.Fatal("raw changed")
		}
		for _, frag := range d.Fragments() {
			if frag.Start < 0 || frag.End < frag.Start || frag.End > len(b) {
				t.Fatal(frag)
			}
		}
		for use := shdata.SMSUpdate; use <= shdata.SCSCFNameData; use++ {
			if _, e := d.Check(use); e != nil {
				t.Fatal(e)
			}
		}
	})
}

func TestEverySelectedLeafAndExtensionBudget(t *testing.T) {
	for _, tc := range []struct{ file, leaf, value string }{
		{"x-sh-pull-01", "SCAddress", "12025550100"},
		{"x-sh-sbi-01", "IP-SM-GW-SBISupportIndication", "1"},
		{"x-sh-mme-01", "UEIPReachabilityMME", "0"},
		{"x-sh-sgsn-01", "UEIPReachabilitySGSN", "0"},
		{"x-sh-amf-01", "UEIPReachabilityAMF3GPP", "0"},
		{"x-sh-amf-01", "UEIPReachabilityAMFnon3GPP", "0"},
	} {
		t.Run(tc.leaf, func(t *testing.T) {
			raw := string(fixture(t, tc.file))
			elem := "<" + tc.leaf + ">" + tc.value + "</" + tc.leaf + ">"
			bad := strings.Replace(raw, elem, elem+elem, 1)
			if _, e := shdata.Parse([]byte(bad), shdata.DefaultLimits()); !errors.Is(e, shdata.ErrStructure) {
				t.Fatal("duplicate", e)
			}
			if tc.leaf != "SCAddress" {
				for _, v := range []string{"", "256", "-1", "2"} {
					bad = strings.Replace(raw, elem, "<"+tc.leaf+">"+v+"</"+tc.leaf+">", 1)
					if _, e := shdata.Parse([]byte(bad), shdata.DefaultLimits()); !errors.Is(e, shdata.ErrStructure) {
						t.Fatal(v, e)
					}
				}
			}
		})
	}
	// Attribute limits apply to opaque children and count xmlns declarations.
	raw := []byte(`<Sh-Data><x:e xmlns:x="urn:x" a="1" b="2"/></Sh-Data>`)
	l := shdata.DefaultLimits()
	l.MaxAttributes = 3
	if _, e := shdata.Parse(raw, l); e != nil {
		t.Fatal(e)
	}
	l.MaxAttributes = 2
	if _, e := shdata.Parse(raw, l); !errors.Is(e, shdata.ErrLimit) {
		t.Fatal(e)
	}
	// A default namespace applies only to its subtree; an unqualified nested
	// lookalike inside an opaque subtree still cannot affect selected data.
	reg := string(fixture(t, "x-sh-reg-01"))
	raw = []byte(strings.Replace(reg, "</SMSRegistrationInfo>", `<IP-SM-GW-Number xmlns="urn:other"><IP-SM-GW-Number xmlns="">bad</IP-SM-GW-Number></IP-SM-GW-Number></SMSRegistrationInfo>`, 1))
	d := parse(t, raw)
	if d.Data().SMS.Value.Number != "12025550999" {
		t.Fatal(d.Data())
	}
	check(t, d, shdata.SMSUpdate, false)
	// Literal comments/CDATA do not change xs:string values, raw remains owned.
	raw = []byte(strings.Replace(reg, "12025550999", "<![CDATA[1202]]><!--split-->5550999", 1))
	d = parse(t, raw)
	check(t, d, shdata.SMSUpdate, true)
	if d.Data().SMS.Value.Number != "12025550999" {
		t.Fatal(d.Data())
	}
}
