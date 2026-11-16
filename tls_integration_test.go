package diameter

import (
	"context"
	stdtls "crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"testing"
	"time"

	dtcp "github.com/skill215/diameterkit/transport/tcp"
	dtls "github.com/skill215/diameterkit/transport/tls"
)

func TestTLSVersionsAndVerification(t *testing.T) {
	for _, version := range []uint16{stdtls.VersionTLS12, stdtls.VersionTLS13} {
		for _, mode := range []string{"valid_mTLS", "untrusted_server", "wrong_name", "missing_client_cert"} {
			t.Run(stdtls.VersionName(version)+"/"+mode, func(t *testing.T) {
				a, b := tcpSockets(t)
				client, server := tlsConfigs(t)
				client.MinVersion = version
				client.MaxVersion = version
				server.MinVersion = version
				server.MaxVersion = version
				switch mode {
				case "untrusted_server":
					client.RootCAs = x509.NewCertPool()
				case "wrong_name":
					client.ServerName = "wrong.example"
				case "missing_client_cert":
					client.Certificates = nil
				}
				type result struct {
					c *dtls.Conn
					e error
				}
				ch := make(chan result, 1)
				ctx := ctxTest(t)
				go func() {
					c, e := dtls.Server(ctx, b, dtls.Config{TLS: server, HandshakeTimeout: time.Second})
					ch <- result{c, e}
				}()
				c, err := dtls.Client(ctx, a, dtls.Config{TLS: client, HandshakeTimeout: time.Second})
				s := <-ch
				if mode == "valid_mTLS" {
					if err != nil || s.e != nil {
						t.Fatalf("client %v server %v", err, s.e)
					}
					if c.ConnectionState().Version != version || s.c.ConnectionState().Version != version {
						t.Fatal("version mismatch")
					}
					if len(c.ConnectionState().VerifiedChains) == 0 || len(s.c.ConnectionState().VerifiedChains) == 0 {
						t.Fatal("unverified mutual identity")
					}
					_ = c.Close()
					_ = s.c.Close()
				} else if mode == "missing_client_cert" {
					// TLS 1.3 clients may finish before receiving the server's certificate alert.
					if s.e == nil {
						t.Fatal("server accepted unauthenticated client")
					}
					if err == nil {
						if _, e := c.ReadPrefix(ctx, time.Second); e == nil {
							t.Fatal("missing alert")
						}
						_ = c.Close()
					}
				} else if err == nil {
					t.Fatal("client accepted invalid server")
				}
			})
		}
	}
}
func TestTLSHandshakeCancellationDeadlineAndEOF(t *testing.T) {
	for _, clientSide := range []bool{false, true} {
		for _, action := range []string{"deadline", "cancel", "peer_close"} {
			name := "server/" + action
			if clientSide {
				name = "client/" + action
			}
			t.Run(name, func(t *testing.T) {
				a, b := tcpSockets(t)
				client, server := tlsConfigs(t)
				cfg := dtls.Config{TLS: server, HandshakeTimeout: 60 * time.Millisecond}
				if clientSide {
					cfg.TLS = client
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				ch := make(chan error, 1)
				go func() {
					var e error
					if clientSide {
						_, e = dtls.Client(ctx, a, cfg)
					} else {
						_, e = dtls.Server(ctx, a, cfg)
					}
					ch <- e
				}()
				if action == "cancel" {
					cancel()
				}
				if action == "peer_close" {
					_ = b.Close()
				}
				select {
				case err := <-ch:
					if err == nil {
						t.Fatal("handshake unexpectedly succeeded")
					}
					if action == "deadline" && !errors.Is(err, context.DeadlineExceeded) {
						t.Fatal(err)
					}
					if action == "cancel" && !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
				case <-time.After(time.Second):
					t.Fatal("handshake did not unblock")
				}
				if _, err := a.Write([]byte{1}); err == nil {
					t.Fatal("failed handshake retained socket")
				}
			})
		}
	}
}
func TestTLSConfigurationFailureClosesOwnedSocket(t *testing.T) {
	for _, cfg := range []dtls.Config{{}, {TLS: &stdtls.Config{}, HandshakeTimeout: 0}, {TLS: &stdtls.Config{MinVersion: stdtls.VersionTLS10}, HandshakeTimeout: time.Second}} {
		a, _ := tcpSockets(t)
		if _, err := dtls.Client(ctxTest(t), a, cfg); err == nil {
			t.Fatal("invalid config accepted")
		}
		if _, err := a.Write([]byte{1}); err == nil {
			t.Fatal("socket not closed")
		}
	}
	if _, err := dtcp.New(nil); err == nil {
		t.Fatal("nil TCP")
	}
	if _, err := dtls.Client(ctxTest(t), nil, dtls.Config{}); err == nil {
		t.Fatal("nil TLS")
	}
}
func TestNetworkAlreadyCanceledDoesNotWrite(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		s, peer, _ := networkStream(t, tls)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r := s.WriteFrame(ctx, frame(20, 1))
		if !errors.Is(r.Err, context.Canceled) || r.Bytes != 0 || r.Uncertain {
			t.Fatalf("%+v", r)
		}
		// No I/O occurred; a fresh context can use the connection.
		r = s.WriteFrame(ctxTest(t), frame(20, 2))
		if r.Err != nil || r.Bytes != 20 {
			t.Fatal(r)
		}
		_ = peer.SetReadDeadline(time.Now().Add(time.Second))
		b := make([]byte, 20)
		if _, err := io.ReadFull(peer, b); err != nil {
			t.Fatal(err)
		}
		if b[19] != 2 {
			t.Fatal("canceled payload leaked")
		}
	})
}
