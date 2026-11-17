// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: External generated consumers preserve manual evidence through repeated ancestors.
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

func TestManualGeneratedExternal(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	d := Definition{Format: 1, Package: "extension", Profile: "manual-test", Version: "1", Sources: []string{"synthetic repeated manual-group evidence"}, Groups: []Group{{Name: "Container", Fields: []Field{{Name: "Proxies", Code: 284, Flags: 64, Type: "proxy", Max: -1}, {Name: "Applications", Code: 260, Flags: 64, Type: "vsai", Max: -1}}}}, Commands: []Command{{Name: "Request", Application: 999, Code: 999, Request: true, Fields: []Field{{Name: "Containers", Code: 900, Flags: 64, Type: "Container", Max: -1}}}}}
	code, e := Generate(d)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(code, []byte("/internal/")) {
		t.Fatal("internal dependency")
	}
	dir := t.TempDir()
	files := map[string]string{"go.mod": "module example.org/manualconsumer\n\ngo 1.26.0\nrequire github.com/skill215/diameterkit v0.0.0\nreplace github.com/skill215/diameterkit => " + filepath.ToSlash(root) + "\n", "extension.go": string(code), "manual_test.go": `package extension
import("bytes";"errors";"testing";"github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/wire";"github.com/skill215/diameterkit/validation")
func part(k uint32,b []byte)[]byte{out:=make([]byte,(8+len(b)+3)&^3);w:=wire.NewEncoder(out,wire.DefaultLimits());w.Bytes(avp.Key{Code:k},64,b);return out}
func TestManualBridge(t *testing.T){
 for _,kind:=range []string{"host","unknown","width","duplicate","missing"}{t.Run(kind,func(t *testing.T){
  group,leaf:=uint32(260),uint32(266);good:=append(part(266,[]byte{0,0,40,175}),part(258,[]byte{1,0,0,1})...);bad:=part(266,[]byte{1,2,3});result:=uint32(5014)
  switch kind {case "host":group=284;leaf=280;good=append(part(280,[]byte("proxy.example")),part(33,[]byte{0,255})...);bad=part(280,[]byte("bad_name"));result=5004
  case "unknown":leaf=999999;bad=append(append([]byte{},good...),part(leaf,[]byte{0,255,1})...);result=5001
  case "duplicate":bad=append(part(266,[]byte{0,0,0,1}),part(266,[]byte{0,0,0,2})...);result=5009
  case "missing":bad=part(258,[]byte{1,0,0,1});result=5005}
  first:=part(900,part(group,good));siblings:=append(part(group,good),part(group,bad)...);raw:=append(first,part(900,siblings)...)
  _,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()));var field *validation.FieldError
  if !errors.As(e,&field)||len(field.Path)!=3||field.Path[0].Code!=900||field.Path[1].Code!=group||field.Path[2].Code!=leaf {t.Fatal("lost nested path",e)}
  if kind=="missing" {var p interface{DiameterMissingAVPV1()(avp.Key,byte,int,error)};if !errors.As(e,&p)||!errors.Is(e,validation.ErrMissing){t.Fatal("missing bridge",e)};for i,k:=range []uint32{900,260,266}{key,flags,width,child:=p.DiameterMissingAVPV1();if key.Code!=k||flags!=64||i==2&&width!=4||i<2&&width!=0{t.Fatal("omission metadata")};if i<2 {p=child.(interface{DiameterMissingAVPV1()(avp.Key,byte,int,error)})}else if child!=nil{t.Fatal("extra omission node")}};return}
  var p error
  if result==5009 {var x interface{error;DiameterExcessAVPV1()(avp.Key,byte,int,int,error)};if !errors.As(e,&x)||!errors.Is(e,validation.ErrDuplicate){t.Fatal("excess bridge",e)};p=x
  }else{var x interface{error;DiameterInvalidAVPV1()(avp.Key,byte,int,uint32,error)};if !errors.As(e,&x){t.Fatal("invalid bridge",e)};p=x}
  for depth:=0;depth<3;depth++ {var key avp.Key;var flags byte;var offset int;var child error
   if result==5009 {var maximum int;key,flags,offset,maximum,child=p.(interface{DiameterExcessAVPV1()(avp.Key,byte,int,int,error)}).DiameterExcessAVPV1();if depth==2&&maximum!=1||depth<2&&maximum!=0{t.Fatal("maximum")}
   }else{var n uint32;key,flags,offset,n,child=p.(interface{DiameterInvalidAVPV1()(avp.Key,byte,int,uint32,error)}).DiameterInvalidAVPV1();if depth==2&&n!=result||depth<2&&n!=0{t.Fatal("result")}}
   scan:=wire.NewDecoder(raw,wire.DefaultLimits());for{v,ok,err:=scan.Next();if err!=nil||!ok{t.Fatal("invented instance",err)};if v.Offset==offset{if key!=v.Key||flags!=v.Flags{t.Fatal("wrong instance")};if depth<2 {if offset==0{t.Fatal("selected first ancestor")};raw=v.Data;p=child}else{if child!=nil{t.Fatal("extra child")};if result==5014&&p.(interface{DiameterLengthAVPV1()string}).DiameterLengthAVPV1()!="u32"{t.Fatal("lost type")};if kind=="host"&&!bytes.Equal(v.Data,[]byte("bad_name")){t.Fatal("wrong host")}};break}}
  }
 })}
}

func TestGroupBridge(t *testing.T){
 for _,kind:=range []string{"both","neither","host","state"}{t.Run(kind,func(t *testing.T){
 group:=uint32(260);key:=avp.Key{};rule:="vsai-"+kind
 good:=append(part(266,make([]byte,4)),part(258,make([]byte,4))...);bad:=part(266,make([]byte,4))
 if kind=="both"{bad=append(append(bad,part(258,make([]byte,4))...),part(259,make([]byte,4))...)}
 if kind=="host"||kind=="state"{group=284;good=append(part(280,[]byte("p")),part(33,nil)...);rule="proxy-"+kind+"-missing";key.Code=280;bad=part(33,nil);if kind=="state"{key.Code=33;bad=part(280,[]byte("p"))}}
 raw:=append(part(900,part(group,good)),part(900,append(part(group,good),part(group,bad)...))...)
 _,e:=decodeRequest(wire.NewDecoder(raw,wire.DefaultLimits()));var field *validation.FieldError;var p interface{DiameterGroupAVPV1()(avp.Key,byte,int,string,error)}
 if !errors.As(e,&field)||!errors.As(e,&p){t.Fatal("missing evidence",e)}
 size:=2;if group==284{size++};if len(field.Path)!=size{t.Fatal("path",field.Path)}
 for depth:=0;depth<2;depth++{k,flags,offset,r,child:=p.DiameterGroupAVPV1();if k!=field.Path[depth]||flags!=64||offset==0||r!=""||child==nil{t.Fatal("parent metadata")};d:=wire.NewDecoder(raw,wire.DefaultLimits());for{v,ok,e:=d.Next();if e!=nil||!ok{t.Fatal("wrong offset")};if v.Offset==offset{if v.Key!=k{t.Fatal("key")};raw=v.Data;break}};p=child.(interface{DiameterGroupAVPV1()(avp.Key,byte,int,string,error)})}
 k,flags,offset,r,child:=p.DiameterGroupAVPV1();if k!=key||flags!=0||offset!=0||r!=rule||child!=nil||!bytes.Equal(raw,bad){t.Fatal("terminal")}
 })}
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
		t.Fatalf("external manual consumer: %v\n%s", e, out)
	}
}
