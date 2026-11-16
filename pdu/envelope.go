// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package pdu

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
)

// Origin is a protocol identity, not a Peer or proof of authorization.
type Origin struct{ Host, Realm string }

// AnswerSources is an explicit, finite offline/source-policy value. The runtime
// obtains it from a frozen binding, never from a received AVP or adjacent DRA.
type AnswerSources struct{ Terminals, ErrorAgents []Origin }

type ProxyInfo struct {
	ProxyHost  string
	ProxyState []byte
	Extra      []avp.Raw
}

// FailedAVP is quoted diagnostic data, not an application AVP tree. Incomplete
// is an observation made during bounded parsing, never an authorization fact.
type FailedAVP struct {
	Data       []byte
	Incomplete bool
}

type ExperimentalResult struct {
	VendorID, ExperimentalResultCode uint32
	Extra                            []avp.Raw
}

type VendorSpecificApplicationID struct {
	VendorID                             uint32
	AuthApplicationID, AcctApplicationID avp.Optional[uint32]
	Extra                                []avp.Raw
}

// Envelope is a synchronous borrowed view returned by compiled operations.
// Retaining it requires an owned message/snapshot; extracting it does not validate it.
type Envelope struct {
	DestinationHost, DestinationRealm avp.Optional[string]
	SessionID                         avp.Optional[string]
	Origin                            Origin
	ProxyInfo                         []ProxyInfo
	AuthSessionState                  avp.Optional[uint32]
}

type AnswerEnvelope struct {
	Command   CommandKey
	SessionID avp.Optional[string]
	Origin    Origin
	ProxyInfo []ProxyInfo
}

// ErrorAnswer uses RFC 6733's E grammar, independently of a normal answer body.
// Command is filled by CompleteAnswer, or must be supplied for direct encoding.
type ErrorAnswer struct {
	Command                     CommandKey
	SessionID                   avp.Optional[string]
	OriginHost, OriginRealm     string
	ResultCode                  uint32
	OriginStateID               avp.Optional[uint32]
	ErrorMessage                avp.Optional[string]
	ErrorReportingHost          avp.Optional[string]
	FailedAVP                   avp.Optional[FailedAVP]
	ExperimentalResult          avp.Optional[ExperimentalResult]
	ProxyInfo                   []ProxyInfo
	AuthSessionState            avp.Optional[uint32]
	Extra                       []avp.Raw
	VendorSpecificApplicationID avp.Optional[VendorSpecificApplicationID]
	RouteRecord                 []string
}

func (m *ErrorAnswer) CommandKey() CommandKey {
	if m == nil {
		return CommandKey{}
	}
	return m.Command
}

// MatchesAnswerHeader only identifies a candidate on an already selected local
// connection/direction. P/T/E semantics, body and source checks follow matching.
// It never looks up a connection, removes pending work or grants completion rights.
func MatchesAnswerHeader(request, answer Header) bool {
	return request.Flags&RequestFlag != 0 && answer.Flags&RequestFlag == 0 &&
		request.CommandCode == answer.CommandCode && request.ApplicationID == answer.ApplicationID &&
		request.HopByHop == answer.HopByHop && request.EndToEnd == answer.EndToEnd
}

func EqualProxies(a, b []ProxyInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ProxyHost != b[i].ProxyHost || !bytes.Equal(a[i].ProxyState, b[i].ProxyState) || len(a[i].Extra) != len(b[i].Extra) {
			return false
		}
		for j, v := range a[i].Extra {
			other := b[i].Extra[j]
			if v.Key != other.Key || v.Flags&^byte(0x1f) != other.Flags&^byte(0x1f) || !bytes.Equal(v.Data, other.Data) {
				return false
			}
		}
	}
	return true
}
