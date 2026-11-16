//go:build ignore

// Run in a disposable module with the original go-diameter dependency.
// No diameterkit code, dictionary or generated expectations are imported.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/datatype"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func load(name string) []byte {
	raw, e := os.ReadFile(filepath.Join(os.Args[1], name+".hex"))
	must(e)
	b, e := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
	must(e)
	return b
}
func field(m *diam.Message, code, vendor uint32) *diam.AVP {
	for _, v := range m.AVP {
		if v.Code == code && v.VendorID == vendor {
			return v
		}
	}
	panic("missing field")
}
func main() {
	if len(os.Args) != 2 {
		panic("absolute testdata/codec directory required")
	}
	for _, name := range []string{"v24-sgd-mt-r", "v24-sgd-mt-a", "v24-sgd-mt-wait", "v24-gdd-mt-r", "v24-gdd-mt-a", "v24-gdd-mt-wait", "v24-era-mt-r", "v24-era-mt-wait", "c27-time-zero-r", "c27-time-zero-a", "n24-tfr-nogmsc-r", "c24-tfr-nomax-r", "n24-tfa-late", "c24-tfa-equal", "n24-tfa-successtime", "n24-tfa-time3", "n24-era-late"} {
		b := load(name)
		m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
		must(e)
		if m.Header.CommandCode != 8388646 || m.Header.ApplicationID != 16777313 {
			panic("command/app")
		}
		// Independently enumerate root TLVs: dictionary type/flags/bytes, no semantic claim.
		for off := 20; off < len(b); {
			n := int(b[off+5])<<16 | int(b[off+6])<<8 | int(b[off+7])
			code := binary.BigEndian.Uint32(b[off:])
			vendor := uint32(0)
			h := 8
			if b[off+4]&0x80 != 0 {
				vendor = binary.BigEndian.Uint32(b[off+8:])
				h = 12
			}
			data := b[off+h : off+n]
			off += (n + 3) &^ 3
			if vendor != 10415 {
				continue
			}
			switch code {
			case 3307, 3330, 3331:
				a := field(m, code, vendor)
				flags := byte(0x80)
				if code == 3307 {
					flags = 0xc0
				}
				if a.Flags != flags {
					panic("Time flags")
				}
				if len(data) == 4 {
					v, ok := a.Data.(datatype.Time)
					if !ok || !bytes.Equal(data, v.Serialize()) {
						panic("Time type/value")
					}
					if binary.BigEndian.Uint32(data) == 0 && !time.Time(v).Equal(time.Date(2036, 2, 7, 6, 28, 16, 0, time.UTC)) {
						panic("legacy era mapping")
					}
				} else {
					if name != "n24-tfa-time3" || code != 3331 || len(data) != 3 {
						panic("unexpected width")
					}
					// Legacy decoder returns a zero *Time and nil error for invalid width.
					if _, ok := a.Data.(*datatype.Time); !ok {
						panic("changed malformed-time behavior")
					}
					if len(a.Data.Serialize()) != 4 {
						panic("changed lossy width")
					}
				}
			case 3332:
				a := field(m, code, vendor)
				if a.Flags != 0x80 {
					panic("GMSC flags")
				}
				v, ok := a.Data.(datatype.OctetString)
				if !ok || !bytes.Equal(data, v.Serialize()) {
					panic("GMSC type/value")
				}
			}
		}
		// Serialize before another ReadMessage can reuse the peer decoder buffer.
		out, e := m.Serialize()
		must(e)
		if m.Header.CommandFlags&0x80 == 0 {
			req := "v24-sgd-mt-r"
			switch {
			case strings.Contains(name, "gdd"):
				req = "v24-gdd-mt-r"
			case strings.Contains(name, "era"):
				req = "v24-era-mt-r"
			case strings.HasPrefix(name, "c27"):
				req = "c27-time-zero-r"
			}
			r, e := diam.ReadMessage(bytes.NewReader(load(req)), dict.Default)
			must(e)
			if r.Header.HopByHopID != m.Header.HopByHopID || r.Header.EndToEndID != m.Header.EndToEndID || !bytes.Equal(field(r, 263, 0).Data.Serialize(), field(m, 263, 0).Data.Serialize()) {
				panic("pair association")
			}
		}
		if name == "n24-tfa-time3" {
			if bytes.Equal(b, out) {
				panic("expected legacy width loss")
			}
			fmt.Printf("PASS %s (%d bytes): observed legacy malformed Time accepted with nil error then 3-to-4-byte payload rewrite; not a rejection\n", name, len(b))
		} else {
			if !bytes.Equal(b, out) {
				panic("exact reencode " + name)
			}
			fmt.Printf("PASS %s (%d bytes): original dictionary Time/OctetString types, flags, pair IDs and exact reencode; no era/fact/semantic certification\n", name, len(b))
		}
	}
}
