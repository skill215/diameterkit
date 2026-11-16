//go:build ignore

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
	for _, name := range []string{"v25-srr-two-r", "v25-sra-sgsn-msc", "v25-sra-mme", "v25-sra-sgsn-n", "v25-sra-sgsn-3", "v25-sra-msc", "v25-sra-msc-comb", "n25-srr-nosmea", "n25-smea-length", "n25-smea-toa", "c27-smea-even", "c27-smea-max15", "c27-smea-min1", "c27-smea-national", "c27-smea-16digits", "c27-smea-bit0-clear", "n27-smea-pad", "n27-smea-digit", "n27-smea-stripped", "n27-smea-tlv"} {
		b := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(e)
		if m.Header.ApplicationID != 16777312 || m.Header.CommandCode != 8388647 {
			panic("command/application")
		}
		u32(field(m.AVP, 277, 0), 1)
		if strings.HasPrefix(name, "v25-sra") {
			text(field(m.AVP, 263, 0), "gmsc.example;25;32")
			text(field(m.AVP, 1, 0), "001010123456789")
			u32(field(m.AVP, 268, 0), 2001)
			for _, code := range []uint32{2401, 2406} {
				outer := field(m.AVP, code, 10415)
				if outer.Flags != 0xc0 {
					panic("node flags")
				}
				children := group(outer, 16777312)
				for _, child := range children {
					if child.VendorID != 10415 {
						panic("child vendor")
					}
					switch child.Code {
					case 2402, 2408, 2409, 2410:
						if _, ok := child.Data.(datatype.DiameterIdentity); !ok {
							panic("identity type")
						}
						want := byte(0xc0)
						if child.Code == 2409 || child.Code == 2410 {
							want = 0x80
						}
						if child.Flags != want {
							panic("identity flags")
						}
						host := "example.net"
						if child.Code == 2402 {
							host = "mme.example"
						}
						if child.Code == 2409 {
							host = "sgsn.example"
						}
						text(child, host)
					case 1489, 1645, 2403:
						if _, ok := child.Data.(datatype.OctetString); !ok {
							panic("number type")
						}
						want := "2120550577f7"
						if child.Code == 1645 {
							want = "2120550588f8"
						}
						if child.Code == 2403 {
							want = "2120550566f6"
						}
						if child.Flags != 0xc0 || hex.EncodeToString(child.Data.Serialize()) != want {
							panic("number flags/value")
						}
					default:
						panic("unexpected node child")
					}
				}
			}
			if name == "v25-sra-sgsn-msc" || name == "v25-sra-msc-comb" {
				if hex.EncodeToString(field(m.AVP, 2400, 10415).Data.Serialize()) != "01234567" {
					panic("LMSI")
				}
			}
			if name == "v25-sra-sgsn-msc" {
				u32(field(m.AVP, 3312, 10415), 0)
				g := group(field(m.AVP, 3102, 10415), 16777312)
				if hex.EncodeToString(field(g, 701, 10415).Data.Serialize()) != "2120550521f4" {
					panic("alert MSISDN")
				}
			}
		} else {
			flags := field(m.AVP, 3310, 10415)
			if _, ok := flags.Data.(datatype.Unsigned32); !ok {
				panic("flags type")
			}
			if name != "n25-srr-nosmea" {
				sme := field(m.AVP, 3309, 10415)
				if _, ok := sme.Data.(datatype.OctetString); !ok {
					panic("SMEA type")
				}
			}
		}
		out, e := m.Serialize()
		must(e)
		if !bytes.Equal(out, b) {
			panic("independent reencode mismatch")
		}
		fmt.Printf("PASS %s (%d bytes): built-in S6c dictionary, selected child types/flags/values, exact re-encode; SMEA remains opaque, no semantic validation claim\n", name, len(b))
	}
}
