//go:build ignore

// Independent offline dictionary evidence, without diameterkit imports.
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
	case 297, 3102:
		g, ok := v.Data.(*diam.GroupedAVP)
		if !ok {
			panic("group type")
		}
		want := uint32(0)
		if v.Code == 3102 {
			want = 10415
		}
		if v.VendorID != want {
			panic("group vendor")
		}
		for _, child := range g.AVP {
			check(child)
		}
	case 266, 298, 268:
		if _, ok := v.Data.(datatype.Unsigned32); !ok {
			panic("result type")
		}
		if v.VendorID != 0 {
			panic("result child vendor")
		}
	case 701, 3300, 3301:
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("number/TPDU type")
		}
		if v.VendorID != 10415 {
			panic("number vendor")
		}
	case 1:
		if _, ok := v.Data.(datatype.UTF8String); !ok {
			panic("IMSI type")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"f24-sgd-mo-5001", "c24-gdd-mo-unknown", "b29-mo-nomsisdn", "b29-mo-othermsisdn", "b29-ofa-origin", "b29-ofa-vendor", "b29-ofa-dupresult", "b29-ofa-dualresult"} {
		raw := load(name)
		original := load("v24-sgd-mo-r")
		session := "mme.example;24;11"
		if name == "c24-gdd-mo-unknown" {
			original = load("v24-gdd-mo-r")
			session = "sgsn.example;24;12"
		}
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 16777313 || m.Header.CommandCode != 8388645 || m.Header.HopByHopID != binary.BigEndian.Uint32(original[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(original[16:20]) {
			panic("header pair")
		}
		found := false
		results := 0
		for _, v := range m.AVP {
			check(v)
			if v.Code == 263 {
				found = true
				if string(v.Data.Serialize()) != session {
					panic("session")
				}
			}
			if v.Code == 297 {
				results++
				g := v.Data.(*diam.GroupedAVP)
				for _, child := range g.AVP {
					want := uint32(5001)
					if child.Code == 266 {
						want = 10415
						if name == "b29-ofa-vendor" {
							want = 0
						}
					}
					if binary.BigEndian.Uint32(child.Data.Serialize()) != want {
						panic("experimental tuple")
					}
				}
			}
		}
		want := 1
		if strings.HasPrefix(name, "b29-mo-") {
			want = 0
		}
		if name == "b29-ofa-dupresult" {
			want = 2
		}
		if !found || results != want {
			panic("presence")
		}
		// Serialize before another ReadMessage can reuse the old decoder's buffer.
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, out) {
			panic("exact reencode")
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary types, experimental tuple, paired header/session and exact reencode; no interface/source/identity semantics or deployment certification\n", name, len(raw))
	}
}
