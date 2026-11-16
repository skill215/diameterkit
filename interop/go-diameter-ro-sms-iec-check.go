//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Standalone offline dictionary evidence; no diameterkit imports.
package main

import (
	"bytes"

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
	case 443, 456, 437, 431, 873, 2000, 2009, 2026, 2002:
		if _, ok := v.Data.(*diam.GroupedAVP); !ok {
			panic(fmt.Sprintf("group type %d/%d: %T", v.VendorID, v.Code, v.Data))
		}
	case 417:
		if v.VendorID == 0 {
			if _, ok := v.Data.(datatype.Unsigned64); !ok {
				panic(fmt.Sprintf("unit type %T", v.Data))
			}
		}
	case 2022:
		if v.VendorID == 10415 {
			if _, ok := v.Data.(datatype.OctetString); !ok {
				panic("refund type")
			}
		}
	case 2010:
		if v.VendorID == 10415 {
			if _, ok := v.Data.(datatype.Address); !ok {
				panic(fmt.Sprintf("address type %T", v.Data))
			}
		}
	}
	if g, ok := v.Data.(*diam.GroupedAVP); ok {
		for _, c := range g.AVP {
			check(c)
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"v-iec-ccr-01", "v-iec-cca-01", "v-iec-cca-02", "b39-r-zero", "b39-r-high", "b39-r-max", "b39-r-duplicate-unit", "b39-a-zero", "b39-a-high", "b39-a-max", "b39-a-duplicate-unit", "b39-r-no-action", "b39-r-refund", "b39-r-initial", "b39-r-sequence", "b39-r-auth-app", "b39-r-empty-rsu", "b39-r-no-rsu", "b39-r-extra-time", "b39-r-top-unit", "b39-r-vendor-unit", "b39-r-duplicate-mscc", "b39-r-sms-result", "b39-r-wrong-sms-parent", "b39-a-unmatched-service", "b39-a-unmatched-number", "b39-a-unmatched-type", "b39-a-no-child-result", "b39-a-no-gsu", "b39-a-empty-gsu", "b39-a-top-failure-child-success", "b39-a-child-denied", "b39-a-empty-token", "b39-a-error", "b39-a-error-app-field"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		if m.Header.HopByHopID != 0xb1b2b3b5 || m.Header.EndToEndID != 0xc1c2c3c5 {
			panic("original IEC pair IDs")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;18;1"
			}
		}
		if !found {
			panic("session")
		}
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(out, raw) {
			panic("exact reencode")
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary Unsigned64/OctetString/Address types, paired header/session and exact reencode; counterexample retention is not protocol semantics or deployment certification\n", name, len(raw))
	}
}
