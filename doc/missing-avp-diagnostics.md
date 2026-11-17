# Bounded missing-AVP diagnostics

This describes the unreleased development checkout after batch 96. The published
core `v0.1.0-rc.2` does not include this extension. The library remains pure Go and
keeps no persistent request, diagnostic or subscriber store.

## Enabled behavior

An inbound command must explicitly select `BaseErrorFailures`. The zero-value
`CloseOnFailure` policy retains its existing behavior. For a generated decoder's
first observed missing required scalar, the enabled policy can return an E answer
with Result-Code 5005 and a single Failed-AVP hierarchy:

| Definition field | Missing value |
| --- | --- |
| `u32`, `i32`, `time`, with `min: 1` | Four zero bytes |
| `u64`, `i64`, with `min: 1` | Eight zero bytes |
| A known enclosing Grouped instance | Its actual key and flags, containing only the next diagnostic child |

The existing five top-level Base placeholders remain supported. A field's missing
flags come from its compiled definition. Parent flags come from the exact instance
whose child decoder failed. The implementation does not search sibling groups for
a matching key, copy subscriber values, or validate the zero placeholder as an
ordinary application value. This follows the omission and ancestor-chain rules in
[RFC 6733 section 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5).

For example, missing Value-Digits in the combined binding's standard money branch
produces `456 / 437 / 413 / 445 / 447`, ending in eight zero bytes. Missing
Feature-List in the second S6c Supported-Features instance produces
`10415:628 / 10415:630`, with that second parent's flags. The diagnostic contains
one minimal hierarchy; it does not encode an occurrence ordinal or siblings.

## Bounds and refusal

The decoder carries keys, flags and fixed widths in private error metadata, with
no reference to request bytes. Generated code includes this metadata locally so
external generated packages continue to compile using public dependencies only.
The receiver freezes at most 64 metadata nodes once before encoding. Metadata is
trusted schema-callback output, not evidence authenticating a remote peer.

The quote accounts for the Failed-AVP wrapper in depth and AVP-count checks. Field,
frame and response-byte limits apply before allocation; the complete response must
also pass the existing encoder, response-size and reservation gates. The existing
bounded rejection slots and resource accounting remain in force.

The connection closes when a safe response cannot be built. This includes
untrusted/duplicate Session-Id, invalid Proxy-Info, illegal parent flags, exhausted
limits, or an unsupported omission. Missing whole groups, variable-length fields
beyond explicitly supported minima, other manual/conditional rules, and
cardinality minima greater than one do not gain guessed placeholders.
[Batch 103](group-constraint-diagnostics.md) adds selected VSAI conditional pairs
and minimum Proxy omission examples. A bare public
`validation.FieldError` path alone cannot create a new deep diagnostic. Batch 97
adds [generated first-excess 5009 diagnostics](excess-avp-diagnostics.md); manual
cardinality and other unproven mappings remain separate. See the
[current gap inventory](remaining-gaps.md) for later selected additions.

## Compatibility and recovery

Generator v14 changes the 11 generated DefinitionDigest constants and therefore
Bundle fingerprints, while leaving message fields and profile slice versions
unchanged. Public signatures and struct layouts are unchanged. The API snapshot
review covers source hashes, those constants and the generator-version banner.
Batch 97 advances the generator to v15 with another fingerprint transition.

Recovery records tied to an earlier fingerprint must continue to fail import under
the new Bundle. Drain or settle outstanding records under their original reviewed
build before switching; do not rewrite fingerprints to bypass import checks. This
intentional recovery boundary requires a new core minor under the project's v0
policy. The planned next core candidate is `v0.2.0-rc.1`; no tag is published by
this batch. The adapter keeps its existing published requirement until a new core
is separately published and installation-tested. Existing immutable tags remain
unchanged. See the [version policy](version-install-policy.md).

## Reproduce acceptance

Use the pinned `.go-version` toolchain on Linux/amd64:

```sh
CGO_ENABLED=0 go test . ./internal/missingavp ./internal/codegen \
  -run 'Test(Missing|RejectDeep|Generated|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./internal/missingavp ./internal/codegen \
  -run 'Test(Missing|RejectDeep|Generated|ProductionImport)' -count=20 -timeout=5m
python3 spec/tools/release_check.py --suite static
python3 spec/tools/release_check.py --suite pure
python3 spec/tools/release_check.py --suite race
python3 spec/tools/release_check.py --suite fuzz
```

TCP/TLS tests cover vendor and non-vendor missing leaves, signed 32/64-bit values,
Time, top-level scalars, repeated parent instances, four nested ancestors, exact
zero payloads, reply limits, no application delivery and resource retirement.
Negative cases retain close behavior. The bounded quote fuzzer joins the existing
checkout fuzz suite. These checks do not establish real target interoperability.

See [machine evidence](../spec/evidence/batch-96-missing-diagnostics.json) and
[remaining requirements](../spec/progress.json).
