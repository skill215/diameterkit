//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Independent offline dictionary verification, not an OCS financial oracle.
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
func check(a *diam.AVP) {
	switch a.Code {
	case 413, 437, 431, 445, 443, 423, 279:
		if _, ok := a.Data.(*diam.GroupedAVP); !ok {
			panic(fmt.Sprintf("group %d %T", a.Code, a.Data))
		}
	case 447:
		if _, ok := a.Data.(datatype.Integer64); !ok {
			panic("digits type")
		}
	case 429:
		if _, ok := a.Data.(datatype.Integer32); !ok {
			panic("exponent type")
		}
	case 425, 439, 268:
		if _, ok := a.Data.(datatype.Unsigned32); !ok {
			panic("uint32 type")
		}
	case 417:
		if _, ok := a.Data.(datatype.Unsigned64); !ok {
			panic("units type")
		}
	case 422, 436, 428:
		if _, ok := a.Data.(datatype.Enumerated); !ok {
			panic("enum type")
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
	names := []string{"a-balance-enough", "a-balance-no", "a-debit", "a-debit-zero", "a-debit-money", "a-refund", "a-refund-cost", "a-4010", "a-4011", "a-4012", "a-5011", "a-5012", "a-5030", "a-5031", "a-e"}
	for _, action := range []string{"balance", "debit", "refund"} {
		for _, suffix := range []string{"", "-zero", "-max", "-money", "-money-min", "-money-max"} {
			names = append(names, "r-"+action+suffix)
		}
	}
	for _, name := range names {
		raw, e := os.ReadFile(filepath.Join(os.Args[1], "b70-"+name+".hex"))
		must(e)
		raw, e = hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
		must(e)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 || m.Header.HopByHopID != 0xb1b2b370 || m.Header.EndToEndID != 0xc1c2c370 {
			panic("header")
		}
		for _, v := range m.AVP {
			check(v)
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(raw, out) {
			panic("exact reencode")
		}
		fmt.Println("PASS", name, len(raw), "bytes: dictionary field types and exact reencode;not service-specific or OCS certification")
	}
}
