package main

import (
	"context"
	"errors"
	"fmt"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/transport/tcp"
	"io"
	"net"
	"time"
)

const backupPeer diameter.PeerID = "backup"

var backupOrigin = pdu.Origin{Host: "backup-ocs.example", Realm: "example"}

// All declarations below describe synthetic endpoints only. No discovery,
// persistent state, shared financial ledger or production authorization exists.
func compatConfigs() ([2]diameter.NodeConfig, diameter.NodeConfig, error) {
	return boundCompatConfigs("combined")
}

func boundCompatConfigs(binding string) ([2]diameter.NodeConfig, diameter.NodeConfig, error) {
	cfg, e := recoveryConfigs()
	if e != nil {
		return cfg, diameter.NodeConfig{}, e
	}
	bp, e := base.Profile()
	if e != nil {
		return cfg, diameter.NodeConfig{}, e
	}
	cp, e := compatProfile(binding)
	if e != nil {
		return cfg, diameter.NodeConfig{}, e
	}
	bundle, e := schema.Compile(bp, cp)
	if e != nil {
		return cfg, diameter.NodeConfig{}, e
	}
	for i := range cfg {
		cfg[i].Bundle = bundle
		// Combined decode ownership has a larger bounded inbound reserve than standard Ro.
		if i == 1 {
			cfg[i].Peers[0].MaxInboundBytes = 128 << 20
		}
		cfg[i].Resources.NodeBytes = 512 << 20
		for j := range cfg[i].Resources.InboundClasses {
			cfg[i].Resources.InboundClasses[j].Bytes = 128 << 20
		}
	}
	secondary, e := endpointConfig(backupOrigin, cfg[0].Origin, false)
	if e != nil {
		return cfg, secondary, e
	}
	secondary.Bundle = bundle
	secondary.Peers[0].MaxInboundBytes = 128 << 20
	secondary.Resources.NodeBytes = 512 << 20
	secondary.Resources.InboundClasses[0].Bytes = 128 << 20
	peer := cfg[0].Peers[0]
	peer.ID = backupPeer
	peer.Origin = backupOrigin
	peer.Commands = append([]diameter.CommandPolicy(nil), peer.Commands...)
	peer.Commands[0].DestinationHost = backupOrigin.Host
	peer.Commands[0].DestinationRealm = backupOrigin.Realm
	peer.Commands[0].Sources = pdu.AnswerSources{Terminals: []pdu.Origin{backupOrigin}}
	cfg[0].Peers = append(cfg[0].Peers, peer)
	cfg[0].RoRecovery.MaxTargets = 2
	cfg[0].Resources.TransportBytes = 64 << 20
	cfg[0].RoTargets = []diameter.RoTargetConfig{
		{TargetID: "ocs", Peers: []diameter.PeerID{remotePeer}, Destination: cfg[1].Origin, ReplayDomain: "synthetic-shared-domain", EvidenceID: "synthetic-primary", Retention: time.Minute, SessionEvidenceID: "synthetic-primary-session"},
		{TargetID: "backup", Peers: []diameter.PeerID{backupPeer}, Destination: backupOrigin, ReplayDomain: "synthetic-shared-domain", EvidenceID: "synthetic-backup", Retention: time.Minute, SessionEvidenceID: "synthetic-backup-session"},
	}
	return cfg, secondary, nil
}

