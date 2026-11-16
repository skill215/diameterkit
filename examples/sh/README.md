# Sh Ref24 public API example

```sh
CGO_ENABLED=0 go run ./examples/sh
CGO_ENABLED=0 go run ./examples/sh -result 5012
CGO_ENABLED=0 go run ./examples/sh -result 3004
```

Two independent Nodes exchange one UDR each and answer the other direction. This
symmetric local demonstration is not a deployment rule permitting an HSS to send
UDR toward an AS. Configure real AS/HSS roles explicitly.

2001 carries synthetic Ref24 SMSRegistrationInfo; 5012 is normal UDA failure;
3004 is an E envelope. Base+Sh and Vendor10415/Auth16777217 capability are explicit.
Successful NewNode construction does not imply negotiated capability.

The responder calls `shdata.MarshalSMSPull`; the consumer explicitly calls Parse
and Check(SMSPull). XML has its own finite byte/depth limits, separate from Node
accounting. Diameter Validation.Complete does not validate the XML or subscriber
permissions. These helpers are not a full XSD validator or HSS database.

[Tests](main_test.go) cover capability negotiation, result/data/VSAI constraints,
XML limits, independent budgets, ownership, cancellation, late answers and cleanup.
The example does not enlarge PUR/SNR/PNR, other references or TLS/SCTP scope.

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
