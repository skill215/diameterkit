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
	case 297, 3102:
		vendor := uint32(10415)
		if v.Code == 297 {
			vendor = 0
		}
		if v.VendorID != vendor {
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
	case 3310, 3312, 3313, 3314, 3315, 266, 298:
		if _, ok := v.Data.(datatype.Unsigned32); !ok {
			panic("flags type")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"v24-sra-zero", "c24-sra-nodiag", "f24-s6c-sr-5552", "f24-s6c-sr-5556", "f24-s6c-sr-5557", "v25-diag-mme", "v25-diag-sgsn", "c25-diag-mme-reserved", "n25-diag-mme-256", "v24-s6c-sr-r", "b28-srr-smea", "b28-sra-ipsm-imsi", "b28-sra-absent-adjuncts", "b28-sra-diag-dup", "b28-sra-diag-width", "b28-sra-diag-wrongdomain", "b28-sra-diag-all", "b28-t4-sra-5550", "b28-t4-sra-5552", "b28-t4-sra-5556", "b28-t4-sra-5557"} {
		raw := load(name)
		original := load("b28-srr-smea")
		command := uint32(8388647)
		session := "gmsc.example;24;5"
		if strings.Contains(name, "25-diag") {
			original = load("v25-srr-two-r")
			session = "gmsc.example;25;32"
		}
		if strings.HasPrefix(name, "b28-t4-") {
			original = load("d21-t4-srr-r")
			session = "gmsc.example;21;10"
		}
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
		if name == "b28-sra-diag-width" {
			expected := bytes.Clone(raw)
			found := false
			for off := 20; off < len(raw); {
				size := int(binary.BigEndian.Uint32(raw[off+4:off+8]) & 0xffffff)
				if binary.BigEndian.Uint32(raw[off:]) == 3314 {
					if size != 15 {
						panic("control width")
					}
					expected[off+7] = 16
					found = true
				}
				off += (size + 3) &^ 3
			}
			if !found || !bytes.Equal(expected, out) {
				panic("unexpected normalization")
			}
			fmt.Printf("PASS %s (%d bytes): legacy decoder accepts three-byte U32 and pads/reencodes length15 as16; explicitly not exact reencode or semantic compliance\n", name, len(raw))
			continue
		}
		if !bytes.Equal(raw, out) {
			panic("exact reencode " + name)
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary types, paired header/session and exact reencode; result/diagnostic controls are not role/UNRI semantics or deployment certification\n", name, len(raw))
	}
}
