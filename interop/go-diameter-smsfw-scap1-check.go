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
	if v.VendorID == 193 {
		switch v.Code {
		case 553, 606, 612, 607:
			if _, ok := v.Data.(*diam.GroupedAVP); !ok {
				panic("SCAP group")
			}
		case 555, 611, 615, 608:
			if _, ok := v.Data.(datatype.Integer32); !ok {
				panic("SCAP signed field")
			}
		case 617:
			if _, ok := v.Data.(datatype.Unsigned64); !ok {
				panic("SCAP unit")
			}
		case 610:
			if _, ok := v.Data.(datatype.Time); !ok {
				panic("SCAP time")
			}
		case 554, 609:
			if _, ok := v.Data.(datatype.OctetString); !ok {
				panic("legacy string byte retention")
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
	if len(os.Args) != 3 {
		panic("absolute fixture directory and explicit offline overlay required")
	}
	raw := load("b45-acr")
	original, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OBSERVED original SCAP1 decode error: %v\n", err)
	} else {
		// Unknown Vendor193 octets must not be represented as typed private groups.
		for _, v := range original.AVP {
			if v.VendorID == 193 && v.Code == 607 {
				if _, ok := v.Data.(*diam.GroupedAVP); ok {
					panic("unexpected built-in SCAP definition")
				}
				fmt.Fprintf(os.Stderr, "OBSERVED original Vendor193/607 type: %T (no private group semantics)\n", v.Data)
				break
			}
		}
	}
	must(dict.Default.LoadFile(os.Args[2]))
	for _, name := range []string{"b45-acr", "b45-acr-envelope", "b45-aca-success", "b45-aca-failed", "b45-aca-diagnostic", "b45-aca-error", "b45-aca-application", "b45-time-zero", "b45-time-max", "b45-subscription-empty", "b45-subscription-raw", "b45-subscription-nul", "b45-missing-vsai", "b45-missing-acct", "b45-missing-subscription", "b45-missing-subscriber", "b45-missing-unit-type", "b45-missing-unit-value", "b45-missing-digits", "b45-missing-number", "b45-missing-record-type", "b45-missing-action", "b45-missing-timestamp", "b45-units-zero", "b45-units-max", "b45-wrong-unit-type", "b45-wrong-record-type", "b45-wrong-number", "b45-wrong-action", "b45-wrong-vendor-value", "b45-wrong-acct", "b45-auth-in-vsai", "b45-wrong-unit-vendor", "b45-standard-action", "b45-standard-time", "b45-unknown-group-child", "b45-unexpected-cc-type", "b45-param-empty", "b45-param-text", "b45-param-raw", "b45-param-nul", "b45-missing-param", "b45-extra-param", "b45-swapped-params", "b45-duplicate-param-type", "b45-origin-copy-mismatch", "b45-wrong-literal", "b45-wrong-timezone", "b45-missing-param-value", "b45-wrong-param-parent", "b45-aca-number", "b45-aca-record-type", "b45-aca-number-negative", "b45-aca-missing-number", "b45-aca-ro-field", "b45-aca-unselected-result"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 0 || m.Header.CommandCode != 271 {
			panic("header")
		}
		h, e := uint32(0x51525354), uint32(0x61626364)
		if m.Header.HopByHopID != h || m.Header.EndToEndID != e {
			panic("SCAP1 pair identities")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;legacy-scap1;1"
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
		fmt.Printf("PASS %s (%d bytes): explicit private dictionary exact reencode and paired header/session; dictionary field types are NOT legacy semantics, negative retention is NOT acceptance\n", name, len(raw))
	}
}
