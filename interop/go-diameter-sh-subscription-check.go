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
func absent(as []*diam.AVP, code, vendor uint32) {
	for _, a := range as {
		if a.Code == code && a.VendorID == vendor {
			panic("unexpected field")
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		panic("argument: absolute fixture directory")
	}
	cases := []struct {
		name        string
		ref, action uint32
	}{
		{"d21-sub11", 11, 0}, {"b19-sub12", 12, 0}, {"b19-unsub25", 25, 1}, {"b19-unsub11", 11, 1}, {"b19-unsub12", 12, 1},
	}
	for _, c := range cases {
		for _, suffix := range []string{"-r", "-a", "-f"} {
			name := c.name + suffix
			raw := load(name)
			m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
			must(e)
			if m.Header.ApplicationID != 16777217 || m.Header.CommandCode != 308 {
				panic("Sh SNR/SNA header")
			}
			vsai := group(field(m.AVP, 260, 0), 16777217)
			u32(field(vsai, 266, 0), 10415)
			u32(field(vsai, 258, 0), 16777217)
			u32(field(m.AVP, 277, 0), 1)
			absent(m.AVP, 628, 10415)
			absent(m.AVP, 1, 0)
			if suffix == "-r" {
				if m.Header.CommandFlags&0x80 == 0 {
					panic("request bit")
				}
				text(field(group(field(m.AVP, 700, 10415), 16777217), 601, 10415), "sip:alice@example.net")
				for _, f := range []struct{ code, value uint32 }{{703, c.ref}, {705, c.action}} {
					a := field(m.AVP, f.code, 10415)
					if _, ok := a.Data.(datatype.Enumerated); !ok {
						panic("enum type")
					}
					u32(a, f.value)
				}
				if c.action == 0 {
					a := field(m.AVP, 710, 10415)
					if _, ok := a.Data.(datatype.Enumerated); !ok {
						panic("SendData type")
					}
					u32(a, 1)
					if a.Flags&64 != 0 {
						panic("SendData M")
					}
				} else {
					absent(m.AVP, 710, 10415)
				}
				if c.ref == 25 {
					a := field(m.AVP, 712, 10415)
					u32(a, 0)
					if a.Flags&64 != 0 {
						panic("OneTime M")
					}
				} else {
					absent(m.AVP, 712, 10415)
				}
			} else {
				if m.Header.CommandFlags&0x80 != 0 {
					panic("answer bit")
				}
				req, e := diam.ReadMessage(bytes.NewReader(load(c.name+"-r")), dict.Default)
				must(e)
				if req.Header.HopByHopID != m.Header.HopByHopID || req.Header.EndToEndID != m.Header.EndToEndID {
					panic("pair IDs")
				}
				text(field(m.AVP, 263, 0), string(field(req.AVP, 263, 0).Data.Serialize()))
				if suffix == "-f" {
					x := group(field(m.AVP, 297, 0), 16777217)
					u32(field(x, 266, 0), 10415)
					u32(field(x, 298, 0), 5104)
					absent(m.AVP, 702, 10415)
				} else {
					u32(field(m.AVP, 268, 0), 2001)
					if c.action == 0 {
						a := field(m.AVP, 702, 10415)
						if _, ok := a.Data.(datatype.OctetString); !ok {
							panic("UserData type")
						}
						want := "<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState></Sh-IMS-Data></Sh-Data>"
						if c.ref == 12 {
							want = "<Sh-Data><Sh-IMS-Data><SCSCFName>sip:scscf.example.net</SCSCFName></Sh-IMS-Data></Sh-Data>"
						}
						text(a, want)
					} else {
						absent(m.AVP, 702, 10415)
					}
				}
			}
			out, e := m.Serialize()
			must(e)
			if !bytes.Equal(out, raw) {
				panic("exact reencode")
			}
			fmt.Printf("PASS %s (%d bytes): independent enum/flags/presence, result and pair checks, exact re-encode; XML opaque\n", name, len(raw))
		}
	}
}
