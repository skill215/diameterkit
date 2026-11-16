// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package sctp

import (
	"encoding/binary"

	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
	upstream "github.com/skill215/sctp"
)

// ordered is a stream0 barrier, not a request to enable a mutable socket mode.
// false allows the configured post-handshake policy; control always stays on0.
func sendInfo(c Config, outbound uint16, b []byte, ordered bool) (upstream.SndInfo, error) {
	if outbound == 0 {
		return upstream.SndInfo{}, ErrConfig
	}
	i := upstream.SndInfo{PPID: 46}
	if c.PPIDPolicy == PPID0Compatibility {
		i.PPID = 0
	}
	code := binary.BigEndian.Uint32(b[4:8]) & 0xffffff
	control := binary.BigEndian.Uint32(b[8:12]) == 0 && (code == 257 || code == 280 || code == 282)
	if ordered || control {
		return i, nil
	}
	if c.StreamMode == UnorderedStream0 {
		i.Flags = upstream.SCTP_UNORDERED
		return i, nil
	}
	key, e := streamHash(b)
	if e != nil {
		return upstream.SndInfo{}, e
	}
	if outbound > 1 {
		i.SID = uint16(1 + key%uint64(outbound-1))
	}
	return i, nil
}

// Fixed FNV-1a64, with domain1 for exact top-level Session-Id bytes, domain2
// for AppID/CommandCode/E2E when no Session exists. HBH/R/T are excluded so
// answers and retransmissions keep their key. This is not an identity proof.
// Reuse wire's bounded top-level AVP walker; no grouped/application decoding,
// retained frame, per-session map, or copied payload is involved.
func streamHash(b []byte) (uint64, error) {
	l := wire.Limits{MaxFrameBytes: wire.MaxWireSize, MaxFieldBytes: wire.MaxWireSize, MaxAVPs: wire.MaxWireSize / 8, MaxRepeated: 1, MaxDepth: 1}
	_, n, e := wire.Header(b, l)
	if e != nil || n != len(b) {
		return 0, transport.ErrFrameState
	}
	d := wire.NewDecoder(b[20:], l)
	var session []byte
	for {
		v, ok, e := d.Next()
		if e != nil {
			return 0, transport.ErrFrameState
		}
		if !ok {
			break
		}
		if v.Key.VendorID == 0 && v.Key.Code == 263 {
			if session != nil || len(v.Data) == 0 {
				return 0, transport.ErrFrameState
			}
			session = v.Data
		}
	}
	h := uint64(14695981039346656037)
	add := func(b []byte) {
		for _, v := range b {
			h = (h ^ uint64(v)) * 1099511628211
		}
	}
	if session != nil {
		add([]byte{1})
		add(session)
	} else {
		add([]byte{2})
		add(b[8:12])
		add(b[5:8])
		add(b[16:20])
	}
	return h, nil
}
