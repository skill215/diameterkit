<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Repeated-group omission diagnostics and reproducible acceptance.
SPDX-AI-Date: 2026-11-18
-->

# Repeated-group omission diagnostics

Batch 109 adds bounded 5005 diagnostics for the combined binding's existing
two-subscription rule. The wire definition still requires at least one group;
the handwritten subset still requires the ordered types 0 and 1.

With neither group present, the decoder's existing minimum check invokes the
explicit `diagnostic_missing_rule` callback. With one valid group present, the
receive-only subset rule supplies the same count evidence. A one-group failure
retains `errors.Is(err, validation.ErrUnsupported)` and additionally identifies
`ErrMissing`. Build/send validation retains its existing rejection behavior.

The consumer independently parses the original top-level groups and verifies
their count, child uniqueness and widths, supported discriminator values and
relative order. Unknown/duplicate types, malformed children, mismatched evidence,
altered paths and unreviewed rules cannot establish a missing-instance quote.
Explicit evidence that fails these checks cannot fall back to older placeholders.

Each absent Subscription-Id produces a 28-byte group containing a four-byte zero
Subscription-Id-Type and an empty Subscription-Id-Data. Two absent groups produce
56 bytes. These are minimum-size examples, not template-valid replacements:
the zero discriminator does not identify the missing business role, and no
subscriber value is synthesized or copied. An actually present empty group keeps
its existing missing-child error. Excess instances retain the existing 5009 path.

Input limits and output byte/count/depth/repetition/field/frame budgets are
checked before allocation. TCP/TLS tests cover zero and either single surviving
subscription in normal/emergency slots, exact bytes and return envelopes,
non-delivery, held charges, valid follow-up traffic and retirement. Default-close
policy remains available. All six existing request templates have offline
omission acceptance.

Generator v23 accepts the callback declaration only on bounded, required,
repeated Grouped fields of requests with a separate handwritten rule. It is
receive-only and runs at the existing structural-minimum failure point.
A nil callback result preserves the original failure; it cannot bypass minimum
validation. Callback code is trusted schema code, not an automatically proved
rule. Synthetic external-module tests exercise the actual generated decoder.

```sh
CGO_ENABLED=0 go test . ./internal/cardinalityavp ./internal/codegen -run 'Test(Cardinality|RejectCardinality)' -count=20
CGO_ENABLED=1 go test -race . ./internal/cardinalityavp ./internal/codegen -run 'Test(Cardinality|RejectCardinality)' -count=20
CGO_ENABLED=0 go test ./internal/cardinalityavp -run '^$' -fuzz '^FuzzCardinalityQuote$' -fuzztime=30s -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to slice `0.1.0-slice.6`. Generator v23 changes all eleven
definition digests and dependent Bundle fingerprints; public signatures/layouts
and operations v9 remain unchanged. Regenerate bindings and retain original
builds for old recovery records. Existing published tags remain immutable.
The library stays pure Go and introduces no persistence.

## SCAP1 ordered parameters

Batch 110 adds bounded 5005 diagnostics for the fixed SCAP1 event dialect's
eleven Service-Parameter-Info groups. The existing minimum and maximum remain
eleven, with ordered Parameter-Type values
`0, 1, 5, 8, 14, 16, 17, 18, 2, 3, 13`.

At the existing minimum failure, an explicit receive-only callback accepts only
a strict ordered subsequence of these types with valid children. The consumer
independently checks the original Vendor193 groups and observed count.
Duplicate, unknown, negative or reordered types cannot prove a missing-instance
diagnostic. Present empty groups retain their missing-child errors; excess groups
retain 5009. Complete requests still undergo all existing fixed-value and
cross-parameter checks. Build/send validation is unchanged.

Each absent group contributes 40 bytes: a Vendor193 group with a zero-filled
four-byte Parameter-Type and a present empty Parameter-Value. All eleven absent
groups produce 440 bytes. This minimum example neither identifies a missing role
nor supplies a valid business value. Output limits include the outer Failed-AVP
container; input and complete-answer budgets apply separately.

All 2047 proper ordered subsets are checked offline. TCP/TLS covers no groups,
only the first group, missing the first group and missing the last group, each
with ordinary and emergency slots. Tests check exact diagnostics, envelopes,
held resources, non-delivery, follow-up, retirement and default-close behavior.
The same bounded fuzz target covers both named rules.

SCAP1 advances to `0.1.0-slice.3`; its definition digest and dependent Bundle
fingerprints change. Generator v23, operations v9 and public signatures/layouts
remain unchanged. Earlier recovery records require their original build.
Other conditional, cardinality, domain and handshake diagnostics remain separate.
