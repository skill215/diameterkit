// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"github.com/skill215/diameterkit/wire"
)

const remotePeer diameter.PeerID = "remote"

// The GMSC sends SRR; the HSS receives and answers it. Adjacent Peer and
// authorized terminal are the same here. A DRA deployment separates the two.
func endpointConfig(local, remote pdu.Origin, requester bool) (diameter.NodeConfig, error) {
	bp, err := base.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	sp, err := s6c.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	command := (&s6c.SRR{}).CommandKey()
	cfg := diameter.NodeConfig{
		Origin: local, Bundle: bundle,
		Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 1024, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16},
		Capabilities: diameter.LocalCapabilities{
			HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-s6c-example",
			// Profile loading alone advertises nothing. Capabilities are separate from application message AVPs.
			VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(command.ApplicationID)}},
		},
		Resources: diameter.ResourcePlan{
			NodeBytes: 256 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096,
			Classes:        []diameter.ClassConfig{{ID: "s6c-routing", Bytes: 64 << 20, Slots: 2}},
			InboundClasses: []diameter.ClassConfig{{ID: "s6c-routing", Bytes: 64 << 20, Slots: 2}},
			MaxPorts:       1, MaxPortOutstanding: 1, MaxDiscardBytes: 4096,
			AdapterBytesPerConnection: tcp.BufferBytes, ListenerBytesPerPeer: tcp.ListenerBufferBytes,
		},
		Timeouts: diameter.Timeouts{
			Handshake: 2 * time.Second, FrameRead: 2 * time.Second, FrameWrite: 2 * time.Second,
			Request: 3 * time.Second, Watchdog: 30 * time.Second, ShutdownTimeout: 3 * time.Second,
			CloseReserve: time.Second, DPAWait: time.Second, ClosingGrace: time.Second,
		},
		Peers: []diameter.PeerConfig{{
			ID: remotePeer, Origin: remote,
			MaxPending: 2, MaxBytes: 32 << 20, MaxInbound: 2, MaxInboundBytes: 32 << 20,
			Commands: []diameter.CommandPolicy{{
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "s6c-routing",
				DestinationHost: remote.Host, DestinationRealm: remote.Realm,
				Sources:         pdu.AnswerSources{Terminals: []pdu.Origin{remote}},
				MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second,
			}},
			InboundCommands: []diameter.InboundPolicy{{
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "s6c-routing",
				Sources: []pdu.Origin{remote}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024,
				ExecutionTimeout: 2 * time.Second, NoCache: true,
			}},
		}},
	}
	if requester {
		cfg.Peers[0].InboundCommands = nil
		cfg.Peers[0].MaxInbound = 0
		cfg.Peers[0].MaxInboundBytes = 0
		cfg.Resources.InboundClasses = nil
	} else {
		cfg.Peers[0].Commands = nil
	}
	return cfg, nil
}

const subscriberIMSI = "001010123456789"

func newRouting(n *diameter.Node) (*s6c.SRR, error) {
	session, err := n.NewSessionID()
	if err != nil {
		return nil, err
	}
	return s6c.NewSRR(s6c.RoutingQuery{SessionID: session, MSISDN: avp.Some("12025550123"), SCAddress: "12025550100", Fields: s6c.SRR{SMRPMTI: avp.Some(uint32(0))}})
}

// The application supplies synthetic routing data and explicitly selects the
// namespace/outcome. The library does not look up subscribers or infer IMSI.
func newAnswer(result uint32) (pdu.Body, error) {
	if result == 3004 {
		return &pdu.ErrorAnswer{ResultCode: result}, nil
	}
	a := &s6c.SRA{AuthSessionState: 1, ResultCode: avp.Some(result)}
	switch result {
	case 2001:
		number, err := avp.EncodeTBCD("12025550199")
		if err != nil {
			return nil, err
		}
		a.UserName = avp.Some(subscriberIMSI)
		a.ServingNode = avp.Some(s6c.ServingNode{MMEName: avp.Some("mme.example"), MMERealm: avp.Some("example"), MMENumber: avp.Some(number)})
	case 5001:
		a.ResultCode = avp.Optional[uint32]{}
		a.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})
	}
	return a, nil
}

func resultName(result uint32) string {
	switch result {
	case 2001:
		return "base:2001"
	case 5012:
		return "base:5012"
	case 5001:
		return "experimental:10415/5001"
	case 3004:
		return "E:3004"
	default:
		return ""
	}
}
