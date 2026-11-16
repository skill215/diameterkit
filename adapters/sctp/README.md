# Diameter SCTP adapter

> **AI-written:** generated and revised with OpenAI Codex under maintainer direction.
> Historical model identities are unrecorded; see [AI disclosure](AI_DISCLOSURE.md).

This independent module uses pure Go 1.26 and the published upstream release
`github.com/skill215/sctp v0.8.0`.
The core module has no SCTP dependency. Linux except 386 is supported by the
pinned upstream message API; other platforms fail explicitly. This batch proves
native Linux/amd64 operation, not native networking on every build target.

Create sockets with `SocketConfig()` for defaults, or `config.SocketConfig()`
for explicit policies/stream requests, before passing them to `New(raw, config)`.
`New` takes exclusive ownership on success **and failure**. Stop using the raw
socket, options and deadlines after transfer. A connection has one reader and
one writer; `Close` may run concurrently and unblocks both. Prebuilt socket
establishment is outside Node's admission budget. Use `NewDialer` and `Listen` below for managed establishment.

Call `config.Requirements()` before constructing Node and set
`Resources.AdapterBytesPerConnection` to the returned `BufferBytes`. Provision
`TransportBytes`/`NodeBytes` for all candidates as usual. The default is 17756
logical bytes on amd64: 9516 from the upstream versioned ConnectionBytes query
+ 4096 adapter state/scratch allowance + 4096 notification + 48 OOB. Node separately reserves 2048 per candidate for
owned event/fact snapshots, plus its existing control and emergency slots.
These are logical library commitments, not heap/RSS/kernel queue ceilings.

Pass the adapter to Node.Connect/Accept. Node handles `*transport.Event`
continuations synchronously, including while receiving/discarding a frame.
Direct users of transport.Conn must handle nonterminal events by resuming the
same method and destination; a terminal event forbids continuation. Notification
tails are never retained. The fixed notification size and rate limits close on
excess; no queue or growing buffer is used.

PPID is 46. CER/CEA/DWR/DWA/DPR/DPA use ordered stream 0. Node enables unordered
stream 0 for application frames after its handshake barrier; the responder
immediately queues a DWR after CEA and keeps applications ordered until a
matched DWA or other valid post-handshake message. The barrier has a fixed
Handshake timeout. Direct `WriteFrame` is always ordered; the optional
`transport.MessageWriter` is for explicit per-frame ordering by an endpoint.
Incoming streams may differ from local outbound streams. Explicit
`OrderedMultiStream` and `PPID0Compatibility` are available as described below.
DTLS and SEND_FAILED attempt association are not enabled.
SEND_FAILED with Context=0 is a bounded diagnostic, never delivery confirmation
or permission to retry. No application sessions, persistence or business replay
are added.

## Explicit stream and PPID policies

`DefaultConfig()` retains `UnorderedStream0` and `PPID46`. The policies are
independent and frozen when a socket/dialer/listener is constructed. No receive
packet, DRA advertisement or notification changes them. `PPID46` sends and
accepts only 46; `PPID0Compatibility` sends 0 and accepts 0 or 46. Both reject 47 and
unknown values. Enable compatibility only for an endpoint with that deployment
agreement; it does not provide DTLS or endpoint authorization.

`OrderedMultiStream` uses ordered stream 0 for Base CER/CEA/DWR/DWA/DPR/DPA and
all frames sent with the handshake barrier. After the barrier, application
frames use `1 + hash % (actual Ostreams - 1)`. Actual Ostreams=1 falls back to
ordered stream 0. `OutboundStreams` and `MaxInboundStreams` request counts from
the kernel, never replace negotiated directional counts. Zero chooses outbound 1
in default mode, outbound 16 in ordered mode, and inbound 16 in either mode.

Mapping is fixed 64-bit FNV-1a: byte 1 followed by exact top-level vendor 0 Session-Id;
without Session, byte 2 followed by AppID(4), CommandCode(3), E2E(4), all in wire
byte order. HBH/R/T are excluded. Only the public wire top-level walker is used;
groups are opaque, padding ignored, and ambiguous/empty Session-Id or malformed
AVP boundaries rejected before send. No application decode, per-session table,
payload copy or stream queue is added. Hash collisions may share a stream;
reconnects may negotiate different counts. This is not cross-connection session
ordering or business deduplication.

`WriteFrame` and `WriteFrameOrdered(..., true)` force ordered stream 0.
`WriteFrameOrdered(..., false)` permits the configured post-handshake policy;
Base control still uses ordered stream 0. Answers independently use local outbound
counts, never the remote request's SID. The existing CEA/DWR barrier remains in
Node; notifications do not clear it. DPR/DPA is not a remote delivery fence
across streams. Receive still accepts any negotiated inbound stream and either
ordered or unordered data, while upstream enforces metadata continuity to EOR.

A returned `transport.Conn` implements `ModeReporter`:
`c.(sctp.ModeReporter).SCTPMode()` returns an immutable `ModeInfo` with both
actual stream counts and `SingleStreamFallback`. Query this before transfer to
Node if needed; no raw socket ownership is exposed. Fallback has ordered HOL,
and level 0 partial delivery retains receive HOL even with multiple streams.
Mode fields and the temporary wire cursor fit the existing fixed reserve;
logical default budgets remain 17756/99676/52620 on Linux/amd64. This says
nothing about kernel per-stream memory or end-to-end latency.

```go
setup := sctp.DefaultSetupConfig()
setup.Transport.StreamMode = sctp.OrderedMultiStream
setup.Transport.OutboundStreams = 8
setup.Transport.MaxInboundStreams = 16
// Set only for an explicitly compatible endpoint:
setup.Transport.PPIDPolicy = sctp.PPID0Compatibility
requirements, err := setup.Requirements()
// Handle err, reserve requirements in NodeConfig, then NewDialer/Listen.
```

