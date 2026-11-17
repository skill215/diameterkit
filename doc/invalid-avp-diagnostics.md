# Invalid and unsupported AVP diagnostics

This guide describes the unreleased batch 98 development checkout. With
`BaseErrorFailures` enabled for a known inbound application command, generated
decoders now provide bounded instance evidence for selected 5004 and 5001 errors.
Production code remains pure Go, with no persistence boundary.

## Supported mappings

| Detection | Result | Evidence |
| --- | --- | --- |
| Generated unknown AVP with M set | 5001 | The actual first unsupported instance encountered |
| Generated declared unsigned enumeration value outside its allowed set | 5004 | The actual decoded value instance |
| Generated UTF-8 or Identity validation failure | 5004 | The original string bytes |
| Generated TBCD value failure within its supported byte-length bounds | 5004 | The original TBCD bytes |
| Unknown AVP with M clear | No new failure | Existing opaque extra-field handling |
| Bare/manual business or cross-field validation error | No new mapping | Existing fail-closed behavior |

The selected result codes require the offending AVP in Failed-AVP. This is based on
[RFC 6733 sections 7 and 7.1.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.1.5).
Only the first detected error is returned; field-value checks now run while the
decoder still knows the exact instance, so a request with multiple faults may
report a different first error than an earlier generator version.

Unknown payloads remain opaque. They are never guessed to be Grouped. Known
Grouped ancestors carry actual offsets and flags, including repeated parents.
The quote retains the offending leaf's complete header, data and padding, inside
minimal ancestor headers. It does not substitute a sibling with the same key.
The copied quote owns its bytes; private error metadata retains no payload or
request-buffer reference. Existing sentinel errors and FieldError paths remain
available through error unwrapping.

## Limits and exclusions

The existing bounded instance walker is shared with the
[5009 extension](excess-avp-diagnostics.md). It verifies framing, offsets, keys,
flags and the metadata path before allocation, and charges depth/count for the
Failed-AVP wrapper. Full response grammar, known diagnostic descendants, reply
size and resource reservations are checked before transmission. Both ordinary
inbound admission and emergency rejection use the same mapping.

Untrusted Session-Id or Proxy-Info, illegal ancestor flags, malformed framing,
resource exhaustion and insufficient reply budgets still close without a partial
answer. `CloseOnFailure` remains the default policy. A manual `ErrUnsupported`
does not establish an unknown Mandatory AVP, and a manual `ErrField` does not
establish a safely quotable isolated value.

Length failures do not acquire a guessed 5004 mapping. The later
[batch 100 extension](length-avp-diagnostics.md) adds selected generated 5014
evidence for fixed-width values, Address and TBCD. Nonempty-octet and manual
domain rules, Proxy/VSAI internals, handshake-specific mappings, malformed-header
5014 and other omission diagnostics remain separate work. Bounded advisory
Redirect 3006 is now covered by [batch 101](redirect-handling.md).
Protocol definition inputs and their declared scope are unchanged.

## Compatibility and acceptance

Generator v16 changes DefinitionDigest constants and Bundle fingerprints. Public
signatures and layouts remain unchanged. Earlier recovery records must be settled
with their original reviewed build; never rewrite fingerprints to import them.
Published core `v0.1.0-rc.2` and adapter `v0.1.0-rc.1` remain unchanged. The planned
next core candidate is still `v0.2.0-rc.1`, subject to separate publication and
immutable cold-install acceptance.

TCP/TLS acceptance covers both result codes, distinct unknown instances, vendor
padding, repeated fields/parents, deep grouping, emergency reserves, no invalid
delivery, envelope refusal, resource retirement and non-M opaque delivery.
Independent evidence tests and the eighth bounded fuzzer exercise mismatched
paths, offsets, flags, result codes, sentinels, truncation and size bounds.

```sh
CGO_ENABLED=0 go test . ./internal/excessavp ./internal/codegen \
  -run 'Test(Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./internal/excessavp ./internal/codegen \
  -run 'Test(Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
CGO_ENABLED=0 go test ./internal/excessavp -run '^$' \
  -fuzz FuzzInvalidQuote -fuzztime=30s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

See the [batch 98 receipt](../spec/evidence/batch-98-invalid-diagnostics.json) and
[current remaining gaps](remaining-gaps.md). These local and hosted checks do not
establish real target interoperability or application recovery permission.

[Batch 102](manual-group-diagnostics.md) extends selected mappings into handwritten
Proxy/VSAI decoders and preserves exact-instance evidence across generated
parents. Unsafe Proxy envelopes still close; conditional and variable omissions
remain separate. Generator v18 changes generated digests and Bundle fingerprints.
