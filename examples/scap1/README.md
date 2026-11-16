# SCAP1 bidirectional public API example

```sh
CGO_ENABLED=0 go run ./examples/scap1
CGO_ENABLED=0 go run ./examples/scap1 -legacy-advertisement
CGO_ENABLED=0 go run ./examples/scap1 -result 5012
CGO_ENABLED=0 go run ./examples/scap1 -legacy-advertisement -result 3004
```

Two independent Nodes each send one ACR and answer the other endpoint over TCP.
SCAP1 is an explicitly named legacy accounting binding: header 0/271 with
Vendor193/Acct19302 capability and SCAP1Accounting policy. It is not Ro App4.

Standard advertisement is the default. `-legacy-advertisement` selects the reviewed
legacy dual-group layout on both peers. Never inject arbitrary dual Auth/Acct
groups into LocalCapabilities; selecting compatibility on one endpoint does not
silently authorize it on the other.

The public path is NewSessionID → scap1.NewEvent → Submit → Receive → Respond →
CompletionPort.Receive. Inputs contain explicit synthetic numbers/text/wire time.
2001 and 5012 use ACA; 3004 uses an independent E answer. Respond restores the
original Session, Proxy and header association. No subscriber or charging facts
are inferred from protocol success.

This example is NoCache/SingleAttempt, with no Ro recovery records or persistent
ledger. Optional bounded duplicate handling elsewhere is not crash-safe deduplication.
Tests cover advertisement combinations, bidirectional results, record association,
capability/direction/identity negatives, slot/byte admission, ownership, cancel,
late answer and port/Node lifecycle. Private ACA/recovery and real deployment
acceptance remain separate.

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
