package diameter

import (
	"errors"
	"time"

	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/wire"
)

var (
	ErrConfig       = errors.New("invalid endpoint configuration")
	ErrClosed       = errors.New("endpoint or completion port closed")
	ErrNotReady     = errors.New("peer is not ready")
	ErrBackpressure = errors.New("endpoint resource limit exceeded")
	ErrIdentifiers  = errors.New("identifier space exhausted")
	ErrProtocol     = errors.New("peer protocol violation")
	ErrUnsupported  = errors.New("runtime capability not supported")
	ErrPeerClosed   = errors.New("peer closed")
	ErrRemoteDPR    = errors.New("remote disconnect request")
	ErrShutdown     = errors.New("local shutdown")
	ErrElectionLost = errors.New("connection lost peer election")
	ErrWatchdog     = errors.New("watchdog suspect")
)

type PeerID string
type ClassID string
type ClassConfig struct {
	ID    ClassID
	Bytes uint64
	Slots int
}
type ResourcePlan struct {
	RoRecoverySlots                            int
	RoRecoveryClasses                          []IdentifierClass
	IdentifierGuardSlots                       int
	IdentifierClasses                          []IdentifierClass
	ListenerBytesPerPeer                       uint64 // Zero disables Serve; otherwise reserves one listener per configured Peer.
	MaxDiscardBytes                            int
	NodeBytes, TransportBytes, IdentifierBytes uint64
	Classes                                    []ClassConfig
	DuplicateClasses                           []ClassConfig // Independent volatile result/cache partition.
	InboundClasses                             []ClassConfig // Independent hard partition; never borrowed by outbound.
	// Ports and connections are bounded at construction, including their rings.
	MaxPorts, MaxPortOutstanding           int
	AdapterBytesPerConnection              uint64
	RejectSlots                            int // Global emergency slots, already charged in TransportBytes.
	EmergencyRequestMax, EmergencyReplyMax int
}
type Timeouts struct{ Handshake, FrameRead, FrameWrite, Request, Watchdog, ShutdownTimeout, CloseReserve, DPAWait, ClosingGrace time.Duration }

// CommandPolicy is an exact, frozen outbound binding with explicit terminal/error
// sources. Optional PeerConfig.RequestTargets supplies reviewed per-request targets;
// the default remains this fixed destination. SCAP1 requires a fixed terminal.
type CommandPolicy struct {
	ApplicationBinding                ApplicationBinding // Zero uses the header application; SCAP1 must opt in.
	Command                           pdu.CommandKey
	Class                             ClassID
	DestinationHost, DestinationRealm string
	Sources                           pdu.AnswerSources
	MaxRequestBytes, MaxAnswerBytes   int
	Timeout                           time.Duration
}
type PeerConfig struct {
	// SCAP1Advertisement changes only this peer's CER/CEA wire compatibility.
	// It requires an explicit SCAP1Accounting command and strict local capability.
	SCAP1Advertisement SCAP1Advertisement
	// RequestTargets is an optional frozen allowlist, at most 64 entries per Peer.
	RequestTargets  []RequestTarget
	Maintain        MaintainPolicy // Zero disables Maintain; manual Dial/Connect remain available.
	ID              PeerID
	Origin          pdu.Origin
	MaxPending      int
	MaxBytes        uint64
	Commands        []CommandPolicy
	MaxInbound      int
	MaxInboundBytes uint64
	InboundCommands []InboundPolicy
	UnknownRequests FailurePolicy
}
type NodeConfig struct {
	AllowRoRecoveryImport bool             // Explicit external-record admission; never automatic replay.
	RoTargets             []RoTargetConfig // Bounded recovery-only targets; ordinary Peer policies are unchanged.
	RoRecovery            RoRecoveryLimits
	Identifiers           IdentifierConfig
	Scheduler             SchedulerConfig
	Origin                pdu.Origin
	// Bundle must include the built-in Base profile. Application profiles remain explicit.
	Bundle    *schema.Bundle
	Limits    wire.Limits
	Resources ResourcePlan
	Timeouts  Timeouts
	Peers     []PeerConfig
	// Host addresses are provided as wire address payloads through LocalCapabilities.
	Capabilities LocalCapabilities
}

// InboundPolicy explicitly enables bounded application request handling. Sources
// are terminal Origin identities, independently of the adjacent peer. Destination
// realm must be local; an absent host is allowed, a present host must be local.
// Select either NoCache (each arrival is delivered independently) or an explicit
// DuplicatePolicy. A zero/partial cache configuration never silently disables it.
type InboundPolicy struct {
	ApplicationBinding              ApplicationBinding // Independent of the adjacent peer and source allowlist.
	Command                         pdu.CommandKey
	Class                           ClassID
	Sources                         []pdu.Origin
	MaxRequestBytes, MaxAnswerBytes int
	ExecutionTimeout                time.Duration
	NoCache                         bool
	Duplicate                       DuplicatePolicy
	FailurePolicy                   FailurePolicy
}
