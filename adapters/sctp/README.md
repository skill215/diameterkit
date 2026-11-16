# Diameter SCTP adapter

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

PPID is 46. CER/CEA/DWR/DWA/DPR/DPA use ordered stream0. Node enables unordered
stream0 for application frames after its handshake barrier; the responder
immediately queues a DWR after CEA and keeps applications ordered until a
matched DWA or other valid post-handshake message. The barrier has a fixed
Handshake timeout. Direct `WriteFrame` is always ordered; the optional
`transport.MessageWriter` is for explicit per-frame ordering by an endpoint.
Incoming streams may differ from local outbound streams. Explicit
`OrderedMultiStream` and `PPID0Compatibility` are available as described below.
DTLS and SEND_FAILED attempt association are not enabled.
SEND_FAILED with Context0 is a bounded diagnostic, never delivery confirmation
or permission to retry. No application sessions, persistence or business replay
are added.

## Explicit stream and PPID policies

`DefaultConfig()` retains `UnorderedStream0` and `PPID46`. The policies are
independent and frozen when a socket/dialer/listener is constructed. No receive
packet, DRA advertisement or notification changes them. `PPID46` sends and
accepts only46; `PPID0Compatibility` sends0 and accepts0 or46. Both reject47 and
unknown values. Enable compatibility only for an endpoint with that deployment
agreement; it does not provide DTLS or endpoint authorization.

`OrderedMultiStream` uses ordered stream0 for Base CER/CEA/DWR/DWA/DPR/DPA and
all frames sent with the handshake barrier. After the barrier, application
frames use `1 + hash % (actual Ostreams - 1)`. Actual Ostreams=1 falls back to
ordered stream0. `OutboundStreams` and `MaxInboundStreams` request counts from
the kernel, never replace negotiated directional counts. Zero chooses outbound1
in default mode, outbound16 in ordered mode, and inbound16 in either mode.

Mapping is fixed FNV-1a64: byte1 followed by exact top-level vendor0 Session-Id;
without Session, byte2 followed by AppID(4), CommandCode(3), E2E(4), all in wire
byte order. HBH/R/T are excluded. Only the public wire top-level walker is used;
groups are opaque, padding ignored, and ambiguous/empty Session-Id or malformed
AVP boundaries rejected before send. No application decode, per-session table,
payload copy or stream queue is added. Hash collisions may share a stream;
reconnects may negotiate different counts. This is not cross-connection session
ordering or business deduplication.

`WriteFrame` and `WriteFrameOrdered(..., true)` force ordered stream0.
`WriteFrameOrdered(..., false)` permits the configured post-handshake policy;
Base control still uses ordered0. Answers independently use local outbound
counts, never the remote request's SID. The existing CEA/DWR barrier remains in
Node; notifications do not clear it. DPR/DPA is not a remote delivery fence
across streams. Receive still accepts any negotiated inbound stream and either
ordered or unordered data, while upstream enforces metadata continuity to EOR.

A returned `transport.Conn` implements `ModeReporter`:
`c.(sctp.ModeReporter).SCTPMode()` returns an immutable `ModeInfo` with both
actual stream counts and `SingleStreamFallback`. Query this before transfer to
Node if needed; no raw socket ownership is exposed. Fallback has ordered HOL,
and level0 partial delivery retains receive HOL even with multiple streams.
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
business requests. Formal v0.8.0 regression is in [batch67](../../spec/history/batch-67.json); migration evidence is in [batch66](../../spec/history/batch-66.json). Prior managed
contracts remain in [batch65](../../spec/history/batch-65.json).

## Development and module verification

The adapter's core requirement `v0.0.0-diameterkit.batch85` is an **offline-only
production-source candidate**, not a published core version. Its source manifest
and archive hash are in the [batch evidence](../../spec/evidence/batch-85-ro-abort-example.json). The 138-file core source package
includes Relay negotiation, bounded per-request terminal bindings and the Ro abort/generic DCCA profiles. Publication
must replace it with a real immutable core version and re-run both modules with
GOWORK=off. A local workspace/replace is not a release dependency lock.

