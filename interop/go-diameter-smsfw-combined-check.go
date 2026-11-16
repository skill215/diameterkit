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
	if v.VendorID == 10415 && v.Code == 2018 {
		if _, ok := v.Data.(datatype.OctetString); !ok {
			panic("2018 not opaque wire")
		}
	}
	if v.VendorID == 0 {
		switch v.Code {
		case 420:
			if _, ok := v.Data.(*diam.GroupedAVP); !ok {
				panic("420 not grouped")
			}
		case 447:
			if _, ok := v.Data.(datatype.Integer64); !ok {
				panic("digits not signed64")
			}
		case 429, 425:
			if _, ok := v.Data.(datatype.Integer32); !ok {
				panic("signed32 lost")
			}
		case 417:
			if _, ok := v.Data.(datatype.Unsigned64); !ok {
				panic("units not uint64")
			}
		case 1292:
			if _, ok := v.Data.(datatype.OctetString); !ok {
				panic("token not octets")
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
		panic("absolute fixture directory and offline overlay XML required")
	}
	// Observe the original standard dictionary before any explicit override.
	raw := load("b42-refund-request")
	m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
	if err != nil {
		fmt.Fprintf(os.Stderr, "OBSERVED original standard App4 dictionary rejects refund420: %v\n", err)
	} else {
		before, e := m.Serialize()
		must(e)
		if bytes.Equal(before, raw) {
			panic("expected standard420 conflict was not observed")
		}
		fmt.Fprintf(os.Stderr, "OBSERVED original standard App4 dictionary changes refund420: %d -> %d bytes\n", len(raw), len(before))
	}
	must(dict.Default.LoadFile(os.Args[2])) // process-local, no third-party file edits
	files, err := filepath.Glob(filepath.Join(os.Args[1], "b47-*.hex"))
	must(err)
	if len(files) != 36 {
		panic("fixture count")
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".hex")
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}

		for _, v := range m.AVP {
			check(v)
		}

		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(out, raw) {
			panic("exact reencode")
		}
		fmt.Printf("PASS %s (%d bytes): explicit offline dictionary override420/425/1292/2018; fixed wire types and exact reencode; not path-rule or OCS acceptance\n", name, len(raw))
	}
}
