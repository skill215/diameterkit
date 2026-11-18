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

SCAP1's eleven ordered service parameters and other conditional/domain/handshake
diagnostics remain separate at this batch.
