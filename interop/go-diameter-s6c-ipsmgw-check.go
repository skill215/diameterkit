//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Independent dictionary evidence: no diameterkit imports or peer connections.
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
func load(name string) []byte {
	b, err := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(err)
	b, err = hex.DecodeString(strings.Join(strings.Fields(string(b)), ""))
	must(err)
	return b
}
func check(v *diam.AVP) {
	switch v.Code {
	case 2401, 2406:
		g, ok := v.Data.(*diam.GroupedAVP)
		if !ok {
			panic("group type")
		}
		for _, child := range g.AVP {
			check(child)
		}
	case 3100:
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("number type")
		}
	case 3101:
		if v.VendorID == 10415 {
			if _, ok := v.Data.(datatype.DiameterIdentity); !ok {
				panic(fmt.Sprintf("identity %d type %T", v.Code, v.Data))
			}
		}
	case 3112:
		if v.VendorID != 10415 {
			if _, ok := v.Data.(datatype.Unknown); !ok {
				panic("vendor-zero shadow type")
			}
			return
		}
		// Legacy dictionary declares Realm as OctetString and predates SBI.
		// Neither supplies semantic identity/enum validation.
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic(fmt.Sprintf("legacy Realm/SBI type %T", v.Data))
		}
	case 3348:
		if _, ok := v.Data.(datatype.Unknown); !ok {
			panic(fmt.Sprintf("SBI unknown type %T", v.Data))
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"c24-add-ipsmgw", "c27-realm-nr", "c27-realm-nhrb", "n27-realm-text", "n27-realm-nouser", "b30-ipsm-n", "b30-ipsm-nh", "b30-ipsm-nb0", "b30-ipsm-nhb1", "b30-ipsm-nhr", "b30-ipsm-sbi2", "b30-ipsm-nonumber", "b30-ipsm-badnumber", "b30-ipsm-duprealm", "b30-ipsm-vendorrealm", "b30-ipsm-unknown", "b30-ipsm-mixed", "b30-ipsm-addn"} {
		raw := load(name)
		original := load("v25-srr-two-r")
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388647 {
			panic("command")
		}
		for _, v := range m.AVP {
			check(v)
		}
		if strings.HasPrefix(name, "b30-") {
			if m.Header.HopByHopID != binary.BigEndian.Uint32(original[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(original[16:20]) {
				panic("pair IDs")
			}
			found := false
			for _, v := range m.AVP {
				if v.Code == 263 {
					found = string(v.Data.Serialize()) == "gmsc.example;25;32"
				}
			}
			if !found {
				panic("session")
			}
		}
		// Serialize before another ReadMessage reuses the old decoder buffer.
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, out) {
			panic("exact reencode " + name)
		}
		fmt.Printf("PASS %s (%d bytes): dictionary Number/Name types and exact reencode; projected pairs checked, old IMS candidates only structural; Realm OctetString and SBI raw, no semantic/compatibility/deployment certification\n", name, len(raw))
	}
}
