package diameter

import (
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

type RequestID struct {
	Boot     [16]byte
	Sequence uint64
}
type ConnectionID struct {
	Boot     [16]byte
	Sequence uint64
}
type Outcome uint8

const (
	NotWritten Outcome = iota
	PartlyWritten
	WriteUncertain
	WrittenNoAnswer
	AnswerReceived
)

type DeliveryEvidence struct {
	Bytes, FrameBytes             int
	Uncertain, MayHaveReachedPeer bool
}
type Result struct {
	RoRecovery      *RoRecoveryResult // Owned segment progress; nil for ordinary Peer/Route requests.
	Validation      validation.Report
	Attempts        []AttemptResult
	SelectedAttempt AttemptID
	StopReason      StopReason
	RequestID       RequestID
	PeerID          PeerID
	ConnectionID    ConnectionID
	Answer          pdu.Message
	Outcome         Outcome
	Delivery        DeliveryEvidence
	Err             error
}
type Completion struct {
	RequestID RequestID
	Result    Result
}
