// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Example or tooling code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Command scap1 runs two bounded, nonpersistent SCAP1 endpoints on loopback TCP.
// Each endpoint sends one event and answers one event using only public APIs.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/compat/scap1"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport/tcp"
)

type endpoint struct {
	node *diameter.Node
	peer *diameter.Peer
	port *diameter.CompletionPort
}
type pair struct {
	left, right endpoint
	cancelServe context.CancelFunc
	serveDone   chan error
}
type report struct {
	Answers     [2]uint32
	Left, Right diameter.ResourceSnapshot
}

// Serve owns the listener, Dial owns the connected transport, and Node owns
// their lifetime. The two Nodes have separate origins, limits and completions.
func openPair(ctx context.Context, leftMode, rightMode diameter.SCAP1Advertisement) (p *pair, err error) {
	p = &pair{}
	defer func() {
		if err != nil {
			err = errors.Join(err, p.close())
			p = nil
		}
	}()
	origins := []pdu.Origin{{Host: "left.example", Realm: "example"}, {Host: "right.example", Realm: "example"}}
	for i, ep := range []*endpoint{&p.left, &p.right} {
		mode := leftMode
		if i == 1 {
			mode = rightMode
		}
		cfg, e := endpointConfig(origins[i], origins[1-i], mode)
		if e != nil {
			return p, e
		}
		ep.node, e = diameter.NewNode(cfg)
		if e != nil {
			return p, e
		}
		ep.peer, _ = ep.node.Peer(remotePeer) // Both configs explicitly contain it.
		ep.port, e = ep.node.NewCompletionPort(diameter.CompletionLimits{MaxOutstanding: 1, MaxOutstandingBytes: 32 << 20})
		if e != nil {
			return p, e
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
	return p, err
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

func exchange(ctx context.Context, from, to *endpoint, result uint32) error {
	request, err := newEvent(from.node)
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
	acr, ok := in.Message().Body.(*scap1.ACR)
	if !ok || acr.SessionID != request.SessionID {
		return fmt.Errorf("unexpected inbound request")
	}
	// The application selects a synthetic outcome. Respond restores the actual
	// arrival's Session/Proxy/header and the configured local Origin.
	var answer pdu.Body = &scap1.ACA{ResultCode: result, AccountingRecordType: acr.AccountingRecordType, AccountingRecordNumber: uint32(acr.AccountingRecordNumber)}
	if result == 3004 {
		answer = &pdu.ErrorAnswer{ResultCode: result}
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
	// A protocol answer is not automatically a successful charging outcome.
	switch a := completion.Result.Answer.Body.(type) {
	case *scap1.ACA:
		if result == 3004 || a.ResultCode != result || a.SessionID != request.SessionID || a.AccountingRecordNumber != 0 {
			return fmt.Errorf("ACA correlation/result mismatch")
		}
	case *pdu.ErrorAnswer:
		if result != 3004 || a.ResultCode != result {
			return fmt.Errorf("E answer mismatch")
		}
	default:
		return fmt.Errorf("unexpected answer type")
	}
	return nil
}

func run(ctx context.Context, mode diameter.SCAP1Advertisement, result uint32, out io.Writer) (r report, err error) {
	if result != 2001 && result != 5012 && result != 3004 {
		return r, fmt.Errorf("result must be 2001, 5012 or 3004")
	}
	p, err := openPair(ctx, mode, mode)
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
	for i, direction := range [][2]*endpoint{{&p.left, &p.right}, {&p.right, &p.left}} {
		if err = exchange(ctx, direction[0], direction[1], result); err != nil {
			return r, err
		}
		r.Answers[i] = result
	}
	// Finish via DPR/DPA on success. Deferred Close/Wait still covers every error.
	if err = p.left.node.Shutdown(ctx); err != nil {
		return r, err
	}
	if err = p.right.node.Shutdown(ctx); err != nil {
		return r, err
	}
	_, err = fmt.Fprintf(out, "left->right result=%d; right->left result=%d\n", r.Answers[0], r.Answers[1])
	return r, err
}

func main() {
	legacy := flag.Bool("legacy-advertisement", false, "explicitly use the reviewed SCAP1 dual advertisement on both peers")
	result := flag.Uint("result", 2001, "synthetic answer: 2001, 5012 or 3004 (E)")
	timeout := flag.Duration("timeout", 10*time.Second, "total run deadline; cleanup has its own finite deadline")
	flag.Parse()
	if *result != 2001 && *result != 5012 && *result != 3004 || *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "use result 2001/5012/3004 and a positive timeout")
		os.Exit(1)
	}
	mode := diameter.SCAP1StandardAdvertisement
	if *legacy {
		mode = diameter.SCAP1LegacyDualAdvertisement
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	r, err := run(ctx, mode, uint32(*result), os.Stdout)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("closed: left=%+v right=%+v\n", r.Left, r.Right)
}
