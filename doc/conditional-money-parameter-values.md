<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated Service-Identifier/Timezone diagnostics, raw counterparts and rule order.
SPDX-AI-Date: 2026-11-19
-->

# Service-Identifier and Timezone values

Batch 128 adds bounded 5004 diagnostics to the existing combined general and
balance request condition. Service-Identifier must equal 1 and Timezone must
equal 0. These are named compatibility constraints, not new interpretations of
standard service identifiers or timezone units.

| Same-condition state | Receive behavior |
| --- | --- |
| Service-Identifier != 1, Timezone == 0, presence conditions satisfied | Quote original Service-Identifier |
| Service-Identifier == 1, Timezone != 0, presence conditions satisfied | Quote original Timezone |
| Both values correct and presence conditions satisfied | Continue existing client/MMS validation |
| Both values wrong, a required counterpart absent, or a forbidden counterpart present | Retain the original bare ErrField without this diagnostic |

The presence conditions require StandardMoney, Service-Identifier, Timezone,
ClientAddressWire and MMS-Information, and forbid Money and Refund-Information.
Refund requests remain outside this slice. Build and Send retain the original
bare `ErrField`. Isolated Receive failures retain that classification and add an
exact field path: `439` for Service-Identifier or `10425:23` for Timezone.
No `ErrUnsupported` classification is added. Evidence is a static discriminator
and contains no request, subscriber, address or financial data.

## Validation order

Sequence, subscription, family markers, Requested-Action, unit-presence and the
money SMS-Node condition keep their earlier positions. Their failures cannot be
replaced by this new diagnostic. A structurally malformed selected or counterpart
field still receives its existing typed-length handling. Nested group rules,
including Address-Type, run before request-level conditions even when the group
is later on the wire.

Within the selected condition, exactly one wrong numeric value is required.
Later ClientAddressWire content checks and MMS field exclusions stay later:
wrong client bytes or a well-formed extra MMS field do not suppress the selected
numeric diagnostic. Tests distinguish those later rules from structural and
nested group validation. No main-decoder scan or rule is reordered.

## Original instances and bounds

`BaseErrorFailures` enables replies; `CloseOnFailure` remains the default. The
consumer requires the named evidence, `ErrField` and exact field path. Complete
source scans check:

- Unique EVENT type, request number 0, action 0 or 2, MSCC, Service-Information,
  Service-Identifier and Timezone. SMS family markers and Refund-Information
  are forbidden. The selected value must be wrong and the other value correct.
- One Requested-Service-Unit, no Used-Service-Unit or Reporting-Reason, one
  StandardMoney and no Money.
- Unique SMS-Information and MMS-Information; SMS-Node must equal 2, ClientAddressWire
  must be present, and the seven forbidden money SMS fields must be absent.

Every selected/counterpart instance must have the correct vendor and flags;
scalar widths are checked independently. A fixed seven-entry root scan and all
nested scans share one source AVP counter. Client bytes and MMS/StandardMoney
payloads remain opaque, preserving the rule boundary. The trusted decoder that
produced the evidence remains responsible for full request validation.

The owned quote is exactly **12 bytes** for Service-Identifier or **16 bytes**
for Timezone: the original numeric AVP, without ancestor or counterpart data.
Including Failed-AVP needs two output AVPs and depth one. The minimal consumer
fixture, with opaque empty client/MMS/StandardMoney payloads, uses thirteen source
AVPs, a 172-byte frame, a largest field of 52 bytes and source depth two. Source
and output budgets are checked before allocation; full answer-envelope and
reply-byte limits remain separate.

Tests cover signed extremes, both field orders, missing/duplicate instances,
wrong flags/vendor, malformed widths, all forbidden raw counterparts, exact
limits, one-less limits, shared counts and ownership. TCP/TLS ordinary, emergency
and default handling verifies exact envelopes and reply budgets, non-delivery,
held charges, valid follow-up and retirement. Mixed errors keep the existing
safe-close behavior when no reviewed diagnostic applies.

## Reproduce

Use the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyParameter|RejectMoneyParameter)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(MoneyParameter|RejectMoneyParameter)' -count=20 -timeout=10m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance counts one affected fuzz target. Hosted acceptance runs all
seventeen bounded targets plus static, pure, race, cross, stress and benchmark
suites. Archive digests, log hashes, test counts and fuzz execution limits are
verified against the exact implementation commit.

## Version and migration

Combined advances to `0.1.0-slice.22`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Preserve original builds for older recovery records; never
relabel records or bypass fingerprints. Pure Go and application-owned persistence
and recovery authorization remain the boundary.

The published rc.1 receipt remains frozen at its original recorded commits;
`6897bebc71e6ebd291b0b15573567a17b56a8259` contains its original checks. It does not
cover this development change. The rc.2 pair stays planned and unpublished and
still needs fresh immutable installation acceptance. See [remaining gaps](remaining-gaps.md),
[completion conditions](release-candidate-checklist.md) and
[batch evidence](../spec/evidence/batch-128-money-parameter-values.json).
