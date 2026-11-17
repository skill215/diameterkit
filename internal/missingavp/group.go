// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded minimum Grouped diagnostic plans and ancestor construction.
// SPDX-AI-Date: 2026-11-17

package missingavp

import (
	"encoding/binary"
	"errors"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

type groupEvidence interface {
	DiameterMissingGroupV1() (avp.Key, byte, string, error)
}
type minimumNode struct {
	key                avp.Key
	flags              byte
	head, size, parent int
}

func quoteGroup(cause error, proof groupEvidence, limits wire.Limits, budget int) ([]byte, bool) {
	var field *validation.FieldError
	if limits.Validate() != nil || budget < 0 || !errors.Is(cause, validation.ErrMissing) || !errors.As(cause, &field) {
		return nil, false
	}
	var parents [64]minimumNode
	depth, headers := 0, 0
	for {
		key, flags, plan, child := proof.DiameterMissingGroupV1()
		if depth >= len(parents) || depth >= len(field.Path) || depth >= limits.MaxDepth || key != field.Path[depth] || !minimumKey(key, flags) {
			return nil, false
		}
		h := minimumHead(key)
		if child != nil {
			if plan != "" {
				return nil, false
			}
			var ok bool
			proof, ok = child.(groupEvidence)
			if !ok {
				return nil, false
			}
			parents[depth] = minimumNode{key: key, flags: flags, head: h}
			depth++
			headers += h
			continue
		}
		if depth+1 != len(field.Path) || len(plan) == 0 || len(plan)%12 != 0 || len(plan) > 64*12 || depth+len(plan)/12+1 > limits.MaxAVPs {
			return nil, false
		}
		var nodes [64]minimumNode
		count := 0
		var parse func(int, int) (int, bool)
		parse = func(parent, level int) (int, bool) {
			if count >= len(plan)/12 || level+depth > limits.MaxDepth {
				return 0, false
			}
			i := count
			r := plan[i*12 : i*12+12]
			count++
			k := avp.Key{VendorID: binary.BigEndian.Uint32([]byte(r[:4])), Code: binary.BigEndian.Uint32([]byte(r[4:8]))}
			flags, width, children := r[8], int(r[9]), int(r[10])
			if !minimumKey(k, flags) || r[11] != 0 || width != 0 && width != 4 && width != 8 || width == 0 && children == 0 || width != 0 && children != 0 {
				return 0, false
			}
			for j := 0; j < i; j++ {
				if nodes[j].parent == parent && nodes[j].key == k {
					return 0, false
				}
			}
			head := minimumHead(k)
			nodes[i] = minimumNode{key: k, flags: flags, head: head, parent: parent}
			size := head + width
			for c := 0; c < children; c++ {
				n, ok := parse(i, level+1)
				if !ok {
					return 0, false
				}
				size += n
			}
			nodes[i].size = size
			return size, true
		}
		size, ok := parse(-1, 1)
		if !ok || count != len(plan)/12 || nodes[0].key != key || nodes[0].flags != flags || plan[9] != 0 {
			return nil, false
		}
		size += headers
		if size > budget || size > limits.MaxFieldBytes || size > limits.MaxFrameBytes-28 {
			return nil, false
		}
		// Metadata and complete output budgets have been checked before allocation.
		out := make([]byte, size)
		pos := 0
		for i := 0; i < depth; i++ {
			p := parents[i]
			minimumHeader(out[pos:], p.key, p.flags, size-pos)
			pos += p.head
		}
		for i := 0; i < count; i++ {
			n := nodes[i]
			minimumHeader(out[pos:], n.key, n.flags, n.size)
			pos += n.head + int(plan[i*12+9])
		}
		return out, true
	}
}
func minimumKey(k avp.Key, flags byte) bool {
	return k.Code != 0 && flags&0x1f == 0 && (k.VendorID != 0) == (flags&avp.VendorFlag != 0)
}
func minimumHead(k avp.Key) int {
	if k.VendorID != 0 {
		return 12
	}
	return 8
}
func minimumHeader(dst []byte, key avp.Key, flags byte, size int) {
	binary.BigEndian.PutUint32(dst, key.Code)
	dst[4] = flags
	wire.Put24(dst[5:8], uint32(size))
	if key.VendorID != 0 {
		binary.BigEndian.PutUint32(dst[8:], key.VendorID)
	}
}
