//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Independent dictionary verification. This is not an OCS financial oracle.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func check(a *diam.AVP) {
	switch a.Code {
	case 456, 431:
		if _, ok := a.Data.(*diam.GroupedAVP); !ok {
			panic(fmt.Sprintf("group %d %T", a.Code, a.Data))
		}
	case 439, 268:
		if _, ok := a.Data.(datatype.Unsigned32); !ok {
			panic("uint32 type")
		}
	case 417:
		if _, ok := a.Data.(datatype.Unsigned64); !ok {
			panic("uint64 type")
		}
	case 2022:
		if a.VendorID != 10415 {
			panic("token vendor")
		}
		if _, ok := a.Data.(datatype.OctetString); !ok {
			panic("token type")
		}
	}
	if g, ok := a.Data.(*diam.GroupedAVP); ok {
		for _, child := range g.AVP {
			check(child)
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute fixture directory required")
	}
	manifest, err := os.ReadFile(filepath.Join(os.Args[1], "manifest.json"))
	must(err)
	var entries []struct{ File, Expect string }
	must(json.Unmarshal(manifest, &entries))
	for _, entry := range entries {
		if !strings.HasPrefix(entry.File, "b71-") || entry.Expect != "pair" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(os.Args[1], entry.File))
		must(err)
		raw, err = hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
		must(err)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
			panic("header")
		}
		for _, a := range m.AVP {
			check(a)
		}
		out, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, out) {
			panic("exact reencode")
		}
		fmt.Println("PASS", entry.File, len(raw), "bytes: dictionary field types and exact reencode;not financial or targetOCS certification")
	}
}
