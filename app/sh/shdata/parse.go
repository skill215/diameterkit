// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package shdata

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/skill215/diameterkit/avp"
)

const xmlNS = "http://www.w3.org/XML/1998/namespace"
const xmlnsNS = "http://www.w3.org/2000/xmlns/"
const xsiNS = "http://www.w3.org/2001/XMLSchema-instance"

var declaration = regexp.MustCompile(`^version[ \t\r\n]*=[ \t\r\n]*(?:"1\.0"|'1\.0')(?:[ \t\r\n]+encoding[ \t\r\n]*=[ \t\r\n]*(?:"(?i:UTF-8)"|'(?i:UTF-8)'))?(?:[ \t\r\n]+standalone[ \t\r\n]*=[ \t\r\n]*(?:"(?:yes|no)"|'(?:yes|no)'))?[ \t\r\n]*$`)

type binding struct{ prefix, uri string }
type frame struct {
	name     xml.Name // QName from RawToken, checked on closing tag.
	path     string
	shape    shape
	last     int
	number   bool
	bindings int
	fragment int // index of outer opaque fragment, or -1
	text     strings.Builder
}
type parser struct {
	doc    Document
	limits Limits
	stack  []frame
	ns     []binding
	offset int64
}

func (p *parser) fail(kind error, path string) error { return &Error{kind, path, p.offset} }
func whitespace(s string) bool                       { return strings.Trim(s, " \t\r\n") == "" }
func validCharacters(s []byte) bool {
	if !utf8.Valid(s) {
		return false
	}
	for _, r := range string(s) {
		if !(r == 9 || r == 10 || r == 13 || r >= 0x20 && r <= 0xd7ff || r >= 0xe000 && r <= 0xfffd || r >= 0x10000 && r <= 0x10ffff) {
			return false
		}
	}
	return true
}
func qname(n xml.Name) bool {
	return ncname(n.Local) && (n.Space == "" || ncname(n.Space))
}

