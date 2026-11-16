//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

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
	names, err := filepath.Glob(filepath.Join(os.Args[1], "b75-router-mt-*.hex"))
	must(err)
	if len(names) != 30 {
		panic("fixture count")
	}
	for _, path := range names {
		raw := load(strings.TrimSuffix(filepath.Base(path), ".hex"))
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 16777313 || m.Header.CommandCode != 8388646 {
			panic("wrong binding")
		}
		for _, a := range m.AVP {
			if a.Code == 3102 && a.VendorID == 10415 {
				for _, child := range group(a, m.Header.ApplicationID) {
					if child.Code == 1 && child.VendorID == 0 {
						if _, ok := child.Data.(datatype.UTF8String); !ok {
							panic("token type")
						}
					}
				}
			}
		}
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, out) {
			panic("wire changed: " + path)
		}
		fmt.Println(filepath.Base(path), len(raw), "exact Router TFR/TFA dictionary decode/reencode; not conditional validation")
	}
}
