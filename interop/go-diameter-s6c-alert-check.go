//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Standalone evidence tool. Run from a disposable module; never a core dependency.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func field(as []*diam.AVP, code, vendor uint32) *diam.AVP {
	var found *diam.AVP
	for _, a := range as {
		if a.Code == code && a.VendorID == vendor {
			if found != nil {
				panic("duplicate field")
			}
			found = a
		}
	}
	if found == nil {
		panic(fmt.Sprintf("missing %d/%d", vendor, code))
	}
	return found
}
func u32(a *diam.AVP, want uint32) {
	b := a.Data.Serialize()
	if len(b) != 4 || binary.BigEndian.Uint32(b) != want {
		panic("u32 mismatch")
	}
}
func text(a *diam.AVP, want string) {
	if string(a.Data.Serialize()) != want {
		panic("text mismatch")
	}
}
func group(a *diam.AVP, app uint32) []*diam.AVP {
	if g, ok := a.Data.(*diam.GroupedAVP); ok {
		return g.AVP
	}
	g, ok := a.Data.(datatype.Grouped)
	if !ok {
		panic(fmt.Sprintf("not a dictionary Grouped: %T", a.Data))
	}
	parsed, err := diam.DecodeGrouped(g, app, dict.Default)
	must(err)
	return parsed.AVP
}

func load(name string) []byte {
	raw, err := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(err)
	b, err := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
	must(err)
	return b
}
func main() {
	if len(os.Args) != 2 {
		panic("argument: absolute testdata/codec directory")
	}
	for _, name := range []string{"v-rdr-01", "v-rda-01", "v-rda-02", "v-rda-03", "v-alr-01", "v-ala-01", "v-ala-02"} {
		b := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(err)
		alert := strings.HasPrefix(name, "v-al")
		code, hbh, e2e, session := uint32(8388649), uint32(0xd1d2d3d4), uint32(0xe1e2e3e4), "gmsc.example;18;1"
		if alert {
			code, hbh, e2e, session = 8388648, 0xd1d2d3d5, 0xe1e2e3e5, "hss.example;18;1"
		}
		if m.Header.CommandCode != code || m.Header.ApplicationID != 16777312 || m.Header.HopByHopID != hbh || m.Header.EndToEndID != e2e {
			panic("command/application/association mismatch")
		}
		text(field(m.AVP, 263, 0), session)
		if name != "v-ala-02" {
			u32(field(m.AVP, 277, 0), 1)
		}
		switch name {
		case "v-rdr-01":
			g := group(field(m.AVP, 3316, 10415), 16777312)
			node := group(field(g, 3317, 10415), 16777312)
			u32(field(node, 3321, 10415), 0)
			identity := group(field(m.AVP, 3102, 10415), 16777312)
			if hex.EncodeToString(field(identity, 701, 10415).Data.Serialize()) != "2120550521f3" {
				panic("report MSISDN")
			}
		case "v-rda-01":
			u32(field(m.AVP, 268, 0), 2001)
			identity := group(field(m.AVP, 3102, 10415), 16777312)
			if hex.EncodeToString(field(identity, 701, 10415).Data.Serialize()) != "2120550521f4" {
				panic("alert MSISDN")
			}
		case "v-rda-02", "v-rda-03":
			g := group(field(m.AVP, 297, 0), 16777312)
			u32(field(g, 266, 0), 10415)
			result := uint32(5558)
			if name == "v-rda-03" {
				result = 5001
			}
			u32(field(g, 298, 0), result)
		case "v-alr-01":
			text(field(m.AVP, 293, 0), "iwmsc.example")
			value := field(m.AVP, 3329, 10415)
			if _, ok := value.Data.(datatype.Time); !ok {
				panic("not dictionary Time")
			}
			u32(value, 0xed27251c)
			identity := group(field(m.AVP, 3102, 10415), 16777312)
			if hex.EncodeToString(field(identity, 701, 10415).Data.Serialize()) != "2120550521f4" {
				panic("alert identity")
			}
		case "v-ala-01":
			u32(field(m.AVP, 268, 0), 2001)
		case "v-ala-02":
			u32(field(m.AVP, 268, 0), 3004)
			if m.Header.CommandFlags&0x20 == 0 {
				panic("missing E flag")
			}
		}
		encoded, err := m.Serialize()
		must(err)
		if !bytes.Equal(encoded, b) {
			panic("independent re-encode mismatch")
		}
		fmt.Printf("PASS %s (%d bytes): independent dictionary decode, selected values, exact re-encode\n", name, len(b))
	}
}
