// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: External generated typed-length metadata contract acceptance.
// SPDX-AI-Date: 2026-11-17

package codegen

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestLengthGeneratedExternal(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	var fields []Field
	for i, kind := range []string{"u32", "i32", "time", "u64", "i64", "tbcd", "address"} {
		fields = append(fields, Field{Name: "Value" + string(rune('A'+i)), Code: uint32(450 + i), Flags: 64, Type: kind, Max: 1})
	}
	d := Definition{Format: 1, Package: "extension", Profile: "length-test", Version: "1", Sources: []string{"synthetic typed-length acceptance"}, Commands: []Command{{Name: "Request", Application: 999, Code: 999, Request: true, Fields: fields}}}
	code, e := Generate(d)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("private import")
	}
	dir := t.TempDir()
	files := map[string]string{"go.mod": "module example.org/lengthconsumer\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "length_test.go": `package extension
import("testing";"errors";"github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/wire";"github.com/skill215/diameterkit/validation")
func TestTypedLength(t *testing.T){
 types:=[]string{"u32","i32","time","u64","i64","tbcd","address"}
 for i,kind:=range types {for n:=0;n<=20;n++ {
  data:=make([]byte,n);want:=false
  switch kind {case "u32","i32","time":want=n!=4;case "u64","i64":want=n!=8;case "tbcd":want=n==0||n>8;if n==8 {data[7]=0xf0};case "address":want=n!=6;if n>=2 {data[1]=1}}
  raw:=make([]byte,128);w:=wire.NewEncoder(raw,wire.DefaultLimits());k:=avp.Key{Code:uint32(450+i)};w.Bytes(k,64,data);if w.Err()!=nil {t.Fatal(w.Err())}
  _,e:=decodeRequest(wire.NewDecoder(raw[:w.Len()],wire.DefaultLimits()))
  if !want {if e!=nil {t.Fatal(kind,n,e)};continue}
  if !errors.Is(e,avp.ErrValue) {t.Fatal("sentinel",kind,n,e)}
  var field *validation.FieldError;if !errors.As(e,&field)||len(field.Path)!=1||field.Path[0]!=k {t.Fatal("path",e)}
  var proof interface{DiameterInvalidAVPV1()(avp.Key,byte,int,uint32,error);DiameterLengthAVPV1()string}
  if !errors.As(e,&proof) {t.Fatal("no typed proof",e)}
  got,flags,offset,result,child:=proof.DiameterInvalidAVPV1()
  if got!=k||flags!=64||offset!=0||result!=5014||child!=nil||proof.DiameterLengthAVPV1()!=kind {t.Fatal("wrong proof")}
 }}
 // Unknown Address families remain opaque; invalid family/domain values must
 // not be relabeled as a type-width mismatch.
 for _,data:=range [][]byte{{0,99,1},{0,99},{0,0,1}} {
  raw:=make([]byte,32);w:=wire.NewEncoder(raw,wire.DefaultLimits());w.Bytes(avp.Key{Code:456},64,data)
  _,e:=decodeRequest(wire.NewDecoder(raw[:w.Len()],wire.DefaultLimits()))
  if len(data)==3&&data[1]==99 {if e!=nil {t.Fatal(e)};continue}
  var proof interface{DiameterInvalidAVPV1()(avp.Key,byte,int,uint32,error)}
  if !errors.Is(e,avp.ErrValue)||errors.As(e,&proof) {t.Fatal("invented Address length evidence",e)}
 }
}
`}
	for name, data := range files {
		if e = os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); e != nil {
			t.Fatal(e)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "CGO_ENABLED=0")
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("external typed-length consumer: %v\n%s", e, out)
	}
}
