<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated combined Multiple-Services-Indicator diagnostics and prior CI evidence closure.
SPDX-AI-Date: 2026-11-19
-->

# Multiple-Services-Indicator values

Batch 124 adds bounded 5004 diagnostics for the existing combined SMS template
rule that Multiple-Services-Indicator must equal 1. The selected IEC, INITIAL and
TERMINATION branch requires both Multiple-Services-Indicator and User-Equipment-Info
to be present, with Service-Identifier, Timezone and Refund-Information absent.
This is a named compatibility constraint, not a universal protocol rule.

| Condition at this rule | Receive behavior |
| --- | --- |
| Valid presence combination, indicator != 1 | Quote the original indicator |
| Valid presence combination, indicator == 1 | Continue existing validation |
| Missing marker or any forbidden field | Retain the existing failure without this diagnostic |

The selected error retains `ErrField`, adds the exact path `455`, and carries a
private static discriminator. Build and Send retain the original bare `ErrField`.
Money templates retain their existing marker-based selection and behavior.
Earlier sequence and subscription checks retain priority. Later identity,
SMS/MMS, timestamp, interface and unit rules do not displace this earlier rule.
Structural decode errors still precede root rules, including malformed field
widths. Mixed faults within the presence condition do not acquire evidence.

## Exact instance and bounds

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default.
The consumer verifies the named evidence, precise path, unique request kind,
indicator and User-Equipment-Info, flags, Integer32 widths and indicator value.
It rejects all three forbidden peers. The entire root AVP area is scanned once
with a shared source counter. The equipment marker's payload and unrelated
siblings stay opaque; the schema decoder validates their structure before the
root rule. Evidence is a trusted validation callback contract, not independent
validation of the entire request.

The owned quote is exactly **12 bytes**, copied from the selected original AVP.
It contains no ancestors or siblings. Including Failed-AVP needs two output
AVPs and depth one. The minimal consumer source fixture needs three AVPs,
a 52-byte frame and a 12-byte largest field. Source and output bounds precede
allocation; full answer-envelope and reply-byte limits are checked separately.
Missing, duplicate or ambiguous required instances, malformed headers or scalar
widths, and altered evidence cannot produce a quote or fall back to an untyped reply.

Tests cover signed extremes, every presence combination and mode, exact bytes,
corruption, shared budgets, mixed-error closure, first-failure ordering and
TCP/TLS ordinary/emergency/default paths. Invalid requests are not delivered;
held resource charges, successful follow-up and retirement are checked. Three
seeds extend the existing bounded `FuzzInvalidQuote` target.

## Reproduce

Use the pinned checkout toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(IndicatorValue|RejectIndicatorValue|SMSValuePrecedence)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(IndicatorValue|RejectIndicatorValue|SMSValuePrecedence)' -count=20 -timeout=10m
python3 spec/tools/release_check.py --suite all
```

## Evidence and version

Batch 123's seven hosted archives have now been downloaded and verified against
their API digests, step-log hashes, test counts and bounded fuzz receipts. The
[follow-up receipt](../spec/evidence/batch-124-prior-ci.json) closes that access
limitation; the original batch record remains historical evidence.

Combined advances to `0.1.0-slice.18`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
are unchanged. Preserve original builds for earlier recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

This development change is not in published v0.2.0-rc.1. The planned rc.2 pair
remains unpublished and needs its own publication/install acceptance. Use commit
`6897bebc71e6ebd291b0b15573567a17b56a8259` for the published pair's original checks.
See [remaining gaps](remaining-gaps.md), [completion conditions](release-candidate-checklist.md)
and [batch evidence](../spec/evidence/batch-124-indicator-values.json).
