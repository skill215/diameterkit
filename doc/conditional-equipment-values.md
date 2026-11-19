<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated combined User-Equipment-Info-Type diagnostics and validation-order boundaries.
SPDX-AI-Date: 2026-11-19
-->

# User-Equipment-Info-Type values

Batch 125 adds bounded 5004 diagnostics for the existing combined group rule
that User-Equipment-Info-Type must equal 0. This is a named compatibility rule,
not a universal interpretation of the AVP's value domain.

| Group condition | Receive behavior |
| --- | --- |
| Type != 0, Value contains no NUL, no extra children | Quote the original Type |
| Type == 0 | Continue existing Value and extra-field validation |
| Wrong Type plus a NUL-containing Value or extra children | Retain the existing failure without this diagnostic |

The isolated error retains `ErrUnsupported` and adds `ErrField`. The complete
request path remains `458 / 459`; Build and Send keep their original error
classification and path. Evidence carries only a static discriminator, never the
equipment identity or request bytes. Empty Value bytes remain valid for this
rule, but the Value AVP itself must be present.

## Validation order

This rule belongs to the User-Equipment-Info group, before request-level template
selection and conditions. It is not gated by a new request-kind or SMS-marker
check. A bad group Type can therefore precede a bad request sequence, subscription
relation, timestamp, interface or unit condition. Those constraints still run
when the group is valid. Templates omitting the group gain no new diagnostic.

Structural failures retain their own wire-order behavior. Missing or duplicate
group children, malformed Type width and unknown Mandatory children keep their
existing diagnostics. An earlier malformed subscription field wins over this
rule; a later malformed SMS field does not. Moving the equipment group after
that malformed field changes which decoder failure is encountered first. Tests
explicitly exercise both orders; no scan reorders the main decoder.

## Original instance and bounds

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default.
The consumer verifies the named evidence, error classification, exact path and
unique group. Its complete child scan requires exactly one Type and one Value,
correct flags, a four-byte nonzero Integer32 and a NUL-free Value. Every other
child is rejected. Root siblings remain opaque. Root and child scans share a
source AVP counter; this is a trusted callback contract, not independent
validation of the entire request.

The owned quote is exactly **20 bytes**: a minimal eight-byte Grouped header and
the original twelve-byte Type AVP. It never copies the equipment identity or
siblings. Including Failed-AVP requires three output AVPs and depth two. The
minimal source fixture, with empty Value bytes, requires three AVPs, a 48-byte
frame and a 20-byte largest field. Longer Value inputs remain bounded by source
limits and do not enlarge the output. Source and output limits precede allocation;
complete answer-envelope and reply-byte checks remain separate.

Missing, duplicate or ambiguous selected/counterpart instances, malformed headers
or scalar width, mixed group faults and altered evidence cannot produce this
quote or fall back to an untyped reply. Tests cover signed extremes, every mode,
variable Value lengths, both child orders, mixed failures, structural precedence,
shared budgets, exact bytes and output ownership. TCP/TLS ordinary, emergency
and default handling checks non-delivery, held charges, follow-up and retirement.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(EquipmentValue|RejectEquipmentValue)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(EquipmentValue|RejectEquipmentValue)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance includes the affected fuzz target; hosted release acceptance
runs all seventeen bounded targets, plus static, pure, race, cross, stress and
benchmark suites. Their archived contents, digests and log hashes are checked
against the exact implementation commit.

## Version and migration

Combined advances to `0.1.0-slice.19`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

The development change is not included in published v0.2.0-rc.1. The planned
rc.2 pair remains unpublished and requires its own publication/install acceptance.
Use commit `6897bebc71e6ebd291b0b15573567a17b56a8259` for the published pair's
original checks. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-125-equipment-values.json).
