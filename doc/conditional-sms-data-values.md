<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated SMS address NUL diagnostics, exact quotes and validation order.
SPDX-AI-Date: 2026-11-23
-->

# SMS address data values

Batch 134 adds bounded 5004 diagnostics for the existing combined SMS group
C-string constraints. The selected fields are Originator-SCCP-Address (2008),
Recipient-SCCP-Address (2010) and SMSC-Address (2017), all Vendor 10415.
Receive supplies named evidence only when exactly one present address contains
NUL and the SMS group has no extra children.

| SMS group state | Behavior |
| --- | --- |
| Exactly one NUL-containing address, no extras | Receive evidence and exact selected leaf path |
| Two or three NUL-containing addresses | Original bare group ErrField; no new evidence |
| NUL-containing address with optional extra children | Original bare group ErrField; no new evidence |
| Missing, duplicate, unknown mandatory or malformed child | Existing structural handling before the callback |
| Empty or NUL-free addresses | No failure from these group constraints |

The error remains `ErrField`, with no `ErrUnsupported` added. The selected
Receive path is now precisely `873/2000/(2008 or 2010 or 2017)` instead of stopping
at the SMS group. Build/Send and mixed errors keep their original group path.
Static private evidence retains no address or request data. `BaseErrorFailures`
enables replies; the default remains `CloseOnFailure`.

## Validation order and wire boundaries

The original callback still checks Originator, Recipient and SMSC in that order.
It emits the selected proof only after the existing error is reached and all
three address fields satisfy the isolated-error condition. Structural scanning
and earlier nested interface errors remain earlier. Later SMS-Node, message-type,
client/address equality and template presence/exclusion rules stay later.
The consumer does not infer a template or operation from address bytes.

ClientAddressWire (2018), protocol bytes, user-data headers and status bytes
remain opaque to this rule, including when they contain NUL. The separate
request-level constraints on client wire bytes are unchanged. A data error in
an address present on a money request precedes its later forbidden-presence
rule, just as the old group callback did.

The consumer scans the root and both ancestors completely with shared counters.
The selected SMS group is closed: all fourteen known fields are unique,
correctly flagged Vendor 10415 instances; six required children must be present;
five scalar kinds must have four-byte data. Of the three address fields,
exactly the selected one must contain NUL. Interface payloads stay opaque to this
consumer because the trusted decoder validates them before issuing evidence.
Unrelated ancestor siblings remain opaque. Bare ErrField alone is insufficient.

## Quote and resource boundaries

For selected data of `n` bytes, the owned quote is **`36 + align4(n)` bytes**:
two minimal 12-byte group headers followed by the entire original padded address
AVP. It includes the selected address bytes, but no other address, client bytes,
interface or sibling data. Original padding is preserved and excluded from the
NUL comparison. A successful quote has at least one data byte and is at least
40 bytes long.

The minimal consumer fixture has nine scanned AVPs, source depth two, a frame
of `140 + align4(n)` bytes and a largest field of `108 + align4(n)` bytes. Its
interface payloads are opaque test fixtures; these sizes do not claim a fully
valid CCR or valid empty interface groups. Failed-AVP requires four output AVPs
and depth three. Source, quote and full-answer budgets are checked independently;
an insufficient budget closes without truncating an AVP or sending a partial reply.

Tests cover every subset of the other seven optional fields, all three address
presence/error combinations, NUL positions, long values, scalar widths, exact and
one-less limits, source counters, original padding, child order and ownership.
TCP/TLS ordinary and emergency acceptance covers all three fields on IEC, INITIAL,
TERMINATION, direct debit, balance and refund requests. It verifies exact envelopes,
non-delivery, held charges, follow-up, short-budget closure and resource retirement.
Moving the SMS group after a malformed MMS scalar preserves first-failure order.
Existing MMS and interface tests now quote an earlier isolated SMS address error.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSData|RejectSMSData|MMSKey|RejectMMSKey|InterfaceValue)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSData|RejectSMSData|MMSKey|RejectMMSKey|InterfaceValue)' -count=20 -timeout=15m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance runs twenty-round pure/race diagnostics and one affected fuzz
target. Hosted CI runs all seventeen bounded targets plus static, pure, race,
cross, stress and benchmarks. Archive digests, contents, log hashes, test counts
and execution budgets are bound to the implementation commit in the
[batch evidence](../spec/evidence/batch-134-sms-data-values.json).

## Version and migration

Combined advances to `0.1.0-slice.28`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
stay unchanged. Keep original builds for older recovery records; never relabel
records or bypass fingerprints. Pure Go, application-owned persistence and
recovery authorization remain the boundary.

The frozen rc.1 receipt recorded in `6897bebc71e6ebd291b0b15573567a17b56a8259`
does not cover this checkout. The rc.2 pair remains planned and unpublished,
requiring fresh immutable publication/install acceptance. See
[remaining gaps](remaining-gaps.md) and
[completion conditions](release-candidate-checklist.md).
