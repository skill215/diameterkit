//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Independent offline dictionary evidence; no diameterkit imports.
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

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func load(name string) []byte {
	b, e := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(e)
	raw, e := hex.DecodeString(strings.Join(strings.Fields(string(b)), ""))
	must(e)
	return raw
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute testdata/codec directory required")
	}
	original := load("v25-t4-ofr-r")
	for _, name := range []string{"v25-t4-ofr-r", "v25-t4-ofa-a", "f25-t4-ofa-5555", "n25-t4-noext", "n25-t4-nested", "n25-t4-nai", "n25-t4-wrongapp", "n25-l092-text", "n25-l092-dup", "n26-ofa-vendor-shadow", "b25-t4-ofa-m0", "b25-t4-fail-m0"} {
		b := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		if name == "n25-t4-wrongapp" {
			if e == nil || e.Error() != "could not find preloaded Command with code 8388645" {
				panic(fmt.Sprintf("wrong-app behavior: %v", e))
			}
			if binary.BigEndian.Uint32(b[8:12]) != 16777311 {
				panic("wrong-app fixture")
			}
			fmt.Printf("PASS %s (%d bytes): original dictionary rejected absent command8388645 in app16777311; no reencode or SCTP claim\n", name, len(b))
			continue
		}
		must(e)
		app := uint32(16777313)
		if name == "n25-t4-wrongapp" {
			app = 16777311
		}
		if m.Header.ApplicationID != app || m.Header.CommandCode != 8388645 {
			panic("header")
		}
		if m.Header.HopByHopID != binary.BigEndian.Uint32(original[12:16]) || m.Header.EndToEndID != binary.BigEndian.Uint32(original[16:20]) {
			panic("pair IDs")
		}
		count := 0
		for _, v := range m.AVP {
			if v.Code != 3111 || v.VendorID != 10415 {
				continue
			}
			count++
			if _, ok := v.Data.(datatype.UTF8String); !ok {
				panic(fmt.Sprintf("3111 type: %T", v.Data))
			}
			flags := byte(0xc0)
			if strings.HasPrefix(name, "b25-") {
				flags = 0x80
			}
			if v.Flags != flags {
				panic("flags")
			}
			// Old dictionary interprets UTF8String without validating NAI or UTF8 bytes.
			want := "device-001@example.net"
			switch name {
			case "n25-t4-nai":
				want = "@example.net"
			case "n25-l092-text":
				want = string([]byte{0xc3, 0x28})
			}
			if string(v.Data.Serialize()) != want {
				panic(fmt.Sprintf("3111 bytes %s: %x", name, v.Data.Serialize()))
			}
		}
		wantCount := 1
		switch name {
		case "v25-t4-ofr-r", "n25-t4-noext", "n25-t4-nested", "n26-ofa-vendor-shadow":
			wantCount = 0
		case "n25-l092-dup":
			wantCount = 2
		}
		if count != wantCount {
			panic("presence")
		}
		// Serialize before another ReadMessage reuses the peer's read buffer.
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(b, out) {
			panic("exact reencode " + name)
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary UTF8String/flags, header IDs and exact reencode; controls are not semantic validation or SCTP evidence\n", name, len(b))
	}
}
