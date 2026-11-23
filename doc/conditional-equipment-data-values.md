<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated equipment value NUL diagnostics, exact quotes and validation order.
SPDX-AI-Date: 2026-11-23
-->

# Equipment identity values

Batch 131 adds bounded 5004 diagnostics for the existing combined
User-Equipment-Info-Value rule: the value must contain no NUL byte. Receive
provides named evidence only when User-Equipment-Info-Type equals 0 and the
selected group has no extra children. Empty and NUL-free values do not fail
this rule; the later subscription/identity relation remains separate.

| Group state | Receive behavior |
| --- | --- |
| Type 0, value contains NUL, no extra children | Bounded value diagnostic |
| Wrong type, NUL-free value, no extra children | Existing type diagnostic |
| Wrong type and value contains NUL | Existing unsupported type error, no new evidence |
| Type 0, value contains NUL and extra children exist | Existing ErrField, no new evidence |
| Structurally missing, duplicated or malformed child | Existing structural handling |

The value error remains `ErrField`, with the existing full path `458/460`.
No `ErrUnsupported` is added. Build and Send retain their previous errors and
paths. The private evidence is static and retains no identity or request data.
This is a named compatibility constraint, not a new C-string wire type or an
identity parser. `BaseErrorFailures` enables replies; the default remains
`CloseOnFailure`.

## Validation order and original instances

The existing group callback checks the type before the value, then extra fields.
The complete structural group scan still precedes that callback. Earlier wire
fields or groups retain priority; later root sequence, subscription, template
and identity-relation conditions are not pulled forward. Reordering the equipment
group after a malformed later group still exposes that earlier wire failure.
Both the old type diagnostic and the new value diagnostic are regression tested.

The consumer scans the entire root and selected group with one shared source
counter. It requires exactly one correctly flagged, vendor-zero group 458,
exactly one four-byte type 459 equal to 0, and exactly one value 460 containing
NUL. Unknown or duplicate group children, wrong flags/vendors and invalid type
widths deny this diagnostic. Unrelated root siblings remain opaque; the trusted
decoder that issued the evidence remains responsible for full validation.

## Quote and resource boundaries

For a value of `n` bytes, the owned quote is exactly **`16 + align4(n)` bytes**:
a minimal group header followed by the entire original value AVP, including
original padding. It contains the selected equipment value, so it can include
identity bytes; it contains no type, subscription or unrelated sibling data.
Padding is preserved but does not participate in the NUL comparison. A successful
diagnostic has at least one value byte and a minimum quote of 20 bytes.

The minimal source fixture has three AVPs, depth one, a frame of
`48 + align4(n)` bytes and a largest field of `20 + align4(n)` bytes. Including
Failed-AVP requires three output AVPs and depth two. Source and output budgets
apply together before allocation; full envelope/reply limits are checked
separately. Insufficient budgets close the connection without truncation.

Tests cover NUL positions, empty/short/long values, signed type extremes,
child ordering, original padding, unique-instance controls, exact and one-less
limits, source counters and ownership. TCP/TLS ordinary and emergency acceptance
covers IEC, INITIAL and TERMINATION, exact envelopes, non-delivery, held charges,
valid follow-up and retirement. Default and mixed errors keep safe-close behavior.
A separate short emergency reply-budget test verifies no partial answer and
subsequent release of the held request on the closed connection.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(EquipmentData|RejectEquipmentData)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(EquipmentData|RejectEquipmentData)' -count=20 -timeout=15m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

The expanded local 20-round diagnostic inventory uses a 15-minute total test
ceiling. Release-suite and counted-fuzz budgets are unchanged. Local acceptance
counts one affected fuzz target; hosted CI runs all seventeen bounded targets
plus static, pure, race, cross, stress and benchmark suites. Archive digests,
log hashes, test counts and execution limits are checked for the implementation
commit.

## Version and migration

Combined advances to `0.1.0-slice.25`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for older recovery records; never
relabel records or bypass fingerprints. Pure Go, application-owned persistence
and recovery authorization remain the boundary.

The rc.1 receipt stays frozen at its original commits recorded in
`6897bebc71e6ebd291b0b15573567a17b56a8259`; it does not cover this checkout.
The rc.2 pair remains planned and unpublished and needs fresh immutable install
acceptance. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-131-equipment-data-values.json).
