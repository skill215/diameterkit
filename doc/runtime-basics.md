# Runtime and outcome semantics

Use the standalone codec for protocol tools and the Node/Peer runtime for live
connections. The complete [Ro configuration](../examples/ro/config.go) and
[lifecycle](../examples/ro/main.go) show the required public setup.

## Configure an endpoint

Compile explicit profiles into a bundle, advertise capabilities, configure
permitted request directions and answer origins, and provision slot/byte budgets.
Establish connections with `Node.Dial` or `Node.Serve`. TCP/TLS are included in the
core; install the [adapter](../adapters/sctp/README.md) only for SCTP.

Profiles describe protocol structure. They do not grant direction or terminal
authorization. A DRA's identity does not implicitly authorize the ultimate
application endpoint. Overlapping 4/272 CCR bindings cannot share a bundle.

## Exchange and consume

| Task | Public API |
| --- | --- |
| Allocate a bounded session identity | `Node.NewSessionID` |
| Submit through an existing completion port | `Peer.Submit` |
| Receive and answer inbound work | `Peer.Receive`, `Inbound.Respond` |
| Consume outcomes and release completion capacity | `CompletionPort.Receive` |
| Cancel an admitted exchange | `Node.Cancel` |
| Drain, close, and join library work | `Node.Shutdown`, `Node.Close`, `Node.Wait` |
| Construct and interpret S6c routing | `s6c.NewSRR`, `s6c.InterpretSRA`, `client/s6c` |
| Prepare and submit authorized Ro recovery | `Node.PrepareRoRecoveryRequest`, `Node.SubmitRoRecovery` |

A Submit context controls admission; canceling it later does not cancel the
exchange. A Receive context stops waiting. Use explicit cancellation for an
admitted exchange. Unconsumed completions retain capacity. `Wait` joins library
work but does not consume application results; close or drain completion ports
as part of cleanup.

## Interpret outcomes

`AnswerReceived` and `Validation.Complete` describe protocol evidence. Inspect
command and service results separately. A timeout or cancellation after writing
does not prove that the remote operation did not occur. Reconnection does not
automatically replay business requests.

Recovery requires explicit policy, identity, target and cumulative-budget
evidence. Earlier recovery records remain tied to their original Bundle
fingerprint. Read [Ro handoff](ro-recovery-handoff.md) and the
[version policy](version-install-policy.md) before migration.

The library owns no database, subscriber lookup, charging ledger, delivery engine
or durable recovery service. Applications own business decisions and persistence.
Local duplicate handling is bounded and volatile. Logical resource reservations
are not heap/RSS or production capacity guarantees.

See [S6c helper semantics](s6c-routing-helpers.md),
[protocol support](diameterkit-support-matrix.md) and the [documentation index](README.md).
