//go:build ignore

// Independent header/body preservation evidence. No diameterkit imports.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func load(dir, name string) []byte {
	b, e := os.ReadFile(filepath.Join(dir, name))
	must(e)
	b, e = hex.DecodeString(strings.Join(strings.Fields(string(b)), ""))
	must(e)
	return b
}
func main() {
	if len(os.Args) != 3 {
		panic("fixture directory and explicit legacy offline overlay required")
	}
	raw, e := os.ReadFile(filepath.Join(os.Args[1], "manifest.json"))
	must(e)
	var entries []struct{ File, Source, Profile string }
	must(json.Unmarshal(raw, &entries))
	count := 0
	for _, profile := range []string{"standard", "legacy"} {
		mode := "original dictionary"
		if profile == "legacy" {
			must(dict.Default.LoadFile(os.Args[2]))
			mode = "explicit legacy420/425/1292/2018 offline overlay"
		}
		for _, f := range entries {
			if !strings.HasPrefix(f.File, "b50-") || f.Profile != profile {
				continue
			}
			b := load(os.Args[1], f.File)
			src := load(os.Args[1], f.Source)
			if !bytes.Equal(b[16:], src[16:]) || !bytes.Equal(b[5:12], src[5:12]) {
				panic("body/E2E/command changed")
			}
			m, e := diam.ReadMessage(bytes.NewReader(b), dict.Default)
			must(e)
			out, e := m.Serialize()
			must(e)
			if !bytes.Equal(b, out) {
				panic("exact reencode")
			}
			if m.Header.ApplicationID != 4 || m.Header.CommandCode != 272 {
				panic("wrong command")
			}
			isT := m.Header.CommandFlags&16 != 0
			if isT != strings.Contains(f.File, "-retransmit") {
				panic("T-bit changed")
			}
			fmt.Printf("PASS %s (%d bytes): %s; identity/application bytes and T preserved, not OCS deduplication evidence\n", strings.TrimSuffix(f.File, ".hex"), len(b), mode)
			count++
		}
	}
	if count != 27 {
		panic("case count")
	}
}
