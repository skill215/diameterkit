package diameter

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	stdtls "crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	dtcp "github.com/skill215/diameterkit/transport/tcp"
	dtls "github.com/skill215/diameterkit/transport/tls"
	"github.com/skill215/diameterkit/wire"
)

func tcpSockets(t *testing.T) (*net.TCPConn, *net.TCPConn) {
	t.Helper()
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	raw, err := (&net.Dialer{}).DialContext(ctx, "tcp4", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	a := raw.(*net.TCPConn)
	t.Cleanup(func() { _ = a.Close() })
	_ = l.SetDeadline(time.Now().Add(3 * time.Second))
	b, err := l.AcceptTCP()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })
	return a, b
}

var certificateOnce sync.Once
var certificate stdtls.Certificate
var certRoots *x509.CertPool
var certificateErr error

func tlsConfigs(t *testing.T) (*stdtls.Config, *stdtls.Config) {
	t.Helper()
	certificateOnce.Do(func() {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			certificateErr = err
			return
		}
		template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"}, DNSNames: []string{"localhost"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}}
		der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
		if err != nil {
			certificateErr = err
			return
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			certificateErr = err
			return
		}
		certificate = stdtls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: cert}
		certRoots = x509.NewCertPool()
		certRoots.AddCert(cert)
	})
	if certificateErr != nil {
		t.Fatal(certificateErr)
	}
	return &stdtls.Config{ServerName: "localhost", RootCAs: certRoots, Certificates: []stdtls.Certificate{certificate}}, &stdtls.Config{Certificates: []stdtls.Certificate{certificate}, ClientCAs: certRoots, ClientAuth: stdtls.RequireAndVerifyClientCert}
}

