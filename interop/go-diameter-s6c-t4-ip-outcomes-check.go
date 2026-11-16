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
	case 1:
		if v.VendorID != 0 {
			panic("IMSI vendor")
		}
		if _, ok := v.Data.(datatype.UTF8String); !ok {
			panic("IMSI type")
		}
	case 3102, 3316, 3317, 3318, 3319, 3320:
		if v.VendorID != 10415 {
			panic("group vendor")
		}
		g, ok := v.Data.(*diam.GroupedAVP)
		if !ok {
			panic("group type")
		}
		for _, child := range g.AVP {
			check(child)
		}
	case 3321:
		if v.VendorID == 10415 {
			if _, ok := v.Data.(datatype.Enumerated); !ok {
				panic(fmt.Sprintf("cause type %T", v.Data))
			}
		} else {
			if _, ok := v.Data.(datatype.Unknown); !ok {
				panic("shadow type")
			}
		}
	case 3322:
		if _, ok := v.Data.(datatype.Unsigned32); !ok {
			panic("diagnostic type")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"b38-rdr-capacity", "b38-rdr-absent", "b38-rdr-diag11", "b38-rdr-diag12", "b38-rdr-success", "b38-rdr-nocause", "b38-rdr-diag0", "b38-rdr-diag13", "b38-rdr-cause3", "b38-rdr-diag256", "b38-rdr-success-diag", "b38-rdr-mme-ip", "b38-rdr-msc-ip", "b38-rdr-sgsn-ip", "b38-rdr-mme-sgsn-ip", "b38-rdr-msc-sgsn-ip", "b38-rdr-dupgroup", "b38-rdr-vendorcause", "b38-rdr-parentdiag", "b38-rdr-borrow-cause", "d21-t4-rdr-r", "d21-t4-rdr-a", "d21-t4-rdr-f"} {
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388649 {
			panic("header")
		}
		if m.Header.HopByHopID != 0x2121000b || m.Header.EndToEndID != 0x3131000b {
			panic("original T4 pair IDs")
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "isg.example;21;11"
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
		fmt.Printf("PASS %s (%d bytes): original dictionary grouped/Enumerated/U32 types, paired header/session and exact reencode; counterexample retention is not protocol semantics or deployment certification\n", name, len(raw))
	}
}
