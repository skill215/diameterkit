# Ro SMS, abort and explicit recovery examples

Run from the repository root:

```sh
CGO_ENABLED=0 go run ./examples/ro
CGO_ENABLED=0 go run ./examples/ro -result 2001/4012
CGO_ENABLED=0 go run ./examples/ro -flow ecur
CGO_ENABLED=0 go run ./examples/ro -flow ecur -term-result 5002
CGO_ENABLED=0 go run ./examples/ro -flow abort -result 5002
CGO_ENABLED=0 go run ./examples/ro -flow recovery -recovery-scenario no-answer
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-binding combined -compat-action debit
```

## Standard SMS IEC and ECUR

The CTF submits CCR to a synthetic OCS over loopback TCP. IEC sends one
DIRECT_DEBITING EVENT_REQUEST with explicit Requested-Action=0, request number 0,
one service and requested unit. The selected Service-Context is `18.32274@3gpp.org`.
The peer has no ledger, rating, authentication service or persistence.

| Result option | Meaning |
| --- | --- |
| 2001 | Command and service success with explicit synthetic grant/token where applicable |
| 2001/4012 | Command success with service rejection; do not treat this as a successful debit |
| 4012 or 5012 | Normal command-level failure |
| 3004 | Independent E error envelope, not normal CCA syntax |

Preserve command-level and MSCC outcomes separately. Refund-Information 2022 is
opaque evidence, not permission to refund. Units stay exact uint64, including
values beyond float's exact integer range. No timeout/cancel/failure branch
proves that charging did not occur or automatically triggers refund/replay.

ECUR sends INITIAL(type=1/number=0) then TERMINATION(type=3/number=1), without
UPDATE. Session, parties and service identity stay associated; each exchange gets
new protocol/request identifiers. The application consumes INITIAL, verifies both
result levels and sufficient grant, then explicitly sends TERM with USU/FINAL.
`CheckECURContinuation` checks selected request continuity, not business permission.
INITIAL rejection suppresses TERM. `-term-result` supports 2001/5002/5012/3004.

## Independent abort

`-flow abort` reverses direction: OCS sends ASR to CTF using an application-supplied
original Session reference (`-session`). ASA can report 2001, unknown Session 5002,
5012 or E3004. It is an independent 4/274 exchange; successful ASA does not finish
a same-Session CCR, generate TERM, or authorize recovery. The application owns the
actual session action. No session database is created.

## Recovery and named compatibility bindings

`-flow recovery` exercises explicit standard debit continuation; scenarios include
`error`, `restart` and `no-answer`. `-flow compat` selects exactly one binding:
`combined`, `smsfwro` or `legacy420`. It maps the supported `debit`, `balance`,
`refund`, `initial` and `term` operations according to that binding; unsupported
combinations fail rather than guessing another profile. Compatibility scenarios
also include `no-answer-import`.

The example holds recovery records/progress in memory and explicitly supplies
policy, cumulative limits, target evidence and original deadlines. Import is not
a storage service. Cross-target continuation requires declared authorization;
late primary answers must not complete a successor. These named legacy mappings
do not authorize standard SMS refunds, SCAP1 recovery or arbitrary OCS failover.

See [consumption.go](consumption.go), [recovery.go](recovery.go),
[compat_recovery.go](compat_recovery.go) and their tests. Public consumer coverage
includes admission negatives, cancellation, late answers, ownership, independent
bindings and no-answer/import flows. [Application handoff](../../doc/ro-recovery-handoff.md)
separates financial decisions from protocol evidence.

## Lifecycle and limits

These examples use public APIs, real loopback TCP and synthetic endpoint data.
They need Go 1.26+, no external service and no database. The default run deadline
is 10 seconds; cleanup has its own finite deadline. Submit contexts govern
admission, Receive contexts govern waiting, and Node.Cancel cancels an admitted
exchange. Unconsumed completions retain capacity. Shutdown/Close/Wait and port
cleanup are explicit; successful runs report zero final resources.

Profiles do not grant direction or terminal authorization. Configure capabilities,
request/answer policies and ultimate endpoint identities explicitly, especially
when a DRA is the adjacent peer. Budget values are logical reservations, not RSS
or production sizing guidance. AnswerReceived and complete protocol validation
do not by themselves establish business success. A canceled/timed-out operation
may already have reached its peer.

See [support](../../doc/diameterkit-support-matrix.md),
[version/installation policy](../../doc/version-install-policy.md) and
[AI disclosure](../../AI_DISCLOSURE.md). Test evidence covers the declared local
workload; real target acceptance and durable application state are separate.