## Managed establishment

`DefaultSetupConfig()` bounds combined local+remote explicit addresses N=16,
zone bytes Z=256 and endpoint text bytes D=4096. Validation occurs before copying
or resolving; accepted text is frozen. Supported networks are sctp/sctp4/sctp6.
Endpoints use slash-separated hosts and a numeric last port; remote port zero,
service names and empty multihoming hosts are rejected. DNS selects one address
per explicit host, without automatic discovery or DNS-expanded multihoming.

Call `setup.Requirements()` before Node creation. Configure
`Resources.AdapterBytesPerConnection` from `ConnectionBufferBytes` and, for
Serve, `Resources.ListenerBytesPerPeer` from `BufferBytes`. Defaults on the
pinned Linux/amd64 ledger are **99676 per physical candidate** and **52620 per
listener**, additive. Upstream `SetupRequirements` reports ledger
`bounded-setup-v2/linux/amd64`, Dialer=8960, Listener=48524, Establish=78380 and
Connection=9516. The adapter adds its own SetupStateReserve=4096 per listener and
candidate, plus StateReserve=4096/notification=4096/OOB=48 per candidate:

- Listener: upstream ListenerBytes + SetupStateReserve.
- Candidate: upstream DialerBytes + max(EstablishBytes, ConnectionBytes)
  + SetupStateReserve + StateReserve + NotificationMax + OOBMin.
- Established socket takeover: upstream ConnectionBytes + StateReserve
  + NotificationMax + OOBMin.

Upstream scratch is already included. The old 8192 reserve and copied address
formula are removed. Candidate declarations remain conservatively reserved after
handoff. Resolver internals, runtime/allocator/stacks and kernel queues remain
outside the logical budget. Provision Node/TransportBytes for all candidates,
including accept waits, election losers and retiring connections.

```go
setup := sctp.DefaultSetupConfig()
requirements, err := setup.Requirements() // No socket I/O.
// Handle err; use requirements in NodeConfig before creating node.
dialer, err := sctp.NewDialer("sctp4", "", "127.0.0.1:3868", setup)
// Handle err; node.Dial(ctx, peerID, dialer) or node.Maintain(ctx, peerID, dialer).
listener, err := sctp.Listen(ctx, "sctp4", "127.0.0.1:3868", setup)
// Handle err; node.Serve(ctx, peerID, listener) takes ownership, even on rejection.
```

The snippet shows construction order, not a complete application. Binding and
its concurrency remain caller-owned. `Listen` reports a frozen bind address,
including an assigned ephemeral port, rather than enumerating wildcard-bound
interfaces. Node reserves before Dial or Accept; resolution, connect and CER/CEA
share the absolute handshake deadline. Idle Accept does not spend Handshake.

The adapter delegates directly to upstream `NewDialer`, `ListenContext`,
`AcceptSCTPContext`, `BoundAddr` and `SetSocketTimeouts`. There is no adapter
socket syscall, address parser or periodic Accept loop. Upstream uses its
existing cancellable poller and listener deadline gate. One adapter Accept
caller is allowed; concurrent calls return ErrConcurrentAccept. A canceled call
leaves the listener reusable; Close permanently ends it. Sockets arriving after
cancellation are closed before the adapter handoff. Connected sockets have legacy
SO timeouts and prior standard deadlines cleared before takeover.

SCTP metadata continuity and directional stream validation now come from
upstream. The adapter retains Diameter message length/EOR validation, explicit PPID
policy, frame deadlines and notification quotas. Valid restart/partial-abort
notifications are projected even when upstream also sets CloseRequired or
ErrAssociationChanged; malformed/failed records never become continuations.
A writer observing the restart latch first retains a transport-termination cause
and its actual byte count, without inventing a receive-event sequence.

An Incoming permits one Establish and idempotent concurrent Close. No setup
cancellation hook survives successful handoff. Serve cancellation closes the
listener and unfinished candidate, leaving established Peer connections usable;
Maintain cancellation retires its connection domain. Core Wait includes setup
borrowers until actual return. Reconnect passes through REOPEN and never replays
business requests. Formal v0.8.0 regression is in [batch 67](../../spec/history/batch-67.json); migration evidence is in [batch 66](../../spec/history/batch-66.json). Prior managed
contracts remain in [batch 65](../../spec/history/batch-65.json).

## Installation and development

The independent adapter candidate is `v0.1.0-rc.1`, targeting core
`v0.1.0-rc.2` and upstream SCTP `v0.8.0`. Consult the
[publication record](../../doc/releases/v0.1.0-rc.md) for actual availability.

```sh
go get github.com/skill215/diameterkit/adapters/sctp@v0.1.0-rc.1
GOWORK=off CGO_ENABLED=0 go -C adapters/sctp test -mod=readonly ./...
GOWORK=off go -C adapters/sctp test -mod=readonly -race ./...
```

Run checkout tests from the repository root on Linux with SCTP support. A root
`go test ./...` does not include this nested module. Missing SCTP support fails
rather than skips. [Checkout CI](../../doc/release-checks.md) explicitly tests both
current modules; [real installation](../../doc/version-install-policy.md) instead
uses two empty external caches, GOWORK=off, no replace and authenticated downloads.
Old offline batch candidates and reconstructed proxies are historical evidence,
not current release dependencies. Public API drift requires review.

## Benchmarks

The 256/4096-byte frame benchmarks measure one-way loopback send plus owned frame
receive. Connection setup is excluded; allocations include receive ownership.
They do not measure remote RTT or production network capacity. See the final
[README table](../../README.md#benchmarks), [raw baseline](../../spec/release/performance-baseline.json)
and [reproduction commands](../../doc/load-checks.md).
