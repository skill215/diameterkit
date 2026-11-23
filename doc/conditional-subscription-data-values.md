<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Repeated subscription data diagnostics, exact instances and validation order.
SPDX-AI-Date: 2026-11-23
-->

# Subscription data values

Batch 133 adds bounded 5004 diagnostics for the existing combined
Subscription-Id-Data rule: data must contain no NUL byte. Receive supplies static
named evidence when the failing group has no extra children. The error remains
`ErrField` at the existing vendor-zero path `443/444`; no `ErrUnsupported` is
added. Build and Send retain their original errors and paths. Empty and NUL-free
data do not fail this group constraint.

| Group or request state | Behavior |
| --- | --- |
| NUL-containing data, no extra children | Receive evidence for the data error |
| NUL-containing data and optional extra children | Existing ErrField, no new proof |
| Missing, duplicate, unknown mandatory or malformed child | Existing structural handling precedes the callback |
| Both subscription groups contain NUL | Quote the first failing group in wire order |
| Earlier group has extra children or a structural error | Keep that earlier failure; never select a later data error |
| Wrong subscription type pairing and NUL data | The data callback still precedes the later request pairing rule |

The proof stores no subscriber data or request. `BaseErrorFailures` enables
replies; the default remains `CloseOnFailure`. This is a compatibility constraint,
not a subscriber parser or a new wire type.

## Repeated groups and validation order

The error path has no occurrence index. The consumer completely scans the root,
retaining at most two Subscription-Id views in a fixed array and enforcing the
configured repetition limit. It then visits these groups in original order.
Every preceding group must be closed and structurally valid with NUL-free data;
the selected group must be closed with NUL-containing data. Each visited group
has exactly one correctly flagged four-byte type 450 and one data 444.

The first failure supplies the quote. Later group payloads remain opaque because
the trusted decoder stopped at the selected callback. Root framing, group flags,
the two-group cap and shared source budgets still apply, including to later
root AVPs. A third group denies this diagnostic rather than bypassing the cap.
Unrelated root payloads stay opaque. This consumer is not a replacement for the
trusted decoder that issued the named proof.

Any signed 32-bit type is structurally valid here. The request-level requirement
for two subscriptions in type order 0 then 1 remains later, as do sequence,
identity and template relations. Earlier wire failures retain priority, and
moving the bad subscription behind a malformed timestamp exposes that width
failure first. Earlier subscription data now supplies its own quote before a
later equipment/address data error; those existing regressions are updated.

## Quote and resource boundaries

For selected data of `n` bytes, the owned quote is **`16 + align4(n)` bytes**:
a minimal eight-byte group header and the entire original padded data AVP.
It includes the selected subscriber bytes, but no type or other subscription.
Padding is preserved and excluded from the NUL comparison. A successful quote
has at least one data byte and a minimum size of 20 bytes.

A single-group fixture needs three source AVPs, depth one, a frame of
`48 + align4(n)` bytes and a largest field of `20 + align4(n)` bytes. Failed-AVP
requires three output AVPs and depth two. With two groups, root scanning counts
both group headers; visiting the first or second failure requires four or six
source AVPs respectively. Two one-byte/two-byte data groups need an 84-byte
frame, a 24-byte largest field and repetition limit two. Source, quote and full
answer limits are independent gates. An insufficient budget closes without a
truncated AVP or partial answer.

Tests cover signed types, NUL positions, long data, original padding, both
instances, reordered and simultaneously failing groups, earlier/selected/later
malformed groups, exact and one-less limits, counters and ownership. TCP/TLS
ordinary and emergency paths cover IEC, INITIAL, TERMINATION, direct debit,
balance check and refund. They verify envelopes, non-delivery, held charges,
valid follow-up, short-budget closure and resource retirement.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SubscriptionData|RejectSubscriptionData|EquipmentData|RejectEquipmentData|AddressData|RejectAddressData)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SubscriptionData|RejectSubscriptionData|EquipmentData|RejectEquipmentData|AddressData|RejectAddressData)' -count=20 -timeout=15m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance runs twenty-round pure/race diagnostics and one affected fuzz
target. Hosted CI runs all seventeen bounded targets plus static, pure, race,
cross, stress and benchmarks. Archive digests, contents, log hashes, test counts
and execution budgets are bound to the implementation commit in the
[batch evidence](../spec/evidence/batch-133-subscription-data-values.json).

## Version and migration

Combined advances to `0.1.0-slice.27`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
stay unchanged. Keep original builds for older recovery records; never relabel
records or bypass fingerprints. Pure Go, application-owned persistence and
recovery authorization remain the boundary.

The frozen rc.1 receipt recorded in `6897bebc71e6ebd291b0b15573567a17b56a8259`
does not cover this checkout. The rc.2 pair remains planned and unpublished,
requiring fresh immutable publication/install acceptance. See
[remaining gaps](remaining-gaps.md) and
[completion conditions](release-candidate-checklist.md).
