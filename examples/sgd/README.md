# SGd/Gdd MO and MT forwarding example

```sh
CGO_ENABLED=0 go run ./examples/sgd
CGO_ENABLED=0 go run ./examples/sgd -interface gdd
CGO_ENABLED=0 go run ./examples/sgd -result 5555
CGO_ENABLED=0 go run ./examples/sgd -interface gdd -result 3004
```

The two endpoints exchange OFR/OFA and TFR/TFA over real loopback TCP. `sgd`
selects MME roles; `gdd` selects SGSN roles under the shared 16777313 application.
Each endpoint has independent outbound/inbound policies and a completion port.
TPDU bytes are opaque and identities are supplied explicitly.

2001 is selected protocol success, 5012 a normal failure, experimental10415/5555
requires its selected cause, and 3004 is an E envelope. Direction-specific causes,
identities and Gdd restrictions are validated. Result acceptance does not establish
SMS delivery. This example adds no automatic retry, charging, subscriber database,
IMS/T4/Router context or retransmission-time negotiation.

Both peers must reach Ready within the finite connection context before the
example starts bidirectional business traffic. Dial returning does not substitute
for the accepting peer's CEA write barrier. Ready remains a mutable snapshot;
applications must still handle later Submit errors.

[Tests](main_test.go) cover both interfaces/directions, results, policy/identity
negatives, payload limits, independent budgets, mutable views, cancellation and
late-answer single completion. Local TCP evidence does not establish target DRA,
access-node or SCTP deployment interoperability.

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
