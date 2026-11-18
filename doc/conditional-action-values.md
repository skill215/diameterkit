<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Conditional EVENT action-value diagnostics and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# EVENT Requested-Action value diagnostics

Batch 118 adds bounded **5004** replies for two existing combined Receive
conditions. With `BaseErrorFailures` enabled, the reply quotes the original
Requested-Action AVP and its value; an invalid request is never delivered to the
application. The default failure policy still closes the connection.

| Existing branch | Value | Receive diagnostic |
| --- | --- | --- |
| IEC EVENT | 0 | Existing template validation continues |
| IEC EVENT | Any nonzero UInt32 | Selected 5004 |
| Money EVENT | 0, 1 or 2 | Existing operation-specific validation continues |
| Money EVENT | Greater than 2 | Selected 5004 |
| EVENT action absent | Absent | Existing [5005 omission](conditional-action-diagnostics.md) |
| INITIAL/TERMINATION action present | Any value | Existing unsupported failure; no new reply mapping |

This changes error reporting for the selected profile, not the accepted actions
or operation selection. A value valid for one branch does not authorize another
branch. Other cross-field mismatches, wrong interfaces and forbidden presence
remain separate. No debit, balance check, refund, retry or persistent state is
created by this diagnostic.

## Evidence and bounds

The private producer contains only one of two static rule names. It retains no
action value, request bytes or financial data. Receive keeps
`errors.Is(err, validation.ErrUnsupported)` and additionally reports
`validation.ErrField` at key 436. Build/Send keep their prior classification.

The consumer rechecks the complete top-level AVP area before allocation:

- Exactly one CC-Request-Type equal to EVENT and one four-byte Requested-Action.
- IEC's Multiple-Services-Indicator and User-Equipment-Info markers, or their
  absence for the money branch; duplicate selectors are rejected.
- The selected range violation, exact one-key public error path and named proof.
- Requested-Action's standard key and Mandatory flags, source framing, scan
  limits and diagnostic budgets.

This is a narrow proof consumer, not another full request validator. The normal
decoder and earlier validation stages establish all other request semantics.
Bare sentinels and unknown proof names do not authorize a reply.

The owned quote is **12 bytes**, copied from the original instance. Including
Failed-AVP it needs two AVPs and depth one. Source frame/field/count budgets and
complete reply budgets apply separately. Missing, duplicate, wrong-length,
altered-path, wrong-branch and insufficient-budget inputs cannot produce this
value quote. No partial quote or substitute value is returned.

## Ordering and compatibility

Structural decoding, typed-length failures and earlier subset rules keep their
order. Sequence/subscriptions precede both action branches. IEC's SMS/MMS and
timestamp checks precede its action condition; Originator checks follow it.
Consequently an invalid action can be reported before a later Originator
omission. The earlier action remains the first fault; this batch does not
construct a combined multi-fault diagnosis.

Typed-length 5014, duplicate 5009 and missing-action 5005 keep their separate
evidence paths. A reply is still subject to trusted Session/Proxy extraction,
the full error grammar and ordinary/emergency resource reservations.

Combined advances to `0.1.0-slice.13` and its definition digest and dependent
Bundle fingerprints change. Public signatures/layouts, generator v23 and ops9
remain unchanged. Recovery records tied to older fingerprints require their
original build. Published tags remain immutable; new publication/cold
installation is a separate gate. Libraries remain pure Go with no persistence.

## Acceptance and reproduction

The checks cover Build/Send/Receive, all six templates and boundary values;
raw-wire value mapping; byte ownership and exact/below budgets; malformed and
ambiguous proof/source controls; first-failure order; and TCP/TLS ordinary and
emergency replies, envelope preservation, non-delivery and resource retirement.
The existing `FuzzInvalidQuote` gains action seeds and assertions without
removing its earlier checks. The seventeen-target counted gate retains its
requested work and timeouts; a separately documented [count correction](count-bounded-fuzz.md#batch-118-count-correction)
accepts only the pinned Go minimization overrun of at most62 calls per target.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(ActionValue|RejectActionValue|ActionMissing|RejectActionMissing|OriginatorMissing|RejectOriginatorMissing)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(ActionValue|RejectActionValue|ActionMissing|RejectActionMissing|OriginatorMissing|RejectOriginatorMissing)' -count=20 -timeout=10m
CGO_ENABLED=0 GOMAXPROCS=2 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

[Batch 118 evidence](../spec/evidence/batch-118-action-values.json) records actual
outcomes, including the initial obsolete close expectation and its explicit
first-failure regression update. See [remaining gaps](remaining-gaps.md).
