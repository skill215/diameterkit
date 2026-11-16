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
func main() {
	if len(os.Args) != 2 {
		panic("argument: fixture directory")
	}
	names, err := filepath.Glob(filepath.Join(os.Args[1], "b76-routing-*.hex"))
	must(err)
	if len(names) != 27 {
		panic("fixture count")
	}
	for _, path := range names {
		raw := load(strings.TrimSuffix(filepath.Base(path), ".hex"))
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		if strings.Contains(path, "short-purpose-r") {
			must(err)
			// The frozen reference returns zero without error for a short Integer32.
			u32(field(m.AVP, 3311, 10415), 0)
			out, err := m.Serialize()
			must(err)
			if bytes.Equal(raw, out) {
				panic("expected malformed-width rewrite")
			}
			fmt.Println(filepath.Base(path), "reference maps malformed enum to zero and rewrites wire; diameterkit rejects three-byte payload")
			continue
		}
		must(err)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388647 {
			panic("wrong binding")
		}
		for _, a := range m.AVP {
			if a.Code == 1 && a.VendorID == 0 {
				if _, ok := a.Data.(datatype.UTF8String); !ok {
					panic("identity type")
				}
			}
			if a.Code == 3311 && a.VendorID == 10415 {
				if _, ok := a.Data.(datatype.Enumerated); !ok {
					panic("purpose type")
				}
			}
		}
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, out) {
			panic("wire changed")
		}
		fmt.Println(filepath.Base(path), len(raw), "exact purpose/identity decode/reencode; not conditional validation")
	}
}
