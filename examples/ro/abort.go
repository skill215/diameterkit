// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/app/ro/abort"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

// This is an explicit synthetic reference to a CTF-owned session, not a new
// session allocated for ASR or a lookup performed by the Diameter library.
const knownAbortSession = "ctf.example;example;original"

func abortConfigs() ([2]diameter.NodeConfig, error) {
	var cfg [2]diameter.NodeConfig
	origins := [2]pdu.Origin{{Host: "ocs.example", Realm: "example"}, {Host: "ctf.example", Realm: "example"}}
	bp, err := base.Profile()
	if err != nil {
		return cfg, err
	}
	rp, err := ro.Profile()
	if err != nil {
		return cfg, err
	}
	ap, err := abort.Profile()
	if err != nil {
		return cfg, err
	}
	bundle, err := schema.Compile(bp, rp, ap)
	if err != nil {
		return cfg, err
	}
	for i := range cfg {
		// Reuse the finite resource plan; authorize only ASR in the reverse business
		// direction. Loading the CCR profile does not authorize sending CCR.
		cfg[i], err = endpointConfig(origins[i], origins[1-i], i == 0)
		if err != nil {
			return cfg, err
		}
		cfg[i].Bundle = bundle
		cfg[i].Capabilities.ProductName = "diameterkit-ro-abort-example"
		for j := range cfg[i].Peers[0].Commands {
			cfg[i].Peers[0].Commands[j].Command = (&abort.ASR{}).CommandKey()
		}
		for j := range cfg[i].Peers[0].InboundCommands {
			cfg[i].Peers[0].InboundCommands[j].Command = (&abort.ASR{}).CommandKey()
		}
	}
	return cfg, nil
}
func openAbortPair(ctx context.Context) (*pair, error) {
	cfg, e := abortConfigs()
	if e != nil {
		return nil, e
	}
	return connectPair(ctx, cfg)
}
func abortResultName(result string) string {
	switch result {
	case "2001":
		return "base:2001"
	case "5002":
		return "base:5002"
	case "5012":
		return "base:5012"
	case "3004":
		return "E:3004"
	default:
		return ""
	}
}

// disposition is the example application's reported outcome, not an instruction
// for Node to stop a session. A real CTF must supply its own session decision and
// coordinate any outstanding CCR, usage reporting or TERM explicitly.
func decideAbort(r *abort.ASR, knownSession, disposition string) (pdu.Body, error) {
	if r == nil || r.SessionID == "" || abortResultName(disposition) == "" {
		return nil, fmt.Errorf("invalid abort decision input")
	}
	if knownSession == "" || r.SessionID != knownSession {
		return &abort.ASA{ResultCode: abort.UnknownSessionID}, nil
	}
	switch disposition {
	case "2001":
		return &abort.ASA{ResultCode: abort.Success}, nil
	case "5002":
		return &abort.ASA{ResultCode: abort.UnknownSessionID}, nil
	case "5012":
		return &abort.ASA{ResultCode: abort.UnableToComply}, nil
	default:
		return &pdu.ErrorAnswer{ResultCode: 3004}, nil
	}
}
func checkAbortAnswer(body pdu.Body, result, session string) error {
	switch a := body.(type) {
	case *abort.ASA:
		if (a.ResultCode != abort.Success && a.ResultCode != abort.UnknownSessionID && a.ResultCode != abort.UnableToComply) || a.SessionID != session || abortResultName(fmt.Sprint(a.ResultCode)) != result || a.OriginStateID.Present {
			return fmt.Errorf("ASA result/session/local state mismatch")
		}
	case *pdu.ErrorAnswer:
		if result != "E:3004" || a.ResultCode != 3004 || !a.SessionID.Present || a.SessionID.Value != session {
			return fmt.Errorf("abort E result/session mismatch")
		}
	default:
		return fmt.Errorf("unexpected abort answer")
	}
	return nil
}
func exchangeAbort(ctx context.Context, p *pair, session, disposition string) (string, error) {
	request := &abort.ASR{SessionID: session, AuthApplicationID: abort.ApplicationID}
	id, e := p.left.peer.Submit(ctx, request, diameter.RequestOptions{}, p.left.port)
	if e != nil {
		return "", e
	}
	in, e := p.right.peer.Receive(ctx)
	if e != nil {
		return "", e
	}
	r, ok := in.Message().Body.(*abort.ASR)
	if !ok || r.SessionID != session || !in.Validation().Complete {
		return "", fmt.Errorf("invalid abort request/validation")
	}
	answer, e := decideAbort(r, knownAbortSession, disposition)
	if e != nil {
		return "", e
	}
	result := "E:3004"
	if a, ok := answer.(*abort.ASA); ok {
		result = abortResultName(fmt.Sprint(a.ResultCode))
	}
	if sent, e := in.Respond(ctx, answer); e != nil || !sent.Committed {
		return "", errors.Join(e, fmt.Errorf("ASA not committed"))
	}
	c, e := p.left.port.Receive(ctx)
	if e != nil {
		return "", e
	}
	if c.RequestID != id || c.Result.Err != nil || c.Result.Outcome != diameter.AnswerReceived || !c.Result.Validation.Complete {
		return "", fmt.Errorf("abort completion: %+v", c.Result)
	}
	if e := checkAbortAnswer(c.Result.Answer.Body, result, session); e != nil {
		return "", e
	}
	return result, nil
}
func runAbort(ctx context.Context, session, disposition string, out io.Writer) (r report, err error) {
	if session == "" || abortResultName(disposition) == "" {
		return r, fmt.Errorf("invalid abort session or disposition")
	}
	p, err := openAbortPair(ctx)
	if err != nil {
		return r, err
	}
	defer func() {
		err = errors.Join(err, p.close())
		r.Left, r.Right = p.left.node.Resources(), p.right.node.Resources()
		if r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
			err = errors.Join(err, fmt.Errorf("resources retained after abort cleanup"))
		}
	}()
	r.Answer, err = exchangeAbort(ctx, p, session, disposition)
	if err != nil {
		return r, err
	}
	if err = p.left.node.Shutdown(ctx); err != nil {
		return r, err
	}
	if err = p.right.node.Shutdown(ctx); err != nil {
		return r, err
	}
	_, err = fmt.Fprintf(out, "ocs->ctf abort=%s\n", r.Answer)
	return r, err
}

// Keep flow-specific arguments explicit instead of silently ignoring them.
func validFlowOptions(flow, result, termResult, session string) bool {
	switch flow {
	case "compat":
		return session == knownAbortSession && termResult == "2001" && (result == "2001" || result == "4012" || result == "5012" || result == "3004")
	case "abort":
		return session != "" && termResult == "2001" && abortResultName(result) != ""
	case "iec", "recovery":
		return session == knownAbortSession && termResult == "2001" && resultName(result) != ""
	case "ecur":
		return session == knownAbortSession && termResultName(termResult) != "" && resultName(result) != ""
	default:
		return false
	}
}
