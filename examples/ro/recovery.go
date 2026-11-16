package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/pdu"
	"io"
	"reflect"
	"time"
)

// All recovery policy is explicit application code in this example. The OCS is
// synthetic; its declaration is not evidence of production deduplication.
func recoveryConfigs() ([2]diameter.NodeConfig, error) {
	var cfg [2]diameter.NodeConfig
	origins := [2]pdu.Origin{{Host: "ctf.example", Realm: "example"}, {Host: "ocs.example", Realm: "example"}}
	for i := range cfg {
		c, e := endpointConfig(origins[i], origins[1-i], i == 0)
		if e != nil {
			return cfg, e
		}
		cfg[i] = c
	}
	cfg[0].RoRecovery = diameter.RoRecoveryLimits{MaxRecordBytes: 65536, MaxTargets: 1, MaxTotalAttempts: 3, MaxSegments: 3, MaxAttemptsPerSegment: 1, MaxReplayAge: time.Minute, MaxResponseWait: time.Second}
	cfg[0].Resources.RoRecoverySlots = 2
	cfg[0].Resources.RoRecoveryClasses = []diameter.IdentifierClass{{ID: "sms-charging", Slots: 2}}
	cfg[0].Resources.IdentifierBytes += 2048
	cfg[0].AllowRoRecoveryImport = true
	return cfg, nil
}
func recoveryOptions() diameter.RoRecoveryOptions {
	return diameter.RoRecoveryOptions{PolicyID: diameter.RoTransientRecoveryPolicy, Targets: []diameter.RoRecoveryTarget{{ID: "ocs", Peer: remotePeer, Retention: time.Minute}}, Limits: diameter.RoRecoveryLimits{MaxRecordBytes: 65536, MaxTargets: 1, MaxTotalAttempts: 3, MaxSegments: 3, MaxAttemptsPerSegment: 1, MaxReplayAge: time.Minute, MaxResponseWait: time.Second}}
}
func recoveryPlan() diameter.RoRecoveryPlan {
	return diameter.RoRecoveryPlan{TargetID: "ocs", Peers: []diameter.PeerID{remotePeer}, MaxAttempts: 1, Trigger: diameter.RoRecoveryInitial, ResponseWait: 500 * time.Millisecond, ErrorContinuation: diameter.RoErrorTransient}
}

type recoveryIntent struct {
	Record   diameter.RoRecoveryRequest
	Plan     diameter.RoRecoveryPlan
	Progress diameter.RoRecoveryProgress
}

// A digest binds this in-memory acknowledgement to the exact intent. It is not
// authentication, durable commit or fencing. Production supplies its own checks.
func intentDigest(i recoveryIntent) ([32]byte, error) {
	b, e := json.Marshal(i)
	return sha256.Sum256(b), e
}
func submitConfirmed(ctx context.Context, n *diameter.Node, i recoveryIntent, confirmed bool, digest [32]byte, port *diameter.CompletionPort) (diameter.RequestID, error) {
	actual, e := intentDigest(i)
	if e != nil {
		return diameter.RequestID{}, e
	}
	if !confirmed || digest != actual {
		return diameter.RequestID{}, errors.New("application intent unconfirmed")
	}
	return n.SubmitRoRecovery(ctx, i.Record, i.Plan, i.Progress, port)
}
func nextRecovery(record diameter.RoRecoveryRequest, result diameter.Result) (recoveryIntent, error) {
	r := result.RoRecovery
	if r == nil || !r.CanContinue || r.NextTrigger == 0 {
		return recoveryIntent{}, diameter.ErrRoRecoveryFinal
	}
	plan := recoveryPlan()
	plan.Trigger = r.NextTrigger
	return recoveryIntent{record, plan, diameter.RoRecoveryProgress{SegmentNo: r.SegmentNo + 1, ReservedBefore: r.ReservedThrough, SegmentAllowance: 1, ReservedThrough: r.ReservedThrough + 1, Delivery: r.CumulativeDelivery}}, nil
}
func receiveRecovery(ctx context.Context, p *pair, id diameter.RequestID) (diameter.Result, error) {
	c, e := p.left.port.Receive(ctx)
	if e != nil {
		return c.Result, e
	}
	if c.RequestID != id || c.Result.RoRecovery == nil {
		return c.Result, errors.New("unexpected recovery completion")
	}
	return c.Result, nil
}

