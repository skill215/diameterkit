// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Typed mixed tree bounds, legacy isolation, exact padding and fuzz.
// SPDX-AI-Date: 2026-11-17

package missingavp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type typedTreeProof struct {
	key   avp.Key
	flags byte
	plan  string
	child error
}

func (*typedTreeProof) Error() string { return "typed minimum tree" }
func (*typedTreeProof) Unwrap() error { return validation.ErrMissing }
func (p *typedTreeProof) DiameterMissingGroupV2() (avp.Key, byte, string, error) {
	return p.key, p.flags, p.plan, p.child
}
func typedRecord(k avp.Key, flags, width, children, kind byte) string {
	p := []byte(planRecord(k, flags, width, children))
	p[11] = kind
	return string(p)
}
func TestMinimumVariableTreeBounds(t *testing.T) {
	for kind := byte(1); kind <= 6; kind++ {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			width := byte(1)
			if kind == 3 || kind == 5 {
				width = 0
			}
			for _, negative := range []string{"exact", "bytes", "field", "frame", "count", "depth", "wrong-width", "unknown-kind", "ambiguous-zero", "typed-group", "leaf-children", "path", "flags", "duplicate", "incomplete", "trailing", "legacy", "mixed-version", "cycle", "bare"} {
				t.Run(negative, func(t *testing.T) {
					parent := avp.Key{VendorID: 193, Code: 9001}
					root := avp.Key{Code: 445}
					leaf := avp.Key{VendorID: 193, Code: 9002}
					fixed := avp.Key{Code: 450}
					last := avp.Key{Code: 9003}
					plan := planRecord(root, 64, 0, 3) + typedRecord(leaf, 128, width, 0, kind) + planRecord(fixed, 64, 4, 0) + typedRecord(last, 64, 1, 0, 6)
					p := &typedTreeProof{key: root, flags: 64, plan: plan}
					outer := &typedTreeProof{key: parent, flags: 192, child: p}
					field := &validation.FieldError{Path: []avp.Key{parent, root}, Cause: outer}
					var cause error = field
					size := 56 + ((int(width) + 3) &^ 3)
					budget := size
					l := wire.DefaultLimits()
					change := func(i int, v byte) { b := []byte(p.plan); b[i] = v; p.plan = string(b) }
					switch negative {
					case "bytes":
						budget--
					case "field":
						l.MaxFieldBytes = size - 1
					case "frame":
						l.MaxFrameBytes = size + 27
						l.MaxFieldBytes = l.MaxFrameBytes
					case "count":
						l.MaxAVPs = 5
						l.MaxRepeated = 5
					case "depth":
						l.MaxDepth = 2
					case "wrong-width":
						change(21, 1-width)
					case "unknown-kind":
						change(23, 255)
					case "ambiguous-zero":
						change(21, 0)
						change(23, 0)
					case "typed-group":
						change(11, 5)
					case "leaf-children":
						change(22, 1)
					case "path":
						field.Path[1].Code++
					case "flags":
						change(20, 64)
					case "duplicate":
						p.plan = planRecord(root, 64, 0, 2) + typedRecord(leaf, 128, width, 0, kind) + typedRecord(leaf, 128, width, 0, kind)
					case "incomplete":
						p.plan = p.plan[:len(p.plan)-1]
					case "trailing":
						p.plan += planRecord(fixed, 64, 4, 0)
					case "legacy":
						field.Cause = &treeProof{key: parent, flags: 192, child: &treeProof{key: root, flags: 64, plan: plan}}
					case "mixed-version":
						outer.child = &treeProof{key: root, flags: 64, plan: plan}
					case "cycle":
						outer.child = outer
					case "bare":
						cause = validation.At(parent, validation.At(root, validation.ErrMissing))
					}
					q, ok := Quote(cause, l, budget)
					if ok != (negative == "exact") || !ok && q != nil {
						t.Fatal("unsafe tree", negative, len(q))
					}
					if !ok {
						return
					}
					if len(q) != size {
						t.Fatal(len(q), size)
					}
					d := wire.NewDecoder(q, l)
					v, _, e := d.Next()
					if e != nil || v.Key != parent || v.Flags != 192 || d.Remaining() != 0 {
						t.Fatal(v, e)
					}
					d, _ = d.Child(v.Data)
					v, _, e = d.Next()
					if e != nil || v.Key != root || d.Remaining() != 0 {
						t.Fatal(v, e)
					}
					data := v.Data
					d, _ = d.Child(data)
					for i, k := range []avp.Key{leaf, fixed, last} {
						v, ok, e := d.Next()
						w := int(width)
						if i == 1 {
							w = 4
						}
						if i == 2 {
							w = 1
						}
						if e != nil || !ok || v.Key != k || !bytes.Equal(v.Data, make([]byte, w)) {
							t.Fatal(v, e)
						}
						head := minimumHead(k)
						if wire.Uint24(data[v.Offset+5:v.Offset+8]) != uint32(head+w) {
							t.Fatal("padded AVP length")
						}
						for _, b := range data[v.Offset+head+w : v.Offset+head+((w+3)&^3)] {
							if b != 0 {
								t.Fatal("padding")
							}
						}
					}
					if d.Remaining() != 0 {
						t.Fatal("extra sibling")
					}
				})
			}
		})
	}
}

