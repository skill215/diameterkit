//go:build ignore

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
	case 443, 456, 437, 431, 446, 458, 873, 874, 2000, 2009, 2026, 2002:
		if _, ok := v.Data.(*diam.GroupedAVP); !ok {
			panic(fmt.Sprintf("group type %d/%d: %T", v.VendorID, v.Code, v.Data))
		}
	case 872:
		if _, ok := v.Data.(datatype.Enumerated); !ok {
			panic("Reporting-Reason type")
		}
	case 21, 22, 23, 460:
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic(fmt.Sprintf("octets type %d: %T", v.Code, v.Data))
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
	for _, name := range []string{"v-sms-ccr-01", "v-sms-cca-01", "v-sms-cca-02", "v-term-ccr-01", "v-term-cca-01", "v-term-cca-02", "b40-term-zero", "b40-term-max", "b40-term-no-usu", "b40-term-empty-usu", "b40-term-dup-usu", "b40-term-dup-unit", "b40-term-no-reason", "b40-term-threshold", "b40-term-dup-reason", "b40-term-usu-reason", "b40-term-top-reason", "b40-term-rsu", "b40-init-usu", "b40-init-action", "b40-init-wrong-number", "b40-term-action", "b40-term-wrong-number", "b40-init-no-rsu", "b40-init-empty-rsu", "b40-init-update", "b40-init-a-no-gsu", "b40-term-a-gsu", "b40-term-a-mscc", "b40-term-a-wrong-number", "b40-term-a-refund", "b40-init-imei15", "b40-init-equipment-type", "b40-init-rat-width", "b40-init-rat-nr", "b40-init-uli-short", "b40-init-uli-plmn", "b40-init-uli-ecgi", "b40-init-timezone-spare", "b40-init-timezone-reserved", "b40-init-nested-equipment", "b40-term-sms-result", "b40-init-a-error", "b40-term-a-error"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		h, e := uint32(0x31323334), uint32(0x41424344)
		if strings.Contains(name, "term") {
			h, e = 0xb1b2b3b4, 0xc1c2c3c4
		}
		if m.Header.HopByHopID != h || m.Header.EndToEndID != e {
			panic("ECUR pair identities")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;17;1"
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