Reconstruct the exact local core package, then test the independent module:

```sh
python3 spec/tools/core_candidate.py . /tmp/diameterkit-core-proxy
GOWORK=off GOMODCACHE=/tmp/diameterkit-sctp-cache \
  GOPROXY=file:///tmp/diameterkit-core-proxy,https://proxy.golang.org \
  GONOSUMDB=github.com/skill215/diameterkit \
  go -C adapters/sctp mod download
GOWORK=off GOMODCACHE=/tmp/diameterkit-sctp-cache CGO_ENABLED=0 \
  go -C adapters/sctp test -mod=readonly ./...
GOWORK=off GOMODCACHE=/tmp/diameterkit-sctp-cache \
  go -C adapters/sctp test -mod=readonly -race ./...
```

Run from the repository root with a kernel allowing local SCTP sockets. Tests
fail when the facility is unavailable; they do not silently skip SCTP coverage.
The batch67 checks downloaded the public v0.8.0 module into a fresh cache,
verified its sum and all70 Go sources, and used no replacement directive.
Only the unpublished core candidate came from a reconstructed offline proxy. Root `go test ./...` does not include this
nested module.

Batch68 verifies Relay-only active/passive handshake, the CEA/DWR barrier and
two selected terminals on one real SCTP association using the new core candidate.
See [DRA integration](../../spec/history/batch-68.json); upstream remains v0.8.0.

Batch69 advances the offline core candidate to `v0.0.0-diameterkit.batch69`
(117 production-package files) for the new Ro abort profile. Upstream v0.8.0
and adapter production code are unchanged; this module is regressed separately.
See [batch69](../../spec/history/batch-69.json). The new ASR/ASA application scenarios
use TCP/TLS; SCTP evidence remains the existing Base/S6c/lifecycle suite.

Batch70 consumes offline core `v0.0.0-diameterkit.batch70` (120 files).
Upstream remains v0.8.0 and adapter production sources are unchanged.
Generic DCCA application tests use TCP/TLS; this module separately reruns its
existing complete SCTP suite. See [batch70](../../spec/history/batch-70.json).

Batch71 consumes offline core `v0.0.0-diameterkit.batch71` (121 files).
Published SCTP v0.8.0 and adapter production sources remain unchanged. New
standard SMS Ro result/2022 network cases use TCP/TLS; the complete existing
SCTP suite is regressed against the new core. See [batch71](../../spec/history/batch-71.json).

Batch72 implements explicit ordered multi-stream and PPID0 policies using the
same core batch71 and published upstream v0.8.0. Real asymmetric negotiation,
single-stream fallback, Base/S6c, CEA/DWR barrier, admission and cancel/close
are covered; see [batch72](../../spec/history/batch-72.json). Public module release and
actual DRA/OCS acceptance remain separate exits.

Batch73 consumes offline core `v0.0.0-diameterkit.batch73` with the explicit legacy SGd Alert binding. Adapter production and upstream v0.8.0 are unchanged. The new application lifecycle tests use TCP/TLS; this module runs its full real SCTP regression separately. See [batch73](../../spec/history/batch-73.json).

Batch74 consumes offline core `v0.0.0-diameterkit.batch74` with explicit Router Alert evidence. Adapter production and upstream v0.8.0 remain unchanged; this module is regressed separately. New Router application flows are tested over TCP/TLS. See [batch74](../../spec/history/batch-74.json).

Batch75 consumes offline core `v0.0.0-diameterkit.batch75` with bounded Router TFR/TFA token output. Adapter production and upstream v0.8.0 are unchanged; separate full regression uses a fresh module cache without replace. New application flows use TCP/TLS; see [batch75](../../spec/history/batch-75.json).

Batch76 consumes offline core `v0.0.0-diameterkit.batch76` for bounded SRA Router/MCCMNC purposes. Adapter production and SCTP v0.8.0 remain unchanged; fresh-cache no-replace regression is independent of the new TCP/TLS application tests. See [batch76](../../spec/history/batch-76.json).