type dualTreeProof struct {
	treeProof
	typed *typedTreeProof
}

func (p *dualTreeProof) DiameterMissingGroupV2() (avp.Key, byte, string, error) {
	return p.typed.DiameterMissingGroupV2()
}
func TestMinimumVariableNoVersionFallback(t *testing.T) {
	k := avp.Key{Code: 445}
	p := &dualTreeProof{treeProof: treeProof{key: k, flags: 64, plan: planRecord(k, 64, 0, 1) + planRecord(avp.Key{Code: 447}, 64, 8, 0)}, typed: &typedTreeProof{key: k, flags: 64, plan: "invalid"}}
	if _, ok := Quote(validation.At(k, p), wire.DefaultLimits(), 1024); ok {
		t.Fatal("invalid V2 fell back to V1")
	}
}
func FuzzMinimumVariableTree(f *testing.F) {
	k := avp.Key{Code: 445}
	for kind := byte(1); kind <= 6; kind++ {
		width := byte(1)
		if kind == 3 || kind == 5 {
			width = 0
		}
		f.Add([]byte(planRecord(k, 64, 0, 2)+typedRecord(avp.Key{VendorID: 193, Code: 9002}, 128, width, 0, kind)+planRecord(avp.Key{Code: 450}, 64, 4, 0)), uint16(128), uint8(4))
	}
	for _, group := range []uint32{437, 446} {
		plan := planRecord(avp.Key{Code: 456}, 64, 0, 1) + planRecord(avp.Key{Code: group}, 64, 0, 1) + planRecord(avp.Key{Code: 417}, 64, 8, 0)
		f.Add([]byte(plan), uint16(32), uint8(3))
		f.Add([]byte(plan), uint16(31), uint8(3))
	}
	f.Fuzz(func(t *testing.T, plan []byte, budget uint16, depth uint8) {
		if len(plan) > 1024 {
			return
		}
		l := wire.DefaultLimits()
		l.MaxAVPs = 64
		l.MaxRepeated = 16
		l.MaxDepth = int(depth % 65)
		root := k
		if len(plan) >= 8 && binary.BigEndian.Uint32(plan[4:8]) == 456 {
			root.Code = 456
		}
		q, ok := Quote(validation.At(root, &typedTreeProof{key: root, flags: 64, plan: string(plan)}), l, int(budget))
		if !ok {
			if q != nil {
				t.Fatal("partial")
			}
			return
		}
		if len(q) > int(budget) || len(q) > l.MaxFieldBytes {
			t.Fatal("unbounded")
		}
		index := 0
		var walk func([]byte)
		walk = func(data []byte) {
			d := wire.NewDecoder(data, l)
			for {
				v, found, e := d.Next()
				if e != nil {
					t.Fatal(e)
				}
				if !found {
					return
				}
				if index >= len(plan)/12 {
					t.Fatal("extra")
				}
				r := plan[index*12 : index*12+12]
				index++
				if r[10] > 0 {
					walk(v.Data)
				} else if len(v.Data) != int(r[9]) || !bytes.Equal(v.Data, make([]byte, len(v.Data))) {
					t.Fatal("payload")
				}
				head := minimumHead(v.Key)
				for _, b := range data[v.Offset+head+len(v.Data) : v.Offset+head+((len(v.Data)+3)&^3)] {
					if b != 0 {
						t.Fatal("padding")
					}
				}
			}
		}
		walk(q)
		if index != len(plan)/12 {
			t.Fatal("lost node")
		}
	})
}
