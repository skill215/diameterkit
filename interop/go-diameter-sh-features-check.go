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

func load(name string) []byte {
	raw, err := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(err)
	b, err := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
	must(err)
	return b
}
func absent(as []*diam.AVP, code, vendor uint32) {
	for _, a := range as {
		if a.Code == code && a.VendorID == vendor {
			panic("unexpected field")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("fixture directory required")
	}
	names := []string{"v25-sh-need-r", "v25-sh-need-a", "f25-sh-5011", "f25-sh-base5001", "v25-sh-partial-r", "v25-sh-partial-a", "f25-sh-partial-5011", "v25-sh-offer-r", "v25-sh-offer-a", "v25-sh-plain-r", "v25-sh-plain-a", "v25-mix-snr-r", "v25-mix-snr-a", "f25-mix-snr", "v25-mix-pnr-r", "v25-mix-pnr-a", "f25-mix-pnr", "b20-sh-partial-r", "b20-sh-partial-a", "b20-sh-partial-f"}
	for _, name := range names {
		raw := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 16777217 {
			panic("application")
		}
		for _, a := range m.AVP {
			if a.Code == 628 && a.VendorID == 10415 {
				children := group(a, 16777217)
				u32(field(children, 266, 0), 10415)
				u32(field(children, 629, 10415), 1)
				bits := field(children, 630, 10415)
				if _, ok := bits.Data.(datatype.Unsigned32); !ok {
					panic("bitmap type")
				}
				if m.Header.CommandFlags&128 == 0 && a.Flags&64 != 0 {
					panic("answer M")
				}
			}
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(out, raw) {
			panic("exact reencode")
		}
		fmt.Printf("PASS %s (%d bytes): built-in Sh dictionary, typed feature children and outer flags, exact re-encode; no negotiation policy or XML claim\n", name, len(raw))
	}
}
