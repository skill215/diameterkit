// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Originator-SCCP-Address exclusion on combined ECUR requests.
// SPDX-AI-Date: 2026-11-24

package excessavp

import (
	"bytes"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type ecurOriginatorProof struct{}

func (ecurOriginatorProof) Error() string                     { return "excluded address" }
func (ecurOriginatorProof) Unwrap() error                     { return validation.ErrUnsupported }
func (ecurOriginatorProof) DiameterExcludedECUROriginatorV1() {}
func ecurOriginatorCause() *validation.FieldError {
	return &validation.FieldError{Path: []avp.Key{{VendorID: 10415, Code: 873}, {VendorID: 10415, Code: 2000}, {VendorID: 10415, Code: 2008}}, Cause: ecurOriginatorProof{}}
}
func ecurOriginatorArea(kind uint32, address []byte, orig, dest int32) []byte {
	number := uint32(0)
	if kind == 3 {
		number = 1
	}
	top := bytes.Join([][]byte{actionWord(416, kind), actionWord(415, number), actionWord(455, 1), unitPart(458, nil)}, nil)
	var sms []byte
	for _, code := range []uint32{2001, 2002, 2007, 2008, 2009, 2010, 2011, 2013, 2015, 2016, 2017, 2018} {
		var data []byte
		switch code {
		case 2001, 2007, 2011, 2016:
			data = make([]byte, 4)
		case 2002:
			data = interfaceLeaf(dest)
		case 2009:
			data = interfaceLeaf(orig)
		case 2008:
			data = address
		}
		sms = append(sms, interfacePart(code, data)...)
	}
	if kind == 3 {
		sms = append(sms, interfacePart(2014, nil)...)
		sms = append(sms, interfacePart(2012, make([]byte, 4))...)
	}
	return append(top, interfacePart(873, interfacePart(2000, sms))...)
}
func TestECUROriginatorConsumer(t *testing.T) {
	data := []byte{255, 2, 3}
	raw := ecurOriginatorArea(1, data, 3, 2)
	for _, name := range []string{"valid", "term", "wrong-kind", "wrong-sequence", "action", "origin", "dest", "both", "nul", "path", "bare", "trailing", "duplicate-selector", "count", "depth", "field", "frame", "short"} {
		t.Run(name, func(t *testing.T) {
			src := bytes.Clone(raw)
			l := wire.DefaultLimits()
			budget := 40
			var cause error = ecurOriginatorCause()
			switch name {
			case "term":
				src = ecurOriginatorArea(3, data, 3, 2)
			case "wrong-kind":
				src[11] = 4
			case "wrong-sequence":
				src[23] = 1
			case "action":
				src = append(src, actionWord(436, 0)...)
			case "origin":
				src = ecurOriginatorArea(1, data, 0, 2)
			case "dest":
				src = ecurOriginatorArea(1, data, 3, 0)
			case "both":
				src = ecurOriginatorArea(1, data, 2, 3)
			case "nul":
				src = ecurOriginatorArea(1, []byte{0}, 3, 2)
			case "path":
				f := ecurOriginatorCause()
				f.Path[1].Code = 877
				cause = f
			case "bare":
				cause = validation.ErrUnsupported
			case "trailing":
				src = append(src, 0)
			case "duplicate-selector":
				src = append(src, actionWord(415, 0)...)
			case "count":
				l.MaxAVPs = 19
				l.MaxRepeated = 19
			case "depth":
				l.MaxDepth = 3
			case "field":
				l.MaxFieldBytes = 39
			case "frame":
				l.MaxFrameBytes = len(src) + 19
				l.MaxFieldBytes = l.MaxFrameBytes
			case "short":
				budget--
			}
			q, c, ok := QuoteInvalid(src, cause, l, budget)
			if ok != (name == "valid" || name == "term") {
				t.Fatal(c, ok)
			}
			if ok {
				if c != 5008 || !bytes.Equal(q, interfacePart(873, interfacePart(2000, interfacePart(2008, data)))) {
					t.Fatal(c, q)
				}
			} else if c != 0 || q != nil {
				t.Fatal(c, q)
			}
		})
	}
}

func fuzzECUROriginator(t *testing.T, raw []byte, budget uint16) {
	l := wire.DefaultLimits()
	l.MaxFrameBytes = 4096
	l.MaxFieldBytes = 2048
	l.MaxAVPs = 64
	l.MaxRepeated = 64
	q, c, ok := QuoteInvalid(raw, ecurOriginatorCause(), l, int(budget))
	if !ok {
		if q != nil || c != 0 {
			t.Fatal("partial quote")
		}
		return
	}
	if c != 5008 || len(q) < 36 || len(q) > int(budget) {
		t.Fatal(c, len(q))
	}
	// Resolve the quoted path independently against the original wire area.
	area := raw
	for _, code := range []uint32{873, 2000, 2008} {
		d := wire.NewDecoder(area, l)
		var selected wire.View
		for {
			v, more, err := d.Next()
			if err != nil {
				t.Fatal(err)
			}
			if !more {
				break
			}
			if v.Key == (avp.Key{VendorID: 10415, Code: code}) {
				if selected.Key.Code != 0 || v.Flags != 192 {
					t.Fatal("ambiguous path")
				}
				selected = v
			}
		}
		if selected.Key.Code != code {
			t.Fatal("missing field", code)
		}
		if code == 2008 {
			end := selected.Offset + ((12 + len(selected.Data) + 3) &^ 3)
			if bytes.IndexByte(selected.Data, 0) >= 0 || !bytes.Equal(q, interfacePart(873, interfacePart(2000, area[selected.Offset:end]))) {
				t.Fatal("original field")
			}
		}
		area = selected.Data
	}
}