// Two pair owners share only the client endpoint; Close is idempotent. Joining
// both owners also joins both servers/listeners and checks all three Nodes.
func openCompatTargets(ctx context.Context, cfg [2]diameter.NodeConfig, secondary diameter.NodeConfig) (primary, backup *pair, err error) {
	primary, err = connectPair(ctx, cfg)
	if err != nil {
		return
	}
	backup = &pair{left: primary.left}
	defer func() {
		if err != nil {
			err = errors.Join(err, primary.close(), backup.close())
		}
	}()
	backup.right.origin = secondary.Origin
	backup.right.node, err = diameter.NewNode(secondary)
	if err != nil {
		return
	}
	backup.right.peer, _ = backup.right.node.Peer(remotePeer)
	raw, e := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		err = e
		return
	}
	listener, e := tcp.NewListener(raw)
	if e != nil {
		err = e
		return
	}
	dialer, e := tcp.NewDialer("tcp4", listener.Addr().String())
	if e != nil {
		_ = listener.Close()
		err = e
		return
	}
	serveCtx, cancel := context.WithCancel(ctx)
	backup.cancelServe = cancel
	backup.serveDone = make(chan error, 1)
	server, done := backup.right.node, backup.serveDone
	go func() { done <- server.Serve(serveCtx, remotePeer, listener) }()
	_, err = primary.left.node.Dial(ctx, backupPeer, dialer)
	return
}
func compatRequest(action, session string) (*combined.CCR, error) {
	general := combined.GeneralInput{SessionID: session, ServiceContextID: "legacy@example.net", SubscriptionMSISDN: "15551234567", SubscriptionIMSI: "001010123456789", OriginatorMSISDN: "15551234567", ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ServiceSpecificUnits: 0, ProtocolID: []byte{0, 255}}
	ec := combined.ECURInput{SessionID: session, ServiceContextID: general.ServiceContextID, SubscriptionMSISDN: general.SubscriptionMSISDN, SubscriptionIMSI: general.SubscriptionIMSI, SMSCAddress: "15550000001", RecipientSCCPAddress: "15550000003", OriginatorMSISDN: general.OriginatorMSISDN, RecipientMSISDN: "15557654321", SMSKey: "synthetic-1", ProtocolID: []byte{0, 255}}
	switch action {
	case "debit":
		return combined.NewDirectDebit(general), nil
	case "balance":
		return combined.NewBalanceCheck(general), nil
	case "refund":
		return combined.NewRefund(combined.RefundInput{SessionID: session, ServiceContextID: general.ServiceContextID, SubscriptionMSISDN: general.SubscriptionMSISDN, SubscriptionIMSI: general.SubscriptionIMSI, ValueDigits: -123, Exponent: -2, CurrencyCode: 840, ProtocolID: []byte{0, 255}, RefundInformation: avp.Some([]byte{0, 255})})
	case "initial":
		return combined.NewInitial(ec), nil
	case "term":
		return combined.NewTermination(combined.TerminationInput{ECURInput: ec, UsedServiceSpecificUnits: 0, Status: []byte{0, 255}}), nil
	default:
		return nil, diameter.ErrConfig
	}
}
func compatAnswer(q *combined.CCR, code uint32) pdu.Body {
	if code == 3004 {
		return &pdu.ErrorAnswer{ResultCode: code}
	}
	a := &combined.CCA{AuthApplicationID: 4, CCRequestType: q.CCRequestType, CCRequestNumber: q.CCRequestNumber, ResultCode: code}
	if code == 2001 && q.RequestedAction.Present && q.RequestedAction.Value == 2 {
		a.CheckBalanceResult = avp.Some(uint32(1))
	}
	return a
}
func compatOptions() diameter.RoRecoveryOptions {
	o := recoveryOptions()
	o.PolicyID = diameter.RoSMSFWRecoveryPolicy
	o.SessionEvidenceID = "synthetic-application-session"
	o.Limits.MaxTargets = 2
	o.Targets = append(o.Targets, diameter.RoRecoveryTarget{ID: "backup", Peer: backupPeer, Retention: time.Minute})
	return o
}

// The first OCS returns E3004; only the explicit application decision submits
// the next segment to the already frozen second target. No implicit failback.
func runCompat(ctx context.Context, action string, code uint32, out io.Writer) (err error) {
	return runBoundCompat(ctx, "combined", action, "error", code, out)
}

