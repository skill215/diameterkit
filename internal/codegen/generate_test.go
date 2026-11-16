package codegen

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func definition(t *testing.T, name string) Definition {
	t.Helper()
	f, err := os.Open("../../spec/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func TestGeneratedFilesCurrentAndDeterministic(t *testing.T) {
	for name, path := range map[string]string{"dcca": "../../app/ro/dcca/zz_generated_messages.go", "roabort": "../../app/ro/abort/zz_generated_messages.go", "combined": "../../compat/smsfwro/combined/zz_generated_messages.go", "scap1": "../../compat/scap1/zz_generated_messages.go", "legacy420": "../../compat/smsfwro/legacy420/zz_generated_messages.go", "smsfwro": "../../compat/smsfwro/zz_generated_messages.go", "ro": "../../app/ro/zz_generated_messages.go", "sh": "../../app/sh/zz_generated_messages.go", "sgd": "../../app/sgd/zz_generated_messages.go", "base": "../../base/zz_generated_messages.go", "s6c": "../../app/s6c/zz_generated_messages.go"} {
		d := definition(t, name)
		got, err := Generate(d)
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("stale %s; run go generate ./...", path)
		}
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		var obj map[string]any
		if err = json.Unmarshal(b, &obj); err != nil {
			t.Fatal(err)
		}
		reordered, err := json.MarshalIndent(obj, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		d2, err := Parse(bytes.NewReader(reordered))
		if err != nil {
			t.Fatal(err)
		}
		other, err := Generate(d2)
		if err != nil || !bytes.Equal(got, other) {
			t.Fatal("JSON key/whitespace affected generation", err)
		}
	}
}
func TestStrictInputAndReferences(t *testing.T) {
	for _, s := range []string{`{"format":1,"format":1}`, `{"Format":1}`, `{"format":1,"unknown":2}`, `null`, `{} {}`, `[]`} {
		if _, err := Parse(strings.NewReader(s)); err == nil {
			t.Fatalf("accepted %s", s)
		}
	}
	cases := map[string]func(*Definition){
		"request error rule":   func(d *Definition) { d.Commands[0].ErrorRule = "checkError" },
		"error rule without E": func(d *Definition) { d.Commands[1].ErrorAnswer = false; d.Commands[1].ErrorRule = "checkError" },
		"invalid error rule":   func(d *Definition) { d.Commands[1].ErrorRule = "bad()" },
		"error rule collision": func(d *Definition) { d.Commands[1].ErrorRule = "validateSRA" },
		"correlation identity": func(d *Definition) { d.Commands[0].Fields[0].Code = 999 },
		"answer replay":        func(d *Definition) { d.Commands[1].TransportReplay = true },
		"base replay":          func(d *Definition) { d.Commands[0].Application = 0 },
		"request E":            func(d *Definition) { d.Commands[0].ErrorAnswer = true },
		"fake Origin binding":  func(d *Definition) { d.Commands[0].Fields[2].Code = 999 },
		"blank package":        func(d *Definition) { d.Package = "_" },
		"duplicate enum":       func(d *Definition) { d.Commands[0].Fields[1].Values = []uint32{1, 1} },
		"rule collision":       func(d *Definition) { d.Commands[0].Rule = "validateSRR" },
		"group envelope":       func(d *Definition) { d.Groups[1].Fields[0].Envelope = "OriginHost" },
		"vendor":               func(d *Definition) { d.Commands[0].Fields[0].Vendor = 10415 },
		"unknown type":         func(d *Definition) { d.Commands[0].Fields[0].Type = "Missing" },
		"duplicate field":      func(d *Definition) { d.Commands[0].Fields = append(d.Commands[0].Fields, d.Commands[0].Fields[0]) },
		"duplicate command":    func(d *Definition) { d.Commands = append(d.Commands, d.Commands[0]) },
		"bad first":            func(d *Definition) { d.Commands[0].Fields[1].First = true },
		"identifier":           func(d *Definition) { d.Package = "package" },
		"cycle":                func(d *Definition) { d.Groups[0].Fields[0].Type = d.Groups[0].Name; d.Groups[0].Fields[0].Values = nil },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			d := definition(t, "s6c")
			mutate(&d)
			if _, err := Generate(d); err == nil {
				t.Fatal("invalid schema accepted")
			}
		})
	}
}

