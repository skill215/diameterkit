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
		panic("argument: absolute fixture directory")
	}
	names := []string{"v-sh-udr-01", "v-sh-uda-01", "v-sh-uda-02", "v-sh-pur-01", "v-sh-pua-01", "v-sh-pua-02", "v-sh-snr-01", "v-sh-sna-01", "v-sh-sna-02", "v-sh-pnr-01", "v-sh-pna-01", "v-sh-pna-02"}
	for _, name := range names {
		raw := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		kind := name[5:7]
		command := map[string]uint32{"ud": 306, "pu": 307, "sn": 308, "pn": 309}[kind]
		if m.Header.ApplicationID != 16777217 || m.Header.CommandCode != command {
			panic("Sh application/command")
		}
		vsai := group(field(m.AVP, 260, 0), 16777217)
		u32(field(vsai, 266, 0), 10415)
		u32(field(vsai, 258, 0), 16777217)
		u32(field(m.AVP, 277, 0), 1)
		request := m.Header.CommandFlags&0x80 != 0
		if request {
			identity := group(field(m.AVP, 700, 10415), 16777217)
			text(field(identity, 601, 10415), "sip:alice@example.net")
			if command != 309 {
				ref := uint32(24)
				if command == 308 {
					ref = 25
				}
				a := field(m.AVP, 703, 10415)
				if _, ok := a.Data.(datatype.Enumerated); !ok {
					panic("Data-Reference not Enumerated")
				}
				u32(a, ref)
			}
			if command == 308 {
				u32(field(m.AVP, 705, 10415), 0)
				a := field(m.AVP, 712, 10415)
				u32(a, 0)
				if a.Flags&0x40 != 0 {
					panic("One-Time M")
				}
				a = field(m.AVP, 709, 10415)
				if _, ok := a.Data.(datatype.Time); !ok {
					panic("Expiry not Time")
				}
				u32(a, 0xee90ff80)
				if a.Flags&0x40 != 0 {
					panic("Expiry M")
				}
			}
			if command == 309 {
				text(field(m.AVP, 293, 0), "as.example")
			}
		} else {
			result := uint32(2001)
			if strings.HasSuffix(name, "02") {
				result = map[string]uint32{"ud": 5102, "pu": 5103, "sn": 5104, "pn": 5100}[kind]
				ex := group(field(m.AVP, 297, 0), 16777217)
				u32(field(ex, 266, 0), 10415)
				u32(field(ex, 298, 0), result)
			} else {
				u32(field(m.AVP, 268, 0), result)
			}
			req, e := diam.ReadMessage(bytes.NewReader(load("v-sh-"+kind+"r-01")), dict.Default)
			must(e)
			if req.Header.HopByHopID != m.Header.HopByHopID || req.Header.EndToEndID != m.Header.EndToEndID {
				panic("pair IDs")
			}
			text(field(m.AVP, 263, 0), string(field(req.AVP, 263, 0).Data.Serialize()))
		}
		if name == "v-sh-uda-01" || name == "v-sh-pur-01" || name == "v-sh-pnr-01" {
			a := field(m.AVP, 702, 10415)
			if _, ok := a.Data.(datatype.OctetString); !ok {
				panic("User-Data not OctetString")
			}
			b := a.Data.Serialize()
			if !bytes.HasPrefix(b, []byte("<Sh-Data>")) {
				panic("opaque fixture bytes")
			}
		}
		encoded, e := m.Serialize()
		must(e)
		if !bytes.Equal(encoded, raw) {
			panic("independent exact re-encode")
		}
		fmt.Printf("PASS %s (%d bytes): independent Sh dictionary decode, selected values, exact re-encode; User-Data opaque\n", name, len(raw))
	}
}