func runBoundCompat(ctx context.Context, binding, action, scenario string, code uint32, out io.Writer) (err error) {
	stage := "open"
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %w", stage, err)
		}
	}()
	if !validBoundCompat(binding, action, scenario) || (code != 2001 && code != 4012 && code != 5012 && code != 3004) {
		return diameter.ErrConfig
	}
	cfg, secondary, e := boundCompatConfigs(binding)
	if e != nil {
		return e
	}
	primary, backup, e := openCompatTargets(ctx, cfg, secondary)
	if e != nil {
		return e
	}
	defer func() {
		err = errors.Join(err, closeCompatTargets(primary, backup))
	}()
	session := "ctf.example;synthetic-existing-session"
	if action != "term" {
		session, e = primary.left.node.NewSessionID()
	}
	if e != nil {
		return e
	}
	body, e := boundCompatRequest(binding, action, session)
	if e != nil {
		return e
	}
	record, e := primary.left.node.PrepareRoRecoveryRequest(ctx, body, compatOptions())
	if e != nil {
		return e
	}
	intent := recoveryIntent{record, recoveryPlan(), diameter.RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: 1, ReservedThrough: 1}}
	digest, e := intentDigest(intent)
	if e != nil {
		return e
	}
	id, e := submitConfirmed(ctx, primary.left.node, intent, true, digest, primary.left.port)
	if e != nil {
		return e
	}
	stage = "primary receive"
	in, e := primary.right.peer.Receive(ctx)
	if e != nil {
		return e
	}
	firstRequest := in.Message()
	originalInbound := in
	noAnswer := scenario == "no-answer" || scenario == "no-answer-import"
	if !noAnswer {
		if _, e = in.Respond(ctx, &pdu.ErrorAnswer{ResultCode: 3004}); e != nil {
			return e
		}
	}
	stage = "primary completion"
	first, e := receiveRecovery(ctx, primary, id)
	if e != nil {
		return e
	}
	if noAnswer {
		if first.Outcome != diameter.WrittenNoAnswer || !first.RoRecovery.ReplayPending || !first.RoRecovery.CumulativeDelivery.MayHaveReachedPeer {
			return fmt.Errorf("invalid no-answer result: %+v", first)
		}
	} else if first.Err != nil || !first.Validation.Complete || first.Outcome != diameter.AnswerReceived {
		return fmt.Errorf("invalid primary result: %+v", first)
	}
	intent, e = nextRecovery(record, first)
	if e != nil {
		return e
	}
	intent.Plan.TargetID = "backup"
	intent.Plan.Peers = []diameter.PeerID{backupPeer}
	if scenario == "restart" || scenario == "no-answer-import" {
		stage = "join old owners"
		if e = closeCompatTargets(primary, backup); e != nil {
			return e
		}
		stage = "open new owners"
		next, nextBackup, openErr := openCompatTargets(ctx, cfg, secondary)
		if openErr != nil {
			return openErr
		}
		primary, backup = next, nextBackup
		// All old Nodes have been joined. This in-process checkpoint is not storage.
		intent.Plan.Import = &diameter.RoRecoveryImport{Checkpoint: first.RoRecovery.Checkpoint(), CheckedAt: time.Now().UTC(), SenderFenced: true, ProgressConfirmed: true, ClockConfirmed: true}
	}
	digest, e = intentDigest(intent)
	if e != nil {
		return e
	}
	id, e = submitConfirmed(ctx, primary.left.node, intent, true, digest, primary.left.port)
	if e != nil {
		return e
	}
	stage = "backup receive"
	in, e = backup.right.peer.Receive(ctx)
	if e != nil {
		return e
	}
	secondRequest := in.Message()
	if !sameCompatRequest(firstRequest.Body, secondRequest.Body) || firstRequest.Header.EndToEnd != secondRequest.Header.EndToEnd || (secondRequest.Header.Flags&pdu.RetransmittedFlag != 0) != noAnswer {
		return errors.New("cross-target identity mismatch")
	}
	if _, e = in.Respond(ctx, boundCompatAnswer(secondRequest.Body, code)); e != nil {
		return e
	}
	stage = "backup completion"
	final, e := receiveRecovery(ctx, primary, id)
	if e != nil {
		return e
	}
	if final.Err != nil || final.Outcome != diameter.AnswerReceived || !final.Validation.Complete || final.RoRecovery.ReservedThrough != 2 || len(final.Attempts) != 1 || final.Attempts[0].TargetID != "backup" {
		return fmt.Errorf("invalid backup result: %+v", final)
	}
	if e = checkCompatAnswer(binding, action, code, final.Answer.Body); e != nil {
		return e
	}
	if scenario == "no-answer" {
		// The original OCS may still execute. Its late answer must not replace the
		// already consumed first completion or the explicit backup completion.
		stage = "late primary answer"
		if _, e = originalInbound.Respond(ctx, boundCompatAnswer(firstRequest.Body, 2001)); e != nil {
			return e
		}
		wait, stop := context.WithTimeout(ctx, 20*time.Millisecond)
		defer stop()
		if _, e = primary.left.port.Receive(wait); !errors.Is(e, context.DeadlineExceeded) {
			return fmt.Errorf("unexpected late completion: %v", e)
		}
	}
	_, err = fmt.Fprintf(out, "%s action=%s scenario=%s first=%v backup=%d reserved=2 continue=%t\n", binding, action, scenario, first.Outcome, code, final.RoRecovery.CanContinue)
	return
}

func validCompatOptions(flow, action string) bool {
	if flow != "compat" {
		return action == "debit"
	}
	switch action {
	case "debit", "balance", "refund", "initial", "term":
		return true
	}
	return false
}
