//go:build ignore

// Offline projection of the frozen simulator construction paths. This does not
// run their services, configuration loader, TPDU library or network client.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/avp"
	"github.com/ctrlzy/go-diameter/v4/diam/basetype"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
	"github.com/ctrlzy/go-diameter/v4/diam/sm/smparser"
	"os"
	"path/filepath"
	"strings"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func load(n string) *diam.Message {
	raw, e := os.ReadFile(filepath.Join(os.Args[1], n+".hex"))
	must(e)
	b, e := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
	must(e)
	m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
	must(e)
	return m
}
func put(n string, m *diam.Message) {
	b, e := m.Serialize()
	must(e)
	must(os.WriteFile(filepath.Join(os.Args[1], "b77-sim-"+n+".hex"), []byte(hex.EncodeToString(b)+"\n"), 0644))
	fmt.Println(n, len(b))
}
func tbcd(s string) datatype.OctetString {
	b := make([]byte, (len(s)+1)/2)
	for i := range b {
		lo := s[2*i] - '0'
		hi := byte(15)
		if 2*i+1 < len(s) {
			hi = s[2*i+1] - '0'
		}
		b[i] = lo | hi<<4
	}
	return datatype.OctetString(b)
}
func answer(r *diam.Message, code uint32, host string, vsai uint32) *diam.Message {
	a := r.Answer(code)
	sid, e := r.FindAVP(avp.SessionID, 0)
	must(e)
	a.InsertAVP(diam.NewAVP(avp.SessionID, avp.Mbit, 0, sid.Data))
	a.NewAVP(avp.AuthSessionState, avp.Mbit, 0, datatype.Enumerated(1))
	a.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(host))
	a.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("example.net"))
	if vsai != 0 {
		a.NewAVP(avp.VendorSpecificApplicationID, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(vsai)), diam.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(10415))}})
	}
	return a
}
func main() {
	if len(os.Args) != 2 {
		panic("fixture directory")
	}
	// Fixed session/header IDs, source literals, and an opaque binary payload.
	// The latter is deliberately not claimed as fkgi/sms output or a captured TPDU.
	host := datatype.DiameterIdentity("desthost.org")
	imsi := datatype.UTF8String("46031013800000000")
	msisdn := tbcd("13800000000")
	ofr := smparser.OFR{SessionId: "simulator;77;ofr", AuthSessionState: 1, OriginHost: "localhost:3868", OriginRealm: "simulator.mme.org", DestinationHost: &host, DestinationRealm: "desthost.realm.org", ScAddress: tbcd("138888888"), UserIdentifier: basetype.UserIdentifier{UserName: &imsi, Msisdn: &msisdn}, SmRpUi: datatype.OctetString(string([]byte{0, 255, 0x30, 0x42, 0x30, 0x44}))}
	for _, n := range []string{"mme-ofr-original", "mme-ofr-header-fixed", "mme-ofr-host-fixed", "mme-ofr-fixed", "mme-ofr-host-invalid"} {
		if n == "mme-ofr-host-fixed" || n == "mme-ofr-fixed" {
			ofr.OriginHost = "mme.example"
		}
		if n == "mme-ofr-fixed" {
			imsi = "460310000000000"
		}
		if n == "mme-ofr-host-invalid" {
			ofr.OriginHost = "localhost:3868"
		}
		m := diam.NewRequest(diam.MOForwardShortMessage, diam.TGPP_SGD_GDD_APP_ID, dict.Default)
		must(m.Marshal(&ofr))
		if n != "mme-ofr-original" {
			m.Header.CommandFlags |= 0x40
		}
		m.Header.HopByHopID = 7701
		m.Header.EndToEndID = 17701
		put(n, m)
	}
	ai := datatype.UTF8String("111111111")
	am := tbcd("13123123")
	event := datatype.Unsigned32(1)
	alr := smparser.ALR{SessionId: "simulator;77;alr", AuthSessionState: 1, OriginHost: "mme.example", OriginRealm: "simulator.mme.org", DestinationHost: &host, DestinationRealm: "desthost.realm.org", ScAddress: tbcd("138888888"), UserIdentifier: basetype.UserIdentifier{UserName: &ai, Msisdn: &am}, SmsGmscAlertEvent: &event}
	for _, n := range []string{"mme-alr-original", "mme-alr-both-identities", "mme-alr-fixed"} {
		if n == "mme-alr-fixed" {
			alr.UserIdentifier.Msisdn = nil
		}
		m := diam.NewRequest(diam.AlertServiceCenter, diam.TGPP_S6C_APP_ID, dict.Default)
		must(m.Marshal(&alr))
		if n != "mme-alr-original" {
			m.Header.CommandFlags |= 0x40
		}
		m.Header.HopByHopID = 7702
		m.Header.EndToEndID = 17702
		put(n, m)
	}
	r := load("v-srr-01")
	var req smparser.SRR
	must(r.Unmarshal(&req))
	name, realm := datatype.DiameterIdentity("test-mme-name"), datatype.DiameterIdentity("test-mme-realm")
	num := tbcd("1234567890")
	id := datatype.UTF8String("123499999999")
	for _, code := range []uint32{2001, 5012} {
		rc := datatype.Unsigned32(code)
		a := r.Answer(code)
		body := smparser.SRA{SessionId: req.SessionId, ResultCode: &rc, AuthSessionState: req.AuthSessionState, OriginHost: *req.DestinationHost, OriginRealm: req.DestinationRealm, UserName: &id, ServingNode: &basetype.ServingNode{MmeName: &name, MmeRealm: &realm, MmeNumberForMtSms: &num}}
		must(a.Marshal(&body))
		put(fmt.Sprintf("mme-sra-%d", code), a)
	}
	put("mme-sra-failure-fixed", answer(r, 5012, "hss.example", 0))
	for _, tc := range []struct{ r, n, h string }{{"v-tfr-01", "mme-tfa", "mme.example"}, {"v-rdr-01", "mme-rda", "hss.example"}} {
		for _, code := range []uint32{2001, 5012} {
			put(fmt.Sprintf("%s-%d", tc.n, code), answer(load(tc.r), code, tc.h, 0))
		}
	}
	r = load("v-ofr-01")
	var mo smparser.OFR
	must(r.Unmarshal(&mo))
	rc := datatype.Unsigned32(2001)
	for _, state := range []datatype.Enumerated{0, 1} {
		a := r.Answer(2001)
		must(a.Marshal(&smparser.OFA{SessionId: mo.SessionId, ResultCode: &rc, AuthSessionState: state, OriginHost: "iwmsc.example", OriginRealm: "example.net"}))
		put(fmt.Sprintf("poc-ofa-state%d", state), a)
	}
	for _, tc := range []struct{ r, n string }{{"v-sh-udr-01", "sh-uda"}, {"v-sh-pur-01", "sh-pua"}, {"v-sh-snr-01", "sh-sna"}} {
		for _, code := range []uint32{2001, 5012} {
			put(fmt.Sprintf("%s-%d", tc.n, code), answer(load(tc.r), code, "hss.example", diam.TGPP_SH_APP_ID))
		}
	}
	a := answer(load("v-sh-udr-01"), 2001, "hss.example", diam.TGPP_SH_APP_ID)
	a.NewAVP(avp.UserData, avp.Mbit|avp.Vbit, 10415, datatype.OctetString("<Sh-Data/>"))
	put("sh-uda-fixed", a)
	put("sh-rda-vsai", answer(load("v-rdr-01"), 2001, "hss.example", diam.TGPP_S6C_APP_ID))
	// Same handler envelope but requests exercise correlation/condition boundaries.
	put("sh-pua-no-proxy", answer(load("v-sh-pur-01"), 2001, "hss.example", diam.TGPP_SH_APP_ID))
	for _, code := range []uint32{316, 318} {
		m := diam.NewRequest(code, 16777251, dict.Default)
		m.Header.HopByHopID = 7703
		m.Header.EndToEndID = 17703
		put(fmt.Sprintf("s6a-%d", code), m)
	}
}
