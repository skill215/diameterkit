//go:build ignore

// Standalone evidence tool. Run from a disposable module; never a core dependency.
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
func field(as []*diam.AVP, code, vendor uint32) *diam.AVP {
	var found *diam.AVP
	for _, a := range as {
		if a.Code == code && a.VendorID == vendor {
			if found != nil {
				panic("duplicate field")
			}
			found = a
		}
	}
	if found == nil {
		panic(fmt.Sprintf("missing %d/%d", vendor, code))
	}
	return found
}
func u32(a *diam.AVP, want uint32) {
	b := a.Data.Serialize()
	if len(b) != 4 || binary.BigEndian.Uint32(b) != want {
		panic("u32 mismatch")
	}
}
func text(a *diam.AVP, want string) {
	if string(a.Data.Serialize()) != want {
		panic("text mismatch")
	}
}
func group(a *diam.AVP, app uint32) []*diam.AVP {
	if g, ok := a.Data.(*diam.GroupedAVP); ok {
		return g.AVP
	}
	g, ok := a.Data.(datatype.Grouped)
	if !ok {
		panic(fmt.Sprintf("not a dictionary Grouped: %T", a.Data))
	}
	parsed, err := diam.DecodeGrouped(g, app, dict.Default)
	must(err)
	return parsed.AVP
}
func main() {
	if len(os.Args) != 2 {
		panic("argument: absolute testdata/codec directory")
	}
	for _, name := range []string{"base-dwr", "base-dwa", "v-srr-01", "v-sra-01", "v-sra-02"} {
		raw, err := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
		must(err)
		b, err := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
		must(err)
		m, err := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(err)
		if m.Header.HopByHopID != 0x51525354 || m.Header.EndToEndID != 0x61626364 {
			panic("pair IDs")
		}
		if strings.HasPrefix(name, "base-") {
			if m.Header.CommandCode != 280 || m.Header.ApplicationID != 0 {
				panic("base identity")
			}
			u32(field(m.AVP, 278, 0), 0)
			if name == "base-dwa" {
				u32(field(m.AVP, 268, 0), 2001)
			}
		} else {
			if m.Header.CommandCode != 8388647 || m.Header.ApplicationID != 16777312 {
				panic("S6c identity")
			}
			text(field(m.AVP, 263, 0), "gmsc.example;17;1")
			u32(field(m.AVP, 277, 0), 1)
			switch name {
			case "v-srr-01":
				u32(field(m.AVP, 3308, 10415), 0)
				if hex.EncodeToString(field(m.AVP, 701, 10415).Data.Serialize()) != "2120550521f3" {
					panic("MSISDN")
				}
			case "v-sra-01":
				u32(field(m.AVP, 268, 0), 2001)
				text(field(m.AVP, 1, 0), "001010123456789")
				g := group(field(m.AVP, 2401, 10415), 16777312)
				text(field(g, 2402, 10415), "mme.example")
				text(field(g, 2408, 10415), "example.net")
				if hex.EncodeToString(field(g, 1645, 10415).Data.Serialize()) != "2120550599f9" {
					panic("MME number")
				}
			case "v-sra-02":
				g := group(field(m.AVP, 297, 0), 16777312)
				u32(field(g, 266, 0), 10415)
				u32(field(g, 298, 0), 5001)
			}
		}
		encoded, err := m.Serialize()
		must(err)
		if !bytes.Equal(encoded, b) {
			panic("independent re-encode mismatch")
		}
		// Compare malformed scalar handling; this is not a conformance oracle.
		if name == "v-srr-01" {
			bad := append([]byte(nil), b...)
			bad[55] = 11
			if _, err = diam.ReadMessage(bytes.NewReader(bad), dict.Default); err == nil {
				fmt.Println("DIFF v-srr-01/Auth-Session-State width=3: go-diameter accepts; diameterkit regression requires rejection")
			}
		}
		fmt.Printf("PASS %s (%d bytes): independent dictionary decode, selected values, exact re-encode\n", name, len(b))
	}
}