Batch77 adds simulator-input tests and documentation only. Core remains the identical 125-file batch76 candidate and SCTP remains v0.8.0. Both adapter CGO0/race suites pass (113 each); new application flows use TCP/TLS, not old simulator processes. See [batch77](../../spec/history/batch-77.json).

Batch78 consumes the new 126-file offline core `v0.0.0-diameterkit.batch78` with peer-scoped SCAP1 handshake compatibility. Fresh-cache, GOWORK=off/no-replace checks retain formal SCTP v0.8.0; no adapter production changes or new SCAP1 SCTP deployment claim. See [batch78](../../spec/history/batch-78.json).

Batch79 consumes offline core `v0.0.0-diameterkit.batch79`, adding the public SCAP1 TCP example and correcting API documentation without runtime behavior changes. Fresh-cache no-replace regression retains SCTP v0.8.0. The example is not a new SCTP deployment test; see [batch79](../../spec/history/batch-79.json).

Batch80 consumes offline core `v0.0.0-diameterkit.batch80`, adding the public Sh Ref24 TCP example. All existing library/adapter production behavior and SCTP v0.8.0 remain unchanged. Fresh-cache no-replace CGO0/race regression passes; this example does not claim a new Sh SCTP deployment. See [batch80](../../spec/history/batch-80.json).

Batch81 consumes offline core `v0.0.0-diameterkit.batch81`, adding the S6c GMSC/HSS public TCP example. Existing core/adapter sources and SCTP v0.8.0 remain unchanged. Fresh-cache no-replace CGO0/race suites pass (113 each); this is not new S6c SCTP deployment evidence. See [batch81](../../spec/evidence/batch-81-s6c-example.json).

Batch82 consumes offline core `v0.0.0-diameterkit.batch82`, adding the SGd/Gdd MO/MT public TCP example. Existing library/adapter Go sources and formal SCTP v0.8.0 stay unchanged. Fresh-cache no-replace CGO0/race regression passes (113 each); this is not new SGd/Gdd SCTP deployment evidence. See [acceptance evidence](../../spec/evidence/batch-82-sgd-example.json).

Batch83 consumes offline core `v0.0.0-diameterkit.batch83`, adding the standard Ro SMS IEC public TCP example. Existing Go sources and formal SCTPv0.8.0 remain unchanged. Fresh-cache no-replace CGO0/race suites pass (113 each); no new Ro SCTP/target OCS deployment claim. See [evidence](../../spec/evidence/batch-83-ro-example.json).

Batch84 consumes offline core `v0.0.0-diameterkit.batch84`, extending the standard Ro example with no-UPDATE ECUR. Library/adapter production behavior and formal SCTPv0.8.0 remain unchanged. Fresh-cache no-replace CGO0/race suites pass (113 each); this TCP example adds no Ro SCTP/target OCS acceptance claim. See [evidence](../../spec/evidence/batch-84-ro-ecur-example.json).

Batch85 consumes offline core `v0.0.0-diameterkit.batch85`, adding the independent Ro ASR/ASA TCP example. Library/adapter runtime and formal SCTPv0.8.0 remain unchanged. Fresh-cache no-replace CGO0/race suites pass (113 each); this adds no Ro SCTP/target acceptance claim. See [evidence](../../spec/evidence/batch-85-ro-abort-example.json).

## Current checkout CI (batch91)

Use `python3 spec/tools/release_check.py` from the repository root for current core+adapter acceptance. The runner creates a temporary workspace with a versioned replacement for the offline core requirement, validates both Main module paths, then runs both CGO0 and race suites. It does not edit this module's dependency lock or claim that the old batch85 candidate is the current release. See [checkout checks](../../doc/release-checks.md) for offline prerequisites, zero-skip native SCTP policy and the separate GOWORK=off release-installation gate.

Batch93 defines the [independent v0 module/API and installation policy](../../doc/version-install-policy.md). The batch85 requirement is still an unpublished historical placeholder. Publish a reviewed core tag first, then update this requirement and real checksums before publishing the adapter tag. The new cold-install gate rejects replacements and fails on the currently missing versions; checkout/probe tests do not establish a published installation.
