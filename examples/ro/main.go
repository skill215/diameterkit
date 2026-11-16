// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Command ro runs bounded, nonpersistent SMS IEC, no-UPDATE ECUR or independent ASR/ASA endpoints.
// CTF and OCS use only public APIs on loopback TCP; flow decisions are explicit.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/ro"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport/tcp"
)

type endpoint struct {
	origin pdu.Origin
	node   *diameter.Node
	peer   *diameter.Peer
	port   *diameter.CompletionPort
}
type pair struct {
	left, right endpoint
	cancelServe context.CancelFunc
	serveDone   chan error
}
type report struct {
	Answer      string
	Left, Right diameter.ResourceSnapshot
}

// Serve owns the listener, Dial owns the connected transport, and Node owns
// their lifetime. The two Nodes have separate origins, limits and completions.
func openPair(ctx context.Context) (*pair, error) {
	origins := [2]pdu.Origin{{Host: "ctf.example", Realm: "example"}, {Host: "ocs.example", Realm: "example"}}
	var configs [2]diameter.NodeConfig
	for i := range configs {
		c, e := endpointConfig(origins[i], origins[1-i], i == 0)
		if e != nil {
			return nil, e
		}
		configs[i] = c
	}
	return connectPair(ctx, configs)
}

func connectPair(ctx context.Context, configs [2]diameter.NodeConfig) (p *pair, err error) {
	p = &pair{}
	defer func() {
		if err != nil {
			err = errors.Join(err, p.close())
			p = nil
		}
	}()
	for i, ep := range []*endpoint{&p.left, &p.right} {
		var e error
		ep.origin = configs[i].Origin
		ep.node, e = diameter.NewNode(configs[i])
		if e != nil {
			return p, e
		}
		ep.peer, _ = ep.node.Peer(remotePeer) // Both configs explicitly contain it.
		if i == 0 {
			ep.port, e = ep.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
			if e != nil {
				return p, e
			}
		}
	}
	raw, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return p, err
	}
	listener, err := tcp.NewListener(raw) // Ownership transfers, even on failure.
	if err != nil {
		return p, err
	}
	dialer, err := tcp.NewDialer("tcp4", listener.Addr().String())
	if err != nil {
		_ = listener.Close()
		return p, err
	}
	serveCtx, cancel := context.WithCancel(ctx)
	p.cancelServe = cancel
	p.serveDone = make(chan error, 1)
	server, done := p.right.node, p.serveDone
	go func() { done <- server.Serve(serveCtx, remotePeer, listener) }()
	p.left.peer, err = p.left.node.Dial(ctx, remotePeer, dialer)
	if err != nil {
		return p, err
	}
	// Mixed CCR/ASR flows may submit immediately from either endpoint. Dial
	// completion does not join the accepting endpoint's CEA write barrier.
	ready := time.NewTicker(time.Millisecond)
	defer ready.Stop()
	for !p.left.peer.Info().Ready || !p.right.peer.Info().Ready {
		select {
		case <-ctx.Done():
			return p, ctx.Err()
		case <-ready.C:
		}
	}
	return p, nil
}