// The opposite endpoint uses Go's raw TCP/TLS stream, not our frame adapter.
func networkStream(t *testing.T, encrypted bool) (transport.Conn, net.Conn, *net.TCPConn) {
	t.Helper()
	a, b := tcpSockets(t)
	if !encrypted {
		s, err := dtcp.New(a)
		if err != nil {
			t.Fatal(err)
		}
		return s, b, a
	}
	client, server := tlsConfigs(t)
	peer := stdtls.Server(b, server)
	errs := make(chan error, 1)
	ctx := ctxTest(t)
	go func() { errs <- peer.HandshakeContext(ctx) }()
	s, err := dtls.Client(ctx, a, dtls.Config{TLS: client, HandshakeTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err = <-errs; err != nil {
		t.Fatal(err)
	}
	if !s.ConnectionState().HandshakeComplete || len(s.ConnectionState().VerifiedChains) == 0 {
		t.Fatal("unverified TLS")
	}
	return s, peer, a
}
func frame(size int, tag byte) []byte {
	b := bytes.Repeat([]byte{tag}, size)
	binary.BigEndian.PutUint32(b[:4], uint32(size)|1<<24)
	return b
}
func rawSend(t *testing.T, c net.Conn, b []byte) {
	t.Helper()
	_ = c.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := c.Write(b); err != nil {
		t.Fatal(err)
	}
}
func readAdapter(ctx context.Context, s transport.Conn) ([]byte, error) {
	p, err := s.ReadPrefix(ctx, time.Second)
	if err != nil {
		return nil, err
	}
	size := int(binary.BigEndian.Uint32(p[:4]) & 0xffffff)
	b := make([]byte, size)
	err = s.ReadFrameInto(ctx, b)
	return b, err
}
func eachNetwork(t *testing.T, f func(*testing.T, bool)) {
	t.Helper()
	t.Run("TCP", func(t *testing.T) { f(t, false) })
	t.Run("TLS", func(t *testing.T) { f(t, true) })
}

func TestNetworkFrameBoundariesAndDiscard(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		s, peer, _ := networkStream(t, tls)
		ctx := ctxTest(t)
		first, drop, last := frame(47, 1), frame(8192, 2), frame(20, 3)
		rawSend(t, peer, first[:3])
		rawSend(t, peer, first[3:17])
		rawSend(t, peer, append(append(first[17:], drop...), last...))
		p, err := s.ReadPrefix(ctx, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.ReadPrefix(ctx, time.Second); err == nil {
			t.Fatal("second prefix accepted")
		}
		if err = s.ReadFrameInto(ctx, make([]byte, 21)); err == nil {
			t.Fatal("wrong dst accepted")
		}
		b := make([]byte, len(first))
		if err = s.ReadFrameInto(ctx, b); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, first) || !bytes.Equal(p[:], first[:20]) {
			t.Fatal("first frame corrupted")
		}
		if _, err = s.ReadPrefix(ctx, time.Second); err != nil {
			t.Fatal(err)
		}
		if err = s.DiscardFrame(ctx, len(drop)); err != nil {
			t.Fatal(err)
		}
		got, err := readAdapter(ctx, s)
		if err != nil || !bytes.Equal(got, last) {
			t.Fatalf("boundary lost %v", err)
		}
		if s.Info().BufferBytes != dtcp.BufferBytes {
			t.Fatal("reserve")
		}
	})
}
func TestNetworkIdleAndAbsoluteFrameDeadline(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		t.Run("idle_is_not_frame_timeout", func(t *testing.T) {
			s, peer, _ := networkStream(t, tls)
			ctx := ctxTest(t)
			ch := make(chan error, 1)
			go func() { _, e := s.ReadPrefix(ctx, 60*time.Millisecond); ch <- e }()
			select {
			case err := <-ch:
				t.Fatalf("idle ended %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			rawSend(t, peer, frame(20, 1))
			if err := <-ch; err != nil {
				t.Fatal(err)
			}
			if err := s.ReadFrameInto(ctx, make([]byte, 20)); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("partial_prefix", func(t *testing.T) {
			s, peer, _ := networkStream(t, tls)
			rawSend(t, peer, []byte{1})
			_, err := s.ReadPrefix(ctxTest(t), 50*time.Millisecond)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("%v", err)
			}
		})
		t.Run("body_does_not_reset_deadline", func(t *testing.T) {
			s, peer, _ := networkStream(t, tls)
			rawSend(t, peer, frame(32, 1)[:20])
			ctx := ctxTest(t)
			if _, err := s.ReadPrefix(ctx, 40*time.Millisecond); err != nil {
				t.Fatal(err)
			}
			time.Sleep(60 * time.Millisecond)
			if err := s.ReadFrameInto(ctx, make([]byte, 32)); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("%v", err)
			}
		})
		t.Run("idle_context_deadline", func(t *testing.T) {
			s, _, _ := networkStream(t, tls)
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
			defer cancel()
			if _, err := s.ReadPrefix(ctx, time.Second); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("%v", err)
			}
		})
	})
}
func TestNetworkCancelAndCloseUnblockReads(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, body := range []bool{false, true} {
			for _, cancelIO := range []bool{false, true} {
				s, peer, _ := networkStream(t, tls)
				ctx, cancel := context.WithCancel(context.Background())
				if body {
					rawSend(t, peer, frame(100, 1)[:20])
					if _, err := s.ReadPrefix(ctx, time.Second); err != nil {
						t.Fatal(err)
					}
				}
				started := make(chan struct{})
				ch := make(chan error, 1)
				go func() {
					close(started)
					if body {
						ch <- s.ReadFrameInto(ctx, make([]byte, 100))
					} else {
						_, err := s.ReadPrefix(ctx, time.Second)
						ch <- err
					}
				}()
				<-started
				if cancelIO {
					cancel()
				} else {
					var wg sync.WaitGroup
					for i := 0; i < 4; i++ {
						wg.Add(1)
						go func() { defer wg.Done(); _ = s.Close() }()
					}
					wg.Wait()
				}
				select {
				case err := <-ch:
					if err == nil || cancelIO && !errors.Is(err, context.Canceled) {
						t.Fatalf("cancel=%v body=%v err=%v", cancelIO, body, err)
					}
				case <-time.After(time.Second):
					t.Fatal("blocked read")
				}
				cancel()
				_ = s.Close()
			}
		}
	})
}
func TestNetworkCompletedContextCannotPoisonNextOperation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		s, peer, _ := networkStream(t, tls)
		b := frame(20, 3)
		for i := 0; i < 30; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			rawSend(t, peer, b)
			got, err := readAdapter(ctx, s)
			if err != nil || !bytes.Equal(got, b) {
				t.Fatalf("read %d: %v", i, err)
			}
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			r := s.WriteFrame(ctx, b)
			cancel()
			if r.Err != nil || r.Bytes != len(b) {
				t.Fatalf("write %d: %+v", i, r)
			}
			_ = peer.SetReadDeadline(time.Now().Add(time.Second))
			if _, err = io.ReadFull(peer, make([]byte, 20)); err != nil {
				t.Fatal(err)
			}
		}
	})
}
func TestNetworkWriteDeadlineAndCloseEvidence(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, action := range []string{"deadline", "cancel", "close"} {
			t.Run(action, func(t *testing.T) {
				s, peer, raw := networkStream(t, tls)
				_ = raw.SetWriteBuffer(1024)
				// The peer never reads application bytes. Eight MiB exceeds TCP buffering.
				_ = peer
				b := frame(8<<20, 9)
				ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
				defer cancel()
				if action != "deadline" {
					ctx, cancel = context.WithCancel(context.Background())
					defer cancel()
				}
				ch := make(chan transport.WriteResult, 1)
				go func() { ch <- s.WriteFrame(ctx, b) }()
				if action != "deadline" {
					time.Sleep(15 * time.Millisecond)
					if action == "cancel" {
						cancel()
					} else {
						_ = s.Close()
					}
				}
				select {
				case r := <-ch:
					if r.Err == nil || r.Bytes < 0 || r.Bytes >= len(b) || r.Uncertain != tls {
						t.Fatalf("bad write evidence %+v", r)
					}
					if action == "cancel" && !errors.Is(r.Err, context.Canceled) {
						t.Fatal(r.Err)
					}
					if action == "deadline" && !errors.Is(r.Err, context.DeadlineExceeded) {
						t.Fatal(r.Err)
					}
					if r2 := s.WriteFrame(ctxTest(t), frame(20, 1)); r2.Err == nil {
						t.Fatal("continued failed stream")
					}
				case <-time.After(time.Second):
					t.Fatal("blocked writer")
				}
			})
		}
	})
}

