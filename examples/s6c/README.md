# S6c GMSC/HSS routing example

```sh
CGO_ENABLED=0 go run ./examples/s6c
CGO_ENABLED=0 go run ./examples/s6c -result 5001
CGO_ENABLED=0 go run ./examples/s6c -result 5012
CGO_ENABLED=0 go run ./examples/s6c -result 3004
```

A GMSC submits an ordinary MSISDN SRR to a synthetic HSS. 2001 returns the selected
routing data; 5001 uses the 10415 experimental namespace, 5012 is a normal failure,
and 3004 is an E answer. The example does not query subscribers, rank candidates,
resolve IP-SM-GW names or deliver SMS.

[config.go](config.go) shows Base+S6c, explicit Vendor10415/Auth16777312 capability,
fixed direction, destination and permitted answer origin. [main.go](main.go) uses
`NewSRR` plus the public Node/Peer lifecycle. [client_test.go](client_test.go)
exercises the thin Route client, synchronous query, async submission and cancellation.

Use [NewSRR/InterpretSRA and client semantics](../../doc/s6c-routing-helpers.md)
for result namespaces, context and ownership. Validate header association and
source before interpreting offline bodies. Primary/Additional and combined nodes
remain distinct; unsupported context must not yield usable candidates. Ordinary
routing does not automatically authorize IMS/T4/Router/MCCMNC behavior.

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
