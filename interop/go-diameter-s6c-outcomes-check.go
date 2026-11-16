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
func check(name string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%v", v)
		}
	}()
	b := load(name)
	m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
	must(e)
	if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388649 || m.Header.HopByHopID != 0x24240006 || m.Header.EndToEndID != 0x34340006 {
		panic("header association")
	}
	text(field(m.AVP, 263, 0), "gmsc.example;24;6")
	u32(field(m.AVP, 277, 0), 1)
	if m.Header.CommandFlags&0x80 != 0 {
		identity := group(field(m.AVP, 3102, 10415), 16777312)
		if hex.EncodeToString(field(identity, 701, 10415).Data.Serialize()) != "2120550521f3" {
			panic("request MSISDN")
		}
		o := group(field(m.AVP, 3316, 10415), 16777312)
		positive := !strings.HasPrefix(name, "n24-")
		expectedNodes := 1
		if name == "b22-rdr-msc-sgsn" || name == "b22-rdr-mme-sgsn" {
			expectedNodes = 2
		}
		if positive && len(o) != expectedNodes {
			panic("node count")
		}
		for _, node := range o {
			if node.VendorID != 10415 || node.Flags != 0xc0 || node.Code < 3317 || node.Code > 3319 {
				panic("node key/flags")
			}
			children := group(node, 16777312)
			causeCount, diagCount := 0, 0
			for _, child := range children {
				if child.VendorID != 10415 || child.Flags != 0xc0 {
					panic("leaf key/flags")
				}
				switch child.Code {
				case 3321:
					if _, ok := child.Data.(datatype.Enumerated); !ok {
						panic("not dictionary Enumerated")
					}
					causeCount++
					if positive {
						want := uint32(1)
						if name == "b22-rdr-mme-sgsn" && node.Code == 3319 {
							want = 2
						}
						u32(child, want)
					}
				case 3322:
					if _, ok := child.Data.(datatype.Unsigned32); !ok {
						panic("not dictionary Unsigned32")
					}
					diagCount++
					if positive {
						want := uint32(0)
						if node.Code == 3317 {
							want = 13
						}
						if node.Code == 3319 {
							want = 5
						}
						u32(child, want)
					}
				default:
					panic("unexpected outcome leaf")
				}
			}
			if positive {
				want := 1
				if name == "c24-rdr-nodiag-r" || name == "b22-rdr-mme-sgsn" && node.Code == 3319 {
					want = 0
				}
				if causeCount != 1 || diagCount != want {
					panic("leaf presence")
				}
			}
		}
	} else {
		if strings.Contains(name, "5001") || strings.Contains(name, "5558") {
			r := group(field(m.AVP, 297, 0), 16777312)
			u32(field(r, 266, 0), 10415)
			v := uint32(5001)
			if strings.Contains(name, "5558") {
				v = 5558
			}
			u32(field(r, 298, 0), v)
		} else {
			u32(field(m.AVP, 268, 0), 2001)
			g := group(field(m.AVP, 3102, 10415), 16777312)
			if name == "n24-rda-emptyid" {
				if len(g) != 0 {
					panic("negative empty identity changed")
				}
			} else {
				if hex.EncodeToString(field(g, 701, 10415).Data.Serialize()) != "2120550521f4" {
					panic("different Alert MSISDN")
				}
			}
		}
	}
	out, e := m.Serialize()
	must(e)
	if name == "n24-rdr-diag3-r" {
		// This checkout silently decodes an invalid-width U32 as zero. Record the
		// exact observed rewrite; this is a peer defect, not negative conformance.
		node := group(field(group(field(m.AVP, 3316, 10415), 16777312), 3318, 10415), 16777312)
		u32(field(node, 3322, 10415), 0)
		pattern := []byte{0, 0, 12, 250, 192, 0, 0, 15, 0, 0, 40, 175, 0, 0, 0, 0}
		off := bytes.Index(b, pattern)
		if off < 0 {
			panic("missing malformed control")
		}
		want := append([]byte(nil), b...)
		want[off+7] = 16
		if !bytes.Equal(out, want) {
			panic("unexpected malformed-width behavior")
		}
		return nil
	}
	if !bytes.Equal(out, b) {
		panic("exact reencode mismatch")
	}
	return nil
}
func main() {
	if len(os.Args) != 2 {
		panic("argument: absolute testdata/codec directory")
	}
	for _, name := range []string{"v24-s6c-rd-r", "v24-s6c-rd-a", "c24-rdr-nodiag-r", "n24-rdr-diag256-r", "n24-rdr-diag3-r", "n24-rdr-dupdiag-r", "n24-rdr-successdiag-r", "n24-rdr-mscmme-r", "n24-rda-emptyid", "f24-s6c-rd-5001", "f24-s6c-rd-5558", "b22-rdr-sgsn-absent", "b22-rda-sgsn-ok", "b22-rdr-msc-sgsn", "b22-rda-pair-5558", "b22-rdr-mme-sgsn", "b22-rda-pair-5001"} {
		e := check(name)
		if name == "n24-rdr-diag3-r" {
			must(e)
			fmt.Printf("PASS %s: observed peer loss: invalid3-byte U32 becomes0 and reencodes as4 bytes; core rejects, not independent negative validation\n", name)
			continue
		}
		must(e)
		fmt.Printf("PASS %s (%d bytes): built-in dictionary types/flags/selected values, exact re-encode; semantic negatives not independently validated\n", name, len(load(name)))
	}
}
