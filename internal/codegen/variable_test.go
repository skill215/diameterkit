// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Declared minimum validation and external variable omission evidence.
// SPDX-AI-Date: 2026-11-17

package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func variableDefinition(kind string, nonempty bool, width int) Definition {
	return Definition{Format: 1, Package: "extension", Profile: "variable-test", Version: "1", Sources: []string{"synthetic"}, Groups: []Group{
		{Name: "Leaf", Fields: []Field{{Name: "Value", Code: 9002, Vendor: 193, Flags: 128, Type: kind, Nonempty: nonempty, Min: 1, Max: 1, DiagnosticMinimumBytes: &width}}},
		{Name: "Container", Fields: []Field{{Name: "Leaves", Code: 9001, Vendor: 193, Flags: 128, Type: "Leaf", Max: -1}}},
	}, Commands: []Command{{Name: "Request", Code: 999, Application: 999, Request: true, Fields: []Field{{Name: "Containers", Code: 9000, Flags: 64, Type: "Container", Max: -1}}}}}
}
func TestVariableMinimumDefinition(t *testing.T) {
	for _, kind := range []string{"identity", "tbcd", "utf8", "octets", "u32", "address", "proxy", "vsai", "diagnostic"} {
		for _, nonempty := range []bool{false, true} {
			for _, width := range []int{-1, 0, 1, 2, 4, 1024} {
				t.Run(fmt.Sprintf("%s/%t/%d", kind, nonempty, width), func(t *testing.T) {
					d := variableDefinition(kind, nonempty, width)
					b, e := json.Marshal(d)
					if e != nil {
						t.Fatal(e)
					}
					parsed, e := Parse(bytes.NewReader(b))
					valid := (kind == "identity" || kind == "tbcd") && width == 1 || (kind == "utf8" || kind == "octets") && (nonempty && width == 1 || !nonempty && width == 0)
					if (e == nil) != valid {
						t.Fatal("minimum declaration", e)
					}
					if !valid {
						return
					}
					if parsed.Groups[0].Fields[0].DiagnosticMinimumBytes == nil || *parsed.Groups[0].Fields[0].DiagnosticMinimumBytes != width {
						t.Fatal("lost explicit zero")
					}
					code, e := Generate(parsed)
					if e != nil || !bytes.Contains(code, []byte("return nil, profileMissingVariable(")) {
						t.Fatal("missing evidence", e)
					}
				})
			}
		}
	}
	for _, minimum := range []int{0, 2} {
		d := variableDefinition("utf8", false, 0)
		d.Groups[0].Fields[0].Min = minimum
		d.Groups[0].Fields[0].Max = 2
		if _, e := Generate(d); e == nil {
			t.Fatal("unsupported presence minimum")
		}
	}
	d := variableDefinition("utf8", false, 0)
	d.Groups[0].Fields[0].DiagnosticMinimumBytes = nil
	code, e := Generate(d)
	if e != nil || bytes.Contains(code, []byte("return nil, profileMissingVariable(")) {
		t.Fatal("inferred opt-in", e)
	}
	d = variableDefinition("utf8", false, 0)
	d.Groups[0].DiagnosticMinimum = true
	if _, e := Generate(d); e == nil {
		t.Fatal("broadened whole-group plan")
	}
}
func TestVariableGeneratedExternal(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		kind     string
		nonempty bool
		width    int
	}{{"identity", false, 1}, {"tbcd", false, 1}, {"utf8", false, 0}, {"utf8", true, 1}, {"octets", false, 0}, {"octets", true, 1}} {
		t.Run(fmt.Sprintf("%s/%t", tc.kind, tc.nonempty), func(t *testing.T) {
			code, e := Generate(variableDefinition(tc.kind, tc.nonempty, tc.width))
			if e != nil {
				t.Fatal(e)
			}
			if bytes.Contains(code, []byte("/internal/")) {
				t.Fatal("internal dependency")
			}
			dir := t.TempDir()
			kind := tc.kind
			if tc.nonempty {
				kind += "-nonempty"
			}
			body := fmt.Sprintf(`package extension
import("errors";"testing";"github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/validation";"github.com/skill215/diameterkit/wire")
func part(code uint32,flags byte,data []byte)[]byte{k:=avp.Key{Code:code};h:=8;if code!=9000{k.VendorID=193;h=12};b:=make([]byte,(h+len(data)+3)&^3);w:=wire.NewEncoder(b,wire.DefaultLimits());w.Bytes(k,flags,data);return b}
func TestVariableBridge(t *testing.T){
 good:=part(9001,128,part(9002,128,[]byte{0x61}))
 raw:=append(part(9000,64,good),part(9000,64,append(good,part(9001,192,nil)...))...)
 _,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()));var field *validation.FieldError;var p interface{DiameterMissingAVPV1()(avp.Key,byte,int,error)}
 if !errors.Is(e,validation.ErrMissing)||!errors.As(e,&field)||len(field.Path)!=3||!errors.As(e,&p){t.Fatal(e)}
 for i,k:=range []uint32{9000,9001,9002}{key,flags,width,child:=p.DiameterMissingAVPV1();if key.Code!=k{t.Fatal("path")};if i<2{if width!=0||i==1&&flags!=192{t.Fatal("wrong parent")};p=child.(interface{DiameterMissingAVPV1()(avp.Key,byte,int,error)})}else{if child!=nil||width!=%d||p.(interface{DiameterMissingVariableV1()string}).DiameterMissingVariableV1()!=%q{t.Fatal("leaf proof")}}}
 empty:=part(9000,64,part(9001,128,part(9002,128,nil)));_,e=decodeRequest(wire.NewDecoder(empty,wire.DefaultLimits()));if errors.Is(e,validation.ErrMissing)||(e==nil)!=%t{t.Fatal("present empty confused with missing",e)}
}
`, tc.width, kind, tc.width == 0)
			files := map[string]string{"go.mod": "module example.org/variable\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "variable_test.go": body}
			for name, data := range files {
				if e = os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); e != nil {
					t.Fatal(e)
				}
			}
			cmd := exec.Command("go", "test", "-timeout=30s", "./...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "CGO_ENABLED=0")
			if out, e := cmd.CombinedOutput(); e != nil {
				t.Fatalf("external: %v\n%s", e, out)
			}
		})
	}
}
