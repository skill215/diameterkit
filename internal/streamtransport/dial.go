package streamtransport

import (
	"context"
	"errors"
	"net"
	"strconv"
)

// ValidateEndpoint bounds the fixed address; it does not resolve or discover peers.
func ValidateEndpoint(network, address string) error {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return errors.New("TCP network required")
	}
	if len(address) > 4096 {
		return errors.New("TCP address too long")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" {
		return errors.New("TCP host and numeric port required")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("TCP port out of range")
	}
	return nil
}
func DialTCP(ctx context.Context, network, address string) (*net.TCPConn, error) {
	raw, err := (&net.Dialer{}).DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(*net.TCPConn)
	if !ok {
		_ = raw.Close()
		return nil, errors.New("dial returned non-TCP connection")
	}
	return conn, nil
}
