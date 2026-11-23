<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated MMS SMSKey NUL diagnostics, exact instances and validation order.
SPDX-AI-Date: 2026-11-23
-->

# MMS SMSKey values

Batch 130 adds bounded 5004 diagnostics for the existing combined MMS rule:
a present SMSKey must contain no NUL byte. Receive supplies named evidence only
when the selected MMS group has no extra children. Empty and absent keys do not
fail this value rule; existing request-level presence rules still apply.

| Selected group state | Behavior |
| --- | --- |
| Present key contains NUL, no extra children | Receive provides bounded 5004 evidence |
| Key contains NUL and extra children are present | Preserve the original ErrField without this evidence |
| Key contains no NUL, including an empty value | Continue existing validation |
| Structural or nested address validation fails | Preserve that earlier failure |

The error remains `ErrField`; no `ErrUnsupported` is added. Build and Send retain
their existing error and path. Receive preserves the full vendor 10415 path
`873/877/1210`. Evidence is a static discriminator and retains no key, address or
request. This is a named compatibility rule, not a new text encoding or C-string
wire type. `BaseErrorFailures` enables replies; the default remains `CloseOnFailure`.

## Ordering and trusted evidence

The existing MMS group callback runs after its complete structural scan and
nested address validation, before the enclosing service and request rules.
Malformed scalar lengths, duplicate/missing required instances and invalid
address groups retain priority even when the key appears earlier on the wire.
Request-level sequence, subscription and SMS/interface conditions remain later.
General and balance requests can therefore report a malformed received key
before their later prohibition on that field; a well-formed forbidden key still
receives the existing unsupported-field behavior. No rule is reordered.

The consumer scans the complete root and Service-Information areas and requires
one correctly flagged selected ancestor each. It closes the selected MMS group:
Originator-Address and SMSKey must each occur once; the five other known children
may each occur at most once. Every child must have the correct vendor and flags;
the four optional scalar fields must have their existing four-byte width.
Unknown children and duplicate instances deny this diagnostic. The original key
must contain NUL. All scans share a source AVP counter. Address payloads and
unrelated ancestor siblings remain opaque to this trusted callback; the decoder
that issued the evidence remains responsible for their full validation.

## Owned quote and bounds

A key of `n` bytes produces exactly **`36 + align4(n)` bytes**: two minimal vendor
group headers and the entire original SMSKey AVP, including its original padding.
The quote retains no address, sibling, subscription or financial data. Padding is
preserved but does not participate in the NUL comparison. An accepted diagnostic
has at least one payload byte, so its minimum quote is 40 bytes.

Source and output budgets must both fit before allocation. Including Failed-AVP
requires four output AVPs and depth three. The minimal consumer fixture, with an
opaque empty Originator-Address, has four source AVPs, source depth two, a frame of
`68 + align4(n)` bytes and a largest field of `36 + align4(n)` bytes. Full answer
envelope and reply-byte limits are checked separately. A short budget closes the
connection; no truncated key or partial answer is sent.

Tests cover all 32 optional-child combinations, NUL positions, empty/short/long
values, original padding and child order, missing/duplicate/vendor/flag/width
controls, exact and one-less limits, shared counts and owned copies. TCP/TLS
ordinary and emergency paths cover five request families, exact envelopes,
non-delivery, held charges, valid follow-up and retirement. Default and mixed
failures close safely; a separate emergency budget test verifies no truncated
answer and subsequent release of a held request on the closed connection.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MMSKey|RejectMMSKey)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MMSKey|RejectMMSKey)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance runs one affected fuzz target; hosted CI runs all seventeen
bounded targets plus static, pure, race, cross, stress and benchmark suites.
Archive digests, log hashes, test counts and execution bounds are verified for
the exact implementation commit.

## Version and migration

Combined advances to `0.1.0-slice.24`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for older recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

The published rc.1 receipt remains frozen at the commits recorded in
`6897bebc71e6ebd291b0b15573567a17b56a8259` and does not cover this checkout. The
rc.2 pair is planned and unpublished and needs fresh immutable installation
acceptance. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-130-mms-key-values.json).
