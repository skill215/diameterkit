# Support and acceptance boundaries

This publication guide distinguishes implementation, public consumer evidence
and actual deployment acceptance. It does not expand the normative specification
or the [machine support matrix](../spec/release/support-matrix.json).

## Selected profiles

| Binding | Slice version | Request application/command | Boundary |
| --- | --- | --- | --- |
| base | `0.1.0-slice.8` | CER 0/257, DWR 0/280, DPR 0/282 | Selected control/envelope rules; not a general proxy or redirect engine. |
| sh | `0.1.0-slice.11` | UDR 16777217/306, PUR 16777217/307, SNR 16777217/308, PNR 16777217/309 | Ref11/12/24/25 payload helpers; no full XSD, HSS database or subscription service. |
| s6c | `0.1.0-slice.23` | SRR 16777312/8388647, RDR 16777312/8388649, ALR 16777312/8388648 | Ordinary and named IMS/T4/Router slices; explicit purpose/role evidence; no delivery or SBI client. |
| sgd | `0.1.0-slice.15` | OFR 16777313/8388645, TFR 16777313/8388646 | Selected MO/MT and named extensions; TPDU opaque; no automatic forwarding/retry service. |
| ro | `0.1.0-slice.6` | CCR 4/272 | SMS IEC debit and no-UPDATE ECUR; no general standard SMS refund/balance sender. |
| dcca | `0.1.0-slice.1` | CCR 4/272 | Independent single-service EVENT debit/balance/refund; not full RFC8506 or SMS semantics. |
| roabort | `0.1.0-slice.1` | ASR 4/274 | Independent ASR/ASA; application owns session action; does not finish a CCR or authorize recovery. |
| smsfwro | `0.1.0-slice.6` | CCR 4/272 | Named legacy IEC/ECUR binding; private answer/token and actual target acceptance remain separate. |
| legacy420 | `0.1.0-slice.6` | CCR 4/272 | Explicit legacy420 debit/balance/refund; private1292 is not standard2022. |
| combined | `0.1.0-slice.25` | CCR 4/272 | Explicit five-template 4/272 binding; mixed2018 wire form requires target agreement. |
| scap1 | `0.1.0-slice.3` | ACR 0/271 | Named 0/271 accounting with Vendor193/Acct19302; private ACA/recovery not established. |
| sgdalert | `legacy-sgd-alert-v1.1` | ALR 16777313/8388648 | Delegated legacy SGd alert, not standard SGd ALR or implicit replay authorization. |

There are 11 generated definitions and one delegated profile. Overlapping command
bindings are counted separately. A bundle cannot load conflicting 4/272 definitions.
Protocol editions, slice versions and module versions are independent.

## Public API and platforms

The reviewed [API snapshot](../spec/release/public-api.json) covers 24 core packages
and one adapter package for Linux/amd64 and Windows/amd64 with CGO=0. Internal,
main and test-only packages are excluded. Snapshot drift requires review; equality
is not proof of semantic compatibility or a stable-v1 promise.

TCP/TLS are pure Go. SCTP uses the published upstream v0.8.0 module and is tested
natively on Linux/amd64. Linux/arm64 has cross-build evidence; Windows and 386 use
unsupported stubs. Build success does not prove native networking on that platform.

The examples use public APIs over real loopback TCP; the adapter suite uses real
kernel SCTP. Recovery consumption covers selected independent bindings, explicit
cross-target authorization and no-answer/import cases. Neither synthetic peers nor
application-provided evidence prove acceptance by a real HSS/MME/SGSN/DRA/OCS.

The [combined server-cache checks](combined-duplicate-acceptance.md) cover selected
normal/E result replay, per-arrival envelopes, cross-template identity conflicts,
bounded admission, expiry and Close on TCP/TLS. They do not establish durable
financial deduplication or full mixed-profile/cache coverage.

## Remaining gates

[Batch 99 duplicate acceptance](cross-peer-duplicate-acceptance.md) covers the
combined binding across two peers and four-profile shared/independent caches on
TCP/TLS. It adds verification without changing production code or API fingerprints.

The unreleased [missing-AVP extension](missing-avp-diagnostics.md) adds bounded
5005 placeholders; [first-excess diagnostics](excess-avp-diagnostics.md) add
generated 5009 instance copies. [Invalid/unsupported diagnostics](invalid-avp-diagnostics.md)
add selected 5004/5001 mappings. [Typed-length diagnostics](length-avp-diagnostics.md) add selected 5014 mappings.
[Redirect handling](redirect-handling.md) adds bounded advisory 3006 fields and
URI validation without automatic routing. Other mappings remain partial.
[Handwritten Proxy/VSAI diagnostics](manual-group-diagnostics.md) add selected
leaf mappings and cross-package evidence. [Batch 103](group-constraint-diagnostics.md)
adds VSAI conditional pairs and bounded Proxy omission examples. [Batch 104](minimum-group-diagnostics.md) adds selected whole-group minimum examples. [Batch 105](variable-missing-diagnostics.md) adds eleven explicit variable-field minima. [Batch 106](variable-group-diagnostics.md) adds typed variable-child group minima with explicit rule audits. Generator v22 and shared operations `ops9` change Bundle fingerprints; earlier recovery records require their original
build. The [current gap list](remaining-gaps.md) separates implementation,
verification and external evidence from historical batch descriptions.

- **Release:** exact available versions, commits, authenticated checksums and cold
  consumers are recorded in the [publication record](releases/v0.2.0-rc.1.md).
- **Private bindings:** original six-template inputs, private ACA/PSP/1292 mappings,
  420/2018 acceptance, deduplication domains and session continuity need named evidence.
- **Deployment:** peer/terminal identities, target recovery permission and real
  business migration belong to each deployment's acceptance.
- **Capacity:** [bounded mixed-load checks](load-checks.md#mixed-application-isolation)
  verify Sh/S6c/SGd/Ro class isolation while a handler or completion consumer stalls.
  Long-soak, deployment-specific multi-tenant and production saturation acceptance
  still require declared workloads and budgets.

All 24 broader requirements retain partial scope in [progress](../spec/progress.json).
A published development candidate does not close these independent gaps. The
library remains pure Go with no persistence boundary.

## Benchmarks

Read the final [README measurements](../README.md#benchmarks) and
[methodology](load-checks.md). Logical budgets, sampled heap and process RSS are
reported separately; none is a substitute for a deployment resource model.
