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
		panic("argument: testdata/codec directory")
	}
	// Explicit old-App command declaration only; standard S6c uses Default.
	const overlay = `<diameter><application id="16777313" type="auth" name="legacy-sgd-alert-v1"><command code="8388648" short="AL" name="Alert-Service-Centre"><request><rule avp="AVP"/></request><answer><rule avp="AVP"/></answer></command></application></diameter>`
	must(dict.Default.Load(strings.NewReader(overlay)))
	names, err := filepath.Glob(filepath.Join(os.Args[1], "b74-router-*.hex"))
	must(err)
	if len(names) != 23 {
		panic("fixture count")
	}
	for _, path := range names {
		name := strings.TrimSuffix(filepath.Base(path), ".hex")
		raw := load(name)
		m, err := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		must(err)
		if m.Header.CommandCode != 8388648 {
			panic("command")
		}
		app := uint32(16777312)
		if strings.Contains(name, "legacy-") {
			app = 16777313
		}
		if m.Header.ApplicationID != app {
			panic("app")
		}
		if m.Header.CommandFlags&0x80 != 0 {
			g := group(field(m.AVP, 3102, 10415), app)
			id := field(g, 1, 0)
			if _, ok := id.Data.(datatype.UTF8String); !ok {
				panic("User-Name type")
			}
			want := "router-alert-21"
			switch strings.TrimPrefix(name, "b74-router-") {
			case "numeric":
				want = "001010123456789"
			case "short":
				want = "7"
			case "unicode":
				want = "令牌-é"
			case "decomposed":
				want = "令牌-e\u0301"
			case "max":
				want = strings.Repeat("x", 256)
			case "long":
				want = strings.Repeat("x", 257)
			case "empty":
				want = ""
			case "nul":
				want = "a\x00b"
			case "invalid-utf8":
				want = "\xff"
			}
			text(id, want)
		} else {
			code := uint32(2001)
			if strings.HasSuffix(name, "failure") {
				code = 5012
			}
			if strings.HasSuffix(name, "-e") {
				code = 3004
			}
			u32(field(m.AVP, 268, 0), code)
		}
		encoded, err := m.Serialize()
		must(err)
		if !bytes.Equal(raw, encoded) {
			panic("wire changed")
		}
		fmt.Println(filepath.Base(path), len(raw), "exact identity/token/result decode/reencode; not conditional validation")
	}
}
