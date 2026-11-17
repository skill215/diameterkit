# diameterkit

**Diameter for Go — typed protocol tools, bounded peers, and explicit application outcomes.**

Build clients, responders and simulators for selected **Sh, S6c, SGd/Gdd and Ro SMS**
procedures. Versioned protocol definitions generate concrete, reflection-free
codecs; TCP/TLS are included and Linux SCTP is a separate module.

**Go 1.26+** · [Tags](https://github.com/skill215/diameterkit/tags) ·
[Apache-2.0](LICENSE) · No third-party runtime dependencies in the core

> **AI-written:** OpenAI Codex generated and revised this library under maintainer
> direction. Complete historical model identities are unrecorded. See
> [AI_DISCLOSURE.md](AI_DISCLOSURE.md) and the Go files' `SPDX-AI-*` headers.

[Core or adapter?](#core-and-sctp-adapter) · [Quick start](#quick-start) · [Go API](#go-api) · [Support](#protocol-support) ·
[Examples](#examples) · [Documentation](#documentation) · [Benchmarks](#benchmarks)

**Published development candidates:** core `v0.1.0-rc.2` and SCTP adapter
`v0.1.0-rc.1`. Both passed independent cold-cache installation and hosted CI.
See the [publication record](doc/releases/v0.1.0-rc.md) for immutable commits and
acceptance evidence. v0 is not a stable-v1 or complete-standards promise.

## What it provides

- **Standalone protocol tools:** owned messages, concrete AVP types, explicit
  presence, normal/error envelopes, bounded decoding and canonical forms.
- **Node and Peer runtime:** capability negotiation, fixed terminal policies,
  inbound requests, completion ports, watchdogs, deadlines and bounded shutdown.
- **Resource admission:** explicit slot/byte budgets, backpressure, bounded
  identifiers and separate inbound, outbound, completion and duplicate budgets.
- **Application helpers:** S6c request construction and result interpretation;
  selected Ro workflows and explicitly authorized recovery records/segments.
- **Transports:** pure-Go TCP/TLS plus a separate pure-Go Linux SCTP adapter using
  `github.com/skill215/sctp v0.8.0`.

The library has no database, subscriber lookup, charging ledger, delivery engine
or durable recovery service. Applications own business decisions and persistence.

## Core and SCTP adapter

**Using TCP/TLS or only encoding and decoding messages? Install the core only.**
**Connecting over SCTP? Add the SCTP adapter on a supported Linux system.**
These are two Go modules in the same project, not separate services.

| Module | Responsibility | Dependencies | When to install |
| --- | --- | --- | --- |
| **Core** — `github.com/skill215/diameterkit` | Diameter codecs, handshake, requests and answers, timeouts, resource limits, application helpers and built-in TCP/TLS | Go standard library; no third-party runtime dependencies | All Diameter use, including standalone protocol tools and TCP/TLS connections |
| **SCTP adapter** — `github.com/skill215/diameterkit/adapters/sctp` | Connects SCTP sockets to the core transport interface; handles message boundaries, streams, PPID and transport notifications | Core plus `github.com/skill215/sctp` | Only when your application communicates with a Diameter peer over SCTP |

The dependency points **from the adapter to the core**. The core does not import
the adapter or the upstream SCTP library. The adapter uses the core for Diameter
protocol handling; it does not replace the core or add charging or SMS delivery
logic. Installing the adapter also resolves its required core and upstream SCTP
versions through Go modules.

This split keeps SCTP dependencies and platform requirements optional and lets
the modules version independently. Both are pure Go and provide no persistence.
SCTP additionally requires Linux kernel support; installing the Go module does
not enable SCTP in the kernel. See the [adapter guide](adapters/sctp/README.md)
for supported platforms and connection setup.

## Quick start

### Install

Use Go 1.26 or newer. The core package name is `diameter`.

For TCP/TLS or standalone protocol tools:

```sh
go get github.com/skill215/diameterkit@v0.1.0-rc.2
```

For SCTP, add the adapter. This version requires core `v0.1.0-rc.2` and upstream
SCTP `v0.8.0`; Go resolves these dependencies automatically:

```sh
go get github.com/skill215/diameterkit/adapters/sctp@v0.1.0-rc.1
```

These exact versions were verified with `GOWORK=off`, no `replace` directives,
empty caches and authenticated public downloads. See the
[release record](doc/releases/v0.1.0-rc.md) for reproduction commands.

### Run a local request and responder

```sh
git clone https://github.com/skill215/diameterkit.git
cd diameterkit
CGO_ENABLED=0 go run ./examples/ro
CGO_ENABLED=0 go run ./examples/ro -result 2001/4012
CGO_ENABLED=0 go run ./examples/s6c
```

Each command runs both endpoints over real loopback TCP in one process and exits.
The Ro example sends an SMS IEC DIRECT_DEBITING request to a synthetic OCS.
`2001/4012` demonstrates command success with service rejection. No real charging
or SMS delivery occurs. Endpoints use bounded cleanup and report released resources.

## Go API

Start from the complete [Ro endpoint configuration](examples/ro/config.go) and
[public lifecycle example](examples/ro/main.go). Compile an explicit profile,
configure capabilities, permitted directions and terminal identities, provision
resource budgets, then establish a connection using `Node.Serve` or `Node.Dial`.

| Task | Public API |
| --- | --- |
| Allocate a bounded session identity | `Node.NewSessionID` |
| Submit a request to an existing completion port | `Peer.Submit` |
| Receive and answer an inbound request | `Peer.Receive`, `Inbound.Respond` |
| Consume outcomes and release completion capacity | `CompletionPort.Receive` |
| Cancel an admitted exchange | `Node.Cancel` |
| Drain, close and join runtime work | `Node.Shutdown`, `Node.Close`, `Node.Wait` |
| Construct/interpret S6c routing | `s6c.NewSRR`, `s6c.InterpretSRA`, `client/s6c` |
| Prepare/submit explicitly authorized Ro recovery | `Node.PrepareRoRecoveryRequest`, `Node.SubmitRoRecovery` |

A Submit context controls admission; canceling it later does not cancel the
exchange. A Receive context stops waiting, not the admitted request. Unconsumed
completions retain their budget. `Wait` joins library work but does not consume
application results; close or drain completion ports as part of cleanup.

`AnswerReceived` and `Validation.Complete` describe protocol evidence. Inspect
command and service-level results separately. A timeout or cancellation after
writing does not prove that the remote operation did not occur. Reconnection does
not automatically replay business requests; recovery requires explicit policy,
identity, target and cumulative-budget evidence.

## Protocol support

| Binding | Selected capability | Important limit |
| --- | --- | --- |
| Base | CER/CEA, DWR/DWA, DPR/DPA and common envelopes | Not a general Diameter router/proxy |
| Sh | UDR/PUR/SNR/PNR; bounded Ref11/12/24/25 Sh-Data helpers | No HSS database or full XML schema validator |
| S6c | SRR/RDR/ALR and named ordinary/IMS/T4/Router slices | Context and role evidence remain explicit |
| SGd/Gdd | OFR/TFR with selected MO/MT and named extensions | TPDU is opaque; no delivery service |
| Standard Ro SMS | IEC debit, no-UPDATE ECUR, selected results/recovery | Not every DCCA/SMS action or full RFC 8506 |
| Generic DCCA / abort | Independent EVENT debit/balance/refund; ASR/ASA | Distinct bindings and application decisions |
| Legacy bindings | Named smsfwro, legacy420, combined, SCAP1 and SGd alert | Private mappings and target acceptance remain separate |

Only one conflicting 4/272 CCR binding can be loaded in a bundle. Profile slice
versions, protocol specification editions and module versions are different.
See the [support matrix](doc/diameterkit-support-matrix.md) and
[definition guide](doc/diameterkit-application-definitions.md).

TCP/TLS run without cgo. SCTP is native on supported Linux kernels; Windows and
386 use explicit unsupported stubs. Cross-compilation is not native network
acceptance. DRA adjacency never implicitly authorizes the ultimate application
endpoint. Local duplicate handling is not a durable charging ledger.

## Examples

| Task | Guide |
| --- | --- |
| SMS IEC, ECUR, ASR and explicit recovery | [Ro](examples/ro/README.md) |
| Query and interpret routing | [S6c](examples/s6c/README.md) |
| MO/MT forwarding with MME or SGSN roles | [SGd/Gdd](examples/sgd/README.md) |
| Bounded Ref24 payload construction and parsing | [Sh](examples/sh/README.md) |
| Named legacy accounting binding | [SCAP1](examples/scap1/README.md) |
| External public binding contracts | [Contracts](examples/contracts/README.md) |
| Independent SCTP transport | [Adapter](adapters/sctp/README.md) |

## Documentation

- [Version/API policy and immutable installation](doc/version-install-policy.md)
- [Publication status and release changes](doc/releases/v0.1.0-rc.md)
- [Dual-module checkout CI](doc/release-checks.md)
- [Bounded fuzz, pressure and measurement methodology](doc/load-checks.md)
- [S6c helper semantics](doc/s6c-routing-helpers.md)
- [Application-owned Ro handoff](doc/ro-recovery-handoff.md)
- [Combined binding: volatile duplicate acceptance](doc/combined-duplicate-acceptance.md)
- [Bounded missing-AVP diagnostics](doc/missing-avp-diagnostics.md)
- [First-excess AVP diagnostics](doc/excess-avp-diagnostics.md)
- [Invalid and unsupported AVP diagnostics](doc/invalid-avp-diagnostics.md)
- [Cross-peer and mixed-profile duplicate acceptance](doc/cross-peer-duplicate-acceptance.md)
- [Typed AVP length diagnostics](doc/length-avp-diagnostics.md)
- [Bounded Redirect 3006 handling](doc/redirect-handling.md)
- [Current remaining gaps](doc/remaining-gaps.md)
- [Machine-readable progress and historical evidence](spec/README.md)
- [AI authorship and review limits](AI_DISCLOSURE.md)

Run `python3 spec/tools/release_check.py --suite all` with the pinned `.go-version`
toolchain on Linux/amd64. This checks both modules, generated output, the reviewed
API, pure/race execution, cross-builds and finite load suites. Installation of
published versions is a separate gate. Detailed historical protocol/design
references retain their original language and are not new release claims.

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