func TestGroupDepthDeterministic(t *testing.T) {
	d := definition(t, "s6c")
	d.Groups = nil
	for i := range 17 {
		name := "G" + string(rune('A'+i))
		typ := "u32"
		if i < 16 {
			typ = "G" + string(rune('A'+i+1))
		}
		d.Groups = append(d.Groups, Group{Name: name, Fields: []Field{{Name: "Value", Code: 1, Flags: 64, Type: typ, Min: 1, Max: 1}}})
	}
	d.Commands = []Command{{Name: "Request", Code: 999, Request: true, Fields: []Field{{Name: "Nested", Code: 2, Flags: 64, Type: "GA", Min: 1, Max: 1}}}}
	for range 20 {
		if _, err := Generate(d); err == nil {
			t.Fatal("deep DAG accepted depending on traversal order")
		}
	}
}
func TestExternalGeneratedConsumer(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	d := Definition{Format: 1, Package: "extension", Profile: "external-test", Version: "1", Sources: []string{"test-only synthetic wire grammar"}, Groups: []Group{{Name: "Group", Fields: []Field{{Name: "Blob", Code: 1, Flags: 64, Type: "octets", Min: 1, Max: 1}, {Name: "Names", Code: 2, Flags: 64, Type: "utf8", Max: -1}, {Name: "Numbers", Code: 3, Flags: 64, Type: "u32", Max: -1}}}}, Commands: []Command{{Name: "Request", Application: 999, Code: 999, Request: true, Fields: []Field{{Name: "One", Code: 10, Flags: 64, Type: "Group", Min: 1, Max: 1}, {Name: "Two", Code: 11, Flags: 64, Type: "Group", Min: 1, Max: 1}, {Name: "Many", Code: 12, Flags: 64, Type: "Group", Max: -1}, {Name: "Optional", Code: 13, Flags: 64, Type: "octets", Max: 1}}}}}
	common := []Field{{Name: "SessionID", Code: 263, Flags: 64, Type: "utf8", Min: 1, Max: 1, First: true, Nonempty: true, Correlation: "session"}, {Name: "OriginHost", Code: 264, Flags: 64, Type: "identity", Min: 1, Max: 1, Envelope: "OriginHost"}, {Name: "OriginRealm", Code: 296, Flags: 64, Type: "identity", Min: 1, Max: 1, Envelope: "OriginRealm"}, {Name: "ProxyInfo", Code: 284, Flags: 64, Type: "proxy", Max: -1, Correlation: "proxy"}}
	d.Groups[0].Fields = append(d.Groups[0].Fields, Field{Name: "Wide", Code: 4, Flags: 64, Type: "u64", Max: -1}, Field{Name: "OptionalWide", Code: 5, Flags: 64, Type: "u64", Max: 1}, Field{Name: "RequiredWide", Code: 6, Flags: 64, Type: "u64", Min: 1, Max: 1})
	d.Groups[0].Fields = append(d.Groups[0].Fields, Field{Name: "Signed", Code: 7, Flags: 64, Type: "i32", Max: -1}, Field{Name: "OptionalSigned", Code: 8, Flags: 64, Type: "i32", Max: 1}, Field{Name: "RequiredSigned", Code: 9, Flags: 64, Type: "i32", Min: 1, Max: 1})
	d.Groups[0].Fields = append(d.Groups[0].Fields, Field{Name: "SignedWide", Code: 14, Flags: 64, Type: "i64", Max: -1}, Field{Name: "OptionalSignedWide", Code: 15, Flags: 64, Type: "i64", Max: 1}, Field{Name: "RequiredSignedWide", Code: 16, Flags: 64, Type: "i64", Min: 1, Max: 1})
	d.Commands[0].Fields = append(append([]Field(nil), common...), d.Commands[0].Fields...)
	d.Commands = append(d.Commands, Command{Name: "Answer", Application: 999, Code: 999, ErrorAnswer: true, Fields: append(append([]Field(nil), common...), Field{Name: "ResultCode", Code: 268, Flags: 64, Type: "u32", Min: 1, Max: 1})})
	code, err := Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("external extension imports private package")
	}
	files := map[string]string{"go.mod": "module example.org/consumer\n\ngo 1.26.0\n\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "extension_test.go": `package extension_test
import("testing"; ext "example.org/consumer"; "github.com/skill215/diameterkit/codec";"github.com/skill215/diameterkit/schema";"github.com/skill215/diameterkit/pdu";"github.com/skill215/diameterkit/avp")
func TestExternal(t *testing.T){p,e:=ext.Profile();if e!=nil{t.Fatal(e)};b,e:=schema.Compile(p);if e!=nil{t.Fatal(e)};c,e:=codec.New(b,codec.Config{});if e!=nil{t.Fatal(e)}
g:=ext.Group{Blob:[]byte{1},Names:[]string{"one"},Numbers:[]uint32{0,7},Wide:[]uint64{0,1<<63,^uint64(0)},OptionalWide:avp.Some(uint64(0)),RequiredWide:^uint64(0),Signed:[]int32{-2147483648,-1,0,2147483647},OptionalSigned:avp.Some(int32(0)),RequiredSigned:-1,SignedWide:[]int64{-9223372036854775808,-1,0,9223372036854775807},OptionalSignedWide:avp.Some(int64(0)),RequiredSignedWide:-9223372036854775808}
r:=&ext.Request{SessionID:"s;1",OriginHost:"a.example",OriginRealm:"example",ProxyInfo:[]pdu.ProxyInfo{{ProxyHost:"proxy.example",ProxyState:[]byte{0,255}}},One:g,Two:g,Many:[]ext.Group{g},Optional:avp.Optional[[]byte]{Value:[]byte{9}}}
m,e:=c.CompleteRequest(r,pdu.RequestEnvelope{HopByHop:1,EndToEnd:2});if e!=nil{t.Fatal(e)};r.Many[0].Blob[0]=5;r.One.Names[0]="changed";r.One.Numbers[0]=8;r.One.SignedWide[0]=123
out:=m.Body.(*ext.Request);if out.Many[0].Blob[0]!=1||out.One.Names[0]!="one"||out.One.Numbers[0]!=0||out.Optional.Value!=nil||out.One.SignedWide[0]!=-9223372036854775808{t.Fatal("clone or absent value")}
buf,e:=c.Append(nil,m);if e!=nil{t.Fatal(e)};decoded,e:=c.Decode(buf);if e!=nil{t.Fatal(e)};if decoded.Body.(*ext.Request).Many[0].Numbers[1]!=7{t.Fatal("nested decode")}
wg:=decoded.Body.(*ext.Request).Many[0];if wg.Wide[2]!=^uint64(0)||wg.Wide[1]!=1<<63||!wg.OptionalWide.Present||wg.OptionalWide.Value!=0||wg.RequiredWide!=^uint64(0){t.Fatal("u64 precision/presence")}
if wg.Signed[0]!=-2147483648||wg.Signed[1]!=-1||wg.Signed[3]!=2147483647||!wg.OptionalSigned.Present||wg.OptionalSigned.Value!=0||wg.RequiredSigned!=-1{t.Fatal("i32 precision/presence")}
if wg.SignedWide[0]!=-9223372036854775808||wg.SignedWide[1]!=-1||wg.SignedWide[3]!=9223372036854775807||!wg.OptionalSignedWide.Present||wg.OptionalSignedWide.Value!=0||wg.RequiredSignedWide!=-9223372036854775808{t.Fatal("i64 precision/presence")}
local:=pdu.Origin{Host:"b.example",Realm:"example"};reply,e:=c.CompleteAnswer(m,&ext.Answer{ResultCode:5012},local);if e!=nil{t.Fatal(e)}
if e=c.ValidateAnswer(m,reply,pdu.AnswerSources{Terminals:[]pdu.Origin{local}});e!=nil{t.Fatal(e)}
reply,e=c.CompleteAnswer(m,&pdu.ErrorAnswer{ResultCode:3002},local);if e!=nil{t.Fatal(e)};wire,e:=c.Append(nil,reply);if e!=nil{t.Fatal(e)}
if _,e=c.DecodeAnswer(m,wire,pdu.AnswerSources{Terminals:[]pdu.Origin{local}});e!=nil{t.Fatal(e)}
}`}
	for name, data := range files {
		if err = os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("external consumer: %v\n%s", err, out)
	}
}

func TestAnswerRuleContract(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Definition)
	}{
		{"request", func(d *Definition) { d.Commands[0].AnswerRule = "checkPair" }},
		{"identifier", func(d *Definition) { d.Commands[3].AnswerRule = "not-valid" }},
		{"symbol", func(d *Definition) { d.Commands[3].AnswerRule = "decodeTFA" }},
		{"same static rule", func(d *Definition) { d.Commands[3].AnswerRule = d.Commands[3].Rule }},
		{"missing pair", func(d *Definition) { d.Commands = append(d.Commands[:2], d.Commands[3]) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := definition(t, "sgd")
			tc.change(&d)
			if _, e := Generate(d); e == nil {
				t.Fatal("accepted invalid answer rule")
			}
		})
	}
}

func TestFeatureContextAndCapturedMandatoryContracts(t *testing.T) {
	for _, mutate := range []func(*Definition){
		func(d *Definition) { d.Commands[0].NoContext = true },
		func(d *Definition) { d.Commands[0].ContextName = "" },
		func(d *Definition) { d.Commands[0].ContextRule = "" },
		func(d *Definition) { d.Commands[0].ContextRule = "bad-name" },
		func(d *Definition) {
			for i := range d.Commands[0].Fields {
				if d.Commands[0].Fields[i].Type == "SupportedFeatures" {
					d.Commands[0].Fields[i].Flags |= 64
				}
			}
		},
		func(d *Definition) { d.Groups[2].Fields[0].Name = "Mandatory" },
	} {
		d := definition(t, "sh")
		mutate(&d)
		if _, err := Generate(d); err == nil {
			t.Fatal("invalid semantic operation definition accepted")
		}
	}
	d := definition(t, "sh")
	a, err := Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	d.Groups[2].CaptureMandatory = false
	b, err := Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("semantic M contract absent from digest")
	}
}
