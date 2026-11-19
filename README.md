<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: README structure and user-facing introduction revised under maintainer direction.
SPDX-AI-Date: 2026-11-19
-->

# diameterkit

**Diameter clients, responders, and protocol tools for Go.**

Build SMS routing and charging integrations with typed messages, bounded
connections, and explicit request outcomes. Includes selected **Sh, S6c,
SGd/Gdd, and Ro** procedures, built-in TCP/TLS, and an optional Linux SCTP adapter.

**Pure Go · Go 1.26+ · Zero third-party runtime dependencies in the core**

[Install](#install) · [Try it](#try-it) · [Examples](#examples) ·
[Documentation](#documentation) · [Benchmarks](#benchmarks)

> **Development preview.** Published core and SCTP adapter: `v0.2.0-rc.1`.
> Independent cold installation is verified. See the [release record](doc/releases/v0.2.0-rc.1.md)
> and [RC scope checklist](doc/release-candidate-checklist.md); this is not stable-v1.
> The development checkout adds [SMS](doc/conditional-sms-values.md) and [unit value diagnostics](doc/conditional-unit-values.md) beyond that published pair.

## Why diameterkit

- **Typed protocol tools.** Concrete Go messages, explicit optional fields,
  owned decoded data, and generated codecs without reflection or runtime dictionaries.
- **Bounded peer runtime.** Capability negotiation, watchdogs, deadlines,
  request correlation, backpressure, and explicit slot/byte budgets.
- **Application helpers.** Construct and interpret S6c routing requests; use
  selected Ro charging flows with explicit command and service results.
- **Explicit ownership.** Your application controls retries, recovery decisions,
  subscriber data, charging state, and persistence.

## Install

### TCP/TLS or standalone codecs

Install the **core**. It contains the protocol types, codecs, Node/Peer runtime,
application helpers, and TCP/TLS transports. The root Go package is `diameter`.

```sh
go get github.com/skill215/diameterkit@v0.2.0-rc.1
```

### SCTP

Add the **SCTP adapter** only when your peer uses SCTP. It connects SCTP sockets
to the same core runtime and requires a supported Linux kernel.

```sh
go get github.com/skill215/diameterkit/adapters/sctp@v0.2.0-rc.1
```

The adapter depends on the core and `github.com/skill215/sctp`; Go resolves both
automatically. **The core does not depend on the adapter.** These are two modules
in one library, and both are pure Go. See [SCTP setup](adapters/sctp/README.md)
and the [version policy](doc/version-install-policy.md).

## Try it

Run a local GMSC/HSS routing exchange with no external service or database:

```sh
git clone https://github.com/skill215/diameterkit.git
cd diameterkit
CGO_ENABLED=0 go run ./examples/s6c
```

The example starts both endpoints over loopback TCP, sends a routing request,
prints the result, and shuts them down. It uses synthetic subscriber data.
Cloning `main` runs the current development checkout.

For a charging exchange, including command success with service rejection:

```sh
CGO_ENABLED=0 go run ./examples/ro
CGO_ENABLED=0 go run ./examples/ro -result 2001/4012
```

## Examples

| Build this | Start here |
| --- | --- |
| Query SMS routing data | [S6c GMSC/HSS](examples/s6c/README.md) |
| Charge for SMS and handle explicit recovery | [Ro IEC / ECUR / abort](examples/ro/README.md) |
| Forward MO/MT messages | [SGd/Gdd](examples/sgd/README.md) |
| Construct and parse selected Sh-Data | [Sh](examples/sh/README.md) |
| Use the named legacy accounting binding | [SCAP1](examples/scap1/README.md) |

These are selected protocol slices. The [support matrix](doc/diameterkit-support-matrix.md)
lists command coverage, compatibility bindings, and platform limits.

## Build your application

Start with the [endpoint configuration](examples/ro/config.go) and
[public lifecycle example](examples/ro/main.go). Choose a protocol bundle, declare
peer identities and permitted operations, set resource budgets, then connect
through `Node.Dial` or `Node.Serve`.

Use `Peer.Submit` and a completion port for outbound work; use `Peer.Receive`
and `Inbound.Respond` for inbound work. Standalone tools can use `codec` and the
typed protocol packages without a Node.

**A received answer is protocol evidence, not automatically business success.**
Inspect command and service results separately. A timeout after writing may mean
the peer already processed the request. See [runtime and outcome semantics](doc/runtime-basics.md)
before adding cancellation, reconnection, or recovery.

## Documentation

- **Learn:** [documentation index](doc/README.md), [protocol support](doc/diameterkit-support-matrix.md), [runtime basics](doc/runtime-basics.md).
- **Ship:** [version and installation policy](doc/version-install-policy.md), [release status](doc/releases/v0.2.0-rc.1.md), [remaining gaps](doc/remaining-gaps.md).
- **Contribute:** [CI and local checks](doc/release-checks.md), [measurement methodology](doc/load-checks.md), [specification and evidence](spec/README.md).

Licensed under [Apache-2.0](LICENSE). OpenAI Codex generated and revised this
library under maintainer direction. Historical model identities are unrecorded;
see [AI disclosure](AI_DISCLOSURE.md) and the Go files' `SPDX-AI-*` headers.

## Benchmarks

Existing measurements from **2026-11-16**, Go **1.26.10**, Linux/amd64 under
WSL2, **AMD Ryzen 7 6800H**. Median of five samples, **200 iterations per sample**,
`-cpu=1`, CGO=0. These are the recorded batch 92 baseline; subsequent checkout
measurements do not replace this historical baseline.

| Workload | Median ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Codec/base-dwr/Decode | 542.9 | 280 | 6 |
| Codec/base-dwr/AppendReuse | 275.3 | 224 | 2 |
| Codec/v-srr-01/Decode | 1,108 | 784 | 9 |
| Codec/v-srr-01/AppendReuse | 539.8 | 224 | 2 |
| Codec/v-srr-01/Canonical | 3,144 | 2,784 | 27 |
| RoRoundTrip | 105,728 | 27,744 | 244 |
| SCTPFrame/256 | 23,845 | 2,068 | 34 |
| SCTPFrame/4096 | 28,053 | 5,908 | 34 |

```sh
python3 spec/tools/release_check.py --suite bench
```

Connection/profile/fixture setup is excluded. Ro includes request construction,
identifiers, codec/network work and completion consumption. SCTP measures one-way
loopback send plus owned receive, not remote RTT. Append reuses output; Decode and
Canonical include owned allocations. **B/op is allocated heap memory, not copied
bytes, retained memory or RSS.** Sample medians are not per-request p50/p99, a
controlled speedup or a production throughput/SLO guarantee.

[All samples and environment](spec/release/performance-baseline.json) ·
[Workloads and resource caveats](doc/load-checks.md)
