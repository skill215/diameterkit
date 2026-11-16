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
func check(v *diam.AVP, name string) {
	switch v.Code {
	case 2401, 2406:
		if v.VendorID != 10415 {
			panic("vendor")
		}
		g, ok := v.Data.(*diam.GroupedAVP)
		if !ok {
			panic(fmt.Sprintf("group %d: %T", v.Code, v.Data))
		}
		for _, child := range g.AVP {
			check(child, name)
		}
	case 1:
		if v.VendorID != 0 || v.Flags != 0x40 {
			panic("IMSI flags/vendor")
		}
		if _, ok := v.Data.(datatype.UTF8String); !ok {
			panic("IMSI type")
		}
		want := "001010123456789"
		if name == "b27-t4-sra-otherimsi" {
			want = "001010123456780"
		}
		if name == "b27-t4-srr-badimsi" {
			want = "not-an-imsi"
		}
		if string(v.Data.Serialize()) != want {
			panic("IMSI value")
		}
	case 701, 1645, 1489, 3300, 3309:
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("MSISDN type")
		}
	case 2402, 2408:
		if _, ok := v.Data.(datatype.DiameterIdentity); !ok {
			panic("node identity type")
		}
	case 3310:
		if _, ok := v.Data.(datatype.Unsigned32); !ok {
			panic("flags type")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"d21-t4-srr-r", "d21-t4-srr-a", "d21-t4-srr-f", "b27-t4-srr-bothid", "b27-t4-srr-noid", "b27-t4-srr-dupimsi", "b27-t4-srr-badimsi", "b27-t4-sra-otherimsi", "b27-t4-sra-noimsi", "b27-t4-sra-nonode", "b27-t4-sra-extra-node", "b27-t4-srr-gprs", "b27-t4-sra-e"} {
		raw := load(name)
		original := load("d21-t4-srr-r")
		command := uint32(8388647)
		session := "gmsc.example;21;10"
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != command || m.Header.HopByHopID != binary.BigEndian.Uint32(original[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(original[16:20]) {
			panic("header/pair")
		}
		found := false
		for _, v := range m.AVP {
			check(v, name)
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