// RawToken validates XML Name, but QName components must each be NCNames.
// In particular, a:1 is an XML Name yet not a namespace-well-formed QName.
func ncname(s string) bool {
	if s == "" || strings.Contains(s, ":") {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' ||
		r >= 0xc0 && r <= 0xd6 || r >= 0xd8 && r <= 0xf6 || r >= 0xf8 && r <= 0x2ff ||
		r >= 0x370 && r <= 0x37d || r >= 0x37f && r <= 0x1fff || r >= 0x200c && r <= 0x200d ||
		r >= 0x2070 && r <= 0x218f || r >= 0x2c00 && r <= 0x2fef || r >= 0x3001 && r <= 0xd7ff ||
		r >= 0xf900 && r <= 0xfdcf || r >= 0xfdf0 && r <= 0xfffd || r >= 0x10000 && r <= 0xeffff
}
func (p *parser) resolve(n xml.Name, attr bool) (xml.Name, bool) {
	if !qname(n) || n.Space == "xmlns" {
		return xml.Name{}, false
	}
	if n.Space == "xml" {
		return xml.Name{Space: xmlNS, Local: n.Local}, true
	}
	if attr && n.Space == "" {
		return n, true
	}
	for i := len(p.ns) - 1; i >= 0; i-- {
		if p.ns[i].prefix == n.Space {
			return xml.Name{Space: p.ns[i].uri, Local: n.Local}, true
		}
	}
	return n, n.Space == ""
}
func isNS(a xml.Attr) bool {
	return a.Name.Space == "xmlns" || a.Name.Space == "" && a.Name.Local == "xmlns"
}
func (p *parser) attributes(t xml.StartElement, path string) (xml.Name, error) {
	// Enforce namespace well-formedness explicitly: encoding/xml accepts unbound
	// prefixes, duplicate attributes and several forbidden namespace bindings.
	seen := make(map[xml.Name]bool, len(t.Attr))
	for _, a := range t.Attr {
		if !qname(a.Name) || seen[a.Name] {
			return xml.Name{}, p.fail(ErrXML, path)
		}
		seen[a.Name] = true
		if !isNS(a) {
			continue
		}
		prefix := a.Name.Local
		if a.Name.Space == "" {
			prefix = ""
		}
		if prefix == "xmlns" || a.Value == xmlnsNS || prefix == "xml" && a.Value != xmlNS || prefix != "xml" && a.Value == xmlNS || prefix != "" && a.Value == "" {
			return xml.Name{}, p.fail(ErrXML, path)
		}
		p.ns = append(p.ns, binding{prefix, a.Value})
	}
	name, ok := p.resolve(t.Name, false)
	if !ok {
		return xml.Name{}, p.fail(ErrXML, path)
	}
	clear(seen)
	for _, a := range t.Attr {
		if isNS(a) {
			continue
		}
		n, ok := p.resolve(a.Name, true)
		if !ok || seen[n] {
			return xml.Name{}, p.fail(ErrXML, path)
		}
		seen[n] = true
	}
	return name, nil
}

// Parse copies raw bytes on success. No partial Document is returned on errors.
// Known selected leaves are checked at their exact namespace and parent path.
// Other recognized Sh fields and allowed external wildcard subtrees are retained
// as bounded raw ranges and reported as uninterpreted; no whole-XSD claim follows.
func Parse(raw []byte, limits Limits) (*Document, error) {
	if !limits.valid() {
		return nil, ErrArgument
	}
	if len(raw) > limits.MaxBytes {
		return nil, &Error{ErrLimit, "/", 0}
	}
	if !validCharacters(raw) {
		return nil, &Error{ErrXML, "/", 0}
	}
	input := raw
	base := 0
	if bytes.HasPrefix(input, []byte{0xef, 0xbb, 0xbf}) {
		input = input[3:]
		base = 3
	}
	p := parser{limits: limits, stack: make([]frame, 0, limits.MaxDepth)}
	p.doc.usage.Bytes = len(raw)
	d := xml.NewDecoder(bytes.NewReader(input))
	d.CharsetReader = func(string, io.Reader) (io.Reader, error) { return nil, ErrUnsupported }
	seenRoot := false
	for {
		start := int(d.InputOffset()) + base
		token, err := d.RawToken()
		p.offset = d.InputOffset() + int64(base)
		if err == io.EOF {
			break
		}
		path := "/"
		if len(p.stack) > 0 {
			path = p.stack[len(p.stack)-1].path
		}
		if err != nil {
			kind := ErrXML
			if errors.Is(err, ErrUnsupported) {
				kind = ErrUnsupported
			}
			return nil, p.fail(kind, path)
		}
		outerOpaque := -1
		if len(p.stack) > 0 {
			outerOpaque = p.stack[len(p.stack)-1].fragment
		}
		switch t := token.(type) {
		case xml.StartElement:
			p.doc.usage.Elements++
			p.doc.usage.Attributes += len(t.Attr)
			if len(p.stack) >= limits.MaxDepth || p.doc.usage.Elements > limits.MaxElements || p.doc.usage.Attributes > limits.MaxAttributes {
				return nil, p.fail(ErrLimit, path)
			}
			nsCount := len(p.ns)
			// Resolve attributes first; applicability is checked after selecting shape.
			name, e := p.attributes(t, path)
			if e != nil {
				return nil, e
			}
			kind := rootShape
			nextPath := "/Sh-Data"
			if len(p.stack) == 0 {
				if seenRoot {
					return nil, p.fail(ErrXML, path)
				}
				if name.Space != "" || name.Local != "Sh-Data" {
					return nil, p.fail(ErrStructure, path)
				}
				seenRoot = true
			} else {
				parent := &p.stack[len(p.stack)-1]
				nextPath = parent.path + "/" + name.Local
				if outerOpaque >= 0 {
					kind = opaque
					nextPath = parent.path
				} else {
					choices, wild := children(parent.shape)
					index := -1
					if name.Space != "" && wild {
						index = len(choices)
						kind = opaque
					} else if name.Space == "" {
						for i, c := range choices {
							if c.name == name.Local {
								index = i
								kind = c.shape
								break
							}
						}
					}
					if index < 0 || index < parent.last || index == parent.last && index < len(choices) && !choices[index].repeat {
						return nil, p.fail(ErrStructure, nextPath)
					}
					parent.last = index
					if parent.shape == sms && kind == numberLeaf {
						parent.number = true
					}
				}
			}
			if kind != opaque {
				for _, a := range t.Attr {
					if isNS(a) {
						continue
					}
					n, _ := p.resolve(a.Name, true)
					if !(n.Space == xsiNS && (n.Local == "schemaLocation" || n.Local == "noNamespaceSchemaLocation")) {
						return nil, p.fail(ErrUnsupported, nextPath)
					}
				}
			}
			if kind == opaque && outerOpaque < 0 {
				outerOpaque = len(p.doc.fragments)
				p.doc.fragments = append(p.doc.fragments, Fragment{Path: nextPath, Namespace: name.Space, Local: name.Local, Start: start})
				p.doc.issues = append(p.doc.issues, Issue{nextPath, Uninterpreted})
			}
			p.stack = append(p.stack, frame{name: t.Name, path: nextPath, shape: kind, last: -1, bindings: nsCount, fragment: outerOpaque})
			if len(p.stack) > p.doc.usage.MaxDepth {
				p.doc.usage.MaxDepth = len(p.stack)
			}
			if kind == sms {
				p.doc.data.SMS.Present = true
			}
			if kind == reach {
				p.doc.data.Reachability.Present = true
			}
		case xml.EndElement:
			if len(p.stack) == 0 || p.stack[len(p.stack)-1].name != t.Name {
				return nil, p.fail(ErrXML, path)
			}
			f := &p.stack[len(p.stack)-1]
			if f.shape == sms && !f.number {
				return nil, p.fail(ErrStructure, path+"/IP-SM-GW-Number")
			}
			if f.shape >= numberLeaf {
				if e := p.leaf(f); e != nil {
					return nil, e
				}
			}
			if outerOpaque >= 0 && (len(p.stack) == 1 || p.stack[len(p.stack)-2].fragment != outerOpaque) {
				p.doc.fragments[outerOpaque].End = int(p.offset)
			}
			p.ns = p.ns[:f.bindings]
			p.stack = p.stack[:len(p.stack)-1]
		case xml.CharData:
			p.doc.usage.TextBytes += len(t)
			if p.doc.usage.TextBytes > limits.MaxTextBytes {
				return nil, p.fail(ErrLimit, path)
			}
			if len(p.stack) == 0 {
				if !whitespace(string(raw[start:int(p.offset)])) {
					return nil, p.fail(ErrXML, path)
				}
			} else if outerOpaque < 0 {
				f := &p.stack[len(p.stack)-1]
				if f.shape >= numberLeaf {
					f.text.Write(t)
				} else if !whitespace(string(t)) {
					return nil, p.fail(ErrStructure, path)
				}
			}
		case xml.Directive:
			return nil, p.fail(ErrUnsupported, path) // DTD and entities never run.
		case xml.ProcInst:
			if strings.EqualFold(t.Target, "xml") {
				if t.Target != "xml" || start != base || !declaration.Match(t.Inst) {
					return nil, p.fail(ErrXML, path)
				}
			}
		case xml.Comment:
		default:
			return nil, p.fail(ErrXML, path)
		}
		if outerOpaque >= 0 {
			p.doc.usage.OpaqueBytes += int(p.offset) - start
			if p.doc.usage.OpaqueBytes > limits.MaxOpaqueBytes {
				return nil, p.fail(ErrLimit, path)
			}
		}
	}
	if !seenRoot || len(p.stack) != 0 {
		return nil, p.fail(ErrXML, "/")
	}
	p.doc.raw = bytes.Clone(raw)
	return &p.doc, nil
}

func (p *parser) leaf(f *frame) error {
	s := f.text.String()
	switch f.shape {
	case scscfLeaf:
		p.doc.data.SCSCFName = avp.Some(s)
		if s != "" && !scscfName(s) {
			p.doc.issues = append(p.doc.issues, Issue{f.path, URIRepresentation})
		}
	case numberLeaf, scLeaf:
		if f.shape == scLeaf && utf8.RuneCountInString(s) > 20 {
			return p.fail(ErrStructure, f.path)
		}
		if s != "" && !number(s) {
			p.doc.issues = append(p.doc.issues, Issue{f.path, NumberRepresentation})
		}
		if f.shape == numberLeaf {
			p.doc.data.SMS.Value.Number = s
		} else {
			p.doc.data.SMS.Value.SCAddress = avp.Some(s)
		}
	default:
		// xs:unsignedByte has whitespace collapse and permits sign/leading zeros.
		// Do not trim xs:string leaves or change the raw payload used for duplicates.
		value := strings.Trim(s, " \t\r\n")
		// XML Schema integer lexical space permits one sign, including -0.
		negative := strings.HasPrefix(value, "-")
		if negative || strings.HasPrefix(value, "+") {
			value = value[1:]
		}
		if len(value) == 0 {
			return p.fail(ErrStructure, f.path)
		}
		maxValue := 0
		if f.shape == sbiLeaf {
			maxValue = 1
		}
		if f.shape == imsStateLeaf {
			maxValue = 3
		}
		n := 0
		for i := range value {
			if value[i] < '0' || value[i] > '9' {
				return p.fail(ErrStructure, f.path)
			}
			n = n*10 + int(value[i]-'0')
			if n > maxValue {
				return p.fail(ErrStructure, f.path)
			}
		}
		if negative && n != 0 {
			return p.fail(ErrStructure, f.path)
		}
		v := avp.Some(uint8(n))
		r := &p.doc.data.Reachability.Value
		switch f.shape {
		case imsStateLeaf:
			p.doc.data.IMSUserState = v
		case sbiLeaf:
			p.doc.data.SMS.Value.SBI = v
		case mmeLeaf:
			r.MME = v
		case sgsnLeaf:
			r.SGSN = v
		case amfLeaf:
			r.AMF3GPP = v
		case nonAMFLeaf:
			r.AMFNon3GPP = v
		}
	}
	return nil
}
