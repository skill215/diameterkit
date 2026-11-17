// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit minimum grammar rejection and external generated consumption.
// SPDX-AI-Date: 2026-11-17

package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func minimumDefinition() Definition {
	return Definition{Format: 1, Package: "extension", Profile: "minimum-test", Version: "1", Sources: []string{"synthetic"}, Groups: []Group{
		{Name: "Value", DiagnosticMinimum: true, Fields: []Field{{Name: "Digits", Code: 447, Flags: 64, Type: "i64", Min: 1, Max: 1}, {Name: "Exponent", Code: 429, Flags: 64, Type: "i32", Min: 1, Max: 1}, {Name: "Optional", Code: 999, Type: "octets", Max: 1}}},
		{Name: "Money", DiagnosticMinimum: true, Fields: []Field{{Name: "Value", Code: 445, Flags: 64, Type: "Value", Min: 1, Max: 1}, {Name: "Currency", Code: 425, Flags: 64, Type: "u32", Min: 1, Max: 1}}},
		{Name: "Container", Fields: []Field{{Name: "Money", Code: 413, Flags: 64, Type: "Money", Min: 1, Max: 1}}},
	}, Commands: []Command{{Name: "Request", Code: 999, Application: 999, Request: true, Fields: []Field{{Name: "Containers", Code: 900, Flags: 64, Type: "Container", Max: -1}}}}}
}
func TestMinimumGrammar(t *testing.T) {
	for _, kind := range []string{"valid", "variable", "unreviewed-child", "capture", "minimum-two", "empty", "too-large", "cycle"} {
		t.Run(kind, func(t *testing.T) {
			d := minimumDefinition()
			switch kind {
			case "variable":
				d.Groups[0].Fields[0].Type = "octets"
			case "unreviewed-child":
				d.Groups[0].DiagnosticMinimum = false
			case "capture":
				d.Groups[0].CaptureMandatory = true
			case "minimum-two":
				d.Groups[0].Fields[0].Min = 2
				d.Groups[0].Fields[0].Max = 2
			case "empty":
				for i := range d.Groups[0].Fields {
					d.Groups[0].Fields[i].Min = 0
				}
			case "too-large":
				for i := 0; i < 22; i++ {
					d.Groups[1].Fields = append(d.Groups[1].Fields, Field{Name: fmt.Sprint("Child", i), Code: uint32(1000 + i), Type: "Value", Min: 1, Max: 1})
				}
			case "cycle":
				d.Groups[0].Fields[0].Type = "Money"
			}
			_, err := Generate(d)
			if (err == nil) != (kind == "valid") {
				t.Fatal(kind, err)
			}
		})
	}
}
func TestMinimumGeneratedExternal(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	code, e := Generate(minimumDefinition())
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("internal import")
	}
	dir := t.TempDir()
	files := map[string]string{"go.mod": "module example.org/minimum\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "minimum_test.go": `package extension
import("errors";"testing";"github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/wire";"github.com/skill215/diameterkit/validation")
func TestExternalMinimum(t *testing.T){
 raw:=make([]byte,8);w:=wire.NewEncoder(raw,wire.DefaultLimits());w.Bytes(avp.Key{Code:900},64,nil)
 _,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()));var field *validation.FieldError
 var p interface{DiameterMissingGroupV2()(avp.Key,byte,string,error)}
 if !errors.Is(e,validation.ErrMissing)||!errors.As(e,&field)||!errors.As(e,&p)||len(field.Path)!=2||field.Path[1].Code!=413{t.Fatal(e)}
 key,flags,plan,child:=p.DiameterMissingGroupV2();if key.Code!=900||flags!=64||plan!=""||child==nil{t.Fatal("ancestor")}
 p=child.(interface{DiameterMissingGroupV2()(avp.Key,byte,string,error)});key,flags,plan,child=p.DiameterMissingGroupV2()
 if key.Code!=413||flags!=64||len(plan)!=60||child!=nil||plan[10]!=2||plan[22]!=2{t.Fatal("minimum plan")}
 if plan[2*12+9]!=8||plan[3*12+9]!=4||plan[4*12+9]!=4{t.Fatal("scalar widths")}
}
`}
	for name, body := range files {
		if e = os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); e != nil {
			t.Fatal(e)
		}
	}
	cmd := exec.Command("go", "test", "-timeout=30s", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "CGO_ENABLED=0")
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("external minimum: %v\n%s", e, out)
	}
}
