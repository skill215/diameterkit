//go:build ignore

// Standalone offline dictionary evidence; no diameterkit imports.
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
	case 3316, 3317, 3318, 3319, 3320:
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
	for _, suffix := range []string{"capacity", "absent", "diag11", "diag12", "success", "msc-sgsn-ip", "mme-ip", "nocause", "diag0", "diag13", "cause3", "diag256", "success-diag", "dupgroup", "vendorcause", "parentdiag"} {
		name := "b31-rdr-" + suffix
		raw := load(name)
		answers := [][]byte{load("v24-s6c-rd-a"), load("f24-s6c-rd-5001"), load("f24-s6c-rd-5558")}
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388649 {
			panic("header")
		}
		for _, a := range answers {
			if m.Header.HopByHopID != binary.BigEndian.Uint32(a[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(a[16:20]) {
				panic("pair IDs")
			}
		}
		found := false
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = string(v.Data.Serialize()) == "gmsc.example;24;6"
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
