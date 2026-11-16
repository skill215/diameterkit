// Command s6c runs two bounded, nonpersistent S6c GMSC/HSS endpoints on loopback TCP.
// GMSC sends one SRR and HSS answers it using only public APIs.
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
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
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
	Answer      string
	Left, Right diameter.ResourceSnapshot
}

// Serve owns the listener, Dial owns the connected transport, and Node owns
// their lifetime. The two Nodes have separate origins, limits and completions.
func openPair(ctx context.Context) (*pair, error) {
	origins := [2]pdu.Origin{{Host: "gmsc.example", Realm: "example"}, {Host: "hss.example", Realm: "example"}}
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
	request, err := newRouting(from.node)
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
	srr, ok := in.Message().Body.(*s6c.SRR)
	if !ok || srr.SessionID != request.SessionID {
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
	// Base, experimental and E results have distinct meanings and namespaces.
	switch a := completion.Result.Answer.Body.(type) {
	case *s6c.SRA:
		if result == 3004 || a.SessionID != request.SessionID {
			return fmt.Errorf("SRA correlation mismatch")
		}
		if result == 5001 {
			if a.ResultCode.Present || !a.ExperimentalResult.Present || a.ExperimentalResult.Value.VendorID != 10415 || a.ExperimentalResult.Value.ExperimentalResultCode != 5001 {
				return fmt.Errorf("experimental result mismatch")
			}
		} else if !a.ResultCode.Present || a.ResultCode.Value != result || a.ExperimentalResult.Present {
			return fmt.Errorf("base result mismatch")
		}
		if result == 2001 {
			if !a.UserName.Present || a.UserName.Value != subscriberIMSI || !a.ServingNode.Present {
				return fmt.Errorf("missing routing identity/node")
			}
			node := a.ServingNode.Value
			if !node.MMEName.Present || node.MMEName.Value != "mme.example" || !node.MMERealm.Present || node.MMERealm.Value != "example" || !node.MMENumber.Present {
				return fmt.Errorf("unexpected MME routing data")
			}
			number, e := avp.DecodeTBCD(node.MMENumber.Value)
			if e != nil {
				return e
			}
			if number != "12025550199" {
				return fmt.Errorf("unexpected MME number")
			}
		} else if a.UserName.Present || a.ServingNode.Present {
			return fmt.Errorf("failure contains routing success data")
		}
	case *pdu.ErrorAnswer:
		if result != 3004 || a.ResultCode != result || !a.SessionID.Present || a.SessionID.Value != request.SessionID {
			return fmt.Errorf("E answer mismatch")
		}
	default:
		return fmt.Errorf("unexpected answer type")
	}

	return nil
}

func run(ctx context.Context, result uint32, out io.Writer) (r report, err error) {
	if resultName(result) == "" {
		return r, fmt.Errorf("result must be 2001, 5012, 5001 (experimental10415) or 3004 (E)")
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
	_, err = fmt.Fprintf(out, "gmsc->hss result=%s\n", r.Answer)
	return r, err
}

func main() {
	result := flag.Uint("result", 2001, "synthetic answer: 2001, 5012, 5001 (experimental10415) or 3004 (E)")
	timeout := flag.Duration("timeout", 10*time.Second, "total run deadline; cleanup has its own finite deadline")
	flag.Parse()
	if *result != 2001 && *result != 5012 && *result != 5001 && *result != 3004 || *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "use result 2001/5012/5001/3004 and a positive timeout")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	r, err := run(ctx, uint32(*result), os.Stdout)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("closed: left=%+v right=%+v\n", r.Left, r.Right)
}
