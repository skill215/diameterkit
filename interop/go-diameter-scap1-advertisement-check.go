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
		panic("fixture directory")
	}
	for _, name := range []string{"strict-cer", "dual-cer", "strict-cea", "dual-cea"} {
		raw := load("b78-scap-" + name)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		count := 0
		for _, a := range m.AVP {
			if a.Code == 260 && a.VendorID == 0 {
				children := group(a, 0)
				u32(field(children, 266, 0), 193)
				u32(field(children, 259, 0), 19302)
				if count == 1 {
					u32(field(children, 258, 0), 19302)
				} else if len(children) != 2 {
					panic("strict shape")
				}
				count++
			}
		}
		want := 1
		if strings.HasPrefix(name, "dual") {
			want = 2
		}
		if count != want {
			panic("group count")
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(out, raw) {
			panic("changed wire")
		}
		fmt.Println(name, "typed VSAI and exact roundtrip; dictionary acceptance is not RFC semantic validation")
	}
}