func readWire(ctx context.Context, s transport.Conn, co *codec.Codec) (pdu.Message, error) {
	b, err := readAdapter(ctx, s)
	if err != nil {
		return pdu.Message{}, err
	}
	return co.Decode(b)
}
func writeWire(ctx context.Context, s transport.Conn, co *codec.Codec, m pdu.Message) error {
	b, err := co.Append(nil, m)
	if err != nil {
		return err
	}
	r := s.WriteFrame(ctx, b)
	if r.Err != nil {
		return r.Err
	}
	if r.Bytes != len(b) {
		return io.ErrShortWrite
	}
	return nil
}

// networkAdapters exercises both public TLS constructors (the framing tests use
// an independent standard-library stream endpoint for fragmentation control).
func networkAdapters(t *testing.T, encrypted bool) (transport.Conn, transport.Conn) {
	t.Helper()
	a, b := tcpSockets(t)
	if !encrypted {
		x, e := dtcp.New(a)
		if e != nil {
			t.Fatal(e)
		}
		y, e := dtcp.New(b)
		if e != nil {
			t.Fatal(e)
		}
		return x, y
	}
	client, server := tlsConfigs(t)
	type result struct {
		c   *dtls.Conn
		err error
	}
	ch := make(chan result, 1)
	ctx := ctxTest(t)
	go func() {
		s, err := dtls.Server(ctx, b, dtls.Config{TLS: server, HandshakeTimeout: time.Second})
		ch <- result{s, err}
	}()
	x, err := dtls.Client(ctx, a, dtls.Config{TLS: client, HandshakeTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	y := <-ch
	if y.err != nil {
		t.Fatal(y.err)
	}
	if len(y.c.ConnectionState().VerifiedChains) == 0 {
		t.Fatal("client not authenticated")
	}
	return x, y.c
}
func TestNetworkNodeHandshakeRequestAndShutdown(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newTestNode(t)
		a, b := networkAdapters(t, tls)
		ctx := ctxTest(t)
		errs := make(chan error, 1)
		go func() {
			req, err := readWire(ctx, b, co)
			if err != nil {
				errs <- err
				return
			}
			ans, err := co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "network-dra", AuthApplicationID: n.cfg.Capabilities.AuthApplicationID, VendorSpecificApplicationID: n.cfg.Capabilities.VendorSpecificApplicationID}, draOrigin)
			if err == nil {
				err = writeWire(ctx, b, co, ans)
			}
			if err == nil {
				req, err = readWire(ctx, b, co)
			}
			if err == nil {
				ans, err = answerFor(co, req)
			}
			if err == nil {
				err = writeWire(ctx, b, co, ans)
			}
			if err == nil {
				req, err = readWire(ctx, b, co)
			}
			if err == nil && req.Header.CommandCode != 282 {
				err = errors.New("expected DPR")
			}
			if err == nil {
				ans, err = co.CompleteAnswer(req, &base.DPA{ResultCode: 2001}, draOrigin)
			}
			if err == nil {
				err = writeWire(ctx, b, co, ans)
			}
			errs <- err
		}()
		p, err := n.Connect(ctx, "dra", a)
		if err != nil {
			t.Fatal(err)
		}
		r, err := p.Request(ctx, requestBody(), RequestOptions{})
		if err != nil || r.Outcome != AnswerReceived {
			t.Fatalf("result %+v %v", r, err)
		}
		if err = n.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		if err = n.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if err = <-errs; err != nil {
			t.Fatal(err)
		}
		if p.Info().LastDPAResult != 2001 {
			t.Fatal("DPA not recorded")
		}
	})
}

