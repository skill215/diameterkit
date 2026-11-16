//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

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
	if v.VendorID == 0 {
		switch v.Code {
		case 456, 423, 445:
			if _, ok := v.Data.(*diam.GroupedAVP); !ok {
				panic("group type")
			}
		case 268, 425:
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
	files, e := filepath.Glob(filepath.Join(os.Args[1], "b48-*.hex"))
	must(e)
	if len(files) != 67 {
		panic("fixture count")
	}
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".hex")
		if name == "b48-child-width" {
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
		fmt.Printf("PASS %s (%d bytes): original dictionary; layered results/cost/DDFH type and byte preservation, not branch or OCS acceptance\n", name, len(raw))
	}
}
