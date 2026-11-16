package transport

import "context"

// ListenerInfo separates fixed listener storage from the maximum per-connection
// adapter storage (including an accepted, not yet established socket).
type ListenerInfo struct {
	BufferBytes, ConnectionBufferBytes uint64
}

// Incoming owns one accepted socket. Establish performs TLS when applicable,
// honors ctx and stops borrowing before returning. It is called at most once.
// Close is idempotent, concurrent-safe and unblocks Establish. Every non-nil
// returned Conn transfers ownership, including on error. On success the caller
// must stop all Incoming.Close callbacks before transferring Conn to the runtime.
type Incoming interface {
	Info() Info
	Establish(context.Context) (Conn, error)
	Close() error
}

// Listener has one Accept caller. Accept only accepts a socket; it must not
// perform TLS or Diameter handshakes. Cancellation releases that call without
// closing the listener. Close unblocks Accept and permanently stops listening.
// Any non-nil Incoming transfers ownership even alongside an error. Methods and
// metadata are trusted; arbitrary uncooperative callbacks cannot be killed.
type Listener interface {
	Info() ListenerInfo
	Accept(context.Context) (Incoming, error)
	Close() error
}