// Keep wire import as a real invariant: transport framing and codec prefix sizes agree.
func TestNetworkInvalidLengthCloses(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		s, peer, _ := networkStream(t, tls)
		b := make([]byte, 20)
		binary.BigEndian.PutUint32(b[:4], 1<<24|19)
		rawSend(t, peer, b)
		if _, err := s.ReadPrefix(ctxTest(t), time.Second); err == nil {
			t.Fatal("invalid size accepted")
		}
		if _, _, err := wire.Header(b, wire.DefaultLimits()); err == nil {
			t.Fatal("codec accepted invalid size")
		}
	})
}

func networkConnectNode(t *testing.T, n *Node, co *codec.Codec, encrypted bool) (*Peer, transport.Conn) {
	t.Helper()
	a, b := networkAdapters(t, encrypted)
	ctx := ctxTest(t)
	ch := make(chan error, 1)
	go func() {
		req, err := readWire(ctx, b, co)
		if err == nil {
			var ans pdu.Message
			ans, err = co.CompleteAnswer(req, &base.CEA{ResultCode: 2001, HostIPAddress: n.cfg.Capabilities.HostIPAddress, ProductName: "network-dra", AuthApplicationID: n.cfg.Capabilities.AuthApplicationID, VendorSpecificApplicationID: n.cfg.Capabilities.VendorSpecificApplicationID}, draOrigin)
			if err == nil {
				err = writeWire(ctx, b, co, ans)
			}
		}
		ch <- err
	}()
	p, err := n.Connect(ctx, "dra", a)
	if err != nil {
		t.Fatal(err)
	}
	if err = <-ch; err != nil {
		t.Fatal(err)
	}
	return p, b
}
func TestNetworkCancelLateAnswerAndCloseRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newTestNode(t)
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 2)
		ctx := ctxTest(t)
		id, err := p.Submit(ctx, requestBody(), RequestOptions{}, port)
		if err != nil {
			t.Fatal(err)
		}
		req, err := readWire(ctx, remote, co)
		if err != nil {
			t.Fatal(err)
		}
		if err = n.Cancel(id); err != nil {
			t.Fatal(err)
		}
		result, err := port.Receive(ctx)
		if err != nil || result.Result.Outcome != WrittenNoAnswer {
			t.Fatal(result, err)
		}
		ans, err := answerFor(co, req)
		if err != nil {
			t.Fatal(err)
		}
		if err = writeWire(ctx, remote, co, ans); err != nil {
			t.Fatal(err)
		}
		// An unrelated late answer must leave the live connection usable.
		id, err = p.Submit(ctx, requestBody(), RequestOptions{}, port)
		if err != nil {
			t.Fatal(err)
		}
		req, err = readWire(ctx, remote, co)
		if err != nil {
			t.Fatal(err)
		}
		ans, err = answerFor(co, req)
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() { defer wg.Done(); _ = n.Cancel(id) }()
		go func() { defer wg.Done(); _ = writeWire(ctx, remote, co, ans) }()
		go func() { defer wg.Done(); _ = n.Close() }()
		wg.Wait()
		result, err = port.Receive(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if result.RequestID != id || !result.Result.Delivery.MayHaveReachedPeer {
			t.Fatal(result)
		}
		if err = n.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if n.Resources().Outstanding != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestNetworkCERHandshakeInterrupted(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, passive := range []bool{false, true} {
			n, _ := newTestNode(t)
			a, b := networkAdapters(t, tls)
			_ = b
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			var err error
			if passive {
				_, err = n.Accept(ctx, "dra", a)
			} else {
				_, err = n.Connect(ctx, "dra", a)
			}
			cancel()
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("passive=%v err=%v", passive, err)
			}
			p, _ := n.Peer("dra")
			if p.Info().Ready || p.Info().Candidates != 0 {
				t.Fatal(p.Info())
			}
			_ = n.Close()
			if err = n.Wait(ctxTest(t)); err != nil {
				t.Fatal(err)
			}
		}
	})
}
func TestNetworkElectionAndCrossedShutdown(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		low, high := testConfig(t), testConfig(t)
		low.Origin.Host = "alpha.example"
		high.Origin.Host = "zulu.example"
		low.Peers[0].Origin = high.Origin
		high.Peers[0].Origin = low.Origin
		a, err := NewNode(low)
		if err != nil {
			t.Fatal(err)
		}
		b, err := NewNode(high)
		if err != nil {
			t.Fatal(err)
		}
		defer a.Close()
		defer b.Close()
		ctx := ctxTest(t)
		ao, bi := networkAdapters(t, tls)
		bo, ai := networkAdapters(t, tls)
		ch := make(chan error, 4)
		go func() { _, e := a.Connect(ctx, "dra", ao); ch <- e }()
		go func() { _, e := b.Connect(ctx, "dra", bo); ch <- e }()
		ap, _ := a.Peer("dra")
		bp, _ := b.Peer("dra")
		waitFor(t, func() bool { return ap.Info().Candidates == 1 && bp.Info().Candidates == 1 })
		go func() { _, e := a.Accept(ctx, "dra", ai); ch <- e }()
		go func() { _, e := b.Accept(ctx, "dra", bi); ch <- e }()
		wins := 0
		for range 4 {
			if <-ch == nil {
				wins++
			}
		}
		if wins != 2 || ap.Info().Protocol != IOpen || bp.Info().Protocol != ROpen {
			t.Fatal(wins, ap.Info(), bp.Info())
		}
		go func() { ch <- a.Shutdown(ctx) }()
		go func() { ch <- b.Shutdown(ctx) }()
		for range 2 {
			if e := <-ch; e != nil {
				t.Fatal(e)
			}
		}
		if err = a.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if err = b.Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if ap.Info().Ready || bp.Info().Ready || ap.Info().Candidates != 0 || bp.Info().Candidates != 0 {
			t.Fatal("not quiesced")
		}
	})
}

