// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package pdu defines transport-independent Diameter message envelopes.
package pdu

const HeaderSize = 20
const (
	RequestFlag       byte = 0x80
	ProxiableFlag     byte = 0x40
	ErrorFlag         byte = 0x20
	RetransmittedFlag byte = 0x10
)

type CommandKey struct {
	ApplicationID, Code uint32
	Request             bool
}

// Body is implementable by external generated packages. A key alone is not type authorization.
type Body interface{ CommandKey() CommandKey }
type Header struct {
	Flags                                          byte
	CommandCode, ApplicationID, HopByHop, EndToEnd uint32
}

func (h Header) Key() CommandKey {
	return CommandKey{h.ApplicationID, h.CommandCode, h.Flags&RequestFlag != 0}
}

type Message struct {
	Header Header
	Body   Body
}
type RequestEnvelope struct {
	OriginHost, OriginRealm, DestinationHost, DestinationRealm string
	HopByHop, EndToEnd                                         uint32
}
