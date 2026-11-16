// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"github.com/skill215/diameterkit/wire"
)

const remotePeer diameter.PeerID = "remote"

// The CTF sends CCR; the OCS receives and answers it. Adjacent Peer and
// authorized terminal are the same here. A DRA deployment separates the two.
func endpointConfig(local, remote pdu.Origin, requester bool) (diameter.NodeConfig, error) {
	bp, err := base.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	sp, err := ro.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	command := (&ro.CCR{}).CommandKey()
	cfg := diameter.NodeConfig{
		Origin: local, Bundle: bundle,
		Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 1024, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16},
		Capabilities: diameter.LocalCapabilities{
			HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-ro-iec-example",
			// Profile loading alone advertises nothing. Capabilities are separate from application message AVPs.
			AuthApplicationID: []uint32{ro.ApplicationID},
		},
		Resources: diameter.ResourcePlan{
			NodeBytes: 256 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096,
			Classes:        []diameter.ClassConfig{{ID: "sms-charging", Bytes: 64 << 20, Slots: 2}},
			InboundClasses: []diameter.ClassConfig{{ID: "sms-charging", Bytes: 64 << 20, Slots: 2}},
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
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "sms-charging",
				DestinationHost: remote.Host, DestinationRealm: remote.Realm,
				Sources:         pdu.AnswerSources{Terminals: []pdu.Origin{remote}},
				MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second,
			}},
			InboundCommands: []diameter.InboundPolicy{{
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "sms-charging",
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

// All charging input is explicit and synthetic. No pricing, subscriber lookup,
// delivery state, session store or ledger is created by the example.
func newDebit(n *diameter.Node) (*ro.CCR, error) {
	session, err := n.NewSessionID()
	if err != nil {
		return nil, err
	}
	return &ro.CCR{SessionID: session, AuthApplicationID: ro.ApplicationID, ServiceContextID: ro.SMSServiceContextID,
		CCRequestType: ro.EventRequest, CCRequestNumber: 0, RequestedAction: avp.Some(ro.DirectDebiting), MultipleServicesIndicator: avp.Some(uint32(1)),
		MultipleServicesCreditControl: []ro.RequestMSCC{{ServiceIdentifier: []uint32{1}, RequestedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))})}},
		ServiceInformation:            avp.Some(ro.ServiceInformation{SMSInformation: avp.Some(ro.SMSInformation{SMSNode: avp.Some(uint32(1)), SMMessageType: avp.Some(uint32(0))})}),
	}, nil
}
func newAnswer(result string) (pdu.Body, error) {
	if result == "3004" {
		return &pdu.ErrorAnswer{ResultCode: 3004}, nil
	}
	top := uint32(2001)
	switch result {
	case "4012":
		top = 4012
	case "5012":
		top = 5012
	}
	a := &ro.CCA{AuthApplicationID: ro.ApplicationID, CCRequestType: ro.EventRequest, CCRequestNumber: 0, ResultCode: top}
	if top == 2001 {
		child := ro.AnswerMSCC{ServiceIdentifier: []uint32{1}, ResultCode: avp.Some(uint32(2001)), GrantedServiceUnit: avp.Some(ro.ServiceUnit{CCServiceSpecificUnits: avp.Some(uint64(1))}), RefundInformation: avp.Some([]byte{0, 255})}
		if result == "2001/4012" {
			child.ResultCode = avp.Some(uint32(4012))
			child.GrantedServiceUnit = avp.Optional[ro.ServiceUnit]{}
			child.RefundInformation = avp.Optional[[]byte]{}
		}
		a.MultipleServicesCreditControl = []ro.AnswerMSCC{child}
	}
	return a, nil
}
func resultName(result string) string {
	switch result {
	case "2001":
		return "base:2001/mscc:2001"
	case "2001/4012":
		return "base:2001/mscc:4012"
	case "4012":
		return "base:4012"
	case "5012":
		return "base:5012"
	case "3004":
		return "E:3004"
	default:
		return ""
	}
}
