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
	if len(os.Args) != 3 {
		panic("arguments: absolute fixture directory and peer dictionary directory")
	}
	// The peer's generated default dictionary is stale for EPS location. Use
	// its existing XML sources without modifying or synthesizing dictionary entries.
	parser, err := dict.NewParser(filepath.Join(os.Args[2], "base.xml"), filepath.Join(os.Args[2], "tgpp_s6c.xml"), filepath.Join(os.Args[2], "tgpp_sgd_gdd.xml"))
	must(err)
	dict.Default = parser // Single-threaded standalone tool only.
	names := []string{"v-ofr-01", "v-ofa-01", "v-ofa-02", "v-tfr-01", "v-tfa-01", "v-tfa-02", "v-gdd-tfr-01", "v-gdd-tfa-01", "v-gdd-tfa-02", "v24-sgd-mo-r", "v24-sgd-mo-a", "v24-gdd-mo-r", "v24-gdd-mo-a"}
	for _, name := range names {
		raw := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(e)
		mt := strings.Contains(name, "tfr") || strings.Contains(name, "tfa")
		code := uint32(8388645)
		if mt {
			code = 8388646
		}
		if m.Header.ApplicationID != 16777313 || m.Header.CommandCode != code {
			panic("header")
		}
		u32(field(m.AVP, 277, 0), 1)
		request := m.Header.CommandFlags&0x80 != 0
		if request {
			identity := m.AVP
			if !mt {
				identity = group(field(m.AVP, 3102, 10415), 16777313)
			}
			text(field(identity, 1, 0), "001010123456789")
			if hex.EncodeToString(field(m.AVP, 3300, 10415).Data.Serialize()) != "2120550501f0" {
				panic("SC address")
			}
			tpdu := field(m.AVP, 3301, 10415).Data.Serialize()
			if len(tpdu) == 0 || len(tpdu) > 200 {
				panic("TPDU")
			}
			if mt {
				target := uint32(1645)
				if strings.Contains(name, "gdd") {
					target = 1489
				}
				a := field(m.AVP, target, 10415)
				if a.Flags&0x40 != 0 {
					panic("MT target M flag")
				}
				u32(field(m.AVP, 3306, 10415), 30)
				a = field(m.AVP, 3307, 10415)
				if _, ok := a.Data.(datatype.Time); !ok {
					panic("Time datatype")
				}
				u32(a, 0xed2723f0)
			}
			if name == "v-ofr-01" {
				eps := group(field(m.AVP, 1496, 10415), 16777313)
				mme := group(field(eps, 1600, 10415), 16777313)
				if hex.EncodeToString(field(mme, 1602, 10415).Data.Serialize()) != "00f11000000001" || hex.EncodeToString(field(mme, 1603, 10415).Data.Serialize()) != "00f1100001" {
					panic("ECGI/TAI")
				}
			}
		} else {
			if strings.HasSuffix(name, "02") {
				r := group(field(m.AVP, 297, 0), 16777313)
				u32(field(r, 266, 0), 10415)
				u32(field(r, 298, 0), 5555)
				cause := group(field(m.AVP, 3303, 10415), 16777313)
				value := uint32(4)
				if mt {
					value = 0
				}
				u32(field(cause, 3304, 10415), value)
			} else {
				u32(field(m.AVP, 268, 0), 2001)
			}
			reqName := map[string]string{"v-ofa-01": "v-ofr-01", "v-ofa-02": "v-ofr-01", "v-tfa-01": "v-tfr-01", "v-tfa-02": "v-tfr-01", "v-gdd-tfa-01": "v-gdd-tfr-01", "v-gdd-tfa-02": "v-gdd-tfr-01", "v24-sgd-mo-a": "v24-sgd-mo-r", "v24-gdd-mo-a": "v24-gdd-mo-r"}[name]
			req, e := diam.ReadMessage(bytes.NewReader(load(reqName)), dict.Default)
			must(e)
			if m.Header.HopByHopID != req.Header.HopByHopID || m.Header.EndToEndID != req.Header.EndToEndID {
				panic("association")
			}
			text(field(m.AVP, 263, 0), string(field(req.AVP, 263, 0).Data.Serialize()))
		}
		encoded, e := m.Serialize()
		must(e)
		if !bytes.Equal(encoded, raw) {
			panic("independent re-encode mismatch")
		}
		fmt.Printf("PASS %s (%d bytes): independent dictionary decode, selected values, exact re-encode\n", name, len(raw))
	}
}