// Cleanup uses its own finite context even if the operation context expired.
// Close stops admission/I/O; port.Close drains retained completions; Wait joins
// transport and listener borrowers. FixedBytes can reach zero only after both.
func (p *pair) close() error {
	if p.cancelServe != nil {
		p.cancelServe()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var errs []error
	for _, ep := range []*endpoint{&p.left, &p.right} {
		if ep.node != nil {
			errs = append(errs, ep.node.Close())
		}
	}
	for _, ep := range []*endpoint{&p.left, &p.right} {
		if ep.port != nil {
			errs = append(errs, ep.port.Close(ctx))
		}
		if ep.node != nil {
			errs = append(errs, ep.node.Wait(ctx))
		}
	}
	if p.serveDone != nil {
		select {
		case e := <-p.serveDone:
			if e != nil && !errors.Is(e, context.Canceled) && !errors.Is(e, context.DeadlineExceeded) && !errors.Is(e, diameter.ErrClosed) && !errors.Is(e, diameter.ErrShutdown) {
				errs = append(errs, e)
			}
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
		}
		p.serveDone = nil
	}
	return errors.Join(errs...)
}

func exchange(ctx context.Context, from, to *endpoint, result string) error {
	request, err := newDebit(from.node)
	if err != nil {
		return err
	}
	id, err := from.peer.Submit(ctx, request, diameter.RequestOptions{}, from.port)
	if err != nil {
		return err
	}
	in, err := to.peer.Receive(ctx)
	if err != nil {
		return err
	}
	if !in.Validation().Complete {
		return fmt.Errorf("incomplete inbound validation: %v", in.Validation())
	}
	ccr, ok := in.Message().Body.(*ro.CCR)
	if !ok || ccr.SessionID != request.SessionID {
		return fmt.Errorf("unexpected inbound request")
	}
	// Respond restores the original Session/Proxy/header and local Origin.
	answer, err := newAnswer(result)
	if err != nil {
		return err
	}
	r, err := in.Respond(ctx, answer)
	if err != nil {
		return err
	}
	if !r.Committed {
		return fmt.Errorf("answer not committed")
	}
	completion, err := from.port.Receive(ctx)
	if err != nil {
		return err
	}
	if completion.RequestID != id || completion.Result.Err != nil || completion.Result.Outcome != diameter.AnswerReceived || !completion.Result.Validation.Complete {
		return fmt.Errorf("request completion: %+v", completion.Result)
	}
	// AnswerReceived and protocol Validation.Complete are not financial success.
	return checkAnswer(completion.Result.Answer.Body, result, request.SessionID)
}

func checkAnswer(body pdu.Body, result, session string) error {
	switch a := body.(type) {
	case *pdu.ErrorAnswer:
		if result != "3004" || a.ResultCode != 3004 || !a.SessionID.Present || a.SessionID.Value != session {
			return fmt.Errorf("E result/correlation mismatch")
		}
	case *ro.CCA:
		if a.SessionID != session || a.CCRequestType != ro.EventRequest || a.CCRequestNumber != 0 {
			return fmt.Errorf("CCA identity mismatch")
		}
		top := uint32(2001)
		switch result {
		case "4012":
			top = 4012
		case "5012":
			top = 5012
		case "3004":
			return fmt.Errorf("expected E answer")
		}
		if a.ResultCode != top || a.ExperimentalResult.Present {
			return fmt.Errorf("CCA result mismatch")
		}
		if top != 2001 {
			if len(a.MultipleServicesCreditControl) != 0 {
				return fmt.Errorf("unexpected failure MSCC")
			}
			return nil
		}
		if len(a.MultipleServicesCreditControl) != 1 {
			return fmt.Errorf("missing service result")
		}
		service := a.MultipleServicesCreditControl[0]
		if len(service.ServiceIdentifier) != 1 || service.ServiceIdentifier[0] != 1 || !service.ResultCode.Present {
			return fmt.Errorf("service association mismatch")
		}
		if result == "2001/4012" {
			if service.ResultCode.Value != 4012 || service.GrantedServiceUnit.Present || service.RefundInformation.Present {
				return fmt.Errorf("service denial mismatch")
			}
		} else {
			if service.ResultCode.Value != 2001 || !service.GrantedServiceUnit.Present || !service.GrantedServiceUnit.Value.CCServiceSpecificUnits.Present || service.GrantedServiceUnit.Value.CCServiceSpecificUnits.Value != 1 {
				return fmt.Errorf("grant mismatch")
			}
			if !service.RefundInformation.Present || !bytes.Equal(service.RefundInformation.Value, []byte{0, 255}) {
				return fmt.Errorf("opaque token changed")
			}
		}
	default:
		return fmt.Errorf("unexpected answer type")
	}
	return nil
}

func run(ctx context.Context, result string, out io.Writer) (r report, err error) {
	if resultName(result) == "" {
		return r, fmt.Errorf("result must be 2001, 2001/4012, 4012, 5012 or 3004 (E)")
	}
	p, err := openPair(ctx)
	if err != nil {
		return r, err
	}
	defer func() {
		err = errors.Join(err, p.close())
		r.Left, r.Right = p.left.node.Resources(), p.right.node.Resources()
		if r.Left != (diameter.ResourceSnapshot{}) || r.Right != (diameter.ResourceSnapshot{}) {
			err = errors.Join(err, fmt.Errorf("resources retained after cleanup"))
		}
	}()
	if err = exchange(ctx, &p.left, &p.right, result); err != nil {
		return r, err
	}
	r.Answer = resultName(result)

	// Finish via DPR/DPA on success. Deferred Close/Wait still covers every error.
	if err = p.left.node.Shutdown(ctx); err != nil {
		return r, err
	}
	if err = p.right.node.Shutdown(ctx); err != nil {
		return r, err
	}
	_, err = fmt.Fprintf(out, "ctf->ocs result=%s\n", r.Answer)
	return r, err
}

func main() {
	compatBinding := flag.String("compat-binding", "combined", "compat only: combined, smsfwro, legacy420 (explicit mutually exclusive binding)")
	compatAction := flag.String("compat-action", "debit", "compat only: debit, balance, refund, initial, term")
	recoveryScenario := flag.String("recovery-scenario", "error", "recovery/compat: error, restart, no-answer; compat also no-answer-import")
	abortSession := flag.String("session", knownAbortSession, "abort only: original CTF session reference")
	flow := flag.String("flow", "iec", "iec, ecur (INITIAL/TERMINATION without UPDATE), abort (ASR/ASA), recovery, or compat (explicit binding and cross-target recovery)")
	termResult := flag.String("term-result", "2001", "ECUR termination:2001,5002,5012 or3004 (E)")
	result := flag.String("result", "2001", "synthetic outcome: IEC/ECUR 2001,2001/4012,4012,5012,3004; abort 2001,5002,5012,3004")
	timeout := flag.Duration("timeout", 10*time.Second, "total run deadline;cleanup has its own finite deadline")
	flag.Parse()
	if !validFlowOptions(*flow, *result, *termResult, *abortSession) || *timeout <= 0 || !validRecoveryOptions(*flow, *recoveryScenario) || !validCompatOptions(*flow, *compatAction) || !validCompatSelection(*flow, *compatBinding, *compatAction, *recoveryScenario) {
		fmt.Fprintln(os.Stderr, "use result 2001, 2001/4012, 4012, 5012 or 3004 and a positive timeout; flow iec/ecur or abort (result 2001/5002/5012/3004, nonempty original session); ECUR term-result 2001/5002/5012/3004")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	var r report
	var err error
	if *flow == "compat" {
		code, _ := strconv.ParseUint(*result, 10, 32)
		err = runBoundCompat(ctx, *compatBinding, *compatAction, *recoveryScenario, uint32(code), os.Stdout)
	} else if *flow == "recovery" {
		r, err = runRecovery(ctx, *recoveryScenario, *result, os.Stdout)
	} else if *flow == "abort" {
		r, err = runAbort(ctx, *abortSession, *result, os.Stdout)
	} else if *flow == "ecur" {
		var er ecurReport
		er, err = runECUR(ctx, *result, *termResult, os.Stdout)
		r.Left, r.Right = er.Left, er.Right
	} else {
		r, err = run(ctx, *result, os.Stdout)
	}
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("closed: left=%+v right=%+v\n", r.Left, r.Right)
}