// scenarios: error uses same-Node continuation; restart imports a settled error;
// no-answer imports an uncertain send after Tr and after closing the old owner.
func runRecovery(ctx context.Context, scenario, result string, out io.Writer) (r report, err error) {
	if scenario != "error" && scenario != "restart" && scenario != "no-answer" {
		return r, diameter.ErrConfig
	}
	if resultName(result) == "" {
		return r, diameter.ErrConfig
	}
	cfg, err := recoveryConfigs()
	if err != nil {
		return r, err
	}
	p, err := connectPair(ctx, cfg)
	if err != nil {
		return r, err
	}
	defer func() {
		err = errors.Join(err, p.close())
		r.Left, r.Right = p.left.node.Resources(), p.right.node.Resources()
		if r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
			err = errors.Join(err, errors.New("recovery resources retained"))
		}
	}()
	body, err := newDebit(p.left.node)
	if err != nil {
		return r, err
	}
	record, err := p.left.node.PrepareRoRecoveryRequest(ctx, body, recoveryOptions())
	if err != nil {
		return r, err
	}
	intent := recoveryIntent{record, recoveryPlan(), diameter.RoRecoveryProgress{SegmentNo: 1, SegmentAllowance: 1, ReservedThrough: 1}}
	digest, err := intentDigest(intent)
	if err != nil {
		return r, err
	}
	id, err := submitConfirmed(ctx, p.left.node, intent, true, digest, p.left.port)
	if err != nil {
		return r, err
	}
	in, err := p.right.peer.Receive(ctx)
	if err != nil {
		return r, err
	}
	original := in.Message()
	if scenario != "no-answer" {
		if _, err = in.Respond(ctx, &pdu.ErrorAnswer{ResultCode: 3004}); err != nil {
			return r, err
		}
	}
	first, err := receiveRecovery(ctx, p, id)
	if err != nil {
		return r, err
	}
	if scenario == "no-answer" {
		if first.Outcome != diameter.WrittenNoAnswer {
			return r, fmt.Errorf("expected no answer: %+v", first)
		}
	} else if first.Outcome != diameter.AnswerReceived || first.Err != nil || !first.Validation.Complete {
		return r, fmt.Errorf("invalid first result: %+v", first)
	}
	// Consume the actual first result separately from the example's explicit
	// recovery decision. A compatibility status never overwrites protocol evidence.
	reason := temporaryError
	if scenario == "no-answer" {
		reason = systemFailure
	}
	decision, err := consumeLocal(businessCharge, reason, fallbackConfig{TemporaryError: 1}, first)
	if err != nil {
		return r, err
	}
	var business businessConsumer
	business.notify(decision)
	intent, err = nextRecovery(record, first)
	if err != nil {
		return r, err
	}
	if scenario != "error" {
		if err = p.close(); err != nil {
			return r, err
		}
		if p.left.node.Resources() != (diameter.ResourceSnapshot{}) || p.right.node.Resources() != (diameter.ResourceSnapshot{}) {
			return r, errors.New("old owner not drained")
		}
		next, e := connectPair(ctx, cfg)
		if e != nil {
			return r, e
		}
		p = next
		// Old Node was joined above; progress is known in this process and checked now.
		intent.Plan.Import = &diameter.RoRecoveryImport{Checkpoint: first.RoRecovery.Checkpoint(), CheckedAt: time.Now().UTC(), SenderFenced: true, ProgressConfirmed: true, ClockConfirmed: true}
	}
	digest, err = intentDigest(intent)
	if err != nil {
		return r, err
	}
	id, err = submitConfirmed(ctx, p.left.node, intent, true, digest, p.left.port)
	if err != nil {
		return r, err
	}
	in, err = p.right.peer.Receive(ctx)
	if err != nil {
		return r, err
	}
	resumed := in.Message()
	a, b := original.Body.(*ro.CCR), resumed.Body.(*ro.CCR)
	if !reflect.DeepEqual(a, b) || original.Header.EndToEnd != resumed.Header.EndToEnd || a.SessionID != b.SessionID || a.CCRequestType != b.CCRequestType || a.CCRequestNumber != b.CCRequestNumber || a.OriginHost != b.OriginHost {
		return r, errors.New("original identity changed")
	}
	if (resumed.Header.Flags&pdu.RetransmittedFlag != 0) != (scenario == "no-answer") {
		return r, errors.New("wrong retransmission lineage")
	}
	answer, err := newAnswer(result)
	if err != nil {
		return r, err
	}
	if _, err = in.Respond(ctx, answer); err != nil {
		return r, err
	}
	final, err := receiveRecovery(ctx, p, id)
	if err != nil {
		return r, err
	}
	if final.Err != nil || !final.Validation.Complete || final.RoRecovery.ReservedThrough != 2 {
		return r, fmt.Errorf("invalid final result: %+v", final)
	}
	if err = checkAnswer(final.Answer.Body, result, body.SessionID); err != nil {
		return r, err
	}
	business.notify(businessDecision{Exchange: final}) // already notified; still consume final protocol result
	r.Answer = resultName(result)
	_, err = fmt.Fprintf(out, "recovery=%s first=%v final=%s reserved=%d continuation=%t\n", scenario, first.Outcome, r.Answer, final.RoRecovery.ReservedThrough, final.RoRecovery.CanContinue)
	return r, err
}

func validRecoveryOptions(flow, scenario string) bool {
	if flow == "compat" {
		return scenario == "error" || scenario == "restart" || scenario == "no-answer" || scenario == "no-answer-import"
	}
	if flow != "recovery" {
		return scenario == "error"
	}
	return scenario == "error" || scenario == "restart" || scenario == "no-answer"
}
