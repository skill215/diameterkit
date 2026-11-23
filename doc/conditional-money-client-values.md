<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Bounded client-wire value diagnostics, original padding and validation order.
SPDX-AI-Date: 2026-11-23
-->

# Money request client bytes

Batch 129 adds bounded 5004 diagnostics for the existing combined general and
balance request rule: ClientAddressWire must equal the six bytes
`00 01 7f 00 00 01`. Every other payload, including empty, shorter and longer
values, fails this named compatibility rule. The field remains opaque bytes;
this change does not introduce an Address parser or general address semantics.

| Condition | Behavior |
| --- | --- |
| Exact six-byte value | Continue existing MMS validation |
| Wrong bytes after all preceding conditions succeed | Receive supplies bounded 5004 evidence |
| Earlier condition or structural/group validation fails | Preserve that earlier failure |
| Client field absent or another required presence condition fails | Preserve existing mixed/presence handling |

Build and Send retain `ErrUnsupported` and their vendor 2018 path. Receive
retains `ErrUnsupported`, adds `ErrField` and names the full vendor 10415 path
`873/2000/2018`. The private evidence is a static discriminator with no retained
request or financial state. Refund and SMS request families are outside this
slice. `BaseErrorFailures` enables answers; `CloseOnFailure` remains the default.

## Validation order and exact instances

Sequence, subscriptions, family markers, Requested-Action, unit/money presence,
SMS-Node, forbidden SMS fields, Service-Identifier/Timezone and all same-condition
presence checks retain priority. Nested structural and group rules, including
Address-Type, still run before request-level rules. Later well-formed MMS field
exclusions stay later and cannot suppress the client-value diagnostic.

The consumer requires the named evidence, both error classifications and the
exact full path. Complete scans independently verify unique EVENT type, request
number 0, action 0 or 2, Service-Identifier 1, Timezone 0, MSCC, Service-Information,
Requested-Service-Unit and StandardMoney. Family markers, Used-Service-Unit,
Reporting-Reason, Money and Refund-Information must be absent. SMS-Information
and MMS-Information must be present, SMS-Node must equal 2, and the seven forbidden
money SMS fields must be absent. Every selected/counterpart instance has its
vendor, flags and applicable scalar width checked with shared source counters.
MMS and StandardMoney payloads remain opaque to this trusted callback; the main
decoder remains responsible for full request validation.

## Variable quote and resource limits

For a client payload of `n` bytes, the owned quote is exactly
**`36 + align4(n)` bytes**: two minimal vendor group headers followed by the
entire original ClientAddressWire AVP, including its original padding. It contains
no sibling, subscription, MMS or financial payload. Nonzero original padding is
preserved, but padding does not participate in the six-byte value comparison.

The entire quote must fit before allocation. Source frame, field, count and depth
limits remain enforced, followed by the full answer envelope and reply-byte
limits. Including Failed-AVP requires four output AVPs and depth three. The minimal
consumer fixture has thirteen source AVPs, a frame of `172 + align4(n)` bytes and
a largest field of `52 + align4(n)` bytes; its source depth is two. Source and
output requirements apply together. Insufficient budgets close the connection;
the consumer never truncates the selected AVP or sends a partial answer.

Tests cover empty/short/long payloads, single-byte mutations, alignment and
padding, reordered roots, exact limits and one-less limits, duplicate/missing
counterparts, flags/vendor/width errors and ownership. TCP/TLS ordinary and
emergency tests check exact envelopes, non-delivery, held charges, valid follow-up
and retirement. A separate emergency reply-budget test holds a valid request,
rejects an oversized diagnostic without sending bytes, then verifies resource
release on the closed connection.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyClient|RejectMoneyClient)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyClient|RejectMoneyClient)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance counts one affected fuzz target; hosted CI runs all seventeen
bounded targets and the static, pure, race, cross, stress and benchmark suites.
Archive digests, individual log hashes, test inventories and execution bounds
are verified against the exact implementation commit. Initial failed test runs
are retained separately from corrected acceptance.

## Version and migration

Combined advances to `0.1.0-slice.23`, updating its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for older recovery records; never
relabel records or bypass fingerprints. Pure Go, application-owned persistence
and recovery authorization remain the boundary.

The published rc.1 receipt is frozen at its original commits, recorded in
`6897bebc71e6ebd291b0b15573567a17b56a8259`; it does not cover this checkout. The
rc.2 pair remains planned and unpublished and needs fresh immutable installation
acceptance. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-129-money-client-values.json).
