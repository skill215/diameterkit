<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Address-Data NUL diagnostics, exact quotes and validation order.
SPDX-AI-Date: 2026-11-23
-->

# Address data values

Batch 132 adds bounded 5004 diagnostics for the existing combined Originator-
and Recipient-Address rule: Address-Data must contain no NUL byte. Receive
supplies static named evidence only when Address-Type equals 1 and the selected
address group has no extra children. Empty and NUL-free data do not fail this
rule. This does not parse, normalize or reinterpret addresses.

| Selected address group | Receive behavior |
| --- | --- |
| Type 1, data contains NUL, no extras | Bounded data diagnostic |
| Wrong type, NUL-free data, no extras | Existing type diagnostic |
| Wrong type and NUL-containing data | Existing unsupported type error, no new evidence |
| Type 1, NUL-containing data and extra children | Existing ErrField, no new evidence |
| Missing, duplicate or malformed child | Existing structural handling |

The data error remains `ErrField` with full Vendor 10415 path
`873/877/(886 or 1201)/897`. No `ErrUnsupported` is added. Build and Send keep
their previous errors and paths. The private proof retains no address or request.
`BaseErrorFailures` enables replies; the default remains `CloseOnFailure`.

## Validation order and original instances

The callback checks type, data and extras in their original order. Structural
checks and earlier wire/group failures remain earlier; later request relations
are not pulled forward. Both address groups are tested together, including
reordered groups and an earlier mixed type/data error. An earlier address NUL
now supplies its own diagnostic before a later MMS SMSKey NUL; it never quotes
the later key. Existing MMS precedence and TCP/TLS acceptance cover this change.

The consumer completely scans each containing area using a shared source
counter. It requires unique correctly flagged Vendor 10415 ancestors and a
closed address group containing exactly one four-byte type 899 equal to 1 and
one data 897 containing NUL. Wrong flags, vendors, type widths, duplicates and
extra children deny this proof. Unrelated ancestor siblings and the other address
payload remain opaque to this consumer; the trusted decoder retains responsibility
for validation before issuing the proof.

## Quote and resource boundaries

For data of `n` bytes, the owned quote is **`48 + align4(n)` bytes**: three minimal
12-byte vendor group headers and the entire original padded data AVP. The quote
contains the selected address bytes, but no type or unrelated sibling data.
Padding is copied unchanged and excluded from the NUL comparison. A successful
diagnostic requires at least one data byte; its minimum quote is 52 bytes.

The minimal source fixture has five AVPs, depth three, a frame of
`84 + align4(n)` bytes and a largest field of `52 + align4(n)` bytes. Including
Failed-AVP requires five output AVPs and depth four. Source and quote limits
are checked before allocation; the full answer envelope has separate checks.
Insufficient budgets close without truncating the selected AVP.

Tests cover signed type extremes, NUL positions, empty and long data, child and
group order, original padding, source ownership, malformed controls and exact
and one-less budgets. TCP/TLS ordinary and emergency paths cover IEC, INITIAL,
TERMINATION and the existing general/balance Originator-Address. They verify
exact envelopes, non-delivery, held charges, valid follow-up and retirement.
Money templates have no MMS submission-time field; its width/order cases apply
to the three SMS templates. A short emergency reply budget verifies closure
without a partial answer and completion of the held request.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(AddressData|RejectAddressData|MMSKey|RejectMMSKey)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(AddressData|RejectAddressData|MMSKey|RejectMMSKey)' -count=20 -timeout=15m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

The expanded local diagnostic inventory runs twenty times in pure and race
modes. Local fuzz acceptance counts one affected target; hosted CI runs all
seventeen bounded targets plus static, pure, race, cross, stress and benchmarks.
Archive digests, contents, log hashes and execution budgets are checked against
the implementation commit. Initial failed tests and their correction are retained
in the [batch evidence](../spec/evidence/batch-132-address-data-values.json).

## Version and migration

Combined advances to `0.1.0-slice.26`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for old recovery records; never
relabel records or bypass fingerprints. Pure Go, application-owned persistence
and recovery authorization remain the boundary.

The frozen rc.1 receipt recorded in `6897bebc71e6ebd291b0b15573567a17b56a8259`
does not cover this checkout. The rc.2 pair remains planned and unpublished,
requiring fresh immutable publication/install evidence. See
[remaining gaps](remaining-gaps.md) and
[completion conditions](release-candidate-checklist.md).
