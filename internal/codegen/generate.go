// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package codegen compiles the first bounded JSON definition format to Go.
// It is a build-time tool; no production codec imports it.
package codegen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"go/token"
	"io"
	"sort"
	"strings"
)

const maxDefinition = 4 << 20
const GeneratorVersion = "20"

type Definition struct {
	Format   int       `json:"format"`
	Package  string    `json:"package"`
	Profile  string    `json:"profile"`
	Version  string    `json:"version"`
	Sources  []string  `json:"sources"`
	Groups   []Group   `json:"groups"`
	Commands []Command `json:"commands"`
}
type Group struct {
	DiagnosticMinimum bool    `json:"diagnostic_minimum,omitempty"`
	CaptureMandatory  bool    `json:"capture_mandatory,omitempty"`
	Name              string  `json:"name"`
	Fields            []Field `json:"fields"`
	Rule              string  `json:"rule,omitempty"`
}
type Command struct {
	ErrorRule       string  `json:"error_rule,omitempty"`
	ContextRule     string  `json:"context_rule,omitempty"`
	ContextName     string  `json:"context_name,omitempty"`
	NoContext       bool    `json:"no_context,omitempty"`
	TransportReplay bool    `json:"transport_replay,omitempty"`
	Name            string  `json:"name"`
	Application     uint32  `json:"application"`
	Code            uint32  `json:"code"`
	Request         bool    `json:"request"`
	Proxiable       bool    `json:"proxiable"`
	Fields          []Field `json:"fields"`
	Rule            string  `json:"rule,omitempty"`
	AnswerRule      string  `json:"answer_rule,omitempty"`
	ErrorAnswer     bool    `json:"error_answer,omitempty"`
	IgnoreVSAI      bool    `json:"ignore_vsai,omitempty"`
}
type Field struct {
	Name        string   `json:"name"`
	Code        uint32   `json:"code"`
	Vendor      uint32   `json:"vendor"`
	Flags       byte     `json:"flags"`
	Type        string   `json:"type"`
	Min         int      `json:"min"`
	Max         int      `json:"max"`
	First       bool     `json:"first,omitempty"`
	Envelope    string   `json:"envelope,omitempty"`
	Values      []uint32 `json:"values,omitempty"`
	Nonempty    bool     `json:"nonempty,omitempty"`
	Correlation string   `json:"correlation,omitempty"`
}

