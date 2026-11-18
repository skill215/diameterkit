// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit missing callback validation and external-module fallback acceptance.
// SPDX-AI-Date: 2026-11-18

package codegen

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func cardinalityDefinition() Definition {
	return Definition{Format: 1, Package: "extension", Profile: "cardinality-test", Version: "1", Sources: []string{"synthetic"},
		Groups:   []Group{{Name: "Item", Fields: []Field{{Name: "Type", Code: 450, Flags: 64, Type: "u32", Min: 1, Max: 1}}}},
		Commands: []Command{{Name: "Request", Code: 999, Application: 999, Request: true, NoContext: true, Rule: "validateSubset", Fields: []Field{{Name: "Items", Code: 443, Flags: 64, Type: "Item", Min: 2, Max: 2, DiagnosticMissingRule: "diagnoseMissing"}}}}}
}
func TestCardinalityMissingRuleDefinition(t *testing.T) {
	for _, kind := range []string{"valid", "optional", "singleton", "unbounded", "primitive", "identifier", "same-rule", "no-rule", "answer", "nested", "symbol", "minimum-plan"} {
		t.Run(kind, func(t *testing.T) {
			d := cardinalityDefinition()
			f := &d.Commands[0].Fields[0]
			switch kind {
			case "optional":
				f.Min = 0
			case "singleton":
				f.Min = 1
				f.Max = 1
			case "unbounded":
				f.Max = -1
			case "primitive":
				f.Type = "u32"
			case "identifier":
				f.DiagnosticMissingRule = "bad-name"
			case "same-rule":
				f.DiagnosticMissingRule = d.Commands[0].Rule
			case "no-rule":
				d.Commands[0].Rule = ""
			case "answer":
				d.Commands[0].Request = false
			case "nested":
				d.Groups = append(d.Groups, Group{Name: "Outer", Fields: []Field{*f}})
				f.DiagnosticMissingRule = ""
			case "symbol":
				f.DiagnosticMissingRule = "decodeRequest"
			case "minimum-plan":
				d.Groups[0].DiagnosticMinimum = true
			}
			_, err := Generate(d)
			if (err == nil) != (kind == "valid") {
				t.Fatal(kind, err)
			}
		})
	}
}
func TestExternalCardinalityMissingRule(t *testing.T) {
	code, err := Generate(cardinalityDefinition())
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("internal dependency")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	body := `package extension
import("errors";"testing";"github.com/skill215/diameterkit/validation";"github.com/skill215/diameterkit/wire";"github.com/skill215/diameterkit/avp")
var marker=errors.New("explicit proof")
var calls int
var enabled=true
func diagnoseMissing(m *Request)error{calls++;if enabled{return marker};return nil}
func validateSubset(m *Request,mode validation.Mode)error{return nil}
func TestHook(t *testing.T){
 for _,have:=range []int{0,1,2}{
  w:=wire.NewEncoder(nil,wire.DefaultLimits());emit:=func(w *wire.Encoder){for i:=0;i<have;i++{g:=w.BeginGroup(avp.Key{Code:443},64);w.Uint32(avp.Key{Code:450},64,0);w.EndGroup(g)}};emit(w);raw:=make([]byte,w.Len());emit(wire.NewEncoder(raw,wire.DefaultLimits()))
  calls=0;enabled=true;_,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()))
  if have<2{if !errors.Is(e,marker)||calls!=1{t.Fatal("hook",e,calls)}}else{if e!=nil||calls!=0{t.Fatal("complete",e,calls)}}
  calls=0;enabled=false;_,e=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()))
  if have<2&&!errors.Is(e,validation.ErrMissing){t.Fatal("nil callback bypassed validation",e)}
 }
 calls=0;_ = validateRequest(&Request{},validation.Build);_ = validateRequest(&Request{},validation.Send);if calls!=0{t.Fatal("nonreceive hook")}
}
`
	files := map[string]string{"go.mod": "module example.org/cardinality\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "hook_test.go": body}
	for name, data := range files {
		if err = os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "test", "-timeout=30s", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
