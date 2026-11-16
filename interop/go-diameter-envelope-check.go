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
		panic("argument: absolute testdata/codec directory")
	}
	for _, name := range []string{"b2-srr", "e28-sra-3002", "e28-sra-5012", "b2-srr-proxy", "b2-sra-normal-proxy", "b2-sra-e-proxy", "b2-sra-e-5001"} {
		b := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(err)
		if m.Header.CommandCode != 8388647 || m.Header.ApplicationID != 16777312 {
			panic("S6c identity")
		}
		if name == "b2-srr" || strings.HasPrefix(name, "e28-") {
			text(field(m.AVP, 263, 0), "gmsc.example;25;32")
			if m.Header.HopByHopID != 0x25250020 || m.Header.EndToEndID != 0x35350020 {
				panic("pair IDs")
			}
		} else {
			text(field(m.AVP, 263, 0), "gmsc.example;17;1")
			if m.Header.HopByHopID != 0x51525354 || m.Header.EndToEndID != 0x61626364 {
				panic("pair IDs")
			}
		}
		isError := strings.HasPrefix(name, "e28-") || strings.Contains(name, "-sra-e-")
		if isError {
			if m.Header.CommandFlags != 0x60 {
				panic("E flags")
			}
			want := uint32(3002)
			if name == "e28-sra-5012" {
				want = 5012
			}
			if name == "b2-sra-e-5001" {
				want = 5001
			}
			u32(field(m.AVP, 268, 0), want)
		} else {
			u32(field(m.AVP, 277, 0), 1)
		}
		if strings.HasSuffix(name, "-proxy") {
			var proxies []*diam.AVP
			for _, a := range m.AVP {
				if a.Code == 284 && a.VendorID == 0 {
					proxies = append(proxies, a)
				}
			}
			if len(proxies) != 2 {
				panic("proxy count")
			}
			a := group(proxies[0], 16777312)
			text(field(a, 280, 0), "dra-a.example")
			if !bytes.Equal(field(a, 33, 0).Data.Serialize(), []byte{0, 255, 'A'}) {
				panic("binary state")
			}
			a = group(proxies[1], 16777312)
			text(field(a, 280, 0), "dra-b.example")
			if len(field(a, 33, 0).Data.Serialize()) != 0 {
				panic("present empty state")
			}
		}
		if name == "b2-sra-e-5001" {
			quote := group(field(m.AVP, 279, 0), 16777312)
			if len(quote) != 1 || quote[0].Code != 99999 {
				panic("Failed-AVP namespace")
			}
		}
		encoded, err := m.Serialize()
		must(err)
		if !bytes.Equal(encoded, b) {
			panic("independent re-encode mismatch")
		}
		fmt.Printf("PASS %s (%d bytes): independent field/envelope/Proxy decode and exact re-encode\n", name, len(b))
	}
	_, err := diam.ReadMessage(bytes.NewReader(load("b2-sra-e-5014-badlength")), dict.Default)
	if err != nil {
		fmt.Println("DIFF b2-sra-e-5014-badlength: go-diameter rejects malformed quoted AVP; diameterkit retains bounded diagnostic evidence")
	} else {
		fmt.Println("PASS b2-sra-e-5014-badlength: peer accepts bounded quoted evidence")
	}
}
