<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated reporting-reason and requested-unit value diagnostics.
SPDX-AI-Date: 2026-11-19
-->

# Reporting-Reason and CC-Service-Specific-Units values

Batch 123 adds bounded 5004 diagnostics for two existing combined SMS template
constraints. These are named compatibility rules, not universal protocol rules.

| Template | Required condition at this rule | Selected error | Receive path |
| --- | --- | --- | --- |
| TERMINATION | Requested-Service-Unit absent; Used-Service-Unit and Reporting-Reason present | Reporting-Reason != 2 | 456 / 10415:872 |
| IEC or INITIAL | Requested-Service-Unit present; Used-Service-Unit and Reporting-Reason absent; Money and StandardMoney absent | CC-Service-Specific-Units != 1 | 456 / 437 / 417 |

Only the isolated value branch gains private evidence. Receive retains
`ErrUnsupported` and adds `ErrField` with the exact path above. Build and Send
retain their original single-key 456 or 437 path and classification. Money
request templates retain their existing behavior. Mixed presence/value errors
remain unsupported without this new diagnostic.

Earlier sequence, subscription, SMS/MMS, timestamp, action and interface errors
retain precedence. Later SMS status/discharge checks do not displace these unit
rules. Structural decoding failures retain their existing diagnostics, including
5014 for a malformed selected or counterpart field width.

## Original instance and bounds

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default. Private
evidence contains only a static discriminator, without retained request bytes or
persistent state. The consumer independently verifies the discriminator, field
path, request kind, template markers, unique ancestors, required counterparts,
forbidden fields, flags, widths and selected value. Complete containing-area
scans share one source AVP counter. Unrelated sibling payloads remain opaque.
Missing, duplicated, malformed or ambiguous selected/counterpart fields and
altered evidence cannot produce this quote or fall back to an untyped reply.

| Diagnostic | Owned quote | Output AVPs including Failed-AVP | Output depth | Minimal source fixture |
| --- | --- | --- | --- | --- |
| Reporting-Reason | 24 bytes | 3 | 2 | 7 AVPs, 100-byte frame, 40-byte largest field, depth 2 |
| CC-Service-Specific-Units | 32 bytes | 4 | 3 | 6 AVPs, 84-byte frame, 32-byte largest field, depth 2 |

Each quote consists of minimal standard Grouped ancestors and the exact original
16-byte leaf. Siblings are not copied. Source and output limits are checked
before allocation. The complete answer envelope and reply-byte budget are
validated separately. The unit quote requires output depth 3 even though the
minimal source fixture needs only depth 2.

Tests cover mode/presence/value matrices, signed and unsigned extremes, exact
instances, corruption at every selected level, counterpart checks, mixed errors,
ordering, shared budgets, ownership and TCP/TLS ordinary/emergency/default
handling. Invalid requests are not delivered; held resource charges, successful
follow-up and retirement are checked. Three seeds extend the existing
`FuzzInvalidQuote` target with independent raw-byte assertions.

## Reproduce

Use the pinned checkout toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(UnitValue|RejectUnitValue)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(UnitValue|RejectUnitValue)' -count=20 -timeout=10m
python3 spec/tools/release_check.py --suite all
```

## Version and migration

Combined advances to `0.1.0-slice.17`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for earlier recovery records; never
relabel records or bypass their fingerprints. The library remains pure Go with
application-owned persistence and recovery authorization.

This development change is not included in published v0.2.0-rc.1. Its
[receipt](releases/v0.2.0-rc.1.md) remains immutable historical evidence. Use commit
`6897bebc71e6ebd291b0b15573567a17b56a8259` to reproduce that published pair's
original checks. The planned v0.2.0-rc.2 pair remains unpublished and requires its
own publication and cold-install acceptance.

See [remaining gaps](remaining-gaps.md), [completion conditions](release-candidate-checklist.md)
and [batch evidence](../spec/evidence/batch-123-unit-values.json).
