//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Independent offline dictionary evidence; no diameterkit imports.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
	"os"
	"path/filepath"
	"strings"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func load(name string) []byte {
	b, e := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(e)
	b, e = hex.DecodeString(strings.Join(strings.Fields(string(b)), ""))
	must(e)
	return b
}
func check(v *diam.AVP) {
	switch v.Code {
	case 3102, 3316, 3317, 3320:
		if v.VendorID != 10415 {
			panic("vendor")
		}
		g, ok := v.Data.(*diam.GroupedAVP)
		if !ok {
			panic(fmt.Sprintf("group %d: %T", v.Code, v.Data))
		}
		for _, child := range g.AVP {
			check(child)
		}
	case 1:
		if v.VendorID != 0 || v.Flags != 0x40 {
			panic("IMSI flags/vendor")
		}
		if _, ok := v.Data.(datatype.UTF8String); !ok {
			panic("IMSI type")
		}
		if string(v.Data.Serialize()) != "001010123456789" {
			panic("IMSI value")
		}
	case 701:
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("MSISDN type")
		}
	case 3308:
		if _, ok := v.Data.(datatype.Unsigned32); !ok {
			panic("cause type")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"d21-t4-rdr-r", "d21-t4-rdr-a", "d21-t4-rdr-f", "d21-t4-alr-r", "d21-t4-alr-a", "d21-t4-alr-f", "b26-t4-rdr-mme", "b26-t4-rda-5001", "b26-t4-rdr-bothid", "b26-t4-rdr-emptyid"} {
		raw := load(name)
		original := load("d21-t4-rdr-r")
		command := uint32(8388649)
		session := "isg.example;21;11"
		if strings.Contains(name, "alr") {
			original = load("d21-t4-alr-r")
			command = 8388648
			session = "hss.example;21;12"
		}
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != command || m.Header.HopByHopID != binary.BigEndian.Uint32(original[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(original[16:20]) {
			panic("header/pair")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = true
				if string(v.Data.Serialize()) != session {
					panic("Session-Id")
				}
			}
		}
		if !found {
			panic("Session-Id missing")
		}
		// Serialize before a later ReadMessage can reuse the legacy decoder buffer.
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(raw, out) {
			panic("exact reencode " + name)
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary types, paired header/session and exact reencode; identity controls are not semantic or deployment certification\n", name, len(raw))
	}
}
