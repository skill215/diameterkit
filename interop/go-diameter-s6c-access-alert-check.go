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
	for _, name := range []string{"v-alr-move-01", "v-ala-move-01", "v-ala-move-02", "v21-sgsn-alr-r", "v21-sgsn-alr-a", "v21-sgsn-alr-f", "v21-router-alr-r", "b23-alr-reachable", "b23-alr-move-nonode", "b23-alr-both", "b23-alr-future-bits", "b23-alr-zero", "b23-alr-noevent", "b23-alr-reachable-node"} {
		b := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(e)
		if m.Header.CommandCode != 8388648 || m.Header.ApplicationID != 16777312 {
			panic("wrong app/command")
		}
		if m.Header.CommandFlags&0x80 != 0 {
			g := group(field(m.AVP, 3102, 10415), 16777312)
			id := field(g, 1, 0)
			if _, ok := id.Data.(datatype.UTF8String); !ok {
				panic("IMSI type")
			}
			want := "001010123456789"
			if name == "v21-router-alr-r" {
				want = "router-alert-21"
			}
			text(id, want)
			eventCount, nodeCount := 0, 0
			for _, v := range m.AVP {
				if v.Code == 3333 && v.VendorID == 10415 {
					eventCount++
					if v.Flags != 0x80 {
						panic("event M bit")
					}
					if _, ok := v.Data.(datatype.Unsigned32); !ok {
						panic("event not bitmap")
					}
					event := uint32(2)
					switch name {
					case "b23-alr-both":
						event = 3
					case "b23-alr-reachable", "b23-alr-reachable-node":
						event = 1
					case "b23-alr-future-bits":
						event = 0x80000002
					case "b23-alr-zero":
						event = 0
					}
					u32(v, event)
				}
				if v.Code == 2401 && v.VendorID == 10415 {
					nodeCount++
					if v.Flags != 0xc0 {
						panic("node flags")
					}
					g := group(v, 16777312)
					if len(g) != 3 {
						panic("node leaf count")
					}
					for _, leaf := range g {
						if leaf.VendorID != 10415 {
							panic("node vendor")
						}
						switch leaf.Code {
						case 2402, 2408, 2409, 2410:
							if _, ok := leaf.Data.(datatype.DiameterIdentity); !ok {
								panic("node identity type")
							}
							flags := byte(0xc0)
							if leaf.Code == 2409 || leaf.Code == 2410 {
								flags = 0x80
							}
							if leaf.Flags != flags {
								panic("node identity flags")
							}
						case 1645, 1489:
							if _, ok := leaf.Data.(datatype.OctetString); !ok {
								panic("node number type")
							}
							if leaf.Flags != 0xc0 {
								panic("number flags")
							}
						default:
							panic("unexpected node leaf")
						}
					}
				}
			}
			wantEvent := 1
			if name == "b23-alr-noevent" {
				wantEvent = 0
			}
			wantNode := 1
			switch name {
			case "b23-alr-reachable", "b23-alr-move-nonode", "b23-alr-zero", "b23-alr-noevent":
				wantNode = 0
			}
			if eventCount != wantEvent || nodeCount != wantNode {
				panic("event/node presence")
			}
		} else {
			result := uint32(2001)
			if m.Header.CommandFlags&0x20 != 0 {
				result = 3004
			}
			u32(field(m.AVP, 268, 0), result)
			source := "v-alr-move-01"
			if strings.HasPrefix(name, "v21-") {
				source = "v21-sgsn-alr-r"
			}
			r, e := diam.ReadMessage(bytes.NewReader(load(source)), dict.Default)
			must(e)
			if r.Header.HopByHopID != m.Header.HopByHopID || r.Header.EndToEndID != m.Header.EndToEndID {
				panic("answer IDs")
			}
			text(field(m.AVP, 263, 0), string(field(r.AVP, 263, 0).Data.Serialize()))
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(out, b) {
			panic("exact reencode mismatch")
		}
		fmt.Printf("PASS %s (%d bytes): original dictionary typed identity/event/node and exact reencode; no role authorization or event-semantics claim\n", name, len(b))
	}
}
