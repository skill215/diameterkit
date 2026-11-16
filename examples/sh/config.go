package main

import (
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"github.com/skill215/diameterkit/wire"
)

const remotePeer diameter.PeerID = "remote"

// Both endpoints explicitly authorize sending AND receiving Sh App16777217/306. This is
// a direct loopback example: adjacent peer and permitted terminal are the same.
// A DRA deployment must configure its adjacency and terminal sources separately.
func endpointConfig(local, remote pdu.Origin) (diameter.NodeConfig, error) {
	bp, err := base.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	sp, err := sh.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	command := pdu.CommandKey{ApplicationID: sh.ApplicationID, Code: 306, Request: true}
	return diameter.NodeConfig{
		Origin: local, Bundle: bundle,
		Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 1024, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16},
		Capabilities: diameter.LocalCapabilities{
			HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-sh-example",
			// Profile loading alone advertises nothing. Sh uses Auth VSAI.
			VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{sh.Application()},
		},
		Resources: diameter.ResourcePlan{
			NodeBytes: 256 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096,
			Classes:        []diameter.ClassConfig{{ID: "sh-pull", Bytes: 64 << 20, Slots: 2}},
			InboundClasses: []diameter.ClassConfig{{ID: "sh-pull", Bytes: 64 << 20, Slots: 2}},
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
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "sh-pull",
				DestinationHost: remote.Host, DestinationRealm: remote.Realm,
				Sources:         pdu.AnswerSources{Terminals: []pdu.Origin{remote}},
				MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second,
			}},
			InboundCommands: []diameter.InboundPolicy{{
				Command: command, ApplicationBinding: diameter.HeaderApplication, Class: "sh-pull",
				Sources: []pdu.Origin{remote}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024,
				ExecutionTimeout: 2 * time.Second, NoCache: true,
			}},
		}},
	}, nil
}

// XML parsing is application work with independent finite limits. These limits
// are not a reservation in Node.Resources and make no claim about heap/RSS.
func dataLimits() shdata.Limits {
	return shdata.Limits{MaxBytes: 512, MaxDepth: 12, MaxElements: 24, MaxAttributes: 8, MaxTextBytes: 128, MaxOpaqueBytes: 128}
}

func newPull(n *diameter.Node) (*sh.UDR, error) {
	session, err := n.NewSessionID()
	if err != nil {
		return nil, err
	}
	return &sh.UDR{SessionID: session, VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1,
		UserIdentity: sh.UserIdentity{PublicIdentity: avp.Some("sip:alice@example.net")}, DataReference: 24}, nil
}

// Both data and result are explicit synthetic application input, not HSS state.
func newAnswer(result uint32) (pdu.Body, error) {
	if result == 3004 {
		return &pdu.ErrorAnswer{ResultCode: result}, nil
	}
	a := &sh.UDA{VendorSpecificApplicationID: sh.Application(), AuthSessionState: 1, ResultCode: avp.Some(result)}
	if result == 2001 {
		raw, err := shdata.MarshalSMSPull(shdata.SMSRegistration{Number: "15551234567", SCAddress: avp.Some("15550000001"), SBI: avp.Some(uint8(0))}, dataLimits())
		if err != nil {
			return nil, err
		}
		a.UserData = avp.Some(raw)
	}
	return a, nil
}
