// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Minimum tree metadata, exact bytes, budgets and bounded fuzz.
// SPDX-AI-Date: 2026-11-17

package missingavp

import (
	"bytes"
	"encoding/binary"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

type treeProof struct {
	key   avp.Key
	flags byte
	plan  string
	child error
}

func (*treeProof) Error() string { return "minimum tree" }
func (*treeProof) Unwrap() error { return validation.ErrMissing }
func (p *treeProof) DiameterMissingGroupV1() (avp.Key, byte, string, error) {
	return p.key, p.flags, p.plan, p.child
}
func planRecord(key avp.Key, flags, width, children byte) string {
	var b [12]byte
	binary.BigEndian.PutUint32(b[:], key.VendorID)
	binary.BigEndian.PutUint32(b[4:], key.Code)
	b[8] = flags
	b[9] = width
	b[10] = children
	return string(b[:])
}
func TestMinimumTreeBounds(t *testing.T) {
	for _, kind := range []string{"exact", "bytes", "negative", "field", "frame", "count", "depth", "path", "bare", "flags", "vendor", "root", "leaf-root", "width", "reserved", "trailing", "incomplete", "duplicate", "empty-group", "too-many", "ancestor-plan", "cycle"} {
		t.Run(kind, func(t *testing.T) {
			root := avp.Key{Code: 445}
			parent := avp.Key{Code: 413}
			a := avp.Key{Code: 447}
			b := avp.Key{VendorID: 193, Code: 429}
			plan := planRecord(root, 64, 0, 2) + planRecord(a, 64, 8, 0) + planRecord(b, 192, 4, 0)
			leaf := &treeProof{key: root, flags: 64, plan: plan}
			outer := &treeProof{key: parent, flags: 64, child: leaf}
			field := &validation.FieldError{Path: []avp.Key{parent, root}, Cause: outer}
			var cause error = field
			l := wire.DefaultLimits()
			budget := 48
			change := func(pos int, v byte) { p := []byte(leaf.plan); p[pos] = v; leaf.plan = string(p) }
			switch kind {
			case "bytes":
				budget = 47
			case "negative":
				budget = -1
			case "field":
				l.MaxFieldBytes = 47
			case "frame":
				l.MaxFrameBytes = 75
				l.MaxFieldBytes = 75
			case "count":
				l.MaxAVPs = 4
				l.MaxRepeated = 4
			case "depth":
				l.MaxDepth = 2
			case "path":
				field.Path[1].Code++
			case "bare":
				cause = validation.At(root, validation.ErrMissing)
			case "flags":
				outer.flags = 65
			case "vendor":
				change(32, 64)
			case "root":
				change(7, 1)
			case "leaf-root":
				leaf.plan = planRecord(root, 64, 4, 0)
			case "width":
				change(21, 3)
			case "reserved":
				change(11, 1)
			case "trailing":
				leaf.plan += planRecord(a, 64, 4, 0)
			case "incomplete":
				leaf.plan = leaf.plan[:24]
			case "duplicate":
				leaf.plan = planRecord(root, 64, 0, 2) + planRecord(a, 64, 8, 0) + planRecord(a, 64, 4, 0)
			case "empty-group":
				leaf.plan = planRecord(root, 64, 0, 0)
			case "too-many":
				leaf.plan = string(bytes.Repeat([]byte(planRecord(a, 64, 4, 0)), 65))
			case "ancestor-plan":
				outer.plan = leaf.plan
			case "cycle":
				outer.child = outer
			}
			q, ok := Quote(cause, l, budget)
			if ok != (kind == "exact") || !ok && q != nil {
				t.Fatal("unsafe minimum", kind, len(q))
			}
			if !ok {
				return
			}
			if len(q) != 48 {
				t.Fatal(len(q))
			}
			d := wire.NewDecoder(q, l)
			v, _, _ := d.Next()
			if v.Key != parent || len(v.Data) != 40 {
				t.Fatal(v)
			}
			d, _ = d.Child(v.Data)
			v, _, _ = d.Next()
			if v.Key != root || len(v.Data) != 32 {
				t.Fatal(v)
			}
			d, _ = d.Child(v.Data)
			for i, k := range []avp.Key{a, b} {
				v, ok, e := d.Next()
				width := 8
				if i == 1 {
					width = 4
				}
				if e != nil || !ok || v.Key != k || !bytes.Equal(v.Data, make([]byte, width)) {
					t.Fatal(v, e)
				}
			}
			if d.Remaining() != 0 {
				t.Fatal("extra")
			}
		})
	}
}
func FuzzMinimumTree(f *testing.F) {
	k := avp.Key{Code: 445}
	f.Add([]byte(planRecord(k, 64, 0, 2)+planRecord(avp.Key{Code: 447}, 64, 8, 0)+planRecord(avp.Key{Code: 429}, 64, 4, 0)), uint16(64), uint8(4))
	f.Add([]byte{}, uint16(0), uint8(0))
	f.Fuzz(func(t *testing.T, plan []byte, budget uint16, depth uint8) {
		if len(plan) > 1024 {
			return
		}
		p := &treeProof{key: k, flags: 64, plan: string(plan)}
		e := validation.At(k, p)
		l := wire.DefaultLimits()
		l.MaxAVPs = 64
		l.MaxRepeated = 16
		l.MaxDepth = int(depth % 65)
		q, ok := Quote(e, l, int(budget))
		if !ok {
			if q != nil {
				t.Fatal("partial quote")
			}
			return
		}
		if len(q) > int(budget) || len(q) > l.MaxFieldBytes {
			t.Fatal("unbounded")
		}
		// Walk using the same declared tree only to validate emitted framing; every
		// payload is either children or a zero-filled fixed-width scalar.
		index := 0
		var walk func([]byte)
		walk = func(data []byte) {
			d := wire.NewDecoder(data, l)
			for {
				v, found, err := d.Next()
				if err != nil {
					t.Fatal(err)
				}
				if !found {
					return
				}
				if index >= len(plan)/12 {
					t.Fatal("extra node")
				}
				r := plan[index*12 : index*12+12]
				index++
				if r[9] == 0 {
					walk(v.Data)
				} else if len(v.Data) != int(r[9]) || !bytes.Equal(v.Data, make([]byte, len(v.Data))) {
					t.Fatal("nonzero scalar")
				}
			}
		}
		walk(q)
		if index != len(plan)/12 {
			t.Fatal("lost node")
		}
	})
}
