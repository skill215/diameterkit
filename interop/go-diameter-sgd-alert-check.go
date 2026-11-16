//go:build ignore

// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Standalone dictionary observation, not a diameterkit production dependency.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctrlzy/go-diameter/v4/diam"
	"github.com/ctrlzy/go-diameter/v4/diam/dict"
)

func main() {
	if len(os.Args) != 2 {
		panic("argument: testdata/codec directory")
	}
	// The existing SGd dictionary has no ALR command grammar. Install an
	// explicit process-local declaration; this is not default compatibility.
	if _, e := dict.Default.FindCommand(16777313, 8388648); e == nil {
		panic("update evidence: default dictionary now has SGd ALR")
	}
	const overlay = `<diameter><application id="16777313" type="auth" name="legacy-sgd-alert-v1"><command code="8388648" short="AL" name="Alert-Service-Centre"><request><rule avp="AVP"/></request><answer><rule avp="AVP"/></answer></command></application></diameter>`
	if e := dict.Default.Load(strings.NewReader(overlay)); e != nil {
		panic(e)
	}
	files, e := filepath.Glob(filepath.Join(os.Args[1], "b73-*.hex"))
	if e != nil {
		panic(e)
	}
	if len(files) != 26 {
		panic("fixture count")
	}
	for _, file := range files {
		h, e := os.ReadFile(file)
		if e != nil {
			panic(e)
		}
		raw, e := hex.DecodeString(strings.Join(strings.Fields(string(h)), ""))
		if e != nil {
			panic(e)
		}
		m, e := diam.ReadMessage(bytes.NewReader(raw), dict.Default)
		if e != nil {
			panic(e)
		}
		if m.Header.ApplicationID != 16777313 || m.Header.CommandCode != 8388648 {
			panic("header identity")
		}
		b, e := m.Serialize()
		if e != nil {
			panic(e)
		}
		if !bytes.Equal(raw, b) {
			panic("wire changed")
		}
		fmt.Println(filepath.Base(file), len(raw), "exact header/AVP decode/reencode; no command/role certification")
	}
}
