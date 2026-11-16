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
	if v.VendorID == 10415 && v.Code == 2014 {
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("status type")
		}
	}
	if v.VendorID == 10415 && v.Code == 2012 {
		if _, ok := v.Data.(datatype.Time); !ok {
			panic("discharge type")
		}
	}
	if v.Code == 417 && v.VendorID == 0 {
		if _, ok := v.Data.(datatype.Unsigned64); !ok {
			panic("unit type")
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
	for _, name := range []string{"b43-init-request", "b43-init-enhanced", "b43-init-a-2001", "b43-init-a-4012", "b43-init-a-5012", "b43-init-a-3004", "b43-term-request", "b43-term-enhanced", "b43-term-a-2001", "b43-term-a-5002", "b43-term-a-5012", "b43-term-a-3004", "b43-init-action", "b43-init-wrong-sequence", "b43-init-wrong-destination-interface", "b43-init-wrong-origin-interface", "b43-init-origin-sccp", "b43-init-wrong-imsi-copy", "b43-init-wrong-time-copy", "b43-term-action", "b43-term-wrong-sequence", "b43-term-wrong-destination-interface", "b43-term-wrong-origin-interface", "b43-term-origin-sccp", "b43-term-wrong-imsi-copy", "b43-term-wrong-time-copy", "b43-term-usage-zero", "b43-term-usage-max", "b43-term-no-usu", "b43-term-empty-usu", "b43-term-no-final", "b43-term-wrong-final", "b43-term-final-parent", "b43-term-rsu", "b43-init-usu", "b43-init-no-rsu", "b43-init-status", "b43-term-no-status", "b43-term-no-discharge", "b43-term-empty-status", "b43-term-zero-discharge", "b43-term-duplicate-status", "b43-term-token", "b43-init-a-gsu", "b43-term-a-gsu", "b43-term-a-wrong-sequence", "b43-term-sender-order"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		h, e := uint32(0x31323334), uint32(0x41424344)
		if strings.HasPrefix(name, "b43-term") {
			h, e = 0xb1b2b3b4, 0xc1c2c3c4
		}
		if m.Header.HopByHopID != h || m.Header.EndToEndID != e {
			panic("ECUR pair identities")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;legacy-ecur;1"
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
