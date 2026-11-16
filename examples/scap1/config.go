package main

import (
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"github.com/skill215/diameterkit/wire"
)

const remotePeer diameter.PeerID = "remote"

// Both endpoints explicitly authorize sending AND receiving App0/271. This is
// a direct loopback example: adjacent peer and permitted terminal are the same.
// A DRA deployment must configure its adjacency and terminal sources separately.
func endpointConfig(local, remote pdu.Origin, mode diameter.SCAP1Advertisement) (diameter.NodeConfig, error) {
	bp, err := base.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	sp, err := scap1.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	command := pdu.CommandKey{ApplicationID: scap1.ApplicationID, Code: scap1.CommandCode, Request: true}
	return diameter.NodeConfig{
		Origin: local, Bundle: bundle,
		Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 1024, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16},
		Capabilities: diameter.LocalCapabilities{
			HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-scap1-example",
			// Local capabilities stay STRICT even when the peer uses the named
			// legacy advertisement. Profile loading alone advertises nothing.
			VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 193, AcctApplicationID: avp.Some(uint32(19302))}},
		},
		Resources: diameter.ResourcePlan{
			NodeBytes: 256 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096,
			Classes:        []diameter.ClassConfig{{ID: "accounting", Bytes: 64 << 20, Slots: 2}},
			InboundClasses: []diameter.ClassConfig{{ID: "accounting", Bytes: 64 << 20, Slots: 2}},
			MaxPorts:       1, MaxPortOutstanding: 1, MaxDiscardBytes: 4096,
			AdapterBytesPerConnection: tcp.BufferBytes, ListenerBytesPerPeer: tcp.ListenerBufferBytes,
		},
		Timeouts: diameter.Timeouts{
			Handshake: 2 * time.Second, FrameRead: 2 * time.Second, FrameWrite: 2 * time.Second,
			Request: 3 * time.Second, Watchdog: 30 * time.Second, ShutdownTimeout: 3 * time.Second,
			CloseReserve: time.Second, DPAWait: time.Second, ClosingGrace: time.Second,
		},
		Peers: []diameter.PeerConfig{{
			ID: remotePeer, Origin: remote, SCAP1Advertisement: mode,
			MaxPending: 2, MaxBytes: 32 << 20, MaxInbound: 2, MaxInboundBytes: 32 << 20,
			Commands: []diameter.CommandPolicy{{
				Command: command, ApplicationBinding: diameter.SCAP1Accounting, Class: "accounting",
				DestinationHost: remote.Host, DestinationRealm: remote.Realm,
				Sources:         pdu.AnswerSources{Terminals: []pdu.Origin{remote}},
				MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second,
			}},
			InboundCommands: []diameter.InboundPolicy{{
				Command: command, ApplicationBinding: diameter.SCAP1Accounting, Class: "accounting",
				Sources: []pdu.Origin{remote}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024,
				ExecutionTimeout: 2 * time.Second, NoCache: true,
			}},
		}},
	}, nil
}

func newEvent(n *diameter.Node) (*scap1.ACR, error) {
	session, err := n.NewSessionID()
	if err != nil {
		return nil, err
	}
	// Explicit synthetic input, including a fixed wire timestamp. These values
	// are not derived by the library from a subscriber, clock or charging store.
	return scap1.NewEvent(scap1.EventInput{
		SessionID: session, SubscriptionMSISDN: avp.Some("15551234567"), Timestamp: avp.Some(avp.Time(3908988800)),
		DataCodingScheme: avp.Some("8"), OriginatorMSISDN: avp.Some("15551234567"), OriginatorVMSC: avp.Some("15550000001"),
		TrafficCase: avp.Some("MO"), DestinationMSISDN: avp.Some("15557654321"),
		ConcatenatedMessages: avp.Some("1"), UserDataLength: avp.Some("140"),
	})
}
