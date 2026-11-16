# Public binding consumer contracts

```sh
CGO_ENABLED=0 go test ./examples/contracts
```

[contracts_test.go](contracts_test.go) is an external-package consumer that can be
copied into a separate module. It uses installed public APIs, without internal
packages, repository fixture files or private root helpers.

Coverage includes profile entry points, pairwise rejection of conflicting 4/272
bindings, independent abort composition, generic DCCA EVENT balance/debit/refund,
new refund sessions, exact amounts/units, explicit private tokens and ECUR request
continuity. Standard 2022 and private 1292 namespaces are never interchangeable.

Message paths use the public Codec for envelope and original-request/context
validation. NO_CREDIT is not grant permission; standard token helpers do not
permit standard SMS refund sending. This is synthetic protocol consumption,
not independent-stack or financial interoperability.

Ro recovery signatures and invalid records are checked here; actual runtime
segments, cross-target and import consumption are exercised in the separate
[Ro example suite](../ro/README.md). The [install gate](../../doc/version-install-policy.md)
copies both suites into isolated consumers and distinguishes checkout rehearsal
from actual published dependency installation.
