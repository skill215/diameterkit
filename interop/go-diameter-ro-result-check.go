//go:build ignore

// Independent offline reencoding with the original go-diameter dictionary.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
	"os"
	"path/filepath"
	"strings"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func check(v *diam.AVP) {
	if v.VendorID == 10415 {
		switch v.Code {
		case 628:
			if _, ok := v.Data.(*diam.GroupedAVP); !ok {
				panic("features group")
			}
		case 629, 630:
			if _, ok := v.Data.(datatype.Unsigned32); !ok {
				panic("features u32")
			}
		}
	}
	if v.VendorID == 0 {
		switch v.Code {
		case 456, 423, 445, 279, 297:
			if _, ok := v.Data.(*diam.GroupedAVP); !ok {
				panic("group type")
			}
		case 268, 425, 266, 298:
			if _, ok := v.Data.(datatype.Unsigned32); !ok {
				panic("unsigned32")
			}
		case 428:
			if _, ok := v.Data.(datatype.Enumerated); !ok {
				panic("DDFH Enumerated")
			}
		case 447:
			if _, ok := v.Data.(datatype.Integer64); !ok {
				panic("signed64")
			}
		case 429:
			if _, ok := v.Data.(datatype.Integer32); !ok {
				panic("signed32")
			}
		}
	}
	if g, ok := v.Data.(*diam.GroupedAVP); ok {
		for _, child := range g.AVP {
			check(child)
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	files, e := filepath.Glob(filepath.Join(os.Args[1], "b49-*.hex"))
	must(e)
	if len(files) != 65 {
		panic("fixture count")
	}
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".hex")
		if name == "b49-exp-width" || name == "b49-rating-quoted-incomplete" {
			continue
		} // deliberately malformed Uint32; local negative only
		text, e := os.ReadFile(f)
		must(e)
		raw, e := hex.DecodeString(strings.Join(strings.Fields(string(text)), ""))
		must(e)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 || m.Header.CommandFlags&128 != 0 {
			panic("answer header")
		}
		for _, v := range m.AVP {
			check(v)
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(raw, out) {
			panic("exact reencode: " + name)
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary; result namespaces/diagnostics and byte preservation, not branch or OCS acceptance\n", name, len(raw))
	}
}
