// Command sgd demonstrates bounded SGd/Gdd MO and MT over loopback TCP.
// The access endpoint and a combined IWMSC/GMSC endpoint use only public APIs.
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
	"time"

	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/app/sgd"
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
	Answers     [2]string
	Left, Right diameter.ResourceSnapshot
}

// Serve owns the listener, Dial owns the connected transport, and Node owns
// their lifetime. The two Nodes have separate origins, limits and completions.
func openPair(ctx context.Context, iface string) (*pair, error) {
	if iface != "sgd" && iface != "gdd" {
		return nil, fmt.Errorf("interface must be sgd or gdd")
	}
	accessHost := "mme.example"
	if iface == "gdd" {
		accessHost = "sgsn.example"
	}
	origins := [2]pdu.Origin{{Host: accessHost, Realm: "example"}, {Host: "sms.example", Realm: "example"}}
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
	if err != nil {
		return p, err
	}
	// The accepting endpoint can finish its CEA write barrier after Dial returns.
	// This bidirectional example must wait before immediately sending MT traffic.
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

func exchange(ctx context.Context, from, to *endpoint, mt, gdd bool, result uint32) error {
	request, err := newRequest(from.node, mt, gdd)
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
	var session string
	var payload []byte
	switch r := in.Message().Body.(type) {
	case *sgd.OFR:
		if mt {
			return fmt.Errorf("unexpected MO request")
		}
		session = r.SessionID
		if session != request.(*sgd.OFR).SessionID {
			return fmt.Errorf("MO session mismatch")
		}
		payload = r.SMRPUI
	case *sgd.TFR:
		if !mt || r.SGSNNumber.Present != gdd || r.MMENumberForMTSMS.Present == gdd {
			return fmt.Errorf("MT target mismatch")
		}
		session = r.SessionID
		if session != request.(*sgd.TFR).SessionID {
			return fmt.Errorf("MT session mismatch")
		}
		payload = r.SMRPUI
	default:
		return fmt.Errorf("unexpected request type")
	}
	if !bytes.Equal(payload, requestPayload()) {
		return fmt.Errorf("opaque request payload changed")
	}
	// Respond restores the original Session/Proxy/header and local Origin.
	answer, err := newAnswer(mt, result)
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
	return checkAnswer(completion.Result.Answer.Body, mt, result, session)
}

func checkAnswer(body pdu.Body, mt bool, result uint32, session string) error {
	if a, ok := body.(*pdu.ErrorAnswer); ok {
		if result != 3004 || a.ResultCode != result || !a.SessionID.Present || a.SessionID.Value != session {
			return fmt.Errorf("E correlation/result mismatch")
		}
		return nil
	}
	var gotSession string
	var rc avp.Optional[uint32]
	var exp avp.Optional[sgd.ExperimentalResult]
	var cause avp.Optional[sgd.SMDeliveryFailureCause]
	var payload avp.Optional[[]byte]
	switch a := body.(type) {
	case *sgd.OFA:
		if mt {
			return fmt.Errorf("unexpected OFA")
		}
		gotSession, rc, exp, cause, payload = a.SessionID, a.ResultCode, a.ExperimentalResult, a.SMDeliveryFailureCause, a.SMRPUI
	case *sgd.TFA:
		if !mt {
			return fmt.Errorf("unexpected TFA")
		}
		gotSession, rc, exp, cause, payload = a.SessionID, a.ResultCode, a.ExperimentalResult, a.SMDeliveryFailureCause, a.SMRPUI
	default:
		return fmt.Errorf("unexpected answer type")
	}
	if gotSession != session || result == 3004 {
		return fmt.Errorf("answer correlation mismatch")
	}
	if result == 5555 {
		expected := uint32(3)
		if mt {
			expected = 0
		}
		if rc.Present || !exp.Present || exp.Value.VendorID != 10415 || exp.Value.ExperimentalResultCode != 5555 || !cause.Present || cause.Value.SMEnumeratedDeliveryFailureCause != expected {
			return fmt.Errorf("experimental result/cause mismatch")
		}
	} else if !rc.Present || rc.Value != result || exp.Present || cause.Present {
		return fmt.Errorf("base result mismatch")
	}
	if result == 2001 {
		if !payload.Present || !bytes.Equal(payload.Value, answerPayload()) {
			return fmt.Errorf("opaque answer payload changed")
		}
	} else if payload.Present {
		return fmt.Errorf("unexpected synthetic failure payload")
	}
	return nil
}

func run(ctx context.Context, iface string, result uint32, out io.Writer) (r report, err error) {
	if resultName(result) == "" {
		return r, fmt.Errorf("result must be 2001, 5012, 5555 (experimental10415) or 3004 (E)")
	}
	p, err := openPair(ctx, iface)
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
		if err = exchange(ctx, direction[0], direction[1], i == 1, iface == "gdd", result); err != nil {
			return r, err
		}
		r.Answers[i] = resultName(result)
	}
	// Finish via DPR/DPA on success. Deferred Close/Wait still covers every error.
	if err = p.left.node.Shutdown(ctx); err != nil {
		return r, err
	}
	if err = p.right.node.Shutdown(ctx); err != nil {
		return r, err
	}
	_, err = fmt.Fprintf(out, "%s MO result=%s; MT result=%s\n", iface, r.Answers[0], r.Answers[1])
	return r, err
}

func main() {
	iface := flag.String("interface", "sgd", "sgd (MME) or gdd (SGSN); shared Diameter Application-ID")
	result := flag.Uint("result", 2001, "synthetic answer: 2001, 5012, 5555 (experimental10415) or 3004 (E)")
	timeout := flag.Duration("timeout", 10*time.Second, "total run deadline; cleanup has its own finite deadline")
	flag.Parse()
	if *result != 2001 && *result != 5012 && *result != 5555 && *result != 3004 || *timeout <= 0 || (*iface != "sgd" && *iface != "gdd") {
		fmt.Fprintln(os.Stderr, "use interface sgd/gdd, result 2001/5012/5555/3004 and a positive timeout")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	r, err := run(ctx, *iface, uint32(*result), os.Stdout)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("closed: left=%+v right=%+v\n", r.Left, r.Right)
}
