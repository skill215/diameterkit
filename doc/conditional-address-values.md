<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated combined Address-Type diagnostics and validation-order boundaries.
SPDX-AI-Date: 2026-11-19
-->

# Address-Type values

Batch 126 adds bounded 5004 diagnostics for the existing combined Address rule
that Address-Type must equal 1. This named compatibility constraint applies to
Originator-Address and Recipient-Address; it does not redefine the AVP's general
value domain.

| Selected Address group | Receive behavior |
| --- | --- |
| Type != 1, Address-Data contains no NUL, no extra children | Quote the original Type |
| Type == 1 | Continue existing data and extra-field validation |
| Wrong Type plus NUL-containing data or extra children | Retain the existing failure without this diagnostic |

The isolated error retains `ErrUnsupported` and adds `ErrField`. Complete paths
are `873 / 877 / 886 / 899` for Originator-Address and
`873 / 877 / 1201 / 899` for Recipient-Address, all with Vendor-ID 10415.
Build and Send retain their original classification and path. Evidence is a
static discriminator and holds no address or request data. Empty Address-Data
bytes remain valid for this rule, but the Address-Data AVP (897) must be present.

## Validation order

Isolation is local to the selected Address group. The other Address group is
not revalidated by the diagnostic consumer. The main decoder still determines
which group fails first: two wrong types select the first group, while a NUL
mixed with the first group's wrong type retains the original unsupported error.
Reordering the groups changes the first failure accordingly.

The group rule precedes request-level sequence, subscription, SMS marker,
timestamp and interface conditions. There is no new request-kind gate. Network
acceptance covers both addresses in IEC, INITIAL and TERMINATION and the
originator in general and balance requests. Templates without an Address group
gain no new diagnostic.

Missing or duplicate children, unknown Mandatory children and malformed scalar
widths retain their existing structural handling. An earlier malformed SMS field
wins; a malformed MMS field after the selected group does not. Moving that group
after the malformed MMS field reverses the outcome. These tests preserve actual
decoder order rather than selecting a more convenient error after decoding.

## Exact instance and bounds

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default. The
consumer verifies named evidence, both error classifications and the complete
path. Complete scans require unique ancestors, correct vendor/flags, exactly one
four-byte Type != 1 and one NUL-free Address-Data, and no extra group children.
All source scans share one AVP counter. Unrelated sibling payloads, including the
other address group, stay opaque. This is a trusted callback contract, not a
replacement for whole-request validation.

The owned quote is exactly **52 bytes**: three minimal twelve-byte vendor group
headers and the original sixteen-byte Type AVP. No address bytes or siblings are
copied. Including Failed-AVP requires five output AVPs and depth four. A minimal
source fixture with empty Address-Data requires five AVPs, an 84-byte frame and
a largest field of 52 bytes; longer addresses stay bounded by source limits and
do not enlarge the quote. Source and output limits are checked before allocation.
Full answer-envelope and reply-byte budgets remain separate checks.

Ambiguous instances, missing counterparts, malformed bytes, wrong flags/vendor,
mixed group errors or altered evidence cannot yield this quote or fall back to
an untyped reply. Tests cover signed extremes, variable data lengths, both child
orders, shared budgets, exact bytes and independent ownership. TCP/TLS ordinary,
emergency and default handling verifies non-delivery, held charges, precise
Session-ID/Proxy-Info, valid follow-up and retirement.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(AddressValue|RejectAddressValue)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(AddressValue|RejectAddressValue)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance includes one affected fuzz target. Hosted acceptance runs all
seventeen bounded targets, plus static, pure, race, cross, stress and benchmark
suites. Archive digests, log hashes and actual test/fuzz counts are checked
against the exact implementation commit.

## Version and migration

Combined advances to `0.1.0-slice.20`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

The development change is not included in published v0.2.0-rc.1. The planned
rc.2 pair remains unpublished and requires its own publication/install acceptance.
Use commit `6897bebc71e6ebd291b0b15573567a17b56a8259` for the published pair's
original checks. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-126-address-values.json).
