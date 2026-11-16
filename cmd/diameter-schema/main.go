// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Command diameter-schema checks and generates the bounded JSON definition format.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/skill215/diameterkit/internal/codegen"
	"os"
	"path/filepath"
)

func run() error {
	in := flag.String("in", "", "definition JSON path")
	out := flag.String("out", "", "generated Go path")
	check := flag.Bool("check", false, "verify existing generated output without modifying it")
	flag.Parse()
	if *in == "" || *out == "" {
		return fmt.Errorf("-in and -out are required")
	}
	f, err := os.Open(*in)
	if err != nil {
		return err
	}
	d, err := codegen.Parse(f)
	f.Close()
	if err != nil {
		return err
	}
	b, err := codegen.Generate(d)
	if err != nil {
		return err
	}
	if *check {
		old, err := os.ReadFile(*out)
		if err != nil {
			return err
		}
		if !bytes.Equal(old, b) {
			return fmt.Errorf("generated output differs: %s", *out)
		}
		return nil
	}
	if err = os.MkdirAll(filepath.Dir(*out), 0755); err != nil {
		return err
	}
	f, err = os.CreateTemp(filepath.Dir(*out), ".diameter-schema-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Chmod(0644); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, *out)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
