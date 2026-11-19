<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Money-template SMS-Node diagnostics, mixed errors and validation order.
SPDX-AI-Date: 2026-11-19
-->

# Money-template SMS-Node values

Batch 127 adds bounded 5004 diagnostics for the existing combined money rule
that SMS-Node must equal 2. The selected requests are general direct debit,
balance check and refund. This named compatibility rule does not redefine the
AVP's general value domain or select a different wire decoder.

| Money SMS condition | Receive behavior |
| --- | --- |
| SMS-Node != 2 and all seven forbidden fields absent | Quote the original SMS-Node |
| SMS-Node == 2 and all seven fields absent | Continue existing money conditions |
| Any forbidden field present, including with a wrong node | Retain the existing unsupported failure without this diagnostic |

The forbidden fields are Originator-SCCP-Address (2008), Recipient-SCCP-Address
(2010), SMSC-Address (2017), Reply-Path-Requested (2011), SM-User-Data-Header
(2015), SM-Status (2014) and SM-Discharge-Time (2012), all Vendor-ID 10415.
SM-Message-Type is not constrained by this branch; its existing behavior remains
unchanged, including nonzero values.

Build and Send retain `ErrUnsupported` at vendor path `2000`. The isolated
Receive branch retains `ErrUnsupported`, adds `ErrField` and supplies the precise
vendor path `873 / 2000 / 2016`. Evidence is a static discriminator without
request, address, subscriber or financial data.

## Validation order

Sequence and subscription validation precede family selection. Money requests
require EVENT, no Multiple-Services-Indicator and no User-Equipment-Info.
Requested-Action validation and requested/used/reason unit-presence checks precede
the node rule. Those earlier failures retain priority. A forbidden SMS field
makes the node condition mixed and suppresses this new evidence.

Later money constraints retain their position: a wrong node precedes a missing
refund field, wrong Service-Identifier, wrong client bytes or missing optional
MMS group. Structural decoding and nested group rules run before request-level
conditions even when their fields occur later on the wire. Malformed SMS scalar
widths and a wrong Address-Type therefore keep their earlier diagnostics. Tests
check the exact selected error and quote, not only whether decoding fails.

## Raw evidence and budgets

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default. The
consumer checks named evidence, both error classifications and the full path.
One shared source AVP counter covers complete scans of the root, MSCC,
Service-Information and SMS-Information areas. It requires:

- Unique EVENT type, request number 0, action in 0..2, MSCC and Service-Information;
  Multiple-Services-Indicator and User-Equipment-Info must be absent.
- One Requested-Service-Unit and no Used-Service-Unit or Reporting-Reason.
- Unique SMS-Information and four-byte SMS-Node != 2, correct vendor/flags and
  absence of all seven forbidden SMS fields.

Required scalar widths and counterpart values are checked against original raw
instances. Requested-unit payloads, subscriptions, financial fields and unrelated
siblings remain opaque. This is a trusted decoder callback contract, not whole-
request validation. Mixed, ambiguous or malformed selected evidence cannot
produce this quote or fall back to an untyped answer.

The owned quote is exactly **40 bytes**: two minimal twelve-byte vendor headers
and the original sixteen-byte SMS-Node AVP. No addresses or unit contents are
copied. Including Failed-AVP requires four output AVPs and depth three. The
minimal consumer fixture, with an opaque empty requested-unit payload, requires
eight source AVPs, a 112-byte frame and an output field budget of 40 bytes.
Source/output limits precede allocation; full answer and reply budgets remain
separate. Tests cover exact limits, one-less budgets, root/child shared counts,
missing/duplicate instances, malformed widths, changed flags/vendor and ownership.
TCP/TLS tests cover ordinary, emergency and default handling, preserved envelopes,
non-delivery, held charges, valid follow-up and retirement.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyNode|RejectMoneyNode)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyNode|RejectMoneyNode)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance counts one affected fuzz target; hosted acceptance runs all
seventeen bounded targets plus static, pure, race, cross, stress and benchmark
suites. Exact-commit archive digests, log hashes and test/fuzz counts are verified
separately from the workflow's success status.

## Version and migration

Combined advances to `0.1.0-slice.21`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for earlier recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

The published rc.1 pair remains frozen. Its original checks are recorded at
`6897bebc71e6ebd291b0b15573567a17b56a8259`; this development change is not covered
by that receipt. The rc.2 pair stays planned and unpublished, with fresh immutable
installation acceptance still required. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-127-money-node-values.json).