// uniqueJSON validates duplicate keys before encoding/json can silently overwrite them.
func uniqueJSON(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("definition nesting limit")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("null is not a definition value")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, err = d.Token()
			if err != nil {
				return err
			}
			k, ok := t.(string)
			if !ok || seen[k] || strings.ToLower(k) != k {
				return fmt.Errorf("duplicate/noncanonical key %q", k)
			}
			seen[k] = true
			if err = uniqueJSON(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err = uniqueJSON(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter")
	}
	_, err = d.Token()
	return err
}
func Parse(r io.Reader) (Definition, error) {
	b, err := io.ReadAll(io.LimitReader(r, maxDefinition+1))
	if err != nil {
		return Definition{}, err
	}
	if len(b) > maxDefinition {
		return Definition{}, fmt.Errorf("definition too large")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	if err = uniqueJSON(d, 0); err != nil {
		return Definition{}, err
	}
	if _, err = d.Token(); err != io.EOF {
		return Definition{}, fmt.Errorf("trailing JSON")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var def Definition
	if err = d.Decode(&def); err != nil {
		return def, err
	}
	return def, validate(def)
}
func identifier(s string) bool {
	return s != "_" && token.IsIdentifier(s) && token.Lookup(s) == token.IDENT
}
func exported(s string) bool { return identifier(s) && s[0] >= 'A' && s[0] <= 'Z' }
func primitive(s string) bool {
	switch s {
	case "utf8", "identity", "u32", "u64", "i32", "i64", "time", "octets", "tbcd", "address", "proxy", "diagnostic", "vsai":
		return true
	}
	return false
}
func validate(d Definition) error {
	if d.Format != 1 || !identifier(d.Package) || d.Profile == "" || d.Version == "" || len(d.Sources) == 0 || len(d.Commands) == 0 || len(d.Commands) > 64 || len(d.Groups) > 64 {
		return fmt.Errorf("invalid definition header")
	}
	names := map[string]bool{"Profile": true, "DefinitionDigest": true, "Version": true, "profileDiagnosticGroups": true, "profileErrorUnsupported": true, "profileOmission": true, "profileMissing": true, "profileWithin": true, "profileExcess": true, "profileTooMany": true, "profileInvalid": true, "profileBadValue": true, "profileBadLength": true, "profileConstraint": true, "profileGroupOmission": true, "profileMissingGroup": true}
	groups := map[string]Group{}
	for _, g := range d.Groups {
		if !exported(g.Name) || names[g.Name] {
			return fmt.Errorf("invalid/duplicate group %q", g.Name)
		}
		names[g.Name] = true
		groups[g.Name] = g
	}
	check := func(name, rule string, fs []Field) error {
		if len(fs) == 0 || len(fs) > 64 || rule != "" && !identifier(rule) {
			return fmt.Errorf("invalid fields/rule for %s", name)
		}
		seenName := map[string]bool{"Extra": true, "CommandKey": true, "Mandatory": true}
		seenKey := map[string]bool{}
		for i, f := range fs {
			key := fmt.Sprintf("%d/%d", f.Vendor, f.Code)
			if !exported(f.Name) || seenName[f.Name] || seenKey[key] || f.Code == 0 || f.Min < 0 || f.Min > 256 || f.Max == 0 || f.Max < -1 || f.Max > 256 || f.Max != -1 && f.Max < f.Min || f.First && (i != 0 || f.Min != 1 || f.Max != 1) {
				return fmt.Errorf("invalid field %s.%s", name, f.Name)
			}
			if f.Flags&0x1f != 0 || (f.Vendor != 0) != (f.Flags&0x80 != 0) {
				return fmt.Errorf("conflicting Vendor/flags in %s.%s", name, f.Name)
			}
			if groups[f.Type].CaptureMandatory && f.Flags&0x40 != 0 {
				return fmt.Errorf("captured M field must have M0 base flags")
			}
			if !primitive(f.Type) {
				if _, ok := groups[f.Type]; !ok {
					return fmt.Errorf("unresolved type %s", f.Type)
				}
			}
			if len(f.Values) > 0 && f.Type != "u32" || f.Nonempty && f.Type != "identity" && f.Type != "utf8" && f.Type != "octets" && f.Type != "tbcd" {
				return fmt.Errorf("invalid constraint %s", f.Name)
			}
			values := make(map[uint32]bool, len(f.Values))
			for _, value := range f.Values {
				if values[value] {
					return fmt.Errorf("duplicate allowed value in %s.%s", name, f.Name)
				}
				values[value] = true
			}
			if f.Envelope != "" {
				switch f.Envelope {
				case "OriginHost", "OriginRealm", "DestinationHost", "DestinationRealm":
				default:
					return fmt.Errorf("invalid envelope")
				}
				if f.Type != "identity" || f.Max != 1 {
					return fmt.Errorf("invalid envelope type")
				}
			}
			if f.Correlation != "" {
				expected := uint32(0)
				typ := ""
				switch f.Correlation {
				case "session":
					expected = 263
					typ = "utf8"
				case "auth":
					expected = 277
					typ = "u32"
				case "proxy":
					expected = 284
					typ = "proxy"
				default:
					return fmt.Errorf("unknown correlation field")
				}
				if f.Code != expected || f.Vendor != 0 || f.Type != typ || f.Correlation == "proxy" && f.Max != -1 || f.Correlation != "proxy" && (f.Min != 1 || f.Max != 1) {
					return fmt.Errorf("invalid correlation binding")
				}
			}
			if f.Envelope != "" {
				expected := map[string]uint32{"OriginHost": 264, "OriginRealm": 296, "DestinationHost": 293, "DestinationRealm": 283}
				if f.Vendor != 0 || f.Code != expected[f.Envelope] {
					return fmt.Errorf("invalid envelope identity binding")
				}
			}
			seenName[f.Name] = true
			seenKey[key] = true
		}
		return nil
	}
	for _, g := range d.Groups {
		if err := check(g.Name, g.Rule, g.Fields); err != nil {
			return err
		}
	}
	keys := map[string]bool{}
	for _, c := range d.Commands {
		if c.ErrorRule != "" && (c.Request || !c.ErrorAnswer || !identifier(c.ErrorRule)) {
			return fmt.Errorf("invalid E answer rule")
		}
		if c.TransportReplay && (!c.Request || c.Application == 0) {
			return fmt.Errorf("transport replay requires an application request")
		}
		if c.AnswerRule != "" {
			if c.Request || !identifier(c.AnswerRule) {
				return fmt.Errorf("invalid answer rule")
			}
			found := false
			for _, r := range d.Commands {
				if r.Request && r.Application == c.Application && r.Code == c.Code {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("answer rule requires request in definition")
			}
		}
		k := fmt.Sprintf("%d/%d/%t", c.Application, c.Code, c.Request)
		if !exported(c.Name) || names[c.Name] || c.Code == 0 || c.Code > 0xffffff || keys[k] {
			return fmt.Errorf("invalid/duplicate command %s", c.Name)
		}
		if c.Request && c.ErrorAnswer {
			return fmt.Errorf("E grammar belongs to answer")
		}
		for _, f := range c.Fields {
			if !c.Request && f.Vendor == 0 && (f.Code == 283 || f.Code == 293) {
				return fmt.Errorf("Destination forbidden in answers")
			}
			if c.IgnoreVSAI && f.Vendor == 0 && f.Code == 260 {
				return fmt.Errorf("ignored VSAI cannot be a semantic field")
			}
		}
		names[c.Name] = true
		keys[k] = true
		if err := check(c.Name, c.Rule, c.Fields); err != nil {
			return err
		}
	}
	active := map[string]bool{}
	heights := map[string]int{}
	var visit func(string) error
	visit = func(n string) error {
		if active[n] {
			return fmt.Errorf("recursive Grouped %s", n)
		}
		if heights[n] > 0 {
			return nil
		}
		active[n] = true
		h := 1
		for _, f := range groups[n].Fields {
			if !primitive(f.Type) {
				if err := visit(f.Type); err != nil {
					return err
				}
				h = max(h, 1+heights[f.Type])
			}
		}
		if h > 16 {
			return fmt.Errorf("Grouped depth limit: %s", n)
		}
		delete(active, n)
		heights[n] = h
		return nil
	}
	for n := range groups {
		if err := visit(n); err != nil {
			return err
		}
	}
	for _, g := range d.Groups {
		if g.DiagnosticMinimum {
			if _, err := minimumPlan(Field{Code: 1, Type: g.Name}, groups); err != nil {
				return err
			}
		}
	}
	symbols := map[string]bool{"Profile": true, "Version": true, "DefinitionDigest": true}
	for name := range names {
		symbols[name] = true
		for _, prefix := range []string{"validate", "encode", "decode", "clone", "complete", "envelope", "answer"} {
			symbols[prefix+name] = true
		}
	}
	for _, g := range d.Groups {
		if symbols[g.Rule] {
			return fmt.Errorf("rule collides with generated symbol")
		}
		for _, f := range g.Fields {
			if f.Envelope != "" || f.Correlation != "" {
				return fmt.Errorf("envelope only supported at command root")
			}
		}
	}
	for _, c := range d.Commands {
		if (c.ContextRule == "") != (c.ContextName == "") || c.ContextRule != "" && (!identifier(c.ContextRule) || !identifier(c.ContextName) || c.NoContext || symbols[c.ContextRule]) {
			return fmt.Errorf("invalid context contract")
		}
		if symbols[c.Rule] || symbols[c.AnswerRule] || symbols[c.ErrorRule] || c.AnswerRule != "" && c.AnswerRule == c.Rule {
			return fmt.Errorf("rule collides with generated symbol")
		}
	}
	if structureBound(d) > 1<<20 {
		return fmt.Errorf("inline structure limit")
	}
	return nil
}
func key(f Field) string { return fmt.Sprintf("avp.Key{VendorID:%d,Code:%d}", f.Vendor, f.Code) }
func keyCommand(c Command) string {
	return fmt.Sprintf("pdu.CommandKey{ApplicationID:%d,Code:%d,Request:%t}", c.Application, c.Code, c.Request)
}
func elem(f Field) string {
	switch f.Type {
	case "utf8", "identity":
		return "string"
	case "octets", "tbcd":
		return "[]byte"
	case "u32":
		return "uint32"
	case "u64":
		return "uint64"
	case "i32":
		return "int32"
	case "i64":
		return "int64"
	case "time":
		return "avp.Time"
	case "address":
		return "avp.Address"
	case "proxy":
		return "pdu.ProxyInfo"
	case "diagnostic":
		return "pdu.FailedAVP"
	case "vsai":
		return "pdu.VendorSpecificApplicationID"
	}
	return f.Type
}
func optional(f Field) bool { return f.Max == 1 && f.Min == 0 }
func repeated(f Field) bool { return f.Max != 1 }
func fieldType(f Field) string {
	t := elem(f)
	if repeated(f) {
		return "[]" + t
	}
	if optional(f) {
		return "avp.Optional[" + t + "]"
	}
	return t
}

type writer struct {
	bytes.Buffer
	groups             map[string]Group
	capture            map[string]bool
	prefix             string
	answer, ignoreVSAI bool
}

func (w *writer) p(s string, args ...any) { fmt.Fprintf(&w.Buffer, s+"\n", args...) }
func (w *writer) each(f Field, fn func(string)) {
	v := "m." + f.Name
	if repeated(f) {
		w.p("for _,v:=range %s {", v)
		fn("v")
		w.p("}")
	} else if optional(f) {
		w.p("if %s.Present {", v)
		fn(v + ".Value")
		w.p("}")
	} else {
		fn(v)
	}
}
func Generate(d Definition) ([]byte, error) {
	if err := validate(d); err != nil {
		return nil, err
	}
	// Slice order specifies deterministic encoding order; JSON object key order does not.
	canonical, _ := json.Marshal(d)
	sum := sha256.Sum256(append([]byte("diameter-schema/"+GeneratorVersion+"\n"), canonical...))
	digest := hex.EncodeToString(sum[:])
	w := &writer{capture: map[string]bool{}, groups: map[string]Group{}}
	for _, g := range d.Groups {
		w.capture[g.Name] = g.CaptureMandatory
		w.groups[g.Name] = g
	}
	if d.Package != "base" {
		w.prefix = "base."
	}
	w.p("// SPDX-AI-Disclosure: ai-generated")
	w.p("// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md")
	w.p("// SPDX-AI-Provider: OpenAI")
	w.p("// SPDX-AI-Scope: Generated protocol declarations from project-authored definitions and generator.")
	w.p("// SPDX-AI-Date: 2026-11-17")
	w.p("")
	w.p("// Code generated by diameter-schema v%s; DO NOT EDIT.", GeneratorVersion)
	w.p("package %s", d.Package)
	w.p(`import ("github.com/skill215/diameterkit/avp";"github.com/skill215/diameterkit/pdu";"github.com/skill215/diameterkit/schema";"github.com/skill215/diameterkit/validation";"github.com/skill215/diameterkit/wire")`)
	needsBase := false
	for _, c := range d.Commands {
		if !c.Request {
			needsBase = true
		}
		for _, f := range c.Fields {
			if f.Type == "proxy" || f.Type == "diagnostic" || f.Type == "vsai" {
				needsBase = true
			}
		}
	}
	for _, g := range d.Groups {
		for _, f := range g.Fields {
			if f.Type == "proxy" || f.Type == "diagnostic" || f.Type == "vsai" {
				needsBase = true
			}
		}
	}
	if needsBase && w.prefix != "" {
		w.p(`import "github.com/skill215/diameterkit/base"`)
	}
	{
		w.p(`import "errors"`)
		// Keep generated extensions consumable outside this module: the error
		// carries versioned structural metadata, not an internal-package import.
		w.p(`type profileOmission struct { key avp.Key; flags byte; width int; child error }
func (*profileOmission) Error() string {return validation.ErrMissing.Error()}
func (*profileOmission) Unwrap() error {return validation.ErrMissing}
func (e *profileOmission) DiameterMissingAVPV1() (avp.Key,byte,int,error) {
 if e.child==nil {return e.key,e.flags,e.width,nil};return e.key,e.flags,e.width,e.child
}
func profileMissing(key avp.Key,flags byte,width int) error {
 if width!=4&&width!=8{return validation.At(key,validation.ErrMissing)}
 return validation.At(key,&profileOmission{key:key,flags:flags,width:width})
}
type profileGroupOmission struct {key avp.Key;flags byte;plan string;child error}
func (*profileGroupOmission) Error()string{return validation.ErrMissing.Error()}
func (*profileGroupOmission) Unwrap()error{return validation.ErrMissing}
func(e *profileGroupOmission) DiameterMissingGroupV1()(avp.Key,byte,string,error){return e.key,e.flags,e.plan,e.child}
func profileMissingGroup(key avp.Key,flags byte,plan string)error{return validation.At(key,&profileGroupOmission{key:key,flags:flags,plan:plan})}
func profileWithin(v wire.View,err error) error {
 var tree interface{error;DiameterMissingGroupV1()(avp.Key,byte,string,error)}
 if errors.As(err,&tree){outer:=validation.At(v.Key,err).(*validation.FieldError);outer.Cause=&profileGroupOmission{key:v.Key,flags:v.Flags,child:tree};return outer}

 var group interface {error; DiameterGroupAVPV1() (avp.Key,byte,int,string,error)}
 if errors.As(err,&group) {
  outer:=validation.At(v.Key,err).(*validation.FieldError)
  outer.Cause=&profileConstraint{key:v.Key,flags:v.Flags,offset:v.Offset,cause:err,child:group}
  return outer
 }
 var invalid interface {error; DiameterInvalidAVPV1() (avp.Key,byte,int,uint32,error)}
 if errors.As(err,&invalid) {
  outer:=validation.At(v.Key,err).(*validation.FieldError)
  outer.Cause=&profileInvalid{key:v.Key,flags:v.Flags,offset:v.Offset,cause:err,child:invalid}
  return outer
 }
 var excess interface {error; DiameterExcessAVPV1() (avp.Key,byte,int,int,error)}
 if errors.As(err,&excess) {
  outer:=validation.At(v.Key,err).(*validation.FieldError)
  outer.Cause=&profileExcess{key:v.Key,flags:v.Flags,offset:v.Offset,child:excess}
  return outer
 }
 var proof interface {error; DiameterMissingAVPV1() (avp.Key,byte,int,error)}
 if !errors.As(err,&proof){return validation.At(v.Key,err)}
 outer:=validation.At(v.Key,err).(*validation.FieldError)
 outer.Cause=&profileOmission{key:v.Key,flags:v.Flags,child:proof}
 return outer
}
type profileConstraint struct {key avp.Key; flags byte; offset int; cause,child error}
func (e *profileConstraint) Error() string {return e.cause.Error()}
func (e *profileConstraint) Unwrap() error {return e.cause}
func (e *profileConstraint) DiameterGroupAVPV1() (avp.Key,byte,int,string,error) {return e.key,e.flags,e.offset,"",e.child}
type profileExcess struct { key avp.Key; flags byte; offset,maximum int; child error }
func (*profileExcess) Error() string {return validation.ErrDuplicate.Error()}
func (*profileExcess) Unwrap() error {return validation.ErrDuplicate}
func (e *profileExcess) DiameterExcessAVPV1() (avp.Key,byte,int,int,error) {
 if e.child==nil {return e.key,e.flags,e.offset,e.maximum,nil};return e.key,e.flags,e.offset,e.maximum,e.child
}
type profileInvalid struct { key avp.Key; flags byte; offset int; code uint32; cause error; child error; lengthType string }
func (e *profileInvalid) Error() string {return e.cause.Error()}
func (e *profileInvalid) Unwrap() error {return e.cause}
func (e *profileInvalid) DiameterInvalidAVPV1() (avp.Key,byte,int,uint32,error) {
 if e.child==nil {return e.key,e.flags,e.offset,e.code,nil};return e.key,e.flags,e.offset,e.code,e.child
}
func (e *profileInvalid) DiameterLengthAVPV1() string {return e.lengthType}
func profileBadLength(v wire.View,kind string,cause error) error {
 return validation.At(v.Key,&profileInvalid{key:v.Key,flags:v.Flags,offset:v.Offset,code:5014,cause:cause,lengthType:kind})
}
func profileBadValue(v wire.View,code uint32,cause error) error {
 return validation.At(v.Key,&profileInvalid{key:v.Key,flags:v.Flags,offset:v.Offset,code:code,cause:cause})
}
func profileTooMany(v wire.View,maximum int) error {
 return validation.At(v.Key,&profileExcess{key:v.Key,flags:v.Flags,offset:v.Offset,maximum:maximum})
}`)
	}
	w.p("const Version = %q", d.Version)
	w.p("const DefinitionDigest = %q", digest)
	w.p("func profileErrorUnsupported() []avp.Key {return []avp.Key{")
	for _, c := range d.Commands {
		for _, f := range c.Fields {
			common := false
			if f.Vendor == 0 {
				switch f.Code {
				case 263, 264, 296, 268, 278, 281, 294, 279, 297, 284, 277, 260, 282, 283, 293:
					common = true
				}
			}
			if !common {
				w.p("%s,", key(f))
			}
		}
	}
	w.p("}}")

	w.p("func profileDiagnosticGroups() []avp.Key {return []avp.Key{{Code:279},{Code:284},{Code:297},{Code:260},")
	for _, c := range d.Commands {
		for _, f := range c.Fields {
			if !primitive(f.Type) {
				w.p("%s,", key(f))
			}
		}
	}
	for _, g := range d.Groups {
		for _, f := range g.Fields {
			if !primitive(f.Type) {
				w.p("%s,", key(f))
			}
		}
	}
	w.p("}}")
	for _, g := range d.Groups {
		w.structure(g.Name, g.Fields, g.Rule)
	}
	for _, c := range d.Commands {
		w.answer = !c.Request
		w.ignoreVSAI = c.IgnoreVSAI
		w.structure(c.Name, c.Fields, c.Rule)
		w.envelope(c)
		w.p("func (*%s) CommandKey() pdu.CommandKey {return %s}", c.Name, keyCommand(c))
		w.complete(c)
	}
	w.p("func Profile() (schema.Profile,error) {return schema.NewProfile(%q,Version,DefinitionDigest,", d.Profile)
	for _, c := range d.Commands {
		w.p("schema.Command{Key:%s,Proxiable:%t,TransportReplay:%t,NoContext:%t,", keyCommand(c), c.Proxiable, c.TransportReplay, c.NoContext)
		if c.ContextRule != "" {
			w.p("CheckContext:%s,ContextName:%q,", c.ContextRule, c.ContextName)
		}
		w.p("Decode:func(d *wire.Decoder)(pdu.Body,error){return decode%s(d)},", c.Name)
		w.p("Encode:func(w *wire.Encoder,b pdu.Body)error{m,ok:=b.(*%s);if !ok||m==nil{return validation.ErrType};return encode%s(w,m)},", c.Name, c.Name)
		w.p("Validate:func(b pdu.Body,mode validation.Mode)error{m,ok:=b.(*%s);if !ok||m==nil{return validation.ErrType};return validate%s(m,mode)},", c.Name, c.Name)
		w.p("Clone:func(b pdu.Body)(pdu.Body,error){m,ok:=b.(*%s);if !ok||m==nil{return nil,validation.ErrType};return clone%s(m),nil},", c.Name, c.Name)
		w.p("Envelope:envelope%s,", c.Name)
		if c.Request {
			w.p("CompleteRequest:complete%s,", c.Name)
			w.p("Canonical:func(src []byte,l wire.Limits,maxBytes int)([]byte,error){return wire.CanonicalRequest(src,l,maxBytes,")
			w.canonicalFields(d, c.Fields)
			w.p("%t)},", c.IgnoreVSAI)
		} else {
			w.p("CompleteAnswer:answer%s,", c.Name)
		}
		if c.AnswerRule != "" {
			for _, r := range d.Commands {
				if r.Request && r.Application == c.Application && r.Code == c.Code {
					w.p("CheckAnswer:func(request,answer pdu.Body)error{r,ok:=request.(*%s);if !ok||r==nil{return validation.ErrType};a,ok:=answer.(*%s);if !ok||a==nil{return validation.ErrType};return %s(r,a)},", r.Name, c.Name, c.AnswerRule)
				}
			}
		}
		if c.ErrorAnswer {
			auth := "avp.Optional[uint32]{}"
			for _, f := range c.Fields {
				if f.Correlation == "auth" && len(f.Values) == 1 {
					auth = fmt.Sprintf("avp.Some(uint32(%d))", f.Values[0])
				}
			}
			rule := c.ErrorRule
			if rule == "" {
				rule = "nil"
			}
			w.p("Error:%sNewErrorOperations(%s,%sErrorOptions{DiagnosticGroups:profileDiagnosticGroups(),AuthSessionState:%s,IgnoreVSAI:%t,UnsupportedFields:profileErrorUnsupported(),Rule:%s}),", w.prefix, keyCommand(c), w.prefix, auth, c.IgnoreVSAI, rule)
		}
		w.p("DecodeBound:func(n int,l wire.Limits)(uint64,error){return wire.DecodeBound(n,l,%d)},},", structureBound(d))
	}
	w.p(")}")
	out, err := format.Source(w.Bytes())
	if err != nil {
		return nil, fmt.Errorf("generated syntax: %w", err)
	}
	return out, nil
}
func structureBound(d Definition) int {
	groups := map[string]Group{}
	for _, g := range d.Groups {
		groups[g.Name] = g
	}
	var bound func([]Field) int
	bound = func(fs []Field) int {
		n := 256
		for _, f := range fs {
			n += 64
			if !primitive(f.Type) && !repeated(f) {
				n += bound(groups[f.Type].Fields)
				if n > 1<<20 {
					return n
				}
			}
		}
		return n
	}
	n := 256
	for _, g := range d.Groups {
		n = max(n, bound(g.Fields))
	}
	for _, c := range d.Commands {
		n = max(n, bound(c.Fields))
	}
	return n
}
func (w *writer) structure(name string, fs []Field, rule string) {
	w.p("type %s struct {", name)
	if w.capture[name] {
		w.p("Mandatory bool // Exact outer M bit, semantically significant for feature negotiation.")
	}
	for _, f := range fs {
		w.p("%s %s", f.Name, fieldType(f))
	}
	w.p("Extra []avp.Raw\n}")
	w.p("func validate%s(m *%s,mode validation.Mode) error {if m==nil{return validation.ErrType}", name, name)
	for _, f := range fs {
		k := key(f)
		if repeated(f) {
			w.p("if len(m.%s)<%d {return validation.At(%s,validation.ErrMissing)}", f.Name, f.Min, k)
			if f.Max != -1 {
				w.p("if len(m.%s)>%d {return validation.At(%s,validation.ErrDuplicate)}", f.Name, f.Max, k)
			}
		}
		if f.Type == "i64" || f.Type == "i32" || f.Type == "u64" || f.Type == "time" || f.Type == "u32" && len(f.Values) == 0 || f.Type == "octets" && !f.Nonempty {
			continue
		}
		if f.Envelope != "" || w.answer && f.Correlation == "session" {
			w.p("if !(mode==validation.Build && %s==\"\") {", func() string {
				if optional(f) {
					return "m." + f.Name + ".Value"
				}
				return "m." + f.Name
			}())
		}
		w.each(f, func(v string) {
			switch f.Type {
			case "identity":
				w.p("if err:=avp.Identity(%s);err!=nil{return validation.At(%s,err)}", v, k)
			case "utf8":
				w.p("if err:=avp.UTF8(%s);err!=nil{return validation.At(%s,err)}", v, k)
			case "tbcd":
				w.p("if err:=avp.CheckTBCD(%s);err!=nil{return validation.At(%s,err)}", v, k)
			case "address":
				w.p("if err:=%s.Validate();err!=nil{return validation.At(%s,err)}", v, k)
			case "u32":
				if len(f.Values) > 0 {
					var vals []string
					for _, n := range f.Values {
						vals = append(vals, fmt.Sprint(n))
					}
					w.p("switch %s {case %s:default:return validation.At(%s,validation.ErrField)}", v, strings.Join(vals, ","), k)
				}
			case "octets":
			case "proxy", "vsai", "diagnostic":
				name := map[string]string{"proxy": "Proxy", "vsai": "VSAI", "diagnostic": "Diagnostic"}[f.Type]
				w.p("if err:=%sValidate%s(&%s,mode);err!=nil{return validation.At(%s,err)}", w.prefix, name, v, k)
			default:
				w.p("if err:=validate%s(&%s,mode);err!=nil{return validation.At(%s,err)}", f.Type, v, k)
			}
			if f.Nonempty {
				w.p("if len(%s)==0{return validation.At(%s,validation.ErrField)}", v, k)
			}
		})
		if f.Envelope != "" || w.answer && f.Correlation == "session" {
			w.p("}")
		}
	}
	w.p("for _,v:=range m.Extra {")
	if w.ignoreVSAI {
		w.p("if v.Key==(avp.Key{Code:260}){if mode==validation.Receive{continue};return validation.ErrUnsupported}")
	}
	w.p("switch v.Key {")
	if w.answer {
		w.p("case avp.Key{Code:283},avp.Key{Code:293}:return validation.At(v.Key,validation.ErrField)")
	}
	for _, f := range fs {
		w.p("case %s:return validation.At(v.Key,validation.ErrDuplicate)", key(f))
	}
	w.p("};if v.Flags&avp.MandatoryFlag!=0{return validation.At(v.Key,validation.ErrUnsupported)};if mode!=validation.Receive&&(v.Flags&0x1f!=0 || (v.Key.VendorID!=0)!=(v.Flags&avp.VendorFlag!=0)){return validation.At(v.Key,wire.ErrAVP)}}")
	if rule != "" {
		w.p("return %s(m,mode)", rule)
	} else {
		w.p("return nil")
	}
	w.p("}")
	w.p("func encode%s(w *wire.Encoder,m *%s) error {", name, name)
	for _, f := range fs {
		if repeated(f) {
			w.p("if len(m.%s)>w.Limits().MaxRepeated{return wire.ErrLimit}", f.Name)
		}
		w.each(f, func(v string) {
			switch f.Type {
			case "identity", "utf8":
				w.p("w.String(%s,%d,%s)", key(f), f.Flags, v)
			case "octets", "tbcd":
				w.p("w.Bytes(%s,%d,%s)", key(f), f.Flags, v)
			case "u32":
				w.p("w.Uint32(%s,%d,%s)", key(f), f.Flags, v)
			case "i64":
				w.p("w.Int64(%s,%d,%s)", key(f), f.Flags, v)
			case "i32":
				w.p("w.Int32(%s,%d,%s)", key(f), f.Flags, v)
			case "u64":
				w.p("w.Uint64(%s,%d,%s)", key(f), f.Flags, v)
			case "time":
				w.p("w.Uint32(%s,%d,uint32(%s))", key(f), f.Flags, v)
			case "address":
				w.p("w.Address(%s,%d,%s)", key(f), f.Flags, v)
			case "diagnostic":
				w.p("w.Diagnostic(%s,%d,%s.Data,profileDiagnosticGroups())", key(f), f.Flags, v)
			case "proxy", "vsai":
				name := map[string]string{"proxy": "Proxy", "vsai": "VSAI"}[f.Type]
				w.p("{mark:=w.BeginGroup(%s,%d);if err:=%sEncode%s(w,&%s);err!=nil{return validation.At(%s,err)};w.EndGroup(mark)}", key(f), f.Flags, w.prefix, name, v, key(f))
			default:
				flags := fmt.Sprint(f.Flags)
				if w.capture[f.Type] {
					w.p("{flags:=byte(%d);if %s.Mandatory {flags |= avp.MandatoryFlag}", f.Flags, v)
					flags = "flags"
				}
				w.p("{mark:=w.BeginGroup(%s,%s);if err:=encode%s(w,&%s);err!=nil{return validation.At(%s,err)};w.EndGroup(mark)}", key(f), flags, f.Type, v, key(f))
				if w.capture[f.Type] {
					w.p("}")
				}
			}
			w.p("if err:=w.Err();err!=nil{return validation.At(%s,err)}", key(f))
		})
	}
	w.p("if len(m.Extra)>w.Limits().MaxRepeated{return wire.ErrLimit};for _,v:=range m.Extra{w.Bytes(v.Key,v.Flags,v.Data)};return w.Err()\n}")
	w.decode(name, fs)
	w.p("func clone%s(m *%s) *%s {c:=*m;c.Extra=make([]avp.Raw,len(m.Extra));for i,v:=range m.Extra{c.Extra[i]=v.Clone()}", name, name, name)
	for _, f := range fs {
		base := "c." + f.Name
		deep := f.Type == "utf8" || f.Type == "identity" || f.Type == "octets" || f.Type == "tbcd" || f.Type == "address" || f.Type == "proxy" || f.Type == "diagnostic" || f.Type == "vsai" || !primitive(f.Type)
		if repeated(f) {
			w.p("%s=append([]%s(nil),m.%s...)", base, elem(f), f.Name)
			if deep {
				w.p("for i:=range %s {", base)
				base += "[i]"
			}
		} else if optional(f) {
			w.p("if %s.Present {", base)
			base += ".Value"
		}
		switch f.Type {
		case "utf8", "identity":
			w.p("%s=avp.CloneString(%s)", base, base)
		case "octets", "tbcd":
			w.p("%s=append([]byte(nil),%s...)", base, base)
		case "address":
			w.p("%s.Data=append([]byte(nil),%s.Data...)", base, base)
		case "proxy", "vsai", "diagnostic":
			name := map[string]string{"proxy": "Proxy", "vsai": "VSAI", "diagnostic": "Diagnostic"}[f.Type]
			w.p("%s=*%sClone%s(&%s)", base, w.prefix, name, base)
		default:
			if !primitive(f.Type) {
				w.p("%s=*clone%s(&%s)", base, f.Type, base)
			}
		}
		if repeated(f) && deep {
			w.p("}")
		}
		if optional(f) {
			w.p("} else {c.%s=avp.Optional[%s]{}}", f.Name, elem(f))
		}
	}
	w.p("return &c\n}")
}
func (w *writer) decode(name string, fs []Field) {
	w.p("func decode%s(d *wire.Decoder) (*%s,error) {m:=new(%s);var seen [%d]int;for {v,ok,err:=d.Next();if err!=nil{return nil,err};if !ok{break};switch v.Key {", name, name, name, len(fs))
	for i, f := range fs {
		w.p("case %s:", key(f))
		w.p("seen[%d]++;if seen[%d]>d.Limits().MaxRepeated{return nil,wire.ErrLimit}", i, i)
		if f.Max != -1 {
			w.p("if seen[%d]>%d{return nil,profileTooMany(v,%d)}", i, f.Max, f.Max)
		}
		if f.First {
			w.p("if v.Offset!=0{return nil,validation.At(v.Key,validation.ErrField)}")
		}
		switch f.Type {
		case "identity", "utf8":
			w.p("x:=string(v.Data)")
		case "octets", "tbcd":
			w.p("x:=v.Data")
		case "u32":
			w.p("x,err:=wire.U32(v.Data);if err!=nil{return nil,profileBadLength(v,\"u32\",err)}")
		case "i64":
			w.p("x,err:=wire.I64(v.Data);if err!=nil{return nil,profileBadLength(v,\"i64\",err)}")
		case "i32":
			w.p("x,err:=wire.I32(v.Data);if err!=nil{return nil,profileBadLength(v,\"i32\",err)}")
		case "u64":
			w.p("x,err:=wire.U64(v.Data);if err!=nil{return nil,profileBadLength(v,\"u64\",err)}")
		case "time":
			w.p("seconds,err:=wire.U32(v.Data);if err!=nil{return nil,profileBadLength(v,\"time\",err)};x:=avp.Time(seconds)")
		case "address":
			w.p("x,err:=avp.ParseAddress(v.Data);if err!=nil{if len(v.Data)<2 || v.Data[0]==0 && (v.Data[1]==1 && len(v.Data)!=6 || v.Data[1]==2 && len(v.Data)!=18){return nil,profileBadLength(v,\"address\",err)};return nil,validation.At(v.Key,err)}")
		case "diagnostic":
			w.p("complete,err:=d.Diagnostic(v.Data,profileDiagnosticGroups());if err!=nil{return nil,validation.At(v.Key,err)};x:=pdu.FailedAVP{Data:v.Data,Incomplete:!complete}")
		case "proxy", "vsai":
			name := map[string]string{"proxy": "Proxy", "vsai": "VSAI"}[f.Type]
			w.p("child,err:=d.Child(v.Data);if err!=nil{return nil,validation.At(v.Key,err)};value,err:=%sDecode%s(child);if err!=nil{return nil,profileWithin(v,err)};x:=*value", w.prefix, name)
		default:
			w.p("child,err:=d.Child(v.Data);if err!=nil{return nil,validation.At(v.Key,err)};value,err:=decode%s(child);if err!=nil{return nil,profileWithin(v,err)};x:=*value", f.Type)
			if w.capture[f.Type] {
				w.p("x.Mandatory=v.Flags&avp.MandatoryFlag!=0")
			}
		}
		// Validate isolated values while the actual wire instance is known.
		// Typed-length errors gain 5014; manual cross-field rules gain no mapping.
		switch f.Type {
		case "identity":
			w.p("if err:=avp.Identity(x);err!=nil{return nil,profileBadValue(v,5004,err)}")
		case "utf8":
			w.p("if err:=avp.UTF8(x);err!=nil{return nil,profileBadValue(v,5004,err)}")
		case "tbcd":
			w.p("if err:=avp.CheckTBCD(x);err!=nil{if len(x)==0||len(x)>8{return nil,profileBadLength(v,\"tbcd\",err)};return nil,profileBadValue(v,5004,err)}")
		case "u32":
			if len(f.Values) > 0 {
				var vals []string
				for _, n := range f.Values {
					vals = append(vals, fmt.Sprint(n))
				}
				w.p("switch x {case %s:default:return nil,profileBadValue(v,5004,validation.ErrField)}", strings.Join(vals, ","))
			}
		}
		if repeated(f) {
			w.p("m.%s=append(m.%s,x)", f.Name, f.Name)
		} else if optional(f) {
			w.p("m.%s=avp.Some(x)", f.Name)
		} else {
			w.p("m.%s=x", f.Name)
		}
	}
	if w.answer {
		w.p("case avp.Key{Code:283},avp.Key{Code:293}:return nil,validation.At(v.Key,validation.ErrField)")
	}
	if w.ignoreVSAI {
		w.p("case avp.Key{Code:260}:if len(m.Extra)>=d.Limits().MaxRepeated{return nil,wire.ErrLimit};m.Extra=append(m.Extra,avp.Raw{Key:v.Key,Flags:v.Flags,Data:v.Data})")
	}
	w.p("default:if v.Flags&avp.MandatoryFlag!=0{return nil,profileBadValue(v,5001,validation.ErrUnsupported)};if len(m.Extra)>=d.Limits().MaxRepeated{return nil,wire.ErrLimit};m.Extra=append(m.Extra,avp.Raw{Key:v.Key,Flags:v.Flags,Data:v.Data})}}")
	for i, f := range fs {
		if f.Min > 0 {
			if f.Min == 1 && w.groups[f.Type].DiagnosticMinimum {
				plan, err := minimumPlan(f, w.groups)
				if err != nil {
					panic(err)
				}
				w.p("if seen[%d]==0{return nil,profileMissingGroup(%s,%d,%q)}", i, key(f), f.Flags, plan)
				continue
			}
			width := -1
			if f.Min == 1 {
				switch f.Type {
				case "u32", "i32", "time":
					width = 4
				case "u64", "i64":
					width = 8
				}
			}
			w.p("if seen[%d]<%d{return nil,profileMissing(%s,%d,%d)}", i, f.Min, key(f), f.Flags, width)
		}
	}
	w.p("if err:=validate%s(m,validation.Receive);err!=nil{return nil,err};return m,nil\n}", name)
}
func (w *writer) complete(c Command) {
	if !c.Request {
		return
	}
	w.p("func complete%s(b pdu.Body,env pdu.RequestEnvelope)(pdu.Body,error){m,ok:=b.(*%s);if !ok||m==nil{return nil,validation.ErrType};out:=clone%s(m)", c.Name, c.Name, c.Name)
	supported := map[string]bool{}
	for _, f := range c.Fields {
		supported[f.Envelope] = true
	}
	for _, name := range []string{"OriginHost", "OriginRealm", "DestinationHost", "DestinationRealm"} {
		if !supported[name] {
			w.p("if env.%s!=\"\"{return nil,validation.ErrEnvelope}", name)
		}
	}
	for _, f := range c.Fields {
		if f.Envelope == "" {
			continue
		}
		v := "out." + f.Name
		if optional(f) {
			v += ".Value"
		}
		w.p("if env.%s!=\"\" {if %s!=\"\"&&%s!=env.%s{return nil,validation.At(%s,validation.ErrEnvelope)};%s=avp.CloneString(env.%s)", f.Envelope, v, v, f.Envelope, key(f), v, f.Envelope)
		if optional(f) {
			w.p("out.%s.Present=true", f.Name)
		}
		w.p("}")
	}
	w.p("return out,nil\n}")
}

// SourceNames returns a stable diagnostic list without interpreting source documents.
func SourceNames(d Definition) []string {
	out := append([]string(nil), d.Sources...)
	sort.Strings(out)
	return out
}