func TestNetworkFirstByteCancelRace(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for range 20 {
			s, peer, _ := networkStream(t, tls)
			ctx, cancel := context.WithCancel(context.Background())
			ch := make(chan error, 1)
			go func() { _, e := s.ReadPrefix(ctx, 10*time.Second); ch <- e }()
			rawSend(t, peer, []byte{1})
			cancel()
			select {
			case err := <-ch:
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("first-byte update lost cancellation")
			}
			_ = s.Close()
		}
	})
}
func TestNetworkPrefixContextDeadlineCarriesIntoBody(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		s, peer, _ := networkStream(t, tls)
		rawSend(t, peer, frame(32, 1))
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
		defer cancel()
		if _, err := s.ReadPrefix(ctx, time.Second); err != nil {
			t.Fatal(err)
		}
		<-ctx.Done()
		// A fresh body context cannot extend the original frame's deadline, even
		// with all body bytes already buffered inside TLS.
		if err := s.ReadFrameInto(ctxTest(t), make([]byte, 32)); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})
}

func TestNetworkAdapterReserveRejectedBeforeCER(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := testConfig(t)
		cfg.Resources.AdapterBytesPerConnection = dtcp.BufferBytes - 1
		n, err := NewNode(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer n.Close()
		a, b := networkAdapters(t, tls)
		if _, err = n.Connect(ctxTest(t), "dra", a); !errors.Is(err, ErrBackpressure) {
			t.Fatal(err)
		}
		if _, err = b.ReadPrefix(ctxTest(t), time.Second); err == nil {
			t.Fatal("rejected socket left open")
		}
		p, _ := n.Peer("dra")
		if p.Info().Candidates != 0 {
			t.Fatal(p.Info())
		}
		_ = n.Close()
		if err = n.Wait(ctxTest(t)); err != nil {
			t.Fatal(err)
		}
	})
}
