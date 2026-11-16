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
	for _, name := range []string{"b41-iec-request", "b41-iec-enhanced", "b41-cca-success", "b41-cca-denied", "b41-cca-failed", "b41-cca-error", "b41-cca-grant-ddfh", "b41-cca-ddfh-zero", "b41-signed-min", "b41-signed-max", "b41-string-nul", "b41-string-nonutf8", "b41-string-empty", "b41-wrong-ue", "b41-wrong-time", "b41-wrong-smsc", "b41-wrong-node", "b41-wrong-action", "b41-wrong-units", "b41-missing-udh", "b41-duplicate-interface", "b41-wrong-vendor", "b41-wrong-parent", "b41-420-scalar", "b41-token1292", "b41-cca-psp", "b41-cca-number", "b41-cca-ddfh-unknown", "b41-subscriptions-reversed"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		h, e := uint32(0x31323334), uint32(0x41424344)
		if m.Header.HopByHopID != h || m.Header.EndToEndID != e {
			panic("IEC pair identities")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "ctf.example;legacy-iec;1"
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
