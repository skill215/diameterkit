package main

import (
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/sgd"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"github.com/skill215/diameterkit/wire"
)

const remotePeer diameter.PeerID = "remote"

// Access node sends MO and receives MT; the combined IWMSC/GMSC example
// endpoint does the reverse. Adjacency and authorized terminal coincide here.
func endpointConfig(local, remote pdu.Origin, access bool) (diameter.NodeConfig, error) {
	bp, err := base.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	sp, err := sgd.Profile()
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	bundle, err := schema.Compile(bp, sp)
	if err != nil {
		return diameter.NodeConfig{}, err
	}
	out, in := (&sgd.OFR{}).CommandKey(), (&sgd.TFR{}).CommandKey()
	if !access {
		out, in = in, out
	}
	return diameter.NodeConfig{
		Origin: local, Bundle: bundle,
		Limits: wire.Limits{MaxFrameBytes: 4096, MaxFieldBytes: 1024, MaxAVPs: 64, MaxDepth: 8, MaxRepeated: 16},
		Capabilities: diameter.LocalCapabilities{
			HostIPAddress: []avp.Address{{Family: 1, Data: []byte{127, 0, 0, 1}}}, ProductName: "diameterkit-sgd-example",
			// Profile loading alone advertises nothing. Application capability and message grammar are separate.
			VendorSpecificApplicationID: []pdu.VendorSpecificApplicationID{{VendorID: 10415, AuthApplicationID: avp.Some(out.ApplicationID)}},
		},
		Resources: diameter.ResourcePlan{
			NodeBytes: 256 << 20, TransportBytes: 32 << 20, IdentifierBytes: 4096,
			Classes:        []diameter.ClassConfig{{ID: "sms-forward", Bytes: 64 << 20, Slots: 2}},
			InboundClasses: []diameter.ClassConfig{{ID: "sms-forward", Bytes: 64 << 20, Slots: 2}},
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
				Command: out, ApplicationBinding: diameter.HeaderApplication, Class: "sms-forward",
				DestinationHost: remote.Host, DestinationRealm: remote.Realm,
				Sources:         pdu.AnswerSources{Terminals: []pdu.Origin{remote}},
				MaxRequestBytes: 1024, MaxAnswerBytes: 1024, Timeout: 2 * time.Second,
			}},
			InboundCommands: []diameter.InboundPolicy{{
				Command: in, ApplicationBinding: diameter.HeaderApplication, Class: "sms-forward",
				Sources: []pdu.Origin{remote}, MaxRequestBytes: 1024, MaxAnswerBytes: 1024,
				ExecutionTimeout: 2 * time.Second, NoCache: true,
			}},
		}},
	}, nil
}

// Synthetic bytes are opaque; this example neither builds a valid SMS TPDU
// nor interprets a returned SM-RP-UI as proof of delivery.
func requestPayload() []byte { return []byte{1, 0, 255, 4} }
func answerPayload() []byte  { return []byte{0, 255, 2} }

func newRequest(n *diameter.Node, mt, gdd bool) (pdu.Body, error) {
	session, err := n.NewSessionID()
	if err != nil {
		return nil, err
	}
	sc, err := avp.EncodeTBCD("12025550100")
	if err != nil {
		return nil, err
	}
	if !mt {
		return &sgd.OFR{SessionID: session, AuthSessionState: 1, UserIdentifier: sgd.UserIdentifier{UserName: avp.Some("001010123456789")}, SCAddress: sc, SMRPUI: requestPayload()}, nil
	}
	number, err := avp.EncodeTBCD("12025550199")
	if err != nil {
		return nil, err
	}
	r := &sgd.TFR{SessionID: session, AuthSessionState: 1, UserName: "001010123456789", SCAddress: sc, SMRPUI: requestPayload()}
	if gdd {
		r.SGSNNumber = avp.Some(number)
	} else {
		r.MMENumberForMTSMS = avp.Some(number)
	}
	return r, nil
}

func newAnswer(mt bool, result uint32) (pdu.Body, error) {
	if result == 3004 {
		return &pdu.ErrorAnswer{ResultCode: result}, nil
	}
	rc := avp.Some(result)
	exp := avp.Optional[sgd.ExperimentalResult]{}
	cause := avp.Optional[sgd.SMDeliveryFailureCause]{}
	payload := avp.Optional[[]byte]{}
	if result == 2001 {
		payload = avp.Some(answerPayload())
	}
	if result == 5555 {
		rc = avp.Optional[uint32]{}
		exp = avp.Some(sgd.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5555})
		code := uint32(3)
		if mt {
			code = 0
		}
		cause = avp.Some(sgd.SMDeliveryFailureCause{SMEnumeratedDeliveryFailureCause: code})
	}
	if mt {
		return &sgd.TFA{AuthSessionState: 1, ResultCode: rc, ExperimentalResult: exp, SMDeliveryFailureCause: cause, SMRPUI: payload}, nil
	}
	return &sgd.OFA{AuthSessionState: 1, ResultCode: rc, ExperimentalResult: exp, SMDeliveryFailureCause: cause, SMRPUI: payload}, nil
}
func resultName(result uint32) string {
	switch result {
	case 2001:
		return "base:2001"
	case 5012:
		return "base:5012"
	case 5555:
		return "experimental:10415/5555"
	case 3004:
		return "E:3004"
	default:
		return ""
	}
}
