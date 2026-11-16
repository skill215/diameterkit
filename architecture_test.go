// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The implemented DAG is checked at source level, including generated files.
// Future layers must extend this map intentionally when their implementation lands.
func TestProductionImportBoundaries(t *testing.T) {
	const module = "github.com/skill215/diameterkit"
	allowed := map[string][]string{
		"client/s6c":               {"app/s6c", "validation"},
		".":                        {"avp", "pdu", "schema", "wire", "validation", "base", "internal/budget", "internal/codecengine", "transport"},
		"internal/streamtransport": {"transport"}, "transport/tcp": {"transport", "internal/streamtransport"}, "transport/tls": {"internal/streamtransport", "transport"},
		"internal/smsidentity": {"avp", "validation"},
		// SCTP uses only wire's bounded top-level walker to select Session-Id's
		// stream; application decoding and core internal state remain forbidden.
		"internal/budget": {}, "transport": {}, "adapters/sctp": {"transport", "wire"}, "internal/testkit": {"transport"}, "avp": {}, "pdu": {"avp"}, "internal/contextvalue": {},
		"validation": {"avp", "internal/contextvalue"}, "wire": {"avp", "pdu"},
		"schema":                   {"avp", "pdu", "wire", "validation"},
		"base":                     {"avp", "pdu", "wire", "validation", "schema"},
		"compat/smsfwro/combined":  {"avp", "pdu", "wire", "validation", "schema", "base", "compat/smsfwro", "compat/smsfwro/legacy420"},
		"compat/scap1":             {"avp", "pdu", "wire", "validation", "schema", "base"},
		"compat/sgdalert":          {"app/s6c", "pdu", "wire", "validation", "schema"},
		"compat/smsfwro/legacy420": {"avp", "pdu", "wire", "validation", "schema", "base"},
		"compat/smsfwro":           {"avp", "pdu", "wire", "validation", "schema", "base"},
		"app/ro/dcca":              {"avp", "pdu", "wire", "validation", "schema", "base"},
		"app/ro/abort":             {"avp", "pdu", "wire", "validation", "schema", "base"},
		"app/ro":                   {"avp", "pdu", "wire", "validation", "schema", "base"},
		"app/sh/shdata":            {"avp"},
		"app/sh":                   {"internal/contextvalue", "avp", "pdu", "wire", "validation", "schema", "base"},
		"app/sgd":                  {"internal/smsidentity", "internal/contextvalue", "avp", "pdu", "wire", "validation", "schema", "base"},
		"app/s6c":                  {"internal/smsidentity", "internal/contextvalue", "avp", "pdu", "wire", "validation", "schema", "base"},
		"internal/codecengine":     {"avp", "pdu", "wire", "validation", "schema"},
		"codec":                    {"internal/codecengine", "pdu", "schema", "wire", "validation"},
	}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		pkg := filepath.ToSlash(filepath.Dir(path))
		if strings.HasPrefix(pkg, "cmd/") || pkg == "internal/codegen" || strings.HasPrefix(pkg, "examples/") || pkg == "interop" {
			return nil
		}
		permit, ok := allowed[pkg]
		if !ok {
			t.Errorf("unreviewed package %s", pkg)
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			s, _ := strconv.Unquote(imp.Path.Value)
			if s == "C" || s == "unsafe" || s == "reflect" || s == "os" || s == "database/sql" || s == "encoding/json" || s == "encoding/xml" && pkg != "app/sh/shdata" {
				t.Errorf("%s imports forbidden production dependency %s", path, s)
			}
			if s == module {
				if pkg != "client/s6c" {
					t.Errorf("forbidden root dependency: %s", path)
				}
			} else if strings.HasPrefix(s, module+"/") {
				target := strings.TrimPrefix(s, module+"/")
				found := false
				for _, a := range permit {
					if a == target {
						found = true
					}
				}
				if pkg != "internal/testkit" && target == "internal/testkit" {
					t.Errorf("production testkit dependency: %s", path)
				}
				if !found {
					t.Errorf("forbidden edge %s -> %s", pkg, target)
				}
			} else if strings.Contains(strings.Split(s, "/")[0], ".") && !(pkg == "adapters/sctp" && s == "github.com/skill215/sctp") {
				t.Errorf("unexpected external dependency %s", s)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
