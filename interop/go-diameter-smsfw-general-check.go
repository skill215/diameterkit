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
	if v.VendorID == 0 {
		switch v.Code {
		case 447:
			if _, ok := v.Data.(datatype.Integer64); !ok {
				panic("digits")
			}
		case 429:
			if _, ok := v.Data.(datatype.Integer32); !ok {
				panic("exponent")
			}
		case 425:
			if _, ok := v.Data.(datatype.Unsigned32); !ok {
				panic("standard currency interpretation")
			}
		case 417:
			if _, ok := v.Data.(datatype.Unsigned64); !ok {
				panic("units")
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
	for _, name := range []string{"b44-general-request", "b44-general-enhanced", "b44-cca-success", "b44-cca-denied", "b44-cca-failed", "b44-cca-error", "b44-cca-cost-units", "b44-cca-cost-absent-exponent", "b44-cca-cost-zero", "b44-cca-cost-limits", "b44-cca-cost-unit", "b44-balance-request", "b44-balance-enough", "b44-balance-no-credit", "b44-balance-grant", "b44-balance-unknown", "b44-failed-cost", "b44-failed-balance", "b44-failed-grant", "b44-money-negative", "b44-currency-negative", "b44-raw-originator", "b44-nul-originator", "b44-string-address", "b44-ipv6-address", "b44-timezone-nonzero", "b44-service-nonliteral", "b44-refund-action", "b44-price-action", "b44-missing-money", "b44-missing-exponent", "b44-missing-currency", "b44-missing-timezone", "b44-missing-service", "b44-missing-client", "b44-missing-mms", "b44-wrong-timezone-vendor", "b44-duplicate-money", "b44-unexpected-token", "b44-cost-missing-unit-value", "b44-cost-missing-digits", "b44-cost-missing-currency", "b44-cost-missing-units", "b44-duplicate-cost", "b44-cost-wrong-parent", "b44-unknown-cost-child", "b44-cca-number"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		h, e := uint32(0x31323334), uint32(0x41424344)
		if m.Header.HopByHopID != h || m.Header.EndToEndID != e {
			panic("General pair identities")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;legacy-general;1"
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
		fmt.Printf("PASS %s (%d bytes): original dictionary exact reencode and paired header/session; dictionary field types are NOT legacy semantics, negative retention is NOT acceptance\n", name, len(raw))
	}
}
