//go:build ignore

// Independent dictionary decode/reencode evidence; no diameterkit imports.
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
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	for _, name := range []string{"asr", "asr-min", "asa-2001", "asa-5002", "asa-5012", "asa-options", "asa-e", "asa-e-diagnostic"} {
		raw, e := os.ReadFile(filepath.Join(os.Args[1], "b69-"+name+".hex"))
		must(e)
		raw, e = hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
		must(e)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 274 || m.Header.HopByHopID != 0xb1b2b369 || m.Header.EndToEndID != 0xc1c2c369 {
			panic("header")
		}
		seen := map[uint32]bool{}
		for _, v := range m.AVP {
			seen[v.Code] = true
			switch v.Code {
			case 263:
				if _, ok := v.Data.(datatype.UTF8String); !ok {
					panic("session type")
				}
				if string(v.Data.Serialize()) != "ctf.example;abort;69" {
					panic("session value")
				}
			case 258, 268, 278:
				if _, ok := v.Data.(datatype.Unsigned32); !ok {
					panic("uint32 type")
				}
			case 264, 296, 283, 293, 294, 282:
				if _, ok := v.Data.(datatype.DiameterIdentity); !ok {
					panic("identity type")
				}
			case 284, 279:
				if _, ok := v.Data.(*diam.GroupedAVP); !ok {
					panic("group type")
				}
			}
		}
		if !seen[263] || !seen[264] || !seen[296] {
			panic("missing envelope")
		}
		if strings.HasPrefix(name, "asr") {
			if !seen[258] || !seen[293] || !seen[283] {
				panic("ASR requirements")
			}
		} else {
			if !seen[268] || seen[258] || seen[415] || seen[416] {
				panic("ASA is not CCA")
			}
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(raw, out) {
			panic("exact reencode")
		}
		fmt.Println("PASS", name, len(raw), "bytes: independent dictionary types/header/session and exact reencode; not business or live OCS certification")
	}
}
