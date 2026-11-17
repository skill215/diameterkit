// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit variable tree minima, rule audit boundaries and external generation.
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

func minimumVariableDefinition() Definition {
	d := minimumDefinition()
	d.Groups[0].Fields = nil
	for i, tc := range []struct {
		kind     string
		nonempty bool
		width    int
	}{{"identity", false, 1}, {"tbcd", false, 1}, {"utf8", false, 0}, {"utf8", true, 1}, {"octets", false, 0}, {"octets", true, 1}} {
		width := tc.width
		d.Groups[0].Fields = append(d.Groups[0].Fields, Field{Name: fmt.Sprintf("Field%d", i), Code: uint32(9010 + i), Vendor: 193, Flags: 128, Type: tc.kind, Nonempty: tc.nonempty, Min: 1, Max: 1, DiagnosticMinimumBytes: &width})
	}
	d.Groups[0].Fields = append(d.Groups[0].Fields, Field{Name: "Fixed", Code: 9020, Flags: 64, Type: "u32", Min: 1, Max: 1}, Field{Name: "Optional", Code: 9021, Type: "octets", Max: 1})
	return d
}
func TestMinimumVariableGrammar(t *testing.T) {
	for _, kind := range []string{"valid", "missing-width", "wrong-width", "optional-declaration", "minimum-two", "unreviewed-group", "missing-rule-audit", "stale-rule-audit", "audit-without-opt-in", "audit-without-rule", "reviewed-rule", "parent-rule", "reviewed-parent-rule", "empty"} {
		t.Run(kind, func(t *testing.T) {
			d := minimumVariableDefinition()
			g := &d.Groups[0]
			switch kind {
			case "missing-width":
				g.Fields[0].DiagnosticMinimumBytes = nil
			case "wrong-width":
				*g.Fields[0].DiagnosticMinimumBytes = 0
			case "optional-declaration":
				g.Fields[0].Min = 0
			case "minimum-two":
				g.Fields[0].Min = 2
				g.Fields[0].Max = 2
			case "unreviewed-group":
				g.DiagnosticMinimum = false
			case "missing-rule-audit":
				g.Rule = "checkValue"
			case "stale-rule-audit":
				g.Rule = "checkValue"
				g.DiagnosticMinimumRule = "oldCheckValue"
			case "audit-without-opt-in":
				g.DiagnosticMinimum = false
				g.Rule = "checkValue"
				g.DiagnosticMinimumRule = g.Rule
			case "audit-without-rule":
				g.DiagnosticMinimumRule = "checkValue"
			case "reviewed-rule":
				g.Rule = "checkValue"
				g.DiagnosticMinimumRule = g.Rule
			case "parent-rule":
				d.Groups[1].Rule = "checkMoney"
			case "reviewed-parent-rule":
				d.Groups[1].Rule = "checkMoney"
				d.Groups[1].DiagnosticMinimumRule = "checkMoney"
			case "empty":
				g.Fields = nil
			}
			_, e := Generate(d)
			want := kind == "valid" || kind == "reviewed-rule" || kind == "reviewed-parent-rule"
			if (e == nil) != want {
				t.Fatal(kind, e)
			}
		})
	}
}
func TestMinimumVariableGeneratedExternal(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	code, e := Generate(minimumVariableDefinition())
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("internal import")
	}
	dir := t.TempDir()
	source := `package extension
import("errors";"testing";"github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/validation";"github.com/skill215/diameterkit/wire")
func part(k avp.Key,flags byte,data []byte)[]byte{h:=8;if k.VendorID!=0{h=12};b:=make([]byte,(h+len(data)+3)&^3);w:=wire.NewEncoder(b,wire.DefaultLimits());w.Bytes(k,flags,data);return b}
func TestMixedMinimum(t *testing.T){
 var value []byte;for i:=0;i<6;i++{value=append(value,part(avp.Key{Code:uint32(9010+i),VendorID:193},128,[]byte{0x61})...)};value=append(value,part(avp.Key{Code:9020},64,make([]byte,4))...)
 money:=append(part(avp.Key{Code:445},64,value),part(avp.Key{Code:425},64,make([]byte,4))...)
 good:=part(avp.Key{Code:900},64,part(avp.Key{Code:413},64,money))
 raw:=append(good,part(avp.Key{Code:900},96,nil)...)
 _,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()));var field *validation.FieldError;var p interface{DiameterMissingGroupV2()(avp.Key,byte,string,error)}
 if !errors.Is(e,validation.ErrMissing)||!errors.As(e,&field)||len(field.Path)!=2||!errors.As(e,&p){t.Fatal(e)}
 key,flags,plan,child:=p.DiameterMissingGroupV2();if key.Code!=900||flags!=96||plan!=""||child==nil{t.Fatal("second parent")}
 p=child.(interface{DiameterMissingGroupV2()(avp.Key,byte,string,error)});key,flags,plan,child=p.DiameterMissingGroupV2()
 if key.Code!=413||flags!=64||len(plan)!=120||child!=nil||plan[10]!=2||plan[22]!=7{t.Fatal("mixed tree")}
 widths:=[]byte{1,1,0,1,0,1};for i:=0;i<6;i++{r:=plan[(i+2)*12:];if r[9]!=widths[i]||r[10]!=0||r[11]!=byte(i+1){t.Fatal("lost explicit kind/minimum")}}
 if plan[8*12+9]!=4||plan[9*12+9]!=4{t.Fatal("fixed siblings")}
 if _,e:=decodeRequest(wire.NewDecoder(good,wire.DefaultLimits()));e!=nil{t.Fatal("optional child gained presence",e)}
}
`
	for name, data := range map[string]string{"go.mod": "module example.org/variabletree\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "minimum_test.go": source} {
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
}
